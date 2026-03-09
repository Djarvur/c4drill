# Project Research Summary

**Project:** C4Drill
**Domain:** Go CLI for C4 Architecture Diagram Generation
**Researched:** 2026-03-09
**Confidence:** HIGH

## Executive Summary

C4Drill is a single-purpose Go CLI tool that transforms TOML architecture definitions into C4 model diagrams (C1-C3 levels) using GraphViz auto-layout. The recommended approach follows a classic compiler-style pipeline: parse TOML, validate semantic integrity, generate filtered views by scope, construct graph structures, and render to DOT/SVG formats using the pure-Go goccy/go-graphviz library.

The key differentiator for C4Drill is interactive drill-down navigation via collapsed/expanded views and explore links between diagram levels. Critical risks include GraphViz layout non-determinism (outputs may vary across runs), go-graphviz memory limitations for large diagrams (WASM constraints), and the complexity of collapsed/expanded state inheritance. All three must be addressed early to avoid rework.

## Key Findings

### Recommended Stack

Go 1.24+ with Cobra for CLI structure, BurntSushi/toml for configuration parsing, and goccy/go-graphviz for pure-Go DOT/SVG rendering without external binary dependencies. This stack prioritizes portable single-binary distribution and aligns with Go ecosystem best practices.

**Core technologies:**
- **Go 1.24+** — Runtime with Swiss Tables optimization, required for modern Go patterns
- **Cobra v1.9+** — CLI framework used by Kubernetes/Docker, provides help generation and shell completions
- **BurntSushi/toml v1.6.0** — Most widely adopted TOML library (15k+ stars), supports TOML v1.1.0
- **goccy/go-graphviz v0.2.9** — Pure Go GraphViz via WASM, no external binary dependency, critical for portability
- **stretchr/testify v1.10+** — Testing assertions and mocks for comprehensive test coverage

### Expected Features

All v1 features center on the core workflow: TOML input, C4 model validation, multi-level view generation, and DOT/SVG output. Collapsed/expanded views and explore links are the primary differentiators versus competitors.

**Must have (table stakes):**
- TOML parsing with nested unit definitions — core input mechanism
- C1-C3 layer generation — fundamental C4 model scope
- All unit types (person, system, db, queue, external variants, box) — complete notation
- Link/relationship definitions — diagrams need connections
- Validation with clear error messages — prevent silent failures
- DOT and SVG output — intermediate and final formats

**Should have (differentiators):**
- Collapsed/expanded views — drill-down navigation without separate files
- Explore links — click-through navigation between diagram levels
- Basic styling with inheritance — visual distinction via properties
- Single CLI command — simple workflow, no subcommand complexity

**Defer (v2+):**
- Tags/stereotypes — adds metadata complexity
- Themes/color schemes — branding is secondary to core functionality
- Sprites/icons — requires external assets, breaks single-file goal
- Watch mode — file watching adds edge case complexity

### Architecture Approach

Pipeline architecture with clear stage separation: parse → validate → view → graph → render → output. The model is the single source of truth; views are projections filtered by scope (C1/C2/C3) and expansion state. This pattern is well-established in tools like Structurizr and enables clean testing at each stage.

**Major components:**
1. **CLI Interface (cmd/c4drill/)** — Cobra-based command parsing, pipeline orchestration
2. **Model (internal/model/)** — Domain types: Workspace, Unit, Link, Properties — no external dependencies
3. **Parser (internal/parser/)** — TOML parsing with BurntSushi/toml, custom unmarshaling for nested structures
4. **Validator (internal/validate/)** — Reference integrity, type constraints, subunit rules
5. **View Generator (internal/view/)** — Scope filtering (C1/C2/C3), expansion handling
6. **Graph Builder (internal/graph/)** — Node/edge/cluster construction from views
7. **Renderer (internal/render/)** — DOT generation and SVG rendering via go-graphviz
8. **Output (internal/output/)** — File writing, path resolution for drill-down file structures

### Critical Pitfalls

1. **GraphViz Layout Non-Determinism** — Same input produces different layouts across runs. Prevention: document variability, generate DOT as primary artifact (deterministic), SVG as secondary. Test by running generation 3x on same input.

2. **TOML Reference Integrity Violations** — Links to non-existent units, circular references, or referencing container units directly. Prevention: two-pass parsing (syntax then semantic), build reference graph, detect cycles, validate leaf-node references only. Comprehensive validation test suite required.

3. **go-graphviz Memory/Segfault Issues** — WASM-embedded GraphViz has constrained memory. Prevention: document limits, graceful error handling, stress testing to find breaking points, provide DOT-only fallback.

4. **Collapsed/Expanded State Inconsistency** — Inheritance rules between global default and per-unit overrides create complex interactions. Prevention: explicit inheritance resolution before rendering, test matrix for all combinations, validate expanded units have subunits.

5. **SVG Interactive Link Generation Failures** — Explore links break when diagrams moved or hosted differently. Prevention: consistent relative paths, test in multiple contexts (file://, localhost, hosted), document expected hosting setup.

## Implications for Roadmap

Based on research, suggested phase structure:

### Phase 1: Core Model and Parser
**Rationale:** Model types have zero dependencies and enable all downstream work. Parser validates the input layer early.
**Delivers:** Domain types, TOML parsing, basic validation
**Addresses:** Unit types, TOML parsing, basic validation from FEATURES.md
**Avoids:** Pitfall 2 (reference integrity) — validation starts here

### Phase 2: Validator and Core Pipeline
**Rationale:** Complete validation before any rendering. Establish error handling patterns early.
**Delivers:** Full reference integrity validation, type constraints, clear error messages with line numbers
**Uses:** BurntSushi/toml, go-playground/validator
**Implements:** Validator component from ARCHITECTURE.md
**Avoids:** Pitfalls 1, 2, 3 — deterministic baseline, validation foundation, error handling

### Phase 3: View Generator and Graph Builder
**Rationale:** Scope filtering and graph construction are intermediate stages between model and output.
**Delivers:** C1/C2/C3 view generation, graph representation with nodes/edges/clusters
**Uses:** Internal model and view types
**Implements:** View Generator and Graph Builder components

### Phase 4: DOT and SVG Rendering
**Rationale:** Output generation completes the core pipeline. Start with DOT, add SVG via go-graphviz.
**Delivers:** DOT output, SVG rendering, file writing
**Uses:** goccy/go-graphviz
**Implements:** Renderer and Output components
**Avoids:** Pitfall 3 (memory issues) — test limits here

### Phase 5: Collapsed/Expanded Views and Explore Links
**Rationale:** This is the primary differentiator. Requires all core pipeline stages working first.
**Delivers:** Interactive drill-down navigation, multi-file output structure
**Uses:** View expansion logic, relative path handling
**Avoids:** Pitfalls 4, 5 (state inconsistency, broken links) — focused testing here

### Phase 6: CLI Polish and Styling
**Rationale:** Final UX improvements after core functionality proven.
**Delivers:** Cobra CLI with flags, help text, basic styling/inheritance, error message formatting
**Uses:** Cobra, fatih/color
**Implements:** CLI Interface component

### Phase Ordering Rationale

- **Model first:** Domain types have no dependencies; everything else builds on them
- **Validation before rendering:** Invalid models produce confusing errors downstream; fail fast
- **DOT before SVG:** DOT is deterministic and debuggable; SVG adds WASM complexity
- **Core views before expansion:** C1-C3 basic generation must work before collapsed/expanded complexity
- **Explore links last:** Requires complete file structure and path handling

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 4:** goccy/go-graphviz API specifics for SVG rendering, memory limits characterization
- **Phase 5:** Relative path handling across platforms, SVG link syntax specifics

Phases with standard patterns (skip research-phase):
- **Phase 1:** Go struct definitions are straightforward
- **Phase 2:** Validation patterns well-documented in Go ecosystem
- **Phase 3:** View filtering is domain logic, not library integration
- **Phase 6:** Cobra CLI patterns are well-established

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All libraries are mature, actively maintained, with official documentation verified |
| Features | MEDIUM | Based on competitor analysis and C4 model principles, but user validation needed |
| Architecture | HIGH | Pipeline pattern is standard for compiler-style tools; Structurizr is proven reference |
| Pitfalls | MEDIUM | Based on go-graphviz issues and general parsing experience; empirical testing needed for limits |

**Overall confidence:** HIGH

### Gaps to Address

- **go-graphviz determinism:** Research explicitly whether deterministic layout is possible; document findings before Phase 4
- **Performance limits:** Empirical testing needed to establish max nodes/edges/nesting depth before memory issues
- **TOML schema finalization:** Exact structure of nested units and link objects needs validation during Phase 1 implementation

## Sources

### Primary (HIGH confidence)
- pkg.go.dev/github.com/goccy/go-graphviz — API reference, v0.2.9
- pkg.go.dev/github.com/BurntSushi/toml — TOML v1.1.0 support, v1.6.0
- pkg.go.dev/github.com/spf13/cobra — CLI framework, v1.9.x
- structurizr.com/help/model — C4 model building blocks
- structurizr.com/help/views — View types and scoping
- c4model.com — C4 model specification

### Secondary (MEDIUM confidence)
- github.com/goccy/go-graphviz/issues — Common rendering problems, memory issues
- github.com/plantuml-stdlib/C4-PlantUML — Feature comparison reference
- docs.likec4.dev — Feature comparison reference
- mermaid.js.org/syntax/c4.html — Feature comparison reference

### Tertiary (LOW confidence)
- Performance limit estimates — Need empirical validation during implementation

---
*Research completed: 2026-03-09*
*Ready for roadmap: yes*
