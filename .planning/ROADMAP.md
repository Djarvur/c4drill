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
- ✅ **v1.15 Hierarchy Wrapping and Granular Keys** — Phase 38 (SHIPPED 2026-08-30) — product release tag: v1.22.0
- 🚧 **v1.16 Edge Style Override** — Phase 39 (in progress)

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

<details>
<summary>✅ v1.15 Hierarchy Wrapping and Granular Keys (Phase 38) — SHIPPED 2026-08-30</summary>

**Goal:** Correct v1.14's scoping after user review — every depicted node on any generated view (regular, boundary, expanded) renders inside its complete ancestor-container chain so nothing hangs in the air (drawing containers only, never extra nodes); add granular CLI switches composing with `--plain`; add a dedicated key to disable labels entirely. Shipped as v1.22.0.

- [x] Phase 38: Hierarchy Wrapping and Granular Keys (6/6 plans) — completed 2026-08-30

**Stats:** 1 phase, 6 plans (phase 37's 7 plans are archived under v1.14 above). Requirements WRAP-01..03, KEY-01..03, LBL-01..03, BC-01, DOC-01..03, REL-01 validated. Outcome note: LBL-01's all-label semantics were narrowed to edge-labels-only post-release by quick task 260831-01u (2026-08-31), which also restored the compact C1 root (flood traced to v1.21.0 CTX-02/03) and made edge identity flag-invariant.

Full details: [milestones/v1.15-ROADMAP.md](milestones/v1.15-ROADMAP.md)

</details>

### 🚧 v1.16 Edge Style Override (In Progress)

**Milestone Goal:** Let users override the edge routing style per invocation via a `--edges <style>` CLI flag — producing variants of the same model (e.g. expanded-with-straight vs non-expanded-with-spline) without editing or duplicating the model file.

- [ ] **Phase 39: Edge Style Override (`--edges` CLI flag)** - Per-invocation edge routing override: persistent flag with loud enum validation, overriding the model `edges` property on every generated view, `--plain` interaction pinned, switch-matrix E2E, backward compat

#### Phase 39: Edge Style Override (`--edges` CLI flag)

**Goal**: Users can override the edge routing style for a whole invocation via `--edges <style>` (`straight|spline|square|ortho`), rendering the same model as per-invocation variants (e.g. expanded-with-straight vs non-expanded-with-spline) without editing or duplicating the model file.
**Depends on**: Nothing within this milestone (builds on the shipped v1.15 baseline — Phase 38, v1.22.0; reuses the PLAIN-01 root + `--expanded` threading pattern at cmd/c4drill/root.go:330-386)
**Requirements**: GEDGE-03, GEDGE-04, GEDGE-05, GEDGE-06, GEDGE-07, GEDGE-08
**Success Criteria** (what must be TRUE):

  1. User can run `c4drill model.toml --edges <style>` with each of `straight|spline|square|ortho` and every generated diagram — C1 root, all drill-down views, and the `--expanded` copy — renders with that routing style, the flag winning over the model's `edges` property (no model edit needed)
  2. User can pass an invalid value (e.g. `--edges diagonal`) and gets a loud error naming the offending value and the allowed enum `straight|spline|square|ortho` — no output produced, no silent fallback
  3. An explicit CLI `--edges <style>` survives `--plain`'s author-format suppression — user intent wins — with the decision pinned by a dedicated test
  4. The switch-matrix E2E covers `--edges` × generation (root / drill-down / `--expanded`) × `--plain`, asserting the graphviz `splines` attribute in RAW dot output for each combination
  5. Without the flag, output is unchanged — all existing canonicalDOT goldens pass untouched

**Plans**: 3 plans
Plans:
**Wave 1**

- [x] 39-01: TDD — `--edges` flag, loud validation, invocation-global override (beats global + per-unit edges), `--plain` survival, flag-off golden invariance (GEDGE-03..06, GEDGE-08)

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 39-02: Switch-matrix E2E — `--edges` × generation × `--plain` asserting `splines=` in RAW dot + golden-safe per-unit fixture (GEDGE-07)

**Wave 3** *(blocked on Wave 2 completion)*

- [ ] 39-03: Docs (README CLI Reference + `--plain` delta + 3 byte-identical SKILL.md copies) + release v1.23.0 (D-07, GEDGE-03/06 docs clause)

**Implementation notes** (from design todo [2026-08-30-add-cli-flag-to-override-edge-routing-style](todos/pending/2026-08-30-add-cli-flag-to-override-edge-routing-style.md)):

- Flow to extend: TOML `edges` → View → `Graph.EdgeStyle` (internal/graph/builder.go:38-49; builder.go:411 for expanded) → `cg.SetSplines` (internal/render/converter.go:264-271; `square`→ortho alias per GEDGE-02 — enum unchanged, no new styles)
- Confirm `--edges` does not collide with existing flags in cmd/c4drill/root.go (naming check from design todo)
- `--plain` interaction is a genuine open decision (PLAIN-02 currently suppresses author TOML edges; "exact union" pin from KEY-02) — settle at plan time, lock with the GEDGE-06 test either way
- Docs (README.adoc usage/flags + 3 SKILL.md copies, byte-identical per 37-06 sync precedent) follow the established repo convention — carried at plan level, not a v1.16 requirement

## Progress

**Execution Order:** Phase 39 (single phase; plans sequenced by plan-phase)

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
| 38. Hierarchy Wrapping and Granular Keys | v1.15 | 6/6 | Complete | 2026-08-30 |
| 39. Edge Style Override (`--edges` flag) | v1.16 | 2/3 | In Progress|  |

**Post-milestone (2026-08-28):** user-directed design review shipped outside any phase as v1.19.0–v1.20.0 — legend reworked into a floating framed node outside an invisible content cluster (REQUIREMENTS.md LEG-01..03 re-specified in place), queue units render as SVG pipes (SHAPE-01, quick task [260828-qbx](.planning/quick/260828-qbx-render-queue-units-as-horizontal-pipe-sh/)). Quick tasks are not tracked in the phase table above (GSD quick-mode convention).

**Post-milestone (2026-08-31):** quick task [260831-01u](.planning/quick/260831-01u-fix-three-rendering-bugs-from-todos-pend/) fixed three post-release rendering bugs TDD-first — compact C1 root restored (flood bisected to v1.21.0 CTX-02/03, not v1.22.0 WRAP), `--no-labels` narrowed to edge labels only, edge identity made flag-invariant via builder-assigned `Edge.Name`; plus repo hardening: all golangci-lint findings resolved, branch protection on master now requires Build/Lint/Test, and the Validate Examples asymmetry (open since 2026-08-14) was fixed.
