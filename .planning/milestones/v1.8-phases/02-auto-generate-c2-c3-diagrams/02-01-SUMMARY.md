---
phase: 02-auto-generate-c2-c3-diagrams
plan: 01
subsystem: diagrams
tags: [c4, c2, c3, golang, graph, gographviz, sub-diagrams, expansion, boundary-nodes]

# Dependency graph
requires:
  - phase: 01-fix-c1-view-scoping
    provides: C1 view scoping fix, resolved-link synthesis, penwidth machinery (WR-01), buildEdges edge work
provides:
  - D-07 guard: expanded-but-empty top-level unit renders as plain node in C1 (one-condition guard at builder.go:68)
  - Regression tests locking D-01/D-02/D-03 (box + containerBox sub-diagram files at unit-key paths), D-04 second half (per-unit expanded container cluster in C2), D-05 (OR expansion), D-06 (silent ignore), D-08 (personExternal actor boundary node in C2)
affects: [verify-work, phase 3 (audit), any future C1 expansion semantics work]

# Tech tracking
tech-stack:
  added: [none — stdlib + testify v1.11.1 only]
  patterns:
    - "Fixture constant blocks keep test literals below the goconst threshold (authComponentName, webUserPath/webUserName reuse)"
    - "Graph-layer guard placement for expansion semantics: view layer stays read-only (scope.go untouched)"

key-files:
  created: []
  modified:
    - internal/graph/builder.go — D-07 one-condition guard in C1 branch (line 68)
    - internal/graph/builder_test.go — TestBuildGraphExpandedEmptyUnitRendersPlainNode, TestBuildGraphC2ExpandedContainerRendersCluster
    - internal/view/scope_test.go — TestGenerateC1View_IsExpandedOrSemantics, TestGenerateC1View_SilentlyIgnoresUnknownExpandedEntries, TestGenerateC2View_ActorBoundaryFromSubunitLinks
    - cmd/c4drill/root_test.go — TestFullPipeline_BoxWithSubunitsGeneratesSubDiagram

key-decisions:
  - "D-07 guard lives in the graph layer (BuildGraph C1 branch), not in isExpandedInC1 — scope.go stays read-only (Phase 1 WR-01 constraint)"
  - "Test fixtures follow existing analog shapes exactly (TestBuildGraphClusters, TestFullPipeline_NestedWithExpanded, TestGenerateC2View_ExternalBoundaryFromSubunitLinks)"
  - "Box pipeline fixture must include inter-unit links — validator ValidateOrphanUnits rejects link-less leaf units"

patterns-established:
  - "D-07 test shape: mirror TestBuildGraphClusters and invert the expectation (0 clusters, 1 plain node)"
  - "C2/C3 boundary-cluster regression guard: TestBuildGraphC2ExpandedContainerRendersCluster fails if builder.go:45 branch is altered"

requirements-completed: [VIEW-03, VIEW-04, VIEW-05]

# Metrics
duration: 26min
completed: 2026-08-06
---

# Phase 02 Plan 01: C2/C3 Sub-diagram Refinement Summary

**D-07 graph-layer guard (expanded-but-empty units render as plain C1 nodes, no empty cluster box) plus 5 new regression tests locking VIEW-03/04/05 semantics — box and containerBox sub-diagram files at unit-key paths (D-01/D-02/D-03), per-unit expanded container clusters inside C2 boundary clusters (D-04), OR expansion precedence (D-05), silent ignore of unknown expanded entries (D-06), and personExternal actor boundary nodes in C2 (D-08)**

## Performance

- **Duration:** 26 min
- **Started:** 2026-08-06T13:44:00Z
- **Completed:** 2026-08-06T14:10:00Z
- **Tasks:** 3 (Task 1 TDD: RED + GREEN; Task 2: 3 tests; Task 3: 2 tests)
- **Files modified:** 4 (1 production, 3 test)

## Accomplishments

- D-07 implemented at its single agreed point: `if entry.IsExpanded` → `if entry.IsExpanded && len(entry.Unit.Subunits) > 0` in BuildGraph's C1 branch (builder.go:68). Expanded-but-empty top-level units now fall through to `buildNode` (plain collapsed node, no "[+]" indicator — `HasSubunits` is false).
- TDD RED/GREEN gate honored: `test(02-01)` commit `eb77ae3` failed against current code (1 empty cluster produced), `feat(02-01)` commit `ad39e93` made it green.
- VIEW-03/04 file layout locked end-to-end: `TestFullPipeline_BoxWithSubunitsGeneratesSubDiagram` runs the full CLI pipeline on an inline TOML with a box and a containerBox, asserting `boxtest.svg`, `boxtest/boxname.svg`, and `boxtest/system/cbox.svg` (unit-key dotted-path naming, auto-detect on subunits alone, no per-unit expanded needed).
- VIEW-05 semantics locked at view layer: OR precedence (properties.expanded ∪ per-unit self-reference) and silent ignore of non-matching entries.
- D-04 second half locked at build layer: per-unit expanded container renders `cluster_app.api` with its 2 components inside the C2 boundary cluster — this test doubles as a guard that the C2/C3 branch (builder.go:45) stays untouched.
- D-08 locked: linked `personExternal` actor appears as an external boundary node in C2 views (uses TypePersonExternal — IsExternalType returns false for plain TypePerson).
- Zero production surface beyond the one-condition guard; internal/view/scope.go diff is empty; Phase 1 WR-01 penwidth tests (`TestBuildEdgesPenwidth*`) green.

## Task Commits

Each task was committed atomically:

1. **Task 1 (RED): D-07 failing test** - `eb77ae3` (test)
2. **Task 1 (GREEN): D-07 guard** - `ad39e93` (feat)
3. **Task 2: C1 OR/silent-ignore + C2 cluster tests** - `c6c131a` (test)
4. **Task 3: box sub-diagram + actor boundary tests** - `ad34af4` (test)
5. **Lint cleanup of new fixtures** - `9e6b1f0` (refactor)

**Plan metadata:** `02-01-SUMMARY.md` (docs: complete plan) — committed with STATE/ROADMAP/REQUIREMENTS updates.

## Files Created/Modified

- `internal/graph/builder.go` - D-07 guard: `entry.IsExpanded && len(entry.Unit.Subunits) > 0` in the C1 branch (:68); C2/C3 boundary-cluster branch (:30-55) and buildEdges untouched
- `internal/graph/builder_test.go` - `TestBuildGraphExpandedEmptyUnitRendersPlainNode` (D-07, "// D-07:" traceability comment), `TestBuildGraphC2ExpandedContainerRendersCluster` (D-04), `authComponentName` fixture constant
- `internal/view/scope_test.go` - `TestGenerateC1View_IsExpandedOrSemantics` (D-05), `TestGenerateC1View_SilentlyIgnoresUnknownExpandedEntries` (D-06), `TestGenerateC2View_ActorBoundaryFromSubunitLinks` (D-08); reuses existing `webUserPath`/`webUserName` constants
- `cmd/c4drill/root_test.go` - `TestFullPipeline_BoxWithSubunitsGeneratesSubDiagram` (D-01/D-02/D-03), no t.Parallel with `//nolint:paralleltest` directive

## Decisions Made

- D-07 guard placed in the graph layer (BuildGraph C1 branch) per RESEARCH Pitfall 1 and 02-PATTERNS.md — `isExpandedInC1` (scope.go) stays read-only; view-layer IsExpanded semantics asserted by `TestGenerateC1View_IsExpandedWhenUnitExpandsSelf` remain valid.
- All new tests clone existing analog shapes (TestBuildGraphClusters, TestFullPipeline_NestedWithExpanded, TestGenerateC2View_ExternalBoundaryFromSubunitLinks) per pattern map.
- Lint fixes applied as a separate `refactor(02-01)` commit rather than amending task commits — keeps the TDD RED/GREEN gate commits pure.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Box pipeline fixture missing links — validator rejects link-less leaf units**
- **Found during:** Task 3 (TestFullPipeline_BoxWithSubunitsGeneratesSubDiagram)
- **Issue:** The plan's inline TOML fixture omitted all `[[...link]]` entries; the validator's `ValidateOrphanUnits` rule rejects every leaf unit with "has no incoming or outgoing links", so `cmd.Execute()` returned `validation failed` and the test could not pass.
- **Fix:** Added `[[boxname.child.link]] peer = "system.cbox.comp"` and the mirror `[[system.cbox.comp.linkFrom]] peer = "boxname.child"` — the exact link/linkFrom style of the existing TestFullPipeline_NestedWithExpanded fixture (:334-390). Nesting rules were verified untouched (box→system is C1-in-C1, system→containerBox is C2-in-system, containerBox→container is C2-in-C2 — all valid per rules.go).
- **Files modified:** cmd/c4drill/root_test.go
- **Verification:** `go test ./cmd/c4drill/ -run 'TestFullPipeline_Box' -count=1` passes; all three file assertions green.
- **Committed in:** ad34af4 (Task 3 commit)

**2. [Rule 1 - Bug] 5 new golangci-lint issues on added test lines**
- **Found during:** Wave gate (`golangci-lint run --new-from-rev=<baseline> ./...`)
- **Issue:** goconst x3 (`"auth"`, `"webUser"`, `"Web User"` fixture literals above threshold), lll x1 (124-char assert line), testifylint x1 (`require.Len(..., 0)` → `require.Empty`).
- **Fix:** Reused existing `webUserPath`/`webUserName` constants in the D-08 test, added `authComponentName` fixture constant in builder_test.go (per scope_test.go convention), wrapped the C3 file assertion, switched to `require.Empty`.
- **Files modified:** internal/view/scope_test.go, internal/graph/builder_test.go, cmd/c4drill/root_test.go
- **Verification:** `golangci-lint run --new-from-rev=5003e3a ./...` → "0 issues."
- **Committed in:** 9e6b1f0 (refactor commit)

---

**Total deviations:** 2 auto-fixed (2 Rule 1 - Bug)
**Impact on plan:** Both fixes confined to test fixtures/cleanup; the single production change (D-07 guard) matches the plan exactly. No scope creep.

## Issues Encountered

- None beyond the two documented deviations. No auth gates, no permission/file-lock issues, no package installs.

## TDD Gate Compliance

Git log confirms the required order for Task 1: `eb77ae3 test(02-01): add failing test for expanded-empty unit as plain node` (RED — failed with "should have 0 item(s), but has 1") precedes `ad39e93 feat(02-01): render expanded-empty units as plain nodes` (GREEN). No REFACTOR commit was needed for the guard itself (one-condition change); the later `9e6b1f0 refactor(02-01)` commit covers lint cleanup of test fixtures only. Gate sequence: PASSED.

## Verification Results

- `go test -v -race ./...` — green, 352 tests across 7 packages
- `golangci-lint run --new-from-rev=5003e3a ./...` — 0 issues
- WR-01 regression spot check: `go test ./internal/graph/ -run 'TestBuildEdgesPenwidth' -count=1` — green (C2/C3 branch and Phase 1 edge work untouched)
- Scope guard: `git diff internal/view/scope.go` — empty (0 lines)
- builder.go:45 still plain `if entry.IsExpanded` (C2/C3 branch untouched); builder.go:68 contains `entry.IsExpanded && len(entry.Unit.Subunits) > 0`

## Known Stubs

None — all tests assert concrete behavior; no placeholder data or unwired components introduced.

## Threat Flags

None — no security-relevant surface added beyond the plan's threat model. The D-07 guard is a slice-length read on parser-validated data; all other changes are test-only.

## Next Phase Readiness

- Phase 02 is complete: VIEW-03/04/05 behaviors are now explicitly locked by tests, ready for verify-work / phase audit.
- The D-04 cluster test (`TestBuildGraphC2ExpandedContainerRendersCluster`) provides a standing regression guard for the C2/C3 boundary-cluster branch — useful for any future expansion-semantics work.

---
*Phase: 02-auto-generate-c2-c3-diagrams*
*Completed: 2026-08-06*

## Self-Check: PASSED

All 5 plan files present (builder.go, builder_test.go, scope_test.go, root_test.go, SUMMARY.md); all 6 commits verified in git log (eb77ae3 RED, ad39e93 GREEN, c6c131a, ad34af4, 9e6b1f0, a20d5cb).
