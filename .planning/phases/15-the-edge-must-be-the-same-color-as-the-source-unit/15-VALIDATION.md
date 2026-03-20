---
phase: 15
slug: the-edge-must-be-the-same-color-as-the-source-unit
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-03-20
---

# Phase 15 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing + stretchr/testify |
| **Config file** | None - Go convention |
| **Quick run command** | `go test ./internal/graph/... ./internal/render/... -v -count=1` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/graph/... ./internal/render/... -v -count=1`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 15-01-01 | 01 | 1 | D-01, D-03 | unit | `go test ./internal/graph/... -run TestEdgeColor -v` | ❌ W0 | ⬜ pending |
| 15-01-02 | 01 | 1 | D-02 | unit | `go test ./internal/render/... -run TestEdgeLabelColor -v` | ❌ W0 | ⬜ pending |
| 15-01-03 | 01 | 1 | Integration | integration | `go test ./internal/render/... -run TestEdgeColorRendering -v` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/graph/builder_test.go` — add TestEdgeColorFromSource tests
- [ ] `internal/render/converter_test.go` — add TestEdgeColorRendering tests
- [ ] No framework install needed — existing test infrastructure

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Visual inspection of SVG output | D-01, D-02 | Color perception is subjective | Generate diagram, verify edge colors match source units visually |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
