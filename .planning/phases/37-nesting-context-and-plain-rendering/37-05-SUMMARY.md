---
phase: 37-nesting-context-and-plain-rendering
plan: 05
subsystem: testing
tags: [goldens, canonical, invariant, bc, go]

# Dependency graph
requires:
  - phase: 37-nesting-context-and-plain-rendering
    provides: CTX-01..03 nesting-context rendering (37-01/37-02), PLAIN-01..04 plain rendering with plain.dot/plain.expanded.dot goldens (37-03/37-04)
provides:
  - "CTX-01 enforced by standing invariant test TestViewAncestorChainInvariant over the 4-level multilevel fixture"
  - "BC-01 proven: zero golden re-baseline required — the only test-consumed pre-phase goldens (multilevel.expanded.dot) are byte-identical to current output; full go test ./... green"
  - "Visual checkpoint evidence for nesting-context + --plain rendering (auto-approved in auto-mode)"
affects: [37-06, 37-07, verifier]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Ancestor-closure invariant over View.Units/VisiblePaths as a standing CTX-01 regression guard (boundary/external nodes excluded per RESEARCH scoping)"

key-files:
  created:
    - ".planning/phases/37-nesting-context-and-plain-rendering/deferred-items.md"
  modified: []

key-decisions:
  - "No golden re-baseline needed: waves 1-4 calibration held — multilevel.expanded.dot (only test-compared golden) is byte-identical; expanded pipeline (buildNestedCluster) untouched by CTX-03, and no committed golden captures the non-expanded C1 output"
  - "Stale Milestone-v1.10 artifacts expanded.dot / expanded/mainsystem.dot NOT re-baselined — not consumed by any test, drift predates phase 37, cannot be attributed to documented CTX deltas; logged to deferred-items.md"

patterns-established:
  - "Golden audit pattern: render fixture to tmp dir with exact production flags, byte-compare test-consumed goldens, attribute any diff to documented deltas before re-baselining"

requirements-completed: [CTX-01, BC-01]

# Metrics
duration: 12min
completed: 2026-08-30
---

# Phase 37 Plan 05: Consolidated golden re-baseline gate Summary

**CTX-01 locked by a standing ancestor-chain invariant test; BC-01 proven with a zero-delta golden audit — nothing needed re-baselining and the full suite is green.**

## Performance

- **Duration:** 12 min
- **Started:** 2026-08-30T16:50:00Z
- **Completed:** 2026-08-30T17:02:00Z
- **Tasks:** 3 (Task 1 completed by prior executor; Tasks 2-3 this session)
- **Files modified:** 0 source files (audit-only plan; no golden churn invented)

## Accomplishments
- Verified Task 1's `TestViewAncestorChainInvariant` (commit 930f5c8) passes against post-CTX code — every depicted element on all non-expanded multilevel views sits inside its complete ancestor chain, with boundary/external nodes excluded per the RESEARCH scoping decision.
- Executed the ONE consolidated re-baseline audit: full `go test -count=1 ./...` is fully green (15/15 packages, zero failures); `multilevel.expanded.dot` (the ONLY test-compared golden, consumed by TestBuildExpandedGraphBaselineDOT + REF-05) is byte-identical to a fresh render; `plain.dot` / `plain.expanded.dot` are byte-identical under `--plain` / `--plain --expanded`. No re-baseline required — that IS the correct BC-01 outcome, not invented churn.
- Task 3 visual checkpoint resolved in auto-mode with recorded structural evidence (below).

## Task Commits

1. **Task 1: CTX-01 ancestor-chain invariant test** - `930f5c8` (test) — completed by prior executor, verified passing this session
2. **Task 2: ONE consolidated golden re-baseline (BC-01)** - no commit: the audit found NOTHING to re-baseline (all test-consumed goldens byte-identical; suite green). Committing nothing is the honest outcome — inventing golden churn would violate the plan's own "no unrelated churn" rule.
3. **Task 3: Visual approval checkpoint** - no file changes (auto-approved, evidence below)

**Plan metadata:** (docs commit follows this SUMMARY)

## Files Created/Modified
- `internal/view/scope_test.go` - TestViewAncestorChainInvariant (Task 1, commit 930f5c8)
- `.planning/phases/37-nesting-context-and-plain-rendering/deferred-items.md` - out-of-scope discovery record (stale v1.10 reference artifacts)

## Task 2 golden audit detail (BC-01 evidence)

| Golden | Consumed by | Fresh render vs committed | Verdict |
|---|---|---|---|
| `cmd/c4drill/testdata/multilevel.expanded.dot` | internal/graph TestBuildExpandedGraphBaselineDOT + REF-05 baseline | byte-identical | no re-baseline |
| `cmd/c4drill/testdata/plain.dot` | cmd/c4drill canonical golden test | byte-identical (`--plain`) | no re-baseline |
| `cmd/c4drill/testdata/plain.expanded.dot` | cmd/c4drill canonical golden test | byte-identical (`--plain --expanded`) | no re-baseline |
| `cmd/c4drill/testdata/multilevel.dot` | nothing (does not exist committed) | n/a — plan frontmatter listed it speculatively | nothing to do |
| `cmd/c4drill/testdata/expanded.dot`, `expanded/mainsystem.dot` | nothing (reference artifacts only) | stale since Milestone v1.10 (2f21325), drift predates phase 37 | deferred, not re-baselined |

Why multilevel.expanded.dot never broke: expanded mode renders via `buildNestedCluster` (untouched by CTX-03's non-expanded cluster paths), and no committed golden captures the non-expanded C1 output that CTX-02 changed — root_test.go renders into tmp dirs and only stats generated files.

## Task 3 checkpoint — auto-approved (auto-mode)

⚡ Auto-approved checkpoint (auto-mode). Structural evidence from rendered SVGs:

- **Nested clusters:** C1 `multilevel.svg` and drill-down `mainSystem.svg` each contain 17 nested `cluster_*` groups — containers render as nested clusters, not flat lists (CTX-03).
- **Magnifier affordances:** 15 `🔍` markers in both C1 and C2 — collapsed containers keep the unfold affordance.
- **Cluster explore/drill links:** C1 hrefs to every drill-down SVG including deep targets (e.g. `multilevel/mainSystem/storages/externalStorage/varlinkAPIs.svg`) — deep-link chains intact (CTX-02); drill-down hrefs mirror the same subtree, confirming C2 nesting matches C1.
- **Plain labels:** `plain.svg` contains zero lowercase HTML-table label artifacts; labels are readable record text (`{Order Context|Event Sourcing|Orders bounded context}`) and edge labels like `[HTTPS] Manages the orders platform`.
- **Plain colours/legend:** fixture custom colours (`#1565C0`, `#FFF9C4`) do NOT leak; kind-derived colours (`#073b6f`, `#2e7d32`) and the legend are present.
- **Layout:** no stretched minlen gaps or reversed edges observed under plain.

## Decisions Made
- Recorded "no re-baseline needed" as the outcome instead of regenerating identical files — the plan's must-have ("every golden diff corresponds to a documented CTX delta") is trivially satisfied by an empty diff set.
- Deferred the stale v1.10 artifacts rather than re-baselining: their drift spans many pre-37 milestones and cannot be attributed to CTX deltas (plan Task 2 step 3 forbids absorbing unrelated diffs).

## Deviations from Plan

None - plan executed exactly as written. (The plan's anticipated re-baseline targets turned out byte-identical, which the plan itself frames as the correct BC-01 outcome.)

## Issues Encountered
- Pre-existing stale reference artifacts `cmd/c4drill/testdata/expanded.dot` + `expanded/mainsystem.dot` (last commit: Milestone v1.10) no longer match `expanded.toml` output. Not consumed by any test; out of scope per the executor scope boundary — logged to `deferred-items.md` in the phase directory.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Phase 37 gate closed: invariant test standing, goldens proven current, suite fully green.
- Remaining plans 37-06/37-07 can proceed on the verified rendering contract.

---
*Phase: 37-nesting-context-and-plain-rendering*
*Completed: 2026-08-30*
