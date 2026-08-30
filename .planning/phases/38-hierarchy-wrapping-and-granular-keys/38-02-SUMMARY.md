---
phase: 38-hierarchy-wrapping-and-granular-keys
plan: 02
subsystem: graph-builder
tags: [cli-flags, granular-suppression, plain, tdd, render-opts]
requires:
  - phase: 38-01
    provides: "post-WRAP builder (wrapper clusters) and an EMPTY expected-failure census — full suite green"
  - phase: 37-nesting-context-and-plain-rendering
    provides: "--plain plumbing precedent (View.Plain → Graph → guards) the granular switches mirror"
provides:
  - "RenderOpts struct (Plain, NoColors, NoStyles, NoLength, NoRank) on Graph, threaded from CLI to builder"
  - "--no-colors/--no-styles/--no-length/--no-rank cobra flags on every view (processView + processExpandedView)"
  - "KEY-02 union lock: TestPlainUnionParity + TestPlainImpliesAllAspects (plain ≡ plain + all four switches)"
affects: [38-04-rebaseline]
tech-stack:
  added: []
  patterns:
    - "RenderOpts suppression-switch struct threaded through builder signatures (replaces scattered plain bool)"
    - "granular guards defer to plain on kind colouring (plain retains semantic colours → union holds)"
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
  - "KEY-02 tension resolution: kind-derived edge colours are semantic and survive --plain (pinned by v1.14 plain goldens), so --no-colors suppresses kind colouring only when plain is NOT set — this makes plain ≡ plain + switches exactly"
  - "D-01 source-border default edge colour is structural, not decoration — --no-colors retains it (otherwise union parity would break)"
  - "converter.go keeps its internal plain bool signatures; only buildCgraph's entry now passes g.Opts.Plain — granular switches never affect labels, so full RenderOpts migration there would be churn without behaviour"
  - "no-styles E2E uses a dedicated inline fixture: plain.toml's author styles are not DOT-observable (pair collapse aggregates dashed→solid; node style emission is fill-driven)"
requirements-completed: [KEY-01, KEY-02]
metrics:
  duration: ~30m
  completed: 2026-08-30
---

# Phase 38 Plan 02: Granular Suppression Switches Summary

Four granular CLI switches (`--no-colors/--no-styles/--no-length/--no-rank`) each suppress exactly one formatting aspect via a threaded `RenderOpts` struct, with `--plain` locked as their exact union by a canonical E2E parity test — full suite green, zero golden re-baseline.

## Performance

- **Duration:** ~30 min
- **Started:** 2026-08-30T15:35:28Z
- **Completed:** 2026-08-30
- **Tasks:** 3
- **Files modified:** 8

## Accomplishments

- `RenderOpts{Plain, NoColors, NoStyles, NoLength, NoRank}` replaces the lone `Graph.Plain`; copied in all three constructors (BuildGraph, BuildExpandedGraph via BuildGraphWithPath); builder signatures (buildNode, buildCluster, createEdge, applyUnitOverrides, applyCollapsedPairStyle) migrated from `plain bool` to `RenderOpts`.
- Guard mapping per pin: NoColors → unit Color/Border fills skipped + author link colour skipped + kindColour suppressed (in createEdge AND applyCollapsedPairStyle); NoStyles → Style/BorderStyle + author link style + collapsed-pair aggregate style override; NoLength → MinLen 0; NoRank → RankReverse/NoConstraint false, automatically neutralizing all three converter emission sites (no converter guards added, asserted instead).
- Four cobra flags registered after `--plain` and threaded onto every view in both processView and processExpandedView. `properties.edges` stays tied to Plain only.
- KEY-02 locked twice: TestPlainImpliesAllAspects (builder level) and TestPlainUnionParity (E2E canonical DOT equality of `--plain` vs `--plain` + all four flags over plain.toml).
- TestSwitchCombination: pairwise + all-four combos over a post-WRAP C3 boundary view suppress exactly the union of their aspects, wrapper-cluster nesting intact.

## Task Commits

1. **Task 1 (RED): per-switch guard tests + plain-parity lock** - `11556ae` (test)
2. **Task 2 (GREEN): RenderOpts struct + flag threading + guards** - `0e6bd24` (feat)
3. **Task 3: combination smoke — switches with each other and WRAP output** - `17aed63` (test)

## Files Created/Modified

- `internal/graph/graph.go` — RenderOpts struct; Graph.Opts replaces Graph.Plain
- `internal/view/view.go` — View.NoColors/NoStyles/NoLength/NoRank beside Plain
- `internal/graph/builder.go` — renderOptsFromView + guard conditions at all suppression sites
- `internal/graph/builder_test.go` — guard tests, union lock, combination lock; existing `g.Plain` assertions migrated
- `internal/render/converter.go` — buildCgraph passes `g.Opts.Plain`
- `internal/render/converter_test.go` — Graph literal fields migrated to `Opts: graph.RenderOpts{...}`
- `cmd/c4drill/root.go` — four flag vars + registrations + per-view assignment
- `cmd/c4drill/root_test.go` — TestPlainUnionParity + TestGranularFlagsE2E

## Decisions Made

- **Kind colours vs union parity:** plain retains semantic kind colouring (v1.14 contract, pinned by the plain.golden asserting `#2E7D32` survives). If `--no-colors` suppressed kind colouring unconditionally, `--plain --no-colors` would differ from `--plain` and break KEY-02. Resolution: kind-colour suppression is active only when `NoColors && !Plain`. Both the goldens and TestPlainUnionParity are green, proving the resolution.
- **Source-border edge colour survives `--no-colors`:** the D-01 default is structural identity, not author decoration; suppressing it would also break union parity (plain keeps it).
- **Converter migration kept minimal:** granular switches never touch labels, so only `buildCgraph`'s entry reads `g.Opts.Plain`; the internal `plain bool` call tree is unchanged.
- **properties.edges remains Plain-only** (documented plan pin — routing is not one of the four aspects).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Test fixture] TestNoStyles E2E switched from plain.toml to a dedicated inline fixture**
- **Found during:** Task 2
- **Issue:** plain.toml's author styles are not observable in DOT output — the admin↔orders.api pair collapses two links (dashed + solid aggregate to solid per AGG-02) and node style emission is fill-colour-driven ("filled" comes from the author colour, not style)
- **Fix:** no-styles subtest renders a small inline TOML where `style="dashed"`/`style="dotted"` reach DOT verbatim; asserts suppression while colour `#1565C0` stays
- **Files modified:** cmd/c4drill/root_test.go
- **Verification:** TestGranularFlagsE2E green
- **Committed in:** 0e6bd24

**2. [Rule 1 - Compile] Existing tests migrated to the new Graph.Opts field**
- **Found during:** Task 2
- **Issue:** `Graph.Plain` removal broke `g.Plain`/`Plain:` literals in builder_test.go and converter_test.go
- **Fix:** mechanical migration to `g.Opts.Plain` / `Opts: graph.RenderOpts{Plain: true}`; gofmt applied
- **Files modified:** internal/graph/builder_test.go, internal/render/converter_test.go
- **Committed in:** 0e6bd24

---

**Total deviations:** 2 auto-fixed (2 Rule 1)
**Impact on plan:** Both necessary for the new API and DOT-observable assertions. No scope creep; no testdata files touched.

## Issues Encountered

- None beyond the deviations above. Committed plain goldens (plain.dot, plain.expanded.dot) stayed green — consistent with 38-01's empty census; nothing re-baselined.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- KEY-01/KEY-02 complete; 38-04 can proceed: census remains EMPTY, so its re-baseline scope stays reduced to (optional) new positive goldens for granular-flag outputs plus the visual spot-check.
- Pre-existing gofmt findings in internal/graph/integration_test.go and internal/graph/shapes.go are untouched, out-of-scope files.

## TDD Gate Compliance

- `test(38-02)` RED commit: 11556ae (compile failure on missing View fields + unknown-flag runtime failures)
- `feat(38-02)` GREEN commit: 0e6bd24 (all guard tests green)
- Task 3 lock test commit: 17aed63. No refactor commit needed.

## Self-Check: PASSED

- Files exist: internal/graph/graph.go, internal/view/view.go, internal/graph/builder.go, cmd/c4drill/root.go — FOUND
- Commits: 11556ae, 0e6bd24, 17aed63 — FOUND in `git log`
- `go test ./... -count=1` — all packages ok; `go vet ./...` clean; gofmt clean on all touched files
- `git status` shows no changes under cmd/c4drill/testdata/
