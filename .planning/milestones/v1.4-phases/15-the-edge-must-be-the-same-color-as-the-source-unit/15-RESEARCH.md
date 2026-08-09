# Phase 15: Edge Coloring - Research

**Researched:** 2026-03-20
**Domain:** Graph rendering enhancement (go-graphviz/cgraph edge styling)
**Confidence:** HIGH

## Summary

This phase adds visual styling to diagram edges so their color matches the source unit's border color. The implementation requires changes across three files: adding a `Color` field to the `Edge` struct, computing the color during edge creation, and applying it during DOT rendering.

The existing codebase already has all the infrastructure needed:
- `GetStyleForType(t, isExternal)` returns border colors by C4 level
- `model.Link.Color` field exists for explicit TOML overrides
- `cgraph.Edge` has `SetColor()` and `SetFontColor()` methods

**Primary recommendation:** Add `Color` field to `graph.Edge`, compute it in `builder.go` using existing color lookup functions, and apply in `converter.go` using go-graphviz edge styling methods.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

**Color Source (D-01):**
- Edge color comes from the source unit's border color
- Border colors are defined in `internal/model/colors.go` per C4 level:
  - C1 (system): `#3C7FC0`
  - C2 (container): `#3C7FC0`
  - C3 (component): `#78A8D8`
- External variants have different colors (gray tones)

**Label Coloring (D-02):**
- Edge labels (technology, description) match the edge color
- Both technology bracket text and description text use the same color as the edge line

**Explicit Color Override (D-03):**
- If `link.color` is explicitly set in TOML, it overrides the source border color
- Fallback chain: explicit color -> source unit border color

### Claude's Discretion

- Exact implementation approach (where to add Color field to Edge struct)
- How to look up source unit's border color during edge creation
- Test coverage strategy

### Deferred Ideas (OUT OF SCOPE)

None - discussion stayed within phase scope.

</user_constraints>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| go-graphviz/cgraph | current | Graphviz bindings for Go | Already used throughout codebase |
| internal/model/colors.go | existing | C4-PlantUML color palette | Canonical color definitions |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| internal/graph/shapes.go | existing | GetStyleForType() | Border color lookup by unit type |

### Key Methods (go-graphviz/cgraph)
| Method | Purpose | Documentation |
|--------|---------|---------------|
| `Edge.SetColor(v string)` | Sets edge line color | `go doc github.com/goccy/go-graphviz/cgraph Edge` |
| `Edge.SetFontColor(v string)` | Sets edge label text color | Same |

## Architecture Patterns

### Current Edge Flow
```
view.Link -> builder.createEdge() -> graph.Edge -> converter.createEdge() -> cgraph.Edge
```

### Required Changes

1. **Add Color field to graph.Edge** (`internal/graph/graph.go`):
```go
type Edge struct {
    Source     string
    Target     string
    Label      *EdgeLabel
    Style      string
    ArrowHead  ArrowDirection
    Color      string  // NEW: edge line and label color
}
```

2. **Compute color during edge creation** (`internal/graph/builder.go`):
```go
func createEdge(source, target string, link model.Link, sourceEntry *view.Entry) *Edge {
    edge := &Edge{
        Source: source,
        Target: target,
        // ... existing fields ...
    }

    // Determine color: explicit override -> source border color
    if link.Color != "" {
        edge.Color = link.Color
    } else if sourceEntry != nil {
        style := GetStyleForType(sourceEntry.Unit.Type, sourceEntry.IsExternal)
        edge.Color = style.BorderColor
    }

    return edge
}
```

3. **Apply color during rendering** (`internal/render/converter.go`):
```go
func createEdge(cg *cgraph.Graph, source, target *cgraph.Node, edge *graph.Edge) error {
    e, err := cg.CreateEdgeByName(edgeName, source, target)
    // ... existing label and style setup ...

    // NEW: Apply edge color (line and label)
    if edge.Color != "" {
        e.SetColor(edge.Color)
        e.SetFontColor(edge.Color)
    }

    return nil
}
```

### Pattern: Looking Up Source Unit Info

The challenge is that `createEdge()` in `builder.go` currently receives only `source` (string path) and `link`. To get the source unit's type and external status, we need access to `view.Units` map.

**Option A: Pass source entry to createEdge** (Recommended)
```go
func processOutgoingLinks(path string, links []model.Link, viewUnits map[string]*view.Entry, seen map[string]bool) []*Edge {
    for _, link := range links {
        sourceEntry := viewUnits[path]  // Get source entry for color lookup
        edge := createEdge(path, link.Peer, link, sourceEntry)
        // ...
    }
}
```

**Option B: Add viewUnits parameter to createEdge**
```go
func createEdge(source, target string, link model.Link, viewUnits map[string]*view.Entry) *Edge
```

Option A is cleaner - pass only what's needed (the source entry).

### Anti-Patterns to Avoid

- **Don't duplicate color logic:** Use existing `GetStyleForType()` from `shapes.go`
- **Don't hardcode colors:** Reference `model.SystemBorder`, `model.ContainerBorder`, etc.
- **Don't forget external types:** External units have different border colors (`SystemExternalBorder`, etc.)

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Color lookup by type | Switch statement for colors | `GetStyleForType(t, isExternal).BorderColor` | Already handles all C4 levels and external variants |
| Level detection | Manual type checking | `LevelForType(t)` | Returns C1/C2/C3 for any unit type |

**Key insight:** The existing `shapes.go` already has all the color mapping logic. Just call `GetStyleForType()` and extract `BorderColor`.

## Common Pitfalls

### Pitfall 1: Missing Source Entry Lookup
**What goes wrong:** Edge color is empty or default because source entry wasn't passed to createEdge()
**Why it happens:** createEdge() currently only receives strings (source/target) and link
**How to avoid:** Modify processOutgoingLinks/processIncomingLinks to pass sourceEntry to createEdge()
**Warning signs:** All edges render with default black color

### Pitfall 2: Forgetting FontColor
**What goes wrong:** Edge line is colored but label text remains black
**Why it happens:** SetColor() only affects the edge line, not label text
**How to avoid:** Call both `e.SetColor(edge.Color)` AND `e.SetFontColor(edge.Color)`
**Warning signs:** Colored arrows with black text labels

### Pitfall 3: Ignoring link.Color Override
**What goes wrong:** Explicit TOML color overrides don't work
**Why it happens:** Always using source border color without checking link.Color first
**How to avoid:** Check `link.Color != ""` before falling back to source border color
**Warning signs:** `link.color = "#FF0000"` in TOML has no effect

### Pitfall 4: Wrong Color for External Types
**What goes wrong:** External system edges use internal blue instead of gray
**Why it happens:** Not passing isExternal flag to GetStyleForType()
**How to avoid:** Use `sourceEntry.IsExternal` from the view entry
**Warning signs:** External boundary units have blue outgoing edges

## Code Examples

### Current createEdge (builder.go:272-295)
```go
// createEdge creates an edge from a link with defaults applied.
func createEdge(source, target string, link model.Link) *Edge {
    edge := &Edge{
        Source: source,
        Target: target,
        Label: &EdgeLabel{
            Technology:  link.Technology,
            Description: link.Description,
            Position:    string(link.LabelPosition),
        },
        Style:     link.Style,
        ArrowHead: ArrowDirection(link.Arrow),
    }

    // Apply defaults
    if edge.Style == "" {
        edge.Style = "solid"
    }

    if edge.Label.Position == "" {
        edge.Label.Position = "middle"
    }

    return edge
}
```

### Current createEdge (converter.go:293-333)
```go
// createEdge creates a cgraph.Edge from a graph.Edge.
func createEdge(cg *cgraph.Graph, source, target *cgraph.Node, edge *graph.Edge) error {
    edgeName := edge.Source + "_to_" + edge.Target
    if edge.Label != nil && edge.Label.Description != "" {
        edgeName += "_" + sanitizeForName(edge.Label.Description)
    }
    e, err := cg.CreateEdgeByName(edgeName, source, target)
    if err != nil {
        return fmt.Errorf("create edge by name: %w", err)
    }

    // Set edge label
    if edge.Label != nil {
        e.SetLabel(buildEdgeLabel(edge.Label))
        e.SetFontSize(fontSizeEdge)
    }

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

### Color Lookup Pattern (from shapes.go:109-115)
```go
// GetStyleForType returns styling based on unit type and external status.
func GetStyleForType(t model.UnitType, isExternal bool) *NodeStyle {
    if isExternal {
        return getExternalStyle(t)
    }
    return getLevelStyle(t)
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Default black edges | Source-colored edges | This phase | Visual flow from source to target |

**Consistent with prior phases:**
- Phase 11: Transparent backgrounds, border colors for nodes
- Phase 13: Cluster labels use border color (fontcolor)

This phase extends the same color-matching pattern to edges.

## Open Questions

1. **Should edges from collapsed units with subunits use the parent's color?**
   - What we know: Collapsed units show the parent's type/style
   - What's unclear: If a collapsed system (C1) contains containers (C2), which color for edges?
   - Recommendation: Use the collapsed unit's type (C1 system -> SystemBorder) since that's what the user sees

2. **Edge label position affects readability - should color affect position logic?**
   - What we know: LabelPosition is configurable (middle, head, tail)
   - What's unclear: No interaction needed
   - Recommendation: No change to label positioning - color is independent

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + stretchr/testify |
| Config file | None - Go convention |
| Quick run command | `go test ./internal/graph/... ./internal/render/... -v -count=1` |
| Full suite command | `go test ./...` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| D-01 | Edge color = source border color | unit | `go test ./internal/graph/... -run TestEdgeColor -v` | No - Wave 0 |
| D-02 | Edge label color matches edge color | unit | `go test ./internal/render/... -run TestEdgeLabelColor -v` | No - Wave 0 |
| D-03 | link.Color overrides source border | unit | `go test ./internal/graph/... -run TestEdgeColorOverride -v` | No - Wave 0 |
| Integration | Rendered DOT contains edge colors | integration | `go test ./internal/render/... -run TestEdgeColorRendering -v` | No - Wave 0 |

### Sampling Rate
- Per task commit: `go test ./internal/graph/... ./internal/render/... -v -count=1`
- Per wave merge: `go test ./...`
- Phase gate: Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/graph/builder_test.go` - add TestEdgeColorFromSource tests
- [ ] `internal/render/converter_test.go` - add TestEdgeColorRendering tests
- [ ] No framework install needed - existing test infrastructure

## Sources

### Primary (HIGH confidence)
- Code review of `internal/graph/builder.go` - edge creation flow
- Code review of `internal/graph/graph.go` - Edge struct definition
- Code review of `internal/render/converter.go` - edge rendering
- Code review of `internal/graph/shapes.go` - GetStyleForType()
- Code review of `internal/model/colors.go` - C4 color definitions
- `go doc github.com/goccy/go-graphviz/cgraph Edge` - SetColor/SetFontColor methods

### Secondary (MEDIUM confidence)
- CONTEXT.md decisions - locked color requirements

### Tertiary (LOW confidence)
- None - all findings verified in codebase

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - all libraries already in use
- Architecture: HIGH - code flow is clear, changes are localized
- Pitfalls: HIGH - derived from existing codebase patterns

**Research date:** 2026-03-20
**Valid until:** 30 days (stable codebase, no external dependencies changing)
