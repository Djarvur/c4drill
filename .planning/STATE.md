---
gsd_state_version: 1.0
milestone: v1.15
milestone_name: Hierarchy Wrapping and Granular Keys
status: executing
last_updated: "2026-08-30T16:01:53.228Z"
last_activity: 2026-08-30
progress:
  total_phases: 2
  completed_phases: 0
  total_plans: 6
  completed_plans: 3
  percent: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-08-30)

**Core value:** Transform simple TOML architecture descriptions into professional C4 diagrams without manual drawing.
**Current focus:** Phase 38 — hierarchy-wrapping-and-granular-keys

## Current Position

Phase: 38 (hierarchy-wrapping-and-granular-keys) — EXECUTING
Plan: 5 of 6
Status: 38-04 complete — KEY-03/BC-01 done, census-empty re-baseline verified, nolabels goldens committed
Last activity: 2026-08-30 -- Plan 38-04 executed (commits 11b15e1, 719388e, e42d278)

Progress: [██████░░░░] 60%

## Performance Metrics

**Velocity (carry-forward):**

- v1.11: 1 phase (34), 4 plans. v1.12: 1 phase (35), 9 plans. v1.13: 1 phase (36), 6 plans. v1.14: 1 phase (37), 7 plans.
- Expect Phase 38 at mid single-digit plans (plan-phase decides).

| Phase | Plans | Notes |
|-------|-------|-------|
| 38 | 38-01: 3 tasks, ~35m. 38-02: 3 tasks, ~30m. 38-03: 3 tasks, ~30m. 38-04: 3 tasks, ~35m | TDD RED→GREEN per plan; 38-04 = matrix + goldens + visual checkpoint |

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

### Pending Todos

See .planning/todos/pending/.

### Blockers/Concerns

- [Phase 38]: WRAP will cause REAL golden re-baselining for models with cross-container links (unlike v1.14's zero-delta outcome) — budget for documented delta churn.
- [Phase 38 — planner must pin]: kind-derived colours / legend coverage under the colours switch; legend behavior under `--no-labels` (default: legend stays — metadata, not an element label).

## Deferred Items

| Category | Item | Status | Deferred At |
|----------|------|--------|-------------|
| *(granular flags item superseded — in scope for v1.15)* | | | |

## Session Continuity

Last session: 2026-08-30 (plan 38-04 execution)
Stopped at: Plan 38-04 complete — next: 38-05
Resume file: .planning/phases/38-hierarchy-wrapping-and-granular-keys/38-04-SUMMARY.md
