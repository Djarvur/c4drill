---
gsd_state_version: 1.0
milestone: v1.7
milestone_name: milestone
current_plan: 2
status: executing
last_updated: "2026-08-06T13:29:40.050Z"
last_activity: 2026-08-06
progress:
  total_phases: 3
  completed_phases: 3
  total_plans: 7
  completed_plans: 7
  percent: 100
---

# v1.8: Proper C1/C2/C3 View Generation

## Current Position

Phase: 03 (compatibility-validation) — EXECUTING
Plan: 4 of 4 (03-04 gap-closure COMPLETE)
**Phase:** 03
**Status:** Executing Phase 03
**Current Plan:** 4 (complete)
**Total Plans in Phase:** 4 (3 original + 03-04 gap-closure)
**Last Activity:** 2026-08-06
**Progress:** [██████████] 100%
**Last session:** 2026-08-06T13:28:18Z
**Stopped At:** Completed 03-04-PLAN.md (C2/C3 navigation gap closure)
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
| Phase 03-compatibility-validation P01 | 5min | 2 tasks | 2 files |
| Phase 03-compatibility-validation P02 | 8min | 3 tasks | 3 files |
| Phase 03-compatibility-validation P04 | 19min | 3 tasks | 8 files |

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
- [Phase ?]: Committed both test-data assets despite render-pipeline order nondeterminism: semantic content (D-02 contract) stable across runs; only sibling ordering and layout geometry flip (pinned go-graphviz fork map-order). 03-02 MUST compare order-insensitively (sort-normalize per RESEARCH Pitfall 1), NOT byte-exact require.Equal
- [Phase ?]: No production code touched: scope guard clean; nondeterminism lives in the pinned external fork (onokonem/go-graphviz cgraph/gvc WASM layout), out of scope for test-data-only plan
- [Phase 03]: Golden comparison is NOT byte-exact require.Equal — replaced with order-insensitive canonicalDOT semantic comparison per DI-1 (5/5 pipeline runs differ byte-wise; sibling cluster/node order + layout geometry flip; semantic content stable)
- [Phase 03]: canonicalDOT strips layout geometry (bb/pos/lp/lheight/lwidth/height/width) and sorts statements/attributes recursively — validated: 2 independent runs + committed golden canonicalize identically; C1 outputs still differ
- [Phase 03]: DOT statement terminator is '];' not ';' — person-node HTML labels contain '&#x1F464;' entities whose semicolons truncate naive scans
- [Phase 03]: TestCompat01 does NOT assert absence of valid/app.dot — Phase 2 auto-detect generates the C2 sub-diagram
- [Phase 03-04]: GraphViz HTML-like labels do NOT support <a href> tags (empirically verified: labels containing them are silently dropped at render time). Clickable HTML labels use <TD HREF=url>, rendered as <a xlink:href> in SVG. Gap 2 implemented via this idiom, not the plan's <a href>+StrdupHTML premise.
- [Phase 03-04]: canonicalDOT golden (multilevel.expanded.dot) regenerated — the graph label is now always HTML-wrapped (label=<...>); benign format change, URLs/edges/clusters unaffected (expanded diagram carries no navigation).
- [Phase 03-04]: Pre-existing C3 breadcrumb bug fixed (Rule 2): computeBreadcrumbURL's index-0 special case assumed first segment = C1 root, so C3 'mainSystem' crumb 404'd on ../multilevel.svg; now delegates to ComputeExploreURL's bidirectional logic -> ../mainSystem.svg.
- [Phase 03-04]: All clickable navigation URLs force .svg regardless of render format (ComputeBackLinkURL/computeBreadcrumbURL ignore the format param, mirroring ComputeExploreURL) — signatures kept to avoid builder ripple.

### Blockers

- 03-02 golden comparison as designed (byte-exact require.Equal vs multilevel.expanded.dot) WILL fail: render pipeline output is order-nondeterministic (sibling cluster/node order + layout geometry flip per run; pinned go-graphviz fork map-order). Use order-insensitive comparison (sort-normalize / strip geometry / structural asserts). See deferred-items.md DI-1
