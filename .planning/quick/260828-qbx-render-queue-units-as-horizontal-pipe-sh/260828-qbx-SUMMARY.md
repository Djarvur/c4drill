---
status: complete
phase: 260828-qbx-render-queue-units-as-horizontal-pipe-sh
plan: 01
subsystem: render
tags: [svg, post-processing, queue, pipe-shape, graphviz]
requires:
  - graph.IsQueueType (internal/graph/shapes.go)
  - render(g, graphviz.SVG) funnel (internal/render/render.go)
provides:
  - internal/render/pipe.go queue SVG post-processor (collectQueueNodeIDs, replaceQueuePolygons, applyPipeRendering)
  - plain-rect (non-rounded) queue node styling with width=1.8 minimum
affects:
  - internal/render/render.go (SVG post-processing hook)
  - skill/examples SVGs + both plugins/c4drill skill copies (re-rendered, synced)
tech-stack:
  added: []
  patterns:
    - byte-level SVG post-processing keyed on <title>ID</title> (wrapSVGInHTML precedent)
    - plain-string title matching for user-controlled IDs (no regexp interpolation, T-QBX-01)
key-files:
  created:
    - internal/render/pipe.go
    - internal/render/pipe_internal_test.go
  modified:
    - internal/graph/shapes.go
    - internal/render/converter.go
    - internal/render/render.go
    - internal/render/render_test.go
    - internal/render/converter_test.go
    - internal/graph/shapes_test.go
    - internal/graph/integration_test.go
    - cmd/c4drill/testdata/multilevel.expanded.dot
    - skill/examples/*.svg (tracked queue examples)
    - examples/cloud-system/*.svg
    - plugins/c4drill/{skills,opencode}/skills/c4drill-toml/ (synced)
decisions:
  - "Queue nodes keep shape=box; the pipe is a byte-level SVG polygon→path replacement inscribed in the same bbox, so cgraph edge anchors stay valid"
  - "Queue nodes drop the 'rounded' style so GraphViz emits a parseable plain-rect <polygon>; with solid border and no fill the whole style attribute is omitted (graphviz default IS plain)"
  - "buildStyleString gains a rounded bool parameter; clusters always pass true — non-queue styling byte-identical"
  - "Pipe geometry: left half-ellipse bulge, straight top/bottom edges, right full-ellipse cap, rx=0.35*ry, fill/stroke/stroke-width/stroke-dasharray copied verbatim from the replaced polygon"
metrics:
  duration: ~30 min
  completed: 2026-08-28
---

# Quick Task 260828-qbx: Render queue units as horizontal pipe shapes Summary

Queue-type units now render as horizontal pipes in SVG and HTML: the renderer drops the rounded queue style (plain-rect polygon, width=1.8 min), drops the ╟/╢ text-bar icon, and a new SVG post-processor (internal/render/pipe.go) replaces each queue polygon with an inscribed pipe path (left half-ellipse bulge, straight top/bottom, right full-ellipse cap) copying the polygon's paint attributes verbatim.

## Tasks Completed

| Task | Name | Commit(s) | Files |
| ---- | ---- | --------- | ----- |
| 1 | Plain-rect queue style + wider min width; queue icon removed; DOT golden re-baselined | 978ab5c (RED), 3630b02 (GREEN), 99dac2f (test alignment) | internal/graph/shapes.go, internal/render/converter.go, shapes_test.go, integration_test.go, converter_test.go, testdata/multilevel.expanded.dot |
| 2 | Pipe post-processor (pipe.go) wired into render() | 45c998a (RED), 1ed7a40 (GREEN) | internal/render/pipe.go, pipe_internal_test.go, render.go |
| 3 | Integration test, example re-render, skill sync, docs sweep, full gate | 04a7e27, c880c7a | internal/render/render_test.go, skill/examples/*.svg, examples/cloud-system/*.svg, plugins copies |

## What Was Built

1. **shapes.go**: `IconForType` returns `""` for all four queue types; doc comment now points at the SVG pipe post-processor.
2. **converter.go**: `buildStyleString(style, rounded bool)` — queue nodes (via `applyNodeStyle(cn, node)`) pass `rounded=false` so GraphViz emits a plain rect `<polygon>`; queue nodes get `width=1.8` minimum; shape stays `box` (edge anchors computed on the box bbox). Clusters unchanged (`rounded=true`).
3. **pipe.go** (new): `collectQueueNodeIDs` walks top-level nodes + nested clusters; `applyPipeRendering`/`replaceQueuePolygons` locate each queue group via **plain string** `<title>ID</title>` search (T-QBX-01: user IDs never enter a regexp), bounded by the group's closing `</g>`; parse the polygon points into a bbox (T-QBX-02: bounded, single-pass) and emit the pipe path `M(bodyL,y0) A(rx,ry) …Z` with fill/stroke/stroke-width/stroke-dasharray copied verbatim (dashed pipes stay dashed — GraphViz 15.1.1 emits `stroke-dasharray="5,2"` on the polygon, verified empirically).
4. **render.go**: `render()` applies `applyPipeRendering` only when `format == graphviz.SVG`; zero-queue graphs pass through unchanged (fast path); `RenderHTML` inherits automatically; DOT/XDOT untouched.
5. **Tests**: TDD RED→GREEN for icon removal, style/width emission, ID collection through nested clusters, path geometry (bbox anchoring, rx=0.35*ry, arc commands), attribute copying, metacharacter-ID safety, zero-queue passthrough, SVG-only wiring, and an SVG+HTML integration test.
6. **Examples/docs**: tracked queue-bearing SVGs re-rendered (06-templates, 10-edge-kinds, cloud-system incl. expanded); non-queue examples (overflow-test, rank-for-better-layout) byte-identical; skill rsynced into both `plugins/c4drill` copies; validation artifacts cleaned (`git clean -f plugins/`, `git clean -x -f skill/examples`) → 0 untracked. Docs sweep: SKILL.md/README.adoc carry no stale icon claims (no changes needed).

## Verification Results

- `go test ./...` — green (full suite)
- `golangci-lint run ./...` — 0 issues
- Golden DOT re-baseline (`go run ./cmd/c4drill cmd/c4drill/testdata/multilevel.toml --format dot --expanded --output cmd/c4drill/testdata`) — **zero diff** (multilevel.toml contains no queue nodes), which proves non-queue rendering is byte-identical (stronger than "only queue-related deltas")
- Visual check (`qlmanage` render of a queue/db/system fixture and `skill/examples/10-edge-kinds.svg`): the pipe reads as a horizontal cylinder — left bulge, straight body, right cap; dashed variant keeps its dash; sits naturally next to the db cylinder; legend unchanged

## Deviations from Plan

**1. [Rule 3 - Blocking lint] pipe_test.go renamed to pipe_internal_test.go**
- **Found during:** Task 2 GREEN
- **Issue:** `.golangci.yml` `testpackage.skip-regexp: _internal_test\.go$` requires external test packages unless the file is named `*_internal_test.go`; the tests access unexported pipe functions.
- **Fix:** renamed the file to `internal/render/pipe_internal_test.go` (rename committed as part of 1ed7a40).
- **Files modified:** internal/render/pipe_internal_test.go

**2. [Rule 1 - Bug in own test] Queue style assertion aligned with omitted empty style**
- **Found during:** Task 1 GREEN
- **Issue:** a solid queue node with no fill carries NO style attribute at all (SafeSet with "" omits it; graphviz's default style IS plain) — the test wrongly required a `style=` attribute to exist.
- **Fix:** assertion now only requires that "rounded" is gone (commit 99dac2f).
- **Files modified:** internal/render/converter_test.go

**3. [Scope note] Task 1 DOT style tests hosted in converter_test.go**
- The behavior block requires DOT style assertions, but the task file list omitted a test file for them; placed in the existing `internal/render/converter_test.go` next to `TestNodeRoundedStyle`.

**4. [Scope note] Golden DOT re-baseline is a no-op**
- cmd/c4drill/testdata/multilevel.toml has no queue units, so the re-baseline produced zero diff (queue-only delta expectation holds vacuously; non-queue byte-stability proven).

## Notes for Verifier

- The `═╦╩═╦═══` ASCII-art row inside queue labels (internal/render/labels.go `buildQueueHTMLLabel`) is **intentionally retained** — labels.go is outside the plan's files_modified scope, and the pipe shape + art row render coherently (the art reads as pipe detailing). Flag for a future quick task if the label row should be dropped now that the shape itself is a pipe.
- Threat mitigations landed as specified: T-QBX-01 (plain-string title match, metachar-ID test), T-QBX-02 (bounded single-pass parsing), T-QBX-03 (no package installs). No new security surface beyond the plan's threat model.

## Self-Check: PASSED

- internal/render/pipe.go — FOUND
- internal/render/pipe_internal_test.go — FOUND
- internal/graph/shapes.go (icon ""), internal/render/converter.go (IsQueueType branch), internal/render/render.go (SVG-only hook) — FOUND
- Commits 978ab5c, 3630b02, 45c998a, 1ed7a40, 99dac2f, 04a7e27, c880c7a — all FOUND in git log
- `git status --porcelain plugins/ skill/examples/` untracked count = 0 — VERIFIED
