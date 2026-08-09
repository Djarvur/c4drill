---
phase: 03-compatibility-validation
plan: 02
subsystem: testing
tags: [compat, golden-baseline, dot, regression, fixture, determinism, go-graphviz]
requires:
  - phase: 03-compatibility-validation
    provides: sanitized public multilevel fixture (multilevel.toml) + committed --expanded DOT golden baseline (multilevel.expanded.dot) from 03-01
  - phase: 01-fix-c1-view-scoping
    provides: AllExpanded exemption machinery (D-02/D-04) the locking tests protect
  - phase: 02-auto-generate-c2-c3-diagrams
    provides: C2/C3 auto-detect behavior the integration test exercises
provides:
  - repointed builder tests (RealToml + BaselineDOT) running on the committed public fixture — zero gitignored paths in test inputs (D-01, CI-enforceable)
  - order-insensitive semantic golden comparison (canonicalDOT) enforcing COMPAT-02 at DOT level (D-01/D-02) despite the render pipeline's map-order nondeterminism (DI-1)
  - D-04 locking test: --expanded mode immune to properties.expanded (poisoned fixture still yields ALL units, no [+] indicators)
  - COMPAT-01 regression: valid.toml renders all-collapsed C1 (no clusters, [+] indicator, no nested nodes)
  - ROADMAP SC4 integration test: multilevel fixture -> 5-node C1 + C2/C3 sub-diagrams end-to-end via the CLI
  - zero "cyp-auth-infra" references in any .go file; render tests deterministic via the synthetic model (private-fixture probing removed)
affects: [04-rendering (future), verify-work, milestone audit — COMPAT-01/COMPAT-02 now enforceable on fresh checkouts]
tech-stack:
  added: []
  patterns:
    - "order-insensitive golden comparison: canonicalDOT parses xdot into statements (kind/head/sorted attrs/recursively sorted children), strips layout geometry (bb/pos/lp/lheight/lwidth/height/width), and compares the normalized semantic form — the D-02 equivalence contract (DI-1)"
    - "locking test: poison m.Properties.Expanded BEFORE GenerateExpandedView and assert the view still contains every unit — passes on first run by construction (D-04)"
    - "synthetic-model render tests: loadCYPAuthInfraModel is a thin wrapper over createSyntheticCYPModel — deterministic on every machine (D-01)"
key-files:
  created: []
  modified:
    - internal/graph/builder_test.go
    - cmd/c4drill/root_test.go
    - internal/render/expanded_internal_test.go
key-decisions:
  - "Golden comparison is NOT byte-exact require.Equal (plan as written) — replaced with order-insensitive canonicalDOT semantic comparison per the DI-1 mandatory deviation (empirically: 5/5 pipeline runs pairwise differ byte-wise; sibling cluster/node order + layout geometry flip; semantic content stable)"
  - "canonicalDOT strips layout geometry (bb/pos/lp/lheight/lwidth/height/width) and sorts statements/attributes recursively — validated: 2 independent runs + committed golden all canonicalize identically, while C1 outputs (different semantics) still differ"
  - "DOT statement terminator is '];' not ';' — person-node HTML labels contain '&#x1F464;' entities whose semicolons would truncate naive scans"
  - "TestCompat01 does NOT assert absence of valid/app.dot — Phase 2 auto-detect generates the C2 sub-diagram (PATTERNS skeleton corrected in plan; verified empirically)"
patterns-established:
  - "canonicalDOT normalization is the reference pattern for any future DOT-level golden comparison in this repo (DI-1 workaround until the pinned go-graphviz fork is fixed/repinned — deferred, Rule 4)"
requirements-completed: [COMPAT-01, COMPAT-02]
duration: 8min
completed: 2026-08-06
---

# Phase 03 Plan 02: Compat Regressions on the Public Fixture + Order-Insensitive Golden Comparison Summary

**Repointed the two private-fixture builder tests onto the committed public fixture, replaced the golden baseline comparison with an order-insensitive semantic canonicalizer (DI-1), locked D-04 (--expanded ignores properties.expanded), locked COMPAT-01 (all-collapsed C1), proved the multilevel 5-node C1 + C2/C3 end-to-end (ROADMAP SC4), and removed every "cyp-auth-infra" reference from .go files — COMPAT-01/COMPAT-02 are now enforceable in CI on fresh checkouts.**

## Performance

- **Duration:** 8 min
- **Started:** 2026-08-06T11:52:00Z
- **Completed:** 2026-08-06T12:00:00Z
- **Tasks:** 3
- **Files modified:** 3 (test-only) + this SUMMARY

## Accomplishments

- **TestBuildExpandedGraphRealToml + TestBuildExpandedGraphBaselineDOT repointed** at `../../cmd/c4drill/testdata/multilevel.toml` with sanitized hardcoded names (`mainSystem.storages.externalStorage.client` → `externalSys`, `length=3` → `minlen=3`) — the last gitignored fixture dependency in test inputs is gone (D-01).
- **COMPAT-02 golden enforcement via canonicalDOT**: the BaselineDOT test compares the freshly rendered expanded DOT against the committed 1041-line `multilevel.expanded.dot` using an order-insensitive semantic normalization (statement kind/head, sorted attributes with layout geometry stripped, recursively sorted children). Validated: two independent pipeline runs and the committed golden canonicalize to identical forms; semantic differences (C1 outputs) still fail — the D-02 contract (node/edge sets, attributes, cluster structure) is enforced without byte-equality.
- **D-04 locked**: `TestBuildExpandedGraphIgnoresPropertiesExpanded` poisons `m.Properties.Expanded = ["mainSystem"]` before `GenerateExpandedView` and asserts all 84 fixture units remain (`require.Len` vs recursive walk), deep leaves `mainSystem.sshAuth.systemd.logind` / `mainSystem.storages.externalStorage.client` render as nodes, and zero labels carry `[+]`. Passed on first run (locking test — production machinery already correct; no RED possible by design).
- **COMPAT-01 locked**: `TestCompat01_ValidTomlAllCollapsed` proves hint-free valid.toml renders a correct all-collapsed C1 — `user`/`app` nodes, `Application [+]` indicator, no `subgraph cluster_`, no `app.api`. Deliberately does not assert sub-diagram absence (valid/app.dot IS generated by Phase 2 auto-detect).
- **ROADMAP SC4 proven end-to-end**: `TestCompat02_MultilevelFixtureFiveNodeC1` runs the real CLI on the public fixture — 5-node C1 (`actorA`/`actorB`/`actorC`/`externalSys`/`mainSystem`), `Main System [+]`, no deep paths, plus C2 `multilevel/mainSystem.dot` and C3 `multilevel/mainSystem/sshAuth.dot`.
- **Repo-wide gate green**: `grep -rn "cyp-auth-infra" --include="*.go" .` returns nothing; `loadCYPAuthInfraModel` is now a thin deterministic wrapper over `createSyntheticCYPModel` (render tests behave identically on every machine, D-01).
- **Full wave gates green**: `go test -race ./...` all 7 packages ok; `golangci-lint run --new-from-rev=f72fbe3 ./...` 0 issues; scope guard `git diff internal/view/scope.go internal/graph/builder.go` empty (production byte-identical).

## Task Commits

Each task was committed atomically:

1. **Task 1: Repoint RealToml + BaselineDOT to public fixture; upgrade BaselineDOT to golden comparison** - `e205403` (test)
2. **Task 2: D-04 locking test — --expanded ignores properties.expanded** - `76fba07` (test)
3. **Task 3: COMPAT-01 regression + multilevel 5-node C1/C2/C3 integration; remove private-fixture probing** - `d393b77` (test)
4. **Lint compliance fixes for the new tests (gocognit/wsl_v5/gosec/staticcheck)** - `23a7910` (test)

**Plan metadata:** pending final commit

## Files Created/Modified

- `internal/graph/builder_test.go` - RealToml repointed to multilevel.toml (sanitized names, minlen=3 assertions); BaselineDOT now a canonicalDOT semantic golden comparison vs multilevel.expanded.dot (penwidth-2.0 loop kept); new TestBuildExpandedGraphIgnoresPropertiesExpanded + helpers countModelUnits/countUnitSubunits/collectExpandedGraphNodes (+cluster/node collectors); canonicalDOT xdot parser/serializer (parseDOTStatements/parseDOTBlock/parseDOTSubgraph/parseDOTAttrStatement/skipDOTWhitespace/normalizeDOTAttrs/isGeometryAttr/serializeDOTStatements/serializeDOTStatement)
- `cmd/c4drill/root_test.go` - TestCompat01_ValidTomlAllCollapsed + TestCompat02_MultilevelFixtureFiveNodeC1 (CLI harness per TestFormatFlag_Dot; no t.Parallel; //nolint:paralleltest + //nolint:gosec with explanations)
- `internal/render/expanded_internal_test.go` - loadCYPAuthInfraModel reduced to a wrapper over createSyntheticCYPModel (paths slice, probing loop, t.Logf fallback deleted); both doc comments reworded without the gated literal; 4 call sites untouched

## Decisions Made

- Golden comparison is order-insensitive canonicalDOT (semantic D-02 form), NOT byte-exact require.Equal — mandated by the DI-1 blocker recorded in 03-01/deferred-items.md (plan's byte-exact design is empirically impossible; 5/5 pipeline runs differ byte-wise).
- Geometry attributes (bb/pos/lp/lheight/lwidth/height/width) are stripped from the canonical form — they are layout output, not D-02 contract content; node/edge sets, key= dedup attrs, labels, minlen/penwidth, and cluster structure are preserved.
- Statement terminator scan uses `];` because `&#x1F464;`-style HTML entities inside person labels contain `;`.
- COMPAT-01 asserts C1 RENDERING only; no os.IsNotExist assertion on valid/app.dot (sub-diagram exists via Phase 2 auto-detect).

## Deviations from Plan

### Mandatory Deviation (DI-1, recorded by 03-01)

**1. [DI-1 - Blocking design change] Golden comparison replaced with order-insensitive semantic comparison**
- **Found during:** Plan review (blocker recorded in 03-01 SUMMARY + STATE.md + deferred-items.md; empirically confirmed again during Task 1: two fresh pipeline runs differ byte-wise, 570 diff lines)
- **Issue:** The plan specified `require.Equal(t, string(expected), string(dotData))` byte-exact equality against the committed golden baseline (03-02-PLAN.md Task 1 / interfaces :103-105). The pinned go-graphviz fork (`onokonem/go-graphviz` cgraph/gvc WASM layout) emits map-order-dependent sibling cluster/node order and layout geometry, so byte-exact comparison fails on the first run. The plan itself was corrected by the project rule: "REPLACE the byte-exact comparison with an order-insensitive, semantics-preserving comparison per D-02."
- **Fix:** Implemented `canonicalDOT` in builder_test.go: parses xdot into statements (kind, head, attributes; subgraph nesting preserved), strips layout geometry, sorts attributes within statements and statements recursively, serializes canonically. The test asserts `require.Equal(t, canonicalDOT(expected), canonicalDOT(dotData))`. Verified: 2 independent runs + committed golden → identical canonical forms; C1 outputs (different semantics) → still unequal.
- **Files modified:** internal/graph/builder_test.go (Task 1 commit)
- **Verification:** go test ./internal/graph/ -run BaselineDOT -count=5 green; race suite green
- **Committed in:** e205403 (Task 1 commit; deviation documented here, in 03-01 SUMMARY, and in deferred-items.md DI-1)

### Auto-fixed Issues

**2. [Rule 1 - Bug] DOT statement terminator scan broke on HTML entity semicolons**
- **Found during:** Task 1 (canonicalDOT design validation)
- **Issue:** The first scan for `;` truncated person-node statements at `&#x1F464;` (emoji HTML entity), producing garbage statements and mismatched canonical forms
- **Fix:** Terminate attribute statements on `];` instead of `;` (HTML labels contain neither `]` nor `];`)
- **Files modified:** internal/graph/builder_test.go
- **Verification:** prototype comparison r1==r2==golden before committing; tests green after
- **Committed in:** e205403 (Task 1 commit)

**3. [Rule 1 - Bug] Lint violations in new test code (gocognit 27>15, wsl_v5 x5, gosec G304 x2, staticcheck S1021)**
- **Found during:** Wave gate lint run (`golangci-lint run --new-from-rev=f72fbe3 ./...`)
- **Issue:** parseDOTBlock cognitive complexity 27; cuddled statements; os.ReadFile on t.TempDir() paths; recursive closure var-then-assign pattern
- **Fix:** Refactored parseDOTBlock into parseDOTSubgraph/parseDOTAttrStatement/skipDOTWhitespace; converted closures to named recursive helpers; added `//nolint:gosec // G304: Test reads from temp directory created by t.TempDir()` (existing repo convention, internal/output/writer_test.go); wsl spacing
- **Files modified:** internal/graph/builder_test.go, cmd/c4drill/root_test.go
- **Verification:** `golangci-lint run --new-from-rev=f72fbe3 ./...` → 0 issues
- **Committed in:** 23a7910 (follow-up lint commit)

---

**Total deviations:** 1 mandatory design change (DI-1) + 2 auto-fixed (2 bugs — one design-blocker-driven, one lint compliance)
**Impact on plan:** The DI-1 change is required for the plan to be executable at all (documented in STATE.md as a blocker since 03-01); the golden comparison remains faithful to the D-02 semantic contract. No scope creep; no production code touched.

## Issues Encountered

- **DI-1 order nondeterminism re-confirmed:** two fresh CLI runs of `--format dot --expanded` on multilevel.toml differ in 570 diff lines (sibling cluster order flips, e.g. systemd vs pam under sshAuth, plus geometry). Semantic content stable in every run — the basis for the canonicalDOT approach.
- **grep/ugrep shell quirks:** `grep -c '\[\+\]'` and `grep '^user\t\|^app\t'` behaved inconsistently under zsh/ugrep during empirical verification; switched to direct file inspection for facts (assertions themselves are Go `assert.Contains`, unaffected).
- **`mise lint` not run** — golangci-lint v2.12.2 invoked directly (same configuration via .golangci.yml); plan's `mise lint` alias was unavailable.

## Stub Tracking

None — all tests are fully wired to committed fixtures and the synthetic model; no placeholders, no skipped paths, no TODO stubs.

## Threat Flags

None — changes are test-only; the new file-access patterns are os.ReadFile/os.Stat on committed testdata (multilevel.expanded.dot, relative from internal/graph) and t.TempDir() outputs, all covered by existing gosec convention. T-03-04 (repo-wide grep gate), T-03-05 (require.NoError on baseline read — missing/corrupt baseline fails loudly), and T-03-06 (poison-before-view assertion order) all hold.

## Next Phase Readiness

- COMPAT-01 and COMPAT-02 are now enforceable in CI on fresh checkouts: no test references the gitignored private fixture; the golden baseline is committed; the comparison is deterministic across runs.
- D-04 (expanded ignores properties.expanded) and the multilevel 5-node C1 + C2/C3 pipeline (ROADMAP SC4) are locked as regressions.
- **Deferred (Rule 4, needs user decision in a future phase):** byte-determinism of the render pipeline — fix/repin the `onokonem/go-graphviz` fork or drop byte-equality permanently (canonicalDOT covers semantic enforcement). See deferred-items.md DI-1.
- Phase 03 (compatibility-validation) is complete: 2/2 plans done; all 4 success criteria of the phase roadmap verified.

---
*Phase: 03-compatibility-validation*
*Completed: 2026-08-06*

## Self-Check: PASSED

All claims verified: commits e205403/76fba07/d393b77/23a7910 exist; modified files present; `go test -race ./...` green (7 packages); `golangci-lint run --new-from-rev=f72fbe3 ./...` 0 issues; `grep -rn "cyp-auth-infra" --include="*.go" .` empty; scope guard `git diff internal/view/scope.go internal/graph/builder.go` empty.
