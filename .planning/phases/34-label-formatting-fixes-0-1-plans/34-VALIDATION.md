---
phase: 34
slug: label-formatting-fixes-0-1-plans
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-08-10
---

# Phase 34 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (built-in) + testify assert/require |
| **Config file** | none — Go conventions (`*_test.go` beside source) |
| **Quick run command** | `go test ./internal/render/ -run '<affected test>'` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~30-60 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/render/ -run '<affected tests>'` (scope to the plan's own tests — avoids cross-plan interference in parallel waves)
- **After every plan wave:** Run `go test ./internal/render/... ./internal/graph/... ./cmd/c4drill/...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** ~60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 34-01-01 | 01 | 1 | LABEL-02 | — | N/A (test re-assertion) | unit | `go test ./internal/render/ -run TestWrapText` (MUST FAIL at RED) | ✅ | ⬜ pending |
| 34-01-02 | 01 | 1 | LABEL-02 | — | N/A (wrap branch) | unit | `go test ./internal/render/ -run TestWrapText` | ✅ | ⬜ pending |
| 34-01-03 | 01 | 1 | LABEL-02, COMPAT-01 | — | N/A (no new input surface) | unit + golden | `grep -rn splitLongWord internal/` (0 hits) && `go test ./cmd/c4drill/... ./internal/graph/...` | ✅ | ⬜ pending |
| 34-02-01 | 02 | 1 | LABEL-01 | — | N/A (test re-assertion) | unit | `go test ./internal/render/ -run TestEdgeLabelGeneration` (MUST FAIL at RED) | ✅ | ⬜ pending |
| 34-02-02 | 02 | 1 | LABEL-01 | T-34-01 (Tampering, escaping) | Reuse `wrapAndEscape`/`html.EscapeString` in edge label rows | unit | `go test ./internal/render/ -run TestEdgeLabelGeneration` | ✅ | ⬜ pending |
| 34-02-03 | 02 | 1 | LABEL-01, COMPAT-01 | T-34-01 | N/A (golden check) | integration/golden | `go test ./cmd/c4drill/... ./internal/graph/...` (canonicalDOT, DI-1) | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/render/wrap_internal_test.go` — TestWrapText "forced character break"/"multi-byte unicode" re-asserted to unsplit-overflow (RED step of plan 01)
- [ ] `internal/render/labels_test.go` — TestEdgeLabelGeneration re-asserted to HTML table form + D-04 rectangle-always cases (RED step of plan 02)
- [ ] No framework install — Go testing built-in

*Existing infrastructure covers all phase requirements; only re-assertions are needed.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Rendered SVG shows edge labels as borderless rectangles | LABEL-01 | Cosmetic confirmation of aspect ratio in actual output | `go run ./cmd/c4drill -f svg <diagram-with-edge-labels>` and eyeball the edge label rectangle proportions |

*Optional — all functional behaviors have automated verification.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** {pending / approved YYYY-MM-DD}
