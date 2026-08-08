---
created: 2026-08-08T00:53:13.640Z
title: Document that `type` is omittable (type-inference is underused)
area: docs
resolves_phase: 33
files:
  - internal/parser/parser.go
  - README.md
  - skill/SKILL.md
  - skill/examples/02-nested.toml
---

## Problem

The parser already infers a unit's `type` from its nesting level, so `type` is optional — but production models don't use it. The production model `cyp-auth-infra/saira-20260320.c2.full.toml` spells `type = "container"` / `type = "component"` on nearly every unit, despite the inference rules covering those exact cases. The feature works; authors just don't know it exists or don't trust it. This is a documentation gap, not a code gap.

## Solution

Surface the type-inference rules prominently in user-facing docs. Pure documentation work — no code change.

### Inference rules to document (source of truth: `internal/parser/parser.go`)

From `defaultTypeForParent` (parser.go:250) — `type` defaults by parent:

| Parent unit type | Default child type |
|---|---|
| (root, no parent) | `system` |
| `system` / `systemExternal` | `container` |
| `box` | `system` (same-level C1 grouping) |
| `container` | `component` |
| `containerBox` | `container` |
| `componentBox` | `component` |

From `inferGenericType` (parser.go:276) — the generic `db` and `queue` types auto-promote by level:

| Nesting level | `db` → | `queue` → |
|---|---|---|
| C1 (root) | `db` | `queue` |
| C2 (inside system/containerBox) | `containerDb` | `containerQueue` |
| C3 (inside container/componentBox) | `componentDb` | `componentQueue` |

### Docs to update

1. **`README.md`** — TOML Format / Unit Types section. Add a callout that `type` is optional and a before/after example:
   ```toml
   # Verbose — type spelled out
   [shop.api]
   type = "container"
   name = "API"

   # Equivalent — type inferred from parent system
   [shop.api]
   name = "API"
   ```
2. **`skill/SKILL.md`** — schema reference. Mark `type` as optional in the required-fields table and add the two inference tables above.
3. **`skill/examples/02-nested.toml`** — add an inline comment demonstrating inference (a unit with no `type` line, with a `# type inferred: container` comment).

### Verification

- Confirm the rules in the tables above still match the code before publishing (re-read parser.go:250 and :276).
- Ensure the `box`/`containerBox`/`componentBox` grouping behavior is mentioned, since those are the non-obvious cases.

## Related

Companion to "TOML authoring ergonomic improvements" (`2026-08-08-toml-authoring-ergonomic-improvements.md`). That todo reduces verbosity via parser features (relative peer, optional name); this one reduces verbosity via documentation of an existing parser feature. Both originated from the same "should we replace TOML with a DSL?" analysis.
