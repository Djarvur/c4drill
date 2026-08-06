# Requirements: C4Drill

**Defined:** 2026-03-29
**Core Value:** Transform simple TOML architecture descriptions into professional C4 diagrams without manual drawing.

## v1.8 Requirements

### View Scoping

- [x] **VIEW-01**: C1 diagram shows only top-level units — no nested subunits appear as nodes
- [x] **VIEW-02**: Links to deeply nested targets resolve to the nearest visible ancestor in the current view level
- [x] **VIEW-03**: C2 diagram auto-generated for each system/box with subunits (written to `{basename}/{unit}.{format}`)
- [x] **VIEW-04**: C3 diagram auto-generated for each container with subunits (written to `{basename}/{unit}.{format}`)
- [x] **VIEW-05**: `properties.expanded` controls which top-level units appear expanded (as clusters) in C1

### Edge Resolution

- [x] **EDGE-01**: Edges from nested subunits to targets outside the current view resolve to the visible ancestor
- [x] **EDGE-02**: Duplicate edges (multiple sub-links resolving to same ancestor pair) collapse into a single edge

### Backward Compatibility

- [x] **COMPAT-01**: Existing TOML files without `properties.expanded` generate correct C1 with all units collapsed
- [x] **COMPAT-02**: `--expanded` flag continues to produce single all-nested diagram unchanged

## v1.0 Requirements (Shipped)

### Parsing & Validation

- ✓ **PARSE-01**: Parse TOML input file with C4 model definition
- ✓ **PARSE-02**: Validate model integrity (references, type rules, subunit constraints)

### Rendering

- ✓ **RENDER-01**: Generate GraphViz DOT output for C1-C3 layers
- ✓ **RENDER-02**: Render SVG output via go-graphviz
- ✓ **RENDER-03**: Support collapsed/expanded unit views
- ✓ **RENDER-04**: Generate explore links for drilling into nested structures
- ✓ **RENDER-05**: Support all unit types (person, system, db, queue, box, container, component + externals)
- ✓ **RENDER-06**: Apply styling: colors, borders, edge routing styles
- ✓ **RENDER-07**: Render HTML output (`-f html`) — SVG inlined in a self-contained HTML document with a JS nav shim for Safari/WebKit compatibility (which silently ignores SVG `<a>` hyperlinks). Default format remains svg.

### CLI

- ✓ **CLI-01**: Single CLI command interface

## Out of Scope

| Feature | Reason |
|---------|--------|
| C4 Layer 4 (Code) | Class/function level diagrams not needed |
| Go library/module | Pure CLI tool, no library interface |
| Manual positioning | Rely on GraphViz auto-layout |
| JSON error output | CLI errors sufficient |
| Live editing/watch mode | Single-shot rendering |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| VIEW-01 | Phase 1 | Complete |
| VIEW-02 | Phase 1 | Complete |
| VIEW-03 | Phase 2 | Complete |
| VIEW-04 | Phase 2 | Complete |
| VIEW-05 | Phase 2 | Complete |
| EDGE-01 | Phase 1 | Complete |
| EDGE-02 | Phase 1 | Complete |
| COMPAT-01 | Phase 3 | Complete |
| COMPAT-02 | Phase 3 | Complete |

**Coverage:**
- v1.8 requirements: 9 total
- Mapped to phases: 9
- Unmapped: 0 ✓

---
*Requirements defined: 2026-03-29*
*Last updated: 2026-03-29 after v1.8 requirements defined*
