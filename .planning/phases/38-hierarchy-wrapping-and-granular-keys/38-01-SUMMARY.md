---
phase: 38-hierarchy-wrapping-and-granular-keys
plan: 01
subsystem: graph-builder
tags: [wrapping, boundary, clusters, c2-c3, tdd]
requires: []
provides:
  - "ancestor wrapper-cluster construction (ensureWrapperCluster) in internal/graph/builder.go"
  - "WRAP-03 node-set / edge-endpoint invariance locks over the multilevel fixture"
  - "expected-failure census for 38-04 (EMPTY — see census section)"
affects: [38-02-granular-keys, 38-04-rebaseline]
tech-stack:
  added: []
  patterns:
    - "wrapper-cluster registry map keyed by dotted path (merge-by-path)"
    - "'wrap_' + path namespaced cluster IDs (T-38-01 mitigation)"
key-files:
  created:
    - internal/graph/wrap_smoke_test.go
  modified:
    - internal/graph/builder.go
    - internal/graph/builder_test.go
decisions:
  - "Wrapper cluster ID = 'wrap_' + dotted path (dots kept; graphviz quotes them; validated with dot(1))"
  - "A boundary chain prefix equal to ExpandedUnit maps onto the expanded unit's boundary cluster instead of a duplicate wrapper"
  - "Fully external = single-segment path; stays top-level (empirically matches resolveBoundaryByDivergence output)"
  - "Wrapper labels from AncestorNames with raw-segment fallback; no icon/technology/description"
metrics:
  duration: ~35m
  completed: 2026-08-30
---

# Phase 38 Plan 01: Hierarchy Wrapping (WRAP-01/02/03) Summary

Ancestor wrapper clusters now enclose the expanded unit and every in-model-ancestor boundary/sibling entry on all C2/C3 views; wrapping is proven cluster-structure-only by node-set and edge-endpoint invariance locks, and the full test suite is green with zero golden re-baseline needed.

## What Was Built

- **RED** (`f94d1de`): six tests over the multilevel fixture — boundary-wrapping, fully-external-stays-top-level, expanded-unit ancestor skeleton (pretty names), wrapper-no-ExploreURL, node-set invariance (WRAP-03), edge-endpoint lock. Two failed for the wrapping reason; the fully-external and invariance locks were green from the start (behavior to preserve, as the plan anticipated).
- **GREEN** (`b2447da`): `buildBoundaryViewGraph` builds the expanded unit's root→parent wrapper skeleton (`ensureWrapperCluster`), nests `buildBoundaryCluster` innermost, and wraps each boundary/external entry in its complete ancestor chain, merging into wrapper clusters by path. Fully external (single-segment) entries stay in `g.Nodes`. Wrapper IDs are namespaced `wrap_<path>` so author unit names cannot collide with real cluster/node IDs (T-38-01); wrappers carry `AncestorNames` pretty labels and no `ExploreURL` — the CTX-03 URL walk keys on `v.Units[cluster.ID]`, which wrapper IDs never hit, so no drill affordance can leak. Node IDs, edge endpoints, dedup/multiplicity keys untouched; view-layer scope.go untouched.
- **Calibration** (`dd798d4`): smoke test rendering C2 + C3 wrapper DOT through the layout engine; `dot -Tsvg` validates the output.

## Deviations from Plan

- **[Rule 1 - Test fixture] TestFullyExternalBoundaryStaysTopLevel originally used the C3 localStorage view, where `externalSys` is not a boundary at all (no link from the localStorage subtree reaches it) — it never appears in the view.** Switched to the C2 view of `mainSystem`, where `externalStorage.client → externalSys` diverges at model root. Files: builder_test.go.
- **[Rule 2 - Test] Added `internal/graph/wrap_smoke_test.go`** (plan files listed only builder_test.go): C2/C3 wrapper DOT must render through `render.RenderDOT` without error — cheap guard against graphviz-invalid cluster names. Files: wrap_smoke_test.go; commit dd798d4.

## Expected-Failure Census for 38-04 (load-bearing)

**Census is EMPTY: zero tests fail after this plan.** `go test ./... -count=1` is fully green, including the cmd E2E goldens.

Why the research's "REAL golden failures expected" did not materialize:
1. The committed goldens (`cmd/c4drill/testdata/multilevel.dot/.svg`, `multilevel.expanded.dot`, `plain.dot`, `plain.expanded.dot`) cover the **C1 and all-expanded** paths (`buildC1ViewGraph` / `BuildExpandedGraph`) — this plan deliberately does not touch those dispatch paths.
2. The cmd E2E suite does generate C2/C3 outputs but asserts them via `Contains`/`NotContains` structural checks, not byte goldens, so added wrapper clusters change nothing it pins.
3. In-package golden comparisons (`TestBuildExpandedGraphBaselineDOT`, REF-02 rendering) likewise cover only expanded/C1 generations.

**Consequence for 38-04:** there is nothing to re-baseline. If 38-04 was planned as "consolidated golden re-baseline of WRAP deltas", it should be reduced to: (a) optionally adding NEW positive goldens capturing the wrapper structure (e.g. a C3 dot golden for `mainSystem.storages.localStorage`), and (b) the visual spot-check. Verified evidence: wrapper nesting appears in rendered DOT as `cluster_wrap_mainSystem ⊃ cluster_wrap_mainSystem.storages ⊃ cluster_mainSystem.storages.localStorage` and `dot(1)` accepts it.

## TDD Gate Compliance

- `test(38-01)` RED commit: f94d1de (tests fail for wrapping reasons, not compile errors)
- `feat(38-01)` GREEN commit: b2447da
- No refactor commit needed.

## Decisions Made

- Wrapper ID keeps dots (`wrap_mainSystem.storages`); graphviz quotes the subgraph name and `dot -Tsvg` validates — simpler than underscore mangling and keeps the merge-by-path mapping obvious.
- Boundary-chain prefix == ExpandedUnit maps onto `boundaryCluster` (that cluster IS the depiction of that ancestor), so C2 sibling boundaries land inside the expanded unit's cluster and no duplicate `wrap_mainSystem` skeleton is created for C2 views.
- "Fully external" implemented as single-segment FullPath — empirically exact: `resolveBoundaryByDivergence` collapses common==0 peers to their top-level segment.

## Self-Check: PASSED

- Files exist: internal/graph/builder.go, internal/graph/builder_test.go, internal/graph/wrap_smoke_test.go — FOUND
- Commits: f94d1de, b2447da, dd798d4 — FOUND in `git log`
- `go test ./... -count=1` — all packages ok

## Threat Flags

None — no new network endpoints, auth paths, or file access; T-38-01 mitigated via `wrap_` namespacing and asserted by the node-set invariance test.
