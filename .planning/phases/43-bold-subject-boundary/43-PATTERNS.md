# Phase 43: Bold Subject Boundary - Pattern Map

**Phase:** 43 - Bold Subject Boundary
**Generated:** 2026-09-03
**Sources:** 43-CONTEXT.md (D-01..D-05), 43-RESEARCH.md, codebase read-only scan

## File Classification

| File | Role | Data Flow | Status |
|------|------|-----------|--------|
| `internal/graph/graph.go` | value type (style carrier) | builder → converter (Cluster.Style) | modify (~line 195 NodeStyle) |
| `internal/graph/builder.go` | construction (emphasis decision) | view → Cluster.Style | modify (`buildBoundaryCluster` :373-441) |
| `internal/render/converter.go` | attribute emission | Cluster.Style → DOT/SVG/HTML | modify (`applyClusterStyle` :656-695) |
| `internal/graph/builder_test.go` | unit test (emphasis + byte guards) | test | modify (exists) |
| `internal/render/converter_test.go` | unit test (raw-DOT cluster attribute multiset) | test | modify (exists) |
| `cmd/c4drill/root_test.go` | integration test (CLI render tree + flag matrix) | test | modify (exists) |
| `cmd/c4drill/testdata/multilevel.toml` | multi-level fixture (C2 + deep-link C3) | fixture | reuse (exists) |
| `cmd/c4drill/testdata/*.dot` | goldens (all expanded/plain/nolabels variants) | fixture | re-baseline ONLY with penwidth-only canonical diff (expected: none) |
| `.planning/PROJECT.md` | milestone docs ("As of v1.18" paragraph) | doc | modify (43-02) |

## Pattern Assignments

### `internal/graph/graph.go` — NodeStyle gains `BorderWidth float64`
**Analog:** `Edge.PenWidth` (graph.go, builder_edge family) — the established 0-means-default numeric field (D-02 mirror contract). `PenWidth 0` = renderer default; builder assigns >0 only where a thicker edge is semantic. `BorderWidth` copies the field contract exactly: `0` = no attribute emitted; only the subject boundary cluster's style is ever non-zero.
**Key excerpt (the contract to mirror):**
```go
// Edge.PenWidth — 0 means "renderer default"; >0 emits penwidth at render
// (see converter.go:899-905 for the 0 -> 1.0 fallback path)
```

### `internal/graph/builder.go` — emphasis assignment in `buildBoundaryCluster`
**Analog:** the function itself, plus the guard evidence in `applyUnitOverrides` (builder.go:624):
```go
func applyUnitOverrides(style *NodeStyle, unit *model.Unit, opts RenderOpts) {
    if opts.Plain || style == nil || unit == nil {
        return
    }
```
The emphasis MUST go in `buildBoundaryCluster` after the overrides call (builder.go:389 `applyUnitOverrides(style, unit, renderOptsFromView(v))`), one `style.BorderWidth = 3.0` before the `return &Cluster{...}` (covers the nil-unit fallback branch too). This is the single construction site per D-01 — no other cluster construction path changes.
**Comment-archaeology convention:** anchor the line with a `BOLD-01`/D-03 reference comment like the existing "fix 260831-01u" comments.

### `internal/render/converter.go` — `penwidth` emission in `applyClusterStyle`
**Analog:** the function's own attribute emission block (converter.go:671-689) — `setClusterAttribute(subgraph, "color", style.BorderColor)` etc. — plus the edge penwidth precedent (converter.go:899-905, `if edge.PenWidth > 0 { e.SetPenWidth(...) }`). Clusters have no `SetPenWidth` (Edge-only in go-graphviz); emission is `setClusterAttribute(subgraph, "penwidth", strconv.FormatFloat(style.BorderWidth, 'f', 1, 64))` gated by `style.BorderWidth > 0` — the "0 = default" guard mirrors the edge precedent.
**Key excerpt (the emission plumbing):**
```go
// converter.go:696-704 — skips empty values; SafeSet(attr, value, "")
func setClusterAttribute(subgraph *cgraph.Graph, attr, value string) error {
    if value == "" {
        return nil
    }
```

### `internal/graph/builder_test.go` — emphasis + byte-identity guards
**Analog:** existing boundary/cluster tests in the same file + the canonical golden consumers (builder_test.go:1233-1241 COMPAT-02, 2846-2851 REF-05 — `multilevel.expanded.dot` must stay green UNCHANGED). Build views via `view.GenerateC2View(m, path)` / `view.GenerateC3View(m, path)` (scope.go:487/594), then `graph.BuildGraph(v)` and `render.RenderDOT(g)` — the direct pipeline used by sibling tests (e.g., TestBuildGraph_ExpandedClusterRendersNestedSubClusters builds full `parser.Model` fixtures in-test; the multilevel fixture file covers CLI-level).

### `internal/render/converter_test.go` — raw-DOT cluster attribute multiset
**Analog:** `dotEdgeMultiset` (converter_test.go:1257-1283) — regexp extraction over the raw DOT with `penwidth=([\d.]+)` (line 1263), sorted multiset compares. New helper: extract per-subgraph attribute sets; assert exactly ONE cluster (`subgraph cluster_<subject>`) carries `penwidth=3.0` and zero others carry any penwidth.

### `cmd/c4drill/root_test.go` — end-to-end flag matrix on the public fixture
**Analog:** `generateMultilevelOutput` harness (root_test.go:680-800) — CLI renders `multilevel.toml` into C1 `multilevel.dot`, C2 `multilevel/mainSystem.dot`, C3 `multilevel/mainSystem/sshAuth.dot`; plus `TestCompat02_MultilevelFixtureFiveNodeC1` structure assertions (root_test.go:693-722). The flag-matrix variant (`--plain`, `--no-styles`, `--no-colors`, `--expanded`) rides this harness for the D-03/D-04 surface — the same shape as the existing penwidth flag-invariance matrix (converter_test.go:1379-1402).

### `.planning/PROJECT.md` — "As of v1.18" paragraph
**Analog:** the "As of v1.15/v1.16/v1.17" paragraphs (PROJECT.md:9-13) — one sentence per milestone capability, placed in the same evolution section; the v1.18 sentence describes the semantic bold subject boundary surviving `--plain`/`--no-styles`. Milestone evolution commits follow the `docs(phase-N): evolve PROJECT.md after phase completion` pattern (git log e.g. `486fe3c`).

## Shared Patterns

- **Testify:** `github.com/stretchr/testify/require` + `assert` — every new/modified test.
- **WASM serialization:** render tests must NOT call `t.Parallel()` — `//nolint:paralleltest` on any test exercising `render.RenderDOT` (root_test.go:689 convention; go-graphviz WASM engine concurrency).
- **Canonical, never raw bytes for whole-DOT compares:** goldens through `internal/testutil/canonical` only (DI-1); targeted attribute assertions may use regexp/multiset extraction on raw DOT.
- **TDD commits:** `test(43-01): ...` (RED) MUST precede `feat(43-01): ...` (GREEN); `refactor(43-01)` only when REFACTOR changes code.
- **Comment archaeology:** anchor changes to decision IDs (D-01..D-05 / BOLD-01..03) in code comments at fix sites.
- **Golden hygiene:** never re-baseline a golden without auditing its canonical diff; this phase's bar: diff = the sole `penwidth` line on the SUBJECT cluster only.