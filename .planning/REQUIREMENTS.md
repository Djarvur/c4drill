# Requirements: C4Drill

**Defined:** 2026-03-10
**Core Value:** Transform simple TOML architecture descriptions into professional C4 diagrams without manual drawing

## v1.1 Requirements

Requirements for AI-Ready milestone. Each maps to roadmap phases.

### AI Documentation (CLAUDE.md)

- [x] **AIDOC-01**: CLAUDE.md contains complete TOML schema reference with all unit types, fields, and link syntax
- [x] **AIDOC-02**: CLAUDE.md includes 3-5 working examples (minimal, medium, complex architectures)
- [x] **AIDOC-03**: CLAUDE.md documents all validation rules with clear error explanations
- [x] **AIDOC-04**: CLAUDE.md provides prompt patterns for AI assistants to generate valid TOML
- [x] **AIDOC-05**: All TOML examples in CLAUDE.md are validated by CI to prevent drift

### All-Expanded Mode

- [x] **EXPD-01**: User can pass `--expanded` flag to CLI to request all-expanded rendering
- [x] **EXPD-02**: All-expanded view renders all units as expanded nested clusters in single diagram
- [x] **EXPD-03**: Cross-level edges (between units at different nesting depths) are visible
- [x] **EXPD-04**: Output saved to `{basename}.expanded.{ext}` format (dot/svg)
- [x] **EXPD-05**: Existing C1/C2/C3 view generation remains unchanged (zero regression)

### Validation Enhancement

- [x] **VAL-01**: Validator rejects TOML files with unlinked (orphan) units - all units must have at least one incoming or outgoing link
- [x] **VAL-02**: Validation error message clearly identifies which units are unlinked

### Parent Links

- [x] **PLNK-01**: A unit with subunits can have Links field without validation error
- [x] **PLNK-02**: A unit with subunits can have LinksFrom field without validation error

### Link List Format

- [x] **LLIST-01**: Link struct has `Peer` field instead of `Target`, with `toml:"peer"` tag
- [x] **LLIST-02**: Unit.Links and Unit.LinksFrom are `[]Link` slices instead of `map[string]Link`
- [x] **LLIST-03**: Parser, validator, view, and graph code updated to iterate slices using `link.Peer`
- [x] **LLIST-04**: All documentation and examples updated to use `[[link]]`/`[[linkFrom]]` array syntax with `peer` field

### Unit Shape and Attributes (v1.2)

- [x] **SHAPE-01**: Collapsed units render with record shape (not HTML labels)
- [x] **SHAPE-02**: All units have transparent backgrounds (no fill colors)

### HTML Labels for All Unit Types

- [x] **HTML-01**: All unit types render with HTML table labels inside record shapes
- [x] **HTML-02**: Each unit type has specific format: Person (icon+name+desc), Database (icon+name+tech+desc), Queue (graphics+name+tech+desc), System/Container/Component (name+tech+desc)

## v1.0 Requirements (Shipped)

Completed in v1.0 Initial Release (2026-03-10). See `.planning/milestones/v1.0-REQUIREMENTS.md` for archive.

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| C4 Layer 4 (Code) | Class/function level diagrams not needed |
| Go library/module | Pure CLI tool, no library interface |
| Manual positioning | Rely on GraphViz auto-layout |
| JSON error output | CLI errors sufficient |
| Live editing/watch mode | Single-shot rendering |
| Multiple commands | Single command does everything |
| Interactive legend in expanded view | Nice-to-have, defer to v1.2 |
| Partial expansion (`--expand=unit1,unit2`) | Adds complexity, defer to v1.2 |
| AI validation workflow integration | Outside CLI scope |
| Edge filtering/aggregation | Future enhancement for cluttered diagrams |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| AIDOC-01 | Phase 7 | Complete |
| AIDOC-02 | Phase 7 | Complete |
| AIDOC-03 | Phase 7 | Complete |
| AIDOC-04 | Phase 7 | Complete |
| AIDOC-05 | Phase 7 | Complete |
| EXPD-01 | Phase 8 | Complete |
| EXPD-02 | Phase 8 | Complete |
| EXPD-03 | Phase 8 | Complete |
| EXPD-04 | Phase 8 | Complete |
| EXPD-05 | Phase 8 | Complete |
| VAL-01 | Phase 9 | Complete |
| VAL-02 | Phase 9 | Complete |
| PLNK-01 | Phase 9 | Complete |
| PLNK-02 | Phase 9 | Complete |
| LLIST-01 | Phase 10 | Complete |
| LLIST-02 | Phase 10 | Complete |
| LLIST-03 | Phase 10 | Complete |
| LLIST-04 | Phase 10 | Complete |
| SHAPE-01 | Phase 11 | Complete |
| SHAPE-02 | Phase 11 | Complete |
| HTML-01 | Phase 12 | Complete |
| HTML-02 | Phase 12 | Complete |

**Coverage:**
- v1.1 requirements: 18 total
- v1.2 requirements: 4 total
- Mapped to phases: 22
- Unmapped: 0 ✓

---
*Requirements defined: 2026-03-10*
*Last updated: 2026-03-13 after Phase 12 plan 01 - v1.2 complete*
