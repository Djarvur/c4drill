---
phase: 41
slug: check-command
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-09-03
---

# Phase 41 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (toolchain 1.26.5) + testify v1.12.1 |
| **Config file** | none — go.mod module tests |
| **Quick run command** | `go test ./cmd/c4drill/ -run 'TestCheck' -count=1` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~10 seconds (check tests never render; full suite adds render goldens) |

---

## Sampling Rate

- **After every task commit:** Run `go test ./cmd/c4drill/ -run 'TestCheck' -count=1`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 41-01-01 | 01 | 1 | CHECK-02 | — | N/A | unit (RED) | `go test ./cmd/c4drill/ -run 'TestCheck' -count=1` | ❌ W1 | ⬜ pending |
| 41-01-02 | 01 | 1 | CHECK-02 | T-41 V5 | fail-closed ext dispatch, render-identical VAL errors | unit (GREEN) | `go test ./cmd/c4drill/ -run 'TestCheck' -count=1` | ❌ W1 | ⬜ pending |
| 41-01-03 | 01 | 1 | CHECK-01 | — | no files written, no output dir required | unit | `go test ./cmd/c4drill/ -run 'TestCheck' -count=1` | ❌ W1 | ⬜ pending |
| 41-01-04 | 01 | 1 | CHECK-03 | — | composed multi-file fixture validates as it renders | unit | `go test ./cmd/c4drill/ -run 'TestCheck' -count=1` | ❌ W1 | ⬜ pending |
| 41-02-01 | 02 | 2 | CHECK-04 | — | N/A | source assertion | `grep -c '=== check' README.adoc` + SKILL.md grep | n/a | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements (Go test framework, testify, cmd/c4drill testdata fixtures). New check_test.go and check_* fixtures are TDD deliverables of Wave 1 tasks, not infrastructure gaps.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| `c4drill check --help` does not advertise render flags (D-05 strong reading) | CHECK-01 | cobra help layout is a human-facing artifact; automated assertion is brittle | Build binary, run `c4drill check --help`, confirm no `-o`/`--format`/`--label-ratio` listed |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
