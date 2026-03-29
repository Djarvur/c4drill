---
phase: 04
slug: rendering-output
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-03-10
---

# Phase 04 — Validation Strategy

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
| 04-01-01 | 01 | 1 | REND-01 | unit | `go test -run TestRenderDOT ./internal/render/` | ❌ W0 | ⬜ pending |
| 04-01-02 | 01 | 1 | REND-02 | unit | `go test -run TestRenderSVG ./internal/render/` | ❌ W0 | ⬜ pending |
| 04-01-03 | 01 | 1 | REND-03 | unit | `go test -run TestFormatFlag ./internal/render/` | ❌ W0 | ⬜ pending |
| 04-02-01 | 02 | 1 | OUTP-01 | unit | `go test -run TestOutputPathC1 ./internal/output/` | ❌ W0 | ⬜ pending |
| 04-02-02 | 02 | 1 | OUTP-02 | unit | `go test -run TestOutputPathExpanded ./internal/output/` | ❌ W0 | ⬜ pending |
| 04-02-03 | 02 | 1 | OUTP-04 | unit | `go test -run TestCreateDir ./internal/output/` | ❌ W0 | ⬜ pending |
| 04-03-01 | 03 | 2 | QUAL-01 | integration | `mise run lint` | ✅ exists | ⬜ pending |
| 04-03-02 | 03 | 2 | QUAL-04 | integration | `go test -cover ./...` | ✅ exists | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/render/render_test.go` — test stubs for REND requirements
- [ ] `internal/output/writer_test.go` — test stubs for OUTP requirements

*Existing infrastructure (mise, go test, testify) covers framework needs.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| DOT visual correctness | REND-01 | Visual verification of generated DOT syntax | Render a sample graph, verify DOT output is well-formed |
| SVG visual correctness | REND-02 | Visual verification of rendered diagram | Open generated SVG in browser, verify shapes and labels |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
