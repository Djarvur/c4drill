# Phase 17: Word-Wrapped Labels for Credit Card Proportions - Context

**Gathered:** 2026-03-23
**Status:** Ready for planning

<domain>
## Phase Boundary

Add word-wrapping to HTML label cells so unit shapes approximate credit card proportions. Long names/descriptions currently stretch units horizontally; wrapping will constrain width to achieve configurable width:height ratios.

**Scope includes:**
1. Word-wrap all label fields (name, technology, description)
2. Calculate target width dynamically based on content height
3. Configurable ratio via CLI flag and environment variable
4. Hybrid wrapping: word-based with forced break for long words

**Out of scope:**
- Changing icon sizes or column widths
- Adding new label fields
- Affecting cluster/container label wrapping (only unit labels)

</domain>

<decisions>
## Implementation Decisions

### Target Proportions
- **D-01:** Default ratio is **8/5 = 1.6:1** (width:height)
- Close to credit card ratio (1.586:1) but simpler to reason about
- User can configure different ratio via `--label-ratio` or `C4DRILL_LABEL_RATIO`

### Width Strategy
- **D-02:** Dynamic width calculation based on content height
- Count rows in label (name + technology + description)
- Calculate target width = height × ratio
- Account for icon column (36px fixed) in total width
- Text column width = total width - icon column width

### Wrapping Approach
- **D-03:** Hybrid wrapping strategy
- **Primary:** Word-based — break at word boundaries (spaces)
- **Fallback:** Character-based — force break if single word exceeds max line length
- Prevents extremely long URLs/identifiers from breaking layout

### Fields to Wrap
- **D-04:** All label fields are wrapped: name, technology, description
- Consistent appearance across all label content
- No special handling per field

### Configuration
- **D-05:** Ratio configurable via both CLI flag and environment variable
- CLI flag: `--label-ratio=1.6` (takes precedence)
- Environment: `C4DRILL_LABEL_RATIO=1.6` (fallback)
- Default: 1.6 if neither specified

### Claude's Discretion
- Exact pixel calculations for height estimation
- Maximum line length before forced character break
- Whether to expose max line length as configurable parameter

</decisions>

<specifics>
## Specific Ideas

- User specified "8/5" as the default ratio — simple fraction, easy to remember
- Dynamic width means each unit's width is derived from its height × ratio
- Icon column (36px) is fixed; wrapping affects text column only

</specifics>

<canonical_refs>
## Canonical References

### Label Structure
- `internal/render/labels.go` — HTML label builders for all unit types
- `internal/graph/label.go` — Label struct with Name, Technology, Description fields

### Icon Integration
- `internal/render/labels.go:14-17` — iconReserve constant and icon column width (36)
- Phase 16 CONTEXT.md — Icon sizing (32x32) and column width (36)

### CLI Configuration
- `cmd/c4drill/root.go` — CLI flag handling

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/render/labels.go` — All label builder functions (buildPersonHTMLLabel, buildDbHTMLLabel, etc.)
- Each builder already handles optional fields (technology, description)
- Icon column pattern well-established (width="36", rowspan)

### Established Patterns
- HTML labels use `<table>` with `<td>` cells
- Text content uses `html.EscapeString()` for safety
- Labels built from `graph.Label` struct

### Integration Points
- Label builders need to wrap text before escaping
- Need function to calculate target width from row count and ratio
- Need word-wrap function with character fallback
- CLI flag parsing in `root.go` needs `--label-ratio` flag
- Environment variable reading for `C4DRILL_LABEL_RATIO`

### Implementation Notes
- Row count depends on which fields are present:
  - Person: 1-2 rows (name, optional description)
  - Others: 1-3 rows (name, optional technology, optional description)
- Height estimation needs to account for font size and row padding
- GraphViz HTML tables support `width` attribute on `<td>` cells

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 17-units-labels-cells-must-be-word-wrapped-to-make-the-unit-shape-proportions-as-close-as-possible-to-credit-card-proportions*
*Context gathered: 2026-03-23*
