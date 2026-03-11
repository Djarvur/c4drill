# Roadmap: C4Drill

## Overview

C4Drill transforms TOML architecture definitions into professional C4 diagrams through a compiler-style pipeline: parse, validate, generate views, build graphs, render to DOT/SVG.

## Milestones

- ✅ **v1.0 Initial Release** — Phases 1-6 (shipped 2026-03-10)
- 🚧 **v1.1 AI-Ready** — Phases 7-9 (in progress)
- 📋 **v2.0** — Future features (planned)

## Phases

<details>
<summary>✅ v1.0 Initial Release (Phases 1-6) — SHIPPED 2026-03-10</summary>

- [x] Phase 1: Foundation & Model — completed 2026-03-09
- [x] Phase 2: Validation — completed 2026-03-09
- [x] Phase 3: Views & Graphs — completed 2026-03-09
- [x] Phase 4: Rendering & Output — completed 2026-03-10
- [x] Phase 5: Navigation — completed 2026-03-10
- [x] Phase 6: CLI & Polish — completed 2026-03-10

See `.planning/milestones/v1.0-ROADMAP.md` for full phase details.

</details>

### 🚧 v1.1 AI-Ready (In Progress)

**Milestone Goal:** Make C4Drill AI-friendly with documentation, single-view expanded diagrams, and stricter validation

- [ ] **Phase 7: AI Documentation** — TOML language manual for AI assistants
- [x] **Phase 8: All-Expanded Mode** — Single-view expanded rendering with cross-level edges (completed 2026-03-10)
- [x] **Phase 9: No Orphan Units** — Validation rule requiring all units to be linked (completed 2026-03-11)

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
**Plans:** 2 plans

Plans:
- [ ] 07-01-PLAN.md — Create skill package (SKILL.md + examples)
- [ ] 07-02-PLAN.md — Add CI validation for examples

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
- [ ] 08-01-PLAN.md — Add core data structures and view generation
- [ ] 08-02-PLAN.md — Add CLI flag integration and recursive cluster building

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
- [x] 09-01-PLAN.md — Add ValidateOrphanUnits rule with tests and integration

### Phase 10: Allow Parent Links
**Goal**: Units with subunits can be linked to directly (remove validation restriction)
**Depends on**: Phase 9 (builds on orphan detection logic)
**Requirements**: PLNK-01, PLNK-02
**Success Criteria** (what must be TRUE):
  1. Users can link to units that have subunits without validation errors
  2. Units with subunits can have their own Links and LinksFrom
  3. Existing valid TOML files continue to validate successfully
  4. Orphan detection still works correctly (units with Links/LinksFrom are not orphans)
**Plans:** 1/1 plans complete

Plans:
- [ ] 10-01-PLAN.md — Remove link restrictions and update tests

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
| 7. AI Documentation | v1.1 | 0/2 | Ready to execute | - |
| 8. All-Expanded Mode | v1.1 | Complete    | 2026-03-10 | - |
| 9. No Orphan Units | v1.1 | Complete    | 2026-03-11 | - |
| 10. Allow Parent Links | 1/1 | Complete    | 2026-03-11 | - |

---

*Roadmap created: 2026-03-09*
*Last updated: 2026-03-11 - Phase 10 added (allow parent links)*
