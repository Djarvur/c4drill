---
phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition
plan: 05
subsystem: parser
tags: [c4d, dsl, tomodel, parity-hub, parseast, inference-parity, duplicate-edges, include-dispatch, mixed-format]

# Dependency graph
requires:
  - phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition (Plan 01)
    provides: typed AST (internal/c4d/ast), comment/Pos-carrying nodes, grammar front-end
  - phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition (Plan 02/03 surface)
    provides: TemplateDef/Instantiation/IncludeDirective Model fields, UseStmt/TemplateDecl/IncludeStmt AST nodes
  - phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition (Plan 04)
    provides: FromModel's Body.Type template-root recording (the deferred gap this plan half-closes)
provides:
  - ToModel(doc *ast.Document) (*parser.Model, error) — the AST->Model conversion with parseUnitWithOrder's exact hook order (D-02/D-21 parity hub closed)
  - c4d.Parse/c4d.ParseFile re-signatured to (*parser.Model, error) — the DSL parses DIRECTLY to the Model hub (D-21)
  - c4d.ParseAST/c4d.ParseASTFile — the exported AST-level entries fmt (35-08) and canonsrc (35-06) compile against (Parse == ToModel(ParseAST))
  - parser.DefaultTypeForParent / parser.InferGenericType — exported inference hooks (D-02)
  - include.Resolve extension dispatch — mixed .toml/.c4d include graphs (D-26), unknown extensions fail closed (T-35-05-01)
affects: [35-06 canonsrc/round-trip (ParseAST + Model twin baselines), 35-07 convert, 35-08 fmt, 35-09 CLI .c4d dispatch]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Front-end parity by hook reuse, not reimplementation: ToModel calls the SAME exported parser.DefaultTypeForParent/InferGenericType/model.Humanize sequence parseUnitWithOrder applies, so require.Equal on parsed Units is achievable by construction"
    - "Two-level parse API: Model-returning Parse composes exported ParseAST + ToModel, keeping the trivia-carrying AST reachable from other packages without exposing grammar internals"
    - "Glyph->Link mapping with `<-` -> LinksFrom{ArrowForward}: the incoming glyph's arrowhead sits at the owner (= the edge target in peer->owner orientation), exactly the shape the validator's mirror of `peer -> owner` carries — authored linkFrom and mirrors stay indistinguishable"
    - "Two-phase use conversion: collect use statements during the unit/template walk, convert after the template registry is complete, so positional args can pair with declared params and forward template references work"

key-files:
  created:
    - internal/c4d/tomodel.go
    - internal/c4d/tomodel_test.go
    - internal/parser/inference_test.go
    - internal/include/resolve_test.go
    - internal/include/testdata/mixed_main.toml
    - internal/include/testdata/shared.c4d
    - internal/include/testdata/mixed_main.c4d
    - internal/include/testdata/shared.toml
  modified:
    - internal/parser/parser.go
    - internal/c4d/parse.go
    - internal/c4d/doc.go
    - internal/c4d/parse_test.go
    - internal/include/resolve.go

key-decisions:
  - "`->` maps to ArrowForward (the glyph IS the arrow spec); the parity twin therefore states `arrow = \"forward\"` explicitly on the TOML side — semantically identical to omitting the key, and exactly what D-22's 'explicit defaults may normalize away' anticipates"
  - "`<-` maps to LinksFrom with ArrowForward (mirror-consistent incoming edge); the `arrow` edge option overrides the glyph default and rides verbatim like TOML values"
  - "Type-led headers without id derive the unit key from the quoted Name (TOML quoted-table twin), falling back to the type keyword; duplicate derived keys hard-error"
  - "m.Instantiations order follows the AST's canonical sections (unit-nested uses before top-level uses) because the 35-01 AST does not preserve top-level statement interleaving; FromModel output round-trips this order exactly"
  - "Positional use args pair with same-file template params at conversion time; a template arriving via include is out of reach there, so positional args for it hard-error with a named-args suggestion"
  - "ToModel honors FromModel's Body.Type/External — the 35-04 template-root-type gap is closed at the Model level (text level remains for 35-06)"
  - "Extension gate lives in resolveOne before the parse: unsupported extension is its own error (naming .toml/.c4d), distinct from include-not-found wrapping"

patterns-established:
  - "parseEntryAny test harness: extension-switched entry parse + include.Resolve (the D-27 CLI dispatch mirrored in tests)"
  - "Committed mixed-format fixture pair under internal/include/testdata for both include directions"

requirements-completed: [D-02, D-10, D-11, D-21, D-26]

# Metrics
duration: 30min
completed: 2026-08-14
---

# Phase 35 Plan 05: C4D-to-Model Parity Hub Summary

**ToModel converts the typed AST into \*parser.Model through the exact inference/humanize hooks the TOML parser applies (require.Equal twin proof), c4d.Parse now returns the Model directly with ParseAST kept as the exported AST entry, duplicate edges/paths hard-error, and include.Resolve dispatches by extension so .toml/.c4d graphs mix freely**

## Performance

- **Duration:** 30 min (17:17–17:47 UTC)
- **Started:** 2026-08-14T17:17:25Z
- **Completed:** 2026-08-14T17:47:23Z
- **Tasks:** 3/3
- **Files modified:** 13 (8 created, 5 modified)

## Accomplishments

- **Task 1 — exported inference hooks:** `parser.DefaultTypeForParent` / `parser.InferGenericType` renamed-export in place with doc comments naming the C4D caller; parser behavior byte-identical (pure rename; `genericDbTypes` stays unexported)
- **Task 2 — ToModel + Model-returning Parse (D-21):** `ToModel(doc *ast.Document) (*parser.Model, error)` walks units recursively applying parseUnitWithOrder's exact hook order (empty type -> DefaultTypeForParent; InferGenericType; empty name -> Humanize). `c4d.Parse`/`ParseFile` re-signatured to `(*parser.Model, error)`; `ParseAST`/`ParseASTFile` keep the exported AST entry (comments attached) and `Parse == ToModel(ParseAST)` is pinned. External modifier -> the four `*External` variants with vocabulary check + Levenshtein suggestion (D-04). Edges: `->`/`<->`/`--` -> Links with the glyph's ArrowDirection, `<-` -> LinksFrom (D-08/D-10), option block -> Link fields with `arrow` override and integer `length`. Duplicate (unit, peer) pairs per list hard-error with the peer name and AST line (D-11). Templates -> TemplateDef with verbatim `${param}` tokens; use -> Instantiation for all three positions (top, unit-nested D-16, template-body D-17); include -> IncludeDirective. 35-01/35-03 test call sites swept to ParseAST/ParseASTFile
- **Parity proof:** the inference-parity test and the template_basic.toml C4D twin test assert `require.Equal` on parsed Unit/TemplateDef/Instantiation structs between the two front-ends; the D-10 twin test asserts identical peer.Resolve errors for unresolvable bare peers
- **Task 3 — mixed-format includes (D-26):** `resolveOne` gates on the included file's extension — `.c4d` -> `c4d.ParseFile`, `.toml` -> `parser.ParseFile`, anything else a hard `*parser.ParseError` naming the accepted extensions (T-35-05-01 fail-closed; no content sniffing). Cycle detection/depth cap/visited dedup unchanged across mixed graphs (T-35-05-02). Full-pipeline integration test: composed `.c4d` (properties + template + use + relative peers + `.c4d` include) through include.Resolve -> template.Expand -> peer.Resolve -> validator.Validate with zero errors and spot-checked substitutions/resolutions
- `go test ./...` 13/13 packages green, golangci-lint 0 issues, TestXC01_PipelineOrdering untouched and passing

## Task Commits

Each task followed TDD RED -> GREEN:

1. **Task 1: export inference helpers** — `22b6c3e` (test, RED: compile-level, names undefined) / `bc45fff` (feat, GREEN)
2. **Task 2: tomodel.go + Parse re-signature + call-site sweep** — `9d53b48` (test, RED: compile-level, ToModel/ParseAST undefined and Parse still AST-returning) / `2df7d1a` (feat, GREEN)
3. **Task 3: include extension dispatch + pipeline proof** — `5ee6615` (test, RED: ./shared.c4d include fails TOML parse; unknown-extension and mixed-cycle assertions fail) / `8b0b8af` (feat, GREEN)

## TDD Gate Compliance

All three `test(35-05)` commits precede their `feat(35-05)` GREEN commits in git order (22b6c3e < bc45fff; 9d53b48 < 2df7d1a; 5ee6615 < 8b0b8af). Tasks 1/2 RED was compile-level (exported symbols undefined — the 35-03/35-04 precedent). Task 3 RED was runtime-level: the mixed-format resolver tests and the pipeline integration test failed on the TOML-only include parse before the dispatch existed; the c4d-includes-toml direction was incidentally green at RED time (its included file is .toml, which parsed already) — the dispatch change is still pinned by the three failing tests in the same commit.

## Files Created/Modified

- `internal/c4d/tomodel.go` (849 lines; min_lines 100 met) — ToModel, buildTemplates/buildUnits/buildUnit/applyUnitFields/applyEdges/linkFromEdge/arrowFromGlyph/applyEdgeOptions, templateDefFromAST, instantiationFromUse (named + positional args), unitKey, externalVariant, applyProperties
- `internal/c4d/parse.go` — Parse/ParseFile return (*parser.Model, error); ParseAST/ParseASTFile exported AST entries; ParseFile attributes conversion errors with the file path
- `internal/parser/parser.go` — DefaultTypeForParent/InferGenericType exported (doc comments name both callers)
- `internal/include/resolve.go` — checkIncludeExtension gate + parseIncludedFile dispatch; package doc records D-26/T-35-05-01
- `internal/c4d/tomodel_test.go` (647 lines) — 12 test functions / 17 subtests covering every behavior bullet of Tasks 2 and 3
- `internal/parser/inference_test.go` — exported-hook value tables (97 lines)
- `internal/include/resolve_test.go` (120 lines) + 4 testdata fixtures — both mixed directions, unknown-extension, mixed-format cycle
- `internal/c4d/parse_test.go`, `internal/c4d/doc.go` — mechanical ParseAST sweep + layering doc

## Decisions Made

- Glyph mapping: `->`/`<->`/`--` carry their ArrowDirection explicitly; `<-` -> LinksFrom{ArrowForward} (the mirror-consistent incoming shape — see key-decisions); `arrow` option overrides verbatim
- Twin-test methodology: the TOML twin states `arrow = "forward"` where the fixture omits the key (the glyph makes the default explicit; identical parsed Link; D-22 normalization territory)
- Derived unit keys for type-led headers (Name, else type keyword) — documented in unitKey; duplicates hard-error
- Two-phase use conversion so the template registry is complete before positional-arg pairing; forward template references work like TOML's extractTemplates-then-extractUses
- Fail-closed TOML-parity checks added beyond the plan's explicit list: duplicate unit paths, duplicate field/property keys, scalar-vs-list shape checks, integer length/lineLength (see Deviations #2)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Plan inaccuracy] emit_test.go needed no c4d.Parse sweep**
- **Found during:** Task 2 call-site survey
- **Issue:** The plan (and 35-04's summary) expected 35-04's emit_test.go fixpoint test to call `c4d.Parse` as its AST entry; the sweep was budgeted for it
- **Fix:** Verified emit_test.go builds its AST corpus by hand (`&ast.Document{...}`) and calls only `parser.Parse` (TOML) — no c4d.Parse call sites existed there. The actual AST consumers were parse_test.go/composition_test.go (35-01/35-03); those were swept mechanically to ParseAST/ParseASTFile
- **Files modified:** internal/c4d/parse_test.go
- **Verification:** repo-wide grep shows zero remaining AST-shape asserts against c4d.Parse; all suites green
- **Committed in:** 2df7d1a

**2. [Rule 2 - Missing critical] conversion-time validation the plan implied but did not enumerate**
- **Found during:** Task 2 implementation
- **Issue:** The AST accepts shapes the TOML front-end rejects (duplicate unit tables, duplicate keys, list literals on scalar fields, non-numeric length/lineLength, `box external` which has no type variant). Converting them silently would produce Models the TOML twin could never produce — breaking the parity contract and downstream invariants (expand's pathTracker assumes no duplicate authored paths)
- **Fix:** Hard `*parser.ParseError` with AST lines for: duplicate unit paths (top-level and subunit), duplicate field/property keys, scalar-vs-list mismatches, non-integer length/lineLength, external modifier on non-externalizable or already-external types (FormatSuggestion against the vocabulary)
- **Files modified:** internal/c4d/tomodel.go, internal/c4d/tomodel_test.go (TestToModelDuplicateUnitError + subtests)
- **Verification:** go test ./internal/c4d/ green; golangci-lint 0 issues
- **Committed in:** 2df7d1a

**3. [Rule 3 - Blocking] godoclint/wrapcheck/lll friction in new files**
- **Found during:** lint passes (all three tasks)
- **Issue:** multi-godoc test packages, unwrapped external-package errors in test helpers/parseIncludedFile, table rows over 120 chars
- **Fix:** single package godoc kept (descriptive comment moved onto the helper), targeted `//nolint:wrapcheck` with justification (repo precedent in include_test.go), shortened table rows; tomodel.go split into sub-60-line functions instead of nolint
- **Files modified:** internal/c4d/tomodel.go, internal/parser/inference_test.go, internal/include/resolve_test.go, internal/include/resolve.go
- **Committed in:** bc45fff, 2df7d1a, 8b0b8af

---

**Total deviations:** 3 auto-fixed (1 plan inaccuracy, 1 missing critical validation, 1 blocking lint friction)
**Impact on plan:** All acceptance criteria and must-have truths met; deviations strengthen the parity contract rather than altering it.

## Threat Model Disposition

| Threat | Disposition | Where |
|--------|-------------|-------|
| T-35-05-01 (extension dispatch tampering) | mitigated | checkIncludeExtension gate BEFORE any parse: unknown extension = hard ParseError naming .toml/.c4d; no fallback parsing, no content sniffing (TestResolveUnknownExtensionHardError) |
| T-35-05-02 (mixed-graph DoS) | mitigated | traversal untouched: maxIncludeDepth=100 + visited-set + cycle stack apply across mixed formats (TestResolveMixedFormatCycleFatal) |
| T-35-05-03 (Model construction elevation) | accepted (as planned) | ToModel builds in-memory structs, no I/O; unknown types fail closed at the grammar and the external vocabulary check |

## Issues Encountered

- The AST's canonical section split (Units/UseStmts as separate slices) loses top-level statement interleaving, so m.Instantiations order is canonical (unit-nested before top-level) rather than source-order — documented in TestToModelUseParents; FromModel output reproduces this order exactly, and arbitrary reordering falls to 35-06's canonical normalization (D-22)
- Template bodies have no root-type syntax, so a C4D template root defaults to system — the integration fixture instantiates at top level for this reason; the Model-level inverse (Body.Type) is honored and pinned by TestToModelTemplateRootTypeFromModel
- A first draft of the integration test used a .toml include (green before the dispatch existed); restructured around a .c4d include so the test genuinely proves Task 3's behavior

## Known Stubs

None. The 35-04 template-root-type gap is half-closed (Model level; text level remains with 35-06 per the updated 35-04-deferred-items.md note).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `c4d.ParseFile` yields a Model that flows through the entire existing pipeline — 35-06 can build canonsrc round-trips on ParseAST + the twin baselines this plan pins
- ParseAST/ParseASTFile are exported and comment-preserving — 35-08's fmt compiles against them from cmd/c4drill
- include.Resolve accepts mixed graphs — 35-07's convert --follow-includes and 35-09's CLI dispatch (D-27) reuse the same extension rule
- No blockers; `go test ./...` 13/13 green, golangci-lint 0 issues

## Self-Check: PASSED

All 8 created files exist on disk; all 6 task commit hashes (22b6c3e, bc45fff, 9d53b48, 2df7d1a, 5ee6615, 8b0b8af) verified in git log.

---
*Phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition*
*Completed: 2026-08-14*
