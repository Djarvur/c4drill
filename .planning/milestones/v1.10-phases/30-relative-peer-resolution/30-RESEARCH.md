# Phase 30: Relative-peer resolution - Research

**Researched:** 2026-08-08
**Domain:** Post-parse model transform (relative-to-absolute peer rewrite) for the C4Drill TOML pipeline
**Confidence:** HIGH (all claims verified against the live codebase in this session; milestone SUMMARY.md §3/§4/§6/§9 already settled the domain)

## Summary

Phase 30 adds a pure post-parse pass, `peer.Resolve(m *parser.Model) error`, that rewrites every `Link.Peer` (on every `Link` in every `Unit.Links` and `Unit.LinksFrom`, including subunits and — once Phase 31 lands — templated units) from a relative bare name to an absolute dotted path. The pass inserts into the existing gap between `parser.ParseFile` (`cmd/c4drill/root.go:112`) and `validator.Validate` (`cmd/c4drill/root.go:118`). It uses stdlib only (`strings.Contains` for the `.` gate; the model's own parent/child structure for the walk-up). It is a **no-op rewrite for every peer that already contains a `.`` (D-16 step 1) and an identity rewrite (`externalSys` → `externalSys`) for bare peers that match a top-level unit (D-15 root step) — so the backward-compat hard contract holds by construction.

The algorithm (D-13/D-14/D-15/D-16) is a nearest-first walk-up over the link-host unit's ancestry: for a bare peer `X` declared by a unit at path `a.b.c`, the resolver tries `a.b.X` (c's siblings), then `a.X` (aunt/cousin at grandparent), then top-level `X`, taking the **first depth with exactly one match**. The walk reads the model's tree structure directly: top-level keys in `Model.Units` and children in `Unit.Subunits` (`internal/model/unit.go:73`). Two hard errors arise: (1) miss-at-root (bare peer resolves nowhere) → "cannot resolve peer `X` from unit `Y`"; (2) same-depth multi-match → see the dead-code finding below.

**Primary recommendation:** Implement `internal/peer/Resolve.go` as a single function `Resolve(m *parser.Model) error` plus unexported tree-walk helpers (`findFirstParentMatch`, `resolveUnit`), with the corpus backward-compat test as the acceptance gate (ERGO-02). The dead-code question about ROADMAP criterion 3 is resolved below: same-depth ambiguity is **structurally unreachable** given `Subunits map[string]*Unit` (unique keys per parent), so the planner should still author the error branch (defensive + documents intent) but it will not be exercised by any reachable model.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions (NON-NEGOTIABLE — do not re-litigate)

- **D-13 (immediate parent):** For a unit at path `a.b.c` declaring a link, the "enclosing parent" whose children the resolver searches first is **`a.b`** — `c`'s immediate parent. Relative resolution tries `a.b.x` (c's sibling). NOT `a`'s children, NOT `c`'s own children.
- **D-14 (walk-up nearest-first):** Bare peer walks UP: immediate parent's children → grandparent's children → ... → top-level (root). **First depth with exactly-one-match wins** (closest ancestor's matching child). Multiple matches at the SAME depth = hard error; first depth with exactly one match resolves. Mirrors lexical scoping.
- **D-15 (walk reaches root):** The walk-up reaches root — top-level units are the outermost scope. So `peer = "messageBus"` from any depth resolves to a top-level `messageBus` if one exists, even with no sibling/aunt match. Miss at root → absolute-fallback (D-16) → hard error.
- **D-16 (unified gate, no top-level short-circuit):** For a peer value:
  1. **If it contains `.`** → treat as absolute. Resolve as-is; miss = existing peer-existence error (`internal/validator/rules.go:14`). No relative resolution attempted.
  2. **If bare (no `.`)** → run the walk-up (D-13/D-14/D-15). First depth with exactly-one-match wins; multiple matches at same depth = hard error; miss at root = hard error ("cannot resolve peer `X` from unit `Y`").
  - NO separate "exact top-level match" check — top-level is the final step of the walk-up. A bare peer matching a top-level unit resolves via the walk-up's root step (identity rewrite).
- **HS-2 (carried from Phase 31 D-01/D-02):** Relative peers authored inside a template resolve at the **instantiation site's** parent, NOT the template's lexical location. The pass runs AFTER template expansion and treats all units uniformly — NO template-special-case logic. (Phase 30 ships before Phase 31's expand call exists, so this is a forward-compatibility contract: the resolver must not assume templates are unexpanded.)
- **Implementation site:** Separate post-parse pass rewriting `Link.Peer` in place on the assembled model before `BuildIndex`/validation. Validator's existing absolute-path logic is untouched. Stdlib only.
- **ERGO-01:** Bare peer (no `.`) resolves against enclosing parent's children (walk-up per D-14).
- **ERGO-02:** Absolute fallback when peer contains `.` OR does not resolve as a relative sibling (D-16 unified gate). Every existing model parses identically — backward-compat is a hard contract.

### Claude's Discretion

- Internal package/location for the resolver (`internal/peer/Resolve` recommended; alternative is a function in `internal/parser/`).
- Likely signature: `func Resolve(m *parser.Model) error` (in-place rewrite of `Link.Peer`; error for unresolvable/ambiguous).
- Error message wording for the two hard-error cases — lean toward naming the peer, the link-host unit, and (for ambiguity) the competing matches.
- Whether to surface a warning (not error) on cross-depth shadowing — optional, planner's call.

### Deferred Ideas (OUT OF SCOPE — do NOT implement)

- **Cross-depth shadowing warning** (nearest-first silently picks nearer) — deferred unless cheap.
- **Compact one-liner link shorthand (ERGO-06)** — Phase 29 (at-risk).
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| ERGO-01 | A `peer` value resolves *relative to the enclosing parent block* when it is a bare name (no `.`) that matches a sibling unit | D-13/D-14/D-15 walk-up algorithm below; tree-walk over `Model.Units` + `Unit.Subunits`; nearest-first resolution |
| ERGO-02 | Relative resolution falls back to absolute when the peer contains a `.` OR exactly matches a top-level unit path OR does not resolve as a relative sibling — every existing model parses identically | D-16 unified gate; corpus backward-compat test asserting byte-identical `(source, resolved-peer)` set is the acceptance criterion (research §"Backward-compat corpus analysis") |
</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Peer string rewrite (bare → absolute) | Pre-validation transform pass | — | Pure function on `*parser.Model`; runs before `validator.BuildIndex` so the validator sees only absolute paths |
| Tree-walk for ancestor lookup | Pre-validation transform pass | — | Reads `Model.Units` (top-level) + `Unit.Subunits` (children); no index needed (the validator's flat index is built later and consumes the rewritten peers) |
| Ambiguity / unresolvable error reporting | Pre-validation transform pass | CLI error path (`cmd/c4drill/root.go` `fmt.Errorf`) | Returns `error` from `Resolve`; CLI wraps with `fmt.Errorf("resolve peers: %w", err)` matching the existing Parse/Validate staging idiom |
| Backward-compat gate (no-op for existing models) | Test suite | — | Corpus test (ERGO-02 acceptance) asserts byte-identical `(source, resolved-peer)` set pre/post |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `strings` (stdlib) | Go 1.26.1 | `strings.Contains(peer, ".")` — the bare-vs-absolute gate (D-16 step 1); `strings.Split`/`Join` for path manipulation | Already used project-wide; no new dependency (research §2 "Zero new external dependencies") |
| `slices` (stdlib) | Go 1.26.1 | `slices.Clone` idiom (already at `parser.go:310`) if the resolver needs to materialize a parent-path stack | Project idiom |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/stretchr/testify` | current | `assert`/`require` in unit tests | All new test files (project idiom — see `internal/parser/parser_test.go`) |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Hand-rolled tree-walk | `path.Dir`/`path.Join` from stdlib | Rejected — `path` is for slash-separated filesystem paths, not dotted TOML keys; manual `strings.Split(path, ".")` + progressive drop is clearer and correct |
| Building a flat index then lookup | `validator.BuildIndex` reuse | Rejected — the validator's index is built AFTER resolve (resolve produces the model the index is built from); also the walk-up semantics (nearest-first, ancestor levels) are not expressible as a flat lookup without re-deriving parent paths anyway |
| Putting `Resolve` in `internal/parser/` | New `internal/peer/` package | Phase 31 chose `internal/template/` as a sibling package (analog decision); `internal/peer/` mirrors that boundary and keeps `internal/parser/` focused on TOML→Model decoding |

**Installation:** none — no new dependency.

**Version verification:** Go 1.26.1 confirmed in `go.mod:3`; testify already in `go.sum`.

## Package Legitimacy Audit

> Not applicable — Phase 30 installs zero external packages (stdlib + existing testify only).

## Architecture Patterns

### System Architecture Diagram

```
                         ┌─────────────────────────────────────────┐
   TOML file ──────────► │ parser.ParseFile (root.go:112)          │
                         │   → *parser.Model                       │
                         └────────────────┬────────────────────────┘
                                          │ (Phase 31 inserts template.Expand HERE)
                                          ▼
                         ┌─────────────────────────────────────────┐
                         │ peer.Resolve (NEW — this phase)         │  ◄── walks Model.Units +
                         │   rewrites Link.Peer: bare → absolute   │      Unit.Subunits tree;
                         │   in place; returns error on miss/ambig │      rewrites Links AND LinksFrom
                         └────────────────┬────────────────────────┘
                                          │ (Phase 29 inserts humanize HERE)
                                          ▼
                         ┌─────────────────────────────────────────┐
                         │ validator.Validate (root.go:118)        │
                         │   BuildIndex(m.Units, "") → flat index  │  ◄── sees only absolute peers now
                         │   ValidateReferences (rules.go:14)      │      (unchanged)
                         └─────────────────────────────────────────┘
```

Data flow: the model enters `peer.Resolve` with a mix of absolute (dotted) and bare (no-dot) peers and leaves with all peers absolute. The validator downstream is structurally unchanged.

### Recommended Project Structure

```
internal/peer/
├── resolve.go          # Resolve(m *parser.Model) error + unexported walk helpers
└── resolve_test.go     # unit + integration tests (TDD per workflow.tdd_mode)
internal/parser/
└── parser_test.go      # (optionally extend — see PATTERNS.md)
cmd/c4drill/
├── root.go             # MODIFY: insert peer.Resolve call in the Parse→Validate gap
└── testdata/
    ├── *.toml          # existing corpus (untouched — backward-compat proof)
    ├── peer_walkup.toml       # NEW: sibling + aunt + top-level walk-up cases
    ├── peer_ambiguous.toml    # NEW: same-depth ambiguity error case (defensive)
    └── peer_unresolvable.toml # NEW: miss-at-root error case
```

### Pattern 1: Tree-walk ancestor enumeration (the walk-up core)

**What:** Given a unit's full dotted path `a.b.c`, enumerate its ancestor scopes nearest-first: `[a.b, a, <root>]`. For each scope, look up the bare peer name in that scope's children map; first scope with exactly one match wins.

**When to use:** Every bare peer (D-16 step 2).

**Pseudocode (verified against `validator.BuildIndex` path mechanics at `internal/validator/index.go:24-43`):**
```go
// Source: derived from D-13/D-14/D-15 + verified against validator/index.go:24-43
// (BuildIndex builds paths as parentPath + "." + name; the walk-up reverses this)

func resolvePeer(hostPath, peer string, m *parser.Model) (string, error) {
    if strings.Contains(peer, ".") {
        return peer, nil          // D-16 step 1: absolute, untouched
    }
    // D-16 step 2: walk-up nearest-first
    scopes := ancestorScopes(hostPath, m) // [a.b, a, root] for hostPath "a.b.c"
    for _, scope := range scopes {
        children := childrenOf(scope, m)  // map[string]*Unit at this scope
        if _, ok := children[peer]; ok {
            // exactly one match (map keys are unique per parent) — resolve
            return joinPath(scope, peer), nil
        }
        // no match at this scope → walk up to next
    }
    return "", &ResolveError{Peer: peer, Host: hostPath}
}
```

**Critical insight — `ancestorScopes` must include the implicit root scope.** The root scope's "children" are `m.Units` itself (top-level keys). For a top-level host (e.g. `actorA`), the walk-up is: `[<root>]` only (no parent). For a host at `a.b.c`, it is `[a.b, a, <root>]`.

### Pattern 2: Iterating all link-hosts in a model

**What:** `Resolve` must visit every `Unit.Links[i].Peer` and `Unit.LinksFrom[i].Peer` across the whole tree (top-level units AND all subunits).

**When to use:** The resolver entry point.

**Pattern (mirrors `validator.BuildIndex` recursion at `internal/validator/index.go:24-43`):**
```go
// Source: derived from validator/index.go:24-43 (BuildIndex recursion shape)
func Resolve(m *parser.Model) error {
    if m == nil { return nil }
    return resolveUnits(m.Units, "", m) // parentPath="" for top-level
}

func resolveUnits(units map[string]*model.Unit, parentPath string, m *parser.Model) error {
    for name, unit := range units {
        fullPath := name
        if parentPath != "" { fullPath = parentPath + "." + name }
        if err := resolveUnitLinks(unit, fullPath, m); err != nil { return err }
        if len(unit.Subunits) > 0 {
            if err := resolveUnits(unit.Subunits, fullPath, m); err != nil { return err }
        }
    }
    return nil
}
```

### Pattern 3: In-place rewrite of `Link.Peer`

**What:** The resolver mutates `Link.Peer` strings in place on the existing `*model.Unit` graph. It does NOT clone units, does NOT touch `UnitOrder`/`SubunitOrder`, does NOT change structure.

**When to use:** Every visited link.

**Why in-place is safe here (unlike templates' HS-1):** Phase 30 runs once per CLI invocation on a freshly-parsed model that no other code has aliased yet (parse produces it, resolve consumes it, validate consumes the result). There is no second instantiation to corrupt. The validator's `populateIncomingLinks` (`internal/validator/index.go:55-86`) appends to `LinksFrom` AFTER resolve — so resolve's `LinksFrom` iteration sees only authored entries (mirrors won't exist yet). [VERIFIED: internal/validator/index.go:55-86 + validator.go:22-24]

### Anti-Patterns to Avoid

- **Building a flat index then doing lookups** — the walk-up semantics (per-scope, nearest-first) are lost. The resolver must walk the tree scope-by-scope.
- **Touching `UnitOrder` or `SubunitOrder`** — load-bearing for rendering order (STATE.md). Resolve only rewrites `Link.Peer` strings.
- **Special-casing templates** — HS-2 says NO template-special-case. The resolver treats every unit identically; template-authored links resolve against the instantiation-site parent because expansion already placed the unit there before resolve runs.
- **Enabling `DisallowUnknownFields()`** — BC-3 (research §9): go-toml/v2 must stay non-strict. (Phase 30 doesn't touch TOML decoding, but the guard applies to any future change.)
- **Reordering the pipeline** — peer.Resolve MUST run after template.Expand (Phase 31) and before humanize (Phase 29) and before Validate. Inserting it elsewhere breaks XC-01.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Bare-vs-absolute gate | A regex or parser | `strings.Contains(peer, ".")` | Trivial, exact, zero cost |
| Ancestor enumeration | A custom path library | `strings.Split(hostPath, ".")` then progressively drop the last segment | Dotted TOML paths are simple; no library fits |
| Recursive tree walk | An iterative queue | Direct recursion mirroring `BuildIndex` | Already the project idiom (`validator/index.go:24-43`); recursion depth bounded by model nesting (typically < 6) |

**Key insight:** The model's parent/child relationships are already implicit in (a) the dotted-path keys and (b) `Unit.Subunits`. No new data structure is needed.

## Common Pitfalls

### Pitfall 1: Forgetting `LinksFrom` peers

**What goes wrong:** The resolver only rewrites `Unit.Links[i].Peer` and misses `Unit.LinksFrom[i].Peer`. Today `LinksFrom` is authored in TOML (`[[unit.linkFrom]] peer = "X"`) AND synthesized by the validator (`populateIncomingLinks`). If a bare peer appears in an authored `linkFrom`, the validator's `ValidateReferences` (rules.go:31-39) will fail because the synthesized flat index won't contain the bare name.
**Why it happens:** `LinksFrom` is easy to overlook because it's primarily validator-synthesized.
**How to avoid:** CONTEXT.md "Rewrite target" explicitly lists both. The resolver visits BOTH `unit.Links` and `unit.LinksFrom` for every unit. The corpus has authored `linkFrom` peers (e.g. `cmd/c4drill/testdata/valid.toml:25`, `expanded.toml:31,49`) — these are all absolute today, but the resolver must handle the bare case too.
**Warning signs:** A model with `[[x.linkFrom]] peer = "y"` (bare) where `y` is a top-level unit fails validation post-resolve with "undefined unit y referenced in linkFrom".

### Pitfall 2: Assuming the corpus has no bare peers (backward-compat test design)

**What goes wrong:** The planner writes the corpus backward-compat test assuming every existing peer is dotted, then discovers bare peers in BOTH corpora: `links.toml` has `webapp`/`user` (root testdata), and `valid.toml`/`expanded.toml`/`multilevel.toml` have `user`/`external`/`externalSys` (cmd/c4drill/testdata). Plus the `invalid*.toml` fixtures have intentionally-undefined bare peers.
**Why it happens:** The valid-corpus bare peers all happen to resolve to top-level units today (the validator's flat index contains the top-level key), so they "just work" pre-resolve. The invalid-corpus bare peers are SUPPOSED to fail.
**How to avoid:** These valid-corpus bare peers are NOT a problem — the walk-up's root step (D-15) resolves each to its identical top-level name, producing an **identity rewrite** (`externalSys` → `externalSys`). The `(source, resolved-peer)` set is still byte-identical. The test must (a) assert the SET comparison, NOT "no peer was rewritten" (which is false — the bare peers ARE rewritten, just to the same string); AND (b) EXCLUDE the `invalid*.toml` fixtures (error-path, expected to fail validation). See "Backward-compat corpus analysis" above for the precise inventory. [VERIFIED: both corpora grepped in this session]
**Warning signs:** Test asserts `len(rewritten) == 0` and fails; OR test includes `invalid_references.toml` and fails because the model doesn't parse/validate.

### Pitfall 3: Same-depth ambiguity is structurally unreachable (criterion 3 dead code)

**What goes wrong:** ROADMAP criterion 3 mandates "Ambiguity at a single nesting depth (two siblings matching the same bare name) is a hard error". The planner designs an elaborate test for it, then finds it cannot be constructed.
**Why it happens:** `Unit.Subunits` is `map[string]*Unit` (`internal/model/unit.go:73`) — map keys are unique within one parent. Two siblings CANNOT share a name. The only way to get "two matches for bare name X at the same ancestor scope" would be if X were a child key AND also... but a single map either has key X once or not at all. So `scope[X]` is always 0-or-1, never 2.
**How to avoid:** Same-depth multi-match is **structurally unreachable** under the walk-up model. The error branch should STILL be authored (defensive coding; documents the intent; guards against a future data-structure change) but it is dead code — the planner must not block on constructing a test that exercises it, and should note this in the plan. Cross-depth "ambiguity" (X exists at both `a.b.X` and `a.X`) is NOT an error — nearest-first (D-14) silently picks `a.b.X`.
**Warning signs:** A test named `TestResolveSameDepthAmbiguity` that cannot be made to pass without violating map uniqueness.

### Pitfall 4: Forgetting the host unit's own subunits are NOT a search scope

**What goes wrong:** The resolver searches the link-host unit's own `Subunits` for the bare peer. Per D-13, the search starts at the IMMEDIATE PARENT's children (the host's siblings), never the host's own children.
**Why it happens:** "Search the enclosing scope" is ambiguous — lexical scoping in most languages searches the local scope first.
**How to avoid:** D-13 explicitly excludes the host's own subunits (a unit doesn't link to its own subunits per VALD-02 anyway). `ancestorScopes(hostPath)` for host `a.b.c` returns `[a.b, a, root]` — never `a.b.c`.
**Warning signs:** A peer `peer = "child"` on unit `a.b` resolves to `a.b.child` instead of erroring (it should error unless `a.child` or top-level `child` exists).

## Code Examples

Verified patterns from the live codebase:

### `BuildIndex` recursion — the canonical tree-walk shape

```go
// Source: internal/validator/index.go:24-43 (VERIFIED — read in this session)
func BuildIndex(units map[string]*model.Unit, parentPath string) map[string]*UnitInfo {
    index := make(map[string]*UnitInfo)
    for name, unit := range units {
        fullPath := name
        if parentPath != "" {
            fullPath = parentPath + "." + name
        }
        index[fullPath] = &UnitInfo{Unit: unit, FullPath: fullPath, Parent: parentPath}
        if len(unit.Subunits) > 0 {
            subIndex := BuildIndex(unit.Subunits, fullPath)
            maps.Copy(index, subIndex)
        }
    }
    return index
}
```

`peer.Resolve`'s recursion mirrors this EXACTLY (Pattern 2 above) — same `fullPath` construction, same `Subunits` recursion — except it rewrites `Link.Peer` instead of building an index.

### CLI pipeline staging idiom — where Resolve inserts

```go
// Source: cmd/c4drill/root.go:111-123 (VERIFIED — read in this session)
// Stage 1: Parse
m, err := parser.ParseFile(inputPath)
if err != nil {
    return fmt.Errorf("parse: %w", err)
}

// Stage 2: Validate
valErrors := validator.Validate(m)
```

Target insertion (between these two stages; Phase 31's `template.Expand` will land in the same gap, ordered before `peer.Resolve`):

```go
// Stage 1.5: Expand templates (Phase 31 — not yet present; will land above this line)
// Stage 1.6: Resolve relative peers (Phase 30)
if err := peer.Resolve(m); err != nil {
    return fmt.Errorf("resolve peers: %w", err)
}
```

### Backward-compat test shape (ERGO-02 acceptance)

```go
// Source: derived from the existing test idiom (internal/parser/parser_test.go:18) + DI-1
// Plan 01 — unit test in internal/peer/resolve_test.go, uses the ROOT testdata corpus
func TestResolveCorpusByteIdentical(t *testing.T) {
    t.Parallel()
    fixtures := []string{"valid.toml", "links.toml", "nested.toml"} // root testdata/, NOT invalid_*.toml
    for _, fix := range fixtures {
        data, err := os.ReadFile("../../testdata/" + fix) // parser-test convention (parser_test.go:18)
        require.NoError(t, err)
        m, err := parser.Parse(data)
        require.NoError(t, err)
        beforePeers := collectPeerSet(m) // map[sourceUnit][]peer across Links AND LinksFrom
        require.NoError(t, peer.Resolve(m))
        afterPeers := collectPeerSet(m)
        assert.Equal(t, beforePeers, afterPeers,
            "corpus %s: peer set changed after Resolve (backward-compat violation)", fix)
    }
}
```

Note `collectPeerSet` captures `(unitPath, link.Peer)` pairs across Links AND LinksFrom for every unit in the tree. The bare peers in the valid corpora (`webapp`, `user` in root testdata; `user`, `external`, `externalSys` in cmd/c4drill/testdata) all resolve to identical top-level strings so the set is unchanged. Plan 02 covers the cmd/c4drill/testdata corpus end-to-end via `TestCLICorpusRendersUnchanged`.

## Backward-compat corpus analysis (load-bearing for ERGO-02)

There are **two testdata corpora** in the repo (verified in this session):

1. **Root `testdata/`** — consumed by `internal/parser/parser_test.go` via `os.ReadFile("../../testdata/<fix>")` (the established convention; `../../testdata/` from `internal/parser/` resolves to the repo-root `/testdata/`). Contains: `valid.toml`, `links.toml`, `nested.toml` (valid fixtures) + `invalid_links.toml`, `invalid_references.toml`, `invalid_subunits.toml` (validator error-path fixtures).

2. **`cmd/c4drill/testdata/`** — consumed by `cmd/c4drill/` integration/render tests. Contains: `valid.toml`, `expanded.toml`, `multilevel.toml` (valid fixtures, with `.dot`/`.svg` goldens) + `invalid.toml` (validator error-path fixture). NOTE: the `valid.toml` files in the two corpora DIFFER (different content — verified by `diff`).

Bare-peer inventory (precise: peer values with no `.` in the value):

**Root `testdata/` corpus:**
| Fixture | Bare peers | Each resolves to |
|---------|------------|------------------|
| `valid.toml` | (none — only `[user]` and `[webapp]` units, no links) | n/a |
| `links.toml` | `peer = "webapp"` (in `[[user.linkFrom]]`), `peer = "user"` (in `[[webapp.link]]`), `peer = "webapp"` (in `[[api.linkFrom]]`) | top-level `[webapp]`, `[user]`, `[webapp]` — identity rewrites |
| `nested.toml` | (none — `[externals]`, `[mainapp]` + subunits, no bare peers) | n/a |
| `invalid_*.toml` | `undefined_db`, `missing_system`, `other`, `parent` (in error-path fixtures) | NOT backward-compat targets — these fixtures are EXPECTED to fail validation; exclude from the corpus test |

**`cmd/c4drill/testdata/` corpus:**
| Fixture | Bare peers | Each resolves to |
|---------|------------|------------------|
| `valid.toml` | 1: `peer = "user"` (in `[[app.api.linkFrom]]`) | top-level `[user]` — identity rewrite |
| `expanded.toml` | 1: `peer = "external"` (in `[[mainsystem.api.link]]`) | top-level `[external]` — identity rewrite |
| `multilevel.toml` | 1: `peer = "externalSys"` (in `[mainSystem.storages.externalStorage.client]`) | top-level `[externalSys]` — identity rewrite |
| `invalid.toml` | (error-path fixture — NOT a backward-compat target) | exclude |

**Conclusion:** Every bare peer in BOTH valid corpora matches a top-level unit. Under D-15 (walk reaches root), each resolves via the root step to its identical top-level name → identity rewrite → the `(source, resolved-peer)` set is byte-identical pre/post. ERGO-02's acceptance criterion is satisfiable for both corpora.

**Test design implication (load-bearing for the planner):**
- Plan 01's `TestResolveCorpusByteIdentical` (unit test in `internal/peer/`) uses the **root `testdata/`** corpus (`valid.toml`, `links.toml`, `nested.toml`) via `../../testdata/` — matching the parser-test convention. This proves the resolver is a no-op on parser-corpus peers.
- Plan 02's `TestCLICorpusRendersUnchanged` (integration test in `cmd/c4drill/`) uses the **`cmd/c4drill/testdata/`** corpus (`valid.toml`, `expanded.toml`, `multilevel.toml`) — proving the CLI renders identically end-to-end.
- Both tests EXCLUDE the `invalid*.toml` fixtures (error-path, not backward-compat targets). [VERIFIED by grep + diff in this session]

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Absolute-only peers (v1.9 and earlier) | Relative peers with walk-up + absolute-fallback (v1.10 Phase 30) | v1.10 | Authors can write `peer = "cache"` from `frontend.api.handlers.auth` and resolve to `frontend.cache` without spelling the full path |

**Deprecated/outdated:** none — the absolute form remains fully supported and is the fall-through for any peer containing `.`.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | `internal/peer/` is the right package location (vs `internal/parser/`) | Standard Stack / Alternatives | Low — both work; `internal/peer/` mirrors Phase 31's `internal/template/` boundary. Planner may override. |
| A2 | The resolver should return a single `error` (not `ValidationErrors`) | Standard Stack | Low — the CLI error path (`fmt.Errorf("...: %w", err)`) handles both; a single error is simpler and resolve fails fast on first unresolvable peer (matching parser's fail-fast). Planner may choose `ValidationErrors` for uniform reporting. |

**If this table is otherwise empty:** all other claims were verified against the live codebase or the milestone research SUMMARY in this session.

## Open Questions

1. **Should the cross-depth shadowing warning be shipped?**
   - What we know: D-14 silently picks the nearer match when `frontend.cache` is shadowed by a nearer `frontend.api.cache`. CONTEXT.md marks this as deferred (Claude's Discretion).
   - What's unclear: whether a non-blocking warning adds enough debugging value to justify the noise.
   - Recommendation: DEFER. Ship the silent nearest-first rule. The corpus test and the resolve-error messages already give authors enough signal. A warning can be added later if users report confusion. Do NOT block the phase on this.

2. **Should `Resolve` fail-fast (first error) or collect all errors?**
   - What we know: `parser.Parse` fails fast (returns first error); `validator.Validate` collects all (`ValidationErrors`).
   - Recommendation: fail-fast (return first `*ResolveError`). Rationale: a model with one unresolvable peer usually has a typo; reporting one clear error beats a wall of errors. Planner may override to collect if uniform reporting is preferred. Either way, the error must name the peer and the host unit.

## Environment Availability

> Step 2.6: SKIPPED — Phase 30 is purely Go code + testdata fixtures, no external dependencies (Go 1.26.1 and testify already present; verified in `go.mod` and `go.sum`).

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + testify (assert/require) |
| Config file | none (standard `go test`) |
| Quick run command | `go test ./internal/peer/` |
| Full suite command | `go test ./...` |
| Estimated runtime | ~3 seconds |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| ERGO-01 | sibling match: `[linuxSystem.localIDP.sessionAPI] peer="sessionManager"` → `linuxSystem.localIDP.sessionManager` | unit | `go test ./internal/peer/ -run TestResolveSibling -x` | ❌ Wave 0 |
| ERGO-01 | aunt match (walk-up): `[frontend.api.handlers.auth] peer="cache"` → `frontend.cache` | unit | `go test ./internal/peer/ -run TestResolveWalkUpAunt -x` | ❌ Wave 0 |
| ERGO-01 | root match: `[linuxSystem.sshAuth.sshd] peer="messageBus"` → top-level `messageBus` | unit | `go test ./internal/peer/ -run TestResolveRoot -x` | ❌ Wave 0 |
| ERGO-01 | nearest-first precedence: `[a.b.c] peer="x"` with both `a.b.x` and `a.x` → `a.b.x` (no error) | unit | `go test ./internal/peer/ -run TestResolveNearestFirst -x` | ❌ Wave 0 |
| ERGO-02 | dotted peer untouched: `peer="a.b.c"` → unchanged | unit | `go test ./internal/peer/ -run TestResolveDottedUntouched -x` | ❌ Wave 0 |
| ERGO-02 | corpus byte-identical: valid/expanded/multilevel peer sets unchanged post-Resolve | unit | `go test ./internal/peer/ -run TestResolveCorpusByteIdentical -x` | ❌ Wave 0 |
| ERGO-02 | unresolvable bare → hard error naming peer + host | unit | `go test ./internal/peer/ -run TestResolveUnresolvableError -x` | ❌ Wave 0 |
| ERGO-01/02 | `LinksFrom` peers also rewritten (authored `[[x.linkFrom]] peer="y"`) | unit | `go test ./internal/peer/ -run TestResolveLinksFrom -x` | ❌ Wave 0 |
| ERGO-01/02 | pipeline: peer.Resolve runs between Parse and Validate; validator sees only absolute peers | integration | `go test ./cmd/c4drill/ -run TestPipelineResolveBeforeValidate -x` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/peer/`
- **Per wave merge:** `go test ./...`
- **Phase gate:** Full suite green + `go vet ./...` clean before `/gsd:verify-work`
- **Max feedback latency:** 5 seconds

### Wave 0 Gaps
- [ ] `internal/peer/resolve_test.go` — all ERGO-01/02 behaviors above. NEW package + tests (TDD: RED first).
- [ ] `cmd/c4drill/testdata/peer_walkup.toml` — sibling + aunt + root + nearest-first cases. NEW fixture.
- [ ] `cmd/c4drill/testdata/peer_unresolvable.toml` — bare peer matching nothing. NEW fixture.
- [ ] `cmd/c4drill/root_test.go` or extend existing — pipeline ordering integration test. NEW or extend.

*(Framework install: none — Go testing + testify already present.)*

## Security Domain

> Phase 30 adds a post-parse string-rewrite pass. It processes untrusted TOML author input. ASVS L1 applies.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | n/a — no auth in this phase |
| V3 Session Management | no | n/a |
| V4 Access Control | no | n/a |
| V5 Input Validation | yes | The resolver validates the peer-name structure implicitly (bare = walk-up, dotted = absolute). It does NOT accept arbitrary input — it only rewrites strings already parsed by `parser.Parse`, which enforces TOML structure. No injection vector: the rewritten peer is a dotted path constructed from existing map keys + the peer string, all of which are TOML-decoded values (no eval, no shell, no SQL). |
| V6 Cryptography | no | n/a |
| V7 Error Handling | yes | Errors must not leak sensitive info; this phase's errors name the peer/host (author-facing diagnostic data, not secrets). Fail-closed: unresolvable peer = hard error (model does not render). |
| V8 Data Protection | no | n/a — no persistent data |

### Known Threat Patterns for {Go CLI TOML transform}

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Path traversal via crafted peer name | Tampering | The resolver constructs dotted paths from map keys (validated by TOML parsing) + the peer string. The output is consumed only by `BuildIndex` (map lookup) and `ValidateReferences`. No filesystem, shell, or network access. A peer like `peer = "../../../etc/passwd"` is just a string that won't match any unit → hard error. No traversal possible. |
| Resource exhaustion (deep walk-up) | Denial of Service | The walk-up depth is bounded by model nesting depth (typically < 6 levels; the parser has no depth cap today, but a pathologically deep model would be caught by Go stack limits long before resolve). Acceptable for a local CLI tool processing trusted author input. |
| Ambiguous/malicious model structure | Tampering | The model structure is produced by `parser.Parse`, which enforces TOML well-formedness. Resolve reads structure, does not reconfigure it. |

**Threat disposition summary:** No high-severity threats. The pass is a pure string transform on already-parsed data with no I/O. All threats are **accept** (low-value local CLI target, no untrusted-network input path).

## Sources

### Primary (HIGH confidence)
- `internal/validator/index.go` (read in full this session) — `BuildIndex` recursion shape, `populateIncomingLinks` mechanics
- `internal/validator/rules.go:1-40` (read this session) — `ValidateReferences` peer-existence check
- `internal/validator/validator.go:1-60` (read this session) — `Validate` staging
- `internal/model/link.go` (read in full this session) — `Link` struct, `Peer` field
- `internal/model/unit.go:40-74` (read this session) — `Unit` struct, `Subunits map[string]*Unit`
- `internal/parser/parser.go:25-96` (read this session) — `Model` struct, `Parse` flow
- `cmd/c4drill/root.go:95-150` (read this session) — pipeline staging, Parse→Validate gap
- `cmd/c4drill/testdata/{valid,expanded,multilevel}.toml` (grepped this session) — corpus peer inventory
- `.planning/research/SUMMARY.md` §3/§4/§6/§9 (milestone research) — pipeline ordering, walk-up semantics, backward-compat gate

### Secondary (MEDIUM confidence)
- `.planning/phases/31-template-expansion/31-PATTERNS.md` — sibling-phase pattern for inserting a transform pass in the same pipeline gap
- `.planning/STATE.md` — D-12 (validator is single gatekeeper), DI-1 (canonicalDOT for goldens), BC-3 (non-strict toml)

### Tertiary (LOW confidence)
- none

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — stdlib only, verified in `go.mod`
- Architecture: HIGH — all integration points read in the live codebase this session
- Pitfalls: HIGH — corpus grepped, dead-code question resolved by reading `unit.go:73`
- Backward-compat: HIGH — all 3 bare corpus peers confirmed to resolve to top-level units (identity rewrite)

**Research date:** 2026-08-08
**Valid until:** 2026-09-07 (stable — pure stdlib, no external API surface)
