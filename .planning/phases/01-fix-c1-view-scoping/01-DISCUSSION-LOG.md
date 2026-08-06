# Phase 1: Fix C1 View Scoping - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-08-06
**Phase:** 1-Fix C1 View Scoping
**Areas discussed:** Duplicate edge collapse, Edges to expanded subunits, Boundary node naming, Collapsed edge labels, Expanded mode dedup, Source-side resolution, Box grouping rules, minlen effectiveness

**Session note:** The v1.8 implementation was found to already exist (commits `f9dc69a`, `c5091f9`, `8f7e21c`, `b6771b0`; STATE.md COMPLETE; tests passing). User chose "Discuss refinements" — discussion focused on behavior refinements to the existing implementation.

---

## Duplicate edge collapse

| Option | Description | Selected |
|--------|-------------|----------|
| Pair-only, first wins | One edge per (source, target) pair; first in definition order wins | ✓ |
| Keep parallel edges | Current behavior — distinct tech/desc render as parallel edges | |
| Pair-only, merged labels | One edge, aggregated technology label ("HTTP · gRPC") | |
| You decide | | |

**User's choice:** Pair-only, first wins
**Notes:** Follow-up questions: applies everywhere (C1/C2/C3) ✓; all attributes first-wins ✓; thicker edge for collapsed multiplicity ✓; binary scaling (2.0) ✓; all contributing links count ✓.

---

## Edges to expanded subunits

| Option | Description | Selected |
|--------|-------------|----------|
| Deepest visible ancestor | Edge targets the visible subunit when parent expanded | ✓ |
| Always parent (current) | All links resolve to top-level ancestor regardless | |
| You decide | | |

**User's choice:** Deepest visible ancestor
**Notes:** Follow-up: replace parent edge (no redundant edges) ✓.

---

## Boundary node naming

| Option | Description | Selected |
|--------|-------------|----------|
| Skip unresolved | Validator is the gatekeeper; view layer skips unresolvable peers | ✓ |
| Keep full-path fallback | Defensive fallback renders boundary node with dotted path | |
| Last component only | Fallback renders last path component | |
| You decide | | |

**User's choice:** Skip unresolved
**Notes:** Follow-up: remove legacy `addExternalBoundaryNodes` + `addExternalBoundaryNodesRecursive` ✓.

---

## Collapsed edge labels

| Option | Description | Selected |
|--------|-------------|----------|
| First-wins, plain | Label = first sub-link tech/desc; penwidth is only multiplicity signal | ✓ |
| Count suffix | "(+2)" suffix on label | |
| You decide | | |

**User's choice:** First-wins, plain

---

## Expanded mode dedup

| Option | Description | Selected |
|--------|-------------|----------|
| Exempt --expanded | Pair-only collapse in C1/C2/C3; --expanded keeps v1.7 tech+desc dedup (COMPAT-02) | ✓ |
| Apply everywhere | Pair-only also in --expanded; forces COMPAT-02/baseline update | |
| You decide | | |

**User's choice:** Exempt --expanded

---

## Source-side resolution

| Option | Description | Selected |
|--------|-------------|----------|
| Resolve source too | Edges emanate from source's deepest visible ancestor | ✓ |
| Parent-level source | Current behavior — source stays at top-level unit | |
| You decide | | |

**User's choice:** Resolve source too
**Notes:** Follow-up: draw within cluster (links between two visible subunits of the same expanded unit render inside the cluster) ✓.

---

## Box grouping rules

| Option | Description | Selected |
|--------|-------------|----------|
| Same rules | Boxes follow deepest-visible-ancestor rules uniformly, no special-casing | ✓ |
| Box-only edges | Edges to box contents always resolve to the box | |
| You decide | | |

**User's choice:** Same rules

---

## minlen effectiveness

| Option | Description | Selected |
|--------|-------------|----------|
| Original pair only | minlen applies only when both drawn endpoints are the link's original units | ✓ |
| Always inherited | minlen first-wins on resolved/collapsed edges like other attributes | |
| Visible endpoints only | minlen whenever both endpoints are visible units | |
| You decide | | |

**User's choice:** Original pair only
**Notes:** User-specified area ("minlen must be effective if both nodes the edge connects are visible only"). For collapsed edges, surviving first link's minlen applies only if its own endpoints were both original.

---

## Claude's Discretion

None — every area was decided with concrete options.

## Deferred Ideas

None.
