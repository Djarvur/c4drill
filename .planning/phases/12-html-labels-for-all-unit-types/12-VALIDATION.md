---
phase: 12
slug: html-labels-for-all-unit-types
status: draft
nyquist_compliant: true
wave_0_complete: false
created: 2026-03-13
---

# Phase 12 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — standard Go testing |
| **Quick run command** | `go test ./internal/render/... -v` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/render/... -v`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 12-00-01 | 00 | 0 | HTML-01 | unit | `go test ./internal/render/... -v -run TestHTML` | ❌ W0 | ⬜ pending |
| 12-01-01 | 01 | 1 | HTML-01 | unit | `go build ./...` | ✅ | ⬜ pending |
| 12-01-02 | 01 | 1 | HTML-01 | unit | `go test ./internal/render/... -v -run TestHTML` | ✅ | ⬜ pending |
| 12-01-03 | 01 | 1 | HTML-02 | unit | `go build ./...` | ✅ | ⬜ pending |
| 12-01-04 | 01 | 1 | HTML-02 | integration | `go test ./... -v` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] **Plan 12-00 created** — Wave 0 test infrastructure plan
- [ ] `internal/render/labels_test.go` — add `TestHTML*` test functions for HTML label builders (12-00-01)
- [ ] Test cases for each unit type: Person, DB, Queue, System, Container, Component

*Wave 0 plan (12-00) created. Tests will be implemented before Wave 1.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| SVG visual rendering | HTML-02 | GraphViz rendering output requires visual inspection | Generate sample diagram with all unit types, open SVG in browser, verify each label format |

*Automated tests verify label string generation; visual rendering is manual.*

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 5s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
