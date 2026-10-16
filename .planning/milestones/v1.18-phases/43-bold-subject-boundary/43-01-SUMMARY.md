---
phase: 43-bold-subject-boundary
plan: 01
subsystem: renderer
tags: [graphviz, dot, penwidth, cluster, semantic-styling, tdd]

# Dependency graph
requires: []
provides:
  - "NodeStyle.BorderWidth field (0 = renderer default, no attribute emitted)"
  - "buildBoundaryCluster assigns BorderWidth = 3.0 at the single subject-boundary construction site"
  - "applyClusterStyle emits penwidth via setClusterAttribute for the subject cluster only"
  - "TDD RED test suite pinning the contract: C2/C3/deep-link raw-DOT assertions + flag matrix"
affects: [43-02, golden audit]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "0-means-renderer-default numeric style field (mirrors Edge.PenWidth, D-02)"
    - "semantic navigation aid decided at graph construction, outside the author-override path (survives --plain/--no-styles, D-03)"

key-files:
  created: []
  modified:
    - internal/graph/graph.go
    - internal/graph/builder.go
    - internal/render/converter.go
    - internal/graph/builder_test.go
    - internal/render/converter_test.go
    - cmd/c4drill/root_test.go

key-decisions:
  - "Emphasis assigned once in buildBoundaryCluster before the return, covering both the unit and nil-unit fallback branches (RESEARCH Q2) — single occurrence, grep -c = 1"
  - "Emitted string pinned to strconv.FormatFloat(w, 'f', 1, 64) = \"3.0\" (D-01 literal, never fmt.Sprint)"
  - "No-penwidth guards scoped to cluster attribute statements (the subgraph's own `graph [...]` block): edge statements legitimately carry penwidth=1/2, so a whole-file regexp would be a false guard"

patterns-established:
  - "Per-cluster raw-DOT attribute extraction (clusterPenwidthValues): stack-walk subgraph blocks, scan each cluster's own graph-[...] statement, collect penwidth values"
  - "assertSubjectBoundaryPenwidth: exactly [3.0] on the subject cluster, empty on every other cluster"

requirements-completed: [BOLD-01, BOLD-02]

# Metrics
duration: 7min
completed: 2026-09-07
---

# Phase 43 Plan 01: Semantic Bold Subject Boundary Summary

**NodeStyle.BorderWidth plumbing emits penwidth=3.0 on the subject boundary cluster of every non-expanded drill-down view (C2, C3, deep-link), surviving --plain/--no-styles, with zero collateral on other clusters, the C1 root, or --expanded goldens**

## Performance

- **Duration:** 7 min
- **Started:** 2026-09-07T07:31:00Z
- **Completed:** 2026-09-07T07:37:48Z
- **Tasks:** 3 (TDD RED → GREEN → REFACTOR)
- **Files modified:** 6

## Accomplishments

- RED suite (3 test functions, 9 assertion subtests) proving the contract fails on current output — subject clusters render with no penwidth
- `NodeStyle.BorderWidth float64` with the D-02 0-means-default contract in graph.go
- Single construction-site assignment `style.BorderWidth = 3.0` in buildBoundaryCluster (covers unit and nil-unit branches; outside applyUnitOverrides so the emphasis is semantic, D-03)
- `applyClusterStyle` emits `penwidth=3.0` via `setClusterAttribute` gated on `BorderWidth > 0` with `FormatFloat('f',1,64)` formatting
- Full module suite green on REFACTOR: `go test ./... -count=1` passes, canonical golden consumers (COMPAT-02/REF-05, multilevel.expanded.dot) unchanged, zero golden drift
- CLI flag matrix proves end-to-end: C1 root and --expanded carry no bold penwidth; C2 (cluster_mainSystem) and deep-link C3 (cluster_mainSystem.sshAuth) carry exactly one penwidth=3.0; --plain keeps it

## Task Commits

Each task was committed atomically:

1. **Task 1: RED — subject-boundary penwidth test suite** - `7ed953f` (test)
2. **Task 2: GREEN — BorderWidth field, construction assignment, penwidth emission** - `bf2007d` (feat)
3. **Task 3: REFACTOR — regression sweep + golden diff audit** - no commit (no code changes needed; full suite already green, zero re-baselines)

**Plan metadata:** `docs(43-01)` pending (committed with this summary)

## Files Created/Modified

- `internal/graph/graph.go` - NodeStyle gained `BorderWidth float64` (0 = no attribute, mirrors Edge.PenWidth)
- `internal/graph/builder.go` - buildBoundaryCluster assigns `style.BorderWidth = 3.0` once before return, with BOLD-01/D-03 anchor comment
- `internal/render/converter.go` - applyClusterStyle emits penwidth gated on `BorderWidth > 0`; strconv imported
- `internal/render/converter_test.go` - `clusterPenwidthValues` helper + `TestSubjectBoundaryClusterPenwidth` (C2, C3, --plain, --no-styles, --no-colors)
- `internal/graph/builder_test.go` - `TestSubjectBoundaryNoCollateral` (C1 root no bold penwidth, --expanded canonical COMPAT-02 equality, deep-link C3)
- `cmd/c4drill/root_test.go` - `TestSubjectBoundaryCLIFlagMatrix` on the multilevel fixture (C1/C2/C3/--expanded/--plain)

## Decisions Made

- **Single assignment before return in buildBoundaryCluster** (not inside the unit branch): covers the nil-unit fallback style too per RESEARCH Q2, keeps the single construction site invariant (grep -c = 1)
- **Value formatting pinned to `FormatFloat('f', 1, 64)`**: emits the D-01 literal `3.0`, not `3`
- **Cluster-scoped no-penwidth guards**: the C1/--expanded assertions use `penwidth=3\.0` absence plus per-cluster attribute extraction instead of a naive whole-file regexp — edge statements legitimately carry `penwidth=1/2`, so only cluster attribute blocks are meaningful for the "no other cluster affected" contract

## Deviations from Plan

### Interpretive Deviations (no code drift)

**1. [Interpretive - Test scope] C1 "no penwidth anywhere" scoped to cluster attributes**
- **Found during:** Task 1 (RED suite authoring)
- **Issue:** The plan's literal acceptance "C1 root contains NO penwidth anywhere (regexp `penwidth=` must not match)" is unsatisfiable as stated — every DOT render already emits `penwidth=1`/`penwidth=2` on EDGE statements (pre-existing D-04 edge contract, converter.go:899-905).
- **Fix:** The C1/--expanded no-collateral guards assert absence of `penwidth=3.0` (the bold value) in the whole file and absence of ANY penwidth in every cluster's own attribute statement via the per-cluster helper — the meaningful D-04 contract.
- **Files modified:** test files only
- **Verification:** Full suite green; C2/C3 subject-scoped assertions still prove single-cluster emphasis
- **Committed in:** 7ed953f (Task 1 commit)

**2. [Interpretive - Test scope] Subject cluster ID naming**
- **Found during:** Task 1 (RED suite authoring)
- **Issue:** The plan's phrasing "subject cluster cluster_auth / cluster_sshAuth" is loose — the boundary cluster ID is `v.ExpandedUnit` (full dotted path), so C3 subjects render as `cluster_mainSystem.auth` / `cluster_mainSystem.sshAuth`.
- **Fix:** Tests assert the full emitted IDs; the assertion helper is ID-agnostic.
- **Files modified:** test files only
- **Verification:** All assertions pass against real emitted IDs
- **Committed in:** 7ed953f (Task 1 commit)

**3. [Pre-existing condition - gofmt] Three pre-existing gofmt drifts left untouched**
- **Found during:** Task 3 (REFACTOR)
- **Issue:** `gofmt -l internal/graph internal/render cmd/c4drill` lists `internal/graph/shapes.go`, `internal/graph/integration_test.go`, `internal/render/integration_test.go` — all three verified pre-existing at HEAD (git show HEAD:... | gofmt -d) and untouched by this phase.
- **Fix:** None — phase-modified files are all gofmt-clean; touching unrelated files would violate D-04 byte-identity scope.
- **Files modified:** none
- **Verification:** gofmt -l on the six phase-modified files prints nothing
- **Committed in:** n/a

---

**Total deviations:** 3 (2 interpretive test-scope clarifications, 1 pre-existing condition)
**Impact on plan:** All three preserve the plan's intent — the D-04/C1 guards are stricter than a naive regexp could be, subject IDs are asserted exactly as emitted, and no unrelated file was reformatted.

## Issues Encountered

- `clusterPenwidthValues` initially returned an empty map in RED state: clusters without the penwidth attribute were never registered. Fixed by always materializing every cluster's entry (empty slice = no penwidth) so `require.Contains(subject)` and the per-cluster empty assertions work in both RED and GREEN states.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 43-02 can now run the golden-delta audit (BOLD-03): the full suite passed with ZERO golden changes, matching RESEARCH A2's expectation
- `go test ./... -count=1` is green and repeatable (~10s warm)
- Expected audit outcome: zero re-baselines; PROJECT.md gains the "As of v1.18" paragraph

---
*Phase: 43-bold-subject-boundary*
*Completed: 2026-09-07*