---
gsd_state_version: 1.0
milestone: v1.14
milestone_name: Nesting Context and Plain Rendering
status: planning
last_updated: "2026-08-30T10:35:40.667Z"
last_activity: 2026-08-30
progress:
  total_phases: 0
  completed_phases: 0
  total_plans: 0
  completed_plans: 0
  percent: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-08-28)

**Core value:** Transform simple TOML architecture descriptions into professional C4 diagrams without manual drawing.
**Current focus:** Between milestones — milestone v1.13 shipped (v1.18.0); post-milestone design-review work shipped as v1.19.0–v1.20.0 (legend rework, queue pipes)

## Current Position

Phase: Not started (defining requirements)
Plan: —
Status: Defining requirements
Last activity: 2026-08-30 — Milestone v1.14 started

## Performance Metrics

**Velocity (carry-forward):**

- v1.11: 1 phase (34), 4 plans. v1.12: 1 phase (35), 9 plans, avg ~27 min/plan.
- Expect Phase 36 at mid single-digit plans (plan-phase decides).

| Phase | Plans | Notes |
|-------|-------|-------|
| 34 (v1.11) | 4 done | TDD RED→GREEN; 28 commits |
| 35 (v1.12) | 9 done | 25 tasks; 42–53 min for grammar/emitter plans |

## Accumulated Context

### Decisions (carry-forward, load-bearing for v1.13)

- **DI-1 / canonicalDOT:** golden comparisons must be ORDER-INSENSITIVE (sort-normalize, strip layout geometry bb/pos/lp/lheight/lwidth/height/width). All v1.13 goldens use `internal/testutil/canonical`, never byte-exact `require.Equal`.
- **v1.13 is ONE phase (36):** all features live in the same render/view/graph/c4d packages and share goldens; splitting would force cross-phase golden churn. Precedent: v1.11 (1 phase/4 plans), v1.12 (1 phase/9 plans). Granularity: standard.
- **BC-01 accepted delta:** default-on legend means models using no new feature do NOT render byte-identical — the legend block is the single user-mandated re-baseline delta. Everything else must be golden-clean.
- **Fix points (pre-confirmed scan):** COLOR → builder.go buildNode/buildCluster + shapes.go, emit via converter.go applyNodeStyle/applyClusterStyle. GEDGE → view/scope.go:377 (C2) + :470 (C3) need Properties.Edges fallback; "square" unimplemented in configureGraphSettings (converter.go:161-169). RANK → createEdge (builder.go:573-611) + converter.go:483-495, swap endpoints + dir=back; 4 view copiers already carry Rank. AGG → first-wins pair merge in processOutgoingLinks/processIncomingLinks. KIND is the widest-touch requirement: new model.Link.Kind → 4 view copiers, validator/index.go mirror, c4d.peg:458 OptionKey + go:generate regen, tomodel applyEdgeOption, frommodel edgeStmtFromLink, emit_toml canonical order, grammar/reserved.go fieldKeywords, testutil/canonsrc.
- **LEGEND fix point SUPERSEDED post-milestone:** the v1.13 approach (legend rows in the top graph-label HTML table) was replaced in v1.19.x by a floating legend node (`__c4drill_legend`, plaintext, framed HTML table) outside an invisible `cluster___content` wrapper; REQUIREMENTS.md LEG-01..03 now describe the current design.
- **TDD mode on** (config workflow.tdd_mode); docs + skill sync land inside the phase (v1.12 precedent); REL-01 tag v1.18.0 is the phase's final task.

### Pending Todos

None pending — v1.13 requirements sourced directly from user request 2026-08-28 (REQUIREMENTS.md).

### Blockers/Concerns

- ~~**Legend-in-top-label constraint:** GraphViz cannot position clusters; the legend renders as a right-aligned column of the top graph-label HTML table.~~ **Resolved post-milestone (v1.19.x):** the legend is now a floating node outside an invisible content cluster — see Quick Tasks and REQUIREMENTS.md LEG-01.
- **Docs-drift item (pre-existing, deferred):** README "Validation Rules" missing VAL-01 orphan rule; root testdata/valid.toml+nested.toml unused. Not v1.13 scope.

### Quick Tasks Completed

| # | Description | Date | Commit | Directory |
|---|-------------|------|--------|-----------|
| 260828-qbx | Render queue units as horizontal pipe shapes via SVG post-processing | 2026-08-28 | c880c7a + bed782d | [260828-qbx-render-queue-units-as-horizontal-pipe-sh](./quick/260828-qbx-render-queue-units-as-horizontal-pipe-sh/) |
| 260828-tgf | Fix pipe end cap: right side renders a full ellipse (coincident-endpoint SVG arc was silently omitted) | 2026-08-28 | afdcbfb | [260828-tgf-fix-pipe-end-cap-right-side-must-render-](./quick/260828-tgf-fix-pipe-end-cap-right-side-must-render-/) |

### Roadmap Evolution

- 2026-08-28: Milestone v1.13 planned as single Phase 36 — 20 requirements (COLOR-01..02, GEDGE-01..02, RANK-01..02, KIND-01..03, AGG-01..03, LEG-01..03, BC-01, DOC-01..03, REL-01), 5 success criteria, 100% coverage.

## Deferred Items

| Category | Item | Status | Deferred At |
|----------|------|--------|-------------|
| Templates | multi-output / `for_each` fan-out | Future (REQUIREMENTS.md) | v1.10 planning |
| Ergonomics | compact-link shorthand variants beyond baseline | Future (REQUIREMENTS.md) | v1.10 planning |
| Styling | properties.color/style/border (page-level) parsed-but-dead | Out of scope (REQUIREMENTS.md) — not the reported defect | v1.13 planning |
| Docs | docs-drift-orphan-rule-testdata — README VAL-01 gap + unused root testdata | confirmed_open, low-severity | v1.10 close |
| Audit | debug-session items re-flagged each close (false positive + doc drift) | acknowledged | v1.12 close |

## Session Continuity

Last session: 2026-08-28
Stopped at: ROADMAP.md created for v1.13 (Phase 36, 20/20 requirements); REQUIREMENTS.md traceability filled
Resume file: None

## Operator Next Steps

- Plan the phase with `/gsd:plan-phase 36`
