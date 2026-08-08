# Phase 29: Optional name humanization - Pattern Map

**Phase:** 29 - Optional name humanization
**Padded phase:** 29
**Phase directory:** .planning/phases/29-optional-name-humanization

Files to create/modify (extracted from CONTEXT.md `<domain>`/`<decisions>` and RESEARCH.md `## Architecture Patterns`):

| File | Role | Create/Modify | Closest analog |
|------|------|---------------|----------------|
| `internal/model/humanize.go` | Pure string utility (`Humanize(segment) string`) | **Create** | `internal/model/colors.go` (small utility file in package `model`), `internal/model/link.go:72 FindLinkByPeer` (a `func` helper in package `model`) |
| `internal/model/humanize_test.go` | Table-driven unit test for the ERGO-04 reference table | **Create** | `internal/parser/parser_test.go` (table-driven `t.Run` + testify `require`/`assert`) — first test in `internal/model/` |
| `internal/parser/parser.go` | Add parse-time `Unit.Name` fallback in `parseUnitWithOrder` | **Modify** (one-line hook) | `internal/parser/parser.go:188-193` (existing parse-time mutation: `unit.Type` defaulting + `inferGenericType`) |
| `internal/parser/parser_test.go` | Add omitted-name + explicit-name-wins + nested-subunit tests | **Modify** (extend) | same file's existing `TestParse*` functions |
| `testdata/optional_name.toml` *(optional)* | Fixture for the integration test | **Create** *(optional)* | `testdata/valid.toml`, `testdata/nested.toml` (minimal TOML fixtures) |
| `README.md` | Document optional `name` + humanize rules + escape hatch | **Modify** | existing TOML format section |
| `skill/SKILL.md` | Document optional `name` in schema reference | **Modify** | existing skill schema reference section |

## Concrete code excerpts to replicate

### Analog 1 — Small utility file in package `model` (`internal/model/colors.go:1-3`)
```go
package model

// Base colors from C4-PlantUML.
const (
    // ElementFontColor is ...
    ElementFontColor = "#FFFFFF"
    ...
)
```
**Pattern to replicate for `humanize.go`:** `package model` header, a doc comment naming the source (ERGO-04), the exported function, no imports beyond stdlib (`strings`, `unicode`). Keep it ~25 LOC.

### Analog 2 — Helper function in package `model` (`internal/model/link.go:72`)
```go
// FindLinkByPeer returns the Link with the given peer and true,
// or nil and false if no such link exists.
func FindLinkByPeer(links []Link, peer string) (*Link, bool) {
    for i := range links {
        if links[i].Peer == peer {
            return &links[i], true
        }
    }
    return nil, false
}
```
**Pattern to replicate:** plain exported `func` in `package model`, godoc comment, pure logic (no I/O, no receiver needed for `Humanize`).

### Analog 3 — Parse-time mutation hook (`internal/parser/parser.go:181-193`)
```go
var unit model.Unit

if err := toml.Unmarshal(unitData, &unit); err != nil {
    return nil, wrapDecodeError(err)
}

// Apply default type if not specified
if unit.Type == "" {
    unit.Type = defaultTypeForParent(parentType)
}

// Infer level-specific type for generic types (db, queue)
unit.Type = inferGenericType(unit.Type, parentType)
```
**Pattern to replicate — insert the humanize fallback HERE, right after type inference:**
```go
// v1.10 ERGO-03/05: derive display name from the identifier segment when omitted.
// Explicit name = always wins (backward-compat). XC-04 (Phase 31) relocates this
// to a post-expansion pass so templated units humanize from their substituted key.
if unit.Name == "" {
    unit.Name = model.Humanize(name) // 'name' is the func param — already the last segment
}
```
The `name` parameter at parser.go:160 (`parseUnitWithOrder(name string, ...)`) is the last path segment for BOTH top-level units (called at parser.go:87 with `name` from `unitOrder`) and subunits (called at parser.go:210 with `subName` from `subunitOrder`). **One insertion point covers both cases.**

### Analog 4 — Table-driven test with testify (`internal/parser/parser_test.go:30-46`)
```go
func TestParseValidUserUnit(t *testing.T) {
    data, err := os.ReadFile("../../testdata/valid.toml")
    require.NoError(t, err, "failed to read test fixture")

    got, err := Parse(data)
    require.NoError(t, err, "Parse() should not error")
    require.Len(t, got.Units, 2, "should have 2 units")

    unit, ok := got.Units["user"]
    require.True(t, ok, "missing 'user' unit")
    ...
}
```
**Pattern to replicate for `parser_test.go` extensions:**
- Read fixture via `os.ReadFile("../../testdata/optional_name.toml")` (or inline a TOML string for brevity).
- `require.NoError(t, err, ...)` on parse.
- Assert on `unit.Name` with `assert.Equal(t, "Local IDP", unit.Name, "humanized name")`.
- For explicit-name-wins: author a unit with `name = "Explicit"` and no segment-derived override; assert `unit.Name == "Explicit"`.
- For nested subunits: define `[parent.child]` with no `name` on `child`; assert `parent.Subunits["child"].Name` equals the humanized `child` segment.

### Analog 5 — Pure-function table test (new convention for `internal/model/humanize_test.go`)
No existing `*_test.go` in `internal/model/` — `humanize_test.go` establishes the pattern. Mirror `parser_test.go`'s `t.Run` + testify style for consistency:
```go
package model

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestHumanize(t *testing.T) {
    cases := []struct{ in, want string }{
        // ... the 9 rows from RESEARCH.md D-01 reference table ...
    }
    for _, c := range cases {
        t.Run(c.in, func(t *testing.T) {
            assert.Equal(t, c.want, Humanize(c.in))
        })
    }
}
```

## Files NOT to touch (confirmed by RESEARCH.md §Architecture Patterns + research SUMMARY §4 "Unchanged Consumers")

- `internal/validator/*` — reads `m.Units` only; `Unit.Name` is just data to it.
- `internal/view/*` — reads `m.Units`, `m.UnitOrder`; displays `Unit.Name` verbatim.
- `internal/render/*`, `internal/graph/*` — consume `*graph.Graph`, never `*parser.Model`.
- `cmd/c4drill/root.go` — pipeline stays `Parse → Validate → views → render`. The humanize hook is inside `Parse`, so `runRoot` is unchanged in this phase. (Phase 31's XC-04 relocation is when `runRoot` may gain a post-expansion stage.)
- `internal/model/unit.go` — `Unit.Name` field exists (line 45); **read-only touch point**, no struct change.
- `internal/parser/parser.go` `isBuiltinField` (line 309) — `name` is already in the allowlist; **no change**.

## Concurrency note (from task brief)

Phase 28 (Reference field) is concurrently editing `internal/parser/parser.go` (adds `reference` to `isBuiltinField`) and `internal/model/unit.go` (adds `Reference` field). Phase 29's edit to `parser.go` is in `parseUnitWithOrder` (line ~193), which is a **different location** from `isBuiltinField` (line 309) — no merge conflict expected. Phase 29 does NOT touch `unit.go` at all. Both phases' commits should merge cleanly; if a rebase is needed, the two edit sites are disjoint.
