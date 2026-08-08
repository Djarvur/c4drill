---
gsd_state_version: 1.0
milestone: v1.10
milestone_name: Model Composition
status: Roadmap created; Phase 28 ready for discuss → plan
last_updated: "2026-08-08T16:44:37.960Z"
last_activity: 2026-08-08 — v1.10 roadmap written (ROADMAP.md + REQUIREMENTS traceability)
progress:
  total_phases: 6
  completed_phases: 0
  total_plans: 0
  completed_plans: 0
  percent: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-08-08)

**Core value:** Transform simple TOML architecture descriptions into professional C4 diagrams without manual drawing.
**Current focus:** v1.10 Model Composition — Phase 28 (Reference field) ready to discuss/plan.

## Current Position

Phase: 28 of 33 (v1.10 spans phases 28-33)
Plan: — (not yet planned)
Status: Roadmap created; Phase 28 ready for discuss → plan
Last activity: 2026-08-08 — v1.10 roadmap written (ROADMAP.md + REQUIREMENTS traceability)

Progress: [░░░░░░░░░░] 0% of v1.10

**Build order:** 28 / 29 / 30 are independent & parallelizable → 31 (templates, carries BC-1 parser change) → 32 (include) → 33 (docs + integration goldens).

## Performance Metrics

**Velocity (v1.8/v1.9 carry-forward):**

- Last 8 plans averaged ~10 min/plan; range 3-26 min.
- v1.9 Phase 27 was 1 plan / 3 tasks / ~3 min.

| Phase | Plans | Notes |
|-------|-------|-------|
| 28-33 (v1.10) | 0 done | Not started |

## Accumulated Context

### Decisions (carry-forward, load-bearing for v1.10)

- **DI-1 / canonicalDOT:** golden comparisons must be ORDER-INSENSITIVE (sort-normalize, strip layout geometry bb/pos/lp/lheight/lwidth/height/width). go-graphviz layout is byte-nondeterministic run-to-run. v1.10 multi-file + templates add ANOTHER ordering axis — all XC-05 goldens must use canonicalDOT, NOT byte-exact `require.Equal`.
- **go-toml/v2 stays NON-strict:** `DisallowUnknownFields()` must stay OFF — the inline-subunit trick (`unit.go:71`, `toml:",inline"`) depends on unknown keys being silently accepted. v1.10 adds new top-level tables; add a guard comment. (research §9 BC-3)
- **Safari/WebKit ignores SVG `<a>` navigation:** the `-f html` format inlines SVG + injects a JS shim. REF-04 (Phase 28) requires the shim to route external `reference` URLs distinctly from internal drill-down URLs.
- **Validator mutates units in place:** `populateIncomingLinks` (index.go:70-81) appends to `Unit.LinksFrom`. HS-1 (Phase 31) requires a real `Unit.Clone()` — shallow copy corrupts the Nth template instantiation. Do NOT round-trip-copy Links (`Link.Mirror` would reset, re-breaking multiplicity counting).
- **Validator is the single gatekeeper:** all four v1.10 features are pure pre-parse/validation passes producing a model structurally identical to a hand-authored single-file model. validator/view/render stay untouched except the reference glyph.

### Pending Todos (now CONSUMED by v1.10 requirements)

The five todos in `.planning/todos/pending/` (reference field, ergonomics, type-omittable docs, unit templates, include directive) are all in scope for v1.10 and mapped to phases 28-33. No separate todo action needed.

### Blockers/Concerns (carry-forward, affect v1.10)

- **Discuss-phase design forks (MUST settle before planning 31 & 32):** HS-2 relative-peer resolution site for template-authored links (recommended: instantiation-parent); forward-reference policy (TM-2/10); unresolved-`${param}` strictness (TM-5); include merge semantics (IN-2/IN-7); diamond behavior (IN-3/INC-05); directive-table naming + reserved-word collision (BC-2). See ROADMAP.md phase Notes + research §6.
- **ERGO-06 (compact-link shorthand) at-risk:** research SUMMARY §3 flags it a v1.10 anti-feature; mapped to Phase 29 for coverage but discuss must decide confirm-vs-defer-to-v2.
- **Parser BC-1 prerequisite:** whichever of templates/include lands first needs coordinated `captureDefinitionOrder` (parser.go:100) + `Parse` rawMap extraction changes so `template`/`include`/`use` tables don't become phantom units. Landed as Plan 1 of Phase 31. `reference` is the only safe single-line `isBuiltinField` addition (Phase 28).
- **KNOWN LIMITATION (cosmetic, deferred):** boundary nodes in C3 clusters draw inside the cluster box (go-graphviz WASM cgraph `agsubnode` on edge creation reassigns root nodes to the cluster subgraph; compound=true doesn't help). Out of scope for v1.10.

## Deferred Items

| Category | Item | Status | Deferred At |
|----------|------|--------|-------------|
| Templates | multi-output / `for_each` fan-out | Future (REQUIREMENTS.md) | v1.10 planning |
| Templates | template nesting (template-instantiating-template) | Future (REQUIREMENTS.md) | v1.10 planning |
| Ergonomics | compact-link shorthand variants beyond baseline | Future (REQUIREMENTS.md) | v1.10 planning |

## Session Continuity

Last session: 2026-08-08T16:44:37.934Z
Stopped at: Phase 31 context gathered
Resume file: .planning/phases/31-template-expansion/31-CONTEXT.md
