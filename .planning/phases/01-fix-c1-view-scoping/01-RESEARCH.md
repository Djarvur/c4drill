# Phase 1: Fix C1 View Scoping - Research

**Researched:** 2026-08-06
**Domain:** Go graph-view resolution & edge deduplication (C4 diagram rendering pipeline)
**Confidence:** HIGH

> **Research provenance note:** The gsd-phase-researcher agent could not be spawned (its
> definition requires mcp__firecrawl__* / mcp__exa__* MCP servers, which are not connected;
> web research is also disabled in config: `brave_search: false, firecrawl: false, exa_search: false`).
> The orchestrator produced this RESEARCH.md directly from codebase analysis performed during
> discuss-phase. All claims are tagged `[VERIFIED: <file>:<line>]` against the actual code.
> No external-library or ecosystem claims are made — the phase adds zero dependencies.

## Summary

Phase 1 refines an ALREADY-IMPLEMENTED C1 view-scoping system (commits `f9dc69a`, `c5091f9`,
`8f7e21c`, `b6771b0`; STATE.md declares v1.8 complete; tests pass). The existing implementation
resolves deep link targets to top-level ancestors (`resolveToTopLevel`), collects resolved links
on entries (`ResolvedLinks`/`ResolvedLinksFrom`), and dedups edges in `buildEdges` by
`source->target:technology:description`.

The 13 locked decisions (D-01..D-13) refine three areas: (1) **pair-only edge collapse** with
first-wins attributes and binary penwidth thickening, scoped to C1/C2/C3 — `--expanded` exempt
to preserve COMPAT-02 byte-identical output; (2) **deepest-visible-ancestor resolution on both
source and target sides** so edges reach visible subunits inside expanded clusters (aligning C1
with the existing C2/C3 `resolveToViewAncestor` semantics) and edges render inside clusters;
(3) **cleanup**: skip-unresolved peers, remove legacy `addExternalBoundaryNodes*`, and gate
minlen to original-pair edges only.

**Primary recommendation:** Modify `internal/view/scope.go` resolution passes and
`internal/graph/builder.go` `buildEdges`/`createEdge`, keeping all changes behind the existing
`ResolvedLinks` override mechanism. Add an `AllExpanded` flag to `view.View` so
`BuildExpandedGraph` retains v1.7 dedup/penwidth behavior.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### Edge Collapse & Deduplication (EDGE-02)
- **D-01:** Pair-only collapse — one edge per (source, target) pair regardless of technology/description. The first contributing link in definition order wins.
- **D-02:** Pair-only collapse applies to resolved views (C1/C2/C3), where edge resolution happens. **`--expanded` mode is exempt** — it keeps the v1.7 tech+desc dedup key so COMPAT-02 output stays identical to v1.7.
- **D-03:** The surviving collapsed edge inherits ALL attributes first-wins (color, style, arrow, length, label position) from the first link in definition order.
- **D-04:** Collapsed edges (2+ relationships collapsed onto the pair) get a binary thicker penwidth (2.0). Single-relationship edges stay at default width.
- **D-05:** Multiplicity counts ALL contributing links (direct links + resolved sub-links) that land on the pair.
- **D-06:** Collapsed edge labels stay plain — first-wins tech/description, no count suffix, no aggregation. Thicker penwidth is the only multiplicity signal.

#### Edge Resolution Semantics (VIEW-01/02, EDGE-01)
- **D-07:** Deepest visible ancestor for targets — when a link target's parent is expanded in C1 (target visible as a node inside the cluster), the edge points at the visible subunit directly. Hidden targets resolve to the top-level ancestor as today. Aligns C1 with C2/C3's existing `resolveToViewAncestor` semantics.
- **D-08:** Replace parent edge — when a link resolves to a visible subunit, the parent-level edge is NOT drawn. The parent edge appears only when a real direct link to the parent exists. No redundant edges.
- **D-09:** Source-side resolution — the source of a link also resolves to its deepest visible ancestor (direct subunit node when the source is inside an expanded top-level unit; the top-level unit otherwise). E.g. `linuxSystem.sshAuth → keycloak` renders as `sshAuth → keycloak`.
- **D-10:** Draw within cluster — links whose source AND target are both visible inside the same expanded top-level unit (e.g. `sshAuth → webAPI`) render as an edge between the two nodes inside the cluster (currently skipped as "internal").
- **D-11:** Box grouping units follow the SAME resolution rules as systems — no special-casing.

#### Boundary Nodes & Dead Code
- **D-12:** Skip unresolved links in the view layer — the validator is the single gatekeeper for undefined targets. Remove the legacy `addExternalBoundaryNodes` + `addExternalBoundaryNodesRecursive` (`internal/view/scope.go`), which are unreachable/no-op for valid models.

#### minlen Semantics
- **D-13:** A link's `length` (GraphViz minlen) takes effect ONLY when both drawn endpoints ARE the link's original units (no resolution on either side). Resolved edges and collapsed edges carry no minlen. For collapsed edges, the surviving first link's minlen applies only if its own endpoints were both original.

### Claude's Discretion
None — every area was decided with concrete options.

### Deferred Ideas (OUT OF SCOPE)
None — discussion stayed within phase scope.

</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| VIEW-01 | C1 diagram shows only top-level units — no nested subunits appear as nodes | Existing `GenerateC1View` + `addC1BoundaryNodes` already deliver this; D-12 removes the legacy path that could violate it |
| VIEW-02 | Links to deeply nested targets resolve to the nearest visible ancestor in the current view level | D-07/D-09 extend `resolveToTopLevel` to deepest-visible-ancestor (both sides), aligning with `resolveToViewAncestor` |
| EDGE-01 | Edges from nested subunits to targets outside the current view resolve to the visible ancestor | D-09 source-side resolution + D-10 within-cluster edges |
| EDGE-02 | Duplicate edges (multiple sub-links resolving to same ancestor pair) collapse into a single edge | D-01..D-06: pair-only dedup key, first-wins attributes, binary penwidth, plain labels |

</phase_requirements>

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Link resolution (deepest visible ancestor) | `internal/view` (scope.go) | — | Resolution operates on model paths + view visibility — view's job |
| Edge deduplication (pair-only) | `internal/graph` (builder.go) | `internal/view` (ResolvedLinks) | buildEdges owns edge construction; consumed links come from view |
| Edge attributes (penwidth, minlen) | `internal/graph` (builder.go) → `internal/render` (converter.go) | — | createEdge sets attributes; converter.go:478 serializes penwidth |
| `--expanded` compat | `internal/view` + `internal/graph` | — | GenerateExpandedView + BuildExpandedGraph share buildEdges — need mode discriminator |
| Undefined-target rejection | `internal/validator` (rules.go) | — | Single gatekeeper per D-12 (unchanged) |

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| Go stdlib (`slices`, `strings`) | 1.26.1 | Path walking, membership checks | Already used in scope.go/builder.go |
| stretchr/testify | v1.11.1 | Assertions in new tests | Existing test convention |
| pelletier/go-toml/v2 | v2.2.4 | TOML parsing (unchanged) | Existing |

### Supporting
None — this phase adds **zero new dependencies**. All changes are to existing view/graph code.

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Extending `resolveToTopLevel` | New standalone resolver package | Overkill — 3 functions to unify, not a subsystem |

## Architecture Patterns

### Pattern 1: Resolved-link override
`buildEdges` prefers `entry.ResolvedLinks` over `entry.Unit.Links` when present
(`internal/graph/builder.go:339-347`). All resolution refinements (D-07..D-10) plug into this
existing seam — resolved links are synthesized `model.Link` values collected on the source
entry. The multiplicity count for D-05 lives here: counting requires the view to record how
many contributing links landed on each pair.

### Pattern 2: Deepest-visible-ancestor walk
`resolveToViewAncestor` (`internal/view/scope.go:716-742`) already implements the exact walk
needed for C1: check peer, then walk up path components until a unit in `v.Units` is found.
C1's `resolveToTopLevel` (`scope.go:368-386`) is the simpler top-level-only variant. D-07/D-09
unify C1 onto the deepest-visible walk with an expanded-visibility predicate.

### Pattern 3: Definition-order iteration
All passes iterate via `UnitOrder`/`SubunitOrder` slices captured by the parser
(`internal/parser/parser.go:100`). "First in definition order wins" (D-01/D-03/D-05) is
deterministic because `resolveAndAddBoundary` (`scope.go:278`) traverses units in
`UnitOrder` and links in slice order.

## Common Pitfalls

### Pitfall 1: C2/C3 cross-subunit regression
**What goes wrong:** Changing the dedup key or resolution shared by C2/C3 passes
(`resolveBoundaryNodeLinks`, `resolveSubunitCrossLinks`, `resolveDescendantCrossLinks`,
`scope.go:652-859`) breaks the cross-subunit edge synthesis (`dacProxy → authModules`)
that v1.8 currently passes.
**How to avoid:** Keep C2/C3 resolution passes untouched; only change the dedup key in
`buildEdges` (shared) and C1's resolution entry points. Verify with existing
`scope_test.go`/`builder_test.go` cross-subunit tests.
**Warning signs:** C2/C3 edge counts change in existing tests.

### Pitfall 2: `--expanded` byte-compat (COMPAT-02)
**What goes wrong:** `BuildExpandedGraph` calls the SAME `buildEdges` (`builder.go:170`).
A pair-only key or penwidth change leaks into `--expanded` output → differs from v1.7.
**How to avoid:** Add an explicit mode discriminator — recommend `view.View.AllExpanded bool`
set `true` in `GenerateExpandedView` (`scope.go:14`) — and branch the dedup key and penwidth
on it. Baseline test: `cmd/c4drill/testdata` expanded output unchanged.
**Warning signs:** Baseline SVG diff after the change.

### Pitfall 3: Penwidth semantics (D-04 grounding)
**What goes wrong:** ALL edges currently render at penwidth 2.0 (`converter.go:478-479`,
set unconditionally in `buildCgraph`). "Default" in D-04 means GraphViz default 1.0;
"thicker" = 2.0 (the v1.7 prominence width). If single edges stay 2.0, collapsed edges need
a NEW higher width — contradicting D-04's binary wording.
**Resolution:** Single edges → penwidth 1.0, collapsed edges → 2.0, in resolved views
(C1/C2/C3). `--expanded` keeps 2.0 everywhere (COMPAT-02). The `graph.Edge` needs a
`PenWidth` field (default 0 → renderer picks 1.0/2.0 by mode), or a boolean
`CollapsedMultiplicity` flag consumed by `converter.go:478`.

### Pitfall 4: minlen leak (D-13)
**What goes wrong:** `createEdge` copies `link.Length` to `Edge.MinLen` unconditionally
(`builder.go:454`). Resolved links inherit the original link's length → spacing hint applies
to an ancestor edge it was never tuned for.
**How to avoid:** `buildEdges`/`processOutgoingLinks` already distinguish resolved vs direct
links (they select `ResolvedLinks` vs `Unit.Links`). Thread a `resolved bool` (or a
`Synthetic bool` on the edge) into `createEdge`; skip `MinLen` when the edge is resolved.
Collapsed pair edges: first-wins minlen only when the surviving link was direct.
**Warning signs:** Edge spacing changes on resolved ancestor edges.

### Pitfall 5: Expanded-unit visibility bookkeeping (D-07..D-10)
**What goes wrong:** `GenerateC1View` adds only top-level units to `v.Units` (`scope.go:88-135`);
`buildCluster` renders direct subunits of expanded units as nodes (`builder.go:265-311`) but
those subunit paths are NOT in `v.Units`. Resolution cannot distinguish "visible subunit" from
"hidden" without new bookkeeping.
**How to avoid:** Track visible subunit paths explicitly — e.g., add `VisiblePaths map[string]bool`
(or entries flagged `IsInCluster`) populated from `isExpandedInC1` results, and make the C1
resolution helper consult it: a peer resolves to its deepest ancestor in `v.Units` OR in
`VisiblePaths`. Also required for source-side resolution (D-09) and within-cluster edges (D-10),
which need the source/target pair to both resolve to visible subunits — note D-10's
within-cluster edges currently fall into the `resolved == sourceAncestor → skip` branch
(`scope.go:300`), so the skip condition must become "resolved source == resolved target"
rather than ancestor equality.
**Warning signs:** Edges from/to cluster labels instead of child nodes in expanded C1.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Peer-path → visible-ancestor resolution | A fourth resolution variant | Unify on the existing `resolveToViewAncestor` walk (scope.go:716) with an expanded-visibility predicate | ARCHITECTURE.md already flags duplicate peer-resolution logic as an anti-pattern; D-12 removes one variant, don't add another |

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go standard `testing` + stretchr/testify v1.11.1 |
| Config file | none — `go test` convention (`.mise.toml` task: `go test -v -race -cover ./...`) |
| Quick run command | `go test ./internal/view/ ./internal/graph/` |
| Full suite command | `go test -v -race -cover ./...` (render tests must NOT use `t.Parallel()` — WASM mutex) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| VIEW-01 | C1 has only top-level units + resolved boundary nodes | unit | `go test ./internal/view/ -run TestC1 -count=1` | ✅ scope_test.go (extend) |
| VIEW-02 | Deep link target resolves to deepest visible ancestor (expanded parent → subunit edge) | unit | `go test ./internal/view/ -run TestC1 -count=1` | ❌ new tests |
| EDGE-01 | Source-side resolution: edge from visible subunit, not parent | unit | `go test ./internal/view/ -run TestC1 -count=1` | ❌ new tests |
| EDGE-02 | Pair-only collapse: distinct tech/desc → one edge, first-wins | unit | `go test ./internal/graph/ -run TestBuildEdges -count=1` | ✅ builder_test.go (extend) |
| EDGE-02 | Collapsed edge penwidth 2.0 vs single 1.0 | unit | `go test ./internal/graph/ -run TestCollapse -count=1` | ❌ new tests |
| COMPAT-02 (constraint) | `--expanded` output identical to v1.7 baseline | integration | `go test ./cmd/c4drill/ -count=1` | ✅ root_test.go baseline |

### Sampling Rate
- **Per task commit:** `go test ./internal/view/ ./internal/graph/ ./cmd/c4drill/`
- **Per wave merge:** `go test -v -race ./...` (or `mise test`)
- **Phase gate:** Full suite green + lint (`mise lint`) before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] New tests for deepest-visible-ancestor resolution in C1 (expanded units)
- [ ] New tests for source-side resolution and within-cluster edges
- [ ] New tests for pair-only collapse + penwidth + minlen gating
- [ ] `--expanded` baseline output test remains green (COMPAT-02)

*(Existing test infrastructure covers the base behavior — no framework gaps)*

## Security Domain

### Applicable ASVS Categories
| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | no | — (local CLI, no users/sessions) |
| V3 Session Management | no | — |
| V4 Access Control | no | — |
| V5 Input Validation | yes | Existing parser/validator pipeline (unchanged); TOML is local trusted input |
| V6 Cryptography | no | — |

### Known Threat Patterns
| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Malformed TOML causing panic | DoS | Existing typed `ParseError`/`ValidationErrors` pipeline rejects before view/graph (unchanged); new resolution code must nil-guard like existing passes |
| Path traversal in dotted output paths | Tampering | Pre-existing `internal/output` concern — unchanged by this phase |

**Threat model note for PLAN.md:** This phase modifies internal graph/view logic only; the CLI
entry, parser, and validator are untouched. The `[BLOCKING]`-level threat is none; keep the
standard nil-guard + no-panic contract in new code.

## Sources

### Primary (VERIFIED — codebase reads this session)
- `internal/view/scope.go` — GenerateC1View (:88), isExpandedInC1 (:144), addC1BoundaryNodes (:253), resolveAndAddBoundary (:278), resolveToTopLevel (:368), resolveToViewAncestor (:716), resolveSubunitCrossLinks (:750), resolveDescendantCrossLinks (:794), GenerateExpandedView (:14), addExternalBoundaryNodes (:154), createExternalBoundaryNode (:220), addUnitRecursive (:55)
- `internal/graph/builder.go` — BuildGraph (:15), buildBoundaryCluster (:79), BuildExpandedGraph (:125), buildNestedCluster (:172), buildCluster (:265), buildEdges (:329), processOutgoingLinks (:362), processIncomingLinks (:388), isTargetInView (:414), createEdge (:423), markSeen (:460), BuildGraphWithPath (:474)
- `internal/render/converter.go:478` — unconditional penwidth 2.0 (D-04 grounding)
- `cmd/c4drill/root.go` — collectExpandedPaths (:155), collectExpandableUnitPaths (:170)
- `internal/validator/rules.go:26` — undefined-reference rejection (D-12 basis)
- `.planning/codebase/ARCHITECTURE.md` — pipeline structure, anti-patterns (duplicate peer-resolution logic)
- `.planning/codebase/CONVENTIONS.md` — lint/test conventions
- `01-CONTEXT.md` — 13 locked decisions

### Secondary
- `01-DISCUSSION-LOG.md` — alternatives considered per area

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | "Default width" in D-04 = GraphViz default 1.0 (current code forces 2.0 on all edges) | Penwidth semantics | Output width changes in resolved views — but this IS the refinement intent; --expanded stays 2.0 per COMPAT-02 |
| A2 | `view.View.AllExpanded` flag is the cleanest `--expanded` discriminator | Pattern 3 | Alternative: pass mode into buildEdges as a parameter — planner may choose either; behavior identical |

## Open Questions

1. **Where to record multiplicity for D-05/D-04?**
   - What we know: resolution happens in `internal/view` (scope.go), edge construction in `internal/graph` (builder.go). The penwidth needs the count at createEdge time.
   - What's unclear: whether the count lives on `view.Entry.ResolvedLinks` items (parallel slice / struct wrapper) or on the synthesized link (e.g., a `Multiplicity int` on the edge-side record).
   - Recommendation: planner picks the least invasive — a per-pair count map computed in `buildEdges` from the resolved+direct link lists, keyed by pair, consulted at `createEdge`. Avoids touching the model package.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — zero new dependencies, verified against go.mod
- Architecture: HIGH — all claims verified against actual code with file:line refs
- Pitfalls: HIGH — grounded in code paths read this session

**Research date:** 2026-08-06
**Valid until:** 2026-09-05 (stable — internal codebase, no external deps)
