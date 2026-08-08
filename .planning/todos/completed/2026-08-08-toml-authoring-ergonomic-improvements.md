---
created: 2026-08-08T00:53:13.640Z
title: TOML authoring ergonomic improvements (relative peer, optional name, compact link)
area: tooling
resolves_phase: [29, 30]
status: completed
resolved_at: 2026-08-08
resolution: "Parts 1 (relative-peer) and 2 (optional-name) shipped in Phases 30 and 29 (v1.10). Part 3 (compact one-liner link, ERGO-06) explicitly deferred per research SUMMARY §3 (anti-feature for v1.10) — moved to Future Requirements. Decisions D-13..D-16 (peer walk-up) and D-01..D-06 (humanize) realized."
files:
  - internal/parser/parser.go
  - internal/model/link.go
  - internal/model/unit.go
---

## Problem

C4Drill's TOML format is correct, but real models carry avoidable verbosity. Analysis was triggered by the question "should we replace TOML with a custom DSL?" — the **decision was NO** (DSL would cost an order of magnitude more than it saves; the data model is flat/tabular and TOML fits; ecosystem/tooling/highlighting comes free with TOML). But the analysis surfaced three specific verbosity pains worth fixing surgically *inside* TOML.

Evidence from the production model `cyp-auth-infra/saira-20260320.c2.full.toml` (499 lines, 62 top-level+dotted units, 56 edges): roughly 40% of the file is mechanical repetition that a DSL would eliminate. The three changes below recover most of that win without writing a parser.

Explicitly **NOT** pursuing: a custom DSL, a `.c4drill` grammar, a lexer/parser/AST. These three items are the TOML-native alternative.

## Solution

Three independent, composable improvements. Listed in priority order (highest payoff first).

### 1. Relative peer resolution (highest value)

Currently `peer` must be an **absolute dotted path** even when the link is declared inside the parent block of a sibling. Example today (`saira` line ~233):

```toml
[linuxSystem.localIDP.sessionAPI]
link = [
  {peer = "linuxSystem.localIDP.sessionManager"},   # absolute, re-spells the whole path
]
```

Goal: `peer` resolves **relative to the enclosing parent block**, falling back to absolute resolution if no relative match:

```toml
[linuxSystem.localIDP.sessionAPI]
link = [
  {peer = "sessionManager"},   # → linuxSystem.localIDP.sessionManager
]
```

Resolution algorithm:
- Try `currentParent + "." + peer` → if that unit exists, use it.
- Else try walking up: `grandparent + "." + peer`, etc.
- Else treat `peer` as absolute (current behavior) — full backward compat.
- Document precedence (relative-first vs absolute-first). Recommend relative-first with absolute fallback so short names Just Work and long names still resolve.

Resolution must happen **after the full model is parsed** (peers can be forward references), so this lives in a post-parse pass, not in the TOML unmarshaling. Touch points: `internal/parser/parser.go` (post-`Parse` at `:47`), validator already runs post-parse so the resolved path must be in place before `internal/validator/rules.go` checks peer existence (`:14`).

Ambiguity risk: if two siblings in different branches share a short name, relative-first picks the nearest. Acceptable — same as any scoping rule. Document it.

### 2. Optional `name` field

Currently `name` is effectively required on every unit, producing pure boilerplate:

```toml
[linuxSystem.localIDP.sessionManager]
name = "Session Manager"          # <- adds nothing the identifier doesn't
```

Goal: if `name` is omitted, derive a humanized display name from the **last path segment** of the identifier:
- `linuxSystem` → "Linux System" (camelCase split + space)
- `localIDP` → "Local IDP"
- `sessionManager` → "Session Manager"
- `grpcAPIs` → "gRPC APIs" (best-effort acronym preservation, or just "Grpc APIs" — pick one and document)

Implementation: humanize function in parser or model layer; called at unit-load time when `Name == ""`. Single source of truth for the humanize rule. Touch point: `internal/model/unit.go` (`Unit` struct `:41`, `Name` field) + parser finalization.

Backward compat: explicit `name =` always wins. No model changes required for existing files.

Decide: does humanization run on the *segment only* (`localIDP` from `linuxSystem.localIDP`) or the *full path*? Segment-only is recommended — the full path is structural, the segment is the "name."

### 3. Compact one-liner link syntax (lowest value, hardest to fit TOML)

The current canonical form is the array-of-tables:

```toml
[[unit.link]]
peer = "sessionManager"
technology = "gRPC"
description = "Calls"
```

…and the inline-array form (`link = [{peer="x", technology="y"}]`) already exists but is cramped on one line for multi-attribute edges. Goal: make the common case (peer + optional technology + optional description) writable compactly.

TOML grammar is rigid, so this is the hardest of the three — options:
- (a) Better document the inline form and provide a formatter/snippet in the skill doc; no code change.
- (b) Allow `link` to accept a **string shorthand** for peer-only edges: `link = ["sessionManager", "cache"]` expands to two peer-only edges. Requires parser-side coercion (string → Link with only Peer set).
- (c) Combine with #1: `link = [{peer="sessionManager", technology="gRPC"}]` becomes short once peers are relative.

Recommend doing #1 first and re-measuring — once peers are short, the inline form may be compact enough that #3 needs no code change. Treat #3 as optional / defer until #1+#2 are in.

## Validation / docs

- All three are **additive and backward-compatible**: existing models parse unchanged.
- Update `README.md` (TOML Format section) and `skill/SKILL.md` (schema reference) for each accepted change.
- Add fixtures under `skill/examples/` demonstrating the new shorthand forms.
- Re-measure saira line count after #1+#2 to validate the verbosity win.

## Sequencing

Ship as independent features, not one big change:
1. Relative peer (#1) — biggest payoff, self-contained post-parse change.
2. Optional name (#2) — tiny, independent.
3. Compact link (#3) — defer; re-evaluate after #1.
