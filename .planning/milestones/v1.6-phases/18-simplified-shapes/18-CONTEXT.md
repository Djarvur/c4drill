# Phase 18: Simplified Shapes - Context

**Gathered:** 2026-03-24
**Status:** Ready for planning

<domain>
## Phase Boundary

Remove SVG icon system and use native GraphViz shapes for DB/Queue units. Simplify labels: Person uses 👤 emoji in 2-column table, System/Box/Container/Component use simple 3-row tables without icon column.

**Scope includes:**
1. Remove entire icon system (icons package, IconExtractor, SVG postprocessing)
2. DB units → `shape=cylinder` (native GraphViz)
3. Queue units → `shape=cylinder` with 90° rotation
4. Person labels → 2-column table with 👤 emoji (`POINT-SIZE="32"`)
5. System/Box/Container/Component labels → 3-row table (name, technology, description)
6. Keep word-wrap functionality from Phase 17

**Out of scope:**
- Adding new unit types
- Changing edge rendering
- Modifying cluster/container label formats

</domain>

<decisions>
## Implementation Decisions

### Icon System Removal
- **D-01:** Remove `internal/render/icons/` package entirely (embed.go and all .svg files)
- **D-02:** Remove `internal/render/icon_extractor.go` and IconExtractor struct
- **D-03:** Remove `.icons/` directory generation from output
- **D-04:** Remove any SVG postprocessing logic that injects icons after rendering

### DB Shape
- **D-05:** DB units use `shape=cylinder` (native GraphViz)
- **D-06:** DB external units also use cylinder shape
- **D-07:** DB label is simple 3-row HTML table (name, technology, description) — no icon column

### Queue Shape
- **D-08:** Queue units use `shape=cylinder` with 90° orientation
- **D-09:** Queue external units also use rotated cylinder shape
- **D-10:** Queue label is simple 3-row HTML table (name, technology, description) — no icon column

### Person Label
- **D-11:** Person units keep 2-column table layout
- **D-12:** First column contains 👤 emoji with `<font POINT-SIZE="32">`
- **D-13:** Second column contains name and description rows
- **D-14:** Person external units use same label format

### System/Box/Container/Component Labels
- **D-15:** All use simple 3-row table (name, technology, description)
- **D-16:** No icon column — single-column layout
- **D-17:** External variants use same format

### Word Wrap (Preserved)
- **D-18:** Keep word-wrap functionality from Phase 17
- **D-19:** `--label-ratio` flag continues to work

### Claude's Discretion
- Exact cylinder orientation syntax (may need `rankdir` or other attrs)
- Fallback if GraphViz doesn't support rotated cylinder
- Handling of iconReserve constant removal

</decisions>

<specifics>
## Specific Ideas

- User confirmed `<font POINT-SIZE="32">` for Person emoji (GraphViz-compatible font sizing)
- Queue cylinder rotation: use native GraphViz orientation support
- All icon-related code removed: package, extractor, column, postprocessing
- DB/Queue labels use same 3-row format as System/Box

</specifics>

<canonical_refs>
## Canonical References

### Label Structure
- `internal/render/labels.go` — HTML label builders for all unit types (to be modified)
- `internal/graph/label.go` — Label struct with Name, Technology, Description fields

### Icon System (To Be Removed)
- `internal/render/icons/embed.go` — SVG icon embedding (REMOVE)
- `internal/render/icons/*.svg` — Icon templates (REMOVE)
- `internal/render/icon_extractor.go` — IconExtractor struct (REMOVE)
- `internal/render/svg_icons.go` — SVG icon handling (REMOVE)
- `internal/render/dot_icons.go` — DOT icon handling (REMOVE)

### Shape Handling
- `internal/graph/shapes.go` — Shape and style definitions for unit types
- `internal/render/converter.go` — Node creation with shape and label

### Word Wrap (To Keep)
- `internal/render/wrap.go` — Word-wrap implementation
- `internal/render/wrap_test.go` — Wrap tests

### Prior Phase Context
- `.planning/phases/17-*/17-CONTEXT.md` — Word-wrap decisions (keep)
- `.planning/phases/16-*/16-CONTEXT.md` — Icon system (remove)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/render/wrap.go` — Word-wrap functions, keep unchanged
- `internal/render/labels.go` — Label builders, need modification to remove icon column
- `internal/graph/shapes.go` — Shape definitions, need cylinder for DB/Queue

### Established Patterns
- HTML labels use `<table>` with `<tr>`/`<td>` structure (no nested padding tables)
- `cellpadding="0" cellspacing="0"` for clean label borders
- Label types dispatch via `buildHTMLLabelForType()` switch statement
- Shape assignment in `shapes.go:GetStyleForType()`
- Font sizing uses `POINT-SIZE` attribute (GraphViz-compatible)

### Integration Points
- `buildPersonHTMLLabel()` — Keep 2-column, use 👤 emoji (no img tag)
- `buildDbHTMLLabel()`, `buildQueueHTMLLabel()` — Change to 3-row single-column
- `buildSystemHTMLLabel()`, `buildContainerHTMLLabel()`, `buildComponentHTMLLabel()` — Change to 3-row single-column (remove SYS/CONT/COMP label column)
- Converter needs to set `shape=cylinder` for DB/Queue types
- Remove iconReserve constant and all IconExtractor usage

### Files to Delete
- `internal/render/icons/embed.go`
- `internal/render/icons/person.svg`
- `internal/render/icons/db.svg`
- `internal/render/icons/pipe.svg`
- `internal/render/icons/system.svg`
- `internal/render/icons/container.svg`
- `internal/render/icons/component.svg`
- `internal/render/icons/icons_test.go`
- `internal/render/icon_extractor.go`
- `internal/render/icon_extractor_test.go`
- `internal/render/icons_integration_test.go`
- `internal/render/svg_icons.go`
- `internal/render/dot_icons.go`

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 18-simplified-shapes*
*Context gathered: 2026-03-24*
