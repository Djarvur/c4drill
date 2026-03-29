---
phase: 03
slug: views-graphs
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-09
---

# Phase 03 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (testing package) |
| **Config file** | .mise.toml (test task) |
| **Quick run command** | `mise run test` |
| **Full suite command** | `go test -v -race -cover ./...` |
| **Estimated runtime** | ~3 seconds |

---

## Sampling Rate

- **After every task commit:** Run `mise run test`
- **After every plan wave:** Run `go test -v -race -cover ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|-----------|-------------------|-------------|--------|
| 03-01-01 | 01 | 1 | VIEW-01 | unit | `go test -run TestC1View ./internal/view/` | ❌ W0 | ⬜ pending |
| 03-01-02 | 01 | 1 | VIEW-02 | unit | `go test -run TestC2View ./internal/view/` | ❌ W0 | ⬜ pending |
| 03-01-03 | 01 | 1 | VIEW-03 | unit | `go test -run TestC3View ./internal/view/` | ❌ W0 | ⬜ pending |
| 03-01-04 | 01 | 1 | VIEW-04 | unit | `go test -run TestCollapsedUnit ./internal/view/` | ❌ W0 | ⬜ pending |
| 03-01-05 | 01 | 1 | VIEW-05 | unit | `go test -run TestExpandedCluster ./internal/view/` | ❌ W0 | ⬜ pending |
| 03-02-01 | 02 | 2 | GRPH-01 | unit | `go test -run TestNodeShapes ./internal/graph/` | ❌ W0 | ⬜ pending |
| 03-02-02 | 02 | 2 | GRPH-02 | unit | `go test -run TestEdgeCreation ./internal/graph/` | ❌ W0 | ⬜ pending |
| 03-02-03 | 02 | 2 | GRPH-03 | unit | `go test -run TestEdgeRouting ./internal/graph/` | ❌ W0 | ⬜ pending |
| 03-02-04 | 02 | 2 | GRPH-04 | unit | `go test -run TestClusterCreation ./internal/graph/` | ❌ W0 | ⬜ pending |
| 03-03-01 | 03 | 3 | QUAL-01 | integration | `mise run lint` | ✅ exists | ⬜ pending |
| 03-03-02 | 03 | 3 | QUAL-04 | integration | `go test -cover ./...` | ✅ exists | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/view/view.go` — View types and generator
- [ ] `internal/view/view_test.go` — test stubs for all VIEW requirements
- [ ] `internal/graph/graph.go` — Graph types and builder
- [ ] `internal/graph/graph_test.go` — test stubs for all GRPH requirements

*Existing infrastructure (mise, go test, testify) covers framework needs.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Node shape visual correctness | GRPH-05 | Visual verification of emoji icons | Render a diagram with all unit types, verify shapes match specification |
| Legend visual placement | VIEW-07 | Subjective positioning quality | Verify legend appears in reasonable location without overlapping |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
