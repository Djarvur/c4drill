# Phase 4: Rendering & Output - Context

**Gathered:** 2026-03-10
**Status:** Ready for planning

<domain>
## Phase Boundary

Generate DOT and SVG files from graph structures with correct file paths. This phase consumes the Graph types from Phase 3 and produces rendered output files. Navigation links (explore, back-links, breadcrumbs) are Phase 5 - this phase renders static diagrams only.

</domain>

<decisions>
## Implementation Decisions

### DOT Generation
- Use go-graphviz native API (not manual DOT string building)
- Build graphviz.Graph object using provided methods
- Render via graphviz.Render() with XDOT format for DOT output
- Keep our graph types (Graph, Node, Edge, Cluster) - renderer converts to go-graphviz
- Two separate render functions: RenderDOT(g *Graph) and RenderSVG(g *Graph)

### HTML Labels
- Use HTML tables (<TABLE><TR><TD>) for node labels
- Matches Phase 3 label design with aligned cells
- Edge labels: two-line format with [Technology] on first line, Description on second

### Cluster Styling
- Expanded units rendered as bordered boxes with labels
- Standard GraphViz cluster style: dashed border, rounded corners, label at top
- Matches C4-PlantUML container visualization

### Output File Structure
- C1 (Context): `{basename}.{format}` (flat file in output directory)
- C2/C3 (expanded): `{basename}/{unit-path}.{format}` (directory hierarchy)
- Nested units: dotted path becomes directory hierarchy
  - Example: `mainapp.api` → `{basename}/mainapp/api.svg`
- Directory structure created recursively as needed

### Error Handling
- Fail fast on file write errors (stop immediately, report error)
- Leave partial output for user diagnostics
- Directory creation failures: report and exit with clear message
- No cleanup on failure (preserves diagnostic information)

### Claude's Discretion
- Exact go-graphviz API usage patterns
- Font specifications in DOT
- Exact spacing and padding in HTML labels
- Legend positioning and format

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/graph/graph.go`: Graph, Node, Edge, Cluster types with all styling info
- `internal/graph/builder.go`: BuildGraph function that creates graph from view
- `internal/graph/shapes.go`: ShapeForType, IconForType, GetStyleForType functions
- `internal/model/colors.go`: C4 color constants (need mapping to go-graphviz)
- `internal/view/scope.go`: GenerateC1View, GenerateC2View, GenerateC3View

### Established Patterns
- Two-phase construction: view then graph
- Map-based traversal for nested structures
- Separate packages for each layer (model, parser, validator, view, graph, render)

### Integration Points
- Renderer receives *graph.Graph from Phase 3
- Output writer handles file creation and directory structure
- CLI integration in Phase 6 will add --format and --output flags

### Dependencies
- go-graphviz: https://github.com/goccy/go-graphviz
- API docs: https://pkg.go.dev/github.com/goccy/go-graphviz

</code_context>

<specifics>
## Specific Ideas

- Use go-graphviz's native graph construction rather than string building
- HTML table labels for proper cell alignment
- Directory hierarchy mirrors C4 architecture hierarchy

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 04-rendering-output*
*Context gathered: 2026-03-10*
