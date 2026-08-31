---
phase: 39-edge-style-override-edges-cli-flag
plan: 02
subsystem: testing
tags: [go-test, e2e, matrix, graphviz, splines, dot, testdata]

requires:
  - phase: 39-01 (--edges override semantics)
    provides: "--edges flag end-to-end, View.EdgesOverride precedence, RAW dot splines attribute forms"
  - phase: 38 (KEY-03 composition matrix)
    provides: generateFixtureOutput harness and matrix organization precedent
provides:
  - "TestEdgesComposition — the GEDGE-07 matrix: --edges × generation × --plain asserting splines in RAW dot (~86 cells)"
  - "edges_override.toml golden-free fixture carrying global AND per-unit edges precedence layers"
affects: [39-03 docs, release]
tech-stack:
  added: []
  patterns:
    - "flat table-driven matrix cells (edgesMatrixCases) instead of nested loops — keeps gocognit/wsl_v5 clean"
key-files:
  created:
    - cmd/c4drill/testdata/edges_override.toml
  modified:
    - cmd/c4drill/root_test.go
key-decisions:
  - "Per-unit edges layer made observable by giving app.api subunits (its own C3 drill-down resolves cmp.Or(unit, properties)) — links retargeted to the leaf app.api.jobs per the expandable-unit VAL rule"
  - "Fixture kept deliberately golden-free (no committed golden references it) — GEDGE-08 invariance untouched"
  - "Matrix rewritten as flat edgesMatrixCases() table after gocognit flagged the nested-loop form at 28 (limit 15); also resolves wsl_v5 nits"
requirements-completed: [GEDGE-05, GEDGE-07]
duration: 12min
completed: 2026-08-31
---

# Phase 39 Plan 02: --edges Switch-Matrix E2E Summary

**GEDGE-07 matrix proving --edges × generation × --plain via RAW-dot splines assertions over a golden-free two-layer precedence fixture**

## Performance

- **Duration:** ~12 min
- **Started:** 2026-08-31 (wave 2)
- **Completed:** 2026-08-31
- **Tasks:** 2
- **Files modified:** 2 (1 created, 1 modified)

## Accomplishments
- Every `--edges` style (spline/straight/ortho/square) asserted across root, C2 drill-down, C3 per-unit drill-down, and `--expanded` — alone AND composed with `--plain` — via exact `splines=` attributes in RAW dot (GEDGE-07)
- D-03 pinned end-to-end: `--edges spline` beats the per-unit `edges = "ortho"` override on that unit's own drill-down
- D-04 pinned: flag-off runs resolve global (`splines=true` at C1/app) and per-unit (`splines=ortho` at api) layers exactly as before; `--expanded` honors the global value
- `square` renders as the documented ortho alias through the flag path (GEDGE-02)
- Full suite green with zero golden churn; golangci-lint 0 issues

## Task Commits

1. **Task 1: golden-safe per-unit fixture** - `a533c97` (test)
2. **Task 2: TestEdgesComposition matrix** - `0e046d3` (test, includes the lint-clean flat-table refactor)

## Files Created/Modified
- `cmd/c4drill/testdata/edges_override.toml` — global `edges = "spline"` + per-unit `edges = "ortho"` on the expanded app.api unit; links target the leaf app.api.jobs per the expandable-unit VAL rule
- `cmd/c4drill/root_test.go` — `edgesMatrixCases()` + `TestEdgesComposition` flat matrix (~86 leaf cells)

## Decisions Made
- Gave app.api subunits so its own drill-down exposes the per-unit layer in RAW dot (links retargeted to the leaf — expandable units cannot be linked directly)
- Flat table-driven matrix (precomputed cells) instead of nested subtest loops — the nested form tripped gocognit (28 > 15) and wsl_v5; the flat form keeps subtest nesting via "/" labels

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Expandable units cannot be linked directly**
- **Found during:** Task 1 (fixture authoring)
- **Issue:** First fixture linked admin/worker to app.api, which needed subunits to expose its per-unit value — the validator rejects links to expandable units
- **Fix:** Retargeted both links to the new leaf app.api.jobs
- **Files modified:** cmd/c4drill/testdata/edges_override.toml
- **Verification:** CLI renders cleanly; validation passes
- **Committed in:** Task 1 commit

---

**Total deviations:** 1 auto-fixed (missing critical)
**Impact on plan:** Necessary for validator compliance; no scope change.

## Issues Encountered
- golangci-lint (gocognit, wsl_v5) flagged the first nested-loop matrix form; resolved by the flat edgesMatrixCases table (same cells, same assertions, verified green after refactor)

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Ready for 39-03 (docs + release): behavior complete and matrix-pinned; README/SKILL docs can state the precedence chain and the --plain delta as verified facts

---
*Phase: 39-edge-style-override-edges-cli-flag*
*Completed: 2026-08-31*
