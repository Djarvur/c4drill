# Phase 12: HTML Labels for All Unit Types - Research

**Researched:** 2026-03-13
**Domain:** GraphViz HTML-like labels with record shapes
**Confidence:** HIGH

## Summary

This phase converts all unit type labels from record-style format to HTML table format with specific layouts per unit type. The key technical challenge is embedding HTML tables inside record shapes, which GraphViz supports but officially discourages due to potential conflicts between the two label schemas.

The implementation requires modifying `internal/render/labels.go` to create HTML table builders for each unit type category, and updating `internal/render/converter.go` to dispatch to the appropriate label builder based on `node.Type`.

**Primary recommendation:** Create dedicated HTML label builder functions for each unit type category (Person, DB, Queue, System/Container/Component/Box), with conditional row omission for optional fields (technology, description).

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### Label Format by Unit Type

**Person (TypePerson, TypePersonExternal):**
- HTML table with icon on left (rowspan=2), name/description on right
- Exact format:
```html
<table>
  <tr align=center>
    <td rowspan=2 valign=middle><font size="+4">👤</font></td>
    <td valign=bottom><b>User name</b></td>
  </tr>
  <tr align=center>
    <td valign=top>Description</td>
  </tr>
</table>
```

**Database (TypeDb, TypeDbExternal, TypeContainerDb, TypeComponentDb):**
- HTML table with icon (rowspan=3), name bold, technology italic, description
- Exact format:
```html
<table>
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

**Queue (TypeQueue, TypeQueueExternal, TypeContainerQueue, TypeComponentQueue):**
- HTML table with 4 rows (no rowspan): graphics, name, technology, description
- Exact format:
```html
<table>
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

**System (TypeSystem, TypeSystemExternal):**
- HTML table with `SYS` label (rowspan=3), name bold, technology italic, description
- Exact format:
```html
<table>
  <tr align=center>
    <td rowspan=3 valign=middle><tt>SYS</tt></td>
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

**Container (TypeContainer):**
- HTML table with `CONT` label (rowspan=3), name bold, technology italic, description
- Exact format:
```html
<table>
  <tr align=center>
    <td rowspan=3 valign=middle><tt>CONT</tt></td>
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

**Component (TypeComponent):**
- HTML table with `COMP` label (rowspan=3), name bold, technology italic, description
- Exact format:
```html
<table>
  <tr align=center>
    <td rowspan=3 valign=middle><tt>COMP</tt></td>
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

**Box (TypeBox):**
- Same format as Container (use `CONT` label)
- Box is a grouping container, visually same as Container

### Shape Requirement
- All units must use `shape=record` (NOT shape=none)
- HTML tables are embedded inside record labels via `<...>` syntax

### Optional Fields
- If technology is empty, omit the technology row
- If description is empty, omit the description row

### Claude's Discretion

None - all specifications are locked.

### Deferred Ideas (OUT OF SCOPE)

None - all requirements specified with exact HTML templates.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| HTML-01 | All unit types render with HTML table labels inside record shapes | GraphViz HTML label syntax documented; record shape embedding supported but officially discouraged |
| HTML-02 | Each unit type has specific format: Person (icon+name+desc), Database (icon+name+tech+desc), Queue (graphics+name+tech+desc), System/Container/Component (name+tech+desc) | Exact HTML templates provided in CONTEXT.md; helper functions needed for each type category |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| go-graphviz | current | GraphViz bindings for Go | Already in use; provides cgraph.Node.SetLabel() for HTML labels |
| stretchr/testify | current | Testing assertions | Existing test pattern |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| strings (stdlib) | - | StringBuilder for HTML generation | All label builders |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| HTML in record shape | shape=none with HTML label | CONTEXT.md explicitly requires shape=record; changing would alter visual appearance |

**Installation:**
No new dependencies required.

## Architecture Patterns

### Recommended Project Structure
```
internal/render/
├── labels.go          # Add HTML label builders for each unit type category
├── converter.go       # Modify createNode() to dispatch by unit type
└── labels_test.go     # Add tests for each new label builder
```

### Pattern 1: Type-Based Label Dispatch
**What:** Select label builder based on unit type category
**When to use:** In `createNode()` when building node labels
**Example:**
```go
// Source: Based on existing converter.go pattern
func createNode(cg *cgraph.Graph, node *graph.Node) (*cgraph.Node, error) {
    // ... existing code ...

    if node.Label != nil {
        label := buildHTMLLabelForType(node.Type, node.Label)
        cn.SetLabel(label)
    }
    // ... rest of function ...
}

func buildHTMLLabelForType(t model.UnitType, label *graph.Label) string {
    switch {
    case graph.IsPersonType(t):
        return buildPersonHTMLLabel(label)
    case isDbType(t):
        return buildDbHTMLLabel(label)
    case isQueueType(t):
        return buildQueueHTMLLabel(label)
    case isSystemType(t):
        return buildSystemHTMLLabel(label)
    case isContainerType(t) || t == model.TypeBox:
        return buildContainerHTMLLabel(label)
    case isComponentType(t):
        return buildComponentHTMLLabel(label)
    default:
        return buildRecordLabel(label) // fallback
    }
}
```

### Pattern 2: HTML Table Builder with Optional Rows
**What:** Build HTML tables with conditional rows based on field presence
**When to use:** For DB, Queue, System, Container, Component types where technology/description may be empty
**Example:**
```go
// Source: CONTEXT.md specification
func buildDbHTMLLabel(label *graph.Label) string {
    var sb strings.Builder
    sb.WriteString("<")
    sb.WriteString(`<table>`)

    // Calculate rowspan based on present fields
    rows := 1 // name always present
    if label.Technology != "" {
        rows++
    }
    if label.Description != "" {
        rows++
    }

    // First row: icon (rowspan) + name
    sb.WriteString(`<tr align=center>`)
    sb.WriteString(fmt.Sprintf(`<td rowspan=%d valign=middle><font size="+4">&#x26C1;</font></td>`, rows))
    sb.WriteString(fmt.Sprintf(`<td valign=bottom><b>%s</b></td>`, escapeHTML(label.Name)))
    sb.WriteString(`</tr>`)

    // Technology row (if present)
    if label.Technology != "" {
        sb.WriteString(`<tr align=center>`)
        sb.WriteString(fmt.Sprintf(`<td valign=middle><i>[%s]</i></td>`, escapeHTML(label.Technology)))
        sb.WriteString(`</tr>`)
    }

    // Description row (if present)
    if label.Description != "" {
        sb.WriteString(`<tr align=center>`)
        sb.WriteString(fmt.Sprintf(`<td valign=top>%s</td>`, escapeHTML(label.Description)))
        sb.WriteString(`</tr>`)
    }

    sb.WriteString(`</table>`)
    sb.WriteString(">")
    return sb.String()
}
```

### Anti-Patterns to Avoid
- **Using shape=none:** CONTEXT.md explicitly requires shape=record; must embed HTML inside record
- **Always including empty rows:** Must omit technology/description rows when fields are empty
- **Hardcoding rowspan=3:** Rowspan must be calculated based on which optional fields are present
- **Using deprecated buildHTMLLabel:** The existing function is deprecated and uses different format

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|------|
| HTML escaping | Custom escape function | `html.EscapeString()` from stdlib | Handles all edge cases (quotes, ampersands, brackets) |
| Type categorization | Manual type checking | Helper functions like `isDbType()`, `isQueueType()` | Reusable, clearer intent, matches existing `graph.IsPersonType()` pattern |

**Key insight:** The project already has `graph.IsPersonType()` and `graph.IsExternalType()` helpers. Follow this pattern for new type category checks.

## Common Pitfalls

### Pitfall 1: GraphViz HTML-in-Record Conflict
**What goes wrong:** GraphViz documentation warns that "adding HTML labels to record-based shapes (record and Mrecord) is discouraged and may lead to unexpected behavior because of their conflicting label schemas"
**Why it happens:** Record shapes have their own field syntax (`{field|field}`) that conflicts with HTML table structure
**How to avoid:** Use `<...>` delimiters for HTML labels (not quotes); ensure HTML is well-formed; test rendering with `dot` command
**Warning signs:** Malformed node shapes, missing content, layout corruption

### Pitfall 2: Incorrect Rowspan Calculation
**What goes wrong:** Icon cell doesn't span correct number of rows when optional fields are omitted
**Why it happens:** Hardcoding rowspan=3 when only 2 rows (name + description) exist
**How to avoid:** Calculate rowspan dynamically: `1 (name) + (1 if technology) + (1 if description)`
**Warning signs:** Icon appears truncated or misaligned with content

### Pitfall 3: Unicode Character Encoding
**What goes wrong:** Icons (person emoji, DB symbol, queue graphics) don't render correctly
**Why it happens:** Unicode characters may need HTML entity encoding in GraphViz HTML labels
**How to avoid:** Use HTML entities for special characters: `&#x1F464;` for person, `&#x26C1;` for DB cylinder, or test with literal Unicode
**Warning signs:** Question marks, boxes, or missing icons in output

### Pitfall 4: Queue Graphics Row
**What goes wrong:** Queue type uses wrong graphics or includes rowspan for graphics
**Why it happens:** Queue format is different from other types (no rowspan, 4 separate rows)
**How to avoid:** Queue has NO rowspan - graphics (`═╦╩═╦══`), name, technology, description are all separate single-row cells
**Warning signs:** Queue nodes look like DB nodes, incorrect layout

## Code Examples

Verified patterns from existing codebase and GraphViz documentation:

### Person HTML Label (Icon with Rowspan)
```go
// Source: CONTEXT.md specification
func buildPersonHTMLLabel(label *graph.Label) string {
    var sb strings.Builder
    sb.WriteString("<")
    sb.WriteString(`<table>`)

    rows := 1 // name always present
    if label.Description != "" {
        rows++
    }

    // First row: icon (rowspan) + name
    sb.WriteString(`<tr align=center>`)
    sb.WriteString(fmt.Sprintf(`<td rowspan=%d valign=middle><font size="+4">&#x1F464;</font></td>`, rows))
    sb.WriteString(fmt.Sprintf(`<td valign=bottom><b>%s</b></td>`, html.EscapeString(label.Name)))
    sb.WriteString(`</tr>`)

    // Description row (if present)
    if label.Description != "" {
        sb.WriteString(`<tr align=center>`)
        sb.WriteString(fmt.Sprintf(`<td valign=top>%s</td>`, html.EscapeString(label.Description)))
        sb.WriteString(`</tr>`)
    }

    sb.WriteString(`</table>`)
    sb.WriteString(">")
    return sb.String()
}
```

### Queue HTML Label (No Rowspan, 4 Rows)
```go
// Source: CONTEXT.md specification
func buildQueueHTMLLabel(label *graph.Label) string {
    var sb strings.Builder
    sb.WriteString("<")
    sb.WriteString(`<table>`)

    // Row 1: Graphics (no rowspan)
    sb.WriteString(`<tr align=center>`)
    sb.WriteString(`<td valign=middle>&#x255F;&#x2569;&#x2562;&#x255F;&#x255F;</td>`)
    sb.WriteString(`</tr>`)

    // Row 2: Name (bold)
    sb.WriteString(`<tr align=center>`)
    sb.WriteString(fmt.Sprintf(`<td valign=bottom><b>%s</b></td>`, html.EscapeString(label.Name)))
    sb.WriteString(`</tr>`)

    // Row 3: Technology (italic, if present)
    if label.Technology != "" {
        sb.WriteString(`<tr align=center>`)
        sb.WriteString(fmt.Sprintf(`<td valign=middle><i>[%s]</i></td>`, html.EscapeString(label.Technology)))
        sb.WriteString(`</tr>`)
    }

    // Row 4: Description (if present)
    if label.Description != "" {
        sb.WriteString(`<tr align=center>`)
        sb.WriteString(fmt.Sprintf(`<td valign=top>%s</td>`, html.EscapeString(label.Description)))
        sb.WriteString(`</tr>`)
    }

    sb.WriteString(`</table>`)
    sb.WriteString(">")
    return sb.String()
}
```

### Type Category Helper Functions
```go
// Source: Following existing graph.IsPersonType() pattern in shapes.go
func isDbType(t model.UnitType) bool {
    return t == model.TypeDb || t == model.TypeDbExternal ||
        t == model.TypeContainerDb || t == model.TypeComponentDb
}

func isQueueType(t model.UnitType) bool {
    return t == model.TypeQueue || t == model.TypeQueueExternal ||
        t == model.TypeContainerQueue || t == model.TypeComponentQueue
}

func isSystemType(t model.UnitType) bool {
    return t == model.TypeSystem || t == model.TypeSystemExternal
}

func isContainerType(t model.UnitType) bool {
    return t == model.TypeContainer
}

func isComponentType(t model.UnitType) bool {
    return t == model.TypeComponent
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Record-style labels for all units | HTML labels for all units | Phase 12 | Consistent visual format, type-specific layouts |
| Person uses `buildPersonRecordLabel()` | Person uses HTML table | Phase 12 | Aligns with other unit types |
| System/Container/Component use `buildRecordLabel()` | Each has distinct HTML label with type indicator | Phase 12 | Visual differentiation via SYS/CONT/COMP labels |

**Deprecated/outdated:**
- `buildHTMLLabel()`: Deprecated in labels.go; will be replaced by type-specific builders
- `buildRecordLabel()`: Will only be used as fallback for unknown types
- `buildPersonRecordLabel()`: Will be replaced by `buildPersonHTMLLabel()`

## Open Questions

1. **Unicode Entity Encoding**
   - What we know: GraphViz supports both literal Unicode and HTML entities
   - What's unclear: Whether entities work better for cross-platform rendering
   - Recommendation: Use HTML entities (`&#x1F464;`) for reliability; test with literal Unicode as fallback

2. **Queue Graphics Character Set**
   - What we know: CONTEXT.md specifies `═╦╩═╦══` as queue graphics
   - What's unclear: Exact Unicode code points for these box-drawing characters
   - Recommendation: Use `&#x255F;&#x2569;&#x2562;&#x255F;&#x255F;` or test with literal characters; verify rendering in SVG output

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + stretchr/testify |
| Config file | None - Go convention |
| Quick run command | `go test ./internal/render/... -run TestHTMLLabel -v` |
| Full suite command | `go test ./internal/render/... -v` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| HTML-01 | All unit types render with HTML table labels | unit | `go test ./internal/render/... -run TestHTMLLabelGeneration -v` | Wave 0 needed |
| HTML-02 | Person format: icon+name+desc | unit | `go test ./internal/render/... -run TestPersonHTMLLabel -v` | Wave 0 needed |
| HTML-02 | DB format: icon+name+tech+desc | unit | `go test ./internal/render/... -run TestDbHTMLLabel -v` | Wave 0 needed |
| HTML-02 | Queue format: graphics+name+tech+desc | unit | `go test ./internal/render/... -run TestQueueHTMLLabel -v` | Wave 0 needed |
| HTML-02 | System format: SYS+name+tech+desc | unit | `go test ./internal/render/... -run TestSystemHTMLLabel -v` | Wave 0 needed |
| HTML-02 | Container format: CONT+name+tech+desc | unit | `go test ./internal/render/... -run TestContainerHTMLLabel -v` | Wave 0 needed |
| HTML-02 | Component format: COMP+name+tech+desc | unit | `go test ./internal/render/... -run TestComponentHTMLLabel -v` | Wave 0 needed |

### Sampling Rate
- **Per task commit:** `go test ./internal/render/... -v`
- **Per wave merge:** `go test ./... -v`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/render/labels_test.go` - Add tests for HTML label builders (TestPersonHTMLLabel, TestDbHTMLLabel, TestQueueHTMLLabel, TestSystemHTMLLabel, TestContainerHTMLLabel, TestComponentHTMLLabel)
- [ ] `internal/render/labels_test.go` - Add test for optional field omission (empty technology, empty description)
- [ ] `internal/render/converter_test.go` - Add test for label dispatch by type

*(Note: Existing test infrastructure in `internal/render/labels_test.go` can be extended for new tests)*

## Sources

### Primary (HIGH confidence)
- [GraphViz Node Shapes Documentation](https://graphviz.org/doc/info/shapes.html) - HTML-like labels syntax, record shape compatibility warning, rowspan/colspan attributes
- Project source code: `internal/render/labels.go`, `internal/render/converter.go`, `internal/graph/shapes.go`

### Secondary (MEDIUM confidence)
- CONTEXT.md - Exact HTML templates for each unit type (user-provided specifications)

### Tertiary (LOW confidence)
- None

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - No new dependencies; using existing Go stdlib and project patterns
- Architecture: HIGH - Clear implementation path following existing patterns
- Pitfalls: HIGH - GraphViz documentation clearly documents HTML-in-record warning

**Research date:** 2026-03-13
**Valid until:** 30 days (stable GraphViz specification)
