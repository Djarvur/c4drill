---
phase: 37-nesting-context-and-plain-rendering
plan: 01
subsystem: graph
tags: [graph, clusters, nesting, recursion, graphviz, go, tdd]

# Dependency graph
requires:
  - phase: 36-edge-semantics-and-legend
    provides: buildNestedCluster recursion, buildCluster flat child loop, BuildGraphWithPath one-level explore-URL walk, node-side 🔍 affordance (buildNode)
provides:
  - "Recursive buildCluster: expanded units render subunit-containers as nested clusters with grandchildren unfolded (leaves as nodes) — CTX-03"
  - "Cluster.ExploreURL field + recursive BuildGraphWithPath walk (assignExploreURLs / assignExploreURLsToClusters + entryShouldHaveExploreLink predicate)"
  - "buildClusterLabel 🔍 guard (HasSubunits && !IsExpanded) — nested container clusters keep the drill affordance"
  - "setClusterLabel emits the subgraph URL attribute via SafeSet(\"URL\", ...) — clusters are clickable like nodes"
  - "TestBuildGraph_ExpandedClusterRendersNestedSubClusters (structure + affordance regression guard)"
affects: [37-02 (C1 deep-link chain unfolding consumes buildCluster recursion), 37-05 (golden re-baseline scope), 37-07 (docs: nested-cluster behavior)]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Child dispatch on HasSubunits with view-entry lookup + inline fallback (copied from buildNestedCluster into buildCluster)"
    - "Cluster-side drill affordance: Cluster.ExploreURL + subgraph URL attribute, mirroring Node.ExploreURL + SetURL"
    - "Recursive explore-URL walk over cluster.Clusters (was one-level)"

key-files:
  created: []
  modified:
    - internal/graph/builder.go
    - internal/graph/graph.go
    - internal/graph/builder_test.go
    - internal/render/converter.go
    - internal/render/converter_test.go
    - internal/view/integration_test.go

key-decisions:
  - "buildCluster dispatches on childEntry.HasSubunits exactly like buildNestedCluster: a container skeleton inside an expanded cluster unfolds FULLY (buildNestedCluster semantics), with 🔍 + ExploreURL marking it as author-drillable — per plan must_haves"
  - "entryShouldHaveExploreLink extracted from shouldHaveExploreLink so nodes and nested container clusters share one drill predicate"
  - "setClusterLabel ignores SafeSet error (_ =) matching applyNodeStyle precedent; URL emitted only when non-empty so flat models stay byte-stable"

patterns-established:
  - "CTX-03 recursion: buildCluster(entry, v) → buildNestedCluster(childEntry, childPath, v) — drill-down views and nesting pictures now share one recursive builder"

requirements-completed: [CTX-03]

# Metrics
duration: 13min
completed: 2026-08-30
---

# Phase 37 Plan 01: Nesting Context — Recursive buildCluster (CTX-03) Summary

**Recursive `buildCluster` renders subunit-containers inside expanded clusters as nested clusters (grandchildren unfolded, leaves as nodes) with the full drill affordance — 🔍 labels, `Cluster.ExploreURL` assigned by a recursive walk, emitted as the GraphViz subgraph `URL` attribute.**

## Performance

- **Duration:** 13 min
- **Started:** 2026-08-30T12:30:39Z
- **Completed:** 2026-08-30T12:44:24Z
- **Tasks:** 3 (RED / GREEN / VERIFY)
- **Files modified:** 6

## Accomplishments
- CTX-03: expanded units now depict nested elements through their intermediate containers — a container child inside an expanded cluster renders as a nested sub-cluster whose own children unfold recursively; leaf children still render as direct nodes
- Cluster drill affordance created end-to-end: `Cluster.ExploreURL` (graph.go) → assigned by a recursive `BuildGraphWithPath` walk (builder.go) → emitted as the subgraph `URL` attribute in `setClusterLabel` (converter.go), mirroring the node-side `SetURL` mechanism
- 🔍 affordance preserved/extended: nested container cluster labels carry 🔍 via the new `buildClusterLabel` guard (`HasSubunits && !IsExpanded`); author-expanded clusters stay 🔍-free; leaf-node 🔍 (buildNode) untouched
- D-07 expanded-but-empty guard preserved at both caller sites (L72-79 / L97-105 untouched); child ordering deterministic (SubunitOrder with map-key fallback)
- Verified end-to-end: CLI C1 output for a model with expanded units shows nested clusters, 15 🔍 markers, and `URL="multilevel/mainSystem/sshAuth.svg"` on the nested cluster

## Expected-Failure List (LOAD-BEARING for 37-02 / 37-05)

**EMPTY — zero failing tests in the repo after GREEN.** The plan anticipated the two committed-golden baselines (`TestBuildExpandedGraphBaselineDOT` + REF-05 alias `TestReference_BackwardCompat`, both reading `cmd/c4drill/testdata/multilevel.expanded.dot`) to fail as the documented CTX-03 delta. They did **not** fail. Verified precisely (uncached `go test -count=1 ./...`: 15 packages ok, 0 failures; both baselines pass by name).

Why the anticipated golden delta did not materialize:
1. Both baselines render the **`--expanded`** view. `GenerateExpandedView` sets `IsExpanded = (len(Subunits) > 0)` on EVERY entry, so (a) `BuildExpandedGraph` routes through `buildNestedCluster` — whose structure this plan does not touch — and (b) the new `buildClusterLabel` 🔍 guard is provably inert there (`HasSubunits && !IsExpanded` is always false). Expanded-mode DOT output is unchanged.
2. `buildCluster` recursion only affects **author-expanded units** in C1/C2/C3 views. `multilevel.toml` has `expanded = ["mainSystem"]`, so the CLI's C1 output for that fixture DOES change — but the committed `cmd/c4drill/testdata/multilevel.dot`/`.svg`/`expanded/` files are generation artifacts, **not test baselines**: no test reads them for comparison (grep-verified; the only golden readers in the repo are the two graph-package tests above). All cmd/c4drill tests are content-based and still pass.
3. The new cluster `URL` attribute is assigned only by `BuildGraphWithPath`; the golden test renders path-less.

**Consequence for plan 37-05:** there may be NO committed goldens left to re-baseline for CTX-03. 37-05 should re-scan; if its purpose was solely the CTX-03 golden delta, expect it to reduce to verifying flat-model stability (BC-01) plus any deltas 37-02/37-03/37-04 introduce. **Consequence for 37-02:** the recursion this plan landed is the unfolding mechanism C1 chain entries will flow into; no expected-failure debt carries over.

## Task Commits

Each task was committed atomically:

1. **Task 1 (RED): nested-sub-cluster structure test** - `b1d59c8` (test) — plus inert `Cluster.ExploreURL` field so the affordance assertion compiles (deviation, see below)
2. **Task 2 (GREEN): recursive buildCluster + cluster drill affordance** - `2636578` (feat)
3. **Task 3 (VERIFY): package suite green** - no commit (verification only, no source edits)

_Note: TDD plan — test → feat commits; no refactor commit needed (implementation landed clean, lint 0 issues)._

## Files Created/Modified
- `internal/graph/builder.go` — `buildCluster(entry, v)` recursion (nested-cluster dispatch, Clusters init, view-entry child lookup with inline fallback), both call sites pass `v`, `buildClusterLabel` 🔍 guard, recursive `BuildGraphWithPath` walk (`assignExploreURLs`, `assignExploreURLsToClusters`, `entryShouldHaveExploreLink`)
- `internal/graph/graph.go` — `Cluster.ExploreURL` field
- `internal/graph/builder_test.go` — `TestBuildGraph_ExpandedClusterRendersNestedSubClusters` (structure + both drill affordances, RED-first)
- `internal/render/converter.go` — `setClusterLabel` emits `SafeSet("URL", cluster.ExploreURL, "")` when non-empty
- `internal/render/converter_test.go` — `TestClusterExploreURLEmission` (cluster URL in DOT; no URL without ExploreURL)
- `internal/view/integration_test.go` — two integration tests updated from old flat-node assertions to nested-cluster assertions (plan Task 2 instruction: update old-structure assertions, do not weaken the new test)

## Decisions Made
- Container children inside an expanded cluster unfold **fully** (buildNestedCluster semantics) regardless of the child's own `IsExpanded`, with 🔍 + ExploreURL distinguishing author-drillable containers — exactly the plan's must_haves ("a container skeleton inside an expanded cluster unfolds fully")
- `setClusterLabel` early-return on empty HTML label retained (URL emission sits after `SetLabelHTML`, per plan); in practice cluster labels are never empty
- Defensive child-entry fallback preserves `isUnitExpanded(entry.Unit, childName)` from the old code so fallback-built entries keep their expansion hint

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Added inert `Cluster.ExploreURL` field in the RED commit**
- **Found during:** Task 1 (RED)
- **Issue:** The plan's Task 1 test asserts `authClusterWithPath.ExploreURL`, but the field is only added by Task 2 — the RED test could not compile ("Test compiles and FAILS" was the Task 1 done criterion, directly conflicting with "No production code changed")
- **Fix:** Added the `ExploreURL string` field to `graph.Cluster` in the RED commit — an inert skeleton (no writers/readers until GREEN, zero behavior change), the standard TDD compile-enabling move
- **Files modified:** internal/graph/graph.go
- **Verification:** RED test compiled and failed on the nested-cluster assertion, not on compilation; rest of package stayed green
- **Committed in:** b1d59c8 (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking). **Impact on plan:** compile-gate conflict inside the plan itself; resolution preserved RED→GREEN semantics exactly. No scope creep.

## Issues Encountered
- Two pre-existing `internal/view` integration tests (`TestBuildGraphExpandedC1VisibleSubunitEdges`, `TestBuildGraphExpandedC1VisibleSubunitExploreLink`) failed after GREEN because they pinned the OLD flat-node rendering of a visible subunit with subunits. This is the exact case Task 2's action anticipates ("If an existing test asserts the OLD flat-child behavior... update that assertion to the recursive structure") — updated to nested-cluster assertions; not counted as a deviation.
- Pre-existing, out-of-scope (not touched, per scope boundary): gofmt comment-formatting drift in `internal/graph/shapes.go` and `internal/graph/integration_test.go` (files this plan did not modify); untracked `.planning/.../37-PATTERNS.md` and phase-36 planning-file deletions predate this execution (orchestrator state, left alone).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- CTX-03 mechanism complete and regression-guarded; 37-02 (C1 deep-link ancestor chains) can rely on `buildCluster` recursion to unfold chain targets' containers
- No expected-failure debt carries forward (see Expected-Failure List above — baselines stayed green; 37-05's re-baseline scope should be re-assessed)
- Full repo suite green (uncached): 15/15 packages; `go build ./...` clean; golangci-lint 0 issues on changed packages; `cmd/c4drill/testdata/` untouched

---

## Self-Check: PASSED

- All 6 modified files exist on disk (verified with `[ -f ]`)
- Commits verified in `git log`: `b1d59c8` (test), `2636578` (feat)
- Plan-level `<verification>` re-run: new test PASS; `go test ./internal/render/` PASS; `go test ./internal/view/` PASS; `go build ./...` OK; internal/graph package fully green (zero golden failures); D-07 + node-side 🔍 tests green
- No file deletions in either commit (`git diff --diff-filter=D` empty)

---
*Phase: 37-nesting-context-and-plain-rendering*
*Completed: 2026-08-30*
