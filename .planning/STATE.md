---
gsd_state_version: 1.0
milestone: v1.10
milestone_name: Model Composition
status: Awaiting next milestone
last_updated: "2026-08-08T20:10:54.179Z"
last_activity: 2026-08-08 — Milestone v1.10 completed and archived
progress:
  total_phases: 6
  completed_phases: 6
  total_plans: 13
  completed_plans: 13
  percent: 100
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-08-08)

**Core value:** Transform simple TOML architecture descriptions into professional C4 diagrams without manual drawing.
**Current focus:** Milestone complete

## Current Position

Phase: Milestone v1.10 complete
Plan: —
Status: Awaiting next milestone
Last activity: 2026-08-08 — Milestone v1.10 completed and archived

## Performance Metrics

**Velocity (v1.8/v1.9 carry-forward):**

- Last 8 plans averaged ~10 min/plan; range 3-26 min.
- v1.9 Phase 27 was 1 plan / 3 tasks / ~3 min.

| Phase | Plans | Notes |
|-------|-------|-------|
| 28-33 (v1.10) | 11 done | 28/29/30/31/32 complete; 33 next |
| Phase 32 | ~28 min | 2 plans | 5 tasks | 9 files (parser IncludeDirective + internal/include.Resolve + pipeline Stage 1a) |

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
- **Phase 33 UNBLOCKED (Phases 31 + 32 shipped 2026-08-08):** Phase 33 is fully planned (4 plans, 33-01..33-04, waves 1+2). **Plan 04 (XC-05 + XC-01 E2E tests) pre-condition now SATISFIED** — `internal/template.Expand` (Phase 31) and `internal/include.Resolve` (Phase 32) both ship in HEAD. Phase 33 may now execute in full. Note: Phase 32's include.Resolve signature is `Resolve(entry, entryDir, entryFile)` (three args — entryFile threads the real entry filename for INC-10/D-12 error attribution); Plan 04's E2E tests must use this signature, not the simplified `Resolve(m)` from the original CONTEXT.

## Deferred Items

| Category | Item | Status | Deferred At |
|----------|------|--------|-------------|
| Templates | multi-output / `for_each` fan-out | Future (REQUIREMENTS.md) | v1.10 planning |
| Templates | template nesting (template-instantiating-template) | Future (REQUIREMENTS.md) | v1.10 planning |
| Ergonomics | compact-link shorthand variants beyond baseline | Future (REQUIREMENTS.md) | v1.10 planning |
| Docs | docs-drift-orphan-rule-testdata — README "Validation Rules" section (line 551) missing VAL-01 orphan rule; root `testdata/valid.toml`+`nested.toml` unused (tests use `cmd/c4drill/testdata/`) | confirmed_open, low-severity | v1.10 close (pre-existing, not a v1.10 regression) |
| Tooling | knowledge-base.md — NOT a debug session; gsd-debugger tool reference doc the audit scanner misclassifies. No action needed. | wontfix (false positive) | v1.10 close |

## Session Continuity

Last session: 2026-08-08T19:30:00.000Z
Stopped at: Phase 32 execution complete (2/2 plans: 32-01 parser IncludeDirective + 32-02 internal/include.Resolve + merge + pipeline Stage 1a wiring)
Resume file: .planning/phases/33-docs-sweep-end-to-end-goldens/33-CONTEXT.md

## Decisions

- [Phase 28]: Phase 28 reference field ships: per-unit `reference` URL renders a clickable 📖 marker via GraphViz native URL attr; external reference wins the single URL slot over drill-down; HTML shim routes external http(s)// to a new tab and no-ops non-http(s) schemes (T-28-02 hardening). Leaf-field isBuiltinField addition only — NO captureDefinitionOrder change (BC-1). REF-05 proven by the unchanged multilevel golden (canonical-DOT, DI-1). — ARCHITECTURE-v1.10.md §6 (6) Option A and the plan locked decisions. multilevel.toml intentionally NOT modified to preserve the COMPAT-02 golden as the REF-05 backward-compat proof; cluster-label + both-present precedence covered by dedicated unit tests instead.

## Operator Next Steps

- Start the next milestone with /gsd-new-milestone
