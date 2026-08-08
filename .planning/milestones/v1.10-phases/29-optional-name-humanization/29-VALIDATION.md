---
phase: 29
slug: optional-name-humanization
status: draft
nyquist_compliant: false
wave_0_complete: false
created: 2026-08-08
---

# Phase 29 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing + testify v1.11.1 |
| **Config file** | none (Go convention) |
| **Quick run command** | `go test ./internal/model/... ./internal/parser/...` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~2 seconds (quick), ~10 seconds (full) |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/model/... ./internal/parser/...`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green
- **Max feedback latency:** ~10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 29-01-01 | 01 | 1 | ERGO-04 | — | N/A (pure string fn) | unit | `go test ./internal/model/... -run TestHumanize` | ❌ W0 (new humanize_test.go) | ⬜ pending |
| 29-01-02 | 01 | 1 | ERGO-03 | — | N/A | unit+integration | `go test ./internal/parser/... -run OmittedName` | ❌ W0 (extend parser_test.go) | ⬜ pending |
| 29-01-03 | 01 | 1 | ERGO-05 | — | N/A (no-overwrite guard) | regression | `go test ./internal/parser/... -run ExplicitNameWins` | ❌ W0 | ⬜ pending |
| 29-01-04 | 01 | 1 | ERGO-03/05 | — | N/A | regression | `go test ./...` (corpus: existing fixtures unchanged) | ✅ (fixtures exist) | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/model/humanize.go` + `internal/model/humanize_test.go` — table-driven test encoding the ERGO-04 reference table (9 rows incl. `gRPC`→"Grpc", `IDPToken`→"Idp Token")
- [ ] `internal/parser/parser_test.go` — extend with `TestParseOmittedName*` (top-level + nested subunit) and `TestParseExplicitNameWins` (backward-compat)
- [ ] optional `testdata/optional_name.toml` fixture for the integration test
- [ ] testify framework — already installed (v1.11.1 in go.mod), no action

*Framework already present — Wave 0 is new test files only, no install step.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Rendered diagram shows humanized name in node label | ERGO-03 | End-to-end visual smoke; automated parser test already proves Name is populated | After execute, run `c4drill testdata/optional_name.toml -f svg` on a fixture with an omitted-name unit and eyeball that the node label shows "Local IDP" not empty |

*Primary verification is automated; this is a one-off smoke.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < ~10s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
