# Phase 8: All-Expanded Mode - Context

**Gathered:** 2026-03-11
**Status:** Ready for planning

<domain>
## Phase Boundary

Add a new rendering mode (`--expanded` CLI flag) that generates a single diagram showing all units at all C4 levels (C1, C2, C3) in nested clusters with all edges visible (including cross-level edges). Output file follows `{basename}.expanded.{ext}` naming convention.

This is an additional mode, not a replacement for existing C1/C2/C3 view generation.

</domain>

<decisions>
## Implementation Decisions

### Expanded view structure
- Use modified C1 view approach, not a standalone GenerateAllExpandedView() function
- Reuse existing View struct structure (no new fields needed)
- Include all units in the model regardless of the expanded attribute
- Add to internal/view/scope.go alongside existing view generators

### Cross-level edge display
- Direct edges between units regardless of nesting depth (e.g., C1 system → C3 component)
- Same visual style for all edges (no distinction between cross-level and same-level)
- Use nested nodes approach with GraphViz edge routing across cluster boundaries
- Always include external boundary nodes for linked units outside the model

### Output file strategy
- Replace normal output when --expanded is used (only create `basename.expanded.{ext}`, skip standard C1/C2/C3 files)
- Share the same `--format` flag (dot/svg)
- Write to output directory root, same level as standard C1 output
- Single file only, no subdirectory structure (no expanded views for specific systems)

### Flag behavior and UX
- Flag name: `--expanded` (boolean flag, simple on/off)
- Combine paths: use `--output` directory, append `.expanded.{ext}` to basename
- Use standard title (properties.name) same as normal diagrams (no "expanded" suffix)

### Performance handling
- No size limits or warnings (generate all-expanded diagrams regardless of model size)
- Silent by default (same as current CLI behavior)
- Use same GraphViz optimization settings as standard views
- Standard error handling if graphviz fails due to complexity (no special suggestions)

### Claude's Discretion
- Exact GraphViz layout settings for nested clusters
- Internal implementation details of how to build nested clusters in graph package
- How to traverse model to include all units with proper nesting hierarchy
- Edge routing optimization details in graph building

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- **cobra CLI framework**: Already used in cmd/c4drill/root.go for flag registration and command structure
- **GenerateC1View()**: Template for view generation in internal/view/scope.go (line 13)
- **GenerateC2View() / GenerateC3View()**: Examples of scoped view generation that handle parent paths
- **collectExpandedPaths()**: Existing path collection logic in root.go (line 122) that can be adapted or simplified for all-expanded mode
- **External boundary node pattern**: createExternalBoundaryNode() in scope.go (line 72) for handling linked units outside model

### Established Patterns
- **Flag registration**: PersistentFlags().StringVarP() pattern in root.go (line 61)
- **Pipeline flow**: Parse → Validate → Collect paths → Process views → Render → Write (root.go runRoot function)
- **View generation**: GenerateXxxView(m *parser.Model, path string) pattern returns *View struct
- **Graph building**: BuildGraphWithPath(v, unitPath, basename, format) takes a View and produces GraphViz graph
- **Error handling**: Static error variables (errGenerateView, errBuildGraph) for consistent error reporting

### Integration Points
- **cmd/c4drill/root.go**: Add `--expanded` flag to PersistentFlags, modify runRoot() to handle all-expanded mode path collection
- **internal/view/scope.go**: Add new view generation function for all-expanded mode alongside existing generators
- **internal/graph**: BuildGraphWithPath() will need to handle nested cluster structures for all-expanded views
- **internal/render**: Existing Render() function should work unchanged with all-expanded graphs

</code_context>

<specifics>
## Specific Ideas

No specific requirements — open to standard approaches that maintain consistency with existing CLI design and view generation patterns.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 08-all-expanded-mode*
*Context gathered: 2026-03-11*
