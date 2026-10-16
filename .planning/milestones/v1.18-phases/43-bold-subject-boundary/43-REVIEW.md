---
phase: 43
slug: bold-subject-boundary
status: clean
files_reviewed: 6
depth: standard
critical: 0
warning: 0
info: 0
total: 0
reviewed: 2026-09-07
---

# Phase 43 Code Review — Bold Subject Boundary

## Scope

Reviewed the six phase files at standard depth (cross-file, call-chain aware):

- `internal/graph/graph.go` — NodeStyle.BorderWidth field addition
- `internal/graph/builder.go` — buildBoundaryCluster emphasis assignment
- `internal/render/converter.go` — applyClusterStyle penwidth emission
- `internal/graph/builder_test.go` — TestSubjectBoundaryNoCollateral + cluster-attr helper
- `internal/render/converter_test.go` — TestSubjectBoundaryClusterPenwidth + cluster-attr helper
- `cmd/c4drill/root_test.go` — TestSubjectBoundaryCLIFlagMatrix + cluster-attr helper

Diff base: `25df2a2^` (docs(43): create phase plan predecessor), i.e. commits `7ed953f` (test) + `bf2007d` (feat) + `a68b77d`/`e45f24c` (docs).

## Findings

No critical, warning, or info findings.

## Checks Performed

1. **Nil-safety**: `applyClusterStyle` guards `style == nil` before the new `style.BorderWidth > 0` dereference; both branches of `buildBoundaryCluster` assign a non-nil style before `style.BorderWidth = 3.0` — no nil dereference path.
2. **Value provenance**: `BorderWidth` is a compile-time constant (3.0) set at graph construction; no model/author input can reach the `penwidth` attribute — no injection surface added (matches STRIDE T-43-01 acceptance).
3. **Formatting determinism**: `strconv.FormatFloat(style.BorderWidth, 'f', 1, 64)` emits exactly `"3.0"` for all non-zero values (D-01 literal); `fmt.Sprint` never used; the `0 = renderer default` contract means the gate `> 0` and the empty-skip in `setClusterAttribute` are consistent.
4. **NaN/negative defense**: a hypothetical NaN or negative `BorderWidth` fails the `> 0` gate and emits nothing — fail-safe, though unreachable (constant only).
5. **Scope isolation**: the emission sits in `applyClusterStyle` only; `createCluster`, the nil-style branch, edges (`applyEdgeAttributes`), labels, URLs, and `BuildExpandedGraph`/`buildC1ViewGraph` are untouched — verified by the full-suite golden zero-drift and the per-cluster raw-DOT assertions.
6. **--plain survival**: the assignment sits outside `applyUnitOverrides` (which early-returns under Plain), so the emphasis is semantic by construction (D-03); the CLI `--plain` subtest pins it end-to-end.
7. **Test-helper robustness**: `clusterPenwidthValues` is a stack-walking DOT parser scoped to each subgraph's own `graph [...]` statement; edge `penwidth=1/2` attributes are correctly ignored; every cluster is materialized (empty slice = no attribute), so RED/GREEN states both yield meaningful assertions. Documented limitation: values spanning lines until `];` assume fixture labels contain no `];` sequence — controlled by test fixtures only, never production code.
8. **Concurrency**: all new render-touching tests carry `//nolint:paralleltest` (WASM engine), matching the file convention.

## Verification Status

- `go vet ./internal/graph/ ./internal/render/ ./cmd/c4drill/` — clean
- `gofmt -l` on the six phase files — clean (three unrelated files carry pre-existing drift at HEAD; untouched by design)
- `go test ./... -count=1` — 19/19 packages pass
- Golden drift `git diff --quiet HEAD -- cmd/c4drill/testdata/` — none (zero re-baselines)