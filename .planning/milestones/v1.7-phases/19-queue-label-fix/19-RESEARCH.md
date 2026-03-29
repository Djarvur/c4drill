# Phase 19: Queue Label Fix - Research

**Researched:** 2026-03-24
**Domain:** GraphViz HTML labels, ASCII art graphics, Go string building
**Confidence:** HIGH

## Summary

Phase 19 fixes the Queue unit rendering by reverting from GraphViz `shape=cylinder` with 90-degree rotation (which doesn't work) to HTML labels with ASCII art graphic representation. The change is straightforward: add a graphic row with `═╦╩═╦═══` pattern as the first row of the Queue HTML table, remove cylinder shape and orientation from the converter.

**Primary recommendation:** Modify `buildQueueHTMLLabel()` to add ASCII art row first, then remove `SetOrientation(90.0)` and cylinder shape for Queue types in converter.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01:** Use `═╦╩═╦═══` as the Queue graphic (classic pipe pattern from Phase 13)
- **D-02:** Graphic row uses center alignment
- **D-03:** Graphic row is NOT word-wrapped (fixed pattern)
- **D-04:** 4-row HTML table: graphic, name, technology, description
- **D-05:** Name/technology/description rows ARE word-wrapped (existing wrap logic)
- **D-06:** Same table attributes as other labels: `border="0" cellpadding="0" cellspacing="0"`
- **D-07:** Queue units use `shape=box, style=rounded` (same as System/Person)
- **D-08:** Remove `SetOrientation(90.0)` call from converter for Queue types
- **D-09:** Queue external units use same format

### Claude's Discretion

- Exact vertical alignment for graphic row (valign=middle recommended)

### Deferred Ideas (OUT OF SCOPE)

None — discussion stayed within phase scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| QUEUE-FIX-01 | Queue units use HTML label with ASCII art graphic (═╦╩═╦═══) | Modify `buildQueueHTMLLabel()` in labels.go to add graphic row |
| QUEUE-FIX-02 | Queue external units use same HTML label format | Same function handles both; no separate function needed |
| QUEUE-FIX-03 | Queue label is 4-row table (graphic, name, technology, description) | Update rowCount calculation to include graphic row for word-wrap |
| QUEUE-FIX-04 | Remove cylinder shape and orientation from Queue units | Modify `createNode()` in converter.go to not set cylinder/orientation for Queue |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| go-graphviz | 0.2.10 (forked) | GraphViz integration | Existing project dependency for DOT generation |
| stretchr/testify | 1.11.1 | Testing assertions | Project standard for all test files |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| Go standard library | 1.26.1 | strings.Builder for HTML construction | All label building functions |

### Alternatives Considered
None needed — this is a targeted fix using existing patterns.

**Installation:**
No new dependencies required.

## Architecture Patterns

### Current Code Structure
```
internal/render/
├── labels.go        # HTML label builders (buildQueueHTMLLabel, buildDbHTMLLabel, etc.)
├── converter.go     # GraphViz node creation (createNode sets shape/orientation)
├── wrap.go          # Word-wrapping utilities (labelMaxCharsNoIcon)
└── *_test.go        # Test files for render package
```

### Pattern 1: HTML Label Building (strings.Builder)
**What:** All label builders use `strings.Builder` to construct HTML table strings.
**When to use:** This pattern is already used; extend it for Queue graphic row.
**Example:**
```go
// From labels.go buildDbHTMLLabel (lines 142-165)
var sb strings.Builder
sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0">`)

// Row 1: Name (bold)
sb.WriteString(`<tr align="center"><td valign="bottom"><b>`)
sb.WriteString(wrapAndEscape(label.Name, maxChars))
sb.WriteString(`</b></td></tr>`)
```

### Pattern 2: Type Detection (graph.IsQueueType)
**What:** Helper functions determine unit type for shape/label routing.
**When to use:** Converter uses this to decide shape; labels.go uses for label routing.
**Example:**
```go
// From converter.go (lines 224-232)
if graph.IsDbType(node.Type) || graph.IsQueueType(node.Type) {
    cn.SetShape(cgraph.CylinderShape)
    if graph.IsQueueType(node.Type) {
        cn.SetOrientation(90.0)
    }
}
```

### Anti-Patterns to Avoid
- **Don't word-wrap the graphic row:** The ASCII art `═╦╩═╦═══` must remain intact (D-03)
- **Don't create separate function for Queue external:** Same function handles both (D-09)
- **Don't forget rowCount adjustment:** Word-wrap calculation must account for 4 rows now, not 3

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Word wrapping | Custom wrap logic | `wrapAndEscape()` from wrap.go | Existing implementation handles word boundaries and HTML escaping |
| Max chars calculation | Manual calculation | `labelMaxCharsNoIcon()` from wrap.go | Existing function calculates based on row count and label ratio |

**Key insight:** The fix is additive — add one row to existing label, remove two lines from converter. No new abstractions needed.

## Common Pitfalls

### Pitfall 1: Incorrect Row Count for Word Wrap
**What goes wrong:** Word-wrap calculation uses wrong row count, causing labels to be too narrow or wide.
**Why it happens:** Adding graphic row increases row count from 3 to 4.
**How to avoid:** Update `rowCount` calculation to start at 2 (graphic + name) instead of 1.
**Warning signs:** Queue labels look disproportionately sized compared to other labels.

### Pitfall 2: HTML-Escaping the Graphic
**What goes wrong:** Graphic characters `═╦╩═╦═══` get HTML-escaped, breaking the visual.
**Why it happens:** Using `wrapAndEscape()` on the graphic row.
**How to avoid:** Write graphic string directly without wrapping or escaping.
**Warning signs:** Queue labels show `&#9552;&#9574;` instead of box drawing characters.

### Pitfall 3: Cylinder Shape Still Applied
**What goes wrong:** Queue nodes still render as cylinders (vertical) despite label change.
**Why it happens:** Incomplete removal of Queue-specific shape handling in converter.
**How to avoid:** Remove Queue from the `IsDbType || IsQueueType` condition, or add else clause for Queue.
**Warning signs:** Queue units appear as vertical cylinders in output.

## Code Examples

### Target Queue Label Format
```html
<!-- Source: CONTEXT.md code_context section -->
<table border="0" cellpadding="0" cellspacing="0">
  <tr align="center"><td valign="middle">═╦╩═╦═══</td></tr>
  <tr align="center"><td valign="bottom"><b>Queue name</b></td></tr>
  <tr align="center"><td valign="middle"><i>[technology]</i></td></tr>
  <tr align="center"><td valign="top">Description</td></tr>
</table>
```

### Current buildQueueHTMLLabel (to be modified)
```go
// Source: internal/render/labels.go lines 169-214
func buildQueueHTMLLabel(label *graph.Label) string {
    if label == nil {
        return ""
    }

    // Calculate max characters for word wrapping (no icon column)
    rowCount := 1 // name
    if label.Technology != "" {
        rowCount++
    }
    if label.Description != "" {
        rowCount++
    }
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

### Current createNode shape logic (to be modified)
```go
// Source: internal/render/converter.go lines 222-232
// Set shape based on unit type
// DB and Queue use cylinder shape, all others use box
if graph.IsDbType(node.Type) || graph.IsQueueType(node.Type) {
    cn.SetShape(cgraph.CylinderShape)
    // Queue uses horizontal cylinder (90 degree rotation)
    if graph.IsQueueType(node.Type) {
        cn.SetOrientation(90.0)
    }
} else {
    cn.SetShape(cgraph.BoxShape)
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Queue with cylinder shape | Queue with HTML label | Phase 19 | Fixes rotation not working |

**Deprecated/outdated:**
- `SetOrientation(90.0)` for Queue: GraphViz doesn't support cylinder rotation — remove this call

## Open Questions

1. **Should rowCount include the graphic row for word-wrap calculation?**
   - What we know: Word-wrap uses rowCount to calculate maxChars per line
   - What's unclear: Graphic row is fixed width; should it affect text row width?
   - Recommendation: Yes, include it (rowCount starts at 2) for consistent label proportions

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + stretchr/testify v1.11.1 |
| Config file | None — tests use t.Run() pattern |
| Quick run command | `go test ./internal/render/... -run TestHTMLQueueLabel -v` |
| Full suite command | `go test ./internal/render/... -v` |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| QUEUE-FIX-01 | Queue units use HTML label with ASCII art graphic | unit | `go test ./internal/render/... -run TestHTMLQueueLabel -v` | ✅ Exists (needs update) |
| QUEUE-FIX-02 | Queue external units use same HTML label format | unit | `go test ./internal/render/... -run TestHTMLQueueLabel -v` | ✅ Exists (covers all Queue types) |
| QUEUE-FIX-03 | Queue label is 4-row table | unit | `go test ./internal/render/... -run TestHTMLQueueLabel -v` | ✅ Exists (needs update) |
| QUEUE-FIX-04 | Remove cylinder shape and orientation | unit | `go test ./internal/render/... -run TestCreateNode -v` | ❌ Wave 0 (add shape verification) |

### Sampling Rate
- **Per task commit:** `go test ./internal/render/... -v -count=1`
- **Per wave merge:** `go test ./internal/render/... -v`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/render/converter_test.go` — Add test verifying Queue nodes use box shape, not cylinder
- [ ] `internal/render/labels_test.go` — Update `TestHTMLQueueLabel` to verify ASCII graphic row presence

*(Existing test infrastructure covers basic label structure; updates needed for new graphic row)*

## Sources

### Primary (HIGH confidence)
- `internal/render/labels.go` - Current buildQueueHTMLLabel implementation (lines 169-214)
- `internal/render/converter.go` - Current createNode shape logic (lines 222-232)
- `.planning/phases/19-queue-label-fix/19-CONTEXT.md` - Locked decisions and code examples

### Secondary (MEDIUM confidence)
- `internal/render/labels_test.go` - Existing test patterns for HTML labels
- `internal/render/wrap.go` - Word-wrap utilities

### Tertiary (LOW confidence)
None — all findings verified against source code.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - No new dependencies, existing patterns well understood
- Architecture: HIGH - Code locations identified, change scope clear
- Pitfalls: HIGH - Based on analysis of existing code patterns

**Research date:** 2026-03-24
**Valid until:** 30 days (stable Go codebase, no external API dependencies)
