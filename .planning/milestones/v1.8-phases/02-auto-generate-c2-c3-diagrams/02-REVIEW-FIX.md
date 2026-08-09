---
phase: 02-auto-generate-c2-c3-diagrams
fixed_at: 2026-08-06T00:00:00Z
review_path: .planning/phases/02-auto-generate-c2-c3-diagrams/02-REVIEW.md
iteration: 1
findings_in_scope: 1
fixed: 1
skipped: 0
status: all_fixed
---

# Phase 2: Code Review Fix Report

**Fixed at:** 2026-08-06T00:00:00Z
**Source review:** `.planning/phases/02-auto-generate-c2-c3-diagrams/02-REVIEW.md`
**Iteration:** 1

**Summary:**
- Findings in scope: 1 (WR-01; INFO findings excluded per fix scope)
- Fixed: 1
- Skipped: 0

## Fixed Issues

### WR-01: D-07 empty-cluster defect persists in the C2/C3 branch

**Files modified:** `internal/graph/builder.go`, `internal/graph/builder_test.go`
**Commits:** `49b9c83` (test), `7074a87` (fix)
**Applied fix:** Applied the D-07 one-condition guard to the C2/C3 boundary-cluster branch of `BuildGraph` (`internal/graph/builder.go:45`): `if entry.IsExpanded` became `if entry.IsExpanded && len(entry.Unit.Subunits) > 0`. An expanded entry naming a subunit without subunits now renders as a plain node inside the boundary cluster instead of an empty cluster box — matching the C1 branch guard at builder.go:70. No other code was touched (scope.go, penwidth/multiplicity machinery, and the C1 branch are unchanged).

**TDD:** Added `TestBuildGraphC2ExpandedEmptySubunitRendersPlainNode` (committed first, `49b9c83`). Confirmed RED against the pre-fix code (boundary cluster contained an empty `cluster_app.api`; `require.Empty(g.Clusters[0].Clusters)` failed), then GREEN after the fix (`7074a87`).

**Verification:**
- `go test ./...` — all packages green, including `TestBuildGraphC2ExpandedContainerRendersCluster` (fixture `app.api` has subunits, still renders as cluster) and all Phase 1 penwidth tests (`TestBuildEdgesPenwidth*`, `TestBuildEdgesExpandedExemption`, `TestBuildExpandedGraphBaselineDOT`, `TestBuildExpandedGraphRealToml`).
- `go vet ./...` — clean.
- `golangci-lint run --new-from-rev=3e258e1 ./...` — 0 new issues.

Note: `cyp-auth-infra/` is gitignored (`.gitignore:72`) and therefore absent from the isolated worktree; the two real-TOML fixture tests (`TestBuildExpandedGraphRealToml`, `TestBuildExpandedGraphBaselineDOT`) were verified with the fixture copied into the worktree.

---

_Fixed: 2026-08-06T00:00:00Z_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
