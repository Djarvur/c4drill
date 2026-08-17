---
phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition
plan: 07
subsystem: cli
tags: [cli, convert, c4d, dispatch, d24-validate-first, d25-follow-includes, d27-extension-dispatch, d28-subcommands, d29-direct-render, d30-output-placement, graph-conversion, mixed-format]
# Dependency graph
requires:
  - phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition (Plan 04)
    provides: EmitTOML/FromModel/EmitC4D deterministic emitters
  - phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition (Plan 05)
    provides: c4d.ParseFile Model-returning API, mixed-format include.Resolve
  - phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition (Plan 06)
    provides: the parity contract (round-trip/order-preservation guarantees convert relies on)
provides:
  - cmd/c4drill extension dispatch — c4drill diagram.c4d renders directly through the full pipeline (D-27/D-29); unknown extensions hard-error naming .toml/.c4d
  - c4drill convert to-toml/to-c4d — validate-first (D-24), swapped-extension output next to input or under -o (D-28/D-30), single-file structure preservation (D-25/D-22)
  - convert --follow-includes — whole-graph conversion with include-path rewriting, once-flag preservation, cycle safety, mixed-format graphs, -o preserving relative directory structure (D-25/D-26)
affects: [35-08 fmt (shares the parseInput extension rule), 35-09 CLI dispatch docs/skill surface]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Two-parse conversion rule: the D-24 gate (parse -> include.Resolve -> template.Expand -> peer.Resolve -> Validate) runs on a model that is DISCARDED after the gate — those stages mutate in place (Includes drained, templates consumed, peers absolutized) — and every twin emits from a FRESH source parse so structure survives verbatim"
    - "Graph twin conversion: deterministic DFS in include-directive order with the include-canonical-path stack/visited pattern; only include-path STRINGS are rewritten (relative form + once preserved); files already in the target format are skipped because conversion is additive (originals stay, so untouched directives keep resolving)"
    - "Cobra subcommand-per-direction with shared persistent flags: convert defines its own -o/--follow-includes as persistent flags; pflag's AddFlagSet duplicate-skip makes the child's -o shadow the root's inherited one deterministically"

key-files:
  created:
    - cmd/c4drill/convert.go
    - cmd/c4drill/convert_test.go
  modified:
    - cmd/c4drill/root.go
    - cmd/c4drill/root_test.go

key-decisions:
  - "Emission NEVER sees the pipeline: the validation copy is discarded and twins re-parse the source fresh — include directives verbatim, template declarations + use instantiations intact (no Expand), authored bare peers non-absolutized (no peer.Resolve), the D-22 round-trip parity with 35-06 upheld by tests"
  - "Graph-mode D-24 gate validates the ORIGINAL entry graph as authored (include.Resolve merges the whole graph; only the merged entry model validates) — included fragments may not be standalone-valid, mirroring the render pipeline; no twin is written if the gate fails"
  - "Files already in the target format get no twin (identity path rewrite): mixed .toml/.c4d graphs stay coherent because conversion is additive — originals are never deleted, so an untouched .c4d fragment's .toml include keeps resolving"
  - "Unknown input extension is a hard parse error naming both accepted extensions (static sentinel errUnsupportedExt); convert's direction gate mirrors it (to-toml on a .toml hard-errors naming the expected .c4d)"
  - "convert tests run WITHOUT t.Parallel despite doing no rendering: cobra flags bind package-level vars, so concurrent Execute calls would race on flag state (root.go precedent)"
  - "Short help now says 'from TOML and C4D' and Use is 'c4drill <input.toml|input.c4d>' — the plan's 'existing root tests untouched' could not survive its own mandated Use change, so the three Use-string assertions were updated (behavior tests untouched)"

requirements-completed: [D-24, D-25, D-27, D-28, D-29, D-30]

# Metrics
duration: 13min
completed: 2026-08-14
---

# Phase 35 Plan 07: CLI Wiring Summary

**Extension dispatch on the render path (.c4d renders directly, unknown extensions fail closed) plus the convert subcommand: validate-first two-parse conversion with swapped-extension output placement and whole-graph --follow-includes migration with include-path rewriting and cycle safety**

## Performance

- **Duration:** 13 min (18:50–19:03 UTC)
- **Started:** 2026-08-14T18:50:30Z
- **Completed:** 2026-08-14T19:03:05Z
- **Tasks:** 3/3
- **Files modified:** 4 (2 created, 2 modified)

## Accomplishments

- **Task 1 — root dispatch (D-27/D-29):** `parseInput(inputPath)` helper branches on `filepath.Ext` — `.toml` -> `parser.ParseFile`, `.c4d` -> `c4d.ParseFile`, default a hard error naming both accepted extensions (static sentinel `errUnsupportedExt`, `parse:` prefix kept on all branches). Everything downstream (include -> expand -> peer -> validate -> views -> render) untouched; TestXC01_PipelineOrdering green unchanged. `TestC4DRenderDirect` renders the hand-written `.c4d` twin of `skill/examples/01-minimal.toml` and its `.toml` original and asserts identical C1 DOT under canonical.Canonical (DI-1). Use/Short/Long updated to the dual-format shape.
- **Task 2 — convert subcommand (D-24/D-28/D-30):** `newConvertCmd()` with nested `to-toml`/`to-c4d` direction subcommands (cobra.ExactArgs(1)), registered via `cmd.AddCommand(newConvertCmd())` in NewRootCmd. TWO-PARSE rule: `validateSourceForConvert` runs the exact root stage composition (parse -> include.Resolve -> template.Expand -> peer.Resolve -> Validate + ReportErrors + `errValidationFailed`) on a model that is discarded; emission re-parses the source fresh and serializes via `c4d.EmitTOML` / `c4d.FromModel + c4d.EmitC4D`. Output = input basename + swapped extension next to the input; `-o` overrides the directory (created when missing). Round-trip tests assert the twin re-parses to the FRESH source Model (canonicalModel comparison with the D-22 defaults filled) plus explicit presence checks: includes verbatim (path + once), template params, template-body bare peers, unit-nested use, authored bare peers non-absolutized. Invalid input (`testdata/invalid_links.toml`, copied to a temp dir) errors with validation reported and NO output file; missing include errors with the `include:` prefix; wrong-direction extension hard-errors naming the expected one.
- **Task 3 — convert --follow-includes (D-25/D-26):** `--follow-includes` bool flag on both directions. `walkIncludeGraph` replicates include's canonical-path stack/visited traversal (DFS in directive order; cycle detection, diamond dedup, maxConvertDepth=100 — T-35-07-02/03 defense-in-depth on top of the gate's own include.Resolve cycle rejection). Per file: skip if already in the target format, else fresh-parse -> rewrite only the include-path strings (`retargetExt` preserves relative form; identity for already-target paths) -> emit -> write next to the source or under `-o` preserving the graph's relative directory structure. The 09-composed test asserts exactly the 3 graph twins (entry.c4d, templates.c4d, domains/auth.c4d; single-file-equivalent.toml gets none), rewritten paths with `once` preserved, and that the converted graph validates clean through the full pipeline (mixed-format include dispatch). Mixed graphs (TOML entry including a .c4d fragment and the mirror) convert whole-graph in both directions. Cycle test pins the "cycle" error and zero partial outputs.
- `go test ./...` 15/15 packages green, golangci-lint 0 issues (cmd + internal), gofmt clean.
- Manual spot check (binary level): `go run ./cmd/c4drill convert to-c4d <valid.toml>` -> render the `.c4d` directly -> `convert to-toml` back; all clean. Note: the repo-root `testdata/valid.toml` is by-design orphan-invalid, so the D-24 gate correctly REFUSES to convert it — the spot check used `cmd/c4drill/testdata/valid.toml`.

## Task Commits

Each task followed TDD RED -> GREEN:

1. **Task 1: root dispatch** — `9450234` (test, RED: .c4d hit the TOML parser, .json parsed as TOML, Use assertions failed) / `f2baeab` (feat, GREEN)
2. **Task 2: convert subcommand** — `015fff0` (test, RED: unknown command "convert") / `4349e85` (feat, GREEN)
3. **Task 3: --follow-includes** — `fbeb126` (test, RED: unknown flag --follow-includes) / `a453dc6` (feat, GREEN)

## TDD Gate Compliance

All three `test(35-07)` commits precede their implementation commits in git order (9450234 < f2baeab; 015fff0 < 4349e85; fbeb126 < a453dc6). All three REDs were runtime-level, not compile-level: Task 1's .c4d input failed inside the TOML parser and the .json error named no extensions; Tasks 2/3 failed on cobra's "unknown command"/"unknown flag" — genuine missing-feature evidence.

## Files Created/Modified

- `cmd/c4drill/convert.go` (404 lines; min_lines 80 met) — newConvertCmd/newDirectionCmd, runConvert, validateSourceForConvert (the D-24 gate), emitConverted, convertOutputPath, convertGraph, walkIncludeGraph, collectIncludeTargets, canonicalizeGraphPath, retargetExt, graphTwinPath
- `cmd/c4drill/convert_test.go` (703 lines) — 13 test functions: both-direction round-trips with preservation checks, -o redirect, invalid-input refusal, include-stage errors, wrong direction, help, graph conversion, graph -o, cycle, default-off, mixed graphs both directions; convertCanonicalModel/requirePreserved/runGraphPipeline helpers
- `cmd/c4drill/root.go` — extToml/extC4d consts, parseInput dispatch helper (extracted from Stage 1, also solving funlen), errUnsupportedExt sentinel, AddCommand(newConvertCmd()), dual-format Use/Short/Long
- `cmd/c4drill/root_test.go` — c4dMinimalTwin + runRenderDot helpers, TestC4DRenderDirect, TestRootInputUnknownExtension, three Use-string assertions updated

## Decisions Made

- The emission model is never validated, included, expanded or peer-resolved — preservation is structural, not best-effort (the plan's STAGE MUTATION WARNING made literal in code comments)
- Graph-mode gate validates the authored entry graph, not the rewritten twins; the plan's instruction to document this choice lives in the runConvert comment referencing D-24
- Twin filenames derive from the input basename ONLY (deriveBasename + ext swap) — no user-controlled filename portion (T-35-07-01); overwrite is the idempotent contract
- -o in graph mode preserves each source's directory relative to the entry's dir (domains/auth.toml -> domains/auth.c4d under -o); paths escaping the entry dir (../x) land at the -o root
- Convert output permission 0644 via plain os.WriteFile (G306-safe, one documented nolint)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Plan inaccuracy] Task 1's "existing root tests untouched and green" contradicted its own mandated Use change**
- **Found during:** Task 1 RED
- **Issue:** TestHelpText/TestHelpSubcommand assert `Contains "c4drill <input.toml>"` and TestNewRootCmd asserts `Equal` on the same string; the mandated Use `"c4drill <input.toml|input.c4d>"` breaks all three (the `>` no longer follows `.toml`)
- **Fix:** updated the three Use-string assertions to the new shape and extended the Short/Long assertions (Short now "…from TOML and C4D"); every BEHAVIOR test stayed untouched and green
- **Files modified:** cmd/c4drill/root_test.go, cmd/c4drill/root.go
- **Committed in:** 9450234, f2baeab

**2. [Rule 1 - Bug] Task 2 test fixtures linked a parent unit directly**
- **Found during:** Task 2 GREEN — the D-24 gate refused the fixtures with "unit api has subunits and cannot be linked to directly" / "cannot have direct links" (the validator's mirror populated an incoming link on the parent)
- **Issue:** the fixture's user->api edge targeted a parent; c4drill forbids direct links to/from parents
- **Fix:** user links the leaf `api.cache` in both the .c4d and .toml fixtures (absolute peer); the gate then passes and the preservation assertions are unaffected
- **Files modified:** cmd/c4drill/convert_test.go
- **Committed in:** 4349e85

**3. [Rule 3 - Plan inaccuracy] Task 3's mixed-graph test could not reuse the 35-05 fixtures**
- **Found during:** Task 3 test design
- **Issue:** internal/include/testdata/mixed_main.* contain units with no links (entrySvc, sharedDb) — orphan-invalid, so the D-24 gate would (correctly) refuse to convert them
- **Fix:** built gate-valid mixed graphs in t.TempDir (linked leaf units, containerDb under system) covering both directions; conversion of already-target-format files is identity by design
- **Files modified:** cmd/c4drill/convert_test.go
- **Committed in:** fbeb126, a453dc6

**4. [Rule 3 - Blocking] lint friction in new code**
- **Found during:** lint passes (all tasks)
- **Issue:** err113 (dynamic extension error), funlen (runRoot grew past 60 lines), wrapcheck (unwrapped parser/c4d returns), gosec G703 on fixture writes, lll over 120 chars, nolintlint (unused paralleltest directive on a helper), gocognit 18 on walkIncludeGraph, test-helper name collision with runConvert
- **Fix:** static sentinel errUnsupportedExt; parseInput extracted from runRoot (dispatch helper, also cleaner for convert reuse); emit errors wrapped with `emit:` prefix; targeted nolints with justifications (repo precedent); collectIncludeTargets extracted from walkIncludeGraph; test helper renamed execConvert
- **Files modified:** cmd/c4drill/root.go, cmd/c4drill/convert.go, cmd/c4drill/convert_test.go
- **Committed in:** f2baeab, 4349e85, a453dc6

---

**Total deviations:** 4 auto-fixed (2 plan inaccuracies, 1 fixture bug, 1 blocking lint friction)
**Impact on plan:** All acceptance criteria and must-have truths met; the gate-refusal incidents are the D-24 contract doing its job.

## Threat Model Disposition

| Threat | Disposition | Where |
|--------|-------------|-------|
| T-35-07-01 (output writes tampering) | mitigated | twin filename = deriveBasename + swapped extension ONLY (no user-controlled filename portion); -o accepts a user-named directory (standard CLI trust); overwrite is the documented idempotent contract |
| T-35-07-02 (traversal DoS) | mitigated | gate's include.Resolve cycle detection runs FIRST; walkIncludeGraph adds stack/visited + maxConvertDepth=100 (pinned by TestConvertGraphCycle: "cycle" error, no partial outputs) |
| T-35-07-03 (symlinked include targets) | mitigated | traversal keys on canonicalized paths (filepath.Clean + Abs, the include.canonicalize precedent); a symlink loop collapses into the cycle error |
| T-35-07-04 (error information disclosure) | accepted (as planned) | stage-prefixed errors echo author-supplied paths/messages only — same surface as the existing root command |

## Issues Encountered

- The plan's manual spot command (`convert to-c4d testdata/valid.toml`) targets the repo-root fixture, which is orphan-invalid by design — the D-24 gate refuses it (correct behavior, worth knowing for docs); the valid cmd corpus fixture converts/renders/round-trips cleanly
- Repo-root testdata files must never be converted in place (twins would land as untracked repo files) — all tests copy fixtures into t.TempDir first
- pflag's AddFlagSet duplicate-skip makes convert's own persistent -o deterministically shadow the root's inherited one; convertOutDir is a separate package var so convert never mutates the render command's outputDir

## Known Stubs

None. All paths are wired end to end: dispatch, both conversion directions, the validation gate, graph traversal, path rewriting, and output placement are exercised through the real cobra command with file-system assertions.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- 35-08 (fmt) can reuse `parseInput`'s extension rule and the cobra subcommand pattern; convert.go's stage-prefix conventions match root exactly
- 35-09 (CLI dispatch docs/skill) can document `.c4d` authoring end to end: direct render, both convert directions, --follow-includes graph migration
- The D-22/D-25 preservation contract is now enforced at CLI level, complementing 35-06's corpus-level parity suites
- No blockers; `go test ./...` 15/15 green, golangci-lint 0 issues

## Self-Check: PASSED

Both created files exist on disk (cmd/c4drill/convert.go, cmd/c4drill/convert_test.go); all 6 task commit hashes (9450234, f2baeab, 015fff0, 4349e85, fbeb126, a453dc6) verified in git log.

---
*Phase: 35-add-a-simple-dsl-alternative-to-the-toml-diagram-definition*
*Completed: 2026-08-14*
