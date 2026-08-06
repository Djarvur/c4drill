---
gsd_state_version: 1.0
milestone: v1.7
milestone_name: milestone
current_plan: 2 of 3
status: executing
last_updated: "2026-08-06T09:14:01.364Z"
last_activity: 2026-08-06
progress:
  total_phases: 3
  completed_phases: 0
  total_plans: 3
  completed_plans: 1
  percent: 0
---

# v1.8: Proper C1/C2/C3 View Generation

## Current Position

**Phase:** 1 (01-fix-c1-view-scoping)
**Status:** Ready to execute
**Current Plan:** 2 of 3
**Total Plans in Phase:** 3
**Last Activity:** 2026-08-06
**Progress:** [███░░░░░░░] 33%
**Last session:** 2026-08-06T09:14:01.331Z
**Stopped At:** None
**Resume File:** None

## Status: **COMPLETE**

## Summary

Fixed the two root-cause bugs:

1. **C1 pollution**: `addExternalBoundaryNodesRecursive` scanned ALL nested links → replaced with `addC1BoundaryNodes` that resolves peers to top-level ancestors
2. **No C2/C3**: `collectExpandedUnitPaths` only checked per-unit `Expanded` → replaced with `collectExpandableUnitPaths` that auto-detects units with subunits

## Evidence

| Metric | Before | After |
|--------|--------|-------|
| C1 nodes (saira TOML) | 105 | 5 |
| Sub-diagrams generated | 0 | 17 |
| All tests passing | ✅ | ✅ |
| `--expanded` flag | ✅ | ✅ |

## Commits

- `f9dc69a` fix: proper C1/C2/C3 view generation with edge resolution
- `c5091f9` chore: update baseline SVG

## Performance Metrics

| Phase | Plan | Duration | Notes |
|-------|------|----------|-------|
| Phase 01-fix-c1-view-scoping P01 | 22min | 3 tasks | 5 files |

## Decisions

- [Phase 01]: Pair-only edge dedup key (source->target) in resolved C1/C2/C3 views; --expanded keeps the v1.7 tech+desc key via new View.AllExpanded flag (D-01/D-02)
- [Phase 01]: Penwidth carried on graph.Edge.PenWidth (0 = renderer default); converter renders PenWidth>0 as-is else 1.0 — collapsed pairs 2.0, single edges 1.0 (D-04)
- [Phase 01]: countPairMultiplicity is mirror-aware: validator LinksFrom mirrors are not double-counted (D-05)
- [Phase 01]: DOT render assertions must extract per-edge attribute blocks — go-graphviz always emits a penwidth=1.0 default edge block
