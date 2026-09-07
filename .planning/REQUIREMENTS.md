# Requirements: C4Drill — Milestone v1.18 Bold Subject Boundary

**Defined:** 2026-09-07
**Core Value:** Transform simple TOML architecture descriptions into professional C4 diagrams without manual drawing.
**Source:** User feedback 2026-09-07 — non-expanded diagrams don't make clear which element they depict.

## v1.18 Requirements

Requirements for this milestone. Each maps to roadmap phases.

### Bold Subject Boundary

- [x] **BOLD-01**: On every non-expanded (drill-down) view, the boundary of the element the diagram depicts is drawn with a bold triple-width border — the user can tell at a glance which element the scheme belongs to
- [x] **BOLD-02**: The bold boundary is a semantic navigation aid, not author formatting — it survives `--plain` and `--no-styles`, and no other cluster on the same view is affected
- [ ] **BOLD-03**: Expanded-mode views, node borders, legend, and edge styling are unchanged; golden updates are limited to the subject-boundary delta

## Future Requirements

Deferred to future milestones. Existing backlog, unchanged by this milestone.

- Template multi-output / `for_each` fan-out
- Compact-link shorthand variants beyond baseline
- C4D polish warnings: WR-03 duplicate `properties {}` last-win, WR-04 skill type-inference table drift, WR-05 quoted-label whitespace trim
- Human follow-ups: v1.17 desktop-window smoke (42-HUMAN-UAT.md); #35 JetBrains live-IDE validation; #34 Zed preview when upstream API lands

## Out of Scope

Explicitly excluded. Documented to prevent scope creep.

| Feature | Reason |
|---------|--------|
| Bold boundary on `--expanded` views | Expanded copies show every unit expanded — no single subject to emphasize; user scoped the request to non-expanded schemes |
| Bold boundary on collapsed C1 root | Root view renders units as collapsed nodes, no boundary clusters |
| Thickening node borders or other clusters | Would reduce the contrast the feature exists to create |
| Configurability of the thickness | Fixed triple-width per request; a future `--border-width` flag can extend this |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| BOLD-01 | Phase 43 | Complete |
| BOLD-02 | Phase 43 | Complete |
| BOLD-03 | Phase 43 | Pending |