---
phase: 39
slug: edge-style-override-edges-cli-flag
status: draft
nyquist_compliant: true
wave_0_complete: false
created: 2026-08-31
---

# Phase 39 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (Go 1.26) + testify v1.12.1 |
| **Config file** | none (go.mod; mise tasks in `.mise.toml`) |
| **Quick run command** | `go test ./cmd/c4drill/ ./internal/graph/ -run 'Edges' -v` |
| **Full suite command** | `go test ./...` (CI form: `go test -v -race -cover ./...`; lint: `golangci-lint run ./...`) |
| **Estimated runtime** | ~30-60 seconds full suite (go-graphviz WASM engine dominates) |

---

## Sampling Rate

- **After every task commit:** Run `go test ./...`
- **After every plan wave:** Run `go test -v -race -cover ./...` + `golangci-lint run ./...`
- **Before `/gsd:verify-work`:** Full suite must be green — critically, ALL existing canonicalDOT goldens pass with ZERO re-baselining (GEDGE-08)
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

Task IDs below follow the expected plan shape (implementation plan 01; adjust IDs if the planner splits differently — the requirement→test mapping is the contract).

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 39-01-01 | 01 | 1 | GEDGE-03 | T-39-01 | `--edges` value allow-listed before any file I/O | unit (TDD) | `go test ./cmd/c4drill/ -run 'TestEdges' -v` | ❌ W0 | ⬜ pending |
| 39-01-02 | 01 | 1 | GEDGE-04 | T-39-01 | Invalid value → loud sentinel error naming value + enum; no output written | unit (TDD) | `go test ./cmd/c4drill/ -run 'TestEdges' -v` | ❌ W0 | ⬜ pending |
| 39-01-03 | 01 | 1 | GEDGE-05 | — | Flag beats global AND per-unit `edges` on C1/drill-down/expanded | unit+e2e (TDD) | `go test ./cmd/c4drill/ ./internal/graph/ -run 'Edges' -v` | ❌ W0 | ⬜ pending |
| 39-01-04 | 01 | 1 | GEDGE-06 | — | `--plain --edges spline` → `splines=true` in RAW dot | e2e pin | `go test ./cmd/c4drill/ -run 'TestEdges' -v` | ❌ W0 | ⬜ pending |
| 39-01-05 | 01 | 1 | GEDGE-07 | — | `--edges` × generation × `--plain` matrix asserts `splines=` in RAW dot | e2e matrix | `go test ./cmd/c4drill/ -run 'TestEdges' -v` | ❌ W0 | ⬜ pending |
| 39-01-06 | 01 | 1 | GEDGE-08 | — | Flag-off: existing goldens untouched (zero re-baseline) | regression | `go test ./...` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `cmd/c4drill/root_test.go` — `TestEdges*` family stubs (accept/reject, plain pin, matrix cells) — same file as KEY-03 matrix
- [ ] `internal/graph/builder_test.go` — builder-level plain×override unit pin (PLAIN-02 interaction)
- [ ] Fixture line for per-unit `edges` (D-03 pin) — extend `cmd/c4drill/testdata/multilevel.toml` or equivalent
- [ ] Framework install: none needed (go test + testify already in place)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Visual routing appearance of ortho/spline variants in SVG | GEDGE-03 (supplementary) | Geometry is graphviz-owned; dot attribute is the automated proxy | `go run ./cmd/c4drill testdata/valid.toml -f svg -o /tmp --edges ortho` and eyeball elbow routing vs `--edges spline` |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending (approves automatically when Wave 0 items land green during execution)
