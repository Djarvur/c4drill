---
gsd_state_version: 1.0
milestone: v1.14
milestone_name: Nesting Context and Plain Rendering
status: executing
last_updated: "2026-08-30T14:39:20.887Z"
last_activity: 2026-08-30
progress:
  total_phases: 2
  completed_phases: 0
  total_plans: 7
  completed_plans: 6
  percent: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-08-30)

**Core value:** Transform simple TOML architecture descriptions into professional C4 diagrams without manual drawing.
**Current focus:** Phase 37 — nesting-context-and-plain-rendering

## Current Position

Phase: 37 (nesting-context-and-plain-rendering) — EXECUTING
Plan: 7 of 7
Status: Ready to execute
Last activity: 2026-08-30

Progress: [█████████░] 86%

## Performance Metrics

**Velocity (carry-forward):**

- v1.11: 1 phase (34), 4 plans. v1.12: 1 phase (35), 9 plans. v1.13: 1 phase (36), 6 plans.
- Expect Phase 37 at mid single-digit plans (plan-phase decides).

| Phase | Plans | Notes |
|-------|-------|-------|
| 34 (v1.11) | 4 done | TDD RED→GREEN; 28 commits |
| 35 (v1.12) | 9 done | 25 tasks; grammar/emitter plans 42–53 min |
| 36 (v1.13) | 6 done | 20/20 requirements; release v1.18.0 |
| Phase 37 P01 | 13min | 3 tasks | 6 files |
| Phase 37 P02 | 32min | 3 tasks | 7 files |
| Phase 37 P03 | 11min | 3 tasks | 5 files |
| Phase 37 P04 | 18min | 3 tasks | 6 files |
| Phase 37 P05 | 12min | 3 tasks | 0 files |

## Accumulated Context

### Decisions (carry-forward, load-bearing for v1.14)

- **DI-1 / canonicalDOT:** golden comparisons ORDER-INSENSITIVE via `internal/testutil/canonical` (strip layout geometry bb/pos/lp/lheight/lwidth/height/width, sort statements + attributes); never byte-exact `require.Equal`.
- **v1.14 is ONE phase (37):** CTX (view semantics: `internal/view/scope.go` + `internal/graph` clusters) and PLAIN (CLI flag in `cmd/c4drill` → render pipeline) share canonicalDOT goldens and emission points; splitting forces cross-phase golden churn. Precedent: v1.11 (1 phase/4 plans), v1.12 (1/9), v1.13 (1/6).
- **BC-01 delta discipline (differs from v1.13):** `--plain` is opt-in → NO universal output delta. Only documented CTX nesting deltas may re-baseline goldens; flat models must render unchanged; full suite green. (v1.13's legend-on-by-default was the accepted universal delta; nothing comparable here.)
- **Fix points (pin in phase research):** CTX — two known loss mechanisms: C1 deep-link target collapse + one-level expanded rendering (flat subunit lists); expected in `internal/view/scope.go` (GenerateC1View, visible subunits, boundary resolution) + `internal/graph` cluster building. PLAIN — suppression at applyNodeStyle/applyClusterStyle (converter.go), edge colour/style/length/rank emission, configureGraphSettings (`properties.edges`), labels.go (plain-text labels, content preserved).
- **Plain-mode boundary:** kind-derived edge colours + legend STAY (semantic, not formatting); collapsed subtrees NOT restructured — chains only for DEPICTED elements; no granular flags; no properties-level keys (Out of Scope, REQUIREMENTS.md).
- **TDD mode on** (config workflow.tdd_mode); docs + skill sync land inside the phase (v1.12/v1.13 precedent); REL-01 tag **v1.21.0** is the phase's final task.

### Pending Todos

None pending — v1.14 requirements sourced directly from user request 2026-08-30 (REQUIREMENTS.md).

### Blockers/Concerns

- **Clarify-in-research:** clarifying questions were auto-skipped (yolo mode); REQUIREMENTS.md scope is the orchestrator's best-judgment reading — phase research must confirm exact fix points against the codebase before planning.
- **Docs-drift item (pre-existing, deferred):** README "Validation Rules" missing VAL-01 orphan rule; root testdata/valid.toml+nested.toml unused. Not v1.14 scope.

### Quick Tasks Completed

See MILESTONES.md post-milestone section (260828-qbx queue pipes, 260828-tgf pipe end cap — shipped as v1.19.0–v1.20.0).

### Roadmap Evolution

- 2026-08-30: Milestone v1.14 planned as single Phase 37 — 11 requirements (CTX-01..03, PLAIN-01..04, BC-01, DOC-01..03, REL-01), 5 success criteria, 100% coverage.

## Deferred Items

| Category | Item | Status | Deferred At |
|----------|------|--------|-------------|
| Templates | multi-output / `for_each` fan-out | Future (REQUIREMENTS.md) | v1.10 planning |
| Ergonomics | compact-link shorthand variants beyond baseline | Future (REQUIREMENTS.md) | v1.10 planning |
| Styling | properties.color/style/border (page-level) parsed-but-dead | Out of scope (REQUIREMENTS.md) — not the reported defect | v1.13 planning |
| Plain | granular per-aspect flags; properties-level ignore keys | Out of scope (REQUIREMENTS.md) — one `--plain` key covers the need | v1.14 planning |
| Docs | docs-drift-orphan-rule-testdata — README VAL-01 gap + unused root testdata | confirmed_open, low-severity | v1.10 close |
| Audit | debug-session items re-flagged each close (false positive + doc drift) | acknowledged | v1.12 close |

## Session Continuity

Last session: 2026-08-30T14:39:20.879Z
Stopped at: Completed 37-05-PLAN.md (visual checkpoint auto-approved in auto-mode)
Resume file: None

## Operator Next Steps

- Plan the phase with `/gsd:plan-phase 37`

## Decisions

- [Phase ?]: CTX-03 landed with zero golden deltas — Expanded-mode baselines stayed green: GenerateExpandedView sets IsExpanded=HasSubunits (buildNestedCluster path untouched, magnifier guard inert) and no test reads the committed cmd .dot/.svg artifacts; 37-05 re-baseline scope must be re-assessed
- [Phase ?]: CTX-02 true-target chains landed with zero committed-golden deltas; seven collapse-pinning content tests updated to the new contract; 37-05 re-baseline scope near-empty, focus BC-01 + CTX-01 proof — chains unfold only under in-scope depicted ancestors; external/sibling boundary nodes keep collapsed resolution per the plan scope guard
- [Phase ?]: PLAIN-01/02 via flag-on-struct threading (View.Plain -> Graph.Plain -> builder guards), no package global; suppression = treat author fields as unset; kind colours + legend survive (semantic), arrow/penwidth unsuppressed (structural) — Matches the AllExpanded precedent and the plan's explicit rejection of the LabelRatio global; one guard point covers every render path
- [Phase ?]: Zero committed-golden debt after 37-03 (15/15 green, expected-failure list EMPTY); 37-05 scope = BC-01 flat-model stability + goldens 37-04 adds; collapsed-pair AGG-02 aggregate style kept inert under plain so author line styles cannot leak — Plain=false keeps the default path byte-stable; the AGG-02 style derives from author link.Style and would otherwise bypass the createEdge guard
- [Phase 37]: PLAIN-03/04 via SetLabel plain-text labels (buildRecordLabel for nodes/clusters, '[Technology] Description' via buildEdgePlainTextLabel for edges) threaded through the converter call tree — true plain text chosen over SetLabelHTML record path per the plan's pinned decision; graphviz escaping covers T-37-07 — PLAIN-03/04 via SetLabel plain-text labels (buildRecordLabel for nodes/clusters, '[Technology] Description' via buildEdgePlainTextLabel for edges) threaded through the converter call tree — true plain text chosen over SetLabelHTML record path per the plan's pinned decision; graphviz escaping covers T-37-07
- [Phase 37]: Cluster drill-down URL emission restructured to run AFTER the plain/HTML label branch in setClusterLabel so the CTX-03 structural affordance survives --plain while the default path early-return (no URL on empty HTML label) stays byte-identical — Cluster drill-down URL emission restructured to run AFTER the plain/HTML label branch in setClusterLabel so the CTX-03 structural affordance survives --plain while the default path early-return (no URL on empty HTML label) stays byte-identical
- [Phase 37]: E2E plain-uniformity assertion uses lowercase HTML markers (<table/<b>/<i> absent) + label=< count == 2 because legend and nav/title legitimately keep UPPERCASE HTML tables under plain — a literal '<TABLE absent' assertion would contradict the locked plain-mode boundary — E2E plain-uniformity assertion uses lowercase HTML markers (<table/<b>/<i> absent) + label=< count == 2 because legend and nav/title legitimately keep UPPERCASE HTML tables under plain — a literal '<TABLE absent' assertion would contradict the locked plain-mode boundary
- [Phase ?]: 37-05: no golden re-baseline needed — only test-compared golden multilevel.expanded.dot byte-identical; suite green (BC-01)
- [Phase ?]: 37-05: stale Milestone-v1.10 artifacts (expanded.dot, expanded/mainsystem.dot) not re-baselined — unconsumed by tests, drift predates phase 37; logged to deferred-items.md
