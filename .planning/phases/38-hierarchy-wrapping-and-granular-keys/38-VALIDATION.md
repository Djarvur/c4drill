---
phase: 38
slug: hierarchy-wrapping-and-granular-keys
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-08-30
---

# Phase 38 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib require/assert per repo convention) |
| **Config file** | go.mod (Go 1.26.1); canonical comparator `internal/testutil/canonical` (DI-1) |
| **Quick run command** | `go test ./internal/view/... ./internal/graph/... ./internal/render/...` |
| **Full suite command** | `go test -count=1 ./...` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** quick run command + the touched package
- **After every plan wave:** full suite
- **Before `/gsd:verify-work`:** Full suite green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------|-------------------|-------------|--------|
| 38-01-xx | 01 | 1 | WRAP-01/02 | T-38-01 | unit (graph+view) | `go test ./internal/graph/ ./internal/view/ -run 'Wrap' -v` | ❌ W0 (new tests) | ⬜ pending |
| 38-0x-xx | 0x | 1 | WRAP-03 | — | unit (invariance) | `go test ./internal/view/ -run 'NodeSetInvariant' -v` | ❌ W0 | ⬜ pending |
| 38-0x-xx | 0x | 2 | KEY-01/02 | T-38-02 | unit (guards) | `go test ./internal/graph/ -run 'NoColors\|NoStyles\|NoLength\|NoRank' -v` | ❌ W0 | ⬜ pending |
| 38-0x-xx | 0x | 2 | LBL-01..03 | T-38-03 | unit + E2E | `go test ./internal/render/ ./cmd/c4drill/ -run 'NoLabels' -v` | ❌ W0 | ⬜ pending |
| 38-0x-xx | 0x | 3 | KEY-03 | — | E2E matrix | `go test ./cmd/c4drill/ -run 'KeyComposition' -v` | ❌ W0 | ⬜ pending |
| 38-0x-xx | 0x | 3 | BC-01 | T-38-04 | canonical goldens | `go test -count=1 ./...` | ✅ (re-baselined once, in-phase) | ⬜ pending |
| 38-0x-xx | 0x | 4 | DOC-01..03 | — | CI parity + render | `diff -r skill plugins/c4drill/skills/c4drill-toml && diff -r skill plugins/c4drill/opencode/skills/c4drill-toml` + renders | ✅ | ⬜ pending |
| 38-0x-xx | 0x | 4 | REL-01 | — | release tag | `git tag v1.22.0` + `gh release view v1.22.0` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] New test stubs land WITH their features (TDD RED first)
- Existing infrastructure covers all needs — no installs

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Wrapper clusters read legibly; labels-off layout is readable | WRAP-01 / LBL-01 | Visual judgment | Render multilevel + a labeled fixture with/without flags → open SVGs |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
