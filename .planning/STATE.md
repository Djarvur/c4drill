---
gsd_state_version: 1.0
milestone: v1.7
milestone_name: milestone
status: unknown
last_updated: "2026-08-06T08:25:51.229Z"
progress:
  total_phases: 3
  completed_phases: 0
  total_plans: 0
  completed_plans: 0
  percent: 0
---

# v1.8: Proper C1/C2/C3 View Generation

## Status: **COMPLETE**

## Summary

Fixed the two root-cause bugs:

1. **C1 pollution**: `addExternalBoundaryNodesRecursive` scanned ALL nested links → replaced with `addC1BoundaryNodes` that resolves peers to top-level ancestors
2. **No C2/C3**: `collectExpandedUnitPaths` only checked per-unit `Expanded` → replaced with `collectExpandableUnitPaths` that auto-detects units with subunits

## Evidence

| Metric | Before | After |
|--------|--------|-------|
| C1 nodes (saira TOML) | 105 | 5 |
| Sub-diagrams generated | 0 | 17 |
| All tests passing | ✅ | ✅ |
| `--expanded` flag | ✅ | ✅ |

## Commits

- `f9dc69a` fix: proper C1/C2/C3 view generation with edge resolution
- `c5091f9` chore: update baseline SVG
