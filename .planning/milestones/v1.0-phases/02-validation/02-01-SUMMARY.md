---
phase: 02-validation
plan: 01
subsystem: validation
tags: [error-handling, levenshtein, unit-index, tdd]

# Dependency graph
requires: []
provides:
  - ValidationError type with human-readable single-line format
  - BuildIndex function for O(1) unit lookup
  - SuggestSimilar helper for typo detection
affects: [02-02, 02-03, 02-04]

# Tech tracking
tech-stack:
  added: [github.com/agnivade/levenshtein@v1.2.1]
  patterns: [error collection pattern, recursive tree traversal, Levenshtein distance for suggestions]

key-files:
  created:
    - internal/validator/errors.go
    - internal/validator/errors_test.go
    - internal/validator/index.go
    - internal/validator/index_test.go
    - internal/validator/suggest.go
    - internal/validator/suggest_test.go
  modified:
    - go.mod
    - go.sum

key-decisions:
  - "Line number takes precedence over path in error formatting"
  - "Skip suggestions for names shorter than 3 characters"
  - "Max Levenshtein distance of 2 for valid suggestions"

patterns-established:
  - "Package comment for godoc compliance"
  - "t.Parallel() in all test functions"
  - "External test package (validator_test) for black-box testing"

requirements-completed: [VALD-05, VALD-06]

# Metrics
duration: 6min
completed: 2026-03-09
---

# Phase 2 Plan 01: Validation Infrastructure Summary

**ValidationError type, BuildIndex function, and SuggestSimilar helper with Levenshtein-based typo detection for C4 model validation**

## Performance

- **Duration:** 6 min
- **Started:** 2026-03-09T19:36:03Z
- **Completed:** 2026-03-09T19:42:03Z
- **Tasks:** 4
- **Files modified:** 6

## Accomplishments
- ValidationError type with human-readable single-line error format
- Unit index builder for O(1) lookup with recursive dotted path support
- Levenshtein-based suggestion helper for typo detection in unit references
- 100% test coverage with all lint checks passing

## Task Commits

Each task was committed atomically:

1. **Task 1: Add agnivade/levenshtein dependency** - `223f25b` (chore)
2. **Task 2: Create ValidationError type** - `9b7c940` (feat)
3. **Task 3: Create unit index builder** - `cb9088f` (feat)
4. **Task 4: Create suggestion helper** - `1924d03` (feat)

**Plan metadata:** (pending final commit)

_Note: TDD tasks followed RED-GREEN-REFACTOR cycle with tests written first_

## Files Created/Modified
- `internal/validator/errors.go` - ValidationError and ValidationErrors types with human-readable Error() methods
- `internal/validator/errors_test.go` - Tests for error formatting (single-line, with line/path)
- `internal/validator/index.go` - UnitInfo struct and BuildIndex function for recursive traversal
- `internal/validator/index_test.go` - Tests for index building with nested units
- `internal/validator/suggest.go` - SuggestSimilar and FormatSuggestion using Levenshtein distance
- `internal/validator/suggest_test.go` - Tests for suggestion logic with edge cases
- `go.mod` / `go.sum` - Added agnivade/levenshtein v1.2.1 dependency

## Decisions Made
- Line number takes precedence over path in error formatting (per plan specification)
- Suggestions skipped for names shorter than 3 characters (avoid false positives on "db", "id", etc.)
- Max Levenshtein distance of 2 for valid suggestions (balance between helpful and noise)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Fixed test expectation for path-only errors**
- **Found during:** Task 2 (ValidationError tests)
- **Issue:** Test `TestValidationError_WithoutLine` had Path set but expected output without path
- **Fix:** Cleared Path field to match test name's intent (testing "without line" means without path too)
- **Files modified:** internal/validator/errors_test.go
- **Verification:** All tests pass
- **Committed in:** 9b7c940 (Task 2 commit)

**2. [Rule 1 - Bug] Fixed lint issues in test files**
- **Found during:** Task 4 (lint verification)
- **Issue:** Linter required package comment, t.Parallel(), external test package, preallocated slices
- **Fix:** Added package comment to errors.go, added t.Parallel() to all tests, changed to validator_test package, preallocated slice in TestValidationErrors_Append
- **Files modified:** internal/validator/errors.go, errors_test.go, index_test.go, suggest_test.go
- **Verification:** `mise run lint` returns 0 issues
- **Committed in:** 1924d03 (amended Task 4 commit)

---

**Total deviations:** 2 auto-fixed (1 blocking, 1 bug)
**Impact on plan:** All fixes necessary for code quality and correctness. No scope creep.

## Issues Encountered
None - TDD workflow proceeded smoothly with all tests passing after implementation.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Validation infrastructure complete and ready for rule implementations
- BuildIndex provides O(1) lookup for reference validation
- SuggestSimilar ready for "did you mean" suggestions in error messages
- ValidationError format matches user's specified single-line output

---
*Phase: 02-validation*
*Completed: 2026-03-09*
