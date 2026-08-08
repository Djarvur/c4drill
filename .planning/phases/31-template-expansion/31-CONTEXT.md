# Phase 31: Template expansion - Context

**Gathered:** 2026-08-08
**Status:** Ready for planning

## Phase Boundary

Users define a parametrized unit template once in a `[template.<name>]` table (which may declare a subunit subtree and a fixed set of links), then instantiate it N times via `[[use]]` directives supplying all-required named parameters. Each instantiation produces one independent concrete unit subtree inserted into the model at a specified parent. Instantiated units behave identically to hand-authored units — they pass validation, appear in auto-generated C1/C2/C3 views, and render the same way.

This phase delivers TMPL-01 through TMPL-10, plus the cross-cutting XC-03 (relative-peer resolution site) and XC-04 (humanization ordering). It also lands the parser prerequisite BC-1 (reserved-table skip) as Plan 1, which Phase 32 (include) depends on.

## Implementation Decisions

### Relative-peer resolution site in templates (HS-2)

- **D-01 (HS-2):** Relative peers authored inside a template resolve against the **instantiation site's** structural parent — NOT the template's lexical location. Rationale: templates are data, not structural location; the instantiation is what gives the produced unit its place in the hierarchy. Template-site resolution would either fail or fabricate silent cross-system edges (the HS-2 corruption case: `peer="cache"` in a template instantiated under both `mainSystem` and `otherSystem` would wrongly resolve both to the same global `cache`).
- **D-02 (HS-2 edge case):** A template instantiated with NO parent (top-level) resolves relative peers against top-level siblings with absolute-fallback — uniform with hand-authored top-level units. No template-special-case logic; the Phase 30 relative-peer pass runs AFTER expansion and treats every unit identically.

### Instantiation syntax & unit placement

- **D-03:** Instantiation uses a **separate `[[use]]` array-of-tables** (Syntax 1), NOT inline-on-unit. Each `[[use]]` carries: `template` (the template name), `parent` (placement, optional — empty/omitted = top-level), and the template's named params as explicit fields (including `name` as a regular param). Rationale: fully explicit, no implicit param injection, `[[use]]` tables never collide with real units, `parent=` is the only placement knob.
- **D-04 (name param semantics):** The `name` param fills the **top-level produced unit's last path segment**. The produced unit's full path = `parent + "." + name` (or just `name` if parent is empty). Subunit keys in the template stay **verbatim** — they are NOT parametrized. Parametrization is field-value-only (substitution into string fields); structural slots (which entities exist, their keys) are never parametrized. This preserves the multi-output / `for_each` anti-feature boundary.
- **D-05 (duplicate-path check, TMPL-07):** After all `[[use]]` expansions, run a **full-path-set check** across hand-authored units + all expanded template instances. ANY collision — templated-vs-templated OR templated-vs-hand-authored — is a single hard error naming both sources (file + location of the `[[use]]` and the conflicting definition). No silent overwrite.

### Reserved-table naming & legacy-collision (BC-1/BC-2)

- **D-06:** Use **bare, ergonomic table names**: `[template.<name>]`, `[[use]]`, and `[[include]]` (the last lands in Phase 32 but is reserved now). No `c4drill.` namespace prefix.
- **D-07 (BC-2 — collision policy):** **No collision-mitigation machinery is needed.** C4Drill is not yet public and there are no existing models to break. The parser treats `template`, `use`, and `include` as reserved top-level keys, full stop — no detection of "legacy unit named `use`", no migration errors, no silent-allow ambiguity. (If/when C4Drill becomes public, revisit; for v1.10 the clean implementation wins.)
- **D-08 (BC-1 — parser changes):** Two coordinated changes land as Plan 1 of this phase, both required before feature code:
  1. **`captureDefinitionOrder`** (`internal/parser/parser.go:100`, skip-rule at `:128`): extend the skip (currently only `properties`) to also skip `[template.*]` (any table whose first key segment is `template`), `[[use]]`, and `[[include]]`. Without this, `[template.microservice]` registers `microservice` as a subunit of phantom parent `template`.
  2. **`Parse` rawMap extraction** (`internal/parser/parser.go:47`): extract these reserved top-level tables from `rawMap` before the unit-processing loop (mirroring the existing `properties` extraction at `parser.go:68-77`), routing them into the new `Model.Templates` / `Model.Instantiations` fields.
  - Note: `isBuiltinField` (`parser.go:309`) does NOT need `template`/`use`/`include` added — those are top-level table names, not unit fields. It only needs `"reference"` added (a leaf field) for Phase 28.

### Already-locked decisions (carried from milestone setup, NOT re-discussed)

These were settled before this discuss session and are listed for downstream-agent awareness — do NOT re-litigate:
- **No param defaults** — every declared param required at every instantiation; missing any = hard error (TMPL-02, TMPL-06).
- **One template = one top-level unit + its declared subunit subtree** (TMPL-04).
- **Fixed link count, parametrized fields** — no fan-out / `for_each` / array expansion (TMPL-03).
- **Forward references allowed** — `[[use]]` may appear textually before `[template.*]` definition (TMPL-09); structured post-parse makes this free.
- **HS-1 deep-copy** — hand-rolled recursive `Unit.Clone()` (~15 LOC) in package model; MUST recurse into `Subunits` (every `*Unit` cloned) and preserve the unexported `Link.Mirror` field (`internal/model/link.go:67`, load-bearing for validator multiplicity at `internal/validator/index.go:70-81`). No reflection/gob/json copier (they silently drop `Mirror`). Three-instantiation regression test required (idempotent re-expand + disjoint `LinksFrom` after validate).
- **Substitution mechanism** — `strings.NewReplacer` over declared params, applied to all string fields of the unit + subunits + links (Name, Description, Technology, Reference, Color, Link.Peer, Link.Description, Link.Technology). No `text/template`.
- **Pipeline ordering** — `include → template-expand → relative-peer-resolve → humanize → validate → generate-views → render`. Template expansion is the 2nd pass.

### Claude's Discretion

- Exact field-name and TOML-shape details for `TemplateDef` / `Instantiation` structs on `parser.Model` (the research recommends dedicated `toml:"-"` fields — `Templates map[string]*TemplateDef`, `Instantiations []Instantiation` — extracted from rawMap; planner may refine).
- Internal package structure for the expansion pass (e.g. `internal/template/Expand` vs. a function in `internal/parser/`).
- Whether to deep-copy-then-substitute or substitute-in-place-on-copy (ordering of the two operations within `Expand`).

## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project planning (load-bearing context)
- `.planning/REQUIREMENTS.md` — TMPL-01..10, XC-03, XC-04 (Phase 31 requirements, fully specified). §Traceability maps each REQ to Phase 31.
- `.planning/ROADMAP.md` — Phase 31 section: goal, requirements, 5 success criteria, discuss-phase blocker notes, BC-1 prerequisite.
- `.planning/todos/pending/2026-08-08-unit-templates-parametrized-definitions.md` — the original feature intent with the full design space, go-metadot + PlantUML reference mechanics, param-semantics table (settled), and verification section.

### Research (load-bearing conclusions)
- `.planning/research/SUMMARY.md` — §2 (zero-deps; hand-rolled `Unit.Clone()` rationale re `Link.Mirror`), §4 (pipeline + Model-extension approach + unchanged consumers), §5 (HS-1 deep-copy aliasing with validator-mutates-in-place detail), §6 design forks (most now settled above), §8 phase breakdown, §9 watch-outs.
- `.planning/research/ARCHITECTURE-v1.10.md` — file:line integration points: pipeline insertion in `cmd/c4drill/root.go` between Parse (`:112`) and Validate (`:118`); Model-extension approach; `captureDefinitionOrder` skip rule (`parser.go:128`); template expansion drain into `Units`/`UnitOrder`.
- `.planning/research/STACK.md` — why hand-rolled deep-copy (reflection libs silently skip unexported `Mirror`), why `strings.NewReplacer` over `text/template`.
- `.planning/research/PITFALLS.md` — HS-1 (deep-copy aliasing, `index.go:70-81` mutation), HS-2 (resolution-site ambiguity, now settled D-01), BC-1 (parser reserves keys in two places), TM-1/TM-3/TM-5 (path collision, recursion cap, unresolved-param strictness).

### Code (integration points — confirmed live)
- `internal/parser/parser.go:47` (`Parse`), `:100` (`captureDefinitionOrder`), `:128` (properties-skip — the BC-1 site), `:309` (`isBuiltinField`) — parser changes land here.
- `internal/model/unit.go:41-72` (`Unit` struct: `Subunits map[string]*Unit` at `:71`, `Links`/`LinksFrom` at `:65`/`:67`, `Expanded` at `:63`) — `Clone()` method lives here.
- `internal/model/link.go:67` (`Mirror` field, `toml:"-"`) — MUST survive deep-copy.
- `internal/validator/index.go:70-81` (`populateIncomingLinks` mutates `LinksFrom` in place) — the HS-1 root cause.
- `cmd/c4drill/root.go:112` (ParseFile) → `:118` (Validate) — the new expansion pass inserts between these.

## Existing Code Insights

### Reusable Assets
- `slices.Clone` (stdlib, already used at `parser.go:310` for the `isBuiltinField` allowlist) — project idiom for shallow slice copy; the recursive `Unit.Clone()` builds on it for the per-element struct copy.
- The existing `properties` extraction pattern (`parser.go:68-77`: pull a known top-level key from `rawMap` before the unit loop) — directly reused for extracting `template`/`use` tables (D-08).
- `captureDefinitionOrder`'s `unstable.Parser` walk (`parser.go:107-150`) — the skip-rule extension (D-08) is a one-condition addition to the existing `properties` skip at `:128`.

### Established Patterns
- **Order preservation is load-bearing** — `captureDefinitionOrder` exists precisely so rendering follows authoring order. Expanded template instances must append to `UnitOrder` at the `[[use]]` site's position (or in `[[use]]` order), NOT alphabetically.
- **Validator is the single gatekeeper** (STATE.md D-12) — the expansion pass must produce a `*parser.Model` whose `.Units` is indistinguishable from a hand-authored one, so `validator.Validate` consumes it unchanged.
- **Non-strict TOML unmarshaling is load-bearing** (PITFALLS BC-3) — `toml:",inline"` subunit trick (`unit.go:71`) depends on unknown keys being accepted. Do NOT enable `DisallowUnknownFields()`. The new reserved tables are extracted explicitly, not via strict mode.

### Integration Points
- **Pipeline insertion** (`cmd/c4drill/root.go`): the expansion pass is a new call between `ParseFile` (`:112`) and `Validate` (`:118`), taking and returning `*parser.Model`. Signature: `func Expand(m *parser.Model) (*parser.Model, error)`.
- **Model extension** (`internal/parser/parser.go:35` `Model` struct): add `Templates map[string]*TemplateDef` and `Instantiations []Instantiation`, both `toml:"-"`, extracted in `Parse`. These fields are NOT read by validator/view/render (confirmed in ARCHITECTURE-v1.10.md §unchanged-consumers).
- **Substitution target fields** — every string field on `Unit` and `Link` (see "Substitution mechanism" above).

## Specific Ideas

- The instantiation syntax sketch (research lean, now confirmed D-03):
  ```toml
  [template.microservice]
  params = ["name", "domain", "tech"]
  name = "${name} Service"
  type = "container"
  technology = "${tech}"
  description = "${name} handles ${domain}"
  [[template.microservice.link]]
  peer = "${upstreamBus}"
  description = "Publishes ${domain} events"

  [[use]]
  template = "microservice"
  parent = "linuxSystem"
  name = "auth"
  domain = "authentication"
  tech = "Go, gRPC"
  upstreamBus = "messageBus"
  ```
- A template may declare a subunit subtree (`[template.svc]` + `[template.svc.api]` + `[template.svc.db]`); one `[[use]]` produces the whole subtree rooted at `parent.name`, subunit keys verbatim (D-04).

## Deferred Ideas

- **Parameter defaults** (trailing-default support) — explicitly out for v1.10 (strictness); revisit if authoring friction emerges. Captured in REQUIREMENTS.md Future.
- **Template multi-output / fan-out / `for_each`** — anti-feature for v1.10; a template declares a fixed unit subtree and fixed link count.
- **Template nesting** (template instantiating another template) — deferred; one level of instantiation in v1.10.
- **Structural parametrization** (parametrized subunit keys/paths, array/conditional link expansion) — would cross into multi-output territory; deliberately out.

### Reviewed Todos (not folded)
- `2026-08-08-add-reference-field-to-units.md` — matched phase 31 fuzzily but is assigned to Phase 28 (`resolves_phase: 28`). The templates phase only needs to ensure substitution covers the `reference` field (TMPL-10), which is captured above.
- `2026-08-08-include-directive-multi-file-diagrams.md` — assigned to Phase 32 (`resolves_phase: 32`). Phase 31 only lands the `[[include]]` parser skip-rule as part of BC-1 (D-08); the include feature itself is Phase 32.
- `2026-08-08-toml-authoring-ergonomic-improvements.md` — assigned to Phases 29/30. Phase 31 consumes the relative-peer pass (runs after expansion per D-01/D-02) but doesn't implement it.
- `2026-08-08-document-type-inference-omittable.md` — assigned to Phase 33 (docs sweep). Not relevant to Phase 31 implementation.

---

*Phase: 31-Template expansion*
*Context gathered: 2026-08-08*
