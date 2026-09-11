---
status: complete
phase: 41-check-command
source: [41-01-SUMMARY.md, 41-02-SUMMARY.md]
started: 2026-09-03
updated: 2026-09-03
---

## Current Test

[testing complete]

## Tests

### 1. Valid model — silent success, zero output files
expected: `c4drill check <valid.c4d>` exits 0, prints nothing, and creates no files or directories.
result: pass

### 2. Invalid model — render-identical errors, exit 1
expected: On the issue #41 orphan model, `check` exits 1 printing byte-identical output to the render path ("unit \"sys.a\" has no incoming or outgoing links…"); verified against an actual render invocation — outputs and exit codes match exactly.
result: pass

### 3. Composed multi-file sources validate as they render
expected: `check` passes the composed examples (skill/examples/08-include/entry.c4d, 09-composed/entry.c4d) — includes resolved, templates expanded, peers resolved — exit 0, silent.
result: pass

### 4. Parse errors surface with stage prefix
expected: A syntactically broken file exits 1 with the render path's stage-prefixed error ("parse: parse error: …").
result: pass

### 5. Documentation
expected: README.adoc documents `check` (section at README.adoc:1474) with usage/exit codes/no-output guarantee; skill/SKILL.md covers it.
result: pass

## Summary

total: 5
passed: 5
issues: 0
pending: 0
skipped: 0

## Gaps

[none]
