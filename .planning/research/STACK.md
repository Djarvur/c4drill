# Technology Stack — v1.10 Model Composition

**Project:** C4Drill
**Milestone:** v1.10 Model Composition (INCLUDE, Unit Templates, TOML Ergonomics, Reference field)
**Researched:** 2026-08-08
**Confidence:** HIGH

## Executive Summary

**Add ZERO new external dependencies.** All four v1.10 features are implementable with the existing stack plus small hand-rolled helpers. The codebase is small (one `Model` struct, one `Unit` struct, one `Link` struct — all known shapes), which makes reflection-based libraries both unnecessary and riskier than targeted code. The one temptation (a config-composition lib like viper/koanf for INCLUDE) is overkill: the merge is a union of two maps plus an append of two slices, not the layered key/value resolution those libraries exist for.

**One housekeeping action:** bump `github.com/pelletier/go-toml/v2` from `v2.2.4` → `v2.4.3` (latest; addresses parser fixes since Feb 2025, no breaking API changes for our usage).

---

## New Dependencies Recommended

**None.** The full feature set maps to hand-rolled code at known integration points.

### Dependency considered and the only candidate worth keeping in mind

If, during implementation, the INCLUDE merge logic turns out to need arbitrary-depth map merging (it currently does not — see analysis below), the only library that would fit C4Drill's "no framework, no bloat" philosophy is **koanf** (`github.com/knadh/koanf/v2` + `parsers/toml`). It is the lightweight alternative to viper with native multi-file merge. This is listed as a contingency, **not** a recommendation — see "Rejected/Overkill" for why it stays out unless the merge shape changes.

---

## Hand-Roll Recommendations

### (a) TOML deep-merge / multi-file composition — HAND-ROLL

**Concern:** Assemble one diagram from multiple TOML files (INCLUDE directive, parse-then-merge approach).

**Why hand-roll:**
1. **go-toml/v2 has no merge capability.** Verified via Context7 resolve and the release notes through v2.4.3 (2026-07-05): the library is a pure parser/encoder. No `Merge`, no compose-config, no struct-overlay. It can re-marshal/unmarshal (which the existing parser already does at `parser.go:69-76` and `parser.go:176-185`), but composing two parsed `Model` structs is not in scope.
2. **The merge is trivial and known-shape.** After parsing each included file into a `parser.Model` (struct at `parser/parser.go:35-42`), the composition is exactly:
   - `Units map[string]*Unit` — union by key (collision = error, or override per design decision)
   - `UnitOrder []string` — append in include order
   - `Properties` — field-wise overlay (included files' properties merged onto base)
   That is ~30 lines of code in a new `internal/compose` (or `internal/parser/include.go`) package. No third-party can beat that in clarity.
3. **Cycle detection and `include_once` are bespoke logic anyway** — no library handles the `include_once` directive or path-canonicalization for relative includes. These must be hand-written regardless.

**Rough approach:**
- New `internal/parser/include.go` (or `internal/compose/compose.go`).
- `ResolveIncludes(rootPath string) (*Model, error)`:
  - Parse root file via existing `ParseFile` (`parser.go:322`).
  - Walk a synthetic `[include]` / `[[include]]` section (extracted before unit parsing, similar to how `[properties]` is special-cased at `parser.go:128`).
  - Maintain `seen map[string]bool` keyed by absolute file path (canonicalize with `filepath.Abs` + `filepath.Clean`) → `include_once` semantics fall out for free.
  - Recurse on each included path; on revisit, either skip (include_once default) or error (cycle / explicit duplicate).
  - Merge function: union `Units` maps, concatenate `UnitOrder` slices, overlay `Properties` (later includes win, or base wins — design decision for discuss phase).
- Integration point: insert between `parser.ParseFile` and `validator.Validate` in `cmd/c4drill/root.go:112-118`. The pipeline comment in `PROJECT.md:42` already specifies `include → template-expand → relative-peer-resolve → validate → render`.

### (b) Go deep-copy for Unit Templates — HAND-ROLL (manual Clone)

**Concern:** Unit templates need to deep-copy a template `Unit` before substituting params.

**Why hand-roll (manual `Clone()` method):**

The `Unit` struct (`internal/model/unit.go:41-72`) has:
- Scalar fields: `Type`, `Name`, `Description`, `Technology`, `Color`, `Style`, `Border`, `Edges`, `Width`, `Height` — trivial value copy.
- `Expanded []string` — copy with builtin `slices.Clone` (Go 1.21+, we're on 1.26.1).
- `Links []Link`, `LinksFrom []Link` — copy with `slices.Clone`; `Link` is a value type with no pointers, so a shallow slice clone is sufficient.
- `SubunitOrder []string` — `slices.Clone`.
- `Subunits map[string]*Unit` — **this is the recursive case.** Iterate and clone each `*Unit` recursively.
- `Link.Mirror` (`internal/model/link.go:67`) is `toml:"-"` and **must survive the copy** — it's read by the validator (`internal/validator/index.go:80`) and graph builder (`internal/graph/builder.go:438`). This is the decisive argument against the popular reflection-based libraries:

| Library | Unexported `Mirror` field | Verdict |
|---|---|---|
| `mohae/deepcopy` | **skipped** (documented: "Unexported field values can't be accessed and will be skipped") | Rejected — silently drops `Mirror` |
| `jinzhu/copier` | partial/field-name based, limitations noted | Rejected — unreliable |
| `ulule/deepcopier` | limited, niche | Rejected — same reflect limitation |
| `encoding/json` round-trip | **lost** — JSON only marshals exported fields, and `Mirror` is unexported | Rejected — loses `Mirror` (and `SubunitOrder` which is also `toml:"-"`) |
| `encoding/gob` round-trip | **preserved** (gob uses reflection on unexported fields) | Works but heavy: ~10x slower than reflect, needs `gob.Register` for the `*Unit` indirection, and forces all fields to be gob-encodable. Overkill for one struct. |
| **manual `Clone()`** | full control, trivially correct | **Recommended** |

**Rough approach** (new method on `Unit`, file `internal/model/unit.go`):
```go
func (u *Unit) Clone() *Unit {
    if u == nil { return nil }
    c := *u                              // scalar copy + slice headers
    c.Expanded = slices.Clone(u.Expanded)
    c.Links = slices.Clone(u.Links)      // Link is a value type → sufficient
    c.LinksFrom = slices.Clone(u.LinksFrom)
    c.SubunitOrder = slices.Clone(u.SubunitOrder)
    if u.Subunits != nil {
        c.Subunits = make(map[string]*Unit, len(u.Subunits))
        for k, sub := range u.Subunits {
            c.Subunits[k] = sub.Clone()  // recursive
        }
    }
    return &c
}
```
~15 lines. `slices.Clone` is already a project idiom (see `slices.Contains` at `parser.go:310`). No new dependency, no reflect gotchas, preserves `Mirror` and `SubunitOrder` by construction.

**Integration point:** new `internal/parser/template.go` (or `internal/template/expand.go`) — the expansion step deep-copies the template Unit, then substitutes params into string fields.

### (c) Named-parameter substitution — HAND-ROLL (simple replacer)

**Concern:** Substitute `${name}` params into Unit string fields during template expansion.

**Why hand-roll:**
- The substitution is **field-level string fill**, not logic. No conditionals, no loops, no partial templates, no pipelines. `text/template` would be 95% unused capability and brings a real attack surface (template injection — `text/template` can call methods, access globals).
- The `${name}` syntax (or `{{name}}` — pick one in discuss phase) is a single `strings.NewReplacer` or one regex. `strings.NewReplacer` is the simplest correct option and is allocation-cheap for the typical 2-4 params.

**Rough approach:**
```go
func substituteParams(s string, params map[string]string) string {
    pairs := make([]string, 0, len(params)*2)
    for k, v := range params {
        pairs = append(pairs, "${"+k+"}", v)
    }
    return strings.NewReplacer(pairs...).Replace(s)
}
```
Then walk the cloned `Unit` and apply to each string field (`Name`, `Description`, `Technology`, `Color`, `Style`, `Border`, `Edges`, and the `Link` string fields). Walking is either a switch on field name (explicit, ~12 lines) or `reflect.Value` iteration over string fields (generic, ~10 lines). Explicit is safer and self-documenting for one struct shape.

**Integration point:** inside the template-expansion step, immediately after `Unit.Clone()`.

### (d) Identifier humanization (camelCase → "Camel Case") — HAND-ROLL

**Concern:** Optional `name` field — when absent, derive a human name from the TOML section key (e.g., `userId` → "User Id" or "User ID"; `gRPC` → "gRPC" not "Grpc").

**Why hand-roll:**
- **No library fits the actual requirement.** Two candidates investigated:
  - `stoewer/go-strcase` — converts between cases (camel/snake/kebab) but has **no acronym awareness** and no "to space-separated words" function. `gRPC` → would mangle.
  - `iancoleman/strcase` — has `ConfigureAcronym` (so "ID" stays "ID"), BUT acronym support only applies to `ToCamel`/`ToLowerCamel`. For the **reverse** direction (splitting into human words) the only path is `ToScreamingDelimited(s, ' ', "", ...)` which yields **"ORDER ID"** (all caps) — not title case "Order Id". There is no `ToWords`/`ToHuman` function. So even with the library, you hand-roll the title-casing on top.
- The input domain is tiny and known: TOML section identifiers (`[userId]`, `[rbacSvc]`, `[apiGateway]`). The rules are: split on case transitions and on `_`/`-`, capitalize the first letter of each token, preserve runs of uppercase as a unit (handles `ID`, `URL`, `gRPC` if we keep a small acronym map).
- A ~25-line function with a small acronym table (`ID`, `URL`, `gRPC`, `IP`, `API`, `UI`, `DB`, `SaaS`, ...) covers the realistic vocabulary and gives full control over edge cases (e.g., `gRPC` stays lowercase-leading).

**Rough approach** (new `internal/parser/humanize.go` or a helper in `internal/model`):
```go
var acronyms = map[string]string{"ID": "ID", "URL": "URL", "gRPC": "gRPC", "API": "API", ...}

func HumanizeIdentifier(s string) string {
    // split on '_', '-', and lower→Upper transitions
    // for each token, lookup acronym map; else Title-case first rune
    // join with spaces
}
```

**Integration point:** post-parse resolver step (after INCLUDE and template expansion, before validate). When a `Unit.Name` is empty, set it from `HumanizeIdentifier(sectionKey)`. Section keys are available where units are built in `parser.go:80-93` and `parser.go:200-216`.

### (e) Reference field + relative peer path — trivial, no dependency

- **Reference field** is one new `string` field on `Unit` (`internal/model/unit.go`) plus an SVG anchor in the renderer. Pure data + a render branch. No library.
- **Relative peer resolution** is string manipulation on `Link.Peer` (`internal/model/link.go:45`) using `path.Join` / `path.Dir` from the stdlib. Post-parse resolver, no dependency.

---

## Rejected / Overkill

| Library | Why not |
|---|---|
| **spf13/viper** | Config framework for env+flags+file layered resolution. Brings a large transitive dep tree (maps, cast, crypt, fsnotify, pflag, etc.). C4Drill needs none of that — we merge two parsed `Model` structs, not resolve a key across N sources. Also: viper's multi-file story is limited (single-file primary model). |
| **knadh/koanf** | Lighter than viper and does support multi-file merge cleanly, but it operates on generic `map[string]interface{}` config blobs, not typed Go structs. Adopting it would mean either (a) re-implementing C4Drill's typed `Model`/`Unit`/`Link` as koanf maps (throwing away the parser we have), or (b) using koanf only for the merge then re-unmarshaling into structs (more code than the 30-line hand-rolled merge). Contingency only if merge requirements grow arbitrary-depth map semantics. |
| **mohae/deepcopy** | Reflection-based; **silently skips unexported fields** (documented). Would drop `Link.Mirror` (`link.go:67`) and `Unit.SubunitOrder` (`unit.go:69`, `toml:"-"`). Both are load-bearing for validation (`validator/index.go:80`) and graph building (`builder.go:438`). Also unmaintained (~6 years stale). |
| **jinzhu/copier** | General struct-to-struct copier, not a pure deep-copy lib; unreliable on unexported fields; brings reflect overhead for a struct we fully control. |
| **ulule/deepcopier** | Niche, low activity, same reflect limitation on unexported fields. |
| **encoding/gob round-trip** | Works (preserves unexported fields) but ~10x slower than reflect, requires `gob.Register` for the `*Unit` pointer indirection, and forces every field to be gob-encodable. Solves a 15-line problem with a sledgehammer. |
| **encoding/json round-trip** | Loses unexported fields (`Mirror`, `SubunitOrder`). Non-starter. |
| **text/template** (stdlib) | Correct but overkill: brings template-injection surface and 95% unused features (conditionals, loops, pipelines, funcs) for what is a flat `${name}` → value string fill. `strings.NewReplacer` is simpler and safer. |
| **stoewer/go-strcase** | No acronym awareness, no to-words function. Wrong shape for the requirement. |
| **iancoleman/strcase** | Closest fit (has `ConfigureAcronym`) but acronym support is forward-only (`ToCamel`/`ToLowerCamel`); reverse direction yields screaming-case ("ORDER ID"), not title-case display. Would still require hand-rolled title-casing on top — so adds a dependency for no net simplification. |
| **gobuffalo/validate** / any validation lib | C4Drill already has a purpose-built validator (`internal/validator/`) with domain rules. New features add rules, not a framework. |

---

## Version Verification

| Dependency | Current (go.mod) | Latest verified | Action | Source |
|---|---|---|---|---|
| Go | 1.26.1 | 1.26.x | Keep | [go.mod:3](../../go.mod) |
| `github.com/pelletier/go-toml/v2` | v2.2.4 | **v2.4.3** (2026-07-05) | **Bump** (minor; no API breaks for our usage of `Unmarshal`/`Marshal`/`unstable.Parser`) | [GitHub releases](https://github.com/pelletier/go-toml/releases) — v2.4.3 Jul 5 2026, v2.4.2 Jun 24, v2.4.1 Jun 22, v2.4.0 Jun 17, v2.3.1 May 2. Notable: v2.3.0 added "UnmarshalText fallbacks to struct unmarshaling for tables and arrays". No merge/compose feature in any release. |
| `github.com/goccy/go-graphviz` (→ `onokonem` replace) | v0.2.10 (replace to onokonem commit `f364b5235161`) | unchanged | Keep | [go.mod:7,28](../../go.mod) |
| `github.com/spf13/cobra` | v1.10.2 | current line | Keep | [go.mod:9](../../go.mod) |
| `github.com/stretchr/testify` | v1.11.1 | current line | Keep | [go.mod:10](../../go.mod) |
| `github.com/agnivade/levenshtein` | v1.2.1 | unchanged | Keep (used by validator "did you mean" hints) | [go.mod:6](../../go.mod) |

**go-toml/v2 verification detail (Context7 + web):** Context7 resolve returned `/burntsushi/toml` and `/iancoleman/strcase` but not `/pelletier/go-toml` as an indexed library, so version verification fell back to the GitHub releases page (authoritative). Confirmed: v2 line is actively maintained, latest v2.4.3, the API surface C4Drill uses (`toml.Unmarshal`, `toml.Marshal`, `unstable.Parser`/`unstable.Table`/`expr.Key`/`keyIter.Node().Data` at `parser.go:57,69,74,107-121,176,183`) is stable across v2.2.4 → v2.4.3. Bump is low-risk.

---

## Integration Summary (where each piece plugs in)

| Feature | New file(s) | Touches | Approach |
|---|---|---|---|
| INCLUDE | `internal/parser/include.go` (or `internal/compose/`) | `cmd/c4drill/root.go:112` (insert resolve step); `parser.go:35` (Model gains an Includes field or pre-parse strip) | Hand-rolled recursive parse-merge + cycle/once via `map[absPath]bool` |
| Unit Templates | `internal/parser/template.go` (or `internal/template/`) | `model/unit.go` (add `Clone()`); pipeline between parse and validate | Hand-rolled `Clone()` + `strings.NewReplacer` param substitution |
| Humanize name | `internal/parser/humanize.go` | `parser.go:80-93,200-216` (set Name when empty) | Hand-rolled splitter + small acronym table |
| Relative peer | `internal/parser/resolve.go` (or fold into include.go) | `model/link.go:45` (Peer string) | `path.Join`/`path.Dir` post-parse resolver |
| Reference field | `model/unit.go` (add field); renderer | SVG anchor render branch | Pure data + render branch |
| go-toml bump | — | `go.mod:8` | `go get github.com/pelletier/go-toml/v2@v2.4.3` |

**Net new external dependencies: 0.** Net new code: ~150-200 lines across 3-4 new files in `internal/parser` (or split into `internal/compose` + `internal/template`).

---

## Sources

- Codebase: [parser.go](../../internal/parser/parser.go), [unit.go](../../internal/model/unit.go), [link.go](../../internal/model/link.go), [properties.go](../../internal/model/properties.go), [go.mod](../../go.mod), [root.go](../../cmd/c4drill/root.go), [PROJECT.md](../PROJECT.md)
- [pelletier/go-toml releases](https://github.com/pelletier/go-toml/releases) — v2.4.3 latest, no merge feature
- [pkg.go.dev go-toml/v2](https://pkg.go.dev/github.com/pelletier/go-toml/v2)
- [knadh/koanf](https://github.com/knadh/koanf) — contingency only
- [Viper vs Koanf comparison](https://itnext.io/golang-configuration-management-library-viper-vs-koanf-eea60a652a22)
- [mohae/deepcopy pkg.go.dev](https://pkg.go.dev/github.com/mohae/deepcopy) — "unexported fields skipped"
- [r/golang: Deep copying done right](https://www.reddit.com/r/golang/comments/e2pd8y/deep_copying_done_right/) — library unexported-field comparison
- [tonybai: Understand deep copy in Go (2024)](https://tonybai.com/2024/09/28/understand-deep-copy-in-go/)
- [JSON vs gob in Golang (2025)](https://blog.vitalvas.com/post/2025/07/23/json-vs-gob-in-golang/)
- [stoewer/go-strcase](https://github.com/stoewer/go-strcase) — no acronym handling
- [iancoleman/strcase pkg.go.dev](https://pkg.go.dev/github.com/iancoleman/strcase) — acronym support is forward-only (ToCamel/ToLowerCamel)
- Context7: `/iancoleman/strcase` docs (acronym config API verified); `/pelletier/go-toml` not indexed in Context7, fell back to GitHub releases for version authority

---
*Stack research for: C4Drill v1.10 Model Composition milestone*
*Researched: 2026-08-08*
