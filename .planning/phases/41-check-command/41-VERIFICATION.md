---
phase: 41-check-command
status: passed
score: 4/4
verified: 2026-09-03
method: inline goal-backward verification (no verifier subagent available in runtime)
requirements_covered: [CHECK-01, CHECK-02, CHECK-03, CHECK-04]
human_verification_count: 0
gaps: []
---

# Phase 41: Check Command — Verification

**Goal:** Users can validate a model without rendering — a fast, render-free `check` command that fits CI and edit loops and reports exactly what the render path would report.

## Observable Truths (roadmap SCs + PLAN must_haves merged)

| # | Truth | Evidence | Status |
|---|-------|----------|--------|
| 1 | `c4drill check <file>` validates without writing any output files or requiring an output directory (SC-1, CHECK-01, D-04) | TestCheckWritesNothing: testdata directory snapshot byte-identical across check runs on valid/invalid/composed fixtures; UAT: `check gap.c4d.toml` in its own dir, no `-o`, directory listing unchanged; check.go constructs no output.Writer | VERIFIED |
| 2 | check exits 0 valid / exits 1 invalid, reporting the same validation errors render reports, incl. VAL orphan-unit rules — pinned TDD-first (SC-2, CHECK-02, D-03) | TestCheckValidModelSilentSuccess (nil error, empty out+err buffers); TestCheckOrphanUnitFailsWithValErrors (`errors.Is(err, errValidationFailed)`, `error: unit "orphan" has no incoming or outgoing links`, `1 error found`); TestCheckOrphanRenderParity: check stderr == render stderr byte-identical for the same fixture; UAT: check exit 1, render exit 1, identical output; TDD order proven: test(41-01) bc4a8e7 precedes feat(41-01) 7e0dc2a | VERIFIED |
| 3 | check runs the SAME pipeline front-half as render — includes, templates, relative peers, then validation — proven by a composed multi-file fixture that checks exactly as it renders (SC-3, CHECK-03, D-02) | ONE shared implementation: `parseValidatedModel` on root.go:249, called by runRoot (root.go:184) and runCheck (check.go:99); runRoot body has zero inline include.Resolve/template.Expand/peer.Resolve/validator.Validate calls; exactly one ReportErrors call in root.go, zero in check.go; TestCheckComposedFixtureValid: bare peer `authService` defined only in the included file resolves (include-before-peers ordering); TestCheckComposedRenderParity: render of the same composed model also succeeds with identical (empty) output | VERIFIED |
| 4 | README documents `check` alongside the existing CLI surface — usage, exit codes, no-output behavior (SC-4, CHECK-04, D-06) | README.adoc `=== check` (line 1474) between `=== fmt` (1452) and `=== serve` (1499), AsciiDoc style, usage string byte-matches the cobra Use line (`c4drill check <file.toml\|file.c4d>`), documents both formats, exit codes 0/1, no-output guarantee, and the `fmt --check . && c4drill check` CI gate; skill/SKILL.md retitled "Converting, Formatting, Validating (convert / fmt / check)" with 2 check usage lines | VERIFIED |

**Score: 4/4 verified** (plan must_haves truths 1-5 all covered by the rows above; D-05 additionally pinned: TestCheckHelpHidesRenderFlags asserts none of the 11 render flags appear in `check --help` — verified live: help shows only `-h, --help`)

## Artifacts

| Artifact | Provides | Contains-check | Status |
|----------|----------|----------------|--------|
| cmd/c4drill/check.go | newCheckCmd + runCheck (ExactArgs(1), SilenceUsage, no render flags) | `func newCheckCmd(`, `cobra.ExactArgs(1)`, `SilenceUsage: true`, `parseValidatedModel(cmd, args[0])` | PASS |
| cmd/c4drill/root.go | shared front-half helper both callers use | `func parseValidatedModel(` (249), `render.LabelRatio = getLabelRatio()` still inside runRoot (172) | PASS |
| cmd/c4drill/check_test.go | 10 behavior pins (valid-silent, orphan-errors, orphan render-parity, parse-prefix, composed-valid, composed render-parity, no-write, unknown-extension, no-args, help-hides-flags) | `func TestCheck` prefix, all passing | PASS |
| cmd/c4drill/testdata/check_orphan.toml | validation-stage failure fixture (ValidateOrphanUnits fires exactly once) | `error: unit "orphan"...` in test assertions | PASS |
| cmd/c4drill/testdata/check_composed_main.toml + check_composed_auth.toml | CHECK-03 cross-file parity fixture | `[[include]] path = "check_composed_auth.toml"`, bare peer `authService` | PASS |
| README.adoc | `=== check` command section | `=== check` (1474) | PASS |
| skill/SKILL.md | check in command examples | `c4drill check` (2 occurrences) | PASS |

## Key Links (wiring)

| From | To | Via | Pattern check | Status |
|------|----|----|---------------|--------|
| cmd/c4drill/check.go runCheck | cmd/c4drill/root.go parseValidatedModel | direct call | `parseValidatedModel\(` (check.go:99) | WIRED |
| cmd/c4drill/root.go runRoot | cmd/c4drill/root.go parseValidatedModel | single call replaces inline stages | `parseValidatedModel\(` (root.go:184) | WIRED |
| README.adoc `=== check` | cmd/c4drill/check.go | usage string equals cobra Use line | `c4drill check <file.toml\|file.c4d>` both sides | WIRED |

## Requirements Coverage

| Requirement | Definition | Status |
|-------------|-----------|--------|
| CHECK-01 | validate without output files or requiring an output directory | Complete (REQUIREMENTS.md, plan 41-01) |
| CHECK-02 | exit 0 valid / non-zero invalid with render-identical errors (orphan-unit VAL rules) | Complete (REQUIREMENTS.md, plan 41-01) |
| CHECK-03 | same pipeline front-half as render; composed sources check exactly as they render | Complete (REQUIREMENTS.md, plan 41-01) |
| CHECK-04 | README documents check alongside the existing CLI surface | Complete (REQUIREMENTS.md, plan 41-02) |

## Gate Results

- `go test ./cmd/c4drill/ -run 'TestCheck' -count=1 -v` — 10/10 PASS
- `go test ./cmd/c4drill/ -count=1` — PASS (pre-existing root/fmt/convert suites green after the extraction)
- `go test ./cmd/... ./internal/validator/... -count=1` — PASS (regression gate over this phase's surface + the directly adjacent validator package)
- `gofmt -l cmd/c4drill/` — clean; `go vet ./cmd/c4drill/` — clean; `go build ./...` — clean
- Code review gate: 41-REVIEW.md status **clean** (0 critical, 0 warning, 2 info)
- Issue #41 gap closed end-to-end (UAT): a format-clean but invalid model passes `c4drill fmt --check` (exit 0) and is caught by `c4drill check` (exit 1) with output byte-identical to render's
- Debt-marker scan (TODO/FIXME/XXX) over phase source files — zero
- Full `go test ./...` note: one failure in `internal/render` (TestDeterministicByteIdenticalOutput) — concurrent phase-40 work in flight, outside this phase's file surface; cmd/ and internal/validator fully green

## Gaps

None.

## Conclusion

Phase 41 achieved its goal: `c4drill check <file.toml|file.c4d>` is a fast render-free validator that shares ONE pipeline front-half with render (single-sourced `parseValidatedModel`), reports byte-identical validation errors, exits 0/1 correctly, writes nothing, needs no output directory, and is documented across README.adoc and skill/SKILL.md.
