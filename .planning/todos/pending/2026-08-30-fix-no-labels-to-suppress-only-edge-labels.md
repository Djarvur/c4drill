---
created: 2026-08-30T20:51:48.317Z
title: Fix --no-labels to suppress only edge labels
area: rendering
files:
  - cmd/c4drill/root.go:112
  - internal/graph/graph.go:54
  - internal/graph/builder.go:197
  - internal/graph/builder.go:464
  - internal/graph/builder.go:1390
  - internal/render/converter.go:407
  - internal/render/converter.go:523
---

## Problem

`--no-labels` currently suppresses ALL label text at once — node labels, edge labels, AND cluster labels. There is no separate switch for edge labels alone. In practice this makes `--expanded` output a diagram of anonymous rectangles: users who only want to declutter edge label text lose the identity of every node and cluster.

Expected semantics: `--no-labels` should mute ONLY edge labels, leaving node, cluster (wrapper), and legend labels intact.

Background: the current all-or-nothing behavior was deliberately implemented in milestone v1.15 (shipped as v1.22.0) under requirement LBL-01, with decisions pinned in phase 38 ("--no-labels suppresses at the GRAPH layer — builder drops Label content; converter empty-label emission is defense-in-depth; legend stays per LBL-03"). Fixing this reverses/reshapes LBL-01's scope, so the requirement text, the committed `nolabels.dot`/`nolabels.expanded.dot` goldens (added in 38-04), the KEY-03 switch matrix, and docs (`README.md`, `skill/SKILL.md:1006-1017` — "shape-only: no element label text") all need re-baselining together.

Touch points for the current behavior:
- `cmd/c4drill/root.go:112` — flag registration and help text; `root.go:332,385` — threading onto views (LBL-02)
- `internal/graph/graph.go:54` — `NoLabels` field doc ("all element labels")
- `internal/graph/builder.go` — suppression sites: wrapper clusters (197-200), boundary clusters (291-292), expanded-unit clusters (404-407, 843-846), nodes (464-468), edge labels (1390-1392)
- `internal/render/converter.go` — defense-in-depth empty-label emission (407-411 nodes, 523-526 clusters)

## Solution

TBD — options to settle at planning:
1. Narrow `--no-labels` to edges only (rename semantics; keep flag name for compatibility), OR
2. Keep `--no-labels` as-is (all labels) and add a granular `--no-edge-labels`, matching the v1.15 granular-switch precedent (KEY-01/LBL-02 threading).

Either way: re-baseline the two nolabels goldens, update the KEY-03 E2E matrix, update README/SKILL.md wording, and supersede the pinned LBL-01 decisions in phase-38 deferred/decision records.
