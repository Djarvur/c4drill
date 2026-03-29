---
phase: 21-fix-box-labels-dashed-borders-validator-for-mixed-external-non-external-color-by-content
plan: 02
subsystem: validator, render
tags: [validation, box-coloring, c4-diagrams, external-units]

# Dependency graph
requires:
  - phase: 21-01
    provides: Box dashed border styling infrastructure
provides:
  - Validation rule to prevent mixing external and non-external units in C1 boxes
  - Content-based box coloring (grey for external, dark blue for internal)
affects: [validator, graph, shapes]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Content-based style determination for boxes"
    - "HasExternalSubunits helper for external subunit detection"
    - "ValidateBoxMixedContents validation pattern"

key-files:
  created: []
  modified:
    - internal/validator/rules.go - Added ValidateBoxMixedContents validation rule
    - internal/validator/rules_test.go - Tests for mixed contents validation
    - internal/validator/validator.go - Added rule call to Validate function
    - internal/graph/shapes.go - Added HasExternalSubunits and GetBoxStyleByContents
    - internal/graph/shapes_test.go - Tests for content-based styling
    - internal/graph/builder.go - Updated buildNode/buildCluster/buildNestedCluster for TypeBox

key-decisions:
  - "C1 boxes cannot contain both external and non-external units (validation rule)"
  - "C1 box color is determined by contents: grey for external, dark blue for internal"
  - "C2 containerBox and C3 componentBox are not validated or styled by this rule"

patterns-established:
  - "Content-based style pattern: GetBoxStyleByContents determines style from subunit types"
  - "External subunit detection: HasExternalSubunits iterates subunits checking IsExternalType"

requirements-completed: []

# Metrics
duration: 5min
completed: 2026-03-24
---
# Phase 21 Plan 02: Mixed Content Validation and Box Color by Contents Summary

**Added validation rule to prevent mixing external/non-external units in C1 boxes and content-based coloring for box borders.**

## Performance

- **Duration:** 5 min
- **Started:** 2026-03-24T18:20:01Z
- **Completed:** 2026-03-24T18:25:00Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments

- C1 boxes with both external and non-external units are now rejected by the validator
- C1 boxes with only external units have grey dashed borders (#8A8A8A)
- C1 boxes with only non-external units have dark blue dashed borders (#073B6F)
- C2 containerBox and C3 componentBox are not affected (continue using standard styling)

## Task Commits

Each task was committed atomically:

1. **Task 1: Add ValidateBoxMixedContents validation rule** - `8346ebe` (feat)
2. **Task 2: Color C1 boxes based on their contents** - `ab90a05` (feat)

## Files Created/Modified

- `internal/validator/rules.go` - Added externalTypes map and ValidateBoxMixedContents function
- `internal/validator/rules_test.go` - Added 6 test cases for ValidateBoxMixedContents
- `internal/validator/validator.go` - Added ValidateBoxMixedContents call in Validate function
- `internal/graph/shapes.go` - Added HasExternalSubunits and GetBoxStyleByContents functions
- `internal/graph/shapes_test.go` - Added TestHasExternalSubunits and TestGetBoxStyleByContents tests
- `internal/graph/builder.go` - Updated buildNode, buildCluster, buildNestedCluster to use GetBoxStyleByContents for TypeBox

## Decisions Made

- Only C1 TypeBox is validated for mixed contents (not containerBox or componentBox)
- Empty boxes are treated as internal (dark blue border)
- External detection uses the existing IsExternalType helper function

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- All box-related validation and styling complete
- All tests passing
- Phase 21 complete

---
*Phase: 21-fix-box-labels-dashed-borders-validator-for-mixed-external-non-external-color-by-content*
*Completed: 2026-03-24*

## Self-Check: PASSED

- SUMMARY.md exists
- Task 1 commit (8346ebe) exists
- Task 2 commit (ab90a05) exists
- All tests passing
