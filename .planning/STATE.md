---
gsd_state_version: 1.0
milestone: v1.7
milestone_name: milestone
current_plan: 3
status: executing
last_updated: "2026-08-06T09:25:04.419Z"
last_activity: 2026-08-06
progress:
  total_phases: 3
  completed_phases: 0
  total_plans: 3
  completed_plans: 2
  percent: 0
---

# v1.8: Proper C1/C2/C3 View Generation

## Current Position

**Phase:** 1 (01-fix-c1-view-scoping)
**Status:** Ready to execute
**Current Plan:** 3
**Total Plans in Phase:** 3
**Last Activity:** 2026-08-06
**Progress:** [███████░░░] 67%
**Last session:** 2026-08-06T09:23:36.311Z
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
| Phase 01-fix-c1-view-scoping P01 | 10min | 3 tasks | 5 files |
| Phase 01-fix-c1-view-scoping P02 | 5min | 3 tasks | 3 files |

## Decisions

- [Phase 01]: Pair-only edge dedup key (source->target) in resolved C1/C2/C3 views; --expanded keeps the v1.7 tech+desc key via new View.AllExpanded flag (D-01/D-02)
- [Phase 01]: Penwidth carried on graph.Edge.PenWidth (0 = renderer default); converter renders PenWidth>0 as-is else 1.0 — collapsed pairs 2.0, single edges 1.0 (D-04)
- [Phase 01]: countPairMultiplicity is mirror-aware: validator LinksFrom mirrors are not double-counted (D-05)
- [Phase 01]: DOT render assertions must extract per-edge attribute blocks — go-graphviz always emits a penwidth=1.0 default edge block
- [Phase 01-fix-c1-view-scoping]: D-12 implemented by deleting the legacy recursive boundary path — validator is the single gatekeeper for undefined peers
- [Phase 01-fix-c1-view-scoping]: D-13 minlen gating at all 6 resolved-link synthesis sites: synthesized links copy Length only when both drawn endpoints are the link's original units; resolved edges carry no minlen
- [Phase 01-fix-c1-view-scoping]: D-02 activated — GenerateExpandedView sets View.AllExpanded=true restoring v1.7 dedup key and 2.0 penwidth in expanded mode (COMPAT-02)
