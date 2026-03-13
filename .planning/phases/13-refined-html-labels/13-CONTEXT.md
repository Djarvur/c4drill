---
phase: 13
slug: refined-html-labels
created: 2026-03-13
status: ready
---

# Phase 13: Refined HTML Labels - Context

**Gathered:** 2026-03-13
**Status:** Ready for planning

<domain>
## Phase Boundary

Refine HTML labels from Phase 12 with:
1. All units: `shape=box, style=rounded` (instead of `shape=none`)
2. Table attributes: `border="0" cellpadding="0" cellspacing="0"`
3. Cluster labels use HTML format (same as corresponding unit)
4. Cluster labels use unit type coloring

This phase covers both node labels AND cluster (subgraph) labels.

</domain>

<decisions>
## Implementation Decisions

### Shape and Style (REFINED-01)

All units MUST render with:
- `shape=box`
- `style=rounded`

This replaces the current `shape=none` approach from Phase 12.

### Table Attributes (REFINED-02)

All HTML tables MUST include:
- `border="0"`
- `cellpadding="0"`
- `cellspacing="0"`

### Cluster Labels (REFINED-03)

**Cluster labels MUST use HTML format** - same as the corresponding unit would have:
- Person cluster → Person HTML label format
- DB cluster → DB HTML label format
- Queue cluster → Queue HTML label format
- System cluster → System HTML label format (SYS label)
- Container cluster → Container HTML label format (CONT label)
- Component cluster → Component HTML label format (COMP label)

**Cluster labels MUST be colored** - same as the corresponding unit would be colored:
- Border color from `GetStyleForType(unitType, isExternal).BorderColor`
- Font color from `GetStyleForType(unitType, isExternal).FontColor`

### Claude's Discretion

- Test structure and organization
- Error handling for missing unit type

</decisions>

<specifics>
## Specific Ideas

User provided exact HTML templates with refined attributes:

### Person Label (rowspan=2, NO technology)

```html
<table border="0" cellpadding="0" cellspacing="0">
  <tr align=center>
    <td rowspan=2 valign=middle><font size="+4">👤</font></td>
    <td valign=bottom><b>User name</b></td>
  </tr>
  <tr align=center>
    <td valign=top>Description</td>
  </tr>
</table>
```

### DB Label (rowspan=3)

```html
<table border="0" cellpadding="0" cellspacing="0">
  <tr align=center>
    <td rowspan=3 valign=middle><font size="+4">⛁</font></td>
    <td valign=bottom><b>DB name</b></td>
  </tr>
  <tr align=center>
    <td valign=middle><i>[technology]</i></td>
  </tr>
  <tr align=center>
    <td valign=top>Description</td>
  </tr>
</table>
```

### Queue Label (NO rowspan - 4 separate rows)

```html
<table border="0" cellpadding="0" cellspacing="0">
  <tr align=center>
    <td valign=middle>═╦╩═╦══</td>
  </tr>
  <tr align=center>
    <td valign=bottom><b>Queue name</b></td>
  </tr>
  <tr align=center>
    <td valign=middle><i>[technology]</i></td>
  </tr>
  <tr align=center>
    <td valign=top>Description</td>
  </tr>
</table>
```

### System Label (rowspan=3, monospace SYS)

```html
<table border="0" cellpadding="0" cellspacing="0">
  <tr align=center>
    <td rowspan=3 valign=middle><font face="monospace">SYS</font></td>
    <td valign=bottom><b>System name</b></td>
  </tr>
  <tr align=center>
    <td valign=middle><i>[technology]</i></td>
  </tr>
  <tr align=center>
    <td valign=top>Description</td>
  </tr>
</table>
```

### Container Label (rowspan=3, monospace CONT)

```html
<table border="0" cellpadding="0" cellspacing="0">
  <tr align=center>
    <td rowspan=3 valign=middle><font face="monospace">CONT</font></td>
    <td valign=bottom><b>Container name</b></td>
  </tr>
  <tr align=center>
    <td valign=middle><i>[technology]</i></td>
  </tr>
  <tr align=center>
    <td valign=top>Description</td>
  </tr>
</table>
```

**Note:** Also used for Box type.

### Component Label (rowspan=3, monospace COMP)

```html
<table border="0" cellpadding="0" cellspacing="0">
  <tr align=center>
    <td rowspan=3 valign=middle><font face="monospace">COMP</font></td>
    <td valign=bottom><b>Component name</b></td>
  </tr>
  <tr align=center>
    <td valign=middle><i>[technology]</i></td>
  </tr>
  <tr align=center>
    <td valign=top>Description</td>
  </tr>
</table>
```

</specifics>

<code_context>
## Existing Code Insights

### Reusable Assets

- **`internal/render/labels.go`** - 6 HTML label builder functions (buildPersonHTMLLabel, buildDbHTMLLabel, etc.)
- **`internal/graph/shapes.go`** - Type helper functions (IsDbType, IsQueueType, IsSystemType, IsContainerType, IsComponentType)
- **`internal/graph/shapes.go:GetStyleForType()`** - Returns NodeStyle with BorderColor, FontColor

### Established Patterns

- **Node HTML labels**: `buildHTMLLabelForType(label, unitType)` dispatches to type-specific builders
- **HTML string handling**: `graph.StrdupHTML()` for HTML label strings (not `SetLabel()`)
- **Monospace**: Use `<font face="monospace">` instead of `<tt>` (GraphViz doesn't support `<tt>`)

### Integration Points

**Cluster struct needs enhancement:**
```go
// internal/graph/graph.go
type Cluster struct {
    ID       string
    Label    *Label
    Nodes    []*Node
    Clusters []*Cluster
    Style    *NodeStyle
    // NEW: UnitType for HTML label dispatch
    Type     model.UnitType  // Add this field
    // NEW: IsExternal for style lookup
    IsExternal bool          // Add this field
}
```

**Cluster builder needs update:**
```go
// internal/graph/builder.go
func buildCluster(entry *view.Entry) *Cluster {
    cluster := &Cluster{
        // ...
        Type:       entry.Unit.Type,      // NEW
        IsExternal: entry.IsExternal,     // NEW
    }
}
```

**Cluster rendering needs HTML labels:**
```go
// internal/render/converter.go
func createCluster(parent *cgraph.Graph, cluster *graph.Cluster, ...) error {
    // OLD: subgraph.SetLabel(cluster.Label.Name)
    // NEW: Use HTML label with coloring
    htmlLabel := buildHTMLLabelForType(cluster.Label, cluster.Type)
    htmlStr, _ := cg.StrdupHTML(htmlLabel)
    subgraph.SetLabel(htmlStr)

    // Set colors from cluster.Style
    if cluster.Style.FontColor != "" {
        subgraph.SetFontColor(cluster.Style.FontColor)
    }
    if cluster.Style.BorderColor != "" {
        subgraph.SafeSet("color", cluster.Style.BorderColor, "")
    }
}
```

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

## Files to Modify

1. **`internal/graph/graph.go`** - Add `Type` and `IsExternal` fields to `Cluster` struct
2. **`internal/graph/builder.go`** - Set `Type` and `IsExternal` in `buildCluster()` and `buildNestedCluster()`
3. **`internal/render/labels.go`** - Add table attributes to all HTML label builders
4. **`internal/render/converter.go`**:
   - Change node shape from `none` to `box` with `style=rounded`
   - Update `createCluster()` to use HTML labels with coloring
5. **`internal/render/labels_test.go`** - Update expected outputs
6. **`internal/graph/graph_test.go`** - Update Cluster test cases

---

## Prior Decisions (Phase 12)

Phase 12 implemented HTML labels with:
- `shape=none` for HTML labels → **Phase 13 changes to `shape=box, style=rounded`**
- `graph.StrdupHTML()` for HTML label strings → **Keep this**
- `<font face="monospace">` instead of `<tt>` → **Keep this**
- Type-specific label builders → **Keep this, add table attributes**

---

*Phase: 13-refined-html-labels*
*Context gathered: 2026-03-13*
