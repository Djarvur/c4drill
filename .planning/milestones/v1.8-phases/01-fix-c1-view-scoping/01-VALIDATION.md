---
phase: 1
slug: fix-c1-view-scoping
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-08-06
---

# Phase 1 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (Go 1.26.1 stdlib testing) + stretchr/testify v1.11.1 |
| **Config file** | none — existing `go test` convention (`.mise.toml` task: `go test -v -race -cover ./...`) |
| **Quick run command** | `go test ./internal/view/ ./internal/graph/ ./cmd/c4drill/` |
| **Full suite command** | `go test -v -race -cover ./...` (render tests must NOT use `t.Parallel()` — WASM mutex) |
| **Estimated runtime** | ~60 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/view/ ./internal/graph/ ./cmd/c4drill/`
- **After every plan wave:** Run `go test -v -race ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 01-01-01 | 01 | 1 | VIEW-01 | — | N/A (no new input surface) | unit | `go test ./internal/view/ -count=1` | ✅ scope_test.go | ⬜ pending |
| 01-01-02 | 01 | 1 | VIEW-02 | — | N/A | unit | `go test ./internal/view/ -run C1 -count=1` | ❌ W0 | ⬜ pending |
| 01-01-03 | 01 | 1 | EDGE-01 | — | N/A | unit | `go test ./internal/view/ -run C1 -count=1` | ❌ W0 | ⬜ pending |
| 01-01-04 | 01 | 1 | EDGE-02 | — | N/A | unit | `go test ./internal/graph/ -run BuildEdges -count=1` | ✅ builder_test.go | ⬜ pending |
| 01-01-05 | 01 | 1 | EDGE-02 (penwidth) | — | N/A | unit | `go test ./internal/graph/ -count=1` | ❌ W0 | ⬜ pending |
| 01-01-06 | 01 | 1 | COMPAT-02 | — | N/A | integration | `go test ./cmd/c4drill/ -count=1` | ✅ root_test.go | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/view/scope_test.go` — new tests: deepest-visible-ancestor resolution (expanded C1), source-side resolution, within-cluster edges
- [ ] `internal/graph/builder_test.go` — new tests: pair-only collapse (distinct tech/desc → one edge), penwidth 1.0/2.0, minlen gating on resolved edges
- [ ] `cmd/c4drill/root_test.go` — `--expanded` baseline output unchanged (COMPAT-02)

*Existing infrastructure covers the base behavior — no framework gaps.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Visual check of expanded-cluster edge placement | VIEW-02, EDGE-01 | GraphViz layout aesthetics not assertable | Render `saira-20260320.c2.full.toml` with `properties.expanded`; confirm edges attach to visible subunit nodes, not cluster labels |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
