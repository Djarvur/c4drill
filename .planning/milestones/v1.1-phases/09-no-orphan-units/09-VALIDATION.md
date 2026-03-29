---
phase: 09
slug: no-orphan-units
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-11
---

# Phase 09 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing + stretchr/testify v1.11.1 |
| **Config file** | None - standard Go test |
| **Quick run command** | `go test ./internal/validator/... -v -run TestValidateOrphanUnits` |
| **Full suite command** | `go test ./internal/validator/... -v` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/validator/... -v`
- **After every plan wave:** Run `go test ./... -v`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 09-01-01 | 01 | 1 | VAL-01 | unit | `go test ./internal/validator/... -v -run TestValidateOrphanUnits` | ❌ W0 | ⬜ pending |
| 09-01-02 | 01 | 1 | VAL-02 | unit | `go test ./internal/validator/... -v -run TestValidateOrphanUnits` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/validator/rules_test.go` — Add TestValidateOrphanUnits_* test cases
- [x] Test framework installed — Go testing + testify already in go.mod
- [x] Shared fixtures — Use existing pattern (inline test data)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Error message clarity | VAL-02 | User judgment on "clearly lists" | Run `c4drill orphan-example.toml` and verify error message format is actionable |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
