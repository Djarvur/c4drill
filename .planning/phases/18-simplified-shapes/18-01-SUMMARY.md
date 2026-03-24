---
phase: 18-simplified-shapes
plan: 01
subsystem: render
tags: [graphviz, cylinder, emoji, html-labels, shapes]

# Dependency graph
requires:
  - phase: 17-units-labels-cells-must-be-word-wrapped-to-make-the-unit-shape-proportions-as-close-as-possible-to-credit-card-proportions
    provides: Word-wrap functionality for labels
provides:
  - Native GraphViz cylinder shapes for DB/Queue units
  - Emoji-based Person labels (no external icon files)
  - Simplified single-column HTML labels for System/Box/Container/Component
  - Removed icon extraction system entirely
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Native GraphViz shapes (cylinder) instead of custom SVG icons
    - HTML emoji entities for Person labels
    - Single-column label layout for non-Person units

key-files:
  created: []
  modified:
    - internal/render/converter.go - Shape assignment with cylinder support
    - internal/render/render.go - Simplified render without icon extraction
    - internal/render/labels.go - Emoji labels and single-column layout
    - internal/render/wrap.go - Added labelMaxCharsNoIcon function
    - internal/render/html_labels_internal_test.go - Updated tests

key-decisions:
  - "Use GraphViz native shape=cylinder for DB units"
  - "Use GraphViz native shape=cylinder with 90deg orientation for Queue units"
  - "Person labels use HTML emoji entity (&#x1F464;) instead of SVG images"
  - "System/Box/Container/Component labels use single-column 3-row layout"

patterns-established:
  - "Cylinder shapes: cn.SetShape(cgraph.CylinderShape)"
  - "Horizontal cylinder: cn.SetOrientation(90.0)"
  - "Emoji in labels: <font size=\"+4\">&#x1F464;</font>"

requirements-completed:
  - ICON-01
  - ICON-02
  - ICON-03
  - ICON-04
  - DB-01
  - DB-02
  - DB-03
  - QUEUE-01
  - QUEUE-02
  - QUEUE-03
  - PERSON-01
  - PERSON-02
  - PERSON-03
  - PERSON-04
  - LABEL-01
  - LABEL-02
  - LABEL-03
  - LABEL-04
  - WRAP-01
  - WRAP-02

# Metrics
duration: 9min
completed: 2026-03-24
---

# Phase 18 Plan 01: Simplified Shapes Summary

**Removed SVG icon system and replaced with native GraphViz cylinder shapes for DB/Queue units, emoji for Person labels, and single-column HTML labels for System/Box/Container/Component types.**

## Performance

- **Duration:** 9 min
- **Started:** 2026-03-24T07:46:54Z
- **Completed:** 2026-03-24T07:56:16Z
- **Tasks:** 3
- **Files modified:** 6

## Accomplishments
- Removed entire icon extraction system (icons package, IconExtractor, SVG/DOT injection)
- Added native GraphViz cylinder shapes for DB units
- Added horizontal cylinder shapes (90deg rotation) for Queue units
- Replaced SVG icons with emoji (&#x1F464;) for Person labels
- Simplified all non-Person labels to single-column 3-row layout
- Preserved word-wrap functionality with --label-ratio flag

## Task Commits

Each task was committed atomically:

1. **Task 1: Remove icon system files** - `a68d1a9` (chore)
2. **Task 2+3: Update converter, render, labels, and tests** - `b113911` (feat)

## Files Created/Modified
- `internal/render/converter.go` - Removed icon parameters, added cylinder shape assignment
- `internal/render/render.go` - Simplified render without icon extraction logic
- `internal/render/labels.go` - Emoji labels and single-column layout for all builders
- `internal/render/wrap.go` - Added labelMaxCharsNoIcon for full-width text calculation
- `internal/render/html_labels_internal_test.go` - Updated tests for new label format
- `internal/render/experiment_internal_test.go` - DELETED (obsolete icon experiments)

## Files Deleted
- `internal/render/icons/embed.go`
- `internal/render/icons/person.svg`
- `internal/render/icons/db.svg`
- `internal/render/icons/pipe.svg`
- `internal/render/icons/system.svg`
- `internal/render/icons/container.svg`
- `internal/render/icons/component.svg`
- `internal/render/icons/icons_test.go`
- `internal/render/icon_extractor.go`
- `internal/render/icon_extractor_test.go`
- `internal/render/icons_integration_test.go`
- `internal/render/svg_icons.go`
- `internal/render/dot_icons.go`

## Decisions Made
- Used GraphViz native cylinder shape instead of custom SVG rendering
- Queue cylinders rotated 90 degrees for horizontal appearance
- Person emoji uses HTML entity encoding for cross-platform compatibility
- Single-column labels use labelMaxCharsNoIcon for full width utilization

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None - all changes compiled and tests passed on first attempt.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Rendering pipeline simplified with no external icon dependencies
- All C4 unit types now render with appropriate shapes
- Word-wrap functionality preserved and working with --label-ratio flag

## Self-Check: PASSED

- All claimed files exist
- All deleted files confirmed removed
- All commit hashes verified
- Build passes
- All tests pass

---
*Phase: 18-simplified-shapes*
*Completed: 2026-03-24*
