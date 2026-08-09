---
phase: 01-fix-c1-view-scoping
plan: 02
subsystem: graph
tags: [view, boundary-nodes, minlen, expanded, c4, tdd]

# Dependency graph
requires:
  - phase: 01-01 (0cd7991/d0af213/3adf383)
    provides: View.AllExpanded field, pair-only dedup key, Edge.PenWidth, countPairMultiplicity
provides:
  - Legacy addExternalBoundaryNodes + addExternalBoundaryNodesRecursive deleted; validator is the single gatekeeper for undefined peers (D-12)
  - GenerateExpandedView sets View.AllExpanded = true — expanded mode re-activated on the v1.7 tech+desc key and 2.0 penwidth (D-02, COMPAT-02)
  - D-13 minlen gating at all six resolved-link synthesis sites: resolved edges carry Length 0, direct original pairs keep it
affects: [01-03 (visible-subunit resolution — will update the `path == sourceAncestor` first operand of the C1 gate), verifier]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Length gate at synthesis sites: a synthesized link copies Length only when (drawn source == original source) AND (drawn peer == original peer)"
    - "originalSource parameter threaded into addResolvedCrossLink/From to distinguish direct subunit links (length-eligible) from descendant-recursion links (never eligible)"

key-files:
  created: []
  modified:
    - internal/view/scope.go
    - internal/view/scope_test.go
    - internal/graph/builder_test.go

key-decisions:
  - "D-12 implemented by deleting the legacy recursive boundary path entirely — createExternalBoundaryNode kept for addC1BoundaryNodes/addResolvedBoundaryNode"
  - "D-13 C1 gate first operand is `path == sourceAncestor` (top-level-only check); plan 01-03's deepest-visible resolution will replace it"
  - "resolveBoundaryNodeLinks gates only on `link.Peer == resolvedPeer` — the boundary-node source is never resolved, only the peer side can resolve"
  - "resolveDescendantCrossLinks always passes fullPath as originalSource — a descendant-authored link never equals the drawn sourcePath, so its length is always dropped"

patterns-established:
  - "Synthesis-site Length gate: `length := 0; if <both endpoints original> { length = link.Length }` — 6 sites, no builder.go change needed (MinLen 0 → converter omits minlen)"
  - "Test guard placement: baseline DOT test asserts per-edge blocks via string Contains (minlen=, penwidth=2) — go-graphviz wraps edge attrs across lines"

requirements-completed: [VIEW-01]

# Metrics
duration: 5min
completed: 2026-08-06
---

# Phase 1 Plan 2: Legacy Boundary Cleanup + AllExpanded Activation + MinLen Gating Summary

**D-12 removal of the legacy recursive external-boundary path from GenerateExpandedView, D-02 activation of View.AllExpanded (restoring v1.7 dedup key and 2.0 penwidth in expanded mode), and D-13 minlen gating at all six resolved-link synthesis sites so length applies only to original-pair edges**

## Performance

- **Duration:** 5 min
- **Started:** 2026-08-06T09:17:00Z (approx., first commit 12:19:21 +0300)
- **Completed:** 2026-08-06T09:22:48Z
- **Tasks:** 3 (TDD: RED + GREEN + regression)
- **Files modified:** 3

## Accomplishments

- D-12: `addExternalBoundaryNodes` + `addExternalBoundaryNodesRecursive` and their call in `GenerateExpandedView` deleted — C1/expanded views can never again be polluted by the legacy recursive boundary path; the validator (internal/validator/rules.go) is the single gatekeeper for undefined peers. `createExternalBoundaryNode` kept (used by addC1BoundaryNodes and addResolvedBoundaryNode).
- D-02: `GenerateExpandedView` now sets `AllExpanded: true` on the view literal — the v1.7 tech+desc dedup key and unconditional 2.0 penwidth are restored for `--expanded` output (COMPAT-02). Doc comment dropped the "It includes external boundary nodes..." sentence.
- D-13: `Length` is copied into a synthesized link ONLY when both drawn endpoints are the link's original units, at all six sites: `resolveAndAddBoundary` (outgoing + incoming, gate `path == sourceAncestor && link.Peer == resolved`), `resolveBoundaryNodeLinks` (outgoing + incoming, gate `link.Peer == resolvedPeer`), and `addResolvedCrossLink`/`addResolvedCrossLinkFrom` (new `originalSource` parameter, gate `originalSource == sourcePath && link.Peer == resolvedPeer`). No builder.go change needed — Length 0 flows to Edge.MinLen 0 and the converter omits minlen.
- Three RED tests flipped GREEN: inverted expanded-boundary test (`TestGenerateExpandedView_SkipsExternalBoundaryNodesForLinkedUnits`), `TestBuildGraphResolvedEdgeMinLen` (resolved edge drops minlen / direct pair keeps it), and `TestBuildExpandedGraphBaselineDOT` (COMPAT-02 guard: every expanded edge keeps penwidth 2.0 and DOT contains minlen=).

## Task Commits

Each task/gate was committed atomically:

1. **Task 1 (RED): failing tests** - `4c8406d` (test)
2. **Task 2 (GREEN): cleanup + AllExpanded + minlen gates** - `5132df9` (feat)
3. **Task 3 (regression): lint fixes on new code** - `1fe07e3` (refactor)

**Plan metadata:** (final docs commit follows this summary)

## Files Created/Modified

- `internal/view/scope.go` - Deleted addExternalBoundaryNodes (:154) + addExternalBoundaryNodesRecursive (:178) + call at :49; `AllExpanded: true` in GenerateExpandedView literal; Length gates at 6 synthesis sites; `originalSource` param added to addResolvedCrossLink/From; call sites in resolveSubunitCrossLinks/resolveDescendantCrossLinks updated
- `internal/view/scope_test.go` - Inverted D-12 test: renamed to TestGenerateExpandedView_SkipsExternalBoundaryNodesForLinkedUnits, asserts NotContains cloudstorage
- `internal/graph/builder_test.go` - New TestBuildGraphResolvedEdgeMinLen (2 subtests) and TestBuildExpandedGraphBaselineDOT (real-TOML COMPAT-02 guard)

## Decisions Made

- Followed the plan exactly: D-12 delete (keeping createExternalBoundaryNode), D-02 AllExpanded literal, D-13 gates with the plan's exact operands (C1: `path == sourceAncestor && link.Peer == resolved`; boundary: `link.Peer == resolvedPeer`; cross-link: `originalSource == sourcePath && link.Peer == resolvedPeer`).
- resolveDescendantCrossLinks passes `fullPath` (the current recursion path) as originalSource — since a descendant's path never equals the drawn sourcePath (the subunit entry), descendant-authored links always drop length, as D-13 requires.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Lint] Four new golangci-lint issues on new code**
- **Found during:** Task 3 (lint gate)
- **Issue:** `golangci-lint run --new-from-rev` reported 4 new issues: funlen on TestBuildGraphResolvedEdgeMinLen, testifylint float-compare (assert.Equal on float64 penwidth), and 2x wsl_v5 "missing whitespace above if" in resolveBoundaryNodeLinks (the re-written `if len(resolved) > 0` blocks).
- **Fix:** Added `//nolint:funlen` with explanation comment, switched to `assert.InDelta(t, 2.0, edge.PenWidth, 0.001, ...)`, added blank lines before the two `if len(resolved) > 0` statements.
- **Files modified:** internal/graph/builder_test.go, internal/view/scope.go
- **Verification:** `golangci-lint run --new-from-rev=c44781d ./...` reports 0 issues; full suite green with -race.
- **Committed in:** 1fe07e3 (Task 3 commit)

**Note on the plan's grep gate:** `grep -c "addExternalBoundaryNodes"` on scope.go returns 4, not 0 — all four hits are `addExternalBoundaryNodesForSubunits`, the C2/C3 helper explicitly NOT removed by D-12 (its name is a superset of the deleted functions' prefix). The exact-name gate (`addExternalBoundaryNodes(`, `addExternalBoundaryNodesRecursive`, the call) is 0. The acceptance criterion's intent — both deleted functions and their call gone — is fully met.

---

**Total deviations:** 1 auto-fixed (lint-gate, 4 issues)
**Impact on plan:** Necessary for the lint gate to pass; no scope creep, no behavior change.

## Issues Encountered

- None — the TDD cycle behaved exactly as designed: all three RED tests failed for the reasons predicted in the plan (cloudstorage still added, MinLen 2 on resolved edge, PenWidth 0 on expanded single edges), and GREEN flipped them all.
- Pre-existing lint config noise observed (gomodguard deprecation warning, unknown `deadcode` nolint directive) — untouched per scope boundary.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 01-03 (visible-subunit resolution, D-07..D-11) can now build on the D-13 gate: its deepest-visible resolution will replace the `path == sourceAncestor` first operand in resolveAndAddBoundary with a source-visible check; the `link.Peer == resolved` second operand stays.
- The AllExpanded activation means expanded-mode fixtures in future plans must expect v1.7 dedup key behavior (penwidth 2.0 everywhere, tech+desc key) — already locked in by TestBuildExpandedGraphBaselineDOT.
- deferred-items.md: the 01-01 "AllExpanded not set" deferred item is now resolved by this plan (D-02). Pre-existing golangci-lint debt (121 issues) remains deferred.

## Self-Check: PASSED

- Files: 01-02-SUMMARY.md exists; scope.go, scope_test.go, builder_test.go modified as planned
- Commits: 4c8406d (test RED), 5132df9 (feat GREEN), 1fe07e3 (refactor) all present
- `go test -v -race ./...` green; `golangci-lint run --new-from-rev=c44781d ./...` reports 0 new issues
- TDD gate sequence verified in git log: test → feat → refactor

---
*Phase: 01-fix-c1-view-scoping*
*Completed: 2026-08-06*
