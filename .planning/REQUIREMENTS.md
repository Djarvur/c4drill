# Requirements: v1.14 Nesting Context and Plain Rendering

**Status:** Active
**Milestone:** v1.14 Nesting Context and Plain Rendering (product release tag: v1.21.0)
**Last updated:** 2026-08-30

Two user-reported improvements define this milestone. First, non-expanded diagrams lose nesting context: a depicted element can appear without the containers it lives in, breaking the mental map as the reader moves between diagram levels. Second, there is no way to render a model ignoring author-custom formatting — for canonical default-styled output (diffing structure, consistent publishing).

Source: user request 2026-08-30. Clarifying questions were auto-skipped (yolo mode); scope below is the orchestrator's best-judgment reading, to be confirmed against the codebase in phase research:

- On non-expanded views, where deep units get depicted today (C1 expanded clusters and their visible subunits, boundary/link resolution in `internal/view/scope.go`), intermediate containers may not render — the phase scan must pin the exact gaps.
- Formatting inputs to ignore under `--plain`: unit `color`/`style`/`border`, link `color`/`style`/`length`/`rank`, `properties.edges`, custom label formatting (`labels.go` HTML rectangles, `labelPosition`).

---

## v1.14 Requirements

### CTX — Nesting context on non-expanded views

- [x] **CTX-01**: Every depicted element on a non-expanded generated diagram renders inside its complete chain of ancestor containers — all intermediate containers render as nested containers around it, so no element ever appears outside its hierarchy.
- [x] **CTX-02**: A link whose target is a deeply nested unit keeps the target's context: the target renders within its container chain and the edge terminates at the target inside those containers, instead of silently collapsing to an anonymous top-level ancestor.
- [x] **CTX-03**: Expanded units render depicted nested elements through their intermediate containers (nested clusters, not flat lists), so the nesting picture on a non-expanded scheme matches the drill-down views — end-to-end recognizability across all diagram levels.

### PLAIN — Formatting-ignoring generation key

- [x] **PLAIN-01**: `c4drill --plain` renders every generated diagram with explicit unit formatting ignored: `color`/`style`/`border` on any unit (including expanded-unit clusters) fall back to the type-palette defaults.
- [x] **PLAIN-02**: `--plain` ignores explicit edge formatting: link `color` and `style` fall back to defaults; `length` and `rank` are ignored (default spacing, forward ranking); global `properties.edges` is ignored. Kind-derived edge colours and the legend are kept — they derive from semantic `kind`, not custom formatting.
- [x] **PLAIN-03**: `--plain` simplifies label formatting: labels render as plain text (no custom HTML-rectangle formatting); label text content (name, technology, description) is preserved.
- [x] **PLAIN-04**: `--plain` applies uniformly to every generated output file — the context diagram and all drill-down views, in all formats (svg/html/dot).

### BC — Backward compatibility

- [x] **BC-01**: Without `--plain`, models that do not exercise the new nesting-context scenarios render unchanged; canonicalDOT goldens are re-baselined only for documented CTX deltas; the full test suite stays green.

### DOC — Documentation and skills

- [ ] **DOC-01**: README.adoc documents the nesting-context behavior and the `--plain` key — what is ignored and what deliberately stays.
- [ ] **DOC-02**: skill/SKILL.md and all plugin copies are synced with the same surface.
- [ ] **DOC-03**: Skill/example fixtures demonstrate both features and render cleanly through the full pipeline.

### REL — Release

- [ ] **REL-01**: Milestone ships as product release **v1.21.0** (git tag; CI release workflow builds artifacts and creates the GitHub release).

---

## Out of Scope

| Feature | Reason |
|---------|--------|
| Granular per-aspect ignore flags (`--no-colors`, `--no-ranks`, …) | One `--plain` key covers the stated need; add granular flags when actually asked for |
| Properties-level ignore keys embedded in the model file | CLI-only keeps the model format lean; the dual-format (TOML + C4D grammar) surface is disproportionate |
| Stripping kind-derived colours or the legend in plain mode | Kind is semantic data-flow information, not custom formatting |
| Restructuring collapsed (non-depicted) subtrees | CTX shows container chains only for DEPICTED elements; collapsed units stay collapsed |
| Manual positioning / layout overrides | Violates the auto-layout design decision (PROJECT.md) |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| CTX-01 | Phase 37 | Complete |
| CTX-02 | Phase 37 | Complete |
| CTX-03 | Phase 37 | Complete |
| PLAIN-01 | Phase 37 | Complete |
| PLAIN-02 | Phase 37 | Complete |
| PLAIN-03 | Phase 37 | Complete |
| PLAIN-04 | Phase 37 | Complete |
| BC-01 | Phase 37 | Complete |
| DOC-01 | Phase 37 | Pending |
| DOC-02 | Phase 37 | Pending |
| DOC-03 | Phase 37 | Pending |
| REL-01 | Phase 37 | Pending |

**Coverage:**
- v1 requirements: 11 total
- Mapped to phases: 11
- Unmapped: 0

---
*Requirements defined: 2026-08-30*
