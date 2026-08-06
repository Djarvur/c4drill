# Phase 2: Auto-generate C2/C3 Diagrams - Research

**Researched:** 2026-08-06
**Domain:** Go C4 diagram sub-diagram generation (CLI rendering pipeline)
**Confidence:** HIGH

> **Research provenance note:** Same as Phase 1 — the gsd-phase-researcher agent cannot spawn
> (its definition requires firecrawl/exa MCP servers; web research disabled in config). The
> orchestrator produced this RESEARCH.md directly from codebase analysis performed during
> discuss-phase. All claims tagged `[VERIFIED: <file>:<line>]`. No external dependencies involved.

## Summary

Phase 2 refines the already-implemented C2/C3 sub-diagram generation. The auto-detect mechanism
(`collectExpandableUnitPaths`) and the view/graph/navigation machinery exist and pass tests.
The 8 locked decisions (D-01..D-08) are predominantly **confirmations of existing behavior**
(D-01 uniform auto-detect incl. boxes, D-02 deep box parity, D-03 unit-key file naming,
D-05 OR expansion precedence, D-06 silent ignore, D-08 actors in C2/C3) plus **one real
behavior refinement**: D-07 — a top-level unit listed in `properties.expanded` with NO subunits
must render as a plain node, not an empty cluster box. D-04 (one expansion level in C1) is
already the behavior of `buildCluster` (direct subunits as nodes, no recursion).

**Primary recommendation:** The phase is verification-heavy — lock the confirmed behaviors with
explicit regression tests (including a box sub-diagram test and an actors-in-C2 test that may
already exist), and implement the single D-07 refinement (guard cluster creation on
`len(Subunits) > 0`). Expect 1–2 small plans, not a big implementation phase.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01:** Uniform auto-detect — ANY unit with subunits gets a sub-diagram, C1 boxes included.
- **D-02:** containerBox/componentBox follow the same uniform rule (deep box parity).
- **D-03:** Sub-diagram files named by unit's TOML section key; URLs derive from the same dotted paths.
- **D-04:** Expansion renders exactly ONE level in C1; per-unit `expanded` renders clusters in C2/C3 views.
- **D-05:** Expansion precedence is OR (properties.expanded ∪ per-unit self-reference).
- **D-06:** Non-matching properties.expanded entries silently ignored.
- **D-07:** Expanded-but-empty units render as plain nodes (no empty cluster box).
- **D-08:** Linked actors render as external boundary nodes in C2/C3 (confirmed v1.0 deferred item).

### Claude's Discretion
None.

### Deferred Ideas (OUT OF SCOPE)
None.

</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| VIEW-03 | C2 diagram auto-generated for each system/box with subunits → `{basename}/{unit}.{format}` | `collectExpandableUnitPaths` (cmd/c4drill/root.go:170) + `writer.Write` (internal/output/writer.go:37) — verified present |
| VIEW-04 | C3 diagram auto-generated for each container with subunits → `{basename}/{unit}.{format}` | Same recursion; `isC3Path` (dotted path) selects GenerateC3View (root.go:209) |
| VIEW-05 | properties.expanded controls which top-level units appear expanded in C1 | `isExpandedInC1` (internal/view/scope.go:144) — OR semantics verified; D-07 is the refinement |

</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Sub-diagram discovery (auto-detect) | `cmd/c4drill` (root.go) | `internal/output` (writer) | collectExpandedPaths/collectExpandableUnitPaths own the path list |
| View generation (C2/C3) | `internal/view` (scope.go) | — | GenerateC2View/C3View project the model subtree |
| Cluster/expansion rendering | `internal/graph` (builder.go) | — | buildCluster (one level), buildNestedCluster (expanded mode) |
| File/URL naming | `internal/output` + `internal/graph/path.go` | — | Dotted paths → dirs → URLs; single source of truth |
| C1 expansion control | `internal/view` + `internal/graph` | — | isExpandedInC1 decides, BuildGraph renders cluster vs node |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib (`slices`, `strings`, `path/filepath`) | 1.26.1 | Path walking, dirs | Already used |
| stretchr/testify | v1.11.1 | Assertions | Existing convention |

### Supporting
None — **zero new dependencies**.

## Architecture Patterns

### Pattern 1: Path-depth generator selection
`processView` (cmd/c4drill/root.go:202-210) selects the generator by path shape: `""` → C1,
no-dots → C2 (`isC2Path` root.go:289), dotted → C3. Boxes at top level produce no-dot paths →
C2 generator — consistent with D-01. This selection is the anchor for box sub-diagram behavior.

### Pattern 2: Recursive auto-detect with definition order
`collectExpandableUnitPaths` (root.go:170-199) recurses `UnitOrder`-driven, appending any unit
with `len(Subunits) > 0`. Deterministic generation order. Boxes and containerBox/componentBox
included with no type checks — D-01/D-02 already satisfied structurally.

### Pattern 3: One-level cluster rendering
`buildCluster` (builder.go:265-311) renders the expanded unit's DIRECT subunits as nodes;
children marked expanded via `isUnitExpanded` (builder.go:324) are still added as nodes —
no recursion in BuildGraph's C1/C2/C3 branch (only `buildNestedCluster` for `--expanded` mode
recurses). D-04 already holds.

### Pattern 4: OR-semantics expansion check
`isExpandedInC1` (scope.go:144-150): `slices.Contains(m.Properties.Expanded, unitPath)` OR
`slices.Contains(unit.Expanded, unitPath)`. D-05 confirmed; D-06 confirmed (deep entries simply
never match — no error path exists).

## Common Pitfalls

### Pitfall 1: D-07 empty-cluster rendering
**What goes wrong:** A top-level unit in `properties.expanded` with no subunits → `IsExpanded:
true` → `BuildGraph` C1 branch calls `buildCluster` (builder.go:60-62) with zero children →
renders an empty cluster box (label with nothing inside).
**How to avoid:** Guard at the expansion decision or graph branch: expanded ⇒ requires
`len(Subunits) > 0`; otherwise render a plain node. Cleanest single point: in `BuildGraph`'s
C1 branch (builder.go:59-67) — `entry.IsExpanded && len(entry.Unit.Subunits) > 0` → cluster,
else node. Must NOT change C2/C3 cluster behavior (boundary clusters are separate — D-04 note).
**Warning signs:** Empty cluster boxes in C1 SVG output.

### Pitfall 2: Regression on Phase 1 C2/C3 edge work
**What goes wrong:** Phase 1 (WR-01 fix) removed pre-dedup in `addResolvedCrossLink`/`From` so
C2/C3 collapsed pairs thicken. Any edit touching `scope.go` C2/C3 passes could reintroduce the
dedup or break `Link.Mirror` preservation.
**How to avoid:** If `scope.go` must change for D-07 (it shouldn't — the fix is graph-level),
re-run the WR-01/WR-02 tests (`TestBuildEdgesPenwidthC2C3CollapsedPairs`,
`TestBuildEdgesPenwidthLinkFromContributions`). Prefer touching only `builder.go`/`view.go`.
**Warning signs:** Penwidth tests fail; C2/C3 cross-subunit synthesis tests fail.

### Pitfall 3: Writer layout assumptions
**What goes wrong:** Tests asserting file layout must use the unit-key convention
(`{basename}/{system}.{format}`, dotted → dirs, writer.go:37-54). Changing naming (D-03 says
keep) would break explore links (path.go) — don't.
**Warning signs:** Explore URL/file mismatch after any naming change.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Expansion decision | A new expansion-tracking system | `isExpandedInC1` + `HasSubunits` guard | Existing flag already computed per entry |

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | go test + testify v1.11.1 |
| Config file | none |
| Quick run command | `go test ./internal/view/ ./internal/graph/ ./cmd/c4drill/` |
| Full suite command | `go test -v -race -cover ./...` (no t.Parallel on render tests — WASM mutex) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| VIEW-03 | C2 file generated for each system/box with subunits (incl. box sub-diagram) | unit/integration | `go test ./cmd/c4drill/ -run Root -count=1` | ✅ root_test.go (verify box case) |
| VIEW-04 | C3 file generated for each container with subunits | unit/integration | `go test ./cmd/c4drill/ -count=1` | ✅ |
| VIEW-05 | properties.expanded expands listed top-level units in C1 (OR semantics) | unit | `go test ./internal/view/ -run C1 -count=1` | ✅ scope_test.go |
| VIEW-05 (D-07) | Expanded-but-empty unit renders as plain node | unit | `go test ./internal/graph/ -run Cluster -count=1` | ❌ new test |
| D-08 | Actor linked from container appears as boundary node in C2 | unit | `go test ./internal/view/ -count=1` | ⚠ verify existing coverage |

### Sampling Rate
- **Per task commit:** quick run command
- **Per wave merge:** `go test -v -race ./...`
- **Phase gate:** full suite green + `mise lint` before verify-work

### Wave 0 Gaps
- [ ] New test: expanded-but-empty unit → plain node (D-07)
- [ ] New/verified test: box with subunits → sub-diagram file generated (D-01)
- [ ] Verified test: actor boundary node in C2 (D-08) — add if absent

## Security Domain

### Applicable ASVS Categories
| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2/V3/V4/V6 | no | — (local CLI, no auth/sessions/crypto) |
| V5 Input Validation | yes | Existing parser/validator pipeline unchanged; new guard is structural (len check), no new input surface |

### Known Threat Patterns
| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| None new | — | Phase touches graph construction only; nil-map/empty-view guards per existing conventions |

**Threat model note for PLAN.md:** minimal — the D-07 guard is a slice-length check; no panic
paths introduced. Keep nil-guard convention.

## Sources

### Primary (VERIFIED — codebase reads this session)
- `cmd/c4drill/root.go` — collectExpandedPaths (:155), collectExpandableUnitPaths (:170-199), processView (:202), isC2Path (:289)
- `internal/view/scope.go` — isExpandedInC1 (:144), GenerateC2View (:390), GenerateC3View (:458)
- `internal/graph/builder.go` — BuildGraph C1 branch (:56-68), buildCluster (:265-311), buildNestedCluster (:172), isUnitExpanded (:324)
- `internal/output/writer.go` — Write layout (:37-54), WriteExpanded (:66)
- `internal/graph/path.go` — URL derivation
- Phase 1 artifacts: 01-CONTEXT.md (D-01/D-02/D-07..D-11), 01-REVIEW.md/01-REVIEW-FIX.md (WR-01/WR-02)

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | D-07 fix belongs in the graph layer (BuildGraph C1 branch) | Pitfall 1 | If expansion decision moves, guard placement differs — behavior identical |
| A2 | Existing tests may already cover box sub-diagrams and actors in C2 | Wave 0 Gaps | If not, tests must be added — plan tasks cover both cases |

## Open Questions (RESOLVED)

None — all areas decided in discussion; open items are test-coverage confirmations (A2).

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — zero new dependencies
- Architecture: HIGH — verified against actual code with file:line refs
- Pitfalls: HIGH — grounded in Phase 1 execution history

**Research date:** 2026-08-06
**Valid until:** 2026-09-05
