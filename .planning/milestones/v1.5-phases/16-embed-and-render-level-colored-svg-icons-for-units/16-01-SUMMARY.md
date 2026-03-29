---
phase: 16-embed-and-render-level-colored-svg-icons-for-units
plan: 01
subsystem: render
tags: [svg, icons, embed, html-labels, graphviz]

# Dependency graph
requires:
  - phase: 15-the-edge-must-be-the-same-color-as-the-source-unit
    provides: Edge color from source unit border
provides:
  - Embedded SVG icons with dynamic color replacement
  - IconExtractor for on-demand icon file generation
  - IMG tags in HTML labels replacing Unicode emojis
  - RenderSVGWithOutput function for icon-enabled rendering
affects: [rendering, cli, output]

# Tech tracking
tech-stack:
  added: [Go embed.FS for SVG icon embedding]
  patterns: [on-demand file extraction, memory + disk caching, graceful degradation]

key-files:
  created:
    - internal/render/icons/embed.go
    - internal/render/icons/person.svg
    - internal/render/icons/db.svg
    - internal/render/icons/pipe.svg
    - internal/render/icons/system.svg
    - internal/render/icons/container.svg
    - internal/render/icons/component.svg
    - internal/render/icons/icons_test.go
    - internal/render/icon_extractor.go
    - internal/render/icon_extractor_test.go
  modified:
    - internal/render/labels.go
    - internal/render/converter.go
    - internal/render/render.go
    - internal/render/html_labels_test.go
    - cmd/c4drill/root.go
    - internal/output/writer.go

key-decisions:
  - "D-01: Use Go embed.FS for SVG icon storage"
  - "D-02: Icons extracted on-demand to {output}/.icons/"
  - "D-03: Icon naming: type-{hexcolor}.svg (e.g., person-3C7FC0.svg)"
  - "D-04: IMG tags at 32x32 pixels in HTML labels"
  - "D-05: Box type uses container icon"
  - "D-06: Empty iconRelPath triggers graceful degradation (no icon column)"

patterns-established:
  - "On-demand extraction: icons created only for types/colors present in diagram"
  - "Dual caching: memory cache for single run, file existence check across runs"
  - "Backward compatibility: empty outputDir = no icons, rendering still works"

requirements-completed: [ICON-01, ICON-02, ICON-03, ICON-04, ICON-05, ICON-06]

# Metrics
duration: 13min
completed: 2026-03-20
---

# Phase 16 Plan 01: SVG Icon Embedding and Rendering Summary

**Embedded SVG icons with dynamic color matching for C4 diagram units, replacing Unicode emojis with 32x32 pixel IMG tags extracted on-demand**

## Performance

- **Duration:** 13 min
- **Started:** 2026-03-20T20:57:13Z
- **Completed:** 2026-03-20T21:10:36Z
- **Tasks:** 5
- **Files modified:** 12

## Accomplishments
- Created icons package with Go embed.FS for 6 SVG templates (person, db, pipe, system, container, component)
- Implemented IconExtractor with dual caching (memory + disk) for on-demand icon extraction
- Updated all 6 HTML label builders to use IMG tags instead of Unicode emojis/monospace text
- Integrated icon extraction with converter pipeline (nodes and clusters)
- Connected CLI to pass output directory for SVG icon extraction

## Task Commits

Each task was committed atomically:

1. **Task 1: Create icons package with embedded SVGs** - `6216449` (feat)
2. **Task 2: Create IconExtractor for on-demand icon extraction** - `9be18b4` (feat)
3. **Task 3: Update HTML label builders to use IMG tags** - `39f7069` (feat)
4. **Task 4: Integrate icon extraction with converter pipeline** - `39e525e` (feat)
5. **Task 5: Update CLI to pass output directory for icon extraction** - `e8e01a9` (feat)

**Test updates:** `230561e` (test)

## Files Created/Modified
- `internal/render/icons/embed.go` - Package with GetTemplate() and Colorize() functions
- `internal/render/icons/*.svg` - 6 SVG icon templates with currentColor placeholder
- `internal/render/icons/icons_test.go` - Tests for template access and colorization
- `internal/render/icon_extractor.go` - IconExtractor with dual caching
- `internal/render/icon_extractor_test.go` - Tests for extraction behavior
- `internal/render/labels.go` - Updated HTML label builders with IMG tags and iconTypeForUnit helper
- `internal/render/converter.go` - Integrated icon extraction in createNode and createCluster
- `internal/render/render.go` - Added RenderSVGWithOutput function
- `internal/render/html_labels_test.go` - Updated tests for new function signatures
- `cmd/c4drill/root.go` - Use RenderSVGWithOutput for SVG format
- `internal/output/writer.go` - Added BaseDir() method

## Decisions Made
- Use embed.FS instead of external files for icon templates
- Extract icons to {output}/.icons/ subdirectory for clean separation
- Filename format type-{hexcolor}.svg enables unique icons per color variant
- 32x32 pixel IMG tags provide consistent icon sizing
- Empty iconRelPath gracefully degrades to skip icon column

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- Pre-existing test failures (TestOutputFlag, TestExpandedViewNestedClusters, etc.) documented in deferred-items.md - out of scope for this plan

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Icon system fully functional and integrated
- Icons extracted on-demand during SVG rendering
- All icon-related tests pass
- Ready for verification and next phase

## Self-Check: PASSED

All key files exist and all task commits found.

---
*Phase: 16-embed-and-render-level-colored-svg-icons-for-units*
*Completed: 2026-03-20*
