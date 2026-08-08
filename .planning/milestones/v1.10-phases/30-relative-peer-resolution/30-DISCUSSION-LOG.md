# Phase 30: Relative-peer resolution - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-08-08
**Phase:** 30-Relative-peer resolution
**Areas discussed:** Enclosing-parent definition, Resolution depth, Bare-vs-absolute gate precedence

---

## Enclosing-parent definition (foundamental)

| Option | Description | Selected |
|--------|-------------|----------|
| Immediate parent (a.b for [a.b.c]) | The enclosing parent whose children we search first is the link-host's immediate parent. [a.b.c]'s peer tries a.b.x (c's sibling) first. Natural lexical-scoping reading. | ✓ |
| Different interpretation | (not chosen) | |

**User's choice:** Immediate parent
**Notes:** Foundational to the depth and gate questions; settled first.

---

## Resolution depth: one-level vs walk-up ancestry

| Option | Description | Selected |
|--------|-------------|----------|
| Walk-up ancestry, nearest-first | Sibling-miss walks UP: immediate parent's children → grandparent's → ... → root. Nearest ancestor's matching child wins. Multiple matches at same depth = hard error. Mirrors lexical scoping; useful in deep models. Research SUMMARY §6 RP-2 recommendation. | ✓ |
| Strict one-level (siblings only) | Sibling-miss → immediately absolute-fallback. peer="x" resolves only if a literal sibling x exists. Simpler but much less useful in deep models. | |
| Let me think | Wanted to consider ambiguity/template interaction. | |

**User's choice:** Walk-up ancestry, nearest-first

### Follow-up: walk boundary

| Option | Description | Selected |
|--------|-------------|----------|
| Walk reaches root (top-level = outermost scope) | Top-level units are the outermost scope; peer="messageBus" resolves from any depth if a top-level messageBus exists. Makes shared services reachable by bare name from anywhere. | ✓ |
| Walk stops below root (no top-level via relative) | Bare peer never resolves to top-level via relative; cross-boundary edges need absolute paths. More explicit but verbose. | |

**User's choice:** Walk reaches root

---

## Bare-vs-absolute gate: exact precedence rules

| Option | Description | Selected |
|--------|-------------|----------|
| Unified: '.' → absolute; bare → walk-up to root | (1) '.' in peer → absolute, resolve as-is, miss = existing validator error. (2) bare → walk-up per D-13/D-14/D-15; top-level is final walk-up step (no separate short-circuit). Backward-compat: every existing model resolves identically. No special cases; cross-depth shadowing handled by nearest-first. | ✓ (best judgment — user did not answer) |
| Three-step: '.' then top-level-exact then relative | (1) '.' → absolute; (2) exact top-level match → absolute short-circuit; (3) otherwise relative walk-up. Introduces shadowing ambiguity (sibling named same as top-level unit resolves to top-level, silently wrong). Adds a special case. | |
| Let me think | Wanted to consider shadowing-across-depths more carefully. | |

**User's choice:** (no response — Claude applied best judgment: unified algorithm)
**Notes:** The unified rule is clearly preferable — no special cases, identical backward-compat, avoids the three-step gate's shadowing pitfall. The user did not answer this question; per workflow guidance, proceeded with best judgment rather than retrying.

---

## Claude's Discretion

- Internal package/location for resolver (`internal/peer/Resolve` vs function in `internal/parser/`).
- Error message wording for the two hard-error cases.
- Optional non-blocking warning on cross-depth shadowing (deferred — planner's call if cheap).

## Deferred Ideas

- **Cross-depth shadowing warning** — nearest-first silently picks the nearer; a warning would aid debugging but adds noise. Deferred.
- **ERGO-06 (compact link shorthand)** — Phase 29, at-risk; not Phase 30's concern.
