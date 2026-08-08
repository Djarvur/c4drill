---
phase: 32
slug: include-directive-multi-file
status: draft
nyquist_compliant: true
wave_0_complete: false
created: 2026-08-08
---

# Phase 32 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go testing + testify (assert/require) |
| **Config file** | none (standard `go test`) |
| **Quick run command** | `go test ./internal/parser/ ./internal/include/` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~5 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./internal/parser/ ./internal/include/`
- **After every plan wave:** Run `go test ./...`
- **Before `/gsd:verify-work`:** Full suite must be green + `golangci-lint run ./...` clean
- **Max feedback latency:** 10 seconds

---

## Per-Task Verification Map

> Task IDs are provisional against the expected 2-plan structure (01 = parser extraction `Model.Includes`; 02 = `internal/include` Resolve+merge + pipeline wiring). The planner finalizes exact task numbers; the Requirement → Test mapping is load-bearing either way.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 32-01-01 | 01 | 1 | INC-01 (parse side) | — | `[[include]]` array extracts into `Model.Includes` (Path/Once), never into `UnitOrder`/`Units` | unit | `go test ./internal/parser/ -run TestParseIncludesExtracted -x` | ❌ W0 | ⬜ pending |
| 32-01-02 | 01 | 1 | INC-01 | — | model with no `[[include]]` leaves `Model.Includes` nil/empty (no regression) | unit | `go test ./internal/parser/ -run TestParseNoIncludes -x` | ❌ W0 | ⬜ pending |
| 32-02-01 | 02 | 2 | INC-01 | — | entry + 1 include merges into one model (Units union, UnitOrder append) | unit | `go test ./internal/include/ -run TestResolveTwoFilesMerge -x` | ❌ W0 | ⬜ pending |
| 32-02-02 | 02 | 2 | INC-02 | IN-5 | paths resolve relative to including file's dir; cd-proof | unit | `go test ./internal/include/ -run TestResolveRelativePathIndependentOfCwd -x` | ❌ W0 | ⬜ pending |
| 32-02-03 | 02 | 2 | INC-03 | — | transitive includes resolve recursively (A→B→C) | unit | `go test ./internal/include/ -run TestResolveTransitive -x` | ❌ W0 | ⬜ pending |
| 32-02-04 | 02 | 2 | INC-04 | — | self-include + mutual (A↔B) → fatal *ParseError naming the cycle | unit | `go test ./internal/include/ -run TestResolveCycleFatal -x` | ❌ W0 | ⬜ pending |
| 32-02-05 | 02 | 2 | INC-05 | IN-3 | diamond (A→B→D, A→C→D) NOT flagged as cycle | unit | `go test ./internal/include/ -run TestResolveDiamondNotCycle -x` | ❌ W0 | ⬜ pending |
| 32-02-06 | 02 | 2 | INC-05/D-11 | IN-3 | cross-file dup unit path (two different files define same path) → hard error naming both files | unit | `go test ./internal/include/ -run TestMergeDuplicateUnitPathError -x` | ❌ W0 | ⬜ pending |
| 32-02-07 | 02 | 2 | INC-06 | — | `once=true` skips re-inclusion of already-included file | unit | `go test ./internal/include/ -run TestResolveOnceDedup -x` | ❌ W0 | ⬜ pending |
| 32-02-08 | 02 | 2 | INC-07 | — | flat merge (no namespacing); same path in two different files after dedup → hard error | unit | `go test ./internal/include/ -run TestMergeFlatNoNamespacing -x` | ❌ W0 | ⬜ pending |
| 32-02-09 | 02 | 2 | INC-08 | — | properties root-wins; included non-zero field conflicting with entry → hard error | unit | `go test ./internal/include/ -run TestMergePropertiesConflictError -x` | ❌ W0 | ⬜ pending |
| 32-02-10 | 02 | 2 | INC-09 | — | UnitOrder = entry units first, then each include's units appended in directive order | unit | `go test ./internal/include/ -run TestMergeUnitOrderAppend -x` | ❌ W0 | ⬜ pending |
| 32-02-11 | 02 | 2 | INC-10/D-12 | — | missing include → fatal *ParseError naming referenced path + including file | unit | `go test ./internal/include/ -run TestResolveMissingIncludeError -x` | ❌ W0 | ⬜ pending |
| 32-02-12 | 02 | 2 | D-10 | — | included file adds subunits to entry-defined parent (cross-file subunit merge) | unit | `go test ./internal/include/ -run TestMergeCrossFileSubunits -x` | ❌ W0 | ⬜ pending |
| 32-02-13 | 02 | 2 | D-11 | IN-3 | same-file diamond (D reached via 2 paths) auto-dedup, no error | unit | `go test ./internal/include/ -run TestResolveSameFileDiamondAutoDedup -x` | ❌ W0 | ⬜ pending |
| 32-02-14 | 02 | 2 | XC-02 | — | templates defined in included file flow into merged `Model.Templates` (visible to `[[use]]`) | unit | `go test ./internal/include/ -run TestMergeCarriesTemplates -x` | ❌ W0 (gated on Phase 31) | ⬜ pending |
| 32-02-15 | 02 | 2 | XC-01/XC-05 | — | include.Resolve runs BEFORE Validate; merged model passes validate + renders | integration | `go test ./cmd/c4drill/ -run TestPipelineIncludeBeforeValidate -x` | ❌ W0 | ⬜ pending |
| 32-02-16 | 02 | 2 | INC-01/XC-05 | DI-1 | multi-file model renders same (canonicalDOT) as equivalent single-file model | integration/golden | `go test ./... -run TestMultiFileGoldenCanonicalDOT -x` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/include/include_test.go` — INC-01..10, D-10, D-11 same-file dedup, XC-02 template carry. NEW package + tests.
- [ ] `internal/include/testdata/` — multi-file fixtures: `main.toml`, `auth.toml`, `templates.toml`, `cycle_a.toml`/`cycle_b.toml`, `diamond_*` (4 files), `missing_ref.toml`, `same_file_diamond_*`, `once_*`. NEW fixtures.
- [ ] `internal/parser/parser_test.go` (extend) — `TestParseIncludesExtracted`, `TestParseNoIncludes`. EXISTING file, add cases (consume the Phase 31 skip).
- [ ] `cmd/c4drill/` pipeline test (extend or add) — INC-01/XC-05 end-to-end: include.Resolve before Validate; multi-file golden.
- [ ] Reuse the existing canonicalDOT order-insensitive comparator helper (STATE.md DI-1) for the golden test — verify it exists in `internal/graph` or test helpers and reuse; do NOT reinvent.

*Framework install: none — Go testing + testify already present.*

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| (none) | — | — | — |

*All phase behaviors have automated verification. Rendered diagram visual inspection is covered indirectly by the canonicalDOT order-insensitive golden pipeline (DI-1) used in the integration test.*

---

## Threat Model

> ASVS L1, `security_block_on: high`. Phase 32 is a local-file TOML composition feature for a single-shot local CLI; input files are author-controlled, not attacker-controlled. No network, no auth, no untrusted-input path.

| Threat | STRIDE | Severity | Mitigation | Residual |
|--------|--------|----------|------------|----------|
| Path traversal via crafted include path | Information Disclosure | low (author already has FS access as invoking user) | `filepath.Clean`+`Abs` canonicalize for cycle/dedup correctness (NOT a security boundary — document this) | Accepted: local tool, no privilege boundary |
| Zip-bomb / diamond-explosion (acyclic, huge) | Denial of Service | medium | `maxIncludeDepth = 100` cap + visited-set ensures each file parsed at most once (O(distinct files)) | Low |
| Symlink-based cycle evasion | Tampering/DoS | low | `filepath.Abs` does not resolve symlinks; document as accepted v1.10 limitation | Accepted: author-controlled; `filepath.EvalSymlinks` is a future hardening path |

**Verdict:** No high-severity threats. DoS mitigations are defense-in-depth. Does not trigger `security_block_on: high`.

---
