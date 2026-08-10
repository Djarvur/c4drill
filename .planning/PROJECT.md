# C4Drill

## What This Is

A Go CLI tool for generating C4 architecture diagrams from TOML definitions. Users describe their system architecture in a structured TOML file (or a set of files composed via `[[include]]`), and C4Drill renders it as GraphViz DOT, SVG, or HTML diagrams. Supports C1 (Context), C2 (Containers), and C3 (Components) layers with collapsed/expanded views and interactive explore links. The HTML format (`-f html`) wraps the SVG in a self-contained document with a JS shim so clickable navigation works in Safari/WebKit (which silently ignores SVG `<a>` hyperlinks).

As of v1.10, models are composable: a diagram can be assembled from multiple TOML files (`[[include]]` with cycle detection and `once` dedup), units can be defined once as parametrized `[template.*]` and instantiated N times via `[[use]]`, `peer` references can be relative (resolving against the enclosing parent's ancestry), `name` is optional (humanized from the identifier), and any unit can carry a `reference` URL rendered as a clickable 📖 marker.

## Core Value

Transform simple TOML architecture descriptions into professional C4 diagrams without manual drawing.

## Requirements

### Validated

- ✓ Parse TOML input file with C4 model definition — v1.0
- ✓ Validate model integrity (references, type rules, subunit constraints) — v1.0
- ✓ Generate GraphViz DOT output for C1-C3 layers — v1.0
- ✓ Render SVG output via go-graphviz — v1.0
- ✓ Support collapsed/expanded unit views — v1.0
- ✓ Generate explore links for drilling into nested structures — v1.0
- ✓ Support all unit types — v1.0
- ✓ Apply styling: colors, borders, edge routing styles — v1.0
- ✓ Single CLI command interface — v1.0

### Validated in Phase 3 (v1.8)

- ✓ Backward compatibility: TOML files without `properties.expanded` render collapsed C1; `--expanded` output identical to v1.7 at DOT level (public fixture + canonicalDOT golden enforcement)

### Validated in Phase 2 (v1.8)

- ✓ C2/C3 files: auto-generated for units with subunits (uniform rule incl. boxes; unit-key file naming)
- ✓ properties.expanded controls default expansion (OR semantics; expanded-but-empty renders as plain node)

### Validated in Phase 1 (v1.8)

- ✓ C1 view: only top-level units, edges resolved to visible ancestors (pair-only edge collapse, deepest-visible-ancestor resolution both sides, binary penwidth for collapsed edges)

### Validated in Phase 28 (v1.10)

- ✓ `reference` field (📖): optional per-unit external-docs URL renders a clickable marker via GraphViz native `URL` attribute; external reference wins the single URL slot over drill-down; HTML shim routes external http(s)// to a new tab and no-ops non-http(s) schemes (XSS hardening); no-reference models render byte-identical to v1.9 (canonical-DOT golden)

### Validated in Phase 29 (v1.10)

- ✓ Optional `name` humanization (ERGO-03/04/05): when a unit omits `name`, the display name is derived from the last path segment via `model.Humanize` — a dumb camelCase split with no acronym preservation (`gRPC` → "Grpc", `IDPToken` → "Idp Token", but trailing pure-upper runs like `localIDP` → "Local IDP" are preserved). Explicit `name =` always wins; existing fixtures parse byte-identically. Zero new deps (stdlib only). Parse-time hook; Phase 31's XC-04 relocates the call to a post-expansion pass.

### Validated in Phase 30 (v1.10)

- ✓ Relative-peer resolution (ERGO-01/02): a bare `peer` value resolves against the enclosing parent's children (the host's siblings), walking up ancestry nearest-first to root; absolute peers (containing `.`) are untouched; a miss at root is a hard error. Ships as a pure pre-parse pass (`internal/peer.Resolve`) between Parse and Validate; absolute-only models parse identically (backward-compat hard contract).

### Validated in Phase 31 (v1.10)

- ✓ Unit templates (TMPL-01..10, XC-03, XC-04): `[template.<name>]` declares a parametrized unit + subunit subtree with named params; `[[use]]` instantiates it N times supplying ALL params (no defaults — missing any is a hard error). `internal/template.Expand` runs between ParseFile and peer.Resolve (pipeline: Parse → Expand → peer.Resolve → Validate), deep-copying each instantiation via hand-rolled recursive `Unit.Clone()` that preserves the unexported `Link.Mirror` (HS-1 — the validator mutates LinksFrom in place). `${param}` substitutes into every string field (Name, Description, Technology, Reference, Color, Link fields) via `strings.NewReplacer`; duplicate-path and residual-token are hard errors. Forward references work. BC-1 parser prerequisite (captureDefinitionOrder skip + rawMap extraction) landed in Plan 01. Three-instantiation regression test (disjoint LinksFrom post-validate + idempotent re-expand) is the HS-1 gate.

### Validated in Phase 32 (v1.10)

- ✓ Include directive (INC-01..10, XC-02): `[[include]]` array-of-tables pulls in other TOML files relative to the including file's directory (INC-02); transitive, cycle-detected (fatal), flat-merge (no namespacing); `once=true` visited-set dedup (INC-06); cross-file subunits (D-10 — included file re-declares parent, subunits attach); root-file-wins properties (INC-08); 3-arg `include.Resolve(entry, entryDir, entryFile)` runs as Stage 1a so templates in included files are visible to `[[use]]` (XC-02). Parser `IncludeDirective` type + `Model.Includes` field + BC-1 captureDefinitionOrder skip landed in Plan 01; resolver + merge + pipeline wiring in Plan 02.

### Validated in Phase 33 (v1.10) — milestone integration + docs

- ✓ Reusable canonicalDOT helper (D-18): `internal/testutil/canonical.Canonical(t, dot)` extracted from `internal/graph/builder_test.go` into a non-`_test.go` package, importable from any `_test.go` file in the repo. DI-1 contract preserved (parse DOT, strip layout geometry bb/pos/lp/lheight/lwidth/height/width, sort statements + attributes recursively, order-insensitive). 4 WR-01/WR-02 regression tests moved with it; 2 existing goldens (COMPAT-02, REF-05) switched to the import.
- ✓ Documentation gap-fill (DOC-01, DOC-02, D-19): README.md + skill/SKILL.md gain the omittable-`type` inference tables, Templates/Multi-File Composition/Relative Peer sections, and a Pipeline Ordering note. Phase 28/29 docs untouched (fill-gaps-only).
- ✓ Example fixtures (DOC-03, D-17): `skill/examples/06-templates.toml`, `07-relative-peer.toml`, `08-include/` (3 files), `09-composed/` (4 files incl. hand-expanded single-file equivalent). All 5 runnable fixtures render cleanly through the full v1.10 pipeline.
- ✓ End-to-end composition proofs (XC-01, XC-05, D-20): `TestXC05_ComposedEquivSingleFile` proves composed multi-file ≡ hand-expanded single-file (canonicalDOT, 3872 bytes identical); `TestXC01_PipelineOrdering` is a behavioral guard that the pipeline order include → template.Expand → peer.Resolve is load-bearing (covers XC-02 + XC-03). Both use `canonical.Canonical` (DI-1), never byte-exact `require.Equal`.

## Current Milestone: v1.11 Label Formatting Fixes

**Goal:** Generated diagram labels render with proper word wrapping and aspect-ratio sizing.

**Target features:**

- **Edge labels: aspect-ratio sizing** — edge labels are formatted like unit labels: wrapped text in a rectangle with the configured aspect ratio (`LabelRatio`), invisible borders. Currently `buildEdgeLabel` (labels.go) emits a plain `\n`-joined string with no wrapping and no ratio sizing.
- **Word-boundary-only line breaking** — lines break only at word boundaries; no mid-word splits. Currently `wrapText` (wrap.go) falls back to char-level breaking (`splitLongWord`) for long words — remove that fallback; authors may reword instead.

**Key context:**
- Unit labels already use HTML `<table border="0">` labels with `maxChars` derived from `LabelRatio` — edge labels should reuse this machinery.
- Backward compatibility matters: existing diagrams must not regress (goldens via canonicalDOT, DI-1).

## Shipped

### v1.10 Model Composition (Shipped: 2026-08-08)

**Goal:** Expand C4Drill's authoring model from a single static TOML file into a composable, parametrized, multi-file format — while preserving backward compatibility and the auto-generated-view philosophy.

**Target features:** include directive (multi-file composition), unit templates (define once, instantiate with params), relative `peer` resolution, optional `name` humanization, `reference` field (📖), omittable-`type` docs.

**Key accomplishments:** pipeline `include → template.Expand → peer.Resolve → validate → render`; 6 phases (28-33), 13 plans, 35 tasks, 119 files changed (+17,703/−626); all 39 requirements validated; canonicalDOT (DI-1) golden enforcement.

### v1.8 Proper C1/C2/C3 View Generation

**Goal:** Generate correct per-level diagrams — C1 shows only top-level units, C2/C3 files are created for units with subunits

**Target features:**

- C1 diagram: only top-level units with edges resolved to C1 ancestors
- C2 diagrams: auto-generated for each system/box with subunits
- C3 diagrams: auto-generated for each container with subunits
- Edge resolution: links to deeply nested targets resolve to the nearest visible ancestor
- properties.expanded controls which top-level units are expanded by default

**Phase 1 complete (2026-08-06):** C1 view scoping refined — pair-only duplicate-edge collapse with binary penwidth, deepest-visible-ancestor resolution on both source and target sides, within-cluster edges for expanded units, minlen gated to original-pair edges, legacy boundary-node code removed. Verified 9/9 must-haves.

**Phase 2 complete (2026-08-06):** C2/C3 auto-generation confirmed and locked — uniform auto-detect (boxes included), unit-key file naming, one expansion level in C1, OR expansion precedence with silent ignore, expanded-but-empty units render as plain nodes (C1 + C2/C3 branches), actors as boundary nodes in deeper views. Verified 6/6 must-haves.

**Phase 3 complete (2026-08-06) — v1.8 milestone COMPLETE:** Backward compatibility locked — sanitized public fixture (`testdata/multilevel.toml`) + CLI-generated DOT golden baseline with order-insensitive `canonicalDOT` enforcement; COMPAT-01 (collapsed C1 without properties.expanded) and COMPAT-02 (`--expanded` identical to v1.7 at DOT level) regression-tested; `--expanded` ignores properties.expanded (v1.7 contract). Verified 12/12 must-haves. All 9 v1.8 requirements shipped.

### Shipped in v1.7

- ✓ Queue units render with ASCII art graphic (═╦╩═╦═══)
- ✓ Helvetica font for all diagram elements
- ✓ Box labels with dashed borders, validation for mixed content, color by content type
- ✓ Link length attribute for edge spacing control
- ✓ Deterministic node/edge ordering
- ✓ Thicker edges (penwidth 2.0) for visual prominence
- ✓ TOML definition order preservation for nodes and edges

### Shipped in v1.6

- ✓ Remove SVG icon extraction system entirely
- ✓ Remove SVG postprocessing
- ✓ DB units → native cylinder shape
- ✓ Person units → 2-column table with 👤 emoji
- ✓ System/Box/Container/Component → 3-row table

### Post-v1.8 Fixes (2026-08-06, unscoped — from post-UAT testing)

- ✓ **Safari SVG link fix**: new `-f html` output format. Safari/WebKit silently ignores SVG `<a>` navigation; the HTML format inlines the SVG and injects a JS shim that restores clickable drill-down navigation. SVG and DOT formats unchanged (default stays svg).
- ✓ **Navigation redesign**: dropped the redundant back-link (breadcrumb-only nav); breadcrumb items now show pretty unit Names (not raw path keys); root context always present in C2/C3 breadcrumbs (navigate to C1 from any level); title visually distinct from nav (14pt vs 10pt gray); breadcrumb items tightly packed (separator merged into item cells to avoid GraphViz column-stretching gaps).

### Out of Scope

- **C4 Layer 4 (Code)** — Class/function level diagrams not needed
- **Go library/module** — Pure CLI tool, no library interface
- **Manual positioning** — Rely on GraphViz auto-layout
- **JSON error output** — CLI errors sufficient
- **Live editing/watch mode** — Single-shot rendering
- **Multiple commands** — Single command does everything

## Context

**Shipped v1.9** — C3 boundary node fix: cross-container links in C3 diagrams now resolve to the sibling container (e.g., RBAC) instead of the parent system (Main System). ~18,700 LOC Go.

**Shipped v1.8** — Proper C1/C2/C3 view generation with auto-detected sub-diagrams, backward-compatible expanded mode, Safari-compatible HTML output, and breadcrumb-only navigation with pretty names.

**Shipped v1.0** with 9,624 LOC Go across 48 files.

C4 model is a lean approach to software architecture documentation created by Simon Brown. It uses four levels of abstraction:

- **C1 (Context)**: System context showing users and external systems
- **C2 (Containers)**: Deployable units within a system (apps, databases, etc.)
- **C3 (Components)**: Logical components within a container
- **C4 (Code)**: Class/function level (out of scope)

The tool uses nested TOML objects where each level contains strictly typed subunits. Systems and boxes can contain subunits; other types cannot. Links define relationships between units with styling options.

## Constraints

- **Tech Stack**: Go 1.26.1
- **Input Format**: TOML — single file
- **Output**: GraphViz DOT, SVG, or HTML via go-graphviz (HTML = SVG inlined in a wrapper with a JS nav shim for Safari compatibility)
- **Diagram Scope**: C1-C3 layers only

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Single CLI command | Simplicity for documentation workflow | ✓ Good — users run `c4drill file.toml` |
| Auto-layout only | Let GraphViz handle positioning | ✓ Good — clean diagrams without manual effort |
| TOML input | Human-readable, supports nested structures | ✓ Good — intuitive authoring |
| go-graphviz library | Native Go, no external graphviz binary needed | ✓ Good — simple deployment |
| Structured post-parse composition (v1.10) | Templates + include + relative-peer run as passes on `*parser.Model` between Parse and Validate, keeping validator/view/render unchanged | ✓ Good — 4 new packages (`internal/{include,peer,template,testutil/canonical}`), zero changes downstream |
| Hand-rolled `Unit.Clone()` (v1.10) | Reflection/JSON/gob deep-copy silently drops the unexported `Link.Mirror` field, which the validator mutates in place | ✓ Good — HS-1 regression test (3 instantiations → disjoint `LinksFrom`) passes |
| Hard-error-everywhere stance (v1.10) | Strictness over convenience: missing include, missing param, duplicate path, unresolved peer all error loudly | ✓ Good — caught real issues early; no silent corruption |
| No LikeC4 feature adoption except `reference` (v1.10) | Custom kinds, tags, icons, metadata, deployment model, user-authored views all fight C4Drill's auto-generation philosophy | ✓ Good — kept the format lean; `reference` was the only clear win |

## TOML Schema (Reference)

### Root Level

```toml
[properties]
name = "Project Name"
description = "Project description"
color = "transparent"        # optional, default transparent
style = "none"               # optional: none, solid, dotted, dashed
border = "transparent"       # optional, default transparent
edges = "straight"           # optional: straight, spline, square
expanded = ["system1", "box1"]  # optional, default empty

[section_name]  # Context-level unit
type = "system"              # optional, default system
name = "Display Name"
description = "Description"
color = "transparent"
style = "none"
border = "transparent"
link = { target = { reverse = false, equal = false, color = "black", style = "solid" } }
linkFrom = { source = { ... } }

# For system and box types only:
edges = "straight"           # optional, inherits from parent
expanded = ["subunit1"]      # optional

[section_name.subunit]       # Container-level unit (inside system)
# Same attributes as context-level units
```

### Unit Types

- `person`, `personExternal` — Actors using the system
- `system`, `systemExternal` — Software systems
- `db`, `dbExternal` — Databases
- `queue`, `queueExternal` — Message queues
- `box` — Grouping container

### Link Object

```toml
link = { "target_unit" = { reverse = false, equal = false, color = "black", style = "solid" } }
```

## Validation Rules

1. Referenced units must be defined
2. Units with subunits cannot be referenced by links
3. Units with subunits cannot have their own links
4. Subunits only allowed for `system` and `box` types

## Rendering Behavior

- **Collapsed**: Single record shape with "explore" link to drill-down file
- **Expanded**: Cluster with subunits rendered inside
- **File structure**: `{basename}.{format}` for context, `{basename}/` directory for expanded units (format = svg, html, or dot)
- **Shapes**: Person, DB, Queue, System each have distinct record shapes

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd:complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-08-10 — v1.11 Label Formatting Fixes milestone started*
