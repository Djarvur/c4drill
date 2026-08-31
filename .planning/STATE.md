---
gsd_state_version: 1.0
milestone: v1.15
milestone_name: Hierarchy Wrapping and Granular Keys
status: Awaiting next milestone
last_updated: "2026-08-31T05:55:58.823Z"
last_activity: 2026-08-31 — Milestone v1.15 completed and archived
progress:
  total_phases: 2
  completed_phases: 1
  total_plans: 6
  completed_plans: 6
  percent: 50
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-08-30)

**Core value:** Transform simple TOML architecture descriptions into professional C4 diagrams without manual drawing.
**Current focus:** Phase 38 — hierarchy-wrapping-and-granular-keys

## Current Position

Phase: Milestone v1.15 complete
Plan: —
Status: Awaiting next milestone
Last activity: 2026-08-31 — Milestone v1.15 completed and archived

## Performance Metrics

**Velocity (carry-forward):**

- v1.11: 1 phase (34), 4 plans. v1.12: 1 phase (35), 9 plans. v1.13: 1 phase (36), 6 plans. v1.14: 1 phase (37), 7 plans.
- Expect Phase 38 at mid single-digit plans (plan-phase decides).

| Phase | Plans | Notes |
|-------|-------|-------|
| 38 | 38-01: 3 tasks, ~35m. 38-02: 3 tasks, ~30m. 38-03: 3 tasks, ~30m. 38-04: 3 tasks, ~35m. 38-05: 2 tasks, ~30m. 38-06: 3 tasks, ~10m | TDD RED→GREEN per plan; 38-04 = matrix + goldens + visual checkpoint; 38-05 = docs + sync + fixture; 38-06 = release v1.22.0 |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- [v1.15 start]: v1.14's scoping decision (boundary/sibling entries top-level) is REVERSED — they must render inside ancestor container chains; fully external entries stay top-level.
- [v1.15 start]: v1.14's deferred-items entry for granular flags is superseded — granular CLI switches are now in scope (CLI-only confirmed by user 2026-08-30).
- [v1.15 start]: single-phase milestone (precedent: v1.11–v1.14; shared packages + goldens).
- [38-01]: wrapper cluster IDs namespaced `wrap_<dotted path>` (T-38-01); dots kept — graphviz quotes them, dot(1) validates.
- [38-01]: boundary-chain prefix equal to ExpandedUnit maps onto the boundary cluster (no duplicate wrapper) — C2 sibling boundaries land inside the expanded unit's cluster.
- [38-02]: kind colours survive --plain (v1.14 golden-pinned), so --no-colors suppresses kind colouring only when plain is unset — --plain stays the exact union (KEY-02, TestPlainUnionParity).
- [38-02]: D-01 source-border default edge colour is structural and survives --no-colors; converter plain call-tree unchanged (only buildCgraph reads g.Opts.Plain).
- [38-01]: committed goldens cover C1/expanded only and cmd E2E asserts C2/C3 via contains-checks → 38-04 golden re-baseline is EMPTY (see 38-01-SUMMARY census).
- [38-03]: --no-labels suppresses at the GRAPH layer (builder drops Label content; converter empty-label emission is defense-in-depth); legend stays per LBL-03 pin; census stays EMPTY for 38-04.
- [38-04]: BC-01 re-baseline verified as NO-OP — zero committed-golden hunks (goldens cover C1/expanded only; WRAP is C2/C3-only; switches opt-in); additive nolabels.dot/nolabels.expanded.dot goldens committed and canonically pinned.
- [38-04]: KEY-03 matrix locked E2E: every switch × C1/drill-down/--expanded × dot/svg/html + --plain/--no-colors compositions; structural dot markers asserted on RAW dot (uppercase sanctioned markup), hexes lowercased.
- [38-05]: docs use the real release tag v1.22 (git tag convention; "v1.15" is GSD milestone numbering only); 13-wrapping.toml ships without a .c4d twin per the 12-plain precedent, expectedExampleTwins manifest untouched.
- [38-06]: release checkpoint auto-selected tag-now (AUTO mode; REL-01 roadmap-approved; v1.21.0 precedent); version stays ldflags-injected — no constant bump.
- [38-06]: pre-existing Validate Examples CI failure (plugin trees track 11-nesting-context drill-down SVGs, skill/ gitignores them) logged to deferred-items — predates phase 38; release workflow is the gate and is green.

### Pending Todos

See .planning/todos/pending/. (1 pending: add CLI flag to override edge routing style — feature request, 2026-08-30)

### Blockers/Concerns

- [Phase 38]: WRAP will cause REAL golden re-baselining for models with cross-container links (unlike v1.14's zero-delta outcome) — budget for documented delta churn.
- [Phase 38 — planner must pin]: kind-derived colours / legend coverage under the colours switch; legend behavior under `--no-labels` (default: legend stays — metadata, not an element label).

### Quick Tasks Completed

| # | Description | Date | Commit | Status | Directory |
|---|-------------|------|--------|--------|-----------|
| 260831-01u | Fix three rendering bugs from todos/pending: root diagram bloated by ancestor wrapping (bisect: ships in v1.21.0 CTX-02/CTX-03, not b2447da), --no-labels narrowed to edge labels only, edge merge made flag-invariant via builder-assigned Edge.Name | 2026-08-30 | 72afbbb | Needs Review (2 human items) | [260831-01u-fix-three-rendering-bugs-from-todos-pend](./quick/260831-01u-fix-three-rendering-bugs-from-todos-pend/) |

## Deferred Items

| Category | Item | Status | Deferred At |
|----------|------|--------|-------------|
| *(granular flags item superseded — in scope for v1.15)* | | | |
| bug | CI Validate Examples asymmetry (11-nesting-context drill-down SVGs tracked in plugin trees, gitignored in skill/) | RESOLVED 2026-08-30 — skill/ force-adds the 5 SVGs per 06-templates precedent (commit 66cd6dc); Validate Examples green | 38-06 |
| debug | docs-drift-orphan-rule-testdata — doc/fixture drift around orphan rule VAL-01 | acknowledged open at v1.15 close (2026-08-31) | v1.15 close |
| debug | knowledge-base — stale note, status unknown | acknowledged open at v1.15 close (2026-08-31) | v1.15 close |
| todo | add CLI flag to override edge routing style (--edges) — feature request | acknowledged open at v1.15 close (2026-08-31); candidate for next milestone/quick | v1.15 close |

## Session Continuity

Last session: 2026-08-30 (plan 38-06 execution — milestone v1.15 complete)
Stopped at: Milestone v1.15 shipped as v1.22.0 — milestone ready for /gsd:complete-milestone
Resume file: .planning/phases/38-hierarchy-wrapping-and-granular-keys/38-06-SUMMARY.md

## Operator Next Steps

- Start the next milestone with /gsd-new-milestone
