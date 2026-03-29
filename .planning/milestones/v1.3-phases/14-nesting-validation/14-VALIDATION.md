---
phase: 14
slug: nesting-validation
status: planned
nyquist_compliant: true
wave_0_complete: true
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
| 14-01-01 | 01 | 1 | NEST-01 | unit | `go test ./internal/validator/... -v -run TestValidateNestingHierarchy` | TDD (Task 1) | ⬜ pending |
| 14-01-02 | 01 | 1 | NEST-02 | unit | `go test ./internal/validator/... -v -run TestValidateNestingHierarchy` | TDD (Task 1) | ⬜ pending |
| 14-01-03 | 01 | 1 | NEST-03 | unit | `go test ./internal/validator/... -v -run TestValidateNestingHierarchy` | TDD (Task 1) | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `internal/validator/rules_test.go` — tests written in Task 1 (TDD approach: tests before implementation)

*Existing infrastructure covers framework and fixtures. TDD approach used - tests created as first task in execution.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| None | — | — | — |

All phase behaviors have automated verification.

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references (TDD approach)
- [x] No watch-mode flags
- [x] Feedback latency < 2s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved
