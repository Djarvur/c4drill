---
status: complete
phase: 40-deterministic-svg-output
source: [40-01-SUMMARY.md, 40-02-SUMMARY.md]
started: 2026-09-03
updated: 2026-09-03
---

## Current Test

[testing complete]

## Tests

### 1. Issue #42 reproducer — repeated renders byte-identical (SVG)
expected: Render the exact `.c4d` from issue #42 (10 containers, %3 edge pattern) three times to separate output dirs; `diff -r` shows zero differences, `id="edge<N>"` groups sequential and stable.
result: pass

### 2. Repeatability at 5 runs
expected: Five consecutive renders to separate dirs are all byte-identical (no run-pair drift).
result: pass

### 3. Determinism holds for dot and html formats
expected: Double renders with `-f dot` and `-f html` are byte-identical per format (REPRO-03).
result: pass

### 4. No semantic change — suite and goldens
expected: Full `go test ./...` green with canonicalDOT goldens untouched; binary-level acceptance rendered against the shipped build (`go build ./cmd/c4drill`).
result: pass

## Summary

total: 4
passed: 4
issues: 0
pending: 0
skipped: 0

## Gaps

[none]
