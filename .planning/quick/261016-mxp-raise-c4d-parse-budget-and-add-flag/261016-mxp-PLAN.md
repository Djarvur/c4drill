---
phase: quick-261016-mxp
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/c4d/parse.go
  - internal/c4d/parse_test.go
  - internal/c4d/errors.go
  - cmd/c4drill/root.go
  - cmd/c4drill/root_test.go
  - internal/lsp/server_test.go
  - README.adoc
autonomous: true
requirements: [MXP-BUDGET-10X, MXP-BUDGET-FLAG, MXP-BUDGET-ERROR-MSG]
must_haves:
  truths:
    - "A valid single-file .c4d model around 110KB parses again: the default pigeon expression budget is 10x the old 1M bound (proven on a 222KB synthetic — fails at the old 1M, parses at the new default in ~1.8s)"
    - "The budget is CLI-tunable: --max-expressions (persistent flag) reaches render, check, convert, fmt, serve/LSP and [[include]] resolution; 0 disables the cap (pigeon semantics)"
    - "The budget-exhausted error names the current limit and the flag that raises it, instead of reading as a syntax error on the file's last line"
    - "C4D parse errors on the named-parse path (check/render via ParseFile) extract the DSL-native line like ParseAST does — the wrap strips the known path prefix pigeon adds"
    - "Full -race suite green; golangci-lint 0 issues"
  artifacts:
    - path: "internal/c4d/parse.go"
      provides: "MaxExpressions tunable knob (var, render.LabelRatio precedent), default 10_000_000"
    - path: "cmd/c4drill/root.go"
      provides: "Persistent --max-expressions flag bound to the knob"
    - path: "internal/c4d/errors.go"
      provides: "Budget-exhausted message enrichment + named-parse prefix stripping in wrapPigeonError"
---

# Quick Task 261016-mxp: Raise C4D Parse Budget 10x and Add --max-expressions — Plan

User report (Yadro neoChat, 2026-10-16): a valid single-file C4 model
(docs/yadro-inference-cluster.readiness.bigpic.c4d, 719 lines / ~110KB) stopped
parsing on c4drill v1.26.0 with `parse error at line 720: max number of
expressions parsed`. The bound is pigeon's MaxExpressions, hardcoded at 1M in
internal/c4d/parse.go — it trips on file volume, not on any syntax problem, so
every growth of the model re-trips it. The reporter asked for: (1) a raised /
size-derived limit, (2) a flag or env knob, (3) an error message naming the
limit and the way out.

## Tasks

1. **MXP-BUDGET-10X** — `internal/c4d/parse.go`: turn the `maxExpressions`
   const into the exported `MaxExpressions` var (render.LabelRatio precedent),
   default 10_000_000.
2. **MXP-BUDGET-FLAG** — `cmd/c4drill/root.go`: persistent
   `--max-expressions uint` flag bound directly to the knob so every parse
   path in the process (render, check, convert, fmt, serve, includes) honors
   it without threading a parameter; 0 = uncapped (pigeon semantics).
3. **MXP-BUDGET-ERROR-MSG** — `internal/c4d/errors.go`: enrich the
   budget-exhausted message with the current limit and the flag; strip the
   `name:` prefix pigeon adds on named parses so line extraction (and the
   enrichment) land on the check/render path too.

## Verification

- knob test (budget 1 fails a valid doc with the enriched message; 0 parses;
  default parses) + named-parse variant
- CLI tests: flag registered with the 10x default; `--max-expressions 1`
  fails a valid .c4d naming the flag, `0` renders
- LSP parity test re-pinned to the canonical TOML-shaped rendering
- synthetic 222KB .c4d: old 1M → budget error; new default → parse+validate
  exit 0 in ~1.8s; `0` → exit 0
- `go test -race ./...` green; `golangci-lint` 0 issues
