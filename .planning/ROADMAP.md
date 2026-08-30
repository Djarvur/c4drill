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
- ✅ **v1.14 Nesting Context and Plain Rendering** — Phase 37 (shipped 2026-08-30) — product release tag: v1.21.0
- 🚧 **v1.15 Hierarchy Wrapping and Granular Keys** — Phase 38 (in progress) — product release tag: v1.22.0

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

**Goal:** Deliver the C4D format — a `.c4d` brace-block D2-inspired alternative to TOML with full feature parity — parseable directly to `*parser.Model` and renderable through the unchanged pipeline, with bidirectional canonical-equivalent converters (`convert to-toml`/`to-c4d`), a gofmt-style comment-preserving formatter (`fmt`) for both formats, nested use and recursive template-instantiating-template expansion, plus full README/skill/example documentation.

- [x] Phase 35: C4D DSL alternative (9/9 plans) — completed 2026-08-14

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

<details>
<summary>✅ v1.14 Nesting Context and Plain Rendering (Phase 37) — SHIPPED 2026-08-30</summary>

**Goal:** Non-expanded diagrams preserve the full nesting context — every depicted element renders inside its complete chain of ancestor containers, deep-link targets keep their container context, and expanded units show nested clusters rather than flat lists — and a `--plain` CLI key renders every diagram with author-custom formatting ignored. Shipped as v1.21.0.

- [x] Phase 37: Nesting Context and Plain Rendering (7/7 plans) — completed 2026-08-30

**Stats:** 1 phase, 7 plans. Requirements CTX-01..03, PLAIN-01..04, BC-01, DOC-01..03, REL-01 validated. Scoping note: boundary/sibling entries were kept top-level by an explicit v1.14 decision — reversed by user review 2026-08-30 and corrected in v1.15.

</details>

### 🚧 v1.15 Hierarchy Wrapping and Granular Keys (In Progress)

**Milestone Goal:** Correct v1.14's scoping after user review — every depicted node on any generated view (regular, boundary, expanded) renders inside its complete ancestor-container chain so nothing hangs in the air (drawing containers only, never extra nodes); add granular CLI switches composing with `--plain`; add a dedicated key to disable labels entirely. Ships as product release **v1.22.0**.

- [ ] **Phase 38: Hierarchy Wrapping and Granular Keys** - Boundary/sibling hierarchy wrapping (v1.14 scoping reversal), granular formatting switches composing with `--plain`, `--no-labels` label-suppression key, golden re-baseline, docs, release v1.22.0

## Phase Details

### Phase 37: Nesting Context and Plain Rendering *(SHIPPED 2026-08-30 — v1.21.0)*

**Goal**: Non-expanded diagrams preserve the full nesting picture and a `--plain` CLI key renders canonical default-styled output. 7 plans (37-01..37-07). Requirements: CTX-01..03, PLAIN-01..04, BC-01, DOC-01..03, REL-01. Landed: recursive cluster unfolding (buildNestedCluster), deep-link ancestor chains (Entry.UnfoldChain + ensureDeepLinkChain), cluster drill affordance (Cluster.ExploreURL), `--plain` threading (View.Plain/Graph.Plain + builder guards). Boundary/sibling entries deliberately kept top-level (v1.14 scoping decision — reversed by v1.15).

### Phase 38: Hierarchy Wrapping and Granular Keys

**Goal**: Every depicted node on any generated view renders inside its complete ancestor-container chain — boundary and sibling entries included (reversing the v1.14 scoping decision), with only containers drawn and no extra nodes — and users gain fine-grained CLI control over formatting: individual switches suppress one formatting aspect each (composing with the master `--plain`), and a dedicated key omits all label text so routing distortion from labels disappears. Ships as product release **v1.22.0**.

**Depends on**: Phase 37 (v1.14 — builds directly on buildNestedCluster, Entry.UnfoldChain/ensureDeepLinkChain, createExternalBoundaryNode/addResolvedBoundaryNode/resolveBoundaryByDivergence in internal/view/scope.go, and the --plain flag plumbing precedent in cmd/c4drill/root.go → View.Plain → Graph.Plain)

**Requirements**: WRAP-01, WRAP-02, WRAP-03, KEY-01, KEY-02, KEY-03, LBL-01, LBL-02, LBL-03, BC-01, DOC-01, DOC-02, DOC-03, REL-01

**Success Criteria** (what must be TRUE):

  1. Boundary and sibling nodes are wrapped: on any generated view (C1, C2/C3 drill-downs, expanded), a boundary or sibling entry with an in-model ancestor renders inside its complete chain of ancestor containers as nested clusters (box, system/container/component chains) — nothing depicted hangs in the air; entries with no in-model ancestor (fully external) stay top-level, and the depicted node set is unchanged from v1.14 (containers drawn only, locked by test). *(WRAP-01, WRAP-02, WRAP-03)*
  2. Granular switches each suppress exactly one aspect: individual CLI switches independently restore defaults for colours (unit `color`/`border` fills + link `color`), styles (unit/link line and border styles), lengths (link `length`), and ranks (link `rank`); with no switch set, behavior is exactly v1.14; `--plain` output is byte/canonically identical to v1.14 (union of all switches — existing plain goldens stay green); the semantic-vs-custom boundary for kind-derived colours and the legend is pinned at planning and honoured. *(KEY-01, KEY-02, KEY-03)*
  3. Every switch composes everywhere: each granular switch and the labels key works across every generation (C1, all drill-downs, `--expanded`) and every format (dot/svg/html), and composes with `--plain` and each other. *(KEY-03, LBL-02, LBL-03)*
  4. Labels can be silenced: `--no-labels` renders all nodes and edges with shapes only — no label content — re-flowing layout without label-induced routing distortion, on drill-down AND `--expanded` generation alike; the legend's behavior under labels-off is pinned at planning and documented (default: legend stays — it is metadata, not an element label). *(LBL-01, LBL-03)*
  5. Only the documented deltas, then shipped: without the new keys, output changes ONLY for the documented WRAP boundary-wrapping deltas (real golden re-baselining expected for models with cross-container links; KEY/LBL are opt-in with zero default-path change; full suite green); README.adoc documents wrapping + every key, skill/SKILL.md and all plugin copies are synced (CI `diff -r` parity), example fixtures demonstrate wrapping and the new keys and render cleanly; the milestone tags product release **v1.22.0**. *(BC-01, DOC-01, DOC-02, DOC-03, REL-01)*

**Plans**: 6 plans (38-01..38-06)
Plans:
**Wave 1**

- [x] 38-01-PLAN.md — WRAP: boundary/sibling ancestor wrapping + node-set invariance (WRAP-01..03) ✅ 2026-08-30

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 38-02-PLAN.md — KEY: granular switches --no-colors/--no-styles/--no-length/--no-rank (KEY-01, KEY-02) ✅ 2026-08-30

**Wave 3** *(blocked on Wave 2 completion)*

- [ ] 38-03-PLAN.md — LBL: --no-labels label suppression (LBL-01..03)

**Wave 4** *(blocked on Wave 3 completion)*

- [ ] 38-04-PLAN.md — composition matrix E2E + consolidated golden re-baseline (KEY-03, BC-01)

**Wave 5** *(blocked on Wave 4 completion)*

- [ ] 38-05-PLAN.md — docs, skill 3-copy sync, examples (DOC-01..03)

**Wave 6** *(blocked on Wave 5 completion)*

- [ ] 38-06-PLAN.md — release v1.22.0 (REL-01)

**Notes** (from milestone context — phase research/codebase scan must pin exact gaps):

- **Why one phase:** the last four milestones were each a single phase (shared packages and goldens); WRAP/KEY/LBL all converge on the same emission points (internal/view/scope.go, internal/graph builder, internal/render/converter.go, cmd/c4drill flags) and share canonicalDOT goldens — splitting would force cross-phase golden churn. No hard dependency boundary exists.
- **WRAP fix points:** createExternalBoundaryNode, addResolvedBoundaryNode, resolveBoundaryByDivergence (internal/view/scope.go) — wrap ancestor chains via the v1.14 buildNestedCluster / UnfoldChain machinery; fully external entries (no in-model ancestor) stay top-level.
- **KEY plumbing precedent:** v1.14's `--plain` (cmd/c4drill/root.go → View.Plain → Graph.Plain + builder guards); granular flags thread the same path. v1.14's deferred-items entry for granular flags is superseded — they are now in scope.
- **LBL fix points:** render/converter.go label emission + graph label construction; labels affect routing, so suppression changes layout (expect its own visual deltas under the flag only).
- **Golden discipline:** real re-baselining expected THIS time for WRAP (unlike v1.14's zero-delta outcome); use canonicalDOT (DI-1) via internal/testutil/canonical; document every delta.
- **Planner must pin:** kind-derived colours / legend coverage under the colours switch; legend behavior under labels-off (default: stays).
- **TDD mode on** (config workflow.tdd_mode); docs + skill sync inside the phase; REL-01 (tag v1.22.0) is the final task.

## Progress

**Execution Order:** Phase 38 (single phase; plans sequenced by plan-phase)

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
| 37. Nesting Context and Plain Rendering | v1.14 | 7/7 | Complete | 2026-08-30 |
| 38. Hierarchy Wrapping and Granular Keys | v1.15 | 2/6 | In Progress|  |

**Post-milestone (2026-08-28):** user-directed design review shipped outside any phase as v1.19.0–v1.20.0 — legend reworked into a floating framed node outside an invisible content cluster (REQUIREMENTS.md LEG-01..03 re-specified in place), queue units render as SVG pipes (SHAPE-01, quick task [260828-qbx](.planning/quick/260828-qbx-render-queue-units-as-horizontal-pipe-sh/)). Quick tasks are not tracked in the phase table above (GSD quick-mode convention).
