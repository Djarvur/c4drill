# Phase 38 Research: Hierarchy Wrapping and Granular Keys

**Researched:** 2026-08-30 (inline — researcher spawn unavailable: MCP `exa` not connected; fix points verified against the post-v1.14 codebase during phase 37 execution)
**Baseline:** v1.21.0 (Phase 37 landed: recursive buildNestedCluster dispatch, Entry.UnfoldChain + ensureDeepLinkChain, Cluster.ExploreURL drill affordance, --plain threading View.Plain → Graph.Plain + builder guards, plain-label converter branches).

## WRAP — wrapping boundary/sibling nodes and the expanded unit in ancestor containers

Current behavior (post-v1.14):
- Boundary/sibling entries are created by `createExternalBoundaryNode` (internal/view/scope.go) with `IsBoundary: true` and rendered at TOP level by `buildBoundaryViewGraph` (`if entry.IsBoundary || entry.IsExternal { node := buildNode(entry); continue }`) — this was the v1.14 scoping decision the user reversed. Regular and deep-link-wrapped content already nests; boundary nodes and the expanded unit's own ancestors do not.
- The expanded unit renders as ONE top-level boundary cluster (`buildBoundaryCluster`, ID = v.ExpandedUnit); its ancestors (e.g. in a C3 view of `mainSystem.storages.localStorage`: `mainSystem` ⊃ `storages`) appear only as breadcrumb text (`View.AncestorNames`), never as clusters.
- Fully external units (top-level in the model) have no in-model ancestors — nothing to wrap.

Recommended design — ONE rule, applied uniformly:
- Build the full top-down container skeleton for each view: wrapper clusters for the expanded unit's ancestor chain (root → parent), then the expanded unit's cluster, and every boundary entry wrapped in ITS ancestor chain. Boundary paths are short (resolveBoundaryByDivergence returns ≤ common+1 depth), so their wrap chains are 1–2 clusters and typically MERGE with the expanded unit's ancestor skeleton (e.g. boundary `mainSystem.sshAuth` lands inside the same `mainSystem` wrapper cluster as the expanded `mainSystem.storages.localStorage` subtree).
- Entry points: `buildBoundaryCluster` + `buildBoundaryViewGraph` (internal/graph/builder.go) gain wrapper-cluster construction (reuse `buildCluster`/cluster-label conventions; ExploreURL on wrapper clusters is NOT wanted — wrappers are containers, not drill targets; the expanded cluster keeps its own affordances).
- Edge endpoints, node IDs, dedup/multiplicity keys unchanged — wrapping is cluster-structure only. `AncestorNames` already carries pretty names for wrapper labels.
- WRAP-03 lock: count/assert the depicted node set before/after (unit test over the multilevel fixture: no new node IDs, only new cluster IDs).

## KEY — granular switches

- Precedent: --plain flag (cmd/c4drill/root.go) → View.Plain → Graph.Plain → guards. Recommended: keep `Plain` as the union and add the aspect fields on the same structs (a `RenderOpts` struct or sibling bools — planner pins, but thread the same way; cover BuildGraph AND BuildExpandedGraph AND BuildGraphWithPath's copy).
- Switch → suppression mapping (pin): `--no-colors` (unit Color/Border fills + link Color + kind-derived colours — i.e. ALL colouring; legend kind rows adjust emergently via legendKindEntries reading final edge colours), `--no-styles` (unit Style/BorderStyle + link Style incl. the applyCollapsedPairStyle aggregate override), `--no-length` (link Length/minlen), `--no-rank` (RankReverse swap + NoConstraint — BOTH render emission sites converter.go ~555/~605/~652). `properties.edges` stays tied to --plain only (not a granular aspect; document).
- KEY-02 lock: existing plain goldens (cmd/c4drill/testdata/plain.dot, plain.expanded.dot) must stay green with `--plain` unchanged.

## LBL — label suppression

- `--no-labels`: omit label TEXT everywhere — nodes, edges, clusters. GraphViz needs shape from label for the record/HTML shapes (person 2-col table etc.), so labels-off means units render as their plain default shape (ShapeForType) with an empty label; edges emit no label; clusters emit no label HTML. Legend is metadata: DEFAULT stays (its own properties.legend setting governs); pin + document.
- Implementation points: buildNode/buildClusterLabel label construction or the converter setNodeLabel/setClusterLabel/edge-label branches (same dispatch sites as v1.14's plain branches — add the NoLabels branch there). Thread like Plain.
- Layout note: suppression must happen at GRAPH layer (drop Label content before emission) so dot itself lays out without label geometry — do not merely hide text in SVG post-processing.
- Applies to all generations incl. --expanded (processExpandedView must thread the flag like Plain).

## BC / goldens

- WRAP changes output for models with boundary nodes — REAL re-baseline expected: cmd/c4drill/testdata/multilevel{,.expanded}.dot, plain.dot/plain.expanded.dot (plain goldens re-render under new structure too), plus in-package canonical expectations. ONE consolidated re-baseline task; diffs audited as cluster-structure-only.
- KEY/LBL are opt-in: default path unchanged except WRAP. New additive goldens for the flags.
- canonical.Canonical (DI-1) everywhere; multilevel.toml remains the E2E fixture.

## Patterns

Reuse phase 37's analog map: `.planning/milestones/v1.13-phases` archive is stale; the live reference is `.planning/phases/37-nesting-context-and-plain-rendering/37-PATTERNS.md` (moved: now committed at that path) — buildNestedCluster dispatch, AllExpanded/Plain field threading, --expanded cobra convention, canonical goldens, skill 3-copy diff -r parity.

## Validation Architecture

| Requirement class | Validation vehicles |
|---|---|
| WRAP-01/02 | scope_test/builder_test: boundary entries wrapped in ancestor clusters; fully external stay top-level; wrapper skeleton for expanded ancestors; edge endpoints unchanged |
| WRAP-03 | node-set invariance test (IDs before == after wrapping) over multilevel fixture |
| KEY-01/02/03 | per-switch guard tests; --plain golden parity (existing goldens green); composition matrix E2E (switch × generation × format spot-checks) |
| LBL-01/02/03 | label-absence assertions in DOT (no HTML tables, empty labels, minimal shapes); --expanded composition test; legend default-stays test |
| BC-01 | one consolidated golden re-baseline (WRAP deltas only) + full suite green |
| DOC-01..03 | README/skill sections, diff -r ×2, fixtures render through pipeline |
| REL-01 | tag v1.22.0 + release workflow green |

Manual-only: none beyond visual spot-check folded into the re-baseline audit (wrapper clusters legible).
