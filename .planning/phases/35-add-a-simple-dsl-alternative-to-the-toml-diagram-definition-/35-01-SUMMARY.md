---
phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition
plan: 01
subsystem: parser
tags: [c4d, dsl, peg, pigeon, parser, ast, grammar]

# Dependency graph
requires:
  - phase: 34-label-formatting-fixes
    provides: stable render/wrap machinery (untouched by this plan)
provides:
  - PEG grammar for core C4D statements (units, nesting, fields, edges, properties, literals, comments)
  - pigeon-generated committed parser (internal/c4d/grammar) with go:generate regeneration chain
  - typed comment/position-aware AST (internal/c4d/ast) — the D-32 fmt groundwork
  - c4d.Parse/c4d.ParseFile front-end returning (*ast.Document, error) with *parser.ParseError contract
  - tools.go pigeon pin (build-time-only dependency pattern)
affects: [35-02 tomodel/emitters, 35-03 reserved-word errors, 35-07 fmt, cmd/c4drill dispatch]

# Tech tracking
tech-stack:
  added: ["github.com/mna/pigeon v1.3.0 (build-time code generator, tools.go pattern)"]
  patterns:
    - "Layered C4D front-end: c4d.peg (package grammar) -> typed AST (package ast) -> c4d.Parse front-end; grammar actions build AST nodes in match order"
    - "Trivia-in-tree comment attachment: leading comments ride the following statement, same-line tails ride the preceding one, orphans ride the enclosing node (D-32)"
    - "Pigeon error strings are the stable error surface: parse 'line:col (offset)[: rule X]:' prefixes into parser.ParseError (pigeon error types are unexported)"

key-files:
  created:
    - internal/c4d/grammar/c4d.peg
    - internal/c4d/grammar/parser_gen.go
    - internal/c4d/grammar/doc.go
    - internal/c4d/ast/ast.go
    - internal/c4d/parse.go
    - internal/c4d/errors.go
    - internal/c4d/doc.go
    - internal/c4d/parse_test.go
    - tools.go
  modified:
    - go.mod
    - go.sum

key-decisions:
  - "pigeon pinned at v1.3.0, not v1.0.0: v1.0.0 lacks the -nolint flag the lint-gate acceptance criteria require (verified empirically)"
  - "Generated parser lives in subpackage internal/c4d/grammar: pigeon emits package-level Parse/ParseFile that would collide with the required typed c4d.Parse/ParseFile in one package"
  - "AST lives in subpackage internal/c4d/ast: grammar actions must construct typed nodes without an import cycle (c4d -> grammar -> ast)"
  - "Same-line trailing comments attach to the PRECEDING statement (gofmt semantics); orphan comments to the enclosing node's trailing list"
  - "Statement terminator accepts an implicit separator when another statement starts (&StmtStart), making one-line nested blocks parse; ';' still required after field values"

patterns-established:
  - "tools.go build-time dependency pin (first use in repo — pigeon D-20)"
  - "go:generate forwarding chain: internal/c4d/doc.go -> grammar/doc.go -> 'pigeon -o parser_gen.go -nolint c4d.peg'"
  - "Defensive action helpers (toAnySlice/toComments/strVal/boolVal) that scan pigeon captures by type instead of relying on sequence value shapes"

requirements-completed: [D-01, D-02, D-03, D-04, D-05, D-06, D-07, D-08, D-09, D-12, D-18, D-20]

# Metrics
duration: 42min
completed: 2026-08-14
---

# Phase 35 Plan 01: C4D Parser Foundation Summary

**Pigeon PEG grammar for the core C4D DSL (units, nesting, arrows, pipe shorthand, properties, literals, comments) producing a typed position/comment-aware AST via a committed generated parser, with the repo's \*parser.ParseError contract on every failure**

## Performance

- **Duration:** 42 min (15:03–15:45 UTC)
- **Started:** 2026-08-14T15:03:08Z
- **Completed:** 2026-08-14T15:45:31Z
- **Tasks:** 3/3
- **Files modified:** 11 (9 created, 2 modified)

## Accomplishments

- Pigeon toolchain: `github.com/mna/pigeon` pinned in go.mod via tools.go, committed generated parser, one-command regeneration (`go generate ./internal/c4d`) proven diff-free (D-20)
- Full core grammar: brace-block units with id-led/type-led/id-only headers, `external` modifier, exact TOML type keywords, all four arrows, the D-09 pipe-shorthand triple (desc-only / `tech |` / both, quoted labels split on first pipe), trailing edge option blocks, `properties { }`, list literals (inline + one-per-line), bareword/quoted/triple-quoted literals with scheme-URL barewords, `;` separators, empty blocks (D-01..D-09, D-12, D-15, D-18)
- Typed AST carrying Pos lines and attached comments — the foundation the D-32 formatter needs
- Front-end: `c4d.Parse`/`c4d.ParseFile` returning `(*ast.Document, error)`, every failure an `errors.As`-decodable `*parser.ParseError` with DSL-native line numbers (D-21), Memoize + MaxExpressions(1M) for bounded work on hostile input (T-35-01-01)

## Task Commits

Each task was committed atomically:

1. **Task 1: Pigeon toolchain + minimal generating grammar** - `cea815f` (feat)
2. **Task 2: Core grammar + typed AST** - `9e6b9a4` (test, RED) / `d3a4c80` (feat, GREEN)
3. **Task 3: Front-end Parse/ParseFile + ParseError wrapping** - `41da547` (test, RED) / `e732bfa` (feat, GREEN)

## TDD Gate Compliance

RED→GREEN discipline held for the behavior-adding tasks: `test(35-01)` commits (9e6b9a4, 41da547) precede their `feat(35-01)` GREEN commits (d3a4c80, e732bfa) in git order. Task 1 (toolchain) has no Go-test surface — its RED/GREEN was command-level: the plan's automated verify (build + vet + go.mod grep + parser_gen.go existence) failed before implementation and passes after.

## Files Created/Modified

- `internal/c4d/grammar/c4d.peg` - PEG grammar for core C4D statements (541 lines)
- `internal/c4d/grammar/parser_gen.go` - committed pigeon-generated parser (4396 lines, -nolint header)
- `internal/c4d/grammar/doc.go` - package doc + `//go:generate pigeon -o parser_gen.go -nolint c4d.peg`
- `internal/c4d/ast/ast.go` - typed AST: Document, PropertiesBlock, UnitNode, EdgeStmt, FieldStmt, Comment, UseStmt placeholder (147 lines)
- `internal/c4d/parse.go` - Parse/ParseFile front-end with Memoize + MaxExpressions options
- `internal/c4d/errors.go` - wrapPigeonError: pigeon errList strings -> *parser.ParseError (49 lines)
- `internal/c4d/doc.go` - package doc + go:generate forwarding to ./grammar
- `internal/c4d/parse_test.go` - 28 black-box test functions (56 cases) covering every Task 2/3 behavior bullet
- `tools.go` - //go:build tools blank import pinning pigeon (18 lines)
- `go.mod` / `go.sum` - pigeon v1.3.0 (+ its x/tools, x/mod indirect pins)

## Decisions Made

- pigeon v1.3.0 instead of the planned v1.0.0 (v1.0.0 has no `-nolint` flag; exact-pin + checksum intent preserved)
- Subpackage layout (`grammar`, `ast`) — forced by Go's lack of overloading against pigeon's generated `Parse`/`ParseFile`; keeps the plan's typed entry-point signatures intact
- Trailing-comment attachment follows gofmt semantics (preceding statement) so fmt preserves same-line comments
- `&StmtStart` implicit separator: `}` plus a following statement start separates statements, matching the plan's own `box "Platform" { api: system { } db: db { } }` example; field values still need `;` to separate (bareword greediness preserves D-18 semantics)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Pigeon version bumped v1.0.0 -> v1.3.0**
- **Found during:** Task 1 (toolchain setup)
- **Issue:** The plan's generate command (`pigeon -nolint`) fails on v1.0.0 — the `-nolint` flag does not exist there (verified via `pigeon -h` on both versions)
- **Fix:** Pinned `github.com/mna/pigeon v1.3.0` (latest tagged v1); threat-model intent (exact pin, go.sum-verified provenance) unchanged
- **Files modified:** go.mod, go.sum, tools.go
- **Verification:** `grep pigeon go.mod` shows exact version; build/vet/lint all green
- **Committed in:** cea815f

**2. [Rule 3 - Blocking] Generated parser + grammar moved to internal/c4d/grammar, AST to internal/c4d/ast**
- **Found during:** Task 1 (first generation attempt)
- **Issue:** Pigeon emits package-level `Parse(filename, []byte, ...Option)`/`ParseFile` in the generated file — a direct redeclaration collision with the plan-required `c4d.Parse(data []byte) (*ast.Document, error)` in the same package; grammar actions also cannot construct AST types living in the package that imports the grammar (import cycle)
- **Fix:** Three-package layering: `internal/c4d/grammar` (peg + generated parser), `internal/c4d/ast` (typed nodes), `internal/c4d` (front-end). Plan artifact paths shifted accordingly (c4d.peg/parser_gen.go under grammar/, ast.go under ast/); all min_lines and rule-name acceptance criteria still met
- **Files modified:** internal/c4d/grammar/\*, internal/c4d/ast/ast.go, internal/c4d/doc.go
- **Verification:** `go build ./...`, full test suite, golangci-lint clean
- **Committed in:** cea815f, d3a4c80

**3. [Rule 1 - Bug] wrapPigeonError extracts positions from pigeon's rendered error strings, not errors.As**
- **Found during:** Task 3 (error wrapping)
- **Issue:** Plan asked to mirror wrapDecodeError's `errors.As` structure, but pigeon's error types (`errList`, `parserError`) are unexported in the generated package — unreachable from package c4d
- **Fix:** Regex-parse the stable rendered prefix `line:col (offset)[: rule X]:` per error line; first positioned entry wins for `Line`; Cause keeps the original pigeon error so Unwrap reaches it. Message/Line/Context only (T-35-01-02)
- **Files modified:** internal/c4d/errors.go
- **Verification:** TestParseSyntaxErrorLineNumbers (line 3 of a 3-line doc), TestParseErrorContract (errors.As, Cause, Error() format)
- **Committed in:** e732bfa

**4. [Rule 2 - Missing critical] Comment attachment semantics pinned for D-32**
- **Found during:** Task 2 (GREEN)
- **Issue:** The initial test left same-line trailing comments nowhere in the tree — the formatter would silently drop them
- **Fix:** gofmt semantics — trailing comments attach to the preceding statement's Comments (after its leading ones); orphan comments attach to the enclosing node's TrailingComments. One committed RED test adjusted to assert the design
- **Files modified:** internal/c4d/parse_test.go, internal/c4d/grammar/c4d.peg, internal/c4d/ast/ast.go
- **Verification:** TestParseCommentLines, TestParseCommentAttachmentInsideBlock
- **Committed in:** d3a4c80

---

**Total deviations:** 4 auto-fixed (2 blocking, 1 bug, 1 missing critical)
**Impact on plan:** All forced by tooling realities (pigeon codegen shape, version flag gap) or correctness (trivia preservation). No scope creep — every plan acceptance criterion is met.

## Issues Encountered

- Zero-width `WS <- [ \t]*` inside `_ <- (WS / EOL)*` looped forever (pigeon star loops do not break on empty-success iterations) — fixed to `[ \t]+`; caught as a test hang, not a silent failure
- `]` was missing from list-bareword stop chars, so `[a, b]` items swallowed the closing bracket — caught by the inline list test
- Pigeon single quotes are single-char literals (`'"""'` invalid) and sequence/optional capture shapes are unspecified — actions use type-scanning helpers (toAnySlice) and escaped double-quoted string literals throughout

## Known Stubs

- `ast.UseStmt` and `UnitNode.UseStmts` are placeholders: the grammar reserves `use`/`include`/`template` as statement keywords (they fail to parse as unit ids) but their statements land in Plans 02/03 per the phase plan — intentional sequencing, not dead code
- `EdgeStmt` option blocks have no trailing-comment slot yet; orphan comments inside option blocks are dropped (fmt-grade trivia inside edge options can be refined in Plan 07)

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `c4d.ParseFile("x.c4d")` returns a typed AST for every core construct; Plan 02 (tomodel) can consume `*ast.Document` and the deferred reserved-keyword/error-suggestion work has a clean seam in `UnitId`/`ReservedKeyword`
- Regeneration contract: edit `internal/c4d/grammar/c4d.peg`, run `go generate ./internal/c4d`, commit the regenerated parser
- No blockers; repo suite 13/13 packages green, golangci-lint 0 issues

## Self-Check: PASSED

All 9 created files exist on disk; all 5 task commit hashes (cea815f, 9e6b9a4, d3a4c80, 41da547, e732bfa) verified in git log.

---
*Phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition*
*Completed: 2026-08-14*
