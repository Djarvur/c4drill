# Phase 2: Auto-generate C2/C3 Diagrams - Context

**Gathered:** 2026-08-06
**Status:** Ready for planning

<domain>
## Phase Boundary

Refine the existing (already-implemented, tested) C2/C3 sub-diagram generation. Every unit with subunits gets an auto-generated sub-diagram written to `{basename}/{unit}.{format}`; `properties.expanded` controls which top-level units appear expanded in C1. Requirements VIEW-03, VIEW-04, VIEW-05. The v1.8 implementation already exists (auto-detect via `collectExpandableUnitPaths`, `GenerateC2View`/`GenerateC3View`, boundary clusters, navigation); this phase refines behavior per the decisions below.

</domain>

<decisions>
## Implementation Decisions

### Sub-diagram Generation (VIEW-03/04)
- **D-01:** Uniform auto-detect — ANY unit with subunits gets a sub-diagram, C1 boxes included (a collapsed box drills down via its explore link to a sub-diagram showing the grouped C1 units). One rule for all units.
- **D-02:** `containerBox`/`componentBox` grouping units follow the same uniform rule — deep box parity (a containerBox with subunits gets a C3 sub-diagram showing its grouped contents).
- **D-03:** Sub-diagram files are named by the unit's TOML section key (path segment), e.g. `{basename}/linuxSystem.svg`. Explore/back/breadcrumb URLs derive from the same dotted paths (`internal/graph/path.go`) so links and files can never diverge. Display names are NOT used in file names.

### C1 Expansion Semantics (VIEW-05)
- **D-04:** Expansion renders exactly ONE level in C1 — an expanded top-level unit shows its direct subunits as cluster nodes. Deeper detail lives in the C2/C3 files via explore links. Per-unit `expanded` still renders clusters in C2/C3 views (an expanded container in its system's C2 diagram shows its components).
- **D-05:** Expansion precedence is OR — a top-level unit expands in C1 if `properties.expanded` contains its path OR its own `expanded` list self-references it (union; no conflict). Per-unit self-reference retained for v1.7 backward compat (COMPAT-01).
- **D-06:** `properties.expanded` entries that match no top-level unit are SILENTLY IGNORED (no error, no output) — permissive for files authored against other tools, consistent with the CLI's "silent per spec" convention.
- **D-07:** A top-level unit listed in `properties.expanded` but with NO subunits renders as a plain collapsed node — no empty cluster box. Expansion only takes effect when there are subunits to show.

### Actors in C2/C3
- **D-08:** Linked actors (person/personExternal — C1-only types) render as external boundary nodes in C2/C3 diagrams via the existing `addExternalBoundaryNodesForSubunits` path. This confirms the v1.0 deferred item ("Person display on C2/C3 diagrams") as shipped behavior — actors are NOT filtered from deeper views.

### Claude's Discretion
None — every area was decided with concrete options.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Requirements & Scope
- `.planning/REQUIREMENTS.md` — VIEW-03, VIEW-04, VIEW-05 (phase 2); COMPAT-01 (constraint on per-unit expansion)
- `.planning/ROADMAP.md` — Phase 2 goal, root-cause analysis, success criteria
- `.planning/phases/01-fix-c1-view-scoping/01-CONTEXT.md` — D-01/D-02 (pair-only collapse in C2/C3), D-07..D-11 (deepest-visible-ancestor both sides, replace-parent-edge, within-cluster edges) — these constrain Phase 2's C2/C3 view behavior
- `.planning/PROJECT.md` — TOML schema (expanded lists), rendering behavior, file structure

### Implementation Files (behavior verified during Phase 1 discussion)
- `cmd/c4drill/root.go` — `collectExpandedPaths` (:155), `collectExpandableUnitPaths` (:170) — the auto-detect at the heart of VIEW-03/04
- `internal/view/scope.go` — `isExpandedInC1` (:144), `GenerateC2View` (:390), `GenerateC3View` (:458), `addExternalBoundaryNodesForSubunits` (:577), `addResolvedBoundaryNode` (:619), `resolveBoundaryNodeLinks` (:652), `resolveSubunitCrossLinks` (:750)
- `internal/graph/builder.go` — `BuildGraphWithPath` (:474), `buildBoundaryCluster` (:79), `buildCluster` (:265, one-level expansion), `buildNestedCluster` (:172), `shouldHaveExploreLink` (:507)
- `internal/graph/path.go` — explore/back/breadcrumb URL derivation from dotted paths (D-03 anchor)
- `internal/output/writer.go` — C2/C3 directory hierarchy layout (`{basename}/{system}/{container}.{format}`)
- `internal/view/integration_test.go`, `internal/graph/builder_test.go` — existing C2/C3 test patterns

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `isExpandedInC1` (`scope.go:144`): OR-semantics expansion check — D-05/D-06/D-07 changes concentrate here and in the graph layer's cluster decision
- `collectExpandableUnitPaths` (`root.go:170`): recursion + auto-detect — D-01/D-02 already satisfied; verify no grouping-type special-casing exists
- `buildCluster` (`builder.go:265`): one-level cluster rendering — D-04 anchor; D-07's plain-node fallback lives in the expansion decision (skip cluster when `len(Subunits)==0`)
- Boundary-node machinery (`addExternalBoundaryNodesForSubunits` etc.): D-08 confirmed behavior, no change expected
- `buildBoundaryCluster` (`builder.go:79`): C2/C3 expanded-unit boundary — label/styling from the unit itself

### Established Patterns
- Definition-order preservation (`UnitOrder`/`SubunitOrder`) for deterministic file generation order
- Dotted paths → directory hierarchy in `internal/output/writer.go`
- Pair-only dedup + first-wins (Phase 1 D-01/D-03) applies to C2/C3 edges — do not regress
- CLI silent-on-success; errors to stderr only (D-06)

### Integration Points
- `processView` (`root.go:202`) → generator selection by path depth (`isC2Path`/`isC3Path`)
- `BuildGraphWithPath` → navigation (explore/back/breadcrumb) for C2/C3
- C1 expansion (`isExpandedInC1` → `BuildGraph` cluster branch) — D-04/D-07 changes here must not affect C2/C3 cluster rendering

</code_context>

<specifics>
## Specific Ideas

No specific external references — decisions captured from discussion with code grounding.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 2-Auto-generate C2/C3 Diagrams*
*Context gathered: 2026-08-06*
