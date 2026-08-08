---
phase: 33
slug: docs-sweep-end-to-end-goldens
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-08-08
---

# Phase 33 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib `testing`) + testify (`require`/`assert`) |
| **Config file** | none (Go convention — `go test` auto-discovers `*_test.go`) |
| **Quick run command** | `go test ./internal/testutil/canonical/ ./internal/graph/ ./cmd/c4drill/ -x` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~15 seconds (canonicalDOT unit tests are instant; E2E goldens exercise go-graphviz ~2-5s each) |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/testutil/canonical/ ./internal/graph/ ./cmd/c4drill/ -x`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

> Task IDs are provisional (TBD until PLAN.md files exist). The Plan/Critical-Task columns will be filled by the planner; the Requirement / Test Type / Command columns are stable from research.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 33-01-01 | 01 | 1 | (infra) | — | N/A — test-only helper | unit (regression) | `go test ./internal/testutil/canonical/ -x` | ❌ W1 (new, moved) | ⬜ pending |
| 33-01-02 | 01 | 1 | (backward-compat) | — | N/A | e2e golden | `go test ./internal/graph/ -run 'TestBuildExpandedGraphBaselineDOT|TestReference_BackwardCompat' -x` | ✅ exists | ⬜ pending |
| 33-02-01 | 02 | 1 | DOC-03 | — | N/A | file existence | `ls skill/examples/06-templates.toml 07-relative-peer.toml 08-include/ 09-composed/` | ❌ W1 (new) | ⬜ pending |
| 33-03-01 | 03 | 1 | DOC-01 | — | N/A | doc assertion | `grep -c 'defaultTypeForParent\|inferGenericType' README.md skill/SKILL.md` (planner sets threshold) | ❌ W1 (docs) | ⬜ pending |
| 33-03-02 | 03 | 1 | DOC-02 | — | N/A | doc assertion | `grep -E 'Templates|Include|Relative.{0,3}Peer' README.md skill/SKILL.md` (planner refines) | ❌ W1 (docs) | ⬜ pending |
| 33-04-01 | 04 | 2 | XC-05 | — | N/A | e2e golden | `go test ./cmd/c4drill/ -run TestXC05_ComposedEquivSingleFile -x` | ❌ W2 (new) | ⬜ pending |
| 33-04-02 | 04 | 2 | XC-01 | — | N/A | e2e behavioral | `go test ./cmd/c4drill/ -run TestXC01_PipelineOrdering -x` | ❌ W2 (new) | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/testutil/canonical/canonical.go` — extracted helper (Plan 1, Wave 1 — this is the "Wave 0" for the test-infra tasks since all Wave 2 tests depend on it)

*Framework install: none needed — Go toolchain + testify already present.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| (none) | — | — | — |

*All phase behaviors have automated verification. DOC-01/02/03 are covered by file-existence + grep assertions (deterministic); XC-01/05 by automated Go tests.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 15s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
