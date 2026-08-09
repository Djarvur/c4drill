# Project Research Summary — v1.10 Model Composition

**Project:** C4Drill (Go CLI rendering C4 diagrams from TOML)
**Milestone:** v1.10 Model Composition
**Features:** include directive · unit templates · TOML ergonomics (relative peer, optional name) · reference field
**Researched:** 2026-08-08
**Confidence:** HIGH (codebase citations live; capability matrices built from fetched authoritative docs)
**Source docs:** [STACK.md](STACK.md), [FEATURES.md](FEATURES.md), [ARCHITECTURE-v1.10.md](ARCHITECTURE-v1.10.md), [PITFALLS.md](PITFALLS.md)

> This summary supersedes the prior v1.1 "AI-Ready" SUMMARY.md. It is the single load-bearing document the roadmapper reads alongside `PROJECT.md` and `REQUIREMENTS.md`. The four source docs are NOT re-read downstream.

---

## 1. Executive Summary

v1.10 makes C4Drill compose a model from **multiple files**, **parametrized templates**, and **shorter, more readable TOML**, and adds per-element **hyperlinks**. The four features are independent at the data level (each adds fields or a parse pass) but form a strict **runtime pipeline** at integration time: `include → template-expand → relative-peer-resolve → validate → generate-views → render`. The headline decisions: **zero new dependencies** (everything is hand-rolled Go in ~150–200 lines); **hard-error everywhere on conflict** (no silent merge/override); a **Model-extension approach** that keeps validator/view/render untouched; and a **single GraphViz `URL` attribute** path for the reference field, making it the cheapest table-stakes item in the milestone. The two design forks most likely to cause silent corruption if left undecided are (a) deep-copy semantics for templates and (b) the resolution site for relative peers authored inside templates — both are HS items below and must be settled in the discuss phase.

---

## 2. Stack Decisions

### Zero new external dependencies

All four features map onto the existing stack with small hand-rolled helpers at known integration points. The codebase's `Model`/`Unit`/`Link` structs are small and known-shape, so reflection-based or config-framework libraries are both unnecessary and riskier than targeted code.

| Hand-roll | Why no library fits | Approx size |
|---|---|---|
| INCLUDE parse-merge | go-toml/v2 has no merge capability (verified through v2.4.3 release notes). The merge is a union of `Units` maps + append of `UnitOrder` + overlay of `Properties` — ~30 lines. `include_once` and cycle detection are bespoke regardless. | ~30 LOC |
| Unit deep-copy (`Unit.Clone()`) | Reflection-based copiers (`mohae/deepcopy`, `jinzhu/copier`, `ulule/deepcopier`) **silently skip unexported fields** — would drop `Link.Mirror` (`link.go:67`, load-bearing for validator at `index.go:80` and graph at `builder.go:438`) and `Unit.SubunitOrder` (`unit.go:69`, `toml:"-"`). JSON round-trip loses both; gob round-trip works but is ~10× slower and forces `gob.Register`. | ~15 LOC |
| `${param}` substitution | `text/template` brings a 95%-unused feature set plus template-injection surface. `strings.NewReplacer` is simpler, safer, and allocation-cheap for 2–4 params. | ~25 LOC |
| Identifier humanization | No library fits: `stoewer/go-strcase` has no acronym awareness; `iancoleman/strcase`'s acronym support is forward-only (so reverse splitting yields screaming case, not title case). The domain is tiny (TOML keys), so a ~25-line splitter + small acronym table wins. | ~25 LOC |
| Relative-peer resolution | `path.Join`/`path.Dir` from stdlib. Pure post-parse pass. | trivial |
| Reference field | One new `string` field + SVG anchor render branch. | trivial |

**Rejected (with rationale):** viper (config framework, wrong shape); koanf (operates on generic maps, throws away typed parser — contingency only if merge grows arbitrary depth); all reflection/gob/json copiers (see above); `text/template` (overkill); `gobuffalo/validate` (C4Drill already has a purpose-built validator); strcase libraries (wrong direction for acronym handling).

### One housekeeping bump

`github.com/pelletier/go-toml/v2`: **v2.2.4 → v2.4.3** (latest, released 2026-07-05). The API surface C4Drill uses (`toml.Unmarshal`, `toml.Marshal`, `unstable.Parser`/`unstable.Table`/`expr.Key`) is stable across the range; notable v2.3.0 added "UnmarshalText fallbacks to struct unmarshaling for tables and arrays." Low-risk minor bump. **No merge/compose feature exists in any release** — confirms the hand-roll decision for INCLUDE.

All other deps (go-graphviz on the `onokonem` replace, cobra, testify, levenshtein) stay at current lines. Go 1.26.1 stays.

---

## 3. Feature Classification

From the capability matrices in FEATURES.md. **Table stakes** = absence feels broken; **Differentiator** = polish or genuine novelty vs comparable tools; **Anti-feature** = fights C4Drill's auto-generated-view / structured-TOML / single-shot-CLI philosophy.

| Feature area | Table stakes | Differentiator | Anti-feature (do NOT ship) |
|---|---|---|---|
| **Include** | directive; transitive includes; **path relative to the including file**; cycle detection (fatal); **hard-error on duplicate unit path** | `once = true` (PlantUML `!include_once` — *essential* for the template-library use case, not optional in practice) | Terraform-override-style silent merge (breaks "TOML is source of truth"); partial import / namespacing; URL includes |
| **Templates** | **named params** (not positional `${1}`); **trailing defaults**; **hard-error on missing required param** (no silent literal `${name}`); substitution into ALL string fields incl. `Link.Peer`, `reference` | **one-template-one-unit** (deliberate deviation from every precedent's multi-output; correct for structured TOML) | multi-output / `for_each` (fights structured TOML); nesting (resolution-order complexity); forward references (requires multi-pass); typed params |
| **Ergonomics** | *(none)* — no comparable tool does either item | **relative-peer resolution** (genuinely novel: no surveyed tool does sibling-scope-walk); **optional `name`** with humanization (more aggressive than every peer — all require explicit names or explicit `title()`); walk-up precedence | acronym preservation (Terraform `title()` proves an unsolved tar pit — ship dumb camelCase split, escape via explicit `name`); compact-link string shorthand |
| **Reference** | per-element URL field; clickable SVG via GraphViz `URL` attribute (near-zero cost — existing backend); **`reference` name** (collision-free — `link` is taken for relationships); HTML-shim coverage for Safari | 📖 visible affordance (matches repo's existing 🔍 collapsed-cluster indicator at `builder.go:258-261`) | URL over-validation (LikeC4 supports any URI scheme); link-specific tooltip (duplicates `description`); per-relationship URL; image/background property |

**Single most relevant borrowed idiom per feature:** Include ← PlantUML `!include_once`; Templates ← named+defaulted params (PlantUML `!procedure`, Jsonnet); Ergonomics ← *no precedent* (C4Drill is more aggressive than peers); Reference ← GraphViz `URL` attribute (zero-cost via existing go-graphviz).

---

## 4. Architecture

### The pipeline (extended)

```
ParseFile ──▶ include.Resolve ──▶ template.Expand ──▶ peer.Resolve ──▶ [humanize-names] ──▶ Validate ──▶ views ──▶ render
  (root.go:112)   (NEW ~:113)      (NEW ~:114)       (NEW ~:115)      (NEW, post-expand)   (:118)    (:209)     (:235)
```

The four features are **pure pre-processing passes inserted between Stage 1 (Parse) and Stage 2 (Validate)** in `cmd/c4drill/root.go:runRoot`. Each takes a `*parser.Model` and returns a `*parser.Model`. The post-chain model is structurally identical to a hand-authored single-file model.

**Ordering is load-bearing for correctness** (not stylistic):

1. **include first** — so templates defined in an included file are visible to the root's `[[use]]` (the "template isolation" motivating use case).
2. **template-expand second** — so peers *inside* a templated unit exist before relative-peer resolution; a templated unit's relative links would otherwise resolve against a model missing the unit being expanded.
3. **relative-peer third** — needs the fully assembled + expanded model to look up peer targets (peers can be forward references across files/templates).
4. **humanize-names** — runs after expand (so templated units get humanized from their substituted instantiation key, not from `${name}`) and before validate (so validator error messages show final names).
5. **validate last** — peer-existence (`rules.go:14`) and level checks (`rules.go:187`) must see the final, concrete unit set; the validator remains the single gatekeeper (STATE.md D-12).

### Model-extension approach (not sidecar structs)

**Dedicated fields on `parser.Model`, all `toml:"-"`, extracted in `Parse` from the raw map before the unit loop** — mirroring the existing `properties` extraction at `parser.go:68-77`:

```go
type Model struct {
    Properties model.Properties
    UnitOrder  []string
    Units      map[string]*model.Unit
    // v1.10 composition fields (not read by validator/view/render):
    Includes       []IncludeDirective   // [[include]] directives, file order
    Templates      map[string]*TemplateDef  // [template.<name>] bodies
    Instantiations []Instantiation      // [[use]] requests, file order
}
```

Sidecar structs (`IncludeGraph`, `TemplateLibrary`, …) were rejected because they would force changing the signatures of `validator.Validate`, the three `view.Generate*` functions, `collectExpandedPaths`, and every caller — wide blast radius for no gain.

### Unchanged consumers (confirmed by reading field accesses)

- **`validator.Validate`** (`validator.go:16`) — calls `BuildIndex(m.Units, "")` at `:22`; reads **`m.Units` only**. All rules in `rules.go` operate on the flat index. **Zero change needed.**
- **`view.GenerateC1View`/`C2`/`C3`/`ExpandedView`** (`scope.go:86/382/454/14`) — read `m.Units`, `m.UnitOrder` (with map-key fallback), `m.Properties`. **Zero change needed.**
- **`render.*`** (`converter.go`, `labels.go`, `render.go`) — consume `*graph.Graph`, never `*parser.Model`. **Zero change needed** (the reference-field glyph is a *new* feature, not a forced change).
- **`collectExpandedPaths`** (`root.go:156`) — reads `m.Units`. **Zero change needed.**

The only existing function whose behavior must change is **`parser.Parse`** (`parser.go:47`) and **`captureDefinitionOrder`** (`parser.go:100`) — to extract the new reserved top-level tables (`include`, `template`, `use`) so they don't pollute the unit namespace.

---

## 5. High-Severity Risks

Both are flagged HS because they cause **silent corruption** — the model parses and renders without error but is wrong, and the failure surfaces far from its cause.

### HS-1: Deep-copy aliasing corrupts the template on the Nth instantiation

**What.** Under templates Option B, the expansion pass deep-copies a template `Unit` and substitutes params. A *shallow* copy shares the template's `Subunits` map (pointers), `Links`/`LinksFrom`/`Expanded` slice headers (shared backing array). Mutating instantiation #1 then mutates the template, so instantiation #2 starts from a corrupted template.

**Codebase-specific exposure.**

- `Unit` (`unit.go:41-72`) has three aliasing fields: `Subunits map[string]*Unit` (`:71`, `toml:",inline"`), `Links []Link` (`:65`), `LinksFrom []Link` (`:67`), `Expanded []string` (`:63`).
- **No existing deep-copy helper anywhere in the codebase** — the template feature introduces the first one.
- **The validator mutates units in place**: `populateIncomingLinks` (`validator/index.go:53-84`) does `targetInfo.Unit.LinksFrom = append(targetInfo.Unit.LinksFrom, model.Link{... Mirror: true})` (`index.go:70-81`). `BuildIndex` stores pointers into the same `Model.Units` graph. With Go slice semantics, if a shared backing array has spare capacity, the append *mutates in place without reallocating* — instantiation #2 silently inherits instantiation #1's mirror links.
- The graph builder reads these same shared structures (`builder.go:220,312` for subunits, `:360-367,410-434` for links, `:342` for `parent.Expanded`), so corruption propagates into rendered output.
- `Link.Mirror` (`link.go:67`, `toml:"-"`) is set only by the validator. An encoder/round-trip copy would silently reset `Mirror=false`, re-breaking multiplicity counting (STATE.md D-05). **Do not round-trip-copy Links.**

**Failure symptom.** First instantiation renders correctly; Nth renders with the first's name/links, or phantom mirror links appear, or output varies between runs depending on slice growth (non-deterministic, distinct from the documented go-graphviz layout nondeterminism).

**Prevention.**

- Hand-rolled recursive `Unit.Clone()` in `package model` (~15 lines, `slices.Clone` is already a project idiom at `parser.go:310`). Recursion allocates fresh `Subunits` map + each `*Unit`; copies `Links`/`LinksFrom` struct-by-struct so `Mirror` survives; allocates fresh `Expanded`.
- **Ordering rule:** deep-copy the template *once per instantiation, then substitute, then discard the copy*. Treat the template registry as immutable after parse.
- **Regression test (required, not optional):** instantiate the same template 3× with distinct params; assert (a) each instantiation's `Links[0].Description` differs, (b) re-running expansion a second time (idempotency) yields identical output, (c) after a full validate pass, two instantiations' `LinksFrom` slices are disjoint.

**Phase:** the Templates implementation phase — this is THE core correctness concern of that feature.

### HS-2: Relative-peer inside a template — template-parent vs instantiation-parent ambiguity

**What.** A relative peer resolves against the unit's *structural parent*. When the link is authored *inside a template*, "parent" is ambiguous: the template's lexical parent (often none — templates are top-level data) or the instantiation site's parent (where `[[use]]` places the concrete unit). Wrong choice produces **silently wrong edges**.

**Concrete case.** Template `[template.svc]` defines a link with `peer = "cache"`. Two instantiations under different parents: `[[use]] name=auth parent=mainSystem` and `[[use]] name=billing parent=otherSystem`, where both `mainSystem.cache` and `otherSystem.cache` exist. Instantiation-site resolution (recommended) gives `billing → otherSystem.cache` (correct). Template-site resolution either fails or resolves both to the *same* `cache`, fabricating a cross-system edge that renders and validates without complaint.

**Prevention.**

- **Decision (Design Fork — see §6):** relative peers authored in a template resolve against the **instantiation site's** structural parent, never the template's lexical location. Templates are data, not structural location.
- **Fallback order:** relative-first (search the unit's own parent's direct children), then absolute-fallback (if `peer` contains `.` OR exactly matches a top-level path, treat as absolute). This is the backward-compat contract: every existing model uses absolute peers and resolves identically.
- **Ambiguity at a single depth is a hard error**, not first-match. "Nearest wins" is for *nesting depth*, not for *ties at the same depth*.
- Implement as a **separate pass that rewrites `Link.Peer` in place** on the post-expansion model before `BuildIndex`/validation, so the validator's existing absolute-path logic is untouched.
- **Test:** template with relative peer instantiated under two different parents; assert each link resolves to its own parent's sibling, not a shared global.

**Phase:** Relative-peer implementation phase — the resolution-site decision must be settled in the discuss phase first (see §6).

### BC-1: Parser treats unknown keys as subunits — in TWO places

**What.** Adding top-level `[template.*]` / `[include]` / `[[use]]` tables without parser changes produces **phantom units** named `template`/`include`/`use` — the validator then emits confusing errors like "unit 'template' has type system which cannot have subunits" or "unit 'include' has no incoming or outgoing links".

**Two coordinated parser changes required:**

1. **`captureDefinitionOrder`** (`parser.go:100`, skip rule at `:128` currently only skips `properties`): extend the skip to also skip `[include]`, `[[include]]`, `[template.*]`, `[[use]]`. Without this, `[template.microservice]` (`len(parts)==2`) registers `microservice` under parent `template` in `subunitOrders["template"]`.
2. **`isBuiltinField`** (`parser.go:309-316`): the allowlist distinguishing struct fields from subunits. **Adding `reference` here is a safe one-liner** (leaf field). But adding `template`/`include`/`use` here is **not** a substitute for change #1 — the two functions have independent detection logic.

The `Parse` loop must extract these reserved tables from `rawMap` before the unit-processing loop (`:80`), exactly as `properties` is extracted at `:68-77`. Rejecting user units literally named `template`/`include`/`use` is itself a design fork (BC-2 — reserved-word collision policy).

---

## 6. Design Decisions Required (for discuss phase)

These forks cannot be resolved by implementation care — the wrong default is silent corruption or a broken backward-compat guarantee. Each should land as a STATE.md decision-log entry when the discuss phase resolves it.

1. **HS-2: Relative-peer resolution site for template-authored links.** *Recommended: instantiation-site parent.* The single most consequential design decision for the milestone.
2. **RP-2 / HS-2 wording: "relative-first" semantics.** Clarify it means relative-search-first *for peers that are not already valid absolute paths* — NOT "rewrite all peers as relative." Pin the exact gate: bare name + no `.` + not a top-level key + not already in index.
3. **TM-2: Forward-reference policy for templates.** *Recommended: allow* (free under Option B — the template registry is built from the fully-parsed Model before any `[[use]]` is processed). The "before use" intuition comes from textual-preprocessor references (Option A / go-metadot) and does not apply to structured post-parse expansion.
4. **TM-5: Unresolved `${param}` policy.** *Recommended: error* (strictness is a feature; structured post-parse gives the affordance). PlantUML/go-metadot leave literals — C4Drill should NOT.
5. **HU-1: Humanization approach and acronym allowlist.** *Recommended: `golang.org/x/text/cases` + hand-rolled C4-relevant acronym allowlist* (`gRPC`, `API`, `IDP`, `OAuth`, `SAML`, `REST`, `GraphQL`, `SQL`, …). Pure hand-roll or "require explicit `name`, no humanize" are the alternatives. The allowlist contents are themselves a mini-decision.
6. **BC-1 / BC-2: Directive-table naming and reserved-word collision policy.** Bare `[include]`/`[template]`/`[use]` (ergonomic, collides with rare legacy units named `use`) **vs** namespaced `[c4.include]`/`[c4drill.template]` (collision-proof, uglier). Backward-compat is non-negotiable, so (b) namespaced or (a) collision-detection are preferred over (c) reserve-and-accept-tiny-break.
7. **IN-2: `UnitOrder`/`SubunitOrder` merge semantics across included files.** *Recommended:* top-level = root units in definition order, then each included file's units appended in include directive order; subunits = included files may append subunits to a root-defined unit, with same-name conflict = hard error.
8. **IN-3: Diamond-include behavior without `once`.** *Recommended: include-twice-and-error-on-dup* (signals the author to add `once=true`); the diamond is legal, not a cycle. Cycle detection uses a **stack** (current chain); `once` dedup uses a **global visited-set** — distinct data structures.

---

## 7. Cross-Feature Dependencies

```
include ──enables──► template-isolation   (template libs in their own file, included with once=true)
include ──enables──► large-model-split    (one file per domain)

template-expand ──requires──► reference field exists            (templates carry parametrized reference URLs)
template-expand ──requires──► relative-peer resolved AFTER expand (templated links use ${name} in peer)
template-expand ──constrains──► one-template-one-unit           (so ${name} peer substitution is well-defined)

relative-peer ──consumes──► template expansion output            (peers inside templated units must resolve)
relative-peer ──orthogonal──► include                           (runs on the merged model; file origin irrelevant)

reference ──orthogonal──► include, relative-peer                 (just a field on the merged unit)
reference ──consumes──► template param substitution              (reference = "https://wiki/${name}" must substitute)

validate ──consumes──► all four                                  (runs last, on the fully assembled + expanded + resolved model)
render   ──consumes──► reference field                          (emits GraphViz URL attribute)
```

**The six concrete interaction cases the plan must test:** (1) template + reference parametrization; (2) include + templates (the motivating use case — `templates.toml` + `[[include]] once=true` + `[[use]]`); (3) template + relative-peer (`peer = "${name}Bus"` — pipeline order required); (4) include + UnitOrder concatenation (templates slot at the use-site's position, not the definition's); (5) optional-name + templates (humanization runs on the instantiation key, not the template segment); (6) reference + render format (📖 marker and SVG `<a>` must render in both `-f svg` and `-f html`; the HTML JS shim must handle external `reference` URLs distinctly from internal drill-down URLs — Safari silently ignores SVG `<a>` for navigation).

---

## 8. Suggested Phase Breakdown

v1.9 shipped as **Phase 27** (`ROADMAP.md:8`). Runtime dependency is `include → template → peer → validate`; the **build/ship order prioritizes low-risk independent work first**. Each pass is independently shippable (each is a no-op for models that don't use the feature).

| Phase | Name | Goal | Depends on |
|---|---|---|---|
| **28** | **Reference field (📖)** | External-docs URL on units: `Unit.Reference`; `isBuiltinField` allowlist update; `buildNode` glyph (Option A — append 📖 to label, matches 🔍 precedent at `builder.go:260`); `Node.ReferenceURL`; README + skill + example. | None — fully independent, parallelizable. |
| **29** | **Optional `name` humanization** | Derive display name from last path segment when `name` omitted: humanize function in parser/model; `Unit.Name` fallback at parse time; acronym allowlist; docs. | None — parallelizable with 28. |
| **30** | **Relative-peer resolution** | Short `peer` names resolve against enclosing parent: `internal/peer/Resolve`; hook in `runRoot` stage 1c; relative-first / absolute-fallback; tests + saira re-measure; backward-compat corpus test. | None (ships as a no-op pass for absolute-only models). |
| **31** | **Template expansion** | `[template.*]` define + `[[use]]` instantiate: `Model.Templates`/`.Instantiations` fields; `captureDefinitionOrder` skip for `template`/`use`; `internal/template/Expand` (deep-copy + `${param}` substitution into Unit + Link fields); drain instantiations into `Units`/`UnitOrder`; **HS-1 regression test**. | Benefits from 30 (templated links can be relative); ships with absolute peers if 30 slips. |
| **32** | **Include directive (multi-file)** | Assemble model from multiple files: `Model.Includes` field; `captureDefinitionOrder` skip for `include`; `internal/include/Resolve` (recursive merge, cycle detection via stack, `Once` via visited-set); merge rules (Units union/conflict-error, `UnitOrder` concatenate, Properties root-wins, Templates union). | 31 — the "template isolation" use case is the primary motivation; include must carry `Templates` through the merge. |
| **33** | **Docs sweep + end-to-end golden tests** | Update `README.md` / `skill/SKILL.md`; add `skill/examples/06-templates.toml`, `07-include/`; golden test: 3-file model produces same SVG (order-insensitive canonicalDOT comparator per STATE.md) as equivalent single-file model. | 30–32. |

**Parallelization:** Phases 28 and 29 touch disjoint code (`model/unit.go` + `graph/builder.go` + `render` vs parser humanization) and can proceed in parallel. Phases 30–32 are sequential due to the runtime ordering constraint.

**Parser-change prerequisite:** whichever of include/templates lands first needs the **two coordinated parser changes** (BC-1: `captureDefinitionOrder` skip + `Parse` rawMap extraction + `isBuiltinField` for `reference`). Land this as its own small, well-tested change before building features on top.

---

## 9. Watch Out For

The top items from PITFALLS, beyond the two HS risks and the parser change:

- **HS-1 (templates):** no existing deep-copy helper in the codebase; the validator mutates `LinksFrom` in place (`index.go:70-81`) — without a correct `Unit.Clone()` the second template instantiation silently inherits the first's state. Three-instantiation regression test is the acceptance criterion.
- **HS-2 (relative-peer in templates):** settle the resolution-site decision (instantiation-parent, not template-parent) in the discuss phase before implementing.
- **BC-1 (phantom units):** the parser treats unknown keys as subunits in TWO independent places (`captureDefinitionOrder` AND `isBuiltinField`) — coordinate both; `reference` is the only safe single-line addition.
- **BC-2 (reserved-word collision):** decide bare vs namespaced directive-table naming in discuss phase; backward-compat for legacy units named `use` is the constraint.
- **BC-3 (strict mode):** go-toml/v2 is non-strict by default and must **stay non-strict** — `DisallowUnknownFields()` is load-bearing OFF because the inline-subunit trick (`unit.go:71`, `toml:",inline"`) depends on unknown keys being silently accepted. Add a guard comment.
- **IN-3 (diamond includes):** a diamond (A→B, A→C, B→D, C→D) is **not a cycle** — use a stack for cycle detection, a separate opt-in visited-set for `once`. Max depth cap (e.g. 100) as defense-in-depth.
- **IN-5 / IN-4 (path resolution):** resolve every include path relative to the *including file's* directory (universal convention), not cwd — "works on my machine, breaks in CI" is the classic failure. Canonicalize paths via `filepath.Abs` + `filepath.Clean` before pushing on the stack/set.
- **TM-1 (path collision):** two `[[use]]` blocks producing the same unit path silently overwrite — expansion must hard-error naming both definitions before insertion.
- **TM-5 (unresolved `${param}`):** after substitution, scan string fields for any remaining `${...}` and error — no silent literals (PlantUML/go-metadot leave them; C4Drill should not).
- **TM-3 (recursion):** template nesting is disallowed — add an expansion-depth cap (e.g. 100) as defense-in-depth even with the ban.
- **RP-2 (backward-compat):** the relative-resolution pass must be a **no-op for every existing model** — corpus test asserting the `(source, resolved-peer)` set is byte-identical to today is the acceptance criterion. Gate: relative-search fires only when the peer is bare + has no `.` + isn't a top-level key + isn't already in the index.
- **HU-2 (humanize ordering):** humanization runs AFTER template expansion (so it sees the final substituted name/key, not `${name}`) and only when `name` is empty (explicit `name` always wins).
- **Golden comparisons:** all multi-file/template golden tests must use the **order-insensitive canonicalDOT** comparator (sort-normalize, strip layout geometry) per STATE.md — NOT byte-exact `require.Equal`. Multi-file and templates add another axis of ordering variance on top of the documented go-graphviz nondeterminism.

---

*Synthesis for: C4Drill v1.10 Model Composition milestone*
*Researched: 2026-08-08*
*Ready for roadmapper + planner: yes*
