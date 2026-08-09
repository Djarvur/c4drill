---
phase: 3
slug: compatibility-validation
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-08-06
---

# Phase 3 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (Go 1.26.1) + stretchr/testify v1.11.1 |
| **Config file** | none — existing `go test` convention |
| **Quick run command** | `go test ./internal/graph/ ./cmd/c4drill/ ./internal/render/` |
| **Full suite command** | `go test -v -race -cover ./...` (no t.Parallel on render tests — WASM mutex) |
| **Estimated runtime** | ~60 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/graph/ ./cmd/c4drill/ ./internal/render/`
- **After every plan wave:** Run `go test -v -race ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 03-01-01 | 01 | 1 | COMPAT-02 (D-01 fixture) | T-03-01 | N/A (test data) | test-data | `go run ./cmd/c4drill cmd/c4drill/testdata/multilevel.toml -f dot` (exit 0) | ❌ W0 | ⬜ pending |
| 03-01-02 | 01 | 1 | COMPAT-02 (D-02 baseline) | T-03-02 | N/A | golden | `go run ... --output /tmp/multilevel-regold --format dot --expanded && cmp -s ...` | ❌ W0 | ⬜ pending |
| 03-02-01 | 02 | 2 | COMPAT-02 (repoint) | T-03-03 | N/A | unit | `go test ./internal/graph/ -count=1` | ❌ repoint | ⬜ pending |
| 03-02-02 | 02 | 2 | COMPAT-02 (D-04 lock) | T-03-04 | N/A | unit | `go test ./internal/graph/ -run Expanded -count=1` | ❌ new | ⬜ pending |
| 03-02-03 | 02 | 2 | COMPAT-01 (D-03) | T-03-05 | N/A | integration | `go test ./cmd/c4drill/ ./internal/render/ -count=1` | ❌ new | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `cmd/c4drill/testdata/multilevel.toml` — sanitized public fixture (saira structure, generic names)
- [x] `cmd/c4drill/testdata/multilevel.expanded.dot` — committed golden DOT baseline
- [x] Repointed tests: no `.go` test references `cyp-auth-infra` after execution (repo-wide gate)
- [x] D-04 locking test (expanded ignores properties.expanded)
- [x] COMPAT-01 regression (valid.toml → C1 all-collapsed)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Visual spot-check of expanded SVG | COMPAT-02 | SVG aesthetics not assertable | `go run ./cmd/c4drill cmd/c4drill/testdata/multilevel.toml --expanded -f svg` — confirm single all-nested diagram |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 60s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending (approved at verify-work)
