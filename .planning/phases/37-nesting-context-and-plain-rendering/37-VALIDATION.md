---
phase: 37
slug: nesting-context-and-plain-rendering
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-08-30
---

# Phase 37 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib require/assert helpers per repo convention) |
| **Config file** | go.mod (Go 1.26.1); canonical comparator `internal/testutil/canonical` (DI-1) |
| **Quick run command** | `go test ./internal/view/... ./internal/graph/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~30 seconds full suite |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/view/... ./internal/graph/...` (plus the package the task touched)
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 37-01-xx | 01 | 1 | CTX-03 | — | N/A | unit (graph) | `go test ./internal/graph/...` | ❌ W0 (new tests in builder_test.go) | ⬜ pending |
| 37-0x-xx | 0x | 1 | CTX-02 | — | N/A | unit (view) | `go test ./internal/view/...` | ❌ W0 (new tests in scope_test.go) | ⬜ pending |
| 37-0x-xx | 0x | 1 | CTX-01 | — | N/A | unit + E2E | `go test ./internal/view/... ./cmd/c4drill/...` | ❌ W0 | ⬜ pending |
| 37-0x-xx | 0x | 2 | PLAIN-01/02 | — | N/A | unit (graph+render) | `go test ./internal/graph/... ./internal/render/...` | ❌ W0 | ⬜ pending |
| 37-0x-xx | 0x | 2 | PLAIN-03/04 | — | N/A | unit + E2E golden | `go test ./internal/render/... ./cmd/c4drill/...` | ❌ W0 | ⬜ pending |
| 37-0x-xx | 0x | 2 | BC-01 | — | N/A | canonicalDOT goldens | `go test ./...` | ✅ (existing goldens) | ⬜ pending |
| 37-0x-xx | 0x | 3 | DOC-01..03 | — | N/A | CI parity + render | `.github/workflows/validate-skill-examples.yml` (diff -r ×2, example renders) | ✅ | ⬜ pending |
| 37-0x-xx | 0x | 3 | REL-01 | — | N/A | release tag | `git tag v1.21.0` + release workflow | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] New test stubs land WITH their features (TDD mode — RED first): scope_test.go (chain entries, true-target resolution), builder_test.go (nested clusters, plain guards), render tests (plain labels/attrs), cmd/c4drill E2E (--plain golden)
- Existing infrastructure covers all framework needs — no installs

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Unfolded-container drill affordance reads well visually | CTX-03 | SVG visual judgment | Render multilevel.toml → open SVG → confirm cluster labels keep explore links and 🔍 markers |
| --plain diagram legibility (plain labels + legend coexistence) | PLAIN-03 | Aesthetic judgment | `c4drill --plain` a styled fixture → open SVG → labels readable, no HTML artifacts |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 30s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
