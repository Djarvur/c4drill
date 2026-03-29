# Phase 3: Views & Graphs - Context

**Gathered:** 2026-03-09
**Status:** Ready for planning

<domain>
## Phase Boundary

Transform validated model into scoped views (C1/C2/C3) and graph structures. This phase does NOT render to DOT/SVG (Phase 4) - it builds the in-memory view and graph structures that the renderer will consume.

</domain>

<decisions>
## Implementation Decisions

### View scoping rules
- C1 = all top-level units + links between them
- C2 = subunits of expanded container + links between subunits
- C3 = subunits of expanded container (C3 level)
- Each expanded unit gets its own view (not combined per-parent)
- Per-unit `expanded` attribute only (no global properties.expanded default)

### External boundary nodes
- Units referenced from in-scope units but outside the view appear as collapsed boundary nodes
- External boundary nodes: same size as regular nodes, dashed border, external palette colors
- Links to external boundary nodes are shown normally

### Link visibility
- Links shown if both endpoints are in view OR one endpoint is in view and other is external boundary
- Scoped to view: link appears only where both endpoints are visible, not in child views
- Bidirectional relationships (link + linkFrom between same units): separate edges, not combined
- Multiple links between same two units: show all separately
- Self-links (unit linking to itself): rejected by validator
- linkFrom renders same as link (just defined from target perspective)

### Link labels
- Technology and description in separate cells
- Technology appears in square brackets: `[HTTP]`
- Label position: default middle, override via link.labelPosition
- Arrow direction: read from link.arrow attribute
- Rank attribute: ignored for layout (GraphViz decides)
- Default link style: solid black line

### Edge routing style
- GraphViz only supports one edge style per diagram
- C1 view: use properties.edges
- Child views: use expanded unit's edges attribute
- Values: straight, spline, square

### Node shape design
- Person: HTML record with 👤 icon: `👤 | { name | description }`
- DB: HTML record with ⛁ icon: `⛁ | { name | description }`
- Queue: HTML record with ╟╢ bars: `╟\n╢ | { name | description }`
- System/Container/Component: HTML record: `{ name | technology | description }`
- Box: same as system shape
- All use HTML-like labels for proper cell formatting

### External variants styling
- Same shape as base type
- Different border and text color (external palette)
- Distinct external color palette (not level colors)

### Cluster rendering
- Expanded units rendered as subgraph clusters with label at top
- Cluster label shows full node label (name, technology, description)
- Children nodes rendered inside cluster

### Style defaults
- No inheritance: each unit's style is independent
- Colors: different levels have different colors from C4-PlantUML palette
- Same level units have same colors
- Border: defaults to level color (same as background)
- Style: solid (default)
- Width/Height: always auto-sized by GraphViz (ignore explicit attributes)

### Label content
- Label order: Name → Technology → Description (top to bottom)
- Empty description: omit the row entirely
- Long text: wordwrap, no truncation
- Technology shown in label table

### Drill-down indicator
- Collapsed units with subunits show `[+]` postfix in name
- Example: `My System [+]` indicates expandable

### Graph layout
- Direction: top-to-bottom (rankdir=TB)
- Title: properties.name displayed as diagram title

### Legend
- Small legend box included in each diagram
- Shows all unit types with their shapes
- Legend position: bottom or corner (Claude's discretion)

### Claude's Discretion
- Exact GraphViz colors mapped from C4-PlantUML palette
- Legend positioning and size
- Exact HTML label formatting (fonts, padding, alignment)
- Node spacing and padding values

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/model/unit.go`: Unit struct with Type, Name, Description, Technology, Subunits, Links, styling attributes
- `internal/model/link.go`: Link struct with Arrow, Rank, Color, Style, Technology, Description, LabelPosition
- `internal/model/colors.go`: C4 color constants (may need mapping to GraphViz)
- `internal/validator/index.go`: BuildIndex pattern for traversing nested units
- `internal/parser/parser.go`: Model struct with Properties and Units map

### Established Patterns
- Type discriminator pattern (UnitType enum)
- Map-based traversal for nested structures
- Separate packages for model, parser, validator

### Integration Points
- View generator receives validated *parser.Model
- Graph builder receives view-scoped unit collections
- Output feeds into Phase 4 renderer (DOT/SVG generation)

</code_context>

<specifics>
## Specific Ideas

- HTML-like record labels for proper cell alignment
- Emoji icons for visual distinction (👤, ⛁, ╟╢)
- C4-PlantUML level colors for visual hierarchy
- External palette clearly distinguishes external systems/actors

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 03-views-graphs*
*Context gathered: 2026-03-09*
