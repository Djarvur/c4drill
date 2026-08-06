# Phase 4 Context: C3 Boundary Node Fix

## Goal

Fix DEFERRED-04: `addResolvedBoundaryNode` in `internal/view/scope.go` walks past the sibling level when resolving cross-container links in C3 views, creating a boundary node for the parent system instead of the sibling container.

## Problem (verified)

C3 view of `mainSystem.localIDP`:
- Component `localIDP.grpcAPIs.authAPI` links to `mainSystem.rbac` (sibling container)
- `addResolvedBoundaryNode(v, m, "mainSystem.rbac", "mainSystem.localIDP")` walks: `mainSystem.rbac` → `mainSystem` → break (idx=0)
- Creates boundary node for `mainSystem` (the parent) — **wrong**
- Should create boundary node for `mainSystem.rbac` (the sibling) — **correct**

## Root Cause

`addResolvedBoundaryNode` (scope.go:625) receives `scopePath` (the container path, e.g., `mainSystem.localIDP`) but does NOT use it to bound the walk-up. The walk-up strips path segments until it finds a visible ancestor or reaches the top level. For C3, the container's parent (`mainSystem`) is not in `v.Units` (only `mainSystem.localIDP.*` entries exist), so the walk overshoots to the top.

## Decision

**D-01: Stop walk-up at the container's parent level.**

The walk-up in `addResolvedBoundaryNode` must stop when the next strip would cross above the container's parent. Concretely: compute the container's parent path from `scopePath` (e.g., `mainSystem` from `mainSystem.localIDP`). During the walk-up, if `peer` (after stripping) equals the container's parent, stop — the current peer (before that strip) is the boundary node.

This means:
- Peer `mainSystem.rbac`, scope `mainSystem.localIDP` → parent of scope is `mainSystem`. Walk: `mainSystem.rbac` → next strip would give `mainSystem` (the scope parent) → STOP. Boundary = `mainSystem.rbac` ✓
- Peer `externalSys`, scope `mainSystem.localIDP` → walk: `externalSys` → idx=0, break. Boundary = `externalSys` ✓ (unchanged — top-level external)
- Peer `mainSystem.localIDP.grpcAPIs.authAPI`, scope `mainSystem.localIDP` → walk: strip to `mainSystem.localIDP.grpcAPIs` → strip to `mainSystem.localIDP` → IN VIEW → return (internal reference, no boundary) ✓ (unchanged)

[auto] D-01 — Q: "How should the walk-up be bounded?" → Selected: "Stop at container's parent level" (recommended default — minimal change, correct for all view levels)

## Scope Boundary

- Fix is in `addResolvedBoundaryNode` only (scope.go:625-651)
- The `scopePath` parameter already carries the container path — no signature changes needed
- C1 and C2 boundary resolution must remain unchanged (their call sites pass `systemPath`/`containerPath` which already bounds correctly via the existing logic)

## Test Plan

- New test: C3 view where a component links to a sibling container → assert the sibling is the boundary node, NOT the parent
- Existing boundary tests (C1 deepest-visible-ancestor) must still pass
- Verify via the `multilevel.toml` fixture: `mainSystem.localIDP` C3 SVG should show `RBAC` as boundary node, not `Main System`

## Files

- `internal/view/scope.go` — `addResolvedBoundaryNode` (the fix)
- `internal/view/scope_test.go` or `integration_test.go` — new regression test
