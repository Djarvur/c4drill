---
gsd_state_version: 1.0
milestone: v1.11
milestone_name: Label Formatting Fixes
status: completed
last_updated: "2026-08-14T14:02:33.503Z"
last_activity: 2026-08-10
progress:
  total_phases: 1
  completed_phases: 1
  total_plans: 4
  completed_plans: 4
  percent: 100
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-08-10)

**Core value:** Transform simple TOML architecture descriptions into professional C4 diagrams without manual drawing.
**Current focus:** Milestone complete

## Current Position

Phase: 34
Plan: Not started
Status: Milestone complete
Last activity: 2026-08-10

## Performance Metrics

**Velocity (v1.8/v1.9 carry-forward):**

- Last 8 plans averaged ~10 min/plan; range 3-26 min.
- v1.9 Phase 27 was 1 plan / 3 tasks / ~3 min.

| Phase | Plans | Notes |
|-------|-------|-------|
| 28-33 (v1.10) | 11 done | 28/29/30/31/32 complete; 33 next |
| Phase 32 | ~28 min | 2 plans | 5 tasks | 9 files (parser IncludeDirective + internal/include.Resolve + pipeline Stage 1a) |
| Phase 34 P01 | 12 | 3 tasks | 2 files |
| Phase 34 P02 | 20 | 3 tasks | 3 files |

## Accumulated Context

### Decisions (carry-forward, load-bearing for v1.11)

- **DI-1 / canonicalDOT:** golden comparisons must be ORDER-INSENSITIVE (sort-normalize, strip layout geometry bb/pos/lp/lheight/lwidth/height/width). go-graphviz layout is byte-nondeterministic run-to-run. All v1.11 goldens/regression checks must use canonicalDOT, NOT byte-exact `require.Equal`.
- **v1.11 is ONE phase (34):** LABEL-01 (edge labels → HTML table via unit-label machinery) and LABEL-02 (remove `splitLongWord` from `wrapText`) both live in `internal/render` and share the wrap machinery — LABEL-01 routes edge labels through `wrapAndEscape`, LABEL-02 changes `wrapText` semantics for ALL labels. Coordinated single phase, no split. Granularity: standard.
- **LABEL-02 changes unit label output in the over-budget case by DESIGN:** `splitLongWord` removal alters unit label output ONLY when a word exceeds maxChars — that is the intended semantic, not a COMPAT-01 regression. Existing goldens (multilevel fixture) contain no over-budget words and NO edge labels, so COMPAT-02/REF-05 should stay byte-stable — run the full canonicalDOT suite to confirm.
- **Existing tests assert the removed/old behavior — plan-phase must budget for re-assertion:** `TestWrapText` "forced character break" + "multi-byte unicode" cases assert `splitLongWord` output (wrap_internal_test.go); `TestEdgeLabelGeneration` "Technology and Description with newline" asserts a plain `\n` separator via `checkNewline` (labels_test.go) — the HTML table form emits `<BR/>` inside `<td>` cells instead.

### Pending Todos (v1.11)

Both captured todos (`.planning/todos/pending/2026-08-10-wrap-edge-labels-like-unit-labels-in-generated-toml.md`, `2026-08-10-wrap-labels-at-word-boundaries-only.md`) are in scope for v1.11 and mapped to Phase 34 (LABEL-01, LABEL-02). No separate todo action needed.

### Blockers/Concerns (carry-forward, affect v1.11)

- **KNOWN LIMITATION (cosmetic, deferred):** boundary nodes in C3 clusters draw inside the cluster box (go-graphviz WASM cgraph `agsubnode` on edge creation reassigns root nodes to the cluster subgraph; compound=true doesn't help). Out of scope for v1.11.
- **Docs-drift item (deferred, pre-existing):** README "Validation Rules" section (line 551) missing VAL-01 orphan rule; root `testdata/valid.toml`+`nested.toml` unused. Not a v1.11 regression.

### Roadmap Evolution

- Phase 35 added: Add a simple DSL alternative to the TOML diagram definition (likec4/d2-style, less verbose syntax) with converters to and from TOML
- Phase 35 planned (2026-08-14): 9 plans / 6 waves. C4D DSL (pigeon PEG per D-20), nested use in both formats (D-16), template-body use lifted from deferral (D-17), canonical-equivalent round-trip contract (D-22), fmt with comment preservation (D-32). Requirements source = CONTEXT.md D-01..D-35.

## Deferred Items

| Category | Item | Status | Deferred At |
|----------|------|--------|-------------|
| Templates | multi-output / `for_each` fan-out | Future (REQUIREMENTS.md) | v1.10 planning |
| Templates | template nesting (template-instantiating-template) | PROMOTED into Phase 35 (D-17, Plan 35-02) | v1.10 planning |
| Ergonomics | compact-link shorthand variants beyond baseline | Future (REQUIREMENTS.md) | v1.10 planning |
| Docs | docs-drift-orphan-rule-testdata — README "Validation Rules" section (line 551) missing VAL-01 orphan rule; root `testdata/valid.toml`+`nested.toml` unused (tests use `cmd/c4drill/testdata/`) | confirmed_open, low-severity | v1.10 close (pre-existing, not a v1.10 regression) |
| Tooling | knowledge-base.md — NOT a debug session; gsd-debugger tool reference doc the audit scanner misclassifies. No action needed. | wontfix (false positive) | v1.10 close |
| Audit | Both debug-session items above re-flagged by audit-open at v1.11 close; acknowledged again (pre-existing, out of v1.11 scope) | acknowledged | v1.11 close |

## Session Continuity

Last session: 2026-08-14T14:02:33.469Z
Stopped at: Phase 35 context gathered
Resume file: .planning/phases/35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition-/35-CONTEXT.md

## Decisions

- [Phase 34]: v1.11 ships as a single phase — edge-label HTML-table formatting (LABEL-01) and word-boundary-only wrapping (LABEL-02) share the `internal/render` wrap machinery and must land coordinated; COMPAT-01 is enforced via the existing canonicalDOT goldens (DI-1), which the multilevel fixture satisfies (no edge labels, no over-budget words). Test re-assertions required: TestWrapText split cases + TestEdgeLabelGeneration newline assertion. See ROADMAP.md Phase 34 Notes.

## Operator Next Steps

- Run /gsd:plan-phase 34 to plan the Label formatting fixes phase
