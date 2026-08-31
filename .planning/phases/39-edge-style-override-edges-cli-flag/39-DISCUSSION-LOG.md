# Phase 39: Edge Style Override (`--edges` CLI flag) - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-08-31
**Phase:** 39-Edge Style Override (`--edges` CLI flag)
**Areas discussed:** Todo fold, Override scope, Flag naming

---

## Todo Fold

| Option | Description | Selected |
|--------|-------------|----------|
| Fold todo | Fold "Add CLI flag to override edge routing style" (score 0.9) into CONTEXT.md as user-validated design input | ✓ |
| Review only | Reference roadmap/requirements instead; don't fold the todo body | |

**User's choice:** Fold (Claude discretion — selection prompt unanswered in autonomous run)
**Notes:** The todo is the milestone's design source, already tagged `resolves_phase: 39` and quoted in ROADMAP.md Phase 39 implementation notes. Not folding would lose the flow-to-extend and file list from the canonical context.

---

## Override Scope

| Option | Description | Selected |
|--------|-------------|----------|
| Invocation-global | `--edges` overrides the resolved `View.Edges` for every view — beats per-unit `edges` AND `[properties] edges` | ✓ |
| Properties-only | `--edges` overrides only `[properties] edges`; per-unit `edges` overrides survive the flag | |

**User's choice:** Invocation-global (Claude discretion — selection prompt unanswered)
**Notes:** Decided on codebase evidence: per-unit edges are live via `cmp.Or(unit.Edges, properties.Edges)` (`internal/view/scope.go:504/:599`); the `--plain`/`--no-*` family is invocation-global; mixed precedence (flag beats properties but loses to unit overrides) would be unpredictable for the motivating variant-rendering use case.

---

## Flag Naming

| Option | Description | Selected |
|--------|-------------|----------|
| `--edges` | Mirrors the TOML key `edges`, like `--expanded` mirrors `expanded` (design todo's choice) | ✓ |
| `--edge-style` | More explicit about taking a style value | |

**User's choice:** `--edges` (Claude discretion — selection prompt unanswered)
**Notes:** Verified no collision against all persistent flag registrations in `cmd/c4drill/root.go:96-114` (`format/-f`, `output/-o`, `expanded`, `plain`, `no-colors`, `no-styles`, `no-length`, `no-rank`, `no-labels`, `label-ratio`). Note the family does use some singular forms (`--no-length`, `--no-rank`), but key-mirroring precedent (`--expanded`) wins for a value-taking flag.

---

## Claude's Discretion

All three decisions above were resolved by Claude: the user answered the milestone-scoping and manager-dispatch prompts but did not respond to the discuss-phase area-selection prompts. Each decision is evidence-backed and veto-friendly — amending `39-CONTEXT.md` before `/gsd:plan-phase 39` is cheap (planning has not started).

## Deferred Ideas

None — discussion stayed within phase scope.
