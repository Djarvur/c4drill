---
phase: 36
slug: edge-semantics-and-legend
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-08-28
---

# Phase 36 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Source: 36-RESEARCH.md §10 Validation Architecture.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib require-style assertions; canonicalDOT via `internal/testutil/canonical`, source round-trips via `internal/testutil/canonsrc`) |
| **Config file** | none — existing per-package `_test.go` infrastructure |
| **Quick run command** | `go test ./internal/graph/... ./internal/render/... ./internal/view/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~60 seconds full suite |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/graph/... ./internal/render/... ./internal/view/...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

Filled per-plan at execution (task IDs below are placeholders mapped from 36-RESEARCH.md §10):

| Area | Requirement | Test Type | Automated Command / Location | Status |
|------|-------------|-----------|------------------------------|--------|
| Unit styling overrides (node + cluster) | COLOR-01/02 | unit (graph+render) | `go test ./internal/graph/... ./internal/render/...` | ⬜ pending |
| Global edges fallback C2/C3 + square→ortho | GEDGE-01/02 | unit + E2E (view, cmd) | `go test ./internal/view/... ./cmd/c4drill/...` | ⬜ pending |
| rank=reverse emission matrix | RANK-01/02 | unit (render money-test: canonical DOT ≡ `<-` + arrow=reverse idiom) | `go test ./internal/graph/... ./internal/render/... ./internal/view/...` | ⬜ pending |
| kind field + palette + precedence | KIND-01/02 | unit (graph+render) | `go test ./internal/graph/... ./internal/render/...` | ⬜ pending |
| kind round-trip both formats + templates | KIND-03 | unit (parser/c4d/template/cmd convert+fmt) | `go test ./internal/parser/... ./internal/c4d/... ./internal/template/... ./cmd/c4drill/...` | ⬜ pending |
| Collapsed kind colour + style precedence + custom-colour suppression | AGG-01..03 | unit (builder pre-scan) | `go test ./internal/graph/...` | ⬜ pending |
| Legend default-on, content, custom lines | LEG-01..03 | unit (parser/view/render) | `go test ./internal/parser/... ./internal/view/... ./internal/render/...` | ⬜ pending |
| Golden re-baseline legend-only + no-feature regression | BC-01 | golden (canonicalDOT) | `go test ./internal/graph/... ./cmd/c4drill/...` | ⬜ pending |
| Docs examples render + skill sync | DOC-01..03 | CLI + diff | `go run ./cmd/c4drill skill/examples/<new>.toml -f dot`; `diff -r skill plugins/c4drill/skills/c4drill-toml plugins/c4drill/opencode/skills/c4drill-toml` | ⬜ pending |
| Release tag v1.18.0 | REL-01 | process | `git tag v1.18.0 && git push origin v1.18.0` (final phase task) | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

*Existing infrastructure covers all phase requirements.* (go test + canonical/canonsrc packages exist; no framework install needed. The pigeon grammar regen step is tool-dependent — `go generate ./internal/c4d/grammar/...` — already part of the repo toolchain.)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Legend palette sign-off (colours legible/distinct in rendered SVG) | LEG-02, KIND-01 | Visual judgment | Render `skill/examples/03-links.toml` + new edge-kinds fixture to SVG, inspect legend swatches and edge colours on white background |
| Unit colour visual sanity (dark fill → readable font) | COLOR-01 | Visual judgment | Render fixture with dark explicit unit colour; check font luminance rule result |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
