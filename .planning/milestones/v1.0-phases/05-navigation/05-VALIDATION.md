---
phase: 05
slug: navigation
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-10
---

# Phase 05 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (testing package) |
| **Config file** | .mise.toml (test task) |
| **Quick run command** | `mise run test` |
| **Full suite command** | `go test -v -race -cover ./...` |
| **Estimated runtime** | ~5 seconds |

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
| 05-01-01 | 01 | 1 | REND-04 | unit | `go test -run TestExploreLink ./internal/graph/...` | ❌ W0 | ⬜ pending |
| 05-01-02 | 01 | 1 | REND-04 | unit | `go test -run TestBuildGraph ./internal/graph/...` | ✅ exists | ⬜ pending |
| 05-02-01 | 02 | 1 | REND-05 | unit | `go test -run TestBackLink ./internal/render/...` | ❌ W0 | ⬜ pending |
| 05-02-02 | 02 | 1 | REND-06 | unit | `go test -run TestBreadcrumbs ./internal/render/...` | ❌ W0 | ⬜ pending |
| 05-02-03 | 02 | 1 | OUTP-05 | unit | `go test -run TestRelativePath ./internal/render/...` | ❌ W0 | ⬜ pending |
| 05-03-01 | 03 | 2 | QUAL-01 | integration | `mise run lint` | ✅ exists | ⬜ pending |
| 05-03-02 | 03 | 2 | QUAL-04 | integration | `go test -cover ./internal/graph/... ./internal/render/...` | ✅ exists | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/graph/navigation_test.go` — test stubs for explore link path computation
- [ ] `internal/render/navigation_test.go` — test stubs for back-link and breadcrumbs

*Existing infrastructure (mise, go test, testify) covers framework needs.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| SVG clickability | REND-04 | Requires visual verification in SVG viewer | Open generated SVG in browser, verify nodes are clickable with correct targets |
| Navigation bar positioning | REND-05, Visual verification of layout | Open SVG, verify back-link and breadcrumbs appear at top of diagram area |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
