---
gsd_state_version: 1.0
milestone: v1.13
milestone_name: Edge Semantics and Legend
status: planning
last_updated: "2026-08-28T10:40:43.820Z"
last_activity: 2026-08-28
progress:
  total_phases: 0
  completed_phases: 0
  total_plans: 0
  completed_plans: 0
  percent: 0
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-08-10)

**Core value:** Transform simple TOML architecture descriptions into professional C4 diagrams without manual drawing.
**Current focus:** Phase 35 — add-a-simple-dsl-alternative-to-the-toml-diagram-definition

## Current Position

Phase: Not started (defining requirements)
Plan: —
Status: Defining requirements
Last activity: 2026-08-28 — Milestone v1.13 started

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
| Phase 35 P01 | 42min | 3 tasks | 11 files |
| Phase 35 P02 | 21min | 2 tasks | 4 files |
| Phase 35 P03 | 24min | 2 tasks | 7 files |
| Phase 35 P04 | 21min | 2 tasks | 4 files |
| Phase 35 P05 | 30min | 3 tasks | 13 files |
| Phase 35 P06 | 53min | 3 tasks | 19 files |
| Phase 35 P07 | 13min | 3 tasks | 4 files |
| Phase 35 P08 | 24min | 3 tasks | 5 files |
| Phase 35 P09 | 15min | 4 tasks | 51 files |

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
| Templates | template nesting (template-instantiating-template) | SHIPPED in Phase 35 (D-17, Plan 35-02) — deferral lifted | v1.10 planning |
| Ergonomics | compact-link shorthand variants beyond baseline | Future (REQUIREMENTS.md) | v1.10 planning |
| Docs | docs-drift-orphan-rule-testdata — README "Validation Rules" section (line 551) missing VAL-01 orphan rule; root `testdata/valid.toml`+`nested.toml` unused (tests use `cmd/c4drill/testdata/`) | confirmed_open, low-severity | v1.10 close (pre-existing, not a v1.10 regression) |
| Tooling | knowledge-base.md — NOT a debug session; gsd-debugger tool reference doc the audit scanner misclassifies. No action needed. | wontfix (false positive) | v1.10 close |
| Audit | Both debug-session items above re-flagged by audit-open at v1.11 close; acknowledged again (pre-existing, out of v1.11 scope) | acknowledged | v1.11 close |
| Audit | Both debug-session items above re-flagged by audit-open at v1.12 close; acknowledged again (pre-existing, out of v1.12 scope) | acknowledged | v1.12 close |

## Session Continuity

Last session: 2026-08-14T19:55:07.059Z
Stopped at: Completed 35-09-PLAN.md (docs+twins+skill surface; phase 35 all 9 plans done)
Resume file: None

## Decisions

- [Phase 34]: v1.11 ships as a single phase — edge-label HTML-table formatting (LABEL-01) and word-boundary-only wrapping (LABEL-02) share the `internal/render` wrap machinery and must land coordinated; COMPAT-01 is enforced via the existing canonicalDOT goldens (DI-1), which the multilevel fixture satisfies (no edge labels, no over-budget words). Test re-assertions required: TestWrapText split cases + TestEdgeLabelGeneration newline assertion. See ROADMAP.md Phase 34 Notes.
- [Phase ?]: 35-01: pigeon pinned at v1.3.0 (v1.0.0 lacks -nolint); generated parser in internal/c4d/grammar + AST in internal/c4d/ast — pigeon's emitted Parse/ParseFile collide with typed c4d.Parse signatures
- [Phase ?]: 35-01: C4D trivia model for fmt — leading comments ride the following statement, same-line tails ride the preceding one, orphans ride the enclosing node
- [Phase 35]: [35-02] All three use forms normalize to one Instantiation mechanism: UseSite document-order capture (narrow ArrayTable admission in the unstable-API pass) + per-path cursor pairing
- [Phase 35]: [35-02] Nested-use explicit parent keys are site-relative; template-body-use parents are clone-root-relative in Expand (basePath joinPath) — produced units never escape the enclosing clone
- [Phase 35]: [35-02] claimSubtree claims every produced subtree path (closes pre-existing TMPL-07 silent-overwrite gap); cycle detection + maxTemplateDepth=100 mirror the include pattern
- [Phase 35]: [35-02] Template-declared types stay fixed at parse time — nested template authors write final level-specific types (containerBox/containerDb), pinned by the validating HS-1 test
- [Phase ?]: [35-03] C4D use args use ONE ordered []Arg{Name, Value} representation (named keys, positional empty Name); positional values containing ':' must be quoted — the named form wins
- [Phase ?]: [35-03] D-19 errors fire from pigeon grammar ACTIONS, not a post-parse walk: action errors record with position but do not fail the match, riding errList through wrapPigeonError; ReservedUnitId's brace lookahead is the false-positive guard (Risk 2)
- [Phase ?]: [35-03] ReservedKeyword single-sourced in grammar/reserved.go (19 words: 14 isBuiltinField + 5 statement keywords); reserved.go lives in package grammar — peg actions compile there (import-cycle forced, 35-01 layout)
- [Phase ?]: [35-04] C4D comment placement is allocation-based: Pos-aware same-line tails + one-step migration to the untailed predecessor — the grammar's StmtEnd attaches the first comment after a statement as its tail, so all-above rendering can never be fixpoint-stable; the allocation renders the placement the parser reproduces verbatim (T-35-04-01)
- [Phase ?]: [35-04] Emitters deterministic by construction: UnitOrder/SubunitOrder walks, sorted template names, sorted [[use]] param keys; newline values emit escaped-\n in TOML and triple-quoted in C4D (D-06 pair); parented uses land inside the parent unit's block in C4D (D-16 native form), top-level [[use]]+parent in TOML
- [Phase ?]: [35-05] C4D glyph->Link mapping: ->/<->/-- carry their ArrowDirection in Links, <- -> LinksFrom{ArrowForward} (mirror-consistent incoming edge); the twin TOML states arrow = "forward" where the fixture omits it (D-22 explicit-defaults normalization)
- [Phase ?]: [35-05] c4d.Parse returns *parser.Model (D-21) composing exported ParseAST+ToModel; ParseAST/ParseASTFile stay exported for fmt (35-08) and canonsrc (35-06); ToModel honors FromModel's Body.Type template-root recording — the 35-04 gap closed at Model level
- [Phase ?]: [35-05] include.Resolve dispatches per included-file extension (.c4d -> c4d.ParseFile, .toml -> parser.ParseFile); unknown extensions hard-error naming .toml/.c4d — mixed-format graphs merge at Model level (D-26/T-35-05-01)
- [Phase ?]: 35-06: D-22 glyph mapping revision — -> maps to the OMITTED arrow default (Arrow ""), not ArrowForward (renderer emits dir=forward only for the explicit value, so the explicit mapping made C4D models render apart from TOML twins); forward/reverse ride -> { arrow: X }, non-default LinksFrom arrows ride <- { arrow: X }
- [Phase ?]: 35-06: template root types ride a template-body type: statement (grammar admits it in template bodies only; unit bodies still reject it; EmitC4D renders Body.Type/External) and PeerRef segments accept ${param} tokens — the 35-04 deferred text gap closed, parametrized peers round-trip
- [Phase ?]: 35-06: pre-existing parser nondeterminism fixed — recordHandAuthored ignored table paths deeper than 2 segments and parseUnitWithOrder looked orders up by short names, so C3+ subunits ordered by Go map iteration (masked until the round-trip contract made order load-bearing); every ancestor pair now records and the recursion carries full lookup paths
- [Phase ?]: 35-07: convert emission NEVER sees the pipeline — D-24 gate (parse->include->expand->peer->validate) runs on a discarded copy; twins emit from a fresh source parse so includes/templates/use/bare peers survive verbatim (D-25/D-22 parity with 35-06)
- [Phase ?]: 35-07: --follow-includes converts each graph file from its own fresh parse, rewriting only include-path strings (once + relative form preserved); already-target-format files are skipped — conversion is additive, so mixed .toml/.c4d graphs stay coherent (D-26)
- [Phase ?]: 35-07: unknown input extension is a hard parse error naming .toml/.c4d (parseInput helper shared by render and convert); convert's direction gate mirrors it
- [Phase ?]: 35-08: fmt preserves the AUTHOR's key order on .toml (contrast with convert's D-23); values render from raw source bytes verbatim — idempotent by construction
- [Phase ?]: 35-08: fmt's T-35-08-01 gate — candidate must re-parse DeepEqual to the original Model BEFORE any write (applyFormatted seam, both failure legs tested)
- [Phase ?]: 35-08: corpus .c4d coverage via internal/include/testdata real fixtures + converted twins of every valid TOML fixture (the plan's four roots ship zero .c4d files)
- [Phase ?]: 35-09: twins are convert-generated, hand-finished with header comments, fmt-canonicalized — TestExampleTwins enforces model parity (12 pairs), render parity (standalone + graphs) and self-contained .c4d include graphs; classification is structure-driven (has-Includes = entry, include-target = fragment)
- [Phase ?]: 35-09: 04-styling.toml formatted in place (Rule 3) so fmt --check skill/examples/ exits 0 — formatting-only normalization, comment text verbatim; c4drill-toml skill name kept (D-35 compat) with dual-format description; both plugin skill copies + all 5 manifests synced

## Operator Next Steps

- Start the next milestone with /gsd-new-milestone
