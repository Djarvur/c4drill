---
phase: 32-include-directive-multi-file
plan: 02
subsystem: include
tags: [multi-file, include-directive, merge, parser, pipeline, templates]

# Dependency graph
requires:
  - phase: 32-01
    provides: parser.IncludeDirective type + Model.Includes field populated by Parse
  - phase: 31
    provides: Model.Templates / Model.Instantiations fields (XC-02 template carry through merge) + template.Expand Stage 1.5 (the pipeline step include.Resolve runs before)
provides:
  - "internal/include.Resolve(entry, entryDir, entryFile) — recursive resolver with cycle stack, once/diamond visited-set, missing-file hard-error, max-depth cap"
  - "internal/include.merge + mergeUnits + mergeSubunits + mergeProperties + mergeTemplates — per-field struct-union (D-09/D-10/D-11/INC-08/XC-02)"
  - "Pipeline Stage 1a wiring in cmd/c4drill/root.go (include.Resolve runs FIRST, before template.Expand)"
affects: [33-04]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Recursive include resolution: stack (cycle, INC-04) + visited-set (once/diamond, INC-06/D-11) keyed by canonical path (filepath.Clean+Abs)"
    - "Per-field struct merge: Units union/conflict, UnitOrder append (D-09), Subunits deep-merge (D-10), Properties root-wins/conflict (INC-08 via reflection), Templates union/conflict (XC-02), Instantiations append"
    - "Committed-fixture tests resolve absolute fixturePath(); writeFiles(t, map) helper for multi-file graphs in t.TempDir()"

key-files:
  created:
    - internal/include/resolve.go
    - internal/include/merge.go
    - internal/include/include_test.go
    - internal/include/testdata/main.toml
    - internal/include/testdata/auth.toml
    - internal/include/testdata/nested_subunits_main.toml
    - internal/include/testdata/nested_subunits_auth.toml
  modified:
    - cmd/c4drill/root.go
    - cmd/c4drill/root_test.go

key-decisions:
  - "Resolve signature is Resolve(entry, entryDir, entryFile) — entryFile threads the real entry filename into error attribution (Plan refined the simplified Resolve(m) from CONTEXT so INC-10/D-12 errors name the including file, not a placeholder)"
  - "D-10 included fixture must declare its own parent: [linuxSystem.auth] WITHOUT [linuxSystem] parses to nothing (captureDefinitionOrder only records parent+child; orphaned subunits are lost). The included file redeclares [linuxSystem]; merge detects the parent already exists in the entry and attaches ONLY the subunits, with the entry's scalar fields authoritative."
  - "Committed-fixture tests use absolute fixturePath() because t.Parallel() runs them concurrently and the cd-proof test (TestResolveRelativePathIndependentOfCwd) mutates the process cwd — relative paths would race."
  - "cd-proof test uses t.Chdir() (Go 1.24+) which auto-restores cwd; incompatible with t.Parallel() so that one test is marked //nolint:paralleltest."
  - "INC-08 properties conflict uses reflection over model.Properties (covers all 8 fields uniformly without enumeration); entry root-wins by virtue of being the merge destination."

patterns-established:
  - "Multi-file test graphs via writeFiles(t, map[string]string) string helper returning a t.TempDir() — keeps cycle/diamond/once/transitive/dup/props/missing fixtures self-contained in the test source (idiomatic Go)"
  - "Error attribution names BOTH files: ParseError.Context = fmt.Sprintf(\"%s (included from %s)\", referencedPath, includingFile)"

requirements-completed: [INC-01, INC-02, INC-03, INC-04, INC-05, INC-06, INC-07, INC-08, INC-09, INC-10, XC-02]

# Metrics
duration: 22min
completed: 2026-08-08
---

# Phase 32 Plan 02: internal/include package + pipeline wiring Summary

**Recursive multi-file include resolver + per-field struct merge land as the pipeline's FIRST pre-processing pass (Stage 1a), delivering INC-01 through INC-10 and XC-02 — templates in included files flow through the merge so [[use]] in the entry file can instantiate them.**

## Performance

- **Duration:** ~22 min
- **Tasks:** 3 (RED tests+fixtures → GREEN resolver+merge → pipeline wiring+integration tests)
- **Files created:** 7 (resolve.go, merge.go, include_test.go, 4 testdata fixtures)
- **Files modified:** 2 (cmd/c4drill/root.go, cmd/c4drill/root_test.go)

## Accomplishments
- `internal/include.Resolve(entry, entryDir, entryFile)` walks `entry.Includes`, recursively merges transitively-included files into one `*parser.Model`. The merged model is structurally indistinguishable from a hand-authored single-file model (validator/view/render unchanged).
- INC-01: multi-file model merges into one logical Model (entry units + included units).
- INC-02: paths resolve relative to the INCLUDING file's dir (cd-proof test passes — identical from two different cwds).
- INC-03: transitive includes resolve recursively (top→mid→leaf chain).
- INC-04: cycle (self + mutual) is fatal `*parser.ParseError` naming the cycle; stack-based detection.
- INC-05: diamond NOT a cycle; same-file diamond auto-deduped (shared unit appears once).
- INC-06: `once=true` skips re-inclusion (visited-set dedup).
- INC-07: cross-file duplicate unit path is a hard error naming both files.
- INC-08: properties root-wins; non-zero conflict hard-errors naming both files (reflection over all 8 Properties fields).
- INC-09: UnitOrder concatenation preserves authoring order (entry first, then each include's units in directive order — D-09 append).
- INC-10 / D-12: missing include is an unconditional hard error naming referenced path + including file (no optional flag).
- D-10: cross-file subunit merge — included file's `[parent.child]` attaches to entry-declared parent, appending to SubunitOrder.
- D-11: same-file diamond auto-dedup (byte-identical by construction); cross-file collision hard-errors.
- XC-02: templates defined in included files flow into merged Model.Templates so `[[use]]` in the entry can instantiate them (verified).
- Pipeline wired: `include.Resolve` is Stage 1a in root.go, running FIRST (before `template.Expand` Stage 1.5 and Validate).

## Task Commits

1. **Task 1: Write failing include tests + fixtures (RED)** — `c3d8c9b` (test)
2. **Task 2: Implement resolve.go + merge.go (GREEN)** — `23e581b` (feat)
3. **Task 3: Wire include.Resolve into pipeline + integration tests** — `0d1449d` (feat)

## Files Created/Modified
- `internal/include/resolve.go` — `Resolve` (exported) + recursive `resolve` worker + `resolveOne` per-directive helper + `canonicalize`. Stack for cycles, visited-set for once/diamond, maxIncludeDepth=100 defense-in-depth.
- `internal/include/merge.go` — `merge` orchestrator + `mergeUnits`/`mergeSubunits`/`mergeProperties`/`mergeTemplates` per-field rules.
- `internal/include/include_test.go` — 12 test functions (13 subtests) covering INC-01..10, D-10, D-11, XC-02; `writeFiles` + `fixturePath` + `parseAndResolve` helpers.
- `internal/include/testdata/{main,auth,nested_subunits_main,nested_subunits_auth}.toml` — committed documentation fixtures for INC-01/09 and D-10.
- `cmd/c4drill/root.go` — Stage 1a wiring: `include.Resolve(m, filepath.Dir(inputPath), inputPath)` between ParseFile and template.Expand.
- `cmd/c4drill/root_test.go` — `TestPipelineIncludeBeforeValidate`, `TestPipelineSingleFileNoRegression32`, `TestCLIMissingIncludeExits`.

## Decisions Made
- **Resolve signature**: Added `entryFile` as a third parameter (refining the simplified `Resolve(m)` from CONTEXT). Without it, error attribution for the entry's own [[include]] directives would use a placeholder `<entry>`; tests assert the real entry filename appears in INC-07/INC-08/INC-10 errors.
- **D-10 fixture shape**: The included file must declare its own parent (`[linuxSystem]`) for its `[linuxSystem.auth]` subunits to parse at all. The merge then detects the parent exists in the entry and attaches ONLY the subunits. This is a parser constraint (captureDefinitionOrder records parent.child as a pair; orphaned subunits are lost), not a merge-design choice.
- **INC-08 reflection**: Used reflection over `model.Properties` for uniform 8-field conflict detection rather than enumerating fields — cleaner and future-proof if the struct grows.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] D-10 fixture as authored does not parse**
- **Found during:** Task 2 (Implement resolve.go + merge.go — GREEN)
- **Issue:** The plan's `nested_subunits_auth.toml` fixture specified `[linuxSystem.auth]` and `[linuxSystem.db]` WITHOUT a `[linuxSystem]` parent in the included file (the plan said "the same shape testdata/nested.toml uses" — but nested.toml always has the parent). Verified empirically: such an included file parses to an empty model because `captureDefinitionOrder` records `linuxSystem.auth` as `subunitOrders["linuxSystem"]=["auth","db"]` but `linuxSystem` never enters `unitOrder` (no top-level `[linuxSystem]` table), so the unit loop skips it and the subunits are orphaned. The cross-file subunit merge then has nothing to attach.
- **Fix:** The included fixture now declares `[linuxSystem]` (parent) + `[linuxSystem.auth]` + `[linuxSystem.db]`. The merge logic (`mergeUnits` → `mergeSubunits`) detects the parent already exists in the entry and attaches ONLY the subunits to the entry's parent, with the entry's scalar fields authoritative (D-10). The fixture now matches the actual `testdata/nested.toml` shape (parent + dotted-path subtables).
- **Files modified:** internal/include/testdata/nested_subunits_auth.toml
- **Verification:** `TestMergeCrossFileSubunits` passes — `merged.Units["linuxSystem"].Subunits` has "auth" and "db" in SubunitOrder, neither leaks as a phantom top-level unit.
- **Committed in:** `23e581b` (Task 2 commit, fixture fix bundled with the GREEN implementation)

**2. [Rule 1 - Bug] Resolve signature needed entryFile for error attribution**
- **Found during:** Task 2 (Implement — initial test run showed errors containing `<entry>` placeholder)
- **Issue:** The plan's CONTEXT specified `Resolve(m)` and the PLAN research refined it to `Resolve(entry, entryDir)`, but both leave the entry file's display name unset for INC-07/INC-08/INC-10 error attribution. Initial implementation used `filepath.Join(entryDir, "<entry>")` and tests asserting `"missing_main.toml"` / `"props_main.toml"` / `"dup_main.toml"` in error messages failed.
- **Fix:** Extended the signature to `Resolve(entry, entryDir, entryFile)` and threaded the real entry path through the recursive worker. The root.go wiring passes `inputPath` as entryFile.
- **Files modified:** internal/include/resolve.go, internal/include/include_test.go, cmd/c4drill/root.go
- **Verification:** All error-attribution tests pass — `TestResolveMissingIncludeError`, `TestMergePropertiesConflictError`, `TestMergeDuplicateUnitPathError` all assert the real entry filename appears in the error.
- **Committed in:** `23e581b` (Task 2) and `0d1449d` (Task 3 wiring)

**3. [Rule 1 - Bug] Committed-fixture tests racy with cd-proof test under t.Parallel**
- **Found during:** Task 2 (initial run — TestMergeCrossFileSubunits failed with "failed to read file")
- **Issue:** Committed-fixture tests used relative `ParseFile("testdata/main.toml")` and `parseAndResolve(t, "testdata/...")`. Under `t.Parallel()`, they run concurrently with `TestResolveRelativePathIndependentOfCwd` which calls `os.Chdir` — mutating the process-global cwd and breaking the relative paths of the concurrent fixture tests.
- **Fix:** Added a `fixturePath(name)` helper that resolves an absolute path to `testdata/<name>`; committed-fixture tests now use `parseAndResolve(t, fixturePath("main.toml"))`. The cd-proof test switched to `t.Chdir()` (Go 1.24+, auto-restores cwd) and is marked `//nolint:paralleltest` since `t.Chdir` is incompatible with `t.Parallel()`.
- **Files modified:** internal/include/include_test.go
- **Verification:** Full include suite passes under `t.Parallel()` for the 11 parallel-safe tests; the single cd-proof test runs serially.
- **Committed in:** `23e581b` (Task 2)

---

**Total deviations:** 3 auto-fixed (3 bugs)
**Impact on plan:** All three were implementation-discovery bugs (parser constraint on D-10 fixtures, error-attribution signature gap, parallel-test/cwd interaction). None changed the plan's semantics or success criteria — INC/D/XC requirements all met as specified. The fixture-shape clarification and Resolve signature refinement are forward-compatible with Phase 33's E2E tests.

## Issues Encountered
None beyond the three deviations above.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Phase 32 complete (2/2 plans). All INC-01..10 + XC-02 delivered.
- Phase 33 Plan 04 (XC-05 + XC-01 E2E tests) is now UNBLOCKED — `internal/include.Resolve` and `internal/template.Expand` both ship in HEAD. Its pre-condition gate can fire.
- Validator/view/render/graph remain unchanged (consume the merged `*parser.Model`), per STATE.md D-12.

## Self-Check: PASSED

- `go test ./internal/include/ -x` — all 12 include-package tests green (INC-01..10, D-10, D-11, XC-02)
- `go test ./internal/parser/ -x` — Plan 01's parser tests still green (no regression from Plan 02)
- `go test ./cmd/c4drill/ -x` — pipeline integration tests green (include.Resolve before Validate; single-file no-regression; missing-include CLI error path)
- `go test ./...` — full suite green (validator/view/render/graph consume the merged model unchanged)
- `golangci-lint run ./internal/include/` — 0 issues (clean); `golangci-lint run ./internal/parser/` — no new findings in parser.go; `golangci-lint run ./cmd/c4drill/` — no new findings in my added lines (one G703 nolint on a corpus-fixture copy, consistent with the existing os.WriteFile test idiom)
- `go vet ./...` — clean
- Pipeline ordering confirmed: `include.Resolve` is Stage 1a (FIRST pre-processing pass), before `template.Expand` (Phase 31, Stage 1.5) and Validate (Stage 2)
- No phantom units: a multi-file model produces a merged Model.Units with only real renderable units
- Multi-file model renders equivalently to the hand-authored single-file equivalent (merge-equivalence asserted at the *parser.Model level; SVG golden comparison deferred to Phase 33 per PATTERNS.md canonicalDOT note)

---
*Phase: 32-include-directive-multi-file*
*Completed: 2026-08-08*
