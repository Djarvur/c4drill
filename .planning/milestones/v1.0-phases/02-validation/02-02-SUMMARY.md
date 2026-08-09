---
phase: 02-validation
plan: 02
subsystem: validation
tags: [validation-rules, reference-checking, type-constraints, link-restrictions, tdd]

# Dependency graph
requires:
  - phase: 02-01
    provides: ValidationError type, BuildIndex function, SuggestSimilar helper
provides:
  - Validate function orchestrating all validation rules
  - ValidateReferences checking unit reference integrity
  - ValidateSubunitRules enforcing type constraints
  - ValidateLinkRules enforcing link restrictions
  - ReportErrors helper for formatted error output
affects: [02-03, 02-04]

# Tech tracking
tech-stack:
  added: []
  patterns: [error collection pattern, rule function separation, io.Writer for output]

key-files:
  created:
    - internal/validator/rules.go
    - internal/validator/rules_test.go
    - internal/validator/validator.go
    - internal/validator/validator_test.go
  modified: []

key-decisions:
  - "Rule functions take only index parameter (units removed as unused)"
  - "ReportErrors accepts io.Writer for flexibility (not hardcoded to stderr)"
  - "Errors preallocated with typicalErrorCount constant for performance"

patterns-established:
  - "Individual rule functions returning ValidationErrors slice"
  - "Main Validate function building index once and passing to all rules"
  - "ReportErrors returns exit code (1 for errors, 0 for success)"

requirements-completed: [VALD-01, VALD-02, VALD-03, VALD-04]

# Metrics
duration: 5min
completed: 2026-03-09
---

# Phase 2 Plan 02: Validation Rules Summary

**Complete validation logic with reference checking, type constraints, and link restrictions using TDD approach**

## Performance

- **Duration:** 5 min
- **Started:** 2026-03-09T19:45:43Z
- **Completed:** 2026-03-09T19:50:43Z
- **Tasks:** 3
- **Files modified:** 6

## Accomplishments
- ValidateReferences detects undefined Links and LinksFrom targets with suggestions
- ValidateSubunitRules enforces that only system/systemExternal/box can have subunits
- ValidateLinkRules prevents links on/from units with subunits
- Validate orchestrates all rules and collects all errors (not fail-fast)
- ReportErrors prints formatted errors with summary to any io.Writer
- 100% test coverage with all lint checks passing

## Task Commits

Each task was committed atomically:

1. **Task 1: Create validation rule functions** - `d2dbebb` (feat)
2. **Task 2: Create main Validate function and ReportErrors** - `7381f42` (feat)
3. **Task 3: Run lint and fix issues** - `0988296` (fix)

**Plan metadata:** (pending final commit)

_Note: TDD tasks followed RED-GREEN-REFACTOR cycle with tests written first_

## Files Created/Modified
- `internal/validator/rules.go` - Three validation rule functions (ValidateReferences, ValidateSubunitRules, ValidateLinkRules)
- `internal/validator/rules_test.go` - Comprehensive tests for each rule function
- `internal/validator/validator.go` - Main Validate function and ReportErrors helper
- `internal/validator/validator_test.go` - Tests for Validate and ReportErrors

## Decisions Made
- Removed unused `units` parameter from rule functions for cleaner API
- ReportErrors accepts io.Writer instead of hardcoding os.Stderr for testability
- Preallocated error slice with named constant (typicalErrorCount=4) to avoid magic number lint error

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Removed unused parameter from rule functions**
- **Found during:** Task 3 (lint verification)
- **Issue:** Rule functions had unused `units` parameter triggering revive linter
- **Fix:** Removed parameter from all three rule functions and updated callers
- **Files modified:** internal/validator/rules.go, internal/validator/validator.go, internal/validator/rules_test.go
- **Verification:** All tests pass, lint clean
- **Committed in:** 0988296 (Task 3 commit)

**2. [Rule 3 - Blocking] Fixed multiple lint issues**
- **Found during:** Task 3 (lint verification)
- **Issue:** errcheck warnings, funlen violations, gocognit complexity, paralleltest/tparallel missing, prealloc warning, magic number
- **Fix:** Ignored fmt.Fprintln return values with `_, _`, split large test functions into smaller ones, added t.Parallel() to all tests, preallocated errors slice, defined typicalErrorCount constant
- **Files modified:** internal/validator/validator.go, internal/validator/validator_test.go, internal/validator/rules_test.go
- **Verification:** `mise run lint` returns 0 issues
- **Committed in:** 0988296 (Task 3 commit)

**3. [Rule 3 - Blocking] Removed duplicate test functions**
- **Found during:** Task 3 (test execution)
- **Issue:** TestValidationErrors_Empty and similar tests duplicated in validator_test.go (already existed in errors_test.go)
- **Fix:** Removed duplicate tests from validator_test.go
- **Files modified:** internal/validator/validator_test.go
- **Verification:** Tests compile and pass
- **Committed in:** 0988296 (Task 3 commit)

---

**Total deviations:** 3 auto-fixed (1 bug, 2 blocking)
**Impact on plan:** All fixes necessary for code quality. No scope creep.

## Issues Encountered
None - TDD workflow proceeded smoothly with all tests passing after implementation.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Validation rules complete and ready for CLI integration
- Validate function can be called from main CLI flow
- ReportErrors provides formatted output to any io.Writer
- All four validation requirements (VALD-01 through VALD-04) implemented

## Self-Check: PASSED

All files and commits verified.

---
*Phase: 02-validation*
*Completed: 2026-03-09*
