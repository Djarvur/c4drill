# Roadmap: C4Drill

## Milestones

- ✅ **v1.0 Initial Release** — Phases 1-6 (shipped 2026-03-10)
- ✅ **v1.7 Queue Label Fix & Visual Improvements** — Phases 7-18 (shipped 2026-03-29)
- ✅ **v1.8 Proper C1/C2/C3 View Generation** — Phases 19-26 (shipped 2026-08-06) → [archive](milestones/v1.8-ROADMAP.md)
- ✅ **v1.9 C3 Boundary Node Fix** — Phase 27 (shipped 2026-08-06) → [archive](milestones/v1.9-ROADMAP.md)
- ✅ **v1.10 Model Composition** — Phases 28-33 (shipped 2026-08-08) → [archive](milestones/v1.10-ROADMAP.md)
- ✅ **v1.11 Label Formatting Fixes** — Phase 34 (shipped 2026-08-10) → [archive](milestones/v1.11-ROADMAP.md)
- ✅ **v1.12 C4D DSL Alternative** — Phase 35 (shipped 2026-08-17) → [archive](milestones/v1.12-ROADMAP.md)

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

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 28. Reference field | v1.10 | 1/1 | Complete | 2026-08-08 |
| 29. Optional name humanization | v1.10 | 2/2 | Complete | 2026-08-08 |
| 30. Relative-peer resolution | v1.10 | 2/2 | Complete | 2026-08-08 |
| 31. Template expansion | v1.10 | 2/2 | Complete | 2026-08-08 |
| 32. Include directive | v1.10 | 2/2 | Complete | 2026-08-08 |
| 33. Docs sweep + goldens | v1.10 | 4/4 | Complete | 2026-08-08 |
| 34. Label formatting fixes | v1.11 | 2/2 | Complete    | 2026-08-10 |
| 35. C4D DSL alternative | v1.12 | 9/9 | Complete | 2026-08-14 |

<details>
<summary>✅ v1.12 C4D DSL Alternative (Phase 35) — SHIPPED 2026-08-17</summary>

**Goal:** Deliver the C4D format — a `.c4d` brace-block D2-inspired DSL with full TOML feature parity — parseable directly to `*parser.Model` and renderable through the unchanged pipeline, with bidirectional canonical-equivalent converters (`convert to-toml`/`to-c4d`), a gofmt-style comment-preserving formatter (`fmt`) for both formats, nested use (`[[unit.use]]` in TOML, `use` in blocks in C4D) and recursive template-instantiating-template expansion, plus full README/skill/example documentation.

- [x] Phase 35: Add a simple DSL alternative to the TOML diagram definition (9/9 plans) — completed 2026-08-14

**Stats:** 1 phase, 9 plans, 25 tasks. Requirements D-01..D-35 satisfied. Verification: 24/24 truths (3 gap fixes at close). UAT: 12/12. Security: 30/30 threats closed (ASVS 1).

Full details: [milestones/v1.12-ROADMAP.md](milestones/v1.12-ROADMAP.md)

</details>

