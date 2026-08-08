# Phase 32: Include directive (multi-file) - Pattern Map

**Mapped:** 2026-08-08
**Source:** CONTEXT.md (D-09/D-10/D-11/D-12 + carry-forward D-06/D-07/D-08) + RESEARCH.md (Architecture Patterns, Code Examples)

Maps every file to be created/modified to its closest existing analog in the codebase, with concrete line references the executor copies from.

## File Classification

| File | Action | Role | Data Flow | Analog |
|------|--------|------|-----------|--------|
| `internal/parser/parser.go` | MODIFY | parser core | transform (TOML→Model) | itself — `properties` extraction + `Model.Properties` field + `UnitOrder` `toml:"-"` field |
| `internal/parser/parser_test.go` | MODIFY (extend) | test | unit | itself — `TestParseValidProperties` |
| `internal/include/resolve.go` | NEW | service/transform | transform (Model→Model, recursive) | `internal/parser/parser.go` `ParseFile` (file read+parse) + `internal/validator/validator.go` `Validate` (Model iterate + structured error) |
| `internal/include/merge.go` | NEW | service/transform | transform (Model+Model→Model) | `internal/parser/parser.go` `Parse` unit loop (Units/UnitOrder assembly) + `internal/validator/index.go` `BuildIndex` (recursive subunit walk) |
| `internal/include/include_test.go` | NEW | test | unit + integration | `internal/parser/parser_test.go` (fixture-driven, t.Parallel, testify) |
| `internal/include/testdata/*.toml` | NEW | fixture | file-I/O | `testdata/valid.toml`, `testdata/nested.toml` |
| `cmd/c4drill/root.go` | MODIFY | pipeline orchestration | request-response (CLI) | itself — Parse/Validate staging at `:112-118` |

## Pattern Assignments

### File: `internal/parser/parser.go` (MODIFY)

**IMPORTANT CONCURRENCY NOTE:** Phase 28 (reference field) and Phase 31 Plan 1 (BC-1 skip + `Model.Templates`/`Instantiations`) edit this file concurrently. Phase 32 touches **only** the `Model` struct (add `Includes` field) and adds **one** extraction block in `Parse` (after the properties extraction at `:68-77`). Phase 32 does **NOT** re-touch `captureDefinitionOrder` — the `[[include]]` skip lands in Phase 31 Plan 1. Keep the diff minimal to avoid merge conflicts.

**Change 1 — `IncludeDirective` type + `Model.Includes` field.** Analog: existing `Model.Properties` field (`parser.go:37`) and the `toml:"-"` tag on `UnitOrder` (`parser.go:39`).

Copy-from excerpt (current):
```go
// internal/parser/parser.go:35-42
type Model struct {
    Properties model.Properties `toml:"properties"`
    UnitOrder  []string         // tracks definition order (not from TOML)
    Units      map[string]*model.Unit
}
```
Target: add `Includes []IncludeDirective` with `toml:"-"` (populated by explicit extraction in `Parse`, matching `UnitOrder`'s `toml:"-"` convention), plus the `IncludeDirective` type definition (`Path string` + `Once bool`, per CONTEXT "Claude's Discretion"). Place the type near the existing `Model` struct. **NOTE:** Phase 31 Plan 1 adds `Templates`/`Instantiations` to the same struct — both land before Phase 32, so by the time Phase 32 executes, the struct already has those fields. Phase 32 adds `Includes` alongside them.

**Change 2 — `[[include]]` extraction in `Parse`.** Analog: the existing `properties` extraction at `parser.go:68-77`.

Copy-from excerpt (current):
```go
// internal/parser/parser.go:68-77
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
Target: mirror this immediately after the properties block (between `:77` and the unit loop at `:80`) — extract `rawMap["include"]` into `&m.Includes` (array-of-tables → slice, preserving directive order), then `delete(rawMap, "include")` to guarantee it never enters the unit loop. See RESEARCH.md Pattern 1 for the exact target code. The executor reads `parser.go:47-96` (full `Parse`) to place the block.

**Do NOT** modify `captureDefinitionOrder` (`parser.go:100-157`) — the `[[include]]` skip lands in Phase 31 Plan 1. Verify the skip is present (a `parts[0]=="include"` condition near `:128`) before merging Phase 32; if Phase 31 has not landed yet, Phase 32 BLOCKS on Phase 31 (declare `depends_on: ["31-01"]` or `depends_on: ["31"]`).

### File: `internal/parser/parser_test.go` (MODIFY — extend)

**Analog:** itself — `TestParseValidProperties` (`parser_test.go:15-28`).

Copy-from test idiom (current):
```go
// internal/parser/parser_test.go:15-28
func TestParseValidProperties(t *testing.T) {
    t.Parallel()
    data, err := os.ReadFile("../../testdata/valid.toml")
    require.NoError(t, err, "failed to read test fixture")
    got, err := parser.Parse(data)
    require.NoError(t, err, "Parse() should not error")
    assert.Equal(t, "Test System", got.Properties.Name, "Properties.Name")
    // ...
}
```
Target: add `TestParseIncludesExtracted` (parse a TOML string with two `[[include]]` blocks; assert `got.Includes` has len 2 with `Path`/`Once` populated, and `"include"` does NOT appear in `got.UnitOrder`/`got.Units`) and `TestParseNoIncludes` (parse existing `testdata/valid.toml`; assert `got.Includes` is nil/empty — no regression). Use inline `parser.Parse([]byte(...))` for the small cases (no new testdata file needed for the parser-side test — the multi-file fixtures live under `internal/include/testdata/`).

### File: `internal/include/resolve.go` (NEW)

**Analogs (two):**
1. `internal/parser/parser.go` `ParseFile` (`parser.go:323-336`) — for the file-read+parse recursion. `include.Resolve` calls `parser.ParseFile(absPath)` for each included file.
2. `internal/validator/validator.go` `Validate` (`validator.go:16`) — for the transform-function shape: takes `*parser.Model`, returns structured error.

Copy-from excerpt (ParseFile, current):
```go
// internal/parser/parser.go:323-336
func ParseFile(path string) (*Model, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, &ParseError{Message: "failed to read file", Context: path, Cause: err}
    }
    return Parse(data)
}
```

Target shape: `func Resolve(entry *parser.Model, entryDir string) (*parser.Model, error)` — see RESEARCH.md "Resolve signature and recursion" for the full ~50-LOC skeleton. Key points the executor copies from the analogs:
- Error wrapping via `*parser.ParseError{Message: ..., Context: <files>, Cause: err}` (matches `errors.go:13-22`).
- Canonicalization via `filepath.Clean(filepath.Abs(filepath.Join(includingDir, dir.Path)))` (stdlib).
- Stack (cycle) is `[]string`; visited-set (once + diamond) is `map[string]bool`. **Three distinct concerns, two data structures** — see RESEARCH.md Pattern 2.

**Resolve signature decision (RESOLVED in RESEARCH Open Questions):** use `Resolve(entry *parser.Model, entryDir string)` — the `entryDir` parameter is MANDATORY for INC-02 (relative-to-including-file). The CONTEXT's simplified `Resolve(m *parser.Model)` signature is incomplete and must NOT be used. The executor reads RESEARCH.md "State of the Art" for this resolution.

### File: `internal/include/merge.go` (NEW)

**Analogs (two):**
1. `internal/parser/parser.go` `Parse` unit loop (`parser.go:80-93`) — for the `Units` map assembly keyed on `UnitOrder`. Merge unions `Units` the same way, but appends to (not initializes) `UnitOrder`.
2. `internal/validator/index.go` `BuildIndex` (`index.go:23-46`) — for the recursive subunit walk. The cross-file subunit merge (D-10) recurses into `unit.Subunits` the same way `BuildIndex` does.

Copy-from excerpt (Parse unit loop, current):
```go
// internal/parser/parser.go:80-93
for _, name := range unitOrder {
    value, ok := rawMap[name]
    if !ok { continue }
    subunitOrder := subunitOrders[name]
    unit, err := parseUnitWithOrder(name, value, "", subunitOrder, subunitOrders)
    if err != nil { return nil, err }
    m.Units[name] = unit
}
```

Copy-from excerpt (BuildIndex recursive subunit walk — for D-10 merge):
```go
// internal/validator/index.go:23-46 (read full function for the recursion shape)
// BuildIndex walks Units + recurses into unit.Subunits; mergeSubunits mirrors this
// to attach an included file's [parent.child] subunits to an entry-defined [parent].
```

Target: `func merge(dst, src *parser.Model, dstFile, srcFile string) (*parser.Model, error)` plus a helper `func mergeSubunits(dstUnit, srcUnit *model.Unit, path, dstFile, srcFile string) error`. Per-field rules table is RESEARCH.md "Pattern 3: Merge as struct-union". Key correctness points:
- **Units union (INC-07):** for each key in `src.Units`: if absent from `dst.Units`, add it + append key to `dst.UnitOrder` (D-09 append order); if present in both, call `mergeSubunits` (D-10) — do NOT overwrite.
- **Subunits (D-10):** for each child key in `srcUnit.Subunits`: if absent from `dstUnit.Subunits`, add + append to `dstUnit.SubunitOrder`; if present in both, hard-error naming both files.
- **Properties (INC-08):** root-wins — only copy a field from `src.Properties` if `dst.Properties`'s field is zero-value; any non-zero/non-zero overlap → hard-error naming both files. (All `model.Properties` fields per CONTEXT discretion — see RESEARCH Open Question 2, resolved.)
- **Templates (XC-02):** union `dst.Templates` with `src.Templates`; dup key → hard-error naming both files. (Templates flow through the merge so Phase 31's `template.Expand` sees included templates.)
- **Instantiations:** `dst.Instantiations = append(dst.Instantiations, src.Instantiations...)` (order preserved; drained by Phase 31's Expand later).
- **Includes:** not merged — `src.Includes` was already resolved by the recursive `Resolve` call before merge.

### File: `internal/include/include_test.go` (NEW)

**Analog:** `internal/parser/parser_test.go` — fixture-driven tests reading `testdata/*.toml`, `t.Parallel()`, testify `require`/`assert`.

Copy-from test idiom (current):
```go
// internal/parser/parser_test.go:1-13 + 172-194 (TestParseLinksOutgoing shape)
package parser_test  // → package include_test

import (
    "os"
    "testing"
    "github.com/Djarvur/c4drill/internal/parser"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestResolveTwoFilesMerge(t *testing.T) {
    t.Parallel()
    entry, err := parser.ParseFile("testdata/main.toml")   // relative to internal/include/
    require.NoError(t, err)
    merged, err := include.Resolve(entry, "testdata")
    require.NoError(t, err)
    // assert merged.Units has both main's and auth's units
    // assert merged.UnitOrder == []string{<main's units>, <auth's units in include order>}
}
```

**Test fixture layout:** fixtures live under `internal/include/testdata/` (per-package, matching `internal/parser`'s `../../testdata/` convention). Multi-file fixtures need sibling files in the same dir so relative paths resolve (e.g. `testdata/main.toml` has `[[include]] path="auth.toml"`). For cycle/diamond tests, create multi-file graphs (`cycle_a.toml` ↔ `cycle_b.toml`; `diamond_*` 4-file graph).

**cd-proof test (INC-02):** save the current cwd via `os.Getwd()`, `os.Chdir` to a temp dir, run `Resolve`, assert success + identical merged output, restore cwd. (Or: resolve the same fixture from two different cwds and assert the merged `*parser.Model` is identical.)

### File: `internal/include/testdata/*.toml` (NEW fixtures)

**Analog:** `testdata/valid.toml` (hand-authored reference, `[properties]`/`[unit]`/`[[link]]` shape) and `testdata/nested.toml` (subunit shape for D-10).

Required fixtures:
- `main.toml` + `auth.toml` + (optional) `billing.toml` — basic 2-3 file merge (INC-01, INC-09).
- `main_transitive.toml` + `mid.toml` + `leaf.toml` — transitive chain (INC-03).
- `cycle_a.toml` + `cycle_b.toml` — mutual cycle (A→B→A) (INC-04). Plus `self_cycle.toml` (self-include) (INC-04).
- `diamond_top.toml` + `diamond_left.toml` + `diamond_right.toml` + `diamond_shared.toml` — diamond (A→B→D, A→C→D) NOT a cycle (INC-05, D-11).
- `once_main.toml` + `once_lib.toml` — `once=true` dedup (INC-06).
- `dup_main.toml` + `dup_other.toml` (both define `[mailAdapter]`) — cross-file dup-path hard-error (INC-05, INC-07, D-11).
- `props_main.toml` + `props_conflict.toml` (both define `[properties] name=`) — properties conflict (INC-08).
- `missing_main.toml` (includes nonexistent `ghost.toml`) — missing include (INC-10, D-12).
- `subunits_main.toml` (`[linuxSystem]`) + `subunits_auth.toml` (`[linuxSystem.auth]`, `[linuxSystem.db]`) — cross-file subunit merge (D-10).
- `templates_lib.toml` (defines `[template.svc]`) + `templates_main.toml` (`[[include]]` + `[[use]] template="svc"`) — XC-02 template isolation (conditional on Phase 31).

Each mirrors the `[properties]`/`[unit]`/`[[link]]`/`[[include]]` TOML shape of existing fixtures. Keep fixtures minimal (1-3 units each) — the tests assert merge semantics, not rendering.

### File: `cmd/c4drill/root.go` (MODIFY)

**Analog:** the existing Parse → Validate staging at `root.go:112-118`.

Copy-from excerpt (current):
```go
// cmd/c4drill/root.go:112-118
// Stage 1: Parse
m, err := parser.ParseFile(inputPath)
if err != nil {
    return fmt.Errorf("parse: %w", err)
}
// Stage 2: Validate
valErrors := validator.Validate(m)
```
Target: insert between `:115` (after ParseFile) and `:118` (before Validate):
```go
// Stage 1a: Resolve includes (recursive, merges *parser.Model structs) — Phase 32; runs FIRST
if m, err = include.Resolve(m, filepath.Dir(inputPath)); err != nil {
    return fmt.Errorf("include: %w", err)
}
```
Add `"github.com/Djarvur/c4drill/internal/include"` to the import block. The executor reads `cmd/c4drill/root.go:100-135` to see the full staging and the `fmt.Errorf("stage: %w", err)` idiom. **Pipeline ordering is load-bearing:** include.Resolve MUST be the FIRST pre-processing pass (before template.Expand when Phase 31 lands, before peer.Resolve when Phase 30 lands) so included templates are visible to `[[use]]` (XC-02). See RESEARCH.md System Architecture Diagram.

**NOTE on Phase 31 concurrency:** Phase 31's `template.Expand` call slots into the SAME insertion point (between ParseFile and Validate). When both Phase 31 and Phase 32 are merged, the order MUST be `include.Resolve` THEN `template.Expand`. Both phases edit `root.go`; the merge requires `include` before `template` — document this in the plan so the executor handles the ordering during integration.

## No Analog Found

(none — every Phase 32 file has a clear codebase analog. The new `internal/include/` package is a composition of existing idioms: `ParseFile` recursion + `Parse` unit-assembly + `BuildIndex` subunit walk + `ParseError` error shape.)

## Shared Patterns

### Error handling / reporting
- All include errors (cycle, missing file, dup path, properties conflict, dup template) are `*parser.ParseError` (`internal/parser/errors.go:13-34`) with `Message`/`Context`/`Cause`. Use `Context` to carry the filename(s) — e.g. `"mailAdapter defined in both main.toml and auth.toml"` or `"ghost.toml (included from main.toml)"`. This matches the CLI's existing `fmt.Errorf("include: %w", err)` path so errors report uniformly. The executor reads `internal/parser/errors.go` (full file, ~80 LOC).

### Non-strict TOML (load-bearing)
go-toml/v2 stays non-strict (STATE.md DI-1 / BC-3, PITFALLS BC-3). The `toml:",inline"` subunit trick (`unit.go:71`) depends on unknown keys being accepted. The `[[include]]` extraction must NOT enable `DisallowUnknownFields()`. Add a guard comment near the new extraction block mirroring the existing non-strict reliance. The executor reads `internal/model/unit.go:71` for the inline tag.

### Canonicalization (load-bearing for cycle/dedup correctness)
Every include path is canonicalized via `filepath.Clean(filepath.Abs(filepath.Join(includingDir, dir.Path)))` BEFORE any stack/visited-set operation. Non-canonical paths (`./x.toml`, `x.toml`, `./a/../x.toml`) would dodge cycle detection and `once`/diamond dedup (PITFALLS IN-4). The canonical form is the map/slice key everywhere. `filepath.Abs` calls `Clean` on its result per Go stdlib docs; the explicit `Clean` is belt-and-suspenders.

### testify test idiom (all test files)
`t.Parallel()` + `require.NoError` for preconditions + `assert.Equal`/`require.Len`/`require.True` for assertions. Fixtures via `os.ReadFile("testdata/...")` (per-package fixtures) or inline `parser.Parse([]byte(...))` for small cases. All new/extended test files follow this. The executor reads `internal/parser/parser_test.go:1-13, 15-28` for the idiom.

### Order preservation (load-bearing)
`captureDefinitionOrder` exists to preserve authoring order (`parser.go:100-157`). Per D-09, the merge CONCATENATES already-ordered `UnitOrder` slices (entry's units first, then each included file's units appended in include-directive order) — it does NOT re-run order capture on the merged model, and does NOT sort. Subunit order (D-10) appends included subunits after entry-authored subunits on the same parent. The executor reads `parser.go:80-93` (unit loop keyed on UnitOrder) and ensures merge appends, never sorts.

### canonicalDOT golden comparison (DI-1 — may not exist yet)
STATE.md DI-1 mandates order-insensitive canonicalDOT comparison (sort-normalize, strip layout geometry bb/pos/lp/lheight/lwidth) for all multi-file goldens — NOT byte-exact `require.Equal`. **Verification:** a grep for `canonicalDOT`/`canonicalDot`/`stripLayout`/`lheight` across `internal/` found NO existing helper. This means the canonicalDOT comparator likely lands in **Phase 33** (which owns "end-to-end golden tests" per ROADMAP). **Planner decision required:** Phase 32 should assert merge-equivalence at the `*parser.Model` level (merged `Units`/`UnitOrder` equals a hand-constructed expectation) rather than at the rendered-SVG level, deferring the SVG golden to Phase 33. Do NOT reinvent the canonicalDOT comparator in Phase 32 — note it as a Phase 33 dependency.

### Validator as single gatekeeper (STATE.md D-12)
`include.Resolve` produces a `*parser.Model` whose `.Units`/`.UnitOrder`/`.Properties` are indistinguishable from a hand-authored single-file model. `validator.Validate` consumes it UNCHANGED — zero changes to `validator/`, `view/`, `render/`, `graph/`. The executor confirms this by reading `internal/validator/validator.go:16-30` (Validate calls `BuildIndex(m.Units, "")` then runs rules — operates on the flat unit set, file-origin-agnostic).
