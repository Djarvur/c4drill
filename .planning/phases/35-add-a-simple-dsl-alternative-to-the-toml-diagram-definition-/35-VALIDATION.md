---
phase: 35
slug: add-a-simple-dsl-alternative-to-the-toml-diagram-definition
status: draft
nyquist_compliant: false
wave_0_complete: false
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
| 35-01-xx | 01 | 1 | D-01..D-19 (grammar/parser) | — | N/A | unit | `go test ./internal/c4d/...` | ❌ W0 | ⬜ pending |
| 35-02-xx | 02 | 1 | D-21/D-22 (round-trip converters) | — | N/A | integration | `go test ./internal/convert/...` | ❌ W0 | ⬜ pending |
| 35-03-xx | 03 | 1 | D-16/D-17 (nested + recursive use) | — | N/A | unit | `go test ./internal/template/...` | ✅ | ⬜ pending |
| 35-04-xx | 04 | 2 | D-28/D-25 (convert CLI) | — | N/A | CLI | `go test ./cmd/c4drill/...` | ✅ | ⬜ pending |
| 35-05-xx | 05 | 2 | D-31/D-32 (fmt CLI) | — | N/A | CLI | `go test ./cmd/c4drill/...` | ✅ | ⬜ pending |
| 35-06-xx | 06 | 3 | D-35 (docs + examples) | — | N/A | fixture | `go test ./...` (fixtures render) | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

Task IDs are indicative — the planner assigns final IDs; every task must map to one of the verification dimensions in `35-RESEARCH.md` §8.

---

## Wave 0 Requirements

- [ ] `internal/c4d/` package skeleton + `c4d.peg` stub with one passing rule (pigeon generates, build passes)
- [ ] Round-trip normalizer package skeleton (`internal/testutil/canonsrc` or similar) — TOML + C4D normalization stubs
- [ ] Edge-case fixture set added under testdata (external types, linkFrom, rank=equal, nested use, template-body use)

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
