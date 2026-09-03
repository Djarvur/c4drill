---
phase: 40-deterministic-svg-output
plan: 02
subsystem: testing
tags: [determinism, graphviz, stable-sort, defense-in-depth, issue-42]

# Dependency graph
requires:
  - phase: 40-01
    provides: D-02 sorted-key mirror synthesis and the D-05 byte-equality regression pin (TestDeterministicByteIdenticalOutput)
provides:
  - Name-ordered g.Edges tail in buildEdges via slices.SortStableFunc — cgraph insertion order is a pure function of model content even if a future walk regresses into map iteration (D-03, REPRO-02 second layer)
  - TestEdgeOrderNameSortedAndStable — unit pin for non-decreasing names, cross-build identity, and name uniqueness
affects: [render-cli, diagram-ci, future-builder-changes]

# Tech tracking
tech-stack:
  added: []
  patterns: [defense-in-depth stable sort on already-unique content keys]

key-files:
  created: []
  modified: [internal/graph/builder.go, internal/graph/builder_test.go]

key-decisions:
  - "D-03 sort keyed on the already-unique sanitized Edge.Name via slices.SortStableFunc — ordering only; name assignment, markSeen first-wins dedup, and pair styling untouched"
  - "Two tests that pinned the superseded definition-order g.Edges slice contract (TestBuildGraphDefinitionOrder, TestBuildGraphDeterministicOrder subtest 2) were updated to the D-03 name-order contract; node definition-order assertions untouched"

patterns-established:
  - "Pattern: g.Edges is name-ordered (D-03) — new builder tests must assert name order, not walk order"

requirements-completed: [REPRO-02]

# Metrics
duration: 12min
completed: 2026-09-03
---

# Phase 40 Plan 02: Defense-in-Depth Edge-Order Stable Sort Summary

**Stable name-order sort of g.Edges at the buildEdges tail (D-03) guarantees edge insertion order — and GraphViz edge\<N\> ids — stay a pure function of model content at a second layer, with the full D-06/D-04 no-semantic-change gate green**

## Performance

- **Duration:** 12 min
- **Started:** 2026-09-03T22:32:00+03:00
- **Completed:** 2026-09-03T22:44:00+03:00
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- `buildEdges` (internal/graph/builder.go) now ends with `slices.SortStableFunc(edges, ...strings.Compare(a.Name, b.Name))` — a no-op on the deterministic post-D-02 walk, belt-and-suspenders against future walk regressions (D-03)
- `TestEdgeOrderNameSortedAndStable` pins: names non-decreasing across the slice, identical sequences across repeated builds, and name uniqueness
- No-semantic-change gate fully green and recorded below (Self-Check)

## Task Commits

Each task was committed atomically:

1. **Task 1 RED: failing edge-order stability test** - `0c4a380` (test)
2. **Task 1 GREEN: name-ordered edge slice at buildEdges tail (D-03)** - `94e0053` (feat)
3. **Contract alignment: edge-order tests updated to D-03 name ordering** - `5720707` (test, deviation auto-fix)

## Files Created/Modified
- `internal/graph/builder.go` - SortStableFunc tail in buildEdges + D-03 comment (11 insertions, ordering only)
- `internal/graph/builder_test.go` - TestEdgeOrderNameSortedAndStable; two definition-order edge assertions updated to the D-03 name-order contract

## Decisions Made
- Sort key is the builder-assigned unique `Edge.Name` (already sanitized via sanitizeEdgeName, threat T-40-03) with `SortStableFunc` — equal names cannot occur (assignEdgeName counters), so stability is purely preventive
- The D-03 sort intentionally supersedes the internal definition-order g.Edges slice contract; the two tests pinning that contract were updated rather than the sort weakened, because the rendered-semantics gates (canonicalDOT goldens, byte pin) define the real no-semantic-change boundary (D-06)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - plan-premise bug] Two existing tests pinned definition-order g.Edges and failed after the D-03 sort**
- **Found during:** Task 2 (no-semantic-change gate, full suite)
- **Issue:** TestBuildGraphDefinitionOrder and TestBuildGraphDeterministicOrder subtest 2 asserted `g.Edges[0].Source` in definition (walk) order — the exact contract D-03 replaces ("g.Edges leaves buildEdges in ascending Edge.Name order" is a plan must-have). The plan's premise that the sort is "a no-op on today's walk" holds only for rendered semantics, not for the internal slice order
- **Fix:** updated both edge assertions to the D-03 contract (exact `Edge.Name` expectations `alpha_to_gamma_1` < `zulu/zeta_to_*_1`, endpoints re-anchored to matching indices); node definition-order assertions and the cross-call determinism subtests kept verbatim
- **Files modified:** internal/graph/builder_test.go
- **Verification:** full suite `go test ./... -count=1` exits 0; both tests pass with the D-03 assertions
- **Committed in:** 5720707

---

**Total deviations:** 1 auto-fixed (plan-premise bug: internal slice-order contract vs D-03)
**Impact on plan:** Expected consequence of the mandated D-03 contract change; rendered semantics unchanged (gate below). No scope creep.

## Issues Encountered
- None beyond the deviation above.

## Self-Check: PASSED

No-semantic-change gate results (D-06/D-04), all four checks:

| # | Check | Command | Result |
|---|-------|---------|--------|
| 1 | Full suite incl. canonicalDOT goldens (DI-1/COMPAT-02/REF-05 order-insensitive gate) | `go test ./... -count=1` | PASS (exit 0, zero failures) |
| 2 | Plan 40-01 byte pin intact post-sort (sort is a no-op on the deterministic walk) | `go test ./internal/render/ -run TestDeterministicByteIdenticalOutput -count=1` | PASS (svg/dot/html ok) |
| 3 | D-04 audit: no explicit edge ids in non-test Go source | `grep -rn 'id="edge' --include='*.go' internal/ cmd/ \| grep -v _test` | PASS (zero hits) |
| 4 | Golden-shift audit | full-suite green — no golden shifted, no D-06 canonicalization/re-baseline needed | PASS (no churn) |

Additional self-checks: key files exist on disk (`internal/graph/builder.go`, `internal/graph/builder_test.go`); `git log --grep="40-02"` returns the RED test commit (`0c4a380`) before the GREEN feat commit (`94e0053`); `git diff` scope for the GREEN commit is the sort + comment only (assignEdgeName, markSeen, applyCollapsedPairStyle bodies unchanged).

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- REPRO-02 now enforced at two layers (D-02 source ordering + D-03 tail sort); phase 40 goal (byte-identical repeated renders) achieved and pinned by tests
- Ready for phase verification

---
*Phase: 40-deterministic-svg-output*
*Completed: 2026-09-03*
