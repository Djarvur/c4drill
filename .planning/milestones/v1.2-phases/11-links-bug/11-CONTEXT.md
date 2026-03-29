# Phase 11: Unit Shape and Attributes - Context

**Gathered:** 2026-03-13
**Status:** Ready for planning

<domain>
## Phase Boundary

Fix unit rendering shapes and fill colors:
- Shape: `record` for **collapsed** units (regardless of whether they have subunits)
- Shape: `cluster` for **expanded** units (showing subunits as nested nodes)
- Fill: Transparent (no background colors for any unit)
- Keep icons and styling differentiation by type and level

**Key distinction:** Shape is determined by VIEW state (expanded/collapsed), NOT by whether subunits exist in the model. A unit with subunits that is collapsed must render as `record`, not `cluster`.

**Current issue:** 
- `ShapeForType()` always returns `ShapeHTML`
- NodeStyle fills use hardcoded background colors
- Shape logic doesn't consider view expansion state

</domain>

<decisions>
## Implementation Decisions

### Shape Logic (View-Based)
- **Collapsed units** → `ShapeRecord` (record shape)
- **Expanded units** → `ShapeCluster` (cluster/subgraph containing nested nodes)
- Shape depends on expansion state in the current view, NOT on whether subunits exist
- Unit with subunits but collapsed = `record` shape
- Unit expanded (showing children) = `cluster` shape

### Fill Logic
- Remove all fill colors from NodeStyle
- All units render with transparent background
- Colors defined in model/colors.go remain for font colors only

### Claude's Discretion
- Exact wording for error messages if needed
- How to handle width/height attributes with transparent fills

</decisions>

<specifics>
## Specific Ideas

- "the unit record or subcluster choice depend on is it expanded on the current picture, not by subunits existence"
- "the unit with subunits but collapsed must be shaped as record"
- "expanded unit must be displayed as subgraph"
- "no fill for any thing" (transparent backgrounds)

</specifics>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/graph/shapes.go`: Contains `ShapeForType()` function (needs modification to consider expansion state)
- `internal/graph/graph.go`: Defines `ShapeRecord`, `ShapeCluster`, `ShapeHTML` constants
- `internal/model/colors.go`: Contains all background color constants (stop using for fills)
- `internal/view/scope.go`: Handles view expansion state (`IsExpanded` checks)

### Established Patterns
- Icons are differentiated by type (person, db, queue) in `IconForType()`
- Styles are differentiated by level and external status in `GetStyleForType()`
- View expansion is tracked per unit in view generation

### Integration Points
- `internal/graph/builder.go:142` - calls `ShapeForType()` when creating nodes (needs expansion state)
- `internal/render/converter.go`: Applies shape via `cn.SetShape()`
- NodeStyle in `internal/graph/shapes.go`: FillColor used by converter (needs to be transparent)

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---
*Phase: 11-unit-shape-attributes*
*Context gathered: 2026-03-13*
