# Phase 37: Nesting Context and Plain Rendering - Pattern Map

**Mapped:** 2026-08-30
**Files analyzed:** 12 (9 code, 3 docs/goldens)
**Analogs found:** 9 / 9 code files (all modify existing files — every fix point has in-file precedent)

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/graph/builder.go` (modify: `buildCluster` recursion + plain guards) | service (graph builder) | transform | same file: `buildNestedCluster` (L212-272) | exact (in-file) |
| `internal/view/scope.go` (modify: CTX-02 ancestor-chain entries) | service (view generator) | transform | same file: visible-subunit population (L156-184), `createExternalBoundaryNode` (L221-248) | exact (in-file) |
| `internal/view/view.go` (modify: add `Plain bool`) | model (view DTO) | transform | same file: `AllExpanded` field (L41-44) | exact (in-file) |
| `internal/graph/graph.go` (modify: add `Plain bool`) | model (graph DTO) | transform | same file: `EdgeStyle` copy source (builder.go L25) | exact |
| `internal/render/converter.go` (modify: plain label branch) | service (renderer) | transform | same file: `buildHTMLLabelForType` dispatch (L16-40) + `setNodeLabel`/`setClusterLabel`/edge-label sites | exact (in-file) |
| `internal/render/labels.go` (modify: plain labels) | utility | transform | same file: `buildRecordLabel` (L13-42) — already the plain-text path | exact (in-file) |
| `cmd/c4drill/root.go` (modify: `--plain` flag + set on views) | controller (cobra) | request-response | same file: `--expanded` flag (L53, L94-95, L188-191) | exact (in-file) |
| `internal/view/scope_test.go`, `internal/graph/builder_test.go`, `cmd/c4drill/root_test.go` (modify: new tests) | test | n/a | existing tests; golden analog `canonical.Canonical` (builder_test.go L1236-1238, root_test.go L1315) | exact |
| `cmd/c4drill/testdata/multilevel{,.expanded}.dot`, `expanded/**` (re-baseline, ONE task) | test fixture (golden) | n/a | `internal/testutil/canonical` contract | exact |
| new plain-mode goldens (new) | test fixture (golden) | n/a | existing goldens + canonical | role-match |
| `README.adoc`, `skill/SKILL.md` + 2 plugin copies, `skill/examples/` (modify/new) | docs | n/a | none (docs; CI enforces `diff -r` parity) | no analog |

## Pattern Assignments

### `internal/graph/builder.go` — CTX-03: make `buildCluster` recursive (service, transform)

**Analog:** `buildNestedCluster` in the same file (L210-272) — the recursion already exists one function over; CTX-03 copies its child-dispatch into `buildCluster`.

**The recursive pattern to copy** (builder.go L244-268 — child loop of `buildNestedCluster`):
```go
	for _, childName := range childOrder {
		childUnit := entry.Unit.Subunits[childName]
		childPath := path + "." + childName

		childEntry, exists := v.Units[childPath]
		if !exists {
			// Create entry if not in view (shouldn't happen, but be defensive)
			childEntry = &view.Entry{
				Unit:        childUnit,
				FullPath:    childPath,
				HasSubunits: len(childUnit.Subunits) > 0,
				IsExternal:  view.IsExternalType(childUnit.Type),
			}
		}

		if childEntry.HasSubunits {
			// Recursively build nested cluster
			nestedCluster := buildNestedCluster(childEntry, childPath, v)
			cluster.Clusters = append(cluster.Clusters, nestedCluster)
		} else {
			// Build node for leaf subunit
			node := buildNode(childEntry)
			node.IsInCluster = true
			cluster.Nodes = append(cluster.Nodes, node)
		}
	}
```

**The fix point it applies to** (builder.go L641-655 — `buildCluster` currently flattens: computes `IsExpanded`/`HasSubunits` then ignores both):
```go
	for _, childName := range childOrder {
		childUnit := entry.Unit.Subunits[childName]
		childPath := entry.FullPath + "." + childName
		childEntry := &view.Entry{
			Unit:        childUnit,
			FullPath:    childPath,
			IsExpanded:  isUnitExpanded(entry.Unit, childName),
			HasSubunits: len(childUnit.Subunits) > 0,
			IsExternal:  view.IsExternalType(childUnit.Type),
		}

		node := buildNode(childEntry)
		node.IsInCluster = true
		cluster.Nodes = append(cluster.Nodes, node)
	}
```
Note: `buildCluster` also lacks the nested `Clusters: make([]*Cluster, 0)` field that `buildNestedCluster` sets (L223-231) — recursion requires it. Keep the D-07 guard at the caller sites (builder.go L72-79 and L97-105: `if entry.IsExpanded && len(entry.Unit.Subunits) > 0` — expanded-but-empty stays a plain node). Keep the 🔍 affordance from `buildNode` (L284-286: `if entry.HasSubunits && !entry.IsExpanded { label.Name += " 🔍" }`) on now-collapsed containers.

**Same-file precedent for the child-order loop** (builder.go L233-242 — SubunitOrder with map-key fallback; copy verbatim into any new recursion):
```go
	var childOrder []string
	if len(entry.Unit.SubunitOrder) > 0 {
		childOrder = entry.Unit.SubunitOrder
	} else {
		// Fallback to map keys for test models without explicit order
		for name := range entry.Unit.Subunits {
			childOrder = append(childOrder, name)
		}
	}
```

---

### `internal/view/scope.go` — CTX-02: deep-link ancestor-chain entries (service, transform)

**Analog A — chain-entry registration:** the visible-subunit population in `GenerateC1View` (scope.go L156-184). New chain entries (depicted ancestor → true target) MUST follow this exact triple registration, or Pitfall 2 (duplicate top-level nodes) fires:
```go
		for _, subName := range subunitOrderOf(entry.Unit) {
			subUnit := entry.Unit.Subunits[subName]
			if subUnit == nil {
				continue
			}

			fullPath := name + "." + subName
			v.Units[fullPath] = &Entry{
				Unit:        subUnit,
				FullPath:    fullPath,
				IsExpanded:  isUnitExpanded(entry.Unit, subName),
				HasSubunits: len(subUnit.Subunits) > 0,
				IsExternal:  IsExternalType(subUnit.Type),
			}
			v.UnitOrder = append(v.UnitOrder, fullPath)
			v.VisiblePaths[fullPath] = true
		}
```
The consumer that relies on this: `buildC1ViewGraph` skips `VisiblePaths` as top-level nodes (builder.go L90-94):
```go
	for _, key := range v.UnitOrder {
		if v.VisiblePaths[key] {
			continue
		}
```

**Analog B — entry construction with real unit data:** `createExternalBoundaryNode` (scope.go L221-248) shows how to wrap a found model unit vs. synthesize a placeholder; chain entries always have real units (`findUnitByPath`), so use the first branch's shape (`IsBoundary: true` only for out-of-scope siblings — chain entries are IN scope and must not set it).

**The fix point — `resolveBoundaryLink`** (scope.go L308-351). Today the resolved link (L336-348) carries only the collapsed `Peer: resolved`; CTX-02 adds chain entries between `resolved` and `link.Peer` and points the edge at the TRUE target:
```go
	resolved := resolveToTopLevel(v, link.Peer)
	if resolved == "" {
		return model.Link{}, false // Internal link (both sides under same ancestor or self-referencing)
	}

	if _, exists := v.Units[resolved]; !exists {
		v.Units[resolved] = createExternalBoundaryNode(m, resolved, path)
		v.UnitOrder = append(v.UnitOrder, resolved)
	}
	...
	resolvedLink := model.Link{
		Peer:          resolved,
		Technology:    link.Technology,
		Description:   link.Description,
		...
	}
```
Apply the same change to the C2/C3 cross-subunit resolution: `addResolvedCrossLink`/`addResolvedCrossLinkFrom` (scope.go L968-1028) which use `resolveToViewAncestor` (L862-888) the same way.

**Ordering pattern — REQUIRED for new chain entries** (Pitfall 4): `sortBoundaryNodesByModelOrder` (scope.go L595-623) — new `Units`/`UnitOrder` appends must end up in deterministic definition order:
```go
	var internal, boundary []string

	for _, path := range v.UnitOrder {
		entry := v.Units[path]
		if entry != nil && (entry.IsBoundary || entry.IsExternal) {
			boundary = append(boundary, path)
		} else {
			internal = append(internal, path)
		}
	}
	// ... sort boundary by index in m.UnitOrder ...
	v.UnitOrder = slices.Concat(internal, boundary)
```
Note: this sorts only the boundary tail. Chain entries are internal (not IsBoundary/IsExternal) — if appended out of order they stay out of order. Either insert them adjacent to their depicted ancestor or extend this function; do not leave append-order dependent on link scan order.

**Dedup-key pitfall site** (builder.go L972-977, L1018-1023): edge keys and pair multiplicity (`countPairMultiplicity` L732-744, `collectPairAggregates` L897-923) all key on `link.Peer` — once `resolveBoundaryLink` emits TRUE targets, these automatically use the new targets (they read the same resolved links), but tests must confirm no pair miscounts.

---

### `internal/view/view.go` + `internal/graph/graph.go` — Plain field threading (model)

**Analog — the `AllExpanded` precedent** (view.go L41-44): the existing flag threaded as a struct field; `Plain` copies this shape:
```go
	// AllExpanded indicates this view contains ALL units at all nesting levels
	// (--expanded mode). When true, edge deduplication keeps the technology+description
	// key and all edges render at penwidth 2.0 (COMPAT-02).
	AllExpanded bool
```
Add `Plain bool` next to `Edges string` (view.go L33) with the same doc style. Mirror in `graph.Graph` (graph.go L33-51) next to `EdgeStyle`.

**Threading pattern — entry points copy view fields into the graph** (builder.go L22-29, identical at L168-175 in `BuildExpandedGraph`):
```go
	g := &Graph{
		Title:     v.Title,
		Direction: "TB",
		EdgeStyle: v.Edges,
		Nodes:     make([]*Node, 0),
		Edges:     make([]*Edge, 0),
		Clusters:  make([]*Cluster, 0),
	}
```
`Plain: v.Plain` goes here in BOTH `BuildGraph` and `BuildExpandedGraph` (Pitfall 5: `--plain` × `--expanded` must both work).

**Guard shape precedent — copy `AllExpanded` guards** (builder.go L736-738):
```go
	// Expanded mode keeps the v1.7 dedup/penwidth behavior (D-02, COMPAT-02).
	if v.AllExpanded {
		return pairCounts
	}
```

**Plain guards apply to:**
- `applyUnitOverrides` (builder.go L320-339) — skip when plain (call sites: buildNode L302, buildCluster L619, buildNestedCluster L221, buildBoundaryCluster L127). Guard INSIDE `applyUnitOverrides` (needs the view/graph plain flag passed in, or guard at the 4 call sites).
- `createEdge` (builder.go L1097-1142) — treat link `Color`/`Style`/`Length`/`Rank` as unset when plain. The fields to neutralize:
```go
	if link.Color != "" {
		edge.Color = link.Color
	} else if colour := kindColour(link.Kind); colour != "" {   // kind colours STAY (semantic)
	...
	edge.MinLen = link.Length                                    // plain: leave 0
	edge.NoConstraint = link.Rank == model.RankEqual             // plain: leave false
	edge.RankReverse = link.Rank == model.RankReverse            // plain: leave false
```
- `EdgeStyle: ""` when plain (cleared at graph construction → `configureGraphSettings` switch at converter.go L236-244 falls through → default splines).

---

### `cmd/c4drill/root.go` — `--plain` flag (controller, request-response)

**Analog — `--expanded` plumbing, copy exactly** (root.go L49-56, L94-95):
```go
 //nolint:gochecknoglobals // Cobra flags require package-level variables for PersistentFlags registration
 var (
 	format     string
 	outputDir  string
 	expanded   bool
 	labelRatio float64
 	version    = "dev"
 )
 ...
 	cmd.PersistentFlags().BoolVar(&expanded, "expanded", false,
 		"Generate all-expanded diagram showing all units")
```
Add `plain bool` to the var block and `cmd.PersistentFlags().BoolVar(&plain, "plain", false, "...")` directly after the `--expanded` registration.

**Dispatch + per-view assignment** (root.go L188-191 dispatch; RESEARCH design sets `v.Plain` in both processors — root.go L295-344 `processView`, L346-383 `processExpandedView`):
```go
	// Handle --expanded mode (skip normal C1/C2/C3 generation)
	if expanded {
		return processExpandedView(m, basename, writer)
	}
```
In `processView` (after `view.GenerateC1View/C2View/C3View` at L299-306, before `graph.BuildGraphWithPath` at L313) and in `processExpandedView` (after `view.GenerateExpandedView` at L350), add `v.Plain = plain`. No generator signature changes.

**Counter-pattern (do NOT follow):** `render.LabelRatio` package global (wrap.go L34-38 `var LabelRatio = defaultLabelRatio`, set at root.go L126 `render.LabelRatio = getLabelRatio()`). RESEARCH explicitly rejected a package global for `--plain` in favor of explicit `View.Plain`/`Graph.Plain` threading.

---

### `internal/render/converter.go` + `internal/render/labels.go` — plain branch (service/utility, transform)

**Label dispatch fix point** (converter.go L16-40 `buildHTMLLabelForType`; called from `setNodeLabel` L382-393 and `setClusterLabel` L475-486). The plain branch selects the record path instead of the HTML-type dispatch:
```go
func buildHTMLLabelForType(label *graph.Label, t model.UnitType) string {
	if label == nil {
		return ""
	}

	switch {
	case graph.IsPersonType(t):
		return buildPersonHTMLLabel(label)
	...
```

**The plain-text label already exists** (labels.go L13-42 `buildRecordLabel`) — plain mode routes node/cluster labels here (content name/technology/description preserved, no HTML table):
```go
func buildRecordLabel(label *graph.Label) string {
	if label == nil {
		return ""
	}

	var parts []string

	// Build name with optional icon
	var nameBuilder strings.Builder
	if label.Icon != "" {
		nameBuilder.WriteString(label.Icon)
		nameBuilder.WriteString(" ")
	}

	nameBuilder.WriteString(label.Name)
	parts = append(parts, nameBuilder.String())

	// Add technology if present
	if label.Technology != "" {
		parts = append(parts, label.Technology)
	}
	...
	return "{" + strings.Join(parts, "|") + "}"
}
```
Note: `buildRecordLabel` output goes through `SetLabelHTML` today via `setNodeLabel`'s fallback (converter.go L37-38 default case). Under plain, decide whether record labels are set via `SetLabel` (true plain text) or keep `SetLabelHTML` — check emitted DOT against goldens.

**Edge-label plain branch** (converter.go L567-580 — HTML rectangle emission):
```go
		if htmlLabel := buildEdgeLabel(edge.Label); htmlLabel != "" {
			e.SetLabelHTML(htmlLabel)
		} else {
			e.SetLabel("")
		}
```
Plain: emit `e.SetLabel(...)` with plain text (technology/description content) — never `SetLabelHTML`. `labelPosition`: `createEdge` already defaults empty to `"middle"` (builder.go L1116-1118); under plain ignore `link.LabelPosition` at that site.

**Rank suppression — BOTH (really three) emission sites** (Pitfall 3). Primary suppression is at the builder (`createEdge` leaves `NoConstraint`/`RankReverse` false under plain), which neutralizes all render sites automatically. The sites tests must assert on:
1. Endpoint swap (converter.go L553-557):
```go
	tail, head := source, target
	if edge.RankReverse {
		tail, head = target, source
	}
```
2. dir inversion (converter.go L604-621 `setEdgeDir` `if edge.RankReverse` branch).
3. constraint suppression (converter.go L652-654):
```go
	if edge.NoConstraint {
		e.SetConstraint(false)
	}
```

**Other attribute consumption unchanged** — `applyEdgeAttributes` (converter.go L639-664): minlen (L647-649) is inert once builder leaves `MinLen 0`; edge Color (L641-644) is inert once builder clears `link.Color`; kind colours (`kindColour`, builder.go L1077-1088) STAY. Legend: `buildLegend`/`legendKindEntries` (builder.go L357-371, L545-570) reads FINAL edge colours, so suppressing custom colours automatically drops kind rows whose colour no longer appears — do not special-case; assert the emergent behavior in tests:
```go
func legendKindEntries(edges []*Edge) []LegendEntry {
	used := make(map[string]bool, len(edges))
	for _, edge := range edges {
		if edge != nil {
			used[edge.Color] = true
		}
	}
```
Queue pipes (pipe.go) and legend/custom lines stay as-is under plain.

---

### Tests — view/graph unit + cmd E2E + goldens (test)

**Golden comparison pattern** (builder_test.go L1236-1238; also root_test.go L1315, L1453):
```go
	// order-insensitive — canonical.Canonical normalizes both sides to a sorted,
	// geometry-stripped semantic form
	require.Equal(t, canonical.Canonical(t, string(expected)), canonical.Canonical(t, string(dotData)),
```
Import: `github.com/Djarvur/c4drill/internal/testutil/canonical`. The comparator (canonical.go L34-43) strips geometry (`bb`, `pos`, `lp`, `lheight`, `lwidth`, `height`, `width` — L238-245) and sorts statements/attrs — use for ALL new goldens and the ONE dedicated re-baseline task.

**Fixture for CTX tests:** `cmd/c4drill/testdata/multilevel.toml` (4-level nesting, cross-level links — named by RESEARCH as the fixture for scope_test.go ancestor-chain assertions and builder_test.go nested-cluster assertions).

**New plain goldens:** styled fixture + `--plain` run → new golden files (do NOT re-baseline existing goldens for plain; flag default false keeps no-plain output untouched).

**Re-baseline pitfall (Pitfall 1):** CTX-03 recursion changes output for any model with expanded clusters containing subunit-containers. Collect ALL golden re-baselines (`multilevel.dot`, `multilevel.expanded.dot`, `expanded/**`, in-package COMPAT-02/REF-05 goldens) into ONE dedicated task; flat models (`testdata/*.toml`: links, nested without deep scenarios) must stay clean.

---

## Shared Patterns

### 1. Flag-on-struct threading (View → Graph → guards)
**Source:** `View.AllExpanded` (view.go L41-44) + `BuildGraph`/`BuildExpandedGraph` field copy (builder.go L22-29, L168-175) + `AllExpanded` guards (builder.go L736-738, L900-902).
**Apply to:** `Plain` end to end — `View.Plain` (set by root.go) → `Graph.Plain` (copied in both entry points) → guards in `applyUnitOverrides`/`createEdge`/label dispatch. Never a package global (rejected LabelRatio pattern).

### 2. Definition-order determinism
**Source:** `sortBoundaryNodesByModelOrder` (scope.go L595-623); SubunitOrder-with-fallback loops (builder.go L233-242, scope.go L107-118).
**Apply to:** every new `Units`/`UnitOrder` append in CTX-02 chain insertion — nondeterministic DOT ordering is the #1 golden-flake risk.

### 3. VisiblePaths triple registration
**Source:** scope.go L174-182 (Units + UnitOrder + VisiblePaths) consumed by builder.go L90-94.
**Apply to:** all CTX-02 chain entries in C1. For C2/C3 chain entries inside the boundary cluster, entries sit in `v.Units` under the expanded unit — confirm `buildBoundaryViewGraph` (builder.go L50-83) dispatches them into nested clusters, not top level (`IsBoundary` must stay false for in-scope chain entries).

### 4. D-07 expanded-but-empty guard
**Source:** builder.go L72-79 and L97-105 (`entry.IsExpanded && len(entry.Unit.Subunits) > 0`).
**Apply to:** recursion must preserve it — expanded-but-empty stays a plain node.

### 5. Canonical goldens
**Source:** `internal/testutil/canonical.Canonical` (canonical.go L34-43); usage builder_test.go L1236-1238, root_test.go L1315.
**Apply to:** all graph/DOT assertions; one consolidated re-baseline task; new plain goldens additive only.

### 6. Cobra flag convention
**Source:** root.go L49-56 (var block with `//nolint:gochecknoglobals`), L94-95 (BoolVar registration).
**Apply to:** `--plain` flag registration and `v.Plain = plain` assignment in `processView` (L295-344) + `processExpandedView` (L346-383).

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `README.adoc` (--plain + nesting sections) | docs | n/a | No prior doc-feature analog in codebase; follow RESEARCH.md DOC-01 and existing README section structure |
| `skill/SKILL.md` + `plugins/c4drill/skills/c4drill-toml` + `plugins/c4drill/opencode/skills/c4drill-toml` | docs | n/a | Content is prose; the binding constraint is mechanical: CI `diff -r` parity (.github/workflows/validate-skill-examples.yml:35-40) — re-sync all 3 copies in the same task |
| `skill/examples/` new fixtures | test fixture (CI-validated examples) | n/a | New content; must render cleanly through the full pipeline (CI validates); pattern = existing examples' .toml + .c4d twin convention |

## Metadata

**Analog search scope:** `internal/view/`, `internal/graph/`, `internal/render/`, `cmd/c4drill/`, `internal/testutil/canonical/`
**Files read in full:** builder.go (1284), scope.go (1028), view.go (98), graph.go (171), converter.go (680), labels.go (259), root.go (405), canonical.go (308)
**Files scanned (grep/ls):** wrap.go, wrap_internal_test.go (LabelRatio counter-pattern), test files inventory, testdata inventory, skill copy inventory
**Pattern extraction date:** 2026-08-30
