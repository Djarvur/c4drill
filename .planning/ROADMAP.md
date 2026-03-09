# Roadmap: C4Drill

## Overview

C4Drill transforms TOML architecture definitions into professional C4 diagrams through a compiler-style pipeline: parse, validate, generate views, build graphs, render to DOT/SVG. The journey starts with model types and parsing, adds validation to catch errors early, builds the view and graph layers, adds rendering output, enables interactive drill-down navigation, and finishes with CLI polish.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Foundation & Model** - Project setup, dev environment, domain types, TOML parsing
- [ ] **Phase 2: Validation** - Reference integrity, type constraints, clear error messages
- [ ] **Phase 3: Views & Graphs** - C1/C2/C3 view generation, graph construction with nodes/edges/clusters
- [ ] **Phase 4: Rendering & Output** - DOT generation, SVG rendering, file output structure
- [ ] **Phase 5: Navigation** - Collapsed/expanded views, explore links, drill-down file structure
- [ ] **Phase 6: CLI & Polish** - Cobra CLI, flags, help text, final integration

## Phase Details

### Phase 1: Foundation & Model

**Goal**: Development environment and core domain model ready for parsing TOML architecture definitions

**Depends on**: Nothing (first phase)

**Requirements**: DEVI-01, DEVI-02, DEVI-03, DEVI-04, DEVI-05, INPT-01, INPT-02, INPT-03, INPT-04, INPT-05, INPT-06, INPT-07, TYPE-01, TYPE-02, TYPE-03, TYPE-04, TYPE-05, TYPE-06, TYPE-07, TYPE-08, QUAL-01, QUAL-02, QUAL-03, QUAL-04, QUAL-05

**Success Criteria** (what must be TRUE):
  1. Developer can run `mise test` and `mise lint` to verify code quality
  2. Developer can parse a valid TOML file and get a populated model struct
  3. All unit types (person, system, db, queue, box, external variants) are defined as Go types
  4. Link objects with target, reverse, equal, color, style attributes are defined
  5. Parser handles nested unit definitions at arbitrary depth

**Plans**: 3 plans

Plans:
- [x] 01-PLAN.md - Development environment (mise, go.mod) and domain model types (Wave 1)
- [x] 02-PLAN.md - TOML parser implementation with go-toml v2 (Wave 2)
- [ ] 03-PLAN.md - CLI entry point and parser tests with 75% coverage (Wave 3, has checkpoint)

### Phase 2: Validation

**Goal**: Invalid TOML files produce clear, actionable error messages before any rendering

**Depends on**: Phase 1

**Requirements**: VALD-01, VALD-02, VALD-03, VALD-04, VALD-05, VALD-06, QUAL-01, QUAL-02, QUAL-03, QUAL-04, QUAL-05

**Success Criteria** (what must be TRUE):
  1. Validator rejects TOML files with references to undefined units
  2. Validator prevents links on units that have subunits
  3. Validator prevents referencing units that contain subunits
  4. Validator prevents subunits on non-system/non-box types
  5. Error messages include line numbers and human-readable context

**Plans**: 3 plans in 3 waves

Plans:
- [x] 02-01-PLAN.md - Validation infrastructure: error types, index builder, suggestion helper (Wave 1)
- [ ] 02-02-PLAN.md - Validation rules implementation: references, subunits, links (Wave 2)
- [ ] 02-03-PLAN.md - CLI integration and quality gates (Wave 3, has checkpoint)

### Phase 3: Views & Graphs

**Goal**: Validated model can be transformed into scoped views (C1/C2/C3) and graph structures

**Depends on**: Phase 2

**Requirements**: VIEW-01, VIEW-02, VIEW-03, VIEW-04, VIEW-05, VIEW-06, VIEW-07, GRPH-01, GRPH-02, GRPH-03, GRPH-04, GRPH-05, GRPH-06, QUAL-01, QUAL-02, QUAL-03, QUAL-04, QUAL-05

**Success Criteria** (what must be TRUE):
  1. Generator creates C1 (Context) view showing top-level units
  2. Generator creates C2 (Containers) view for expanded systems
  3. Generator creates C3 (Components) view for expanded containers
  4. Collapsed units render as single nodes; expanded units render as clusters
  5. Graph builder creates nodes with type-appropriate shapes (person icon, db cylinder, etc.)
  6. Graph builder creates edges with configured routing styles (straight, spline, square)

**Plans**: 3 plans in 3 waves

Plans:
- [x] 03-01-PLAN.md - View package: View/ViewUnit types, GenerateC1View/GenerateC2View/GenerateC3View (Wave 1)
- [x] 03-02-PLAN.md - Graph package: Graph/Node/Edge/Cluster types, BuildGraph, shape/style mapping (Wave 2)
- [x] 03-03-PLAN.md - Integration tests and quality verification (Wave 3)

### Phase 4: Rendering & Output

**Goal**: Graph structures render to valid DOT and SVG files with correct output paths

**Depends on**: Phase 3

**Requirements**: REND-01, REND-02, REND-03, OUTP-01, OUTP-02, OUTP-04, QUAL-01, QUAL-02, QUAL-03, QUAL-04, QUAL-05

**Success Criteria** (what must be TRUE):
  1. Renderer generates valid GraphViz DOT format from graph structures
  2. Renderer generates SVG output via go-graphviz library
  3. User can select output format via --format flag (dot|svg)
  4. Context level renders to {basename}.{format}
  5. Expanded units render to {basename}/{unit-name}.{format}
  6. Output directory structure is created recursively as needed

**Plans**: TBD

### Phase 5: Navigation

**Goal**: Users can navigate between diagram levels via explore links and back-links

**Depends on**: Phase 4

**Requirements**: REND-04, REND-05, REND-06, OUTP-05, QUAL-01, QUAL-02, QUAL-03, QUAL-04, QUAL-05

**Success Criteria** (what must be TRUE):
  1. Collapsed units include explore link pointing to drill-down file
  2. All diagrams include back-link to parent level
  3. All diagrams include breadcrumb trail showing navigation path
  4. Relative paths are used for explore and back links (portable across contexts)

**Plans**: TBD

### Phase 6: CLI & Polish

**Goal**: Users interact with a polished CLI that handles all input/output scenarios

**Depends on**: Phase 5

**Requirements**: CLII-01, CLII-02, CLII-03, CLII-04, CLII-05, CLII-06, OUTP-03, QUAL-01, QUAL-02, QUAL-03, QUAL-04, QUAL-05

**Success Criteria** (what must be TRUE):
  1. User can run `c4drill <input.toml>` to generate diagrams
  2. User can specify output directory with --output flag
  3. Help text shows usage examples and flag descriptions
  4. Tool exits with code 0 on success, non-zero on failure
  5. Errors are written to stderr (not stdout)

**Plans**: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5 → 6

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation & Model | 2/3 | Done | 2026-03-09 |
| 2. Validation | 2/3 | Done | 2026-03-09 |
| 3. Views & Graphs | 3/3 | Done | 2026-03-09 |
| 4. Rendering & Output | 0/TBD | Not started | - |
| 5. Navigation | 0/TBD | Not started | - |
| 6. CLI & Polish | 0/TBD | Not started | - |

---
*Roadmap created: 2026-03-09*
*Granularity: standard*
*Last updated: 2026-03-09 - Phase 3 complete (view + graph packages with 89%+ coverage)*
