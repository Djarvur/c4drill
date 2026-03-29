# Phase 18: Simplified Shapes - Research

**Researched:** 2026-03-24
**Domain:** GraphViz native shapes, HTML labels, icon system removal
**Confidence:** HIGH

## Summary

Phase 18 removes the SVG icon system and replaces it with native GraphViz shapes for DB/Queue units. The primary changes involve:

1. **Icon System Removal**: Delete the `internal/render/icons/` package, `icon_extractor.go`, `svg_icons.go`, `dot_icons.go`, and related tests. Remove `iconReserve` constant and all IconExtractor usage from converter.go.

2. **Native Cylinder Shapes**: Use GraphViz's built-in `shape=cylinder` for DB units, and `shape=cylinder` with `orientation=90` for Queue units (horizontal cylinder).

3. **Simplified Labels**: Person labels keep 2-column layout with 👤 emoji (using `<font size="+4">`). System/Box/Container/Component/DB/Queue labels become simple 3-row tables without icon columns.

4. **Preserve Word Wrap**: Keep Phase 17's word-wrap functionality intact.

**Primary recommendation:** The go-graphviz library (v0.2.10 with fork) supports `CylinderShape` constant and `SetOrientation(float64)` method for nodes. Implementation is straightforward: modify shape assignment in converter.go and simplify label builders in labels.go.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** Remove `internal/render/icons/` package entirely (embed.go and all .svg files)
- **D-02:** Remove `internal/render/icon_extractor.go` and IconExtractor struct
- **D-03:** Remove `.icons/` directory generation from output
- **D-04:** Remove any SVG postprocessing logic that injects icons after rendering
- **D-05:** DB units use `shape=cylinder` (native GraphViz)
- **D-06:** DB external units also use cylinder shape
- **D-07:** DB label is simple 3-row HTML table (name, technology, description) — no icon column
- **D-08:** Queue units use `shape=cylinder` with 90° orientation
- **D-09:** Queue external units also use rotated cylinder shape
- **D-10:** Queue label is simple 3-row HTML table (name, technology, description) — no icon column
- **D-11:** Person units keep 2-column table layout
- **D-12:** First column contains 👤 emoji with `<font size="+4">` (same as before)
- **D-13:** Second column contains name and description rows
- **D-14:** Person external units use same label format
- **D-15:** All use simple 3-row table (name, technology, description)
- **D-16:** No icon column — single-column layout
- **D-17:** External variants use same format
- **D-18:** Keep word-wrap functionality from Phase 17
- **D-19:** `--label-ratio` flag continues to work

### Claude's Discretion
- Exact cylinder orientation syntax (may need `rankdir` or other attrs)
- Fallback if GraphViz doesn't support rotated cylinder
- Handling of iconReserve constant removal

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| ICON-01 | Remove icons package (internal/icons) entirely | Delete `internal/render/icons/` directory with embed.go and all .svg files |
| ICON-02 | Remove IconExtractor from converter | Remove `icon_extractor.go`, remove IconExtractor usage from `converter.go` |
| ICON-03 | Remove .icons/ directory generation | Remove from IconExtractor.Extract() calls and output generation |
| ICON-04 | Remove SVG postprocessing logic | Delete `svg_icons.go` and `dot_icons.go` |
| DB-01 | DB units render with GraphViz native `shape=cylinder` | Use `cgraph.CylinderShape` in converter.go |
| DB-02 | DB external units also use cylinder shape | Same shape for all DB types (IsDbType) |
| DB-03 | DB label is simple 3-row table | Modify `buildDbHTMLLabel()` to remove icon column |
| QUEUE-01 | Queue units render with `shape=cylinder` rotated 90° | Use `cn.SetOrientation(90.0)` after `SetShape(CylinderShape)` |
| QUEUE-02 | Queue external units also use rotated cylinder | Same shape for all Queue types (IsQueueType) |
| QUEUE-03 | Queue label is simple 3-row table | Modify `buildQueueHTMLLabel()` to remove icon column |
| PERSON-01 | Person units use 2-column table layout | Keep existing 2-column structure |
| PERSON-02 | First column contains 👤 emoji at font size +4 | Use `<font size="+4">&#x1F464;</font>` in first column |
| PERSON-03 | Second column contains name and description rows | Keep existing structure, remove img tag |
| PERSON-04 | Person external units use same label format | Same label builder for both |
| LABEL-01 | System units use simple 3-row table | Modify `buildSystemHTMLLabel()` to single-column |
| LABEL-02 | Box units use same 3-row table format | Box uses Container label builder |
| LABEL-03 | Container and Component units use same 3-row table | Modify both builders to single-column |
| LABEL-04 | No icon column in system/box/container/component labels | Remove icon column logic from all builders |
| WRAP-01 | All label text is word-wrapped | Keep existing `wrap.go` unchanged |
| WRAP-02 | Existing --label-ratio flag continues to work | No changes to CLI or wrap logic |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| goccy/go-graphviz | v0.2.10 (fork) | GraphViz bindings for Go | Project dependency, supports CylinderShape |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| stretchr/testify | v1.11.1 | Testing assertions | All unit tests |

### Key Shape Constants (from go-graphviz/cgraph)
```go
// Available in cgraph package:
CylinderShape  Shape = "cylinder"

// Methods for nodes:
func (n *Node) SetShape(v Shape) *Node
func (n *Node) SetOrientation(v float64) *Node  // 0.0 to 360.0
```

## Architecture Patterns

### Current Label Builder Pattern
Each unit type has a dedicated builder function in `internal/render/labels.go`:
- `buildPersonHTMLLabel()` - 2-column with icon
- `buildDbHTMLLabel()` - 2-column with icon
- `buildQueueHTMLLabel()` - 2-column with icon
- `buildSystemHTMLLabel()` - 2-column with icon
- `buildContainerHTMLLabel()` - 2-column with icon
- `buildComponentHTMLLabel()` - 2-column with icon

### Target Label Builder Pattern (After Phase 18)
- `buildPersonHTMLLabel()` - 2-column with emoji (no img tag)
- `buildDbHTMLLabel()` - single-column 3-row (no icon)
- `buildQueueHTMLLabel()` - single-column 3-row (no icon)
- `buildSystemHTMLLabel()` - single-column 3-row (no icon)
- `buildContainerHTMLLabel()` - single-column 3-row (no icon)
- `buildComponentHTMLLabel()` - single-column 3-row (no icon)

### Shape Assignment Pattern
In `internal/render/converter.go`:
```go
// Current: All nodes use BoxShape
cn.SetShape(cgraph.BoxShape)

// After: DB/Queue use CylinderShape
if graph.IsDbType(node.Type) || graph.IsQueueType(node.Type) {
    cn.SetShape(cgraph.CylinderShape)
    if graph.IsQueueType(node.Type) {
        cn.SetOrientation(90.0)
    }
} else {
    cn.SetShape(cgraph.BoxShape)
}
```

### Anti-Patterns to Avoid
- **Don't modify wrap.go**: Word-wrap functionality must remain unchanged
- **Don't change iconReserve without cleanup**: Remove the constant AND all references
- **Don't forget icon column width**: Remove `iconColumnWidth` constant usage from simplified labels

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Cylinder shape | Custom SVG or HTML | `cgraph.CylinderShape` | Native GraphViz support |
| Rotated cylinder | Manual rotation math | `cn.SetOrientation(90.0)` | Built-in orientation attribute |
| Emoji rendering | Base64 encoded images | `<font size="+4">&#x1F464;</font>` | HTML entity works in GraphViz labels |

**Key insight:** GraphViz has native cylinder support since at least version 2.37. No custom shape implementation needed.

## Common Pitfalls

### Pitfall 1: Cylinder Orientation May Not Render in All Engines
**What goes wrong:** The `orientation` attribute may behave differently across layout engines (dot vs neato vs fdp).
**Why it happens:** Not all GraphViz engines support node rotation equally.
**How to avoid:** Test with the default `dot` engine. If issues arise, consider alternative visual cues for Queue vs DB.
**Warning signs:** Queue cylinders appearing vertical instead of horizontal.

### Pitfall 2: Emoji Rendering in GraphViz
**What goes wrong:** Emojis may not render consistently across all systems or output formats.
**Why it happens:** GraphViz relies on system fonts for emoji support.
**How to avoid:** Use HTML entity `&#x1F464;` with explicit font size. Test on target platforms.
**Warning signs:** Person labels showing empty boxes or fallback characters.

### Pitfall 3: Forgetting to Update labelMaxChars
**What goes wrong:** After removing icon column, label width calculations still subtract `iconColumnWidth`.
**Why it happens:** `calculateTextWidth()` in wrap.go subtracts `iconColumnWidth` from total width.
**How to avoid:** Update `calculateTextWidth()` to accept a parameter for whether icon column is present, or create separate functions for 2-column vs single-column labels.
**Warning signs:** Labels appearing narrower than expected, excessive wrapping.

### Pitfall 4: Incomplete Icon System Removal
**What goes wrong:** Tests fail due to missing icon imports or references.
**Why it happens:** Icon references are spread across multiple files.
**How to avoid:** Systematically delete all files and then run full test suite. Grep for "icon" to find stragglers.
**Warning signs:** Compilation errors about missing icons package.

## Code Examples

### Setting Cylinder Shape for DB Nodes
```go
// Source: go-graphviz/cgraph/attribute.go
// In converter.go createNode():

if graph.IsDbType(node.Type) {
    cn.SetShape(cgraph.CylinderShape)
} else if graph.IsQueueType(node.Type) {
    cn.SetShape(cgraph.CylinderShape)
    cn.SetOrientation(90.0)
} else {
    cn.SetShape(cgraph.BoxShape)
}
```

### Simplified 3-Row Label (DB/Queue/System/Container/Component)
```go
// Source: Modified from internal/render/labels.go pattern
func buildSimpleHTMLLabel(label *graph.Label) string {
    if label == nil {
        return ""
    }

    rowCount := 1 // name
    if label.Technology != "" {
        rowCount++
    }
    if label.Description != "" {
        rowCount++
    }

    // For single-column: use full width, no iconColumnWidth subtraction
    maxChars := labelMaxCharsNoIcon(rowCount)

    var sb strings.Builder
    sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0">`)

    // Row 1: Name (bold)
    sb.WriteString(`<tr align="center"><td valign="bottom"><b>`)
    sb.WriteString(wrapAndEscape(label.Name, maxChars))
    sb.WriteString(`</b></td></tr>`)

    // Row 2: Technology (if present, italic in brackets)
    if label.Technology != "" {
        sb.WriteString(`<tr align="center"><td valign="middle"><i>[`)
        sb.WriteString(wrapAndEscape(label.Technology, maxChars))
        sb.WriteString(`]</i></td></tr>`)
    }

    // Row 3: Description (if present)
    if label.Description != "" {
        sb.WriteString(`<tr align="center"><td valign="top">`)
        sb.WriteString(wrapAndEscape(label.Description, maxChars))
        sb.WriteString(`</td></tr>`)
    }

    sb.WriteString(`</table>`)
    return sb.String()
}
```

### Person Label with Emoji (No Image)
```go
// Source: Modified from internal/render/labels.go buildPersonHTMLLabel
func buildPersonHTMLLabel(label *graph.Label) string {
    if label == nil {
        return ""
    }

    rowCount := 1 // name
    if label.Description != "" {
        rowCount++
    }

    // Person keeps 2-column, so use existing calculation
    maxChars := labelMaxChars(rowCount)

    var sb strings.Builder
    sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0">`)

    rowspan := 1
    if label.Description != "" {
        rowspan = 2
    }

    // Row 1: Emoji + Name
    sb.WriteString(`<tr align="center">`)

    // Emoji column (instead of img tag)
    sb.WriteString(`<td width="36" rowspan="`)
    sb.WriteString(strconv.Itoa(rowspan))
    sb.WriteString(`" valign="middle"><font size="+4">&#x1F464;</font></td>`)

    sb.WriteString(`<td valign="bottom"><b>`)
    sb.WriteString(wrapAndEscape(label.Name, maxChars))
    sb.WriteString(`</b></td>`)
    sb.WriteString(`</tr>`)

    // Row 2: Description (if present)
    if label.Description != "" {
        sb.WriteString(`<tr align="center">`)
        sb.WriteString(`<td valign="top">`)
        sb.WriteString(wrapAndEscape(label.Description, maxChars))
        sb.WriteString(`</td>`)
        sb.WriteString(`</tr>`)
    }

    sb.WriteString(`</table>`)
    return sb.String()
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| SVG icon extraction | Native GraphViz shapes | Phase 18 | Simpler rendering, no external files |
| 2-column labels with icons | Single-column labels (except Person) | Phase 18 | Cleaner appearance, more text space |
| `<img>` tags in labels | Emoji for Person only | Phase 18 | No external icon dependencies |

**Deprecated/outdated:**
- `iconReserve` constant: No longer needed, remove from labels.go
- `iconColumnWidth` for non-Person labels: Remove from width calculations
- `iconTypeForUnit()` function: Can be removed entirely

## Open Questions

1. **Queue Cylinder Orientation Support**
   - What we know: `SetOrientation(90.0)` exists in go-graphviz
   - What's unclear: Whether all output formats (SVG, PNG) render rotated cylinders correctly
   - Recommendation: Test with `dot -Tsvg` and `dot -Tpng` to verify. If rotation fails, fallback to same cylinder shape as DB (vertical).

2. **labelMaxChars Calculation for Single-Column Labels**
   - What we know: Current `calculateTextWidth()` subtracts `iconColumnWidth`
   - What's unclear: Whether to modify existing function or create new one
   - Recommendation: Create `labelMaxCharsNoIcon(rowCount)` that uses full width, or add boolean parameter to existing function.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + stretchr/testify v1.11.1 |
| Config file | None — standard Go test pattern |
| Quick run command | `go test ./internal/render/... -v -short` |
| Full suite command | `go test ./... -v` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| ICON-01 | Remove icons package | unit | `go build ./...` | N/A - deletion |
| ICON-02 | Remove IconExtractor | unit | `go test ./internal/render/... -run IconExtractor` | Needs update |
| ICON-03 | Remove .icons directory | integration | Visual inspection | N/A |
| ICON-04 | Remove SVG postprocessing | unit | `go test ./internal/render/...` | Needs update |
| DB-01 | DB units use cylinder shape | unit | `go test ./internal/render/... -run Cylinder` | Wave 0 |
| DB-02 | DB external units use cylinder | unit | `go test ./internal/render/... -run Cylinder` | Wave 0 |
| DB-03 | DB label is 3-row table | unit | `go test ./internal/render/... -run DbHTMLLabel` | Update existing |
| QUEUE-01 | Queue units use rotated cylinder | unit | `go test ./internal/render/... -run Queue` | Wave 0 |
| QUEUE-02 | Queue external units use rotated cylinder | unit | `go test ./internal/render/... -run Queue` | Wave 0 |
| QUEUE-03 | Queue label is 3-row table | unit | `go test ./internal/render/... -run QueueHTMLLabel` | Update existing |
| PERSON-01 | Person units use 2-column table | unit | `go test ./internal/render/... -run PersonHTMLLabel` | Update existing |
| PERSON-02 | Person has emoji font +4 | unit | `go test ./internal/render/... -run PersonHTMLLabel` | Update existing |
| PERSON-03 | Person has name/description | unit | `go test ./internal/render/... -run PersonHTMLLabel` | Update existing |
| PERSON-04 | Person external same format | unit | `go test ./internal/render/... -run PersonHTMLLabel` | Update existing |
| LABEL-01 | System 3-row table | unit | `go test ./internal/render/... -run SystemHTMLLabel` | Update existing |
| LABEL-02 | Box 3-row table | unit | `go test ./internal/render/... -run ContainerHTMLLabel` | Update existing |
| LABEL-03 | Container/Component 3-row table | unit | `go test ./internal/render/... -run Container` | Update existing |
| LABEL-04 | No icon column | unit | `go test ./internal/render/... -run HTMLLabel` | Update existing |
| WRAP-01 | Word-wrap preserved | unit | `go test ./internal/render/... -run Wrap` | Existing |
| WRAP-02 | --label-ratio works | unit | `go test ./internal/render/... -run LabelRatio` | Existing |

### Sampling Rate
- **Per task commit:** `go test ./internal/render/... -v -short`
- **Per wave merge:** `go test ./... -v`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/render/cylinder_test.go` — tests for DB/Queue cylinder shapes with orientation
- [ ] Update `internal/render/html_labels_internal_test.go` — remove img tag assertions, add emoji assertions
- [ ] Delete `internal/render/icon_extractor_test.go` — no longer needed
- [ ] Delete `internal/render/icons_integration_test.go` — no longer needed
- [ ] Delete `internal/render/icons/icons_test.go` — no longer needed

## Sources

### Primary (HIGH confidence)
- [GraphViz Node Shapes](https://graphviz.org/doc/info/shapes.html) - Confirmed cylinder shape is native polygon-based shape
- [go-graphviz cgraph/attribute.go] - Confirmed `CylinderShape` constant and `SetOrientation(float64)` method exist

### Secondary (MEDIUM confidence)
- Existing codebase analysis of `internal/render/labels.go`, `internal/render/converter.go`, `internal/graph/shapes.go`

### Tertiary (LOW confidence)
- None — all findings verified from source code or official documentation

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Verified go-graphviz supports CylinderShape and SetOrientation
- Architecture: HIGH - Existing codebase patterns are clear and well-documented
- Pitfalls: MEDIUM - Emoji rendering and orientation support need runtime verification

**Research date:** 2026-03-24
**Valid until:** 30 days (stable GraphViz shapes)
