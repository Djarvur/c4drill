# Phase 1: Fix C1 View Scoping - Pattern Map

**Mapped:** 2026-08-06
**Files analyzed:** 8 (all EXISTING files — refinement phase over v1.8 implementation)
**Analogs found:** 8 / 8 (self-analog / sibling-pass analog — every modification target already exists)

> This is a REFINEMENT phase: every file below already exists and already implements a first
> version of the behavior. The "analog" for each file is its own current implementation plus its
> sibling passes (C2/C3 resolution in `scope.go`, `--expanded` mode in `builder.go`). The planner
> should produce DELTA actions against the exact line ranges cited, not rewrites.

## Decision → Function Map (D-01..D-13)

| Decision | Primary site (file:line) | Supporting site |
|----------|--------------------------|-----------------|
| D-01 pair-only dedup key | `internal/graph/builder.go:377` and `:403` (edgeKey in processOutgoing/processIncoming) | buildEdges `:329-359` |
| D-02 `--expanded` exempt | `internal/view/view.go:20` (add `AllExpanded bool` to View) | `internal/view/scope.go:19` (GenerateExpandedView View literal), `builder.go:329` (dedup key branch), `converter.go:478` (penwidth branch) |
| D-03 first-wins attrs | `builder.go:423` createEdge (already copies all link attrs) + `builder.go:460` markSeen (already rejects later duplicates) | `scope.go:301-311` / `:330-340` resolvedLink field copy |
| D-04 binary penwidth | `internal/graph/graph.go:74` (add field to Edge) + `builder.go:423` createEdge | `internal/render/converter.go:478-479` (conditional SetPenWidth) |
| D-05 multiplicity count | `builder.go:329` buildEdges (per-pair count map over merged link lists) | `scope.go:312`/`:341` (ResolvedLinks appends per contributing link — count survives because resolveAndAddBoundary does NOT dedup) |
| D-06 plain labels | `builder.go:427-434` (no change; verify no count suffix added) | — |
| D-07 deepest-visible-ancestor targets | `internal/view/scope.go:368` resolveToTopLevel (replace/unify) | `scope.go:716` resolveToViewAncestor (existing walk to reuse), `scope.go:88` GenerateC1View (populate visible-subunit bookkeeping) |
| D-08 replace parent edge | `scope.go:300` / `:329` (append condition in resolveAndAddBoundary) | — |
| D-09 source-side resolution | `scope.go:280-283` (sourceAncestor in resolveAndAddBoundary) | visible-subunit bookkeeping (Pitfall 5) |
| D-10 within-cluster edges | `scope.go:300` (skip condition `resolved != sourceAncestor` must become "resolved source == resolved target") | `builder.go:372`/`:397` (isTargetInView must accept visible subunit paths) |
| D-11 box grouping | Same sites as D-07..D-10 — no special-case branch for `model.TypeBox` | `model/unit.go:16` TypeBox |
| D-12 remove legacy boundary nodes | `scope.go:49` (call in GenerateExpandedView) + delete `scope.go:154-174` + `scope.go:178-215` | `internal/validator/rules.go:14-45` (single gatekeeper, unchanged); KEEP `createExternalBoundaryNode` `scope.go:220-245` (used by `:295`/`:324`/`:642`) |
| D-13 minlen gate | `builder.go:454` (`edge.MinLen = link.Length`) + `builder.go:362`/`:388` (thread resolved-flag into createEdge) | `converter.go:474-476` (renderer only serializes `MinLen > 0` — no change needed there) |

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/view/scope.go` | view generator (resolution passes) | transform (path→visible-ancestor resolution) | `resolveToViewAncestor` (scope.go:716) + `resolveBoundaryNodeLinks` (scope.go:652) | exact (sibling pass) |
| `internal/view/view.go` | model (struct defs) | data | itself (View/Entry structs) | exact |
| `internal/graph/builder.go` | graph builder (edge construction) | transform (view→graph) | itself (buildEdges/createEdge/markSeen) | exact |
| `internal/graph/graph.go` | model (struct defs) | data | itself (Edge struct) | exact |
| `internal/render/converter.go` | renderer (graph→cgraph) | transform | itself (createEdge at :426) | exact |
| `internal/view/scope_test.go` | test | — | itself (TestIntegrationC1EdgeResolution in sibling integration_test.go:468) | exact |
| `internal/graph/builder_test.go` | test | — | itself (TestBuildGraphMultipleLinks :252, TestBuildGraphEdgeLength :928) | exact |
| `cmd/c4drill/root_test.go` | test (COMPAT-02 guard) | — | itself (expanded.toml fixtures :488, :553, :574, :603) | exact |

## Pattern Assignments

### `internal/view/scope.go` (view generator, transform)

**Analog:** `resolveToViewAncestor` (scope.go:716-742) — the deepest-visible-ancestor walk D-07/D-09 must unify onto; `resolveBoundaryNodeLinks` (scope.go:652-711) — the C2/C3 resolved-link synthesis pattern.

**D-07/D-09 target pattern — deepest-visible-ancestor walk to reuse (scope.go:716-742):**
```go
// resolveToViewAncestor resolves a peer path to its nearest ancestor
// (or itself) that is present in the view.
// Returns "" if no ancestor is found in the view.
func resolveToViewAncestor(v *View, peer string) string {
	// Check if peer itself is in the view
	if _, exists := v.Units[peer]; exists {
		return peer
	}

	// Walk up the peer's path to find nearest visible ancestor
	for {
		idx := strings.LastIndex(peer, ".")
		if idx <= 0 {
			break
		}

		peer = peer[:idx]

		if _, exists := v.Units[peer]; exists {
			return peer
		}
	}

	// Check if the top-level peer (no dots) is in the view
	if _, exists := v.Units[peer]; exists {
		return peer
	}

	return ""
}
```
Executor guidance: C1's `resolveToTopLevel` (scope.go:368-386) is the simpler top-level-only variant to REPLACE. The unified helper must add a visibility predicate so a peer also resolves to a visible subunit path (an expanded unit's direct child rendered by `buildCluster`). Per Pitfall 5 those paths are NOT in `v.Units` — add `VisiblePaths map[string]bool` to `View` (view.go:20), populated in `GenerateC1View` (scope.go:119-126) from `isExpandedInC1` results (walk each expanded entry's `SubunitOrder`, add `path+"."+subName`).

**D-09/D-10 target pattern — current source-ancestor + append logic to change (scope.go:278-314):**
```go
// resolveAndAddBoundary recursively scans a unit and all its subunits for links.
func resolveAndAddBoundary(v *View, m *parser.Model, path string, unit *model.Unit) {
	// Find the top-level ancestor of this path (source for resolved links)
	sourceAncestor := path
	if idx := strings.Index(path, "."); idx > 0 {
		sourceAncestor = path[:idx]
	}

	sourceEntry := v.Units[sourceAncestor]

	// Process outgoing links
	for _, link := range unit.Links {
		resolved := resolveToTopLevel(v, link.Peer)
		if resolved == "" {
			continue // Internal link (both sides under same ancestor or self-referencing)
		}

		if _, exists := v.Units[resolved]; !exists {
			v.Units[resolved] = createExternalBoundaryNode(m, resolved, path)
			v.UnitOrder = append(v.UnitOrder, resolved)
		}

		// Add resolved outgoing link to the source ancestor
		if sourceEntry != nil && resolved != sourceAncestor {
			resolvedLink := model.Link{
				Peer:          resolved,
				Technology:    link.Technology,
				Description:   link.Description,
				Style:         link.Style,
				Arrow:         link.Arrow,
				Rank:          link.Rank,
				LabelPosition: link.LabelPosition,
				Color:         link.Color,
				Length:        link.Length,
			}
			sourceEntry.ResolvedLinks = append(sourceEntry.ResolvedLinks, resolvedLink)
		}
	}
	// ... same structure for unit.LinksFrom (:317-343), then recursion :346-362
}
```
Executor guidance:
- D-09: `sourceAncestor` (`:280-283`) must become the deepest VISIBLE ancestor of `path` (walk `path` segments; stop at first segment in `v.Units` OR `v.VisiblePaths`). Append the resolved link to THAT entry (may be a subunit entry inside an expanded cluster).
- D-10: the skip condition `resolved != sourceAncestor` (`:300`, `:329`) becomes "resolvedSource == resolvedPeer" — when both resolve inside the SAME expanded unit the edge must still be recorded (currently dropped as "internal"). Keep the `resolved == ""` internal-link skip.
- D-08: follows automatically — the parent-level edge is not added for a link that resolved to a subunit; the parent edge only appears when a direct link to the parent exists (its resolved peer IS the parent).
- D-03/D-05: the `model.Link` copy (`:301-311`, `:330-340`) copies ALL attributes first-wins; the per-contributing-link append (no peer dedup here — unlike `addResolvedCrossLink` `:838-842`) preserves the multiplicity count for buildEdges.

**D-12 target pattern — legacy code to delete (scope.go:49, :154-174, :178-215):**
```go
	// Add external boundary nodes for referenced units not in the model
	addExternalBoundaryNodes(v, m)   // scope.go:49 — DELETE call, set v.AllExpanded = true instead

// addExternalBoundaryNodes — scope.go:154-174 — DELETE entire function
// addExternalBoundaryNodesRecursive — scope.go:178-215 — DELETE entire function
```
Executor guidance: keep `createExternalBoundaryNode` (scope.go:220-245) — still called by `addC1BoundaryNodes` (:295, :324) and `addResolvedBoundaryNode` (:642). Validator gatekeeper `internal/validator/rules.go:14-45` (`ValidateReferences`) already rejects undefined peers, so the removed path is unreachable for valid models.

### `internal/view/view.go` (model, data)

**Analog:** itself. Add fields with doc comments (every struct field commented per CONVENTIONS.md).

- `View` struct (view.go:20-41): add `AllExpanded bool` (D-02 — set true only in `GenerateExpandedView`, consumed by buildEdges dedup key + penwidth branch) and `VisiblePaths map[string]bool` (Pitfall 5 — visible subunit paths inside expanded top-level clusters).
- `Entry` (view.go:45-62): existing `ResolvedLinks`/`ResolvedLinksFrom` fields (`:58`, `:61`) are the seam — unchanged; if the executor prefers an `IsInCluster` flag over `VisiblePaths` it must be added here instead.

### `internal/graph/builder.go` (graph builder, transform)

**Analog:** itself — buildEdges/createEdge/markSeen are the exact functions D-01..D-06/D-13 modify.

**D-01/D-02 target pattern — dedup key + override seam (builder.go:329-359):**
```go
// buildEdges creates edges from view links.
func buildEdges(v *view.View) []*Edge {
	edges := make([]*Edge, 0)
	seen := make(map[string]bool) // Track processed links

	// Process edges in definition order
	for _, path := range v.UnitOrder {
		entry := v.Units[path]

		// Use resolved links when available (for C1 views with edge resolution),
		// otherwise fall back to the unit's direct links.
		outLinks := entry.Unit.Links
		if entry.ResolvedLinks != nil {
			outLinks = entry.ResolvedLinks
		}

		inLinks := entry.Unit.LinksFrom
		if entry.ResolvedLinksFrom != nil {
			inLinks = entry.ResolvedLinksFrom
		}

		// Process outgoing links
		outEdges := processOutgoingLinks(path, outLinks, v.Units, seen)
		edges = append(edges, outEdges...)

		// Process incoming links (linkFrom)
		inEdges := processIncomingLinks(path, inLinks, v.Units, seen)
		edges = append(edges, inEdges...)
	}

	return edges
}
```
Executor guidance: the `ResolvedLinks != nil` override (`:339-347`) is the extension point for ALL resolution refinements. For D-05, build a per-pair count map here by iterating `outLinks`/`inLinks` BEFORE dedup and counting occurrences of the pair key; consult it in `processOutgoingLinks`/`processIncomingLinks` (or pass it down) so `createEdge` can set penwidth. For D-13, thread a `resolved bool` (links came from `ResolvedLinks`) — but note GenerateExpandedView never populates ResolvedLinks, so `v.AllExpanded` alone can also discriminate; the pair-count and minlen logic must branch on `v.AllExpanded` first (D-02).

**D-01/D-04/D-13 target pattern — edge construction sites (builder.go:362-411, :423-468):**
```go
// processOutgoingLinks processes outgoing links from a unit.
func processOutgoingLinks(
	path string,
	links []model.Link,
	viewUnits map[string]*view.Entry,
	seen map[string]bool,
) []*Edge {
	edges := make([]*Edge, 0)
	sourceEntry := viewUnits[path] // Get source entry for color lookup

	for _, link := range links {
		if !isTargetInView(viewUnits, link.Peer) {
			continue
		}

		edge := createEdge(path, link.Peer, link, sourceEntry)
		edgeKey := path + "->" + link.Peer + ":" + link.Technology + ":" + link.Description

		if markSeen(seen, edgeKey) {
			edges = append(edges, edge)
		}
	}

	return edges
}
```
- D-01: `edgeKey` (`:377`, and `:403` for incoming: `link.Peer + "->" + path + ":" + ...`) drops the `:technology:description` suffix → `path + "->" + link.Peer` for resolved views; keep the full key when `v.AllExpanded` (D-02, COMPAT-02). First-wins is automatic: `buildEdges` iterates `v.UnitOrder` and links in slice order, and `markSeen` (`:460-468`) rejects the 2nd+ occurrence.
- D-10: `isTargetInView(viewUnits, link.Peer)` (`:372`, `:397`) must also accept paths in `v.VisiblePaths` — change signature to take the view (or the visibility set) so within-cluster edges pass the gate.
- D-13: `createEdge` (`:423-457`) — the MinLen copy at `:454` (`edge.MinLen = link.Length`) becomes conditional: skip when the edge is resolved (any side) OR when `v.AllExpanded` is false and the pair is collapsed with a resolved surviving link.

**D-04 penwidth — Edge struct (graph.go:74-89):**
```go
// Edge represents a connection between two nodes.
type Edge struct {
	// Source is the source node ID.
	Source string
	// Target is the target node ID.
	Target string
	// Label contains edge label information.
	Label *EdgeLabel
	// Style is the line style (solid, dashed, dotted).
	Style string
	// ArrowHead is the arrow direction.
	ArrowHead ArrowDirection
	// Color is the edge line and label color (from source unit's border color or explicit override).
	Color string
	// MinLen is the minimum length (minlen attribute) for the edge.
	MinLen int
}
```
Add `PenWidth float64` (0 = renderer default) — set 2.0 only when the pair collapsed (2+ contributors, D-04/D-05) in resolved views; leave 0 elsewhere so the renderer uses 1.0 in resolved views and 2.0 in `--expanded` (D-02/COMPAT-02). Per RESEARCH Pitfall 3, single edges must render at 1.0 in resolved views — the current unconditional 2.0 at converter.go:478-479 is what changes.

### `internal/render/converter.go` (renderer, transform)

**Analog:** itself — `createEdge` (converter.go:426-482). Change ONLY the penwidth block (`:478-479`):
```go
	// Set edge penwidth to 2.0 (twice the default node border width of 1.0)
	e.SetPenWidth(2.0)
```
Executor guidance: make this `if edge.PenWidth > 0 { e.SetPenWidth(edge.PenWidth) } else { e.SetPenWidth(1.0) }` for resolved views, but KEEP 2.0 default for `--expanded` output — the discriminator is `v.AllExpanded` on the view. Note the renderer has no access to the view; the mode must arrive via the graph (e.g., a `PenWidth` defaulted per mode on each edge, or a Graph-level flag). MinLen serialization at `:474-476` (`if edge.MinLen > 0 { e.SetMinLen(edge.MinLen) }`) needs no change — gating at builder.go:454 suffices (D-13).

### `internal/view/scope_test.go` (test)

**Analog:** `internal/view/integration_test.go:468-519` (`TestIntegrationC1EdgeResolution` — asserts `ResolvedLinks[0].Peer`) and `:346-455` (`TestIntegrationC1ViewNoNestedBoundaryPollution` — the 5-node C1 fixture that must stay green).

Patterns to replicate for new tests (D-07..D-11):
```go
func TestIntegrationC1EdgeResolution(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"webUser": {
				Type: model.TypePersonExternal,
				Name: "Web User",
				// Link targets a deeply nested component
				Links: []model.Link{
					{Peer: "system.api.handler", Technology: "HTTP"},
				},
			},
			// ... nested system with api.handler
		},
	}

	v := view.GenerateC1View(m)
	require.NotNil(t, v)

	// webUser should have a resolved link to system (not system.api.handler)
	if assert.NotNil(t, v.Units["webUser"].ResolvedLinks) {
		assert.Equal(t, "system", v.Units["webUser"].ResolvedLinks[0].Peer)
	}
}
```
- Expand-a-unit fixture: set `Expanded: []string{"linuxSystem"}` on the top-level unit (self-referencing — `isExpandedInC1` scope.go:144-150 accepts both properties-level and unit-level `Expanded`).
- Assert visible-subunit resolution: `ResolvedLinks[0].Peer == "linuxSystem.sshAuth"` for a link `webUser → linuxSystem.sshAuth.sshd` when linuxSystem is expanded (D-07); source-side: `v.Units["linuxSystem.sshAuth"].ResolvedLinks` non-nil (D-09); within-cluster: pair `linuxSystem.sshAuth → linuxSystem.webAPI` present (D-10).
- D-12 REQUIRES updating `TestGenerateExpandedView_AddsExternalBoundaryNodesForLinkedUnits` (scope_test.go:784-818) — it asserts the deleted behavior (`cloudstorage` boundary node added by `addExternalBoundaryNodes`). Delete or invert this test (expanded view no longer synthesizes boundary nodes; validator rejects undefined peers upstream).
- Existing tests that MUST stay green: `TestGenerateC1View_ExternalBoundaryNodesForReferencedUnits` (:160), `TestGenerateC1View_NoDuplicateExternalBoundaryNodes` (:211), `TestIntegrationC1ViewNoNestedBoundaryPollution` (integration_test.go:346), definition-order test (:924).

### `internal/graph/builder_test.go` (test)

**Analog:** `TestBuildGraphMultipleLinks` (builder_test.go:252-287 — "Test 15: Multiple links between same units shown separately" — this WILL CHANGE to pair-collapse semantics for resolved views), `TestBuildGraphEdgeLength` (:928-1023 — the MinLen assert pattern `assert.Equal(t, 2, g.Edges[0].MinLen)`), `TestBuildGraphEdgeColor` (:742-925 — first-wins attribute assert pattern).

New-test patterns (D-01, D-04, D-05, D-13):
```go
//nolint:funlen // Test functions with model setup are naturally longer
func TestBuildGraphEdgeLength(t *testing.T) {
	t.Parallel()

	// Test 1: Edge with length > 0 has MinLen set
	t.Run("edge with length > 0 has MinLen set", func(t *testing.T) {
		t.Parallel()
		// ... model with Links: []model.Link{{Peer: "db", Length: 2}}
		v := view.GenerateC1View(m)
		g := graph.BuildGraph(v)

		require.Len(t, g.Edges, 1)
		assert.Equal(t, 2, g.Edges[0].MinLen)
	})
}
```
- Pair-collapse fixture (D-01): one source, two `Links` to the same peer with different Technology/Description → `require.Len(t, g.Edges, 1)`, edge carries the FIRST link's tech/desc (definition order = slice order).
- Penwidth fixture (D-04/D-05): two contributing links → `assert.Equal(t, 2.0, g.Edges[0].PenWidth)`; one link → `assert.Zero(t, ...)` or 1.0 per chosen encoding.
- Minlen fixture (D-13): link `Length: 2` whose target resolves (e.g., peer `system.api.handler` in C1) → `assert.Zero(t, g.Edges[0].MinLen)`; direct-pair link with Length → `assert.Equal(t, 2, ...)`.
- Expanded-mode exemption (D-02/COMPAT-02): `TestBuildExpandedGraphRealToml` (:698-739) asserts `MinLen` survives in `--expanded` mode (render.RenderDOT at :734) — MUST stay green. Also keep `TestBuildGraphMultipleLinks` semantics for expanded mode only if the pair key branches on `v.AllExpanded`.

### `cmd/c4drill/root_test.go` (test, COMPAT-02 guard)

**Analog:** itself — `testdata/expanded.toml` fixtures used at :488, :553, :574, :603. No golden-file byte comparison exists today; per RESEARCH the COMPAT-02 check is that expanded-mode DOT output is unchanged after the phase. Planner should add a baseline test rendering `testdata/expanded.toml` with `--expanded` and asserting `minlen=` + default 2.0 penwidth appear (guards against pair-only key or penwidth leaking into expanded mode).

## Shared Patterns

### Resolved-link override seam
**Source:** `internal/graph/builder.go:339-347`
**Apply to:** all buildEdges changes (D-01, D-04, D-05, D-13)
```go
outLinks := entry.Unit.Links
if entry.ResolvedLinks != nil {
	outLinks = entry.ResolvedLinks
}
```
All resolution refinements plug into this seam — `GenerateExpandedView` never populates ResolvedLinks, so `v.AllExpanded` cleanly discriminates D-02.

### Definition-order iteration with map-key fallback
**Source:** `internal/view/scope.go:101-110` (and `:346-353`, builder.go:284-292)
**Apply to:** any new resolution pass in scope.go
```go
// Determine iteration order: use UnitOrder if available, otherwise fallback to map keys
var unitOrder []string
if len(m.UnitOrder) > 0 {
	unitOrder = m.UnitOrder
} else {
	// Fallback for test models or models without explicit order
	for name := range m.Units {
		unitOrder = append(unitOrder, name)
	}
}
```
"First in definition order wins" (D-01/D-03/D-05) is deterministic ONLY because every pass iterates these slices, never map keys directly.

### markSeen dedup
**Source:** `internal/graph/builder.go:460-468`
**Apply to:** D-01 (key change only — mechanism unchanged)
```go
// markSeen marks an edge key as seen and returns true if it was not already seen.
func markSeen(seen map[string]bool, edgeKey string) bool {
	if seen[edgeKey] {
		return false
	}

	seen[edgeKey] = true

	return true
}
```

### Nil guards at public API boundaries
**Source:** `internal/view/scope.go:88-91`, `internal/graph/builder.go:15-18`
**Apply to:** any new exported function
```go
func GenerateC1View(m *parser.Model) *View {
	if m == nil {
		return nil
	}
```
Also mandated by RESEARCH Security Domain: new resolution code must nil-guard like existing passes (`//nolint:` directives always with explanation).

### Test conventions (from CONVENTIONS.md + RESEARCH)
- `require` for fatal conditions, `assert` for value checks, with field-path message args; external test packages (`package view_test`).
- `t.Parallel()` on all view/graph tests; render-touching tests avoid it ONLY where the go-graphviz WASM SVG engine is used — DOT rendering is parallel-safe (precedent: builder_test.go:699 calls `render.RenderDOT` under `t.Parallel()` with no nolint).
- Long test-model setup functions get `//nolint:funlen // Test functions with model setup are naturally longer` (precedent builder_test.go:289).
- Real-TOML regression pattern: `parser.ParseFile("../../cyp-auth-infra/...")` + `validator.Validate(m)` + `require.Empty(t, valErrors)` (builder_test.go:698-739).
- Per-task check: `go test ./internal/view/ ./internal/graph/ ./cmd/c4drill/`; gate: `mise test` + `mise lint`.

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| (none) | — | — | All 8 modification targets already exist; new test functions extend existing files with the patterns above |

Notable: there is no existing pair-collapse or penwidth test — those are genuinely new test functions, but they extend `builder_test.go`/`scope_test.go` which supply the full pattern (fixture construction, require/assert, t.Parallel).

## Metadata

**Analog search scope:** `internal/view/`, `internal/graph/`, `internal/render/`, `internal/model/`, `internal/parser/`, `internal/validator/`, `cmd/c4drill/`
**Files scanned:** 18 (scope.go, builder.go, converter.go, view.go, graph.go, link.go, unit.go, parser.go, rules.go, root.go + 7 test files)
**Pattern extraction date:** 2026-08-06
**Key risk points encoded:** Pitfall 1 (C2/C3 passes `resolveBoundaryNodeLinks`/`resolveSubunitCrossLinks` scope.go:652-859 untouched), Pitfall 2 (`v.AllExpanded` discriminator), Pitfall 3 (penwidth 1.0/2.0 semantics, converter.go:478), Pitfall 4 (minlen gate builder.go:454), Pitfall 5 (visible-subunit bookkeeping — skip condition scope.go:300 becomes "resolved source == resolved target")
