---
created: 2026-08-30T20:53:45.906Z
title: Fix non-expanded root diagram bloated by ancestor wrapping
area: rendering
files:
  - internal/graph/builder.go
  - .planning/phases/38-hierarchy-wrapping-and-granular-keys/38-01-SUMMARY.md
  - .planning/STATE.md
---

## Problem

Regression introduced by phase 38 / v1.22.0, commit b2447da (2026-08-30, "feat(38-01): wrap boundary entries and expanded unit in ancestor wrapper clusters (WRAP-01/02)", +136/-10 in `internal/graph/builder.go`).

Reproduction on the SAME model commit, only the c4drill binary differs:

| binary | titles | canvas |
|--------|--------|--------|
| old (pre-v1.22.0) | 55 | 1130×1991 |
| new (v1.22.0) | 260 | 4121×5964 |

Model properties are identical between runs (expanded, edges, legend); removing `expanded: [actors, yic]` changes nothing. The ancestor wrapping rebuilt boundary-graph assembly — the expanding unit and its environment now get wrapped into ancestor-cluster chains — and as a side effect the ROOT (non-expanded) diagram stopped being a compact C1 view and prints the entire model. Net result: non-expanded output practically coincides with expanded, defeating the purpose of both.

Consequence for past analysis: the earlier "no flags is better" verdict was measured on a broken basis — the 93.8 Mpt² figure described the already-bloated non-expanded render, not the compact one that exists in the model commit. Any v1.22.0-era size/compactness comparisons are suspect.

Notable: committed goldens cover C1/expanded only and 38-04 reported a zero-delta BC-01 re-baseline ("WRAP is C2/C3-only"), yet the root diagram visibly ballooned — the fixture models evidently lack the structure that triggers whole-model printing (STATE.md already warned WRAP "will cause REAL golden re-baselining for models with cross-container links"). New fixtures replicating the user's model shape are needed, or this regresses silently again.

## Solution

TBD — to settle at planning:
1. Diagnose in `internal/graph/builder.go` (b2447da hunk) why root-level boundary assembly pulls ancestor chains covering the whole model instead of stopping at the compact C1 neighborhood.
2. Restore the invariant: non-expanded root stays a compact C1; wrapper clusters appear in drill-down/expanded contexts (WRAP-01/02 intent) without flooding the root.
3. Add a fixture reproducing the reported shape (cross-container links + `expanded: [actors, yic]`) and pin compact-root goldens (title count / canvas size) so the regression is caught by CI.
4. Re-run the compactness comparison that produced the 93.8 Mpt² / "no flags" conclusion on the fixed binary.
