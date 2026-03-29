# Milestones

## v1.7 Queue Label Fix & Visual Improvements (Shipped: 2026-03-29)

**Phases completed:** 8 phases (19–26, excluding skipped 24), 7 plans

**Stats:** 15,163 LOC Go, all tests passing

**Key accomplishments:**

- Queue units render with ASCII art graphic (═╦╩═╦═══) instead of broken rotated cylinder
- Helvetica font for all diagram elements
- Box labels with dashed borders, validation for mixed content, color by content type
- Link length attribute for edge spacing control
- Deterministic node/edge ordering
- Thicker edges (penwidth 2.0) for visual prominence
- TOML definition order preservation for nodes and edges

---

## v1.0 Initial Release (Shipped: 2026-03-10)

**Phases completed:** 6 phases, 16 plans

**Stats:** 9,624 LOC Go, 48 files, 28 feature commits

**Key accomplishments:**

- TOML parser with nested unit definitions and error handling
- C4 model validation with clear error messages and line numbers
- C1/C2/C3 view generation with collapsed/expanded rendering
- GraphViz DOT and SVG rendering via go-graphviz
- Interactive navigation with explore links, back-links, and breadcrumbs
- Production-ready Cobra CLI with help text and error handling

---

