---
phase: 41-check-command
status: clean
depth: standard
reviewed_by: inline orchestrator review (gsd-code-reviewer agent surface unavailable in this runtime)
date: 2026-09-03
files_reviewed: 7
critical: 0
warning: 0
info: 2
total: 2
---

# Phase 41 Code Review Report

**Scope (from SUMMARY key-files):** cmd/c4drill/check.go, cmd/c4drill/check_test.go, cmd/c4drill/root.go, cmd/c4drill/testdata/check_orphan.toml, cmd/c4drill/testdata/check_composed_main.toml, cmd/c4drill/testdata/check_composed_auth.toml, README.adoc, skill/SKILL.md

## Mechanical Checks

| Check | Result |
|-------|--------|
| go vet ./cmd/c4drill/ | clean |
| gofmt -l cmd/c4drill/ | clean |
| go test ./cmd/c4drill/ -count=1 | PASS |
| Docs usage string vs cobra Use line | exact match (`check <file.toml|file.c4d>`) |
| Error path audit | runCheck returns the shared front-half error verbatim; main() exits 1 on any error, 0 implicit — no swallowed errors, no double reporting |

## Review Analysis (standard depth)

**Bugs:** None found. `parseValidatedModel` returns `(nil, err)` on every failure and a non-nil validated model on success (parseInput cannot produce a nil model with nil error); `runRoot` and `runCheck` both gate on err before use. Extraction from runRoot is a verbatim move — stage order, `fmt.Errorf` prefixes (`parse:`/`include:`/`expand:`/`resolve peers:`), the `cmd.OutOrStderr()` writer, and `errValidationFailed` are preserved (confirmed by diff review and the render-parity tests).

**Security:** No new surface. check is read-parse-report with zero writes (pinned by TestCheckWritesNothing directory-snapshot test); validator messages name unit paths only — identical exposure to render (D-03). Extension dispatch fails closed via the shared parseInput. No installs, no subprocess, no network. Threat register dispositions T-41-01/T-41-02 (accept) hold.

**Correctness of the D-05 help mechanism:** `helpCheck` hides the 11 inherited render flags only for the duration of check's own help render, restoring via defer — required because cobra shares `*flag.Flag` objects between root and subcommands (permanent hiding regressed root --help and was caught by TestHelpText during GREEN). Safe: command instances are per-invocation and help rendering is single-threaded and synchronous.

**Tests:** 10 behavior pins cover every CHECK contract, including byte-identical render parity on both an invalid and a valid composed fixture. RED→GREEN commit order verified (bc4a8e7 before 7e0dc2a).

## Findings

### IR-1 (Info) — renderOnlyFlags duplicates root's flag-name list
cmd/c4drill/check.go — the hidden-flag names are a literal list; a future render flag added to root would need a matching entry here or it would appear in `check --help`. TestCheckHelpHidesRenderFlags pins today's names but cannot catch a future addition. Suggested (optional, future): derive the list from root's persistent flags minus help/version. No action this phase — the flag surface is stable and documented.

### IR-2 (Info) — help-time flag mutation relies on single-threaded help rendering
cmd/c4drill/check.go — safe under every current caller (CLI single render, tests construct fresh commands); would misbehave only if the same command instance rendered help concurrently from multiple goroutines. Documented in code comments. No action.

## Verdict

**Clean** — no critical or warning findings; both info findings are documented, non-blocking observations with no current failure mode.
