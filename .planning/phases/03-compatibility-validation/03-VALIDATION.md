---
phase: 3
slug: compatibility-validation
status: draft
nyquist_compliant: false
wave_0_complete: false
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
| **Quick run command** | `go test ./internal/graph/ ./cmd/c4drill/` |
| **Full suite command** | `go test -v -race -cover ./...` (no t.Parallel on render tests — WASM mutex) |
| **Estimated runtime** | ~60 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/graph/ ./cmd/c4drill/`
- **After every plan wave:** Run `go test -v -race ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 03-01-01 | 01 | 1 | COMPAT-02 (fixture) | — | N/A | test-data | `go test ./internal/graph/ -count=1` | ❌ W0 | ⬜ pending |
| 03-01-02 | 01 | 1 | COMPAT-02 (baseline) | — | N/A | unit | `go test ./internal/graph/ -run BaselineDOT -count=1` | ❌ repoint | ⬜ pending |
| 03-01-03 | 01 | 1 | COMPAT-02 (D-04) | — | N/A | unit | `go test ./internal/graph/ -run Expanded -count=1` | ❌ new | ⬜ pending |
| 03-01-04 | 01 | 1 | COMPAT-01 | — | N/A | integration | `go test ./cmd/c4drill/ -count=1` | ⚠ verify/add | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `cmd/c4drill/testdata/` — sanitized public fixture (saira structure, generic names, `length` attrs)
- [ ] `cmd/c4drill/testdata/` — committed golden DOT baseline for the public fixture
- [ ] Repointed tests: no test reads `../../cyp-auth-infra/` paths
- [ ] D-04 regression test (expanded ignores properties.expanded)
- [ ] COMPAT-01 regression test (valid.toml → C1 all-collapsed)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Visual spot-check of expanded SVG | COMPAT-02 | SVG aesthetics not assertable | `go run ./cmd/c4drill testdata/expanded.toml --expanded -f svg` — confirm single all-nested diagram |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
