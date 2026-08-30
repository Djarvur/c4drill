---
phase: 37-nesting-context-and-plain-rendering
plan: 03
subsystem: graph
tags: [plain, cli, graph, go, tdd]

# Dependency graph
requires:
  - phase: 37-nesting-context-and-plain-rendering
    plan: 01
    provides: recursive buildCluster/buildNestedCluster + cluster drill affordance — the nested-cluster styling paths the applyUnitOverrides plain guard now covers
    plan2: 02
    provides: UnfoldChain dispatch + resolved deep-link edges — the new edge paths createEdge's plain neutralization must cover
provides:
  - "PLAIN-01: `c4drill --plain` cobra flag threaded to View.Plain on every generated view (processView AND processExpandedView)"
  - "PLAIN-02: builder-level suppression — applyUnitOverrides skips author color/style/border; createEdge neutralizes link color/style/length/rank/label-position; properties.edges ignored (EdgeStyle cleared in BOTH graph entry points)"
  - "Kind-derived edge colours + legend proven to survive plain mode (semantic, not formatting)"
  - "BC-01 locked at builder level: Plain=false reproduces today's output exactly (TestBuildGraph_DefaultPathUnchanged + full suite green)"
  - "Tests: TestBuildGraph_PlainSkipsUnitOverrides, TestBuildGraph_PlainNeutralizesEdgeFormatting, TestBuildGraph_PlainKeepsKindColourAndLegend, TestBuildGraph_PlainCopiedFromView, TestBuildGraph_DefaultPathUnchanged"
affects: [37-04 (PLAIN-03 render-side label simplification + PLAIN-04 E2E plain goldens build on Graph.Plain), 37-05 (golden re-baseline scope)]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Flag-on-struct threading (View.Plain → Graph.Plain → guards) following the AllExpanded precedent — explicitly NOT a package global (LabelRatio counter-pattern rejected)"
    - "Plain neutralization by 'treat author fields as unset' (linkColor/linkStyle/labelPosition locals) so the existing default-application logic stays the single code path"

key-files:
  created: []
  modified:
    - internal/view/view.go
    - internal/graph/graph.go
    - internal/graph/builder.go
    - internal/graph/builder_test.go
    - cmd/c4drill/root.go

key-decisions:
  - "applyUnitOverrides takes a plain bool and returns immediately when set — one guard covers all four call sites (buildNode, buildCluster, buildNestedCluster, buildBoundaryCluster), so nested-cluster styling (37-01 code paths) is covered without extra branches"
  - "buildNode/buildCluster gained a plain bool parameter threaded from callers where v is in scope; createEdge/applyCollapsedPairStyle likewise — no generator signature changes, no render changes"
  - "Collapsed-pair AGG-02 aggregate style override kept inert under plain (deviation 1): the aggregate line style derives from author link.Style fields, so letting it through would leak custom styles; the aggregate's kind/source-border COLOUR logic still applies because kind colours survive plain mode"
  - "RED accepted as compile failure (missing View.Plain/Graph.Plain) — the plan explicitly documents this as the acceptable RED for a field-addition feature, so unlike 37-01/37-02 no inert-field deviation was needed"

patterns-established:
  - "Plain-mode guard contract: author formatting suppressed at the builder (the single point every render path consumes); render-side label simplification stays in 37-04"

requirements-completed: [PLAIN-01, PLAIN-02]

# Metrics
duration: 11min
completed: 2026-08-30
---

# Phase 37 Plan 03: Plain Rendering — --plain Threading + Builder Suppression Guards (PLAIN-01/PLAIN-02) Summary

**`c4drill --plain` threads from a cobra flag through View.Plain into Graph.Plain (copied by BOTH BuildGraph and BuildExpandedGraph) and suppresses author-custom formatting at the builder level — unit color/style/border, link color/style/length/rank, label position, and properties.edges all fall back to defaults — while kind-derived edge colours and the legend survive.**

## Performance

- **Duration:** 11 min
- **Started:** 2026-08-30T13:23:46Z
- **Completed:** 2026-08-30T13:34:45Z
- **Tasks:** 3 (RED / GREEN / VERIFY)
- **Files modified:** 5

## Accomplishments
- PLAIN-01: `--plain` registered as a cobra PersistentFlag (directly after `--expanded`) and assigned to every generated view in `processView` AND `processExpandedView` — `--plain` × `--expanded` works (Pitfall 5)
- PLAIN-02 builder guards: `applyUnitOverrides(style, unit, plain)` returns immediately under plain — one guard covering plain nodes, expanded clusters, nested clusters (37-01 recursion), and the C2 boundary cluster; `createEdge(..., plain)` treats link Color/Style/LabelPosition as unset and ignores Length/Rank (no minlen, no endpoint swap, no constraint suppression); `EdgeStyle` cleared at both graph entry points so `configureGraphSettings` falls through to default routing
- Boundary locked by test: kind-derived edge colours (kindColour) and the legend survive plain mode untouched — no plain branch added to kindColour, buildLegend, or legendKindEntries
- Default path pinned: `TestBuildGraph_DefaultPathUnchanged` asserts Plain=false reproduces today's overrides/formatting exactly; full uncached suite green (15/15 packages) including both committed goldens — zero default-output deltas
- CLI smoke run (multilevel.toml → /tmp, testdata untouched): `--plain` DOT carries no minlen/constraint/dashed/custom-colour attributes, legend present, output differs from the default run — flag effective end to end

## Expected-Failure List (LOAD-BEARING for 37-04 / 37-05)

**EMPTY — zero failing tests in the repo after GREEN.** Uncached `go test -count=1 ./...`: 15/15 packages ok, 0 failures, including both committed-golden baselines (`TestBuildExpandedGraphBaselineDOT` + REF-05 `TestReference_BackwardCompat`). Plain defaults to false, so no default-path delta is possible; the `--plain` delta only materializes when the flag is set, and 37-04's E2E plain goldens (PLAIN-04) are additive by design (PATTERNS.md: "new plain goldens additive only"). **37-05's re-baseline scope: verify flat-model stability (BC-01) + the goldens 37-04 adds; nothing to re-baseline from this plan.**

## Task Commits

1. **Task 1 (RED): Plain-guard tests** - `ce7b277` (test) — package fails to compile on the missing View.Plain/Graph.Plain fields (the plan's documented acceptable RED; no production code changed, no compile-gate deviation needed)
2. **Task 2 (GREEN): Plain threading + builder guards** - `8dc2277` (feat)
3. **Task 3 (VERIFY): default output untouched, suite categorized** - no commit (verification only, no source edits)

_Note: TDD plan — test → feat commits; no refactor commit needed (implementation landed clean: golangci-lint 0 issues, gofmt clean, vet clean on all changed packages)._

## Files Created/Modified
- `internal/view/view.go` — `View.Plain bool` next to `Edges`, documented in the AllExpanded style
- `internal/graph/graph.go` — `Graph.Plain bool` next to `EdgeStyle`
- `internal/graph/builder.go` — `Plain: v.Plain` + EdgeStyle-clearing local in BuildGraph AND BuildExpandedGraph; `plain bool` params on applyUnitOverrides/buildNode/buildCluster/createEdge/applyCollapsedPairStyle with all call sites updated (dispatch in buildBoundaryViewGraph, buildC1ViewGraph, BuildExpandedGraph, buildNestedCluster recursion); kindColour/buildLegend/legendKindEntries untouched
- `internal/graph/builder_test.go` — five Plain-guard tests (12 subtests total incl. collapsed-pair style guard and C2 boundary cluster)
- `cmd/c4drill/root.go` — `plain bool` var + `--plain` PersistentFlag registration + `v.Plain = plain` in processView and processExpandedView

## Decisions Made
- Neutralization implemented as "treat author fields as unset" (locals `linkColor`/`linkStyle`/`labelPosition` cleared under plain) so the existing default-application chain (style→solid, position→middle, colour→kind→source-border) remains the single code path — no duplicated default logic
- `applyCollapsedPairStyle` keeps its kind/source-border colour logic under plain but skips the aggregate line-style override — kind colouring for collapsed pairs is semantic (survives), the aggregate style is author formatting (suppressed)
- Edge `ArrowHead` (link.Arrow) NOT suppressed: arrow direction is data-flow semantics, not formatting; the plan lists only color/style/length/rank/labelPosition/edges
- `PenWidth` (multiplicity thickness) NOT suppressed: D-04 multiplicity is structural, not author formatting; plan does not list it

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing critical] Collapsed-pair aggregate style kept inert under plain**
- **Found during:** Task 2 (GREEN)
- **Issue:** The plan's action list guards `createEdge` only, but collapsed pairs (2+ parallel links) get their surviving edge's Style overridden afterwards by `applyCollapsedPairStyle` (AGG-02) from the pair aggregate — which is derived from author `link.Style` fields. Without a guard, two dashed links would render a dashed edge under `--plain`, violating the plan's own must_have truth ("link color/style fall back to defaults")
- **Fix:** Added `plain bool` to `applyCollapsedPairStyle`; the style assignment is skipped when plain while the colour logic (kind colour / source-border default) still applies; added the `collapsed pair keeps the default style` subtest to TestBuildGraph_PlainNeutralizesEdgeFormatting
- **Files modified:** internal/graph/builder.go, internal/graph/builder_test.go
- **Verification:** subtest passes (collapsed dashed pair renders solid under plain); default-path pair-aggregation tests (TestPairAggregation) all still green
- **Committed in:** 8dc2277

---

**Total deviations:** 1 auto-fixed (missing critical). **Impact on plan:** closes a leak the plan's must_haves forbid but its action list missed; no scope creep — same file, one parameter + one guard. Notably, NO compile-gate deviation was needed this time: the plan pre-authorized compile-failure RED for the field additions.

## Issues Encountered
- One batched edit to `buildNode`'s signature was initially missed (its edit block also carried the buildNestedCluster child-loop change, which didn't match); the compiler caught it immediately and the signature edit was applied before any test run. No impact on the final state.
- Pre-existing, out-of-scope (untouched per scope boundary): gofmt comment-formatting drift in `internal/graph/shapes.go` and `internal/graph/integration_test.go` (files this plan did not modify); untracked `.planning/.../37-PATTERNS.md` and phase-36 planning-file deletions predate this execution (orchestrator state, left alone).

## Known Stubs

None — every suppression guard is wired end to end. Note for the verifier: `--plain`'s help text promises "plain-text labels", but render-side label simplification is PLAIN-03, deliberately scheduled for plan 37-04 (plan objective: "Render-side label simplification (PLAIN-03) and the E2E goldens (PLAIN-04) land in plan 37-04"). Labels still render through the HTML path today, in plain mode too.

## TDD Gate Compliance
- RED gate: `ce7b277` `test(37-03)` — package failed to compile on the missing Plain fields (documented acceptable RED for a field-addition feature; the five tests cannot pass until the fields exist)
- GREEN gate: `8dc2277` `feat(37-03)` — all five Plain tests PASS; full repo suite green (15/15 packages, uncached)
- REFACTOR gate: not needed (lint/gofmt/vet clean on first landing)

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- 37-04 (PLAIN-03 label simplification + PLAIN-04 E2E plain goldens) can key off `Graph.Plain` — the flag already reaches the render pipeline with zero render changes in this plan, so 37-04 stays render-only as designed (the LabelPosition-ignore guard landed here in createEdge per plan)
- 37-05: no committed-golden debt from this plan (see Expected-Failure List); verify BC-01 flat-model stability + scan the goldens 37-04 introduces
- Full repo suite green (uncached): 15/15 packages; `go build ./...` clean; golangci-lint 0 issues on graph/view/cmd; `cmd/c4drill/testdata/` untouched (verified via git status)

---

## Self-Check: PASSED

- All 5 modified files exist on disk (verified with `[ -f ]` via the edits landing)
- Commits verified in `git log`: `ce7b277` (test), `8dc2277` (feat)
- Plan-level `<verification>` re-run: `go test ./internal/graph/ -run 'Plain' -v` → 5/5 PASS; `go run ./cmd/c4drill --help` lists `--plain`; `go test ./internal/view/ ./internal/render/` PASS; `go vet ./internal/graph/ ./cmd/c4drill` clean; `go test ./internal/graph/` PASS (failure list unchanged — empty)
- No file deletions in either commit (`git diff --diff-filter=D` empty)
- `cmd/c4drill/testdata/` untouched (git status clean of testdata paths)

---
*Phase: 37-nesting-context-and-plain-rendering*
*Completed: 2026-08-30*
