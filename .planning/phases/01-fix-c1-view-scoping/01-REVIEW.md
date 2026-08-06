---
phase: 01-fix-c1-view-scoping
reviewed: 2026-08-06T09:54:52Z
depth: standard
files_reviewed: 8
files_reviewed_list:
  - internal/view/view.go
  - internal/graph/graph.go
  - internal/graph/builder.go
  - internal/render/converter.go
  - internal/graph/builder_test.go
  - internal/view/scope.go
  - internal/view/scope_test.go
  - internal/view/integration_test.go
findings:
  critical: 0
  warning: 2
  info: 6
  total: 8
status: issues_found
---

# Phase 1: Code Review Report — Fix C1 View Scoping

**Reviewed:** 2026-08-06T09:54:52Z
**Depth:** standard
**Files Reviewed:** 8
**Status:** issues_found

## Summary

Reviewed the D-01..D-13 refinement over the v1.8 baseline (`b6771b0..HEAD`): pair-only edge dedup with binary penwidth, `--expanded` exemption (`AllExpanded`), mirror-aware multiplicity counting, legacy `addExternalBoundaryNodes*` removal, D-13 minlen gating, deepest-visible-ancestor resolution (`resolveToViewAncestor` unification + `VisiblePaths`), within-cluster edges, and box parity.

Verification performed: full test suite and `-race` pass for `internal/view`, `internal/graph`, `internal/render`; `go vet` clean; probe tests written, executed, and removed to confirm the two Warning findings empirically; golangci-lint run at baseline and HEAD (lint debt decreased overall, 353 -> 344 issues; no new blocker-level lint regressions in changed files; new `//nolint` directives carry explanations per convention).

Confirmed-correct behavior (spot checks): D-07/D-09/D-10 resolution in C1 (webUser -> sshAuth, sshAuth -> keycloak, within-cluster sshAuth -> webAPI, D-08 no parent edge), D-13 minlen gating in all four synthesis paths, pair-only dedup first-wins, `--expanded` tech+desc dedup with 2.0 penwidth, C1 boundary-node pollution guard (5-node C1), nil-map safety (`VisiblePaths`, `pairCounts`), and definition-order determinism. Render tests using `t.Parallel()` (e.g., `TestBuildEdgesPenwidthRendered`, `TestBuildGraphExpandedC1VisibleSubunitEdges`) are safe: all rendering is serialized by `wasmMutex` in `internal/render/render.go` (confirmed with `-race`).

The two Warning findings are deviations from locked decisions D-04/D-05: the penwidth-thickening signal is silently absent in C2/C3 views and for authored `linkFrom` relationships, because multiplicity is counted from already-deduped resolved link slices and a mirror-discrimination heuristic that cannot distinguish validator mirrors from authored links.

## Warnings

### WR-01: D-04/D-05 penwidth thickening never fires in C2/C3 views

**File:** `internal/view/scope.go:836-840` (and `873-877`)
**Issue:** D-05 requires multiplicity to count ALL contributing links (direct + resolved sub-links) landing on a pair, and D-04 (which D-02 scopes to all resolved views C1/C2/C3) requires collapsed pairs (2+) to thicken to penwidth 2.0. In C1 this works because `resolveAndAddBoundary` appends every contributing link to `ResolvedLinks` without dedup. In C2/C3, `addResolvedCrossLink`/`addResolvedCrossLinkFrom` dedup by resolved peer (`if existing.Peer == resolvedPeer { return }`) BEFORE `countPairMultiplicity` ever sees the links, so a pair with two direct links (`api -> db` x2) or two descendant-contributed links collapses to a single `ResolvedLinks` entry and `pairCounts` records 1. Empirically confirmed with a probe: a C2 view with two `api -> db` links yields one collapsed edge with `PenWidth == 0` (renders 1.0), while the identical authoring in C1 yields 2.0. The C2/C3 paths were intentionally not modified, but the new D-04 feature silently does not apply there — an inconsistency with the locked decisions and with C1 output for identical input.
**Fix:** Count multiplicity from the pre-dedup link sets (`Unit.Links`/`Unit.LinksFrom`) in `countPairMultiplicity` instead of the deduped `ResolvedLinks`/`ResolvedLinksFrom`, or perform the pair count inside `resolveSubunitCrossLinks`/`resolveDescendantCrossLinks` before the peer-dedup, and store it on the entry (e.g., a per-entry `PairCounts` map consumed by `buildEdges`).

### WR-02: Mirror-discrimination heuristic undercounts authored `linkFrom` relationships

**File:** `internal/graph/builder.go:430-432` (`countIncomingPairs`)
**Issue:** `if pairCounts[key] == 0 { pairCounts[key]++ }` assumes any incoming entry on a pair that already has an outgoing contribution is a validator mirror (`internal/validator/index.go` dedups mirrors per pair with `FindLinkByPeer`). But the TOML schema supports authored `linkFrom` entries, and `populateIncomingLinks` skips the mirror when one already exists — so an authored second relationship on an already-outgoing pair is indistinguishable from a mirror and is discarded from the count. D-05 says count ALL contributing links; the count becomes authoring-dependent: two links both authored on the source thicken to 2.0, but one link on the source plus one authored `linkFrom` on the target renders 1.0. Empirically confirmed with a probe (PenWidth 0 for the mixed authoring). Consequence: the multiplicity signal (the only D-06 multiplicity signal) depends on where in the model the relationships were authored.
**Fix:** Count incoming contributions unconditionally and subtract only actual mirrors (e.g., count both directions, then subtract 1 per pair that has both an outgoing link and a `LinksFrom` entry whose `Peer`/attributes exactly match the validator's mirror shape), or have the validator mark mirrored links so they can be excluded reliably.

## Info

### IN-01: Phantom edges for parent<->visible-child links in unvalidated views

**File:** `internal/view/scope.go:274` (guard `resolved != resolvedSource`)
**Issue:** The D-09 source-side resolution narrowed the internal-link detection: a link from an expanded top-level unit to its own visible subunit, or from a hidden descendant to its own top-level ancestor, now satisfies `resolved != resolvedSource` and is appended as a resolved link whose source or target is rendered as a CLUSTER, not a node. `createEdges` (`internal/render/converter.go:137-139`) then silently drops the edge (nodeMap lookup fails), so it remains as a dead entry in `g.Edges`. Empirically confirmed with a probe (1 phantom edge). Unreachable for validated models (VALD-02/VALD-03 in `internal/validator/rules.go` forbid links on/at units with subunits), but hand-built or unvalidated views hit it.
**Fix:** Skip the resolved-link append when the resolved peer's nearest visible ancestor equals the source's resolved ancestor (i.e., `resolveToViewAncestor(v, resolved) == resolvedSource`), restoring the pre-phase internal-link detection.

### IN-02: Stale comment in `TestBuildEdgesExpandedExemption`

**File:** `internal/graph/builder_test.go:1595-1596`
**Issue:** The comment claims "GenerateExpandedView does not set AllExpanded yet (plan 01-02), so the view is constructed literally." `GenerateExpandedView` now sets `AllExpanded: true` (`internal/view/scope.go:22`); the literal-view construction is redundant and the comment misleads future readers about why the fixture is hand-built.
**Fix:** Update the comment to state the fixture is hand-built to isolate the exemption behavior, or build it via `GenerateExpandedView`.

### IN-03: Misleading `PenWidth` doc comment on `Edge`

**File:** `internal/graph/graph.go:89-91`
**Issue:** "0 means the renderer applies the default (1.0 in resolved views, 2.0 in --expanded mode)" — the renderer always applies 1.0 for `PenWidth == 0` (`internal/render/converter.go:481-485`); expanded mode never emits 0 because the builder forces 2.0 there. The comment promises behavior the renderer does not implement.
**Fix:** Reword: "0 means the renderer applies the default 1.0; the builder sets 2.0 for collapsed pairs (D-04) and for all --expanded edges."

### IN-04: Dead code and odd identifier in `joinLabels`

**File:** `internal/render/converter.go:199-215`
**Issue:** `result := ""` followed by `result += resultSb120.String()` is dead accumulation (result is only read at the return), and the builder is named `resultSb120` for no reason. Pre-existing, but the function sits in a modified file and violates the project's lint conventions.
**Fix:** Return `resultSb120.String()` directly and rename to `result`.

### IN-05: Dead `resolved == ""` guard and misleading comment

**File:** `internal/view/scope.go:262-264`
**Issue:** `resolveToTopLevel` never returns `""` (its fallback returns the peer as-is), so `if resolved == "" { continue // Internal link ... }` is unreachable and the comment misdescribes when the internal-link skip actually happens (it happens via `resolved != resolvedSource` below).
**Fix:** Remove the guard, or make `resolveToTopLevel` return `""` for the no-ancestor case and rely on the guard.

### IN-06: Complexity growth in modified functions without `//nolint`

**File:** `internal/view/scope.go:85, 251`
**Issue:** `GenerateC1View` (gocognit 21) and `resolveAndAddBoundary` (gocognit 31, up from 24 at baseline) grew beyond the configured threshold in this phase. The baseline already carries lint debt, but new complexity in modified code should carry an explanatory `//nolint:gocognit` per project convention, and the duplicated outgoing/incoming resolution blocks in `resolveAndAddBoundary` (and `resolveBoundaryNodeLinks`) are candidates for a shared helper.
**Fix:** Add `//nolint:gocognit` with explanation, or extract the resolved-link construction (including the D-13 length gate) into a helper used by all four synthesis paths.

---

_Reviewed: 2026-08-06T09:54:52Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
