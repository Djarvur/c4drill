---
gsd_state_version: 1.0
milestone: v1.7
milestone_name: milestone
current_plan: 1
status: ready_to_plan
last_updated: 2026-08-06T11:03:26.331Z
last_activity: 2026-08-06
progress:
  total_phases: 3
  completed_phases: 2
  total_plans: 4
  completed_plans: 4
  percent: 67
stopped_at: Phase 02 complete (1/1) — ready to discuss Phase 3
---

# v1.8: Proper C1/C2/C3 View Generation

## Current Position

Phase: 02 (auto-generate-c2-c3-diagrams) — EXECUTING
Plan: 1 of 1
**Phase:** 3
**Status:** Ready to plan
**Current Plan:** Not started
**Total Plans in Phase:** 1
**Last Activity:** 2026-08-06
**Progress:** [██████████] 100%
**Last session:** 2026-08-06T10:50:54.247Z
**Stopped At:** Completed 02-01-PLAN.md
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
| Phase 01-fix-c1-view-scoping P03 | 6min | 3 tasks | 5 files |
| Phase 02-auto-generate-c2-c3-diagrams P01 | 26min | 3 tasks | 4 files |

## Decisions

- [Phase 01]: Pair-only edge dedup key (source->target) in resolved C1/C2/C3 views; --expanded keeps the v1.7 tech+desc key via new View.AllExpanded flag (D-01/D-02)
- [Phase 01]: Penwidth carried on graph.Edge.PenWidth (0 = renderer default); converter renders PenWidth>0 as-is else 1.0 — collapsed pairs 2.0, single edges 1.0 (D-04)
- [Phase 01]: countPairMultiplicity is mirror-aware: validator LinksFrom mirrors are not double-counted (D-05)
- [Phase 01]: DOT render assertions must extract per-edge attribute blocks — go-graphviz always emits a penwidth=1.0 default edge block
- [Phase 01-fix-c1-view-scoping]: D-12 implemented by deleting the legacy recursive boundary path — validator is the single gatekeeper for undefined peers
- [Phase 01-fix-c1-view-scoping]: D-13 minlen gating at all 6 resolved-link synthesis sites: synthesized links copy Length only when both drawn endpoints are the link's original units; resolved edges carry no minlen
- [Phase 01-fix-c1-view-scoping]: D-02 activated — GenerateExpandedView sets View.AllExpanded=true restoring v1.7 dedup key and 2.0 penwidth in expanded mode (COMPAT-02)
- [Phase 01-fix-c1-view-scoping]: Visible subunits (direct children of expanded top-level C1 units) are added to v.Units + UnitOrder and marked in v.VisiblePaths; BuildGraph skips them as top-level nodes because buildCluster already renders them inside the parent cluster (Pitfall 5)
- [Phase 01-fix-c1-view-scoping]: resolveAndAddBoundary sources from resolveToViewAncestor(v, path) — the deepest VISIBLE ancestor of the source path; append condition resolved != resolvedSource keeps within-cluster edges (D-10) and suppresses parent edges (D-08)
- [Phase 01-fix-c1-view-scoping]: D-13 gate first operand updated to path == resolvedSource: a visible subunit's own links stay length-eligible when the peer also resolves unchanged
- [Phase 01-fix-c1-view-scoping]: resolveToTopLevel unified onto resolveToViewAncestor with the peer-as-is fallback for truly-external units (boundary-node contract preserved)
- [Phase 02-auto-generate-c2-c3-diagrams]: D-07 guard lives in the graph layer (BuildGraph C1 branch), not in isExpandedInC1 — scope.go stays read-only (Phase 1 WR-01 constraint)
- [Phase 02-auto-generate-c2-c3-diagrams]: Box pipeline fixture must include inter-unit links — validator ValidateOrphanUnits rejects link-less leaf units
