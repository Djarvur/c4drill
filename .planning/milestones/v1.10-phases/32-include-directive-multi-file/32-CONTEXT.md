# Phase 32: Include directive (multi-file) - Context

**Gathered:** 2026-08-08
**Status:** Ready for planning

## Phase Boundary

Users assemble a single diagram from multiple TOML files via `[[include]]` directives. Included files merge into one logical model (flat, no namespacing), with recursive transitive includes, cycle detection, `once=true` dedup, relative-to-including-file path resolution, and clear conflict errors. The primary motivating use case is template isolation (template libraries in their own file, included by the models that use them, per XC-02) and large-model splitting (one file per domain).

This phase delivers INC-01 through INC-10 plus XC-02 (templates in included files visible to `[[use]]`). It depends on Phase 31 (template expansion) — the BC-1 parser skip-rule for `[[include]]` already lands in Phase 31 Plan 1, so this phase only adds `Model.Includes` + `internal/include/Resolve`.

## Implementation Decisions

### Include ordering (UnitOrder assembly)

- **D-09:** The `[[include]]` directive is an **import, not a splice**. The entry file's top-level units come first in their authored order; each included file's top-level units append in include-directive order. Placement of `[[include]]` directives within the entry file does NOT affect ordering (an author typically groups them at top or bottom). Aligns with research SUMMARY §6 IN-2 recommendation and the module-import idiom (Terraform, Go imports).
  - *Note: this was initially discussed as splice-in-place (C #include / PlantUML !include) and the user revised to append after reflection. The append model is simpler to reason about — the entry file is the "main" and includes are additive libraries.*
- **D-10 (cross-file subunits):** An included file may define subunits under a top-level unit declared in the entry file (or another include). E.g. `main.toml` has `[linuxSystem]`; `auth.toml` has `[linuxSystem.auth]` and `[linuxSystem.db]`. The merge appends these subunits to the existing `linuxSystem`, preserving the include-file's subunit order. Subunits concatenate onto the parent regardless of which file the parent or the subunit lives in. This is essential for the "one file per domain, all under one system" decomposition pattern.

### Diamond-include behavior

- **D-11:** **Dedup same-file diamonds; hard-error cross-file collisions.** This refines INC-05 (which specified strict-error unconditionally).
  - If the SAME file is reached via two include paths (true diamond, e.g. A→B→D and A→C→D where both reach D), its units are byte-identical by construction (same file on disk) — silently dedup them, no error.
  - If two DIFFERENT files define the same unit path (e.g. `auth.toml` and `billing.toml` both define `[mailAdapter]`), hard-error naming both files — a genuine name collision the author must resolve.
  - `once = true` (INC-06) remains the explicit opt-in visited-set dedup; the same-file automatic dedup is complementary, not a replacement. Distinction: cycle detection uses a **stack** (current chain); `once` dedup uses a **global visited-set**; same-file diamond dedup is automatic when both inclusion paths resolve to the identical canonical path (`filepath.Abs` + `filepath.Clean`).

### Missing-include behavior

- **D-12:** **Hard-error unconditionally, no `optional` flag.** Any missing include file is a fatal error naming the referenced path and the including file (INC-10 as written). No `optional = true` escape hatch. Consistent with the milestone's hard-error-everywhere / strictness-is-a-feature stance (same rationale as no-param-defaults in TMPL-02, no-silent-literals in TMPL-06). A missing file is almost always a typo, uncommitted file, or wrong working directory — better to fail loudly than render an incomplete model silently. The local-override / per-developer layering use case (the main motivator for `optional`) is rare at C4Drill's current scale.

### Already-locked decisions (carried from Phase 31 / REQUIREMENTS, NOT re-discussed)

These were settled before this discuss session — do NOT re-litigate:
- **D-06/D-07 from Phase 31 (spans both phases):** Bare `[[include]]` table name (no `c4drill.` namespace prefix). No collision-mitigation machinery — C4Drill is not public, no legacy models to break. Parser treats `include` as a reserved top-level key, full stop.
- **D-08 BC-1:** The `captureDefinitionOrder` skip-rule for `[[include]]` lands in Phase 31 Plan 1 (shared parser change covering `template`/`use`/`include`). Phase 32 does NOT re-touch `captureDefinitionOrder` — it consumes that skip. Phase 32 adds: `Model.Includes` field (`[]IncludeDirective`, `toml:"-"`) extracted in `Parse` from rawMap, + `internal/include/Resolve` (recursive merge, cycle detection via stack, `once` via visited-set, same-file diamond dedup per D-11).
- **INC-02 (path resolution):** Include paths resolve relative to the INCLUDING file's directory (not CLI cwd), canonicalized via `filepath.Abs` + `filepath.Clean`. Universal convention; prevents "works on my machine, breaks in CI."
- **INC-03 (transitive):** An included file may itself contain `[[include]]` directives, resolved recursively.
- **INC-04 (cycle detection):** A direct or transitive include cycle (A→B→A) is a fatal error naming the cycle. Stack-based detection; max-depth cap (e.g. 100) as defense-in-depth.
- **INC-07 (flat merge):** No namespacing/prefixing — included units merge into one namespace; a unit path defined in two DIFFERENT files (after dedup) is a hard error naming both files.
- **INC-08 (properties root-wins):** The entry file's `name`/`description` are authoritative; conflicting `[properties]` from an included file is a hard error.
- **Pipeline ordering:** include runs **first** (`include → template-expand → relative-peer-resolve → humanize → validate → generate-views → render`). Include must carry `Templates`/`Instantiations` through the merge so Phase 31's expansion sees templates from included files (XC-02).

### Claude's Discretion

- Exact struct shape for `IncludeDirective` (likely `Path string`, `Once bool` — minimal; planner may refine if more fields emerge).
- Internal package structure for the resolver (research suggests `internal/include/Resolve`; planner may choose `internal/compose/` or a function in `internal/parser/`).
- Whether `properties` conflict detection compares the full `Properties` struct or just `name`/`description` (INC-08 names those two; other fields like `color`/`edges` defaults could be first-wins or conflict-error — planner's call, lean conflict-error for safety).

## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project planning (load-bearing context)
- `.planning/REQUIREMENTS.md` — INC-01..10, XC-02 (Phase 32 requirements, fully specified). §Traceability maps each REQ to Phase 32.
- `.planning/ROADMAP.md` — Phase 32 section: goal, requirements, 5 success criteria, discuss-phase blocker notes (now resolved by D-09/D-11/D-12).
- `.planning/todos/pending/2026-08-08-include-directive-multi-file-diagrams.md` — the original feature intent with the full design space, go-metadot (`metadot.pl:839` `registerInclude`, `@incChain` cycle detection) + PlantUML (`!include`/`!include_once`/`!includesub`) reference mechanics, candidate directive syntaxes, and merge-semantics table.
- `.planning/phases/31-template-expansion/31-CONTEXT.md` — D-06/D-07 (bare reserved names, no collision machinery) and D-08 (BC-1 parser skip-rule for `[[include]]` lands in Phase 31 Plan 1). Phase 32 consumes both.

### Research (load-bearing conclusions)
- `.planning/research/SUMMARY.md` — §3 (include table stakes: directive, transitive, relative-to-including-file, cycle fatal, hard-error-on-dup; `once=true` essential day-one); §4 (pipeline + Model-extension approach + unchanged consumers); §6 IN-2/IN-3 design forks (now settled: D-09 append, D-11 same-file dedup); §8 phase breakdown (Phase 32); §9 watch-outs (IN-3 diamond-not-cycle, IN-5 relative paths, BC-3 non-strict toml stays).
- `.planning/research/ARCHITECTURE-v1.10.md` — file:line integration points: pipeline insertion (include.Resolve is the FIRST pre-processing pass, before template.Expand); `captureDefinitionOrder` skip-rule already lands in Phase 31; `Model.Includes` field + `internal/include/Resolve` are the new code; validator/view/render unchanged (consume the merged `*parser.Model`).
- `.planning/research/PITFALLS.md` — IN-3 (diamond-not-cycle, stack vs visited-set distinction — now encoded in D-11); IN-5 (relative path resolution); BC-3 (non-strict toml load-bearing, do NOT enable DisallowUnknownFields); the canonicalDOT golden comparison requirement for multi-file models.

### Code (integration points — confirmed live)
- `cmd/c4drill/root.go:112` (ParseFile) — `include.Resolve` is the new FIRST call after ParseFile, before template.Expand and Validate (`:118`). Signature: `func Resolve(m *parser.Model) (*parser.Model, error)`.
- `internal/parser/parser.go:47` (`Parse`), `:100` (`captureDefinitionOrder`), `:128` (properties-skip — the BC-1 site, already extended for `include` in Phase 31 Plan 1) — Phase 32 adds `Model.Includes` extraction in `Parse` from rawMap.
- `internal/parser/parser.go:35` (`Model` struct) — add `Includes []IncludeDirective` (`toml:"-"`).

## Existing Code Insights

### Reusable Assets
- The existing `properties` extraction pattern (`parser.go:68-77`: pull a known top-level key from `rawMap` before the unit loop) — directly reused for extracting `include` directives (D-08, lands in Phase 31).
- `filepath.Abs` + `filepath.Clean` (stdlib) — for canonicalizing include paths (cycle detection, same-file diamond dedup, `once` visited-set). Idiomatic Go; no new dep.
- Phase 31's `captureDefinitionOrder` skip-rule extension (D-08) — already covers `[[include]]`; Phase 32 consumes it without re-touching the parser skip.

### Established Patterns
- **Order preservation is load-bearing** — `captureDefinitionOrder` exists so rendering follows authoring order. Per D-09 (append), the merge concatenates already-ordered `UnitOrder` slices; it does NOT re-run order capture on the merged model.
- **Validator is the single gatekeeper** (STATE.md D-12) — `include.Resolve` must produce a `*parser.Model` whose `.Units`/`.UnitOrder`/`.Properties` are indistinguishable from a hand-authored single-file model, so `validator.Validate` consumes it unchanged.
- **Non-strict TOML unmarshaling is load-bearing** (PITFALLS BC-3) — do NOT enable `DisallowUnknownFields()`.
- **CanonicalDOT for multi-file golden tests** (STATE.md DI-1) — multi-file goldens MUST use order-insensitive comparison (sort-normalize, strip layout geometry), NOT byte-exact `require.Equal`. Multi-file adds another axis of ordering variance on top of go-graphviz nondeterminism.

### Integration Points
- **Pipeline insertion** (`cmd/c4drill/root.go`): `include.Resolve` is the new FIRST call between `ParseFile` (`:112`) and `template.Expand`/`Validate` (`:118`). Must carry `Templates`/`Instantiations` through the merge (XC-02).
- **Model extension** (`internal/parser/parser.go:35`): add `Includes []IncludeDirective` (`toml:"-"`), extracted in `Parse`.
- **Merge semantics** — union `Units` maps (conflict = hard-error per D-11), concatenate `UnitOrder` (append per D-09), merge `Subunits` cross-file (D-10), root-wins `Properties` (INC-08), union `Templates`/`Instantiations` (conflict = hard-error).

## Specific Ideas

- Directive syntax (research lean, consistent with the bare-names decision D-06):
  ```toml
  [[include]]
  path = "templates/common.toml"
  once = true   # include_once semantics — safe to re-include from multiple model files

  [[include]]
  path = "domains/auth.toml"
  ```
- The motivating end-to-end test (XC-02, also lands in Phase 33): `templates.toml` defines a template; `model.toml` does `[[include]] path="templates.toml" once=true` then `[[use]]` instantiates it — produces the same output as a single-file model with the template inline.

## Deferred Ideas

- **`optional = true` for missing-file skip** — explicitly rejected for v1.10 (D-12, strictness). Revisit if the local-override / per-developer layering use case becomes common.
- **Include namespacing / prefixing** — anti-feature (Out of Scope per REQUIREMENTS); flat merge only.
- **URL includes** — anti-feature (Out of Scope per REQUIREMENTS); local-file include only.
- **Splice-in-place ordering** — initially discussed, user revised to append (D-09). Could revisit if a "split a big file and preserve exact order" use case emerges, but append is locked for v1.10.

### Reviewed Todos (not folded)
- `2026-08-08-add-reference-field-to-units.md` — assigned to Phase 28 (`resolves_phase: 28`), already planned.
- `2026-08-08-toml-authoring-ergonomic-improvements.md` — assigned to Phases 29/30.
- `2026-08-08-unit-templates-parametrized-definitions.md` — assigned to Phase 31 (`resolves_phase: 31`), context gathered, currently being planned.
- `2026-08-08-document-type-inference-omittable.md` — assigned to Phase 33 (docs sweep).

---

*Phase: 32-Include directive (multi-file)*
*Context gathered: 2026-08-08*
