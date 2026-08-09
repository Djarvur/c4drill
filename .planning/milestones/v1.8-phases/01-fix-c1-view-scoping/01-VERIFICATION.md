---
phase: 01-fix-c1-view-scoping
verified: 2026-08-06T00:00:00Z
status: passed
score: 9/9 must-haves verified
overrides_applied: 0
---

# Phase 1: Fix C1 View Scoping Verification Report

**Phase Goal:** C1 diagram shows only top-level units; links to nested targets resolve to nearest visible ancestor; duplicate edges collapse.
**Verified:** 2026-08-06
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| #   | Truth   | Status     | Evidence       |
| --- | ------- | ---------- | -------------- |
| 1   | C1 diagram for `saira-20260320.c2.full.toml` has exactly 5 nodes (webUser, sshUser, adminUser, keycloak, linuxSystem) | ✓ VERIFIED | Ran the actual CLI (`go run ./cmd/c4drill cyp-auth-infra/saira-20260320.c2.full.toml -f dot`): parsed DOT contains exactly 5 top-level node IDs. `TestIntegrationC1ViewNoNestedBoundaryPollution` (integration_test.go:348) asserts the same shape for the mimic model. |
| 2   | All C1 edges connect only top-level units (no deeply nested targets) | ✓ VERIFIED | CLI-generated C1 DOT has 4 edges: webUser→linuxSystem, sshUser→linuxSystem, adminUser→linuxSystem, linuxSystem→keycloak — every endpoint is top-level (no dotted paths). The linuxSystem→keycloak edge is the deep authAPI→keycloak link resolved to its top-level ancestor. |
| 3   | Multiple links to subunits of the same parent collapse to a single edge between parents | ✓ VERIFIED | webUser's 2 links (authAPI, sessionAPI) → 1 edge; adminUser's 3 links (rbac, dacProxy, sessionAPI) → 1 edge (real fixture TOML confirmed 3 adminUser links). Both collapsed edges render penwidth=2 in the CLI DOT; single-link pairs render penwidth=1. `TestBuildEdgesPairCollapse` (builder_test.go:1494) asserts 1 edge with first-wins attrs. |
| 4   | Pair-only dedup key, first-wins attributes, plain labels, binary penwidth (D-01/D-03/D-04/D-06) | ✓ VERIFIED | builder.go:457/496 pair-only key `path+"->"+peer`; markSeen first-wins; createEdge copies plain tech/desc (no count suffix); penWidth 2.0 iff AllExpanded or pairCounts>=2 (builder.go:464-467, 503-506); converter.go:481-485 conditional SetPenWidth. Tests TestBuildEdgesPairCollapse/TestBuildEdgesPenwidth/TestBuildEdgesPenwidthRendered all green (uncached). |
| 5   | `--expanded` exempt: v1.7 tech+desc key, penwidth 2.0, minlen preserved (D-02, COMPAT-02) | ✓ VERIFIED | `AllExpanded: true` set by GenerateExpandedView (scope.go:22); countPairMultiplicity early-returns (builder.go:383); edgeKey appends tech+desc only when AllExpanded (builder.go:458-460, 497-499). CLI expanded DOT: 6 `minlen=`, per-edge `penwidth=2` on all edges (the single `penwidth=1.0` is go-graphviz's global default edge block, confirmed at file line 21-24). TestBuildEdgesExpandedExemption asserts 2 edges same pair in expanded mode; TestBuildExpandedGraphBaselineDOT asserts every edge PenWidth==2.0. |
| 6   | Deepest-visible-ancestor resolution for targets AND sources; within-cluster edges; no redundant parent edges; box parity (D-07..D-11) | ✓ VERIFIED | scope.go:255 `resolvedSource := resolveToViewAncestor(v, path)`; resolveToTopLevel delegates to resolveToViewAncestor (scope.go:359-366); append condition `resolved != resolvedSource` (scope.go:274, 311) both keeps within-cluster edges and suppresses parent edges; GenerateC1View populates visible subunits into v.Units + VisiblePaths (scope.go:132-165); BuildGraph skips them (builder.go:63-65). 7 tests green uncached: TestGenerateC1View_ResolvesToVisibleSubunit, _SourceResolvesToVisibleSubunit, _WithinClusterEdge, _NoRedundantParentEdge, _BoxResolutionParity, _ExpandedUnitExposesVisibleSubunits, TestBuildGraphExpandedC1VisibleSubunitEdges. |
| 7   | Legacy `addExternalBoundaryNodes` + `addExternalBoundaryNodesRecursive` deleted; validator is the single gatekeeper (D-12) | ✓ VERIFIED | grep gates: 0 matches for `addExternalBoundaryNodes(` and `addExternalBoundaryNodesRecursive` in scope.go; only the C2/C3 helper `addExternalBoundaryNodesForSubunits` remains (4 hits, correctly kept). Inverted test `TestGenerateExpandedView_SkipsExternalBoundaryNodesForLinkedUnits` (scope_test.go:825) asserts `NotContains(v.Units, "cloudstorage")` and passes. |
| 8   | minlen gated at all 6 synthesis sites — resolved edges carry no length (D-13) | ✓ VERIFIED | scope.go gates: resolveAndAddBoundary outgoing/incoming `path == resolvedSource && link.Peer == resolved` (:278, :315); resolveBoundaryNodeLinks `link.Peer == resolvedPeer` (:653, :685); addResolvedCrossLink/From `originalSource == sourcePath && link.Peer == resolvedPeer` (:843, :876). TestBuildGraphResolvedEdgeMinLen (2 subtests: resolved drops minlen / direct pair keeps it) green uncached. |
| 9   | Review fixes WR-01 (C2/C3 collapsed-pair penwidth) and WR-02 (authored linkFrom multiplicity) genuinely resolved | ✓ VERIFIED | WR-01: addResolvedCrossLink/From no longer pre-dedup by peer — every contributing link appended (scope.go:834-892), dedup left to builder markSeen; TestBuildEdgesPenwidthC2C3CollapsedPairs + TestGenerateC2View_ResolvedLinksKeepMultiplicity green. WR-02: `Link.Mirror` field `toml:"-"` (model/link.go:62-67), set by validator populateIncomingLinks (index.go:80), preserved through all 3 resolved-incoming sites (scope.go:329, 699, 890), `countIncomingPairs` counts unconditionally minus Mirror (builder.go:416-437); TestBuildEdgesPenwidthLinkFromContributions green. |
| 10  | Existing tests continue to pass; suite green | ✓ VERIFIED | `go test -count=1 ./...` green across all 8 packages (ran myself); `go test -race -count=1 ./internal/graph/ ./internal/view/` green; `go vet ./...` clean; `golangci-lint run --new-from-rev=c454ed5 ./...` reports 0 issues on new phase code. All claimed commits present in git log (0cd7991/d0af213/3adf383, 4c8406d/5132df9/1fe07e3, 000819f/40b635e/a58b6f0, 90e28d8/c828b95, 788237d/0e71b92). |

**Score:** 9/9 truths verified

### Required Artifacts

| Artifact | Expected    | Status | Details |
| -------- | ----------- | ------ | ------- |
| `internal/view/view.go` | `AllExpanded bool` + `VisiblePaths map[string]bool` with doc comments | ✓ VERIFIED | view.go:41-48 (both fields, doc comments present) |
| `internal/graph/graph.go` | `Edge.PenWidth float64` | ✓ VERIFIED | graph.go:89-91 (doc comment present; wording slightly stale per review IN-03 — cosmetic, excluded from fix scope) |
| `internal/graph/builder.go` | `countPairMultiplicity`, pair-only key, penwidth assignment, BuildGraph VisiblePaths skip | ✓ VERIFIED | builder.go:379-437 (counter), 457/496 (pair key), 464-467 (penwidth), 63-65 (skip) |
| `internal/render/converter.go` | Conditional SetPenWidth (PenWidth>0 else 1.0) | ✓ VERIFIED | converter.go:481-485 |
| `internal/model/link.go` + `internal/validator/index.go` | `Link.Mirror` marker set by populateIncomingLinks | ✓ VERIFIED | link.go:62-67, index.go:70-81 |
| `internal/view/scope.go` | AllExpanded literal, D-12 deletions, D-13 gates, visible-subunit population, unified resolution | ✓ VERIFIED | scope.go:22, 132-165, 255, 278/315/653/685/843/876, 359-366 |
| Tests | TestBuildEdges*, TestGenerateC1View_*, TestGenerateExpandedView_Skips*, TestBuildGraphResolvedEdgeMinLen, TestBuildExpandedGraphBaselineDOT, TestBuildEdgesPenwidthC2C3CollapsedPairs, TestBuildEdgesPenwidthLinkFromContributions, TestBuildGraphExpandedC1VisibleSubunitEdges | ✓ VERIFIED | All present with substantive assertions; all green uncached |

### Key Link Verification

| From | To  | Via | Status | Details |
| ---- | --- | --- | ------ | ------- |
| buildEdges (builder.go:343) | countPairMultiplicity | pairCounts computed before edge loop | ✓ WIRED | builder.go:343 |
| processOutgoing/IncomingLinks | createEdge | penWidth argument | ✓ WIRED | builder.go:469, 509 → createEdge (builder.go:531) |
| converter createEdge | graph.Edge.PenWidth | conditional SetPenWidth | ✓ WIRED | converter.go:481-485 |
| GenerateExpandedView (scope.go:22) | buildEdges AllExpanded branch | `AllExpanded: true` → tech+desc key + penwidth 2.0 | ✓ WIRED | scope.go:22 → builder.go:383, 458-460, 464-467 |
| resolveAndAddBoundary | resolveToViewAncestor | resolvedSource source-side resolution | ✓ WIRED | scope.go:255 |
| GenerateC1View | v.Units + v.VisiblePaths | visible-subunit entries before addC1BoundaryNodes | ✓ WIRED | scope.go:132-165 |
| BuildGraph C1 branch | v.VisiblePaths | skip entries rendered inside parent cluster | ✓ WIRED | builder.go:63-65 |
| validator populateIncomingLinks | countIncomingPairs | Mirror flag excluded from multiplicity | ✓ WIRED | index.go:80 → builder.go:429-431 |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| Edge.PenWidth → DOT penwidth | pairCounts | countPairMultiplicity over real model links | ✓ FLOWING | CLI saira C1 DOT: collapsed pairs render penwidth=2, singles penwidth=1; expanded DOT: all per-edge blocks penwidth=2 |
| Edge.MinLen → DOT minlen | link.Length | D-13-gated synthesis sites | ✓ FLOWING | CLI expanded DOT contains 6 `minlen=`; resolved C1 edges carry 0 (converter omits) per TestBuildGraphResolvedEdgeMinLen |
| C1 resolved edges | ResolvedLinks | resolveAndAddBoundary over all model links | ✓ FLOWING | Real-fixture CLI output shows deep links (authAPI→keycloak, sshd→nss internal) resolved to top-level edges |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| saira C1 has exactly 5 nodes | `go run ./cmd/c4drill <toml> -f dot -o /tmp/...` + node-ID parse | 5 top-level node IDs | ✓ PASS |
| All C1 edges top-level only | same output, edge parse | 4 edges, no dotted endpoints | ✓ PASS |
| Duplicates collapse + binary penwidth | same output, penwidth parse | webUser→linuxSystem penwidth=2 (2 links), adminUser→linuxSystem penwidth=2 (3 links), singles penwidth=1 | ✓ PASS |
| Expanded mode keeps minlen + 2.0 penwidth | `--expanded -f dot` + grep | 6 `minlen=`, per-edge `penwidth=2` | ✓ PASS |
| Full test suite | `go test -count=1 ./...` | all 8 packages ok | ✓ PASS |
| Race detector | `go test -race -count=1 ./internal/graph/ ./internal/view/` | ok | ✓ PASS |
| Vet | `go vet ./...` | clean | ✓ PASS |
| Lint on new code | `golangci-lint run --new-from-rev=c454ed5 ./...` | 0 issues | ✓ PASS |

### Probe Execution

No probes were declared in PLAN files and none exist under `scripts/` (no scripts/ directory). Step 7c: N/A — CLI-based behavioral spot-checks above substitute for probe discovery.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ---------- | ----------- | ------ | -------- |
| VIEW-01 | 01-02, 01-03 | C1 diagram shows only top-level units — no nested subunits appear as nodes | ✓ SATISFIED | CLI saira C1: exactly 5 top-level nodes, no dotted node IDs; TestIntegrationC1ViewNoNestedBoundaryPollution; addC1BoundaryNodes resolution prevents nested boundary pollution |
| VIEW-02 | 01-03 | Links to deeply nested targets resolve to the nearest visible ancestor in the current view level | ✓ SATISFIED | resolveToTopLevel→resolveToViewAncestor (scope.go:359-366); D-07 test asserts webUser→sshAuth (visible subunit), not linuxSystem |
| EDGE-01 | 01-03 | Edges from nested subunits to targets outside the current view resolve to the visible ancestor | ✓ SATISFIED | D-09 source-side resolution (scope.go:255); sshAuth→keycloak test; CLI: authAPI→keycloak renders as linuxSystem→keycloak |
| EDGE-02 | 01-01, 01-03 | Duplicate edges (multiple sub-links resolving to same ancestor pair) collapse into a single edge | ✓ SATISFIED | Pair-only key + first-wins (builder.go:457, 496); CLI: webUser 2 links→1 edge, adminUser 3 links→1 edge; TestBuildEdgesPairCollapse |

All 4 phase requirement IDs are claimed by plan frontmatter (01-01: EDGE-02; 01-02: VIEW-01; 01-03: VIEW-01, VIEW-02, EDGE-01, EDGE-02) and all are accounted for in REQUIREMENTS.md with Phase 1 mapping (status Complete). No orphaned requirements. COMPAT-02 is honored as a phase constraint (AllExpanded exemption verified above); REQUIREMENTS.md schedules its formal completion in Phase 3, which is not a phase-1 gap.

### Decision Coverage (D-01..D-13 + review fixes)

| Decision | Status | Evidence |
| -------- | ------ | -------- |
| D-01 pair-only dedup, first-wins | ✓ | builder.go:457, 496 + markSeen; TestBuildEdgesPairCollapse |
| D-02 --expanded exemption | ✓ | AllExpanded flag scope.go:22, builder.go:383/458-460 |
| D-03 first-wins all attrs | ✓ | createEdge copies from first link (definition-order iteration + markSeen) |
| D-04 binary penwidth 1.0/2.0 | ✓ | builder.go:464-467, 503-506; converter.go:481-485; DOT spot-checks |
| D-05 count all contributing links | ✓ | countPairMultiplicity; WR-02 Mirror exclusion |
| D-06 plain labels, no count suffix | ✓ | TestBuildEdgesPairCollapse asserts SQL/"first", no suffix |
| D-07 deepest visible ancestor (targets) | ✓ | resolveToTopLevel→resolveToViewAncestor; test green |
| D-08 no redundant parent edges | ✓ | `resolved != resolvedSource`; TestGenerateC1View_NoRedundantParentEdge |
| D-09 source-side resolution | ✓ | resolveAndAddBoundary resolvedSource; test green |
| D-10 within-cluster edges | ✓ | same condition passes; TestGenerateC1View_WithinClusterEdge |
| D-11 box parity | ✓ | TestGenerateC1View_BoxResolutionParity green; no type special-casing |
| D-12 legacy boundary removal | ✓ | grep gates 0; inverted test green |
| D-13 minlen gated at 6 sites | ✓ | scope.go:278/315/653/685/843/876; TestBuildGraphResolvedEdgeMinLen |
| WR-01 C2/C3 collapsed-pair penwidth | ✓ | pre-dedup removed; 2 tests green |
| WR-02 authored linkFrom multiplicity | ✓ | Mirror marker; test green |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| internal/graph/graph.go | 89-91 | Stale `PenWidth` doc comment ("1.0 in resolved views, 2.0 in --expanded mode") | ℹ️ Info | Review finding IN-03; explicitly excluded from review-fix scope (01-REVIEW-FIX.md fixes only WR-01/WR-02). Comment is misleading but behavior is correct. Non-blocking. |
| — | — | TBD/FIXME/XXX markers | none found | grep across all 10 phase files: 0 matches |
| — | — | Placeholder/stub returns | none found | only doc-comment mention of "placeholder" in createExternalBoundaryNode (real functionality, not a stub) |
| — | — | Pre-existing golangci-lint debt (~121 issues) | ℹ️ Info | Documented in deferred-items.md; untouched per scope boundary; `--new-from-rev` shows 0 new issues from this phase |

### Human Verification Required

None. The phase goal is fully verifiable programmatically: node/edge scoping is deterministic graph structure (verified at DOT level end-to-end via the real CLI), and the render pipeline is covered by existing render tests with RenderDOT success asserted. No external services, real-time behavior, or visual aesthetics are part of this phase's success criteria.

### Gaps Summary

No gaps found. All 10 truths verified, all 4 requirement IDs satisfied, all 13 decisions (D-01..D-13) plus both review warnings (WR-01, WR-02) implemented and verified in the actual code. The phase goal — "C1 diagram shows only top-level units; links to nested targets resolve to nearest visible ancestor; duplicate edges collapse" — is observably true in the codebase and confirmed end-to-end against the real saira fixture through the actual CLI binary.

Non-blocking notes for downstream phases:
- Review INFO findings IN-01..IN-06 (phantom edges on unvalidated views, stale test comment, stale PenWidth doc, joinLabels dead code, dead `resolved == ""` guard, gocognit nolint) remain open by design — the fix scope covered only WR-01/WR-02.
- Explore-link side effect (visible subunits with subunits now receive ExploreURLs) is a documented behavioral consequence of 01-03, asserted by TestBuildGraphExpandedC1VisibleSubunitExploreLink.

---

_Verified: 2026-08-06_
_Verifier: Claude (gsd-verifier)_
