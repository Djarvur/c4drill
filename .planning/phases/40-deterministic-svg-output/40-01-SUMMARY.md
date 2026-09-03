---
phase: 40-deterministic-svg-output
plan: 01
subsystem: testing
tags: [determinism, graphviz, svg, tdd, byte-equality, issue-42]

# Dependency graph
requires:
  - phase: 31-edge-identity
    provides: builder-assigned unique Edge.Name (fix 260831-01u) — reused as the stable sort key in Plan 40-02
provides:
  - Sorted-key iteration in populateIncomingLinks — LinksFrom mirror order is a pure function of model content (D-02, REPRO-02 source layer)
  - TestMirrorOrderDeterministicAcrossRuns — validator unit test pinning mirror order across 20 runs
  - TestDeterministicByteIdenticalOutput — D-05 byte-equality regression pin over the issue #42 reproducer for svg/dot/html (REPRO-01, REPRO-03)
affects: [40-02, render-cli, diagram-ci]

# Tech tracking
tech-stack:
  added: []
  patterns: [sorted-key map iteration for deterministic synthesis, id-canonicalized byte-equality regression pins]

key-files:
  created: [internal/render/deterministic_test.go]
  modified: [internal/validator/index.go, internal/validator/index_test.go]

key-decisions:
  - "D-02 fix at the source: populateIncomingLinks collects index keys, sort.Strings, then iterates the sorted slice — loop body unchanged"
  - "Byte pin uses the C2 drill-down view (unitPath \"sys\"): the collapsed C1 view hides permuted mirrors behind link resolution and is deterministic even pre-fix (verified empirically 0/12 vs 12/12 unstable run-pairs)"
  - "Byte pin canonicalizes the root <g id=\"a_graph0_N\"> suffix before comparison: the suffix is a per-process global counter in the go-graphviz WASM engine (renderer-global state), not model content — D-06 id-canonicalization precedent"

patterns-established:
  - "Pattern: byte-equality regression pins canonicalize renderer-global generated ids (a_graph0_N) while pinning every model-derived byte including edge<N> ids"
  - "Pattern: in-process double-render probes to localize nondeterminism (view-shape vs validator vs renderer)"

requirements-completed: [REPRO-01, REPRO-02, REPRO-03]

# Metrics
duration: 17min
completed: 2026-09-03
---

# Phase 40 Plan 01: Deterministic Incoming-Mirror Synthesis Summary

**Sorted-key mirror synthesis in populateIncomingLinks makes repeated renders byte-identical (issue #42): validator unit pin + full-pipeline svg/dot/html byte-equality regression over the issue #42 reproducer**

## Performance

- **Duration:** 17 min
- **Started:** 2026-09-03T22:10:00+03:00
- **Completed:** 2026-09-03T22:27:00+03:00
- **Tasks:** 3
- **Files modified:** 3

## Accomplishments
- `populateIncomingLinks` (internal/validator/index.go) now iterates a sorted slice of index keys instead of ranging the map — LinksFrom mirror order, global edge insertion order, and GraphViz's generated `id="edge<N>"` SVG ids are pure functions of model content (D-02)
- `TestMirrorOrderDeterministicAcrossRuns` pins per-target mirror Peer order across 20 fresh BuildIndex+populateIncomingLinks runs, asserting Mirror-flagged entries only (authored linkFrom excluded)
- `TestDeterministicByteIdenticalOutput` renders the issue #42 reproducer (10 containers, (i+j)%3==0 link predicate, ext→sys.c0 tail) twice in-process through the full pipeline and requires byte-identical svg, dot, and html output

## Task Commits

Each task was committed atomically:

1. **Task 1: RED — validator mirror-order determinism test** - `50e0a28` (test)
2. **Task 2: RED — issue #42 byte-equality regression test** - `8431228` (test)
3. **Task 2 follow-up: canonicalize renderer-global graph id suffix in byte pin** - `4b29eb6` (test, TDD fail-fast investigation)
4. **Task 3: GREEN — sorted-key mirror synthesis** - `fa658a5` (feat)

_Note: TDD tasks may have multiple commits (test → feat → refactor)_

## Files Created/Modified
- `internal/validator/index.go` - sorted-key iteration in populateIncomingLinks (D-02 source fix, issue #42)
- `internal/validator/index_test.go` - converted to internal package validator to host TestMirrorOrderDeterministicAcrossRuns (populateIncomingLinks is unexported); existing BuildIndex tests unchanged
- `internal/render/deterministic_test.go` - D-05 byte-equality regression pin (REPRO-01/03) with inline issue #42 model const

## Decisions Made
- Fixed at the source per D-02 (locked): sorted `sort.Strings` key slice, mirror-append body byte-identical (FindLinkByPeer dedup, full model.Link literal, Mirror: true)
- Byte pin runs the C2 drill-down view (processView's unitPath "sys" branch): empirical probes showed the collapsed C1 view is deterministic pre-fix (0/12 unstable run-pairs) because link resolution lifts internal mirrors away; the drill-down exposes them (12/12 unstable pre-fix)
- Canonicalized `a_graph0_\d+` (root graph id suffix) in the byte pin: it is a per-process global WASM-renderer counter, not model-derived; all model-derived bytes including `id="edge<N>"` are pinned exactly

## Deviations from Plan

### Auto-fixed Issues

**1. [TDD fail-fast rule 1 - unexpected GREEN] Byte pin originally used the top-level C1 view**
- **Found during:** Task 2 (byte-equality regression test)
- **Issue:** the collapsed C1 view resolves internal ci→cj mirrors away before buildEdges, so the svg assertion passed on pre-fix code — not a RED carrier
- **Fix:** switched the pin to the C2 drill-down view (GenerateC2View(m, "sys") + BuildGraphWithPath(v, "sys", ...)), mirroring processView's drill-down branch; this is the view shape that exhibits issue #42
- **Files modified:** internal/render/deterministic_test.go
- **Verification:** pre-fix svg assertion fails with `id="edge17"` vs `id="edge16"` diffs (the documented issue #42 evidence); dot passes pre-fix (documented asymmetry)
- **Committed in:** 8431228

**2. [TDD fail-fast rule 1 - renderer-global id] Root graph id suffix differs across in-process renders**
- **Found during:** Task 3 (post-fix GREEN run)
- **Issue:** after the D-02 fix the only remaining in-process diff was `<g id="a_graph0_0">` vs `a_graph0_1` — a per-process global counter in the go-graphviz WASM emitter, not model nondeterminism; a raw byte pin can never pass across two in-process renders
- **Fix:** canonicalize the `a_graph0_\d+` suffix before comparison (D-06 id-canonicalization precedent); edge ids and every other byte stay pinned; pre-fix the canonicalized assertion still fails (edge id permutation)
- **Files modified:** internal/render/deterministic_test.go
- **Verification:** canonicalized pin is green and stable (`-count=3`), full suite green
- **Committed in:** 4b29eb6

**3. [Rule 1 - test placement] index_test.go switched from package validator_test to package validator**
- **Found during:** Task 1 (validator determinism test)
- **Issue:** populateIncomingLinks is unexported and the plan requires the test in index_test.go (package validator, internal test); the file was external (validator_test)
- **Fix:** converted the file's package to validator and dropped the `validator.` qualifiers; all existing BuildIndex tests kept verbatim
- **Files modified:** internal/validator/index_test.go
- **Verification:** `go vet ./internal/validator/` clean; full validator suite green
- **Committed in:** 50e0a28

---

**Total deviations:** 3 auto-fixed (2 TDD fail-fast investigations, 1 test-package placement)
**Impact on plan:** All deviations are test-side accommodations required by the TDD RED gate and renderer-global state; the product fix is exactly the plan's D-02 change. No scope creep.

## Issues Encountered
- None beyond the deviations above.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Ready for Plan 40-02: defense-in-depth name-sorted tail on buildEdges (D-03) plus the D-06/D-04 no-semantic-change gate; the D-05 byte pin from this plan is the no-op proof for the sort
- REPRO-01/02/03 hold at the source layer; Plan 40-02 adds the second enforcement layer

---
*Phase: 40-deterministic-svg-output*
*Completed: 2026-09-03*
