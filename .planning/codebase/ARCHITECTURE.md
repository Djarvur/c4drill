<!-- refreshed: 2026-08-05 -->
# Architecture

**Analysis Date:** 2026-08-05

## System Overview

C4Drill is a single-binary Go CLI that transforms a TOML architecture description into
C4-style diagrams (SVG or DOT) with automatic drill-down navigation. The tool is a
**staged pipeline**: parse → validate → project views → build graph → render → write.
Each stage is a separate package under `internal/`, and `cmd/c4drill` orchestrates them.

```text
┌─────────────────────────────────────────────────────────────────────────────┐
│                           CLI Orchestration (cobra)                          │
│                          cmd/c4drill/root.go                                 │
│       runRoot → collectExpandedPaths → processView per unit path             │
├────────────────┬────────────────┬──────────────────┬─────────────────────────┤
│  Stage 1       │  Stage 2       │  Stage 3-6       │  --expanded mode        │
│  Parse         │  Validate      │  per-path loop   │  single all-nested view  │
└───────┬────────┴───────┬────────┴────────┬─────────┴──────────┬──────────────┘
        │                │                 │                    │
        ▼                ▼                 ▼                    ▼
┌──────────────┐  ┌──────────────┐  ┌─────────────────┐  ┌──────────────────┐
│ parser       │  │ validator    │  │ view            │  │ view             │
│ parser.go    │  │ validator.go │  │ scope.go        │  │ GenerateExpanded │
│ TOML→Model   │  │ rules.go     │  │ C1/C2/C3 views  │  │ View             │
└──────┬───────┘  └──────────────┘  └────────┬────────┘  └────────┬─────────┘
       │                                     │                    │
       ▼                                     ▼                    ▼
┌──────────────┐                    ┌─────────────────┐  ┌──────────────────┐
│ model        │                    │ graph           │  │ graph            │
│ unit.go      │                    │ builder.go      │  │ BuildExpandedGraph│
│ link.go      │                    │ shapes.go       │  │                  │
│ (domain)     │                    │ path.go (nav)   │  │                  │
└──────────────┘                    └────────┬────────┘  └────────┬─────────┘
                                             │                    │
                                             ▼                    ▼
                                    ┌──────────────────────────────────────┐
                                    │ render (go-graphviz WASM, serialized)│
                                    │ render.go / converter.go / labels.go │
                                    └──────────────────┬───────────────────┘
                                                       ▼
                                    ┌──────────────────────────────────────┐
                                    │ output (Writer) → SVG/DOT file tree  │
                                    │ output/writer.go                     │
                                    └──────────────────────────────────────┘
```

## Component Responsibilities

| Component | Responsibility | File |
|-----------|----------------|------|
| `cmd/c4drill` | CLI flags, pipeline orchestration, expanded-path auto-detection | `cmd/c4drill/root.go` |
| `model` | Domain types: `Unit` tree, `Link`, `Properties`, color constants | `internal/model/unit.go`, `internal/model/link.go`, `internal/model/colors.go` |
| `parser` | TOML → `parser.Model` (3-pass parse), default/inferred type resolution | `internal/parser/parser.go` |
| `validator` | Semantic integrity rules, incoming-link backfill, typo suggestions | `internal/validator/validator.go`, `internal/validator/rules.go`, `internal/validator/suggest.go` |
| `view` | Scoped projections (C1/C2/C3/Expanded), boundary nodes, link resolution | `internal/view/scope.go`, `internal/view/view.go` |
| `graph` | Render-agnostic graph (nodes/edges/clusters), navigation paths, styles | `internal/graph/builder.go`, `internal/graph/shapes.go`, `internal/graph/path.go` |
| `render` | Graph → DOT/SVG via go-graphviz (WASM), HTML labels, word wrap | `internal/render/render.go`, `internal/render/converter.go`, `internal/render/labels.go`, `internal/render/wrap.go` |
| `output` | File writes, directory hierarchy from dotted paths | `internal/output/writer.go` |

## Pattern Overview

**Overall:** Staged pipeline (functional decomposition) with a shared domain model. Each
stage consumes the previous stage's output and produces the next stage's input. No stage
skips ahead; `cmd/c4drill` is the only composer.

**Key Characteristics:**
- Strict one-directional dependency flow: `model → parser → validator` and
  `model → parser → view → graph → render → output`, with `cmd/c4drill` depending on all.
- The domain model (`model.Unit`) is a **tree** keyed by dotted paths (e.g., `mainapp.api`).
- Each stage introduces its own enriched representation: `model.Unit` (parse tree) →
  `view.Entry` (scoped projection) → `graph.Node`/`graph.Cluster` (render graph).
- Definition order is preserved at every stage via explicit `UnitOrder` / `SubunitOrder`
  slices (TOML maps are unordered).
- Rendering is centralized on go-graphviz's WASM engine guarded by a global mutex.

## Layers

**Model Layer (`internal/model/`):**
- Purpose: Domain types shared by all stages.
- Location: `internal/model/`
- Contains: `Unit` (recursive tree node), `Link`, `Properties`, `ArrowDirection`,
  `RankDirection`, `LabelPosition`, color constants.
- Depends on: nothing internal (stdlib only).
- Used by: every other package.

**Parsing Layer (`internal/parser/`):**
- Purpose: Convert TOML bytes into a `parser.Model` with ordered units.
- Location: `internal/parser/`
- Contains: `Parse`/`ParseFile`, definition-order capture via the go-toml `unstable` API,
  default-type and generic-type (`db`/`queue`) inference, `ParseError`.
- Depends on: `internal/model`, go-toml v2.
- Used by: `cmd/c4drill`, `validator`, `view`.
- Note: `parser.Model` is the de-facto root domain type; `view` and `validator` accept
  `*parser.Model` directly, coupling those layers to the parser package (see Anti-Patterns).

**Validation Layer (`internal/validator/`):**
- Purpose: Semantic checks that the parsed model is a legal C4 hierarchy.
- Location: `internal/validator/`
- Contains: `Validate` (runs all rules, collects errors, not fail-fast), rules
  (references, subunit rules, link rules, orphans, nesting hierarchy, box mixed
  contents), `BuildIndex` (flat dotted-path lookup), `populateIncomingLinks` (backfills
  `LinksFrom` from other units' `Links`), Levenshtein-based `SuggestSimilar`.
- Depends on: `internal/model`, `internal/parser`.
- Used by: `cmd/c4drill`.

**View Layer (`internal/view/`):**
- Purpose: Project the model tree into a scoped, flat view for one diagram.
- Location: `internal/view/`
- Contains: `View` (level, title, ordered entries), `Entry` (unit + path + expansion/
  external flags + `ResolvedLinks`), generators `GenerateC1View` / `GenerateC2View` /
  `GenerateC3View` / `GenerateExpandedView`, boundary-node creation, and several
  link-resolution passes that rewrite deep peer paths to nearest visible ancestors.
- Depends on: `internal/model`, `internal/parser`.
- Used by: `cmd/c4drill`, `graph`, `render` (via `view.View`).

**Graph Layer (`internal/graph/`):**
- Purpose: Turn a `view.View` into a `graph.Graph` ready for rendering.
- Location: `internal/graph/`
- Contains: `Graph`/`Node`/`Edge`/`Cluster`/`Label`/`NodeStyle`/`Navigation` types,
  `BuildGraph` (C1 flat + C2/C3 boundary cluster), `BuildExpandedGraph` (nested
  clusters), `BuildGraphWithPath` (adds `ExploreURL` + navigation), shape/style/icon
  dispatch, URL computation (`ComputeExploreURL`, `ComputeBackLinkURL`,
  `BuildBreadcrumbPath`).
- Depends on: `internal/model`, `internal/view`.
- Used by: `cmd/c4drill`, `render`.

**Render Layer (`internal/render/`):**
- Purpose: Serialize `graph.Graph` to DOT or SVG bytes using go-graphviz.
- Location: `internal/render/`
- Contains: `Render`/`RenderSVG`/`RenderDOT`, the graphviz-to-cgraph converter
  (`buildCgraph`, node/edge/cluster creation), HTML table label builders per unit type,
  word-wrapping with `LabelRatio`, navigation label builder.
- Depends on: `internal/graph`, `internal/model`.
- Used by: `cmd/c4drill`.

**Output Layer (`internal/output/`):**
- Purpose: Persist rendered bytes to disk with the documented file layout.
- Location: `internal/output/`
- Contains: `Writer` with `Write` (C1 flat file, C2/C3 directory hierarchy) and
  `WriteExpanded`.
- Depends on: stdlib only.
- Used by: `cmd/c4drill`.

## Data Flow

### Primary Request Path (default mode)

1. Entry — `main()` calls `NewRootCmd().Execute()` (`cmd/c4drill/main.go:8`).
2. Flag/args validation and defaults — `runRoot` (`cmd/c4drill/root.go:84`): validates
   `--format`, resolves `--output` default to input file's directory, sets
   `render.LabelRatio` from `--label-ratio` / `C4DRILL_LABEL_RATIO` / 1.6
   (`cmd/c4drill/root.go:296`).
3. Stage 1 Parse — `parser.ParseFile(inputPath)` (`cmd/c4drill/root.go:111` →
   `internal/parser/parser.go:322`). Three passes: `captureDefinitionOrder` (unstable
   API, `parser.go:100`), raw-map unmarshal, then ordered struct build with recursive
   subunit parsing (`parseUnitWithOrder`, `parser.go:160`).
4. Stage 2 Validate — `validator.Validate(m)` (`internal/validator/validator.go:16`).
   On errors, `validator.ReportErrors` prints them and `runRoot` returns
   `errValidationFailed` (`cmd/c4drill/root.go:117`).
5. Diagram discovery — `collectExpandedPaths(m)` (`cmd/c4drill/root.go:155`): always
   includes `""` (C1) plus every unit with subunits (recursive auto-detect).
6. Per-diagram pipeline — `processView(m, unitPath, basename, writer)`
   (`cmd/c4drill/root.go:202`):
   - View generation — `view.GenerateC1View` (`internal/view/scope.go:88`),
     `view.GenerateC2View` (`scope.go:390`), or `view.GenerateC3View` (`scope.go:458`)
     selected by `isC2Path` (`cmd/c4drill/root.go:289`).
   - Graph building — `graph.BuildGraphWithPath(v, unitPath, basename, format)`
     (`internal/graph/builder.go:474`), which calls `BuildGraph` and adds
     `ExploreURL`s plus `Navigation` for C2/C3.
   - Rendering — `render.RenderSVGWithOutput(g, baseDir)` or `render.Render(g, format)`
     (`cmd/c4drill/root.go:231` → `internal/render/render.go:47`).
   - Writing — `writer.Write(basename, unitPath, format, data)`
     (`internal/output/writer.go:37`).
7. Silent success on completion (per spec); non-zero exit + error text on failure.

### Expanded Mode (`--expanded`)

1. `processExpandedView` (`cmd/c4drill/root.go:251`):
   - `view.GenerateExpandedView(m)` (`internal/view/scope.go:14`) adds every unit at all
     nesting levels depth-first.
   - `graph.BuildExpandedGraph(v)` (`internal/graph/builder.go:125`) builds recursive
     nested clusters via `buildNestedCluster` (`builder.go:172`).
   - Rendered and written to `{basename}.expanded.{format}` via `writer.WriteExpanded`
     (`internal/output/writer.go:66`).
   - No navigation is attached (single-file mode).

### Link Resolution Flow (C1/C2/C3)

Deep-nested links are rewritten so edges connect the units actually visible in a diagram:

1. `view.GenerateC1View` runs `addC1BoundaryNodes` → `resolveAndAddBoundary`
   (`internal/view/scope.go:253`, `scope.go:278`): every link (recursive) is resolved
   to its nearest top-level ancestor via `resolveToTopLevel` (`scope.go:368`), boundary
   nodes are created for absent units, and the resolved links are collected on the
   top-level source's `ResolvedLinks`/`ResolvedLinksFrom`.
2. `view.GenerateC2View`/`GenerateC3View` run three passes: boundary nodes from subunit
   links (`addExternalBoundaryNodesForSubunits`, `scope.go:577`), boundary-node link
   resolution (`resolveBoundaryNodeLinks`, `scope.go:652`), and cross-subunit link
   synthesis (`resolveSubunitCrossLinks`, `scope.go:750`).
3. `graph.buildEdges` (`internal/graph/builder.go:329`) prefers `entry.ResolvedLinks`
   over `entry.Unit.Links` when present, deduplicates via a `seen` key map, and skips
   peers not present in the view (`isTargetInView`, `builder.go:414`).

### Output File Layout

- C1: `{basename}.{format}` (`internal/output/writer.go:41`)
- C2: `{basename}/{system}.{format}`
- C3: `{basename}/{system}/{container}.{format}`
- Expanded: `{basename}.expanded.{format}` (`internal/output/writer.go:67`)
- Dotted paths become directory hierarchies (`strings.ReplaceAll(unitPath, ".", sep)`).
- Clickable URLs in SVG always use `.svg` regardless of `--format` (browser navigation)
  (`internal/graph/path.go:16`).

**State Management:**
- No external state store. Everything is in-memory per invocation; the only mutable
  package-level state is `render.LabelRatio` (CLI-config) and the `wasmMutex` guard.
- Unit/section ordering is threaded through the pipeline via `UnitOrder`/`SubunitOrder`
  slices so map iteration order never affects output.

## Key Abstractions

**`parser.Model` (parse tree root):**
- Purpose: Root of the parsed document — `[properties]` plus ordered top-level units.
- Location: `internal/parser/parser.go:35`
- Pattern: Aggregate; owns `Units map[string]*model.Unit` and `UnitOrder []string`.
- Note: despite living in `parser`, it is used as the domain entry point by `validator`
  and `view`.

**`model.Unit` (recursive tree node):**
- Purpose: A C4 element at any level with typed metadata and nested subunits.
- Location: `internal/model/unit.go:41`
- Pattern: Composite; `Subunits map[string]*Unit` plus `SubunitOrder`.
- Links are bidirectional-ish: `Links` (outgoing) and `LinksFrom` (incoming backfilled
  by the validator at `internal/validator/index.go:53`).

**`view.View` / `view.Entry` (scoped projection):**
- Purpose: Flat, ordered snapshot of the units visible in one diagram plus link
  resolution results.
- Location: `internal/view/view.go:20`, `view.go:45`
- Pattern: Flattening of the tree with metadata enrichment (`IsExpanded`,
  `IsExternal`, `HasSubunits`, `ResolvedLinks`).

**`graph.Graph` / `Node` / `Edge` / `Cluster` (render graph):**
- Purpose: Render-agnostic intermediate representation consumed by the render layer.
- Location: `internal/graph/graph.go:34`
- Pattern: DTO graph; clusters represent expanded units (subgraph clusters), nodes are
  individual units, edges carry label/style/arrow/color/minlen.

**`graph.Navigation` / `BackLink` / `BreadcrumbItem`:**
- Purpose: Drill-down navigation metadata for C2/C3 diagrams.
- Location: `internal/graph/navigation.go:5`
- Pattern: Value objects produced by `buildNavigation` (`internal/graph/builder.go:528`)
  and rendered to an HTML clickable bar by `BuildNavigationLabel`
  (`internal/render/navigation.go:13`).

## Entry Points

**CLI binary:**
- Location: `cmd/c4drill/main.go:7`
- Triggers: `go run ./cmd/c4drill`, `go install`, built binary.
- Responsibilities: Build the cobra root command and execute it; exit code 1 on error.

**Cobra root command:**
- Location: `cmd/c4drill/root.go:46` (`NewRootCmd`)
- Triggers: any CLI invocation; `RunE: runRoot` (`root.go:84`) is the pipeline entry.
- Flags: `--format/-f` (dot|svg), `--output/-o`, `--expanded`, `--label-ratio`.

## Architectural Constraints

- **Threading:** Single-threaded rendering. go-graphviz uses a WASM engine that is not
  thread-safe; all render calls are serialized by `wasmMutex`
  (`internal/render/render.go:20`). Consequently render-invoking tests must NOT call
  `t.Parallel()` (see `cmd/c4drill/root_test.go:15` and
  `internal/render/integration_test.go:9`).
- **Global state:** `render.LabelRatio` (`internal/render/wrap.go:36`) is a package var
  set by the CLI before rendering; `wasmMutex` (`internal/render/render.go:20`).
  Both carry `//nolint:gochecknoglobals`.
- **Package layering:** `internal/model` is a strict leaf. `internal/output` is a strict
  leaf (stdlib only). No package below `cmd/c4drill` may depend on `cmd/c4drill`.
- **Order preservation:** TOML maps are unordered; the pipeline relies on
  `UnitOrder`/`SubunitOrder` slices captured by the parser (`internal/parser/parser.go:100`)
  and propagated into views (`view.UnitOrder`) and the graph (`graph.Nodes` iteration).
- **go.mod replace directive:** `github.com/goccy/go-graphviz` is replaced with a fork
  `github.com/onokonem/go-graphviz` (`go.mod`); `.golangci.yml` allow-lists this replace.
- **Lint gate:** golangci-lint v2 with `default: all` linters (see `.golangci.yml`).
  New code must satisfy all enabled linters; intentional globals use
  `//nolint:gochecknoglobals`.

## Anti-Patterns

### Parser-coupled Model Type

**What happens:** The root parse-tree type `parser.Model` is consumed directly by
`validator.Validate` (`internal/validator/validator.go:16`) and all view generators
(`internal/view/scope.go:88`), so `validator` and `view` import `parser`.
**Why it's wrong:** The view layer is conceptually downstream of parsing but cannot be
reused without the parser package; domain identity leaks into the parser package.
**Do this instead:** Move the root model type into `internal/model` (e.g.,
`model.Document`/`model.Model`) and have `parser` return it, keeping `view`/`validator`
dependent on `model` only.

### Duplicate Peer-Resolution Logic

**What happens:** Peer-path → nearest-visible-ancestor resolution is implemented
several times with slightly different semantics: `resolveToTopLevel`
(`internal/view/scope.go:368`), `resolveToViewAncestor` (`scope.go:716`),
`addResolvedBoundaryNode` (`scope.go:619`), plus the three orchestration passes
(`resolveAndAddBoundary`, `resolveBoundaryNodeLinks`, `resolveSubunitCrossLinks`).
**Why it's wrong:** High cognitive load; subtle differences in behavior (e.g., whether
the bare top-level name is re-checked after trimming) make the resolution pipeline hard
to reason about and test exhaustively.
**Do this instead:** Extract a single resolution helper (input: view units + peer path;
output: resolved peer or "") and drive all passes from it, with the C1 ancestor mode and
C2/C3 ancestor mode as explicit options.

### Parallel Enums for Arrows

**What happens:** `model.ArrowDirection` (`internal/model/link.go:7`: `forward`,
`reverse`, `bidirectional`, `none`) and `graph.ArrowDirection`
(`internal/graph/graph.go:19`: `forward`, `reverse`, `both`, `none`) model the same
concept with different value sets; `builder.createEdge` casts one to the other
(`internal/graph/builder.go:431`: `ArrowDirection(link.Arrow)`).
**Why it's wrong:** Two sources of truth; `bidirectional` vs `both` naming divergence
invites conversion bugs if one side changes.
**Do this instead:** Keep one enum in `internal/model` and reference it from `graph`.

### Unused/legacy API surface in render

**What happens:** `RenderSVGWithOutput` and `RenderWithOutput` accept an `outputDir`
parameter that is documented as unused (`internal/render/render.go:47`, `render.go:69`);
`iconTypeForUnit` is dead code (`internal/render/labels.go:423`); `joinLabels` contains
a vestigial `resultSb120` variable and dead assignment (`internal/render/converter.go:200`).
**Why it's wrong:** Confusing for new contributors; implies icon extraction and multiple
render entry points that no longer exist.
**Do this instead:** Remove the unused parameter variants (update callers in
`cmd/c4drill/root.go`), delete the deprecated function, and clean up `joinLabels`.

### Duplicate Expanded-Detection Helpers

**What happens:** `collectExpandedPaths`/`collectExpandableUnitPaths`
(`cmd/c4drill/root.go:155`) duplicate the "has subunits ⇒ has sub-diagram" logic that
also appears in `view` (e.g., `HasSubunits` flags in `internal/view/scope.go`) and the
graph builder (`buildCluster` in `internal/graph/builder.go:265`).
**Why it's wrong:** Diagram discovery lives in the CLI layer instead of the view layer,
so behavior can drift between `collectExpandedPaths` and what the view generators mark
as expandable.
**Do this instead:** Add an exported `view.ExpandablePaths(m)` (or similar) and call it
from the CLI.

## Error Handling

**Strategy:** Fail-fast staged pipeline in `cmd/c4drill` with wrapped errors
(`fmt.Errorf("%w: ...")`). Static sentinel errors for control flow:
`errInvalidFormat`, `errValidationFailed`, `errGenerateView`, `errBuildGraph`
(`cmd/c4drill/root.go:23`); render-layer sentinels `ErrNilGraph`, `ErrUnsupportedFormat`
(`internal/render/render.go:23`).

**Patterns:**
- Parse errors are typed `*parser.ParseError` with line/context and `Unwrap`
  (`internal/parser/errors.go:13`).
- Validation collects ALL errors (not fail-fast) into `ValidationErrors` and reports
  them via `ReportErrors` (`internal/validator/validator.go:47`).
- Render returns wrapped errors from each cgraph-building step
  (`internal/render/converter.go:94`).

## Cross-Cutting Concerns

**Logging:** No logging framework. CLI is silent on success ("silent per spec",
`cmd/c4drill/root.go:148`); errors print to stderr. Tests use `t.Logf` sparingly
(e.g., `internal/render/expanded_internal_test.go:27`).
**Validation:** Two-phase — syntax (parser) then semantics (validator). Semantic rules
are independent functions in `internal/validator/rules.go`, aggregated by `Validate`.
**Authentication:** None (local CLI tool, no network or auth).
**Navigation (cross-cutting):** URL generation is centralized in
`internal/graph/path.go`; all explore/back/breadcrumb links derive from dotted paths,
`basename`, and format so file layout and links cannot diverge.

---

*Architecture analysis: 2026-08-05*
