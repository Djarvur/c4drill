# Phase 4: Rendering & Output - Research

**Researched:** 2026-03-10
**Domain:** GraphViz DOT/SVG rendering with go-graphviz library
**Confidence:** HIGH

## Summary

Phase 4 implements rendering of graph structures to DOT and SVG files using the go-graphviz library. The library provides a pure Go implementation with WebAssembly-embedded Graphviz, eliminating external dependencies. The key implementation involves converting our `graph.Graph` types to go-graphviz's `cgraph.Graph` objects, then rendering to files with proper directory structure.

**Primary recommendation:** Create a new `internal/render` package with two main functions: `RenderDOT(g *graph.Graph) ([]byte, error)` and `RenderSVG(g *graph.Graph) ([]byte, error)`. Use go-graphviz's native API to build graphs programmatically rather than string concatenation.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **DOT Generation:** Use go-graphviz native API (not manual DOT string building)
- Build graphviz.Graph object using provided methods
- Render via graphviz.Render() with XDOT format for DOT output
- Keep our graph types (Graph, Node, Edge, Cluster) - renderer converts to go-graphviz
- Two separate render functions: RenderDOT(g *Graph) and RenderSVG(g *Graph)

- **HTML Labels:** Use HTML tables (<TABLE><TR><TD>) for node labels
- Matches Phase 3 label design with aligned cells
- Edge labels: two-line format with [Technology] on first line, Description on second

- **Cluster Styling:** Expanded units rendered as bordered boxes with labels
- Standard GraphViz cluster style: dashed border, rounded corners, label at top
- Matches C4-PlantUML container visualization

- **Output File Structure:**
  - C1 (Context): `{basename}.{format}` (flat file in output directory)
  - C2/C3 (expanded): `{basename}/{unit-path}.{format}` (directory hierarchy)
  - Nested units: dotted path becomes directory hierarchy
  - Example: `mainapp.api` -> `{basename}/mainapp/api.svg`
  - Directory structure created recursively as needed

- **Error Handling:** Fail fast on file write errors (stop immediately, report error)
- Leave partial output for user diagnostics
- Directory creation failures: report and exit with clear message
- No cleanup on failure (preserves diagnostic information)

### Claude's Discretion
- Exact go-graphviz API usage patterns
- Font specifications in DOT
- Exact spacing and padding in HTML labels
- Legend positioning and format

### Deferred Ideas (OUT OF SCOPE)
None - discussion stayed within phase scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| REND-01 | Renderer generates valid GraphViz DOT format | go-graphviz native API with `graphviz.XDOT` format |
| REND-02 | Renderer generates SVG via go-graphviz library | go-graphviz `graphviz.SVG` format |
| REND-03 | Output format controlled by --format flag (dot|svg) | Separate RenderDOT/RenderSVG functions |
| OUTP-01 | Context level renders to {basename}.{format} | Output path construction logic |
| OUTP-02 | Expanded units render to {basename}/{unit-name}.{format} | Directory hierarchy from dotted paths |
| OUTP-04 | Directory structure created recursively as needed | `os.MkdirAll()` for parent directories |
| QUAL-01 | All lint errors must be fixed before commit | golangci-lint v2 configuration |
| QUAL-02 | Lint config (.golangci.yml) MUST NOT be adjusted | Existing config preserved |
| QUAL-03 | nolint directives require explicit user confirmation | Avoid unless absolutely necessary |
| QUAL-04 | Minimum 75% test coverage required | Table-driven tests with testify |
| QUAL-05 | Coverage enforced in CI/quality gate | `go test -cover` verification |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| go-graphviz | latest | GraphViz rendering | Pure Go, WASM-embedded, no external deps |
| cgraph (subpkg) | latest | Graph construction API | Type-safe node/edge/cluster creation |
| gvc (subpkg) | latest | Rendering context | Layout and output control |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| stretchr/testify | v1.11.1 (existing) | Testing assertions | All test files |
| os (stdlib) | - | File I/O | Writing output files |
| path/filepath (stdlib) | - | Path manipulation | Directory structure creation |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| go-graphviz | Manual DOT string building | String building is error-prone, no validation, harder to maintain |
| go-graphviz | graphviz command-line | Requires external dependency, harder to distribute |
| XDOT format | Custom DOT serialization | XDOT is standard, ensures GraphViz compatibility |

**Installation:**
```bash
go get github.com/goccy/go-graphviz
```

## Architecture Patterns

### Recommended Project Structure
```
internal/
├── render/           # NEW: Rendering package
│   ├── render.go     # Main render functions (RenderDOT, RenderSVG)
│   ├── converter.go  # Convert graph.Graph to cgraph.Graph
│   ├── labels.go     # HTML label generation
│   ├── render_test.go
│   ├── converter_test.go
│   └── labels_test.go
├── output/           # NEW: Output file handling
│   ├── writer.go     # File writing with directory creation
│   └── writer_test.go
└── graph/            # EXISTING: Graph types (input to render)
```

### Pattern 1: Graphviz Initialization and Cleanup
**What:** go-graphviz requires context and proper cleanup via defer
**When to use:** Every render operation
**Example:**
```go
// Source: https://github.com/goccy/go-graphviz README
func RenderDOT(g *graph.Graph) ([]byte, error) {
    ctx := context.Background()
    gv, err := graphviz.New(ctx)
    if err != nil {
        return nil, fmt.Errorf("create graphviz: %w", err)
    }
    defer gv.Close()

    graph, err := buildGraphvizGraph(gv, g)
    if err != nil {
        return nil, fmt.Errorf("build graph: %w", err)
    }
    defer graph.Close()

    var buf bytes.Buffer
    if err := gv.Render(ctx, graph, graphviz.XDOT, &buf); err != nil {
        return nil, fmt.Errorf("render: %w", err)
    }
    return buf.Bytes(), nil
}
```

### Pattern 2: Node Creation with HTML Labels
**What:** Use `SetLabel()` with HTML table syntax, `SetShape(PlainTextShape)` for HTML
**When to use:** Creating nodes from our graph.Node type
**Example:**
```go
// Source: go-graphviz cgraph/attribute.go types
func createNode(gvGraph *cgraph.Graph, node *graph.Node) error {
    n, err := gvGraph.CreateNodeByName(node.ID)
    if err != nil {
        return err
    }

    // HTML labels require shape=plaintext (or no shape)
    n.SetShape(cgraph.PlainTextShape)
    n.SetLabel(buildHTMLLabel(node.Label))
    n.SetFillColor(node.Style.FillColor)
    n.SetFontColor(node.Style.FontColor)
    n.SetStyle(cgraph.FilledNodeStyle)

    if node.Style.BorderStyle == "dashed" {
        // Note: For node borders, use penwidth and color
        n.SetPenWidth(1.0)
    }
    return nil
}

func buildHTMLLabel(label *graph.Label) string {
    // HTML table format for C4-style labels
    var sb strings.Builder
    sb.WriteString(`<<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0" CELLPADDING="4">`)
    sb.WriteString(`<TR><TD>`)
    if label.Icon != "" {
        sb.WriteString(label.Icon)
        sb.WriteString(" ")
    }
    sb.WriteString(label.Name)
    sb.WriteString(`</TD></TR>`)
    if label.Technology != "" {
        sb.WriteString(`<TR><TD>`)
        sb.WriteString(label.Technology)
        sb.WriteString(`</TD></TR>`)
    }
    if label.Description != "" {
        sb.WriteString(`<TR><TD>`)
        sb.WriteString(label.Description)
        sb.WriteString(`</TD></TR>`)
    }
    sb.WriteString(`</TABLE>>`)
    return sb.String()
}
```

### Pattern 3: Cluster (Subgraph) Creation
**What:** Use `CreateSubGraphByName()` with "cluster_" prefix for clusters
**When to use:** Creating clusters from our graph.Cluster type
**Example:**
```go
// Source: go-graphviz cgraph/cgraph.go
func createCluster(parent *cgraph.Graph, cluster *graph.Cluster) (*cgraph.Graph, error) {
    // Name must start with "cluster_" for GraphViz to render as cluster
    subgraph, err := parent.CreateSubGraphByName(cluster.ID)
    if err != nil {
        return nil, err
    }

    // Cluster styling
    subgraph.SetLabel(cluster.Label.Name)
    subgraph.SetStyle(cgraph.FilledGraphStyle)
    subgraph.SetFillColor(cluster.Style.FillColor)

    // Create nodes inside cluster
    for _, node := range cluster.Nodes {
        if err := createNodeInCluster(subgraph, node); err != nil {
            return nil, err
        }
    }
    return subgraph, nil
}
```

### Pattern 4: Edge Creation with Labels
**What:** Use `CreateEdgeByName()` with multi-line labels
**When to use:** Creating edges from our graph.Edge type
**Example:**
```go
// Source: go-graphviz cgraph/cgraph.go + attribute.go
func createEdge(gvGraph *cgraph.Graph, source, target *cgraph.Node, edge *graph.Edge) error {
    e, err := gvGraph.CreateEdgeByName(
        edge.Source + "_to_" + edge.Target,
        source,
        target,
    )
    if err != nil {
        return err
    }

    // Two-line edge label format
    var label string
    if edge.Label.Technology != "" {
        label = "[" + edge.Label.Technology + "]"
        if edge.Label.Description != "" {
            label += "\n" + edge.Label.Description
        }
    } else if edge.Label.Description != "" {
        label = edge.Label.Description
    }
    e.SetLabel(label)
    e.SetFontSize(10.0)

    // Edge style
    switch edge.Style {
    case "dashed":
        e.SetStyle(cgraph.DashedEdgeStyle)
    case "dotted":
        e.SetStyle(cgraph.DottedEdgeStyle)
    default:
        e.SetStyle(cgraph.SolidEdgeStyle)
    }

    // Arrow direction
    switch edge.ArrowHead {
    case graph.ArrowForward:
        e.SetDir(cgraph.ForwardDir)
    case graph.ArrowReverse:
        e.SetDir(cgraph.BackDir)
    case graph.ArrowBoth:
        e.SetDir(cgraph.BothDir)
    case graph.ArrowNone:
        e.SetDir(cgraph.NoneDir)
    }

    return nil
}
```

### Pattern 5: Graph-Level Settings
**What:** Configure layout direction, edge routing, fonts
**When to use:** During graph initialization
**Example:**
```go
// Source: go-graphviz cgraph/attribute.go
func configureGraph(gvGraph *cgraph.Graph, g *graph.Graph) {
    // Layout direction (TB=top-to-bottom, LR=left-to-right)
    gvGraph.SetRankDir(cgraph.TBRank)
    if g.Direction == "LR" {
        gvGraph.SetRankDir(cgraph.LRRank)
    }

    // Edge routing style
    // Options: "true" (spline), "false" (line), "none", "ortho", "polyline"
    switch g.EdgeStyle {
    case "spline":
        gvGraph.SetSplines("true")
    case "straight":
        gvGraph.SetSplines("false")
    case "ortho":
        gvGraph.SetSplines("ortho")
    }

    // Font settings (use cross-platform fonts)
    gvGraph.SetFontName("Helvetica")
    gvGraph.SetFontSize(14.0)

    // Spacing
    gvGraph.SetNodeSeparator(0.5)
    gvGraph.SetRankSeparator(0.75)
    gvGraph.SetPad(0.1)
}
```

### Pattern 6: File Output with Directory Creation
**What:** Create parent directories before writing file
**When to use:** Writing rendered output to disk
**Example:**
```go
// Source: Standard Go patterns
func WriteOutput(basename, unitPath, format string, data []byte) error {
    var outputPath string
    if unitPath == "" {
        // C1: flat file
        outputPath = fmt.Sprintf("%s.%s", basename, format)
    } else {
        // C2/C3: directory hierarchy
        outputPath = fmt.Sprintf("%s/%s.%s", basename, unitPath, format)
    }

    // Create parent directories
    dir := filepath.Dir(outputPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("create directory %s: %w", dir, err)
    }

    // Write file (fail fast on error)
    if err := os.WriteFile(outputPath, data, 0644); err != nil {
        return fmt.Errorf("write file %s: %w", outputPath, err)
    }

    return nil
}
```

### Anti-Patterns to Avoid
- **Manual DOT string building:** Error-prone, no validation, special character escaping issues
- **Forgetting to close resources:** Always defer `gv.Close()` and `graph.Close()`
- **Cluster naming without "cluster_" prefix:** Subgraphs won't render as visual clusters
- **Using shape=record with HTML labels:** Use shape=plaintext for HTML tables
- **Ignoring error returns:** go-graphviz returns errors for invalid operations

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| DOT generation | String concatenation | go-graphviz cgraph | Escaping, validation, format correctness |
| Graph layout | Custom positioning | go-graphviz layout engines | Complex algorithms, edge routing |
| SVG generation | Custom SVG builder | go-graphviz Render() | Full GraphViz rendering pipeline |
| Directory creation | Manual path parsing | os.MkdirAll() | Handles nested paths, permissions |

**Key insight:** go-graphviz handles all the complexity of DOT syntax, escaping, layout, and rendering. Attempting to build DOT strings manually introduces subtle bugs with special characters, quoting, and format changes.

## Common Pitfalls

### Pitfall 1: HTML Label Escaping
**What goes wrong:** Special characters in labels break DOT parsing
**Why it happens:** HTML-like labels need proper escaping for `<`, `>`, `&`
**How to avoid:** Use go-graphviz's `StrdupHTML()` for HTML label strings, or ensure labels are sanitized
**Warning signs:** DOT parsing errors, malformed output

### Pitfall 2: Cluster Node Association
**What goes wrong:** Nodes appear outside their clusters
**Why it happens:** Nodes must be created within the subgraph context, not the parent graph
**How to avoid:** Create nodes using the subgraph's `CreateNodeByName()`, not the parent graph's method
**Warning signs:** Nodes floating outside cluster boundaries in output

### Pitfall 3: Resource Cleanup
**What goes wrong:** Memory leaks, resource exhaustion in long-running processes
**Why it happens:** go-graphviz uses WASM resources that must be explicitly freed
**How to avoid:** Always defer `Close()` on both graphviz instance and graph objects
**Warning signs:** Increasing memory usage over time

### Pitfall 4: Directory Path Handling
**What goes wrong:** File write fails on nested paths
**Why it happens:** Parent directories don't exist
**How to avoid:** Always call `os.MkdirAll()` before writing, handle dotted paths correctly
**Warning signs:** "no such file or directory" errors on C2/C3 output

### Pitfall 5: Context Cancellation
**What goes wrong:** Render operations hang or fail unexpectedly
**Why it happens:** go-graphviz uses context for cancellation, WASM operations can be slow
**How to avoid:** Use `context.Background()` for simple cases, add timeout for long operations
**Warning signs:** Timeout errors, interrupted renders

## Code Examples

Verified patterns from official sources:

### Complete Render Function
```go
// Source: go-graphviz README + cgraph API
package render

import (
    "bytes"
    "context"
    "fmt"

    "github.com/goccy/go-graphviz"
    "github.com/goccy/go-graphviz/cgraph"
    "github.com/Djarvur/c4drill/internal/graph"
)

// RenderDOT renders a graph to DOT format.
func RenderDOT(g *graph.Graph) ([]byte, error) {
    return render(g, graphviz.XDOT)
}

// RenderSVG renders a graph to SVG format.
func RenderSVG(g *graph.Graph) ([]byte, error) {
    return render(g, graphviz.SVG)
}

func render(g *graph.Graph, format graphviz.Format) ([]byte, error) {
    ctx := context.Background()

    gv, err := graphviz.New(ctx)
    if err != nil {
        return nil, fmt.Errorf("create graphviz instance: %w", err)
    }
    defer gv.Close()

    cg, err := buildCgraph(gv, g)
    if err != nil {
        return nil, fmt.Errorf("build graph: %w", err)
    }
    defer cg.Close()

    var buf bytes.Buffer
    if err := gv.Render(ctx, cg, format, &buf); err != nil {
        return nil, fmt.Errorf("render output: %w", err)
    }

    return buf.Bytes(), nil
}

func buildCgraph(gv *graphviz.Graphviz, g *graph.Graph) (*cgraph.Graph, error) {
    cg, err := gv.Graph()
    if err != nil {
        return nil, err
    }

    // Configure graph-level settings
    configureGraphSettings(cg, g)

    // Build node lookup map
    nodeMap := make(map[string]*cgraph.Node)

    // Create top-level nodes
    for _, node := range g.Nodes {
        cn, err := createNode(cg, node)
        if err != nil {
            return nil, fmt.Errorf("create node %s: %w", node.ID, err)
        }
        nodeMap[node.ID] = cn
    }

    // Create clusters with their nodes
    for _, cluster := range g.Clusters {
        ccg, err := createCluster(cg, cluster, nodeMap)
        if err != nil {
            return nil, fmt.Errorf("create cluster %s: %w", cluster.ID, err)
        }
        _ = ccg // Cluster created, nodes added to map
    }

    // Create edges
    for _, edge := range g.Edges {
        source := nodeMap[edge.Source]
        target := nodeMap[edge.Target]
        if source == nil || target == nil {
            continue // Skip edges with missing endpoints
        }
        if err := createEdge(cg, source, target, edge); err != nil {
            return nil, fmt.Errorf("create edge %s->%s: %w", edge.Source, edge.Target, err)
        }
    }

    return cg, nil
}
```

### Output Writer
```go
// Source: Standard Go patterns
package output

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

// Writer handles file output with directory creation.
type Writer struct {
    baseDir string
}

func NewWriter(baseDir string) *Writer {
    return &Writer{baseDir: baseDir}
}

// Write writes rendered data to the appropriate output path.
// For C1: {basename}.{format}
// For C2/C3: {basename}/{unit-path}.{format}
func (w *Writer) Write(basename, unitPath, format string, data []byte) error {
    var relPath string
    if unitPath == "" {
        relPath = fmt.Sprintf("%s.%s", basename, format)
    } else {
        // Convert dotted path to directory hierarchy
        dirPath := strings.ReplaceAll(unitPath, ".", string(filepath.Separator))
        relPath = filepath.Join(basename, dirPath+"."+format)
    }

    fullPath := filepath.Join(w.baseDir, relPath)

    // Create parent directories (fail fast on error)
    dir := filepath.Dir(fullPath)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("create output directory %s: %w", dir, err)
    }

    // Write file (fail fast on error)
    if err := os.WriteFile(fullPath, data, 0644); err != nil {
        return fmt.Errorf("write output file %s: %w", fullPath, err)
    }

    return nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| graphviz CLI + os/exec | go-graphviz WASM | 2020+ | No external dependency, pure Go |
| Manual DOT strings | cgraph API | Always | Type-safe, validated output |
| Single output format | Multiple formats | Always | DOT, SVG, PNG, JPG support |

**Deprecated/outdated:**
- graphviz CLI invocation: Requires external installation, harder distribution
- String template DOT generation: No validation, escaping issues

## Open Questions

1. **Legend Implementation**
   - What we know: Legend struct exists in graph.Graph but is empty placeholder
   - What's unclear: Exact format, positioning, content for C4 legend
   - Recommendation: Implement basic legend in Phase 4, refine in Phase 5 with navigation

2. **Font Selection**
   - What we know: Cross-platform fonts needed for consistent rendering
   - What's unclear: Whether to use Helvetica, Arial, or system default
   - Recommendation: Use "Helvetica" as default, SVG viewers typically handle font substitution

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + stretchr/testify v1.11.1 |
| Config file | none - standard go test |
| Quick run command | `go test ./internal/render/... -v -short` |
| Full suite command | `go test ./... -cover -race` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| REND-01 | Generates valid DOT | unit | `go test ./internal/render/... -run TestRenderDOT -v` | Wave 0 |
| REND-02 | Generates valid SVG | unit | `go test ./internal/render/... -run TestRenderSVG -v` | Wave 0 |
| OUTP-01 | C1 flat output path | unit | `go test ./internal/output/... -run TestWriteC1 -v` | Wave 0 |
| OUTP-02 | C2/C3 directory hierarchy | unit | `go test ./internal/output/... -run TestWriteC2C3 -v` | Wave 0 |
| OUTP-04 | Recursive directory creation | unit | `go test ./internal/output/... -run TestMkdirAll -v` | Wave 0 |
| QUAL-04 | 75% coverage | coverage | `go test ./internal/render/... -cover` | Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/render/... ./internal/output/... -v -short`
- **Per wave merge:** `go test ./... -cover -race`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/render/render.go` - RenderDOT, RenderSVG functions
- [ ] `internal/render/converter.go` - buildCgraph function
- [ ] `internal/render/labels.go` - buildHTMLLabel function
- [ ] `internal/render/render_test.go` - Unit tests for rendering
- [ ] `internal/output/writer.go` - Write function with directory creation
- [ ] `internal/output/writer_test.go` - Unit tests for output writing
- [ ] Framework install: `go get github.com/goccy/go-graphviz` - Add to go.mod

## Sources

### Primary (HIGH confidence)
- https://raw.githubusercontent.com/goccy/go-graphviz/master/README.md - Library overview, usage examples
- https://raw.githubusercontent.com/goccy/go-graphviz/master/graphviz.go - Main API (New, Render, Graph methods)
- https://raw.githubusercontent.com/goccy/go-graphviz/master/cgraph/cgraph.go - Graph, Node, Edge, SubGraph methods
- https://raw.githubusercontent.com/goccy/go-graphviz/master/cgraph/attribute.go - Attribute types and setters

### Secondary (MEDIUM confidence)
- Project source files: internal/graph/graph.go, internal/graph/builder.go - Input types for renderer
- Project test patterns: internal/graph/*_test.go - Test structure conventions

### Tertiary (LOW confidence)
- N/A - All critical information from primary sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - go-graphviz is well-documented, pure Go, actively maintained
- Architecture: HIGH - Clear API patterns from source code review
- Pitfalls: HIGH - Common patterns identified from API documentation

**Research date:** 2026-03-10
**Valid until:** 30 days - Library API stable, patterns unlikely to change
