---
created: 2026-08-30T22:22:16.409Z
title: Add CLI flag to override edge routing style
area: rendering
files:
  - cmd/c4drill/root.go:102
  - internal/graph/graph.go:66
  - internal/graph/builder.go:38
  - internal/render/converter.go:264
---

## Problem

Edge routing style (graphviz `splines`) is configurable ONLY in the TOML model via the `edges` property (`straight|spline|square|ortho`; "square" is a documented alias of "ortho" per GEDGE-02). There is no command-line override, so producing per-invocation variants — the motivating case: render the SAME model as expanded-with-straight and non-expanded-with-spline — requires editing the model file or maintaining two copies.

Flow to extend: TOML `edges` property → `View` → `Graph.EdgeStyle` (internal/graph/builder.go:38-49 copies it with the PLAIN-02 caveat: properties.edges is IGNORED under --plain; also builder.go:411 for expanded) → `cg.SetSplines(...)` in internal/render/converter.go:264-271 (spline→"true", straight→"false", ortho/square→"ortho").

Design constraints from existing patterns:
- Flag threading precedent: granular switches thread onto every generated view (KEY-01/LBL-02; PLAIN-01 threading at cmd/c4drill/root.go:330-386 shows the root + --expanded copy pattern).
- Values must reuse the existing enum (straight|spline|square|ortho) and its converter mapping — no new routing styles.
- Precedence to settle at planning: CLI flag must override model properties; open question — does an explicit CLI --edges survive `--plain` (PLAIN-02 currently suppresses author TOML edges)? Leaning yes: plain ignores AUTHOR formatting, an explicit CLI request is user intent, but it changes documented --plain semantics ("exact union" pin from KEY-02) — decide and pin with a test either way.
- Naming: `--edges <style>` mirrors the TOML key; check no collision with existing flags in root.go.

## Solution

TBD — outline:
1. Add persistent flag (suggested `--edges`) in cmd/c4drill/root.go with enum validation (loud error naming the offending value — same UX as the existing LOUD hard-error precedent for bad values).
2. Thread onto both views (root + expanded) following the PLAIN-01 pattern; override model `edges` when the flag is set.
3. Settle and pin the --plain interaction with a dedicated test.
4. Extend the KEY-03-style switch matrix E2E (flag × C1/expanded × --plain composition) asserting the graphviz `splines` attribute in RAW dot output.
5. Docs: README.adoc usage/flags section + 3 SKILL.md copies (byte-identical, per 37-06 sync precedent).
