---
phase: 40-deterministic-svg-output
status: passed
score: 8/8
verified: 2026-09-03
method: inline goal-backward verification (no verifier subagent available in runtime)
requirements_covered: [REPRO-01, REPRO-02, REPRO-03]
human_verification_count: 0
gaps: []
---

# Phase 40: Deterministic SVG Output — Verification

**Goal:** Rendering the same model twice produces byte-identical output — generated edge ids derive deterministically from model content, so diagrams can be diffed, committed, and cached reliably (issue #42).

## Observable Truths (roadmap SCs + PLAN must_haves merged)

| # | Truth | Evidence | Status |
|---|-------|----------|--------|
| 1 | Rendering the issue #42 reproducer twice in-process produces byte-identical SVG including generated `id="edge<N>"` group ids | `go test ./internal/render/ -run TestDeterministicByteIdenticalOutput -count=1` → ok (svg subtest green; RED pre-fix with `id="edge17"` vs `id="edge16"` evidence) | VERIFIED |
| 2 | Byte-equality holds for dot, svg, and html alike | same test: all three format subtests green (`-count=3` stability re-run also ok) | VERIFIED |
| 3 | Validator-synthesized LinksFrom mirror order is identical across repeated pipeline runs | `go test ./internal/validator/ -run TestMirrorOrder -count=1` → ok (20 fresh BuildIndex+populateIncomingLinks runs, identical mirror Peer sequence) | VERIFIED |
| 4 | No DOT-output change: no explicit edge ids emitted into DOT (D-04) | `grep -rn 'id="edge' --include='*.go' internal/ cmd/ \| grep -v _test` → zero hits | VERIFIED |
| 5 | g.Edges leaves buildEdges in ascending Edge.Name order (D-03, insertion order = pure function of model content) | `go test ./internal/graph/ -run TestEdgeOrderNameSortedAndStable -count=1` → ok (non-decreasing names, cross-build identity, uniqueness); `slices.SortStableFunc` tail present in builder.go | VERIFIED |
| 6 | The D-03 sort is a no-op on the deterministic walk — Plan 40-01 byte pin passes post-sort | `go test ./internal/render/ -run TestDeterministicByteIdenticalOutput -count=1` after 94e0053 → ok | VERIFIED |
| 7 | No rendered-semantics change: canonicalDOT goldens (DI-1/COMPAT-02/REF-05) and full suite green (D-06) | `go test ./... -count=1` → 19/19 packages ok, zero failures, zero golden churn | VERIFIED |
| 8 | TDD gate compliance: RED test commits precede GREEN feat commits in both plans | git log: 50e0a28 + 8431228 (test RED) before fa658a5 (feat); 0c4a380 (test RED) before 94e0053 (feat) | VERIFIED |

**Score: 8/8 verified**

## Artifacts

| Artifact | Provides | Contains-check | Status |
|----------|----------|----------------|--------|
| internal/validator/index.go | Sorted-key iteration in populateIncomingLinks (D-02) | `sort.Strings` (line 65), no bare `range index` driving mirror appends | PASS |
| internal/render/deterministic_test.go | Byte-equality regression pin from the issue #42 reproducer (D-05) | 175 lines (min 60), `TestDeterministicByteIdenticalOutput`, inline model const | PASS |
| internal/validator/index_test.go | Mirror-order determinism unit pin | `TestMirrorOrderDeterministicAcrossRuns` | PASS |
| internal/graph/builder.go | Stable name-order tail sort (D-03) | `SortStableFunc` + `strings.Compare(a.Name, b.Name)` final statement before return | PASS |
| internal/graph/builder_test.go | Edge-order unit pin | `TestEdgeOrderNameSortedAndStable` | PASS |

## Key Links (wiring)

| From | To | Via | Pattern check | Status |
|------|----|----|---------------|--------|
| internal/validator/index.go (populateIncomingLinks) | internal/graph/builder.go (buildEdges inLinks) | LinksFrom slice order feeds edge insertion order | buildEdges reads `entry.Unit.LinksFrom` (now deterministic) | WIRED |
| internal/render/deterministic_test.go | render.Render / graph.BuildGraphWithPath | full pipeline invoked twice per assertion | `BuildGraphWithPath(` + `render.Render(` present | WIRED |
| internal/graph/builder.go (buildEdges tail sort) | internal/render/converter.go (createEdges) | g.Edges iterated verbatim for cgraph insertion | createEdges loops `for _, edge := range edges` in order | WIRED |
| internal/graph/builder_test.go | graph.BuildGraph | builds views, asserts g.Edges order | `graph.BuildGraph(` present | WIRED |

## Requirements Coverage

| Requirement | Definition | Evidence | Status |
|-------------|-----------|----------|--------|
| REPRO-01 | Same model rendered twice → byte-identical SVG incl. `id="edge<N>"` | TestDeterministicByteIdenticalOutput/svg (issue #42 reproducer, in-process double render) | SATISFIED |
| REPRO-02 | Edge ids deterministic from model content, never Go map iteration | Two layers: sorted-key mirror synthesis (D-02, fa658a5) + name-sorted edge tail (D-03, 94e0053); pinned by TestMirrorOrderDeterministicAcrossRuns + TestEdgeOrderNameSortedAndStable | SATISFIED |
| REPRO-03 | Determinism holds for every supported format (dot, svg, html) | TestDeterministicByteIdenticalOutput asserts all three formats | SATISFIED |

REQUIREMENTS.md traceability: REPRO-01/02/03 marked complete (phase 40 rows updated).

## Deviations Reviewed (from SUMMARYs)

1. Byte pin uses the C2 drill-down view instead of collapsed C1 (TDD fail-fast: C1 is deterministic pre-fix — verified 0/12 vs 12/12 unstable run-pairs) — justified, documented in 40-01-SUMMARY.md.
2. Byte pin canonicalizes the renderer-global `a_graph0_N` root id suffix (per-process WASM counter, not model content; D-06 precedent) — justified, documented.
3. index_test.go converted to internal package validator — required to host the unexported-function test in the plan-mandated file.
4. Two graph contract tests updated to the D-03 name-order slice contract (40-02 deviation) — correct: rendered-semantics gates stayed green.

All deviations are test-side accommodations; the product surface is exactly the planned D-02/D-03 ordering changes.

## Human Verification

None required — the phase goal (byte-identical repeated renders) is fully automatable and pinned by three regression tests. CI reproducibility follows from in-process determinism plus stable sorted walks.

## Verdict

**PASSED — Phase 40 goal achieved.** Byte-identical repeated output enforced at two layers (validator mirror synthesis + buildEdges tail sort), zero golden churn, full suite green.
