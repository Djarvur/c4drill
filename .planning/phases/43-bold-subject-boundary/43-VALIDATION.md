---
phase: 43
slug: bold-subject-boundary
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-09-03
---

# Phase 43 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (toolchain 1.26.x) + testify v1.12.1 |
| **Config file** | none — go.mod module tests |
| **Quick run command** | `go test ./internal/graph/ ./internal/render/ -run 'SubjectBoundary|Bold|BoundaryCluster' -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~10 seconds (warm cache; phases 40/41 precedent) |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/graph/ ./internal/render/ -count=1`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 43-01-01 | 01 | 1 | BOLD-01, BOLD-02 | — | constant `penwidth=3.0` emitted only on subject boundary cluster; N/A semantics | unit (TDD RED) | `go test ./internal/graph/ ./internal/render/ -run 'SubjectBoundary|Bold' -count=1` | ❌ W1 | ⬜ pending |
| 43-01-02 | 01 | 1 | BOLD-01, BOLD-02 | — | emphasis set at construction + emitted via setClusterAttribute; survives --plain/--no-styles | unit (GREEN) | `go test ./internal/graph/ ./internal/render/ -count=1` | ❌ W1 | ⬜ pending |
| 43-01-03 | 01 | 1 | BOLD-03 | — | existing canonical goldens green; C1 root + --expanded unchanged | integration + regression | `go test ./... -count=1` | ❌ W1 | ⬜ pending |
| 43-02-01 | 02 | 2 | BOLD-03 | — | golden re-baseline audit: per-golden canonical diff = sole `penwidth` line on subject cluster (expected: zero) | integration + source audit | `go test ./... -count=1` | ❌ W1 | ⬜ pending |
| 43-02-02 | 02 | 2 | BOLD-03 | — | PROJECT.md v1.18 "As of" paragraph documents semantic bold boundary | source assertion | `grep -c 'As of v1.18' .planning/PROJECT.md` | n/a | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements: `cmd/c4drill/testdata/multilevel.toml` (multi-level fixture with C2 + deep-link C3 drill-downs), `generateMultilevelOutput` CLI harness (root_test.go:680-800), canonical comparator (`internal/testutil/canonical`), raw-DOT attribute extraction precedent (`dotEdgeMultiset`, converter_test.go:1257). New test files are TDD RED deliverables of task 43-01-01, not infrastructure gaps.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| (optional visual) bold `penwidth=3.0` cluster frame in the generated SVG/PNG | BOLD-01 | graphviz rendering quality is a human-felt check; raw-DOT assertion is the automated gate | Open a generated C2/C3 SVG; confirm the subject boundary frame renders visibly thicker than other clusters |

All phase behaviors have automated verification; the visual check above is optional confirmation, never a gate.

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 15s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending