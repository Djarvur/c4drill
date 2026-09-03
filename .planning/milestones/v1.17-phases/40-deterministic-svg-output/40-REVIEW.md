---
phase: 40
reviewed_files:
  - internal/validator/index.go
  - internal/validator/index_test.go
  - internal/render/deterministic_test.go
  - internal/graph/builder.go
  - internal/graph/builder_test.go
depth: standard
status: clean
critical: 0
warnings: 0
info: 2
reviewed: 2026-09-03
commits: [50e0a28, 8431228, 4b29eb6, fa658a5, 0c4a380, 94e0053, 5720707]
---

# Phase 40 — Code Review

Scope: files changed by Phase 40 per 40-01-SUMMARY.md and 40-02-SUMMARY.md (1 created, 4 modified; commits 50e0a28..5720707). Reviewer note: executed inline (standard depth) — runtime provides no reviewer subagent.

## internal/validator/index.go (modified, +15/−1)

- D-02 fix verified correct: `populateIncomingLinks` now collects index keys into a pre-sized `[]string`, `sort.Strings`, and iterates the sorted slice with `index[sourcePath]` lookup — semantically identical to the removed `range` (same `*UnitInfo` pointers), differing only in deterministic order. Loop body (FindLinkByPeer dedup check, full `model.Link` literal with `Mirror: true`) byte-identical to pre-fix.
- Dedup outcome is order-independent: the "already exists" check tests the target's `LinksFrom` for the specific `sourcePath`, and distinct sources never collide on that key, so insertion order cannot change which mirrors survive.
- No new allocation concerns (single O(n) slice, O(n log n) sort per Validate call, negligible at model scale). `sort` is stdlib — no dependency change (T-40-SC honored).

## internal/graph/builder.go (modified, +11/−0)

- D-03 tail sort verified: `slices.SortStableFunc` on the already-materialized edge slice keyed by `strings.Compare(a.Name, b.Name)`. `Edge.Name` is unique per build (assignEdgeName per-pair counters over one shared `nameCounters` map), so the stable variant is strictly preventive and the comparator never sees equal keys.
- Ordering-only diff: `assignEdgeName`, `markSeen` (first-wins), and `applyCollapsedPairStyle` bodies untouched (diff-verified). Sort runs before `buildLegend`, so legend rows see the same final slice consumers use — consistent.
- Threat check: sort key is the sanitized `Edge.Name` (sanitizeEdgeName precedent, T-Q1-01) — author-controlled paths never flow raw (T-40-03 mitigated). O(n log n) on a materialized slice — negligible (T-40-04 accepted).

## internal/render/deterministic_test.go (created, +175)

- D-05 pin structure sound: inline deterministic model const (issue #42 generator translation, (i+j)%3==0 predicate and ext→sys.c0 tail verified against the research), full CLI-equivalent pipeline (parse → peer.Resolve → validator.Validate → C2 view → BuildGraphWithPath → render.Render) run twice per format for svg/dot/html.
- Canonicalization is correctly scoped: `a_graph0_\d+` normalizes only the renderer-global root graph id suffix (per-process WASM counter, not model content); every model-derived byte — including all `id="edge<N>"` group ids — is asserted exactly. Failure path logs the first differing byte index for diagnosis.
- Test hygiene: no `t.Parallel()` (WASM engine convention), no testdata files, no randomness, no new dependencies.

## internal/validator/index_test.go / internal/graph/builder_test.go (modified)

- index_test.go package switch to internal `validator` is the only way to host the unexported-function test in the plan-mandated file; existing BuildIndex tests kept verbatim (qualifiers only). TestMirrorOrderDeterministicAcrossRuns rebuilds fresh unit trees per run (mirrors would otherwise accumulate), filters Mirror-flagged entries only, and does 20 comparison runs — sound RED detection for map-iteration permutation.
- TestEdgeOrderNameSortedAndStable asserts all three contract points (non-decreasing names, cross-build identity, uniqueness). The two updated definition-order tests correctly encode the D-03 name-order contract with exact `Edge.Name` expectations while keeping node-order and cross-call determinism assertions.

## Findings

| # | Severity | File | Line | Finding | Disposition |
|---|----------|------|------|---------|-------------|
| 1 | Info | internal/render/deterministic_test.go | 91 | Canonicalization matches `a_graph0_\d+` anywhere in output bytes; a model whose rendered text contained that literal would also be normalized | Accept (renderer-namespace literal, not a realistic author identifier; documented in-test) |
| 2 | Info | internal/validator/index.go | 4 | Uses `sort.Strings` while sibling code leans on the newer `slices` package — consistency nit only, both stdlib | Accept (no action) |

**Verdict:** clean — no critical or warning findings. The product change is 26 added lines across two files (sorted-key iteration + stable tail sort), both ordering-only, with the full suite, canonicalDOT goldens, and the byte-equality pin green.
