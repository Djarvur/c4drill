# Phase 2: Auto-generate C2/C3 Diagrams - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-08-06
**Phase:** 2-Auto-generate C2/C3 Diagrams
**Areas discussed:** Box sub-diagrams, Nested expansion depth, File naming, Expansion precedence, Expanded-but-empty, Actors in C2/C3, Deep box parity

**Session note:** Refinement discussion over the existing v1.8 implementation (auto-detect via collectExpandableUnitPaths). Phase 1 decisions D-01/D-02/D-07..D-11 carried forward as constraints.

---

## Box sub-diagrams

| Option | Description | Selected |
|--------|-------------|----------|
| Generate for boxes | Uniform auto-detect; collapsed boxes drill down via explore links | ✓ |
| Exclude boxes | Boxes are pure grouping; contents only visible when expanded in C1 | |

**User's choice:** Generate for boxes

---

## Nested expansion depth

| Option | Description | Selected |
|--------|-------------|----------|
| One level in C1 | Expanded top-level unit → direct subunits as cluster nodes; deeper via explore links | ✓ |
| Recursive clusters | Nested expansion renders recursively; C1 grows toward C2/C3 | |

**User's choice:** One level in C1
**Notes:** Per-unit `expanded` still renders clusters in C2/C3 views.

---

## File naming

| Option | Description | Selected |
|--------|-------------|----------|
| Unit key (current) | Files and URLs derive from the same dotted paths; links never break | ✓ |
| Display name | Pretty names; requires sanitization, breaks links on rename | |

**User's choice:** Unit key

---

## Expansion precedence

| Option | Description | Selected |
|--------|-------------|----------|
| OR — either wins | properties.expanded ∪ per-unit self-reference; v1.7 compat (COMPAT-01) | ✓ |
| properties wins | Single source of truth when properties.expanded present | |

**User's choice:** OR — either wins
**Follow-up:** non-matching properties.expanded entries → silently ignore (recommended) ✓; no stderr warning (silent-per-spec).

---

## Expanded-but-empty

| Option | Description | Selected |
|--------|-------------|----------|
| Plain node | No-subunit expanded units render as normal collapsed nodes | ✓ |
| Empty cluster | Preserves author intent; visually odd | |

**User's choice:** Plain node

---

## Actors in C2/C3

| Option | Description | Selected |
|--------|-------------|----------|
| Keep actors | Linked actors render as external boundary nodes in C2/C3 (v1.0 deferred item) | ✓ |
| Hide actors | Persons only in C1; software-only deeper diagrams | |

**User's choice:** Keep actors

---

## Deep box parity

| Option | Description | Selected |
|--------|-------------|----------|
| Uniform rule | containerBox/componentBox get sub-diagrams like C1 boxes | ✓ |
| C1 boxes only | Deep grouping types never generate files | |

**User's choice:** Uniform rule

---

## Claude's Discretion

None — every area was decided with concrete options.

## Deferred Ideas

None.
