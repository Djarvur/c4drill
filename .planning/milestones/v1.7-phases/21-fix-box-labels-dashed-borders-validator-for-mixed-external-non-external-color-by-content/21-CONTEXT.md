# Phase 21: Box Label, Border, Validator, and Color Fixes - Context

**Gathered:** 2026-03-24
**Status:** Ready for planning
**Source:** Auto-captured from user requirements

<domain>
## Phase Boundary

Fix box unit rendering and validation:
1. Remove unnecessary curly brackets from box labels
2. Make box borders dashed by default
3. Add validator to prevent mixing external and non-external units in C1 boxes
4. Color C1 boxes based on their contents (grey for externals, dark blue for non-externals)

This affects the render and validator subsystems only.

</domain>

<decisions>
## Implementation Decisions

### Box Labels
- Box labels should NOT have curly brackets (`{...}`)
- Use HTML label builder (like `buildContainerHTMLLabel`) instead of `buildRecordLabel`
- Box label format: name (bold) / [technology] italic / description (same as container/component)
- Applies to all box types: TypeBox, TypeContainerBox, TypeComponentBox

### Box Borders
- Box borders must be **dashed** by default (not solid)
- Applies to both collapsed and expanded views
- Applies to all box types at all levels (C1, C2, C3)
- Border width remains default (no change needed)

### C1 Box Validation
- C1 box (TypeBox) must NOT contain both external AND non-external units simultaneously
- Add new validation rule: `ValidateBoxMixedContents`
- Error message: `box "{path}" cannot contain both external and non-external units`
- External units are: personExternal, systemExternal, dbExternal, queueExternal
- Non-external units are: person, system, db, queue, box

### C1 Box Color by Content
- C1 box containing ANY external units → grey border color (PersonExternalBorder: #808080)
- C1 box containing ONLY non-external units → dark blue border color (PersonBorder: #073B6F)
- This requires analyzing box contents to determine color
- Color applies to both collapsed and expanded views

### Claude's Discretion
- Exact implementation of content analysis (iterate subunits to check for external)
- Whether to add helper function `hasExternalSubunits(box *UnitInfo, index) bool`
- Test file organization (single test file vs multiple)

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Rendering
- `internal/render/labels.go` — HTML label builders (buildContainerHTMLLabel pattern to follow)
- `internal/render/converter.go` — buildHTMLLabelForType dispatch, buildStyleString for border styles
- `internal/graph/shapes.go` — getLevelStyle, getExternalStyle functions, NodeStyle struct

### Validation
- `internal/validator/rules.go` — Validation rule patterns (ValidateNestingHierarchy as reference)
- `internal/validator/validator.go` — Validator integration

### Model
- `internal/model/unit.go` — Unit type constants, border color constants (PersonBorder, PersonExternalBorder)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `buildContainerHTMLLabel` — Same 3-row HTML table format box labels should use
- `getLevelStyle` / `getExternalStyle` — Pattern for returning NodeStyle with border settings
- `ValidateNestingHierarchy` — Pattern for validation rule that checks parent-child relationships
- `IsExternalType` — Helper to check if a type is external

### Established Patterns
- HTML labels use `<table>` with `<tr align="center">` rows
- Border style set via `BorderStyle` field in `NodeStyle` struct
- Validators return `ValidationErrors` slice (not fail-fast)
- External types detected via suffix "External" in type name or explicit checks

### Integration Points
- `buildHTMLLabelForType` in converter.go:15-37 — Add box type case before default
- `getLevelStyle` in shapes.go:124-157 — Modify for box types to return dashed style
- Validator rules called from `Validate` function in validator.go
- Graph builder creates nodes/clusters with style from `GetStyleForType`

### Key Files to Modify
1. `internal/render/labels.go` — Add `buildBoxHTMLLabel` function
2. `internal/render/converter.go` — Add box case to `buildHTMLLabelForType`
3. `internal/graph/shapes.go` — Modify `getLevelStyle` for dashed box borders
4. `internal/validator/rules.go` — Add `ValidateBoxMixedContents` rule
5. `internal/graph/builder.go` — May need to pass box contents info for color determination

</code_context>

<specifics>
## Specific Ideas

- User explicitly stated: "box label has unnecessary curly brackets"
- User explicitly stated: "box border must be dashed by default, in expanded and collapsed view"
- User explicitly stated: "c1 box must NOT contain external and non-external units in the same time"
- User explicitly stated: "c1 box contain externals default color must be grey as externals are"
- User explicitly stated: "c1 box for non-externals is dark blue"

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 21-fix-box-labels-dashed-borders-validator-for-mixed-external-non-external-color-by-content*
*Context gathered: 2026-03-24 via auto capture*
