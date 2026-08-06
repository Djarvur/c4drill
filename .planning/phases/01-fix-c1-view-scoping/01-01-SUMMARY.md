---
phase: 01-fix-c1-view-scoping
plan: 01
subsystem: graph
tags: [graph, edges, dedup, penwidth, dot, c4]

# Dependency graph
requires:
  - phase: v1.8 milestone (f9dc69a/c5091f9/8f7e21c/b6771b0)
    provides: resolved-link override seam (Entry.ResolvedLinks), buildEdges/createEdge/markSeen, GenerateExpandedView
provides:
  - Pair-only edge dedup key for resolved views (C1/C2/C3) with first-wins attributes
  - Binary penwidth: collapsed pairs 2.0, single edges 1.0 (renderer default), --expanded keeps 2.0
  - View.AllExpanded mode discriminator (D-02) + Edge.PenWidth field
  - countPairMultiplicity mirror-aware per-pair counter (D-05)
affects: [01-02 (minlen gating + GenerateExpandedView sets AllExpanded), 01-03 (visible-subunit resolution), verifier]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Mode discriminator on the view (AllExpanded) instead of threading parameters into buildEdges"
    - "Per-pair count map computed before the dedup loop (countPairMultiplicity)"
    - "DOT assertions must target per-edge attribute blocks — go-graphviz always emits a penwidth=1.0 default edge block and wraps edge attrs across lines"

key-files:
  created: []
  modified:
    - internal/view/view.go
    - internal/graph/graph.go
    - internal/graph/builder.go
    - internal/render/converter.go
    - internal/graph/builder_test.go

key-decisions:
  - "Pair-only dedup key (source->target) in resolved views; --expanded keeps the v1.7 tech+desc key via View.AllExpanded (D-01/D-02)"
  - "Penwidth is carried on graph.Edge.PenWidth (0 = renderer default 1.0); converter sets PenWidth>0 as-is else 1.0 (D-04, Pitfall 3)"
  - "countPairMultiplicity counts outgoing links per pair, then incoming links only for pairs without outgoing contributions — validator LinksFrom mirrors are not double-counted (D-05)"
  - "go-graphviz DOT output always includes a default edge block with penwidth=1.0 and wraps edge attributes across multiple lines — render assertions must extract the per-edge block"

patterns-established:
  - "countPairMultiplicity: two-pass counter (outgoing, then mirror-aware incoming) computed in buildEdges before the edge loop"
  - "Edge test fixture: literal view.View construction with FullPath set so renderer node lookup finds endpoints"

requirements-completed: [EDGE-02]

# Metrics
duration: 22min
completed: 2026-08-06
---

# Phase 1 Plan 1: Pair-Only Edge Collapse with Binary Penwidth Summary

**Pair-only edge dedup for resolved C1/C2/C3 views with first-wins attributes (D-01/D-03/D-06), binary penwidth 1.0/2.0 via new Edge.PenWidth (D-04), mirror-aware multiplicity counting (D-05), and a View.AllExpanded discriminator keeping --expanded byte-compatible with v1.7 (D-02/COMPAT-02)**

## Performance

- **Duration:** 22 min
- **Started:** 2026-08-06T09:05:38Z
- **Completed:** 2026-08-06T09:10:23Z
- **Tasks:** 3 (TDD: RED + GREEN + regression)
- **Files modified:** 5

## Accomplishments

- Pair-only edge collapse: multiple links on the same (source, target) pair now produce ONE edge in resolved views, carrying the first contributing link's technology/description/color/style/length (first-wins via definition-order iteration + markSeen).
- Binary penwidth: collapsed pairs (2+ contributing links) render at 2.0, single-relationship edges at the 1.0 default; `--expanded` edges keep the v1.7 2.0 prominence — converter.go:478 unconditional `SetPenWidth(2.0)` replaced with an `edge.PenWidth > 0` conditional.
- New `View.AllExpanded` flag (D-02 mode discriminator, consumed by buildEdges — set by GenerateExpandedView in plan 01-02) and `Edge.PenWidth` field (0 = renderer default).
- `countPairMultiplicity` counts contributing links per pair WITHOUT double-counting validator LinksFrom mirrors (index.go:53-80 mirrors every outgoing link into the target's LinksFrom).
- Four new RED tests (TestBuildEdgesPairCollapse, TestBuildEdgesPenwidth, TestBuildEdgesExpandedExemption, TestBuildEdgesPenwidthRendered) cover D-01..D-06/COMPAT-02, including DOT-level penwidth assertions.
- Expanded-mode regression green: TestBuildExpandedGraphRealToml still renders `minlen=N` for the real TOML (the saira fixture has zero duplicate-pair links, so the interim pair-key state is inert until 01-02 activates the v1.7 key via AllExpanded).

## Task Commits

Each task/gate was committed atomically:

1. **Task 1 (RED): skeleton fields + failing tests** - `0cd7991` (test)
2. **Task 2 (GREEN): pair-only dedup key, multiplicity counting, penwidth** - `d0af213` (feat)
3. **Task 3 (regression): lint fixes on new code** - `3adf383` (refactor)

**Plan metadata:** (final docs commit follows this summary)

## Files Created/Modified

- `internal/view/view.go` - Added `View.AllExpanded bool` (D-02 discriminator, doc comment per CONVENTIONS)
- `internal/graph/graph.go` - Added `Edge.PenWidth float64` (0 = renderer default)
- `internal/graph/builder.go` - `countPairMultiplicity`/`countOutgoingPairs`/`countIncomingPairs` helpers; `buildEdges` computes pairCounts before the edge loop; `processOutgoingLinks`/`processIncomingLinks` take `v *view.View` + pairCounts, pair-only key unless AllExpanded, penWidth 2.0 for AllExpanded or count>=2; `createEdge` gained a `penWidth` param setting `Edge.PenWidth`
- `internal/render/converter.go` - Penwidth conditional: `PenWidth > 0` as-is, else 1.0 (Pitfall 3)
- `internal/graph/builder_test.go` - Four new test functions + `edgeBlockFromDOT` helper (multi-line edge attribute block extraction)

## Decisions Made

- Followed the plan exactly: pair-only key for resolved views, v1.7 key for `--expanded` (D-02), binary penwidth (D-04), mirror-aware counting (D-05), plain labels (D-06), MinLen copy untouched (D-13 gating lands in plan 01-02).
- Literal expanded-mode view fixtures set `FullPath` on entries — without it the renderer's nodeMap lookup (converter.go createEdges skips edges with missing endpoints) would drop the edges and the render assertions would be vacuous.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Test correctness] DOT penwidth assertion was a false positive**
- **Found during:** Task 1 (RED verification)
- **Issue:** `assert.Contains(dot, "penwidth=1")` passed BEFORE implementation because go-graphviz always emits a default edge block `edge [..., penwidth=1.0, ...]` regardless of per-edge values — the test was not testing what it claimed (TDD fail-fast).
- **Fix:** Assertions now target the per-edge attribute block via a new `edgeBlockFromDOT` helper; discovered the DOT writer also wraps edge attributes across multiple lines, so the helper collects lines from the edge's first line through the closing `];`.
- **Files modified:** internal/graph/builder_test.go
- **Verification:** At RED the per-edge block contains `penwidth=2` (unconditional) so "penwidth=1" fails; at GREEN it contains `penwidth=1`. Edge-block assertions confirmed via raw DOT dumps.
- **Committed in:** 0cd7991 (part of RED commit, before GREEN)

**2. [Rule 3 - Blocking] `mise lint` auto-fix polluted the working tree**
- **Found during:** Task 3 (lint gate)
- **Issue:** The `mise lint` task runs `golangci-lint run --fix` (lint-fix) first, which auto-modified 12 files — including 9 files outside this plan's scope (path.go, shapes.go, parser.go, errors.go, labels_test.go, wrap.go, html_labels_internal_test.go, integration_test.go, scope.go) and pre-existing code regions in plan files.
- **Fix:** Reverted all lint-fix changes per the scope boundary (pre-existing issues are out of scope), then fixed ONLY the issues in new code: extracted countOutgoingPairs/countIncomingPairs (gocognit), assert.InDelta (testifylint float-compare), FailNowf (testifylint formatter), strings.Builder (modernize), wsl whitespace, funlen nolint per convention.
- **Files modified:** internal/graph/builder.go, internal/graph/builder_test.go
- **Verification:** `golangci-lint run --new-from-rev=HEAD~2 ./...` reports 0 new issues; full suite green; pre-existing count unchanged (125 at base vs 121 now — net -4).
- **Committed in:** 3adf383

---

**Total deviations:** 2 auto-fixed (1 test-correctness, 1 lint-gate)
**Impact on plan:** Both were necessary for the plan's tests and lint gate to be meaningful. No scope creep.

## Issues Encountered

- go-graphviz (onokonem fork) `SetPenWidth` uses `fmt.Sprint`, so 1.0/2.0 serialize as `penwidth=1`/`penwidth=2` — the DOT assertions rely on this.
- The repo carries pre-existing golangci-lint debt (125 issues at plan base: goconst 50, gocognit 10, wsl_v5 16, etc.). Left untouched per scope boundary; logged to deferred-items.md.

## Deferred Items

Logged to `.planning/phases/01-fix-c1-view-scoping/deferred-items.md`:
- Pre-existing golangci-lint debt (121 issues: goconst, gocognit, wsl_v5, mnd, nestif, lll, etc.) across ~20 files — not caused by this plan; `mise lint` will keep failing until a dedicated cleanup plan addresses it.
- `View.AllExpanded` is not yet set by `GenerateExpandedView` (plan 01-02) — until then, expanded views run with the pair-only dedup key (interim state; inert for the saira fixture which has no duplicate-pair links).

## Next Phase Readiness

- Plan 01-02 can now set `v.AllExpanded = true` in `GenerateExpandedView` (re-activating the v1.7 tech+desc key and 2.0 penwidth for expanded mode = COMPAT-02) and implement D-13 minlen gating with full knowledge of the buildEdges signature (v, path, links, seen, pairCounts).
- Plan 01-03 (visible-subunit resolution) builds on the unchanged `isTargetInView` gate — its signature still takes `v.Units`, so visible-subunit paths must land in `v.Units` per the plan.

## Self-Check: PASSED

- Files: 01-01-SUMMARY.md, deferred-items.md exist
- Commits: 0cd7991 (test RED), d0af213 (feat GREEN), 3adf383 (refactor) all present
- `go test -v -race ./...` green; `golangci-lint run --new-from-rev=HEAD~2 ./...` reports 0 new issues

---
*Phase: 01-fix-c1-view-scoping*
*Completed: 2026-08-06*
