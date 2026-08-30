---
phase: 37-nesting-context-and-plain-rendering
verified: 2026-08-30T16:00:00Z
status: passed
score: 12/12 requirements verified (5/5 roadmap success criteria)
overrides_applied: 0
gaps: []
deferred:
  - truth: "Stale v1.10 reference artifacts cmd/c4drill/testdata/expanded.dot and expanded/mainsystem.dot do not match current output"
    addressed_in: "Deferred (out of scope, pre-existing)"
    evidence: "deferred-items.md: not consumed by any test; drift predates phase 37 (last touched in v1.10 commit 2f21325); BC-01 delta discipline forbids re-baselining undocumented deltas"
---

# Phase 37: Nesting Context and Plain Rendering — Verification Report

**Phase Goal:** Non-expanded diagrams preserve the full nesting picture; `--plain` renders canonical default-styled output; ships as v1.21.0.
**Verified:** 2026-08-30
**Status:** PASSED
**Re-verification:** No — initial verification

## Goal Achievement

### Behavioral Spot-Checks (live binary, go build + run)

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Full suite green | `go test -count=1 ./...` | all 16 packages ok, exit 0 | PASS |
| `go vet` clean | `go vet ./...` | exit 0 | PASS |
| PLAIN-01/02: `--plain` suppresses custom unit fill | `c4drill 12-plain.toml --plain -f dot` | 0 `fillcolor="#` hits (default: 2) | PASS |
| CTX-01/02: deep-link target rendered in container chain | inspect `/tmp/ctx-out/11-nesting-context.dot` | `subgraph cluster_platform` nested; node `platform.gateway.router` present; edge `admin -> "platform.gateway.router"` terminates at true target (line 143) | PASS |
| CTX-03: cluster drill affordance | `grep -c "URL=" 11-nesting-context.dot` | 4 URL emissions on clusters/nodes | PASS |
| Standing invariant + key tests | `go test -run "TestViewAncestorChainInvariant|TestBuildGraph_ExpandedClusterRendersNestedSubClusters|TestBuildGraph_DeepLinkChainRendersInsideUnfoldedCluster|TestGenerateC1View_DeepLinkTargetKeepsContainerChain"` | ok (view + graph) | PASS |

### Requirements Coverage

| Requirement | Verdict | Evidence |
| ----------- | ------- | -------- |
| CTX-01 | VERIFIED | `TestViewAncestorChainInvariant` (internal/view/scope_test.go:1746) — standing ancestor-closure invariant over 4-level fixture; suite green |
| CTX-02 | VERIFIED | `ensureDeepLinkChain` + `depicted.UnfoldChain = true` (scope.go:353,422); dispatch guards `entry.IsExpanded \|\| entry.UnfoldChain` (builder.go:92,122); `TestGenerateC1View_DeepLinkTargetKeepsContainerChain`, `TestBuildGraph_DeepLinkChainRendersInsideUnfoldedCluster` pass; live dot shows edge to true target |
| CTX-03 | VERIFIED | Recursive `buildNestedCluster` (builder.go:224,242,291,700); `Cluster.ExploreURL` (graph.go:136); `assignExploreURLsToClusters` recursive walk (builder.go:1253,1278); nested-cluster test passes; live dot shows nested subgraph + URLs |
| PLAIN-01 | VERIFIED | `View.Plain` (view.go:39) → `Graph.Plain` (graph.go:44) → `applyUnitOverrides(style, unit, v.Plain)` guards (builder.go:150,251); live run: 0 custom hexes under --plain |
| PLAIN-02 | VERIFIED | `createEdge(..., v.Plain)` neutralization + `applyCollapsedPairStyle(..., v.Plain)` (builder.go:1047-1098); legend/kind colours preserved (legend SetLabelHTML path untouched) |
| PLAIN-03 | VERIFIED | converter.go plain branches: node/cluster/edge labels via `SetLabel`/record path (converter.go:414-418, 521-524, 626-643), never HTML under Plain |
| PLAIN-04 | VERIFIED | `v.Plain = plain` in both processView and processExpandedView incl. `--plain --expanded` (root.go:317,364); goldens `plain.dot`/`plain.expanded.dot` canonical-compared in cmd tests |
| BC-01 | VERIFIED | 37-05: zero golden re-baseline — `multilevel.expanded.dot` byte-identical; full `go test ./...` green |
| DOC-01 | VERIFIED | README.adoc: `--plain` documented (lines 1232-1263: what's ignored, what stays, examples) |
| DOC-02 | VERIFIED | `diff -r skill plugins/...{skills,opencode}` — zero differing tracked files (only gitignored generated `*.dot` artifacts exist solely in skill/, per `.gitignore:71`; convention matches v1.19/v1.20) |
| DOC-03 | VERIFIED | Fixtures `11-nesting-context` (3-level + deep links + nested cluster) and `12-plain.toml` present in skill/ AND both plugin copies; twin pinned in internal/c4d/parity_test.go:909; CI-style render sweep green per 37-06 |
| REL-01 | VERIFIED | tag `v1.21.0` exists (4691efb); GitHub release published (not draft, 2026-08-30T14:43:30Z); CI run 33317613032 conclusion=success |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| (none) | — | No TBD/FIXME/XXX/PLACEHOLDER in phase-touched files (grep exit 1) | — | — |

### Security Assessment

- All 7 plans carry `<threat_model>` STRIDE registers (ASVS L1 enforced by workflow config).
- Phase surface is CLI-only additions: a boolean flag and graph-structure changes; no new input parsing, network, or file-write surface beyond existing `-o` handling.
- `ExploreURL` on clusters is inert path text in dot output (same trust model as existing node URLs); no path traversal introduced — computed via existing `ComputeExploreURL`.
- No unresolved threats. No SECURITY.md required for this surface (consistent with phase scope decision).

### Known Accepted Deviations (all confirmed benign)

- 37-01 inert ExploreURL in RED commit — superseded by GREEN (verified live emission).
- 37-02/37-05 test updates, map-order flake fix, transient flake — suite green on `-count=1` re-run.
- 37-03 `applyCollapsedPairStyle` inert under plain — intentional style-override guard, verified threaded with `v.Plain`.
- 37-07 manifest pin of 11-nesting-context.toml — present in parity_test.go.

### Deferred Items

Stale v1.10 goldens (`cmd/c4drill/testdata/expanded.dot`, `expanded/mainsystem.dot`): pre-existing drift, consumed by no test, correctly logged in deferred-items.md rather than silently re-baselined. Suggested follow-up: regenerate or delete.

### Human Verification Required

None blocking. Visual SVG quality was covered by 37-05's auto-approved visual checkpoint (auto-mode); rendering is byte/canonical-compared by tests and confirmed via live dot inspection.

## Verdict

All 5 roadmap success criteria and all 12 requirements are observably true in the live codebase, confirmed by a green full test suite (`go test -count=1 ./...`, exit 0), targeted test runs, live CLI behavioral spot-checks, docs/skill parity diffs, and published release v1.21.0 with successful CI. Phase goal achieved.

_Verified: 2026-08-30_
_Verifier: Claude (gsd-verifier)_
