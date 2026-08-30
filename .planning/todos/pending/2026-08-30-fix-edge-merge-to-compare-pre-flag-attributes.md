---
created: 2026-08-30T20:59:06.975Z
title: Fix edge merge to compare pre-flag arrow attributes
area: rendering
files:
  - internal/graph/builder.go:1225
  - internal/graph/builder.go:980
  - internal/graph/builder.go:1449
  - internal/validator/index.go
---

## Problem

User report on non-expanded mode: when two units are connected by identical arrows, the arrows merge into ONE edge with increased (binary) thickness. The comparison treats arrows as identical based on shape, color, label, and direction.

The critical finding: **the comparison runs AFTER the format-disabling flags are applied.** This means `--no-colors`, `--no-labels` (and compositions) change which arrows count as identical — edges that are distinct in the default render (differing only in color or label) collapse into a single thickened edge once the flags strip those attributes. The flags stop being pure presentation switches and silently alter graph semantics (edge count/multiplicity), so the flag-on render is not the same diagram with formatting off.

Related observation from the report session: this interacts with the pending `--no-labels` semantics bug (2026-08-30-fix-no-labels-to-suppress-only-edge-labels.md) — if `--no-labels` stops killing node/cluster labels its effect on the merge key changes too.

Code discrepancy to reconcile during debugging: the user observes comparison by shape/color/label/direction, but the dedup key at `internal/graph/builder.go:1227-1229` is pair-based — `path + "->" + link.Peer`, extended with `Technology + Description` only in expanded mode (D-01/D-02 pair-only key in resolved views); `markSeen` (1449) dedups, `countPairMultiplicity` (980, D-05) counts per pair, `applyCollapsedPairStyle` sets binary penwidth 2.0 (D-04). Either a second comparison layer exists elsewhere (validator mirror synthesis at internal/validator/index.go?), or actual merging is coarser than reported — both directions change the fix. Related decisions: D-01..D-05, WR-02, COMPAT-02.

## Solution

TBD — to settle at planning:
1. Establish the invariant first: format-disabling flags must not change merge outcomes — the comparison key must be computed from pre-flag (source model) attributes. Pin it with a test: same model, with/without flags → identical edge topology (only styling differs).
2. Reconcile the reported comparison key (shape/color/label/direction) with the pair-based edgeKey in builder.go; decide the canonical key (and whether label/color belong in it at all, given they are flag-suppressible).
3. Decide whether post-merge thickness/multiplicity display stays as-is under the new key.
