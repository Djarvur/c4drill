# Roadmap: C4Drill v1.9 — C3 Boundary Node Fix

**Created:** 2026-08-06
**Milestone:** v1.9 C3 Boundary Node Fix

## Overview

Fix DEFERRED-04: C3 boundary node resolution walks past the sibling level and creates a boundary node for the parent system instead of the sibling container.

**1 phase** | **3 requirements** | All covered ✓

| # | Phase | Goal | Requirements | Success Criteria |
|---|-------|------|--------------|------------------|
| 4 | C3 Boundary Node Fix | Fix addResolvedBoundaryNode to stop at sibling level in C3 views | BOUND-01, BOUND-02, BOUND-03 | 3 |

---

## Phase 4: C3 Boundary Node Fix

**Goal:** Fix `addResolvedBoundaryNode` so C3 views resolve cross-container links to the sibling container, not the parent system. Add a regression test.

**Requirements:** BOUND-01, BOUND-02, BOUND-03

**Success criteria:**
1. C3 diagram for `mainSystem.localIDP` shows `RBAC` (sibling container) as boundary node, not `Main System` (parent)
2. C1 and C2 boundary resolution unchanged (existing boundary tests pass)
3. New regression test covers the C3 cross-container sibling case

**Context:**
- Bug diagnosed in DEFERRED-04 (`.planning/milestones/v1.8-phases/03-compatibility-validation/deferred-items.md`)
- Fix is in `internal/view/scope.go` (`addResolvedBoundaryNode`, line 625)
- The function needs to know the container scope (the C3 view's expanded container path) and stop walking up before crossing above the container's parent

**Key decisions:**
- D-01: The walk-up should stop when the peer path reaches the same parent level as the expanded container. A peer like `mainSystem.rbac` shares parent `mainSystem` with `mainSystem.localIDP`, so `rbac` is the boundary — not `mainSystem`.
