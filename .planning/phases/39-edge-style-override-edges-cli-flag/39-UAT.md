---
status: complete
phase: 39-edge-style-override-edges-cli-flag
source: [39-01-SUMMARY.md, 39-02-SUMMARY.md, 39-03-SUMMARY.md]
started: 2026-08-31T08:10:25Z
updated: 2026-08-31T08:22:40Z
---

## Current Test

[testing complete]

## Tests

### 1. Help shows the --edges flag
expected: Run `go run ./cmd/c4drill --help` — the flags list includes `--edges string` documented as the edge routing override (straight|spline|square|ortho).
result: pass
evidence: `--edges string        Override edge routing style for every generated diagram (straight|spline|square|ortho)`

### 2. Flag overrides the model's global edges
expected: Run `go run ./cmd/c4drill cmd/c4drill/testdata/edges_override.toml -f dot --edges straight -o /tmp/uat39` — the root dot file (edges_override.dot) contains `splines=false` (straight), even though the model's global edges is spline (which would emit `splines=true`).
result: pass
evidence: root dot contains `splines=false` (model global is spline)

### 3. Flag beats a per-unit edges override
expected: Same command as test 2 — the drill-down dot for the api unit (edges_override/app/api.dot), whose unit declares `edges = "ortho"`, ALSO contains `splines=false` (the flag wins over the per-unit value, which alone would emit `splines=ortho`).
result: pass
evidence: api.dot contains `splines=false` (unit declares ortho)

### 4. square is accepted as the ortho alias
expected: Run with `--edges square` — the root dot contains `splines=ortho` (square is the documented alias of ortho, never the literal word "square").
result: pass
evidence: root dot contains `splines=ortho`

### 5. Invalid value fails loudly
expected: Run with `--edges diagonal` — the command exits non-zero printing an error that names the offending value ("diagonal") and the allowed enum (straight, spline, square, ortho); no output files are written.
result: pass
evidence: `Error: invalid edges: must be straight, spline, square, or ortho: "diagonal"`, exit 1, 0 files written

### 6. Explicit flag survives --plain
expected: Run with `--plain --edges spline` — the root dot still contains `splines=true` (user intent wins). A control run with `--plain` alone emits NO splines attribute (author suppression unchanged).
result: pass
evidence: plain+flag → `splines=true`; plain alone → no splines attribute

### 7. Without the flag, output is unchanged
expected: Run with NO --edges flag — the root dot shows the global spline (`splines=true`) and the api drill-down shows the per-unit ortho (`splines=ortho`) exactly as before this milestone; `go test ./...` passes with zero golden churn.
result: pass
evidence: flag-off root `splines=true`, api `splines=ortho` (pre-milestone fallback chain); full suite ok incl. cmd/c4drill; clean worktree for code

## Summary

total: 7
passed: 7
issues: 0
pending: 0
skipped: 0

## Gaps

[none]
