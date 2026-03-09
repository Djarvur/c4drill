# Requirements: C4Drill

**Defined:** 2026-03-09
**Core Value:** Transform simple TOML architecture descriptions into professional C4 diagrams without manual drawing.

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Input/Parsing

- [ ] **INPT-01**: CLI accepts path to TOML input file
- [ ] **INPT-02**: Parser handles nested unit definitions with arbitrary depth
- [ ] **INPT-03**: Parser extracts properties section with name, description, styling
- [ ] **INPT-04**: Parser extracts context-level units (person, system, db, queue, box types)
- [ ] **INPT-05**: Parser handles external variants (personExternal, systemExternal, etc.)
- [ ] **INPT-06**: Parser extracts link and linkFrom definitions with styling attributes
- [ ] **INPT-07**: Parser handles expanded list for collapsed/expanded rendering control

### Model/Types

- [ ] **TYPE-01**: System defines person type with name, description, styling
- [ ] **TYPE-02**: System defines personExternal type (external actor)
- [ ] **TYPE-03**: System defines system type (can contain subunits for containers)
- [ ] **TYPE-04**: System defines systemExternal type (external system)
- [ ] **TYPE-05**: System defines db and dbExternal types (database storage)
- [ ] **TYPE-06**: System defines queue and queueExternal types (message queues)
- [ ] **TYPE-07**: System defines box type (grouping container, can contain context-level units)
- [ ] **TYPE-08**: Link object defines target, reverse, equal, color, style attributes

### Validation

- [ ] **VALD-01**: Validator checks all referenced units are defined
- [ ] **VALD-02**: Validator prevents links on units that have subunits
- [ ] **VALD-03**: Validator prevents referencing units that have subunits
- [ ] **VALD-04**: Validator prevents subunits on non-system/non-box types
- [ ] **VALD-05**: Error messages include line numbers and context
- [ ] **VALD-06**: Error messages use human-readable format (not JSON)

### View Generation

- [ ] **VIEW-01**: Generator creates C1 (Context) level view from model
- [ ] **VIEW-02**: Generator creates C2 (Containers) level view for expanded systems
- [ ] **VIEW-03**: Generator creates C3 (Components) level view for expanded containers
- [ ] **VIEW-04**: Collapsed units render as single record shape
- [ ] **VIEW-05**: Expanded units render as clusters with subunits inside
- [ ] **VIEW-06**: View respects expanded list from properties and unit-level overrides
- [ ] **VIEW-07**: Styling (color, border, style, edges) inherits from parent with override

### Graph Construction

- [ ] **GRPH-01**: Builder creates nodes for each unit with type-appropriate shapes
- [ ] **GRPH-02**: Builder creates edges for each link definition
- [ ] **GRPH-03**: Builder applies edge routing style (straight, spline, square)
- [ ] **GRPH-04**: Builder creates clusters for expanded units
- [ ] **GRPH-05**: Shapes: person uses 👤 icon, db uses ⛁ icon, queue uses ╟╢ bars
- [ ] **GRPH-06**: System shape includes name, description, explore link

### Rendering

- [ ] **REND-01**: Renderer generates valid GraphViz DOT format
- [ ] **REND-02**: Renderer generates SVG via go-graphviz library
- [ ] **REND-03**: Output format controlled by --format flag (dot|svg)
- [ ] **REND-04**: Collapsed units include explore link pointing to drill-down file
- [ ] **REND-05**: All diagrams include back-link to parent level
- [ ] **REND-06**: All diagrams include breadcrumb trail showing path

### Output

- [ ] **OUTP-01**: Context level renders to {basename}.{format}
- [ ] **OUTP-02**: Expanded units render to {basename}/{unit-name}.{format}
- [ ] **OUTP-03**: Output directory controlled by --output flag (default: current directory)
- [ ] **OUTP-04**: Directory structure created recursively as needed
- [ ] **OUTP-05**: Relative paths used for explore and back links

### CLI

- [ ] **CLII-01**: Single command: c4drill <input.toml> [flags]
- [ ] **CLII-02**: --format flag selects output format (dot|svg)
- [ ] **CLII-03**: --output flag specifies output directory
- [ ] **CLII-04**: Help text with usage examples
- [ ] **CLII-05**: Exit code 0 on success, non-zero on failure
- [ ] **CLII-06**: Errors written to stderr

### Development Environment

- [ ] **DEVI-01**: Go version updated to 1.26.1 in go.mod and all config files before development
- [ ] **DEVI-02**: Mise config includes tasks for running tests
- [ ] **DEVI-03**: Mise config includes tasks for running lint
- [ ] **DEVI-04**: Mise installs golangci-lint v2 into sandbox (not global)
- [ ] **DEVI-05**: Modern Go plugin loaded via /use-modern-go before any development task

### Quality Gates

- [ ] **QUAL-01**: All lint errors must be fixed before commit
- [ ] **QUAL-02**: Lint config (.golangci.yml) MUST NOT be adjusted to silence errors
- [ ] **QUAL-03**: nolint directives require explicit user confirmation before adding
- [ ] **QUAL-04**: Minimum 75% test coverage required
- [ ] **QUAL-05**: Coverage enforced in CI/quality gate

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Advanced Features

- **TAGS-01**: Tags/stereotypes metadata on units
- **THME-01**: Predefined color themes/schemes
- **THME-02**: Custom theme loading from file
- **WTCH-01**: Watch mode for auto-regeneration on file change
- **WTCH-02**: Live reload integration for browser preview

### Extended Output

- **OUTP-06**: PNG output format
- **OUTP-07**: PDF output format

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| C4 Layer 4 (Code) | Class/function level not needed for architecture docs |
| Go library/module | Pure CLI tool per user specification |
| Manual positioning | Rely on GraphViz auto-layout |
| JSON error output | CLI errors sufficient for v1 |
| Multiple subcommands | Single command workflow per user specification |
| Real-time collaboration | Single-user tool |
| Diagram editing UI | Text-based input only |
| Cloud hosting | Local file generation only |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| INPT-01 | Phase 1 | Pending |
| INPT-02 | Phase 1 | Pending |
| INPT-03 | Phase 1 | Pending |
| INPT-04 | Phase 1 | Pending |
| INPT-05 | Phase 1 | Pending |
| INPT-06 | Phase 1 | Pending |
| INPT-07 | Phase 1 | Pending |
| TYPE-01 | Phase 1 | Pending |
| TYPE-02 | Phase 1 | Pending |
| TYPE-03 | Phase 1 | Pending |
| TYPE-04 | Phase 1 | Pending |
| TYPE-05 | Phase 1 | Pending |
| TYPE-06 | Phase 1 | Pending |
| TYPE-07 | Phase 1 | Pending |
| TYPE-08 | Phase 1 | Pending |
| VALD-01 | Phase 2 | Pending |
| VALD-02 | Phase 2 | Pending |
| VALD-03 | Phase 2 | Pending |
| VALD-04 | Phase 2 | Pending |
| VALD-05 | Phase 2 | Pending |
| VALD-06 | Phase 2 | Pending |
| VIEW-01 | Phase 3 | Pending |
| VIEW-02 | Phase 3 | Pending |
| VIEW-03 | Phase 3 | Pending |
| VIEW-04 | Phase 3 | Pending |
| VIEW-05 | Phase 3 | Pending |
| VIEW-06 | Phase 3 | Pending |
| VIEW-07 | Phase 3 | Pending |
| GRPH-01 | Phase 3 | Pending |
| GRPH-02 | Phase 3 | Pending |
| GRPH-03 | Phase 3 | Pending |
| GRPH-04 | Phase 3 | Pending |
| GRPH-05 | Phase 3 | Pending |
| GRPH-06 | Phase 3 | Pending |
| REND-01 | Phase 4 | Pending |
| REND-02 | Phase 4 | Pending |
| REND-03 | Phase 4 | Pending |
| REND-04 | Phase 5 | Pending |
| REND-05 | Phase 5 | Pending |
| REND-06 | Phase 5 | Pending |
| OUTP-01 | Phase 4 | Pending |
| OUTP-02 | Phase 4 | Pending |
| OUTP-03 | Phase 6 | Pending |
| OUTP-04 | Phase 4 | Pending |
| OUTP-05 | Phase 5 | Pending |
| CLII-01 | Phase 6 | Pending |
| CLII-02 | Phase 6 | Pending |
| CLII-03 | Phase 6 | Pending |
| CLII-04 | Phase 6 | Pending |
| CLII-05 | Phase 6 | Pending |
| CLII-06 | Phase 6 | Pending |
| MISE-01 | Phase 1 | Pending |
| MISE-02 | Phase 1 | Pending |
| DEVI-01 | Phase 1 | Pending |
| DEVI-05 | All Phases | Pending |
| QUAL-01 | All Phases | Pending |
| QUAL-02 | All Phases | Pending |
| QUAL-03 | All Phases | Pending |
| QUAL-04 | All Phases | Pending |
| QUAL-05 | All Phases | Pending |

**Coverage:**
- v1 requirements: 60 total
- Mapped to phases: 60
- Unmapped: 0 ✓

---
*Requirements defined: 2026-03-09*
*Last updated: 2026-03-09 after initial definition*
