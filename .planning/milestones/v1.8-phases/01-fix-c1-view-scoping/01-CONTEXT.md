# Phase 1: Fix C1 View Scoping - Context

**Gathered:** 2026-08-06
**Status:** Ready for planning

<domain>
## Phase Boundary

Refine the existing (already-implemented, tested) C1 view generation so the diagram shows only top-level units with edges resolved to nearest visible ancestors and duplicate edges collapsed. The phase covers the edge-resolution and deduplication behavior of the C1 view and its interplay with expanded top-level units (VIEW-05), plus cleanup of dead boundary-node code. Requirements VIEW-01, VIEW-02, EDGE-01, EDGE-02.

The v1.8 implementation already exists (commits `f9dc69a`, `c5091f9`, `8f7e21c`, `b6771b0`; STATE.md declares complete; tests pass). This phase refines behavior per the decisions below, NOT a from-scratch re-implementation.

</domain>

<decisions>
## Implementation Decisions

### Edge Collapse & Deduplication (EDGE-02)
- **D-01:** Pair-only collapse — one edge per (source, target) pair regardless of technology/description. The first contributing link in definition order wins.
- **D-02:** Pair-only collapse applies to resolved views (C1/C2/C3), where edge resolution happens. **`--expanded` mode is exempt** — it keeps the v1.7 tech+desc dedup key so COMPAT-02 output stays identical to v1.7.
- **D-03:** The surviving collapsed edge inherits ALL attributes first-wins (color, style, arrow, length, label position) from the first link in definition order.
- **D-04:** Collapsed edges (2+ relationships collapsed onto the pair) get a binary thicker penwidth (2.0). Single-relationship edges stay at default width.
- **D-05:** Multiplicity counts ALL contributing links (direct links + resolved sub-links) that land on the pair.
- **D-06:** Collapsed edge labels stay plain — first-wins tech/description, no count suffix, no aggregation. Thicker penwidth is the only multiplicity signal.

### Edge Resolution Semantics (VIEW-01/02, EDGE-01)
- **D-07:** Deepest visible ancestor for targets — when a link target's parent is expanded in C1 (target visible as a node inside the cluster), the edge points at the visible subunit directly. Hidden targets resolve to the top-level ancestor as today. Aligns C1 with C2/C3's existing `resolveToViewAncestor` semantics.
- **D-08:** Replace parent edge — when a link resolves to a visible subunit, the parent-level edge is NOT drawn. The parent edge appears only when a real direct link to the parent exists. No redundant edges.
- **D-09:** Source-side resolution — the source of a link also resolves to its deepest visible ancestor (direct subunit node when the source is inside an expanded top-level unit; the top-level unit otherwise). E.g. `linuxSystem.sshAuth → keycloak` renders as `sshAuth → keycloak`.
- **D-10:** Draw within cluster — links whose source AND target are both visible inside the same expanded top-level unit (e.g. `sshAuth → webAPI`) render as an edge between the two nodes inside the cluster (currently skipped as "internal").
- **D-11:** Box grouping units follow the SAME resolution rules as systems — no special-casing (collapsed box: edges resolve to the box; expanded box: edges resolve to visible grouped units; source-side resolution applies).

### Boundary Nodes & Dead Code
- **D-12:** Skip unresolved links in the view layer — the validator is the single gatekeeper for undefined targets (`internal/validator/rules.go` rejects them). Remove the legacy `addExternalBoundaryNodes` + `addExternalBoundaryNodesRecursive` (`internal/view/scope.go`), which are unreachable/no-op for valid models (still called from `GenerateExpandedView` but can never fire since all units are in the view).

### minlen Semantics
- **D-13:** A link's `length` (GraphViz minlen) takes effect ONLY when both drawn endpoints ARE the link's original units (no resolution on either side). Resolved edges and collapsed edges carry no minlen. For collapsed edges, the surviving first link's minlen applies only if its own endpoints were both original. (User-specified area: "minlen must be effective if both nodes the edge connects are visible only".)

### Claude's Discretion
None — every area was decided with concrete options.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements & Scope
- `.planning/REQUIREMENTS.md` — VIEW-01, VIEW-02, EDGE-01, EDGE-02 (phase 1), COMPAT-02 (constraint on `--expanded` exemption)
- `.planning/ROADMAP.md` — Phase 1 goal, root-cause analysis, success criteria (5-node C1, resolved edges, collapsed duplicates)
- `.planning/PROJECT.md` — TOML schema (link object incl. `length`), rendering behavior, validation rules

### Implementation Files (behavior verified during discussion)
- `internal/view/scope.go` — `GenerateC1View`, `addC1BoundaryNodes`, `resolveAndAddBoundary`, `resolveToTopLevel` (to be replaced/unified with deepest-visible-ancestor), `resolveToViewAncestor` (existing C2/C3 helper to align with), `GenerateExpandedView` (legacy call site), `createExternalBoundaryNode`
- `internal/graph/builder.go` — `buildEdges`, `processOutgoingLinks`/`processIncomingLinks` (dedup key change), `markSeen`, `createEdge` (minlen gate + penwidth), `buildCluster` (one-level expansion rendering), `buildBoundaryCluster`, `BuildGraph`
- `cmd/c4drill/root.go` — `collectExpandedPaths`/`collectExpandableUnitPaths` (C2/C3 auto-detection, unchanged by this phase)
- `internal/validator/rules.go` — undefined-unit reference rejection (basis for D-12)
- `internal/view/scope_test.go`, `internal/graph/builder_test.go` — existing test patterns for new tests

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `Entry.ResolvedLinks`/`ResolvedLinksFrom` mechanism (`internal/view/scope.go`): resolved links collected on entries, consumed by `buildEdges` — the natural extension point for source-side resolution and pair-collapse counting
- `resolveToViewAncestor` (`internal/view/scope.go:716`): existing deepest-visible-ancestor helper — unify C1 resolution with it instead of duplicating
- `markSeen` + `seen` map (`internal/graph/builder.go:460`): dedup mechanism; key change to pair-only
- `createEdge` (`internal/graph/builder.go:423`): single edge-construction point — gate minlen (D-13) and set penwidth (D-04) here
- `buildCluster` (`internal/graph/builder.go:265`): renders expanded top-level unit as cluster with direct subunits as visible nodes — defines what "visible" means for C1

### Established Patterns
- Definition-order preservation everywhere (`UnitOrder`/`SubunitOrder` slices) — "first wins" is deterministic via these
- `ResolvedLinks != nil` overrides `Unit.Links` in `buildEdges` (`builder.go:339`)
- Nil-guard at public API boundaries; `//nolint:` directives always with explanation
- Edge color = source border color, `link.color` overrides (D-01/D-03 in earlier context)

### Integration Points
- `GenerateC1View` (`scope.go:88`) → `addC1BoundaryNodes` — resolution logic lives here
- `buildEdges` (`builder.go:329`) — dedup key + penwidth + minlen changes
- `GenerateExpandedView` (`scope.go:14`) — remove `addExternalBoundaryNodes` call
- C2/C3 paths (`resolveSubunitCrossLinks`, `resolveBoundaryNodeLinks`) — verify D-02/D-09 don't regress cross-subunit synthesis

</code_context>

<specifics>
## Specific Ideas

No specific external references — the user's one custom requirement (minlen effective only when both edge endpoints are the original visible units) is captured as D-13.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 1-Fix C1 View Scoping*
*Context gathered: 2026-08-06*
