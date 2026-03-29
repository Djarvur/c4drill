---
phase: 02
slug: validation
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-09
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (testing package) |
| **Config file** | .mise.toml (test task) |
| **Quick run command** | `mise run test` |
| **Full suite command** | `go test -v -race -cover ./...` |
| **Estimated runtime** | ~2 seconds |

---

## Sampling Rate

- **After every task commit:** Run `mise run test`
- **After every plan wave:** Run `go test -v -race -cover ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 3 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 02-01-01 | 01 | 1 | VALD-01 | unit | `go test -run TestValidate ./internal/validator/` | ❌ W0 | ⬜ pending |
| 02-01-02 | 01 | 1 | VALD-02 | unit | `go test -run TestValidate ./internal/validator/` | ❌ W0 | ⬜ pending |
| 02-01-03 | 01 | 1 | VALD-03 | unit | `go test -run TestValidate ./internal/validator/` | ❌ W0 | ⬜ pending |
| 02-01-04 | 01 | 1 | VALD-04 | unit | `go test -run TestValidate ./internal/validator/` | ❌ W0 | ⬜ pending |
| 02-02-01 | 02 | 1 | VALD-05 | unit | `go test -run TestValidationError ./internal/validator/` | ❌ W0 | ⬜ pending |
| 02-02-02 | 02 | 1 | VALD-06 | unit | `go test -run TestErrorFormat ./internal/validator/` | ❌ W0 | ⬜ pending |
| 02-03-01 | 03 | 2 | QUAL-01 | integration | `mise run lint` | ✅ exists | ⬜ pending |
| 02-03-02 | 03 | 2 | QUAL-04 | integration | `go test -cover ./...` | ✅ exists | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/validator/validator.go` — validator struct and Validate function
- [ ] `internal/validator/errors.go` — ValidationError type with line, context, suggestion
- [ ] `internal/validator/validator_test.go` — test stubs for all VALD requirements

*Existing infrastructure (mise, go test, testify) covers framework needs.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Error message clarity | VALD-06 | Subjective readability | Run validator on invalid TOML, verify messages are human-readable |
| Suggestion helpfulness | VALD-05 | "did you mean" quality | Verify suggestions appear for typos and are actually helpful |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 3s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
