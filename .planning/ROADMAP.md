# Roadmap: C4Drill

## Milestones

- ✅ **v1.0 Initial Release** — Phases 1-6 (shipped 2026-03-10)
- ✅ **v1.7 Queue Label Fix & Visual Improvements** — Phases 7-18 (shipped 2026-03-29)
- ✅ **v1.8 Proper C1/C2/C3 View Generation** — Phases 19-26 (shipped 2026-08-06) → [archive](milestones/v1.8-ROADMAP.md)
- ✅ **v1.9 C3 Boundary Node Fix** — Phase 27 (shipped 2026-08-06) → [archive](milestones/v1.9-ROADMAP.md)
- ✅ **v1.10 Model Composition** — Phases 28-33 (shipped 2026-08-08) → [archive](milestones/v1.10-ROADMAP.md)
- ✅ **v1.11 Label Formatting Fixes** — Phase 34 (shipped 2026-08-10) → [archive](milestones/v1.11-ROADMAP.md)
- ✅ **v1.12 C4D DSL Alternative** — Phase 35 (shipped 2026-08-17) → [archive](milestones/v1.12-ROADMAP.md)
- ✅ **v1.13 Edge Semantics and Legend** — Phase 36 (shipped 2026-08-28) → [archive](milestones/v1.13-ROADMAP.md) — product release tag: v1.18.0
- 🚧 **v1.14 Nesting Context and Plain Rendering** — Phase 37 (in progress) — product release tag: v1.21.0

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

<details>
<summary>✅ v1.13 Edge Semantics and Legend (Phase 36) — SHIPPED 2026-08-28</summary>

**Goal:** Make edge/colour semantics trustworthy and expressive — unit styling renders, global edge style everywhere, `rank = "reverse"`, edge kinds with collapse aggregation, default-on legend. Shipped as v1.18.0, followed by post-milestone design review v1.19.0–v1.20.0 (floating legend node, queue pipes).

- [x] Phase 36: Edge Semantics and Legend (6/6 plans) — completed 2026-08-28

**Stats:** 1 phase, 6 plans. All 20 requirements validated (COLOR-01..02, GEDGE-01..02, RANK-01..02, KIND-01..03, AGG-01..03, LEG-01..03, BC-01, DOC-01..03, REL-01). Release: v1.18.0.

Full details: [milestones/v1.13-ROADMAP.md](milestones/v1.13-ROADMAP.md)

</details>

### 🚧 v1.14 Nesting Context and Plain Rendering (In Progress)

**Milestone Goal:** Non-expanded diagrams preserve the full nesting context — every depicted element renders inside its complete chain of ancestor containers, deep-link targets keep their container context, and expanded units show nested clusters rather than flat lists, so the nesting picture is recognizable across all generated views — and a `--plain` CLI generation key renders every diagram with author-custom formatting ignored, for canonical default-styled output. Ships as product release **v1.21.0**.

- [ ] **Phase 37: Nesting Context and Plain Rendering** - Full ancestor chains on all non-expanded views, `--plain` formatting-ignoring generation key, backward-compat goldens, docs + release v1.21.0

## Phase Details

### Phase 37: Nesting Context and Plain Rendering

**Goal**: Non-expanded diagrams preserve the full nesting picture — every depicted element renders inside its complete chain of ancestor containers, links to deeply nested units keep the target's container context, and expanded units render nested clusters instead of flat lists so the nesting picture matches the drill-down views — and a `--plain` CLI key renders every generated diagram with author-custom formatting ignored, for canonical default-styled output. Ships as product release **v1.21.0**.

**Depends on**: Phase 36 (v1.13 — plain mode must interact correctly with the v1.13 styling surface: unit colour/style/border are suppressed under `--plain`, while kind-derived edge colours and the legend stay)

**Requirements**: CTX-01, CTX-02, CTX-03, PLAIN-01, PLAIN-02, PLAIN-03, PLAIN-04, BC-01, DOC-01, DOC-02, DOC-03, REL-01

**Success Criteria** (what must be TRUE):

  1. Full ancestor chains on non-expanded views: every depicted element on any non-expanded generated diagram renders inside its complete chain of ancestor containers — intermediate containers appear as nested containers around it, and no depicted element ever appears outside its hierarchy. *(CTX-01)*
  2. Deep links keep target context: a link whose target is a deeply nested unit renders that target within its container chain, with the edge terminating at the target inside those containers — never silently collapsing to an anonymous top-level ancestor. *(CTX-02)*
  3. Expanded units render nested clusters, matching drill-downs: depicted nested elements appear through their intermediate containers (nested clusters, not flat lists), so the nesting picture on a non-expanded scheme matches the drill-down views and is recognizable end-to-end across all diagram levels. *(CTX-03)*
  4. `--plain` uniformly strips custom formatting, keeps semantics: `c4drill --plain` renders every generated file — the context diagram and all drill-down views, in svg/html/dot — with unit `color`/`style`/`border` fallen back to type-palette defaults, link `color`/`style` at defaults, `length`/`rank` ignored (default spacing, forward ranking), `properties.edges` ignored, and labels as plain text preserving name/technology/description content — while kind-derived edge colours and the legend remain. *(PLAIN-01, PLAIN-02, PLAIN-03, PLAIN-04)*
  5. Nothing else changed, documented, released: without `--plain`, models that do not exercise the new nesting-context scenarios render unchanged (canonicalDOT goldens re-baselined only for documented CTX deltas; full test suite green); README.adoc documents both features (what is ignored, what deliberately stays), skill/SKILL.md and all plugin copies are synced, example fixtures demonstrate both features and render cleanly through the full pipeline; the milestone tags product release **v1.21.0**. *(BC-01, DOC-01, DOC-02, DOC-03, REL-01)*

**Plans**: 7 plans in 6 waves
Plans:
**Wave 1**

- [ ] 37-01-PLAN.md — CTX-03: recursive clusters in buildCluster (expanded units render nested clusters, not flat lists)

**Wave 2** *(blocked on Wave 1 completion)*

- [ ] 37-02-PLAN.md — CTX-02: deep-link ancestor-chain entries in C1/C2/C3 view resolution (true-target edges)

**Wave 3** *(blocked on Wave 2 completion)*

- [ ] 37-03-PLAN.md — PLAIN-01/02: --plain flag threading (View.Plain → Graph.Plain) + builder-level formatting suppression

**Wave 4** *(blocked on Wave 3 completion)*

- [ ] 37-04-PLAN.md — PLAIN-03/04: plain-text labels in the renderer + E2E --plain goldens and uniformity across all outputs

**Wave 5** *(blocked on Wave 4 completion)*

- [ ] 37-05-PLAN.md — CTX-01 invariant + ONE consolidated golden re-baseline (BC-01) + visual checkpoint
- [ ] 37-06-PLAN.md — DOC-01..03: README, skill sync (3 byte-identical copies), example fixtures

**Wave 6** *(blocked on Wave 5 completion)*

- [ ] 37-07-PLAN.md — REL-01: tag and ship product release v1.21.0

**Notes** (from milestone context — phase research/codebase scan must pin exact gaps):

- **Why one phase:** CTX (view semantics: `internal/view/scope.go` — GenerateC1View, visible subunits, boundary resolution; `internal/graph` cluster building) and PLAIN (CLI flag in `cmd/c4drill`, applied through the render pipeline) are largely disjoint in code but share canonicalDOT goldens and the same emission points; splitting would force cross-phase golden churn (single-phase precedent: v1.11, v1.12, v1.13 — STATE.md decision carry-forward).
- **CTX scan targets:** the two known mechanisms losing nesting context today — C1 deep-link target collapse (target resolved to an anonymous top-level ancestor) and one-level expanded rendering (depicted nested elements as flat lists inside expanded clusters). Fix points expected in `internal/view/scope.go` + `internal/graph` cluster building.
- **PLAIN suppression points:** unit styling at `applyNodeStyle`/`applyClusterStyle` (internal/render/converter.go), edge `color`/`style`/`length`/`rank` at edge emission, `properties.edges` in graph settings, label formatting in `internal/graph/labels.go` (plain-text label path; content preserved).
- **Plain-mode boundary:** kind-derived edge colours and the legend STAY — they are semantic, not custom formatting; collapsed (non-depicted) subtrees are NOT restructured — container chains render only for DEPICTED elements (Out of Scope).
- **BC-01 delta discipline (differs from v1.13):** `--plain` is opt-in, so there is NO universal output delta — only CTX's documented nesting deltas may re-baseline goldens; flat models must render unchanged; comparisons use canonicalDOT (DI-1, order-insensitive) via `internal/testutil/canonical`.
- **TDD mode on** (config workflow.tdd_mode); docs + skill sync land inside the phase (v1.12/v1.13 precedent); REL-01 (tag v1.21.0) is the phase's final task, not a separate phase.

## Progress

**Execution Order:** Phase 37 (single phase; plans sequenced by plan-phase)

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
| 37. Nesting Context and Plain Rendering | v1.14 | 0/7 | Not started | - |

**Post-milestone (2026-08-28):** user-directed design review shipped outside any phase as v1.19.0–v1.20.0 — legend reworked into a floating framed node outside an invisible content cluster (REQUIREMENTS.md LEG-01..03 re-specified in place), queue units render as SVG pipes (SHAPE-01, quick task [260828-qbx](.planning/quick/260828-qbx-render-queue-units-as-horizontal-pipe-sh/)). Quick tasks are not tracked in the phase table above (GSD quick-mode convention).
