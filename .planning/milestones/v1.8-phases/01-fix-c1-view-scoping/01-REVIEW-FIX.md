---
phase: 01-fix-c1-view-scoping
fixed_at: 2026-08-06T00:00:00Z
review_path: .planning/phases/01-fix-c1-view-scoping/01-REVIEW.md
iteration: 1
findings_in_scope: 2
fixed: 2
skipped: 0
status: all_fixed
---

# Phase 1: Code Review Fix Report

**Fixed at:** 2026-08-06
**Source review:** `.planning/phases/01-fix-c1-view-scoping/01-REVIEW.md`
**Iteration:** 1

**Summary:**
- Findings in scope: 2 (WR-01, WR-02 — WARNING scope; IN-01..IN-06 excluded per fix scope)
- Fixed: 2
- Skipped: 0

## Fixed Issues

### WR-01: D-04/D-05 penwidth thickening never fires in C2/C3 views

**Files modified:** `internal/view/scope.go`
**Test commits:** `90e28d8` (tests), implementation `c828b95`
**Applied fix:** Removed the resolved-peer pre-dedup from `addResolvedCrossLink` and `addResolvedCrossLinkFrom` (`internal/view/scope.go`). Every contributing link is now appended to `ResolvedLinks`/`ResolvedLinksFrom`, so `countPairMultiplicity` (D-05) sees the full pre-dedup set and collapsed pairs (2+) thicken to penwidth 2.0 (D-04) exactly as they already did in C1 (whose `resolveAndAddBoundary` never deduped). Edge dedup remains the builder's job: `processOutgoingLinks`/`processIncomingLinks` already use pair-only `markSeen` (D-01 first-wins), so rendered output is unchanged apart from the penwidth. C2/C3 passes' other semantics (resolution, D-13 minlen gating, first-wins attributes) are untouched. Verified by `TestBuildEdgesPenwidthC2C3CollapsedPairs` (C2 direct duplicates, C2 descendant-contributed links, C3 direct duplicates — all render 2.0) and `TestGenerateC2View_ResolvedLinksKeepMultiplicity` (both contributing links survive synthesis).

### WR-02: Mirror-discrimination heuristic undercounts authored `linkFrom` relationships

**Files modified:** `internal/model/link.go`, `internal/validator/index.go`, `internal/view/scope.go`, `internal/graph/builder.go`
**Test commits:** `788237d` (tests), implementation `0e71b92`
**Applied fix:** Replaced the guess-the-mirror heuristic with reliable mirror marking:
- `model.Link` gains a `Mirror bool` field (`toml:"-"`, never serialized) documented as validator-synthesized bookkeeping.
- `populateIncomingLinks` (`internal/validator/index.go`) marks the reverse entries it synthesizes with `Mirror: true`. The existing `FindLinkByPeer` guard still skips a mirror when an authored `linkFrom` exists, so authored entries are never marked.
- The flag is preserved through the resolved-incoming-link constructions in `internal/view/scope.go` (`resolveAndAddBoundary` incoming, `resolveBoundaryNodeLinks` incoming, `addResolvedCrossLinkFrom`).
- `countIncomingPairs` (`internal/graph/builder.go`) now counts incoming contributions unconditionally and excludes only `Mirror`-flagged entries; the old `pairCounts[key] == 0` heuristic is gone.

D-05 multiplicity now counts ALL authored contributing links regardless of where they were authored, and validator mirrors never double-count. Verified by `TestBuildEdgesPenwidthLinkFromContributions`: authored `linkFrom` on an already-outgoing pair thickens to 2.0; a validated bidirectional pair with mirrors only stays at default width.

---

**Verification performed:** `go test ./...` green (incl. `-race` on `internal/view`, `internal/graph`, `internal/validator`); `golangci-lint run --new-from-rev=262d6b9 ./...` reports 0 new issues. Note: two pre-existing tests (`TestBuildExpandedGraphRealToml`, `TestBuildExpandedGraphBaselineDOT`) read the gitignored private fixture `cyp-auth-infra/` and fail only in checkouts lacking it (environmental, unrelated to these fixes).

_Fixed: 2026-08-06_
_Fixer: Claude (gsd-code-fixer)_
_Iteration: 1_
