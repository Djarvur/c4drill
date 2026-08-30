---
phase: 38-hierarchy-wrapping-and-granular-keys
plan: 03
subsystem: graph-builder
tags: [cli-flags, no-labels, label-suppression, shape-only, tdd, render-opts]
requires:
  - phase: 38-02
    provides: "RenderOpts struct threaded CLI→View→Graph (NoLabels rides the same path) and an EMPTY census — suite green"
  - phase: 38-01
    provides: "wrapper clusters (wrap_*) whose labels are suppressed too under --no-labels"
provides:
  - "--no-labels flag on every view (processView + processExpandedView), threaded View.NoLabels → Graph.Opts.NoLabels"
  - "Graph-layer label suppression: bare-shape nodes, label-less edges, unlabelled clusters (incl. wrapper/boundary); legend + URL attributes survive"
  - "LBL-01/02/03 complete: everywhere, composing, legend pinned"
affects: [38-04-rebaseline]
tech-stack:
  added: []
  patterns:
    - "converter label sites take graph.RenderOpts (replacing the internal plain bool) so NoLabels and Plain branch beside each other"
    - "count-based `label=<` HTML-label assertions (robust discriminator vs the uppercase sanctioned graph/legend markup)"
key-files:
  created: []
  modified:
    - internal/view/view.go
    - internal/graph/graph.go
    - internal/graph/builder.go
    - internal/graph/builder_test.go
    - internal/render/converter.go
    - internal/render/converter_test.go
    - cmd/c4drill/root.go
    - cmd/c4drill/root_test.go
decisions:
  - "Suppression at the GRAPH layer per plan pin: builder drops Label content before emission; converter NoLabels branch (SetLabel(\"\")) is defense-in-depth for hand-built graphs"
  - "NoLabels node keeps Style (colour/border) and ShapeForType default shape — only label TEXT is suppressed; 🔍/📖 glyphs are label content and go with it, ReferenceURL stays (structural)"
  - "Converter label sites migrated from `plain bool` to graph.RenderOpts so the NoLabels and Plain branches live beside each other (supersedes 38-02's minimal-threading decision where NoLabels made it behaviour-relevant)"
  - "Legend STAYS under --no-labels (LBL-03 planner pin): metadata governed by properties.legend, to be documented in 38-05"
requirements-completed: [LBL-01, LBL-02, LBL-03]
metrics:
  duration: ~30m
  completed: 2026-08-30
---

# Phase 38 Plan 03: --no-labels (Label Suppression) Summary

`--no-labels` silences every element label at the graph layer — nodes render as their bare ShapeForType shapes, edges and clusters (wrapper + boundary included) emit no label text — so dot lays out without label geometry, on every generation and format, while the legend and drill/reference URL attributes survive; census stays EMPTY.

## Performance

- **Duration:** ~30 min
- **Started:** 2026-08-30 (after 38-02)
- **Completed:** 2026-08-30
- **Tasks:** 3
- **Files modified:** 8

## Accomplishments

- `RenderOpts.NoLabels` + `View.NoLabels` threaded exactly like the 38-02 switches: cobra `--no-labels` registered after `--no-rank`, assigned onto every view in processView AND processExpandedView, copied by renderOptsFromView into both graph constructors.
- Builder suppression: buildNode drops label content entirely (bare shape, style/colour semantics kept), createEdge leaves `Edge.Label` nil, buildCluster/buildNestedCluster/buildBoundaryCluster/ensureWrapperCluster skip label construction while retaining cluster IDs and nesting (WRAP structure intact).
- Converter defense-in-depth: label sites (setNodeLabel/setClusterLabel/createEdge) now take `graph.RenderOpts`; under NoLabels they emit empty labels even for hand-built graphs carrying non-nil Labels. Cluster ExploreURL and node Reference/Explore URL attributes survive.
- Legend untouched — its own builder path keeps it under labels-off (LBL-03 pin, asserted at builder, converter and E2E levels).
- Tests: 9 new builder/converter tests + 4 E2E tests (all generations × formats, composition with --plain and granular switches, opt-in locks at both builder and converter levels).

## Task Commits

1. **Task 1 (RED): label-absence + composition tests** - `bdff1c0` (test)
2. **Task 2 (GREEN): NoLabels threading + graph-layer suppression** - `4936563` (feat)
3. **Task 3: full-suite calibration** - no code change; census delta recorded here (below)

## Files Created/Modified

- `internal/graph/graph.go` — RenderOpts.NoLabels
- `internal/view/view.go` — View.NoLabels
- `internal/graph/builder.go` — renderOptsFromView copies NoLabels; label guards at buildNode/createEdge/all four cluster construction sites
- `internal/render/converter.go` — RenderOpts threaded through node/cluster/edge label emission; NoLabels → SetLabel("")
- `cmd/c4drill/root.go` — noLabels var + flag + per-view assignment (both process functions)
- `internal/graph/builder_test.go` — TestNoLabels{NodesAreBareShapes, EdgesHaveNoLabel, ClustersUnlabelled, LegendStays, CopiedFromView, OptInDefaultPathUntouched}
- `internal/render/converter_test.go` — TestNoLabelsDOTEmitsNoLabelMarkup + TestNoLabelsOptInConverterDefaultUnchanged
- `cmd/c4drill/root_test.go` — TestNoLabelsAllGenerationsAndFormats, TestNoLabelsComposesWithPlainAndSwitches, TestNoLabelsOptIn

## Decisions Made

- **Graph-layer suppression (plan pin honoured):** the builder drops Label CONTENT; the converter's NoLabels branch exists only as defense-in-depth. Layout therefore re-flows without label geometry — the whole point of LBL-01.
- **What survives labels-off:** unit shapes (ShapeForType), colour/style semantics, cluster structure (IDs, nesting, wrapper chains), ExploreURL/ReferenceURL attributes, and the legend. What goes: node/edge/cluster label text and the 🔍/📖 glyphs (label content).
- **Converter signature migration:** `plain bool` → `graph.RenderOpts` at the label sites — NoLabels needed the context anyway, and 38-02's "minimal threading" note is superseded where labels became behaviour-relevant.
- **Test discriminator:** element HTML labels are the only LOWERCASE `<table` source; the sanctioned graph/legend markup stays UPPERCASE (phase-37 pattern). E2E additionally pins `label=""` emission and a `label=<` count of exactly 2 (graph label + legend).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Test] noLabelsModel self-expanded "app" fixed to a separate expandable box**
- **Found during:** Task 2
- **Issue:** TestNoLabelsNodesAreBareShapes looked for node "app" at top level, but `Expanded: ["app"]` correctly renders it as a cluster, not a node
- **Fix:** model now has a plain referenced system "app" plus a separate expanded box "svc" containing "api"; the default-path lock's glyph expectation updated to "Application 📖"
- **Files modified:** internal/graph/builder_test.go
- **Committed in:** 4936563

**2. [Rule 1 - Test] markup assertions moved off lowercased DOT text**
- **Found during:** Task 2
- **Issue:** the drill-down's nav breadcrumb legitimately carries "Order Context" (navigation is metadata, not an element label), and lowercasing the DOT before `<table` checks also matched the sanctioned UPPERCASE graph/legend markup after ToLower
- **Fix:** "Order Context" absence asserted on the C1 file only; `<table` checks run on the RAW dot (lowercase = element markup); added `label=""` presence pins
- **Files modified:** cmd/c4drill/root_test.go
- **Committed in:** 4936563

---

**Total deviations:** 2 auto-fixed (2 Rule 1, test-side only — implementation matched the plan as written)
**Impact on plan:** None on behaviour; test precision fixes only. No testdata touched; no new goldens (additive nolabels goldens remain deferred to 38-04).

## Issues Encountered

- One transient full-suite failure of `TestIntegrationC1EdgeResolution` (internal/view) during Task 3's first calibration run; it passes alone and in 3/3 subsequent full-suite runs, and this plan adds no logic to internal/view beyond a struct field — treated as an unrelated parallel-scheduling flake, not a regression.

## Expected-Failure Census for 38-04

**Census remains EMPTY after this plan:** `go test -count=1 ./...` fully green (verified ×3). The plan's note that cmd golden E2E tests "may fail pending 38-04" did NOT materialize — --no-labels is opt-in and every golden/comparison asserts default-path output. 38-04's scope stays reduced to (optional) NEW positive goldens for granular/nolabels outputs plus the visual spot-check.

## Next Phase Readiness

- LBL-01..03 complete; KEY-01/KEY-02 (38-02) and WRAP-01..03 (38-01) intact — all guard tests green.
- 38-05 documentation plan should document: legend stays under `--no-labels` (properties.legend governs), and the flag's composition semantics (union with --plain and the four granular switches).
- Pre-existing gofmt findings in internal/graph/integration_test.go and internal/graph/shapes.go remain untouched (out of scope).

## TDD Gate Compliance

- `test(38-03)` RED commit: bdff1c0 (compile failure on missing View.NoLabels/RenderOpts.NoLabels + unknown-flag runtime failures)
- `feat(38-03)` GREEN commit: 4936563 (all NoLabels tests + 38-02 guards green)
- Task 3 was calibration-only (no code delta → no commit). No refactor commit needed.

## Self-Check: PASSED

- Files exist: internal/view/view.go, internal/graph/graph.go, internal/graph/builder.go, internal/render/converter.go, cmd/c4drill/root.go — FOUND
- Commits: bdff1c0, 4936563 — FOUND in `git log`
- `go test -count=1 ./...` — all packages ok (×3 consecutive runs); `go vet ./...` clean; gofmt clean on all touched files
- `git status` shows no changes under cmd/c4drill/testdata/
