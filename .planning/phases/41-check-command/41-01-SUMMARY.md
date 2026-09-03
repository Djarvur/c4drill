---
phase: 41-check-command
plan: 01
subsystem: cli
tags: [cobra, validation, tdd, pipeline, issue-41]

requires:
  - phase: 32-include-resolution
    provides: include.Resolve stage of the pipeline front-half
  - phase: 31-template-expansion
    provides: template.Expand stage
  - phase: 30-relative-peers
    provides: peer.Resolve stage
provides:
  - `c4drill check <file.toml|file.c4d>` subcommand — render-free validation, silent exit 0 / render-identical exit 1
  - shared pipeline front-half helper `parseValidatedModel` on root.go (single-sourced stages 1→2 for check AND render)
  - behavior pins (10 tests) + 3 TOML fixtures (orphan validation failure, two-file composed model)
affects: [phase-42 (gui could surface check), docs (README/SKILL in 41-02), future CI gates]

tech-stack:
  added: []
  patterns:
    - "Shared front-half extraction: one helper owns stage order + error prefixes + validation writer for every CLI verb that validates"
    - "Help-surface scoping: check-scoped help func hides inherited render flags, restoring the shared flag objects after render (cobra shares *flag.Flag across commands)"

key-files:
  created:
    - cmd/c4drill/check.go
    - cmd/c4drill/check_test.go
    - cmd/c4drill/testdata/check_orphan.toml
    - cmd/c4drill/testdata/check_composed_main.toml
    - cmd/c4drill/testdata/check_composed_auth.toml
  modified:
    - cmd/c4drill/root.go

key-decisions:
  - "parseValidatedModel(cmd, inputPath) extracted verbatim from runRoot stages 1→2; runRoot calls it, check calls it — no copied stage sequence (D-02)"
  - "Orphan-unit fixture carries [[user.link]] because linkFrom is NOT normalized onto the peer unit — without it user would also fire ValidateOrphanUnits (2 errors, not 1)"
  - "Render flags hidden from check --help via help-time hide/restore instead of permanent Hidden=true: cobra flag objects are shared with root, permanent hiding regressed root --help (caught by TestHelpText)"

requirements-completed: [CHECK-01, CHECK-02, CHECK-03]

duration: ~25 min
completed: 2026-09-03
---

# Phase 41 Plan 01: Check Command Summary

`c4drill check <file>` — render-free model validation sharing ONE extracted pipeline front-half with render (`parseValidatedModel` on root.go), TDD-first: silent exit 0 when valid, exit 1 with byte-identical render validation errors when not; nothing written, no output directory required (issue #41, CHECK-01/02/03).

## What Was Built

- **cmd/c4drill/root.go** — extracted runRoot's stages 1→2 (parseInput → include.Resolve → template.Expand → peer.Resolve → validator.Validate + ReportErrors + errValidationFailed) verbatim into `parseValidatedModel(cmd *cobra.Command, inputPath string) (*parser.Model, error)`; runRoot now makes one helper call; render-only concerns (validateOutputFlags, `render.LabelRatio = getLabelRatio()`) stay in runRoot. Registered the check subcommand with the `(issue #41)` comment convention.
- **cmd/c4drill/check.go** — `newCheckCmd` (Use `check <file.toml|file.c4d>`, ExactArgs(1), SilenceUsage, Long documenting same-front-half/exit codes/no-output) + `runCheck` (returns the shared front-half error verbatim, discards the model, silent on success) + D-05 help scoping (`helpCheck`/`setRenderFlagsHidden`/`renderOnlyFlags`).
- **cmd/c4drill/check_test.go** — 10 behavior pins: valid-silent, orphan VAL errors + errValidationFailed sentinel, orphan render-parity (byte-identical stderr), parse-stage prefix, composed multi-file valid, composed render-parity, no-write directory snapshot, unknown-extension fail-closed, ExactArgs(1), help hides 11 render flags.
- **3 fixtures** — check_orphan.toml (validation-stage failure: exactly one orphan), check_composed_main.toml + check_composed_auth.toml (cross-file bare peer resolvable only because include.Resolve precedes peer.Resolve).

## Tasks

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 | RED — fixtures + failing check tests | bc4a8e7 | check_test.go + 3 fixtures |
| 2 | GREEN — shared front-half + check.go | 7e0dc2a | root.go, check.go |
| 3 | REFACTOR — gofmt nolint placement | 11e1900 | check.go |

## Verification

- `go test ./cmd/c4drill/ -run 'TestCheck' -count=1 -v` — 10/10 PASS
- `go test ./cmd/c4drill/ -count=1` — PASS (all pre-existing root/fmt/convert tests green; extraction regression-proof)
- `gofmt -l cmd/c4drill/` clean; `go vet ./cmd/c4drill/` clean
- TDD gate: test(41-01) commit bc4a8e7 precedes feat(41-01) commit 7e0dc2a
- CLI smoke: `check testdata/valid.toml` → exit 0 silent; `check checkdata/check_orphan.toml` → exit 1 with `error: unit "orphan" has no incoming or outgoing links in orphan` + `1 error found`; `check --help` shows no render flags
- `go test ./...` — one failure in internal/render (TestDeterministicByteIdenticalOutput): concurrent phase-40 work outside this phase's file surface (concurrent-phase noise, not a regression from this plan); everything under cmd/ passes

## Deviations from Plan

- **[Rule 1 - Mechanical] Flag-hiding moved from newCheckCmd's constructor body to a check-scoped help func (hide at help time, restore after).** Found during: Task 2 GREEN | Issue: cobra shares one *flag.Flag object between root PersistentFlags and every subcommand's inherited set, so permanently setting Hidden=true inside newCheckCmd (pre-registration: no parents visible yet) or post-registration stripped render flags from root --help and failed the pre-existing TestHelpText. | Fix: `cmd.SetHelpFunc(helpCheck)` — check's help renders with the 11 render flags hidden, restoring them immediately after; root/sibling help untouched, D-05 behavior as pinned by the tests. | Files: check.go | Verification: TestCheckHelpHidesRenderFlags + TestHelpText both green | Commit 7e0dc2a.
- **Total deviations:** 1 auto-fixed (mechanical, cobra flag-ownership constraint). **Impact:** none on behavior — D-05 pinned by tests, root help regression caught and fixed within the same task.

## Self-Check: PASSED

- key-files.created exist on disk: yes (5/5)
- git log --grep 41-01: 3 commits (test/feat/refactor)
- All task acceptance criteria re-run: PASS
- Plan <verification> commands re-run: PASS (full-suite caveat documented above)
