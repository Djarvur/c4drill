# Phase 40: Deterministic SVG Output - Pattern Map

**Phase:** 40 - Deterministic SVG Output
**Generated:** 2026-09-03
**Sources:** 40-CONTEXT.md (D-01..D-06), 40-RESEARCH.md, codebase read-only scan

## File Classification

| File | Role | Data Flow | Status |
|------|------|-----------|--------|
| `internal/validator/index.go` | source-fix (mirror synthesis) | model → LinksFrom mirrors | modify |
| `internal/graph/builder.go` | defense-in-depth (edge slice ordering) | view → g.Edges | modify |
| `internal/validator/index_test.go` | unit test (mirror-order determinism) | test | modify (exists) |
| `internal/render/deterministic_test.go` | integration test (byte-equality pin, issue #42 reproducer) | model → graph → bytes | create |
| `internal/graph/builder_test.go` | unit test (name-sorted edge order) | test | modify (exists) |

## Pattern Assignments

### `internal/validator/index.go` (validator, map→slice materialization)
**Analog:** itself — the fix shape is a sorted-keys loop replacing a bare map range (index.go:54).
**Copy from:** the repo's canonical definition-order walk, `modelUnitOrder` (internal/view/scope.go:92-105) captures definition order into a slice; `sort.Strings` precedent lives in `internal/testutil/canonical/canonical.go` (stdlib `sort` import, no new deps).
**Key excerpt (current code, the bug):**
```go
// internal/validator/index.go:53-56
func populateIncomingLinks(index map[string]*UnitInfo) {
	for sourcePath, sourceInfo := range index {
		for _, link := range sourceInfo.Unit.Links {
```
The mirror-append body (index.go:56-84, `model.FindLinkByPeer` dedup + `Mirror: true` append) stays byte-identical — only the outer iteration becomes sorted.

### `internal/graph/builder.go` (graph builder, slice finalization)
**Analog:** `buildEdges` tail (builder.go:1030-1074) plus the existing stable-sort precedent `sort.SliceStable` usage in `internal/testutil/canonical` and `slices.Contains` usage at builder.go:1026 (Go 1.26 `slices` package is already imported in this file's package family — verify import block before use).
**Key excerpt (sort key already materialized):**
```go
// internal/graph/builder.go:1094-1103 (assignEdgeName)
// names are unique per drawn edge: sanitizeEdgeName(source)_to_(target)_n
```
Insert `slices.SortStableFunc(edges, ...)` comparing `strings.Compare(a.Name, b.Name)` as the final statement of `buildEdges` before `return edges` (per D-03).

### `internal/validator/index_test.go` (unit test, testify style)
**Analog:** its own package's `errors_test.go` naming (`TestValidationError_WithLine` — table-less single-concern tests) and `internal/validator/index_test.go` existing helpers. Determinism loop test pattern: run `BuildIndex` + `populateIncomingLinks` N times over the same model, collect the mirrored `LinksFrom` peer sequences, assert all runs equal run 1 (use `testify/assert`, per repo convention).
**Note:** `populateIncomingLinks` is unexported — the test lives in package `validator` (internal test), which is the existing convention in this package.

### `internal/render/deterministic_test.go` (integration test, byte-equality pin) — NEW FILE
**Analog:** `internal/render/integration_test.go` (full-pipeline render tests in the render package) + `cmd/c4drill/convert_test.go:458` `runGraphPipeline` for the pipeline sequence shape: `c4d.Parse` → `include.Resolve` → `template.Expand` → `peer.Resolve` → `validator.Validate` → `view.GenerateC1View` → `graph.BuildGraphWithPath` → `render.Render{SVG,DOT,HTML}`.
**Model fixture:** build the issue #42 reproducer string in-test (10 containers, `-> sys.c{j}` for `j != i && (i+j)%3 == 0`, plus `ext -> sys.c0`) — no testdata file needed; write to `t.TempDir()` only if a file path is required by the parse entry (`c4d.Parse([]byte)` accepts bytes directly).
**Assertions:** two independent pipeline runs → `require.Equal` on the byte slices for `svg`, `dot`, `html` (REPRO-01/03). Expect RED: svg/html differ pre-fix; dot may pass even in RED (per-issue evidence) — the svg assertion carries the RED gate.
**Marker:** `//nolint:paralleltest` if the pipeline touches shared state (follow `cmd/c4drill` convention when in doubt; render package tests are isolated per call under `wasmMutex`).

### `internal/graph/builder_test.go` (unit test, name-sorted order)
**Analog:** `TestBuildGraph_EdgeOverride` (builder_test.go:4706) — build a small view via `graph.BuildGraph(v)`, assert on `g.Edges`. New test: after the D-03 sort, `g.Edges[i].Name <= g.Edges[i+1].Name` for all i, and the sequence is identical across repeated `buildEdges` invocations on freshly built views.

## Shared Patterns

- **Testify:** `github.com/stretchr/testify/require` + `assert` — every new test.
- **TDD commits:** `test(40-01): ...` / `feat(40-01): ...` / `refactor(40-02): ...` per workflow.tdd_mode gate enforcement (RED commit MUST precede GREEN).
- **Comment archaeology:** this repo anchors changes to fix IDs (e.g., "fix 260831-01u") — reference issue #42 and decision IDs (D-02, D-03) in code comments at fix sites.
- **No golden churn:** canonicalDOT comparisons go through `internal/testutil/canonical` — never raw-byte DOT comparisons; SVG id shifts (if any) are audited by canonicalizing `id="edge<N>"` (D-06).
- **WASM serialization:** all render calls funnel through `wasmMutex` inside `render()` — tests must not spawn parallel renders (no `t.Parallel()` on render integration tests).
