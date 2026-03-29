---
phase: 05-navigation
plan: 01
subsystem: graph
tags: [navigation, url, path, breadcrumb, backlink, graphviz]

# Dependency graph
requires:
  - phase: 04-rendering-output
    provides: Output writer with file path structure for computing relative URLs
provides:
  - Navigation types (Navigation, BackLink, BreadcrumbItem)
  - Path computation utilities (ComputeExploreURL, ComputeBackLinkURL, BuildBreadcrumbPath)
  - BuildGraphWithPath function for graphs with navigation data
  - Node.ExploreURL field for drill-down links
  - Graph.Navigation field for breadcrumb and back-link info
affects: [rendering]

# Tech tracking
tech-stack:
  added: []
  patterns: [url-encoding, relative-path-computation, breadcrumb-trail]

key-files:
  created:
    - internal/graph/navigation.go
    - internal/graph/path.go
    - internal/graph/navigation_test.go
    - internal/graph/path_test.go
  modified:
    - internal/graph/graph.go
    - internal/graph/builder.go
    - internal/graph/builder_test.go

key-decisions:
  - "Only system and box types get explore links (not person/db/queue)"
  - "Path segments URL-encoded individually to preserve directory separators"
  - "C3 back-link uses parent directory name, not last segment"

patterns-established:
  - "TDD approach: write failing tests first, then implement"
  - "Path computation uses dotted notation for unit paths"

requirements-completed: [REND-04, OUTP-05]

# Metrics
duration: 10min
completed: 2026-03-10
---

# Phase 5 Plan 1: Navigation Types Summary

**Navigation types and path computation utilities for clickable C4 diagram links with explore URLs, back-links, and breadcrumb trails**

## Performance

- **Duration:** 10 min
- **Started:** 2026-03-10T11:19:49Z
- **Completed:** 2026-03-10T11:29:02Z
- **Tasks:** 3
- **Files modified:** 6

## Accomplishments

- Added Navigation, BackLink, and BreadcrumbItem types for diagram navigation
- Implemented path computation utilities (ComputeExploreURL, ComputeBackLinkURL, BuildBreadcrumbPath, URLEncodePath)
- Extended Graph and Node types with navigation fields (Navigation, ExploreURL)
- Created BuildGraphWithPath function that computes navigation metadata

## Task Commits

Each task was committed atomically:

1. **Task 1: Add navigation types to graph package** - `779e062` (test/feat)
2. **Task 2: Implement path computation utilities** - `26163df` (feat)
3. **Task 3: Extend builder to compute navigation URLs** - `8afd4aa` (feat)

## Files Created/Modified

- `internal/graph/navigation.go` - Navigation, BackLink, BreadcrumbItem types
- `internal/graph/path.go` - Path computation utilities
- `internal/graph/navigation_test.go` - Tests for navigation types
- `internal/graph/path_test.go` - Tests for path computation
- `internal/graph/graph.go` - Added Navigation and ExploreURL fields
- `internal/graph/builder.go` - Added BuildGraphWithPath function
- `internal/graph/builder_test.go` - Tests for BuildGraphWithPath

## Decisions Made

- Only system and box types get explore links (not person/db/queue) - per CONTEXT.md decision
- Path segments URL-encoded individually to preserve directory separators
- C3 back-link uses parent directory name (e.g., `../mainsystem.svg` for `mainapp.api`)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed C3->C2 back-link URL test expectation**

- **Found during:** Task 2 (Path computation tests)
- **Issue:** Plan specified `../api.svg` for `mainapp.api` back-link, but actual file structure uses `../mainsystem.svg`
- **Fix:** Updated test expectation to match actual output writer file structure
- **Files modified:** internal/graph/path_test.go
- **Verification:** All tests pass
- **Committed in:** 26163df (Task 2 commit)

**2. [Rule 1 - Bug] Fixed URL encoding to preserve directory separators**

- **Found during:** Task 2 (Path computation implementation)
- **Issue:** url.PathEscape encodes slashes, breaking directory hierarchy
- **Fix:** Encode each path segment separately, then join with slashes
- **Files modified:** internal/graph/path.go
- **Verification:** Tests for C2->C3 navigation pass
- **Committed in:** 26163df (Task 2 commit)

---

**Total deviations:** 2 auto-fixed (2 bug fixes)

**Impact on plan:** Both fixes corrected plan specification to match actual file structure. No scope creep.

## Issues Encountered

None - implementation followed TDD approach smoothly.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Navigation types ready for renderer integration
- Path computation utilities tested and verified
- BuildGraphWithPath provides all navigation data needed for SVG rendering

## Self-Check: PASSED

- All 6 key files verified to exist
- All 3 task commits verified in git history (779e062, 26163df, 8afd4aa)

---
*Phase: 05-navigation*
*Completed: 2026-03-10*
