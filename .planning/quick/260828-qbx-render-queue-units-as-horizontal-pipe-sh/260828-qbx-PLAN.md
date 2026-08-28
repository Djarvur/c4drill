---
phase: 260828-qbx-render-queue-units-as-horizontal-pipe-sh
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/graph/shapes.go
  - internal/graph/shapes_test.go
  - internal/graph/integration_test.go
  - internal/render/converter.go
  - internal/render/render.go
  - internal/render/pipe.go
  - internal/render/pipe_test.go
  - internal/render/render_test.go
  - cmd/c4drill/testdata/multilevel.expanded.dot
  - skill/examples/*.svg
autonomous: true
requirements: [QUICK-PIPE-01, QUICK-ICON-01, QUICK-DOCS-01]
must_haves:
  truths:
    - "Queue-type units (queue, queueExternal, containerQueue, componentQueue) render as a horizontal pipe (SVG <path>) in SVG and HTML output"
    - "Non-queue nodes are visually unchanged: rounded <path> or <polygon> outlines still emitted"
    - "DOT output keeps plain boxes and no longer carries the U+255F/U+2562 queue bars icon"
    - "Edge anchors stay valid: queue nodes remain shape=box in cgraph; the pipe is inscribed in the same bbox"
    - "All tracked example SVGs re-rendered in sync; no untracked artifacts under plugins/ or skill/examples/ (CI diff -r parity)"
  artifacts:
    - path: "internal/render/pipe.go"
      provides: "Queue-node SVG post-processor: queue-ID collection (nested clusters) + polygon-to-pipe-path replacement"
    - path: "internal/render/converter.go"
      provides: "Queue nodes emit style WITHOUT 'rounded' (plain rect polygon) and wider min width"
      contains: "IsQueueType"
    - path: "internal/graph/shapes.go"
      provides: "IconForType returns empty string for queue types; doc comment updated"
    - path: "cmd/c4drill/testdata/multilevel.expanded.dot"
      provides: "Re-baselined golden DOT without queue bars icon / rounded queue style"
  key_links:
    - from: "internal/render/render.go render()"
      to: "internal/render/pipe.go"
      via: "post-process SVG bytes when format == graphviz.SVG"
      pattern: "applyPipeRendering|pipeRequeue|replaceQueuePolygons"
    - from: "internal/render/converter.go createNode"
      to: "graph.IsQueueType"
      via: "style branch dropping 'rounded' for queue nodes"
      pattern: "IsQueueType"
    - from: "internal/render/RenderHTML"
      to: "pipe post-processor"
      via: "RenderHTML calls render(g, graphviz.SVG) so HTML inherits the pipe"
      pattern: "render\\(g, graphviz\\.SVG\\)"
---

<objective>
Render queue-type units as horizontal pipe shapes in SVG/HTML output by post-processing the GraphViz SVG bytes, replacing the rounded-box + text-bar-icon proxy.

Purpose: Queues currently render as rounded boxes with a ╟/╢ text-bar cue, which does not read as a queue. GraphViz 15.1.1 (pinned) cannot rotate shape=cylinder and lacks shape=tape, so the pipe is drawn by replacing the queue node's `<polygon>` with a hand-computed `<path>` in the SVG bytes — the same post-processing precedent as wrapSVGInHTML.

Output: internal/render/pipe.go post-processor + tests, plain-rect queue styling in converter.go, empty queue icon in shapes.go, re-baselined golden DOT, re-rendered example SVGs, synced skill plugins.
</objective>

<execution_context>
@/Users/nil/DiskD/W/Djarvur/c4drill/.zcode/get-shit-done/workflows/execute-plan.md
@/Users/nil/DiskD/W/Djarvur/c4drill/.zcode/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md
@internal/render/render.go
@internal/render/converter.go
@internal/graph/shapes.go
@internal/graph/builder_test.go

<interfaces>
<!-- Key contracts the executor needs. Extracted from codebase. -->

From internal/render/render.go:
```go
// render is the internal render function that handles all formats.
// This is the post-processing hook point: RenderSVG and RenderHTML both
// funnel through render(g, graphviz.SVG); RenderDOT uses graphviz.XDOT.
func render(g *graph.Graph, format graphviz.Format) ([]byte, error)
func wrapSVGInHTML(svgBytes []byte) []byte // precedent for byte post-processing
```

From internal/render/converter.go:
```go
func buildStyleString(style *styleInfo) []string // line 61: ALWAYS starts with "rounded" — queue nodes must bypass this
func createNode(cg *cgraph.Graph, node *graph.Node) (*cgraph.Node, error) // line 280: sets shape (IsDbType -> CylinderShape, else BoxShape), label, style
func applyNodeStyle(cn *cgraph.Node, style *graph.NodeStyle) // line 333: joins buildStyleString output into style attr
```

From internal/graph/shapes.go:
```go
func IsQueueType(t model.UnitType) bool // TypeQueue, TypeQueueExternal, TypeContainerQueue, TypeComponentQueue
func IconForType(t model.UnitType) string // queue case returns "\u255F\n\u2562" — change to ""
```

From internal/graph/graph.go:
```go
type Graph struct { Nodes []*Node; Edges []*Edge; Clusters []*Cluster; Legend *Legend; ... }
type Node struct { ID string; Label *Label; Shape Shape; Type model.UnitType; Style *NodeStyle; IsInCluster bool; ... }
type Cluster struct { ID string; Nodes []*Node; Clusters []*Cluster; ... } // nested clusters exist (--expanded all-levels)
```

GraphViz SVG emission (verified via dot-CLI experiments, locked design):
- Plain (non-rounded) box node → `<g id="nodeN"><title>NODE_ID</title><polygon points="x,y x,y ..." .../></g>`
- Rounded node → `<path d="..."/>` instead of polygon
- Dashed style → stroke-dasharray attribute on the shape element (mirror whatever the pinned build emits; inspect one dashed fixture before finalizing)
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Queue nodes get plain-rect style + wider min width; queue text-bar icon removed; DOT golden re-baselined</name>
  <files>internal/graph/shapes.go, internal/graph/shapes_test.go, internal/graph/integration_test.go, internal/render/converter.go, cmd/c4drill/testdata/multilevel.expanded.dot</files>
  <behavior>
    - Test: graph.IconForType(model.TypeQueue) == "" (same for TypeQueueExternal, TypeContainerQueue, TypeComponentQueue); non-queue icons unchanged (person emoji, db U+26C1)
    - Test: Label.Icon for queue nodes in built graphs is "" (internal/graph/integration_test.go:447-448 currently asserts "\u255F\n\u2562" — flip to "")
    - Test: DOT output for a queue node's style attribute does NOT contain "rounded"; a dashed queue node's style DOES contain "dashed"; non-queue nodes still emit "rounded"
  </behavior>
  <action>
1. In internal/graph/shapes.go IconForType: return "" for the four queue types (per locked design — the pipe replaces the cue). Update the function doc comment: the "Queue: U+255F/U+2562 (box drawing characters for bars)" line is stale — replace with a note that queues are drawn as a horizontal pipe by the SVG post-processor (internal/render/pipe.go), so no icon is emitted.
2. In internal/render/converter.go: queue nodes must NOT get the "rounded" style so GraphViz emits a parseable `<polygon>` instead of a rounded `<path>`. Minimum viable change respecting current structure: make the "rounded" prefix conditional — e.g. buildStyleString gains a rounded bool parameter (or a variant helper), and applyNodeStyle passes graph.IsQueueType-derived flag down from createNode; keep "filled" (only when FillColor set) and "dashed"/"dotted" behavior exactly as today. Do not change styling for non-queue nodes.
3. In createNode: for queue nodes (graph.IsQueueType(node.Type)) set a wider minimum width so pipes read as pipes: `cn.SafeSet("width", "1.8", "")` (graphviz width = minimum width in inches; graphviz may still grow the node for long labels).
4. Keep shape=box for queue nodes in createNode (edge anchors are computed on the box bbox; the pipe is inscribed in the same bbox — do NOT touch SetShape logic for queues).
5. Update the icon assertions in internal/graph/shapes_test.go (lines 44-47) and internal/graph/integration_test.go (lines 447-448) to expect "".
6. Re-baseline the DOT golden via the documented command (COMPAT-02): `go run ./cmd/c4drill cmd/c4drill/testdata/multilevel.toml --format dot --expanded --output cmd/c4drill/testdata` — the diff on cmd/c4drill/testdata/multilevel.expanded.dot must show ONLY queue-related deltas: icon lines gone from queue labels, style without "rounded" on queue nodes, width="1.8" on queue nodes.
  </action>
  <verify>
    <automated>go test ./internal/graph/... ./internal/render/... 2>&1 | tail -5</automated>
  </verify>
  <done>IconForType returns "" for all four queue types; queue DOT style has no "rounded" but keeps dashed when BorderStyle is dashed; queue nodes carry width=1.8; non-queue styling byte-identical to before; golden re-baselined with queue-only deltas; package tests green</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Pipe post-processor (pipe.go) replacing queue polygons with horizontal-pipe paths, wired into render()</name>
  <files>internal/render/pipe.go, internal/render/pipe_test.go, internal/render/render.go</files>
  <behavior>
    - Test: collectQueueNodeIDs returns all queue node IDs from a graph with top-level queue nodes AND queues nested inside clusters AND clusters nested in clusters; returns nothing for non-queue graphs
    - Test: bbox-to-path helper given polygon points "10,20 90,20 90,60 10,60" produces a path anchored at bbox (x0=10,y0=20)-(x1=90,y1=60) with cy=40, ry=20, rx≈7 (0.35*ry); path starts/ends with arc commands (A/a) for the left bulge and right cap
    - Test: polygon replacement copies fill, stroke, stroke-width (and stroke-dasharray when present) from the replaced polygon onto the new path; non-queue node groups are untouched (still contain polygon/path as emitted)
    - Test: input with zero queue IDs returns bytes unchanged
  </behavior>
  <action>
1. Create internal/render/pipe.go with (suggested, names flexible): collectQueueNodeIDs(g *graph.Graph) []string — walk g.Nodes then g.Clusters recursively (Cluster.Nodes, Cluster.Clusters) filtering graph.IsQueueType(node.Type); and a replace func queuePolygonsToPipes(svg []byte, queueIDs []string) []byte.
2. Geometry (locked design): for each queue node locate its group via `<title>ID</title>` (use plain string search or regexp.QuoteMeta — node IDs come from user TOML and may contain regex metacharacters), find the `<polygon points="..."/>` inside that group, parse points into a bbox (x0,y0)-(x1,y1) from min/max of the coordinates, and replace the polygon with a `<path d="..."/>` drawing a horizontal pipe inscribed in the bbox: cy=(y0+y1)/2, ry=(y1-y0)/2, rx=ry*0.35; left end = half-ellipse arc bulging left, straight top and bottom edges, right end = full ellipse cap. Iterate the path visually with `dot -Tpng` (render a queue fixture via the CLI) until it reads as a pipe next to db's cylinder.
3. Attribute copying: take fill, stroke, stroke-width from the replaced polygon element and re-emit on the path; when the node style is dashed the polygon carries stroke-dasharray — copy it verbatim (mirror whatever GraphViz 15.1.1 emits; inspect a dashed fixture's SVG before finalizing). Preserve the `<title>` and any `<a>`/text siblings in the group untouched.
4. Wire into render() in internal/render/render.go: after gv.Render succeeds, apply the post-processing ONLY when format == graphviz.SVG (RenderDOT/XDOT keeps plain boxes per the locked constraint; RenderHTML inherits automatically because it calls render(g, graphviz.SVG)). Skip the pass entirely when collectQueueNodeIDs finds none (zero-alloc fast path for queue-less diagrams). Keep the function under funlen ≤ 60 — split helpers rather than adding //nolint:funlen.
5. House style: gochecknoglobals is tolerated with //nolint + justification (see .golangci.yml and render.go's wasmMutex precedent) if you need package-level compiled regexps.
  </action>
  <verify>
    <automated>go test ./internal/render/... -run 'Pipe|Queue' -v 2>&1 | tail -15</automated>
  </verify>
  <done>pipe.go converts each queue node's polygon to an inscribed horizontal-pipe path preserving fill/stroke/dash attributes; queue IDs collected through nested clusters; XDOT/DOT output untouched; zero-queue SVGs pass through unchanged; unit tests green</done>
</task>

<task type="auto">
  <name>Task 3: Render integration test, example SVG re-render + skill sync, docs truthfulness, full gate</name>
  <files>internal/render/render_test.go, skill/examples/*.svg, skill/SKILL.md, README.adoc, plugins/c4drill/skills/c4drill-toml/, plugins/c4drill/opencode/skills/c4drill-toml/</files>
  <action>
1. Integration test (internal/render/render_test.go or a new internal render test file, following existing render test patterns): build a small *graph.Graph containing one queue node, one db node, one plain system node; RenderSVG; assert the queue node's `<title>ID</title>` group contains a `<path` element with the pipe geometry (arc commands present, no `<polygon`), and the non-queue nodes still render polygon or rounded path elements. Also assert RenderHTML output contains the same pipe path (HTML inherits via wrapSVGInHTML).
2. Docs truthfulness sweep: skill/SKILL.md and README.adoc were grepped during planning — neither mentions the queue bars icon (only type-list lines like "* `queue` - Message queue" which stay true). Re-grep both for icon/bars/255F/╟ references and update ONLY if any stale icon claim surfaced. shapes.go doc comment was already fixed in Task 1.
3. Re-render tracked example SVGs (build the binary first: `go build -o c4drill ./cmd/c4drill`):
   `for f in skill/examples/*.toml skill/examples/*.c4d examples/*/*.toml; do ./c4drill "$f"; done`
   plus expanded variants: `./c4drill examples/cloud-system/cloud-system.toml --expanded` and `./c4drill skill/examples/06-templates.toml --expanded`.
4. CI parity hygiene (CI runs diff -r on a pristine checkout — no untracked artifacts may remain): `git clean -f plugins/` and `git clean -x -f skill/examples`, then sync the skill:
   `rsync -a --delete skill/ plugins/c4drill/skills/c4drill-toml/ && rsync -a --delete skill/ plugins/c4drill/opencode/skills/c4drill-toml/`
5. Full gate: `go test ./...` green (golden re-baseline from Task 1 is the only expected churn) and `golangci-lint run ./...` reports 0 issues (funlen ≤ 60; gochecknoglobals only with //nolint + justification). Visually sanity-check one rendered example (e.g. open skill/examples SVG containing a queue, or `dot -Tpng` a queue fixture) so the pipe reads as a pipe; also confirm examples/overflow-test still lays out sanely (cramped fixture — wider queue min width must not explode the layout).
  </action>
  <verify>
    <automated>go test ./... 2>&1 | tail -8 && golangci-lint run ./... 2>&1 | tail -3 && git status --porcelain plugins/ skill/examples/ | grep -c '^??' || true</automated>
  </verify>
  <done>Integration test proves queue→pipe path and non-queue preservation in SVG and HTML; tracked example SVGs re-rendered and committed; plugins synced and free of untracked files; go test ./... green; golangci-lint 0 issues; pipe visually reads as a horizontal cylinder</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| user TOML → node ID → SVG title search | Queue node IDs originate from user-authored TOML and are matched against GraphViz-emitted `<title>` elements |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-QBX-01 | Tampering | pipe.go title/group matching | mitigate | Match `<title>ID</title>` via plain string search or regexp.QuoteMeta(ID) — never interpolate raw user-controlled IDs into a regexp pattern |
| T-QBX-02 | DoS | pipe.go bbox parsing | mitigate | Polygon parsing is bounded by GraphViz-generated input (trusted generator, finite points per node); no unbounded loops — single pass over SVG bytes |
| T-QBX-03 | Tampering | npm/pip/cargo installs | accept | No package-manager installs in this task set; stdlib + existing deps only |
</threat_model>

<verification>
- `go test ./...` green (only queue-golden churn, re-baselined in Task 1)
- `golangci-lint run ./...` = 0 issues
- Queue node in rendered SVG: `<path>` pipe inscribed in former polygon bbox; non-queue nodes unchanged
- DOT golden diff contains only queue-related deltas (icon removal, style, width)
- `git status --porcelain plugins/ skill/examples/` shows no untracked files after clean+rsync
- Golden DOT re-baseline command: `go run ./cmd/c4drill cmd/c4drill/testdata/multilevel.toml --format dot --expanded --output cmd/c4drill/testdata`
</verification>

<success_criteria>
Queue-type units render as horizontal pipes (SVG path: left half-ellipse bulge, straight top/bottom, right full ellipse cap) in SVG and HTML output; the ╟/╢ text-bar icon is gone from labels and DOT goldens; queue nodes keep shape=box anchoring and plain (non-rounded) rect SVG polygons that the post-processor replaces; non-queue rendering is byte-identical; all tracked example SVGs and skill plugin copies are in sync with no untracked artifacts; full test suite and lint pass.
</success_criteria>

<output>
Create `.planning/quick/260828-qbx-render-queue-units-as-horizontal-pipe-sh/260828-qbx-SUMMARY.md` when done
</output>
