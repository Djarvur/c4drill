# Phase 16: Embed and render level-colored SVG icons for units - Context

**Gathered:** 2026-03-20
**Status:** Ready for planning

<domain>
## Phase Boundary

Replace text-based Unicode icons (emojis like `\U0001F464`, `\u26C1`) with embedded SVG images that match each unit's C4 level colors. Icons are extracted to `{output}/.icons/` directory during rendering and referenced via `<img>` tags in HTML table labels.

**Scope includes:**
1. Move SVG icons from `data/` to embeddable location in renderer
2. Apply C4 level colors (C1/C2/C3) to icon strokes dynamically
3. All icons use consistent `type-{hexcolor}.svg` naming (e.g., `person-3C7FC0.svg`)
4. Extract icons on-demand during rendering
5. Reference icons via `<img>` tags with relative paths

**Out of scope:**
- Adding new icon types beyond the existing 6
- PNG conversion (keep SVG format)
- Icon size/configuration in TOML (fixed at 32x32)

</domain>

<decisions>
## Implementation Decisions

### Icon Storage Location
- **D-01:** Use Go's `embed.FS` to embed SVG icons in the renderer package
- Icons travel with the executable for single-binary distribution
- Source SVGs moved from `data/` to `internal/render/icons/`

### Icon Extraction Strategy
- **D-02:** Extract icons on-demand per diagram type
- Only extract icons for types that appear in the current diagram
- Extract to `{output-base}/.icons/` directory (e.g., `diagram/.icons/`)
- Keeps output clean, no unused icon files

### Icon Color Mapping
- **D-03:** Dynamic `currentColor` replacement at render time
- Template SVGs use `currentColor` placeholder
- Replace `currentColor` with actual hex color based on unit's C4 level
- **Consistent naming:** Always `type-{hexcolor}.svg` format
  - Standard colors: `person-3C7FC0.svg`, `db-3C7FC0.svg`, `pipe-78A8D8.svg`
  - Custom colors: `person-FF0000.svg`, `db-00FF00.svg`
  - No special naming for standard vs custom — always use hex color

### Image Reference in HTML Labels
- **D-04:** Use `<img src='...'>` tags in HTML table labels
- Path is relative to rendered SVG file: `.icons/person-3C7FC0.svg`
- GraphViz supports `<img>` in HTML labels

### Icon Size and Placement
- **D-05:** Icons rendered at 32x32 pixels
- **D-06:** Icon column with rowspan (same pattern as SYS/CONT/COMP text labels)
- Icon spans all content rows (name, technology, description)

### Icon Scope
- **D-07:** All 6 unit types display icons: person, db, pipe, system, container, component
- External variants use external border colors (gray tones)
- Box type uses container icon

### Claude's Discretion
- Exact icon column width and padding
- Fallback behavior if icon extraction fails
- Whether to cache extracted icons across multiple renders

</decisions>

<specifics>
## Specific Ideas

- SVG icons in `data/` already use `currentColor` — perfect for color inheritance
- Existing icons: person.svg, db.svg, pipe.svg, system.svg, container.svg, component.svg
- Color palette already defined in `internal/model/colors.go` (SystemBorder, ContainerBorder, ComponentBorder, etc.)
- `IconForType()` in `internal/graph/shapes.go` currently returns Unicode — will be replaced

</specifics>

<canonical_refs>
## Canonical References

### Icon Design
- `data/*.svg` — Source SVG templates with `currentColor` placeholders

### Color Definitions
- `internal/model/colors.go` — C4 level border colors (SystemBorder, ContainerBorder, ComponentBorder) and external variants

### Current Icon System
- `internal/graph/shapes.go:26-42` — `IconForType()` function (to be replaced)
- `internal/render/labels.go:156-267` — HTML label builders (to add icon column)

### Rendering Pipeline
- `internal/render/converter.go:14-31` — `buildHTMLLabelForType()` dispatcher
- `internal/render/converter.go:163-187` — Node creation with HTML labels

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/model/colors.go` — Color constants already defined per C4 level
- `internal/graph/shapes.go:GetStyleForType()` — Returns border color for any unit type
- HTML table label pattern in `labels.go` — Well-established rowspan/column pattern

### Established Patterns
- HTML labels use `<table>` with `<tr>`/`<td>` structure
- Label types dispatch via `buildHTMLLabelForType()` switch statement
- `graph.IsPersonType()`, `graph.IsDbType()`, etc. — Type checking helpers exist

### Integration Points
- `buildPersonHTMLLabel()`, `buildDbHTMLLabel()`, etc. — Each needs icon column added
- Icon extraction should happen in converter before node creation
- Output path available from `RenderDOT()` context

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 16-embed-and-render-level-colored-svg-icons-for-units*
*Context gathered: 2026-03-20*
