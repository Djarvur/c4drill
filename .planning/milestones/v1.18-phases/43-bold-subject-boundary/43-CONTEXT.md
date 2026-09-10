# Phase 43: Bold Subject Boundary - Context

**Gathered:** 2026-09-07
**Status:** Ready for planning

<domain>
## Phase Boundary

On every non-expanded (drill-down) view, the boundary group of the element the diagram depicts renders with a bold triple-width border (`penwidth=3`), so the viewer can tell at a glance which element the scheme belongs to. Emphasis is semantic (survives `--plain`/`--no-styles`); nothing else changes — expanded views, collapsed C1 root, other clusters, nodes, edges, legend.

</domain>

<decisions>
## Implementation Decisions

### Emphasis mechanism
- **D-01:** Triple width = `penwidth=3.0` graphviz cluster attribute. Applied ONLY to the subject boundary cluster created by `buildBoundaryCluster` (internal/graph/builder.go:373) — the single construction site whose comment states "the boundary frame IS the unit on its own child diagram". This cluster is exactly the "group the scheme depicts" on C2, C3, and deep-link drill-down views.
- **D-02:** `graph.NodeStyle` (graph.go:195) gains a `BorderWidth float64` field — 0 means "renderer default" (no attribute emitted), >0 emits `penwidth`. `applyClusterStyle` (converter.go:656) emits it via `SafeSet("penwidth", ...)`. Edge precedent: `e.SetPenWidth(edge.PenWidth)` with 0→1.0 fallback at converter.go:899-905 — cluster side mirrors that contract.

### Semantics (not author formatting)
- **D-03:** The bold boundary is a semantic navigation aid, NOT author formatting: it survives `--plain` and `--no-styles` — the same rationale as kind-derived edge colours and the legend (PLAIN-02 explicitly keeps those). No new CLI flag; no model key. `--no-colors` does not interact (penwidth is not colour).

### Scope guard
- **D-04:** Everything else stays byte-identical: `--expanded` copies (no single subject), collapsed C1 root (no boundary clusters), non-subject clusters, node shapes, legend. Locked by byte-equality assertions on those render paths, not by hope.

### Verification contract
- **D-05:** TDD RED→GREEN: raw-DOT assertion that the subject boundary cluster carries `penwidth=3` and NO other cluster on the same diagram does. Multi-level fixture (C2 + deep-link C3) proves the emphasis follows the drilled target. Goldens: re-baseline only where a subject boundary exists; the per-golden diff must contain nothing beyond the `penwidth` attribute line (canonicalDOT keeps attributes — attribute-set changes ARE visible; geometry stays stripped).

### Claude's Discretion
- Exact fixture to base the multi-level assertion on (existing goldens carriers preferred).
- Whether `BorderWidth` belongs on `NodeStyle` vs a dedicated cluster field — planner's call, D-02 is the recommendation.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Code
- `internal/graph/builder.go` §buildBoundaryCluster (line ~373) — the single subject-boundary construction site (style from `GetStyleForType` + `applyUnitOverrides`)
- `internal/graph/graph.go` §NodeStyle (line ~195) — gains the BorderWidth field
- `internal/render/converter.go` §applyClusterStyle (line ~656) + §applyEdgeAttributes penwidth precedent (lines 899-905) — attribute emission contract
- `internal/view/view.go` — View flags (Plain/NoStyles semantics, PLAIN-02/KEY-01)

### Project docs
- `.planning/REQUIREMENTS.md` — BOLD-01..03 definitions
- `.planning/PROJECT.md` — "As of v1.15/1.16/1.17" paragraphs (formatting-key family, PLAIN/KEY semantics)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `SafeSet` / `setClusterAttribute` — attribute emission path already shared by cluster styles
- Existing golden suite (canonicalDOT, order-insensitive) — the no-collateral-change gate
- `GetStyleForType` + `applyUnitOverrides` — boundary cluster style pipeline the new field rides

### Established Patterns
- 0-means-default numeric fields (Edge.PenWidth) — D-02 mirrors it
- Semantic-survives-plain precedent (kind colours, legend; PLAIN-02)
- TDD RED→GREEN with raw-DOT attribute assertions (GEDGE-07 switch-matrix precedent)

### Integration Points
- `buildBoundaryCluster` → `NodeStyle` → `applyClusterStyle` → SVG/DOT emission
- Golden re-baselining for drill-down views only

</code_context>

<specifics>
## Specific Ideas

From the user: "если можно — даже тройной толщины" — triple width explicitly requested and feasible (`penwidth=3`).

</specifics>

<deferred>
## Deferred Ideas

- A future-configurable border-width flag (`--border-width <N>`) — this phase hard-codes 3 per request; noted, not built.

</deferred>

---

*Phase: 43-Bold Subject Boundary*
*Context gathered: 2026-09-07*