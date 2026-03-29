# Phase 3: Views & Graphs - Research

**Researched:** 2026-03-09
**Domain:** C4 view generation and graph data structure construction
**Confidence:** HIGH

## Summary

This phase transforms the validated C4 model into scoped views (C1/C2/C3) and constructs graph data structures for later rendering. The key insight is that this phase does NOT generate DOT/SVG output - it builds intermediate in-memory representations that the Phase 4 renderer will consume.

The view generator creates different diagram scopes based on the C4 hierarchy: C1 shows all top-level units, C2 shows subunits of expanded containers, and C3 shows components within expanded containers. The graph builder then creates node and edge structures with type-appropriate shapes and styling attributes.

**Primary recommendation:** Implement two separate packages (`internal/view` and `internal/graph`) with clear data structures. Use the existing `BuildIndex` pattern from validator for traversing nested units.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### View scoping rules
- C1 = all top-level units + links between them
- C2 = subunits of expanded container + links between subunits
- C3 = subunits of expanded container (C3 level)
- Each expanded unit gets its own view (not combined per-parent)
- Per-unit `expanded` attribute only (no global properties.expanded default)

#### External boundary nodes
- Units referenced from in-scope units but outside the view appear as collapsed boundary nodes
- External boundary nodes: same size as regular nodes, dashed border, external palette colors
- Links to external boundary nodes are shown normally

#### Link visibility
- Links shown if both endpoints are in view OR one endpoint is in view and other is external boundary
- Scoped to view: link appears only where both endpoints are visible, not in child views
- Bidirectional relationships (link + linkFrom between same units): separate edges, not combined
- Multiple links between same two units: show all separately
- Self-links (unit linking to itself): rejected by validator
- linkFrom renders same as link (just defined from target perspective)

#### Link labels
- Technology and description in separate cells
- Technology appears in square brackets: `[HTTP]`
- Label position: default middle, override via link.labelPosition
- Arrow direction: read from link.arrow attribute
- Rank attribute: ignored for layout (GraphViz decides)
- Default link style: solid black line

#### Edge routing style
- GraphViz only supports one edge style per diagram
- C1 view: use properties.edges
- Child views: use expanded unit's edges attribute
- Values: straight, spline, square

#### Node shape design
- Person: HTML record with icon: `| { name | description }`
- DB: HTML record with icon: `| { name | description }`
- Queue: HTML record with bars: `|\n| { name | description }`
- System/Container/Component: HTML record: `{ name | technology | description }`
- Box: same as system shape
- All use HTML-like labels for proper cell formatting

#### External variants styling
- Same shape as base type
- Different border and text color (external palette)
- Distinct external color palette (not level colors)

#### Cluster rendering
- Expanded units rendered as subgraph clusters with label at top
- Cluster label shows full node label (name, technology, description)
- Children nodes rendered inside cluster

#### Style defaults
- No inheritance: each unit's style is independent
- Colors: different levels have different colors from C4-PlantUML palette
- Same level units have same colors
- Border: defaults to level color (same as background)
- Style: solid (default)
- Width/Height: always auto-sized by GraphViz (ignore explicit attributes)

#### Label content
- Label order: Name -> Technology -> Description (top to bottom)
- Empty description: omit the row entirely
- Long text: wordwrap, no truncation
- Technology shown in label table

#### Drill-down indicator
- Collapsed units with subunits show `[+]` postfix in name
- Example: `My System [+]` indicates expandable

#### Graph layout
- Direction: top-to-bottom (rankdir=TB)
- Title: properties.name displayed as diagram title

#### Legend
- Small legend box included in each diagram
- Shows all unit types with their shapes
- Legend position: bottom or corner (Claude's discretion)

### Claude's Discretion
- Exact GraphViz colors mapped from C4-PlantUML palette
- Legend positioning and size
- Exact HTML label formatting (fonts, padding, alignment)
- Node spacing and padding values

### Deferred Ideas (OUT OF SCOPE)
None - discussion stayed within phase scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| VIEW-01 | Generator creates C1 (Context) level view from model | View package with `GenerateC1View(model) -> View` |
| VIEW-02 | Generator creates C2 (Containers) level view for expanded systems | View package with `GenerateC2View(model, expandedUnit) -> View` |
| VIEW-03 | Generator creates C3 (Components) level view for expanded containers | View package with `GenerateC3View(model, expandedUnit) -> View` |
| VIEW-04 | Collapsed units render as single record shape | Graph package with `NodeType.Collapsed = false` |
| VIEW-05 | Expanded units render as clusters with subunits inside | Graph package with `Cluster` struct containing `Nodes` |
| VIEW-06 | View respects expanded list from properties and unit-level overrides | View package reads `unit.Expanded` field |
| VIEW-07 | Styling (color, border, style, edges) inherits from parent with override | Each node carries own style, no inheritance per decisions |
| GRPH-01 | Builder creates nodes for each unit with type-appropriate shapes | Graph package with `ShapeForType(UnitType)` function |
| GRPH-02 | Builder creates edges for each link definition | Graph package with `Edge` struct referencing source/target |
| GRPH-03 | Builder applies edge routing style (straight, spline, square) | Graph package with `EdgeStyle` enum |
| GRPH-04 | Builder creates clusters for expanded units | Graph package with `Cluster` struct |
| GRPH-05 | Shapes: person uses icon, db uses cylinder icon, queue uses bars | Graph package with `ShapeForType()` mapping |
| GRPH-06 | System shape includes name, description, explore link | Node `Label` struct with Name, Technology, Description fields |
| QUAL-01 | All lint errors must be fixed before commit | Standard Go linting via golangci-lint |
| QUAL-02 | Lint config MUST NOT be adjusted to silence errors | Use existing `.golangci.yml` |
| QUAL-03 | nolint directives require explicit user confirmation | Document in code review |
| QUAL-04 | Minimum 75% test coverage required | Test files in each package |
| QUAL-05 | Coverage enforced in CI/quality gate | Run `go test -cover` in CI |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/pelletier/go-toml/v2 | v2.2.4 | TOML parsing | Already in use, proven |
| github.com/stretchr/testify | v1.11.1 | Testing assertions | Already in use, idiomatic |

### Supporting
| Library | Purpose | When to Use |
|---------|---------|-------------|
| (standard library) | All graph/view structures | This phase - no external graph libraries needed |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Custom graph types | gonum/graph | gonum is overkill for our simple DAG; we only need nodes, edges, clusters |
| External graph library | dominikh/go-graph | Adds dependency for trivial data structures; hand-roll is cleaner |

**Key insight:** This phase does NOT need a graph algorithm library. We are building a simple graph *representation* for later DOT generation, not performing graph algorithms. Custom structs are simpler and more maintainable than adapting to a generic graph interface.

**Installation:**
No new dependencies required. Use existing `go.mod`.

## Architecture Patterns

### Recommended Project Structure
```
internal/
├── view/           # View generation package (NEW)
│   ├── view.go     # View struct and generation functions
│   ├── view_test.go
│   ├── scope.go    # View scoping logic (C1/C2/C3)
│   └── scope_test.go
├── graph/          # Graph construction package (NEW)
│   ├── graph.go    # Graph, Node, Edge, Cluster structs
│   ├── graph_test.go
│   ├── builder.go  # BuildGraph(view) -> Graph
│   ├── builder_test.go
│   ├── shapes.go   # ShapeForType, label formatting
│   └── shapes_test.go
├── model/          # Existing - domain types
├── parser/         # Existing - TOML parsing
└── validator/      # Existing - validation rules
```

### Pattern 1: View Scoping (C1/C2/C3)

**What:** A View represents a scoped subset of the model with specific units and links visible.

**When to use:** All view generation - C1 for context, C2 for containers, C3 for components.

**Data Structure:**
```go
// internal/view/view.go
package view

import "github.com/Djarvur/c4drill/internal/model"

// View represents a scoped view of the architecture model.
type View struct {
    // Level indicates the C4 level (C1, C2, C3).
    Level Level
    // Title is the diagram title (from properties.name or parent name).
    Title string
    // Units are the units visible in this view (keyed by full path).
    Units map[string]*ViewUnit
    // Edges are the edge routing style for this view.
    Edges string
    // Parent is the parent unit path for C2/C3 views (empty for C1).
    Parent string
    // ExpandedUnit is the unit being expanded (for C2/C3).
    ExpandedUnit string
}

// Level represents the C4 hierarchy level.
type Level int

const (
    LevelC1 Level = iota // Context level
    LevelC2              // Container level
    LevelC3              // Component level
)

// ViewUnit represents a unit within a view.
type ViewUnit struct {
    // Unit is the underlying model unit.
    Unit *model.Unit
    // FullPath is the dotted path (e.g., "mainapp.api").
    FullPath string
    // IsExpanded indicates if this unit is expanded (shows subunits).
    IsExpanded bool
    // IsExternal indicates if this is an external boundary node.
    IsExternal bool
    // HasSubunits indicates if this unit has children (for [+] indicator).
    HasSubunits bool
}

// GenerateC1View creates a C1 (Context) level view.
func GenerateC1View(m *parser.Model) *View {
    // Implementation: all top-level units, links between them
}

// GenerateC2View creates a C2 (Containers) view for an expanded system.
func GenerateC2View(m *parser.Model, systemPath string) *View {
    // Implementation: subunits of expanded system + links between them
}

// GenerateC3View creates a C3 (Components) view for an expanded container.
func GenerateC3View(m *parser.Model, containerPath string) *View {
    // Implementation: subunits of expanded container + links between them
}
```

### Pattern 2: Graph Construction

**What:** Build a graph representation from a view, with nodes, edges, and clusters.

**When to use:** After view generation, before DOT rendering.

**Data Structure:**
```go
// internal/graph/graph.go
package graph

// Graph represents a graph structure for rendering.
type Graph struct {
    // Title is the diagram title.
    Title string
    // Direction is the layout direction (default: "TB").
    Direction string
    // EdgeStyle is the edge routing style.
    EdgeStyle string
    // Nodes are all nodes in the graph.
    Nodes []*Node
    // Edges are all edges in the graph.
    Edges []*Edge
    // Clusters are subgraph clusters for expanded units.
    Clusters []*Cluster
    // Legend contains legend information.
    Legend *Legend
}

// Node represents a single node in the graph.
type Node struct {
    // ID is a unique identifier (full path).
    ID string
    // Label contains the formatted label parts.
    Label *Label
    // Shape is the node shape (determined by type).
    Shape Shape
    // Style contains visual styling.
    Style *NodeStyle
    // IsExternal indicates external boundary node.
    IsExternal bool
    // IsInCluster indicates node is inside a cluster.
    IsInCluster bool
}

// Label represents a node label with multiple parts.
type Label struct {
    // Name is the primary name (with [+] indicator if expandable).
    Name string
    // Technology is the technology string (optional).
    Technology string
    // Description is the description text (optional).
    Description string
    // Icon is the emoji/icon prefix (optional).
    Icon string
}

// NodeStyle contains visual styling for a node.
type NodeStyle struct {
    // FillColor is the background color.
    FillColor string
    // BorderColor is the border color.
    BorderColor string
    // FontColor is the text color.
    FontColor string
    // BorderStyle is "solid" or "dashed".
    BorderStyle string
}

// Edge represents a connection between two nodes.
type Edge struct {
    // Source is the source node ID.
    Source string
    // Target is the target node ID.
    Target string
    // Label contains edge label information.
    Label *EdgeLabel
    // Style is the line style.
    Style string
    // ArrowHead is the arrow direction.
    ArrowHead ArrowDirection
}

// EdgeLabel contains label information for an edge.
type EdgeLabel struct {
    // Technology appears in brackets.
    Technology string
    // Description is the relationship description.
    Description string
    // Position is where the label appears (middle, head, tail).
    Position string
}

// ArrowDirection represents arrow placement.
type ArrowDirection string

const (
    ArrowForward ArrowDirection = "forward"  // Arrow at target
    ArrowReverse ArrowDirection = "reverse"  // Arrow at source
    ArrowBoth    ArrowDirection = "both"     // Arrows at both ends
    ArrowNone    ArrowDirection = "none"     // No arrow
)

// Cluster represents a subgraph cluster for expanded units.
type Cluster struct {
    // ID is the cluster identifier.
    ID string
    // Label is the cluster label (parent unit info).
    Label *Label
    // Nodes are the nodes inside this cluster.
    Nodes []*Node
    // Style contains cluster styling.
    Style *NodeStyle
}

// Shape represents node shape types.
type Shape string

const (
    ShapeRecord   Shape = "record"   // Default rectangular
    ShapeHTML      Shape = "html"    // HTML-like label
    ShapeCluster   Shape = "cluster" // Subgraph cluster
)
```

### Pattern 3: Shape Determination

**What:** Map unit types to shapes and icons.

**Example:**
```go
// internal/graph/shapes.go
package graph

import "github.com/Djarvur/c4drill/internal/model"

// ShapeForType returns the appropriate shape for a unit type.
func ShapeForType(t model.UnitType) Shape {
    // All types use HTML-like labels for proper formatting
    return ShapeHTML
}

// IconForType returns the emoji icon for a unit type.
func IconForType(t model.UnitType) string {
    switch t {
    case model.TypePerson, model.TypePersonExternal:
        return ""
    case model.TypeDb, model.TypeDbExternal,
         model.TypeContainerDb, model.TypeComponentDb:
        return ""
    case model.TypeQueue, model.TypeQueueExternal,
         model.TypeContainerQueue, model.TypeComponentQueue:
        return ""
    default:
        return "" // System, Container, Component, Box
    }
}

// IsExternalType returns true if the type is an external variant.
func IsExternalType(t model.UnitType) bool {
    switch t {
    case model.TypePersonExternal, model.TypeSystemExternal,
         model.TypeDbExternal, model.TypeQueueExternal:
        return true
    default:
        return false
    }
}

// LevelForType returns the C4 level for a unit type.
func LevelForType(t model.UnitType) int {
    switch t {
    case model.TypePerson, model.TypePersonExternal,
         model.TypeSystem, model.TypeSystemExternal,
         model.TypeDb, model.TypeDbExternal,
         model.TypeQueue, model.TypeQueueExternal,
         model.TypeBox:
        return 1 // C1
    case model.TypeContainer, model.TypeContainerDb,
         model.TypeContainerQueue:
        return 2 // C2
    case model.TypeComponent, model.TypeComponentDb,
         model.TypeComponentQueue:
        return 3 // C3
    default:
        return 1
    }
}
```

### Anti-Patterns to Avoid

- **Combining view and graph logic in one package:** Keep view scoping separate from graph construction for clarity
- **Inheriting styles from parent:** Decisions explicitly state each unit's style is independent
- **Using external graph libraries for simple data structures:** Custom structs are cleaner for our use case
- **Generating DOT in this phase:** DOT generation belongs in Phase 4 (Rendering)

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Color constants | Define new colors | `internal/model/colors.go` | Already defined with C4-PlantUML palette |
| Unit traversal | New traversal code | `validator.BuildIndex()` pattern | Proven pattern, handles nesting |
| Unit types | New type definitions | `model.UnitType` constants | Already defined |
| Link types | New link types | `model.Link`, `model.ArrowDirection` | Already defined |

**Key insight:** Reuse existing model types and validator patterns. The `BuildIndex` function already demonstrates how to traverse nested units - adapt this pattern for view scoping.

## Common Pitfalls

### Pitfall 1: Confusing View Scope with Model Structure

**What goes wrong:** Trying to render the entire model in one view instead of scoping to C1/C2/C3.

**Why it happens:** The model has all units nested, so it's easy to forget to filter.

**How to avoid:** Always start from a View struct that explicitly defines which units are in scope. C1 = top-level only, C2 = children of expanded system, C3 = children of expanded container.

**Warning signs:** C1 view showing container/component nodes; C2 view showing top-level context nodes.

### Pitfall 2: Missing External Boundary Nodes

**What goes wrong:** Links to units outside the view disappear instead of showing external boundary nodes.

**Why it happens:** Forgetting to check if link targets are in-scope and create boundary nodes for out-of-scope references.

**How to avoid:** When building a view, scan all links from in-scope units. For any link target not in the current view, create an external boundary node.

**Warning signs:** Links with only one visible endpoint; missing dependencies in diagrams.

### Pitfall 3: Combining bidirectional links

**What goes wrong:** Merging `link` and `linkFrom` between the same two units into a single bidirectional edge.

**Why it happens:** Attempting to "simplify" the graph representation.

**How to avoid:** Per the decisions, bidirectional relationships (link + linkFrom between same units) are shown as separate edges, not combined.

**Warning signs:** Only one edge between units that should have two relationships.

### Pitfall 4: Forgetting the [+] Indicator

**What goes wrong:** Collapsed units with subunits look identical to leaf units.

**Why it happens:** Not checking if a unit has subunits when collapsed.

**How to avoid:** When building ViewUnit, set `HasSubunits = len(unit.Subunits) > 0`. When building Label, append `[+]` to name if `HasSubunits && !IsExpanded`.

**Warning signs:** Users cannot tell which units are expandable.

### Pitfall 5: Style Inheritance

**What goes wrong:** Child units inheriting colors/styles from parent when they should be independent.

**Why it happens:** Assuming inheritance simplifies configuration.

**How to avoid:** Per the decisions, "no inheritance: each unit's style is independent". Read style directly from unit fields, use level-appropriate defaults if not specified.

**Warning signs:** Container nodes with system colors; inconsistent styling across same-level units.

## Code Examples

### Example: C1 View Generation

```go
// internal/view/scope.go
package view

import (
    "github.com/Djarvur/c4drill/internal/parser"
)

// GenerateC1View creates a C1 (Context) level view showing all top-level units.
func GenerateC1View(m *parser.Model) *View {
    view := &View{
        Level:   LevelC1,
        Title:   m.Properties.Name,
        Edges:   m.Properties.Edges,
        Units:   make(map[string]*ViewUnit),
    }

    // Add all top-level units
    for name, unit := range m.Units {
        view.Units[name] = &ViewUnit{
            Unit:        unit,
            FullPath:    name,
            IsExpanded:  isExpanded(unit, m.Properties.Expanded, name),
            HasSubunits: len(unit.Subunits) > 0,
            IsExternal:  isExternalType(unit.Type),
        }
    }

    // Add external boundary nodes for links to undefined top-level units
    addExternalBoundaryNodes(view, m)

    return view
}

// isExpanded checks if a unit should be expanded in this view.
func isExpanded(unit *model.Unit, globalExpanded []string, path string) bool {
    // Check unit-level expanded list first
    for _, exp := range unit.Expanded {
        if exp == path {
            return true
        }
    }
    return false
}
```

### Example: Graph Builder

```go
// internal/graph/builder.go
package graph

import (
    "github.com/Djarvur/c4drill/internal/model"
    "github.com/Djarvur/c4drill/internal/view"
)

// BuildGraph constructs a graph from a view.
func BuildGraph(v *view.View) *Graph {
    g := &Graph{
        Title:      v.Title,
        Direction:  "TB", // top-to-bottom
        EdgeStyle:  v.Edges,
        Nodes:      make([]*Node, 0),
        Edges:      make([]*Edge, 0),
        Clusters:   make([]*Cluster, 0),
    }

    // Build nodes and clusters
    for path, vu := range v.Units {
        if vu.IsExpanded {
            // Create cluster for expanded unit
            cluster := buildCluster(vu, v)
            g.Clusters = append(g.Clusters, cluster)
        } else {
            // Create single node
            node := buildNode(vu)
            g.Nodes = append(g.Nodes, node)
        }
    }

    // Build edges
    g.Edges = buildEdges(v)

    return g
}

// buildNode creates a node from a view unit.
func buildNode(vu *view.ViewUnit) *Node {
    style := getStyleForType(vu.Unit.Type, vu.IsExternal)

    label := &Label{
        Name:        vu.Unit.Name,
        Technology:  vu.Unit.Technology,
        Description: vu.Unit.Description,
        Icon:        IconForType(vu.Unit.Type),
    }

    // Add [+] indicator for collapsed units with subunits
    if vu.HasSubunits && !vu.IsExpanded {
        label.Name = label.Name + " [+]"
    }

    return &Node{
        ID:         vu.FullPath,
        Label:      label,
        Shape:      ShapeForType(vu.Unit.Type),
        Style:      style,
        IsExternal: vu.IsExternal,
    }
}

// getStyleForType returns styling based on unit type and external status.
func getStyleForType(t model.UnitType, isExternal bool) *NodeStyle {
    if isExternal {
        return getExternalStyle(t)
    }
    return getLevelStyle(t)
}

func getLevelStyle(t model.UnitType) *NodeStyle {
    switch {
    case isC1Type(t):
        return &NodeStyle{
            FillColor:   model.SystemBackground,
            BorderColor: model.SystemBorder,
            FontColor:   model.FontColorC1C2,
            BorderStyle: "solid",
        }
    case isC2Type(t):
        return &NodeStyle{
            FillColor:   model.ContainerBackground,
            BorderColor: model.ContainerBorder,
            FontColor:   model.FontColorC1C2,
            BorderStyle: "solid",
        }
    case isC3Type(t):
        return &NodeStyle{
            FillColor:   model.ComponentBackground,
            BorderColor: model.ComponentBorder,
            FontColor:   model.FontColorC3,
            BorderStyle: "solid",
        }
    default:
        return &NodeStyle{
            FillColor:   model.SystemBackground,
            BorderColor: model.SystemBorder,
            FontColor:   model.FontColorC1C2,
            BorderStyle: "solid",
        }
    }
}
```

### Example: Edge Building with External Boundaries

```go
// internal/graph/builder.go (continued)

// buildEdges creates edges from view links.
func buildEdges(v *view.View) []*Edge {
    edges := make([]*Edge, 0)
    seen := make(map[string]bool) // Track processed links

    for path, vu := range v.Units {
        // Process outgoing links
        for target, link := range vu.Unit.Links {
            edgeKey := path + "->" + target
            if seen[edgeKey] {
                continue
            }
            seen[edgeKey] = true

            // Only include if target is in view (as regular or boundary node)
            if _, exists := v.Units[target]; !exists {
                continue
            }

            edge := &Edge{
                Source: path,
                Target: target,
                Label: &EdgeLabel{
                    Technology:  link.Technology,
                    Description: link.Description,
                    Position:    string(link.LabelPosition),
                },
                Style:     link.Style,
                ArrowHead: ArrowDirection(link.Arrow),
            }

            if edge.Style == "" {
                edge.Style = "solid"
            }
            if edge.Label.Position == "" {
                edge.Label.Position = "middle"
            }

            edges = append(edges, edge)
        }

        // Process incoming links (linkFrom)
        for source, link := range vu.Unit.LinksFrom {
            edgeKey := source + "->" + path
            if seen[edgeKey] {
                continue
            }
            seen[edgeKey] = true

            // Only include if source is in view
            if _, exists := v.Units[source]; !exists {
                continue
            }

            edge := &Edge{
                Source: source,
                Target: path,
                Label: &EdgeLabel{
                    Technology:  link.Technology,
                    Description: link.Description,
                    Position:    string(link.LabelPosition),
                },
                Style:     link.Style,
                ArrowHead: ArrowDirection(link.Arrow),
            }

            edges = append(edges, edge)
        }
    }

    return edges
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Gonum/graph for all graph work | Custom structs for view/graph | Phase 3 design | Simpler code, no external dependency |
| Inheritance-based styling | Independent unit styling | Phase 3 decisions | More explicit, less magical |

**Deprecated/outdated:**
- Using global `properties.expanded` defaults: Per decisions, only per-unit `expanded` attribute is used
- Combining bidirectional edges: Keep them separate per decisions

## Open Questions

1. **Should view and graph packages be separate or combined?**
   - What we know: They have different responsibilities (scoping vs. structure building)
   - What's unclear: If combined package is simpler for this codebase size
   - Recommendation: Keep separate for clarity and testability

2. **How to handle legend generation?**
   - What we know: Legend shows all unit types with shapes
   - What's unclear: Exact positioning and formatting (Claude's discretion)
   - Recommendation: Defer to Phase 4, but prepare data structure in graph package

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing package + stretchr/testify |
| Config file | None - standard Go test pattern |
| Quick run command | `go test ./internal/view/... ./internal/graph/... -v` |
| Full suite command | `go test ./... -cover` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| VIEW-01 | C1 view generation | unit | `go test ./internal/view/... -run TestGenerateC1View -v` | Wave 0 |
| VIEW-02 | C2 view generation | unit | `go test ./internal/view/... -run TestGenerateC2View -v` | Wave 0 |
| VIEW-03 | C3 view generation | unit | `go test ./internal/view/... -run TestGenerateC3View -v` | Wave 0 |
| VIEW-04 | Collapsed units as single nodes | unit | `go test ./internal/graph/... -run TestBuildNode -v` | Wave 0 |
| VIEW-05 | Expanded units as clusters | unit | `go test ./internal/graph/... -run TestBuildCluster -v` | Wave 0 |
| VIEW-06 | Expanded list respected | unit | `go test ./internal/view/... -run TestExpanded -v` | Wave 0 |
| VIEW-07 | Independent styling | unit | `go test ./internal/graph/... -run TestNodeStyle -v` | Wave 0 |
| GRPH-01 | Nodes with type-appropriate shapes | unit | `go test ./internal/graph/... -run TestShapeForType -v` | Wave 0 |
| GRPH-02 | Edges for links | unit | `go test ./internal/graph/... -run TestBuildEdges -v` | Wave 0 |
| GRPH-03 | Edge routing style | unit | `go test ./internal/graph/... -run TestEdgeStyle -v` | Wave 0 |
| GRPH-04 | Clusters for expanded units | unit | `go test ./internal/graph/... -run TestCluster -v` | Wave 0 |
| GRPH-05 | Icons for person/db/queue | unit | `go test ./internal/graph/... -run TestIconForType -v` | Wave 0 |
| GRPH-06 | System shape with name/tech/desc | unit | `go test ./internal/graph/... -run TestNodeLabel -v` | Wave 0 |
| QUAL-01 | Lint errors fixed | lint | `golangci-lint run` | Existing |
| QUAL-04 | 75% coverage | coverage | `go test ./... -cover` | Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/view/... ./internal/graph/... -v`
- **Per wave merge:** `go test ./... -cover`
- **Phase gate:** Full suite green + 75% coverage before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/view/view.go` - View struct and types
- [ ] `internal/view/view_test.go` - View generation tests
- [ ] `internal/view/scope.go` - C1/C2/C3 scoping logic
- [ ] `internal/view/scope_test.go` - Scoping tests
- [ ] `internal/graph/graph.go` - Graph, Node, Edge, Cluster structs
- [ ] `internal/graph/graph_test.go` - Graph construction tests
- [ ] `internal/graph/builder.go` - BuildGraph function
- [ ] `internal/graph/builder_test.go` - Builder tests
- [ ] `internal/graph/shapes.go` - Shape/icon mapping
- [ ] `internal/graph/shapes_test.go` - Shape mapping tests

## Sources

### Primary (HIGH confidence)
- Existing codebase: `internal/model/`, `internal/validator/`, `internal/parser/` - Verified patterns and types
- CONTEXT.md decisions - Locked implementation choices from user discussion

### Secondary (MEDIUM confidence)
- GraphViz DOT format knowledge - Based on official documentation patterns (https://graphviz.org/doc/info/lang.html)
- C4 model patterns - Based on C4-PlantUML conventions

### Tertiary (LOW confidence)
- None - All findings verified against existing codebase or locked decisions

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Using existing dependencies, no new libraries needed
- Architecture: HIGH - Patterns established by existing codebase (BuildIndex, model types)
- Pitfalls: HIGH - Derived from locked decisions in CONTEXT.md

**Research date:** 2026-03-09
**Valid until:** 30 days - Go patterns and C4 model are stable
