---
gsd_state_version: 1.0
milestone: v1.1
milestone_name: AI-Ready
status: completed
last_updated: "2026-03-25T13:27:00.000Z"
last_activity: 2026-03-25 — Completed 23-01-PLAN.md
progress:
  total_phases: 16
  completed_phases: 16
  total_plans: 23
  completed_plans: 22
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-03-24)

**Core value:** Transform simple TOML architecture descriptions into professional C4 diagrams without manual drawing.
**Current focus:** Phase 23 — Deterministic Node and Edge Creation Order

## Current Position

Phase: 23 (Complete)
Plan: 23-01 (Complete)
Status: Phase 23 complete - All map iterations use sorted keys for deterministic output
Last activity: 2026-03-25 — Completed 23-01-PLAN.md

## Accumulated Context

### Decisions

Prior decisions from v1.0-v1.6:

- Use GraphViz native `shape=cylinder` for DB units (works)
- Person labels: 2-column table with emoji (&#x1F464;) instead of SVG icons
- System/Box/Container/Component labels: 3-row single-column table (name, technology, description)
- Icon extraction system removed
- SVG postprocessing removed

**Decisions for v1.7:**

- Queue units: Use HTML label with ASCII art (═╦╩═╦═══) instead of rotated cylinder (GraphViz doesn't support cylinder rotation)

**New decision for Phase 20:**

- Use Helvetica font for all text rendering (nodes, edges, clusters)
- [Phase 20-helvetica-font]: Use Helvetica font for all diagram elements (graph, clusters, edges)
- [Phase 21]: Box labels use same 3-row HTML table format as container/component units — Consistent visual appearance across all unit types that don't have special shapes
- [Phase 21]: All box types (TypeBox, TypeContainerBox, TypeComponentBox) get dashed borders — Visual differentiation from other unit types like system/container/component
- [Phase 21]: C1 boxes cannot contain both external and non-external units (validation rule)
- [Phase 21]: C1 box color determined by contents — grey (#8A8A8A) for external, dark blue (#073B6F) for internal
- [Phase 22]: Link.Length field maps to Edge.MinLen which maps to GraphViz minlen attribute — Users can control edge rank spacing via length attribute in TOML
- [Phase 23]: All map iterations in builder.go use slices.Sorted(maps.Keys()) for deterministic node/edge/cluster ordering — Same input produces identical output order every time

### Pending Todos

None.

### Blockers/Concerns

None.

### Deferred Items

- TestOutputFlag test failure (pre-existing, out of scope) - see deferred-items.md

### Roadmap Evolution

- Phase 21 added: box label has unnecessary curly brackets, box border must be dashed by default
- Phase 21 plan 02 added: validator for mixed external/non-external, color by content
- Phase 22 added: links must have length attribute
- Phase 23 added: deterministic node and edge creation order

## Session Continuity

Last session: 2026-03-25T13:18:00.000Z
Status: Phase 23 complete
Next: Project complete - all planned phases finished
