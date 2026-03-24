# Requirements: C4Drill

**Defined:** 2026-03-10
**Core Value:** Transform simple TOML architecture descriptions into professional C4 diagrams without manual drawing

## v1.6 Requirements

Requirements for Simplified Shapes milestone. Replace icon system with native GraphViz shapes.

### Icon Removal

- [ ] **ICON-01**: Remove icons package (internal/icons) entirely
- [ ] **ICON-02**: Remove IconExtractor from converter
- [ ] **ICON-03**: Remove .icons/ directory generation from output
- [ ] **ICON-04**: Remove SVG postprocessing logic

### DB Shape

- [ ] **DB-01**: DB units render with GraphViz native `shape=cylinder`
- [ ] **DB-02**: DB external units also use cylinder shape
- [ ] **DB-03**: DB label is simple 3-row table (name, technology, description)

### Queue Shape

- [ ] **QUEUE-01**: Queue units render with GraphViz native `shape=cylinder` rotated 90°
- [ ] **QUEUE-02**: Queue external units also use rotated cylinder shape
- [ ] **QUEUE-03**: Queue label is simple 3-row table (name, technology, description)

### Person Label

- [ ] **PERSON-01**: Person units use 2-column table layout
- [ ] **PERSON-02**: First column contains 👤 emoji at font size 8
- [ ] **PERSON-03**: Second column contains name and description rows
- [ ] **PERSON-04**: Person external units use same label format

### System/Box Label

- [ ] **LABEL-01**: System units use simple 3-row table (name, technology, description)
- [ ] **LABEL-02**: Box units use same 3-row table format
- [ ] **LABEL-03**: Container and Component units use same 3-row table format
- [ ] **LABEL-04**: No icon column in system/box/container/component labels

### Word Wrap

- [ ] **WRAP-01**: All label text is word-wrapped to maintain credit card proportions
- [ ] **WRAP-02**: Existing --label-ratio flag continues to work

## v1.1 Requirements (Shipped)

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

## v1.3 Requirements (Shipped)

Requirements for Validation Enhancements milestone.

### Nesting Validation

- [x] **NEST-01**: Top-level units must be C1 types (person, system, db, queue, box + external variants); C2/C3 types at top level are rejected with clear error
- [x] **NEST-02**: Units inside system/systemExternal/box must be C2 types (container, containerDb, containerQueue); C3 types inside system/box are rejected
- [x] **NEST-03**: Units inside container must be C3 types (component, componentDb, componentQueue); C2 types inside container are rejected

## v1.0 Requirements (Shipped)

Completed in v1.0 Initial Release (2026-03-10). See `.planning/milestones/v1.0-REQUIREMENTS.md` for archive.

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature                     | Reason                                              |
|-----------------------------|-----------------------------------------------------|
| C4 Layer 4 (Code)           | Class/function level diagrams not needed            |
| Go library/module           | Pure CLI tool, no library interface                 |
| Manual positioning          | Rely on GraphViz auto-layout                        |
| JSON error output           | CLI errors sufficient                               |
| Live editing/watch mode     | Single-shot rendering                               |
| Multiple commands           | Single command does everything                      |
| Interactive legend          | Nice-to-have, defer                                 |
| Partial expansion           | Adds complexity, defer                              |
| AI validation workflow      | Outside CLI scope                                   |
| Edge filtering/aggregation  | Future enhancement for cluttered diagrams           |
| SVG icons                   | Replaced with native shapes and emoji (v1.6)        |
| Custom HTML shapes          | Using GraphViz native shapes instead (v1.6)         |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status  |
|-------------|-------|---------|
| ICON-01     | 18    | Pending |
| ICON-02     | 18    | Pending |
| ICON-03     | 18    | Pending |
| ICON-04     | 18    | Pending |
| DB-01       | 18    | Pending |
| DB-02       | 18    | Pending |
| DB-03       | 18    | Pending |
| QUEUE-01    | 18    | Pending |
| QUEUE-02    | 18    | Pending |
| QUEUE-03    | 18    | Pending |
| PERSON-01   | 18    | Pending |
| PERSON-02   | 18    | Pending |
| PERSON-03   | 18    | Pending |
| PERSON-04   | 18    | Pending |
| LABEL-01    | 18    | Pending |
| LABEL-02    | 18    | Pending |
| LABEL-03    | 18    | Pending |
| LABEL-04    | 18    | Pending |
| WRAP-01     | 18    | Pending |
| WRAP-02     | 18    | Pending |
| AIDOC-01    | 7     | Complete |
| AIDOC-02    | 7     | Complete |
| AIDOC-03    | 7     | Complete |
| AIDOC-04    | 7     | Complete |
| AIDOC-05    | 7     | Complete |
| EXPD-01     | 8     | Complete |
| EXPD-02     | 8     | Complete |
| EXPD-03     | 8     | Complete |
| EXPD-04     | 8     | Complete |
| EXPD-05     | 8     | Complete |
| VAL-01      | 9     | Complete |
| VAL-02      | 9     | Complete |
| PLNK-01     | 9     | Complete |
| PLNK-02     | 9     | Complete |
| LLIST-01    | 10    | Complete |
| LLIST-02    | 10    | Complete |
| LLIST-03    | 10    | Complete |
| LLIST-04    | 10    | Complete |
| SHAPE-01    | 11    | Complete |
| SHAPE-02    | 11    | Complete |
| HTML-01     | 12    | Complete |
| HTML-02     | 12    | Complete |
| NEST-01     | 14    | Complete |
| NEST-02     | 14    | Complete |
| NEST-03     | 14    | Complete |

**Coverage:**
- v1.6 requirements: 20 total
- v1.1 requirements: 18 total (shipped)
- v1.3 requirements: 3 total (shipped)
- Mapped to phases: 41
- Unmapped: 0 ✓

---
*Requirements defined: 2026-03-10*
*Last updated: 2026-03-24 after v1.6 milestone started*
