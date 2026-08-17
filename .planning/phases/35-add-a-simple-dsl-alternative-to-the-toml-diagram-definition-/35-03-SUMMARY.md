---
phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition
plan: 03
subsystem: parser
tags: [c4d, dsl, peg, pigeon, grammar, templates, use, include, reserved-words, levenshtein]

# Dependency graph
requires:
  - phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition (Plan 01)
    provides: core grammar/AST/front-end layout (grammar + ast subpackages), pigeon v1.3.0 chain, *parser.ParseError contract
  - phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition (Plan 02)
    provides: Instantiation{Template, Parent, Params} shape the C4D UseStmt maps onto; isBuiltinField "use" reservation
provides:
  - Full C4D authoring surface at parse level — template declarations with params, use in all three positions, include with once, both list forms (D-13, D-14, D-15, D-16/D-17 grammar surface)
  - grammar.ReservedKeywords()/CheckUnitID/UnknownFieldError — the D-19 reserved-word machinery with Levenshtein suggestions, reusable by tomodel/fmt plans
  - ${param} tokens captured verbatim in literal values (substitution stays the TemplateDef/Expand contract)
affects: [35-04/05 tomodel + emitters (consume TemplateDecl/UseStmt/IncludeStmt), 35-07 fmt (comments + Positions on new nodes), README/skill docs]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Pigeon action errors as quality errors: an action returning (value, err) records err in pigeon's errList WITH position but does not fail the match — the value parses, the parse still returns the error (D-19 parse-level suggestions without a post-parse walk)"
    - "Brace-lookahead disambiguation: ReservedUnitId matches only unit-shaped reserved-led statements (&(_ \"{\")) so field statements whose value starts with a type keyword keep parsing as fields (Risk 2)"
    - "Single-sourced keyword set: ReservedKeyword is `kw:Ident &{ isReservedKeyword(kw) }` — reserved.go is the only list; UnitId exclusion, StmtStart lookahead and the error check all read it"
    - "Context-specific bareword stop sets: ListBareword stops at ']', ArgBareword stops at ')' ',' whitespace — parenthesized/bracketed contexts need their own single-token variant"

key-files:
  created:
    - internal/c4d/composition_test.go
    - internal/c4d/grammar/reserved.go
  modified:
    - internal/c4d/grammar/c4d.peg
    - internal/c4d/grammar/parser_gen.go
    - internal/c4d/ast/ast.go
    - internal/c4d/errors.go
    - internal/c4d/parse_test.go

key-decisions:
  - "Arg representation is ONE ordered []Arg{Name, Value ast.Literal}: named args carry the key, positional args get empty Name and pair with TemplateDecl.Params at expansion (documented in ast.go); positional values containing ':' must be quoted — the named form wins on a bare key: value shape"
  - "TemplateDecl.Body reuses *UnitNode (plan's offered choice): the template body IS a unit body, so Fields/Edges/Subunits/UseStmts carry it via the same BodyPart machinery — full unit grammar by construction (D-13)"
  - "Reserved-keyword errors fire from grammar actions, not a post-parse walk: pigeon records action errors with position and returns them even when the value parses, so c4d.Parse surfaces them through the existing wrapPigeonError contract (Line extracted from the position prefix)"
  - "reserved.go lives in internal/c4d/grammar (not internal/c4d as planned): peg actions compile into package grammar; package c4d imports grammar, so the helper on the c4d level would be an import cycle — same forced layout as 35-01's subpackage split"
  - "Exact collisions suggest the reserved word itself (FormatSuggestion over the full candidate list, per the plan's message template); near-miss ids like descripton remain legal"

patterns-established:
  - "ReservedUnitId placed BEFORE FieldStmt in statement dispatch (a reserved-led unit-shaped statement must beat FieldStmt's PEG choice commitment) with brace lookahead as the pollution guard"
  - "UnknownFieldStmt as the LAST statement alternative: an Ident ':' Literal that survives every legal form is by definition an error, its action always errors"

requirements-completed: [D-13, D-14, D-15, D-19]

# Metrics
duration: 24min
completed: 2026-08-14
---

# Phase 35 Plan 03: Composition Grammar + Reserved Words Summary

**Template/use/include statements complete the C4D grammar (use in all three positions, ${param} tokens verbatim, both list forms) and D-19 reserved-word collisions become hard *parser.ParseError's with Levenshtein suggestions fired from grammar actions via pigeon's positioned error list**

## Performance

- **Duration:** 24 min (16:17–16:41 UTC)
- **Started:** 2026-08-14T16:17:20Z
- **Completed:** 2026-08-14T16:41:12Z
- **Tasks:** 2/2
- **Files modified:** 7 (2 created, 5 modified)

## Accomplishments

- D-13: `template name(p1, p2) { ... }` declarations parse with the FULL unit grammar in the body (fields, edges with relative peers, nesting at any depth, use statements); `${param}` tokens ride literal values verbatim — a new TemplateToken bareword component keeps `{`/`}` from truncating values, and substitution remains the TemplateDef/Expand contract (asserted by test)
- D-13/D-16/D-17: `use name(args)` parses at top level (Document.UseStmts), inside unit blocks (UnitNode.UseStmts — the enclosing unit is the parent) and inside template bodies (Body.UseStmts — parent relative to the template unit root); one ordered `[]Arg{Name, Value}` representation serves both named and positional forms
- D-14: `include path` parses with bare or quoted paths (`./`, `../`, `/` prefixes, `.c4d`/`.toml` extensions) and the optional `once` modifier; the path text is captured as written — resolution stays include.Resolve's job
- D-15: `expanded: [c1, c2]` inline and one-per-line bracket forms both produce `Literal{Kind: KindList, List: [...]}` in unit bodies (machinery from 35-01, now pinned at the unit-field level)
- D-19: all 19 reserved words (14 isBuiltinField strings copied verbatim from internal/parser/parser.go + once/use/include/template/properties) error as unit ids in BOTH header forms — `unit id "description" is a reserved word (did you mean "description"?)` — as *parser.ParseError with the DSL-native Line; unknown field keys error with reserved+field-key suggestion candidates; near-miss ids stay legal
- Every failure remains a *parser.ParseError through the unchanged wrapPigeonError contract; Memoize + MaxExpressions(1M) untouched (T-35-03-02)

## Task Commits

Each task followed TDD RED → GREEN:

1. **Task 1: template/use/include grammar + AST nodes + list forms** - `6267710` (test, RED) / `53b6e8d` (feat, GREEN)
2. **Task 2: Reserved-keyword enforcement with Levenshtein suggestions** - `c6e2352` (test, RED) / `ba80e3a` (feat, GREEN)

## TDD Gate Compliance

Both tasks committed `test(35-03)` RED before their `feat(35-03)` GREEN commits in git order. RED evidence: Task 1 failed at compile level (Document.Templates/UseStmts/Includes, ast.Arg undefined); Task 2 failed at compile level (grammar.ReservedKeywords undefined) plus behaviorally (`description: system { }` parsed without error before the enforcement existed). Two guard-style tests (near-miss-legal, field-statements-still-parse) passed pre-implementation by design — they pin existing behavior against false positives, and their companion error tests carried the RED.

## Files Created/Modified

- `internal/c4d/grammar/c4d.peg` - TemplateDecl, ParamList, UseStmt, ArgList, Arg, ArgValue/ArgBareword, IncludeStmt/IncludeValue/OnceModifier, TemplateToken bareword component, ReservedKeyword predicate form, ReservedUnitId, UnknownFieldStmt, StmtStart via ReservedKeyword
- `internal/c4d/grammar/parser_gen.go` - regenerated via `go generate ./internal/c4d`, committed
- `internal/c4d/grammar/reserved.go` - ReservedKeywords() (19), FieldKeywords(), CheckUnitID, UnknownFieldError (77 lines; parser import aliased c4dparser — pigeon declares its own `parser` type)
- `internal/c4d/ast/ast.go` - Document.{Templates,UseStmts,Includes}; TemplateDecl; IncludeStmt; Arg; UseStmt.Args; UseStmt placeholder comment replaced with the 1:1 Instantiation mapping doc
- `internal/c4d/composition_test.go` - 14 test functions (34 cases incl. 38 subtests in the reserved table): every Task 1/2 behavior bullet
- `internal/c4d/errors.go` - nestedParseErrorPrefix strip so action-level ParseErrors are not double-prefixed
- `internal/c4d/parse_test.go` - 35-01 reserved-keyword pin updated to id-collision forms (statements now parse by design)

## Decisions Made

- Arg canonical representation: ordered []Arg with named keys / empty-Name positional (documented in ast.go and the peg)
- TemplateDecl.Body = *UnitNode — full unit grammar by construction, one body machinery
- Action-level error wiring (not a post-parse walk) — pigeon records action errors with position while the match still succeeds, so quality errors surface without restructuring the front-end
- ReservedUnitId brace lookahead (`&(_ "{")`) is the false-positive guard: `description: system handles auth` stays a FieldStmt, only unit-shaped collisions error (Risk 2 pinned in both directions by tests)
- reserved.go in package grammar (import-cycle forced; plan path internal/c4d/reserved.go adjusted)
- An exact reserved collision suggests the word itself — the plan's message template over the full candidate list; the suggestion proves the machinery and message shape

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] reserved.go placed in internal/c4d/grammar, not internal/c4d**
- **Found during:** Task 2 implementation
- **Issue:** CheckUnitID/UnknownFieldError are called from c4d.peg actions, which compile into package grammar; package c4d imports grammar, so helpers at the planned path would create an import cycle
- **Fix:** internal/c4d/grammar/reserved.go (77 lines, min_lines 20 met); ReservedKeywords() stays exported for later plans
- **Files modified:** internal/c4d/grammar/reserved.go
- **Verification:** go build ./..., full suite green
- **Committed in:** ba80e3a

**2. [Rule 1 - Bug] internal/parser import aliased in reserved.go**
- **Found during:** Task 2 implementation
- **Issue:** pigeon's generated parser_gen.go declares its own package-level `parser` type — importing internal/parser unaliased in the same package collides
- **Fix:** `c4dparser "github.com/Djarvur/c4drill/internal/parser"` alias
- **Files modified:** internal/c4d/grammar/reserved.go
- **Committed in:** ba80e3a

**3. [Rule 1 - Bug] Argument barewords needed a paren-context stop set**
- **Found during:** Task 1 GREEN
- **Issue:** `use helper(x: 1)` mis-parsed — the value Bareword does not stop at `)`, so `1)` was consumed as the value and the arg list never closed
- **Fix:** ArgValue/ArgBareword (ListBareword pattern): argument barewords are single tokens stopping at `)` `,` whitespace; values needing spaces are quoted
- **Files modified:** internal/c4d/grammar/c4d.peg
- **Verification:** TestParseUseInsideTemplateBody, TestParseUseTopLevelPositional
- **Committed in:** 53b6e8d

**4. [Rule 2 - Missing critical] 35-01 reserved-keyword test pin updated to the new contract**
- **Found during:** Task 1 GREEN
- **Issue:** TestParseReservedKeywordsNotUnitIds pinned "use/include/template statements do not parse" — obsolete the moment the composition statements landed; leaving it would fail the suite and misdocument the surface
- **Fix:** The pin now asserts the id-COLLISION forms (`use: system { }`, `use { }`, ...) error — the D-19 invariant that survives both plans
- **Files modified:** internal/c4d/parse_test.go
- **Verification:** TestParseReservedKeywordsNotUnitIds green; Task 2's table supersedes it in coverage
- **Committed in:** 53b6e8d

**5. [Rule 1 - Bug] Doubled "parse error at line N:" prefix on action-level errors**
- **Found during:** Task 2 GREEN (manual error-surface review)
- **Issue:** CheckUnitID returns *parser.ParseError; pigeon renders it inside its positioned error list via Error(), and wrapPigeonError re-wrapped the result — "parse error at line 1: parse error at line 1: unit id ..."
- **Fix:** wrapPigeonError strips a leading nested `parse error at line N: ` from each positioned message
- **Files modified:** internal/c4d/errors.go
- **Verification:** manual sample run: `parse error at line 1: unit id "description" is a reserved word (did you mean "description"?)`; full suite green (existing error-contract tests unaffected)
- **Committed in:** ba80e3a

**6. [Rule 3 - Blocking] pigeon semantic predicate needs an explicit `return ..., nil`**
- **Found during:** Task 2 implementation
- **Issue:** `&{ isReservedKeyword(kw) }` generated a function body without a return ("missing return"); pigeon inlines predicate code verbatim into a `(bool, error)` function
- **Fix:** `&{ return isReservedKeyword(kw.(string)), nil }`
- **Files modified:** internal/c4d/grammar/c4d.peg
- **Committed in:** ba80e3a

---

**Total deviations:** 6 auto-fixed (3 blocking, 3 bug — one of them a prior-plan test pin)
**Impact on plan:** All acceptance criteria met; two artifact-path adjustments both forced by the 35-01 subpackage layout. No scope creep.

## Threat Model Disposition

| Threat | Disposition | Where |
|--------|-------------|-------|
| T-35-03-01 (reserved disambiguation tampering) | mitigated | Exact-match table over all 19 words in both header forms + near-miss-legal test + field-statements-still-parse guard (Risk 2 pinned both directions) |
| T-35-03-02 (composition combinatorics DoS) | mitigated | Memoize(true) + MaxExpressions(1M) unchanged in parse.go; list/arg forms bounded by input size |
| T-35-03-03 (suggestion output disclosure) | accepted | FormatSuggestion echoes only the author's own token against a fixed public keyword list (verified in sample output) |

## Issues Encountered

- Pigeon action semantics had to be verified from the generated code before designing the error wiring: action errors are recorded WITH position but do NOT fail the match (parseActionExpr adds to errList and returns ok=true) — this is what makes action-level quality errors work, and also why the brace-lookahead pollution guard matters
- Pigeon labels bind sequence values as []any — ReservedUnitId's rid needed the existing strVal type-scanning helper (35-01 lesson re-applied)
- `properties { }` at top level is a valid empty properties block, so the id-only reserved table skips exactly that one form

## Known Stubs

None - both tasks landed complete implementations. The remaining Plan-01 placeholders (`UnitNode.UseStmts`) are now populated by the grammar; no unwired data slots remain in the c4d packages.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- The grammar accepts the full C4D authoring surface (D-01..D-19); Plans 04/05 can convert `*ast.Document` (Templates/UseStmts/Includes included) to `*parser.Model` — the UseStmt → Instantiation mapping is documented 1:1 in ast.go
- grammar.ReservedKeywords()/FieldKeywords() are exported for tomodel/fmt validation reuse
- `go test ./...` 14/14 packages green, golangci-lint 0 issues; no blockers

## Self-Check: PASSED

Both created files exist on disk (internal/c4d/composition_test.go, internal/c4d/grammar/reserved.go); all 4 task commit hashes (6267710, 53b6e8d, c6e2352, ba80e3a) verified in git log.

---
*Phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition*
*Completed: 2026-08-14*
