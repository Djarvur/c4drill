---
status: complete
phase: 260828-tgf-fix-pipe-end-cap-right-side-must-render-
plan: 01
subsystem: render
tags: [svg, graphviz, queue, pipe, path-geometry, post-processing]

# Dependency graph
requires:
  - phase: 260828-qbx (quick task)
    provides: queue pipe SVG post-processor (internal/render/pipe.go) whose single-subpath d ended in a degenerate coincident-endpoint arc
provides:
  - Two-subpath pipe d: closed clockwise body outline + right cap-face subpath forming a visible full ellipse
  - Regression tests proving no coincident-endpoint arc can ever be emitted and the cap face subpath exists
  - Re-rendered example SVGs (10-edge-kinds, 06-templates family, cloud-system expanded + amazon/sqs) and synced plugin skill copies
affects: [future queue-pipe visual work, skill/plugin example parity]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Two-subpath SVG path in ONE d attribute so copied paint attributes (fill/stroke/dasharray) apply to both subpaths — no structural change to the polygon replacement"

key-files:
  created: []
  modified:
    - internal/render/pipe.go
    - internal/render/pipe_internal_test.go
    - skill/examples/10-edge-kinds.svg
    - skill/examples/10-edge-kinds/app.svg
    - skill/examples/06-templates.svg
    - skill/examples/06-templates.expanded.svg
    - skill/examples/06-templates/platform.svg
    - skill/examples/06-templates/platform/users.svg
    - examples/cloud-system/cloud-system.expanded.svg
    - examples/cloud-system/cloud-system/amazon/sqs.svg
    - plugins/c4drill/skills/c4drill-toml/examples/ (6 mirrored SVGs)
    - plugins/c4drill/opencode/skills/c4drill-toml/examples/ (6 mirrored SVGs)

key-decisions:
  - "Cap face is a second subpath in the SAME d attribute (not a second element) so copied polygon paint attrs apply to it; fill is unaffected (body outline already covers the silhouette) and the stroked face arc is what makes the ellipse visible"
  - "GraphViz edge-group IDs vary per process run (verified with a pre-task binary): all edge-ID churn in re-rendered SVGs was restored to the committed arrangement so the diff is purely pipe geometry (amazon.svg needed no pipe change and stayed byte-identical)"

patterns-established:
  - "Empirical sweep-flag verification: rasterize with qlmanage and visually confirm arc directions before locking test expectations"

requirements-completed: [QUICK-PIPE-CAP-01]

# Metrics
duration: 13min
completed: 2026-08-28
---

# Quick Task 260828-tgf: Fix Pipe End Cap — Right Side Must Render Summary

**Queue pipes now end in a full ellipse: the d attribute carries a closed clockwise body outline (right outer arc sweep 1 through (x1,cy)) plus a sweep-0 cap-face subpath replacing the old degenerate coincident-endpoint arc that SVG renderers omitted**

## Performance

- **Duration:** 13 min
- **Started:** 2026-08-28T18:18:36Z
- **Completed:** 2026-08-28T18:31:04Z
- **Tasks:** 2
- **Files modified:** 22 (2 Go files, 8 tracked example SVGs, 12 plugin-mirrored SVGs)

## Accomplishments
- Root cause fixed: the old trailing "full ellipse" arc had coincident start/end points (`A rx,ry 0 1,1 bodyR,y0` right after arriving at bodyR,y0) — SVG omits such arcs, so the right cap drew nothing and rendered as a capsule side
- `pipePathFmt`/`pipePathFromBBox` now emit two subpaths in one d attribute: `M bodyL,y0 L bodyR,y0 A 0 0,1 bodyR,y1 L bodyL,y1 A 0 0,1 bodyL,y0 Z M bodyR,y0 A 0 0,0 bodyR,y1` — the pipe still fills the exact polygon bbox (left cap reaches x0, outer arc reaches x1), so edge anchors stay valid
- Regression tests: `TestPipePathFromPoints_NoCoincidentArc` (pen-point walk proving no arc ever has coincident endpoints) and `TestPipePathFromPoints_CapFaceSubpath` (second M after Z, sweep 0, (bodyR,y0)→(bodyR,y1)); full-string d equality pins the geometry
- Empirically verified via qlmanage raster of 10-edge-kinds.svg: full-ellipse right cap, plain half-bulge left end, db cylinders/person/box/legend untouched
- All tracked example SVGs re-rendered, both plugin skill copies synced, `diff -r` parity confirmed

## Task Commits

1. **Task 1: Two-subpath pipe path with coincident-arc regression test**
   - `aa2512e` (test — RED gate: failing full-equality + regression tests)
   - `afdcbfb` (feat — GREEN gate: two-subpath pipePathFmt/pipePathFromBBox)
   - `86a395d` (test — pipeSubpaths helper fix + unused-nolint removal)
2. **Task 2: Re-render example SVGs, sync skill plugins, full gates**
   - `1e6f939` (feat — re-rendered SVGs + synced plugin copies)

## Files Created/Modified
- `internal/render/pipe.go` — two-subpath `pipePathFmt`, new Sprintf arg order, rewritten doc comments incl. the bodyR-2rx >= x0 invariant
- `internal/render/pipe_internal_test.go` — full-string d equality, NoCoincidentArc + CapFaceSubpath + ArcsBulgeOutward regressions, pipeArcs/pipeSubpaths/pipePointAt helpers, updated TestReplaceQueuePolygons assertions
- `skill/examples/10-edge-kinds.svg`, `10-edge-kinds/app.svg`, `06-templates*.svg` (4), `examples/cloud-system/cloud-system.expanded.svg`, `cloud-system/amazon/sqs.svg` — queue pipe d now two-subpath (sqs has two queue pipes)
- `plugins/c4drill/{skills,opencode}/skills/c4drill-toml/examples/...` — 12 mirrored SVGs synced via rsync

## Decisions Made
- Kept the single `fmt.Sprintf(pipePathFmt, ...)` const form (plan-preferred) — same shape as before, lint-clean
- Both subpaths share ONE `<path>` d attribute so `copiedPipeAttrs` output applies to the face too; nonzero fill rule keeps the silhouette identical, the stroked face arc provides the visible ellipse
- Restored committed edge-group IDs in re-rendered SVGs (see Deviations) instead of committing random GraphViz churn

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] pipeSubpaths test helper split at every command letter**
- **Found during:** Task 1 (GREEN verification)
- **Issue:** Splitting on every M/L/A/Z letter yielded 8 "subpaths" instead of 2; subpath boundaries are M commands
- **Fix:** Accumulate segments, start a new subpath at each M (absolute commands only, so concatenation reproduces the original d)
- **Files modified:** internal/render/pipe_internal_test.go
- **Verification:** TestPipePathFromPoints_CapFaceSubpath passes; all pipe tests green
- **Committed in:** 86a395d

**2. [Rule 1 - Bug] Unused nolint directive tripped nolintlint**
- **Found during:** Task 2 (full lint gate)
- **Issue:** `//nolint:gochecknoglobals` on the test-file regexp was unused (gochecknoglobals does not flag test vars), so nolintlint reported 1 issue
- **Fix:** Removed the directive; golangci-lint back to 0 issues
- **Files modified:** internal/render/pipe_internal_test.go
- **Verification:** `golangci-lint run ./...` → 0 issues
- **Committed in:** 86a395d (amended into the helper-fix commit)

### Plan-Documentation Corrections (not code deviations)

- **Expected-churn list was incomplete:** the 06-templates family (06-templates.svg, 06-templates.expanded.svg, platform.svg, platform/users.svg) also contains queue pipes and legitimately churned with the same transformation. `10-edge-kinds/ext.svg` has no queue and correctly did NOT change (plan listed it as expected churn).
- **GraphViz edge-ID nondeterminism (pre-existing, out of scope):** `<g id="edgeN">` numbering varies per process run — proven by rendering with a binary built from the pre-task commit (a28f1a1) in a throwaway detached worktree, which flip-flopped between orderings across runs. This task's change cannot cause it (the post-processor never touches edge groups). All edge-ID-only churn was restored to the committed arrangement; the committed diff contains only pipe `path d=` lines. amazon.svg (no queue on that page) ended byte-identical to HEAD. Worth a future hardening task if byte-stable example re-renders matter.

---

**Total deviations:** 2 auto-fixed (2 x Rule 1, both test-only) + 2 plan-documentation corrections
**Impact on plan:** Auto-fixes were test-helper correctness only; production geometry landed exactly as the locked design specified (empirical render confirmed the sweep flags needed no flips). No scope creep.

## Issues Encountered
- A pre-task binary (built to diagnose the edge-ID question) transiently overwrote amazon.svg / sqs.svg with old-code output during diagnosis; the final full re-render with the current binary plus edge-ID restoration left the tree in the intended state (verified by git status + diff inspection).

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Queue pipes render correctly end to end; no known stubs introduced
- Deferred (out of scope, pre-existing): GraphViz edge-ID nondeterminism in generated SVGs; legend and labels.go untouched per plan

---
*Phase: 260828-tgf-fix-pipe-end-cap-right-side-must-render-*
*Completed: 2026-08-28*

## Self-Check: PASSED

- All 8 spot-checked created/modified files exist on disk (pipe.go, pipe_internal_test.go, 10-edge-kinds.svg, 06-templates.svg, cloud-system.expanded.svg, amazon/sqs.svg, both plugin-mirrored 10-edge-kinds.svg)
- All 4 commits found in git log: aa2512e, afdcbfb, 86a395d, 1e6f939
- Gates at finish: `go test ./...` 15/15 packages ok, `golangci-lint run ./...` 0 issues, `git status --porcelain` empty, `git clean -n -d plugins/ skill/examples` empty, `diff -r skill <both plugin copies>` parity OK
