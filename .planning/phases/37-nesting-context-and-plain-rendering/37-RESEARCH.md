# Phase 37 Research: Nesting Context and Plain Rendering

**Researched:** 2026-08-30 (inline codebase scan — gsd-phase-researcher spawn unavailable: MCP server `exa` not connected; phase needs no web research, this is a pure codebase analysis)
**Confidence:** High — every fix point below is read from current source.

## Current-Behavior Findings (codebase scan)

### CTX — where nesting context is lost today

1. **C1 deep-link collapse (CTX-02).** `internal/view/scope.go`: `addC1BoundaryNodes` → `resolveAndAddBoundary` → `resolveBoundaryLink` (scope.go:256-351) resolves every link peer via `resolveToTopLevel` → `resolveToViewAncestor` (scope.go:353-364, D-07) to the nearest VISIBLE ancestor. A link to `linuxSystem.sshAuth.sshd` in C1 becomes an edge to `linuxSystem` (or the deepest visible subunit inside an expanded cluster). The resolved link (scope.go:336-348) carries no trace of the true deeper target — the container chain between the depicted ancestor and the real endpoint is invisible. This collapse was a deliberate anti-pollution choice ("prevents C1 from being polluted with ~100 boundary nodes", scope.go:250-255) — CTX-02 reverses it for link TARGETS.

2. **Expanded clusters flatten children (CTX-03).** `internal/graph/builder.go` `buildCluster` (610-658): for each child it computes `IsExpanded: isUnitExpanded(entry.Unit, childName)` (line 647) and `HasSubunits` — then IGNORES both and appends `buildNode(childEntry)` to `cluster.Nodes` (652-654). A nested container inside an expanded cluster renders as a flat collapsed box; grandchildren are never depicted in C1/C2/C3 views. The recursive pattern already exists one function over: `buildNestedCluster` (212-272, used by `BuildExpandedGraph`) recurses `HasSubunits → nested cluster; leaf → node`. `buildBoundaryViewGraph` (50-83) has the same flattening for C2/C3 internal entries (one buildCluster level max).

3. **CTX-01 invariant audit.** Inside any view's scope, depicted descendants already sit inside their container chain EXCEPT via mechanisms (1) and (2). Boundary nodes (sibling containers, external units) resolve to the nearest meaningful container by path divergence (`addResolvedBoundaryNode`/`resolveBoundaryByDivergence`, scope.go:714-793, the v1.9 fix) and render at TOP level by design (`Entry.IsBoundary`, view.go:76-82). **Scoping decision:** CTX-01's "complete ancestor chain" applies to elements depicted inside the view's expansion semantics (visible subunits, deep-link targets, expanded-cluster children). Sibling/external boundary nodes stay top-level — their out-of-scope position IS their context; wrapping them in ancestor clusters would unfold the whole model into every drill-down view and contradict the drill-down design.

### CTX — recommended design

- **CTX-03 (mechanical, do first):** make `buildCluster` recursive like `buildNestedCluster`: a child with subunits renders as a nested sub-cluster; leaves render as nodes. Container skeletons unfold fully; leaf units collapse (keep the 🔍 affordance on collapsed containers with subunits). Expanded-but-empty stays a plain node (D-07 guard). Decide the drill-down affordance for unfolded containers: keep the explore/explicit drill URL on the cluster label cell (nav bar uses HTML `<TD HREF>` labels — v1.8 mechanism), since the focused C2/C3 files still exist.
- **CTX-02:** when a link's true target lies under a depicted ancestor but is not itself depicted, the view generator adds the target's ancestor chain (from the depicted ancestor down to the target) as visible subunit entries (registered in `View.VisiblePaths`) and resolves the edge to the TRUE target. Applies to C1 (`resolveBoundaryLink`) and the C2/C3 cross-subunit resolution (`resolveSubunitCrossLinks`). The chain entries render via the CTX-03 recursive clusters — the target sits inside its full container chain.
- **CTX-01:** falls out of CTX-02+CTX-03 (every depicted element ends up inside its chain). Add an invariant test over the multilevel fixture.

### PLAIN — consumption points for every formatting input (scan results)

| Input | Render-side consumption | Plain behavior |
|---|---|---|
| Unit `color`/`style`/`border` | `applyUnitOverrides` (builder.go:320-339), called from buildNode:302, buildCluster:619, buildNestedCluster:221, buildBoundaryCluster:127 | Skip → type palette |
| Link `color`/`style` | createEdge (builder.go:1097+), pair aggregates `styleFor`/`kindColourFor` (builder.go:803-896) | Defaults; kind-derived colours STAY (`kindColour`, builder.go:1077 — semantic) |
| Link `length` | D-13 minlen, converter.go applyEdgeAttributes (~635-651) | Ignore → default spacing |
| Link `rank` | `RankReverse` (converter.go:555, 605 — endpoint swap + dir=back), `NoConstraint` (652-653) | Forward: no swap, no constraint suppression — must suppress BOTH emission sites |
| `properties.edges` | `View.Edges` → `Graph.EdgeStyle` → `configureGraphSettings` (converter.go:227-232) | Ignore → default splines |
| Label formatting | unit HTML tables (render/labels.go buildX HTML labels; record path buildRecordLabel:13), edge HTML rectangles (`buildEdgeLabel` labels.go:56, emitted SetLabelHTML converter.go:568-574), `labelPosition` | Plain-text labels (record/plain path), content (name/technology/description) preserved, default middle position |
| Legend + custom lines | buildLegend (builder.go:357) | STAYS (semantic; not custom formatting) |
| Queue pipes | render/pipe.go SVG post-processing | STAYS (type-default rendering, not custom formatting) |

### PLAIN — recommended design: explicit flag threading (no model mutation)

1. `cmd/c4drill/root.go`: new `--plain` bool PersistentFlag (next to `--expanded`); `runRoot` sets it on every generated view right after `view.Generate*` — `v.Plain = plain` in `processView` (295-344) and `processExpandedView` (346-383). No generator signature changes.
2. `internal/view/view.go`: new `View.Plain bool` field.
3. `internal/graph`: `Graph.Plain bool` (graph.go); `BuildGraph`/`BuildExpandedGraph` copy `v.Plain`. Guards: skip `applyUnitOverrides`; treat link Color/Style/Length/Rank as unset in edge building; `EdgeStyle: ""` when plain.
4. `internal/render/converter.go`: plain branch selects plain-text unit labels (record path) + plain edge labels (no SetLabelHTML) + middle labelPosition; style/attribute emission unchanged otherwise (defaults already flow from cleared graph fields).
5. Dual-format surface untouched — `--plain` is CLI-only (TOML/C4D model files unchanged; `convert`/`fmt` unaffected).

**Alternative considered:** a pre-render model sanitizer pass (clear Unit.Color/Style/Border, Link.Color/Style/Length/Rank/LabelPosition, Properties.Edges after Validate). Rejected as primary: HTML label formatting and rank endpoint-swap still need render-side switches anyway, and model mutation complicates golden reasoning. Flag threading keeps the model source-stable.

### BC / goldens

- canonicalDOT enforcement via `internal/testutil/canonical` (DI-1, order-insensitive). Existing goldens: `cmd/c4drill/testdata/multilevel{,.expanded}.dot`, `cmd/c4drill/testdata/expanded/**` (C2/C3 dot), plus in-package goldens (COMPAT-02, REF-05).
- CTX changes alter output only for models with expanded clusters containing subunit-containers and/or deep links — expect documented re-baselines of multilevel goldens. Flat models (testdata/*.toml: links, nested without deep scenarios) must stay clean.
- `--plain` is opt-in: default output path untouched → add new plain-mode goldens (styled fixture with --plain) rather than re-baselining existing ones.
- Note: with custom edge colours suppressed under plain, `buildLegend` (reads FINAL edge colours, "a convention absent from the view is not listed") automatically adjusts kind rows — verify this emergent behavior in tests.

### Docs surface

- `README.adoc` — features + CLI reference (add `--plain`, nesting-context behavior).
- `skill/SKILL.md` ↔ `plugins/c4drill/skills/c4drill-toml` ↔ `plugins/c4drill/opencode/skills/c4drill-toml` — CI enforces `diff -r` parity (.github/workflows/validate-skill-examples.yml:35-40); re-sync all copies in the same tasks.
- `skill/examples/` — add a nesting-context demo fixture (+.c4d twin where applicable) and a custom-formatting fixture for `--plain` before/after; examples must render cleanly through the full pipeline (CI validates).
- Release: tag v1.21.0 final task (CI release workflow builds artifacts).

## Pitfalls

1. **buildCluster recursion re-baselines every expanded golden** — plan ONE dedicated re-baseline task; do not let each task re-baseline ad hoc.
2. **Edge endpoints move deeper** (CTX-02): dedup keys, pair multiplicity counting (`countPairMultiplicity`), and kind aggregation (`collectPairAggregates`) must use the same new targets; `buildC1ViewGraph` skips `VisiblePaths` entries (builder.go:90-107, Pitfall 5 from v1.8) — new chain entries MUST be registered there or they duplicate as top-level nodes.
3. **Rank suppression has two emission sites** (converter.go:555 swap and 652 constraint) — suppress both or edges render reversed with default constraint.
4. **C1 view population order**: deep-link chain entries appended to `Units`/`UnitOrder` must respect definition-order determinism (sortBoundaryNodesByModelOrder pattern) or DOT ordering becomes nondeterministic.
5. **`--plain` × `--expanded`**: BuildExpandedGraph must honor Plain too (PLAIN-04 covers every generated file).
6. **TDD mode is on** (config) — each feature task needs RED→GREEN.

## Validation Architecture

| Requirement class | Validation vehicles |
|---|---|
| CTX-01/02/03 view semantics | `internal/view/scope_test.go` unit tests: ancestor-chain entries present for deep-link targets; resolved link targets the TRUE target; VisiblePaths registration. Fixture: `cmd/c4drill/testdata/multilevel.toml` (4-level nesting, cross-level links) |
| CTX graph rendering | `internal/graph/builder_test.go`: nested-cluster structure for expanded units with subunit-containers (canonical.Canonical, DI-1); unfolded-container explore affordance |
| PLAIN-01..04 | Per-suppression-point unit tests (overrides skipped, edge attrs defaulted, both rank sites neutral, plain labels emitted, EdgeStyle cleared, kind colours + legend retained); E2E CLI test `--plain` on a styled fixture: canonical DOT contains no explicit custom colours, no minlen, no reversed ranks, plain labels, legend block present |
| BC-01 | Existing canonicalDOT goldens green except documented CTX re-baselines; `--plain` absent = unchanged pipeline (flag default false; no-plain goldens untouched) |
| DOC-01..03 | README/skill sections exist; CI `diff -r` parity across 3 skill copies; example fixtures render through full pipeline (CI job) |
| REL-01 | `git tag v1.21.0` final task; release workflow green |

Validation is executable per-plan: view/graph unit tests + cmd E2E + canonical goldens (all existing patterns in this repo).
