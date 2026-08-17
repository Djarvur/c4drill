---
phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition
plan: 04
subsystem: parser
tags: [c4d, dsl, emitter, canonical-order, tomlemit, frommodel, compact-leaf, comments, determinism]

# Dependency graph
requires:
  - phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition (Plan 01)
    provides: typed AST (internal/c4d/ast), c4d.Parse AST front-end, comment/Pos-carrying nodes (D-32)
  - phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition (Plan 03)
    provides: TemplateDecl/UseStmt/IncludeStmt nodes, composition grammar surface
provides:
  - EmitTOML(m *parser.Model) (string, error) — deterministic Model->TOML in the fixed D-23 canonical field order, fixture-shaped tables
  - FromModel(m *parser.Model) *ast.Document — the canonical Model->AST inverse of 35-05's toModel (comments nil by design)
  - EmitC4D(doc *ast.Document) string — comment-aware compact-leaf C4D printer (D-33), fixpoint-stable at the AST level
  - quoteTOML/quoteC4D escaping helpers + c4dBarewordSafe canonical literal rule (T-35-04-02)
affects: [35-05 tomodel (FromModel is its inverse; Parse API sweep includes emit_test.go), 35-06 canonsrc/round-trip, 35-07 convert, 35-08 fmt (comment placement machinery)]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Self-consistent comment placement: split attached comments Pos-aware (a comment at/below its statement's line is the same-line tail), then migrate each statement's first lead to the untailed predecessor — exactly where the grammar's StmtEnd re-attaches it. This is the only rendering family for which emit(parse(emit(doc))) is byte-identical."
    - "Kind-preserving literal rendering: EmitC4D renders per Literal.Kind so parse(emit) reproduces the same kinds; FromModel assigns kinds via the same bareword-safety rule the printer uses, so convert output is canonical AND fixpoint-stable."
    - "Canonical section order both emitters share: properties, units (UnitOrder/SubunitOrder recursive), templates (sorted names), uses, includes; [[use]] params sorted by key (the Model stores them as a map)."

key-files:
  created:
    - internal/c4d/emit_toml.go
    - internal/c4d/emit_c4d.go
    - internal/c4d/frommodel.go
    - internal/c4d/emit_test.go
  modified: []

key-decisions:
  - "Comment placement is allocation-based, not all-above: the grammar's StmtEnd lets the first comment after a statement (across any whitespace, even blank lines) ride that statement as its tail, so rendering every comment above its statement makes comments migrate backward on every re-parse — no fixpoint exists in that form. The Pos-aware tail split + one-step migration renders the placement the parser reproduces verbatim (verified by probe before implementing)."
  - "[[use]] params emit sorted by key and template names sort alphabetically: the Model carries both as maps, sorted order is the only deterministic choice (D-23 pins field order, not arg/name order)"
  - "ArrowReverse Links render through the linkFrom statement form \"<-\" on the owning unit (the plan's literal glyph mapping); LinksFrom entries also emit \"<-\" — 35-05's toModel owns the inverse direction of that mapping"
  - "Top-level instantiations with a Parent land INSIDE that unit's block (D-16 C4D form of TOML's parent key); unresolvable parents fall back to top level so an instantiation is never dropped"
  - "Width/Height are not emitted: outside the D-23 canonical field set and outside the C4D grammar"
  - "Template root Name rides as a body `name` field (a legal FieldKey); template root Type/External are recorded on Body but NOT rendered — the grammar has no template-root-type syntax (known gap, deferred to 35-05/06)"

patterns-established:
  - "renderedStmt render-list + allocateTails: one statement-list abstraction drives document, block and template-body rendering (kind-tagged for blank rules, closure renderers per statement)"
  - "Mirror-link skip via Link.IsMirror() in BOTH emitters; test constructs mirrors through validator.Validate (the only way to observe Mirror outside package model)"

requirements-completed: [D-23, D-33]

# Metrics
duration: 21min
completed: 2026-08-14
---

# Phase 35 Plan 04: Emitters Summary

**EmitTOML renders any \*parser.Model as deterministic TOML in the fixed D-23 canonical field order (fixture-shaped tables), while FromModel + EmitC4D push Models through the typed AST into compact-leaf C4D (D-33) with a self-consistent comment placement that makes emit(parse(emit(doc))) byte-identical**

## Performance

- **Duration:** 21 min (16:51–17:12 UTC)
- **Started:** 2026-08-14T16:51:46Z
- **Completed:** 2026-08-14T17:12:47Z
- **Tasks:** 2/2
- **Files modified:** 4 (all created; plus 1 deferred-items note)

## Accomplishments

- **EmitTOML (D-23):** fixed canonical orders — unit fields `type, name, description, technology, reference, color, style, border, edges, expanded`; link fields `peer, arrow, rank, color, style, technology, description, labelPosition, length`; properties `name, description, color, style, border, edges, lineLength, expanded`. Empty values omitted; sections `[properties]` -> unit tables (UnitOrder/SubunitOrder, dotted subunit paths, links BEFORE subunit tables for TOML table-context safety) -> `[template.<name>]` (params + body + `[[template.<name>.<path>.use]]` site form) -> `[[use]]` (parent key omitted when empty) -> `[[include]]` (once only when set). Newline values pinned to the escaped-`\n` single-line basic string and re-parse to the identical value (D-06); `quoteTOML` escapes all control bytes (T-35-04-02); double-emit is byte-identical.
- **FromModel:** the Model->AST inverse — units in UnitOrder with external types split (`systemExternal` -> `system` + `external`), links/linksFrom to edge statements with the four-glyph inverse, templates to declarations (root name as body field), instantiations placed inside their parent unit's block (D-16) or at top level, template-body uses under their root-relative subunit path (D-17), includes verbatim. Comments nil by design (convert loses trivia; only fmt preserves, D-32).
- **EmitC4D (D-33):** compact one-line leaves (no subunits/edges/uses, <=3 single-line-able fields — `db: db { description: cache }`), multi-line nesting at 2 spaces per depth, D-23 field and option sorting, D-09 edge inverse (`"tech | desc"`, `"tech |"` trailing pipe, desc-only without pipe, trailing `{ color: red }` option block), template/use/include statements, one blank line between top-level units and across kind changes. Newline values emit as triple-quoted blocks that force multi-line units (D-06 inverse).
- **Fixpoint (T-35-04-01):** AST-level `EmitC4D(Parse(EmitC4D(doc)))` is byte-identical for a corpus covering every statement kind plus comments — achieved through Kind-preserving literal rendering and the allocation-based comment placement (see deviations).
- `go test ./...` 14/14 packages green, golangci-lint 0 issues; the emitted TOML of testdata/valid.toml and skill/examples/06-templates.toml both re-parse through parser.Parse with templates/instantiations intact.

## Task Commits

Each task followed TDD RED → GREEN:

1. **Task 1: emit_toml.go — Model to TOML canonical emission** — `7518715` (test, RED: compile-level, c4d.EmitTOML undefined) / `37c513f` (feat, GREEN)
2. **Task 2: frommodel.go + emit_c4d.go — Model to AST to C4D text** — `8a0ddf6` (test, RED: compile-level, FromModel/EmitC4D undefined) / `db704c8` (feat, GREEN)

## TDD Gate Compliance

Both tasks committed `test(35-04)` RED before their `feat(35-04)` GREEN commits in git order (7518715 < 37c513f; 8a0ddf6 < db704c8). RED evidence was compile-level (the exported functions did not exist), matching the 35-03 precedent. Two committed test expectations were corrected during Task 2 GREEN where the fixture shape contradicted the pinned style rules (a 1-field subunit is a D-33 leaf and must emit one-line; a hand-built unsafe KindBareword is out of the Kind-preserving contract) — the corrected tests assert the same plan behaviors, weakened nowhere.

## Files Created/Modified

- `internal/c4d/emit_toml.go` (396 lines; min_lines 80 met) — EmitTOML, section emitters, quoteTOML/quoteTOMLArray, sortedTemplateNames/sortedParamKeys, joinDotted
- `internal/c4d/emit_c4d.go` (667 lines; min_lines 80 met) — EmitC4D, renderedStmt/stmtComments render-list machinery, allocateTails/splitStmtComments, unit/properties/template/edge/use/include renderers, literalC4D/quoteC4D/c4dBarewordSafe/includeBarewordSafe, D-23 rank tables
- `internal/c4d/frommodel.go` (333 lines; min_lines 60 met) — FromModel, properties/unit/template/use builders, edgeStmtFromLink + arrowGlyphFor, splitExternalType, findUnitByPath, literalFor
- `internal/c4d/emit_test.go` (854 lines) — 13 TestEmitTOML* + 15 TestFromModel*/TestEmitC4D*/mirror functions covering every behavior bullet of both tasks

## Decisions Made

- Comment placement by allocation (the load-bearing design decision — see Deviations #1): Pos-aware same-line tails plus one-step migration to the untailed predecessor is the unique self-consistent rendering under the 35-01 grammar's StmtEnd attachment rule
- Determinism sources: UnitOrder/SubunitOrder walks, sorted template names, sorted [[use]] param keys — never map iteration
- Nested-use placement: parented instantiations emit inside the parent's block (the C4D-native D-16 form) rather than as top-level `[[use]]` with a parent key; the TOML emitter uses the top-level `[[use]]` + parent key form per the plan
- Mirror links skipped in both emitters; test constructs them via validator.Validate (only Mirror observer outside package model)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Comment emission could not be "own line before the statement" and stay fixpoint-stable**
- **Found during:** Task 2 GREEN (TestEmitC4DFixpointASTLevel failed)
- **Issue:** The plan's behavior bullet says attached comments emit on their own line immediately before their statement. But the grammar's `StmtEnd <- _ cm:Comment Term` lets `_` eat newlines, so the FIRST comment after any statement (even across blank lines) re-attaches as THAT statement's tail — a probe confirmed `a: system { }\n# c\nb: ...` puts `# c` on `a`. Rendering comments above their statement therefore migrates every comment one statement backward per re-parse: no fixpoint exists in that form, and the pinned fixpoint test can never pass.
- **Fix:** Allocation-based placement — each statement splits its comments Pos-aware (a comment at or below the statement's own line is its same-line tail, gofmt semantics), then each statement's first lead migrates to the previous statement when that one has no tail — exactly where re-parsing attaches it. The rendered text re-parses to the placement the emitter assumed, so emit(parse(emit(doc))) is byte-identical. Leads still emit on their own line immediately before their statement; only tails moved (same-line, after).
- **Files modified:** internal/c4d/emit_c4d.go, internal/c4d/emit_test.go
- **Verification:** TestEmitC4DComments (exact placement + its own fixpoint assertion), TestEmitC4DFixpointASTLevel (byte-identity over a corpus with comments)
- **Committed in:** db704c8

**2. [Rule 1 - Bug] properties fields were not sorted canonically (lint-caught)**
- **Found during:** Task 2 lint pass
- **Issue:** `golangci-lint` flagged `propertyFieldRank` unused — the refactored emitPropertiesC4D looped over `props.Fields` in source order, silently dropping D-23 normalization for the properties block
- **Fix:** Sort via `sortedFields(props.Fields, propertyFieldRank)`
- **Files modified:** internal/c4d/emit_c4d.go
- **Verification:** golangci-lint 0 issues; TestFromModelValidFixtureToC4D pins the property field order
- **Committed in:** db704c8

**3. [Rule 3 - Blocking] Mirror-link test needed a real producer**
- **Found during:** Task 2 test writing
- **Issue:** `model.Link.Mirror` is unexported with no setter — a black-box test cannot construct a mirror link, so the "Mirror links are never emitted" bullet was untestable as written
- **Fix:** Build a valid two-unit model and run `validator.Validate` (exported), whose populateIncomingLinks synthesizes a mirror `LinksFrom` entry in place; assert both emitters skip it
- **Files modified:** internal/c4d/emit_test.go
- **Committed in:** 8a0ddf6

**4. [Rule 1 - Bug] `gofmt -w internal/c4d/` reformatted two pre-existing files**
- **Found during:** Task 2 formatting pass
- **Issue:** composition_test.go and grammar/reserved.go were unformatted before this plan (gofmt-unclean but lint-clean); my directory-wide gofmt touched them — out of scope
- **Fix:** `git checkout --` both files; tree clean before the Task 2 commit
- **Committed in:** n/a (reverted)

---

**Total deviations:** 4 auto-fixed (2 bug, 1 blocking, 1 accidental-formatting reverted)
**Impact on plan:** All acceptance criteria and must-have truths met. Deviation #1 is a design refinement forced by the grammar's actual attachment semantics — it strengthens the pinned fixpoint contract rather than weakening the comment bullet.

## Threat Model Disposition

| Threat | Disposition | Where |
|--------|-------------|-------|
| T-35-04-01 (emitter determinism tampering) | mitigated | Order-slice walks only (UnitOrder/SubunitOrder/sorted names/keys); determinism test (double-emit byte-equal) + AST-level fixpoint test |
| T-35-04-02 (quoted string injection) | mitigated | quoteTOML escapes every control byte + quotes/backslashes; quoteC4D escapes `"`/`\`/newlines; unit names always quoted so `}` cannot break a block; bareword rule excludes every grammar stop char |
| T-35-04-03 (canonical order repudiability) | accepted | D-23 rank tables + assertOrdered tests pin the exact sequences |

## Issues Encountered

- The StmtEnd tail-attachment behavior (deviation #1) was diagnosed empirically with a throwaway probe test before redesigning the comment machinery — the failing fixpoint output alone showed comments migrating backward but not why
- Test-shape corrections during GREEN: the nested-indent fixture's subunit was a 1-field leaf (must emit one-line per D-33 — gave it an edge); the quoting test's third case used a hand-built unsafe KindBareword (out of the Kind-preserving contract — replaced with a KindQuoted comma value plus a FromModel subtest pinning literalFor's safety rule)
- `quoteC4D` leaves non-EOL control bytes raw inside quoted strings — the DoubleQuoted grammar accepts them (`!'"' !EOL .`), so no corruption is possible

## Known Stubs

None in code. One recorded format gap (see `.planning/phases/35-.../35-04-deferred-items.md`): the C4D grammar has no template-root-type syntax, so a TOML template's non-default root `type` is recorded on `TemplateDecl.Body.Type/External` by FromModel but not rendered by EmitC4D — TOML->C4D->TOML round-trips lose it. The 35-06 fixtures use the root-default `system`, and D-22's explicit-defaults normalization covers them; 35-05/35-06 own the resolution. Edge labels containing literal pipes are likewise unrepresentable in C4D (D-09 splits on the first pipe) — documented, fixtures are pipe-free.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `EmitTOML`, `FromModel`, `EmitC4D` are deterministic, fixture-shaped and fixpoint-stable; 35-07's convert composes them directly and 35-08's fmt reuses EmitC4D's comment placement
- 35-05 note: emit_test.go calls `c4d.Parse` as the AST entry (the 35-01 shape) — 35-05 Task 2's call-site sweep should switch it to `c4d.ParseAST` when Parse becomes Model-returning; the tests consume only *ast.Document so the sweep is mechanical
- No blockers; `go test ./...` 14/14 green, golangci-lint 0 issues

## Self-Check: PASSED

All 4 created files exist on disk; all 4 task commit hashes (7518715, 37c513f, 8a0ddf6, db704c8) verified in git log.

---
*Phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition*
*Completed: 2026-08-14*
