---
phase: 31
slug: template-expansion
status: draft
nyquist_compliant: true
wave_0_complete: false
created: 2026-08-08
---

# Phase 31 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing + testify (assert/require) |
| **Config file** | none (standard `go test`) |
| **Quick run command** | `go test ./internal/parser/ ./internal/model/ ./internal/template/` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/parser/ ./internal/model/ ./internal/template/`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green + `golangci-lint run` clean
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 31-01-01 | 01 | 1 | TMPL-09 | — | reserved tables never enter unitOrder (no phantom units) | unit | `go test ./internal/parser/ -run TestParseReservedTablesSkipped -x` | ❌ W0 | ⬜ pending |
| 31-01-02 | 01 | 1 | TMPL-01 | — | `[template.<name>]` routes into Model.Templates | unit | `go test ./internal/parser/ -run TestParseTemplateTable -x` | ❌ W0 | ⬜ pending |
| 31-01-03 | 01 | 1 | TMPL-02 | — | `[[use]]` routes into Model.Instantiations in document order | unit | `go test ./internal/parser/ -run TestParseUseArray -x` | ❌ W0 | ⬜ pending |
| 31-01-04 | 01 | 1 | TMPL-09 | — | `[[use]]` textually before `[template.*]` parses (forward ref) | unit | `go test ./internal/parser/ -run TestParseUseBeforeTemplate -x` | ❌ W0 | ⬜ pending |
| 31-02-01 | 02 | 2 | TMPL-08 | — | Clone preserves unexported Link.Mirror | unit | `go test ./internal/model/ -run TestClonePreservesMirror -x` | ❌ W0 | ⬜ pending |
| 31-02-02 | 02 | 2 | TMPL-08 | — | Clone recurses into Subunits map | unit | `go test ./internal/model/ -run TestCloneRecursesSubunits -x` | ❌ W0 | ⬜ pending |
| 31-02-03 | 02 | 2 | TMPL-08 | HS-1 | 3× instantiate → idempotent re-expand + disjoint LinksFrom post-validate | unit | `go test ./internal/template/ -run TestExpandThreeInstantiationsHS1 -x` | ❌ W0 | ⬜ pending |
| 31-02-04 | 02 | 2 | TMPL-03 | — | `${param}` substitutes into all string fields + link fields | unit | `go test ./internal/template/ -run TestExpandSubstitution -x` | ❌ W0 | ⬜ pending |
| 31-02-05 | 02 | 2 | TMPL-04 | — | template with subunit subtree expands whole subtree | unit | `go test ./internal/template/ -run TestExpandSubtree -x` | ❌ W0 | ⬜ pending |
| 31-02-06 | 02 | 2 | TMPL-06 | — | missing param → hard error naming template+param+site | unit | `go test ./internal/template/ -run TestExpandMissingParamNames -x` | ❌ W0 | ⬜ pending |
| 31-02-07 | 02 | 2 | TMPL-07 | — | duplicate path across uses/hand-authored → hard error naming both | unit | `go test ./internal/template/ -run TestExpandDuplicatePath -x` | ❌ W0 | ⬜ pending |
| 31-02-08 | 02 | 2 | XC-03 | — | expanded unit's parent = `[[use]]` parent (instantiation site) | unit | `go test ./internal/template/ -run TestExpandParentPlacement -x` | ❌ W0 | ⬜ pending |
| 31-02-09 | 02 | 2 | TMPL-10 | — | `reference` field substitutes params (conditional on Phase 28) | unit | `go test ./internal/template/ -run TestExpandReferenceField -x` | ❌ W0 | ⬜ pending |
| 31-02-10 | 02 | 2 | TMPL-05/XC-04 | — | Expand runs before Validate; expanded units pass validate + appear in views | integration | `go test ./cmd/c4drill/ -run TestPipelineExpandBeforeValidate -x` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/model/unit_test.go` — Clone tests (Mirror preservation, Subunits recursion). NEW file (no model tests exist today).
- [ ] `internal/template/expand_test.go` — TMPL-02..10, XC-03. NEW package + tests.
- [ ] `internal/parser/parser_test.go` (extend) — reserved-table skip, extraction, forward-ref cases. EXISTING file, add cases.
- [ ] `cmd/c4drill/` pipeline test (extend or add) — TMPL-05/XC-04 end-to-end ordering.
- [ ] `testdata/template_*.toml` fixtures — basic, subtree, 3x-instantiate, missing-param, duplicate-path, forward-ref. NEW fixtures.

*Framework install: none — Go testing + testify already present.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| (none) | — | — | — |

*All phase behaviors have automated verification. Rendered diagram visual inspection is covered indirectly by the existing canonicalDOT golden pipeline (DI-1) used in integration tests.*

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 10s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
