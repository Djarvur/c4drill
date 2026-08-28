# Phase 36 Research: Edge Semantics and Legend

**Phase:** 36 (v1.13, single-phase milestone, product release v1.18.0)
**Requirements:** COLOR-01..02, GEDGE-01..02, RANK-01..02, KIND-01..03, AGG-01..03, LEG-01..03, BC-01, DOC-01..03, REL-01
**Method:** Read-only codebase trace (builder → view → render → c4d → parser → testutil), go-graphviz v0.10.0 API inspection, full test-suite green-run baseline (`go test ./...` = all 17 packages ok, 2026-08-28).
**Baseline note:** suite is green at research time; the ONLY asserted DOT golden today is `cmd/c4drill/testdata/multilevel.expanded.dot`.

---

## 1. Verified fix points (ROADMAP scan confirmed, with corrections)

Every pre-confirmed fix point in ROADMAP.md was verified against source. Two corrections/refinements:

1. **The C2/C3 `properties.edges` fallback is 2 lines in view generators, not a builder change.** `internal/view/scope.go:377` (`Edges: systemUnit.Edges`) and `:470` (`Edges: containerUnit.Edges`) are the only gaps; C1 (`:128`) and expanded (`:22`) already read `m.Properties.Edges`.
2. **There are FIVE link-copy sites that must learn `Link.Kind` (not four):** the 4 view copiers in `scope.go` (`resolveBoundaryLink` :331, `resolveBoundaryLinks` :831, `addResolvedCrossLink` :970, `addResolvedCrossLinkFrom` :1003) **plus** the validator mirror `internal/validator/index.go:70-81` (`populateIncomingLinks`). Missing the validator mirror silently loses kind on validator-synthesized incoming links (which `buildEdges` consumes in views without ResolvedLinks).

Full fix-point table:

| Req | Fix point | File:line (verified) |
|---|---|---|
| COLOR | Unit.Color/Style/Border parsed (TOML struct tags `internal/model/unit.go:55-59`; C4D `FieldKey` grammar + `unitStringField` `internal/c4d/tomodel.go:465-486`) with **zero render-side reads**. Palette-only styling at `buildNode` (`internal/graph/builder.go:266-302`), `buildCluster` (:305-351), `buildNestedCluster` (:205-263), `buildBoundaryCluster` (:108-145) via `GetStyleForType`/`GetBoxStyleByContents` (`internal/graph/shapes.go:111-257`). Emission already generic: `applyNodeStyle` (`internal/render/converter.go:299-318`) and `applyClusterStyle` (:389-427) emit whatever `NodeStyle` carries. | builder.go + shapes.go |
| GEDGE | `Edges` copy: scope.go:128 (C1 ok), :377 (C2 gap), :470 (C3 gap), scope.go:22 (expanded ok). `"square"` unimplemented: `configureGraphSettings` switch `internal/render/converter.go:162-169` handles only spline/straight/ortho. | scope.go, converter.go |
| RANK | `model.Link.Rank` parses + round-trips (`internal/model/link.go:48-49`, D-22 canonical defaults normalize `""`↔`"forward"` in `internal/c4d/canonequal.go:98-100` and canonsrc). Only `rank="equal"` is consumed: `builder.go:608` → `Edge.NoConstraint` → `converter.go:517-519` `SetConstraint(false)`. Edge dir emission: `converter.go:483-493`. | builder.go, converter.go |
| KIND | No `Link.Kind` anywhere. Touch list in §6. | §6 |
| AGG | Pair dedup is first-wins: `markSeen` on `edgeKey = path+"->"+peer` in `processOutgoingLinks`/`processIncomingLinks` (builder.go:482-559). A per-pair pre-scan precedent already exists: `countPairMultiplicity` (builder.go:421-479) with the `v.AllExpanded` opt-out. | builder.go |
| LEGEND | `graph.Legend` placeholder (`internal/graph/graph.go:151-155`) referenced only by `graph.Graph.Legend` (:48) and a trivial struct test (`graph_test.go:21`). Top-label HTML table: `configureGraphSettings` converter.go:183-243 (`SetLabelHTML` + `SetLabelLocation(cgraph.TopLocation)`), nav TDs from `internal/render/navigation.go`. Pipeline: `cmd/c4drill/root.go` `processView` :295-338 (BuildGraphWithPath), `processExpandedView` :341-378 (BuildExpandedGraph). | converter.go, graph.go |

**Plumbing fact the planner must design around:** `graph.BuildGraph`/`BuildExpandedGraph` receive only `*view.View` — never the model. New global settings (legend on/off, custom lines) must ride `view.View` as new fields, populated by the 4 view generators from `m.Properties` (the exact pattern `Edges:` already uses).

---

## 2. RANK-01 — exact reverse-rank semantics (deep dive)

### 2.1 What today's `"<-"` + `arrow = "reverse"` idiom produces at DOT level

Traced end-to-end (TOML `[[unit.linkFrom]] peer="b"` on unit `a`, `arrow="reverse"`):

1. Parser → `unit.LinksFrom` entry {Peer:"b", Arrow:"reverse"}.
2. Builder `processIncomingLinks` (builder.go:531) → `createEdge(link.Peer, path, ...)` → `Edge{Source:"b", Target:"a", ArrowHead:ArrowReverse}`.
3. Renderer `createEdges` → `CreateEdgeByName(name, bNode, aNode)`; `ArrowReverse` case → `e.SetDir(cgraph.BackDir)` (converter.go:487-488).
4. **DOT: `b -> a [dir=back, ...]`** — b ranks ABOVE a (GraphViz ranks tail above head in `rankdir=TB`), arrowhead drawn at the tail (b). Visually: line from a pointing AT b, b on top.

**`rank = "reverse"` must produce exactly this DOT shape for a plain outgoing link `a -> b`:** swap endpoints (`b -> a`) + `dir=back`. Same visual arrow (pointing at b), opposite vertical order (b on top). This is the equivalence the requirement demands, and it is byte-exact at the canonicalDOT level (same attr set the old idiom produces), so a regression test can literally compare canonical forms of the two idioms.

### 2.2 API verification (go-graphviz v0.10.0, the pinned fork)

- `(*Graph).CreateEdgeByName(name string, start, end *Node)` — endpoint order is a plain argument; swapping is trivial (cgraph/cgraph.go:979).
- `(*Edge).SetDir(v DirType)` → `SafeSet("dir", v, ForwardDir)` (attribute.go:593) — **`SetDir(ForwardDir)` emits nothing** (value equals the declared default), which is why the omitted-arrow default (`""`) and explicit `forward` produce identical DOT today. Preserve this: do NOT start emitting `dir=forward`.
- `DirType` consts: Forward/Back/Both/None (attribute.go:581-584).

### 2.3 Recommended mechanism (renderer-side swap, logical Edge unchanged)

Add `Edge.RankReverse bool` (set in builder `createEdge` from `link.Rank == model.RankReverse`), keep `Edge.Source/Target` **logical** (a→b). In `converter.createEdge` (converter.go:443-498), when `RankReverse`:

- Call `CreateEdgeByName(edgeName, target, source)` (swap at emission only). Keep `edgeName` built from the logical pair — names are layout-irrelevant and keeping them logical preserves test readability.
- Map dir: authored `""`/`forward` → `SetDir(BackDir)`; authored `reverse` → `SetDir(ForwardDir)` (emits nothing — arrowhead at head = original source, exactly the authored semantics); `bidirectional` → `BothDir`; `none` → `NoneDir` (both swap-invariant).

**Direction matrix (the interaction rules to encode in tests):**

| authored arrow | rank=default/forward | rank=reverse |
|---|---|---|
| "" (omitted) | `a -> b` (no dir attr) | `b -> a [dir=back]` |
| forward | `a -> b` (no dir attr) | `b -> a [dir=back]` |
| reverse | `a -> b [dir=back]` | `b -> a` (no dir attr) |
| bidirectional | `a -> b [dir=both]` | `b -> a [dir=both]` |
| none | `a -> b [dir=none]` | `b -> a [dir=none]` |

**rank=reverse × rank=equal:** impossible by construction — `Link.Rank` is a single enum (`forward|reverse|equal`, model/link.go:18-28); `NoConstraint` and `RankReverse` are mutually exclusive branches on the same field. No extra rule needed; a parse-level test pins it.

**minlen/constraint under swap:** `minlen` is symmetric (rank distance); the swapped edge keeps default `constraint=true`, which is precisely what forces the flipped ranking. Penwidth/label/style are direction-agnostic.

### 2.4 Survival through resolution and expanded views (RANK-02)

- All 4 scope.go copiers **already copy `Rank`** (:337, :835, :976+1008) and the validator mirror copies it (index.go:73) — nothing to add for Rank itself; add copy-site regression tests only if the plan touches the copiers for Kind anyway (recommended: one table-driven test per copier covering Kind+Rank).
- Expanded mode (`v.AllExpanded`): links are consumed raw from `Unit.Links`; dedup key stays `path->peer:tech:desc` (logical direction) — the swap happens purely at emission, so parallel-edge dedup and penwidth are untouched.
- Pair/multiplicity keys (`countPairMultiplicity`, `markSeen`) all use logical direction — do not key on the swapped pair.

### 2.5 Docs consequence

`rank = "reverse"` replaces the `"<-"` + `arrow="reverse"` idiom. README places: line 159 ("Effect of Rank" section), lines 528-529 (link attribute table), lines 960/984 (C4D side-by-side using `arrow: reverse`). The old idiom keeps working (arrow semantics unchanged) — document reverse as the *convenient* form, do not remove the idiom.

---

## 3. KIND — palette recommendation and home

**Recommended constants (add to `internal/model/colors.go`, the established palette home):**

```go
// Link-kind colours (KIND-01). Dark enough for legible edge-label text
// (fontcolor rides the same value) on the white SVG background; all three
// are distinct from the blue C4 node palette (#08427B/#1168BD/#438DD5/#85BBF0),
// the gray external palette (#999999/#B3B3B3/#CCCCCC), and the default edge
// colors (border blues + #666666).
const (
    LinkReadColour      = "#2E7D32" // green  (Material green 800, ~5.1:1 on white)
    LinkWriteColour     = "#C62828" // red    (Material red 800,   ~5.9:1 on white)
    LinkReadWriteColour = "#6A1B9A" // blend: purple — hue-distinct from BOTH green
                                    // and red, ~7.4:1 on white, no collision with
                                    // node/edge palette; reads clearly in a legend row
)
```

Rationale for purple as "blend": a literal green↔red blend is muddy brown (illegible, ambiguous with grays); purple is hue-orthogonal to both, satisfies "blend colour distinct from both", and stays distinct in the legend. Alternative if the user objects: dark orange `#E65100` (between green and red on the wheel, also distinct from the blue palette) — planner picks one; do not pick both.

**Mapping location:** a small `kindColour(kind model.LinkKind) string` helper in `internal/graph` (builder side, next to `createEdge`), consuming the model constants. The renderer stays palette-free — it just emits `edge.Color`. Legend defaults (LEG-02) read the same constants so the legend can never drift from the renderer ("matching the palette actually used by the renderer" — put the single source in `model/colors.go` and have both graph builder and legend builder consume it).

**Per-edge colour precedence (KIND-01/KIND-02) — new `createEdge` logic:**

```
link.Color != ""        -> link.Color            (KIND-02: explicit wins; kind still round-trips)
link.Kind  != ""        -> kindColour(link.Kind) (KIND-01)
else                    -> source unit border colour (D-01, unchanged — BC-01 safe)
```

Edge labels: `applyEdgeAttributes` (converter.go:504-529) already sets both `SetColor` + `SetFontColor` from `edge.Color`, so kind colours automatically colour the label text too. No renderer change for KIND beyond nothing — **the entire KIND render change is one precedence branch in builder `createEdge`.**

**`LinkKind` type:** `type LinkKind string` + consts `KindRead/KindWrite/KindReadWrite` in `internal/model/link.go` (mirror the `RankDirection` pattern). Values are free strings at parse (consistent with `style`/`rank` today — no validation rule exists for rank values either); unknown kinds simply don't colour. Optional stretch: a validator warning for unknown kind — out of scope unless cheap.

---

## 4. AGG — collapsed-edge aggregation: where and how

### 4.1 Computation point: a second pre-scan in `buildEdges` (builder.go)

Do **not** change `view/scope.go`. The collapse already happened by the time `buildEdges` sees the view (ResolvedLinks/CrossLinks); aggregation is a *pair-level* decision over the pre-dedup link lists the builder already consumes — exactly the data `countPairMultiplicity` walks today. Recommended shape:

```go
type pairAggregate struct {
    kinds          []model.LinkKind // every constituent's kind ("" included)
    styles         []string         // effective styles (after "solid" default)
    anyExplicitClr bool             // any constituent link.Color != ""
}
// collectPairAggregates(v) map[string]pairAggregate — skip when v.AllExpanded
// (expanded mode shows every parallel link separately; COMPAT-02 unchanged).
```

Consume it in `processOutgoingLinks`/`processIncomingLinks` at the point where `markSeen` returns **true** (the surviving first edge of the pair): if the pair has 2+ constituents, override the first edge's colour/style per the precedence below. Labels/technology keep today's first-wins (no requirement touches them). Single-link pairs keep pure per-edge semantics (§3 precedence) — AGG only *overrides* on collapsed pairs.

### 4.2 Precedence implementations

- **AGG-01 colour:** all constituents same kind → that kind colour; any mixture of kinds (incl. read+write, read+read-write) → `LinkReadWriteColour`. Only applies when **every** constituent has a kind (recommended resolution for partial-kind pairs: an unset kind breaks the aggregate → fall through to the AGG-03/default path; this keeps no-kind models byte-identical).
- **AGG-02 style:** all equal → that style; else any `"solid"` → solid; else any `"dashed"` → dashed; else dotted. (Constituents with no style default to solid first, so any unstyled constituent forces solid — deterministic and matches "all-same" semantics.)
- **AGG-03 custom-colour suppression:** "custom" is simply `link.Color != ""` — kind colours are computed at build time and never occupy `link.Color`, so **no flag/sentinel is needed** (answer to the open question). If any constituent has an explicit colour → no kind colour on the collapsed edge; fall back to today's D-01 source-border default of the drawn edge (the pair's logical source), which keeps no-kind models unchanged (BC-01).

### 4.3 Interaction with RANK

A collapsed pair whose surviving link has `rank="reverse"` keeps the swap (emission-level); aggregation only touches colour/style. Orthogonal.

---

## 5. LEGEND — implementation design

### 5.1 Data path (model → view → graph → render)

1. **model:** extend `graph.Legend` (graph.go:151) from placeholder to `type Legend struct { Entries []LegendEntry }`, `type LegendEntry struct { Label, Color, Style string }` (`Style` empty for colour swatch rows; set for line-style sample rows and custom lines).
2. **view:** add `View.Legend *graph.Legend`-shaped fields — recommended `View.ShowLegend bool` + `View.LegendLines []model.LegendLine` — populated by all FOUR generators (GenerateC1View scope.go:125, GenerateC2View :373, GenerateC3View :466, GenerateExpandedView :19) from `m.Properties` (same pattern as `Edges:`). Alternatively resolve `*graph.Legend` directly in the view package; either is fine — keep all four generators symmetric.
3. **model (properties):**
   ```go
   // Properties additions
   Legend      *bool        `toml:"legend"`     // nil = enabled (default true)
   LegendLines []LegendLine `toml:"legendLine"` // custom lines, after defaults
   type LegendLine struct { Label, Color, Style string } // in model (new file or properties.go)
   ```
   `*bool` is required because Go zero-value `false` cannot express "default on". Nil → true normalization must be applied in `c4d.CanonicalEqual` (`canonicalModelForCompare` copies `Properties` by value at canonequal.go:37 — add `if clone.Properties.Legend == nil { ... = &trueVal }`) and in canonsrc's property writer, or `convert`/parity round-trips fail when one side states the default explicitly.
4. **builder:** `BuildGraph`/`BuildExpandedGraph` assemble `g.Legend` when the view says enabled: default rows first (Read/Write/Read-Write swatch rows + Solid/Dashed/Dotted style rows), then custom lines verbatim.
5. **render:** `configureGraphSettings` (converter.go:183-243) appends legend rows to the existing top HTML table. **Change the guard** at :210: today `len(navTDs) == 0 && !hasTitle → return nil` (no label at all); legend must also trigger label emission (matters for models with empty `properties.name` — C1/expanded get no label today, only C2/C3 get nav).

### 5.2 Which views emit the label today (LEG-01 coverage check)

| View | Navigation | Title | Label today | Legend lands via |
|---|---|---|---|---|
| C1 | none | `properties.name` (may be empty) | only if name set | same guard change |
| C2/C3 | breadcrumbs (BuildGraphWithPath :651-655) | view title (always non-empty) | always | existing table |
| expanded | none (comment "no navigation for expanded view") | `properties.name` | only if name set | same guard change |

So: **all four view kinds get the legend through the single `configureGraphSettings` table** — no per-view work beyond the view-field plumbing.

### 5.3 Position: the GraphViz constraint (confirmed)

GraphViz cannot absolutely position the graph label; `labeljust` affects only plain (non-HTML) labels; HTML-label tables cannot span the canvas width (no `width=100%`). Therefore "upper-right" renders as: **top-of-diagram label block (existing behaviour), legend right-aligned *within* that block.** Implementation: legend rows as additional `<TR>`s whose single `<TD ALIGN="RIGHT" COLSPAN=n>` holds a nested 2-column borderless table (swatch cell + label cell), mirroring the COLSPAN technique the title row already uses. Every row must wrap content in explicit `<FONT POINT-SIZE="10">` tags — the documented quirk that silently drops rows mixing font sizes (converter.go:196-206). Swatch cells: `<TD BGCOLOR="#2E7D32" WIDTH="12" HEIGHT="8"> </TD>` (BGCOLOR+fixed size is the reliable GraphViz swatch mechanism; the queue icon already proves box-drawing glyphs render). Line-style samples cannot be drawn as dashed/dotted *strokes* inside HTML labels — use text glyphs in the default edge colour `#666666`: e.g. `───` (U+2500), `- - -`, `· · ·` (planner picks exact glyphs; keep them in one constant block next to `navFont*` in navigation.go). Custom label text: `html.EscapeString` (T-03-04-02 discipline); colours land in an attribute → escape too.

### 5.4 Parse design — TOML

`[[properties.legendLine]]` array-of-tables is **parse-safe today**: `captureDefinitionOrder` only records `expr.Kind == unstable.Table` headers as units (parser.go:653-658; ArrayTable is skipped — the "[[a.link]] etc. are not units" comment), so `[[properties.legendLine]]` cannot register a phantom `properties` unit. The `[properties]` extraction path (rawMap → `toml.Marshal` → `Unmarshal` into `model.Properties`, parser.go:150-159) picks up both new fields automatically via struct tags. `legend = false` scalar rides the same path into `*bool`.

Canonical emission (BC-critical): `emitPropertiesTOML` (emit_toml.go:64-105) must emit `legend = false` **only when explicitly false** (nil/true omitted → source-level byte stability for models not using the feature) and `[[properties.legendLine]]` blocks after `expanded` (define the D-23 slot; also update `propertyFieldRank` in emit_c4d.go:37-46 and the canon-src sorted-key writer, which needs the nil==true normalization).

### 5.5 Parse design — C4D

- `properties { }` grammar: `PropertyKey` is a closed alternation (c4d.peg:666). Add `"legend"` and `"legendLine"` to it + regen (`//go:generate pigeon -o parser_gen.go -nolint c4d.peg`, grammar/doc.go:12; pigeon pinned in tools.go). `tomodel.applyPropertyField` (tomodel.go:285-312) gains `case "legend"` (parse bareword `"true"/"false"`, error on anything else) and `case "legendLine"` (list value).
- **C4D has no boolean literals today** (`once` is a trailing keyword, c4d.peg:612). So `legend: false` is a bareword string parsed by `applyPropertyField` — fine, but pin it with tests for `legend: true/false/garbage`.
- **Custom lines authoring (recommended):** list-of-strings with pipe-split fields, symmetric with the D-09 label shorthand and the `expanded: [...]` list precedent — zero grammar changes beyond PropertyKey:
  ```
  legendLine: ["Reads from archive|#2E7D32", "Batch import|#C0392B|dashed"]
  ```
  split each item on the first pipe → label, colour, optional style. Constraint to document: label must not contain `|` (same class of constraint as C4D edge labels). The alternative (repeatable `legendLine { }` sub-blocks) requires grammar/emitter surgery beyond PropertyKey — not worth it.
- `frommodel.propertiesBlockFromModel` (frommodel.go:64-99) emits `legend` (only when false) and a composed `legendLine` list; `emit_c4d` field-rank table updated.

---

## 6. KIND-03 — the full touch list (widest requirement; use as plan checklist)

1. `internal/model/link.go` — `LinkKind` type + 3 consts + `Link.Kind` field (place after `Rank`, before `Color`, to define emission order).
2. `internal/validator/index.go:70-81` — mirror copy gains `Kind:` (else kind vanishes on validator-synthesized LinksFrom).
3. `internal/view/scope.go` — all 4 copiers gain `Kind:` (:331, :831, :970, :1003).
4. `internal/c4d/grammar/c4d.peg:458` — `OptionKey` alternation += `"kind"`; regen parser_gen.go via pigeon.
5. `internal/c4d/tomodel.go:701-728` — `applyEdgeOption` `case "kind": link.Kind = model.LinkKind(v)`.
6. `internal/c4d/frommodel.go:183-223` — `edgeStmtFromLink` `addOpt("kind", string(link.Kind))` in canonical option order (arrow, rank, **kind**, color, style, labelPosition, length).
7. `internal/c4d/emit_toml.go:203-241` — `emitLinkTOML` writes `kind` after `rank`; update the D-23 doc comment (and canonsrc `writeLinkCanonTOML` for parity tests).
8. `internal/c4d/emit_c4d.go:44-51` — `edgeOptionRank` += `"kind": 2` (shift others).
9. `internal/c4d/grammar/reserved.go:33-39` — `fieldKeywords` += `"kind"` (unknown-key suggestions).
10. `internal/testutil/canonsrc` — TOML link writer + C4D edge-option writer carry `kind`.
11. `internal/template/expand.go:348-355` — **decision:** `applySubstitutionLink` substitutes Peer/Description/Technology/Color/Style but NOT the enum-ish fields (Arrow/Rank/LabelPosition/Length). Recommended: do NOT substitute `Kind` (follow the enum precedent — substitution could silently produce an invalid kind), and interpret KIND-03's "${param} substitution inside templates" as *round-trip coverage*: a template-body link carrying a literal `kind` must survive Clone+instantiation intact (it does — Link value-copy) and get a `template/expand_test.go` case. If the planner instead reads it as "kind must be substitutable", it's a one-line add + test — flag as an explicit plan decision.
12. `internal/c4d/canonequal.go` — no change needed for Kind (no default to normalize); Properties changes need the `Legend` nil==true normalization (§5.1.3).

---

## 7. COLOR-01/02 — mechanism confirmation

Only **builder-side** work; the emission path is already generic and untouched:

- `buildNode` (builder.go:266): after the palette/box-style lookup, apply overrides from `entry.Unit`:
  - `Color` → `NodeStyle.FillColor` (render's `buildStyleString` auto-adds `"filled"` when fillColor non-empty — converter.go:60-72 — so background colour needs nothing else).
  - `Border` → `NodeStyle.BorderColor` (README documents `border` as a *color*, `style` as the border *style* — "solid|dashed|dotted", README §Styling ~line 610).
  - `Style` → `NodeStyle.BorderStyle`; renderer `buildStyleString` currently only maps dashed → must also map `dotted` (one-line extension; `dotted` is a legal GraphViz node style).
- **Font colour decision (needed because there is no unit fontColour field):** when an explicit dark `Color` fills a node/cluster, the existing font colour (level border blue on transparent) may be unreadable. Recommended: luminance rule — fill luminance < 0.5 → `FontColor = #FFFFFF`, else keep the level default (C1/C2 border-coloured, C3 black per `getLevelStyle`). One small helper in graph; test with `#4A90D9` (light → keep) and `#08427B`-like darks (→ white).
- `buildCluster` (:305), `buildNestedCluster` (:205), `buildBoundaryCluster` (:108) get the same override (clusters emit via `applyClusterStyle`: `SetBackgroundColor`, `fontcolor`, `color`, style string — all already conditional on non-empty, converter.go:389-427).
- **Box interaction:** `GetBoxStyleByContents` (shapes.go:111) computes content-based box styles — explicit author fields should win over computed box styling (author intent beats heuristics); apply overrides AFTER the box branch in all four build fns.
- `skill/examples/04-styling.toml` (+ .c4d twin) is exactly the author expectation surface: light fill colours + dark borders per unit, `edges = "ortho"` per-unit — it renders today with all styling silently dropped. It becomes the COLOR-01/02 golden fixture; extend it with the legend custom lines for DOC-03.

---

## 8. GEDGE-01/02 — confirmation

- **GEDGE-01:** `internal/view/scope.go:377` → `Edges: cmp.Or(systemUnit.Edges, m.Properties.Edges)` (same at :470; `cmp.Or` is available, go.mod says go 1.26.1). Per-unit wins already by this ordering; expanded/C1 already global. The view-level `Edges` applies view-wide (that IS "per-unit edges" semantics today — the drilled unit's field styles the whole diagram), so no builder change.
- **GEDGE-02:** `SetSplines` is a raw string SafeSet (attribute.go:2541) — the engine accepts `true|false|line|spline|ortho|polyline|compound`; any other value (like today's `"square"`) is silently ignored. **Decision: map `"square"` → `SetSplines("ortho")`** in the converter switch (converter.go:162-169) — the documented value gets real behaviour, docs stay truthful. Note the GraphViz caveat for docs: `splines=ortho` ignores/misplaces edge labels in some engine versions (acceptable; fixture suite uses spline).

---

## 9. BC-01 golden impact inventory

**Asserted goldens (must re-baseline):**
- `cmd/c4drill/testdata/multilevel.expanded.dot` — the ONLY committed DOT golden under comparison, used by 2 canonical tests: `internal/graph/builder_test.go:1209` (TestBuildExpandedGraphBaselineDOT) and `:2385` (REF-05 backward-compat). Regenerate: `go run ./cmd/c4drill cmd/c4drill/testdata/multilevel.toml --format dot --expanded`. `multilevel.toml` sets no new properties → default-on legend → the graph `label=<...>` attr changes → this is THE BC-01 accepted delta. Re-baseline in a dedicated commit.

**Committed but NOT asserted (stale artifacts; regenerate for hygiene, optional):**
- `cmd/c4drill/testdata/expanded.dot`, `expanded/mainsystem.dot`, `expanded/mainsystem/webapp.dot`, `multilevel/*.dot`, and the 18 `.svg` files — no test compares them (root_test only does existence checks and greps files it generates into tmp dirs). Recommend regenerating alongside the golden commit or noting them as decorative.
- `examples/**/*.svg` (README `image::` refs: cloud-system, overflow-test, rank-for-better-layout) and `skill/examples/06-templates.svg(+expanded)` — regenerated for DOC-03/docs accuracy; not test-enforced.

**In-test expectations that change (update as part of TDD, not re-baselining):**
- `internal/render/converter_test.go` / `navigation_test.go` — any test asserting the combined label HTML or the `len(navTDs)==0 && !hasTitle` early-return shape (legend adds rows + changes the guard).
- `internal/testutil/canonsrc` round-trip fixtures — once `legend`/`legendLine`/`kind` join the canonical writers, existing normalize tests stay green only if the writers emit nothing for absent fields (they do — putString pattern).
- `cmd/c4drill/root_test.go:1447` TestC4DRenderDirect + `internal/c4d/parity_test.go:842` — both sides render with the same legend → comparisons stay green without edits.
- `graph_test.go:21` Legend placeholder assertions — trivially satisfied by the extended struct.
- `cmd/c4drill/testdata` corpus tests (TestCLICorpusRendersUnchanged etc.) — validate, don't compare bytes → green.

**Byte-stable guarantees to protect:** models using no new feature must produce identical DOT *except* the graph label attr (legend), identical SVG except the label table, and byte-identical canonical source emission (legend/legendLine/kind omitted when unset) — cover with a dedicated no-feature regression test (parse valid.toml → EmitTOML → byte-equal to pre-phase output captured in-test).

---

## 10. Validation Architecture

TDD mode is ON (config workflow.tdd_mode) — every task RED→GREEN. All DOT assertions use `internal/testutil/canonical` (DI-1: order-insensitive, geometry-stripped). Source round-trips use `internal/testutil/canonsrc`. Prefer in-test canonical comparisons over new committed golden files.

| Req | Proving tests (package → file) |
|---|---|
| COLOR-01/02 | `internal/graph` → builder_test.go: explicit Color/Style/Border override NodeStyle on plain node AND cluster (incl. buildNestedCluster + boundary cluster); box-with-override wins over content style; unset fields keep palette. `internal/render` → converter_test.go: applyNodeStyle/applyClusterStyle emit fillcolor/fontcolor/color/style(+dotted) attrs in DOT (canonical compare). |
| GEDGE-01 | `internal/view` → scope_test.go: C2/C3 `Edges` precedence (unit edges wins; falls back to properties.edges; C1/expanded unchanged). `cmd/c4drill` → root_test.go: end-to-end `properties.edges="straight"` appears as `splines=false` in generated C2/C3 DOT. |
| GEDGE-02 | `internal/render` → converter_test.go: `"square"` → `splines=ortho` attr; spline/straight/ortho regressions. |
| RANK-01 | `internal/graph` → builder_test.go: `Edge.RankReverse` set from link; logical endpoints preserved. `internal/render` → converter_test.go: **the money test** — canonical DOT of `a→b rank=reverse` equals canonical DOT of today's `linkFrom + arrow="reverse"` idiom; the full arrow×rank matrix from §2.3; no `dir=forward` attr ever emitted. |
| RANK-02 | `internal/view` → scope_test.go: each of the 4 copiers carries Rank(+Kind) (table-driven). `internal/graph` → builder_test.go: rank=reverse survives C1 boundary resolution and renders swapped in `--expanded`. |
| KIND-01/02 | `internal/graph` → builder_test.go: per-edge colour precedence (explicit > kind > source border) with the §3 hex values; `internal/render` → converter_test.go: color+fontcolor emitted. |
| KIND-03 | `internal/parser` → parser_test.go (TOML link kind round-trip); `internal/c4d` → parse_test/tomodel_test (option incl. `${param}` body case), emit_test (canonical option order incl. kind), parity_test (both front-ends → equal Model + canonical src), representable_test (no new constraint); `cmd/c4drill` → convert_test/fmt_test (convert + fmt round-trips both directions); grammar suggestion: unknown-key error suggests "kind" (c4d parse_test, reserved.go fieldKeywords); `internal/template` → expand_test.go: kind survives instantiation (+ substitution decision pinned). |
| AGG-01..03 | `internal/graph` → builder_test.go: hand-built views with 2+ ResolvedLinks per pair — all-read → #2E7D32, all-write → #C62828, mixed → #6A1B9A; style matrix (all-same/any-solid/any-dashed/else-dotted); explicit colour on any constituent → source-border default; 1-link pair unaffected; AllExpanded unaffected. |
| LEG-01 | `internal/parser` → parser_test.go: `legend=false` parses; absent → nil (default-on semantics). `internal/view` → scope_test.go: all 4 generators populate ShowLegend/LegendLines from properties. `internal/render` → converter_test.go: label table present on C1 (even with empty name), C2, C3, expanded; absent when disabled. |
| LEG-02 | `internal/render` → converter_test.go: default rows carry the §3 hexes + 3 style samples, sourced from model consts (single-source check). |
| LEG-03 | `internal/parser` + `internal/c4d` parse tests for `[[properties.legendLine]]` and C4D `legendLine: [...]`; `internal/render` → custom lines render AFTER defaults, escaped; emit_toml/emit_c4d/canonsrc round-trip. |
| BC-01 | `internal/graph` → builder_test.go golden pair (re-baselined) + a NEW no-feature regression test (no-feature model → EmitTOML byte-stable; render DOT differs only in the graph label statement). `cmd/c4drill` → full corpus suite green. |
| DOC-01..03 | Docs sweep with a rendered-example check: every new README/SKILL snippet actually runs (`go run ./cmd/c4drill skill/examples/<new>.toml -f dot`); sync check `diff -r skill plugins/c4drill/skills/c4drill-toml plugins/c4drill/opencode/skills/c4drill-toml`. |
| REL-01 | Tag `v1.18.0` triggers `.github/workflows/release.yml` (on: push tags v*) — verify workflow file unchanged and final tag push is the last phase task. |

---

## 11. Docs surface inventory (DOC-01..03)

**README.adoc sections to touch:**
- `=== Properties Section` (~195): add `legend` + `[[properties.legendLine]]` to the properties table; clarify `edges` values (`square` = ortho routing).
- `==== Link Attributes` (~520): add `kind` row; annotate `rank` with the reverse semantics.
- NEW subsection after Links (suggest `==== Edge Kinds and the Legend`): kind colours + legend setting + custom lines (both formats).
- `=== Styling` (~602): keep the triple, note it now actually renders (and dotted border support).
- `=== Effect of Rank` (~155): document `rank = "reverse"` as the single-knob replacement for the `"<-"` + `arrow="reverse"` dance (old idiom still works).
- `== C4D Format` → `=== Fields` (~800) + `=== Edges` (~847): kind option in the option block; legend/legendLine properties; `=== Side by Side with TOML` (~945) rows for the new keys.
- Regenerate referenced example SVGs (`examples/cloud-system`, `overflow-test`, `rank-for-better-layout`).

**skill/SKILL.md sections to sync (source of truth `skill/SKILL.md`):**
- `### C4D Syntax Cheat-Sheet` (~101), `### [properties] Section` (~172), `### Link Syntax` (~419), `### Link Attributes` (~464), `### Examples` (~771), `## Generation Guidelines` (~805) + `## Common Mistakes` (~827) (mention legend default-on and kind colours when advising on colours).

**Plugin copies (all must land in the same commit — DOC-02):**
- `/Users/nil/DiskD/W/Djarvur/c4drill/skill/SKILL.md` + `skill/examples/` (source)
- `/Users/nil/DiskD/W/Djarvur/c4drill/plugins/c4drill/skills/c4drill-toml/` (SKILL.md + examples/)
- `/Users/nil/DiskD/W/Djarvur/c4drill/plugins/c4drill/opencode/skills/c4drill-toml/` (SKILL.md + examples/)
- **Pre-existing drift to resolve while syncing:** plugin copies' `04-styling.toml` differ (comment spacing) and `06-templates`/`06-templates*.svg` exist only under `skill/examples/` — decide sync-all or document the delta; no CI test enforces sync today.

**Example fixtures (DOC-03):** extend `skill/examples/03-links.toml/.c4d` with `kind`, `skill/examples/04-styling.toml/.c4d` with a custom legend line (and they now visibly render); optionally add `10-edge-kinds.toml/.c4d` demonstrating kinds + aggregation + rank=reverse. Note: `.github/workflows/validate-skill-examples.yml` validates only `skill/examples/*.toml` (+ `examples/*/*.toml`) — new `.c4d` fixtures are NOT CI-validated (optional tiny workflow extension; flag for the planner).

**REL-01:** existing release workflow triggers on `v*` tags — no CI change needed; tag `v1.18.0` at phase close.

---

## 12. Open decisions for plan-phase (with recommendations)

1. **Legend boolean representation** — `*bool` nil-defaults-true (recommended; needs canonequal/canonsrc normalization) vs inverted `DisableLegend bool` (no normalization, uglier docs).
2. **C4D custom legend lines** — pipe-split list literal (recommended, zero grammar surgery) vs new sub-statement grammar.
3. **kind in template `${param}` substitution** — not substituted, enum precedent (recommended) vs substitutable; either way KIND-03's round-trip coverage needs an expand_test case.
4. **Legend content policy** — fixed default block always shown (recommended: deterministic goldens, matches LEG-01/02 wording) vs only-what-the-diagram-uses.
5. **AGG partial-kind pairs** — aggregate only when ALL constituents have a kind (recommended; BC-safe) vs treating unset as wildcard.
6. **Explicit-unit colour vs box content styling precedence** — author wins (recommended).
7. **Font colour under dark explicit fills** — luminance rule (recommended) vs fixed level default.
8. **Palette sign-off** — `#2E7D32` / `#C62828` / `#6A1B9A` (recommended) needs user-visible confirmation in the legend mock during UAT.
9. **Stale non-asserted goldens** under cmd/c4drill/testdata — regenerate (recommended) or declare decorative.
10. **Suggested plan slicing** (from dependency order): (a) model+builder colour/kind/rank core with render emission, (b) view plumbing GEDGE + copier Kind/Rank + AGG, (c) legend model→view→graph→render + parse/emission both formats + golden re-baseline, (d) docs/skill sync + examples + tag. KIND touches the most files; the C4D grammar regen (pigeon) is the only tool-dependent step.
