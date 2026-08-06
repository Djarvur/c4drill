---
phase: 01-fix-c1-view-scoping
plan: 03
subsystem: view
tags: [view, c1, resolution, visible-subunits, clusters, c4, tdd]

# Dependency graph
requires:
  - phase: 01-01 (0cd7991/d0af213/3adf383)
    provides: pair-only dedup key, Edge.PenWidth, View.AllExpanded, countPairMultiplicity
  - phase: 01-02 (4c8406d/5132df9/1fe07e3)
    provides: D-13 minlen gates at all six synthesis sites (C1 gate first operand was `path == sourceAncestor` — this plan replaced it with `path == resolvedSource`), AllExpanded activation
provides:
  - D-07..D-11: deepest-visible-ancestor resolution for C1 link targets AND sources via View.VisiblePaths + visible-subunit entries in v.Units
  - Within-cluster edges (D-10) recorded instead of dropped; no redundant parent edges (D-08)
  - Box grouping units follow the same resolution rules as systems (D-11)
  - BuildGraph C1 branch skips VisiblePaths entries (no duplicate node IDs in DOT)
affects: [verifier, future C1 rendering refinements, explore-link behavior (visible subunits now get ExploreURLs)]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Visible-subunit bookkeeping: View.VisiblePaths map + v.Units entries appended to UnitOrder AFTER the top-level loop, BEFORE boundary resolution — resolution reads v.Units, BuildGraph reads VisiblePaths"
    - "Unified deepest-visible walk: resolveToTopLevel delegates to the existing resolveToViewAncestor (Don't Hand-Roll directive); external-peer fallback preserved"
    - "D-13 C1 gate first operand is now path == resolvedSource (both-original check on the deepest visible source)"

key-files:
  created: []
  modified:
    - internal/view/view.go
    - internal/view/scope.go
    - internal/graph/builder.go
    - internal/view/scope_test.go
    - internal/view/integration_test.go

key-decisions:
  - "Visible subunits (direct children of expanded top-level C1 units) are added to v.Units + UnitOrder and marked in v.VisiblePaths; BuildGraph skips them as top-level nodes because buildCluster already renders them inside the parent cluster (Pitfall 5)"
  - "resolveAndAddBoundary sources from resolveToViewAncestor(v, path) — the deepest VISIBLE ancestor of the source path; append condition resolved != resolvedSource keeps within-cluster edges (D-10) and suppresses parent edges (D-08)"
  - "D-13 gate first operand updated to path == resolvedSource: a visible subunit's own links stay length-eligible when the peer also resolves unchanged"
  - "resolveToTopLevel unified onto resolveToViewAncestor with the peer-as-is fallback for truly-external units (boundary-node contract preserved)"

patterns-established:
  - "Expanded-C1 test fixture: top-level unit with Expanded: [self] + SubunitOrder children; hidden grandchildren (sshAuth.sshd) exercise D-07"
  - "Graph-level C1 assertions: cluster node membership + edge endpoints + RenderDOT success (duplicate node IDs would fail rendering)"
  - "Shared fixture path constants (webUserPath/keycloakPath/linuxSystemPath/sshAuthPath/webAPIPath) keep repeated literals below the goconst threshold"

requirements-completed: [VIEW-01, VIEW-02, EDGE-01, EDGE-02]

# Metrics
duration: 6min
completed: 2026-08-06
---

# Phase 1 Plan 3: Deepest-Visible-Ancestor Resolution in C1 Summary

**D-07..D-11 implemented: C1 links resolve to the deepest VISIBLE ancestor on BOTH sides — visible subunit entries (View.VisiblePaths + v.Units) inside expanded clusters receive subunit-level edges (D-07/D-09), within-cluster edges are recorded instead of dropped (D-10), parent edges are never duplicated (D-08), boxes follow the same rules (D-11), and BuildGraph skips VisiblePaths entries to prevent duplicate node IDs in DOT (Pitfall 5)**

## Performance

- **Duration:** 6 min
- **Started:** 2026-08-06T09:31:17Z (first commit 12:31:17 +0300)
- **Completed:** 2026-08-06T09:36:49Z
- **Tasks:** 3 (TDD: RED + GREEN + regression)
- **Files modified:** 5

## Accomplishments

- Visible-subunit bookkeeping: `GenerateC1View` now appends direct subunits of expanded top-level units to `v.Units` + `UnitOrder` (entry construction mirrors buildCluster's child-entry contract) and marks them in the new `View.VisiblePaths` map, populated between the top-level loop and `addC1BoundaryNodes` so resolution reads them (Pitfall 5).
- Unified resolution: `resolveToTopLevel` now delegates to the existing `resolveToViewAncestor` walk (Don't Hand-Roll directive) — a peer resolves to a visible subunit path when its parent cluster is expanded, with the peer-as-is fallback preserving the external-boundary-node contract.
- Source-side resolution (D-09): `resolveAndAddBoundary` sources from `resolveToViewAncestor(v, path)` instead of the raw top-level ancestor — links authored inside an expanded cluster land on the visible subunit entry.
- Within-cluster edges (D-10) + no redundant parent edges (D-08): the append condition became `resolved != resolvedSource`; the D-13 minlen gate's first operand became `path == resolvedSource` (visible subunit's own links stay length-eligible).
- BuildGraph C1 branch skips `v.VisiblePaths[key]` entries — nil-map reads keep C2/C3/expanded/hand-built views unaffected.
- Box parity (D-11) verified with a TypeBox fixture producing identical resolution behavior.
- Seven RED tests flipped GREEN, including a graph-level test asserting the subunit node lives inside `cluster_linuxSystem` (not top-level), the D-07 edge `webUser → linuxSystem.sshAuth`, RenderDOT success, and a new explore-link side-effect test (visible subunit WITH subunits now receives an ExploreURL via BuildGraphWithPath).

## Task Commits

Each task/gate was committed atomically:

1. **Task 1 (RED): skeleton field + failing tests** - `000819f` (test)
2. **Task 2 (GREEN): visible-subunit bookkeeping + unified resolution + BuildGraph skip** - `40b635e` (feat)
3. **Task 3 (regression): lint fixes on new test code** - `a58b6f0` (refactor)

**Plan metadata:** (final docs commit follows this summary)

## Files Created/Modified

- `internal/view/view.go` - Added `View.VisiblePaths map[string]bool` with doc comment (visible-subunit bookkeeping, Pitfall 5)
- `internal/view/scope.go` - GenerateC1View populates visible subunits (entry + UnitOrder + VisiblePaths) before addC1BoundaryNodes; resolveAndAddBoundary sources from resolveToViewAncestor with `resolved != resolvedSource` skip and `path == resolvedSource` D-13 gate; resolveToTopLevel unified onto resolveToViewAncestor
- `internal/graph/builder.go` - BuildGraph C1 branch skips v.VisiblePaths entries (prevents duplicate node IDs)
- `internal/view/scope_test.go` - Shared expanded-C1 fixture helpers (expandedC1Model/expandedC1BoxModel + path constants) and six D-07..D-11 tests
- `internal/view/integration_test.go` - Graph-level TestBuildGraphExpandedC1VisibleSubunitEdges + explore-link side-effect test

## Decisions Made

- Followed the plan exactly: VisiblePaths + v.Units entries, unified resolveToTopLevel, resolvedSource semantics, BuildGraph skip. C2/C3 resolution passes (resolveBoundaryNodeLinks, resolveSubunitCrossLinks, resolveDescendantCrossLinks — scope.go:652-859) received ZERO edits (Pitfall 1, verified via git diff hunk review).
- The explore-link side effect is documented as a sibling test (`TestBuildGraphExpandedC1VisibleSubunitExploreLink`) rather than a subtest of the view-level test — it exercises BuildGraphWithPath, which is graph-level (plan explicitly allowed "or a sibling test").
- Shared fixture constants introduced for repeated path strings — also the goconst fix (see deviations).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Lint] 11 new golangci-lint issues on new test code**
- **Found during:** Task 3 (lint gate)
- **Issue:** `golangci-lint run --new-from-rev=ff3451f` reported 11 issues on the new tests: 5x goconst (repeated fixture literals "webUser"/"keycloak"/"linuxSystem"/"linuxSystem.sshAuth"/"Web User" across the package) and 6x wsl_v5 (missing whitespace above range loops and asserts in the graph-level integration tests).
- **Fix:** Extracted shared fixture path constants (webUserPath/keycloakPath/linuxSystemPath/sshAuthPath/webAPIPath/webUserName) used by both test files; added blank lines before the flagged `for` loops and follow-up asserts per wsl_v5.
- **Files modified:** internal/view/scope_test.go, internal/view/integration_test.go
- **Verification:** `golangci-lint run --new-from-rev=ff3451f ./...` reports 0 issues; full suite green with -race.
- **Committed in:** a58b6f0 (Task 3 commit)

**2. [Rule 3 - Blocking] `mise lint` not run as-is (lint-fix pollution)**
- **Found during:** Task 3 (lint gate)
- **Issue:** The `mise lint` task runs `golangci-lint run --fix ./...` (lint-fix) first, which auto-modifies files outside this plan's scope — the exact problem 01-01 documented (12 files modified, 9 out of scope). Pre-existing lint debt (~121 issues) is logged to deferred-items.md.
- **Fix:** Used `golangci-lint run --new-from-rev=ff3451f ./...` (no --fix) per the phase convention established in 01-01/01-02; fixed only the new-code issues manually.
- **Verification:** 0 new issues vs wave-2 HEAD; full suite green with -race; pre-existing debt untouched.
- **Committed in:** a58b6f0 (Task 3 commit)

---

**Total deviations:** 2 auto-fixed (1 test-code lint, 1 lint-gate methodology)
**Impact on plan:** Both necessary for the lint gate to be meaningful; no scope creep, no production behavior change from either.

## Issues Encountered

- None — the TDD cycle behaved exactly as designed: all seven RED tests failed for the predicted reasons (visible subunits absent from v.Units, resolution returning "linuxSystem", within-cluster links dropped as internal, edge pointing at the parent), and GREEN flipped them all. The D-08 no-redundant-parent-edge and D-10 within-cluster semantics were verified simultaneously: `resolved != resolvedSource` both keeps the within-cluster edge AND suppresses the parent edge.
- `TestGenerateC1View_NoRedundantParentEdge` and `TestGenerateC1View_ResolvesToVisibleSubunit` share the same fixture — intentional: one asserts single-entry ResolvedLinks with the subunit peer, the other asserts the peer value explicitly (both are required by the plan).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 1 (01-fix-c1-view-scoping) is now complete: all 13 decisions D-01..D-13 are implemented and the full suite is green with -race.
- The explore-link side effect is a behavioral consequence to be aware of in later phases: visible subunits WITH subunits (inside expanded C1 clusters) now receive ExploreURLs — drill-down to auto-generated C2 diagrams.
- The `resolveToTopLevel` function is now a thin delegation to `resolveToViewAncestor`; future phases may consider retiring it, but it remains the documented C1 entry point per this plan.
- deferred-items.md: pre-existing golangci-lint debt (~121 issues) remains deferred.

## Self-Check: PASSED

- Files: 01-03-SUMMARY.md exists in the plan directory
- Commits: 000819f (test RED), 40b635e (feat GREEN), a58b6f0 (refactor) all present
- TDD gate sequence verified in git log: test → feat → refactor
- `go test -v -race ./...` green; `golangci-lint run --new-from-rev=ff3451f ./...` reports 0 new issues
- C2/C3 resolution passes untouched (Pitfall 1): TestIntegrationC2ViewNestedSystem, TestIntegrationC3ViewNestedContainer, TestIntegrationBuildGraphFromC2ViewWithClusters, boundary/cross-subunit tests green

---
*Phase: 01-fix-c1-view-scoping*
*Completed: 2026-08-06*
