# Roadmap: C4Drill v1.8

**Created:** 2026-03-29
**Milestone:** v1.8 Proper C1/C2/C3 View Generation

## Overview

Fix view generation so C1 shows only top-level units, C2/C3 diagrams are auto-generated for units with subunits, and edges resolve correctly across levels.

**3 phases** | **9 requirements** | All covered ✓

| # | Phase | Goal | Requirements | Success Criteria |
|---|-------|------|--------------|------------------|
| 1 | Fix C1 View Scoping | 3/3 | Complete    | 2026-08-06 |
| 2 | Auto-generate C2/C3 | 1/1 | Complete   | 2026-08-06 |
| 3 | Compatibility & Validation | Existing TOML files and --expanded still work | COMPAT-01, COMPAT-02 | 2 |

---

## Phase 1: Fix C1 View Scoping

**Goal:** C1 diagram shows only top-level units; links to nested targets resolve to nearest visible ancestor; duplicate edges collapse.

**Requirements:** VIEW-01, VIEW-02, EDGE-01, EDGE-02

### Root Cause Analysis

Two bugs in `internal/view/scope.go`:

1. **`addExternalBoundaryNodes`** recursively scans ALL links (including from deeply nested subunits) and creates boundary nodes for any target not already in the view. When `webUser` links to `linuxSystem.localIDP.grpcAPIs.authAPI`, it creates that deeply nested component as a C1 node.

2. **Edge building in `builder.go`** passes all links through without resolving targets to visible ancestors.

### Success Criteria

1. C1 diagram for `saira-20260320.c2.full.toml` has exactly 5 nodes (webUser, sshUser, adminUser, keycloak, linuxSystem)
2. All edges in C1 connect only top-level units (no deeply nested targets)
3. Multiple links to subunits of the same parent collapse to a single edge between parents
4. Existing tests continue to pass

### Key Files

- `internal/view/scope.go` — Fix `addExternalBoundaryNodes` to only consider top-level links
- `internal/graph/builder.go` — Add edge resolution logic to map nested targets to visible ancestors
- `internal/view/scope_test.go` — New tests for C1 scoping
- `internal/graph/builder_test.go` — New tests for edge resolution

**Plans:** 3 plans
Plans:
**Wave 1**

- [x] 01-01-PLAN.md — Pair-only edge collapse with binary penwidth (D-01..D-06)

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 01-02-PLAN.md — Remove legacy boundary nodes, activate AllExpanded, gate minlen (D-12, D-02, D-13)

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 01-03-PLAN.md — Deepest-visible-ancestor resolution both sides (D-07..D-11)

---

## Phase 2: Auto-generate C2/C3 Diagrams

**Goal:** Automatically generate C2/C3 diagram files for every unit that has subunits. `properties.expanded` controls which top-level units show expanded in C1.

**Requirements:** VIEW-03, VIEW-04, VIEW-05

### Root Cause Analysis

`collectExpandedPaths` in `cmd/c4drill/root.go` only checks per-unit `Expanded` field for self-referencing paths, ignoring `m.Properties.Expanded`. It should:

- Use `m.Properties.Expanded` to determine which top-level units are expanded in C1
- Auto-detect ALL units with subunits and generate C2/C3 diagrams for them

### Success Criteria

1. Running `c4drill saira-20260320.c2.full.toml` creates a directory `saira-20260320.c2.full/` with C2/C3 sub-diagrams
2. Each system/box with subunits gets a C2 diagram
3. Each container with subunits gets a C3 diagram
4. `properties.expanded` list controls which top-level units appear as clusters (expanded) in C1
5. Units not in `properties.expanded` appear as collapsed nodes with `[+]` in C1

### Key Files

- `cmd/c4drill/root.go` — Fix `collectExpandedPaths` to use `m.Properties.Expanded` and auto-detect all expandable units
- `internal/view/scope.go` — Ensure `GenerateC1View` uses `properties.expanded` for top-level expansion
- `cmd/c4drill/root_test.go` — Test C2/C3 file generation

**Plans:** 1 plan
Plans:
**Wave 1**

- [x] 02-01-PLAN.md — D-07 plain-node guard + lock D-01..D-06/D-08 with regression tests

---

## Phase 3: Compatibility & Validation

**Goal:** Ensure backward compatibility and validate with real TOML files.

**Requirements:** COMPAT-01, COMPAT-02

### Success Criteria

1. `cmd/c4drill/testdata/valid.toml` (no `properties.expanded`) generates correct C1 with all units collapsed
2. `--expanded` flag produces identical output to v1.7 for the same input
3. All existing tests pass without modification (except tests verifying the buggy behavior)
4. `saira-20260320.c2.full.toml` generates correct C1 (5 nodes) plus C2/C3 sub-diagrams

### Key Files

- `cmd/c4drill/root_test.go` — Integration tests
- `internal/view/integration_test.go` — View integration tests

---

## Requirement Coverage

| Requirement | Phase | Status |
|-------------|-------|--------|
| VIEW-01 | Phase 1 | Pending |
| VIEW-02 | Phase 1 | Pending |
| EDGE-01 | Phase 1 | Pending |
| EDGE-02 | Phase 1 | Complete |
| VIEW-03 | Phase 2 | Pending |
| VIEW-04 | Phase 2 | Pending |
| VIEW-05 | Phase 2 | Pending |
| COMPAT-01 | Phase 3 | Pending |
| COMPAT-02 | Phase 3 | Pending |

**Coverage:** 9/9 requirements mapped ✓
