# Phase 13: Refined HTML Labels - Research

**Researched:** 2026-03-13
**Domain:** Graph rendering bug fix and HTML label refinements
**Confidence:** HIGH

## Summary

This phase addresses two distinct concerns:

**Part A (Critical Bug):** The expanded view graph builder (`BuildExpandedGraph`) in `internal/graph/builder.go` correctly processes nested clusters via `buildNestedCluster()`, but the `createCluster()` function in `internal/render/converter.go` does NOT recursively render nested clusters. It only iterates over `cluster.Nodes`, completely ignoring `cluster.Clusters`.

**Part B (Label Refinements):** Minor enhancements to HTML label rendering including shape changes, table attributes, and cluster label improvements.

**Primary recommendation:** Fix `createCluster()` in `converter.go` to recursively render nested clusters, then add the `Type` and `IsExternal` fields to the `Cluster` struct for proper HTML label dispatch.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Part A: Bug Fixes (Critical)**
Fix bugs in expanded view graph generation where nested containers are missing:
1. `server.pam` cluster not rendered
2. `server.pam.unix` and `server.pam.cyp` components missing
3. Links to/from nested components not rendered

**Part B: Label Refinements**
1. All units: `shape=box, style=rounded`
2. Table attributes: `border="0" cellpadding="0" cellspacing="0"`
3. Cluster labels use HTML format (same as corresponding unit)
4. Cluster labels use unit type coloring

### Claude's Discretion

None - all decisions were locked in CONTEXT.md.

### Deferred Ideas (OUT OF SCOPE)

None - discussion stayed within phase scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| BUG-01 | `server.pam` cluster not rendered in expanded view | Root cause: `createCluster()` in converter.go does not recurse into nested clusters |
| BUG-02 | `server.pam.unix` and `server.pam.cyp` components missing | Same root cause as BUG-01 |
| BUG-03 | Links to/from nested components not rendered | Edge building is correct; bug is nodes not existing due to missing clusters |
| TEST-01 | Add automated tests for DOT verification | Use existing test patterns with `strings.Contains` for substring assertions |
| REFINED-01 | All units use `shape=box, style=rounded` | Modify `createNode()` in converter.go |
| REFINED-02 | Table attributes for HTML labels | Modify all HTML label builders in labels.go |
| REFINED-03 | Cluster labels use HTML format with type coloring | Add `Type` and `IsExternal` to Cluster struct, modify `createCluster()` |
</phase_requirements>

## Root Cause Analysis

### BUG-01/02/03: Nested Clusters Not Rendered

**Location:** `internal/render/converter.go:224-270` (`createCluster` function)

**The Bug:**
```go
// Current implementation (lines 259-267)
for _, node := range cluster.Nodes {
    cn, err := createNode(subgraph, node)
    // ...
    nodeMap[node.ID] = cn
}
// MISSING: No iteration over cluster.Clusters!
```

**Root Cause:** The `Cluster` struct in `internal/graph/graph.go:97-109` correctly supports nested clusters:
```go
type Cluster struct {
    ID       string
    Label    *Label
    Nodes    []*Node
    Clusters []*Cluster  // <-- This exists!
    Style    *NodeStyle
}
```

And `buildNestedCluster()` in `internal/graph/builder.go:87-123` correctly populates `cluster.Clusters`:
```go
if childEntry.HasSubunits {
    nestedCluster := buildNestedCluster(childEntry, childPath, v)
    cluster.Clusters = append(cluster.Clusters, nestedCluster)
}
```

**But `createCluster()` in converter.go never iterates over `cluster.Clusters`**, so nested clusters are never rendered.

**Fix:** Add recursive cluster creation in `createCluster()`:
```go
// After creating nodes, recursively create nested clusters
for _, nestedCluster := range cluster.Clusters {
    if err := createCluster(subgraph, nestedCluster, nodeMap); err != nil {
        return fmt.Errorf("create nested cluster %s: %w", nestedCluster.ID, err)
    }
}
```

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| go-graphviz | v0.x | DOT/SVG rendering | Only Go library with full GraphViz support |
| stretchr/testify | v1.x | Assertions | Project standard for tests |

### Test Patterns
| Pattern | Location | When to Use |
|---------|----------|-------------|
| `strings.Contains(dotStr, "pattern")` | integration_test.go | Simple substring verification |
| Custom model builders | integration_test.go | Reusable test fixtures |
| `require.NotNil`, `assert.Contains` | All test files | Standard assertions |

## Architecture Patterns

### Recommended Fix Structure

```
internal/graph/
    graph.go          # Add Type, IsExternal to Cluster struct
    builder.go        # No changes needed (already correct)

internal/render/
    converter.go      # Fix createCluster() to recurse
    labels.go         # Add table attributes to HTML builders
```

### Pattern 1: Recursive Cluster Rendering

**What:** `createCluster()` must recursively call itself for nested clusters
**When:** Always - whenever a cluster has nested `Clusters`
**Example:**
```go
// Source: Fix for converter.go createCluster()
func createCluster(parent *cgraph.Graph, cluster *graph.Cluster, nodeMap map[string]*cgraph.Node) error {
    subgraph, err := parent.CreateSubGraphByName("cluster_" + cluster.ID)
    if err != nil {
        return fmt.Errorf("create subgraph: %w", err)
    }

    // ... set label and style ...

    // Create nodes inside cluster
    for _, node := range cluster.Nodes {
        cn, err := createNode(subgraph, node)
        if err != nil {
            return fmt.Errorf("create node %s in cluster: %w", node.ID, err)
        }
        nodeMap[node.ID] = cn
    }

    // FIX: Create nested clusters recursively
    for _, nestedCluster := range cluster.Clusters {
        if err := createCluster(subgraph, nestedCluster, nodeMap); err != nil {
            return fmt.Errorf("create nested cluster %s: %w", nestedCluster.ID, err)
        }
    }

    return nil
}
```

### Pattern 2: HTML Label Dispatch for Clusters

**What:** Clusters need `Type` and `IsExternal` fields to select correct HTML label builder
**When:** Creating cluster labels in `createCluster()`
**Example:**
```go
// Add to Cluster struct in graph.go
type Cluster struct {
    ID         string
    Label      *Label
    Nodes      []*Node
    Clusters   []*Cluster
    Style      *NodeStyle
    Type       model.UnitType  // NEW: For HTML label dispatch
    IsExternal bool            // NEW: For external styling
}

// Then in converter.go createCluster():
if cluster.Label != nil {
    htmlLabel := buildHTMLLabelForType(cluster.Label, cluster.Type)
    if htmlLabel != "" {
        htmlStr, err := parent.StrdupHTML(htmlLabel)
        if err != nil {
            return fmt.Errorf("create HTML cluster label: %w", err)
        }
        subgraph.SetLabel(htmlStr)
    }
}
```

### Anti-Patterns to Avoid

- **Forgetting recursive call:** Only processing `cluster.Nodes` and not `cluster.Clusters`
- **Infinite recursion:** Ensure recursion terminates at leaf clusters (those with empty `Clusters` slice)
- **Missing nodeMap updates:** Nested cluster nodes must be added to `nodeMap` for edge creation

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| DOT parsing in tests | Custom regex parser | `strings.Contains` | Simple substring checks sufficient for verification |
| Graph traversal | Custom walker | Existing recursive pattern | `buildNestedCluster()` already shows the pattern |

**Key insight:** The fix is a 5-line addition to existing code, not new infrastructure.

## Common Pitfalls

### Pitfall 1: Not Updating builder.go to Set Type on Clusters

**What goes wrong:** Adding `Type` field to Cluster struct but not setting it in `buildNestedCluster()`
**Why it happens:** Easy to forget the new field when cluster is created
**How to avoid:** Update `buildNestedCluster()` to include:
```go
cluster := &Cluster{
    ID:         "cluster_" + path,
    Label:      buildClusterLabel(entry),
    Nodes:      make([]*Node, 0),
    Clusters:   make([]*Cluster, 0),
    Style:      GetStyleForType(entry.Unit.Type, entry.IsExternal),
    Type:       entry.Unit.Type,     // NEW
    IsExternal: entry.IsExternal,    // NEW
}
```
**Warning signs:** Cluster labels use wrong color or format

### Pitfall 2: Edge Endpoints Missing After Cluster Fix

**What goes wrong:** Edges still not rendered even after cluster fix
**Why it happens:** `nodeMap` doesn't contain nodes from nested clusters
**How to avoid:** Ensure recursive `createCluster()` adds all nodes to `nodeMap`
**Warning signs:** `isTargetInView` returns false for valid nested unit paths

### Pitfall 3: GraphViz Cluster Naming Convention

**What goes wrong:** Clusters don't render as boxes
**Why it happens:** GraphViz requires subgraph names to start with "cluster_"
**How to avoid:** Always use `"cluster_" + cluster.ID` for subgraph names (already correct in code)
**Warning signs:** Clusters appear as regular nodes or not at all

## Code Examples

### Fix for createCluster() - Recursive Cluster Rendering

```go
// Source: internal/render/converter.go - createCluster()
// Add after the node creation loop (around line 267)

// Create nested clusters recursively
for _, nestedCluster := range cluster.Clusters {
    if err := createCluster(subgraph, nestedCluster, nodeMap); err != nil {
        return fmt.Errorf("create nested cluster %s: %w", nestedCluster.ID, err)
    }
}
```

### Fix for Cluster Struct - Add Type and IsExternal

```go
// Source: internal/graph/graph.go - Cluster struct
// Add these fields:

type Cluster struct {
    ID         string
    Label      *Label
    Nodes      []*Node
    Clusters   []*Cluster
    Style      *NodeStyle
    Type       model.UnitType  // NEW: Unit type for HTML label dispatch
    IsExternal bool            // NEW: For external node styling
}
```

### Fix for buildNestedCluster() - Set Type and IsExternal

```go
// Source: internal/graph/builder.go - buildNestedCluster()
// Update cluster creation (line 88-94):

cluster := &Cluster{
    ID:         "cluster_" + path,
    Label:      buildClusterLabel(entry),
    Nodes:      make([]*Node, 0),
    Clusters:   make([]*Cluster, 0),
    Style:      GetStyleForType(entry.Unit.Type, entry.IsExternal),
    Type:       entry.Unit.Type,     // NEW
    IsExternal: entry.IsExternal,    // NEW
}
```

### Fix for HTML Table Attributes

```go
// Source: internal/render/labels.go - All HTML label builders
// Change from:
sb.WriteString("<table>")
// To:
sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0">`)
```

### Fix for Node Shape

```go
// Source: internal/render/converter.go - createNode()
// Change from:
cn.SetShape(cgraph.NoneShape)
// To:
cn.SetShape(cgraph.BoxShape)
// Note: style=rounded is already set on line 213
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| `shape=none` for HTML labels | `shape=box, style=rounded` | Phase 13 | Better visual appearance |
| Simple cluster labels | HTML cluster labels | Phase 13 | Consistent styling with nodes |
| Flat cluster structure | Nested cluster structure | Phase 8 | Now needs recursive rendering |

**Deprecated/outdated:**
- `shape=none` for HTML labels: Use `shape=box, style=rounded` instead (per CONTEXT.md decision)

## Open Questions

1. **Should we also update `buildCluster()` (non-expanded) to set Type/IsExternal?**
   - What we know: `buildCluster()` is used for single-level expansion in C1/C2/C3 views
   - What's unclear: Whether clusters in non-expanded views also need HTML labels
   - Recommendation: Yes, for consistency. Add the same fields to `buildCluster()` output.

2. **Do we need to update `buildCluster()` in builder.go for nested clusters?**
   - What we know: `buildCluster()` doesn't handle nested clusters (single-level only)
   - What's unclear: If a unit is manually expanded but contains nested containers, should they also expand?
   - Recommendation: No - `buildCluster()` is for single-level expansion. Only `buildNestedCluster()` needs the fix.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + stretchr/testify |
| Config file | None - Go conventions |
| Quick run command | `go test ./internal/render/... -run TestExpanded -v` |
| Full suite command | `go test ./internal/... -v` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| BUG-01 | Nested cluster `server.pam` renders | integration | `go test ./internal/render/... -run TestExpandedViewNestedClusters -v` | Wave 0 |
| BUG-02 | Nested components appear in DOT | integration | `go test ./internal/render/... -run TestExpandedViewNestedComponents -v` | Wave 0 |
| BUG-03 | Edges to nested components render | integration | `go test ./internal/render/... -run TestExpandedViewNestedEdges -v` | Wave 0 |
| REFINED-01 | Nodes use shape=box, style=rounded | unit | `go test ./internal/render/... -run TestNodeShape -v` | Wave 0 |
| REFINED-02 | HTML tables have correct attributes | unit | `go test ./internal/render/... -run TestHTMLTableAttributes -v` | Wave 0 |
| REFINED-03 | Cluster labels use HTML format | integration | `go test ./internal/render/... -run TestClusterHTMLLabels -v` | Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/render/... -v`
- **Per wave merge:** `go test ./internal/... -v`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/render/expanded_test.go` - New test file for expanded view tests using `cyp-auth-infra.toml` fixture
- [ ] Test helper to load `cyp-auth-infra/cyp-auth-infra.toml` as test fixture
- [ ] Tests for nested cluster rendering
- [ ] Tests for HTML table attributes
- [ ] Tests for cluster HTML labels

### Recommended Test Implementation

```go
// internal/render/expanded_test.go

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestExpandedViewNestedClusters(t *testing.T) {
    // Load cyp-auth-infra.toml fixture
    m := loadCYPAuthInfraModel(t)

    v := view.GenerateExpandedView(m)
    g := graph.BuildExpandedGraph(v)

    output, err := render.RenderDOT(g)
    require.NoError(t, err)

    dotStr := string(output)

    // BUG-01: server.pam cluster should exist
    assert.Contains(t, dotStr, "cluster_server.pam",
        "DOT should contain nested cluster for server.pam")

    // BUG-02: server.pam.unix and server.pam.cyp should exist
    assert.Contains(t, dotStr, "server.pam.unix",
        "DOT should contain nested component server.pam.unix")
    assert.Contains(t, dotStr, "server.pam.cyp",
        "DOT should contain nested component server.pam.cyp")

    // BUG-03: Edges to nested components should exist
    // Note: Edges use the format "source_to_target"
    assert.Contains(t, dotStr, "server.sshd_to_server.pam",
        "DOT should contain edge from sshd to pam components")
}
```

## Sources

### Primary (HIGH confidence)
- `internal/graph/builder.go` - Lines 87-123 (`buildNestedCluster`) - Verified recursive cluster creation is correct
- `internal/graph/graph.go` - Lines 97-109 (`Cluster` struct) - Verified `Clusters` field exists
- `internal/render/converter.go` - Lines 224-270 (`createCluster`) - Verified missing recursive call
- `internal/view/scope.go` - Lines 14-51 (`GenerateExpandedView`) - Verified all units are added to view

### Secondary (MEDIUM confidence)
- `cyp-auth-infra/cyp-auth-infra.expanded.dot` - Confirmed bug by verifying missing patterns
- `internal/graph/builder_test.go` - Verified existing test patterns for expanded view

### Tertiary (LOW confidence)
- None - all findings verified against source code

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Using existing project libraries and patterns
- Architecture: HIGH - Root cause clearly identified in source code
- Pitfalls: HIGH - Based on direct code analysis
- Bug fix approach: HIGH - Simple recursive call addition

**Research date:** 2026-03-13
**Valid until:** 30 days - Codebase is stable, fix approach is straightforward
