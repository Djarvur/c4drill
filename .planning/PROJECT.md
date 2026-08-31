# C4Drill

## What This Is

A Go CLI tool for generating C4 architecture diagrams from TOML definitions. Users describe their system architecture in a structured TOML file (or a set of files composed via `[[include]]`), and C4Drill renders it as GraphViz DOT, SVG, or HTML diagrams. Supports C1 (Context), C2 (Containers), and C3 (Components) layers with collapsed/expanded views and interactive explore links. The HTML format (`-f html`) wraps the SVG in a self-contained document with a JS shim so clickable navigation works in Safari/WebKit (which silently ignores SVG `<a>` hyperlinks).

As of v1.10, models are composable: a diagram can be assembled from multiple TOML files (`[[include]]` with cycle detection and `once` dedup), units can be defined once as parametrized `[template.*]` and instantiated N times via `[[use]]`, `peer` references can be relative (resolving against the enclosing parent's ancestry), `name` is optional (humanized from the identifier), and any unit can carry a `reference` URL rendered as a clickable 📖 marker.

As of v1.15, rendering is CLI-controllable: every depicted node renders inside its complete ancestor-container chain (boundary and sibling entries included — containers only, no extra nodes), and formatting is suppressible per aspect via granular switches (`--no-colors`, `--no-styles`, `--no-lengths`, `--no-ranks`) plus `--plain` (all of them) and `--no-labels` (edge labels only — node/cluster labels and the legend stay). The compact C1 root invariant (root shows only the visible neighborhood, never the whole model) is pinned by fixture + golden.

As of v1.16, edge routing is per-invocation overridable: `--edges <style>` (`straight|spline|square|ortho`) beats both the global `[properties] edges` and per-unit `edges` on every generated view, and an explicit flag survives `--plain` (user intent beats author-format suppression — a documented delta to the exact-union contract).

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

### Validated in Phase 34 (v1.11) — label formatting fixes

- ✓ LABEL-01 (edge labels as wrapped rectangles): `buildEdgeLabel` (labels.go) now emits a borderless HTML-table rectangle (`<table border="0" cellpadding="0" cellspacing="0">`) with the `[Technology]` row and the Description wrapped below via `wrapAndEscape`; width derived from `LabelRatio` via `labelMaxCharsNoIcon` with a 2-row floor (D-03); rectangle always emitted, tech-only and description-only edges included (D-04). `createEdge` emits via `e.SetLabelHTML` (graphviz-13 HTML-ness preserved; empty labels keep `SetLabel("")` for golden parity).
- ✓ LABEL-02 (word-boundary-only breaking): `wrapText` over-budget branch emits the whole word unsplit on its own line; `splitLongWord` deleted (0 references); no character-level fallback anywhere (D-05).
- ✓ COMPAT-01 (no regression): unit labels byte-identical absent over-budget words; all canonicalDOT goldens (COMPAT-02, REF-05, DI-1) pass unchanged; `go test ./...` green (12/12 packages).

### Validated in Phases 37–38 (v1.14–v1.15) — hierarchy wrapping & formatting keys

- ✓ Nesting context (CTX-01..03): recursive cluster unfolding (`buildNestedCluster`), deep-link ancestor chains (`Entry.UnfoldChain` + `ensureDeepLinkChain`), cluster drill affordance (`Cluster.ExploreURL`); `--plain` renders canonical default-styled output (PLAIN-01..04) — v1.21.0
- ✓ Full hierarchy wrapping (WRAP-01..03): boundary and sibling entries render inside their complete ancestor-container chains on every view; containers only, depicted node set unchanged (locked by test); fully external entries stay top-level — v1.22.0
- ✓ Granular CLI switches (KEY-01..03): `--no-colors`/`--no-styles`/`--no-lengths`/`--no-ranks` each suppress one formatting aspect, compose with `--plain` (which stays the exact union) and with each other, across every generation and format; switch matrix locked E2E — v1.22.0
- ✓ Label suppression (LBL-01..03): `--no-labels` on every generation/format; legend stays (metadata, not an element label). Outcome note: narrowed post-release to edge-labels-only by quick task 260831-01u (2026-08-31) — node/cluster labels survive
- ✓ Post-release correctness (260831-01u): compact C1 root restored (flood bisected to v1.21.0 CTX-02/03 whole-subtree unfolding; pinned by `deepcross` fixture + golden invariant); edge identity made flag-invariant via builder-assigned `Edge.Name` (label-derived find-or-create had silently merged parallel edges under `--no-labels`)
- ✓ BC-01: without the new keys, default output changes only for the documented WRAP deltas; docs (README.adoc + 3 SKILL.md copies) synced with CI `diff -r` parity (DOC-01..03)

### Validated in Phase 39 (v1.16) — edge style override

- ✓ `--edges` CLI override (GEDGE-03..05): invocation-global routing-style override accepting the unchanged enum (`straight|spline|square|ortho`, `square` = ortho alias per GEDGE-02); beats BOTH the global `properties.edges` value AND per-unit `edges` overrides on every generated view (C1 root, all drill-downs, `--expanded` copy) via a dedicated `View.EdgesOverride` carrier applied after the PLAIN-02 zeroing; invalid values fail loudly naming the value and the enum before any output (GEDGE-04)
- ✓ `--plain` composition delta (GEDGE-06, D-05): an explicit `--edges` is user intent and SURVIVES `--plain`'s author-format suppression — pinned by `TestEdgesSurvivesPlain` and stated explicitly in the README `--plain` docs (amendment to the KEY-02 exact-union contract); `--plain` with no flag still suppresses author edges
- ✓ Switch-matrix E2E (GEDGE-07): `--edges` × generation (root / drill-down / `--expanded`) × `--plain` asserted via the graphviz `splines` attribute in RAW dot (~86 cells, `TestEdgesComposition` over the golden-free `edges_override.toml` fixture carrying both precedence layers)
- ✓ Backward compat (GEDGE-08): without the flag, all existing canonicalDOT goldens pass untouched — zero re-baselining; scope.go resolution and converter mapping unchanged

## Previous Milestone: v1.16 Edge Style Override — COMPLETE (2026-08-31)

**Goal:** Let users override the edge routing style per invocation via a `--edges <style>` CLI flag — producing variants of the same model (e.g. expanded-with-straight vs non-expanded-with-spline) without editing or duplicating the model file.

**Target features:**

- **`--edges` CLI flag** — persistent flag accepting the existing enum (`straight|spline|square|ortho`), loud error on bad values (same UX as existing LOUD hard-error precedent); no new routing styles.
- **Override semantics** — the flag overrides the model's `edges` property on every generated view (root + `--expanded` copy), threaded per the PLAIN-01 pattern.
- **`--plain` interaction pinned** — decide and lock with a test whether an explicit CLI `--edges` survives `--plain`'s author-format suppression ("exact union" contract from KEY-02).

**Key context:**
- Source: feature request captured 2026-08-30 as todo (`todos/completed/2026-08-30-add-cli-flag-to-override-edge-routing-style.md`) with design sketch and file list; design decisions D-01..D-07 locked in phase 39 CONTEXT.md (override scope invocation-global; `--edges` naming).
- Flow to extend: TOML `edges` → `View` → `Graph.EdgeStyle` → `cg.SetSplines`; threading precedent KEY-01/LBL-02 + PLAIN-01 root + `--expanded` copy at `cmd/c4drill/root.go:330-386`.
- `square` is implemented as the documented ortho alias (GEDGE-02) — enum unchanged.
- Naming check: `--edges` must not collide with existing flags in `root.go`.
- Extend the KEY-03-style switch matrix E2E asserting graphviz `splines` in RAW dot output.

**Status:** ✅ COMPLETE (2026-08-31) — Phase 39 shipped as product release v1.23.0; 1 phase, 3 plans, all 6 requirements (GEDGE-03..08) validated. Full record: [MILESTONES.md](MILESTONES.md).

## Previous Milestone: v1.15 Hierarchy Wrapping and Granular Keys — COMPLETE (2026-08-30)

**Goal:** Correct v1.14's scoping after user review — every depicted node on any generated view (regular, boundary, expanded) renders inside its complete ancestor-container chain so nothing hangs in the air (drawing containers only, never extra nodes); add granular CLI switches on top of `--plain`; add a dedicated key to disable labels entirely (labels distort routing and can make diagrams unreadable), honored by drill-down AND expanded generation.

**Target features:**

- **Full hierarchy wrapping** — boundary and sibling nodes too are wrapped in their ancestor container chains (box, system/container/component) in all views; only containers are drawn, no extra nodes beyond those that belong on the scheme. This REVERSES the v1.14 scoping decision that kept sibling/external boundary nodes top-level.
- **Granular formatting keys** — individual CLI switches (colours, styles, labels, lengths, ranks) composing with the master `--plain`.
- **Label suppression key** — a key that omits labels on the scheme entirely (nodes/edges render without label text), for drill-down and `--expanded` generation alike.

**Key context:**
- Source: user review 2026-08-30 of the v1.14 orchestrator decisions (the questions auto-skipped in yolo mode were answered retroactively). CLI-only keys confirmed correct.
- The v1.14 deferred-items entry for granular flags is superseded: granular switches are now in scope.
- Boundary wrapping will change golden output for models with cross-container links — expect real re-baselining this time.

**Status:** ✅ COMPLETE (2026-08-30) — phases 37+38 shipped as product releases v1.21.0 and v1.22.0; post-release quick task 260831-01u (2026-08-31) fixed three rendering bugs and hardened CI. Full record: [MILESTONES.md](MILESTONES.md).

## Current Focus

Planning the next milestone — nothing active. Remaining candidate backlog:
- Template multi-output / `for_each` fan-out (Future, REQUIREMENTS archive)
- Compact-link shorthand variants beyond baseline (Future, REQUIREMENTS archive)
- C4D polish warnings: WR-03 duplicate `properties {}` last-win, WR-04 skill type-inference table drift, WR-05 quoted-label whitespace trim
- Debug: docs-drift around the orphan rule (VAL-01) testdata; stale `knowledge-base` debug note
- Human UAT follow-ups from 260831-01u: re-render the reporter's real model (compact root), eyeball re-baselined goldens

## Previous Milestone: v1.14 Nesting Context and Plain Rendering — COMPLETE (2026-08-30)

**Goal:** Make edge/colour semantics trustworthy and expressive — fix silently-dropped custom unit colours, make the global edge style apply to every generated diagram, give edge direction/ranking a single clear knob (`rank = "reverse"`), introduce edge kinds (`read`/`write`/`read-write`) with kind-derived colours that survive collapse aggregation, and add a default-on upper-right legend showing the colour semantics plus author-defined lines.

**Target features:**

- **Unit colour fix** — `Unit.Color`/`Style`/`Border` (parsed since v1.0) are dead at render; units must actually render with author-specified styling (nodes and clusters), falling back to the type palette when unset.
- **Global edge style everywhere** — `properties.edges` must be respected by ALL diagram generations (C1, C2, C3, expanded); C2/C3 currently read only the drilled-into unit's own `edges` field and silently ignore the global setting.
- **Convenient rank reversal** — a single `rank = "reverse"` link option replaces today's unclear `"<-"` + `arrow = "reverse"` dance; `rank = "forward"` (default) and `"equal"` (existing constraint=false) complete the set.
- **Edge kinds** — `kind = "read" | "write" | "read-write"` on links with kind-derived colours (read=green, write=red, read-write=a blend), overridable by explicit `color`.
- **Collapsed-edge kind semantics** — when edges collapse to a visible ancestor, colour derives from the constituent kinds; line style follows precedence (any solid → solid, else any dashed → dashed, else dotted); explicit custom colours suppress kind colouring (default edge colour).
- **Legend** — global setting (default enabled) rendering a legend in the upper-right of every diagram: default colour explanations + author-defined custom lines.

**Status:** Complete (2026-08-28) — Phase 36 delivered all 6 plans; 20/20 requirements; full suite green at close; release v1.18.0 tagged.

**Key context:**
- Root cause confirmed by codebase scan: `buildNode`/`buildCluster` style exclusively via `GetStyleForType` palettes; `Unit.Color/Style/Border` and `Properties.Color/Style/Border` have zero render-side reads. Fix point: `internal/graph/builder.go` + `internal/graph/shapes.go`, emission via `internal/render/converter.go` (`applyNodeStyle`/`applyClusterStyle`).
- `rank = "forward"/"reverse"` already parse and round-trip but are consumed nowhere — only `rank = "equal"` → `constraint=false` works.
- `properties.edges` reaches C1/expanded via `view.View.Edges`; C2 (`scope.go:377`) / C3 (`scope.go:470`) copy only `unit.Edges` — no global fallback. `"square"` is documented but unimplemented in `configureGraphSettings` (spline/straight/ortho only).
- `model.Link` gains `Kind`; touches the 4 view copiers, validator mirror, C4D grammar (`OptionKey` rule + regen), `applyEdgeOption`, emitters (TOML + C4D), `canonsrc`.
- Legend: `graph.Graph.Legend` placeholder struct exists; render via the top graph-label HTML table (right-aligned legend column) — GraphViz has no cluster positioning.
- Release tag for this milestone: **v1.18.0** (product tags v1.13.0–v1.17.0 already exist; GSD milestone numbering is internal).

## Current State (2026-08-31)

**Shipped:** v1.16 Edge Style Override — 1 phase (39), 3 plans, 8 tasks, product release v1.23.0. Invocation-global `--edges` routing override (beats global + per-unit edges, survives `--plain`), switch-matrix E2E (~86 cells), zero golden churn. Verification 5/5, UAT 7/7. ~50.3k LOC Go, all tests green, CI at 0 lint issues.

**Previously:** v1.15 Hierarchy Wrapping and Granular Keys — 2 phases (37, 38), 13 plans, product releases v1.21.0 + v1.22.0; post-release quick task 260831-01u restored the compact C1 root, narrowed `--no-labels` to edge labels, made edge identity flag-invariant.

## Next Milestone Goals

*v1.16 shipped; nothing active. Candidate backlog for later:*
- Template multi-output / `for_each` fan-out (Future, REQUIREMENTS archive)
- Compact-link shorthand variants beyond baseline (Future, REQUIREMENTS archive)
- C4D polish warnings: WR-03 duplicate `properties {}` last-win, WR-04 skill type-inference table drift, WR-05 quoted-label whitespace trim
- Docs drift: README "Validation Rules" VAL-01 orphan rule + unused root testdata (pre-existing, acknowledged)

## Shipped

### v1.12 C4D DSL Alternative (Shipped: 2026-08-17)

**Goal:** Deliver the C4D format — a `.c4d` brace-block D2-inspired DSL with full TOML feature parity — parseable directly to `*parser.Model` and renderable through the unchanged pipeline, with bidirectional canonical-equivalent converters (`convert to-toml`/`to-c4d`), a gofmt-style comment-preserving formatter (`fmt`) for both formats, nested use and recursive template-instantiating-template expansion, plus full README/skill/example documentation.

**Key accomplishments:** 1 phase (35), 9 plans, 25 tasks. Full C4D grammar (pigeon PEG, committed generated parser) + typed AST + `c4d.Parse` front-end (D-01..D-21); composition surface in both formats with recursive template expansion — v1.10 nesting deferral lifted (D-13..D-19); canonical-equivalent converters with `--follow-includes` graph migration and 29-fixture parity corpus (D-22..D-30); comment-preserving `fmt` for both formats with `--check` + semantic safety gate (D-31..D-33); README C4D section, 12 render-parity-enforced twins, dual-format skill synced to all plugin copies (D-34/D-35). Close-out: 3 verification blockers fixed (representability + write gate, width/height parity, `-o` structure preservation); 24/24 truths, 12/12 UAT, 30/30 security threats closed.

**Goal:** Generated diagram labels render with proper word wrapping and aspect-ratio sizing — edge labels formatted like unit labels (wrapped rectangle with `LabelRatio` aspect ratio, invisible borders), and line breaks at word boundaries only (no mid-word splits).

**Key accomplishments:** 1 phase (34), 4 plans (all TDD, RED→GREEN discipline), 28 commits; LABEL-01 edge labels as borderless HTML-table rectangles via `e.SetLabelHTML` (D-01..D-04); LABEL-02 `splitLongWord` removed, over-budget words unsplit (D-05); UAT fix 34-03 punctuation-aware tokenizer (`->`, `:`, `-`, `_` break opportunities — "any punctuation must be considered word boundary"); UAT fix 34-04 self-consistent aspect-ratio sizing (closed-form quadratic `labelMaxCharsForText`, edge + unit labels, measured edge label ≈1.84:1 vs ≈0.2:1); COMPAT-01 enforced — canonicalDOT goldens re-baselined with documented label-only deltas (incl. REF-05 empty-label attribute emission preserved).

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
- **Input Format**: TOML or C4D — single file or composed via `[[include]]` / `[template.*]`+`[[use]]`
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
| Visible-paths-only unfolding (v1.15 close, 260831-01u) | CTX-02/03 whole-subtree unfolding flooded the non-expanded root with the entire model; root must stay a compact C1 neighborhood | ✓ Good — pinned by deepcross fixture invariant; deep-link chains still unfold to true targets |
| `--no-labels` = edge labels only (v1.15 close, 260831-01u) | Suppressed node/cluster labels made `--expanded` render anonymous rectangles; users wanted decluttered edges, not anonymous shapes | ⚠️ Revisit if a true all-labels-off key is ever wanted — supersede LBL-01 deliberately |
| Builder-assigned `Edge.Name` for cgraph identity (v1.15 close, 260831-01u) | Converter derived edge names from labels; find-or-create then silently merged parallel edges once labels were suppressed — flags must never change topology | ✓ Good — `TestEdgeTopologyFlagInvariant` proves identical edge multisets across flag compositions |
| Branch protection requires Build/Lint/Test on master (v1.15 close) | PR Sanity lint had gone red unnoticed; PR merges must be gated on the sanity suite | ✓ Good — strict up-to-date check; enforce_admins stays off so direct master pushes keep working |
| Invocation-global `--edges` override (v1.16, D-03) | Per-unit `edges` surviving the flag would create unpredictable mixed routing across views; `--plain`/`--no-*` family precedent is invocation-global | ✓ Good — pinned E2E (`--edges` beats per-unit ortho on its own drill-down) |
| `View.EdgesOverride` carrier applied post-PLAIN-02 (v1.16, D-05 mechanism) | Explicit CLI flag is user intent and must beat `--plain`'s author-format zeroing; a dedicated empty-default field makes flag-off invariance structural, not just tested | ✓ Good — `TestEdgesSurvivesPlain` pins the KEY-02 delta; zero golden churn flag-off |

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
*Last updated: 2026-08-31 — milestone v1.16 COMPLETE (Edge Style Override shipped as v1.23.0)*
