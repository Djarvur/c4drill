---
phase: 2
slug: auto-generate-c2-c3-diagrams
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-08-06
---

# Phase 2 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (Go 1.26.1 stdlib testing) + stretchr/testify v1.11.1 |
| **Config file** | none — existing `go test` convention |
| **Quick run command** | `go test ./internal/view/ ./internal/graph/ ./cmd/c4drill/` |
| **Full suite command** | `go test -v -race -cover ./...` (no t.Parallel on render tests — WASM mutex) |
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
| 02-01-01 | 01 | 1 | VIEW-05 (D-07) | T-02-01 | N/A (slice-length guard) | unit (TDD) | `go test ./internal/graph/ -run ExpandedEmpty -count=1` | ❌ W0 | ⬜ pending |
| 02-01-02 | 01 | 1 | VIEW-05 (D-04/D-05/D-06) | T-02-02 | N/A | unit | `go test ./internal/view/ ./internal/graph/ -count=1` | ❌ W0 | ⬜ pending |
| 02-01-03 | 01 | 1 | VIEW-03, VIEW-04 (D-01/D-02/D-03/D-08) | T-02-03 | N/A | unit | `go test ./cmd/c4drill/ ./internal/view/ -count=1` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/graph/builder_test.go` — expanded-but-empty unit renders as plain node (D-07)
- [ ] `cmd/c4drill/root_test.go` — box with subunits generates a sub-diagram file (D-01)
- [ ] `internal/view/scope_test.go` — actor boundary node in C2 view (D-08) if not already covered

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Visual check of box sub-diagram rendering | D-01 | GraphViz aesthetics | Render a TOML with a box containing systems; open `{basename}/{box}.svg` |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
