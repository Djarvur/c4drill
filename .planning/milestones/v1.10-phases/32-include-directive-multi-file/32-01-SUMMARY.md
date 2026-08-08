---
phase: 32-include-directive-multi-file
plan: 01
subsystem: parser
tags: [toml, parser, include-directive, multi-file]

# Dependency graph
requires:
  - phase: 31-template-expansion
    provides: BC-1 reserved-table parser skip for [[include]] in captureDefinitionOrder + Model struct extension convention (Templates/Instantiations as toml:"-")
provides:
  - "parser.IncludeDirective type (Path string, Once bool)"
  - "parser.Model.Includes field ([]IncludeDirective, toml:\"-\") populated by Parse"
  - "extractIncludes + parseIncludeDirective helpers in parser.Parse"
affects: [32-02, 33-04]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Reserved-table extraction for array-of-tables uses a []any walker (extractInstantiations style), NOT the properties marshal/unmarshal pattern (which fails on top-level arrays)"

key-files:
  created: []
  modified:
    - internal/parser/parser.go
    - internal/parser/parser_test.go

key-decisions:
  - "Used an extractIncludes+parseIncludeDirective walker for [[include]] (mirrors extractInstantiations) instead of the properties marshal/unmarshal pattern — go-toml/v2 does not marshal a top-level []any back into valid TOML"

patterns-established:
  - "Array-of-tables extraction: walk []any directly, type-assert each map[string]any entry, decode fixed fields by key switch, accept unknown keys (non-strict toml, BC-3)"

requirements-completed: [INC-01]

# Metrics
duration: 6min
completed: 2026-08-08
---

# Phase 32 Plan 01: IncludeDirective extraction (parser-side) Summary

**IncludeDirective type + Model.Includes field landed so `[[include]]` array-of-tables route into the Model in document order with zero phantom units — the foundation Plan 02's resolver consumes.**

## Performance

- **Duration:** ~6 min
- **Tasks:** 2
- **Files modified:** 2 (internal/parser/parser.go, internal/parser/parser_test.go)

## Accomplishments
- `IncludeDirective` struct (`Path string toml:"path"`, `Once bool toml:"once"`) defined next to the Model struct.
- `Model.Includes []IncludeDirective` field (`toml:"-"`) added, populated by `Parse` via the new `extractIncludes` helper.
- `[[include]]` blocks now extract into `Model.Includes` in document order with correct Path/Once values; `Once` defaults false when the key is omitted.
- Zero phantom units: `[[include]]` does not appear in `UnitOrder` or `Units` (consumes the Phase 31 BC-1 skip; `delete(rawMap, "include")` is belt-and-suspenders).
- No regression: single-file models (no `[[include]]`) parse identically — `Includes` is nil/empty, `Properties`/`UnitOrder`/`Units` unchanged.
- `captureDefinitionOrder` and `isBuiltinField` UNCHANGED (verified: diff is pure additions, no edits to those regions).

## Task Commits

1. **Task 1: Write failing parser tests for [[include]] extraction (RED)** — `281137e` (test)
2. **Task 2: Implement IncludeDirective type, Model.Includes field, and extraction (GREEN)** — `058c465` (feat)

## Files Created/Modified
- `internal/parser/parser.go` — `IncludeDirective` type, `Model.Includes` field, `extractIncludes` + `parseIncludeDirective` helpers, extraction call wired into `Parse` between `extractInstantiations` and the unit loop.
- `internal/parser/parser_test.go` — `TestParseIncludesExtracted`, `TestParseIncludesOnceDefaultsFalse`, `TestParseNoIncludes` (3 new tests).

## Decisions Made
- Used `extractIncludes` + `parseIncludeDirective` walker for `[[include]]` instead of the properties marshal/unmarshal pattern. The plan's `<implementation>` suggested the marshal/unmarshal pattern, but `extractInstantiations` is the correct analog for array-of-tables — go-toml/v2 cannot re-marshal a top-level `[]any` into valid TOML.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Extraction pattern for array-of-tables**
- **Found during:** Task 2 (Implement extraction — GREEN)
- **Issue:** The plan's `<implementation>` and `<interfaces>` described mirroring the `properties` marshal/unmarshal extraction (parser.go:106-115). The first implementation did exactly that: `toml.Marshal(inc)` then `toml.Unmarshal(incData, &m.Includes)`. This produced `toml: invalid character at start of key: ]` on every test with `[[include]]` (including the pre-existing `TestParseIncludeReservedSkipped` from Phase 31). Root cause: `rawMap["include"]` is a `[]any` of array-of-tables; marshaling a bare `[]any` at the top of a document is not valid TOML (arrays must live inside a table). The properties case works because `rawMap["properties"]` is a single `map[string]any` (one table).
- **Fix:** Switched to the `extractInstantiations` pattern — walk the `[]any`, type-assert each `map[string]any` entry, decode the `path`/`once` keys via a per-entry helper. Extracted `parseIncludeDirective` as a separate function to keep cognitive complexity under the linter threshold (gocognit 18 → well below 15 after the split).
- **Files modified:** internal/parser/parser.go
- **Verification:** `go test ./internal/parser/` fully green (3 new tests + all pre-existing incl. `TestParseIncludeReservedSkipped`); `go test ./...` fully green; `golangci-lint run ./internal/parser/` reports zero new findings in parser.go (gocognit and wsl_v5 from the first attempt both gone); `go vet ./internal/parser/` clean.
- **Committed in:** `058c465` (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** The plan's `<interfaces>` extraction analog was misidentified (properties is a table, include is an array-of-tables). The fix aligns with the correct existing analog (`extractInstantiations`), produces clean lint, and matches the established codebase pattern. No scope creep.

## Issues Encountered
None beyond the deviation above.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- `Model.Includes` is now available for Plan 32-02's `internal/include.Resolve(entry, entryDir)` to walk.
- Plan 32-02 is unblocked (Wave 2).
- `captureDefinitionOrder`/`isBuiltinField` confirmed untouched — safe to proceed to the include package without parser conflicts.

## Self-Check: PASSED

- `go test ./internal/parser/ -x` — all parser tests green (3 new + all existing, no regression)
- `go test ./...` — full suite green
- `golangci-lint run ./internal/parser/` — no new findings in parser.go (pre-existing test-file findings unchanged)
- `go vet ./internal/parser/` — clean
- `git diff internal/parser/parser.go` confined to the Model struct + IncludeDirective type + new helper + extraction call (pure additions, 109 insertions, 0 deletions; `captureDefinitionOrder` and `isBuiltinField` unchanged)

---
*Phase: 32-include-directive-multi-file*
*Completed: 2026-08-08*
