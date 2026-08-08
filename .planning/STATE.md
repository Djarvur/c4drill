---
gsd_state_version: 1.0
milestone: v1.9
milestone_name: C3 Boundary Node Fix
status: Awaiting next milestone
last_updated: "2026-08-06T18:17:46.366Z"
last_activity: 2026-08-06 — Milestone v1.9 completed and archived
progress:
  total_phases: 1
  completed_phases: 1
  total_plans: 1
  completed_plans: 1
  percent: 100
---

# v1.9: C3 Boundary Node Fix

## Current Position

Phase: Milestone v1.9 complete
Plan: —
Status: Awaiting next milestone
Last activity: 2026-08-06 — Milestone v1.9 completed and archived

## Status: **COMPLETE**

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
| Phase 04-c3-boundary-node-fix P01 | 3min | 3 tasks | 2 files |

## Decisions

- [Phase 04]: D-01 realized via v.ExpandedUnit (NOT scopePath): addResolvedBoundaryNode bounds its peer walk-up at the expanded unit's parent so C3 cross-container links resolve to sibling containers (e.g. mainSystem.rbac) not the parent system (mainSystem). The plan's scopePath premise only holds at the top-level call — addExternalBoundaryNodesForSubunits recurses and rebinds scopePath to each link-host's immediate parent, so a scopePath-derived bound fired at the wrong level and the RED test stayed failing. v.ExpandedUnit is the stable view root. Signature unchanged; C2 (top-level ExpandedUnit -> "") and C1 (not called) behavior preserved (BOUND-01/02/03). Commits d3a56ff (RED), 710440b (GREEN).
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
- [Post-v1.8]: Safari/WebKit silently ignores <a> element navigation inside SVG (both xlink:href and plain href, standalone and inline). Chromium honors them. This is a WebKit limitation, not an href-value bug. Fix: new `-f html` format inlines the SVG in an HTML document and injects a JS shim (addEventListener click -> window.location.href) that restores navigation in Safari. SVG/DOT formats unchanged.
- [Post-v1.8]: GraphViz HTML-label quirk — when a multi-row TABLE label has row 1 with <FONT POINT-SIZE="10"> content and row 2 with plain (default-size) text, GraphViz SILENTLY DROPS the title row from the SVG render. Fix: wrap the title in explicit <FONT POINT-SIZE="14"> so both rows carry FONT tags.
- [Post-v1.8]: GraphViz HTML-label column-stretching — separate separator cells (<TD>&gt;</TD> between breadcrumb items) stretch to the column width (sized to the widest cell in that column across all rows), creating large visual gaps. Fix: merge the ">" separator INTO each item's cell as trailing inline text (<TD HREF=url>Name &gt;</TD>).
- [Post-v1.8]: Navigation redesigned — back-link dropped (it duplicated the breadcrumb's nearest-ancestor entry: same destination, same label). Breadcrumb-only now. C3 root context breadcrumb added (was missing — C3 couldn't navigate to C1 directly). Pretty unit Names resolved via new View.AncestorNames map (populated by view generators which have model access; graph builder consumes).
- [Post-v1.8]: canonicalDOT golden (multilevel.expanded.dot) regenerated AGAIN — nav redesign changed the graph label format (ALIGN="CENTER" on TABLE, FONT POINT-SIZE="14" wrapper on title). Benign format change, semantic content unaffected.

## Accumulated Context

### Pending Todos

- **Add 'reference' field to C4 units with 📖 marker** (`tooling`) — `.planning/todos/pending/2026-08-08-add-reference-field-to-units.md`. Optional per-unit URL field (LikeC4 `link` port; only LikeC4 feature adopted). Units with non-empty `reference` render a 📖 symbol. Rejected: custom kinds, tags, icons, metadata, deployment model, user-authored views.
- **TOML authoring ergonomic improvements** (`tooling`) — `.planning/todos/pending/2026-08-08-toml-authoring-ergonomic-improvements.md`. Three independent fixes (relative `peer` resolution, optional `name`, compact `link`) as the TOML-native alternative to a rejected custom DSL. Sequenced: relative peer first, optional name second, compact link deferred.

## Operator Next Steps

- Start the next milestone with /gsd-new-milestone
- [Post-v1.9]: KNOWN LIMITATION — boundary nodes in C3 diagrams are drawn inside the cluster box visually. Root cause: go-graphviz's WASM cgraph engine calls agsubnode when creating edges between root-level nodes and cluster-internal nodes, which reassigns the root node to the cluster subgraph. This is standard cgraph behavior (edges require both endpoints in the same subgraph context). compound=true does not prevent it. The boundary nodes ARE top-level in the graph structure (verified via debug), the edges DO render (verified: all 16 sub-diagrams have ≥3 edges), but GraphViz draws the cluster bounding box around them. Fix would require either: (a) forking the go-graphviz WASM to avoid agsubnode on edge creation, (b) using lhead/ltail to draw edges to cluster boundaries, or (c) replacing the cluster subgraph with a manually-drawn boundary box. Deferred — cosmetic, edges and nodes are correct.
- [Post-v1.9]: Boundary resolution v1.9 fix (expandedParent guard) was incomplete — only caught direct siblings, not cross-subtree peers. Fixed with path-prefix divergence algorithm: find common prefix between peer and v.ExpandedUnit, boundary = peer's ancestor one level below divergence. Verified: all 16 sub-diagrams clean.
- [Post-v1.9]: IsExternal overloading bug — setting IsExternal=true on all boundary nodes fixed placement (top-level) but broke styling (regular containers rendered gray instead of blue). Fix: added Entry.IsBoundary (view-scope signal for placement) separate from IsExternal (unit-type signal for styling). Graph builder checks IsBoundary || IsExternal for placement; GetStyleForType still uses IsExternal for colors.
- [Post-v1.9]: Boundary nodes placed inside cluster visually (KNOWN LIMITATION, see above). Root cause confirmed: go-graphviz WASM cgraph calls agsubnode on edge creation between root and cluster nodes, reassigning the root node to the cluster subgraph. compound=true does not help. Functional behavior (edges, node structure) is correct; only the cluster box bounding is wrong.
