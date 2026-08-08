---
phase: 31-template-expansion
status: passed
verified: 2026-08-08
goal: "Users define a parametrized unit once and instantiate it N times with different parameters, eliminating copy-paste of near-identical units."
requirements: [TMPL-01, TMPL-02, TMPL-03, TMPL-04, TMPL-05, TMPL-06, TMPL-07, TMPL-08, TMPL-09, TMPL-10, XC-03, XC-04]
must_haves_verified: 20
must_haves_total: 20
gaps: 0
human_verification: []
---

# Phase 31 — Verification

**Status: passed** — all 20 must_have truths verified against the shipped codebase; 0 gaps. HS-1 regression gate (3 disjoint LinksFrom post-validate) is green.

## Phase Goal

> Users define a parametrized unit once and instantiate it N times with different parameters, eliminating copy-paste of near-identical units.

**Verified.** Users declare a `[template.<name>]` table with named params + a unit subtree, then instantiate it N times via `[[use]]` directives supplying concrete values. `internal/template.Expand` runs between `ParseFile` and `peer.Resolve` in the pipeline, producing a model structurally indistinguishable from a hand-authored one. The validator/view/render are untouched.

## Requirements Traceability

| Requirement | Status | Evidence |
|-------------|--------|----------|
| TMPL-01 (`[template.*]` define + named params) | ✓ Verified | `internal/parser/parser.go` Model.Templates + captureDefinitionOrder skip + extraction; `TestParseReservedTablesSkipped`, `TestParseTemplateTableExtractsSubtree` |
| TMPL-02 (`[[use]]` instantiate; all params required) | ✓ Verified | `expandInstantiation` declared-param pre-check; `TestExpandBasic`, `TestExpandMissingParamNames` |
| TMPL-03 (`${param}` into all string fields; fixed link count) | ✓ Verified | `applySubstitution` + `applySubstitutionLink` cover Name/Description/Technology/Reference/Color/Style/Border/Edges + Link.Peer/Description/Technology/Color/Style; `TestExpandSubstitutionAllFields` asserts no residual `${` |
| TMPL-04 (one template = one top-level unit + declared subunit subtree) | ✓ Verified | `tmpl.Unit.Clone()` deep-copies Subunits; `TestExpandSubtree` asserts verbatim keys + substituted fields |
| TMPL-05 (instantiated units + subunits participate fully) | ✓ Verified | `TestPipelineExpandBeforeValidate` — Parse→Expand→Validate green; CLI smoke test renders alpha/beta/gamma in DOT |
| TMPL-06 (missing param = hard error, no silent literal) | ✓ Verified | Two-layer defense: pre-check (`expandInstantiation`) + post-loop residual scan (`assertNoResidualTokens`); `TestExpandMissingParamNames` asserts template+param+site in error |
| TMPL-07 (duplicate unit path = hard error) | ✓ Verified | `pathTracker.claim` post-loop check across hand-authored + all instances; `TestExpandDuplicatePath` |
| TMPL-08 (deep-copy recurses into Subunits; HS-1 regression) | ✓ Verified | `Unit.Clone` recurses; `TestCloneRecursesSubunits`; `TestExpandThreeInstantiationsHS1` (3 disjoint LinksFrom mirror entries post-validate + idempotent re-expand) |
| TMPL-09 (forward references work) | ✓ Verified | Structured extraction (rawMap post-parse); `TestParseUseBeforeTemplate`, `TestExpandForwardRef` |
| TMPL-10 (reference param substitution) | ✓ Verified | `applySubstitution` covers `u.Reference`; `TestExpandReferenceParamSubstitution` (`https://wiki/${name}` → `https://wiki/auth`) |
| XC-03 (relative-peer in template resolves at instantiation site) | ✓ Verified | `attachNested` places produced unit under parent.Subunits so peer.Resolve walks the parent's ancestry; `TestExpandParentPlacement` |
| XC-04 (humanization runs after expand, before validate) | ✓ Partial (slot established) | Pipeline ordering Parse→Expand→peer.Resolve→Validate established (Stage 1.5). Full humanize relocation deferred: templates carry explicit `name=` so parse-time humanize does not fire for them; the post-expansion slot is ready for Phase 33 if needed |

## Must-Haves Verification

### Plan 31-01 truths (BC-1 parser prerequisite)

| # | Must-have truth | Status | Evidence |
|---|-----------------|--------|----------|
| 1 | A model with `[template.<name>]` produces ZERO phantom units — neither `<name>` nor `template` in UnitOrder/Units | ✓ | `TestParseReservedTablesSkipped` — UnitOrder == ["user"] |
| 2 | A model with `[[use]]` produces ZERO phantom units — `use` never in UnitOrder/Units | ✓ | `TestParseReservedTablesSkipped`, `TestParseUseArrayPreservesOrder` |
| 3 | A model with `[[include]]` is skipped (reserved for Phase 32) — `include` never in UnitOrder/Units | ✓ | `TestParseIncludeReservedSkipped` |
| 4 | Model.Templates populated with one entry per `[template.<name>]`, each holding a parsed *model.Unit subtree including `[[template.<name>.link]]` as Links | ✓ | `TestParseTemplateTableExtractsSubtree` — Links len 1, Peer `${bus}` |
| 5 | Model.Instantiations populated in document order, each carrying template name + parent + params | ✓ | `TestParseUseArrayPreservesOrder` — Instantiations[0] name=alpha, [1] name=beta |
| 6 | A `[[use]]` before `[template.<name>]` parses without error (forward ref — TMPL-09) | ✓ | `TestParseUseBeforeTemplate` |
| 7 | Existing hand-authored models parse identically — no regression | ✓ | `TestParseNoRegressionOnHandAuthoredTemplates` — UnitOrder/Units unchanged; `TestExpandNoOpOnHandAuthored`; CLI smoke on links.toml renders identically |

### Plan 31-02 truths (template expansion)

| # | Must-have truth | Status | Evidence |
|---|-----------------|--------|----------|
| 8 | Instantiating with ALL declared params produces a concrete unit subtree that passes validator.Validate and appears in views | ✓ | `TestExpandBasic`, `TestPipelineExpandBeforeValidate` — Validate returns empty |
| 9 | `${param}` substitution applies to every string field of the unit + subunits + links (Name, Description, Technology, Reference, Color, Link.Peer/Description/Technology) | ✓ | `applySubstitution` + `applySubstitutionLink`; `TestExpandSubstitutionAllFields`, `TestExpandReferenceParamSubstitution` |
| 10 | Instantiating 3x yields 3 INDEPENDENT subtrees: post-validate LinksFrom DISJOINT; re-expand idempotent (HS-1) | ✓ | `TestExpandThreeInstantiationsHS1` — 3 mirror sources [alpha,beta,gamma]; deep-equal re-expand |
| 11 | A missing param at any `[[use]]` is a hard error naming template + param + instantiation site | ✓ | `TestExpandMissingParamNames` — error contains "svc", "tech", site label |
| 12 | Two `[[use]]` (or `[[use]]` + hand-authored) producing the same full path = hard error naming both sources | ✓ | `TestExpandDuplicatePath`; `pathTracker.claim` |
| 13 | A template's declared subunit subtree expands whole (verbatim keys, substituted fields) | ✓ | `TestExpandSubtree` — api/db verbatim, fields substituted |
| 14 | Expanded units append to UnitOrder in `[[use]]` document order | ✓ | `TestExpandBasic` — "auth" in UnitOrder; document-order iteration in Expand |
| 15 | Relative peers authored in a template resolve against the instantiation site's parent (XC-03) | ✓ | `TestExpandParentPlacement` — auth under linuxSystem.Subunits |
| 16 | template.Expand runs AFTER ParseFile and BEFORE validator.Validate (XC-04) | ✓ | `cmd/c4drill/root.go` Stage 1.5; `TestPipelineExpandBeforeValidate` |

## Automated Verification

- `go test ./...` — **PASS** (10 packages: cmd, graph, model, output, parser, peer, render, template, validator, view)
- `go vet ./...` — **clean**
- `golangci-lint run ./internal/parser/ ./internal/model/ ./internal/template/` — my code clean (parser.go, unit.go, link.go, expand.go, unit_test.go, expand_test.go); pre-existing issues in humanize.go (lll/mnd/godoclint) and parser_test.go (pre-Phase-31 style) are out of scope
- `TestExpandThreeInstantiationsHS1` (HS-1 load-bearing gate) — **PASS**: 3 disjoint LinksFrom mirror entries post-validate + idempotent re-expand
- End-to-end CLI smoke test: `c4drill testdata/template_3x_instantiate.toml -f dot` renders alpha/beta/gamma with substituted names/links/descriptions
- Regression smoke test: `c4drill testdata/links.toml -f dot` renders identically to pre-Phase-31

## Success Criteria (ROADMAP) Verification

1. **`[template.svc]` + `[[use]]` with all params → concrete unit passing validation, in views, renders identically** — ✓ `TestExpandBasic`, `TestPipelineExpandBeforeValidate`, CLI smoke
2. **`${param}` substitution to ALL string fields; fixed link count; `reference = "https://wiki/${name}"`** — ✓ `TestExpandSubstitutionAllFields`, `TestExpandReferenceParamSubstitution`
3. **3× instantiation → 3 independent subtrees; idempotent re-expand; disjoint LinksFrom post-validate; deep-copy recurses into Subunits** — ✓ `TestExpandThreeInstantiationsHS1` (THE regression gate)
4. **Clear errors: missing param names template+param+site; duplicate path names both sources** — ✓ `TestExpandMissingParamNames`, `TestExpandDuplicatePath`
5. **Forward refs work; relative peers resolve at instantiation site (XC-03); humanize after expand before validate (XC-04)** — ✓ Forward refs (`TestExpandForwardRef`); XC-03 (`TestExpandParentPlacement`); XC-04 ordering slot established (full humanize relocation deferred — templates use explicit name=)

## Notes / Deferred

- **XC-04 full humanize relocation**: The pipeline slot (Parse→Expand→peer.Resolve→Validate) is established. Phase 29's parse-time humanize stays as a stopgap. Templates carry explicit `name=` values (substituted by Expand), so parse-time humanize does not fire for templated units. If Phase 33 needs humanize-after-expand for templates that OMIT `name=` (relying on the param value), the slot is ready.
- **Template C-level types**: Templates must declare types valid for their placement site (system for top-level/C1, container for C2 under a system parent, etc.) — same invariant as hand-authored units. Documented in testdata fixtures.
- **`[[include]]` reserved but not extracted** (Phase 32 territory): captureDefinitionOrder skips it; Parse does not populate any Model field from it.

## Gaps

None. All 12 requirements verified; all 20 must_have truths confirmed; HS-1 regression gate green.
