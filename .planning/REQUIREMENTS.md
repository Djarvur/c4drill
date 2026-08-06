# Requirements: C4Drill v1.9 — C3 Boundary Node Fix

**Defined:** 2026-08-06
**Milestone:** Fix DEFERRED-04 — C3 boundary node resolves to parent instead of sibling container.

## Core Problem

When a C3 diagram (e.g., `mainSystem.localIDP`) contains a component that links to a sibling container (e.g., `mainSystem.rbac`), the boundary node shows the **parent system** (`Main System [+]`) instead of the **sibling container** (`RBAC`). This is confusing — the parent system is the expanded container being viewed, not an external entity.

## Root Cause (diagnosed in DEFERRED-04)

`addResolvedBoundaryNode` (`internal/view/scope.go:625`) walks the peer path up to find the nearest visible ancestor. For peer `mainSystem.rbac` in a C3 view of `mainSystem.localIDP`, it walks: `mainSystem.rbac` → `mainSystem` (breaks at top level). Since `mainSystem` is not in the view (only `mainSystem.localIDP.*` entries exist), it creates a boundary node for `mainSystem` — the parent. It should stop at `mainSystem.rbac` (the sibling level).

## Requirements

### Boundary Resolution

- [ ] **BOUND-01**: C3 diagram boundary nodes for cross-container links resolve to the sibling container, not the parent system
- [ ] **BOUND-02**: C3 boundary resolution does not regress C1 or C2 boundary behavior (deepest-visible-ancestor still works)
- [ ] **BOUND-03**: Regression test covers C3 cross-container boundary resolution (sibling as boundary node, not parent)

## Out of Scope

- Changing C1 boundary resolution (works correctly — deepest visible ancestor)
- Changing C2 boundary resolution (works correctly — sibling containers)
- The boundary node visual styling (separate concern)
