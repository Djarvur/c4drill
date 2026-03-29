# Phase 15: Edge Coloring - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-03-20
**Phase:** 15-the-edge-must-be-the-same-color-as-the-source-unit
**Areas discussed:** Color source, Label coloring, Explicit color

---

## Color Source

| Option | Description | Selected |
|--------|-------------|----------|
| Source border color | Edge takes border color of unit it comes FROM | ✓ |
| Source font color | Edge takes font color of source unit (same as border currently) | |
| Explicit link.color only | Use existing link.color field, user must specify | |

**User's choice:** Source border color (Recommended)
**Notes:** Natural "flow" visualization - edge inherits visual identity from its source.

---

## Label Coloring

| Option | Description | Selected |
|--------|-------------|----------|
| Default black | Labels in black for readability | |
| Match edge color | Labels match edge color | ✓ |
| Muted version | Labels use muted gray version of edge color | |

**User's choice:** Match edge color
**Notes:** Creates visual consistency - both line and text share the same color.

---

## Explicit Color Override

| Option | Description | Selected |
|--------|-------------|----------|
| Override when set | If TOML has link.color, use it; else source border | ✓ |
| Always source color | Always use source border, deprecate color field | |
| Ignore explicit | TOML color field ignored entirely | |

**User's choice:** Override when set (Recommended)
**Notes:** Fallback chain: explicit color → source unit border color

---

## Claude's Discretion

- Exact location to add Color field to Edge struct
- How to look up source unit's border color during edge creation
- Test coverage strategy

## Deferred Ideas

None — discussion stayed within phase scope.
