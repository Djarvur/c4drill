---
phase: 40
slug: deterministic-svg-output
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-09-03
---

# Phase 40 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go standard `testing` + testify v1.12.1 (go1.26 toolchain) |
| **Config file** | none — `go test` conventions |
| **Quick run command** | `go test ./internal/validator/ ./internal/graph/ ./internal/render/ -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~8s full suite (verified 2026-09-03); quick run ~2s |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/validator/ ./internal/graph/ ./internal/render/ -count=1`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 40-01-01 | 01 | 1 | REPRO-01 | — | N/A | integration (TDD RED) | `go test ./internal/render/ -run TestDeterministicByteIdenticalOutput -count=1` | ❌ W0 | ⬜ pending |
| 40-01-02 | 01 | 1 | REPRO-02 | — | N/A | unit (TDD RED) | `go test ./internal/validator/ -run TestMirrorOrder -count=1` | ❌ W0 | ⬜ pending |
| 40-01-03 | 01 | 1 | REPRO-01/02/03 | — | N/A | fix (TDD GREEN) | `go test ./internal/validator/ ./internal/render/ -count=1` | ❌ W0 | ⬜ pending |
| 40-02-01 | 02 | 2 | REPRO-02 | T-Q1-01 | edge cgraph names stay sanitized | unit | `go test ./internal/graph/ -run TestEdgeOrder -count=1` | ❌ W0 | ⬜ pending |
| 40-02-02 | 02 | 2 | REPRO-01/03 (D-06 gate) | — | N/A | regression gate | `go test ./... -count=1` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/render` byte-equality regression test (issue #42 reproducer, dot/svg/html) — TDD RED
- [ ] `internal/validator` mirror-order determinism unit test — TDD RED

*Framework install: none needed.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| (none) | — | All phase behaviors have automated verification | — |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
