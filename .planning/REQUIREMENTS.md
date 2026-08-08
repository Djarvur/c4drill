# Requirements: v1.10 Model Composition

**Status:** Active
**Milestone:** v1.10 Model Composition
**Last updated:** 2026-08-08

C4Drill expands from a single static TOML file into a composable, parametrized, multi-file format. All features are additive and backward-compatible; existing single-file models must parse and render unchanged. The four features form a strict runtime pipeline: `include → template-expand → relative-peer-resolve → validate → generate-views → render`.

Source research: [.planning/research/SUMMARY.md](research/SUMMARY.md)

---

## v1.10 Requirements

### INCLUDE — Multi-file composition

- [ ] **INC-01**: User can assemble a diagram from multiple TOML files using an `[[include]]` directive; the entry file plus all transitively-included files merge into one logical model.
- [ ] **INC-02**: Include paths resolve relative to the *including file's* directory (not the CLI working directory), so models work identically regardless of where `c4drill` is invoked.
- [ ] **INC-03**: Transitive includes work — an included file may itself contain `[[include]]` directives, resolved recursively.
- [ ] **INC-04**: Cycle detection is fatal — a file that includes itself directly or transitively (A→B→A) produces a clear error naming the cycle, not infinite recursion or silent corruption.
- [ ] **INC-05**: Diamond includes are *not* cycles — a file reachable by two paths (A→B→D, A→C→D) is legal; behavior without `once` is hard-error on the resulting duplicate unit paths (signals the author to add `once=true`).
- [ ] **INC-06**: `once = true` on an `[[include]]` directive skips re-inclusion of an already-included file (PlantUML `!include_once` semantics), so shared template/definition libraries can be safely included from multiple model files without double-definition errors.
- [ ] **INC-07**: Merge is flat (no namespacing/prefixing) — included units merge into one namespace; a unit path defined in two files (after `once` dedup) is a hard error naming both files, not a silent override.
- [ ] **INC-08**: `[properties]` follows root-file-wins semantics — the entry file's `name`/`description` are authoritative; conflict on `[properties]` from an included file is a hard error.
- [ ] **INC-09**: `UnitOrder` concatenation preserves authoring order — entry file's units first, then each included file's units appended in include-directive order (so rendering order is deterministic and matches author intent).
- [ ] **INC-10**: A missing include file produces a clear error naming the referenced path and the including file, not a generic parse failure.

### TMPL — Unit templates

- [ ] **TMPL-01**: User can define a reusable, parametrized unit template in a `[template.<name>]` table with declared named parameters.
- [ ] **TMPL-02**: User can instantiate a template via a `[[use]]` directive, supplying concrete values for ALL of the template's declared parameters. **No parameter defaults** — every declared param is required at every instantiation; missing any is a hard error. (Simpler than PlantUML/go-metadot trailing defaults; strictness is a feature.)
- [ ] **TMPL-03**: `${param}` substitution applies to *all* string fields of the instantiated unit — `Name`, `Description`, `Technology`, `Reference`, `Color`, and every `Link` field (`Peer`, `Description`, `Technology`) — so a param can parameterize any text the unit carries. The template declares a **fixed set of links** (fixed count); each instantiation fills their fields from params (so the same template produces units linked differently — different peer/target/technology per instantiation — but always the same number of links). No array/conditional expansion (no `for_each`/fan-out).
- [ ] **TMPL-04**: A template may define a **single top-level unit plus its declared subunit subtree** — e.g. `[template.svc]` with `[template.svc.api]`, `[template.svc.db]`. One instantiation produces that whole subtree. Substitution applies to subunit fields too. (Relaxed from the earlier "exactly one unit" formulation.)
- [ ] **TMPL-05**: Instantiated units (and their subunits) participate in the model exactly like hand-authored units — they pass validation, appear in auto-generated C1/C2/C3 views, and render identically.
- [ ] **TMPL-06**: Instantiating a template with a missing parameter is a hard error naming the template, the parameter, and the instantiation site — no silent `${param}` literals in output (deliberate divergence from go-metadot/PlantUML).
- [ ] **TMPL-07**: Two `[[use]]` instantiations producing the same unit path (e.g. same `name`/`parent`) is a hard error naming both instantiation sites, not a silent overwrite.
- [ ] **TMPL-08**: Template deep-copy is correct under repeated instantiation — instantiating the same template N times with distinct params yields N independent unit subtrees; the 2nd instantiation never inherits state from the 1st (HS-1 from research: validator mutates `LinksFrom` in place, so a shallow copy corrupts subsequent instantiations). Deep-copy MUST recurse into `Subunits` (every `*Unit` in the map cloned, each subunit's links cloned) since templates may declare subunit subtrees (TMPL-04).
- [ ] **TMPL-09**: Forward references work — a `[[use]]` may appear textually before its `[template.<name>]` definition in the same file (structured post-parse expansion makes this free; the "define-before-use" restriction from go-metadot's textual preprocessor does not apply).
- [ ] **TMPL-10**: A templated unit's `reference` field substitutes params correctly (e.g. `reference = "https://wiki/${name}"`), so reference URLs can be parameterized (composes with the Phase 28 reference field).

### ERGO — TOML authoring ergonomics

- [ ] **ERGO-01**: A `peer` value resolves *relative to the enclosing parent block* when it is a bare name (no `.`) that matches a sibling unit — e.g. inside `[linuxSystem.localIDP]`, `peer = "sessionManager"` resolves to `linuxSystem.localIDP.sessionManager`.
- [ ] **ERGO-02**: Relative resolution falls back to absolute when the peer contains a `.` OR exactly matches a top-level unit path OR does not resolve as a relative sibling — so every existing model with absolute peers parses identically (backward-compat is a hard contract).
- [ ] **ERGO-03**: The `name` field is optional — when omitted, the display name is derived from the last path segment of the unit's identifier (e.g. `localIDP` → "Local IDP", `sessionManager` → "Session Manager").
- [ ] **ERGO-04**: Humanization is a dumb camelCase split — acronym preservation is explicitly out of scope (Terraform's `title()` proves it's an unsolved tar pit); `gRPC` humanizes to "Grpc" and authors escape via explicit `name =`.
- [ ] **ERGO-05**: Explicit `name =` always wins over humanization (backward-compat for every existing model).
- [ ] **ERGO-06**: A compact one-liner link shorthand is available so common edges (peer + optional technology/description) are writable without the multi-line array-of-tables form — exact syntax TBD in discuss/plan, but the authoring cost of a single edge drops from 3+ lines to 1.

### REF — Reference field

- [x] **REF-01**: A unit may carry an optional `reference` field whose value is a URL (external doc, runbook, ADR, etc.).
- [x] **REF-02**: A unit with a non-empty `reference` renders a visible **📖** marker so readers can see at a glance which elements have external documentation.
- [x] **REF-03**: The 📖 marker (or the node) is clickable in SVG output, linking to the reference URL via GraphViz's native `URL` attribute.
- [x] **REF-04**: The reference field renders correctly in both `-f svg` and `-f html` output; the HTML JS shim handles external `reference` URLs distinctly from internal drill-down navigation (Safari silently ignores SVG `<a>` for navigation).
- [x] **REF-05**: Units without a `reference` render exactly as before (backward-compat).

### DOC — Documentation

- [ ] **DOC-01**: README.md and skill/SKILL.md document that `type` is optional, with the full type-inference rules (default type by parent at parser.go:250; generic db/queue promotion by nesting level at parser.go:276) and a before/after example.
- [ ] **DOC-02**: README.md and skill/SKILL.md document all four new features (include, templates, ergonomics, reference) with syntax and examples.
- [ ] **DOC-03**: New example fixtures demonstrate each feature (`skill/examples/06-templates.toml`, a multi-file include example, etc.).

### XCOMP — Cross-feature integration

- [ ] **XC-01**: The pipeline ordering `include → template-expand → relative-peer-resolve → validate → generate-views → render` is enforced in code and documented; reordering is detected (e.g. via tests) as a regression.
- [ ] **XC-02**: Templates defined in an included file are visible to `[[use]]` instantiations in the including file (the "template isolation" motivating use case).
- [ ] **XC-03**: Relative peers authored inside a template resolve against the instantiation site's parent, not the template's lexical location (HS-2 from research — must be settled in discuss phase before implementation).
- [ ] **XC-04**: Humanization runs after template expansion (so it sees the substituted instantiation key, not `${name}`) and before validation (so error messages show final names).
- [ ] **XC-05**: A multi-file model using include + templates + relative peers produces the same rendered output (order-insensitive canonicalDOT comparison) as the equivalent hand-expanded single-file model.

---

## Future Requirements (deferred)

- **Compact link shorthand variants** beyond the v1.10 baseline (e.g. string-only `link = ["a", "b"]` peer-only shorthand) — re-evaluate after relative-peer + optional-name land and authoring verbosity is re-measured.
- **Template multi-output / fan-out** (`for_each`-style one-template-many-units, or array/conditional link expansion) — deliberately deferred; a template declares a fixed set of links and a single top-level unit + its declared subunit subtree (TMPL-04).
- **Parameter defaults** — every declared param is required at every instantiation (TMPL-02); trailing-default support deferred (PlantUML/go-metadot idiom rejected for v1.10 strictness).
- **Template nesting** (template instantiating another template) — deferred; one level of instantiation in v1.10.

## Out of Scope

- **Custom element kinds** — fixed 17-type enum stays (from LikeC4 comparison).
- **Tags, metadata/props** — no render target; would be dead data.
- **Icons** — contrary to current `dev/shapes-no-icons` branch direction.
- **Acronym-preserving humanization** — unsolved tar pit; dumb split ships instead.
- **Deployment model** — large scope expansion, decide on its own merits separately.
- **User-authored views + include/exclude predicates** — fights the auto-generated-view thesis; this is a different product, not a feature.
- **View-scoped styling, dynamic views, `extends`** — depend on user-authored views; moot.
- **Include namespacing / prefixing** — adds a namespacing layer the format doesn't have; complicates peer resolution. Flat merge only.
- **URL includes** — network fetch/caching/offline concerns out of scope for v1.10; local-file include only.
- **Custom color palette system** — per-element `color` + `[properties]` defaults are adequate.
- **`summary` field** (LikeC4's short-on-node label) — deferred nice-to-have.

---

## Traceability

*Phase mapping filled by roadmapper (2026-08-08). 39/39 v1.10 requirements mapped (TMPL reqs renumbered after dropping param-defaults requirement).*

| REQ-ID | Phase | Notes |
|--------|-------|-------|
| REF-01 | 28 | reference field on unit |
| REF-02 | 28 | 📖 marker render |
| REF-03 | 28 | clickable via GraphViz URL |
| REF-04 | 28 | svg + html coverage (Safari shim) |
| REF-05 | 28 | backward-compat (no reference = unchanged) |
| ERGO-03 | 29 | optional name → humanized from last path segment |
| ERGO-04 | 29 | dumb camelCase split (no acronym preservation) |
| ERGO-05 | 29 | explicit name always wins |
| ERGO-06 | 29 | AT-RISK: compact-link shorthand; research §3 flags as anti-feature — discuss-phase confirm-vs-defer-to-v2 |
| ERGO-01 | 30 | bare peer resolves against enclosing parent |
| ERGO-02 | 30 | absolute-fallback (backward-compat) |
| TMPL-01 | 31 | `[template.*]` define + named params |
| TMPL-02 | 31 | `[[use]]` instantiate; all params required (no defaults) |
| TMPL-03 | 31 | `${param}` into all string fields; fixed link count, parametrized fields |
| TMPL-04 | 31 | one template = one top-level unit + its declared subunit subtree |
| TMPL-05 | 31 | instantiated units + subunits participate fully |
| TMPL-06 | 31 | missing param = hard error (no silent literal) |
| TMPL-07 | 31 | duplicate unit path = hard error |
| TMPL-08 | 31 | deep-copy recurses into Subunits (HS-1 regression test) |
| TMPL-09 | 31 | forward references work |
| TMPL-10 | 31 | reference param substitution |
| XC-03 | 31 | relative-peer in template resolves at instantiation site (HS-2 — discuss MUST settle first) |
| XC-04 | 31 | humanization runs after expand, before validate (humanize hook lands here; end-to-end test in 33) |
| INC-01 | 32 | assemble from multiple files |
| INC-02 | 32 | path relative to including file |
| INC-03 | 32 | transitive includes |
| INC-04 | 32 | cycle detection fatal |
| INC-05 | 32 | diamond not cycle; dup = hard error |
| INC-06 | 32 | once=true dedup |
| INC-07 | 32 | flat merge; dup path = hard error |
| INC-08 | 32 | properties root-file-wins |
| INC-09 | 32 | UnitOrder concatenation preserves order |
| INC-10 | 32 | missing include = clear error |
| XC-02 | 32 | templates in included files visible to `[[use]]` (template-isolation use case) |
| DOC-01 | 33 | document type-inference |
| DOC-02 | 33 | document all four features |
| DOC-03 | 33 | example fixtures |
| XC-01 | 33 | pipeline ordering enforced + regression test |
| XC-05 | 33 | multi-file ≡ single-file golden (canonicalDOT) |
