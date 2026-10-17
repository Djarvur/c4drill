---
status: complete
phase: quick-261016-mxp
plan: 01
subsystem: c4d parser (pigeon budget) + CLI flags + parse error wrapping
tags: [bug-fix, feature, parser, c4d, cli-flag, error-messages, lsp]
requires: []
provides:
  - "10x default pigeon expression budget (1M → 10M) — ~110KB single-file models parse again"
  - "--max-expressions persistent CLI flag (render/check/convert/fmt/serve/includes); 0 disables the cap"
  - "budget-exhausted parse errors name the limit and the flag instead of posing as syntax errors"
  - "named-parse (ParseFile) error wrapping extracts DSL-native lines — check/render C4D errors match the TOML front-end shape"
affects: [internal/c4d, cmd/c4drill, internal/lsp]
tech-stack:
  added: []
  patterns:
    - "CLI-tunable package knob bound directly from a cobra persistent flag (render.LabelRatio precedent)"
    - "known-prefix stripping before the positioned-error regex in wrapPigeonError"
key-files:
  created:
    - .planning/quick/261016-mxp-raise-c4d-parse-budget-and-add-flag/261016-mxp-PLAN.md
    - .planning/quick/261016-mxp-raise-c4d-parse-budget-and-add-flag/261016-mxp-SUMMARY.md
  modified:
    - internal/c4d/parse.go
    - internal/c4d/parse_test.go
    - internal/c4d/errors.go
    - cmd/c4drill/root.go
    - cmd/c4drill/root_test.go
    - internal/lsp/server_test.go
    - README.adoc
decisions:
  - "Knob = exported var c4d.MaxExpressions (not a parameter through the pipeline): one binding covers every parse path including [[include]] recursion and the LSP, matching the render.LabelRatio precedent; written once at flag-parse time, read-only during parses (race-safe)"
  - "0 means uncapped — pigeon's native MaxExpressions(0) semantics, documented in the flag help instead of rejected"
  - "Named-parse fix strips the exact `context + \":\"` prefix rather than generalizing the regex — no path/line mis-splitting on digit-suffixed or Windows-style paths"
  - "LSP C4D parity test re-pinned to the TOML-shaped rendering (parse error at line N, no path echo) — the old pin preserved the raw-pigeon quirk"
metrics:
  duration: ~1h
  completed: 2026-10-16
  tasks: 3
  files: 7
---

# Quick Task 261016-mxp: Raise C4D Parse Budget 10x and Add --max-expressions — Summary

All three reporter asks delivered in commit 605a892: the default pigeon
expression budget is 10x the old 1M bound, `--max-expressions` makes it
tunable everywhere, and the budget error now explains itself. A fourth,
unasked-for fix fell out: named parses (check/render) had silently lost the
DSL-native line extraction because pigeon prefixes error lines with the file
path and `wrapPigeonError`'s regex expected the bare positioned form.

## What Changed

- **internal/c4d/parse.go** — `maxExpressions` const (1M) → exported
  `MaxExpressions` var, default 10_000_000; the knob doc records why (the old
  bound was a file-size limit in disguise for ~110KB single-file models).
- **cmd/c4drill/root.go** — persistent `--max-expressions uint` bound directly
  to the knob; inherited by check, convert, fmt, serve; help documents `0` =
  uncapped.
- **internal/c4d/errors.go** — (a) the budget-exhausted message becomes
  `max number of expressions parsed: the expression budget (N) is a resource
  cap against runaway parses, not a syntax limit — raise it with
  --max-expressions (0 disables the cap)`; (b) `wrapPigeonError` strips the
  known `name:` prefix before the positioned-error regex, so Line extracts on
  the named path too (C4D `check` errors now render `parse error at line N: …`
  like the TOML front-end instead of the raw pigeon line).
- **README.adoc** — CLI Reference block re-aligned (real cobra output), new
  `--max-expressions` bullet; LSP section unaffected.

## Tests

- `TestParseMaxExpressionsKnob` (internal/c4d): budget 1 fails a valid doc
  naming the limit and the flag, on both the AST (unnamed) and named
  (ParseNamed) entries with the DSL line intact; 0 parses; default parses.
- `TestMaxExpressionsFlag_Exists` / `TestMaxExpressionsFlag_LimitsC4DParsing`
  (cmd/c4drill): flag registered with default 10000000; `--max-expressions 1`
  fails E2E naming the flag, `0` renders minimal.dot.
- `TestC4DParseErrorParity` (internal/lsp): re-pinned — C4D diagnostics now
  carry the TOML-shaped `parse: parse error at line 1:` prefix, no path echo,
  Range anchored at 0-based line 0.
- Reproducer (synthetic 222KB / 1500-unit .c4d): old 1M budget →
  `max number of expressions parsed` at the furthest position; new default →
  parse+validate exit 0 in ~1.8s; `--max-expressions 0` → exit 0.

## Verification

- `go test -race ./...` — all packages ok.
- `golangci-lint run` — 0 issues.
- Out of scope (reporter's alternative option 3): replacing the budget with a
  time/depth limit — not attempted; the raised, tunable cap keeps the
  hang-guard semantics the budget was added for (T-35-01-01).

## Known Stubs

None.

## Self-Check: PASSED

- All must_haves truths hold (see PLAN.md); verifier evidence: the knob test,
  the CLI flag tests, the 222KB reproducer transcript in this summary.
- Full suite green at commit 605a892; release: product tag v1.27.0.
