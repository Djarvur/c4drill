---
phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition
plan: 08
subsystem: cli
tags: [fmt, tomlfmt, formatter, d31, d32, d33, idempotency, comment-preservation, check-mode, ci-gate, safety-gate, go-toml-unstable, corpus-sweep]

# Dependency graph
requires:
  - phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition (Plan 04)
    provides: EmitC4D compact-leaf canonical printer
  - phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition (Plan 05)
    provides: exported c4d.ParseAST (the trivia-aware AST entry fmt consumes)
  - phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition (Plan 07)
    provides: cobra subcommand conventions, parseInput extension rule, root AddCommand wiring
provides:
  - internal/tomlfmt — comment-preserving, idempotent TOML formatter over the go-toml unstable API with KeepComments (D-31/D-32)
  - c4drill fmt — gofmt-style in-place formatter for BOTH authoring formats with recursive dir walk, --check CI mode (exit 1 listing offenders), and the T-35-08-01 semantic safety gate
  - fmt∘fmt == fmt proven over the full shipped corpus (both formats, incl. converted twins)
affects: [35-09 CLI dispatch docs/skill surface (fmt is documentable end to end)]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Event-stream TOML formatting: statements render as offset-anchored events; blank-line grouping is decided by counting newlines in the SOURCE gap between consecutive event spans (>= 2 renders one blank line), so the author's grouping survives and runs collapse deterministically to a fixpoint"
    - "go-toml unstable KeepComments semantics: same-line trailing comments ride expr.Next() (finishLine attaches them as the expression root's next sibling); standalone comments are their own expressions; values render from raw source bytes (composite value nodes carry partial raws — only the KeyValue root raw spans the full statement)"
    - "Gate-then-rewrite: the candidate output must re-parse (DeepEqual) to the original file's Model BEFORE any byte is written; the gate is a separate testable seam (applyFormatted) so the corruption path is provable without an emitter bug"

key-files:
  created:
    - internal/tomlfmt/tomlfmt.go
    - internal/tomlfmt/tomlfmt_test.go
    - cmd/c4drill/fmt.go
    - cmd/c4drill/fmt_test.go
  modified:
    - cmd/c4drill/root.go

key-decisions:
  - "fmt on .toml preserves the AUTHOR's key order (documented package-doc contrast with convert's D-23 canonical order) — formatting normalizes only whitespace/indent/style: column-0 statements, one space around '=', tight table brackets, blank-line runs collapse to one"
  - "TOML values render from raw source bytes, verbatim: quoting style, multi-line strings/arrays and comments inside them pass through untouched (idempotent by construction — the raw slice of the formatted output is the same bytes)"
  - "The C4D path reuses the exported ParseAST + EmitC4D with a separate Model-parse baseline (convert's two-parse rule): the gate compares c4d.Parse(original) vs c4d.Parse(candidate) DeepEqual — same-format comparison, so no cross-format defaults handling is needed"
  - "fmt hard-errors on a Model-refusing file (e.g. duplicate unit path .c4d) BEFORE formatting — the safety gate runs on the original parse; syntax-invalid corpus fixture (cmd/c4drill/testdata/invalid.toml) pinned as the fail-closed refusal case"
  - "fmt on a directory with zero matching files errors loudly (errFmtNoTargets) — the v1.10 hard-error stance; direct file args with unknown extensions reuse the root errUnsupportedExt sentinel naming .toml/.c4d (T-35-08-02)"
  - "Task 3's 'both formats across all four corpus roots' adapted to reality (zero shipped .c4d files there): the sweep adds internal/include/testdata (the two REAL .c4d fixtures, 35-05) as a fifth root AND converts every valid TOML fixture to a .c4d twin via FromModel+EmitC4D — the C4D leg covers the whole corpus semantics"

requirements-completed: [D-31, D-32, D-33]

# Metrics
duration: 24min
completed: 2026-08-14
---

# Phase 35 Plan 08: Formatter Summary

**internal/tomlfmt (comment-preserving, idempotent TOML formatting over the go-toml unstable KeepComments API) + the c4drill fmt subcommand (both formats in place, recursive walk, --check CI mode, semantic safety gate) with fmt∘fmt == fmt proven over the entire shipped corpus**

## Performance

- **Duration:** 24 min (19:07–19:32 UTC)
- **Started:** 2026-08-14T19:07:56Z
- **Completed:** 2026-08-14T19:31:59Z
- **Tasks:** 3/3
- **Files modified:** 5 (4 created, 1 modified)

## Accomplishments

- **Task 1 — internal/tomlfmt (D-31/D-32):** `Format(data []byte) ([]byte, error)` walks the go-toml v2 unstable API with `KeepComments: true` (the same API internal/parser uses for definition-order capture — the plan's probe confirmed comments ARE recoverable with positions, so the tree-walk strategy shipped, not the fallback). Statements render as offset-anchored events: table headers from key-segment raws, `key = value` with exactly one space around '=', values VERBATIM from raw source bytes (composite value nodes carry partial raws; only the KeyValue root raw spans the full statement). Same-line trailing comments ride `expr.Next()` (finishLine attaches them as the expression root's next sibling); blank-line grouping is decided by newline counts in the source gap between event spans (>= 2 renders one blank line; longer runs collapse — that IS the normalization, and the fixpoint). Malformed TOML fails closed with zero output bytes. Tests: comment text-list equality (header runs, above tables, same-line tails, standalone, inside multi-line values), a whitespace/order golden (author key order kept — NOT convert's D-23 canonical), corpus idempotency over 39 fixtures + pinned syntax-invalid refusal, parser.Parse model equality, already-formatted stability.
- **Task 2 — fmt subcommand (D-31):** `newFMTCmd()` with `--check`, registered in NewRootCmd beside convert. Arg expansion: files used directly (extension must be .toml/.c4d — root's errUnsupportedExt sentinel, T-35-08-02), directories via filepath.WalkDir (never follows symlinked dirs) collecting both extensions. Per file: .c4d -> ParseAST + EmitC4D (comments ride the AST, D-32/D-33; NOT a Model re-emit) with a separate c4d.Parse baseline; .toml -> tomlfmt.Format with a parser.Parse baseline. SAFETY GATE (T-35-08-01): the candidate must re-parse to a Model DeepEqual to the original's BEFORE any write — reparse failure and Model drift are both hard errors leaving the file untouched (the gate is the testable `applyFormatted` seam). --check prints offenders one per line and exits 1 with zero byte writes (T-35-08-04); a Model-broken .c4d (duplicate unit path) hard-errors at command level; empty target sets error loudly.
- **Task 3 — corpus sweep (D-32 enforcer):** TestFmtCorpusIdempotency over 53 fixtures (52 TOML + 2 real .c4d from internal/include/testdata, minus the pinned syntax-invalid one) across five roots: every fixture asserts second-format no-op AND Model equality, both formats; the C4D leg additionally runs converted .c4d twins (FromModel+EmitC4D of every valid TOML fixture), covering the whole corpus's semantics in C4D form. TestFmtCorpusCheckClean copies the corpus (with twins) to t.TempDir, runs the fmt COMMAND in place, then `--check` exits 0 — the CI gate proven over the real corpus. The sweep passed on arrival: Tasks 1-2 left no fixture gaps (honest green for a test-only contract-enforcer task — nothing existed to make RED; documented under TDD Gate Compliance).
- `go test ./...` 15/15 packages green, golangci-lint 0 issues (cmd + internal), gofmt clean.
- Manual spot check (binary level): scratch corpus copy -> `fmt` exit 0 -> `--check` exit 0 -> re-`fmt` exit 0; 04-styling.toml comment count identical (31 = 31) with only the canonical trailing-comment spacing normalized (`  #` -> ` #`); testdata/valid.toml byte-identical (already formatted).

## Task Commits

Tasks 1-2 followed TDD RED -> GREEN; Task 3 is test-only (contract enforcer):

1. **Task 1: internal/tomlfmt** — `f62dcbd` (test, RED: package undefined, compile-level — 35-06 precedent) / `46aecf9` (feat, GREEN)
2. **Task 2: fmt subcommand** — `cb08e8c` (test, RED: applyFormatted/newFMTCmd undefined, compile-level) / `37494e7` (feat, GREEN)
3. **Task 3: corpus sweep** — `87da93d` (test; green-on-arrival, see TDD Gate Compliance)

## TDD Gate Compliance

Tasks 1-2: both `test(35-08)` commits precede their implementation commits in git order (f62dcbd < 46aecf9; cb08e8c < 37494e7). Task 1 RED was compile-level (package undefined); Task 2 RED was compile-level (applyFormatted seam undefined — the command tests would fail at "unknown command fmt", but the seam reference blocks compilation first).

Task 3 is a test-only contract-enforcer task (its `<files>` list contains only fmt_test.go): the sweep verifies Tasks 1-2's output over the shipped corpus and passed on first run. Per the fail-fast rule this was investigated, not waved through: the green is genuine coverage, not a vacuous test — the sweep exercises previously-untested surface (six internal/include TOML fragments, the two real .c4d fixtures, 50+ converted twins, command-level --check over a corpus copy) and its guards (tomlCount > 10, c4dCount > 0) pin that the walk actually collects fixtures. No RED was possible without manufacturing a failure; the plan's own action anticipates this ("if any fixture fails idempotency, fix the emitter/formatter" — none did).

## Files Created/Modified

- `internal/tomlfmt/tomlfmt.go` (248 lines; min_lines 60 met) — Format, collectEvents, commentEvent, tableEvent, keyValueEvent, rawValue, render, blankBetween; static sentinels errMalformedTOML/errUnsupportedExpr
- `internal/tomlfmt/tomlfmt_test.go` (374 lines) — 7 tests incl. corpus walk (40 fixtures), comment extractor over the unstable API, pinned syntax-invalid refusal set
- `cmd/c4drill/fmt.go` (269 lines; min_lines 60 met) — newFMTCmd, runFMT, expandFMTArgs/expandFMTArg/walkFMTDir, formatFile, applyFormatted (the gate seam)
- `cmd/c4drill/fmt_test.go` (585 lines) — 13 tests: in-place both formats with comment assertions, dir walk, --check both exits, multiple args, unknown extension, Model-broken refusal, gate seam (broken reparse + Model drift), no-targets, help, corpus sweep + corpus --check
- `cmd/c4drill/root.go` — cmd.AddCommand(newFMTCmd()) beside convert

## Decisions Made

- See key-decisions frontmatter; the load-bearing ones: author-order preservation (vs D-23), raw-value verbatim rendering, gate-then-rewrite with a DeepEqual same-format comparison, the fifth corpus root + converted twins for .c4d coverage

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] rawValue guard broke composite TOML values**
- **Found during:** Task 1 GREEN — inline-table/array fixtures errored "key-value without a value"
- **Issue:** the initial guard compared the value-start offset against the VALUE NODE's raw end, but go-toml gives composite value nodes partial raws (an InlineTable node's raw covers only `{`); only the KeyValue ROOT raw spans the full statement
- **Fix:** dropped the valueEnd parameter — the parser guarantees a value exists and the expression raw end bounds the slice; empty-slice guard retained for fail-closed safety
- **Files modified:** internal/tomlfmt/tomlfmt.go
- **Committed in:** 46aecf9

**2. [Rule 3 - Plan inaccuracy] the corpus contains a syntax-invalid TOML fixture**
- **Found during:** Task 1 corpus run — cmd/c4drill/testdata/invalid.toml fails at TOML SYNTAX level ("expected ']' to close table name"), so "Format succeeds over the corpus" could not hold unconditionally
- **Fix:** pinned it as syntaxInvalidCorpus (the 35-06 invalidCorpusFixtures pattern) and asserted the FAIL-CLOSED half of the contract instead: Format returns an error and parser.Parse agrees — a refusal documented as behavior, not worked around
- **Files modified:** internal/tomlfmt/tomlfmt_test.go, cmd/c4drill/fmt_test.go (fmtSyntaxInvalid)
- **Committed in:** 46aecf9, 87da93d

**3. [Rule 3 - Plan inaccuracy] zero .c4d files exist in the plan's four corpus roots**
- **Found during:** Task 3 design — "sweep covers both formats across all four corpus roots" is unsatisfiable as written (the corpus is TOML-only; the repo's only real .c4d files live in internal/include/testdata)
- **Fix:** added internal/include/testdata as a fifth root (2 real .c4d fixtures) AND swept converted .c4d twins of every valid TOML fixture (FromModel+EmitC4D) — the C4D leg covers the full corpus semantics, which is stronger than the literal four-root walk
- **Files modified:** cmd/c4drill/fmt_test.go
- **Committed in:** 87da93d

**4. [Rule 3 - Blocking] lint friction in new code**
- **Found during:** lint passes (all tasks)
- **Issue:** exhaustive on the expression-kind switch, gosec G304 (corpus reads) / G306 (fmt write) / G703 (tempdir writes), errcheck on Fprintln, gocognit 21 on expandFMTArgs, revive error-return on the exec helper, paralleltest (cobra flag vars), testifylint negative-positive, lll, wsl_v5
- **Fix:** targeted nolints with justifications (repo precedent), expandFMTArg/walkFMTDir extraction, error-last return order, assert.Positive, helper restructuring
- **Files modified:** internal/tomlfmt/tomlfmt.go, cmd/c4drill/fmt.go, cmd/c4drill/fmt_test.go
- **Committed in:** 46aecf9, 37494e7, 87da93d

---

**Total deviations:** 4 auto-fixed (1 bug, 2 plan inaccuracies, 1 blocking lint friction)
**Impact on plan:** All acceptance criteria and must-have truths met; deviations 2-3 turned plan inaccuracies into stronger documented contracts.

## Threat Model Disposition

| Threat | Disposition | Where |
|--------|-------------|-------|
| T-35-08-01 (in-place rewrite tampering) | mitigated | applyFormatted gate: candidate must re-parse DeepEqual to the original Model BEFORE any write; both failure legs (reparse error, Model drift) unit-tested via the seam; Model-broken originals hard-error in formatFile |
| T-35-08-02 (recursive walk tampering) | mitigated | filepath.WalkDir (no symlinked-dir following), *.c4d/*.toml filter only; direct file args reuse the root errUnsupportedExt sentinel naming both accepted extensions (tested) |
| T-35-08-03 (pathological files DoS) | accepted (as planned) | local dev tool over own files; parse costs bounded by MaxExpressions (C4D) and go-toml (TOML) |
| T-35-08-04 (--check repudiation) | mitigated | offenders print one per line to stdout + exit 1; TestFMTCheckMisformattedExitsOne pins path output AND zero byte change |

## Issues Encountered

- The unstable API's arena resets on every NextExpression (Expression() returns &p.nodes[0]) — every string/offset is copied out before advancing; no node pointers escape collectEvents
- The go-toml ExampleParser_comments output was the decisive documentation: comments render as standalone expressions AND as the expression root's next sibling (finishLine), which fixed the event design before any code was written
- Task 1's first golden had an authoring slip (omitted the authored blank line after the header comment) — the implementation was right (blank-line grouping preserved); corrected the golden during RED->GREEN iteration
- 04-styling.toml's `  #` double-space trailing comments normalize to single-space ` #` — canonical style change on first format (comment TEXT is verbatim; count identical 31=31)

## Known Stubs

None. Both format paths, the walk, the gate and --check are wired end to end and exercised through the real cobra command with file-system assertions.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- 35-09 (CLI dispatch docs/skill) can document fmt end to end: dual-format in-place formatting, --check CI gate, author-order/comment-preservation semantics
- The fmt surface completes the tooling triangle (render, convert, fmt) over both authoring formats; the corpus sweep keeps the contract enforced as fixtures grow
- No blockers; `go test ./...` 15/15 green, golangci-lint 0 issues

## Self-Check: PASSED

All 4 created files exist on disk; all 5 task commit hashes (f62dcbd, 46aecf9, cb08e8c, 37494e7, 87da93d) verified in git log.

---
*Phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition*
*Completed: 2026-08-14*
