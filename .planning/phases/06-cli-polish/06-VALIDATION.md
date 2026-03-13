---
phase: 06
slug: cli-polish
status: draft
nyquist_compliant: false
wave_0_complete: true
created: 2026-03-10
---

# Phase 06 — Validation Strategy

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
| 06-01-01 | 01 | 1 | CLII-01, CLII-02, CLII-03 | unit | `go test -run TestRootCmd ./cmd/c4drill/...` | ✅ exists | ⬜ pending |
| 06-01-02 | 01 | 1 | CLII-04 | unit | `go test -run TestHelp ./cmd/c4drill/...` | ✅ exists | ⬜ pending |
| 06-01-03 | 01 | 1 | CLII-05, CLII-06 | unit | `go test -run TestExitCodes ./cmd/c4drill/...` | ✅ exists | ⬜ pending |
| 06-02-01 | 02 | 1 | CLII-01, OUTP-03 | integration | `go test -run TestFullPipeline ./cmd/c4drill/...` | ✅ exists | ⬜ pending |
| 06-02-02 | 02 | 2 | QUAL-04 | integration | `go test -cover ./cmd/c4drill/...` | ✅ exists | ⬜ pending |
| 06-02-03 | 02 | 2 | QUAL-01 | integration | `mise run lint` | ✅ exists | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `cmd/c4drill/main.go` — existing CLI entry point to refactor

*Existing infrastructure (mise, go test, testify) covers framework needs.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Help text display | CLII-04 | Visual verification of formatting | Run `c4drill --help`, verify usage examples and flag descriptions appear |
| Error output to stderr | CLII-06 | Shell redirection test | Run with invalid input, verify `2>&1` shows errors, `1>/dev/null` does not |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 5s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
