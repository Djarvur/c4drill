# Phase 8: All-Expanded Mode - Research

**Researched:** 2026-03-11
**Domain:** GraphViz nested clusters, recursive model traversal, CLI flag integration
**Confidence:** HIGH

## Summary

Phase 8 adds an `--expanded` flag that generates a single diagram showing all units at all C4 levels (C1, C2, C3) as nested clusters with cross-level edges visible. The implementation requires: (1) a new CLI flag in `cmd/c4drill/root.go`, (2) a new view generator `GenerateAllExpandedView()` in `internal/view/scope.go` that recursively includes all units with proper nesting, (3) modifications to `internal/graph/builder.go` to support deeply nested clusters, and (4) an updated output path for the expanded file.

GraphViz supports nested clusters natively through subgraphs with `cluster_` prefix. Edges between nodes in different clusters (including at different nesting depths) work automatically with `compound=true` graph setting. The existing `goccy/go-graphviz` library already supports nested subgraphs via `CreateSubGraphByName()`.

**Primary recommendation:** Create `GenerateAllExpandedView()` that builds a recursive view structure with all units expanded. Extend `Cluster` struct to support nested clusters, then leverage existing rendering pipeline unchanged.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **Expanded view structure**: Use modified C1 view approach, not a standalone GenerateAllExpandedView() function
- **Reuse existing View struct**: No new fields needed in View struct
- **Include all units**: Regardless of the expanded attribute
- **Location**: Add to internal/view/scope.go alongside existing view generators
- **Cross-level edge display**: Direct edges between units regardless of nesting depth, same visual style for all edges
- **Nested nodes approach**: Use GraphViz edge routing across cluster boundaries
- **External boundary nodes**: Always include for linked units outside the model
- **Output file strategy**: Replace normal output when --expanded is used (only create `basename.expanded.{ext}`, skip standard C1/C2/C3 files)
- **Format flag**: Share the same `--format` flag (dot/svg)
- **Output location**: Write to output directory root, same level as standard C1 output
- **Single file only**: No subdirectory structure
- **Flag name**: `--expanded` (boolean flag, simple on/off)
- **Combine paths**: use `--output` directory, append `.expanded.{ext}` to basename
- **Title**: Use standard title (properties.name) same as normal diagrams
- **Performance**: No size limits or warnings, silent by default, same GraphViz optimization settings
- **Error handling**: Standard error handling if graphviz fails due to complexity

### Claude's Discretion
- Exact GraphViz layout settings for nested clusters
- Internal implementation details of how to build nested clusters in graph package
- How to traverse model to include all units with proper nesting hierarchy
- Edge routing optimization details in graph building

### Deferred Ideas (OUT OF SCOPE)
None
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| EXPD-01 | User can pass `--expanded` flag to CLI | Cobra `PersistentFlags().Bool()` pattern in root.go:61 |
| EXPD-02 | All-expanded view renders all units as expanded nested clusters | GraphViz nested clusters with `cluster_` prefix; recursive traversal pattern from scope.go |
| EXPD-03 | Cross-level edges visible | GraphViz handles edges between clusters automatically with `compound=true` |
| EXPD-04 | Output saved to `{basename}.expanded.{ext}` | Writer.Write() pattern in writer.go; new method for expanded path |
| EXPD-05 | Existing C1/C2/C3 view generation unchanged | Separate code path in runRoot(); no modification to existing generators |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/spf13/cobra | v1.10.2 | CLI framework | Already used for flag registration and command structure |
| github.com/goccy/go-graphviz | v0.2.10 | GraphViz binding | Pure Go, WASM-based, supports nested subgraphs |
| github.com/pelletier/go-toml/v2 | v2.2.4 | TOML parsing | Already used for input parsing |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/stretchr/testify | v1.11.1 | Testing | Unit tests for new view generator and graph builder |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Nested Cluster struct | Flat cluster list with parent IDs | Parent IDs require more complex rendering; nested struct is cleaner |
| Separate AllExpanded renderer | Reuse BuildGraph() | Same graph structure works; only view generation differs |

## Architecture Patterns

### Recommended Project Structure
```
internal/
├── view/
│   └── scope.go           # Add GenerateAllExpandedView()
├── graph/
│   ├── graph.go           # Extend Cluster to support nested clusters
│   └── builder.go         # Add buildNestedCluster() for recursive cluster building
├── output/
│   └── writer.go          # Add WriteExpanded() for {basename}.expanded.{ext}
cmd/c4drill/
└── root.go                # Add --expanded flag, modify runRoot() flow
```

### Pattern 1: Recursive Unit Traversal
**What:** Walk the entire unit tree depth-first, collecting all units with their full paths
**When to use:** Building all-expanded view that includes every unit at every level
**Example:**
```go
func collectAllUnits(m *parser.Model) map[string]*Entry {
    units := make(map[string]*Entry)
    for name, unit := range m.Units {
        collectUnitRecursive(units, name, unit)
    }
    return units
}

func collectUnitRecursive(units map[string]*Entry, path string, unit *model.Unit) {
    units[path] = &Entry{
        Unit:        unit,
        FullPath:    path,
        IsExpanded:  len(unit.Subunits) > 0, // Always expanded if has subunits
        HasSubunits: len(unit.Subunits) > 0,
        IsExternal:  IsExternalType(unit.Type),
    }
    for subName, subUnit := range unit.Subunits {
        collectUnitRecursive(units, path+"."+subName, subUnit)
    }
}
```

### Pattern 2: Nested Cluster Structure
**What:** Extend Cluster to contain child Clusters for deeply nested structures
**When to use:** Rendering all-expanded view with units at multiple nesting depths
**Example:**
```go
type Cluster struct {
    ID       string
    Label    *Label
    Nodes    []*Node
    Clusters []*Cluster  // NEW: nested clusters
    Style    *NodeStyle
}
```

### Pattern 3: CLI Flag Integration
**What:** Add boolean flag alongside existing flags, modify control flow
**When to use:** Adding --expanded mode to existing CLI
**Example:**
```go
var expanded bool

func NewRootCmd() *cobra.Command {
    cmd := &cobra.Command{...}
    cmd.PersistentFlags().BoolVar(&expanded, "expanded", false, "Generate all-expanded diagram")
    // ... existing flags
    return cmd
}

func runRoot(cmd *cobra.Command, args []string) error {
    // ... parse and validate
    
    if expanded {
        return processExpandedView(m, basename, writer)
    }
    
    // ... existing C1/C2/C3 flow
}
```

### Anti-Patterns to Avoid
- **Creating separate View struct variant**: Locked decision says reuse existing View struct
- **Modifying existing view generators**: Must keep C1/C2/C3 unchanged for EXPD-05
- **Adding cluster depth limit**: Locked decision says no size limits or warnings
- **Special edge styling for cross-level**: Locked decision says same visual style for all edges

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Nested cluster rendering | Custom DOT generation | goccy/go-graphviz CreateSubGraphByName() | Library handles cluster nesting, edge routing |
| Cross-cluster edges | Manual edge coordinate calculation | Standard edge creation between nodes | GraphViz routes edges across cluster boundaries automatically |
| CLI flag parsing | Manual argument parsing | Cobra PersistentFlags | Already integrated, consistent with existing flags |
| File path construction | String concatenation | filepath.Join() | Platform-independent path handling |

**Key insight:** GraphViz's `compound=true` setting and subgraph nesting handles all complexity of cross-cluster edges. The Go library already exposes this through `CreateSubGraphByName()`.

## Common Pitfalls

### Pitfall 1: Edge Visibility in Nested Clusters
**What goes wrong:** Edges between deeply nested nodes don't appear
**Why it happens:** Missing `compound=true` graph attribute
**How to avoid:** Ensure `compound=true` is set on the root graph (already set in current implementation via SetCompound if needed)
**Warning signs:** Edges shown in debug but missing in rendered output

### Pitfall 2: Cluster Naming Conflicts
**What goes wrong:** Clusters with same ID overwrite each other
**Why it happens:** Using simple names like "cluster_api" when multiple systems have "api" subunit
**How to avoid:** Use full dotted path in cluster ID: `cluster_mainapp.api` not `cluster_api`
**Warning signs:** Missing clusters in output, unexpected nesting

### Pitfall 3: External Boundary Node Duplication
**What goes wrong:** Same external unit appears multiple times
**Why it happens:** External boundary nodes created per-link rather than deduplicated
**How to avoid:** Use existing `addExternalBoundaryNodes()` pattern which checks for existence before adding
**Warning signs:** Multiple identical external nodes in diagram

### Pitfall 4: Forgetting to Skip Normal Output
**What goes wrong:** Both expanded and normal C1/C2/C3 files generated
**Why it happens:** Not returning early from runRoot() after expanded generation
**How to avoid:** Use early return pattern: `if expanded { return processExpanded(...); }`
**Warning signs:** Extra files in output directory

## Code Examples

### GenerateAllExpandedView Implementation Pattern
```go
func GenerateAllExpandedView(m *parser.Model) *View {
    if m == nil {
        return nil
    }
    
    v := &View{
        Level: LevelC1,
        Title: m.Properties.Name,
        Edges: m.Properties.Edges,
        Units: make(map[string]*Entry),
    }
    
    // Recursively add all units
    for name, unit := range m.Units {
        addUnitRecursive(v, name, unit)
    }
    
    addExternalBoundaryNodes(v, m)
    return v
}

func addUnitRecursive(v *View, path string, unit *model.Unit) {
    v.Units[path] = &Entry{
        Unit:        unit,
        FullPath:    path,
        IsExpanded:  len(unit.Subunits) > 0, // Always show as expanded if has children
        HasSubunits: len(unit.Subunits) > 0,
        IsExternal:  IsExternalType(unit.Type),
    }
    
    for subName, subUnit := range unit.Subunits {
        addUnitRecursive(v, path+"."+subName, subUnit)
    }
}
```

### Nested Cluster Building Pattern
```go
func buildNestedCluster(entry *view.Entry, path string) *Cluster {
    cluster := &Cluster{
        ID:       "cluster_" + path,
        Label:    buildClusterLabel(entry),
        Nodes:    make([]*Node, 0),
        Clusters: make([]*Cluster, 0), // For nested clusters
        Style:    GetStyleForType(entry.Unit.Type, entry.IsExternal),
    }
    
    for childName, childUnit := range entry.Unit.Subunits {
        childPath := path + "." + childName
        childEntry := &view.Entry{
            Unit:        childUnit,
            FullPath:    childPath,
            IsExpanded:  len(childUnit.Subunits) > 0,
            HasSubunits: len(childUnit.Subunits) > 0,
            IsExternal:  view.IsExternalType(childUnit.Type),
        }
        
        if childEntry.IsExpanded {
            // Recursively build nested cluster
            cluster.Clusters = append(cluster.Clusters, buildNestedCluster(childEntry, childPath))
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

### GraphViz Nested Cluster Rendering
```go
func createClusterRecursive(parent *cgraph.Graph, cluster *graph.Cluster, nodeMap map[string]*cgraph.Node) error {
    subgraph, err := parent.CreateSubGraphByName(cluster.ID)
    if err != nil {
        return err
    }
    
    subgraph.SetLabel(cluster.Label.Name)
    subgraph.SetStyle(cgraph.FilledGraphStyle)
    
    if cluster.Style != nil && cluster.Style.FillColor != "" {
        subgraph.SetBackgroundColor(cluster.Style.FillColor)
    }
    
    // Create nodes in this cluster
    for _, node := range cluster.Nodes {
        cn, err := createNode(subgraph, node)
        if err != nil {
            return err
        }
        nodeMap[node.ID] = cn
    }
    
    // Recursively create nested clusters
    for _, nestedCluster := range cluster.Clusters {
        if err := createClusterRecursive(subgraph, nestedCluster, nodeMap); err != nil {
            return err
        }
    }
    
    return nil
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| GraphViz native binary dependency | WASM-embedded GraphViz | goccy/go-graphviz v0.2+ | No external dependencies needed |
| Flat cluster list | Nested cluster structure | This phase | Enables arbitrary nesting depth |

**Deprecated/outdated:**
- Calling external `dot` binary: Use goccy/go-graphviz library instead

## Open Questions

1. **Should BuildGraph() be extended or should we create BuildAllExpandedGraph()?**
   - What we know: CONTEXT.md says use modified C1 view approach
   - What's unclear: Whether modification happens in BuildGraph or a new function
   - Recommendation: Extend BuildGraph() to handle nested clusters via recursive cluster building; cleaner than duplicating logic

2. **How to handle the Cluster struct extension?**
   - What we know: Current Cluster has Nodes []*Node
   - What's unclear: Should we add Clusters []*Cluster field or create new NestedCluster type
   - Recommendation: Add Clusters []*Cluster to existing Cluster struct; maintains backward compatibility, existing code ignores the new field

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | stretchr/testify v1.11.1 |
| Config file | None — Go testing convention |
| Quick run command | `go test ./internal/view/... -v -run TestGenerateAllExpanded` |
| Full suite command | `go test ./... -v` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| EXPD-01 | --expanded flag accepted | unit | `go test ./cmd/c4drill/... -run TestExpandedFlag` | ❌ Wave 0 |
| EXPD-02 | All units appear as nested clusters | unit | `go test ./internal/view/... -run TestGenerateAllExpandedView` | ❌ Wave 0 |
| EXPD-03 | Cross-level edges visible | integration | `go test ./internal/graph/... -run TestCrossLevelEdges` | ❌ Wave 0 |
| EXPD-04 | Output file naming correct | unit | `go test ./internal/output/... -run TestWriteExpanded` | ❌ Wave 0 |
| EXPD-05 | C1/C2/C3 unchanged | regression | `go test ./...` (existing tests must pass) | ✅ Existing |

### Sampling Rate
- **Per task commit:** `go test ./internal/view/... ./internal/graph/... ./cmd/c4drill/... -v`
- **Per wave merge:** `go test ./... -v`
- **Phase gate:** Full suite green before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] `internal/view/scope_test.go` — add TestGenerateAllExpandedView* tests (EXPD-02)
- [ ] `internal/graph/builder_test.go` — add TestBuildNestedCluster tests (EXPD-03)
- [ ] `internal/output/writer_test.go` — add TestWriteExpanded tests (EXPD-04)
- [ ] `cmd/c4drill/root_test.go` — add TestExpandedFlag tests (EXPD-01)
- [ ] Test fixture: Create `testdata/nested.toml` with 3+ levels of nesting

## Sources

### Primary (HIGH confidence)
- GraphViz documentation - https://graphviz.org/doc/info/lang.html (cluster syntax, compound attribute)
- GraphViz gallery - https://graphviz.org/Gallery/directed/cluster.html (nested cluster examples)
- goccy/go-graphviz source - https://github.com/goccy/go-graphviz (CreateSubGraphByName API)
- Project source code - internal/view/scope.go, internal/graph/builder.go, cmd/c4drill/root.go

### Secondary (MEDIUM confidence)
- Web search verified with official GraphViz docs - nested subgraph and cross-cluster edge behavior

### Tertiary (LOW confidence)
None - all findings verified against primary sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - All libraries already in use in the project
- Architecture: HIGH - Existing patterns well-documented in codebase, GraphViz nested clusters well-understood
- Pitfalls: HIGH - Based on GraphViz documentation and existing codebase patterns

**Research date:** 2026-03-11
**Valid until:** 2026-04-11 (stable libraries, well-established GraphViz patterns)
