---
phase: 04-c3-boundary-node-fix
plan: 01
subsystem: view
tags: [view, boundary, c3, go, tdd]

# Dependency graph
requires:
  - phase: 01-fix-c1-view-scoping
    provides: addResolvedBoundaryNode + addExternalBoundaryNodesForSubunits (the C2/C3 boundary node synthesis path this plan fixes)
provides:
  - "addResolvedBoundaryNode now bounds its peer walk-up at the expanded unit's parent, so C3 cross-container links resolve to sibling containers, not the parent system"
  - "TestGenerateC3View_SiblingContainerIsBoundaryNotParent regression test (BOUND-03)"
affects: [view, c3-diagrams, boundary-resolution, multilevel-fixture]

# Tech tracking
tech-stack:
  added: []  # pure stdlib strings change — no new deps (T-04-SC mitigated)
  patterns:
    - "Bound peer walk-up: a boundary-node resolver must stop its ancestor walk at the view root's parent, not at the link host's immediate parent"

key-files:
  created: []
  modified:
    - internal/view/scope.go
    - internal/view/scope_test.go

key-decisions:
  - "D-01 implemented via v.ExpandedUnit, not scopePath: scopePath (the link host's immediate parent) varies by recursion depth and does not carry the container path for nested links. v.ExpandedUnit is the stable view root (the expanded container/system). Signature unchanged."
  - "C2/C1 invariance preserved naturally: C2 v.ExpandedUnit is a top-level name (no dot) -> scopeParent '' -> guard inert; C1 never calls addResolvedBoundaryNode (routes through resolveAndAddBoundary)."

patterns-established:
  - "Boundary-node walk-up bound: derive the stop-level from the view's ExpandedUnit (single source of truth), never from a per-link-host parameter that drifts with recursion depth."

requirements-completed: [BOUND-01, BOUND-02, BOUND-03]

# Metrics
duration: 3min
completed: 2026-08-06
---

# Phase 4 Plan 01: C3 Boundary Node Fix Summary

**addResolvedBoundaryNode now stops its peer walk-up at the expanded container's parent, so C3 cross-container links surface the sibling container (e.g. mainSystem.rbac) as the boundary node instead of the parent system (mainSystem)**

## Performance

- **Duration:** 3 min
- **Started:** 2026-08-06T18:09:54Z
- **Completed:** 2026-08-06T18:13:06Z
- **Tasks:** 3 (RED / GREEN / verify)
- **Files modified:** 2 (`internal/view/scope.go`, `internal/view/scope_test.go`)

## Accomplishments

- **BOUND-01:** A C3 view of a container resolves a cross-container link to the sibling container, not the parent system. Verified end-to-end on `cmd/c4drill/testdata/multilevel.toml`: the `mainSystem/localIDP` C3 sub-diagram now shows `mainSystem.rbac` (Name "RBAC") plus the other sibling containers (authModules, storages, sshAuth) as boundary nodes, and the parent `mainSystem` node is absent.
- **BOUND-02:** C1 and C2 boundary behavior is unchanged. `go test ./...` is fully green; the named boundary tests pass unchanged (`TestIntegrationC1ViewNoNestedBoundaryPollution`, `TestGenerateC2View_ExternalBoundaryFromSubunitLinks`, `TestIntegrationExternalBoundaryNodes`).
- **BOUND-03:** A regression test `TestGenerateC3View_SiblingContainerIsBoundaryNotParent` locks in the sibling-as-boundary behavior (written first, confirmed RED against the buggy code, then GREEN after the fix).
- No signature change, no new imports, no golden/canonical DOT files modified.

## TDD Gate Compliance

- **RED gate:** `d3a56ff test(04-01): add failing C3 sibling-boundary regression test` — test written first; confirmed FAIL against current code (map contained `mainSystem`, absent `mainSystem.rbac`).
- **GREEN gate:** `710440b fix(04-01): bound addResolvedBoundaryNode walk-up at the expanded unit's parent` — minimal implementation; RED test now PASSes; full `internal/view` package green.
- **REFACTOR gate:** N/A — Task 3 was verification-only (no source edits), no cleanup needed.
- MVP+TDD gate predicate (`task.is-behavior-adding`) returned `false` for this plan (correctness bug fix, not net-new MVP behavior) — gate did not trip, no halt required.

## Task Commits

Each task was committed atomically (TDD cycle):

1. **Task 1 (RED): Add C3 sibling-boundary regression test** — `d3a56ff` (test)
2. **Task 2 (GREEN): Bound addResolvedBoundaryNode walk-up at the expanded unit's parent** — `710440b` (fix)
3. **Task 3 (REFACTOR/verify): Full suite + multilevel C3 fixture** — no commit (verification only; no source edits per plan)

## Files Created/Modified

- `internal/view/scope_test.go` — Added `TestGenerateC3View_SiblingContainerIsBoundaryNotParent`: builds a minimal model mirroring the multilevel.toml sibling case (`mainSystem.localIDP.grpcAPIs.authAPI` -> `mainSystem.rbac`), calls `view.GenerateC3View(m, "mainSystem.localIDP")`, asserts the sibling container is the boundary node (with real unit data Name "RBAC") and the parent system is not surfaced.
- `internal/view/scope.go` — Modified `addResolvedBoundaryNode`: computes `expandedParent` from `v.ExpandedUnit` (the view root) and adds a guard `if stripped == expandedParent { break }` inside the walk-up loop so the peer is not stripped past the expanded unit's parent. The trailing boundary-node creation block is unchanged.

## Decisions Made

- **D-01 realization uses `v.ExpandedUnit`, not `scopePath`.** The plan's `<interfaces>` stated "scopePath argument carries the container path", which holds only at the top-level call from GenerateC2View/GenerateC3View. `addExternalBoundaryNodesForSubunits` recurses and passes each link-host's *immediate* parent as `parentPath`/`scopePath`, so for a link authored on `mainSystem.localIDP.grpcAPIs.authAPI` the `scopePath` is `mainSystem.localIDP.grpcAPIs` (the componentBox), not `mainSystem.localIDP` (the container). A `scopePath`-derived bound therefore fires at the wrong level and the bug persists. `v.ExpandedUnit` is the stable view root and yields the correct stop-level (`mainSystem`). This is documented as a deviation (Rule 1) below.
- **C2/C1 invariance is structural, not special-cased.** No view-level branch was added: for C2 `v.ExpandedUnit` is a top-level name with no dot, so `expandedParent` is `""` and the guard never matches; C1 routes through `resolveAndAddBoundary` -> `resolveToTopLevel`/`resolveToViewAncestor` and never calls `addResolvedBoundaryNode`. This matches BOUND-02's requirement exactly.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] scopeParent source corrected from scopePath to v.ExpandedUnit**
- **Found during:** Task 2 (GREEN implementation)
- **Issue:** The plan's `<action>` specified computing `scopeParent` from the `scopePath` parameter (`if idx := strings.LastIndex(scopePath, "."); idx > 0 { scopeParent = scopePath[:idx] }`). Implemented literally, the RED test STILL FAILED: at the recursion depth where the sibling link lives, `scopePath` is the link host's immediate parent (`mainSystem.localIDP.grpcAPIs`), so the derived bound was `mainSystem.localIDP` — and the peer strips to `mainSystem` which is `!= mainSystem.localIDP`, so the guard never fired and the parent system was still surfaced. The plan's premise that `scopePath` carries the container path is only true for the top-level call, not for nested-link recursion.
- **Fix:** Derive the stop-level from `v.ExpandedUnit` (the stable view root set by GenerateC2View/GenerateC3View) instead of `scopePath`. The `scopePath` parameter is retained in the signature (still used by the trailing `createExternalBoundaryNode(m, peer, scopePath)` call) — no signature change.
- **Files modified:** `internal/view/scope.go`
- **Verification:** `TestGenerateC3View_SiblingContainerIsBoundaryNotParent` now PASSes; `TestGenerateC3View_ExternalBoundaryFromSubunitLinks` (top-level external peer) still PASSes; full `internal/view` package green; full `go test ./...` green. Fixture render confirms `mainSystem.rbac` present and parent `mainSystem` absent.
- **Committed in:** `710440b` (Task 2 GREEN commit)

---

**Total deviations:** 1 auto-fixed (1 bug — Rule 1)
**Impact on plan:** The deviation realizes the plan's *intent* (D-01: "stop walk-up at the container's parent level") with a correct data source. The plan's literal `<action>` was unimplementable as written because its premise about `scopePath` was incorrect at recursion depth. No scope creep — the fix is smaller and more robust than the planned change (no new branch logic, single source of truth). Behavior matches BOUND-01/02/03 exactly.

## Issues Encountered

- The plan's `<interfaces>` block asserted `scopePath` carries the container path at call sites. This is true only at the top-level `addExternalBoundaryNodesForSubunits` invocation from GenerateC2View/GenerateC3View; the helper recurses and rebinds `parentPath` to each nested unit's full path. Resolved by using `v.ExpandedUnit` as the bound source (see deviation above). Confirmed empirically: the scopePath-based implementation left the RED test failing, then the v.ExpandedUnit-based implementation turned it green.

## Verification Evidence

- `go test ./...` — all packages green (cmd/c4drill, internal/graph, internal/output, internal/parser, internal/render, internal/validator, internal/view).
- Named boundary tests pass: `TestIntegrationC1ViewNoNestedBoundaryPollution`, `TestGenerateC2View_ExternalBoundaryFromSubunitLinks`, `TestIntegrationExternalBoundaryNodes` (BOUND-02).
- Fixture end-to-end (BOUND-01): `go run ./cmd/c4drill cmd/c4drill/testdata/multilevel.toml -f dot -o /tmp/c4drill-phase04-verify`; the `mainSystem/localIDP.dot` C3 sub-diagram node set is exactly the 5 internal components plus 4 sibling containers (rbac, authModules, sshAuth, storages); parent `mainSystem` node absent; `grep -c 'RBAC' .../localIDP.dot` == 1; SVG render likewise contains `RBAC`.
- Pre/post comparison via a detached worktree at the RED commit (`d3a56ff`, buggy code) confirmed the pre-fix behavior: the `localIDP` C3 sub-diagram contained a spurious parent `mainSystem [+]` boundary node (URL `../mainSystem.svg`) with edges from authValidateScenario/authProvidersClient/grpcAPIs/actorC, and `mainSystem.rbac` was absent. The worktree was removed after comparison; no commits were made in it.

## User Setup Required

None — no external service configuration required. Pure stdlib Go change.

## Next Phase Readiness

- The C3 boundary resolution defect (DEFERRED-04) is closed; sibling containers now render correctly as C3 boundary nodes.
- No golden/canonical DOT files were modified, so no downstream regolden is needed.
- No blockers. The milestone v1.9 (C3 Boundary Node Fix) single plan is complete.

---
*Phase: 04-c3-boundary-node-fix*
*Completed: 2026-08-06*

## Self-Check: PASSED

- Files exist: `internal/view/scope.go`, `internal/view/scope_test.go`, `04-01-SUMMARY.md`.
- Commits exist: `d3a56ff` (RED), `710440b` (GREEN).
- Implementation content present: `expandedParent` bound source derived from `v.ExpandedUnit`; regression test `TestGenerateC3View_SiblingContainerIsBoundaryNotParent` present.
- No golden/canonical/DOT fixture files modified between plan-start (`1ca06df`) and HEAD.
