---
phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition
plan: 06
subsystem: parser
tags: [c4d, dsl, parity, round-trip, canonsrc, d22, canonical, render-equivalence, composed-graph, fixtures, grammar, determinism]

# Dependency graph
requires:
  - phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition (Plan 04)
    provides: EmitTOML/FromModel/EmitC4D deterministic emitters
  - phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition (Plan 05)
    provides: c4d.Parse (Model hub), exported ParseAST/ParseASTFile, ToModel, mixed-format include.Resolve
provides:
  - internal/testutil/canonsrc — NormalizeTOML/NormalizeC4D canonical source normalizers (D-22, the DI-1 precedent applied to source formats)
  - internal/c4d/parity_test.go — full-corpus round-trip suites (both directions), composed include-graph conversion, invalid-fixture refusal guards, canonicalDOT render-equivalence suite
  - testdata/c4d/ — six-fixture edge-case corpus (external types, linkFrom, multiline, rank equal, nested use, unicode)
  - grammar template-body `type:` statement + PeerRef ${param} segments (round-trip-enabling language surface)
affects: [35-07 convert (the parity contract it must uphold), 35-08 fmt (canonical orders), 35-09 CLI dispatch]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Canonical-source normalization via the semantic hub: NormalizeTOML parses through parser.Parse and re-serializes, so representation differences die at the Model boundary; NormalizeC4D does the same through ParseAST with grammar-shaped output (fixpoint-stable)"
    - "Contract-enforcer parity suite: the corpus walker with a >15 tally guard and per-fixture subtests turns every emitter/parser gap into a named failing fixture — four gaps surfaced and were fixed in the offending plans' files"
    - "Compare-before-Validate in model-equality tests: validator.Validate mutates in place (populateIncomingLinks appends mirrors in map-iteration order), so equality assertions run on pre-validation models and validity is proven on fresh parses"

key-files:
  created:
    - internal/testutil/canonsrc/canonsrc.go
    - internal/testutil/canonsrc/canonsrc_test.go
    - internal/testutil/canonsrc/canonsrc_internal_test.go
    - internal/c4d/parity_test.go
    - testdata/c4d/external-types.toml
    - testdata/c4d/linkfrom.toml
    - testdata/c4d/multiline-strings.toml
    - testdata/c4d/rank-equal.toml
    - testdata/c4d/template-nested-use.toml
    - testdata/c4d/unicode-strings.toml
  modified:
    - internal/c4d/grammar/c4d.peg
    - internal/c4d/grammar/parser_gen.go
    - internal/c4d/emit_c4d.go
    - internal/c4d/frommodel.go
    - internal/c4d/tomodel.go
    - internal/c4d/tomodel_test.go
    - internal/c4d/emit_test.go
    - internal/parser/parser.go
    - .planning/phases/35-.../35-04-deferred-items.md

key-decisions:
  - "NormalizeTOML routes through parser.Parse (the plan's preferred Model route): two TOML texts are canonically equal iff their Models are, and the D-22 explicit-defaults list (arrow=forward, rank=forward, labelPosition=middle) drops at serialization"
  - "`->` maps to the OMITTED arrow default (Arrow \"\"), not ArrowForward: the renderer emits dir=forward for the explicit value and nothing for \"\", so 35-05's mapping made C4D-parsed models render apart from their TOML twins — the glyph IS the default, `arrow: forward` states it explicitly"
  - "Template root types ride a template-body `type:` statement (the deferred-items note's suggested fix): grammar admits it in template bodies only, EmitC4D renders Body.Type/External — the 35-04 text-level gap is closed"
  - "PeerRef segments accept ${param} tokens: parametrized template link peers (XC-03) are core v1.10 surface and must round-trip"
  - "Round-trip models compare RAW (m2 == m3, canonicalModel(m1) == canonicalModel(m3) with only the defaults list filled): everything else — UnitOrder, SubunitOrder, values, template roots — survives exactly"
  - "Invalid corpus fixtures assert refusal through the D-24 gate mirror (parse/expand/peer/validate must error somewhere); the six testdata/c4d fixtures prove the positive path through the same gate"

patterns-established:
  - "canonicalModel test helper: deep-clone + fill D-22 defaults for require.Equal comparisons across the DSL hop"
  - "Corpus walker split (flat dirs + recursive examples walk with SkipDir) with a repo-root-anchored path space and a pinned invalid list"

requirements-completed: [D-22, D-26]

# Metrics
duration: 53min
completed: 2026-08-14
---

# Phase 35 Plan 06: Parity Contract Summary

**canonsrc normalizers (D-22) + a full-corpus round-trip/render-equivalence suite that closed four real gaps: the template-root-type text form (grammar + emitter), parametrized ${param} peers, the arrow-glyph mapping that rendered apart from TOML twins, and a pre-existing parser nondeterminism in deep subunit order**

## Performance

- **Duration:** 53 min (17:52–18:45 UTC)
- **Started:** 2026-08-14T17:52:22Z
- **Completed:** 2026-08-14T18:45:13Z
- **Tasks:** 3/3
- **Files modified:** 19 (10 created, 9 modified)

## Accomplishments

- **Task 1 — canonsrc (D-22):** `NormalizeTOML(t, src)` parses via parser.Parse and re-serializes canonically (unit tables in UnitOrder, links before subunit tables, sorted-key fields, sorted-name templates, ordered [[use]]/[[include]]); `NormalizeC4D(t, src)` parses via the exported `c4d.ParseAST` and emits a grammar-shaped canonical text (sections in AST order, comments/Pos dropped, literal KINDS dropped — values canonical). Both normalize the D-22 defaults list away; newline-containing values serialize through ONE form per format (TOML multi-line basic string with trailing-newline escape; C4D raw triple-quote); both outputs re-parse and re-normalize identically (fixpoint pinned by tests); malformed input fails via t.Fatalf (canonical.Canonical contract; error path unit-tested through the internal helpers).
- **Task 2 — fixture corpus:** six fixtures under testdata/c4d/ — external-types (all four C1 external variants), linkfrom (authored [[...linkFrom]] + bidirectional + none arrows), multiline-strings (multi-line basic + escaped-\n values + an 84-char word), rank-equal (rank = "equal" on every link), template-nested-use ([template.*] + [[platform.use]] D-16 + [[template.worker.use]] D-17, unit-nested site authored before top-level to match canonical instantiation order), unicode-strings (multi-byte names/descriptions + a long Cyrillic word; ASCII unit keys per D-07). Guards assert every fixture parses and passes expand -> peer.Resolve -> validate clean (the D-24 convert gate mirrored at model level) plus per-fixture coverage pins.
- **Task 3 — parity suite (the phase contract):** forward round-trip over the FULL corpus (testdata/, testdata/c4d/, cmd/c4drill/testdata/ top-level, skill/examples/ recursive minus 09-composed; 29 valid fixtures, tally-guarded > 15) asserting canonical TOML text equality AND exact Model equality; reverse direction asserting canonical C4D stability across the loop; 09-composed include graph converted file-by-file to .c4d twins with include paths rewritten to the twin extension (D-26 groundwork) expanding to the SAME model; invalid fixtures asserting refusal through the gate; canonicalDOT render equivalence for the pinned 10-fixture set through the real view/graph/render composition (no t.Parallel, WASM rule).
- **Four gaps found and fixed** (details under Deviations): template root type text form, ${param} peers, arrow-glyph mapping, parser deep-subunit nondeterminism.
- `go test ./...` 14/14 packages green (verified stable across repeated runs), golangci-lint 0 issues, gofmt clean on all touched files.

## Task Commits

Each task followed TDD RED -> GREEN:

1. **Task 1: canonsrc normalizers** — `dcc977c` (test, RED: compile-level, package undefined) / `697b526` (feat, GREEN)
2. **Task 2: fixture corpus + guards** — `b24d2b1` (test, RED: fixtures missing) / `8ad9283` (feat, GREEN)
3. **Task 3: parity suite + gap fixes** — `aa3a546` (test, RED: 20 failing subtests) / `099aa19` (fix, GREEN)

## TDD Gate Compliance

All three `test(35-06)` commits precede their implementation commits in git order (dcc977c < 697b526; b24d2b1 < 8ad9283; aa3a546 < 099aa19). Task 1 RED was compile-level (35-04/35-05 precedent). Task 2 RED was runtime-level (missing fixture files). Task 3 RED was runtime-level with 20 failing subtests across all four suites — genuine gap evidence, not compile failure; the GREEN commit fixed the gaps in the offending plans' files per the plan's contract-enforcer mandate.

## Files Created/Modified

- `internal/testutil/canonsrc/canonsrc.go` (733 lines; min_lines 60 met) — NormalizeTOML/NormalizeC4D, canonical serializers, value forms, D-22 default rules
- `internal/c4d/parity_test.go` (840 lines; min_lines 80 met) — fixture guards, corpus walker, round-trip suites, composed-graph conversion, invalid-refusal, render equivalence, canonicalModel helper
- `internal/c4d/grammar/c4d.peg` + `parser_gen.go` — TemplateBodyPart/TemplateBodyStmt/TemplateTypeStmt, templateType carrier, PeerSeg/ParamToken, duplicate-type error
- `internal/c4d/emit_c4d.go` — emitTemplateC4D renders the root `type:` statement
- `internal/c4d/frommodel.go` — exact arrow forms (arrowOptionValue/arrowGlyphFor rewrite), doc updates
- `internal/c4d/tomodel.go` — `->`/`<-` map to the omitted default; UnitOrder initialized as an empty slice (parser.Parse nil-ness parity)
- `internal/parser/parser.go` — recordHandAuthored records every ancestor pair (any depth); parseUnitWithOrder carries the full lookup path
- `internal/c4d/tomodel_test.go`, `emit_test.go` — glyph mappings re-pinned; TestToModelTemplateTypeStatement + TestToModelParametrizedPeerRoundTrip added
- `testdata/c4d/*.toml` — six fixtures

## Decisions Made

- The plan's schematic "a = x / b = 2" normalizer case is expressed over real c4drill documents: top-level bare scalars are outside the Model (parser.Parse skips non-table expressions), so a literal pair would compare two empty models — vacuous
- NormalizeC4D's canonical text mirrors grammar shape (unquoted structural keywords, quoted values) so the fixpoint property is testable against the real parser
- Render-equivalence fixtures run expand + peer.Resolve but not a validate REQUIREMENT: the classic pinned fixtures (testdata/valid.toml) contain by-design orphans, and the render path does not depend on validation
- Model comparisons run before any Validate call (mirrors append in map-iteration order — comparing validated models flaked before this rule)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Template root type had no C4D text form (35-04 deferred gap, closed)**
- **Found during:** Task 3 RED — 06-templates.toml, template_reserved.toml (root `type = "container"`) and the 09-composed graph failed Model and text round-trips; the converted 09-composed graph did not even validate (componentDb under a system root)
- **Issue:** FromModel recorded TemplateDecl.Body.Type but EmitC4D could not render it — the grammar had no template-root-type syntax, so non-default root types were lost in text
- **Fix:** Template-body `type: <TypeKeyword> [external]` statement (grammar TemplateBodyStmt admits TemplateTypeStmt ahead of the shared forms; unit bodies still reject `type:`; duplicates hard-error), EmitC4D renders it — exactly the deferred-items note's suggested option
- **Files modified:** internal/c4d/grammar/c4d.peg, parser_gen.go, emit_c4d.go; tomodel_test.go (TestToModelTemplateTypeStatement); 35-04-deferred-items.md (marked CLOSED)
- **Verification:** corpus round-trips green; TestToModelTemplateTypeStatement pins parse/duplicate/unit-rejection
- **Committed in:** 099aa19

**2. [Rule 1 - Bug] Parametrized link peers (${param}) were inexpressible in C4D**
- **Found during:** Task 3 — template_reserved/06-templates emitted C4D failed to re-parse (`-> ${bus}`); PeerRef accepted only path characters
- **Issue:** XC-03's load-bearing feature (parametrized template link peers) had no C4D spelling, so any template using it could not round-trip
- **Fix:** PeerRef segments accept ${param} tokens (PeerSeg/ParamToken), captured verbatim; mixed segments work
- **Files modified:** internal/c4d/grammar/c4d.peg, parser_gen.go; tomodel_test.go (TestToModelParametrizedPeerRoundTrip)
- **Committed in:** 099aa19

**3. [Rule 1 - Bug] `->` -> ArrowForward made C4D models render apart from TOML twins**
- **Found during:** Task 3 RED — template_basic render equivalence failed on a `dir=forward` edge-defaults attribute present only on the round-tripped side
- **Issue:** the renderer maps Arrow "" to no dir attribute and "forward" to dir=forward, so 35-05's explicit mapping produced a different render for the same document; the reverse-arrow `<-` mapping also moved Links{reverse} to LinksFrom{forward}, breaking Model equality (03-links.toml)
- **Fix:** `->` maps to the omitted default (Arrow ""), `<-` to LinksFrom{""}; forward/reverse ride `-> { arrow: X }`, non-default LinksFrom arrows ride `<- { arrow: X }` — every Model shape now round-trips exactly
- **Files modified:** internal/c4d/tomodel.go, frommodel.go; tomodel_test.go, emit_test.go re-pinned
- **Committed in:** 099aa19

**4. [Rule 1 - Bug, pre-existing] parser.Parse subunit order was nondeterministic at depth >= 3**
- **Found during:** Task 3 — multilevel.toml round-trips flipped subunit order between runs (five runs, three different orders)
- **Issue:** recordHandAuthored ignored table paths deeper than two segments, and the recursion looked subunit orders up by SHORT names while keys were full paths — deep subunits fell into the raw-map fallback branch (Go map iteration). Previously masked by DI-1's order-insensitive canonicalDOT comparisons; fatal for the order-preservation contract
- **Fix:** every ancestor pair of a unit table path records (document order, any depth); parseUnitWithOrder carries the full dotted lookup path (template bodies keep their relative keying, which already matched)
- **Files modified:** internal/parser/parser.go
- **Verification:** repeated full-suite runs stable; all pre-existing parser/cmd/golden tests green unchanged
- **Committed in:** 099aa19

**5. [Rule 3 - Plan inaccuracy] Task 2's fixture guards live in parity_test.go**
- **Found during:** Task 2 — the plan's `<files>` list only names the six fixtures, but its `<verify>` runs `go test ./internal/c4d/ -run TestFixtures`, so the guard test must exist at Task 2's commit
- **Fix:** created parity_test.go with the fixture guards in Task 2 (RED), extended it in Task 3 as planned
- **Committed in:** b24d2b1, 8ad9283

---

**Total deviations:** 5 auto-fixed (4 bugs — one pre-existing — and 1 plan inaccuracy)
**Impact on plan:** All acceptance criteria and must-have truths met; the fixes are the parity contract doing its job — the plan's action text explicitly directs fixing exposed gaps in the offending plans' files.

## Threat Model Disposition

| Threat | Disposition | Where |
|--------|-------------|-------|
| T-35-06-01 (corpus walker tampering) | mitigated | filepath.WalkDir (never follows symlinked dirs), *.toml filter, repo-root-anchored pinned path space |
| T-35-06-02 (render-equivalence DoS) | accepted (as planned) | serial render tests (no t.Parallel, nolint:paralleltest per repo rule); pinned 10-fixture set |
| T-35-06-03 (parity failures repudiated) | mitigated | every round-trip failure names its fixture path as a failing subtest; no silent skips — the only skip is 09-composed (converted whole by its own test) with an explicit comment |

## Issues Encountered

- C4D triple-quoted literals capture their body VERBATIM (no leading-newline trim, unlike TOML multi-line strings) — the canonsrc three-form test's first spelling was wrong and corrected against the real grammar behavior
- The plan's Model-equality wording needed operational care: raw require.Equal works only after the glyph-mapping fix; the canonicalModel helper (defaults filled on deep clones) keeps m1-vs-m3 comparisons honest where fixtures author defaults implicitly
- TestComposedGraphRoundTrip initially flaked (4-of-8 runs) because Validate populates mirror LinksFrom in map-iteration order — fixed by comparing pre-validation models and validating fresh parses (documented pattern)
- EmitTOML writes template-body uses root-first while ToModel reads subunit-uses first — a latent order asymmetry for templates mixing root- and subunit-level body uses; no corpus fixture mixes them (noted for 35-07/35-08 awareness)

## Known Stubs

None. The 35-04 template-root-type deferred gap is fully closed (text level included; deferred-items note updated). The C4D edge-label pipe constraint (desc-only values containing `|`) remains a documented authoring constraint — the canonical quoted label keeps tech|desc pipes round-trip safe, and all corpus fixtures are pipe-free.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- The parity contract is enforced: 35-07's convert composes EmitTOML/FromModel/EmitC4D under exactly the guarantees these suites pin (canonical-equal round-trips, order preservation, D-24 refusal on invalid input)
- 35-08's fmt inherits stable canonical orders and the grammar's template-root-type surface
- The parser determinism fix makes every downstream consumer's UnitOrder/SubunitOrder stable — 35-09's CLI dispatch and any future golden work benefit
- No blockers; `go test ./...` 14/14 green, golangci-lint 0 issues

## Self-Check: PASSED

All 10 created files exist on disk; all 6 task commit hashes (dcc977c, 697b526, b24d2b1, 8ad9283, aa3a546, 099aa19) verified in git log.

---
*Phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition*
*Completed: 2026-08-14*
