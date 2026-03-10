# Requirements: C4Drill

**Defined:** 2026-03-10
**Core Value:** Transform simple TOML architecture descriptions into professional C4 diagrams without manual drawing

## v1.1 Requirements

Requirements for AI-Ready milestone. Each maps to roadmap phases.

### AI Documentation (CLAUDE.md)

- [ ] **AIDOC-01**: CLAUDE.md contains complete TOML schema reference with all unit types, fields, and link syntax
- [ ] **AIDOC-02**: CLAUDE.md includes 3-5 working examples (minimal, medium, complex architectures)
- [ ] **AIDOC-03**: CLAUDE.md documents all validation rules with clear error explanations
- [ ] **AIDOC-04**: CLAUDE.md provides prompt patterns for AI assistants to generate valid TOML
- [ ] **AIDOC-05**: All TOML examples in CLAUDE.md are validated by CI to prevent drift

### All-Expanded Mode

- [ ] **EXPD-01**: User can pass `--expanded` flag to CLI to request all-expanded rendering
- [ ] **EXPD-02**: All-expanded view renders all units as expanded nested clusters in single diagram
- [ ] **EXPD-03**: Cross-level edges (between units at different nesting depths) are visible
- [ ] **EXPD-04**: Output saved to `{basename}.expanded.{ext}` format (dot/svg)
- [ ] **EXPD-05**: Existing C1/C2/C3 view generation remains unchanged (zero regression)

### Validation Enhancement

- [ ] **VAL-01**: Validator rejects TOML files with unlinked (orphan) units - all units must have at least one incoming or outgoing link
- [ ] **VAL-02**: Validation error message clearly identifies which units are unlinked

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
| AIDOC-01 | Phase 7 | Pending |
| AIDOC-02 | Phase 7 | Pending |
| AIDOC-03 | Phase 7 | Pending |
| AIDOC-04 | Phase 7 | Pending |
| AIDOC-05 | Phase 7 | Pending |
| EXPD-01 | Phase 8 | Pending |
| EXPD-02 | Phase 8 | Pending |
| EXPD-03 | Phase 8 | Pending |
| EXPD-04 | Phase 8 | Pending |
| EXPD-05 | Phase 8 | Pending |
| VAL-01 | Phase 9 | Pending |
| VAL-02 | Phase 9 | Pending |

**Coverage:**
- v1.1 requirements: 12 total
- Mapped to phases: 12
- Unmapped: 0 ✓

---
*Requirements defined: 2026-03-10*
*Last updated: 2026-03-10 after v1.1 milestone started*
