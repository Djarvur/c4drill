---
phase: 03-compatibility-validation
verified: 2026-08-06T12:23:24Z
status: passed
score: 12/12 must-haves verified
overrides_applied: 0
---

# Phase 3: Compatibility & Validation Verification Report

**Phase Goal:** Ensure backward compatibility and validate with real TOML files.
**Verified:** 2026-08-06T12:23:24Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| #   | Truth   | Status     | Evidence       |
| --- | ------- | ---------- | -------------- |
| 1   | Fresh checkout runs the full suite without the gitignored private fixture — the sanitized public fixture (multilevel.toml) stands in for the saira structure with generic names (D-01) | ✓ VERIFIED | `multilevel.toml` (499 lines) committed and tracked; zero private identifiers (grep for linuxSystem\|keycloak\|webUser\|sshUser\|adminUser = 0 matches); `cyp-auth-infra/` gitignored with 0 tracked files; repo-wide gate `grep -rn "cyp-auth-infra" --include="*.go" .` returns nothing (exit 1); render tests use the synthetic model (`loadCYPAuthInfraModel` is a thin wrapper over `createSyntheticCYPModel`, no file probing) |
| 2   | The committed golden baseline captures the v1.7-compatible expanded DOT: minlen=3 and penwidth=2 edge attributes present, all units at all levels included (D-02) | ✓ VERIFIED | `multilevel.expanded.dot` (1041 lines, XDOT `digraph "" {` + rankdir=TB) committed and tracked; client→externalSys edge block at :1030-1041 with `minlen=3` + `penwidth=2` + `style=solid`; 4 minlen=2 edge blocks; 56 penwidth=2 occurrences; deep node `"mainSystem.sshAuth.systemd.logind"` present (2 occurrences); zero `[+]` |
| 3   | The fixture carries NO expansion hints (neither properties.expanded nor per-unit expanded) | ✓ VERIFIED | `grep -c "expanded" multilevel.toml` = 0 (exit 1); `valid.toml` also contains zero "expanded" occurrences (COMPAT-01 fixture basis) |
| 4   | The fixture preserves the structural skeleton: 5 top-level units, 4-level nesting, cross-level length links | ✓ VERIFIED | Exactly 5 top-level units (actorA/actorB/actorC personExternal, externalSys systemExternal, mainSystem system); deep 4-level path mainSystem.sshAuth.systemd.logind; cross-level link `{peer = "externalSys", length=3}` at :398; four length=2 links preserved (:175, :259, :260, :270); inline `link = [{...}]` array syntax |
| 5   | TestBuildExpandedGraphRealToml passes against the public fixture with sanitized names — client → externalSys edge keeps MinLen=3 and DOT contains minlen=3 | ✓ VERIFIED | Test at builder_test.go:803 repointed to `../../cmd/c4drill/testdata/multilevel.toml`; asserts unit path `mainSystem.storages.externalStorage.client`, peer `externalSys`, `edge.MinLen == expectedLength`, DOT contains `minlen=3`; PASSES |
| 6   | TestBuildExpandedGraphBaselineDOT compares against the committed golden baseline via canonicalDOT semantic comparison (sorted, geometry-stripped) — with WR-01 off-by-one and WR-02 terminator fixes present | ✓ VERIFIED | Test at builder_test.go:1217 keeps penwidth InDelta 2.0 loop and ends with `require.Equal(canonicalDOT(expected), canonicalDOT(dotData))`; canonicalDOT parses statements, strips geometry attrs (bb/pos/lp/lheight/lwidth/height/width) via isGeometryAttr, sorts attrs + statements recursively; WR-01 fix `normalizeDOTAttrs(text[open+1:])` (full block, no len-1 truncation); WR-02 helpers scanDOTValueEnd/findDOTAttrTerminator/findDOTBlockOpen are quote- and HTML-aware; PASSES 5/5 repeated runs (deterministic) |
| 7   | D-04 locking test: with properties.expanded poisoned, the expanded view still contains ALL units and no `[+]` indicators anywhere | ✓ VERIFIED | TestBuildExpandedGraphIgnoresPropertiesExpanded (builder_test.go:2314) poisons `m.Properties.Expanded = []string{"mainSystem"}` BEFORE `view.GenerateExpandedView`, asserts `require.Len(v.Units, total)` vs recursive walk (countModelUnits), deep leaves render as nodes, zero labels contain `[+]`; PASSES |
| 8   | COMPAT-01: valid.toml (no properties.expanded) generates a correct all-collapsed C1 — no clusters, no app.api node, `[+]` present (no absence assertion on valid/app.dot) | ✓ VERIFIED | TestCompat01_ValidTomlAllCollapsed (root_test.go:607) asserts `user\t[`, `app\t[`, `Application [+]`, no `subgraph cluster_`, no `app.api`; contains NO os.IsNotExist assertion on valid/app.dot; PASSES |
| 9   | The multilevel fixture produces a 5-node C1 plus C2/C3 sub-diagram files end-to-end via the CLI (ROADMAP success criterion 4) | ✓ VERIFIED | TestCompat02_MultilevelFixtureFiveNodeC1 (root_test.go:638) asserts all 5 top-level node lines, no clusters, no `mainSystem.sshAuth` deep path, `Main System [+]`, C2 `multilevel/mainSystem.dot` and C3 `multilevel/mainSystem/sshAuth.dot` exist; PASSES; independent CLI run `go run ./cmd/c4drill cmd/c4drill/testdata/multilevel.toml --output /tmp/... --format dot` exits 0 and reproduces the same C1 (5 top-level nodes, 0 clusters) + C2/C3 files |
| 10  | No test references ../../cyp-auth-infra/ after this phase — CI enforceability gate | ✓ VERIFIED | `grep -rn "cyp-auth-infra" --include="*.go" .` returns nothing (exit 1); `git check-ignore cyp-auth-infra/` confirms gitignored, 0 tracked files |
| 11  | Production files (scope.go, builder.go) byte-identical to phase base | ✓ VERIFIED | `git log 346cafc..HEAD -- internal/view/scope.go internal/graph/builder.go internal/render/converter.go` = no commits; `git diff 346cafc HEAD` on those files = empty |
| 12  | All existing tests pass unmodified (except fixture-path updates); full suite green | ✓ VERIFIED | `go test -count=1 ./...` run by verifier: all 7 packages ok (cmd/c4drill, internal/graph, internal/output, internal/parser, internal/render, internal/validator, internal/view); phase tests re-run individually all PASS |

**Score:** 12/12 truths verified

### Required Artifacts

| Artifact | Expected    | Status | Details |
| -------- | ----------- | ------ | ------- |
| `cmd/c4drill/testdata/multilevel.toml` | Sanitized public fixture: 5 top-level units, 4-level nesting, cross-level length links, validator-clean | ✓ VERIFIED | 499 lines, committed (bb536a6); CLI parse+validate exits 0; sanitization gate clean; zero expansion hints; contains `mainSystem.storages.externalStorage.client` and length=3 link |
| `cmd/c4drill/testdata/multilevel.expanded.dot` | Committed golden baseline for COMPAT-02 (DOT-level equivalence contract) | ✓ VERIFIED | 1041 lines, committed (67f1098); XDOT header, rankdir=TB; minlen=3 + penwidth=2 edge block; deterministic semantic baseline (canonicalDOT comparison stable 5/5) |
| `internal/graph/builder_test.go` | Repointed RealToml + BaselineDOT (golden comparison) + D-04 regression + canonicalDOT helper | ✓ VERIFIED | All present with correct assertions; WR-01/WR-02 fixes at :1376/:1384-1452; 4 WR regression tests present and passing; zero cyp-auth-infra references |
| `cmd/c4drill/root_test.go` | COMPAT-01 regression + multilevel 5-node C1 / C2/C3 integration | ✓ VERIFIED | TestCompat01_ValidTomlAllCollapsed (:607), TestCompat02_MultilevelFixtureFiveNodeC1 (:638); both pass; no absence assertion on valid/app.dot; //nolint:paralleltest on both |
| `internal/render/expanded_internal_test.go` | Private-fixture probing removed (deterministic synthetic model) | ✓ VERIFIED | loadCYPAuthInfraModel (:18-22) is a thin wrapper over createSyntheticCYPModel; comments reworded without the gated literal; 4 render tests pass |

### Key Link Verification

| From | To  | Via | Status | Details |
| ---- | --- | --- | ------ | ------- |
| builder_test.go TestBuildExpandedGraphBaselineDOT | cmd/c4drill/testdata/multilevel.expanded.dot | os.ReadFile + require.Equal on canonicalDOT forms | ✓ WIRED | builder_test.go:1238-1246; passes 5/5 runs |
| cmd/c4drill/root_test.go TestCompat01_ValidTomlAllCollapsed | cmd/c4drill/testdata/valid.toml | filepath.Join("testdata", "valid.toml") CLI run | ✓ WIRED | root_test.go:611-615; executes and passes |
| internal/graph/builder_test.go TestBuildExpandedGraphIgnoresPropertiesExpanded | internal/view/scope.go GenerateExpandedView | poison m.Properties.Expanded before view generation (READ-ONLY consumption) | ✓ WIRED | builder_test.go:2324-2328; require.Len(v.Units, total) proves no consultation |
| cmd/c4drill/testdata/multilevel.toml | parser/validator pipeline | CLI parse+validate run | ✓ WIRED | `go run ./cmd/c4drill ... --format dot` exits 0 (verifier re-ran) |
| cmd/c4drill/testdata/multilevel.toml | cmd/c4drill/testdata/multilevel.expanded.dot | exact pipeline ParseFile → Validate → GenerateExpandedView → BuildExpandedGraph → RenderDOT | ✓ WIRED | Golden naming per writer.go:66 ({basename}.expanded.{format}); semantic reproduction confirmed by the passing golden comparison test |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| -------- | ------------- | ------ | ------------------ | ------ |
| TestBuildExpandedGraphRealToml | m (parsed model) | `parser.ParseFile("../../cmd/c4drill/testdata/multilevel.toml")` | ✓ (real committed fixture; validator require.Empty passes; MinLen flows from TOML length attribute to edge to rendered minlen) | ✓ FLOWING |
| TestBuildExpandedGraphBaselineDOT | dotData vs expected | in-process pipeline vs committed golden file | ✓ (canonical forms equal; golden itself generated by the same pipeline) | ✓ FLOWING |
| TestCompat01/TestCompat02 | C1 DOT + sub-diagram files | real CLI run on committed fixtures, output in t.TempDir() | ✓ (assertions on real rendered content; C2/C3 files exist on disk) | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| -------- | ------- | ------ | ------ |
| Full suite green | `go test -count=1 ./...` | all 7 packages ok | ✓ PASS |
| RealToml + BaselineDOT + D-04 + WR regressions | `go test ./internal/graph/ -run 'TestBuildExpandedGraphRealToml\|TestBuildExpandedGraphBaselineDOT\|TestBuildExpandedGraphIgnoresPropertiesExpanded\|TestCanonicalDOT' -count=1 -v` | 7/7 PASS | ✓ PASS |
| COMPAT regressions | `go test ./cmd/c4drill/ -run 'TestCompat' -count=1 -v` | 2/2 PASS | ✓ PASS |
| Render tests deterministic | `go test ./internal/render/ -count=1` | ok | ✓ PASS |
| Golden comparison stability | `go test ./internal/graph/ -run 'TestBuildExpandedGraphBaselineDOT' -count=5` | ok (5/5) | ✓ PASS |
| Fixture validator-clean + C1/C2/C3 | `go run ./cmd/c4drill cmd/c4drill/testdata/multilevel.toml --output /tmp/multilevel-verify --format dot` | exit 0; C1 = 5 top-level nodes, 0 clusters, "Main System [+]"; C2/C3 files generated | ✓ PASS |
| No cyp-auth-infra in .go files | `grep -rn "cyp-auth-infra" --include="*.go" .` | no matches (exit 1) | ✓ PASS |
| Production scope guard | `git log 346cafc..HEAD -- internal/view/scope.go internal/graph/builder.go` + `git diff 346cafc HEAD` | no commits, no diff | ✓ PASS |

### Probe Execution

No probes declared by the phase plans; the phase's verification gates were CLI runs and targeted `go test` invocations, all re-run by the verifier above. Step 7c: not applicable (no probe-*.sh files in scripts/, none referenced in PLAN/SUMMARY).

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| ----------- | ---------- | ----------- | ------ | -------- |
| COMPAT-01 | 03-02-PLAN | Existing TOML files without `properties.expanded` generate correct C1 with all units collapsed | ✓ SATISFIED | TestCompat01_ValidTomlAllCollapsed passes (no clusters, no app.api, `[+]` indicators); D-03 OR semantics intentionally not duplicated (locked in Phase 2 by TestGenerateC1View_IsExpandedOrSemantics); REQUIREMENTS.md status Complete |
| COMPAT-02 | 03-01-PLAN, 03-02-PLAN | `--expanded` flag continues to produce single all-nested diagram unchanged | ✓ SATISFIED | Committed golden baseline + canonicalDOT semantic comparison (DOT-level equivalence per D-02) passes deterministically; TestBuildExpandedGraphRealToml + D-04 locking test enforce the expanded-view contract; REQUIREMENTS.md status Complete |

Both requirement IDs claimed by plans are accounted for; no orphaned requirements (REQUIREMENTS.md maps only COMPAT-01/COMPAT-02 to Phase 3, both claimed).

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| ---- | ---- | ------- | -------- | ------ |
| — | — | none | — | No TBD/FIXME/XXX/TODO/HACK markers, no placeholder strings, no empty implementations, no console.log-only handlers in any phase-modified file |

### Review Findings Resolution (WR-01 / WR-02)

| Finding | Status | Evidence |
| ------- | ------ | -------- |
| WR-01: canonicalDOT off-by-one drops last char of last attribute | ✓ RESOLVED | parseDOTAttrStatement:1376 passes full block `text[open+1:]`; regression tests TestCanonicalDOTPreservesLastAttribute (canonical form contains `style=solid`) and TestCanonicalDOTFinalAttributeDriftDetected (`penwidth=1` vs `penwidth=2` canonicalize differently) both PASS (commits 94da191, 48f4632) |
| WR-02: statement terminators not value-aware | ✓ RESOLVED | scanDOTValueEnd (:1384) handles quoted strings (backslash escapes) and nested HTML `<...>`; findDOTAttrTerminator (:1422) and findDOTBlockOpen (:1441) use it; regression tests TestCanonicalDOTQuotedValuesDoNotTruncate (`description="SSH [session]; admin"`, `label="uses {braces}"` survive) and TestCanonicalDOTHTMLLabelDoesNotTruncate both PASS |

### Human Verification Required

None. This is a test-infrastructure phase with no production code changes; all success criteria are programmatically verifiable. The DI-1 deviation (byte-exact → canonicalDOT order-insensitive semantic comparison) is a documented, mandated design change (STATE.md, deferred-items.md DI-1) already accepted by the executor and reviewer; the comparison is deterministic across repeated runs.

### Gaps Summary

No gaps. All 12 must-have truths verified against the actual codebase: both committed test-data assets exist and are validator/golden-clean, all six phase tests pass (independently re-run), the WR-01/WR-02 review fixes are present with passing regression tests, the repo-wide cyp-auth-infra gate is clean, production files are byte-identical to the phase base, and the full suite (`go test -count=1 ./...`) is green. ROADMAP success criteria 1-4 all satisfied at the DOT-level equivalence contract (D-02).

---

_Verified: 2026-08-06T12:23:24Z_
_Verifier: Claude (gsd-verifier)_
