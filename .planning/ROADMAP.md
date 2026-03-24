# Roadmap: C4Drill

## Overview

C4Drill transforms TOML architecture definitions into professional C4 diagrams through a compiler-style pipeline: parse, validate, generate views, build graphs, render to DOT/SVG.

## Milestones

- **v1.0 Initial Release** -- Phases 1-6 (shipped 2021-03-10)
- **v1.1 AI-Ready** -- Phases 7-10 (shipped 2021-03-13)
- **v1.2 Bug Fixes** -- Phases 11-13 (shipped 2021-03-14)
- **v1.3 Validation Enhancements** -- Phase 14 (shipped 2021-03-17)
- **v1.4 Edge Coloring** -- Phase 15 (shipped 2021-03-18)
- **v1.5 SVG Icons** -- Phase 16-17 (shipped 2021-03-23)
- **v1.6 Simplified Shapes** -- Phase 18 (shipped 2021-03-24)
- **v1.7 Queue Label Fix** -- Phase 19 (in progress)

## Phases

<details>
<summary>v1.0 Initial Release (Phases 1-6) -- SHIPPED 2021-03-10</summary>

- [x] Phase 1: Foundation & Model -- completed 2021-03-09
- [x] Phase 2: Validation -- completed 2021-03-09
- [x] Phase 3: Views & Graphs -- completed 2021-03-09
- [x] Phase 4: Rendering & Output -- completed 2021-03-10
- [x] Phase 5: Navigation -- completed 2021-03-10
- [x] Phase 6: CLI & Polish -- completed 2021-03-10

See `.planning/milestones/v1.0-ROADMAP.md` for full phase details.

</details>

### v1.1 AI-Ready (In Progress)

**Milestone Goal:** Make C4Drill AI-friendly with documentation, single-view expanded diagrams, and stricter validation

- [x] **Phase 7: AI Documentation** -- TOML language manual for AI assistants
- [x] **Phase 8: All-Expanded Mode** -- Single-view expanded rendering with cross-level edges (completed 2021-03-10)
- [x] **Phase 9: No Orphan Units** -- Validation rule requiring all units to be linked (completed 2021-03-11)
- [x] **Phase 10: Link List Format** -- Change Links/LinksFrom from maps to lists with explicit target/source (completed 2021-03-13)

---

## Phase Details

### Phase 7: AI Documentation
**Goal**: AI assistants can generate valid C4Drill TOML without prior training
**Depends on**: Nothing (documentation-only, independent of Phase 8)
**Requirements**: AIDOC-01, AIDOC-02, AIDOC-03, AIDOC-04, AIDOC-05
**Success Criteria** (what must be TRUE):
  1. User can read SKILL.md and understand the complete TOML schema with all unit types, fields, and link syntax
  2. User can copy any example TOML from skill/examples/ and have it parse and validate successfully
  3. AI assistants given the skill produce syntactically valid TOML files
  4. All validation rules are documented with clear explanations of what triggers each error
  5. CI validates all TOML examples against the actual parser to prevent drift
**Plans:** 2/2 plans complete

Plans:
- [x] 07-01-PLAN.md -- Create skill package (SKILL.md + examples)
- [x] 07-02-PLAN.md -- Add CI validation for examples

### Phase 8: All-Expanded Mode
**Goal**: Users can generate a single diagram showing all units expanded with cross-level edges
**Depends on**: Nothing (independent of Phase 7, can run in parallel)
**Requirements**: EXPD-01, EXPD-02, EXPD-03, EXPD-04, EXPD-05
**Success Criteria** (what must be TRUE):
  1. User can run `c4drill input.toml --expanded` and receive expanded diagram output
  2. All units in the model appear as nested clusters in a single diagram
  3. Edges between units at different nesting depths are visible in the diagram
  4. Output file is saved with `{basename}.expanded.{ext}` naming convention
  5. Existing C1/C2/C3 view generation produces identical output when `--expanded` is not used
**Plans:** 2/2 plans complete

Plans:
- [x] 08-01-PLAN.md -- Add core data structures and view generation
- [x] 08-02-PLAN.md -- Add CLI flag integration and recursive cluster building

### Phase 9: No Orphan Units
**Goal**: All units in the architecture must be connected via links (no isolated units)
**Depends on**: Nothing (independent, can run in parallel)
**Requirements**: VAL-01, VAL-02
**Success Criteria** (what must be TRUE):
  1. Validator rejects TOML files containing units with no incoming or outgoing links
  2. Error message clearly lists all unlinked (orphan) units by name
  3. Existing valid TOML files continue to validate successfully
**Plans:** 1/1 plans complete

Plans:
- [x] 09-01-PLAN.md -- Add ValidateOrphanUnits rule with tests and integration

### Phase 10: Link List Format
**Goal**: Change Links and LinksFrom from maps to lists with explicit peer field
**Depends on**: Nothing (breaking change, independent)
**Requirements**: LLIST-01, LLIST-02, LLIST-03, LLIST-04
**Success Criteria** (what must be TRUE):
  1. Links is a list of Link objects with explicit `peer` field (the target)
  2. LinksFrom is a list of Link objects with explicit `peer` field (the source)
  3. Multiple links to the same peer are supported
  4. All existing TOML examples updated to new syntax
  5. Parser, validator, graph builder, and all tests updated
**Plans:** 3/3 plans complete

Plans:
- [x] 10-01-PLAN.md -- Model + Parser + testdata + parser tests
- [x] 10-02-PLAN.md -- Documentation (SKILL.md + examples)
- [x] 10-03-PLAN.md -- Validator + View + Graph + their tests

### Phase 11: Unit Shape and Attributes
**Goal**: Fix unit rendering to use record shapes for collapsed units (not HTML) with transparent backgrounds for all units
**Depends on**: Phase 10
**Requirements**: SHAPE-01, SHAPE-02
**Success Criteria** (what must be TRUE):
  1. Collapsed units render with `shape=record` (not HTML labels)
  2. Expanded units render as clusters (subgraphs)
  3. All units have transparent backgrounds (no fill colors)
  4. Icons and styling remain differentiated by type and level
  5. All tests pass with new shape logic
**Plans:** 1/1 plans complete

Plans:
- [x] 11-01-PLAN.md -- Shape logic (ShapeRecord) and fill logic (transparent backgrounds)

---

## Progress

**Execution Order:**
Phases 7, 8, and 9 are independent and can run in parallel.

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. Foundation & Model | v1.0 | 3/3 | Complete | 2021-03-09 |
| 2. Validation | v1.0 | 2/2 | Complete | 2021-03-09 |
| 3. Views & Graphs | v1.0 | 4/4 | Complete | 2021-03-09 |
| 4. Rendering & Output | v1.0 | 3/3 | Complete | 2021-03-10 |
| 5. Navigation | v1.0 | 2/2 | Complete | 2021-03-10 |
| 6. CLI & Polish | v1.0 | 2/2 | Complete | 2021-03-10 |
| 7. AI Documentation | v1.1 | 2/2 | Complete | 2021-03-10 |
| 8. All-Expanded Mode | v1.1 | 2/2 | Complete | 2021-03-10 |
| 9. No Orphan Units | v1.1 | 1/1 | Complete | 2021-03-11 |
| 10. Link List Format | v1.1 | 3/3 | Complete | 2021-03-13 |
| 11. Unit Shape and Attributes | v1.1 | 1/1 | Complete | 2021-03-13 |
| 12. HTML labels for all unit types | v1.2 | 2/2 | Complete | 2021-03-13 |
| 13. Refined HTML Labels | v1.2 | Complete    | 2021-03-14 | 2021-03-14 |
| 14. Nesting Validation | v1.3 | Complete    | 2021-03-17 | 2021-03-17 |
| 15. Edge Coloring | v1.4 | 1/1 | Complete | 2021-03-18 |
| 16. SVG Icons | v1.5 | 1/1 | Complete | 2021-03-18 |
| 17. Word-wrapped Labels | v1.5 | 1/1 | Complete | 2021-03-23 |
| 18. Simplified Shapes | v1.6 | Complete    | 2021-03-24 | 2021-03-24 |
| 19. Queue Label Fix | v1.7 | 0/1 | Planning | — |
| 20. Helvetica Font | v1.7 | 1/1 | Complete | 2021-03-24 |
| 21. Box Fixes | v1.7 | Complete    | 2026-03-24 | 2026-03-24 |

### Phase 12: HTML labels for all unit types

**Goal:** Convert all unit type labels from record-style format to HTML table format with specific layouts per unit type
**Requirements**: HTML-01, HTML-02
**Depends on:** Phase 11
**Plans:** 2/2 plans complete
**Status:** Complete

Plans:
- [x] 12-00-PLAN.md -- Add test file and stub HTML label builder functions
- [x] 12-01-PLAN.md -- HTML label builders with type-specific formats (Person icon, DB icon, Queue 4-row, SYS/CONT/COMP labels)

### Phase 13: Refined HTML Labels

**Goal:** Fix bugs in expanded view where nested containers are missing, plus label refinements
**Requirements**: BUG-01, BUG-02, BUG-03, TEST-01, REFINED-01, REFINED-02, REFINED-03
**Depends on:** Phase 12
**Plans:** 1/1 plans complete
**Status:** Complete

**HTML Label Specifications:**

All units: `shape=box, style=rounded`

Person label:
```html
<table border="0" cellpadding="0" cellspacing="0">
  <tr align=center>
    <td rowspan=2 valign=middle><font size="+4">👤</font></td>
    <td valign=bottom><b>User name</b></td>
  </tr>
  <tr align=center>
    <td valign=top>Description</td>
  </tr>
</table>
```

DB label:
```html
<table border="0" cellpadding="0" cellspacing="0">
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

Queue label:
```html
<table border="0" cellpadding="0" cellspacing="0">
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

System label:
```html
<table border="0" cellpadding="0" cellspacing="0">
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

Container label:
```html
<table border="0" cellpadding="0" cellspacing="0">
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

Component label:
```html
<table border="0" cellpadding="0" cellspacing="0">
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

Plans:
- [x] 13-01-PLAN.md -- Fix nested cluster rendering and refine HTML labels

### Phase 14: Nesting Validation

**Goal:** Enforce C4 model hierarchy by validating that units are nested at the correct level
**Requirements:** NEST-01, NEST-02, NEST-03
**Depends on:** Nothing (standalone validation enhancement)
**Plans:** 1/1 plans complete
**Status:** Complete

**Nesting Rules:**

- Top level: C1 types only (person, system, db, queue, box)
- Inside system/box: C2 types only (container, containerDb, containerQueue)
- Inside container: C3 types only (component, componentDb, componentQueue)
- C3 types: No subunits (leaf nodes)

Plans:
- [x] 14-01-PLAN.md -- Add ValidateNestingHierarchy rule with tests

### Phase 15: the edge must be the same color as the source unit

**Goal:** Edges render with color matching the source unit's border color
**Requirements:** EDGE-01, EDGE-02, EDGE-03
**Depends on:** Phase 14
**Plans:** 1 plan

**Edge Coloring Rules:**

- D-01: Edge color comes from the source unit's border color
- D-02: Edge labels (technology, description) match the edge color
- D-03: If link.color is explicitly set in TOML, it overrides source border color

Plans:
- [x] 15-01-PLAN.md -- Add Color field to Edge struct and apply color in converter

### Phase 16: Embed and render level-colored SVG icons for units

**Goal:** Replace text-based Unicode icons with embedded SVG images that match each unit's C4 level colors
**Requirements:** ICON-01, ICON-02, ICON-03, ICON-04, ICON-05, ICON-06
**Depends on:** Phase 15
**Plans:** 1/1 plans complete
**Status:** Complete

**Icon System Design:**

- D-01: Use Go's embed.FS to embed SVG icons in renderer package
- D-02: Extract icons on-demand per diagram type to {output-base}/.icons/
- D-03: Dynamic currentColor replacement, naming: type-{hexcolor}.svg
- D-04: Use `<img src='...'>` tags in HTML table labels
- D-05: Icons at 32x32 pixels
- D-06: Icon column with rowspan (same pattern as SYS/CONT/COMP text labels)
- D-07: All 6 unit types get icons: person, db, pipe, system, container, component

Plans:
- [x] 16-01-PLAN.md -- Create icons package, IconExtractor, update HTML labels, integrate with converter

### Phase 17: Units labels cells must be word-wrapped to make the unit shape proportions as close as possible to credit card proportions

**Goal:** Add word-wrapping to HTML label cells so unit shapes approximate credit card proportions
**Requirements:** WRAP-01, WRAP-02, WRAP-03, WRAP-04, WRAP-05
**Depends on:** Phase 16
**Plans:** 1/1 plans complete
**Status:** Complete

**Word-Wrapping Design:**

- D-01: Default ratio is 8/5 = 1.6:1 (width:height)
- D-02: Dynamic width calculation based on content height
- D-03: Hybrid wrapping: word-based with forced character break for long words
- D-04: All label fields wrapped: name, technology, description
- D-05: Ratio configurable via `--label-ratio` CLI flag and `C4DRILL_LABEL_RATIO` env var

Plans:
- [x] 17-01-PLAN.md -- Implement word-wrap functions, CLI flag, and integrate with HTML label builders

---

### v1.6 Simplified Shapes (Complete)

**Milestone Goal:** Replace icon system with native GraphViz shapes and simplified labels

### Phase 18: Simplified Shapes

**Status:** Complete
**Plans:** 1/1 plans complete

**Goal:** Remove SVG icons, use native GraphViz cylinder shapes for DB/Queue, simplify labels
**Requirements:** ICON-01, ICON-02, ICON-03, ICON-04, DB-01, DB-02, DB-03, QUEUE-01, QUEUE-02, QUEUE-03, PERSON-01, PERSON-02, PERSON-03, PERSON-04, LABEL-01, LABEL-02, LABEL-03, LABEL-04, WRAP-01, WRAP-02
**Depends on:** Phase 17
**Plans:** 1 plan

**Simplified Shapes Design:**

- D-01: Remove icons package (internal/icons) entirely
- D-02: Remove IconExtractor from converter
- D-03: Remove .icons/ directory generation
- D-04: Remove SVG postprocessing logic
- D-05: DB units use `shape=cylinder` (native GraphViz)
- D-06: Queue units use `shape=cylinder` with 90 rotation
- D-07: Person labels: 2-column table with emoji (font size +4)
- D-08: System/Box/Container/Component labels: 3-row table (name, technology, description)
- D-09: Keep word-wrap functionality from Phase 17

Plans:
- [x] 18-01-PLAN.md -- Remove icon system, add native shapes, update label builders

---

### v1.7 Queue Label Fix (In Progress)

**Milestone Goal:** Fix Queue rendering - GraphViz cylinder rotation doesn't work, use HTML labels

### Phase 19: Queue Label Fix

**Status:** Planning
**Plans:** 0/1 plans complete

**Goal:** Revert Queue units to HTML labels with ASCII art graphic (═╦╩═╦═══)
**Requirements:** QUEUE-FIX-01, QUEUE-FIX-02, QUEUE-FIX-03, QUEUE-FIX-04
**Depends on:** Phase 18
**Plans:** 1 plan

**Queue Label Fix Design:**

- D-01: Queue units use HTML label with ASCII art graphic (═╦╩═╦═══)
- D-02: Queue external units use same HTML label format
- D-03: Queue label is 4-row table (graphic, name, technology, description)
- D-04: Remove cylinder shape and SetOrientation from Queue units in converter

Plans:
- [ ] 19-01-PLAN.md -- Update Queue label builder, remove cylinder shape for Queue

### Phase 20: Helvetica Font

**Status:** Complete
**Plans:** 1/1 plans complete

**Goal:** Use Helvetica font for all text rendering in diagrams
**Requirements:** FONT-01, FONT-02
**Depends on:** Phase 19
**Plans:** 1 plan

**Helvetica Font Design:**

- D-01: Set fontname="Helvetica" for all nodes and edges
- D-02: Set fontname="Helvetica" for all cluster (subgraph) labels
- D-03: Ensure consistent font family across all diagram elements

### Phase 21: Box Fixes - Labels, Borders, Validation, Color by Content

**Status:** Complete
**Plans:** 2/2 plans complete

**Goal:** Fix box unit rendering: remove curly brackets from labels, add dashed borders, validate C1 box contents, and color C1 boxes based on contents
**Requirements:** BOX-01, BOX-02, BOX-03, BOX-04
**Depends on:** Phase 20
**Plans:** 2 plans

**Box Fixes Design:**

- D-01: Box labels use HTML table format (no curly brackets) - same as container/component
- D-02: All box types (box, containerBox, componentBox) have dashed borders
- D-03: C1 boxes cannot contain both external and non-external units (validation rule)
- D-04: C1 box color based on contents: grey for externals, dark blue for non-externals

Plans:
- [x] 21-01-PLAN.md -- Box HTML labels and dashed borders
- [x] 21-02-PLAN.md -- C1 Box validation and color by content

---

*Roadmap created: 2021-03-09*
*Last updated: 2026-03-24 - Phase 21 complete (Box Fixes)*
