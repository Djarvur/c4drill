# Roadmap: C4Drill

## Milestones

- ✅ **v1.0 Initial Release** — Phases 1-6 (shipped 2026-03-10)
- ✅ **v1.7 Queue Label Fix & Visual Improvements** — Phases 7-18 (shipped 2026-03-29)
- ✅ **v1.8 Proper C1/C2/C3 View Generation** — Phases 19-26 (shipped 2026-08-06) → [archive](milestones/v1.8-ROADMAP.md)
- ✅ **v1.9 C3 Boundary Node Fix** — Phase 27 (shipped 2026-08-06) → [archive](milestones/v1.9-ROADMAP.md)
- 🚧 **v1.10 Model Composition** — Phases 28-33 (in progress)

## Phases

<details>
<summary>✅ v1.9 C3 Boundary Node Fix (Phase 27) — SHIPPED 2026-08-06</summary>

- [x] Phase 27: C3 Boundary Node Fix (1/1 plan) — completed 2026-08-06

Full details: [milestones/v1.9-ROADMAP.md](milestones/v1.9-ROADMAP.md)

</details>

---

### 🚧 v1.10 Model Composition (In Progress)

**Milestone Goal:** Expand C4Drill's authoring model from a single static TOML file into a composable, parametrized, multi-file format — while preserving backward compatibility and the auto-generated-view philosophy. Four additive features form a strict runtime pipeline: `include → template-expand → relative-peer-resolve → humanize → validate → generate-views → render`.

**Granularity:** standard (6 phases)
**Coverage:** 40/40 v1.10 requirements mapped ✓
**Build/ship order:** low-risk independent work first (28, 29, 30 parallelizable); templates then include follow the runtime pipeline (31 → 32); docs + integration goldens last (33).

- [x] **Phase 28: Reference field (📖)** - Per-unit external-docs URL; renders a clickable 📖 marker via GraphViz `URL` (completed 2026-08-08)
- [ ] **Phase 29: Optional name humanization** - Omit `name`; display name derived from identifier (camelCase split); +at-risk compact-link
- [ ] **Phase 30: Relative-peer resolution** - Short `peer` names resolve against the enclosing parent; absolute-fallback preserves backward-compat
- [ ] **Phase 31: Template expansion** - `[template.*]` define + `[[use]]` instantiate parametrized units (deep-copy + `${param}`); parser BC-1 prerequisite
- [ ] **Phase 32: Include directive (multi-file)** - `[[include]]` assembles a model from multiple TOML files (merge, cycle detection, `once`)
- [ ] **Phase 33: Docs sweep + end-to-end goldens** - Document all four features; prove multi-file+templates+peers ≡ single-file

## Phase Details

### Phase 28: Reference field (📖)
**Goal**: Users can attach an external-docs URL to any unit and readers see a clickable 📖 marker in the rendered diagram.
**Depends on**: Nothing (fully independent, parallelizable with 29)
**Requirements**: REF-01, REF-02, REF-03, REF-04, REF-05
**Success Criteria** (what must be TRUE):
  1. A unit with `reference = "https://..."` shows a visible 📖 marker in the SVG output, matching the existing 🔍 collapsed-cluster affordance style (builder.go:258-261)
  2. Clicking the 📖 marker (or the node) opens the reference URL via GraphViz's native `URL` attribute — a single render-path change, near-zero cost
  3. The 📖 marker and clickable link render correctly in BOTH `-f svg` and `-f html`; the HTML JS shim routes external `reference` URLs distinctly from internal drill-down navigation (Safari silently ignores SVG `<a>`)
  4. A unit WITHOUT a `reference` field renders exactly as before — byte-identical to v1.9 output (backward-compat hard contract)
  5. Any unit type (system, container, db, queue, person, box) accepts the optional `reference` field
**Notes**: `isBuiltinField` (parser.go:309) gets the one safe single-line addition for `reference`. No `captureDefinitionOrder` change needed (reference is a leaf field, not a reserved table). This is the cheapest table-stakes item per research §3.
**Plans**: 1 plan

Plans:
- [x] 28-01-PLAN.md — Add `Unit.Reference` field + isBuiltinField entry; render 📖 glyph via buildNode/buildClusterLabel; wire external URL through Node.ReferenceURL + cn.SetURL; branch htmlNavShim for external http(s) vs internal drill-down; backward-compat canonical-DOT golden; README/SKILL docs + example fixtures

### Phase 29: Optional name humanization
**Goal**: Users can omit the `name` field and get a readable display name derived from the unit's identifier, reducing boilerplate.
**Depends on**: Nothing (parallelizable with 28)
**Requirements**: ERGO-03, ERGO-04, ERGO-05, ERGO-06
**Success Criteria** (what must be TRUE):
  1. A unit defined as `[linuxSystem.localIDP]` with no `name` displays as "Local IDP" — derived from the last path segment of the identifier
  2. Humanization is a dumb camelCase split — `sessionManager` → "Session Manager", `gRPC` → "Grpc"; acronym preservation is explicitly out of scope (Terraform `title()` proves it unsolved), authors escape via explicit `name =`
  3. An explicit `name = "..."` always wins over humanization — every existing model renders identically to v1.9 (backward-compat hard contract)
  4. *(AT-RISK — ERGO-06)* A common single edge can be written on one line instead of the multi-line array-of-tables form. **Research SUMMARY §3 flags compact-link shorthand as a v1.10 anti-feature; the discuss phase must decide confirm-in-v1.10 vs defer-to-v2.** If deferred, ERGO-06 moves to Future Requirements and this criterion drops.
**Notes**: Humanization runs AFTER template expansion and BEFORE validation (XC-04 — enforced in Phase 31 when the humanize hook relocates to a post-expansion pass; Phase 29 ships a parse-time fallback since templates do not exist yet, and `model.Humanize` is the stable artifact Phase 31 reuses). The research §6 HU-1 acronym-allowlist design fork is **resolved against the allowlist** by ERGO-04 (dumb split, no acronyms) — see 29-CONTEXT.md D-01. **ERGO-06 (compact-link shorthand) is deferred to Future Requirements** per research SUMMARY §3 (classified a v1.10 anti-feature) and the original todo's sequencing ("defer #3; re-evaluate after #1+#2"); success criterion 4 drops.
**Plans**:
- `29-01` (wave 1, TDD): `model.Humanize` + parse-time `Unit.Name` fallback in `parseUnitWithOrder`; covers ERGO-03/04/05. Zero new deps; validator/view/render untouched.
- `29-02` (wave 2, depends on 29-01): README + skill/SKILL.md docs for optional `name` + humanize rules + acronym escape hatch.
- **Wave 2** *(blocked on Wave 1 completion)*

### Phase 30: Relative-peer resolution
**Goal**: Users can write short `peer` values that resolve against the enclosing parent block, eliminating repetitive absolute paths.
**Depends on**: Nothing (ships as a no-op pass for absolute-only models)
**Requirements**: ERGO-01, ERGO-02
**Success Criteria** (what must be TRUE):
  1. Inside `[linuxSystem.localIDP]`, a link with `peer = "sessionManager"` resolves to `linuxSystem.localIDP.sessionManager` — the sibling of the enclosing block
  2. A `peer` that contains a `.`, OR exactly matches a top-level unit path, OR does not resolve as a relative sibling, falls back to absolute resolution — every existing model parses identically to v1.9 (a corpus test asserting the (source, resolved-peer) set is byte-identical to today is the acceptance criterion)
  3. Ambiguity at a single nesting depth (two siblings matching the same bare name) is a hard error, not a silent first-match — "nearest wins" is for nesting depth, not ties at the same depth
**Notes**: Implemented as a separate post-parse pass that rewrites `Link.Peer` in place on the assembled model before `BuildIndex`/validation, so the validator's existing absolute-path logic is untouched. Gate for relative-search: bare name + no `.` + not a top-level key + not already in the index (research §6 RP-2 / §9 RP-2). HS-2 (relative-peer inside a TEMPLATE) is settled jointly with Phase 31 in discuss — the resolution-site decision (instantiation-parent, not template-parent) must be pinned before implementation.
**Plans**: TBD

### Phase 31: Template expansion
**Goal**: Users define a parametrized unit once and instantiate it N times with different parameters, eliminating copy-paste of near-identical units.
**Depends on**: Benefits from Phase 30 (templated links may be relative; HS-2 resolution site shared). Ships with absolute peers if 30 slips.
**Requirements**: TMPL-01, TMPL-02, TMPL-03, TMPL-04, TMPL-05, TMPL-06, TMPL-07, TMPL-08, TMPL-09, TMPL-10, XC-03, XC-04
**Success Criteria** (what must be TRUE):
  1. A `[template.svc]` table with declared named params can be instantiated via a `[[use]]` directive supplying concrete values for ALL params (no defaults — every declared param is required), producing a concrete unit (+ its declared subunit subtree) that passes validation, appears in auto-generated C1/C2/C3 views, and renders identically to a hand-authored unit
  2. `${param}` substitution applies to ALL string fields of the instantiated unit AND its subunits — Name, Description, Technology, Reference, Color, and every Link field (Peer, Description, Technology). The template declares a FIXED set of links; each instantiation fills their fields from params (same template → units linked differently, but always the same link count). No array/conditional fan-out. Includes `reference = "https://wiki/${name}"`.
  3. Instantiating the same template 3× with distinct params yields 3 INDEPENDENT unit subtrees — the 2nd never inherits the 1st's name/links/LinksFrom-mirror-links (HS-1 regression test: re-running expansion is idempotent; after a full validate pass two instantiations' LinksFrom slices are disjoint). Deep-copy recurses into Subunits.
  4. Errors are clear and point to cause: a missing param names the template + parameter + instantiation site (no silent `${param}` literals — deliberate divergence from go-metadot/PlantUML); two `[[use]]` blocks producing the same unit path name both instantiation sites (no silent overwrite)
  5. Forward references work — a `[[use]]` may appear textually before its `[template.<name>]` definition in the same file (free under structured post-parse). Relative peers authored in a template resolve against the instantiation site's parent, not the template's lexical location (XC-03 / HS-2). Humanization runs AFTER expansion so the substituted instantiation key is what gets humanized, and BEFORE validation so error messages show final names (XC-04).
**Notes**:
- **DISCUSS-PHASE BLOCKER (design forks — MUST settle before planning):** forward-reference policy (TM-2/TM-09), unresolved-`${param}` strictness (TM-5/TMPL-06), relative-peer resolution site for template-authored links (HS-2/XC-03 — the single most consequential decision for the milestone, recommended: instantiation-site parent).
- **Parser prerequisite (BC-1) lands as Plan 1 of this phase** — two coordinated changes so the new reserved top-level tables don't become phantom units: extend `captureDefinitionOrder` skip (parser.go:100/:128) to also skip `[template.*]`, `[[use]]`, `[[include]]` (added for include here even though the feature lands in 32, since it's one skip-rule path), AND extend `Parse` rawMap extraction before the unit loop (mirroring properties extraction at parser.go:68-77). Land as a small well-tested change before feature code.
- **HS-1 (deep-copy aliasing) is THE core correctness concern** — hand-rolled recursive `Unit.Clone()` (~15 LOC) in package model; the validator mutates LinksFrom in place (index.go:70-81) so a shallow copy corrupts the Nth instantiation. Three-instantiation regression test is required, not optional. Clone MUST recurse into Subunits since templates may declare subunit subtrees (TMPL-04).
**Plans**: TBD

### Phase 32: Include directive (multi-file)
**Goal**: Users assemble a single diagram from multiple TOML files, isolating template libraries into their own files and splitting large models across files.
**Depends on**: Phase 31 — the "template isolation" motivating use case is primary; include must carry `Templates` through the merge so templates defined in an included file are visible to `[[use]]` in the including file (XC-02).
**Requirements**: INC-01, INC-02, INC-03, INC-04, INC-05, INC-06, INC-07, INC-08, INC-09, INC-10, XC-02
**Success Criteria** (what must be TRUE):
  1. An entry file with `[[include]]` directives plus all transitively-included files merge into one logical model that renders identically to a single-file equivalent (an included file may itself contain `[[include]]` directives, resolved recursively)
  2. Include paths resolve relative to the INCLUDING file's directory (not the CLI working directory), so the model renders identically regardless of where `c4drill` is invoked — paths canonicalized via `filepath.Abs` + `filepath.Clean`
  3. A direct or transitive include cycle (A→B→A) produces a clear FATAL error naming the cycle (stack-based detection, max-depth cap as defense-in-depth); a diamond (A→B→D, A→C→D) WITHOUT `once` is legal but hard-errors on the resulting duplicate unit path; `once = true` skips re-inclusion of an already-included file (PlantUML `!include_once`, visited-set dedup)
  4. Merge conflicts produce clear errors: a unit path defined in two files (after `once` dedup) names both files (flat merge, no namespacing); a conflicting `[properties]` from an included file is a hard error (root-file-wins); a missing include file names the referenced path and the including file (not a generic parse failure); `UnitOrder` concatenation preserves authoring order (entry file's units first, then each included file's units appended in include-directive order)
  5. Templates defined in an included file are visible to `[[use]]` instantiations in the including file (XC-02 — the "template isolation" motivating use case)
**Notes**:
- **DISCUSS-PHASE BLOCKER (design forks — MUST settle before planning):** merge semantics for `UnitOrder`/`SubunitOrder` across files (IN-2/INC-09), diamond-include behavior without `once` (IN-3/INC-05), directive-table naming + reserved-word collision policy for legacy units named `use` (BC-2 — bare vs namespaced; backward-compat non-negotiable).
- `captureDefinitionOrder` skip for `[[include]]` already landed in Phase 31 Plan 1 (shared parser change). This phase adds `Model.Includes` field + `internal/include/Resolve` (recursive merge, cycle detection via stack, `Once` via visited-set).
**Plans**:
- `32-01` — IncludeDirective extraction (parser-side): `IncludeDirective` type + `Model.Includes` field + `[[include]]` rawMap extraction in `Parse` (consumes the Phase 31 BC-1 skip; does NOT re-touch `captureDefinitionOrder`). Wave 1.
- `32-02` — `internal/include` package + pipeline wiring: `resolve.go` (recursive Resolve with cycle stack, once/diamond visited-set, missing-file hard-error) + `merge.go` (per-field struct-union per D-09/D-10/D-11/INC-08) + comprehensive tests + `cmd/c4drill/root.go` Stage 1a insertion. Wave 2 *(blocked on Wave 1 completion)*.

### Phase 33: Docs sweep + end-to-end goldens
**Goal**: All four features are documented with runnable examples and proven to compose correctly end-to-end.
**Depends on**: Phases 30-32 (all feature work done)
**Requirements**: DOC-01, DOC-02, DOC-03, XC-01, XC-05
**Success Criteria** (what must be TRUE):
  1. README.md and skill/SKILL.md document that `type` is optional, with the full type-inference rules (default type by parent at parser.go:250; generic db/queue promotion by nesting level at parser.go:276) and a before/after example
  2. README.md and skill/SKILL.md document all four features (include, templates, ergonomics, reference) with syntax and examples; new example fixtures demonstrate each (`skill/examples/06-templates.toml`, a multi-file include example, etc.)
  3. A multi-file model using include + templates + relative peers produces the SAME rendered output (order-insensitive canonicalDOT comparison — sort-normalize, strip layout geometry per STATE.md DI-1) as the equivalent hand-expanded single-file model
  4. The pipeline ordering `include → template-expand → relative-peer-resolve → humanize → validate → generate-views → render` is enforced in code and a test detects reordering as a regression (XC-01)
**Notes**: Hosts the cross-cutting integration tests whose feature code landed in earlier phases — XC-02's end-to-end test (include+template), XC-04's humanize-after-expand regression, XC-05's full multi-file golden. All multi-file/template goldens MUST use the order-insensitive canonicalDOT comparator, NOT byte-exact `require.Equal` (go-graphviz layout nondeterminism + an added ordering axis from multi-file/templates).
**Plans**: TBD

## Progress

**Execution Order:**
Phases 28, 29, 30 are independent and parallelizable. Phases 31 → 32 are sequential (pipeline + template-isolation use case). Phase 33 last (integration). Within 31, the BC-1 parser change (Plan 1) must land before template feature code.

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 28. Reference field | v1.10 | 1/1 | Complete    | 2026-08-08 |
| 29. Optional name humanization | v1.10 | 0/TBD | Not started | - |
| 30. Relative-peer resolution | v1.10 | 0/TBD | Not started | - |
| 31. Template expansion | v1.10 | 0/TBD | Not started | - |
| 32. Include directive | v1.10 | 0/TBD | Not started | - |
| 33. Docs sweep + goldens | v1.10 | 0/TBD | Not started | - |
