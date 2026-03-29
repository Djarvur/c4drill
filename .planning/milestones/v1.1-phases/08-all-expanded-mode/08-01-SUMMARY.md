---
phase: 08-all-expanded-mode
plan: 01
subsystem: view
tags: [expanded-mode, recursive-traversal, nested-clusters, output-writer]

# Dependency graph
requires: []
provides:
  - GenerateExpandedView function for recursive unit collection
  - Cluster.Clusters field for arbitrary nesting depth
  - WriteExpanded method for expanded output path
affects: [08-02]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Recursive unit traversal with dotted path construction
    - Nested cluster structure for GraphViz rendering
    - Expanded output file naming convention

key-files:
  created: []
  modified:
    - internal/view/scope.go
    - internal/graph/graph.go
    - internal/output/writer.go

key-decisions:
  - "GenerateExpandedView uses LevelC1 (same as C1 view) per locked decision"
  - "IsExpanded set to true when unit has subunits (always show nested structure)"
  - "addExternalBoundaryNodesRecursive scans all nested subunits for links"
  - "Cluster.Clusters field is non-breaking (nil slice is safe for existing code)"

patterns-established:
  - "Pattern: Recursive traversal with addUnitRecursive helper for depth-first collection"
  - "Pattern: External boundary nodes scanned recursively from all nesting levels"

requirements-completed: [EXPD-02, EXPD-04]

# Metrics
duration: 4 min
completed: 2026-03-10
---

# Phase 08 Plan 01: Core Data Structures Summary

**Added GenerateExpandedView function, extended Cluster struct for nested clusters, and WriteExpanded method for expanded output path handling.**

## Performance

- **Duration:** 4 min
- **Started:** 2026-03-10T23:03:05Z
- **Completed:** 2026-03-10T23:07:23Z
- **Tasks:** 3
- **Files modified:** 3

## Accomplishments

- GenerateExpandedView recursively collects all units at all nesting levels with proper dotted paths
- Cluster struct extended with Clusters field for arbitrary nesting depth
- WriteExpanded method creates output at {basename}.expanded.{ext} path

## Task Commits

Each task was committed atomically:

1. **Task 1: Add GenerateExpandedView function to scope.go** - `4faa0ff` (test) + `ebdd18c` (feat)
2. **Task 2: Extend Cluster struct with nested Clusters field** - `4ff5545` (feat)
3. **Task 3: Add WriteExpanded method to writer.go** - `3415cdc` (feat)

**Plan metadata:** pending (will be added after summary creation)

_Note: Task 1 was TDD with RED (test commit) then GREEN (implementation commit)_

## Files Created/Modified

- `internal/view/scope.go` - Added GenerateExpandedView, addUnitRecursive, addExternalBoundaryNodesRecursive
- `internal/graph/graph.go` - Added Clusters []*Cluster field to Cluster struct
- `internal/output/writer.go` - Added WriteExpanded method for {basename}.expanded.{ext} output
- `internal/view/scope_test.go` - Added 8 tests for GenerateExpandedView
- `internal/output/writer_test.go` - Added 5 tests for WriteExpanded

## Decisions Made

- **Level for expanded view:** Uses LevelC1 (consistent with locked decision to use modified C1 approach)
- **IsExpanded logic:** Set to true when unit has subunits (always show nested structure in expanded view)
- **External boundary nodes:** Recursive scanning of all nested subunits for links (not just top-level)
- **Non-breaking Cluster extension:** Clusters field is nil by default, safe for existing code

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Core data structures in place for all-expanded mode
- GenerateExpandedView ready to be consumed by graph builder (next plan)
- WriteExpanded ready to be called from CLI (next plan)
- Cluster.Clusters field ready for nested cluster building (next plan)

---
*Phase: 08-all-expanded-mode*
*Completed: 2026-03-10*

## Self-Check: PASSED

- All 3 modified files exist on disk
- All 4 task commits found in git history
- All tests pass
