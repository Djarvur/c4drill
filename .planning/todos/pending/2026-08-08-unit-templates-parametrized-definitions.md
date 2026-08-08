---
created: 2026-08-08T00:53:13.640Z
title: Unit templates — parametrized unit definitions (define + instantiate)
area: tooling
files:
  - internal/parser/parser.go
  - internal/model/unit.go
  - internal/model/link.go
  - internal/validator/rules.go
---

## Problem

Real C4 models often contain many near-identical units. Today each must be spelled out in full — a repetitive, error-prone authoring burden. The user wants a templating mechanism: define an almost-complete unit once with parametrized fields, then instantiate it with concrete parameter values to produce concrete units in the diagram. Primary use case: "handy to define the number of almost-identical units."

Two reference implementations inform the design. Both are "textual macro" systems whose idioms should be adapted to C4Drill's structured TOML, not copied literally.

### Reference 1: go-metadot macros

go-metadot (`/Users/nil/DiskD/W/Djarvur/go-metadot`) implements this as a `define`/`enddef`/`&NAME args` macro system. **IMPORTANT: the Go port does NOT implement macros** — it parses them into AST structs but silently drops them at `internal/graph/graph.go:535` ("Эти элементы должны быть обработаны на уровне парсера"). The working implementation lives in the Perl original `metadot.pl`; that is the real spec.

#### go-metadot macro mechanics (reference, NOT a binding spec)

The Perl macro engine is a **line-oriented text preprocessor** interleaved with parsing:
- `define NAME` … body lines … `enddef` — body captured verbatim (`metadot.pl:888` `appendToDefine`)
- `&NAME arg1 arg2` — instantiate; args are positional, 1-indexed, shell-split (`metadot.pl:911` `Text::ParseWords`)
- `${1}`, `${2}`, … — positional substitution (`metadot.pl:1187` `doSubst`); unresolved `${N}` left literally in place
- `${uid}` — auto-injected unique id `NAME_<counter>` per instantiation (`metadot.pl:915`)
- One `&` call emits as many nodes/edges as the body has lines (NOT exactly one node) — a template is a multi-element sub-graph
- No defaults, all-string params, no arity validation, recursion cap 100, no nested `define` (fatal), forward references to later defines do NOT work
- Substitution is purely textual: a param can be used in labels, identifiers, styles, goto targets, concatenated with suffixes

### Reference 2: PlantUML `!procedure`

PlantUML's preprocessor ([docs](https://plantuml.com/preprocessing)) has a mature, widely-used procedure/macro system. It is the more ergonomic of the two references and the better model for C4Drill's named-parameter direction.

#### PlantUML `!procedure` mechanics (reference, NOT a binding spec)

- **Define:** `!procedure $name($arg1, $arg2) … !endprocedure`. Procedure and argument names are `$`-prefixed.
- **Call:** invoke by name with parens: `$name("foo1", "foo2")`.
- **Arguments — named AND positional, with defaults:** PlantUML supports both positional calls and Python-style keyword calls (`$element(myalias, $size=10, $technology="Java")`). Defaults allowed: `!function $inc($value, $step=1)` — "only arguments at the end of the parameter list can have default values."
- **Body emits diagram text directly** — a procedure is void (no return); whatever it contains is spliced into the diagram. A body can emit multiple elements (nodes, interfaces, arrows). Verbatim multi-element example from the docs:
  ```text
  !unquoted procedure COMP_TEXTGENCOMP(name)
  [name] << Comp >>
  interface Ifc << IfcType >> AS name##Ifc
  name##Ifc - [name]
  !endprocedure
  COMP_TEXTGENCOMP(dummy)
  ```
- **`!procedure` vs `!function`:** procedures emit text (void); functions return a value via `!return` and "do not output any text." C4Drill's template feature maps to `!procedure` (emit a unit), not `!function`.
- **`!unquoted`** keyword marks a procedure whose args don't need quotes — relevant if C4Drill wants shorthand invocation.
- **Procedures can call other procedures** (nesting allowed); local variables are scoped to the procedure body.
- **Deprecation lineage:** `!define`/`!definelong` are the deprecated predecessors; the docs say "Use `!function`, `!procedure` or variable definition instead." So `!procedure` is the *current recommended* PlantUML idiom — worth aligning C4Drill's naming/semantics with it for familiarity.

#### What to borrow from PlantUML specifically (vs go-metadot)

| Concern | go-metadot | PlantUML `!procedure` | C4Drill lean |
|---|---|---|---|
| Positional vs named | positional `${1}` | **both** (positional + keyword) | **named**, optionally with defaults — matches TOML map args and is far more readable |
| Defaults | none | **yes**, trailing args | **yes** — adopt PlantUML's "trailing args may have defaults" rule |
| Multi-element emission | yes (one line = one element) | yes (body splices arbitrary text) | under Option B: one template = one unit + its inline links (deliberate C4Drill constraint) |
| Typing | all-string | ints + strings | all-string (field-level substitution) |
| Nesting | no (fatal) | **yes** (procedures call procedures) | keep disallowed for v1; revisit if needed |

**Why this matters for C4Drill:** both references are textual macro systems embedded in a line-oriented DSL. C4Drill is structured TOML — a faithful text-level port of either does NOT fit. PlantUML's named-args + defaults idiom is the more natural fit for TOML's map syntax and should inform the C4Drill design. The adaptation must still choose where in the pipeline substitution happens — see Solution.

## Solution

### Open design question (resolve before implementing)

C4Drill's TOML is parsed into Go structs by `pelletier/go-toml/v2` (`internal/parser/parser.go:47`). Substitution can happen at one of three layers. Each has different ergonomics and cost. **A decision is required before implementation; do not assume one.**

**Option A — Text pre-processing (closest to go-metadot):** substitute `${N}` in the raw TOML bytes *before* handing them to go-toml. Pros: maximal power (params can expand into entire tables, multiple units, partial table fragments — anything text can express). Cons: fights TOML's structure — a param spanning table boundaries can produce invalid TOML; no line/column error fidelity (errors point to the post-substitution text); breaks `captureDefinitionOrder` (`parser.go:100`) if substitution changes line counts. Likely overkill and fragile for C4Drill's tabular model.

**Option B — Structured post-parse expansion (RECOMMENDED):** let go-toml parse templates into a dedicated `[template.<name>]` table as *data* (Unit struct + param spec), then run a Go-level expansion pass that deep-copies the template Unit, substitutes params into its string fields (`Name`, `Description`, `Technology`, `Color`, etc.) and into `Link.Peer`/`Link.Description`, and inserts the resulting concrete Units into `Model.Units` + `Model.UnitOrder`. Pros: type-safe, error-fidelity preserved, composes cleanly with existing validator and relative-peer resolution (if that lands first); keeps TOML valid; the expansion is a small, testable Go function. Cons: a param can only fill a *field value*, not structural slots (e.g. cannot expand one template into 3 sibling units via a single param) — but that matches "parametrize fields," which is the stated requirement.

**Option C — go text/template engine:** embed `text/template` and allow template bodies to use `{{.Param1}}` with conditionals/ranges. Pros: most powerful, familiar to Go devs. Cons: overkill for the stated use case; the DSL-in-TOML-meta-language gets complicated; harder to validate; precedent (go-metadot) deliberately uses dumb substitution, not a template engine.

**Recommendation: Option B.** It fits C4Drill's structured-data model, avoids the text-vs-structure fight, and the stated use case ("parametrize fields, instantiate to produce a unit") maps 1:1 to deep-copy + field substitution. Capture the chosen option as a decision in the eventual plan.

### Sketch of Option B (the likely shape, subject to design confirmation)

A new top-level table for definitions:

```toml
[template.microservice]
# params declared with optional defaults (PlantUML-style trailing defaults)
# params without defaults are required at instantiation
params = { name = "<required>", domain = "<required>", tech = "Go, gRPC" }

name = "${name} Service"
type = "container"
technology = "${tech}"
description = "${name} handles ${domain}"
# links inside a template use the same ${name} substitution
[[template.microservice.link]]
peer = "messageBus"
description = "Publishes ${domain} events"
```

Instantiation — a new top-level `[use]` section or per-unit `template` field. Named args (matching PlantUML's keyword-call idiom, not go-metadot's positional `${1}`):

```toml
# syntax 1: explicit use blocks
[[use]]
template = "microservice"
name = "auth"
domain = "authentication"
# → produces a unit at top level

# syntax 2: inline on the unit (cleaner, but constrains placement)
[auth]
template = "microservice"
params = { name = "auth", domain = "authentication" }
```

Syntax 1 (separate `[use]`) keeps template instantiation out of the unit table namespace and avoids collisions with the relative-peer and optional-name work; syntax 2 is terser but harder to place a templated unit *inside* a parent (which is the common case — you want N microservices inside a system). **Leaning toward syntax 1 with a `parent = "linuxSystem"` field** for placement. Resolve in plan.

### Param semantics to pin down (from go-metadot, adapt as fits C4Drill)

| Question | go-metadot | C4Drill proposal |
|---|---|---|
| Positional vs named | positional `${1}` | **named** `${name}` (TOML has maps; named is more readable and matches syntax sketch) |
| Defaults | none (go-metadot); **yes, trailing** (PlantUML) | **yes, PlantUML-style trailing defaults** — params without defaults are required; supplying them errors (clearer than go-metadot's silent-literal) |
| Typed | all-string | all-string (substitution is field-level) |
| Use in multiple fields | yes | yes — any string field |
| One template → multiple units | yes (line-based) | **no under Option B** — one template instantiates one Unit (+ its inline links). If multi-unit emission is needed, reconsider Option A. State this limitation explicitly. |
| `${uid}` / uniqueness | auto-injected `NAME_<count>` | consider: the instantiation table key provides uniqueness (e.g. `[[use]] name="auth"` → unit `auth`); may not need `${uid}` at all |
| Nested templates | no (fatal in go-metadot) | no — keep it simple; one level of instantiation |
| Forward references | no | no — templates must be defined before use (consistent with file-order semantics) |
| Recursion guard | cap 100 | cap, but if nested is disallowed, recursion is impossible — guard may be unnecessary |

### Touch points (Option B)

- **`internal/model/unit.go`** — `Unit` struct gains no field; templates live in a separate `Template` struct (`Name string`, `Params []string`, `Unit Unit`). Or model templates as a Unit with a `Params []string` field and a `IsTemplate bool` marker — decide in plan.
- **`internal/parser/parser.go`** — new top-level table handling for `[template.<name>]`; skip templates in `captureDefinitionOrder` (they're not renderable units); new post-parse expansion pass that walks `[use]` blocks, deep-copies, substitutes, inserts into `Model.Units`/`UnitOrder`. Update builtin field allowlist (`parser.go:309`) if templates use reserved keys.
- **`internal/validator/rules.go`** — validate: referenced template exists; required params supplied; instantiation produces a unit at a legal nesting level (C1/C2/C3 check at `rules.go:187`); templated unit's links resolve (existing peer check `rules.go:14` runs *after* expansion).
- **`internal/model/link.go`** — substitution must also apply to `Link.Peer`, `Link.Description`, `Link.Technology` etc., not just unit fields.
- **README.md / skill/SKILL.md** — document the feature with a before/after example (5 near-identical microservices spelled out → one template + 5 `[[use]]` blocks).
- **`skill/examples/`** — add a `06-templates.toml` demonstrating the pattern.

### Interaction with other pending todos

- **Relative-peer resolution** (`2026-08-08-toml-authoring-ergonomic-improvements.md`): templated units contain links; if both land, template substitution should produce *relative* peers that then resolve, OR template peers should be absolute-only. Ordering: ship relative-peer first, then templates can rely on it.
- **Optional `name`** (same todo): if `name` becomes optional via humanization, template units can omit `name` and inherit the instantiation key — complementary.
- **`reference` field** (`2026-08-08-add-reference-field-to-units.md`): templates can include a `reference` field with a parametrized URL (`reference = "https://wiki/${name}"`) — verify substitution covers it.

### Verification

- Unit tests for the expansion pass: param substitution into every string field, missing-param error, multi-instantiation producing distinct units, templated-unit links resolving post-expansion.
- End-to-end test: a model with a template + N instantiations produces the same SVG as a hand-expanded model with N spelled-out units (golden comparison, order-insensitive per the existing canonicalDOT approach — see STATE.md decision log).
- Confirm validator runs *after* expansion so peer-existence and level checks apply to the expanded model, not the template skeleton.
