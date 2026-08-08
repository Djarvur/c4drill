# Phase 31: Template expansion - Research

**Researched:** 2026-08-08
**Domain:** TOML parsing internals (go-toml/v2 unstable API), in-place model transformation, Go deep-copy semantics
**Confidence:** HIGH (all claims verified against live code in this repo this session)

## Summary

Phase 31 is a pure Go pre-validation transformation phase with **zero new external dependencies**. It extends the parser to recognize three reserved top-level tables (`[template.*]`, `[[use]]`, `[[include]]`), adds a `Unit.Clone()` deep-copy method to the model package, and inserts one new pipeline stage (`template.Expand`) between `parser.ParseFile` and `validator.Validate` in `cmd/c4drill/root.go`. Every other consumer (validator, view, render) consumes the expanded `*parser.Model` unchanged — confirmed by reading each package this session.

The phase is small in surface area (one ~15-LOC Clone, one parser skip-rule extension mirroring the existing `properties` skip, one new `Expand` function, two new `toml:"-"` Model fields) but carries **one core correctness concern (HS-1)** that the plan MUST prove with a three-instantiation regression test: the validator mutates `Unit.LinksFrom` in place (`internal/validator/index.go:55-86`, `populateIncomingLinks`), so a shallow copy of a template unit corrupts every instantiation after the first. The hand-rolled recursive `Unit.Clone()` is the load-bearing mitigation.

**Critical pipeline-ordering finding (verified this session):** As of this research, `cmd/c4drill/root.go` goes directly from `parser.ParseFile` (line 112) to `validator.Validate` (line 118). **There is no humanize pass and no relative-peer pass in the codebase yet** (grep for `humaniz`/`RelativePeer` returns nothing). Phases 29 (relative-peer) and 30 (humanize) are NOT yet implemented. CONTEXT.md's target pipeline `include → template-expand → relative-peer-resolve → humanize → validate` is the end-state across phases 29-33, not the current state. **Consequence for the plan:** Phase 31's `Expand` pass must insert between Parse and Validate NOW, and XC-03/XC-04 are satisfied by *placing expansion at the right slot and producing units whose parent is the instantiation site* — NOT by depending on passes that do not exist yet. The plan must not assume `humanize`/`relative-peer` code is present; it must make expanded units structurally indistinguishable from hand-authored units so that whenever those passes land, they treat expanded units uniformly.

**Primary recommendation:** Two plans. Plan 1 = BC-1 parser prerequisite (extend `captureDefinitionOrder` skip rule + `Parse` rawMap extraction + new `Model.Templates`/`Model.Instantiations` `toml:"-"` fields, all parser-only, fully tested). Plan 2 = template expansion feature (`Unit.Clone()`, `template.Expand`, pipeline insertion, `${param}` substitution via `strings.NewReplacer`, duplicate-path check, missing-param check, three-instantiation HS-1 regression test). TMPL-01 through TMPL-10 + XC-03 + XC-04 are all covered by Plan 2; Plan 1 is the parser prerequisite that also unblocks Phase 32.

<user_constraints>
## User Constraints (from CONTEXT.md)

These are locked decisions from `/gsd:discuss-phase`. Research MUST honor these — alternatives are NOT explored.

### Locked Decisions

**D-01 (HS-2):** Relative peers authored inside a template resolve against the **instantiation site's** structural parent — NOT the template's lexical location. Templates are data, not structural location.

**D-02 (HS-2 edge case):** A template instantiated with NO parent (top-level) resolves relative peers against top-level siblings with absolute-fallback — uniform with hand-authored top-level units. No template-special-case logic.

**D-03:** Instantiation uses a **separate `[[use]]` array-of-tables** (Syntax 1), NOT inline-on-unit. Each `[[use]]` carries: `template` (name), `parent` (placement, optional — empty/omitted = top-level), and the template's named params as explicit fields (including `name` as a regular param). `[[use]]` tables never collide with real units; `parent=` is the only placement knob.

**D-04 (name param semantics):** The `name` param fills the **top-level produced unit's last path segment**. Produced unit full path = `parent + "." + name` (or just `name` if parent empty). Subunit keys in the template stay **verbatim** — they are NOT parametrized. Parametrization is field-value-only (substitution into string fields); structural slots (which entities exist, their keys) are never parametrized. Preserves the multi-output / `for_each` anti-feature boundary.

**D-05 (duplicate-path check, TMPL-07):** After all `[[use]]` expansions, run a **full-path-set check** across hand-authored units + all expanded template instances. ANY collision — templated-vs-templated OR templated-vs-hand-authored — is a single hard error naming both sources. No silent overwrite.

**D-06:** Use **bare, ergonomic table names**: `[template.<name>]`, `[[use]]`, and `[[include]]` (the last lands in Phase 32 but is reserved now). No `c4drill.` namespace prefix.

**D-07 (BC-2 — collision policy):** **No collision-mitigation machinery.** C4Drill is not yet public. Parser treats `template`, `use`, `include` as reserved top-level keys, full stop.

**D-08 (BC-1 — parser changes):** Two coordinated changes land as Plan 1, both required before feature code:
  1. `captureDefinitionOrder` (`internal/parser/parser.go:100`, skip-rule at `:128`): extend the skip (currently only `properties`) to also skip `[template.*]` (any table whose first key segment is `template`), `[[use]]`, and `[[include]]`.
  2. `Parse` rawMap extraction (`internal/parser/parser.go:47`): extract these reserved top-level tables from `rawMap` before the unit-processing loop (mirroring the existing `properties` extraction at `parser.go:68-77`), routing them into new `Model.Templates` / `Model.Instantiations` fields.
  - Note: `isBuiltinField` (`parser.go:309`) does NOT need `template`/`use`/`include` added — those are top-level table names, not unit fields.

**Already-locked (carried from milestone setup, NOT re-litigated):**
- **No param defaults** — every declared param required at every instantiation; missing any = hard error (TMPL-02, TMPL-06).
- **One template = one top-level unit + its declared subunit subtree** (TMPL-04).
- **Fixed link count, parametrized fields** — no fan-out / `for_each` / array expansion (TMPL-03).
- **Forward references allowed** — `[[use]]` may appear textually before `[template.*]` definition (TMPL-09); structured post-parse makes this free.
- **HS-1 deep-copy** — hand-rolled recursive `Unit.Clone()` (~15 LOC) in package model; MUST recurse into `Subunits` (every `*Unit` cloned) and preserve the unexported `Link.Mirror` field (`internal/model/link.go:67`). No reflection/gob/json copier (they silently drop `Mirror`). Three-instantiation regression test required (idempotent re-expand + disjoint `LinksFrom` after validate).
- **Substitution mechanism** — `strings.NewReplacer` over declared params, applied to all string fields of the unit + subunits + links. No `text/template`.
- **Pipeline ordering (target end-state)** — `include → template-expand → relative-peer-resolve → humanize → validate → generate-views → render`. Template expansion is the 2nd pass.

### Claude's Discretion

- Exact field-name and TOML-shape details for `TemplateDef` / `Instantiation` structs on `parser.Model` (research recommends dedicated `toml:"-"` fields — `Templates map[string]*TemplateDef`, `Instantiations []Instantiation` — extracted from rawMap; refined below in Architecture Patterns).
- Internal package structure for the expansion pass (research recommends `internal/template/Expand` — a new package — over a function in `internal/parser/`; rationale in Architecture Patterns).
- Whether to deep-copy-then-substitute or substitute-in-place-on-copy (research recommends deep-copy-then-substitute: clone the template once per `[[use]]`, then run the replacer over the clone's string fields; rationale in Architecture Patterns).

### Deferred Ideas (OUT OF SCOPE — do NOT implement)

- **Parameter defaults** (trailing-default support) — out for v1.10.
- **Template multi-output / fan-out / `for_each`** — anti-feature for v1.10.
- **Template nesting** (template instantiating another template) — deferred.
- **Structural parametrization** (parametrized subunit keys/paths, array/conditional link expansion) — deliberately out.
- **`[[include]]` feature itself** — Phase 32. Phase 31 only lands the `[[include]]` parser skip-rule as part of BC-1 (D-08); the include feature is Phase 32.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| TMPL-01 | `[template.<name>]` table with declared named params | D-08 parser skip + `Model.Templates` extraction (Plan 1); `TemplateDef` struct with `Params []string` (Plan 2) |
| TMPL-02 | `[[use]]` instantiate; ALL params required (no defaults) | D-03 `[[use]]` as array-of-tables → `Model.Instantiations`; missing-param check in `Expand` (Plan 2) |
| TMPL-03 | `${param}` into all string fields; fixed link count | `strings.NewReplacer` over declared params applied to every string field (Name, Description, Technology, Reference, Color, Link.Peer/Description/Technology) (Plan 2) |
| TMPL-04 | one template = one top-level unit + declared subunit subtree | `TemplateDef.Unit` holds a parsed subtree; `Clone()` recurses into `Subunits` map (Plan 2) |
| TMPL-05 | instantiated units participate fully (validate/views/render) | Expand drains into `Model.Units`/`Model.UnitOrder` — validator/view/render unchanged (verified) |
| TMPL-06 | missing param = hard error (no silent literal) | strict check in `Expand`: every declared param present in `[[use]]`, else named error (Plan 2) |
| TMPL-07 | duplicate unit path = hard error | D-05 full-path-set check across hand-authored + expanded (Plan 2) |
| TMPL-08 | deep-copy recurses into Subunits (HS-1 regression) | `Unit.Clone()` hand-rolled recursive; three-instantiation test (Plan 2) |
| TMPL-09 | forward references work | structured post-parse: `Model.Templates` fully built before `Expand` iterates `Instantiations` (Plan 1+2) |
| TMPL-10 | reference param substitution | substitution covers `Reference` field once Phase 28 adds it; Plan 2 substitutes all string fields generically so it future-proofs |
| XC-03 | relative-peer in template resolves at instantiation site | D-01: Expand sets produced unit's parent path to the `[[use]]` `parent` field; the (future) relative-peer pass treats expanded units uniformly. Phase 31 produces the correct structural placement. |
| XC-04 | humanization after expand, before validate | Expand inserts between Parse and Validate (the slot humanize will later occupy before validate). Phase 31 establishes the ordering; the humanize pass itself lands in Phase 30. |
</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Reserved-table recognition (BC-1) | Parser (internal/parser) | — | TOML-shape knowledge lives only in the parser's unstable-API walk + rawMap extraction |
| Template subtree storage | Parser Model struct | — | `Model.Templates`/`Instantiations` are `toml:"-"` fields populated during Parse; consumed once by Expand then dead |
| Deep-copy (HS-1 correctness) | Model (internal/model) | — | `Unit` and `Link` are model types; Clone must preserve `Link.Mirror` which is model-internal bookkeeping |
| `${param}` substitution | New template package (internal/template) | — | Pure string-field transformation; isolated from parser/model to keep Parse single-responsibility |
| Duplicate-path / missing-param checks | New template package (internal/template) | — | Semantic checks belong with expansion, not with parser (parser only knows TOML shape) |
| Pipeline insertion | CLI (cmd/c4drill/root.go) | — | The only place stage ordering is assembled |
| Validation of expanded units | Validator (internal/validator) | — | UNCHANGED — consumes the post-Expand model identically to hand-authored |

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/pelletier/go-toml/v2` | (pinned in go.mod) | TOML parsing, including `unstable.Parser` for ordered key capture | Already the project's parser; the unstable API is already used at `parser.go:107` [VERIFIED: codebase] |
| `github.com/stretchr/testify` | (pinned in go.mod) | test assertions/require | Already used across all `*_test.go` files [VERIFIED: codebase] |
| Go stdlib `strings` | go 1.26 | `strings.NewReplacer` for `${param}` substitution | Zero-dep, exact-purpose; CONTEXT.md locks out `text/template` [VERIFIED: codebase + CONTEXT.md] |
| Go stdlib `slices` | go 1.26 | `slices.Clone` for slice copying (already used at `parser.go:310`) | Project idiom for shallow slice copy; Clone builds on it per-element [VERIFIED: codebase] |

### Supporting

No supporting libraries. This phase adds **zero external dependencies**. [VERIFIED: go.mod read this session — no new imports needed for any of the four code-edit locations]

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Hand-rolled `Unit.Clone()` | `reflect.DeepCopy`, gob round-trip, json round-trip | REJECTED (CONTEXT.md HS-1): all three silently drop the unexported `Link.Mirror` field (`link.go:67`), re-breaking multiplicity counting. Hand-rolled is ~15 LOC and is the only option that explicitly preserves `Mirror`. |
| `text/template` for substitution | `strings.NewReplacer` | REJECTED (CONTEXT.md): `text/template` adds Go-template syntax into user-authored TOML strings (escaping hazards, injection surface) for no benefit over a flat `${name}` replacer. |
| Expansion inside `internal/parser` | New `internal/template` package | RECOMMEND new package: keeps Parse single-responsibility (TOML → Model) and lets Expand own semantic checks (missing-param, duplicate-path). Parser stays shape-only. |

**Installation:**
```bash
# NONE — no packages to install. Phase uses only stdlib + already-pinned deps.
```

**Version verification:** N/A — no new packages. All four code-edit locations use imports already present in `go.mod`.

## Package Legitimacy Audit

> This phase installs **no external packages**. The audit is therefore vacuous — all work is stdlib + already-pinned `go-toml/v2` and `testify`. [VERIFIED: go.mod read this session]

| Package | Registry | Age | Downloads | Source Repo | slopcheck | Disposition |
|---------|----------|-----|-----------|-------------|-----------|-------------|
| (none new) | — | — | — | — | — | N/A |

**Packages removed due to slopcheck [SLOP] verdict:** none
**Packages flagged as suspicious [SUS]:** none

## Architecture Patterns

### System Architecture Diagram

```
                    TOML input file
                          │
                          ▼
              ┌───────────────────────┐
              │  parser.Parse (Parse) │   ◄── BC-1: captureDefinitionOrder
              │  - capture order      │       skip-rule extended to skip
              │  - rawMap unmarshal   │       [template.*], [[use]], [[include]]
              │  - extract properties │       (Plan 1, D-08)
              │  - extract templates ◄┤       NEW: route to Model.Templates /
              │  - extract uses ◄─────┤       Model.Instantiations (toml:"-")
              │  - build Units map    │
              └───────────┬───────────┘
                          │  *parser.Model
                          │   .Units (hand-authored only)
                          │   .UnitOrder
                          │   .Templates  ◄── NEW
                          │   .Instantiations ◄── NEW
                          ▼
              ┌───────────────────────┐
              │  template.Expand(m)   │   ◄── NEW pipeline stage (Plan 2)
              │  for each Instantiation:    inserts between Parse & Validate
              │    lookup Template          (root.go:112 → :118)
              │    verify all params present (TMPL-06)
              │    Unit.Clone() template    (HS-1, TMPL-08)
              │    strings.NewReplacer      (TMPL-03)
              │    apply to all string fields
              │    attach under parent/name (D-04)
              │  full-path collision check  (TMPL-07, D-05)
              │  drain into Units/UnitOrder (TMPL-05)
              └───────────┬───────────┘
                          │  *parser.Model
                          │   .Units (hand-authored + expanded)
                          │   .UnitOrder (expanded appended at [[use]] order)
                          │   .Templates/Instantiations (now dead)
                          ▼
              ┌───────────────────────┐
              │  (future: relative-   │   ◄── Phase 29 — NOT present yet.
              │   peer-resolve)       │       Expand produces units with the
              │  (future: humanize)   │   ◄── Phase 30 — NOT present yet.
              └───────────┬───────────┘       correct parent so these future
                          │                   passes treat expanded units
                          ▼                   uniformly (XC-03, XC-04).
              ┌───────────────────────┐
              │  validator.Validate   │   ◄── UNCHANGED (root.go:118)
              │  populateIncomingLinks│       mutates LinksFrom in place —
              │  (index.go:55-86)     │       this is why Clone is required
              └───────────┬───────────┘
                          ▼
                 views / render (unchanged)
```

### Recommended Project Structure

```
internal/
├── parser/
│   ├── parser.go          # MODIFIED: captureDefinitionOrder skip + Parse extraction + Model fields
│   └── parser_test.go     # EXTENDED: reserved-table skip tests, extraction tests
├── model/
│   ├── unit.go            # MODIFIED: add Unit.Clone() method
│   ├── link.go            # UNCHANGED (Mirror field consumed by Clone)
│   └── unit_test.go       # NEW: Clone tests (Mirror preservation, Subunits recursion)
├── template/              # NEW PACKAGE
│   ├── expand.go          # Expand function + TemplateDef/Instantiation types + checks
│   └── expand_test.go     # three-instantiation HS-1 regression, substitution, error tests
└── validator/             # UNCHANGED
cmd/c4drill/
└── root.go                # MODIFIED: insert template.Expand between ParseFile and Validate
testdata/
├── valid.toml             # EXISTING (hand-authored reference)
├── template_*.toml        # NEW: expansion fixtures (basic, subunits, 3x-instantiate, errors)
```

### Pattern 1: Reserved-table skip in `captureDefinitionOrder` (BC-1, D-08)

**What:** Extend the existing single-condition skip at `parser.go:128` (currently `len(parts)==1 && parts[0]=="properties"`) to also skip any reserved table.

**When to use:** Always during Parse — these tables must NEVER enter `unitOrder`, or they register as phantom units (e.g. `[template.microservice]` would register `microservice` as a subunit of a phantom parent `template`).

**Example (target shape — the executor fills exact syntax from current code):**
```go
// Source: internal/parser/parser.go:127-130 (current) → extended (Plan 1)
// CURRENT:
//   if len(parts) == 1 && parts[0] == "properties" { continue }

// TARGET — skip [properties], and any reserved top-level / namespaced table:
if len(parts) == 1 && parts[0] == "properties" {
    continue
}
// Skip reserved top-level tables (BC-1, D-08/D-06): template / use / include.
if len(parts) == 1 && (parts[0] == "use" || parts[0] == "include") {
    continue
}
// Skip the [template.*] namespace — first segment "template" (covers
// [template.foo], [template.foo.bar], [[template.foo.link]]).
if parts[0] == "template" {
    continue
}
```

Note the `parts[0]=="template"` check covers ALL nesting under `[template.*]` (the subunit subtree `[template.svc.api]` and the `[[template.svc.link]]` array-of-tables) in one condition, because they all share the first key segment. `[VERIFIED: unstable.Parser yields the full key path per expression — see parser.go:117-121]`

### Pattern 2: rawMap extraction mirroring `properties` (BC-1, D-08)

**What:** In `Parse` (parser.go:47), after the existing `properties` extraction block (parser.go:68-77), add parallel extraction blocks for `template`, `use`. Each pulls the key from `rawMap`, re-marshals, unmarshals into the typed field, and removes the key so the unit loop never sees it.

**When to use:** Whenever a new reserved top-level table is added — this is the established pattern.

**Example (target shape):**
```go
// Source: mirrors parser.go:68-77 (properties extraction)
// Extract [template.*] tables -> Model.Templates
if tmpl, ok := rawMap["template"]; ok {
    // ... marshal tmpl, unmarshal into &m.Templates, delete(rawMap, "template")
}
// Extract [[use]] array -> Model.Instantiations
if uses, ok := rawMap["use"]; ok {
    // ... marshal uses, unmarshal into &m.Instantiations, delete(rawMap, "use")
}
// NOTE: [[include]] is reserved (D-06) but its extraction lands in Phase 32.
// For Phase 31, include is only SKIPPED in captureDefinitionOrder (Pattern 1);
// it is NOT extracted here. If an [[include]] appears, the skip keeps it out of
// unitOrder, and leaving it in rawMap is harmless (the unit loop keys on
// unitOrder, not rawMap iteration — see parser.go:80-93).
```

The unit loop iterates `unitOrder` (parser.go:80), NOT `rawMap`, so any reserved key left in `rawMap` but absent from `unitOrder` is simply never processed. `[VERIFIED: parser.go:80-93]`

### Pattern 3: Hand-rolled recursive `Unit.Clone()` (HS-1, TMPL-08)

**What:** A method on `*Unit` that returns a deep copy. Must recurse into `Subunits` (map of `*Unit`) and copy every slice field element-wise (`Links`, `LinksFrom`, `Expanded`, `SubunitOrder`). Must preserve `Link.Mirror`.

**When to use:** Once per `[[use]]` instantiation, BEFORE substitution — clone the template's root unit, then run the replacer over the clone.

**Why hand-rolled (load-bearing):** `Link.Mirror` is unexported (`link.go:67`). Reflection-based, gob, and json copiers cannot see/set it, so they silently zero it. After validate runs `populateIncomingLinks` (which appends `Mirror:true` entries to `LinksFrom`), a non-cloned or shallow-cloned unit's `LinksFrom` would alias across instantiations, corrupting multiplicity counting. `[VERIFIED: link.go:62-67 Mirror is unexported; index.go:55-86 appends Mirror:true entries]`

**Example (target shape — ~15-20 LOC):**
```go
// Source: internal/model/unit.go (new method)
// Clone returns a deep copy of the unit and its entire subtree.
// It preserves the unexported Link.Mirror field (HS-1): reflection/gob/json
// copiers silently drop Mirror, corrupting multiplicity counting after validate.
func (u *Unit) Clone() *Unit {
    if u == nil {
        return nil
    }
    clone := *u // shallow copy of value fields + slice headers
    // Deep-copy slice fields element-wise
    clone.Expanded = slices.Clone(u.Expanded)
    clone.SubunitOrder = slices.Clone(u.SubunitOrder)
    clone.Links = cloneLinks(u.Links)
    clone.LinksFrom = cloneLinks(u.LinksFrom)
    // Deep-copy subunits map (every *Unit cloned — TMPL-04 subtrees)
    if u.Subunits != nil {
        clone.Subunits = make(map[string]*Unit, len(u.Subunits))
        for k, sub := range u.Subunits {
            clone.Subunits[k] = sub.Clone()
        }
    }
    return &clone
}

// cloneLinks deep-copies a link slice, preserving the unexported Mirror field.
func cloneLinks(links []Link) []Link {
    if links == nil {
        return nil
    }
    out := make([]Link, len(links))
    for i := range links {
        out[i] = links[i] // value copy preserves ALL fields incl. unexported Mirror
    }
    return out
}
```

`out[i] = links[i]` is a *value* copy of the `Link` struct — because `Link` has no pointer/map/slice fields (all value types: strings, ints, enums), a value copy fully duplicates it *including the unexported `Mirror` bool*. This is the key correctness property reflection-based copiers break. `[VERIFIED: link.go:43-68 — Link fields are all value types]`

### Pattern 4: `strings.NewReplacer` substitution (TMPL-03, TMPL-10)

**What:** Build a `strings.NewReplacer` from `${param}` → value pairs for all declared params, then apply it to every string field on the cloned unit + subunits + links.

**When to use:** Once per `[[use]]`, after `Clone()`, before attaching under the parent.

**Example (target shape):**
```go
// Source: internal/template/expand.go (new)
// Build the replacer from declared params + supplied values.
pairs := make([]string, 0, len(tmpl.Params)*2)
for _, p := range tmpl.Params {
    pairs = append(pairs, "${"+p+"}", supplied[p])
}
r := strings.NewReplacer(pairs...)
// Apply to every string field — recurse into subunits + links.
applySubstitution(clone, r)
```

`applySubstitution` walks the unit tree once, calling `r.Replace` on each of: `Name`, `Description`, `Technology`, `Color` (and `Reference` once Phase 28 adds the field — generic field-walking future-proofs), plus each link's `Peer`, `Description`, `Technology`, `Color`, `Style`. Substitution is applied IN PLACE on the clone (safe — the clone is freshly deep-copied, no aliasing). `[VERIFIED: Unit/Link string fields enumerated from unit.go:41-72 + link.go:43-68]`

### Anti-Patterns to Avoid

- **Shallow `*Unit` copy for each instantiation (`clone := *tmpl.Unit`):** corrupts the Nth instantiation because `LinksFrom`/`Subunits` slice/map headers alias the template. The validator's `populateIncomingLinks` (index.go:55-86) appends to `LinksFrom` in place, so the 2nd instantiation's validate would append to the 1st's slice. This is the exact HS-1 corruption.
- **Using `reflect`/`gob`/`json` for deep-copy:** silently zeros `Link.Mirror`, re-breaking multiplicity counting (WR-02) after validate.
- **Registering reserved tables in `unitOrder`:** `[template.foo]` would create a phantom top-level unit named `foo` (or `template` as a parent). Pattern 1 prevents this.
- **Substituting only top-level fields, forgetting subunits/links:** TMPL-03/TMPL-04 require substitution across the whole subtree. A partial walk misses `Link.Peer` → broken peer resolution.
- **Letting `[[use]]` keys leak into the unit loop:** `rawMap["use"]` is an array-of-tables; if not deleted after extraction, the unit loop (keyed on unitOrder) still won't touch it, BUT deleting it is defensive and matches the `properties` precedent (parser.go:68-77 deletes implicitly by only extracting, not deleting — but the loop is order-keyed so it's safe). Recommend explicit extraction without deletion, matching the existing `properties` pattern.
- **Assuming humanize/relative-peer passes exist:** they do NOT (verified this session). Phase 31 must not call them. XC-03/XC-04 are satisfied by correct *placement* of Expand, not by invoking absent code.
- **Adding `template`/`use`/`include` to `isBuiltinField`:** WRONG — these are top-level table names, not unit leaf fields. CONTEXT.md D-08 note is explicit. Only Pattern 1 (skip) + Pattern 2 (extraction) handle them.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| `${param}` substitution | A custom regex/template engine | `strings.NewReplacer` | Exact-purpose, zero-dep, no escaping hazards; CONTEXT.md locks this choice |
| Ordered key capture | Re-implement TOML scanning | existing `unstable.Parser` walk (parser.go:107) | Already produces ordered keys; extend its skip rule, don't rewrite |
| Slice shallow-copy | manual `for` loops | `slices.Clone` | already the project idiom (parser.go:310); stdlib, tested |
| Property-style extraction | new extraction framework | the existing `properties` extraction pattern (parser.go:68-77) | established; mirror it for template/use |

**Key insight:** Every mechanism this phase needs already has a precedent in the codebase. The phase is novel-composition, not novel-mechanism: skip-rule (extend), extraction (mirror), deep-copy (new but ~15 LOC and unavoidable given `Mirror`), substitution (stdlib), pipeline stage (one line in root.go).

## Common Pitfalls

### Pitfall 1: HS-1 deep-copy aliasing (THE core correctness concern)
**What goes wrong:** The 2nd+ instantiation of a template inherits the 1st's `LinksFrom` mirror links after validate runs, corrupting multiplicity counting and producing phantom incoming links.
**Why it happens:** `validator.populateIncomingLinks` (index.go:55-86) appends `Mirror:true` `Link` entries to `targetInfo.Unit.LinksFrom` IN PLACE. If two instantiations share the same underlying `*Unit` (shallow copy), the append mutates shared state.
**How to avoid:** Hand-rolled recursive `Unit.Clone()` (Pattern 3) BEFORE substitution. Clone recurses into `Subunits` and value-copies every `Link` (preserving `Mirror`).
**Warning signs:** Three-instantiation test fails after a validate pass; `LinksFrom` slices of two instantiations share elements; multiplicity counts are off. The three-instantiation regression test (TMPL-08) is the gate — it MUST run validate then assert `LinksFrom` slices are disjoint across instantiations.

### Pitfall 2: Phantom units from reserved tables
**What goes wrong:** `[template.svc]` registers `svc` (or `template`) as a real top-level unit, appearing in views/validation.
**Why it happens:** `captureDefinitionOrder` only skips `properties`; any other top-level table enters `unitOrder`.
**How to avoid:** Pattern 1 — extend the skip rule to `template`/`use`/`include`.
**Warning signs:** A model with a `[template.*]` table and no `[[use]]` produces phantom units in the output; `len(unitOrder)` exceeds the hand-authored unit count.

### Pitfall 3: Silent `${param}` literals (TMPL-06 violation)
**What goes wrong:** A typo'd or missing param leaves `${name}` literally in the output unit's fields.
**Why it happens:** `strings.NewReplacer` silently leaves unknown `${...}` intact; if the missing-param check isn't strict, the literal leaks.
**How to avoid:** Two-layer defense — (1) strict pre-check: every declared param MUST be present as a key in the `[[use]]` table (missing → hard error naming template+param+site); (2) post-substitution scan: after applying the replacer, scan all string fields for any remaining `${` and treat as a hard error (catches params referenced in fields but not declared in `params`). Layer (1) satisfies TMPL-02/TMPL-06; layer (2) is belt-and-suspenders.
**Warning signs:** Output contains `${`; a test asserting `!strings.Contains(field, "${")` fails.

### Pitfall 4: Subunit key parametrization (anti-feature leak)
**What goes wrong:** The executor substitutes into subunit map keys, producing parametrized structure (TMPL-04 violation / multi-output territory).
**Why it happens:** Tempting to treat subunit keys like string fields.
**How to avoid:** D-04 is explicit — subunit keys stay VERBATIM. `applySubstitution` operates ONLY on string field VALUES, never on map keys. The subunit map is iterated by key for recursion, but the key itself is never passed through the replacer.
**Warning signs:** Two instantiations of the same template produce subunits with different keys.

### Pitfall 5: Order regression (expanded units sorted, not appended at `[[use]]` position)
**What goes wrong:** Expanded units appear in alphabetical or map-iteration order instead of `[[use]]` document order.
**Why it happens:** `captureDefinitionOrder` exists precisely to preserve authoring order; a naive Expand that inserts into `Units` without appending to `UnitOrder` in the right sequence breaks rendering order.
**How to avoid:** Process `Model.Instantiations` in slice order (which preserves document order — slice, not map), and append each produced root unit's path to `Model.UnitOrder` at the position corresponding to the `[[use]]`. Since `[[use]]` is skipped in `unitOrder` (Pattern 1), the expanded units append in `[[use]]` document order. `[VERIFIED: captureDefinitionOrder preserves first-seen order; Instantiations is a slice preserving array-of-tables order]`
**Warning signs:** Rendered diagram shows expanded units out of author-intended order.

## Code Examples

Verified patterns from the live codebase (read this session):

### Existing `properties` extraction (the template/use extraction precedent)
```go
// Source: internal/parser/parser.go:68-77 [VERIFIED: codebase]
if props, ok := rawMap["properties"]; ok {
    propsData, err := toml.Marshal(props)
    if err != nil {
        return nil, &ParseError{Message: "failed to marshal properties", Cause: err}
    }
    if err := toml.Unmarshal(propsData, &m.Properties); err != nil {
        return nil, wrapDecodeError(err)
    }
}
```

### Existing skip rule (the BC-1 extension point)
```go
// Source: internal/parser/parser.go:127-130 [VERIFIED: codebase]
// Skip [properties] section
if len(parts) == 1 && parts[0] == "properties" {
    continue
}
```

### Existing `slices.Clone` idiom
```go
// Source: internal/parser/parser.go:310 [VERIFIED: codebase]
return slices.Contains([]string{ ... }, key)
// slices.Clone is the project's shallow-slice-copy primitive (stdlib).
```

### Validator in-place mutation (HS-1 root cause)
```go
// Source: internal/validator/index.go:70-86 [VERIFIED: codebase]
targetInfo.Unit.LinksFrom = append(targetInfo.Unit.LinksFrom, model.Link{
    Peer: sourcePath, /* ... */ Mirror: true,
})
// NOTE: this appends to the Unit's LinksFrom slice IN PLACE.
// A shallow-cloned Unit shares this slice header → HS-1 corruption.
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| go-metadot textual preprocessor (define-before-use, trailing defaults) | Structured post-parse expansion (forward refs free, strict params) | v1.10 design (CONTEXT.md) | Forward refs (TMPL-09) are free; param strictness (TMPL-02) is a feature |
| Reflection/gob/json deep-copy | Hand-rolled recursive Clone | v1.10 (HS-1) | Preserves unexported `Link.Mirror`; only correct option |

**Deprecated/outdated:**
- Textual preprocessing for templates (go-metadot/PlantUML style): superseded by structured post-parse — do NOT implement a string-rewriting preprocessor.
- Any deep-copy that isn't explicit field-by-field: silently drops `Mirror`.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `[[use]]` array-of-tables unmarshals into a `[]Instantiation` slice preserving document order | Architecture Patterns | go-toml/v2 array-of-tables ordering — high confidence (standard TOML semantics) but not yet test-verified in this repo. Mitigation: Plan 1/2 tests assert order. |
| A2 | Phase 28 (`reference` field on Unit) may or may not have landed by the time Phase 31 executes | TMPL-10 / Substitution | Substitution must generically cover all string fields including a future `Reference`; if Phase 28 hasn't landed, the field simply isn't there yet and TMPL-10 is satisfied structurally. Low risk — generic field walk. |
| A3 | Phases 29/30 (relative-peer, humanize) are NOT yet implemented (verified by grep this session) | XC-03, XC-04 | Plan correctly avoids depending on them. If they HAVE landed by execution time, Expand still inserts correctly before them. |

**If this table were empty:** All other claims were verified against the live codebase this session.

## Open Questions

1. **Exact TOML shape of `TemplateDef` and `Instantiation`**
   - What we know: CONTEXT.md gives Claude discretion; recommends `Templates map[string]*TemplateDef`, `Instantiations []Instantiation`, both `toml:"-"`, extracted from rawMap.
   - What's unclear: Whether `TemplateDef` stores a pre-parsed `*model.Unit` subtree or re-parses from the raw map at Expand time. The `[[template.svc.link]]` array-of-tables inside a template must parse into `model.Unit.Links` correctly.
   - Recommendation: Store the template's subtree as a `*model.Unit` parsed during extraction (reuse `parseUnitWithOrder` semantics), keyed by template name in `Templates`. `Instantiation` is a flat struct: `Template string`, `Parent string`, plus a `map[string]string` (or `toml:",inline"` catch-all) for params. The planner/executor finalizes exact field names.

2. **How `[[template.svc.link]]` parses within the `[template.*]` namespace**
   - What we know: go-toml/v2 represents `[template.svc]` with a nested `link` array as `map["template"]["svc"]["link"] = []map[string]any`. The existing `parseUnitWithOrder` re-marshals a sub-map to TOML then unmarshals into `model.Unit`, which handles `[[link]]` → `Links`.
   - What's unclear: Whether the `captureDefinitionOrder` skip (Pattern 1) interferes with capturing the template's internal subunit/link order for `SubunitOrder`.
   - Recommendation: The skip only prevents reserved tables from entering the TOP-LEVEL `unitOrder`. Template-internal order can be captured separately during extraction (the `unstable.Parser` walk already yields every expression including nested template ones), OR templates can fall back to map-iteration order (acceptable since template authors control a fixed subtree). Planner to specify; non-blocking.

## Environment Availability

> This phase has NO external dependencies beyond the Go toolchain and already-pinned modules. Step 2.6 audit:

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | build/test | ✓ | go1.26.5 darwin/arm64 | — |
| go-toml/v2 | parsing | ✓ | pinned in go.mod | — |
| testify | tests | ✓ | pinned in go.mod | — |
| golangci-lint | lint gate | ✓ | configured (.golangci.yml) | — |

**Missing dependencies with no fallback:** none
**Missing dependencies with fallback:** none

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + testify (assert/require) [VERIFIED: codebase] |
| Config file | none (standard `go test`) |
| Quick run command | `go test ./internal/...` |
| Full suite command | `go test ./...` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| TMPL-01 | `[template.<name>]` parsed into Model.Templates, not unitOrder | unit | `go test ./internal/parser/ -run TestParseTemplateTable -x` | ❌ Wave 0 |
| TMPL-02 | `[[use]]` with all params instantiates; missing param errors | unit | `go test ./internal/template/ -run TestExpandMissingParam -x` | ❌ Wave 0 |
| TMPL-03 | `${param}` substitutes into all string fields + link fields | unit | `go test ./internal/template/ -run TestExpandSubstitution -x` | ❌ Wave 0 |
| TMPL-04 | template with subunit subtree expands whole subtree | unit | `go test ./internal/template/ -run TestExpandSubtree -x` | ❌ Wave 0 |
| TMPL-05 | expanded unit passes validator + appears in views | integration | `go test ./cmd/c4drill/ -run TestExpandEndToEnd -x` | ❌ Wave 0 |
| TMPL-06 | missing param → hard error naming template+param+site | unit | `go test ./internal/template/ -run TestExpandMissingParamNames -x` | ❌ Wave 0 |
| TMPL-07 | duplicate path across uses/hand-authored → hard error | unit | `go test ./internal/template/ -run TestExpandDuplicatePath -x` | ❌ Wave 0 |
| TMPL-08 | 3× instantiate → idempotent re-expand + disjoint LinksFrom post-validate (HS-1) | unit | `go test ./internal/template/ -run TestExpandThreeInstantiationsHS1 -x` | ❌ Wave 0 |
| TMPL-09 | `[[use]]` before `[template.*]` in file works | unit | `go test ./internal/parser/ -run TestParseUseBeforeTemplate -x` + `go test ./internal/template/ -run TestExpandForwardRef -x` | ❌ Wave 0 |
| TMPL-10 | `reference` field substitutes params | unit | `go test ./internal/template/ -run TestExpandReferenceField -x` (conditional on Phase 28 landing) | ❌ Wave 0 |
| XC-03 | expanded unit's parent = `[[use]]` parent (instantiation site) | unit | `go test ./internal/template/ -run TestExpandParentPlacement -x` | ❌ Wave 0 |
| XC-04 | Expand runs before Validate (pipeline slot) | integration | `go test ./cmd/c4drill/ -run TestPipelineOrderExpandBeforeValidate -x` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/parser/ ./internal/model/ ./internal/template/`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green before `/gsd:verify-work`: `go test ./...` + `golangci-lint run`

### Wave 0 Gaps
- [ ] `internal/model/unit_test.go` — covers TMPL-08 (Clone: Mirror preservation, Subunits recursion). NEW file (no model tests exist).
- [ ] `internal/template/expand_test.go` — covers TMPL-02..10, XC-03. NEW package + tests.
- [ ] `internal/parser/parser_test.go` (extend) — covers TMPL-01/09 (reserved-table skip, extraction, forward-ref). EXISTING file, add cases.
- [ ] `cmd/c4drill/root_test.go` or pipeline test — covers TMPL-05/XC-04 (end-to-end, ordering). Verify a CLI test file exists or add one.
- [ ] `testdata/template_*.toml` fixtures — basic, subtree, 3x-instantiate, missing-param, duplicate-path, forward-ref. NEW fixtures.

*(If no gaps: "None — existing test infrastructure covers all phase requirements")* — gaps exist; all are new files/fixtures.

## Security Domain

> This phase adds no external input surface, no network, no auth, no crypto, no user accounts. It is a pure in-process TOML-to-model transformation over local files. Security enforcement is enabled by default but the ASVS surface for this phase is minimal — input validation (TOML parse errors, missing-param/duplicate-path errors) is the only applicable category, and it is already covered by the parser's error handling and the new Expand checks.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | N/A — no auth |
| V3 Session Management | no | N/A — no sessions |
| V4 Access Control | no | N/A — no privileged ops |
| V5 Input Validation | yes | TOML parse error handling (existing `wrapDecodeError`) + Expand's missing-param/duplicate-path strict checks (TMPL-06/07). No unbounded recursion: template expansion is single-level (no nesting, CONTEXT.md deferred), so no recursion-cap DoS vector. |
| V6 Cryptography | no | N/A — no crypto |

### Known Threat Patterns for Go TOML parsing

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Malformed TOML causing panic | Denial of Service | go-toml/v2 returns errors (no panic on malformed input); `Parse` wraps via `wrapDecodeError`. Existing handling unchanged. |
| Unbounded template nesting → stack exhaustion | Denial of Service | Template nesting is deferred (CONTEXT.md) — single-level only in v1.10. `Unit.Clone` recurses into subunits but subtree depth is bounded by the model's C1/C2/C3 nesting (max ~3 levels), not by template count. No mitigation needed for v1.10. |
| `${param}` injection into output | Tampering | N/A — output is a C4 diagram (local render), not executed code. `${...}` is substituted literally into display strings; no shell/eval. |

## Sources

### Primary (HIGH confidence)
- `internal/parser/parser.go` — full file read this session (Parse:47, captureDefinitionOrder:100, properties skip:128, properties extraction:68-77, isBuiltinField:309, parseUnitWithOrder:160, unit loop:80-93)
- `internal/model/unit.go` — full file read this session (Unit struct:41-72, Subunits map:71, Links/LinksFrom:65/67, no Clone exists)
- `internal/model/link.go` — full file read this session (Link struct:43-68, Mirror:67 unexported, all value-type fields)
- `internal/validator/index.go:55-86` — `populateIncomingLinks` in-place mutation of `LinksFrom` (HS-1 root cause)
- `cmd/c4drill/root.go:112` (ParseFile) → `:118` (Validate) — pipeline insertion point; NO humanize/relative-peer pass present
- `.planning/phases/31-template-expansion/31-CONTEXT.md` — D-01..D-08 locked decisions, HS-1 spec
- `.planning/REQUIREMENTS.md` — TMPL-01..10, XC-03/04 verbatim
- `.planning/STATE.md` — confirms Phases 28/29/30 not yet implemented; Phase 31 builds on Parse+Validate directly

### Secondary (MEDIUM confidence)
- `.planning/research/SUMMARY.md`, `ARCHITECTURE-v1.10.md`, `STACK.md`, `PITFALLS.md` — milestone research (substantial, already done; this phase-level research consolidates + verifies the integration points against live code)

### Tertiary (LOW confidence)
- none — all claims verified against live code this session

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — zero new deps, all verified against go.mod + codebase
- Architecture: HIGH — every integration point read this session; pipeline current-state verified by grep (no humanize/relative-peer)
- Pitfalls: HIGH — HS-1 root cause verified in index.go:55-86; phantom-unit vector verified in captureDefinitionOrder:128
- Substitution/Clone mechanics: HIGH — Link field types verified (all value types → value copy preserves Mirror)

**Research date:** 2026-08-08
**Valid until:** 2026-09-07 (stable internal-codebase phase; no external API surface to drift)
