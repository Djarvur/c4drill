# Requirements: v1.15 Hierarchy Wrapping and Granular Keys

**Status:** Active
**Milestone:** v1.15 Hierarchy Wrapping and Granular Keys (product release tag: v1.22.0)
**Last updated:** 2026-08-30

User review (2026-08-30) of the shipped v1.14 corrected three orchestrator decisions. (1) The hierarchy invariant was scoped too narrowly: boundary/sibling nodes are drawn WITHOUT their containers in drill-down views — regular nodes, boundary nodes, and expanded elements alike must render inside their full container chains (box, system/container/component) so nodes visibly belong to a hierarchy; ONLY containers may be drawn — no extra nodes beyond those that belong on the scheme. (2) Granular per-aspect formatting keys are wanted alongside the master `--plain` (v1.14 deferred them). (3) A dedicated key must be able to disable labels entirely — labels heavily distort routing and can make a scheme unreadable — for drill-down AND expanded generation.

Source: user review 2026-08-30 (retroactive answers to the yolo-skipped questions). CLI-only keys confirmed correct. Research skipped: corrections to just-shipped behavior; fix points mapped in v1.14 (internal/view/scope.go boundary resolution + createExternalBoundaryNode, internal/graph builder dispatch, render/converter.go emission, cmd/c4drill root.go flags). v1.14 traceability preserved in git history (b794f9f).

---

## v1.15 Requirements

### WRAP — Full hierarchy wrapping for all depicted nodes

- [x] **WRAP-01**: Every depicted node on every generated view (C1, C2/C3 drill-downs, expanded) renders inside its complete chain of ancestor containers — boundary and sibling nodes included; box, system, container and component container chains all render as nested clusters around their contents.
- [x] **WRAP-02**: Boundary resolution keeps its v1.14 semantics for WHAT is depicted (deepest sibling-level container, deep-link true target); the wrapping is additive — ancestor containers of a boundary entry render as nested clusters around it. Entries with no in-model ancestor (fully external) stay top-level — there is nothing to wrap them in.
- [x] **WRAP-03**: Wrapping draws containers only — the depicted node set is unchanged from v1.14 (no extra nodes appear anywhere); locked by test.

### KEY — Granular formatting switches

- [x] **KEY-01**: Individual CLI switches exist, each independently suppressing one author-custom formatting aspect with defaults restored: colours (unit `color`/`border` fills + link `color`), styles (unit/link line and border styles), lengths (link `length`), ranks (link `rank`). Unset switch = exactly current behavior.
- [x] **KEY-02**: `--plain` renders identically to v1.14 — the union of all granular switches (locked by existing plain goldens staying green).
- [x] **KEY-03**: Every switch composes with every generation (C1, all drill-downs, `--expanded`) and format (dot/svg/html). Kind-derived colours and the legend follow the v1.14 semantic-vs-custom boundary unless the colours switch explicitly covers them (planner pins; documented either way).

### LBL — Label suppression key

- [x] **LBL-01**: A CLI key (e.g. `--no-labels`) omits label text from all nodes and edges on the scheme — shapes render minimally with no label content, re-flowing the layout without label-induced routing distortion.
- [x] **LBL-02**: The key applies to every generation — C1, all drill-down views, and `--expanded` — in all formats.
- [x] **LBL-03**: The key composes with `--plain` and the granular switches. Whether the legend is suppressed with labels-off is pinned at planning and documented (default: legend stays, controlled by its own setting — it is metadata, not an element label).

### BC — Backward compatibility

- [x] **BC-01**: Without the new keys, output changes ONLY for the documented WRAP deltas (boundary wrapping — real re-baselining expected this time); KEY/LBL switches are opt-in with zero default-path change; the full suite stays green.

### DOC — Documentation and skills

- [x] **DOC-01**: README.adoc documents boundary wrapping and every new key (per-aspect meaning, composition, legend behavior under labels-off).
- [x] **DOC-02**: skill/SKILL.md and all plugin copies are synced (CI `diff -r` parity).
- [x] **DOC-03**: Skill/example fixtures demonstrate the wrapping and the new keys; render cleanly through the full pipeline.

### REL — Release

- [ ] **REL-01**: Milestone ships as product release **v1.22.0** (git tag; CI release workflow builds artifacts and creates the GitHub release).

---

## Out of Scope

| Feature | Reason |
|---------|--------|
| Properties-level (model-file) keys | CLI-only confirmed by user 2026-08-30 |
| Suppressing the deep-link/expanded depiction semantics from v1.14 | User review confirmed nodes "that should be on the scheme" stay; only wrapping is corrected |
| Manual positioning / layout overrides | Violates the auto-layout design decision (PROJECT.md) |

## Traceability

Filled during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| WRAP-01 | Phase 38 | ✅ 38-01 |
| WRAP-02 | Phase 38 | ✅ 38-01 |
| WRAP-03 | Phase 38 | ✅ 38-01 |
| KEY-01 | Phase 38 | Complete |
| KEY-02 | Phase 38 | Complete |
| KEY-03 | Phase 38 | Complete |
| LBL-01 | Phase 38 | Complete |
| LBL-02 | Phase 38 | Complete |
| LBL-03 | Phase 38 | Complete |
| BC-01 | Phase 38 | Complete |
| DOC-01 | Phase 38 | ✅ 38-05 |
| DOC-02 | Phase 38 | ✅ 38-05 |
| DOC-03 | Phase 38 | ✅ 38-05 |
| REL-01 | Phase 38 | Pending |

**Coverage:**
- v1 requirements: 13 total
- Mapped to phases: 13
- Unmapped: 0 ✓

---
*Requirements defined: 2026-08-30*
