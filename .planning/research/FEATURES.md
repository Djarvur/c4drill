# Feature Research

**Domain:** LikeC4 → C4Drill DSL converter (v1.11 LikeC4 Compatibility Layer)
**Researched:** 2026-08-08
**Confidence:** HIGH — source of truth is `.planning/research/likec4-dsl-brief.md` (full DSL brief, scraped from likec4.dev + bigbank.c4 canonical example); C4Drill side grounded against `internal/model/unit.go` and `internal/parser/parser.go`.

> Supersedes the v1.10 FEATURES.md (model-composition landscape, now shipped). This document scopes the v1.11 converter: for each LikeC4 construct, classify how it maps to C4Drill — table stakes, differentiator, or anti-feature.

## Feature Landscape

North star: "Render any model" — accept ANY valid LikeC4 source, produce a reasonable diagram, NEVER crash on unsupported constructs. Unsupported constructs are dropped with deduplicated stderr warnings (one warning per construct type, not per occurrence).

Three tiers:
- **Table stakes** — the converter is useless without these. Every real LikeC4 file uses them.
- **Differentiators** — valuable, low-cost wins that improve fidelity without expanding scope. v1.x candidates.
- **Anti-features** — explicitly DROPPED. Documented so future contributors don't waste cycles re-litigating.

### Table Stakes (Users Expect These)

Features users assume work. Missing these = converter produces garbage on the canonical bigbank.c4 example.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| Top-level block routing | Every `.c4` file has `specification {}`, `model {}`, and usually `views {}` blocks; the converter must dispatch on these keywords. | LOW | Hand-written recursive-descent parser. `specification` builds the kind→shape/style registry; `model` is the only block that becomes a `*parser.Model`. `views`, `deployment`, `global` are recognized keywords → dropped with one warning each (dedup). Multiple blocks of the same type concatenate (brief §1). |
| Lexical `{}` nesting → dotted-path subunits | LikeC4 parent-child is established **purely by braces** (brief §4); there is no `parent` attribute. C4Drill's subunit tree IS dotted-path keys. | MEDIUM | Parser tracks a scope stack: on `name = kind {`, push `name`; on `}`, pop. FQN = join(stack, "."). Both declaration syntaxes (`component backend { }` kind-first AND `backend = component { }` name-first) must reduce to the same AST node — bigbank uses kind-first. Watch the `spa = mobileApp "..."` case where kind and name share a token (different namespaces, legal). |
| `->` relationships (infix form) | 25 of 25 relationships in bigbank use the `a -> b 'title'` form. | MEDIUM | Two parse sites: top-level inside `model {}` and nested inside element bodies. Inline triple is `[title] [description] [technology]` (positional, optional). Strings are single OR double quoted (`'uses'` and `"uses"` both appear in bigbank). Emit a `model.Link` keyed by absolute target FQN; resolve short names via scope-hoisting before emission. |
| `this` / `it` / sourceless `->` resolution | Nested relationships use parent as implicit source/target (brief §3). `it -> frontend`, `frontend -> this`, bare `-> frontend` all legal. | MEDIUM | Resolve at parse time: `this`/`it` → current scope's enclosing element FQN; bare `-> target` → source = enclosing element. Must run BEFORE converting to absolute FQN targets. Failure to resolve is a hard error (consistent with LikeC4 rejecting ambiguous short names). Depends on lexical-nesting feature for the scope stack. |
| Kind name → C4Drill UnitType fuzzy mapping | LikeC4 has NO built-in kinds (brief §2); every kind is user-declared. C4Drill has 6 + External variants (`internal/model/unit.go`: person/system/db/queue/box + container/component composites). | MEDIUM | Static lookup table. Exact matches first, then case-insensitive contains: `person`→person, `database`/`db`/`rdbms`→db, `queue`/`kafka`/`topic`→queue, `container`→container, `system`/`enterprise`/`softwaresystem`/`spa`/`app`→system, `component`→component, fallback→box (only `system`/`box` can have subunits — non-fallback kinds that nest must promote to `box`). bigbank's 10 kinds (`person/enterprise/staff/existingsystem/softwaresystem/spa/mobileApp/container/component/database`) all map cleanly with this table. |
| `title` / `description` / `technology` field mapping | Every element in bigbank carries `description`; ~half carry `technology`. These are the only element properties C4Drill renders. | LOW | Direct field copy: LikeC4 `title` → C4Drill `name` (when present; otherwise name humanized from FQN last segment, the v1.10 ERGO path), `description` → `description`, `technology` → `technology`. Drop `summary` (C4Drill has no separate summary field; LikeC4 falls back from summary→description anyway). Inline positional and nested `description: "..."` forms (with optional colon, brief §10) must both parse. |
| `extend <element> {}` cross-file merging | LikeC4's ONLY explicit cross-file mechanism (brief §9). Real workspaces split a model across files and re-open scopes with `extend cloud { ... }`. | HIGH | Multiple `extend` blocks targeting the same FQN must accumulate children/properties (tags appended, metadata merged with dup→array dedup, links appended). Converter must merge into the same `*parser.Unit` node. Single-file `extend` is table stakes (common even within one file). Cross-file `extend` requires either synthesizing C4Drill `[[include]]` directives OR accepting multiple `.c4` inputs — defer the cross-file form to v1.x. |
| Comments (`//` and `/* */`) | Every real `.c4` file has comments. Parser must skip them or it crashes. | LOW | Lexer strips `//` to EOL and `/* */` blocks. Bigbank uses `//` mid-block. Trivial but a parser that doesn't handle this is DOA. |
| Graceful-degradation warning emitter | North-star contract: NEVER crash on unsupported constructs. Each dropped construct type emits ONE deduplicated warning. | LOW | Warning set with a `map[string]bool` dedup gate in the converter. Distinct keys per construct class: `deployments`, `custom-kind:<kindName>` (one per kind), `icons`, `tags`, `metadata`, `views`, `view-predicates`, `auto-layout`, `relationship-style`. Emit to stderr (NOT stdout — keeps DOT/SVG output clean for piping). |

### Differentiators (Competitive Advantage)

Valuable, not required for v1. Low-cost fidelity wins that ride existing C4Drill features.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Relationship kind preservation in `technology` | LikeC4 `-[async]->` carries semantic info; C4Drill has no kind taxonomy but `technology` is a free-text field already on links. | LOW | Concatenate: if relationship has both kind and technology, emit `"<kind>: <technology>"` (e.g. `"async: HTTPS"`); if kind only, emit `"<kind>"`. Zero new fields. Depends on `->` parsing. |
| `<->` bidirectional as two links | LikeC4 `<->` and `-[kind]<->` mean "both directions"; C4Drill links are strictly directional. | LOW | On encountering `<->`, emit two `model.Link` entries: `from→to` and `to→from`. One warning if `style.multiple` or `head`/`tail` arrow decorations are present (we drop them). Edge case: `<->` with `style { multiple true }` should still emit just two links (parallel-edges styling is meaningless in C4Drill's GraphViz output). |
| `link <url>` → C4Drill `reference` field | v1.10 added `reference` (📖 marker) per-unit URLs (PROJECT.md Phase 28). LikeC4 `link` is the direct analog. | LOW | LikeC4 allows MULTIPLE links per element; C4Drill allows ONE `reference` per unit (single URL slot). Take the FIRST `link http(s)://...`; warn-once if additional links are dropped. Non-http(s) schemes (`ssh://`, relative paths like `../src/index.ts`) get a warning and are dropped (matches v1.10 HTML-shim XSS hardening). |
| Shape mapping (LikeC4 `style.shape` → C4Drill type) | LikeC4 lets `style { shape cylinder }` override geometry independent of kind (brief §6). C4Drill derives shape from UnitType. | MEDIUM | Override the fuzzy-kind-mapped type when an explicit shape is present: `shape person`→person, `shape cylinder`→db, `shape queue`→queue, `shape storage`→db, `shape browser`/`mobile`→system (browser is just a styled box to C4Drill), others→leave kind-mapped type. Only meaningful when the converter reads `style {}` blocks (which the table-stakes path otherwise ignores). Warn-once for `shape component`/`document`/`bucket` (no C4Drill analog). |
| `name = kind` humanization fallback | LikeC4 allows `component api` with no title; the display name is the identifier. | LOW | Already solved by v1.10 ERGO-03 `model.Humanize` (PROJECT.md Phase 29). The converter just omits `name` and lets the existing humanization pass derive "Api" from `api`. Zero new code — just don't synthesize a name. |
| LikeC4 `specification` kind metadata as defaults | A kind declaration can carry default `title`/`description`/`technology` applied to all instances (brief §2). | MEDIUM | When converting an element, if a field is absent on the instance but present on its kind declaration, copy from the kind. Requires the parser to retain the `specification` block as a kind→defaults map. Useful for the bigbank case where `technology` lives on the kind not the instance. |

### Anti-Features (Commonly Requested, Often Problematic)

Explicitly dropped. These either fight C4Drill's auto-generation philosophy or have no analog. Documented once so the warning emitter dedups them and contributors don't re-propose.

| Feature | Why Requested | Why Problematic | Alternative |
|---------|---------------|-----------------|-------------|
| `views {}` DSL | LikeC4's signature feature — 8 views in bigbank alone. | C4Drill's CORE philosophy is auto-generated C1/C2/C3 from the model tree (PROJECT.md). Replaying LikeC4 view predicates (`include cloud.**`, `where kind is microservice`, `with { color muted }`) would mean re-implementing LikeC4's predicate engine inside C4Drill — a multi-month tangent. | Drop the entire `views {}` block with one warning. C4Drill auto-generates equivalent coverage. Brief §5 confirms: "a TOML-emitting converter would drop the entire `views {}` block." |
| `deployment {}` / `deploymentNode` | LikeC4 has a whole separate physical-layer model (brief §8). | C4Drill has zero deployment concept — no node types, no `instanceOf`, no deployment views. Mapping would require a parallel C4Drill model. | Drop `deployment {}` blocks with one warning. `specification { deploymentNode X }` declarations also dropped. |
| Icons (`icon`, `iconColor`, `iconSize`, `iconPosition`) | LikeC4 ships bundled icon libraries (aws/azure/gcp/tech/bootstrap — brief §6). | Project is explicitly on `dev/shapes-no-icons` branch (PROJECT.md v1.6: "Remove SVG icon extraction system entirely"). Icons were a v1.5 feature that was DELETED. Re-adding them contradicts a shipped decision. | Silently drop (no warning — icons are visual sugar, not semantic; user already knows the icon won't render because C4Drill has no icon slot). Brief §6 confirms shapes and icons are independent: dropping `icon*` doesn't break shape mapping. |
| `tags` / `metadata` visual rendering | LikeC4 tags drive view predicates and styling; metadata shows in detail dialogs. | C4Drill has no tag/metadata-driven rendering. Tags only matter inside `views {}` predicates (which we drop). Metadata has no UI surface in C4Drill output. | Drop with one warning each (`tags`, `metadata`). The kind→shape mapping uses kind NAME, not tags, so dropping tags doesn't affect diagram correctness. |
| View predicates, auto-layout directives, rank constraints | LikeC4 `autoLayout LeftRight 120 110`, `rank same { a, b }`, `group 'Label' {}` (brief §5). | C4Drill relies on GraphViz auto-layout (PROJECT.md Key Decision: "Auto-layout only — Let GraphViz handle positioning ✓ Good"). Honoring LikeC4 layout hints would fight GraphViz. | Drop silently as part of the `views {}` block drop. No separate warning (already covered by the `views` warning). |
| Custom element-kind inheritance / subtyping | Plausible — "microservice extends service" seems natural. | Confirmed NONEXISTENT in LikeC4 (brief §2 negative finding: "No element-kind subtyping / `kind.of` hierarchy"). The grammar has no such syntax; `element.kind` is a flat scalar. | The fuzzy-name mapper handles this implicitly: `microservice`→system by contains-match. Document the mapping table; don't invent inheritance. |
| Relationship arrow decorations (`head`, `tail`, `line`) | LikeC4 supports `head diamond`, `tail crow`, `line dotted` (brief §6). | C4Drill edges derive style from a fixed `style` enum (none/solid/dotted/dashed) on links — no per-arrowhead control. GraphViz `arrowhead`/`arrowtail` would require renderer changes. | Drop. The `line` value maps loosely to C4Drill's link `style` (dotted/dashed/solid) as a stretch differentiator; `head`/`tail` have no analog. |
| `navigateTo <viewId>` on relationships | LikeC4 lets a relationship drill into a dynamic view. | We drop `views {}` entirely, so `navigateTo` targets nothing. | Drop silently (covered by the `views` warning). |
| File-level `import` / `#include` | Tempting for multi-file workspaces. | Confirmed NONEXISTENT in LikeC4 DSL (brief §9). Composition is implicit file-merge + `extend`. | v1.11 is single-file; multi-file workspace support rides C4Drill's `[[include]]` (v1.10) if added later. Don't invent DSL syntax. |

## Feature Dependencies

```
[Lexical {} nesting → dotted paths]
    └──requires──> [Comments // and /* */]
    └──requires──> [Top-level block routing]
    └──enables───> [this/it/sourceless -> resolution]
                       └──requires──> [-> relationship parsing]
    └──enables───> [Kind name fuzzy mapping]
                       └──enhanced by──> [Shape mapping override]
                       └──enhanced by──> [specification kind-default metadata]

[extend <element> {}]
    └──requires──> [Lexical {} nesting]  (single-file)
    └──requires──> [Multi-file include pipeline (v1.10)]  (cross-file)

[Graceful-degradation warning emitter]
    └──required by──> ALL anti-features (deployments/views/icons/tags/metadata)

[Relationship kind preservation in technology]
    └──requires──> [-> relationship parsing]

[<-> bidirectional]
    └──requires──> [-> relationship parsing]

[link → reference field mapping]
    └──requires──> [-> relationship parsing]  (links attach to relationships too)
    └──requires──> [Lexical {} nesting]  (links also attach to elements)

[Anti-features (views/deployments/icons/tags/metadata)]
    ──all conflict with──> [C4Drill auto-generation philosophy]
    ──all depend on──> [Graceful-degradation warning emitter]
```

### Dependency Notes

- **Lexical `{}` nesting requires Comments + Top-level block routing:** A recursive-descent parser can't track brace depth if it chokes on comments or doesn't know which top-level block it's inside. These three are the parser's foundation.
- **`this`/`it`/sourceless resolution requires `->` parsing:** Both ride the same scope stack; the resolver just back-fills the source slot. Building them together avoids a second AST walk.
- **Kind fuzzy mapping enables, and is enhanced by, Shape mapping:** The kind table is the floor; explicit `style.shape` overrides it. Both feed the same UnitType slot on `*parser.Unit`.
- **`extend` cross-file requires C4Drill's `[[include]]` pipeline (v1.10):** For single-file `extend`, the converter merges in-process; for cross-file workspace `extend`, the converter would need to either (a) synthesize C4Drill `[[include]]` directives, or (b) accept multiple `.c4` inputs. v1.11 ships single-file only.
- **All anti-features require the warning emitter:** The north-star "NEVER crash" contract is enforced by routing every dropped construct through the dedup warning gate. Without it, the converter is silent — users won't know why their `views {}` produced no separate diagrams.
- **Anti-features conflict with C4Drill's auto-generation philosophy:** PROJECT.md Key Decision (v1.10): "No LikeC4 feature adoption except `reference` — Custom kinds, tags, icons, metadata, deployment model, user-authored views all fight C4Drill's auto-generation philosophy ✓ Good." This is a locked architectural stance, not a v1.11 limitation.

## MVP Definition

### Launch With (v1.11)

Minimum viable converter — must render bigbank.c4 into a reasonable C1/C2/C3 hierarchy.

- [ ] Top-level block routing (`specification`/`model`/`views`/`deployment`/`global` dispatch) — without this, nothing parses
- [ ] Lexical `{}` nesting → dotted-path subunits — LikeC4's primary composition mechanism
- [ ] `->` relationship parsing (infix form, inline triple `[title] [description] [technology]`) — every relationship in bigbank uses this
- [ ] `this`/`it`/sourceless `->` resolution — nested relationships are common
- [ ] Kind name → C4Drill UnitType fuzzy mapping — without it, every kind falls back to `box` and the diagram is shapeless
- [ ] `title`/`description`/`technology` field mapping — the only element properties C4Drill renders
- [ ] Comments (`//` and `/* */`) — parser doesn't work without this
- [ ] Graceful-degradation warning emitter — north-star "NEVER crash" contract
- [ ] Extension-based routing (`.c4`/`.likec4` → converter; `.toml` → existing parser) — zero new flags, hard BC contract
- [ ] Single-file `extend <element> {}` merging — common in real workspaces
- [ ] Anti-feature drops: `views {}`, `deployments {}`, icons, tags/metadata — each with dedup warning

### Add After Validation (v1.x)

Low-cost differentiators once v1.11 ships and the bigbank fixture renders cleanly.

- [ ] Relationship kind preservation in `technology` — trigger: user reports losing `async`/`sync` semantics
- [ ] `<->` bidirectional as two links — trigger: first `<->` in the wild
- [ ] `link <url>` → `reference` field mapping — trigger: user wants clickable external doc links (rides v1.10 📖 feature)
- [ ] Shape mapping override (`style.shape cylinder`) — trigger: user reports wrong shape from kind name alone
- [ ] `specification` kind-default metadata inheritance — trigger: bigbank-style files where `technology` lives on the kind

### Future Consideration (v2+)

Defer until converter has users and feedback.

- [ ] Cross-file `extend` workspace merging — defer until multi-file `.c4` workspace demand materializes; would ride C4Drill's `[[include]]` pipeline
- [ ] LikeC4 `relationship style.line` → C4Drill link `style` — defer; partial analog, low value
- [ ] `metadata` key-value passthrough into C4Drill tooltips/HTML — defer until C4Drill has a tooltip surface (currently none)

## Feature Prioritization Matrix

| Feature | User Value | Implementation Cost | Priority |
|---------|------------|---------------------|----------|
| Top-level block routing | HIGH | LOW | P1 |
| Lexical `{}` nesting → dotted paths | HIGH | MEDIUM | P1 |
| `->` relationship parsing | HIGH | MEDIUM | P1 |
| `this`/`it`/sourceless resolution | HIGH | MEDIUM | P1 |
| Kind name fuzzy mapping | HIGH | MEDIUM | P1 |
| `title`/`description`/`technology` mapping | HIGH | LOW | P1 |
| Comments | HIGH | LOW | P1 |
| Graceful-degradation warning emitter | HIGH | LOW | P1 |
| Extension-based routing (`.c4`/`.toml`) | HIGH | LOW | P1 |
| Single-file `extend` merging | HIGH | HIGH | P1 |
| Anti-feature drops (views/deployments/icons/tags/metadata) | HIGH | LOW | P1 |
| Relationship kind → `technology` | MEDIUM | LOW | P2 |
| `<->` bidirectional | MEDIUM | LOW | P2 |
| `link` → `reference` mapping | MEDIUM | LOW | P2 |
| Shape mapping override | MEDIUM | MEDIUM | P2 |
| Kind-default metadata inheritance | MEDIUM | MEDIUM | P2 |
| Cross-file `extend` workspace merge | LOW | HIGH | P3 |
| Relationship `line` style mapping | LOW | LOW | P3 |
| Metadata → tooltip passthrough | LOW | HIGH | P3 |

**Priority key:**
- P1: Must have for v1.11 launch — without these, bigbank.c4 fails
- P2: Should have, add in v1.x when trigger fires
- P3: Nice to have, defer to v2+ pending user demand

## Competitor Feature Analysis

| Feature | LikeC4 (source DSL) | Structurizr (Simon Brown's reference) | C4Drill (our target) |
|---------|----------------------|---------------------------------------|----------------------|
| Element kinds | User-declared, flat, no hierarchy | Fixed set (person/softwareSystem/container/component) | Fixed set + External variants + box grouping |
| Nesting | Lexical `{}` only | Explicit `parent` or model-tree API | Dotted-path TOML keys (lexical-equivalent) |
| Relationships | `->` infix, kinded, `<->`, `this`/`it` | Directed, typed | Directed only, `technology` free-text |
| Views | Full predicate DSL, auto-layout, scoped | Auto-generated + manual | AUTO-GENERATED ONLY (core philosophy) |
| Deployment | Separate `deployment {}` layer | Separate deployment model | NONE (out of scope) |
| Icons | Bundled libraries (aws/azure/gcp/tech) | Limited | REMOVED in v1.6 (deliberate) |
| Cross-file | Implicit merge + `extend` | Workspace JSON includes | `[[include]]` directive (v1.10) |
| Our approach | Accept `.c4`, drop what doesn't fit, render the rest | N/A — different input format | Converter bridges LikeC4 → C4Drill TOML model in-process |

## Sources

- `.planning/research/likec4-dsl-brief.md` — full LikeC4 DSL technical brief (scraped from likec4.dev + bigbank.c4 canonical example)
- `.planning/PROJECT.md` — C4Drill project history, validated requirements, locked key decisions
- `internal/model/unit.go` — C4Drill UnitType enum (person/system/db/queue/box + container/component composites + External variants)
- `internal/parser/parser.go` — C4Drill TOML parser pipeline, generic-type inference, parent-default-type rules
- likec4.dev DSL docs (specification/model/relationships/views/styling/deployment/extend/references)
- `github.com/likec4/likec4/apps/playground/src/examples/bigbank/bigbank.c4` — 228-line canonical real-world example (10 kinds, 3 nesting levels, 25 relationships, 8 views)

---
*Feature research for: LikeC4 → C4Drill DSL converter (v1.11 LikeC4 Compatibility Layer)*
*Researched: 2026-08-08*
