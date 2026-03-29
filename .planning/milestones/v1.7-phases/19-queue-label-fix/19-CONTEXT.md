# Phase 19: Queue Label Fix - Context

**Gathered:** 2026-03-24
**Status:** Ready for planning

<domain>
## Phase Boundary

Fix Queue rendering - GraphViz cylinder shapes don't support rotation. Revert Queue units to HTML labels with ASCII art graphic (═╦╩═╦═══) in a 4-row table format.

**Scope includes:**
1. Remove `shape=cylinder` from Queue units in converter
2. Remove `SetOrientation(90.0)` call for Queue units
3. Update Queue label builder to add ASCII art graphic row
4. Queue label format: 4-row table (graphic, name, technology, description)

**Out of scope:**
- Changing DB cylinder shape (works correctly)
- Changing Person emoji labels (work correctly)
- Changing System/Box/Container/Component labels (work correctly)
- Modifying word-wrap functionality

</domain>

<decisions>
## Implementation Decisions

### ASCII Graphic Pattern
- **D-01:** Use `═╦╩═╦═══` as the Queue graphic (classic pipe pattern from Phase 13)
- **D-02:** Graphic row uses center alignment
- **D-03:** Graphic row is NOT word-wrapped (fixed pattern)

### Queue Label Structure
- **D-04:** 4-row HTML table: graphic, name, technology, description
- **D-05:** Name/technology/description rows ARE word-wrapped (existing wrap logic)
- **D-06:** Same table attributes as other labels: `border="0" cellpadding="0" cellspacing="0"`

### Shape and Converter Changes
- **D-07:** Queue units use `shape=box, style=rounded` (same as System/Person)
- **D-08:** Remove `SetOrientation(90.0)` call from converter for Queue types
- **D-09:** Queue external units use same format

### Claude's Discretion
- Exact vertical alignment for graphic row (valign=middle recommended)

</decisions>

<specifics>
## Specific Ideas

- Graphic pattern `═╦╩═╦═══` matches the horizontal pipe/tube representation from Phase 13
- 4-row table: Row 1 = graphic, Row 2 = name (bold), Row 3 = technology (italic in brackets), Row 4 = description

</specifics>

<canonical_refs>
## Canonical References

### Queue Label Implementation
- `internal/render/labels.go` — buildQueueHTMLLabel function (to be modified)
- `internal/render/converter.go` — Shape and orientation for Queue types (to be modified)

### Word Wrap (Keep)
- `internal/render/wrap.go` — Word-wrap implementation (no changes needed)

### Prior Phase Context
- `.planning/phases/13-refined-html-labels/13-CONTEXT.md` — Original Queue label format reference
- `.planning/phases/18-simplified-shapes/18-CONTEXT.md` — Current (broken) Queue implementation

</canonical_refs>

<code_context>
## Existing Code Insights

### Current Queue Label (Phase 18)
```go
// buildQueueHTMLLabel - 3-row table: name, technology, description
// NO graphic row - this is what needs to be added
```

### Target Queue Label (Phase 13 style)
```html
<table border="0" cellpadding="0" cellspacing="0">
  <tr align="center"><td valign="middle">═╦╩═╦═══</td></tr>
  <tr align="center"><td valign="bottom"><b>Queue name</b></td></tr>
  <tr align="center"><td valign="middle"><i>[technology]</i></td></tr>
  <tr align="center"><td valign="top">Description</td></tr>
</table>
```

### Converter Changes
- Line ~227-229 in converter.go sets `SetOrientation(90.0)` for Queue — REMOVE
- Queue should fall through to `shape=box, style=rounded` like System/Person

### Integration Points
- `buildQueueHTMLLabel()` in labels.go — Add graphic row as first row
- `buildQueueExternalHTMLLabel()` in labels.go — Same change
- Converter shape handling for Queue types — Remove cylinder/orientation

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 19-queue-label-fix*
*Context gathered: 2026-03-24*
