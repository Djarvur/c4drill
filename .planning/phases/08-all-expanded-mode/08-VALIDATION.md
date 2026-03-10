---
phase: 08
slug: all-expanded-mode
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-11
---

# Phase 08 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | stretchr/testify v1.11.1 (Go testing) |
| **Config file** | None — Go testing convention |
| **Quick run command** | `go test ./internal/view/... ./internal/graph/... ./cmd/c4drill/... -v` |
| **Full suite command** | `go test ./... -v` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/view/... ./internal/graph/... ./cmd/c4drill/... -v`
- **After every plan wave:** Run `go test ./... -v`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 08-01-01 | 01 | 1 | EXPD-02 | unit | `go test ./internal/view/... -run TestGenerateAllExpandedView -v` | ❌ W0 | ⬜ pending |
| 08-01-02 | 01 | 1 | EXPD-03 | unit | `go test ./internal/graph/... -run TestBuildNestedCluster -v` | ❌ W0 | ⬜ pending |
| 08-02-01 | 02 | 1 | EXPD-04 | unit | `go test ./internal/output/... -run TestWriteExpanded -v` | ❌ W0 | ⬜ pending |
| 08-02-02 | 02 | 1 | EXPD-01 | unit | `go test ./cmd/c4drill/... -run TestExpandedFlag -v` | ❌ W0 | ⬜ pending |
| 08-02-03 | 02 | 1 | EXPD-05 | regression | `go test ./... -v` | ✅ Existing | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/view/scope_test.go` — add TestGenerateAllExpandedView* tests (EXPD-02)
- [ ] `internal/graph/builder_test.go` — add TestBuildNestedCluster tests (EXPD-03)
- [ ] `internal/output/writer_test.go` — add TestWriteExpanded tests (EXPD-04)
- [ ] `cmd/c4drill/root_test.go` — add TestExpandedFlag tests (EXPD-01)
- [ ] Test fixture: Create `testdata/nested.toml` with 3+ levels of nesting

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Visual cluster nesting quality | EXPD-02 | Visual inspection required | Run `c4drill testdata/nested.toml --expanded -f svg`, open SVG, verify nested clusters appear correctly |
| Cross-level edge routing | EXPD-03 | Visual inspection required | In SVG output, verify edges between units at different nesting depths are visible and not obscured |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
