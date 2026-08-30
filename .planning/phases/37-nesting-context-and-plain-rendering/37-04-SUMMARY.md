---
phase: 37-nesting-context-and-plain-rendering
plan: 04
subsystem: render
tags: [plain, render, labels, e2e, go, tdd]

# Dependency graph
requires:
  - phase: 37-nesting-context-and-plain-rendering
    plan: 03
    provides: Graph.Plain threaded from the --plain cobra flag through View.Plain into BOTH graph entry points, plus builder-level suppression of author color/style/border/length/rank/labelPosition/edges — the flag and neutralized data this plan's render-side label switches consume
    plan2: 01
    provides: recursive clusters + UnfoldChain — the drill-down cluster paths setClusterLabel's URL-preserving plain branch must cover
provides:
  - "PLAIN-03: render-side plain label branches — setNodeLabel/setClusterLabel emit buildRecordLabel via SetLabel (no HTML-table dispatch), createEdge emits buildEdgePlainTextLabel via SetLabel (never the HTML rectangle); name/technology/description content preserved at all three sites"
  - "Cluster drill-down URL emission (SafeSet URL) survives the plain branch — structural affordance held"
  - "PLAIN-04: E2E proof that --plain applies uniformly to EVERY generated file — C1 + drill-downs in dot, svg, html, including --plain --expanded (Pitfall 5)"
  - "NEW committed goldens cmd/c4drill/testdata/plain.dot + plain.expanded.dot (canonical-compared); styled fixture plain.toml; zero existing goldens touched"
  - "Tests: TestConverter_PlainNodeLabelsArePlainText, TestConverter_PlainClusterLabelsArePlainText, TestConverter_PlainEdgeLabelsArePlainText, TestConverter_NonPlainLabelsUnchanged, TestConverter_PlainKeepsLegendAndKindColour, TestPlainFlagC1Golden, TestPlainFlagExpandedGolden, TestPlainFlagAppliesToAllGeneratedFiles, TestPlainFlagAllFormats, TestPlainFlagOptIn"
affects: [37-05 (BC-01 flat-model stability + scan the goldens this plan adds), docs/skill sync]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Plain label routing via a plain bool threaded through the converter call tree (buildCgraph -> createTopLevelNodes/createCluster/createEdges -> createNode/setNodeLabel/createEdge/setClusterLabel) — same flag-on-struct philosophy, no globals"
    - "Plain edge-label text format pinned: '[Technology] Description' via buildEdgePlainTextLabel in converter.go (labels.go deliberately untouched)"
    - "E2E HTML-marker discriminator: unit/edge HTML labels are the only LOWERCASE html markup in DOT; nav/title graph label + legend carry UPPERCASE tables and stay under plain — label=< count == 2 per dot file"

key-files:
  created:
    - cmd/c4drill/testdata/plain.toml
    - cmd/c4drill/testdata/plain.dot
    - cmd/c4drill/testdata/plain.expanded.dot
  modified:
    - internal/render/converter.go
    - internal/render/converter_test.go
    - cmd/c4drill/root_test.go

key-decisions:
  - "Plain node/cluster labels emit buildRecordLabel via SetLabel (true plain text, graphviz-escaped) instead of keeping the SetLabelHTML record path — the plan's interfaces block pinned this choice and the new goldens lock it (T-37-07: no new string concatenation into DOT, SetLabel escapes author text)"
  - "setClusterLabel restructured so the CTX-03 drill-down URL emission (SafeSet URL) runs AFTER the plain/HTML label branch — URL survives plain mode; default path keeps the early return on empty HTML label (no URL) exactly as before"
  - "Plain edge-label text = '[Technology] Description' (buildEdgePlainTextLabel, converter.go) — mirrors the content the HTML rectangle carried; labels.go NOT modified per plan"
  - "E2E '<TABLE absent' plan assertion interpreted as the lowercase label-formatting markers (<table/<b>/<i>) because the legend and nav/title legitimately keep UPPERCASE HTML tables under plain (locked boundary); label=< count == 2 pins the sanctioned pair"

patterns-established:
  - "Plain-mode label contract: FORMATTING simplifies (record text, no HTML), CONTENT and structural affordances (URLs, kind colours, legend) are preserved"

requirements-completed: [PLAIN-03, PLAIN-04]

# Metrics
duration: 18min
completed: 2026-08-30
---

# Phase 37 Plan 04: Plain Label Rendering + E2E Plain Goldens (PLAIN-03/PLAIN-04) Summary

**Under `--plain` the converter now emits true plain-text labels — record-path node/cluster labels and `[Technology] Description` edge labels via SetLabel, with the HTML-table/rectangle paths demoted to default mode only — proven uniform across the context diagram, all drill-downs, and svg/html/dot (including `--plain --expanded`) by two NEW committed goldens and a styled fixture.**

## Performance

- **Duration:** 18 min
- **Started:** 2026-08-30T13:38:37Z
- **Completed:** 2026-08-30T13:56:54Z
- **Tasks:** 3 (RED / GREEN / E2E)
- **Files modified:** 6 (3 created testdata, 3 modified source/test)

## Accomplishments
- PLAIN-03 at the render layer: `g.Plain` threaded through `buildCgraph` into `createTopLevelNodes`/`createCluster`/`createEdges`/`createNode`/`createEdge`; `setNodeLabel`/`setClusterLabel` route to `buildRecordLabel` via `SetLabel` under plain (skip the type-specific HTML dispatch entirely); edge labels emit `buildEdgePlainTextLabel` via `SetLabel`, never `SetLabelHTML`; empty-label fallback behavior preserved
- Cluster drill-down URL emission preserved under plain — the SafeSet("URL") block moved after the plain/HTML branch so the structural affordance (CTX-03) survives while the default path's early-return stays byte-identical
- PLAIN-04 proven at E2E: `TestPlainFlagAppliesToAllGeneratedFiles` walks EVERY generated .dot (C1 + orders drill-down) asserting author hexes/minlen/dir=back/constraint=false/splines=false/lowercase-HTML absent while legend, kind colour `#2E7D32`, and plain-text label content are present; `TestPlainFlagAllFormats` covers svg + html existence/non-emptiness; `--plain --expanded` locked by its own golden (Pitfall 5)
- Opt-in locked from both directions: `TestConverter_NonPlainLabelsUnchanged` (unit) and `TestPlainFlagOptIn` (E2E — author colour, HTML labels, minlen all still emitted without the flag)
- Full uncached suite green: 15/15 packages, zero failures — **expected-failure list remains EMPTY** for 37-05; no pre-existing golden modified (git status shows only the three new testdata files)

## Expected-Failure List (LOAD-BEARING for 37-05)

**EMPTY — zero failing tests in the repo after this plan.** Uncached `go test -count=1 ./...`: 15/15 packages ok, 0 failures, all pre-existing committed goldens untouched (`git status` shows only new `testdata/plain{,.expanded,.toml}` files). 37-05's re-baseline scope: BC-01 flat-model stability verification + scanning the goldens this plan adds (`plain.dot`, `plain.expanded.dot`); nothing to re-baseline.

## Task Commits

1. **Task 1 (RED): plain label unit tests** - `1c0030a` (test) — Tests 1-3 fail because plain graphs still emit HTML labels; Test 4 (default-path lock) and Test 5 (legend + kind colour survive) pass by design, pinning the boundary RED-side
2. **Task 2 (GREEN): converter plain label branches** - `0cf7987` (feat) — all 5 unit tests pass; full render package green; vet/gofmt clean
3. **Task 3 (E2E): styled fixture + new goldens + uniformity tests** - `fd46451` (feat) — 5 E2E tests pass; full repo suite green

_Note: TDD plan — test → feat commits; no refactor commit needed (implementation landed clean: gofmt/vet clean, labels.go untouched as planned)._

## Files Created/Modified
- `internal/render/converter.go` — plain bool threaded through the call tree; plain branches in setNodeLabel/setClusterLabel (buildRecordLabel via SetLabel) and createEdge (buildEdgePlainTextLabel via SetLabel); new buildEdgePlainTextLabel helper; URL emission restructured to survive plain
- `internal/render/converter_test.go` — five plain label unit tests (three RED-verified, two boundary locks)
- `cmd/c4drill/root_test.go` — generatePlainFixtureOutput/collectGeneratedFiles helpers + five E2E tests (goldens, uniformity, formats, opt-in)
- `cmd/c4drill/testdata/plain.toml` — styled fixture: person with color/border/style, expanded box with nested subunit, links carrying color/style/length=3/rank=reverse/labelPosition=tail and rank=equal/kind=read, properties.edges=straight
- `cmd/c4drill/testdata/plain.dot` — NEW committed `--plain` C1 golden (canonical-compared)
- `cmd/c4drill/testdata/plain.expanded.dot` — NEW committed `--plain --expanded` golden (Pitfall 5)

## Decisions Made
- Plain labels use `SetLabel` (true plain text) rather than keeping `SetLabelHTML` with record content — the plan's interfaces section pinned this decision ("the decision is SetLabel — this plan pins that choice and the new goldens lock it"); graphviz escaping covers T-37-07
- The fixture's expanded unit is a top-level `box` (04-styling precedent) rather than a `container`: the validator rejects non-C1 types at top level ("unit orders has type container which is not allowed at top level"); box + nested system exercises the same expanded-cluster-with-subunit surface
- Pair-aggregation safety checked before fixture design: pair keys are directional (`path->peer`), so the fixture's two opposite-direction links (admin->orders.api, orders.api->admin) never aggregate — the kind colour survives per-edge in every view
- `penwidth=2` in plain.expanded.dot is COMPAT-02 expanded-mode semantics (structural), consistent with 37-03's decision not to suppress PenWidth

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] E2E "<TABLE absent" assertion contradicted the legend-stays boundary**
- **Found during:** Task 3 (E2E)
- **Issue:** The plan's uniformity behavior asks every dot file to assert "`<TABLE` absent" while the same must_haves keep the legend (an HTML table via SetLabelHTML) and the nav/title graph label under plain — a literal assertion can never pass
- **Fix:** Asserted the precise discriminators instead: lowercase `<table`/`<b>`/`<i>` absent (only unit/edge HTML labels emit lowercase markup — verified navigation.go/legend use uppercase) plus `label=<` count == 2 per file (graph label + legend, the sanctioned pair). Label-level table absence on no-title/no-legend graphs remains pinned exactly by the Task 1 unit tests
- **Files modified:** cmd/c4drill/root_test.go (assertions only)
- **Verification:** TestPlainFlagAppliesToAllGeneratedFiles passes on both generated dot files; catches any future HTML-label leak
- **Committed in:** fd46451

---

**Total deviations:** 1 auto-fixed (bug in planned assertion, test-only). **Impact on plan:** none on shipped behavior — the interpretation follows the plan's own must_have truth that legend/nav stay under plain; production code matches the plan exactly.

## Issues Encountered
- First fixture draft used `type = "container"` at top level and failed validation (C1 types only) — resolved by switching to `box` + nested `system` (see Decisions). No other issues; goldens generated on the first CLI run after the fix.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- 37-05 (final plan: BC-01 flat-model stability + golden scan + docs/skill sync + REL-01 tag): suite fully green with EMPTY expected-failure list; the only new goldens to scan are `cmd/c4drill/testdata/plain.dot` and `plain.expanded.dot`
- Plain surface is now complete end to end (flag -> view -> builder -> render labels); default path locked by unit + E2E opt-in tests
- Threat mitigations landed: T-37-07 (SetLabel escaping path, asserted by content-preservation tests), T-37-08/T-37-SC N/A by design

---
## Self-Check: PASSED

- All created files exist: `cmd/c4drill/testdata/plain.toml`, `plain.dot`, `plain.expanded.dot` verified on disk
- Commits verified in `git log`: `1c0030a` (test), `0cf7987` (feat), `fd46451` (feat)
- Plan-level `<verification>` re-run: `go test ./internal/render/ -run 'Plain' -v` → 5/5 PASS; `go test ./internal/render/ -count=1` → ok; `go test ./cmd/c4drill/ -run 'TestPlainFlag' -v` → 5/5 PASS; `go test -count=1 ./...` → 15/15 packages ok
- No deletions in any commit (`git diff --diff-filter=D` empty); pre-existing goldens untouched

---
*Phase: 37-nesting-context-and-plain-rendering*
*Completed: 2026-08-30*
