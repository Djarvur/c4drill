---
phase: 14
slug: nesting-validation
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-18
---

# Phase 14 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — existing infrastructure |
| **Quick run command** | `go test ./internal/validator/... -v -run TestValidateNesting` |
| **Full suite command** | `go test ./internal/validator/... -v` |
| **Estimated runtime** | ~2 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/validator/... -v -run TestValidateNesting`
- **After every plan wave:** Run `go test ./internal/validator/... -v`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 2 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 14-01-01 | 01 | 1 | NEST-01 | unit | `go test ./internal/validator/... -v -run TestValidateNestingHierarchy` | ❌ W0 | ⬜ pending |
| 14-01-02 | 01 | 1 | NEST-02 | unit | `go test ./internal/validator/... -v -run TestValidateNestingHierarchy` | ❌ W0 | ⬜ pending |
| 14-01-03 | 01 | 1 | NEST-03 | unit | `go test ./internal/validator/... -v -run TestValidateNestingHierarchy` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/validator/rules_test.go` — add TestValidateNestingHierarchy test cases for NEST-01, NEST-02, NEST-03

*Existing infrastructure covers framework and fixtures.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| None | — | — | — |

All phase behaviors have automated verification.

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 2s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
