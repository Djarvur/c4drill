# Phase 32: Include directive (multi-file) - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-08-08
**Phase:** 32-Include directive (multi-file)
**Areas discussed:** UnitOrder splice-in-place vs append (revisited), Diamond-include behavior, Missing-include behavior

---

## UnitOrder splice-in-place vs append

| Option | Description | Selected |
|--------|-------------|----------|
| Append (module-import) | Entry file's units first in authored order; each included file's units append in include-directive order. Simpler; entry-file order stable regardless of [[include]] placement. Research SUMMARY §6 IN-2 recommendation. | ✓ (after revision) |
| Splice-in-place (C #include) | Included file's units insert at the [[include]] directive's position. Splitting one big file into chunks preserves original order. | ✓ initially, then revised |

**User's choice:** Initially splice-in-place; revised to append after reflection.
**Notes:** The append model is simpler to reason about (entry file = "main", includes = additive libraries). Recorded as a deliberate override of the splice intuition. Cross-file subunits still append onto existing parents regardless (D-10).

### Follow-up: cross-file subunits

| Option | Description | Selected |
|--------|-------------|----------|
| Cross-file subunits append onto existing parent | An included file may define subunits under a top-level unit declared in the entry file. Essential for "one file per domain, all under one system." | ✓ |
| Subunits must be co-located with parent | A subunit must be in the same file as its parent (or a file it includes). More restrictive; blocks the natural domain-per-file decomposition. | |

**User's choice:** Cross-file subunits append onto existing parent
**Notes:** Essential for the primary decomposition pattern.

---

## Diamond-include: strict-error vs dedup-identical

| Option | Description | Selected |
|--------|-------------|----------|
| Dedup same-file diamonds; error cross-file collisions | Same file reached via two paths → byte-identical by construction → silently dedup. Two different files defining the same path → hard-error (genuine collision). Refines INC-05. | ✓ |
| Strict dup-error always | Any duplicate unit path from a diamond without once=true is a hard error, even same-file. Forces once=true for every diamond. INC-05 as originally written. | |

**User's choice:** Dedup same-file diamonds; error cross-file collisions
**Notes:** Same-file diamonds are harmless (identical by construction); cross-file collisions are genuine name clashes. Distinction encoded in D-11. `once=true` remains the explicit opt-in visited-set; same-file dedup is automatic and complementary.

---

## Missing-include: hard-error vs optional=true

| Option | Description | Selected |
|--------|-------------|----------|
| Hard-error unconditionally, no optional flag | Any missing include = fatal error naming path + including file (INC-10). Consistent with milestone's hard-error/strictness stance. Simplest; no knob to misuse. | ✓ |
| optional=true escape hatch | Support optional=true to skip missing files silently. Enables local-override/per-developer layering. Con: can mask genuinely-required-but-typo'd includes, causing silent omission. | |

**User's choice:** Hard-error unconditionally, no optional flag
**Notes:** Strictness wins. A missing file is almost always a typo/uncommitted/wrong-cwd; better to fail loudly. Local-override use case rare at current scale.

---

## Claude's Discretion

- Exact `IncludeDirective` struct shape (likely `Path string`, `Once bool`).
- Internal package structure for resolver (`internal/include/Resolve` vs `internal/compose/` vs function in `internal/parser/`).
- `properties` conflict detection scope (full struct vs just name/description; lean conflict-error).

## Deferred Ideas

- **`optional = true`** for missing-file skip — rejected for v1.10 (strictness). Revisit if local-override use case emerges.
- **Include namespacing/prefixing** — anti-feature (Out of Scope).
- **URL includes** — anti-feature (Out of Scope).
- **Splice-in-place ordering** — revised to append (D-09); could revisit if "preserve exact split order" use case emerges.
