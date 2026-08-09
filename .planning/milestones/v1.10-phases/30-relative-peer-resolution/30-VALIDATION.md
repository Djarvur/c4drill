---
phase: 30
slug: relative-peer-resolution
status: draft
nyquist_compliant: true
wave_0_complete: false
created: 2026-08-08
---

# Phase 30 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing + testify (assert/require) |
| **Config file** | none (standard `go test`) |
| **Quick run command** | `go test ./internal/peer/` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~3 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/peer/`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green + `go vet ./...` clean
- **Max feedback latency:** 5 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 30-01-01 | 01 | 1 | ERGO-01 | — | sibling match rewrites bare peer to parent's child path | unit | `go test ./internal/peer/ -run TestResolveSibling -x` | ❌ W0 | ⬜ pending |
| 30-01-02 | 01 | 1 | ERGO-01 | — | aunt match via walk-up (grandparent's child) | unit | `go test ./internal/peer/ -run TestResolveWalkUpAunt -x` | ❌ W0 | ⬜ pending |
| 30-01-03 | 01 | 1 | ERGO-01 | — | root match: bare peer resolves to top-level unit from any depth | unit | `go test ./internal/peer/ -run TestResolveRoot -x` | ❌ W0 | ⬜ pending |
| 30-01-04 | 01 | 1 | ERGO-01 | — | nearest-first: nearer match wins over deeper match (no error) | unit | `go test ./internal/peer/ -run TestResolveNearestFirst -x` | ❌ W0 | ⬜ pending |
| 30-01-05 | 01 | 1 | ERGO-02 | — | dotted peer left untouched (absolute, no walk-up) | unit | `go test ./internal/peer/ -run TestResolveDottedUntouched -x` | ❌ W0 | ⬜ pending |
| 30-01-06 | 01 | 1 | ERGO-02 | — | unresolvable bare peer → hard error naming peer + host unit | unit | `go test ./internal/peer/ -run TestResolveUnresolvableError -x` | ❌ W0 | ⬜ pending |
| 30-01-07 | 01 | 1 | ERGO-01/02 | — | authored `LinksFrom` peers also rewritten (not just `Links`) | unit | `go test ./internal/peer/ -run TestResolveLinksFrom -x` | ❌ W0 | ⬜ pending |
| 30-01-08 | 01 | 1 | ERGO-02 | — | parser-corpus byte-identical: testdata/{valid,links,nested}.toml peer sets unchanged post-Resolve | unit | `go test ./internal/peer/ -run TestResolveCorpusByteIdentical -x` | ❌ W0 | ⬜ pending |
| 30-02-01 | 02 | 1 | ERGO-01/02 | — | pipeline: Resolve runs after Parse, before Validate; validator sees only absolute peers | integration | `go test ./cmd/c4drill/ -run TestPipelineResolveBeforeValidate -x` | ❌ W0 | ⬜ pending |
| 30-02-02 | 02 | 1 | ERGO-02 | — | CLI: model with unresolvable peer exits non-zero with clear error (no panic) | integration | `go test ./cmd/c4drill/ -run TestCLIUnresolvablePeerExits -x` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/peer/resolve_test.go` — all ERGO-01/02 behaviors (TDD: RED first). NEW package + tests.
- [ ] `cmd/c4drill/testdata/peer_walkup.toml` — sibling + aunt + root + nearest-first cases. NEW fixture.
- [ ] `cmd/c4drill/testdata/peer_unresolvable.toml` — bare peer matching nothing. NEW fixture.
- [ ] `cmd/c4drill/` pipeline test (extend or add `root_test.go`) — pipeline ordering + CLI error path. NEW or extend.

*Framework install: none — Go testing + testify already present.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| (none) | — | — | — |

*All phase behaviors have automated verification. The walk-up traces in CONTEXT.md "Specific Ideas" are covered by unit tests asserting the exact resolved path; rendered diagram visual inspection is covered indirectly by the corpus byte-identical test (ERGO-02 acceptance) — if the peer set is unchanged, the rendered output is unchanged.*

---

## Notes for the Planner

- **Same-depth ambiguity (ROADMAP criterion 3) is structurally unreachable** under the walk-up model (`Subunits map[string]*Unit` has unique keys per parent — see RESEARCH.md Pitfall 3). Author the error branch defensively but do NOT block on a test that exercises it; it is dead code. The `TestResolve*` matrix above deliberately omits a same-depth-ambiguity case.
- **The corpus has bare peers in BOTH testdata directories** (root `testdata/`: `links.toml` has `webapp`/`user`; `cmd/c4drill/testdata/`: `valid.toml` `user`, `expanded.toml` `external`, `multilevel.toml` `externalSys`), each resolving to a top-level unit. The byte-identical test (30-01-08) asserts the `(source, peer)` SET is unchanged, NOT that no rewrite happened — the rewrites are identity rewrites. Plan 01's unit test uses the root `testdata/` corpus; Plan 02's integration test uses the `cmd/c4drill/testdata/` corpus. Both EXCLUDE the `invalid*.toml` error-path fixtures. See RESEARCH.md "Backward-compat corpus analysis" + Pitfall 2.

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 5s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
