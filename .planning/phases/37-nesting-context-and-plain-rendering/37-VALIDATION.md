---
phase: 37
slug: nesting-context-and-plain-rendering
status: approved
nyquist_compliant: true
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
| 37-01-01..03 | 01 | 1 | CTX-03 | T-37-01 | N/A | unit (graph, TDD RED→GREEN) | `go test ./internal/graph/ -run 'TestBuildGraph_ExpandedClusterRendersNestedSubClusters' -v` | ❌ W0 (new test in builder_test.go) | ⬜ pending |
| 37-02-01..03 | 02 | 2 | CTX-02 | T-37-03 | N/A | unit (view+graph, TDD) | `go test ./internal/view/ -run 'DeepLink|DeepCrossLink' -v && go test ./internal/graph/ -run 'TestBuildGraph_DeepLinkTargetRendersInsideUnfoldedCluster' -v` | ❌ W0 (new tests in scope_test.go + builder_test.go) | ⬜ pending |
| 37-03-01..03 | 03 | 3 | PLAIN-01/02 | T-37-05 | N/A | unit (graph, TDD) | `go test ./internal/graph/ -run 'Plain' -v` | ❌ W0 (new tests in builder_test.go) | ⬜ pending |
| 37-04-01..03 | 04 | 4 | PLAIN-03/04 | T-37-07 | N/A | unit + E2E golden (TDD + canonical) | `go test ./internal/render/ -run 'Plain' -v && go test ./cmd/c4drill/ -run 'Plain' -v` | ❌ W0 (converter_test.go, root_test.go, plain.dot goldens) | ⬜ pending |
| 37-05-01 | 05 | 5 | CTX-01 | T-37-10 | N/A | invariant unit + E2E goldens | `go test ./internal/view/ -run 'TestViewAncestorChainInvariant' -v` | ❌ W0 (new test in scope_test.go) | ⬜ pending |
| 37-05-02 | 05 | 5 | BC-01 | T-37-09 | N/A | canonicalDOT goldens | `go test ./...` | ✅ (existing goldens re-baselined once, here) | ⬜ pending |
| 37-06-01..03 | 06 | 5 | DOC-01..03 | T-37-11 | N/A | CI parity + render | `diff -r skill plugins/c4drill/skills/c4drill-toml && diff -r skill plugins/c4drill/opencode/skills/c4drill-toml` + example renders | ✅ (workflow exists) | ⬜ pending |
| 37-07-01..02 | 07 | 6 | REL-01 | T-37-13 | N/A | release tag | `git tag v1.21.0` + `gh release view v1.21.0` | ✅ (release.yml exists) | ⬜ pending |

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

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 30s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-08-30 (plan-checker iteration 2 — VERIFICATION PASSED)
