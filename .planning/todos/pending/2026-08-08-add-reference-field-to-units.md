---
created: 2026-08-08T00:53:13.640Z
title: Add 'reference' field to C4 units with 📖 marker
area: tooling
resolves_phase: 28
files:
  - internal/model/unit.go
  - internal/render/converter.go
  - internal/parser/parser.go
---

## Problem

C4Drill model units currently have no way to associate an external documentation link (runbook, Confluence page, ADR, etc.) with an element. During a feature comparison between LikeC4 and C4Drill, the **one** LikeC4 feature judged worth adopting was LikeC4's element `link` (a clickable URL hyperlink on a node).

C4Drill already uses `link`/`linkFrom` for *relationships*, so the field must be named differently to avoid collision → **`reference`**.

Scope note: this is the **only** LikeC4 feature being adopted. The following were explicitly considered and **REJECTED** in the same discussion — do not implement them as part of this work:
- custom element kinds (fixed type enum stays)
- tags
- icons (note: current branch is `dev/shapes-no-icons`; icons are contrary to current direction)
- metadata / props
- deployment model
- user-authored views + include/exclude predicates
- custom color palette system
- view-scoped styling
- `summary` field (judged nice-to-have, deferred)

## Solution

Add an optional `reference` field to model units.

### 1. Model layer (`internal/model/unit.go`)
- Add `Reference string \`toml:"reference"\`` to the `Unit` struct.
- Add `"reference"` to the builtin field allowlist in `internal/parser/parser.go` (around line 309, the list is: `type, name, description, technology, color, style, border, edges, width, height, expanded, link, linkFrom`). Without this, the parser treats unknown keys as subunits.

### 2. Rendering (`internal/render/`)
- When a unit has a non-empty `reference`, render a **📖 (open book)** symbol next to/within the node label so the reader can see at a glance which elements have external docs.
- Preferably make the 📖 (or the whole node) a clickable `<a href="reference">` link in the SVG, since SVG natively supports anchors. Confirm whether the existing renderer emits GraphViz HTML-like labels or raw DOT; wire the symbol + link accordingly.
- Precedent in the repo: the collapsed-cluster indicator was recently changed from `[+]` to a 🔍 magnifier symbol (commit `2ac5202`), so emoji-as-indicator is an established pattern here.

### 3. Docs
- Update `README.md` (Unit fields table) and `skill/SKILL.md` (the schema reference) to document the new field. Note that the README's link-syntax section is already known-stale — leave that alone, only add `reference`.

### 4. Examples / fixtures
- Add `reference = "https://..."` to one or two units in `skill/examples/` (e.g. `05-ecommerce.toml`) and/or `cmd/c4drill/testdata/multilevel.toml` so the feature is visible and tested end-to-end.

### Validation rules to consider
- Empty string vs. omitted: treat both as "no reference" (no 📖).
- URL validation: optional — a basic `http(s)://` prefix check is enough; don't over-engineer.
