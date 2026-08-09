# Codebase Structure

**Analysis Date:** 2026-08-05

## Directory Layout

```
c4drill/
├── cmd/
│   └── c4drill/           # CLI entry point (package main, cobra)
│       ├── main.go        # main() → NewRootCmd().Execute()
│       ├── root.go        # Cobra command, flags, pipeline orchestration
│       ├── root_test.go   # CLI-level tests (help, flags, end-to-end writes)
│       └── testdata/      # CLI fixtures (valid.toml, invalid.toml, expanded.toml, expanded.dot)
├── internal/
│   ├── model/             # Domain types (Unit tree, Link, Properties, colors) — leaf package
│   │   ├── unit.go        # Unit struct, UnitType constants (C1/C2/C3 types)
│   │   ├── link.go        # Link struct, ArrowDirection/RankDirection/LabelPosition
│   │   ├── properties.go  # [properties] section struct
│   │   └── colors.go      # C4-PlantUML-derived color constants
│   ├── parser/            # TOML → parser.Model (3-pass parse)
│   │   ├── parser.go      # Parse/ParseFile, order capture, type inference
│   │   └── errors.go      # ParseError with line/context
│   ├── validator/         # Semantic validation rules
│   │   ├── validator.go   # Validate() entry, ReportErrors
│   │   ├── rules.go       # Individual rules (references, nesting, orphans, boxes)
│   │   ├── index.go       # BuildIndex (dotted-path flat map), populateIncomingLinks
│   │   ├── suggest.go     # Levenshtein typo suggestions
│   │   └── errors.go      # ValidationError / ValidationErrors
│   ├── view/              # Scoped diagram views (C1/C2/C3/Expanded)
│   │   ├── view.go        # View/Entry types, Level constants
│   │   └── scope.go       # View generators, boundary nodes, link resolution
│   ├── graph/             # Render-agnostic graph + navigation
│   │   ├── graph.go       # Graph/Node/Edge/Cluster/Label/NodeStyle types
│   │   ├── builder.go     # BuildGraph, BuildExpandedGraph, BuildGraphWithPath
│   │   ├── shapes.go      # Shape/icon/style dispatch per unit type
│   │   ├── path.go        # Explore URL / back-link / breadcrumb computation
│   │   └── navigation.go  # Navigation, BackLink, BreadcrumbItem types
│   ├── render/            # Graph → DOT/SVG via go-graphviz WASM
│   │   ├── render.go      # Render/RenderSVG/RenderDOT, wasmMutex
│   │   ├── converter.go   # graph.Graph → cgraph conversion (nodes/edges/clusters)
│   │   ├── labels.go      # Per-type HTML table label builders
│   │   ├── wrap.go        # Word wrapping, LabelRatio, char-estimation math
│   │   └── navigation.go  # BuildNavigationLabel (HTML nav bar)
│   └── output/            # File persistence
│       └── writer.go      # Writer.Write (C1/C2/C3 layout), WriteExpanded
├── skill/                 # "c4drill-toml" authoring skill + examples
│   ├── SKILL.md           # Skill definition (unit types, TOML grammar)
│   └── examples/          # 01-minimal.toml … 05-ecommerce.toml (+ generated .dot/.svg)
├── data/                  # Legacy SVG icons (person.svg, db.svg, …) — unused by render
├── testdata/              # Shared fixtures: valid/nested/links/invalid_* .toml + .dot baselines
├── cyp-auth-infra/        # Large real-world sample project + generated outputs
├── .github/workflows/     # CI: validate-skill-examples.yml (build + run examples)
├── .golangci.yml          # golangci-lint v2 config (default: all, tuned exclusions)
├── .planning/             # GSD planning docs (milestones, phases, codebase, research)
├── go.mod / go.sum        # Go 1.26.1 module; graphviz replaced with WASM fork
└── README.md              # Usage, TOML format docs
```

## Directory Purposes

**`cmd/c4drill/`:**
- Purpose: The only executable package. Owns CLI surface and pipeline composition.
- Contains: Cobra root command, flag definitions, staged orchestration
  (`runRoot`, `processView`, `processExpandedView`), expanded-path discovery.
- Key files: `main.go` (entry), `root.go` (all logic), `root_test.go`, `testdata/`.

**`internal/model/`:**
- Purpose: Pure domain types — no I/O, no logic beyond constants and small helpers.
- Contains: `Unit` (recursive tree), `Link`, `Properties`, color constants,
  `FindLinkByPeer`.
- Key files: `unit.go`, `link.go`, `properties.go`, `colors.go`.

**`internal/parser/`:**
- Purpose: TOML syntax → domain tree with order preserved and types resolved.
- Contains: 3-pass parse (`captureDefinitionOrder`, raw unmarshal, ordered build),
  default type per nesting level, generic `db`/`queue` → level-specific inference,
  `ParseError`.
- Key files: `parser.go`, `errors.go`.

**`internal/validator/`:**
- Purpose: Semantic integrity enforcement; collects all errors, not fail-fast.
- Contains: Six rules in `rules.go`, flat index + `LinksFrom` backfill in `index.go`,
  Levenshtein suggestions in `suggest.go`.
- Key files: `validator.go`, `rules.go`, `index.go`.

**`internal/view/`:**
- Purpose: Project the model tree into a flat, ordered, scope-filtered `View` with
  boundary nodes and resolved links.
- Contains: C1/C2/C3/Expanded generators, external boundary node creation, and link
  resolution passes.
- Key files: `view.go`, `scope.go`.

**`internal/graph/`:**
- Purpose: Build render-agnostic graphs (with clusters and navigation) from views.
- Contains: Graph DTOs, cluster building, style/shape/icon dispatch, URL math.
- Key files: `builder.go`, `shapes.go`, `path.go`, `graph.go`, `navigation.go`.

**`internal/render/`:**
- Purpose: Emit DOT/SVG bytes. go-graphviz WASM usage is serialized by `wasmMutex`.
- Contains: format dispatch, cgraph conversion, per-type HTML labels, wrapping.
- Key files: `render.go`, `converter.go`, `labels.go`, `wrap.go`, `navigation.go`.

**`internal/output/`:**
- Purpose: Write files/directories per the documented layout.
- Contains: `Writer` with `Write` and `WriteExpanded`.
- Key files: `writer.go`.

**`skill/`:**
- Purpose: Model-agnostic authoring skill teaching valid C4Drill TOML; examples double
  as golden fixtures (validated by CI).
- Key files: `SKILL.md`, `examples/01-minimal.toml` … `examples/05-ecommerce.toml`.

**`cyp-auth-infra/`:**
- Purpose: Large real-world sample TOML (`cyp-auth-infra.toml`,
  `saira-20260320.c2.full.toml`) with committed generated outputs (`.svg`, `.dot`,
  `.expanded.*`). Used as the load-bearing test fixture across many integration tests
  (`internal/render/expanded_internal_test.go:18` locates it via relative paths).
- Key files: `cyp-auth-infra.toml`, generated `.svg`/`.dot` baselines.

**`data/`:**
- Purpose: Legacy SVG icon files. No longer referenced by render (icons are emoji +
  native GraphViz shapes). Considered vestigial.

**`testdata/`:**
- Purpose: Shared, small fixtures used by unit/integration tests across packages
  (`.toml` inputs + expected `.dot` baselines).

## Key File Locations

**Entry Points:**
- `cmd/c4drill/main.go`: process entry (`main()`).
- `cmd/c4drill/root.go`: `NewRootCmd()` (command definition) and `runRoot()` (pipeline).

**Configuration:**
- `.golangci.yml`: lint configuration (golangci-lint v2, `default: all`).
- `go.mod`: module, Go version, dependencies, graphviz replace directive.
- `.github/workflows/validate-skill-examples.yml`: CI (build + run `skill/examples/*.toml`).

**Core Logic:**
- Parse: `internal/parser/parser.go`
- Validate: `internal/validator/validator.go`, `internal/validator/rules.go`
- Views: `internal/view/scope.go`
- Graph: `internal/graph/builder.go`, `internal/graph/path.go`
- Render: `internal/render/converter.go`, `internal/render/labels.go`
- Output: `internal/output/writer.go`

**Testing:**
- CLI: `cmd/c4drill/root_test.go` (+ `cmd/c4drill/testdata/`)
- Shared fixtures: `testdata/`, `cyp-auth-infra/`
- Per-package: co-located `*_test.go` files in every package directory.

## Naming Conventions

**Files:**
- snake_case for source files matching the primary type/function: `unit.go`,
  `properties.go`, `scope.go`, `writer.go`.
- Test files: `{source}_test.go` (e.g., `parser_test.go`, `scope_test.go`).
- Internal (white-box) tests use the `_internal_test.go` suffix:
  `expanded_internal_test.go`, `html_labels_internal_test.go`,
  `multirender_internal_test.go`.

**Directories:**
- Single-word lowercase package names under `internal/`: `model`, `parser`,
  `validator`, `view`, `graph`, `render`, `output`.
- `cmd/<binary-name>/` for the executable package.

**Packages (Go):**
- `package main` only in `cmd/c4drill`.
- Public tests use the `<pkg>_test` external package form (`package graph_test`,
  `package render_test`, `package output_test`).

## Where to Add New Code

**New Feature (e.g., a new output format):**
- Flag/validation: `cmd/c4drill/root.go` (`runRoot` format switch, `errInvalidFormat`).
- Format dispatch: `internal/render/render.go` (`Render`, `RenderWithOutput`).
- File layout: `internal/output/writer.go`.
- Tests: `cmd/c4drill/root_test.go`, `internal/output/writer_test.go`.

**New Unit Type (e.g., `cache`):**
- Type constant: `internal/model/unit.go`.
- Shape/icon/style/level: `internal/graph/shapes.go` (`IconForType`, `LevelForType`,
  `GetStyleForType`, `Is*Type` helpers).
- HTML label: `internal/render/labels.go` (`buildHTMLLabelForType` dispatch +
  dedicated builder).
- Validation: `internal/validator/rules.go` (nesting `c1Types`/`c2Types`/`c3Types`,
  allowed subunit types, box-mixing rules).
- Type inference (if generic): `internal/parser/parser.go` (`inferGenericType`,
  `defaultTypeForParent`).
- Docs: `skill/SKILL.md`.

**New Validation Rule:**
- Implementation: `internal/validator/rules.go` (function returning `ValidationErrors`).
- Registration: `internal/validator/validator.go` (`Validate` aggregation).
- Tests: `internal/validator/rules_test.go`.

**New View Type (e.g., a C4 level view):**
- Generator: `internal/view/scope.go` (mirror `GenerateC2View`/`GenerateC3View`).
- Type/constants: `internal/view/view.go` (`Level`).
- Dispatch: `cmd/c4drill/root.go` (`processView` switch + `collectExpandedPaths`).

**New Graph Transform / Navigation Feature:**
- Graph building: `internal/graph/builder.go`.
- URL/path math: `internal/graph/path.go` (all layout-sensitive code lives here).
- Tests: `internal/graph/builder_test.go`, `internal/graph/path_test.go`,
  `internal/graph/navigation_test.go`.

**Shared Utilities:**
- `internal/model/` for anything domain-shaped; no top-level `util` package exists.
- Keep `internal/output` stdlib-only and `internal/model` dependency-free.

## Special Directories

**`.planning/`:**
- Purpose: GSD workflow state — `milestones/`, `phases/`, `codebase/`, `research/`,
  `debug/`, plus top-level documents (`1-CONTEXT.md`, `PROJECT.md`, `ROADMAP.md`,
  `STATE.md`, `MILESTONES.md`, `REQUIREMENTS.md`).
- Generated: Yes (by GSD tooling).
- Committed: Yes.

**`.gsd/`, `.claude/`, `.cursor/`, `.agent/`, `.zcode/`, `.serena/`:**
- Purpose: AI-tool configuration and GSD skill/agent definitions for various IDEs.
- Generated: Tooling-injected.
- Committed: Mostly (`.gsd/runtime/` is gitignored per `git log` 6c7f266).

**`data/`:**
- Purpose: Legacy SVG icons.
- Generated: No.
- Committed: Yes (vestigial — no production references).

**`cmd/c4drill/testdata/` and `testdata/`:**
- Purpose: Test fixtures.
- Generated: No (hand-authored TOML + expected DOT).
- Committed: Yes.

**`cyp-auth-infra/`:**
- Purpose: Real-world sample + generated baselines used by integration tests.
- Generated: Partially (`.svg`/`.dot` outputs are committed baselines).
- Committed: Yes.

**`.artifacts/`:**
- Purpose: Browser-session artifacts from UI testing.
- Generated: Yes.
- Committed: Not relevant to build (runtime test output).

---

*Structure analysis: 2026-08-05*
