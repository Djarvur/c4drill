---
phase: 35
slug: add-a-simple-dsl-alternative-to-the-toml-diagram-definition
status: planned
nyquist_compliant: true
wave_0_complete: true
created: 2026-08-14
---

# Phase 35 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (stdlib) + testify |
| **Config file** | none needed — existing repo layout (`.mise.toml` tasks) |
| **Quick run command** | `go test ./internal/c4d/... ./internal/parser/... 2>/dev/null || go test ./...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~30-60 seconds |

**Conventions:** render-invoking tests must NOT call `t.Parallel()` (WASM engine mutex — repo-wide rule, 91 existing annotations). New parser/converter tests CAN parallelize (no render involvement).

---

## Sampling Rate

- **After every task commit:** Run `go build ./... && go test ./...`
- **After every plan wave:** Run `go test ./...` (12/12 packages green)
- **Before `/gsd:verify-work`:** Full suite green + round-trip corpus green
- **Max feedback latency:** ~60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 35-01-1..3 | 01 | 1 | D-01..D-09, D-12, D-18, D-20 (toolchain + core grammar + errors) | T-35-01-* | MaxExpressions/DoS cap, pinned pigeon | unit (TDD) | `go test ./internal/c4d/...` | ❌ W0 | ⬜ pending |
| 35-02-1..2 | 02 | 1 | D-16, D-17 (nested use + recursive expansion) | T-35-02-* | cycle + depth cap | unit (TDD) | `go test ./internal/template/ ./internal/parser/` | ✅ | ⬜ pending |
| 35-03-1..2 | 03 | 2 | D-13, D-14, D-15, D-19 (composition grammar + reserved) | T-35-03-* | reserved-set pinning | unit (TDD) | `go test ./internal/c4d/...` | ✅ (35-01 creates) | ⬜ pending |
| 35-04-1..2 | 04 | 3 | D-23, D-33 (emitters + canonical order + compact style) | T-35-04-* | determinism, escaping | unit (TDD) | `go test ./internal/c4d/ -run TestEmit` | ✅ | ⬜ pending |
| 35-05-1..3 | 05 | 3 | D-02, D-10, D-11, D-21, D-26 (toModel + mixed includes) | T-35-05-* | ext-dispatch fail closed | unit+integration (TDD) | `go test ./internal/c4d/ ./internal/include/` | ✅ | ⬜ pending |
| 35-06-1..3 | 06 | 4 | D-22, D-26 (round-trip + render equivalence) | T-35-06-* | corpus walk hygiene | integration (TDD) | `go test ./internal/c4d/ -run 'TestRoundTrip|TestRenderEquivalence'` | ❌ W0 | ⬜ pending |
| 35-07-1..3 | 07 | 4 | D-24, D-25, D-27..D-30 (dispatch + convert CLI) | T-35-07-* | validate-first, cycle-safe walk | CLI (TDD) | `go test ./cmd/c4drill/ -run TestConvert` | ✅ | ⬜ pending |
| 35-08-1..3 | 08 | 5 | D-31, D-32 (fmt + tomlfmt) | T-35-08-* | semantic safety gate, walk filter | unit+CLI (TDD) | `go test ./internal/tomlfmt/ ./cmd/c4drill/ -run TestFmt` | ❌ W0 | ⬜ pending |
| 35-09-1..3 | 09 | 6 | D-34, D-35 (docs + twins + skill) | T-35-09-* | real flags only | fixture + grep gates | `go test ./...` + README/skill greps | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

Task IDs are indicative — the planner assigns final IDs; every task must map to one of the verification dimensions in `35-RESEARCH.md` §8.

---

## Wave 0 Requirements

- [x] `internal/c4d/` skeleton folded into Plan 35-01 Task 1 (toolchain + generating peg stub — first commit of the plan)
- [x] Round-trip normalizer (`internal/testutil/canonsrc`) folded into Plan 35-06 Task 1 (TDD: normalizer tests are the first artifact)
- [x] Edge-case fixture set folded into Plan 35-06 Task 2 (testdata/c4d/: external types, linkFrom, rank=equal, template+nested use, unicode)
- [x] tomlfmt package skeleton folded into Plan 35-08 Task 1 (probe test decides unstable-API vs fallback strategy)

Wave-0 coverage note: every plan is TDD-ordered — RED tests precede implementation within each task, satisfying the Nyquist requirement without separate scaffold commits.

*Existing infrastructure (go test, testify, canonical.Canonical DI-1, fixture corpus) covers the rest.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| README C4D section readability | D-35 | Prose quality | Human review of rendered README |
| Example `.c4d` twins ergonomics | D-35 | Authoring feel | Compare each twin against its `.toml` source for verbosity win |

*All functional behaviors have automated verification (round-trip, render-equivalence, fmt idempotency, CLI exit codes).*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
