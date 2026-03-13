# Roadmap: C4Drill

## Overview

C4Drill transforms TOML architecture definitions into professional C4 diagrams through a compiler-style pipeline: parse, validate, generate views, build graphs, render to DOT/SVG.

## Milestones

- **v1.0 Initial Release** -- Phases 1-6 (shipped 2026-03-10)
- **v1.1 AI-Ready** -- Phases 7-11 (in progress)
- **v2.0** -- Future features (planned)

## Phases

<details>
<summary>v1.0 Initial Release (Phases 1-6) -- SHIPPED 2026-03-10</summary>

- [x] Phase 1: Foundation & Model -- completed 2026-03-09
- [x] Phase 2: Validation -- completed 2026-03-09
- [x] Phase 3: Views & Graphs -- completed 2026-03-09
- [x] Phase 4: Rendering & Output -- completed 2026-03-10
- [x] Phase 5: Navigation -- completed 2026-03-10
- [x] Phase 6: CLI & Polish -- completed 2026-03-10

See `.planning/milestones/v1.0-ROADMAP.md` for full phase details.

</details>

### v1.1 AI-Ready (In Progress)

**Milestone Goal:** Make C4Drill AI-friendly with documentation, single-view expanded diagrams, and stricter validation

- [x] **Phase 7: AI Documentation** -- TOML language manual for AI assistants
- [x] **Phase 8: All-Expanded Mode** -- Single-view expanded rendering with cross-level edges (completed 2026-03-10)
- [x] **Phase 9: No Orphan Units** -- Validation rule requiring all units to be linked (completed 2026-03-11)
- [x] **Phase 10: Link List Format** -- Change Links/LinksFrom from maps to lists with explicit target/source (completed 2026-03-13)

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
| 1. Foundation & Model | v1.0 | 3/3 | Complete | 2026-03-09 |
| 2. Validation | v1.0 | 2/2 | Complete | 2026-03-09 |
| 3. Views & Graphs | v1.0 | 4/4 | Complete | 2026-03-09 |
| 4. Rendering & Output | v1.0 | 3/3 | Complete | 2026-03-10 |
| 5. Navigation | v1.0 | 2/2 | Complete | 2026-03-10 |
| 6. CLI & Polish | v1.0 | 2/2 | Complete | 2026-03-10 |
| 7. AI Documentation | v1.1 | 2/2 | Complete | 2026-03-10 |
| 8. All-Expanded Mode | v1.1 | 2/2 | Complete | 2026-03-10 |
| 9. No Orphan Units | v1.1 | 1/1 | Complete | 2026-03-11 |
| 10. Link List Format | v1.1 | 3/3 | Complete | 2026-03-13 |
| 11. Unit Shape and Attributes | v1.1 | 1/1 | Complete | 2026-03-13 |
| 12. HTML labels for all unit types | v1.2 | 2/2 | Complete | 2026-03-13 |
| 13. Refined HTML Labels | v1.2 | 0/1 | Not Started | - |

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

**Goal:** Refine HTML labels with shape=box style=rounded and updated table attributes (border="0" cellpadding="0" cellspacing="0")
**Requirements**: REFINED-01, REFINED-02
**Depends on:** Phase 12
**Plans:** 0 plans
**Status:** Not Started

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
- [ ] TBD (run /gsd:discuss-phase 13 to gather context)

---

*Roadmap created: 2026-03-09*
*Last updated: 2026-03-13 - Phase 13 added for refined HTML labels*
