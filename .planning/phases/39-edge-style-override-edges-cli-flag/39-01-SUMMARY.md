---
phase: 39-edge-style-override-edges-cli-flag
plan: 01
subsystem: cli
tags: [cobra, graphviz, splines, dot, edge-routing, tdd]

requires:
  - phase: 38-hierarchy-wrapping-and-granular-keys (v1.15)
    provides: PLAIN-01 threading pattern, PLAIN-02 plain suppression in builders, KEY-03 switch-matrix harness, canonicalDOT goldens
provides:
  - "--edges <style> persistent CLI flag (straight|spline|square|ortho) with loud enum validation (errInvalidEdges + validateOutputFlags)"
  - "View.EdgesOverride carrier applied after PLAIN-02 zeroing in BOTH BuildGraph and BuildExpandedGraph — beats global AND per-unit model edges AND --plain"
  - "TestEdgesFlagValidation / TestEdgesFlagOverridesModel / TestEdgesSurvivesPlain / TestEdgesFlagOffInvariant pins"
affects: [39-02 matrix E2E, docs surface, release]
tech-stack:
  added: []
  patterns:
    - "explicit-CLI-beats-plain precedence: dedicated default-empty view override field applied post-suppression in every builder site"
    - "enum flag validation extracted to validateOutputFlags + validEdgesStyle helpers (gocognit-safe)"
key-files:
  created: []
  modified:
    - cmd/c4drill/root.go
    - internal/view/view.go
    - internal/graph/builder.go
    - internal/graph/builder_test.go
    - cmd/c4drill/root_test.go
key-decisions:
  - "D-05 mechanism = dedicated View.EdgesOverride field (RESEARCH Option A) applied after plain zeroing — structural flag-off invariance instead of a provenance bool"
  - "Flag validation extracted into validateOutputFlags() covering --format and --edges together — runRoot was pre-existing at gocognit 14/15 and any added branch tripped the lint"
  - "RAW dot asserts use unquoted splines=true/false/ortho forms — the graphviz attribute form verified against actual output, not the planned quoted form"
requirements-completed: [GEDGE-03, GEDGE-04, GEDGE-05, GEDGE-06, GEDGE-08]
duration: 15min
completed: 2026-08-31
---

# Phase 39 Plan 01: --edges Override Semantics Summary

**Invocation-global `--edges <style>` flag with loud enum validation that beats global and per-unit model edges and survives `--plain`, implemented TDD with zero golden churn**

## Performance

- **Duration:** ~15 min
- **Started:** 2026-08-31 (wave 1)
- **Completed:** 2026-08-31
- **Tasks:** 3 (RED / GREEN / pins)
- **Files modified:** 5

## Accomplishments
- `c4drill model.toml --edges straight|spline|square|ortho` overrides the routing style for every generated diagram (C1, all drill-downs, `--expanded`), winning over both `properties.edges` and per-unit `edges` values (D-03/GEDGE-05)
- Invalid values fail loudly before any file I/O — `invalid edges: must be straight, spline, square, or ortho: "diagonal"` — with zero output written (D-02/GEDGE-04)
- `--plain --edges spline` renders `splines=true`: explicit user intent survives author-format suppression, pinned by TestEdgesSurvivesPlain (D-05/GEDGE-06); plain-without-flag suppression unchanged
- Flag-off runs are byte-identical to before — every existing canonicalDOT golden passes untouched (D-04/GEDGE-08); scope.go and converter.go have zero diffs

## Task Commits

Each task was committed atomically (TDD gate order):

1. **Task 1: RED — failing tests** - `0cfb75a` (test)
2. **Task 2: GREEN — implementation** - `01a086c` (feat)
3. **Task 3: pins + invariance** - `200b016` (test)

## Files Created/Modified
- `cmd/c4drill/root.go` — `--edges` flag registration, `errInvalidEdges` sentinel, `validateOutputFlags`/`validEdgesStyle` helpers, threading in both PLAIN-01 blocks
- `internal/view/view.go` — `EdgesOverride string` carrier field (D-03/D-05 doc comment)
- `internal/graph/builder.go` — post-PLAIN-02 override application in BuildGraph AND BuildExpandedGraph
- `internal/graph/builder_test.go` — `TestBuildGraph_EdgeOverride` (6 subtests across both builders)
- `cmd/c4drill/root_test.go` — `TestEdgesFlagValidation`, `TestEdgesFlagOverridesModel`, `TestEdgesSurvivesPlain`, `TestEdgesFlagOffInvariant`

## Decisions Made
- Implemented RESEARCH Option A (dedicated `EdgesOverride` field) — D-04 becomes structural (empty field = today's behavior) rather than merely tested
- Extracted flag validation into `validateOutputFlags()` because `runRoot` sat at gocognit 14/15 pre-change; the inline enum chain tripped the lint at 16
- Test expectations corrected to the verified raw-dot attribute forms (`splines=true`, `splines=false`, `splines=ortho` — unquoted); verified against actual CLI output for all four values before pinning

## Deviations from Plan

None - plan executed exactly as written. (The `validateOutputFlags` extraction and the unquoted-attribute test correction are GREEN-phase iterations within the plan's own TDD iteration contract, not scope deviations.)

## Issues Encountered
- `generateFixtureOutput` callers must pass the fixture name WITH extension (`plain.toml`) per the D-27 extension dispatch — first RED-to-GREEN run used a bare `plain` and hit the extension error
- `gocognit` (repo lint gate) flagged `runRoot` at 16 > 15 after the inline validation branch; resolved by the helper extraction above

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Ready for 39-02 (switch-matrix E2E): the flag, precedence, and plain survival are live end-to-end; RAW dot attribute forms verified (`splines=true|false|ortho`, plain-alone emits no attribute)
- Note for 39-02 matrix assertions: use the unquoted forms verified here

---
*Phase: 39-edge-style-override-edges-cli-flag*
*Completed: 2026-08-31*
