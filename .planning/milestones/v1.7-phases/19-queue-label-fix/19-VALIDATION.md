---
phase: 19
slug: queue-label-fix
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-24
---

# Phase 19 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing + stretchr/testify v1.11.1 |
| **Config file** | None — tests use t.Run() pattern |
| **Quick run command** | `go test ./internal/render/... -run TestHTMLQueueLabel -v` |
| **Full suite command** | `go test ./internal/render/... -v` |
| **Estimated runtime** | ~2 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/render/... -v -count=1`
- **After every plan wave:** Run `go test ./internal/render/... -v`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 19-01-01 | 01 | 1 | QUEUE-FIX-01 | unit | `go test ./internal/render/... -run TestHTMLQueueLabel -v` | ✅ Exists | ⬜ pending |
| 19-01-02 | 01 | 1 | QUEUE-FIX-02 | unit | `go test ./internal/render/... -run TestHTMLQueueLabel -v` | ✅ Exists | ⬜ pending |
| 19-01-03 | 01 | 1 | QUEUE-FIX-03 | unit | `go test ./internal/render/... -run TestHTMLQueueLabel -v` | ✅ Exists | ⬜ pending |
| 19-01-04 | 01 | 1 | QUEUE-FIX-04 | unit | `go test ./internal/render/... -run TestCreateNode -v` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/render/converter_test.go` — Add test verifying Queue nodes use box shape, not cylinder
- [ ] `internal/render/labels_test.go` — Update `TestHTMLQueueLabel` to verify ASCII graphic row presence

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Visual inspection of Queue label | QUEUE-FIX-01 | ASCII art rendering depends on font | Generate SVG and verify `═╦╩═╦═══` appears correctly |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
