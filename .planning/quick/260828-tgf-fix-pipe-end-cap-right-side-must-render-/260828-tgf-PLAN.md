---
phase: 260828-tgf-fix-pipe-end-cap-right-side-must-render-
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - internal/render/pipe.go
  - internal/render/pipe_internal_test.go
  - skill/examples/10-edge-kinds.svg
  - skill/examples/10-edge-kinds/app.svg
  - skill/examples/10-edge-kinds/ext.svg
  - examples/cloud-system/cloud-system.expanded.svg
  - examples/cloud-system/cloud-system/amazon/sqs.svg
  - plugins/c4drill/skills/c4drill-toml/
  - plugins/c4drill/opencode/skills/c4drill-toml/
autonomous: true
requirements: [QUICK-PIPE-CAP-01]
must_haves:
  truths:
    - "A queue node's right end renders as a FULL ELLIPSE (outer silhouette arc + inner face arc), not a capsule side"
    - "The emitted d attribute never contains an arc command whose start and end points coincide (SVG omits such arcs — the root cause)"
    - "The pipe still fills the original polygon bbox exactly (left cap reaches x0, outer arc reaches x1) so edge anchors stay valid"
    - "db cylinder and all non-queue rendering untouched; DOT goldens unaffected (multilevel.expanded.dot is queue-free, zero diff)"
    - "All tracked example SVGs re-rendered; plugin skill copies synced; no untracked artifacts under plugins/ or skill/examples/ (CI diff -r parity)"
  artifacts:
    - path: "internal/render/pipe.go"
      provides: "Two-subpath pipe d: closed body outline (top edge, right outer arc, bottom edge, left cap arc) + cap-face subpath (inner half-ellipse), both in one d attribute"
      contains: "0 0,0"
    - path: "internal/render/pipe_internal_test.go"
      provides: "Updated path assertions + regression test: no coincident-endpoint arc ever emitted; cap face present as separate subpath with sweep 0"
    - path: "skill/examples/10-edge-kinds.svg"
      provides: "Re-rendered example showing the full-ellipse right cap"
  key_links:
    - from: "internal/render/pipe.go pipePathFromBBox"
      to: "internal/render/pipe.go pipePathFmt"
      via: "fmt.Sprintf with the new arg order (body outline first, then cap face subpath)"
      pattern: "fmt\\.Sprintf\\(pipePathFmt"
    - from: "internal/render/pipe.go replaceQueuePolygon"
      to: "single <path> element"
      via: "both subpaths share ONE d attribute so copied paint attrs (fill/stroke/dasharray) apply to face too — no structural change to replacement"
      pattern: "<path d=\""
---

<objective>
Fix the queue pipe right end cap: render a full ellipse (outer silhouette arc + inner face arc) instead of the current capsule (two symmetric side arcs).

Purpose: The third arc in pipePathFmt is emitted with coincident start/end points ("A rx,ry 0 1,1 bodyR,y0" immediately after arriving at bodyR,y0). Per the SVG spec an arc segment whose endpoints coincide is omitted, so the intended "full ellipse cap" draws nothing — leaving a capsule. Root cause is already diagnosed and locked; this plan implements the locked two-subpath fix.

Output: Corrected pipePathFmt/pipePathFromBBox in internal/render/pipe.go, updated + new regression tests in internal/render/pipe_internal_test.go, re-rendered example SVGs, synced skill plugins, all gates green.
</objective>

<execution_context>
@/Users/nil/DiskD/W/Djarvur/c4drill/.zcode/get-shit-done/workflows/execute-plan.md
@/Users/nil/DiskD/W/Djarvur/c4drill/.zcode/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/STATE.md
@internal/render/pipe.go
@internal/render/pipe_internal_test.go

<interfaces>
<!-- Key contracts the executor needs. Extracted from codebase. -->

From internal/render/pipe.go:
```go
const pipeEndRatio = 0.35 // rx = 0.35*ry — unchanged
func pipePathFromPoints(points string) (string, bool)            // unchanged signature
func pipePathFromBBox(x0, y0, x1, y1 float64) string             // same signature; new d body
func replaceQueuePolygon(s, id string) string                    // unchanged — swaps polygon for <path d="..." + copied attrs/>
func copiedPipeAttrs(element string) string                      // unchanged — fill/stroke/stroke-width/stroke-dasharray
```

Locked new geometry (bbox (x0,y0)-(x1,y1), cy=(y0+y1)/2, ry=(y1-y0)/2, rx=0.35*ry, bodyL=x0+rx, bodyR=x1-rx):
- Subpath 1 — body outline, single closed subpath, clockwise in y-down SVG coords:
  `M bodyL,y0  L bodyR,y0  A rx,ry 0 0,1 bodyR,y1  L bodyL,y1  A rx,ry 0 0,1 bodyL,y0  Z`
  top edge → right OUTER arc (sweep 1, passes through (x1,cy)) → bottom edge → left cap arc (sweep 1, passes through (x0,cy)).
- Subpath 2 — right cap face, appended to the SAME d attribute:
  `M bodyR,y0  A rx,ry 0 0,0 bodyR,y1`
  inner half of the cap ellipse (sweep 0, passes through (bodyR-rx,cy)). Outer arc + inner arc together form the visible full ellipse.
- Invariant to note in a comment: bodyR-2rx >= x0 for queue nodes (min width 1.8in), so the inner face arc never crosses the left cap. (Fixture bbox width 129.6pt vs 3rx≈21.8pt — ample margin.)

Expected exact d for test bbox (10,20)-(90,60) (cy=40, ry=20, rx=7, bodyL=17, bodyR=83):
`M17.00,20.00 L83.00,20.00 A7.00,20.00 0 0,1 83.00,60.00 L17.00,60.00 A7.00,20.00 0 0,1 17.00,20.00 Z M83.00,20.00 A7.00,20.00 0 0,0 83.00,60.00`
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Two-subpath pipe path (body outline + right cap face) with coincident-arc regression test</name>
  <files>internal/render/pipe.go, internal/render/pipe_internal_test.go</files>
  <behavior>
    - Test: pipePathFromPoints("10,20 90,20 90,60 10,60") returns EXACTLY "M17.00,20.00 L83.00,20.00 A7.00,20.00 0 0,1 83.00,60.00 L17.00,60.00 A7.00,20.00 0 0,1 17.00,20.00 Z M83.00,20.00 A7.00,20.00 0 0,0 83.00,60.00" (deterministic geometry — full-string equality is the strongest guard)
    - Regression test (the bug): for the generated d, NO arc command has identical start and end points — parse each A command's endpoint and compare against the point the pen was at before the arc; an arc with coincident endpoints is omitted by SVG renderers and must never be emitted
    - Regression test: the cap face exists as a SECOND subpath — an "M" command after the "Z" — whose arc runs from (bodyR,y0) to (bodyR,y1) with sweep flag 0 (0 0,0)
    - Test: left cap arc and right outer arc both bulge OUTWARD: right outer arc endpoint (83.00,60.00) with sweep 1 from (83.00,20.00); left cap arc closes at (17.00,20.00) with sweep 1 from (17.00,60.00)
    - Test: TestReplaceQueuePolygons fixture (bbox (0,-57.6)-(129.6,-16) → bodyL=7.28, bodyR=122.32, rx=7.28, ry=20.80) produces d starting "M7.28,-57.60 L122.32,-57.60 A7.28,20.80 0 0,1" and containing the face subpath "M122.32,-57.60 A7.28,20.80 0 0,0 122.32,-16.00"
    - Existing tests TestPipePathFromPoints_Degenerate, TestReplaceQueuePolygons_NoQueueIDs, TestReplaceQueuePolygons_RegexMetacharID, TestApplyPipeRendering_Wiring stay green unchanged
  </behavior>
  <action>
1. RED first: update internal/render/pipe_internal_test.go to the new expectations above (they fail against current code — the current d ends with the degenerate "A7.00,20.00 0 1,1 83.00,20.00"). Run `go test ./internal/render/ -run Pipe` and confirm failure.
2. GREEN: rewrite pipePathFmt in internal/render/pipe.go (lines 36-47) to the locked two-subpath format from <interfaces>. Keep it a single fmt.Sprintf const (preferred — same shape as today, lint-clean) OR switch pipePathFromBBox to explicit strings.Builder construction; your call, but: funlen <= 60, no new package-level vars without //nolint + justification (gochecknoglobals precedent at pipe.go:33). Update pipePathFromBBox's Sprintf argument list to the new order and rewrite the stale doc comments on both pipePathFmt (describes "full-ellipse arc (right cap face, drawn in place)" — no longer true) and pipePathFromBBox (describes the old three-arc capsule) to describe the two-subpath structure: closed body outline + cap face subpath forming the visible full ellipse. Include the invariant comment: bodyR-2rx >= x0 holds for queue nodes (min width 1.8in), so the inner face arc never crosses the left cap.
3. Add the regression tests from <behavior>: a coincident-endpoint arc scanner (parse the d string's A commands, track the current pen point through M/L/A commands, assert every arc endpoint differs from the pen point) and a cap-face-subpath assertion (second M after Z, sweep 0, endpoints (bodyR,y0)→(bodyR,y1)). Name it so intent is obvious, e.g. TestPipePathFromPoints_NoCoincidentArc, TestPipePathFromPoints_CapFaceSubpath.
4. Update TestPipePathFromPoints's stale assertions: prefix becomes "M17.00,20.00 L83.00,20.00" (moveto + top edge, not an arc); bottom edge assertion becomes "L17.00,60.00" (bottom edge now runs right→left); the "A7.00,20.00" count stays 3 (right outer, left cap, cap face); the hasSuffix assertion is replaced by full-string equality.
5. Empirical sweep-flag verification (mandated by the locked design): do NOT trust the flags on paper — rebuild (`go build -o c4drill ./cmd/c4drill`), render a queue fixture (`./c4drill skill/examples/10-edge-kinds.c4d`), rasterize (`mkdir -p /tmp/pipe-check && qlmanage -t -s 1400 -o /tmp/pipe-check skill/examples/10-edge-kinds.svg`), and Read the PNG. Verify: right end shows a full ellipse (silhouette bulge right to x1 PLUS inner face line bulging left into the body), left end shows only the half-ellipse bulge, db cylinders (rds etc.) untouched. If an arc bulges the wrong way, flip that arc's sweep flag and re-render until correct — then make the test expectations match the VERIFIED flags (the values in <interfaces> are the design intent; empirical render wins).
6. Out of scope — do NOT touch: the legend (level-qualified rows, invisible content cluster), labels.go, pipeEndRatio, replaceQueuePolygon's ID matching, or any non-pipe rendering path.
  </action>
  <verify>
    <automated>go test ./internal/render/ -run 'Pipe|Queue|Replace' -v 2>&1 | tail -15</automated>
  </verify>
  <done>pipe.go emits the locked two-subpath d (closed clockwise body outline + cap-face subpath with sweep 0); no arc command with coincident endpoints can be emitted (regression-tested); all pipe/queue tests in internal/render green; qlmanage PNG of 10-edge-kinds shows the full-ellipse right cap next to untouched db cylinders</done>
</task>

<task type="auto">
  <name>Task 2: Re-render example SVGs, sync skill plugins, full test + lint gates</name>
  <files>skill/examples/*.svg, examples/cloud-system/*.svg, examples/cloud-system/cloud-system/amazon/sqs.svg, plugins/c4drill/skills/c4drill-toml/, plugins/c4drill/opencode/skills/c4drill-toml/</files>
  <action>
1. Rebuild the binary (skip if Task 1 already built it after the final code state): `go build -o c4drill ./cmd/c4drill`.
2. Re-render all tracked examples (exact loop from the locked design):
   `for f in skill/examples/*.toml skill/examples/*.c4d examples/*/*.toml; do ./c4drill "$f"; done`
   plus expanded variants: `./c4drill examples/cloud-system/cloud-system.toml --expanded` and `./c4drill skill/examples/06-templates.toml --expanded`.
   Expected churn: only queue-pipe SVGs change (skill/examples/10-edge-kinds.svg, 10-edge-kinds/app.svg, 10-edge-kinds/ext.svg, examples/cloud-system/cloud-system.expanded.svg, examples/cloud-system/cloud-system/amazon/sqs.svg) — the pipe d gains the top-edge L command, sweep-flag changes, and the trailing face subpath. Everything else must be byte-identical.
3. Spot-check the re-rendered queue SVG still shows the full ellipse: re-rasterize 10-edge-kinds.svg with qlmanage into /tmp/pipe-check and Read the PNG (final visual confirmation on committed-to-be bytes).
4. CI parity hygiene (CI runs diff -r on a pristine checkout — no untracked artifacts may remain): `git clean -f plugins/` and `git clean -x -f skill/examples` (this also removes qlmanage PNG byproducts), then sync the skill:
   `rsync -a --delete skill/ plugins/c4drill/skills/c4drill-toml/ && rsync -a --delete skill/ plugins/c4drill/opencode/skills/c4drill-toml/`
5. Full gates: `go test ./...` green — DOT goldens must show ZERO diff (multilevel.expanded.dot is queue-free; canonical goldens unaffected by an SVG-only post-processor) — and `golangci-lint run ./...` reports 0 issues.
6. Do NOT commit the /tmp/pipe-check artifacts (they live outside the repo); commit the re-rendered SVGs + synced plugins with the code change.
  </action>
  <verify>
    <automated>go test ./... 2>&1 | tail -8 && golangci-lint run ./... 2>&1 | tail -3 && git status --porcelain plugins/ skill/examples/ | grep -c '^??' || true</automated>
  </verify>
  <done>Only queue-pipe SVGs churned; plugins synced and free of untracked files; go test ./... green with zero golden churn; golangci-lint 0 issues; the rendered queue pipe reads as a pipe with a full-ellipse right end cap</done>
</task>

</tasks>

<threat_model>
## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| user TOML → node ID → SVG title search | Unchanged from 260828-qbx: queue node IDs from user-authored TOML are matched via plain string search (T-QBX-01 mitigation untouched) |

## STRIDE Threat Register

| Threat ID | Category | Component | Disposition | Mitigation Plan |
|-----------|----------|-----------|-------------|-----------------|
| T-TGF-01 | Tampering | pipe.go d-string generation | mitigate | Geometry emitted only from parsed GraphViz polygon coordinates (floats via strconv.ParseFloat, already validated in parsePointsBBox); no user-controlled strings are interpolated into the d attribute — the new subpath adds only computed numeric coordinates |
| T-TGF-02 | DoS | pipe.go path building | accept | Same bounded single-pass construction as before; the second subpath adds O(1) work per queue node |
| T-TGF-SC | Tampering | npm/pip/cargo installs | accept | No package-manager installs in this task set; stdlib + existing deps only |
</threat_model>

<verification>
- `go test ./...` green; DOT goldens zero diff (queue-free)
- `golangci-lint run ./...` = 0 issues
- Regression test proves no coincident-endpoint arc is ever emitted and the cap face subpath (sweep 0, (bodyR,y0)→(bodyR,y1)) exists
- qlmanage PNG of skill/examples/10-edge-kinds.svg shows the right end as a full ellipse; db cylinders untouched
- `git status --porcelain plugins/ skill/examples/` shows no untracked files after clean+rsync
</verification>

<success_criteria>
Queue pipes render with a full ellipse on the right end — the single <path> d attribute carries a closed clockwise body outline (top edge, right outer arc sweep 1 through (x1,cy), bottom edge, left cap arc sweep 1 through (x0,cy)) plus a cap-face subpath (M bodyR,y0, arc sweep 0 to bodyR,y1 through (bodyR-rx,cy)); no degenerate coincident-endpoint arc can be emitted (regression-tested); the pipe still fills the exact polygon bbox; non-queue rendering, legend, labels.go, and DOT output are untouched; example SVGs re-rendered and plugins synced with full gates green.
</success_criteria>

<output>
Create `.planning/quick/260828-tgf-fix-pipe-end-cap-right-side-must-render-/260828-tgf-SUMMARY.md` when done
</output>
