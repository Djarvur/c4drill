---
phase: 02-auto-generate-c2-c3-diagrams
reviewed: 2026-08-06T00:00:00Z
depth: standard
files_reviewed: 4
files_reviewed_list:
  - internal/graph/builder.go
  - internal/graph/builder_test.go
  - internal/view/scope_test.go
  - cmd/c4drill/root_test.go
findings:
  critical: 0
  warning: 1
  info: 3
  total: 4
status: issues_found
---

# Phase 2: Code Review Report

**Reviewed:** 2026-08-06T00:00:00Z
**Depth:** standard
**Files Reviewed:** 4
**Status:** issues_found

## Summary

Refinement phase over the existing v1.8 implementation. The single production change is the D-07 guard in the C1 branch of `BuildGraph` (`internal/graph/builder.go:68-70`) — verified via `git diff` against the phase base that the C2/C3 boundary-cluster branch (builder.go:30-55), `scope.go`, and all Phase 1 dedup/penwidth/mirror machinery are byte-identical to before the phase.

The D-07 guard itself is correct and regression-free: it only changes behavior when `len(entry.Unit.Subunits) == 0`, where `buildCluster` previously produced an empty cluster box (0 nodes) — no rendered content is ever lost. The new test `TestBuildGraphExpandedEmptyUnitRendersPlainNode` is a genuine RED/GREEN test (pre-fix code produces `g.Clusters == [cluster_app]`, failing both `require.Empty(g.Clusters)` and `require.Len(g.Nodes, 1)`).

All four new test functions are meaningful (not tautological) and their fixtures are valid per the validator rules: the box fixture in `root_test.go` passes `ValidateOrphanUnits` (every unit has links or subunits — `boxname.child` and `system.cbox.comp` carry the authored link/linkFrom the executor added), `ValidateNestingHierarchy` (box→system is C1-in-C1; system→containerBox→container is C1→C2→C2), `ValidateBoxMixedContents` (box contains only internal units), and `ValidateReferences` (both peers defined). The asserted output paths (`boxtest/boxname.svg`, `boxtest/system/cbox.svg`) match the writer's dotted-path layout (writer.go:37-47) and `collectExpandableUnitPaths` auto-detect (root.go:170-194, boxes included).

Verification performed: full test suites for the three affected packages pass; `go vet` clean; new tests confirmed green individually; the render-based tests are parallel-safe because `render.RenderDOT` serializes on the package `wasmMutex` (render.go:15-20), so the `t.Parallel()` calls in builder tests do not contradict the cmd-package convention of suppressing `paralleltest` (cmd tests were left non-parallel conservatively — consistent, no action needed).

Findings below are limited to one residual defect of the same class D-07 fixes (deliberately left in the C2/C3 branch per the plan's edit-scope guard) and three minor test-quality observations.

## Warnings

### WR-01: D-07 empty-cluster defect persists in the C2/C3 branch

**File:** `internal/graph/builder.go:45` (contrast with the D-07 guard at :70)
**Issue:** The C2/C3 branch still uses the bare `if entry.IsExpanded {` condition. An entry that is expanded but has no subunits renders an empty cluster box inside the boundary cluster — the exact defect class D-07 eliminates for C1. This is reachable from real TOML: a parent's `expanded` list naming a child key whose unit has no subunits (e.g., `[app] expanded = ["api"]` where `[app.api]` declares no children) produces an empty `cluster_app.api` in the C2 diagram. The phase plan explicitly scoped this out ("Do NOT touch the C2/C3 branch"), so this is a documented residual rather than a regression — but the D-07 rationale ("no empty cluster box") applies equally here, and the fix would be the same one-condition guard. Track for a follow-up phase.
**Fix:** (out of this phase's scope, for follow-up) apply the same guard at builder.go:45:
```go
if entry.IsExpanded && len(entry.Unit.Subunits) > 0 {
    cluster := buildCluster(entry)
    boundaryCluster.Clusters = append(boundaryCluster.Clusters, cluster)
} else {
    node := buildNode(entry)
    node.IsInCluster = true
    boundaryCluster.Nodes = append(boundaryCluster.Nodes, node)
}
```

## Info

### IN-01: Edit-scope guard claim in C2 cluster test is overstated

**File:** `internal/graph/builder_test.go:283-285`
**Issue:** The comment on `TestBuildGraphC2ExpandedContainerRendersCluster` claims "it fails if the C2/C3 branch is altered." That is only partially true: if a future executor misapplied the D-07 guard to the C2/C3 branch, this test would still pass, because `app.api` HAS subunits (`len(Subunits) > 0`), so `IsExpanded && len(Subunits) > 0` still builds the cluster. The test only locks the expanded-with-subunits path of the branch.
**Fix:** Soften the comment to "fails if the C2/C3 branch stops rendering expanded containers as clusters" — or, if the empty-expanded case is decided to keep current behavior, add an explicit assertion (e.g., a C2 fixture with a parent `expanded` entry naming a subunit without subunits) documenting the branch's intentional asymmetry with C1.

### IN-02: New constant surfaces pre-existing goconst suggestions

**File:** `internal/graph/builder_test.go:280` (constant), flagged literals at :497, :942, :1484
**Issue:** Adding `const authComponentName = "auth"` makes goconst flag the pre-existing `"auth"` map-key literals (5 occurrences now) with "such constant already exists". The file already carries ~50 goconst findings (not gating), so this is noise-level — but the phase's own refactor commit (9e6b1f0 "clean up new test fixtures for linters") signals intent to keep this file tidy.
**Fix:** Replace the three remaining `"auth": {` literals in `TestBuildExpandedGraph` (recursive clusters subtest), the C3 edge-color test, and the C3 navigation test with `authComponentName:`.

### IN-03: Minor coverage gaps in the two pipeline/expansion tests

**File:** `cmd/c4drill/root_test.go:453-460` and `internal/graph/builder_test.go:272-275`
**Issue:** (a) `TestFullPipeline_BoxWithSubunitsGeneratesSubDiagram` asserts the box C2 and containerBox C3 files but not `boxtest/system.svg` — the C2 for the plain `system` unit — so the D-01 "any unit with subunits" claim is only transitively locked for plain systems (via the `cbox` C3 file, which requires recursing through `system`). (b) The D-07 test does not assert that the plain node renders without the "[+]" indicator (plan states `HasSubunits=false` means no "[+]"); and the per-unit self-reference variant of the empty-expanded case (`unit.Expanded = ["self"]` with no subunits) is untested — only the `properties.expanded` side of the OR is exercised for empty units.
**Fix:** Optionally add `assert.FileExists(t, filepath.Join(outputDir, "boxtest", "system.svg"))` and one assertion that `g.Nodes[0].Label.Name` does not contain "[+]" in the D-07 test. Both are nice-to-have; the current tests are not misleading.

---

_Reviewed: 2026-08-06T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
