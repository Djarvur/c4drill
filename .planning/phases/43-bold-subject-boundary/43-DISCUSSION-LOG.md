# Phase 43: Bold Subject Boundary - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-09-07
**Phase:** 43-Bold Subject Boundary
**Areas discussed:** Emphasis mechanism, Semantics, Scope guard, Verification contract
**Mode:** `--auto` (yolo runtime) — selections auto-chosen and logged

---

## Emphasis mechanism

| Option | Description | Selected |
|--------|-------------|----------|
| penwidth=3 on the subject boundary cluster | Triple width per request; one construction site (buildBoundaryCluster) | ✓ |
| Thicken all cluster borders | Would blur the very contrast the feature creates | |
| Node-level border | The subject with subunits is a cluster, not a node | |

**User's choice:** [auto] Q: "How and where to draw the bold boundary?" → Selected: "penwidth=3 on the subject boundary cluster" (recommended default — user requested triple width; scout confirmed the single construction site).

---

## Semantics

| Option | Description | Selected |
|--------|-------------|----------|
| Semantic navigation aid — survives --plain/--no-styles | Same rationale as kind colours/legend (PLAIN-02) | ✓ |
| Author-style key (suppressed by --plain) | Enters the formatting-key family and dies under --plain | |

**User's choice:** [auto] Q: "Is the bold boundary author formatting or a semantic aid?" → Selected: "Semantic aid" (recommended default — it labels WHAT the diagram is, not how the author styled content).

---

## Scope guard

| Option | Description | Selected |
|--------|-------------|----------|
| Only non-expanded drill-down views | --expanded has no single subject; C1 root has no boundary clusters | ✓ |
| Include --expanded copies | No subject concept there — would be arbitrary emphasis | |

**User's choice:** [auto] Q: "Where does the emphasis apply?" → Selected: "Only non-expanded drill-down views" (recommended default — matches the user's request verbatim).

---

## Verification contract

| Option | Description | Selected |
|--------|-------------|----------|
| Raw-DOT pin + byte-equality for untouched paths + limited golden re-baseline | RED→GREEN per tdd_mode; per-golden diff = penwidth line only | ✓ |
| Visual-only verification | Not assertable in CI | |

**User's choice:** [auto] Q: "How to pin the feature?" → Selected: "Raw-DOT pin + byte-equality for untouched paths + limited golden re-baseline" (recommended default).

---

## Claude's Discretion

- Multi-level fixture choice for the deep-link (C3) assertion.
- Field placement (NodeStyle vs dedicated cluster field) — CONTEXT D-02 recommends NodeStyle.

## Deferred Ideas

- `--border-width <N>` configurability — hard-coded 3 per request; noted for a future flag if users ask.