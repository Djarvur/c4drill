# Roadmap: C4Drill

## Milestones

- ✅ **v1.0 Initial Release** — Phases 1-6 (shipped 2026-03-10)
- ✅ **v1.7 Queue Label Fix & Visual Improvements** — Phases 7-18 (shipped 2026-03-29)
- ✅ **v1.8 Proper C1/C2/C3 View Generation** — Phases 19-26 (shipped 2026-08-06) → [archive](milestones/v1.8-ROADMAP.md)
- ✅ **v1.9 C3 Boundary Node Fix** — Phase 27 (shipped 2026-08-06) → [archive](milestones/v1.9-ROADMAP.md)
- ✅ **v1.10 Model Composition** — Phases 28-33 (shipped 2026-08-08) → [archive](milestones/v1.10-ROADMAP.md)
- ✅ **v1.11 Label Formatting Fixes** — Phase 34 (shipped 2026-08-10) → [archive](milestones/v1.11-ROADMAP.md)
- ✅ **v1.12 C4D DSL Alternative** — Phase 35 (shipped 2026-08-17) → [archive](milestones/v1.12-ROADMAP.md)
- 🚧 **v1.13 Edge Semantics and Legend** — Phase 36 (in progress) — product release tag: v1.18.0

## Phases

<details>
<summary>✅ v1.9 C3 Boundary Node Fix (Phase 27) — SHIPPED 2026-08-06</summary>

- [x] Phase 27: C3 Boundary Node Fix (1/1 plan) — completed 2026-08-06

Full details: [milestones/v1.9-ROADMAP.md](milestones/v1.9-ROADMAP.md)

</details>

<details>
<summary>✅ v1.10 Model Composition (Phases 28-33) — SHIPPED 2026-08-08</summary>

**Goal:** Expand C4Drill's authoring model from a single static TOML file into a composable, parametrized, multi-file format. Four additive features form a strict runtime pipeline: `include → template-expand → relative-peer-resolve → humanize → validate → generate-views → render`.

- [x] Phase 28: Reference field (📖) (1/1 plan) — completed 2026-08-08
- [x] Phase 29: Optional name humanization (2/2 plans) — completed 2026-08-08
- [x] Phase 30: Relative-peer resolution (2/2 plans) — completed 2026-08-08
- [x] Phase 31: Template expansion (2/2 plans) — completed 2026-08-08
- [x] Phase 32: Include directive (multi-file) (2/2 plans) — completed 2026-08-08
- [x] Phase 33: Docs sweep + end-to-end goldens (4/4 plans) — completed 2026-08-08

**Stats:** 6 phases, 13 plans, 35 tasks, 119 files changed (+17,703/−626). All 39 requirements validated.

Full details: [milestones/v1.10-ROADMAP.md](milestones/v1.10-ROADMAP.md)

</details>

<details>
<summary>✅ v1.11 Label Formatting Fixes (Phase 34) — SHIPPED 2026-08-10</summary>

**Goal:** Generated diagram labels render with proper word wrapping and aspect-ratio sizing — edge labels formatted like unit labels (wrapped rectangle with `LabelRatio` aspect ratio, invisible borders), and line breaks at word boundaries only (no mid-word splits).

- [x] Phase 34: Label formatting fixes (4/4 plans) — completed 2026-08-10

**Stats:** 1 phase, 4 plans, 28 commits. All 3 requirements validated (LABEL-01, LABEL-02, COMPAT-01). UAT: 3 gaps found and fixed (punctuation tokenizer, ratio sizing). Security: 9/9 threats closed.

Full details: [milestones/v1.11-ROADMAP.md](milestones/v1.11-ROADMAP.md)

</details>

<details>
<summary>✅ v1.12 C4D DSL Alternative (Phase 35) — SHIPPED 2026-08-17</summary>

**Goal:** Deliver the C4D format — a `.c4d` brace-block D2-inspired DSL with full TOML feature parity — parseable directly to `*parser.Model` and renderable through the unchanged pipeline, with bidirectional canonical-equivalent converters (`convert to-toml`/`to-c4d`), a gofmt-style comment-preserving formatter (`fmt`) for both formats, nested use (`[[unit.use]]` in TOML, `use` in blocks in C4D) and recursive template-instantiating-template expansion, plus full README/skill/example documentation.

- [x] Phase 35: Add a simple DSL alternative to the TOML diagram definition (9/9 plans) — completed 2026-08-14

**Stats:** 1 phase, 9 plans, 25 tasks. Requirements D-01..D-35 satisfied. Verification: 24/24 truths (3 gap fixes at close). UAT: 12/12. Security: 30/30 threats closed (ASVS 1).

Full details: [milestones/v1.12-ROADMAP.md](milestones/v1.12-ROADMAP.md)

</details>

### 🚧 v1.13 Edge Semantics and Legend (In Progress)

**Milestone Goal:** Make edge/colour semantics trustworthy and expressive — fix silently-dropped custom unit colours, make the global edge style apply to every generated diagram, give edge direction/ranking a single clear knob (`rank = "reverse"`), introduce edge kinds (`read`/`write`/`read-write`) with kind-derived colours that survive collapse aggregation, and add a default-on upper-right legend showing the colour semantics plus author-defined lines. Ships as product release **v1.18.0**.

- [ ] **Phase 36: Edge Semantics and Legend** - Unit styling renders, global edge style everywhere, `rank = "reverse"`, edge kinds with collapse aggregation, default-on legend, docs + release v1.18.0

## Phase Details

### Phase 36: Edge Semantics and Legend

**Goal**: Edge/colour semantics are trustworthy and expressive: author-specified unit styling actually renders (nodes and clusters), the global edge style reaches every generated diagram, edge direction gets one clear knob (`rank = "reverse"`), edges can be coloured by data-flow kind (`read`/`write`/`read-write`) with kind identity surviving collapse aggregation, and every diagram carries a default-on upper-right legend explaining the colour conventions plus author-defined lines.

**Depends on**: Phase 35 (v1.12 — C4D DSL; `kind` must be added to both formats' grammar/emitters)

**Requirements**: COLOR-01, COLOR-02, GEDGE-01, GEDGE-02, RANK-01, RANK-02, KIND-01, KIND-02, KIND-03, AGG-01, AGG-02, AGG-03, LEG-01, LEG-02, LEG-03, BC-01, DOC-01, DOC-02, DOC-03, REL-01

**Success Criteria** (what must be TRUE):
  1. Custom unit styling renders: a unit with explicit `color`/`style`/`border` shows those overrides in the rendered diagram — as a plain node AND as an expanded-unit cluster — while units without them keep the type-palette defaults. *(COLOR-01, COLOR-02)*
  2. The global edge style applies to every diagram: setting `properties.edges` (e.g. disabling splines) visibly changes edge routing in C1, C2, C3, and expanded views alike; per-unit `edges` still wins where set; every documented value maps to real behavior (no silent no-op). *(GEDGE-01, GEDGE-02)*
  3. Edge direction and colour have single clear knobs: `rank = "reverse"` alone flips vertical ordering while keeping the arrow direction (works in collapsed and expanded views); `kind = "read" | "write" | "read-write"` colours edges (green / red / distinct blend), an explicit `color` overrides the kind colour, and `kind` works and round-trips canonically in both TOML and C4D (convert/fmt/templates, `${param}` substitution). *(RANK-01, RANK-02, KIND-01, KIND-02, KIND-03)*
  4. Collapsed edges keep kind identity: a collapsed edge's colour derives from constituent kinds (all read → read, all write → write, mixed → read-write), line style follows precedence (any solid → solid, else any dashed → dashed, else dotted), and explicit custom colours suppress kind colouring to the default edge colour. *(AGG-01, AGG-02, AGG-03)*
  5. Diagrams explain themselves, and nothing else changed: every diagram renders a legend in the upper-right (default kind colours + default line styles, then custom author lines), on by default and disable-able via one properties-level setting; models using none of the new features render identically except for the legend block (canonicalDOT goldens re-baselined for the legend only, full suite green); README/skill/example fixtures document the whole surface in both formats; the milestone tags product release **v1.18.0**. *(LEG-01, LEG-02, LEG-03, BC-01, DOC-01, DOC-02, DOC-03, REL-01)*

**Plans**: TBD (plan-phase determines; precedent: v1.11 = 4 plans, v1.12 = 9 plans — expect mid single digits)

**Notes** (from pre-confirmed codebase scan):
- **COLOR fix point:** `buildNode`/`buildCluster` (internal/graph/builder.go) + type palettes (internal/graph/shapes.go); emission via `applyNodeStyle`/`applyClusterStyle` (internal/render/converter.go). `Unit.Color/Style/Border` and `Properties.Color/Style/Border` currently have zero render-side reads.
- **GEDGE fix point:** C2 (`view/scope.go:377`) and C3 (`scope.go:470`) copy only `unit.Edges` — add `Properties.Edges` fallback. `"square"` documented but unimplemented in `configureGraphSettings` (converter.go:161-169) — implement as `ortho` or remove from docs.
- **RANK:** `model.Link.Rank` already parses `RankForward`/`RankReverse` (dead today; only `"equal"` → `constraint=false` works). Implement at `createEdge` (builder.go:573-611) + render (converter.go:483-495) by swapping endpoints + `dir=back`; must survive the 4 view copiers in `view/scope.go` (they already copy Rank).
- **KIND:** new `model.Link.Kind` field — touches the 4 view copiers, `validator/index.go` mirror, C4D grammar (`c4d.peg:458` OptionKey + `go:generate` regen), `c4d/tomodel.go` applyEdgeOption, `c4d/frommodel.go` edgeStmtFromLink, `emit_toml.go` canonical field order, `grammar/reserved.go` fieldKeywords, `testutil/canonsrc`.
- **AGG:** view copiers carry style/colour; merge first-wins per pair in `processOutgoingLinks`/`processIncomingLinks` (builder.go) — kind colour + style precedence computed there.
- **LEGEND:** `graph.Graph.Legend` placeholder exists (internal/graph/graph.go:151-155); render via the top graph-label HTML table (`SetLabelHTML` + `SetLabelLocation(TopLocation)`, converter.go:183-240), legend as a right-aligned column — GraphViz cannot position clusters. Default enabled = new properties field (default true) + custom lines array.
- **BC-01 golden discipline:** canonicalDOT (DI-1) order-insensitive comparisons; legend-on-by-default is the one accepted, user-mandated re-baseline delta.
- **TDD mode on** (config workflow.tdd_mode); docs + skill sync land inside the phase (v1.12 precedent); REL-01 (tag v1.18.0) is the final small item of this phase, not a separate phase.

## Progress

**Execution Order:** Phase 36 (single phase; plans sequenced by plan-phase)

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 28. Reference field | v1.10 | 1/1 | Complete | 2026-08-08 |
| 29. Optional name humanization | v1.10 | 2/2 | Complete | 2026-08-08 |
| 30. Relative-peer resolution | v1.10 | 2/2 | Complete | 2026-08-08 |
| 31. Template expansion | v1.10 | 2/2 | Complete | 2026-08-08 |
| 32. Include directive | v1.10 | 2/2 | Complete | 2026-08-08 |
| 33. Docs sweep + goldens | v1.10 | 4/4 | Complete | 2026-08-08 |
| 34. Label formatting fixes | v1.11 | 4/4 | Complete | 2026-08-10 |
| 35. C4D DSL alternative | v1.12 | 9/9 | Complete | 2026-08-14 |
| 36. Edge Semantics and Legend | v1.13 | 6/6 | Complete | 2026-08-28 |

**Post-milestone (2026-08-28):** user-directed design review shipped outside any phase as v1.19.0–v1.20.0 — legend reworked into a floating framed node outside an invisible content cluster (REQUIREMENTS.md LEG-01..03 re-specified in place), queue units render as SVG pipes (SHAPE-01, quick task [260828-qbx](.planning/quick/260828-qbx-render-queue-units-as-horizontal-pipe-sh/)). Quick tasks are not tracked in the phase table above (GSD quick-mode convention).
