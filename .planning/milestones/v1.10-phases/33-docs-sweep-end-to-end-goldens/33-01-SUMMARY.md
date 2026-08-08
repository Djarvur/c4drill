---
phase: 33-docs-sweep-end-to-end-goldens
plan: 01
subsystem: testing
tags: [gotest, refactor, dot-canonical, golden-helper]

# Dependency graph
requires:
  - phase: 26-graph-builder
    provides: original canonicalDOT prior art in builder_test.go (WR-01/WR-02 hardened)
provides:
  - "Reusable canonical.Canonical(t, dot) order-insensitive DOT comparator at internal/testutil/canonical/ (DI-1, D-18)"
  - "Importable from any _test.go file in the repo (Plan 04's XC-05/XC-01 E2E goldens + future milestone goldens)"
affects: [33-04, future-milestone-goldens]

# Tech tracking
tech-stack:
  added: []
  patterns: ["Extract-and-Delegate for test helpers trapped in _test.go files (cross-package import boundary forced internal/testutil/ location)"]

key-files:
  created:
    - internal/testutil/canonical/canonical.go
    - internal/testutil/canonical/canonical_test.go
  modified:
    - internal/graph/builder_test.go

key-decisions:
  - "D-18 realized: canonical.Canonical(t, dot) at internal/testutil/canonical/ — the only location Go's import rules allow cross-package _test.go sharing under the internal/ prefix"
  - "4 regression tests moved, not 3 — plan's line range (1249-1591) missed TestCanonicalDOTHTMLLabelDoesNotTruncate (WR-02, extends to line 1592); leaving any behind would not compile (references removed local canonicalDOT)"
  - "Failure path switched from require.True(t, ok, ...) to t.Fatalf(...) inside Canonical — keeps the package stdlib-only (no testify dep) per threat model T-33-01-SC; behavior-equivalent for test callers"

patterns-established:
  - "internal/testutil/<name>/ as the home for reusable test helpers that must cross package boundaries (Go _test.go import exclusion workaround)"

requirements-completed: []

# Metrics
duration: 8 min
completed: 2026-08-08
---

# Phase 33 Plan 01: Reusable canonicalDOT helper (D-18) Summary

**Extracted the order-insensitive canonicalDOT comparator (DI-1) from internal/graph/builder_test.go into a reusable internal/testutil/canonical/ package, importable from any _test.go file — unblocking Plan 04's cross-package E2E goldens.**

## Performance

- **Duration:** ~8 min
- **Started:** 2026-08-08 (inline execution)
- **Completed:** 2026-08-08
- **Tasks:** 2 (1 implementation + 1 verification-only)
- **Files modified:** 3 (2 created, 1 modified)

## Accomplishments
- Reusable `canonical.Canonical(t, dot)` ships at `internal/testutil/canonical/` — the D-18 foundational deliverable every Wave 2 golden depends on
- Helper is importable from ANY `_test.go` file (lives in a non-`_test.go` package under `internal/` — Go's `_test.go` import exclusion no longer traps it)
- DI-1 contract preserved verbatim: parse DOT, strip layout geometry (bb/pos/lp/lheight/lwidth/height/width), sort statements + attributes recursively, order-insensitive comparison
- 4 WR-01/WR-02 regression tests moved and green in new home
- 2 existing goldens (COMPAT-02 TestBuildExpandedGraphBaselineDOT, REF-05 TestReference_BackwardCompat) still green via import — zero backward-compat regression
- `internal/graph/builder_test.go` no longer carries the local copy (single source of truth)

## Task Commits

Each task was committed atomically:

1. **Task 1: Extract canonicalDOT into internal/testutil/canonical/ package** - `1ef0db9` (refactor)
2. **Task 2: Full-suite regression gate after extraction** - (verification only — suite already green from Task 1; no commit needed per plan)

## Files Created/Modified
- `internal/testutil/canonical/canonical.go` - package `canonical`; exported `Canonical(t, dot)` + unexported `dotStatement` struct + 12 parse/serialize helpers (verbatim bodies, DI-1/D-02/WR-01/WR-02 provenance comments preserved)
- `internal/testutil/canonical/canonical_test.go` - the 4 WR-01/WR-02 regression tests moved verbatim (local `canonicalDOT(t, ...)` calls switched to `Canonical(t, ...)`)
- `internal/graph/builder_test.go` - import added; 2 golden call sites switched to `canonical.Canonical`; local block (1249-1593) deleted; unused `sort` import dropped

## Decisions Made
- Moved **4** regression tests, not the **3** named in the plan. The plan's `<interfaces>` block cited lines "1529-1591" but the actual block runs to line 1592 (the plan's prose also said "approximately lines 1249-1591"). `TestCanonicalDOTHTMLLabelDoesNotTruncate` (lines 1582-1592) is a WR-02 HTML-label regression guard and is part of the same canonical contract — leaving it behind would fail to compile (it calls the removed local `canonicalDOT`). All 4 moved together.
- Switched `Canonical`'s parse-failure path from `require.True(t, ok, "...")` to `t.Fatalf("...")`. The original lived in `graph_test` which imports testify; the new package aims to be stdlib-only (threat model T-33-01-SC: zero new deps). `t.Fatalf` is behavior-equivalent for callers (fails the test with a message). testify remains a test-scope dep of `canonical_test.go` only.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Plan understated the regression-test count (3 vs 4)**
- **Found during:** Task 1 (extraction)
- **Issue:** Plan's `<interfaces>` and `<tasks>` named 3 tests (`TestCanonicalDOTPreservesLastAttribute`, `TestCanonicalDOTFinalAttributeDriftDetected`, `TestCanonicalDOTQuotedValuesDoNotTruncate`) but the cited line range 1249-1591 actually contains 4 — `TestCanonicalDOTHTMLLabelDoesNotTruncate` (WR-02, lines 1582-1592) extends just past 1591.
- **Fix:** Moved all 4 tests verbatim to `canonical_test.go`. Leaving any behind would not compile (the orphaned test would reference the deleted local `canonicalDOT`).
- **Files modified:** internal/testutil/canonical/canonical_test.go
- **Verification:** `grep -c '^func Test' internal/testutil/canonical/canonical_test.go` returns 4; `go test ./internal/testutil/canonical/` green.
- **Committed in:** 1ef0db9 (Task 1 commit)

**2. [Rule 3 - Blocking] Dropped now-unused `sort` import in builder_test.go**
- **Found during:** Task 1 (extraction) — `go vet` flagged it
- **Issue:** `sort` was imported in builder_test.go only for the 12 helpers being moved out. After deletion, vet reported "sort imported and not used".
- **Fix:** Removed `sort` from the import block. `strings` stays (still used 4× elsewhere).
- **Files modified:** internal/graph/builder_test.go
- **Verification:** `go vet ./internal/graph/...` clean; `grep -c '\bsort\.' internal/graph/builder_test.go` returns 0.
- **Committed in:** 1ef0db9 (Task 1 commit)

**3. [Rule 3 - Blocking] gofmt fixed a double blank line left at the deletion boundary**
- **Found during:** Task 1 (extraction) — `gofmt -l` flagged it
- **Issue:** Deleting lines 1250-1593 left two consecutive blank lines between the end of `TestBuildExpandedGraphBaselineDOT` and the `//nolint:funlen` of `TestBuildGraphDeterministicOrder`.
- **Fix:** `gofmt -w internal/graph/builder_test.go` collapsed the doublespace.
- **Files modified:** internal/graph/builder_test.go
- **Verification:** `gofmt -l internal/graph/builder_test.go` returns empty.
- **Committed in:** 1ef0db9 (Task 1 commit)

---

**Total deviations:** 3 auto-fixed (1 bug in plan's test count, 2 blocking cleanup from extraction mechanics)
**Impact on plan:** All auto-fixes necessary for a compiling, vet-clean, gofmt-clean extraction. Zero scope creep — the canonicalization logic, helper names, and DI-1 contract are byte-identical to the source. The 4th test move was load-bearing (plan's line range was slightly off, not a design change).

## Issues Encountered
None - extraction was mechanical; full suite green on first complete run after the 3 auto-fixes.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Plan 04 (Wave 2) can now import `github.com/Djarvur/c4drill/internal/testutil/canonical` from `cmd/c4drill/root_test.go` for the XC-05/XC-01 E2E goldens.
- Future milestone golden tests reuse `canonical.Canonical` instead of re-implementing the DI-1 contract.

---
*Phase: 33-docs-sweep-end-to-end-goldens*
*Completed: 2026-08-08*

## Self-Check: PASSED

- [x] `internal/testutil/canonical/canonical.go` exists with `package canonical` and exported `func Canonical(t *testing.T, dot string) string`
- [x] `internal/testutil/canonical/canonical_test.go` exists with the regression tests (4 moved, plan named 3 — see deviation 1)
- [x] `grep -c 'func canonicalDOT\|func parseDOTStatements\|func serializeDOTStatements\|func isGeometryAttr\|type dotStatement' internal/graph/builder_test.go` returns 0
- [x] `canonical.Canonical` appears at the 2 golden call sites in builder_test.go (3 lines, 6 occurrences — both sites call it twice for expected+actual)
- [x] `go test ./internal/testutil/canonical/ -x` exits 0 — all 4 regression tests pass in the new home
- [x] `go test ./internal/graph/ -run 'TestBuildExpandedGraphBaselineDOT|TestReference_BackwardCompat'` exits 0 — COMPAT-02 + REF-05 backward-compat preserved
- [x] `go test ./internal/graph/` exits 0 — no other graph test regressed
- [x] `go vet ./...` exits 0
- [x] `go build ./...` exits 0
- [x] `go test ./...` exits 0 — full suite green, no cross-package regression
- [x] Commit message starts with `refactor(33-01):`
