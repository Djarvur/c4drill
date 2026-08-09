# Architecture Integration: C4Drill v1.1 AI-Ready

**Domain:** CLI tool for C4 architecture diagrams
**Researched:** 2026-03-10
**Confidence:** HIGH (based on direct codebase analysis of v1.0 implementation)

## Overview

This document describes how the two v1.1 features integrate with the existing v1.0 architecture:

1. **TOML Language Manual** - AI-focused documentation for TOML authoring
2. **All-Expanded Mode** - Single-view rendering with all units expanded

---

## Existing v1.0 Architecture Reference

The v1.0 architecture follows a clean pipeline pattern (see ARCHITECTURE.md for full details):

```
TOML File
    |
    v
+-------------+     +-------------+     +-------------+
|   parser    | --> |  validator  | --> |    view     |
| (TOML->Model)|    | (rules)     |    | (scope gen) |
+-------------+     +-------------+     +-------------+
                                              |
                                              v
+-------------+     +-------------+     +-------------+
|   output    | <-- |   render    | <-- |   graph     |
| (file write)|     | (DOT/SVG)   |     | (builder)   |
+-------------+     +-------------+     +-------------+
```

**Key packages (all under `internal/`):**
- `parser` - TOML parsing via `pelletier/go-toml/v2`
- `validator` - Reference integrity, type rules, subunit constraints
- `view` - C1/C2/C3 view generation with `view.Entry` structs
- `graph` - Builds `graph.Graph` with nodes/edges/clusters
- `render` - Converts graph to DOT/SVG via `goccy/go-graphviz`
- `output` - File writing with directory structure

---

## Feature 1: TOML Language Manual

### Description

A documentation file (CLAUDE.md) serving as:
1. AI prompt context for generating C4Drill TOML files
2. Human reference for TOML authoring

### Integration Point: NEW DOCUMENTATION

**Location:** Project root as `CLAUDE.md`

**Rationale:**
- Not code - pure documentation
- Must be discoverable at repository root for AI tools
- Follows emerging convention (CLAUDE.md for Claude-specific context)
- No code dependencies or modifications required

**Structure:**

```
CLAUDE.md
├── Purpose statement
├── TOML schema reference
│   ├── [properties] section
│   ├── Unit types and their attributes
│   ├── Link syntax (link, linkFrom)
│   └── Nested subunit syntax
├── Validation rules (what will fail)
├── Examples by use case
│   ├── Simple C1 diagram
│   ├── Multi-level with drill-down
│   └── Complex nested system
└── AI prompt templates
    ├── "Generate a C4 diagram for..."
    └── Error recovery patterns
```

**Build Order Impact:** NONE - Documentation only, can be written independently.

---

## Feature 2: All-Expanded Rendering Mode

### Description

A new `--expanded` CLI flag that renders a single diagram showing:
1. ALL units expanded to their deepest level
2. ALL cross-level edges (edges between units at different nesting depths)
3. Output to `{basename}.expanded.{ext}`

### Integration Points

#### 2.1 CLI Layer: MODIFY `cmd/c4drill/root.go`

**New flag:**

```go
var expanded bool  // Add to existing flag variables

// In NewRootCmd():
cmd.PersistentFlags().BoolVarP(&expanded, "expanded", "e", false,
    "Render all units expanded with cross-level edges")
```

**Pipeline modification in `runRoot()`:**

```go
// After validation, before view generation
if expanded {
    return processExpandedView(m, basename, writer)
}
// Existing path for normal multi-view rendering continues unchanged
```

**New function `processExpandedView()`:**

```go
func processExpandedView(m *parser.Model, basename string, writer *output.Writer) error {
    v := view.GenerateExpandedView(m)
    if v == nil {
        return fmt.Errorf("%w: expanded view", errGenerateView)
    }

    g := graph.BuildExpandedGraph(v)
    if g == nil {
        return fmt.Errorf("%w: expanded graph", errBuildGraph)
    }

    data, err := render.Render(g, format)
    if err != nil {
        return fmt.Errorf("render: %w", err)
    }

    return writer.WriteExpanded(basename, format, data)
}
```

#### 2.2 View Layer: NEW FILE `internal/view/expanded.go`

**New constant in `internal/view/view.go`:**

```go
const (
    LevelC1 Level = iota
    LevelC2
    LevelC3
    LevelExpanded  // NEW
)
```

**New function in `internal/view/expanded.go`:**

```go
// GenerateExpandedView creates a single view with ALL units expanded.
// Unlike C1/C2/C3 views which scope to a level, this recursively
// includes all subunits at all depths.
func GenerateExpandedView(m *parser.Model) *View
```

**Key differences from existing views:**

| Aspect | C1/C2/C3 Views | Expanded View |
|--------|---------------|---------------|
| Units included | Scoped to level | All units at all depths |
| Clustering | Single level | Nested clusters (recursive) |
| Edge scope | Within view only | Cross-level edges included |
| Navigation | Back-links, breadcrumbs | None (single view) |
| Entry.IsExpanded | Based on `expanded` attr | Always true if HasSubunits |

**Implementation approach:**

```go
func GenerateExpandedView(m *parser.Model) *View {
    v := &View{
        Level: LevelExpanded,
        Title: m.Properties.Name + " (Expanded)",
        Edges: m.Properties.Edges,
        Units: make(map[string]*Entry),
    }

    // Recursively add all units
    for name, unit := range m.Units {
        addExpandedUnits(v, name, unit, "")
    }

    return v
}

func addExpandedUnits(v *View, name string, unit *model.Unit, parentPath string) {
    fullPath := name
    if parentPath != "" {
        fullPath = parentPath + "." + name
    }

    v.Units[fullPath] = &Entry{
        Unit:        unit,
        FullPath:    fullPath,
        IsExpanded:  len(unit.Subunits) > 0,  // Always expanded if has subunits
        HasSubunits: len(unit.Subunits) > 0,
        IsExternal:  IsExternalType(unit.Type),
    }

    // Recurse into subunits
    for subName, subUnit := range unit.Subunits {
        addExpandedUnits(v, subName, subUnit, fullPath)
    }
}
```

#### 2.3 Graph Layer: MODIFY `internal/graph/graph.go`

**Cluster struct needs nested clusters:**

```go
type Cluster struct {
    ID       string
    Label    *Label
    Nodes    []*Node
    Clusters []*Cluster  // NEW: support nested clusters
    Style    *NodeStyle
}
```

#### 2.4 Graph Layer: MODIFY `internal/graph/builder.go`

**Option A: New function `BuildExpandedGraph()`**

```go
// BuildExpandedGraph constructs a graph with nested clusters for expanded view.
func BuildExpandedGraph(v *view.View) *Graph {
    g := &Graph{
        Title:     v.Title,
        Direction: "TB",
        EdgeStyle: v.Edges,
        Nodes:     make([]*Node, 0),
        Edges:     make([]*Edge, 0),
        Clusters:  make([]*Cluster, 0),
    }

    // Build nested clusters recursively
    for _, entry := range v.Units {
        // Only process top-level units (no parent)
        if !strings.Contains(entry.FullPath, ".") {
            if entry.HasSubunits {
                cluster := buildNestedCluster(entry, v)
                g.Clusters = append(g.Clusters, cluster)
            } else {
                node := buildNode(entry)
                g.Nodes = append(g.Nodes, node)
            }
        }
    }

    // Build cross-level edges
    g.Edges = buildAllEdges(v)

    return g
}

func buildNestedCluster(entry *view.Entry, v *view.View) *Cluster {
    cluster := &Cluster{
        ID:       "cluster_" + entry.FullPath,
        Label:    buildClusterLabel(entry),
        Nodes:    make([]*Node, 0),
        Clusters: make([]*Cluster, 0),
        Style:    GetStyleForType(entry.Unit.Type, entry.IsExternal),
    }

    for childName, childUnit := range entry.Unit.Subunits {
        childPath := entry.FullPath + "." + childName
        childEntry := v.Units[childPath]  // Look up from view

        if childEntry.HasSubunits {
            // Recursively build nested cluster
            nestedCluster := buildNestedCluster(childEntry, v)
            cluster.Clusters = append(cluster.Clusters, nestedCluster)
        } else {
            // Build leaf node
            node := buildNode(childEntry)
            node.IsInCluster = true
            cluster.Nodes = append(cluster.Nodes, node)
        }
    }

    return cluster
}
```

**Edge handling for cross-level edges:**

```go
// buildAllEdges creates edges for all links regardless of depth.
func buildAllEdges(v *view.View) []*Edge {
    edges := make([]*Edge, 0)
    seen := make(map[string]bool)

    for path, entry := range v.Units {
        // Process outgoing links
        for target, link := range entry.Unit.Links {
            // Target may be at any depth; look up in view
            if _, exists := v.Units[target]; exists {
                edge := createEdge(path, target, link)
                edgeKey := path + "->" + target + ":" + link.Technology
                if markSeen(seen, edgeKey) {
                    edges = append(edges, edge)
                }
            }
        }

        // Process incoming links
        for source, link := range entry.Unit.LinksFrom {
            if _, exists := v.Units[source]; exists {
                edge := createEdge(source, path, link)
                edgeKey := source + "->" + path + ":" + link.Technology
                if markSeen(seen, edgeKey) {
                    edges = append(edges, edge)
                }
            }
        }
    }

    return edges
}
```

#### 2.5 Render Layer: MODIFY `internal/render/converter.go`

**createCluster() needs recursion:**

```go
func createCluster(parent *cgraph.Graph, cluster *graph.Cluster, nodeMap map[string]*cgraph.Node) error {
    subgraph, err := parent.CreateSubGraphByName("cluster_" + cluster.ID)
    if err != nil {
        return fmt.Errorf("create subgraph: %w", err)
    }

    // ... existing styling code (SetLabel, SetStyle, etc.) ...

    // Create leaf nodes
    for _, node := range cluster.Nodes {
        cn, err := createNode(subgraph, node)
        if err != nil {
            return fmt.Errorf("create node %s in cluster: %w", node.ID, err)
        }
        nodeMap[node.ID] = cn
    }

    // NEW: Recursively create nested clusters
    for _, nested := range cluster.Clusters {
        if err := createCluster(subgraph, nested, nodeMap); err != nil {
            return fmt.Errorf("create nested cluster %s: %w", nested.ID, err)
        }
    }

    return nil
}
```

#### 2.6 Output Layer: MODIFY `internal/output/writer.go`

**New method:**

```go
// WriteExpanded writes the expanded view output.
// Output path: {basename}.expanded.{format}
func (w *Writer) WriteExpanded(basename, format string, data []byte) error {
    relPath := fmt.Sprintf("%s.expanded.%s", basename, format)
    fullPath := filepath.Join(w.baseDir, relPath)

    dir := filepath.Dir(fullPath)
    if err := os.MkdirAll(dir, dirPermission); err != nil {
        return fmt.Errorf("create output directory %s: %w", dir, err)
    }

    if err := os.WriteFile(fullPath, data, filePermission); err != nil {
        return fmt.Errorf("write output file %s: %w", fullPath, err)
    }

    return nil
}
```

---

## Component Integration Summary

### New Components

| Component | Location | Purpose |
|-----------|----------|---------|
| TOML Language Manual | `CLAUDE.md` (root) | AI prompt + human reference |
| Expanded view generator | `internal/view/expanded.go` | Generate all-units-expanded view |

### Modified Components

| Component | File | Changes | Complexity |
|-----------|------|---------|------------|
| CLI | `cmd/c4drill/root.go` | Add `--expanded` flag, `processExpandedView()` | Low |
| View types | `internal/view/view.go` | Add `LevelExpanded` constant | Trivial |
| Graph types | `internal/graph/graph.go` | Add `Clusters []*Cluster` to `Cluster` struct | Low |
| Graph builder | `internal/graph/builder.go` | Add `BuildExpandedGraph()`, `buildNestedCluster()`, `buildAllEdges()` | Medium |
| Render converter | `internal/render/converter.go` | Modify `createCluster()` for recursion | Medium |
| Output writer | `internal/output/writer.go` | Add `WriteExpanded()` method | Low |

### Data Flow for Expanded Mode

```
TOML File
    |
    v
+-------------+     +-------------+
|   parser    | --> |  validator  |
+-------------+     +-------------+
                          |
          +---------------+---------------+
          |                               |
          v (normal)                      v (--expanded)
    Multi-view path                 +-------------+
    (existing)                      | view        |
                                    | (expanded)  |
                                    +-------------+
                                          |
                                          v
                                    +-------------+
                                    | graph       |
                                    | (nested)    |
                                    +-------------+
                                          |
                                          v
                                    +-------------+
                                    | render      |
                                    +-------------+
                                          |
                                          v
                                    +-------------+
                                    | output      |
                                    | .expanded.  |
                                    +-------------+
```

---

## Suggested Build Order

Based on dependency analysis:

### Phase 1: Documentation (No Dependencies)
1. **CLAUDE.md** - Create TOML Language Manual
   - Can be done in parallel with code changes
   - Validates understanding of schema

### Phase 2: View Layer (Core Logic)
2. **internal/view/view.go** - Add `LevelExpanded` constant (1 line)
3. **internal/view/expanded.go** - Implement `GenerateExpandedView()` with `addExpandedUnits()`
   - Unit tests for recursive unit collection
   - Test with deeply nested TOML

### Phase 3: Graph Layer (Structural Changes)
4. **internal/graph/graph.go** - Add `Clusters []*Cluster` field to `Cluster` struct
5. **internal/graph/builder.go** - Implement `BuildExpandedGraph()`
6. **internal/graph/builder.go** - Implement `buildNestedCluster()` with recursion
7. **internal/graph/builder.go** - Implement `buildAllEdges()` for cross-level edges
   - Unit tests for nested cluster structure
   - Unit tests for edge resolution at all depths

### Phase 4: Render Layer (Output Changes)
8. **internal/render/converter.go** - Modify `createCluster()` for recursive handling
   - Integration tests with nested clusters
   - Visual verification of SVG output

### Phase 5: Output + CLI (Integration)
9. **internal/output/writer.go** - Add `WriteExpanded()` method
10. **cmd/c4drill/root.go** - Add `--expanded` flag and `processExpandedView()`
    - End-to-end tests for expanded mode

### Phase 6: Integration Testing
11. **Full pipeline tests** - Complex nested TOML with cross-level edges
    - Verify nested cluster rendering in SVG
    - Verify all edges appear correctly
    - Verify output file naming (`{basename}.expanded.{ext}`)

---

## Anti-Patterns to Avoid

### Anti-Pattern 1: Duplicating View Logic
**What:** Copy-pasting C1/C2/C3 generation code for expanded view
**Why bad:** Maintenance burden, drift between code paths
**Instead:** Extract common unit-to-entry conversion to shared helper (e.g., `entryFromUnit()`)

### Anti-Pattern 2: Flat Edge List for Nested Structures
**What:** Treating all edges as flat when units are deeply nested
**Why bad:** GraphViz creates confusing edge routing, crossing cluster boundaries incorrectly
**Instead:** Ensure edge endpoints reference the correct node IDs at their actual depth

### Anti-Pattern 3: Ignoring GraphViz Cluster Depth Limits
**What:** Assuming unlimited nesting depth works
**Why bad:** GraphViz has practical limits; deeply nested clusters may render poorly
**Instead:** Test with realistic depth (3-4 levels), consider warning if exceeded

### Anti-Pattern 4: Missing Edge Cases in Cross-Level Links
**What:** Only handling links at the same nesting level
**Why bad:** Cross-level edges (e.g., component to external system) won't render
**Instead:** `buildAllEdges()` must look up targets at any depth in the view

---

## Scalability Considerations

| Concern | Small Model (10 units) | Medium (50 units) | Large (200+ units) |
|---------|------------------------|-------------------|---------------------|
| Memory | Trivial | Moderate | May need optimization |
| SVG file size | ~10KB | ~100KB | May reach MB range |
| Layout quality | Good | Good | May need manual hints |
| Rendering time | Instant | ~1s | Several seconds |

**Recommendation:** For large models, consider warning when expanded mode exceeds threshold.

---

## Sources

- Direct codebase analysis of `/Users/nil/DiskD/W/Djarvur/c4drill/internal/`
- PROJECT.md specification for v1.1 milestone requirements
- Test data examples in `testdata/` directory

---
*Architecture integration research for: C4Drill v1.1 AI-Ready*
*Researched: 2026-03-10*
