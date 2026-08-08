# Phase 31: Template expansion - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-08-08
**Phase:** 31-Template expansion
**Areas discussed:** Relative-peer resolution site (HS-2), Instantiation syntax & unit placement, Reserved-table naming & legacy-collision (BC-1/BC-2)

---

## Relative-peer resolution site in templates (HS-2)

| Option | Description | Selected |
|--------|-------------|----------|
| Instantiation-site | Relative peers resolve against the parent where the [[use]] places the concrete unit. Templates are data, not location. Composes cleanly with Phase 30's pass (runs after expansion, treats all units identically). | ✓ |
| Template-site | Relative peers resolve against the template's lexical location. Since templates are top-level data with no parent, this effectively forbids relative peers in templates (must use absolute). | |
| Let me think / more detail | Wanted to think through a specific aspect before locking. | |

**User's choice:** Instantiation-site
**Notes:** Recommended by research; the structural argument (templates are data, instantiations have location) is decisive.

### Follow-up: parentless instantiation edge case

| Option | Description | Selected |
|--------|-------------|----------|
| Top-level = resolve against top-level siblings | Treat relative peers in a parentless instantiation the same as any top-level unit: bare name matches a top-level sibling, else absolute-fallback. Uniform, no template-special-case. | ✓ |
| Hard-error on relative peer in parentless instantiation | Forbid relative peers when instantiated without a parent. Stricter but catches a likely mistake. | |

**User's choice:** Top-level = resolve against top-level siblings
**Notes:** Uniformity wins — no special template logic in the relative-peer pass.

---

## Instantiation syntax & unit placement

| Option | Description | Selected |
|--------|-------------|----------|
| Syntax 1: separate [[use]] + parent + explicit name param | Dedicated [[use]] array-of-tables. Author explicitly names template, parent (placement), and ALL params incl. name-as-param. Nothing implicit. | ✓ |
| Syntax 2: inline [parent.name] + params map | Instantiation IS a unit table; path provides placement + identity, last segment becomes implicit ${name}. Terser but magic param injection; collision risk if template has no ${name} param. | |
| Let me think / describe my own | Wanted to think through a hybrid or describe a specific syntax. | |

**User's choice:** Syntax 1 (separate [[use]] + explicit name param)
**Notes:** Fully explicit preferred over terseness; no implicit param injection; [[use]] tables never collide with real units.

### Follow-up: name param vs subunit subtree

| Option | Description | Selected |
|--------|-------------|----------|
| name = top unit's segment; subunits keep declared keys | name param fills the top-level unit's last path segment. Subunit keys stay verbatim (NOT parametrized). Parametrization is field-value-only — preserves the multi-output/for_each anti-feature boundary. | ✓ |
| Subunit paths themselves can be parametrized | Template could declare [template.svc.${suffix}] with suffix param. More flexible but blurs field-value vs structural-slot line; approaches multi-output anti-feature. | |
| Let me think / more detail | Wanted to think about subunit naming. | |

**User's choice:** name = top unit's segment; subunits keep declared keys
**Notes:** Clean separation between field-value parametrization (allowed) and structural parametrization (forbidden).

### Follow-up: duplicate-path check scope (TMPL-07)

| Option | Description | Selected |
|--------|-------------|----------|
| Full-path-set check, all collisions hard-error | After all [[use]] expansions, check the full resulting path set (hand-authored + all expanded instances). Any collision (templated/templated or templated/hand-authored) = single hard error naming both sources. | ✓ |
| Templated-vs-templated only | Only check templated/templated; templated overwriting hand-authored is warning or silent last-wins. Looser, risks silent data loss. | |

**User's choice:** Full-path-set check, all collisions hard-error
**Notes:** Catches every collision in one pass; matches the milestone's hard-error stance.

---

## Reserved-table naming & legacy-collision (BC-1/BC-2)

| Option | Description | Selected |
|--------|-------------|----------|
| Bare names: [template.*], [[use]], [[include]] | Shortest, most readable, matches go-metadot/PlantUML idiom. Collision risk for existing models with units named template/use/include — mitigated by detection + migration error. | ✓ |
| Namespaced: [c4drill.template.*], [[c4drill.use]] | Zero collision risk (no real unit can be named c4drill.template). Verbose/ugly; every directive carries a meaningless prefix. | |
| Bare + top-level-only reservation | Reserve bare names only at top level; subunit named use/template stays legal. Smaller collision surface but inconsistent (name reserved in one position, not another). | |
| Let me think / propose my own | Considering alternatives (leading underscore, single [c4] wrapper). | |

**User's choice:** Bare names: [template.*], [[use]], [[include]]

### Follow-up: collision policy for existing models

| Option | Description | Selected |
|--------|-------------|----------|
| Detect + clear migration error | Parser detects legacy unit named template/use/include and emits migration error pointing at the rename fix. | |
| Silent allow if directive fields absent | Treat [use] as a legacy unit if no template field present. Zero breakage but ambiguity + complicates parser. | |
| Deprecation warning, continue as unit | Warn but continue treating as unit; reserve right to hard-error later. | |
| (free-text) | "there are NO existing models yet, we are not public" | ✓ |

**User's choice:** No collision-mitigation machinery needed — C4Drill is not public, no existing models to break.
**Notes:** This dissolved the entire collision sub-question. Parser treats template/use/include as reserved top-level keys, full stop. No detection, no migration errors, no silent-allow ambiguity. BC-2 effectively resolved: no legacy to protect.

---

## Claude's Discretion

- Exact `TemplateDef` / `Instantiation` struct shapes on `parser.Model` (research recommends dedicated `toml:"-"` fields; planner may refine).
- Internal package structure for the expansion pass (`internal/template/Expand` vs. function in `internal/parser/`).
- Whether to deep-copy-then-substitute or substitute-in-place-on-copy (operation ordering within `Expand`).

## Deferred Ideas

- **Parameter defaults** (trailing-default support) — out for v1.10 (strictness); revisit if authoring friction emerges.
- **Template multi-output / fan-out / `for_each`** — anti-feature for v1.10.
- **Template nesting** (template instantiating another template) — deferred; one level in v1.10.
- **Structural parametrization** (parametrized subunit keys/paths, array/conditional link expansion) — would cross into multi-output territory; deliberately out.
