# Phase 15: Edge Coloring - Context

**Gathered:** 2026-03-20
**Status:** Ready for planning

<domain>
## Phase Boundary

Apply visual styling to diagram edges (arrows) so their color matches the source unit's color scheme. This is a rendering enhancement - no changes to TOML schema or parsing required.

</domain>

<decisions>
## Implementation Decisions

### Color Source
- **D-01:** Edge color comes from the source unit's border color
- Border colors are defined in `internal/model/colors.go` per C4 level:
  - C1 (system): `#3C7FC0`
  - C2 (container): `#3C7FC0`
  - C3 (component): `#78A8D8`
  - External variants have different colors (gray tones)

### Label Coloring
- **D-02:** Edge labels (technology, description) match the edge color
- Both technology bracket text and description text use the same color as the edge line

### Explicit Color Override
- **D-03:** If `link.color` is explicitly set in TOML, it overrides the source border color
- Fallback chain: explicit color → source unit border color

### Claude's Discretion
- Exact implementation approach (where to add Color field to Edge struct)
- How to look up source unit's border color during edge creation
- Test coverage strategy

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Color definitions
- `internal/model/colors.go` — C4 level border colors (SystemBorder, ContainerBorder, ComponentBorder, external variants)

### Edge creation and rendering
- `internal/graph/builder.go` — `createEdge()` function creates edges (add color here)
- `internal/graph/graph.go` — `Edge` struct definition (add Color field here)
- `internal/render/converter.go` — `createEdge()` renders edges to DOT (apply color here)
- `internal/graph/shapes.go` — `GetStyleForType()` returns border colors for unit types

### Link model
- `internal/model/link.go` — `Link` struct with existing `Color` field

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `GetStyleForType(t, isExternal)` in `shapes.go` — returns `NodeStyle` with `BorderColor`
- Level detection via `LevelForType(t)` — maps unit type to C1/C2/C3 level
- `model.Link.Color` field already exists for explicit color override

### Established Patterns
- Node colors use border color from C4-PlantUML palette
- Transparent backgrounds (Phase 11 decision)
- Font color matches border color (Phase 13 decision for clusters)

### Integration Points
- `buildEdges(v)` in `builder.go` — creates all edges from view links
- `createEdge(source, target, link)` — where edge struct is built
- `createEdge(cg, sourceNode, targetNode, edge)` in converter — where edge is rendered to cgraph

</code_context>

<specifics>
## Specific Ideas

- Edge color should visually indicate the "owner" of the relationship (the source)
- Creates visual flow from source to target
- Consistent with cluster label coloring approach from Phase 13

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 15-the-edge-must-be-the-same-color-as-the-source-unit*
*Context gathered: 2026-03-20*
