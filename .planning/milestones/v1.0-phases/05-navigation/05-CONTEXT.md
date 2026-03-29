# Phase 5: Navigation - Context

**Gathered:** 2026-03-10
**Status:** Ready for planning

<domain>
## Phase Boundary

Add clickable navigation links between diagram levels. Collapsed units link to their expanded views (explore links). All diagrams except C1 have back-links to parent. Breadcrumb trails show navigation path. Uses GraphViz URL attributes for SVG interactivity.

</domain>

<decisions>
## Implementation Decisions

### Explore Links
- Node name shows "System Name [+]" where [+] indicates drill-down capability
- Only expandable node types get explore links (systems, boxes - not persons, databases, queues)
- Links use relative paths (e.g., "./mainapp/api.svg")
- Clickable nodes set cursor:pointer via GraphViz (no visual style change)
- Use GraphViz URL attribute (SetURL) for whole-node clickability

### Back-Links
- Text link at top-left: "← Back to {parent_name}"
- Positioned outside diagram area (in margin/padding, not part of layout)
- C1 (Context) diagram has no back-link (it's the root)
- Uses parent's actual name in link text

### Breadcrumbs
- Show unit names: "Main System › API Container › Auth Service"
- Chevron (›) separator between items
- Same line as back-link: "← Back to Main | Main System › API Container"
- Current level item is NOT clickable (only ancestors are links)
- No limit on breadcrumb length - show full path

### Path Handling
- Paths pre-computed during graph building stage (stored in graph types)
- URL encode paths for special characters (url.PathEscape)
- Relative paths for portability across hosting contexts

### Claude's Discretion
- Exact positioning of navigation bar (margin values)
- Font size and styling for navigation elements
- Whether to use HTML label or graph label for navigation bar

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/graph/graph.go`: Graph, Node types - need to add URL field for links
- `internal/graph/builder.go`: BuildGraph function - extend to compute paths
- `internal/render/converter.go`: Node conversion - add SetURL calls
- `internal/render/labels.go`: HTML label building - add [+] indicator
- `internal/output/writer.go`: File path construction - already handles nested paths

### Established Patterns
- Two-phase construction: view then graph
- Relative path handling in writer (dotted path → directory hierarchy)
- HTML table labels for node formatting
- go-graphviz native API for rendering

### Integration Points
- Graph builder needs current file path context to compute relative target paths
- Node type needs URL field for explore links
- Graph type needs parent info for back-link and breadcrumbs
- Renderer calls SetURL on nodes that have explore links

### Data Model Changes
- Node struct: add `ExploreURL string` field
- Graph struct: add `Parent *ParentInfo` field (nil for C1)
- ParentInfo struct: `Name string`, `BackURL string`, `Breadcrumb []BreadcrumbItem`
- BreadcrumbItem struct: `Name string`, `URL string` (URL empty for current)

</code_context>

<specifics>
## Specific Ideas

- [+] indicator follows C4-PlantUML convention for expandable nodes
- Navigation bar positioned in graph's label area or as invisible node
- Standard web breadcrumb patterns (chevron, current not clickable)

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 05-navigation*
*Context gathered: 2026-03-10*
