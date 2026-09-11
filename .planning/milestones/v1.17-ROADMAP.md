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
- ✅ **v1.16 Edge Style Override** — Phase 39 (SHIPPED 2026-08-31) → [archive](milestones/v1.16-ROADMAP.md) — product release tag: v1.23.0
- 🚧 **v1.17 Issue Sweep** — Phases 40-42 (IN PROGRESS, started 2026-09-03)

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

<details>
<summary>✅ v1.16 Edge Style Override (Phase 39) — SHIPPED 2026-08-31</summary>

**Goal:** Let users override the edge routing style per invocation via a `--edges <style>` CLI flag — producing variants of the same model (e.g. expanded-with-straight vs non-expanded-with-spline) without editing or duplicating the model file. Shipped as v1.23.0.

- [x] Phase 39: Edge Style Override (`--edges` CLI flag) (3/3 plans) — completed 2026-08-31

**Stats:** 1 phase, 3 plans, 8 tasks, 29 commits (+485/−2 Go, 5 files; repo at ~50.3k LOC). All 6 requirements validated (GEDGE-03..08); verification 5/5 success criteria + 11/11 plan truths; UAT 7/7 (zero issues). Key outcomes: invocation-global override beats global AND per-unit `edges` (dedicated `View.EdgesOverride` carrier applied post-PLAIN-02 in both builders); explicit `--edges` survives `--plain` — a deliberate, documented delta to the KEY-02 exact-union contract, pinned by `TestEdgesSurvivesPlain`; `square`→ortho alias confirmed through the flag path; zero golden churn flag-off.

Full details: [milestones/v1.16-ROADMAP.md](milestones/v1.16-ROADMAP.md)

</details>

<details>
<summary>🚧 v1.17 Issue Sweep (Phases 40-42) — IN PROGRESS (started 2026-09-03)</summary>

**Goal:** Close the three actionable open GitHub issues — byte-reproducible SVG output (#42), a render-free `check` command (#41), and the Wails desktop binding fix (#38). Three independent issue families, one phase each; phases can execute in any order.

- [x] **Phase 40: Deterministic SVG Output** - Rendering the same model twice yields byte-identical output; edge ids derive from model content, not map iteration (issue #42) (completed 2026-09-03)
- [x] **Phase 41: Check Command** - `c4drill check <file>` validates a model without rendering, with render-identical errors and exit codes (issue #41) (completed 2026-09-03)
- [x] **Phase 42: Desktop GUI Binding Fix** - Frontend RPC aligned to the Wails-generated namespace of the actually bound struct, `main.desktop` (issue #38) (completed 2026-09-03)

### Phase 40: Deterministic SVG Output

**Goal**: Rendering the same model twice produces byte-identical output — generated edge ids derive deterministically from model content, so diagrams can be diffed, committed, and cached reliably.
**Depends on**: Nothing (fully independent)
**Requirements**: REPRO-01, REPRO-02, REPRO-03
**Success Criteria** (what must be TRUE):

  1. Rendering the same model file twice in separate invocations produces byte-identical SVG output, including the generated `id="edge<N>"` group ids — pinned by a byte-equality regression test built from the issue #42 reproducer (TDD: RED on current code, GREEN after the fix)
  2. Edge ids are assigned in a deterministic order derived from model content (sorted/insertion order), never Go map iteration order — repeated renders never permute or renumber edge ids
  3. The byte-equality-across-repeated-runs guarantee is asserted for every supported output format: `dot`, `svg`, and `html`
  4. Existing canonicalDOT goldens (DI-1/COMPAT-02/REF-05) and the full test suite stay green — the fix changes only id-assignment ordering, not rendered semantics

**Plans**: 2 plans
Plans:
**Wave 1**

- [x] 40-01-PLAN.md — Deterministic mirror synthesis (D-02) + issue #42 byte-equality regression pin (D-05) — TDD, wave 1

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 40-02-PLAN.md — Defense-in-depth name-sorted edge slice (D-03) + no-semantic-change gate (D-04/D-06) — wave 2

### Phase 41: Check Command

**Goal**: Users can validate a model without rendering — a fast, render-free `check` command that fits CI and edit loops and reports exactly what the render path would report.
**Depends on**: Nothing (fully independent)
**Requirements**: CHECK-01, CHECK-02, CHECK-03, CHECK-04
**Success Criteria** (what must be TRUE):

  1. `c4drill check <file>` validates a model and exits without writing any output files or requiring an output directory
  2. `check` exits 0 when the model is valid and non-zero when invalid, reporting the same validation errors the render path reports (e.g. VAL orphan-unit rules) — pinned TDD-first with valid and invalid fixtures
  3. `check` runs the same pipeline front-half as render — includes resolved, templates expanded, relative peers resolved, then validation — proven by a composed multi-file fixture that checks exactly as it renders
  4. README documents the `check` command alongside the existing CLI surface (usage, exit codes, no-output behavior)

**Plans**: 2 plans

Plans:
**Wave 1**

- [x] 41-01-PLAN.md — check subcommand via shared render front-half (TDD: RED/GREEN/REFACTOR)

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 41-02-PLAN.md — README.adoc + skill/SKILL.md document check

### Phase 42: Desktop GUI Binding Fix

**Goal**: Desktop-window mode works again — the frontend RPC layer calls the Wails-generated namespace that matches the actually bound Go struct (`main.desktop`).
**Depends on**: Nothing (fully independent)
**Requirements**: GUI-01, GUI-02
**Success Criteria** (what must be TRUE):

  1. `internal/gui/frontend/src/rpc.ts` calls `window.go.main.desktop.Dispatch` — the namespace Wails generates for the bound struct (cmd/c4drill-gui `main.desktop`, Wails `Bind`) — and no references to the phantom `window.go.main.App` / `go.backend.App` namespaces remain
  2. Desktop-window RPC works again: every method rpc.ts invokes exists on the generated `main.desktop` binding, restoring the desktop transport broken since the #31/#37 restructure (verified via frontend build + binding-shape assertions; TDD where testable)
  3. The `--serve` HTTP fallback path is unchanged and its existing e2e suite stays green

**Plans**: 1 plan
**UI hint**: yes
Plans:

- [x] 42-01-PLAN.md — Align frontend resolver to `window.go.main.desktop.Dispatch` (TDD: RED binding-shape tests → GREEN resolver fix → D-04 structural + regression gates)

</details>

## Progress

**Execution Order:** Phases 40-42 (independent issue families — default order 40 → 41 → 42; plans sequenced by plan-phase)

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
| 39. Edge Style Override (`--edges` flag) | v1.16 | 3/3 | Complete    | 2026-08-31 |
| 40. Deterministic SVG Output | v1.17 | 2/2 | Complete    | 2026-09-03 |
| 41. Check Command | v1.17 | 2/2 | Complete    | 2026-09-03 |
| 42. Desktop GUI Binding Fix | v1.17 | 1/1 | Complete    | 2026-09-03 |

**Post-milestone (2026-08-28):** user-directed design review shipped outside any phase as v1.19.0–v1.20.0 — legend reworked into a floating framed node outside an invisible content cluster (REQUIREMENTS.md LEG-01..03 re-specified in place), queue units render as SVG pipes (SHAPE-01, quick task [260828-qbx](.planning/quick/260828-qbx-render-queue-units-as-horizontal-pipe-sh/)). Quick tasks are not tracked in the phase table above (GSD quick-mode convention).

**Post-milestone (2026-08-31):** quick task [260831-01u](.planning/quick/260831-01u-fix-three-rendering-bugs-from-todos-pend/) fixed three post-release rendering bugs TDD-first — compact C1 root restored (flood bisected to v1.21.0 CTX-02/03, not v1.22.0 WRAP), `--no-labels` narrowed to edge labels only, edge identity made flag-invariant via builder-assigned `Edge.Name`; plus repo hardening: all golangci-lint findings resolved, branch protection on master now requires Build/Lint/Test, and the Validate Examples asymmetry (open since 2026-08-14) was fixed.
