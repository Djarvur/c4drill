# Feature Landscape: v1.10 Model Composition

**Domain:** C4Drill v1.10 Milestone — how comparable tools implement the four target features
**Researched:** 2026-08-08
**Scope:** INCLUDE/multi-file composition, TEMPLATES/parametrized definitions, ERGONOMIC SHORTHAND (relative peer + optional name), ELEMENT HYPERLINKS (reference field)
**Confidence:** HIGH (authoritative docs fetched for each tool; design context cross-checked against the four pending todos)

> This document broadens beyond the captured go-metadot/PlantUML references in the todos to distinguish **table stakes** (users expect this; absence feels broken) from **differentiators** (polish) from **anti-features** (would fight C4Drill's auto-generated-view philosophy). The requirements-definition step will consume the classifications here to scope REQ-IDs.

## C4Drill's load-bearing philosophy (the filter for every classification)

Three properties constrain what is and is not a fit. Anti-features below are tied back to one of these.

1. **Auto-generated views.** The user authors a *model* (units + links); C4Drill *derives* C1/C2/C3 automatically (`internal/view`, `internal/graph/builder.go`). There is no user-authored view DSL, no include/exclude predicates, no per-view styling. Structurizr and LikeC4 both have rich view DSLs; C4Drill deliberately does not.
2. **Structured TOML, not a custom DSL.** The DSL-vs-TOML question was settled NO (`2026-08-08-toml-authoring-ergonomic-improvements.md`). Features must compose with TOML's map/table model; textual-preprocessor tricks that span table boundaries are anti-patterns here.
3. **Single-shot CLI, single entry file.** `c4drill <file.toml>` (`cmd/c4drill/root.go`). Anything that requires a project/workspace directory or a build orchestrator is out of scope.

---

## 1. INCLUDE / MULTI-FILE COMPOSITION

**Research question:** How do config-driven tools assemble a model from multiple files — include vs include_once semantics, conflict resolution, path resolution, transitive includes?

### Capability matrix

| Capability | Structurizr DSL | D2 | Terraform (override files) | Terraform (modules) | CUE | Jsonnet | Puppet (legacy) | PlantUML |
|---|---|---|---|---|---|---|---|---|
| Directive to splice another file | `!include <file\|dir\|url>` ([includes](https://docs.structurizr.com/dsl/includes)) | `@file` or `@"file.d2"` ([imports](https://d2lang.com/tour/imports/)) | filename `*_override.tf` auto-detected ([override](https://developer.hashicorp.com/terraform/language/files/override)) | `module "x" { source = ... }` | `package` import (`"pkg"` in cue) | `import 'file.libsonnet'` | `import` (removed in Puppet 4) | `!include` / `!include_many` / `!include_once` ([preprocessing](https://plantuml.com/preprocessing)) |
| Path resolution | Relative to including file; must be in same dir or subdir; HTTPS for URLs | **Relative to the file doing the importing**, not the cwd ([imports](https://d2lang.com/tour/imports/)) | Same dir as main config (override files are co-located) | Module source path; registry/hosted/relative | Module root from `cue.mod/module.cue` | Relative to importing file | Was relative to `site.pp` | Relative to including file; search dirs via `!include <searchdir:file>` |
| Transitive (included files can include) | Yes — content is inlined, then re-processed | Implied yes (no doc statement found) | No chaining — overrides applied sequentially against main config | Yes — modules compose freely | Yes — packages import packages (Go-like) | Yes — "a given Jsonnet file can be recursively imported" ([spec](https://jsonnet.org/ref/spec.html)) | Yes (historically) | Yes |
| Conflict on duplicate entity | Not documented; inline-merge model suggests last-wins or error | Not documented | **Well-defined:** override's same-named arg *replaces* original; nested blocks *replace* wholesale; multiple overrides compound in lexicographic filename order ([override](https://developer.hashicorp.com/terraform/language/files/override)) | **Hard error** on duplicate resource address ([terraform](https://oneuptime.com/blog/post/2026-02-23-how-to-fix-duplicate-resource-address-error-in-terraform/view)) | Structurally merged (unification); conflicting concrete values = error | Object concatenation merges fields; conflicting values last-wins unless explicitly error | Duplicate resource in scope = error (parametrized-class workaround) | `!includesub` PREFIX/STRIP/KEEP for namespace conflicts |
| `include_once` / dedup | **No** — no dedup keyword | Not documented | N/A (override is a different model) | N/A (modules are explicit calls) | N/A (package identity is dedup'd by Go-style import) | N/A (imports are pure references, not splices) | N/A | **Yes — `!include_once` and `!include_many`** are the canonical primitives ([preprocessing](https://plantuml.com/preprocessing)) |
| Cycle detection | Not documented | Not documented | N/A | Terraform errors on module cycles | **Errors with `import cycle`** ([cue/load](https://pkg.go.dev/cuelang.org/go/cue/load), [issue #990](https://github.com/cue-lang/cue/issues/990)) | Errors on import cycle (Go-like) | Was a runtime stack check | PlantUML preprocessing has no documented guard; relies on user discipline |
| Partial import (subset of file) | No | **Yes — `@"file".shapeName`** targets one object ([imports](https://d2lang.com/tour/imports/)) | No (whole override block) | Yes — module outputs are referenced selectively | Yes — fields are addressable | Yes — `importstr`, `import` return the object; you index it | No | `!includesub` for sub-diagrams |

### Key cross-tool patterns

- **Two distinct models of "multi-file":** (a) *textual splice* (Structurizr `!include`, PlantUML `!include`, Puppet `import`) — the file content is inlined at the directive site, in order; (b) *named reference* (D2 `@`, CUE `import`, Jsonnet `import`, Terraform modules) — the imported file is parsed as a standalone unit and its *named exports* are referenced. C4Drill's todos lean toward splice (Option B: parse-then-merge structs), matching go-metadot's `registerInclude` and PlantUML's `!include`.
- **Conflict handling splits cleanly by model.** Splice tools either error (Terraform module address collision) or merge-last-wins (Terraform override, PlantUML `!includesub`). Reference tools unification/merge at the value level (CUE) or concatenate at object level (Jsonnet). **No tool silently double-defines an entity** — that is a universal bug, not a design choice.
- **`include_once` is rare and PlantUML-specific.** Only PlantUML ships it as a first-class directive. It exists specifically to support *definition-library* files that many consumers include safely — exactly C4Drill's template-isolation use case. This is the single most relevant borrowed idiom for C4Drill.
- **Path resolution is universally "relative to the including file"** across splice tools (Structurizr, D2 explicit, PlantUML, go-metadot `registerInclude`). This is the strongest consensus in the whole matrix.

### Classification

| Sub-feature | Label | Rationale |
|---|---|---|
| `include` directive (assemble multiple files) | **Table stakes** | Every config tool with files-at-scale supports this. A 499-line production model (`cyp-auth-infra/saira`) is the direct evidence. Structurizr, D2, PlantUML, CUE, Jsonnet, Terraform all have it. |
| Transitive includes (included files can include) | **Table stakes** | Universal. Without it, a template library that includes a shared `icons.toml` is impossible. Implied by splice model; trivial to support. |
| Path resolution relative to the *including* file | **Table stakes** | Universal convention. Resolving relative to cwd would surprise every user coming from any other tool. |
| Cycle detection (fatal on re-enter) | **Table stakes** | Otherwise infinite loops. Both references (go-metadot `@incChain`, PlantUML preprocessing) and CUE's `import cycle` do this. Cheap to implement (a stack). |
| `include_once` / dedup flag | **Differentiator** | PlantUML-only. Not expected by users coming from D2/Structurizr/CUE — but *essential* for the template-library use case C4Drill explicitly targets. Worth the small cost (a seen-set + a per-include `once` flag). |
| Conflict on duplicate unit path = **hard error** | **Table stakes** | Matches Terraform module-address behavior and avoids silent shadowing. The todos already lean this way; this research confirms it is the safe, expected choice. |
| Conflict resolution via last-wins *merge* (Terraform override style) | **Anti-feature** | Override semantics fight the auto-generated-view philosophy: silent last-wins lets a faraway included file rewrite a unit's `name` or `link`, producing a diagram that disagrees with the model file the user is looking at. C4Drill's value is "the TOML *is* the source of truth"; override files break that contract. **Recommend hard-error instead.** |
| Partial import (`@"file".unit` à la D2) | **Anti-feature** (for v1.10) | Adds a namespace/addressing layer for marginal gain at C4Drill's scale. C4Drill's unit namespace is flat dotted paths; a partial-import feature would need a collision story and a name-resolution layer. Defer; revisit only if template libraries grow large. |
| URL includes (`!includeurl`) | **Anti-feature** | Introduces a network dependency and a security surface (untrusted remote TOML) for a single-shot local CLI. Structurizr allows it because its tooling is workspace-centric; C4Drill's single-file CLI ethos says no. |
| Namespacing (`!includesub PREFIX` à la PlantUML) | **Anti-feature** (for v1.10) | C4Drill's dotted-path unit identity would need a parallel namespace concept. Adds cognitive load for a problem (cross-file same-name units) that hard-error-on-conflict already solves correctly. |

### Recommendations for C4Drill — Include

1. **Adopt the splice model** (Option B from the todo: parse-then-merge structs). It matches both references (go-metadot, PlantUML) and the universal convention among splice tools.
2. **Path resolution: relative to the including file.** Non-negotiable; every comparable tool does this.
3. **Cycle detection: fatal error** with an include-stack error message naming the chain.
4. **Conflict on duplicate unit path: hard error.** Do *not* implement Terraform-override-style silent merge.
5. **Ship `once = true`** (include_once) on day one — it is the enabler for the template-library use case that motivates the whole feature. PlantUML precedent.
6. **Use the array-of-tables directive shape** (`[[include]] path = ... once = true`) from the todo — it scales to future flags without redesign, and mirrors PlantUML's `!include` family.
7. **Explicitly reject** URL includes, override-style merge, and namespacing in the plan's "Out of scope" section, with the auto-generated-view / single-file-CLI rationale.

---

## 2. TEMPLATES / PARAMETRIZED DEFINITIONS

**Research question:** Named vs positional params, defaults, scoping, multi-output emission, the UX, recursion/nesting.

### Capability matrix

| Capability | PlantUML `!procedure` | Jsonnet | Helm named templates | Dhall | CUE | Terraform modules | Terraform `for_each` | Ansible roles | Structurizr DSL |
|---|---|---|---|---|---|---|---|---|---|
| Param style | **Both positional + named** ([preprocessing](https://plantuml.com/preprocessing)) | **Both positional + named**; positional must precede named ([language ref](https://jsonnet.org/ref/language.html)) | Single dict arg emulating named params ([named templates](https://helm.sh/docs/chart_template_guide/named_templates)) | Positional λ-calculus functions | Implicit (record fields); no functions as first-class | Named via `variable {}` blocks | `each.key` / `each.value` | Named via `defaults/main.yml` | **None — no parametrized elements** |
| Defaults | **Yes, trailing args only** ([preprocessing](https://plantuml.com/preprocessing)) | Yes, `param=default` ([language ref](https://jsonnet.org/ref/language.html)) | No native; emulated via `default` filter | Yes, via `\(x ? default)` | Yes, disjunction defaults `field: value \| default` ([defaults](https://cuelang.org/docs/howto/specify-a-default-value-for-a-field/)) | Yes, `default =` in variable block | N/A | Yes, `defaults/main.yml` lowest precedence | N/A |
| Multi-output (one call → many entities) | **Yes — body splices arbitrary text** (multiple nodes/edges) ([preprocessing](https://plantuml.com/preprocessing)) | Yes — function returns any value | Yes — partial emits any text | Yes — function returns any record | Yes — definition is just a value | Yes — module emits many resources | Yes — one block → many resources | Yes — role = many resources | **Yes (deployment only):** `softwareSystemInstance` / `containerInstance` instantiate *existing* elements for deployment nodes; not parametrized creation ([language](https://docs.structurizr.com/dsl/language)) |
| Typing | Ints + strings | Untyped (duck) | Untyped (Go template) | **Strongly typed**, total | **Strongly typed**, schemas | Type-constrained variables | N/A | Untyped | Tags only |
| Nesting (templates call templates) | **Yes** (procedures call procedures) | Yes (closures, recursion) | Yes | Yes (functions compose) | Yes (values reference values) | Yes (nested module calls) | No (single-level meta-arg) | Yes (role dependencies) | N/A |
| Scoping | Local vars scoped to body | Lexical closures | Single `. ` arg; helper convention | Lexical | Lexical + hidden fields `_foo` | Per-module isolation | `each` scope | Per-role defaults isolation | Element-scoped tags |
| Forward references (use before define) | No (define-before-use) | No (top-level order independent for objects, not for locals) | No | No | Yes (order-independent) | Yes (modules resolve after load) | No | N/A | Yes (DSL is multi-pass) |
| One-call-one-entity (the C4Drill v1.10 proposal) | No (multi-output is the norm) | No | No | No | No | No | No | No | N/A (no templating) |

### Key cross-tool patterns

- **Named + defaulted params is the dominant idiom.** PlantUML, Jsonnet, Helm (via dict), Dhall, CUE, Terraform all converge on it. Positional-only systems (go-metadot `${1}`) are the historical exception, not the modern baseline. Users coming from any of these tools will *expect* named params with defaults.
- **Defaults are universally "trailing args may have defaults."** PlantUML and Jsonnet both enforce positional-before-named and defaults-only-at-tail. This is the single most consistent micro-rule across the matrix.
- **Multi-output is the norm; one-template-one-entity is the exception.** Every tool except C4Drill's proposed Option B lets one call produce many entities. This is the biggest place C4Drill would *differ* from precedent — and it is a *deliberate, defensible* difference (see anti-feature row below).
- **Structurizr — the closest C4 sibling — has NO parametrized element creation.** Its `softwareSystemInstance`/`containerInstance` are *deployment-time instances of already-defined elements*, not factories. Its closest reuse mechanism (`archetypes`) only adds tags/defaults to existing types. **This means C4Drill's templates feature would be a genuine differentiator vs Structurizr**, not table stakes within the C4-tool family.
- **Terraform's `for_each` is the most relevant precedent for "N near-identical units."** The `for_each` meta-argument exists specifically because users wanted "several similar objects without writing a separate block for each" ([for_each](https://developer.hashicorp.com/terraform/language/meta-arguments/for_each)). The stated C4Drill motivation ("define the number of almost-identical units") is word-for-word the Terraform `for_each` use case. `for_each` is preferred over `count` because map keys are stable identifiers — directly analogous to C4Drill unit paths.

### Classification

| Sub-feature | Label | Rationale |
|---|---|---|
| Named params (`${name}` not `${1}`) | **Table stakes** | Universal across modern tools (PlantUML, Jsonnet, Helm, Dhall, CUE, Terraform). Positional-only reads as archaic. TOML's map syntax makes named params the natural fit (per todo). |
| Optional params with defaults (trailing only) | **Table stakes** | PlantUML + Jsonnet + Dhall + CUE + Terraform all support this with the same trailing-defaults rule. Users expect "omit a param → sensible default." |
| Error on missing required param (no silent literal `${name}`) | **Table stakes** | go-metadot's silent-literal behavior is a footgun; every typed tool errors instead. The todo already leans this way. |
| One template → one unit (+ its inline links) under Option B | **Differentiator (deliberate)** | Differs from every precedent (all allow multi-output) but is the *correct* call for C4Drill: keeps the expansion pass a small, testable Go function; preserves `captureDefinitionOrder` semantics; matches "parametrize fields, instantiate a unit." Document the limitation explicitly so users don't expect Terraform-`for_each`-level power. |
| Substitute params into *all* string fields (Name, Description, Technology, Color, Link.Peer, reference) | **Table stakes** | Field-level substitution is the whole point. Any tool that omits a field forces the user back to hand-expansion. Must include `Link.Peer` and the new `reference` field per the cross-feature dependency. |
| Multi-output (one template → N sibling units via one call) | **Anti-feature** (for v1.10) | PlantUML/go-metadot allow this because they are text preprocessors; C4Drill is structured TOML. Multi-output under TOML requires either a `count`/`for_each` meta-field (Terraform-level complexity: stable key derivation, index-vs-key semantics) or a text-preprocess escape hatch (fights the structured-data model). The stated use case (5 near-identical microservices) is served by 5 `[[use]]` blocks — verbose but explicit and on-philosophy. **Defer; do not design v1.10 around it.** |
| Nesting (templates instantiate other templates) | **Anti-feature** (for v1.10) | PlantUML allows it; Jsonnet/CUE allow it via composition. But C4Drill templates are field-substitution, not compositional — nesting adds a resolution-order problem (which template expands first?) for marginal reuse gain. The todo already proposes disallowing it; this research confirms precedent allows but does not require it. One level is enough. |
| Forward references (use template before its definition) | **Anti-feature** (for v1.10) | Only CUE/Terraform-modules/Structurizr allow forward refs, and all of them are multi-pass by design. C4Drill's `captureDefinitionOrder` is single-pass file-order; allowing forward refs would require either a two-pass parser or a fixpoint, both of which fight the simplicity thesis. Define-before-use is the go-metadot/PlantUML/Jsonnet-local/Helm norm. |
| Typed params (int vs string) | **Anti-feature** | CUE and Dhall prove types catch bugs, but they require a type system. C4Drill substitution is field-level strings; all-string params (PlantUML/Jsonnet/Helm model) is the right scope. Adding types is a different project. |
| `${uid}` auto-injected uniqueness | **Differentiator (defer)** | go-metadot injects `NAME_<counter>`. C4Drill likely doesn't need it: the `[[use]] name="auth"` table key *is* the unique identity. Defer unless real models hit collisions. |

### Recommendations for C4Drill — Templates

1. **Named params with trailing defaults** — the universal modern idiom, fits TOML maps. Reject positional `${1}`.
2. **One template → one unit + inline links** (Option B from the todo). This is the *correct* deviation from precedent for a structured-TOML tool; document the one-to-one limit explicitly.
3. **Substitute into every string field**, including `Link.Peer`, `Link.Description`, and the new `reference` URL field — these are cross-feature dependencies (see §5).
4. **Hard-error on missing required param.** No silent literals.
5. **No nesting, no forward references, no multi-output, no typed params** in v1.10. Each is a defensible "no" tied to keeping the expansion pass small and the structured-data model intact.
6. **Re-evaluate multi-output (`for_each`-style) post-v1.10** if the 5-`[[use]]`-blocks pattern proves too verbose in practice. Terraform `for_each` is the model to copy *if* this is ever revisited.

---

## 3. ERGONOMIC SHORTHAND — relative peer + optional name

**Research question:** How do tools handle relative references in nested config, and how do they humanize machine IDs into display names?

### Capability matrix

| Capability | Terraform HCL | CUE | Jsonnet | Kustomize | Helm | Terraform `title()` | GraphViz | Mermaid |
|---|---|---|---|---|---|---|---|---|
| Relative reference to sibling | `aws_instance.web.id` — fully-qualified traversal; **no implicit "sibling"** | Nearest-enclosing-scope reference; `self`/`super` for object context ([scope](https://cuetorials.com/overview/scope-and-visibility/)) | `self` / `super` / `$` for object context; locals are file-scoped | `namePrefix` / `nameSuffix` applied uniformly; resources referenced by post-transform name | `.` context + `{{ .Values.x }}`; no implicit sibling | N/A | N/A | Node IDs are flat; no nesting |
| Humanize ID → display name | Resource names are snake_case (`postgresql_database`); display via `title()` function ([title](https://developer.hashicorp.com/terraform/language/functions/title)); no auto-humanization | No convention; values are literal | No convention | `namePrefix`/`nameSuffix` are explicit transforms | `{{ .Values.nameOverride }}` is explicit | `title()` capitalizes first letter of each word; **does not change other letters** ([title](https://developer.hashicorp.com/terraform/language/functions/title)) | `label` is free text; ID is `node_id` | Node label is free text in `[brackets]`; ID is separate |
| Acronym preservation in humanization | `title()` does NOT preserve — "gRPC" → "Grpc" ([title SO](https://stackoverflow.com/questions/65112736/in-terraform-is-it-possible-to-change-the-casing-of-values)); users write custom logic | N/A | N/A | N/A | N/A | Same limitation | N/A | N/A |
| Scope-walk on unresolved reference | Error — no walk-up | Walk-up via lexical scope | Walk-up via lexical scope | N/A (transform is global) | Error | N/A | N/A | N/A |

### Key cross-tool patterns

- **Relative references are surprisingly uncommon.** Terraform — the most widely-used config language — has *no* implicit sibling reference: every reference is a fully-qualified traversal (`aws_instance.web.id`). Implicit context (`self`, `super`, `$`) exists only in object-oriented config langs (Jsonnet, CUE) and only for *the current object's* fields, not for siblings. **C4Drill's proposed relative-peer (resolve `sessionManager` against the enclosing parent) is genuinely novel** among the tools surveyed — none of them do sibling-resolution-by-scope-walk. This makes it a *differentiator*, not table stakes.
- **Humanization is universally explicit, not implicit.** No surveyed tool auto-derives a display name from a machine ID by default. Terraform ships a `title()` function the user *calls explicitly*. Kustomize/Helm require explicit `nameOverride`. GraphViz and Mermaid keep ID and label as separate free-text fields. **C4Drill's proposed "omit `name` → humanize from segment" would be more aggressive than any comparable tool** — every other tool makes the user opt in.
- **Acronym preservation is a known unsolved problem.** Terraform's `title()` mangles "gRPC" → "Grpc" ([SO](https://stackoverflow.com/questions/65112736/in-terraform-is-it-possible-to-change-the-casing-of-values)). There is no community-standard acronym-aware humanizer. If C4Drill attempts acronym preservation (the todo raises `grpcAPIs` → "gRPC APIs"), it inherits an unsolved general problem.
- **Terraform `title()`'s exact rule** — "capitalizes the first letter of each word; does not change any other letters" — is the minimal viable humanizer. For input that is already camelCase (`sessionManager`), this rule alone produces "SessionManager" (no space inserted). C4Drill needs *camelCase-splitting* on top, which Terraform's `title()` does not do. So even the closest precedent requires custom logic.

### Classification

| Sub-feature | Label | Rationale |
|---|---|---|
| Relative-peer resolution (short `peer` resolved against enclosing parent, walking up, with absolute fallback) | **Differentiator** | No comparable tool does sibling-scope-walk resolution. It is a real ergonomic win (40% of the saira model is re-spelled paths) but users won't *expect* it from prior tool experience. Ship it, document it loudly, make relative-first-with-absolute-fallback the documented precedence. |
| Optional `name` with humanization from segment | **Differentiator** | More aggressive than any comparable tool (all require explicit display names or explicit `title()` calls). High payoff (kills boilerplate) but should be **explicit opt-out**: if humanization misfires (acronym case), the user adds `name =` and wins. Backward-compatible (explicit name always wins). |
| Acronym preservation in humanization (gRPC, IDP, API) | **Anti-feature** (for v1.10) | Terraform's `title()` proves this is an unsolved general problem. No surveyed tool does it well. Any heuristic C4Drill ships will misfire on edge cases (is "IDP" an acronym? "GRPC"? "REST"?). The escape hatch (`name =` always wins) makes a dumb humanizer acceptable. **Ship the dumb camelCase-split + title-case; do not chase acronyms.** |
| Compact one-liner link syntax (`link = ["a", "b"]` for peer-only edges) | **Anti-feature** (for v1.10) | TOML grammar is rigid; coercing `string → Link` is parser-special-casing for marginal gain. The todo already defers this pending relative-peer's impact. Re-measure saira after relative-peer lands; the inline-array form may be compact enough. |
| Walk-up precedence (relative-first vs absolute-first) | **Differentiator** | The todo recommends relative-first-with-absolute-fallback. This is the *opposite* of Terraform (which has no relative at all) and is a genuine design choice. Document it; the ambiguity risk (two siblings in different branches share a short name) is acceptable and matches every scoping rule's nearest-wins semantics. |

### Recommendations for C4Drill — Ergonomics

1. **Ship relative-peer** as a post-parse pass. It is the highest-payoff item and a genuine differentiator vs every comparable tool.
2. **Ship optional `name` with dumb humanization** (camelCase split + title-case on the *segment only*, not the full path). Explicit `name =` always wins. Backward-compatible.
3. **Do not chase acronym preservation.** Ship the dumb version; document the `name =` escape hatch. Terraform's `title()` precedent shows this is a tar pit.
4. **Defer compact-link syntax** until relative-peer is measured. Likely unnecessary after.
5. **Order in pipeline:** relative-peer runs *after* template expansion (per the todo) so templated-unit links resolve correctly. This ordering is load-bearing.

---

## 4. ELEMENT HYPERLINKS — the `reference` field

**Research question:** How do diagram tools attach URLs to nodes, and are they clickable in SVG?

### Capability matrix

| Tool | Field/keyword | Syntax | Clickable in SVG? | Visual indicator | Tooltip companion |
|---|---|---|---|---|---|
| **GraphViz** (C4Drill's backend) | **`URL`** (or `href`) attribute on node/edge/cluster | `node1 [URL="https://..."];` | **Yes — emits `<a>` anchor wrapping the shape** ([URL attr](https://graphviz.org/docs/attrs/URL/)) | None by default; pair with `tooltip` and a label glyph | **`tooltip` attribute** — separate, orthogonal |
| Structurizr DSL | `url` keyword on element/relationship | `url https://example.com` inside element block ([language](https://docs.structurizr.com/dsl/language)) | Rendered by Structurizr's viewers (cloud/on-prem) | Element gets a visual link indicator in the viewer | `description` is the companion context |
| LikeC4 | **`link`** keyword on element | `link https://...` or `link https://... 'Label'` inside element block ([model](https://likec4.dev/dsl/model/)) | Yes (rendered diagrams); supports any URI scheme incl. `ssh://` and relative source links | Implicit | `description` / `summary` |
| D2 | **`link`** shape attribute + **`tooltip`** shape attribute | `MyShape: { link: https://...; tooltip: ... }` ([interactive](https://d2lang.com/tour/interactive/)) | **Yes — "click to go to an external link"** | **Icon shown on shapes with tooltip/link** ([release 0.1.4](https://d2lang.com/releases/0.1.4/)) | `tooltip` is a first-class sibling attribute |
| PlantUML | `[[url]]` after element, or `url of X is [[url]]` | `class Car [[http://...]]` ([link](https://plantuml.com/link)) | Yes — emits clickable SVG anchor; "still clickable" even when not underlined | Optional label `{tooltip}` syntax | **`{tooltip}` in braces** — composable with link |
| Mermaid | `click NodeID href "URL" "tooltip"` directive | `click A href "https://..." "tooltip"` | Yes (wraps node in SVG anchor); requires `securityLevel` config for JS callbacks ([GH issue 6314](https://github.com/mermaid-js/mermaid/issues/6314)) | None native | Tooltip arg in click directive |

### Key cross-tool patterns

- **Every diagram tool supports clickable node URLs.** This is universal table stakes in the diagram-tool family. GraphViz, Structurizr, LikeC4, D2, PlantUML, Mermaid all ship it. *Not* having it is the gap C4Drill is closing.
- **The field name splits into two camps.** (a) `url` — Structurizr, GraphViz attribute; (b) `link` — LikeC4, D2, Mermaid-adjacent. C4Drill's `link` is already taken for *relationships*, so `reference` is the correct collision-avoiding name. This is a C4Drill-specific constraint, not a deviation from convention — the *concept* is universal, only the name differs.
- **GraphViz's `URL` attribute is the implementation path of least resistance.** C4Drill already emits DOT and renders via go-graphviz; setting the node's `URL` attribute generates the SVG `<a>` anchor for free. This is the single lowest-cost feature in the whole milestone — the research confirms the renderer already speaks this natively.
- **D2's `link` + `tooltip` pair is the most polished UX.** D2 shows an *icon* on shapes that have a link/tooltip so the affordance is discoverable. C4Drill's proposed 📖 marker is the same idea — a visible affordance — and the repo already has precedent (the 🔍 collapsed-cluster indicator, commit `2ac5202`).
- **PlantUML's tooltip syntax** (`[[url {tooltip}]]`) shows link and tooltip are composable but orthogonal. C4Drill's `description` field already serves the tooltip role; the `reference` field should be link-only and not try to also be a tooltip.

### Classification

| Sub-feature | Label | Rationale |
|---|---|---|
| Per-element URL field | **Table stakes** | Every diagram tool has it. The LikeC4 comparison already judged this *the one* worth adopting. Not having it is the gap. |
| Named `reference` (not `link`/`url`) to avoid collision with relationship `link` | **Table stakes** | C4Drill-specific naming constraint. The concept maps to `url`/`link` elsewhere; the name is forced by the existing relationship field. |
| Render as clickable SVG `<a>` via GraphViz `URL` attribute | **Table stakes** | Native GraphViz capability ([URL attr](https://graphviz.org/docs/attrs/URL/)); go-graphviz already wraps it. Near-zero implementation cost. This is the cheapest table-stakes item in the milestone. |
| Visible affordance (📖 marker next to label) | **Differentiator** | D2 shows an icon; C4Drill's 📖 matches the repo's existing 🔍 precedent. Not strictly required (GraphViz links work invisibly) but is the polish that makes the feature discoverable. Low cost, high visibility. |
| HTML format's JS nav shim extends to reference links | **Table stakes** | C4Drill already ships `-f html` with a JS shim because Safari silently ignores SVG `<a>` for *navigation* links ([PROJECT.md](../PROJECT.md)). The same shim must handle `reference` clicks or the feature is broken in Safari. Verify in implementation; likely free if the shim is generic. |
| URL validation beyond `http(s)://` prefix | **Anti-feature** | LikeC4 deliberately supports `ssh://`, `mailto:`, relative source links. Over-validating rejects legitimate URIs. A bare prefix check (or none) is the right scope. |
| Tooltip tied to the link (PlantUML `{tooltip}` style) | **Anti-feature** | C4Drill's `description` field already carries the context. A link-specific tooltip duplicates it and adds rendering complexity. Keep `reference` link-only. |
| Per-relationship URL ( Structurizr allows `url` on relationships too) | **Anti-feature** (for v1.10) | Scope creep. The todo scopes this to elements only. Relationships already have `description`/`technology` labels. Defer unless a concrete use case appears. |
| Image/background property (LikeC4 discussion #1458) | **Anti-feature** | Explicitly rejected in the LikeC4 comparison (icons are contrary to the current `dev/shapes-no-icons` branch). Do not implement. |

### Recommendations for C4Drill — Reference

1. **Ship it as `reference`** (collision-free name) with the 📖 affordance.
2. **Implement via GraphViz `URL` attribute** — native, near-zero cost, emits clickable SVG anchors automatically. This is the lowest-effort table-stakes win in the milestone.
3. **Verify the `-f html` JS shim handles `reference` clicks**, not just drill-down navigation. Safari silently ignores SVG `<a>`; the shim is the workaround. If the shim is keyed on href patterns, `reference` URLs (external) vs drill-down URLs (relative `.svg` files) may need distinction.
4. **No URL validation** beyond an optional `http(s)://`-or-similar affordance. LikeC4 supports arbitrary URI schemes; do not be stricter than that.
5. **Elements only**, not relationships, in v1.10. Matches the todo scope.

---

## 5. CROSS-FEATURE DEPENDENCIES

The four features are not independent; their interactions are load-bearing for the pipeline ordering `include → template-expand → relative-peer-resolve → validate → render` (captured in the include todo).

### Dependency graph

```
include ──enables──► template-isolation (template libs in their own file, included with once=true)
include ──enables──► large-model-split (one file per domain)

template-expand ──requires──► reference field exists (templates can carry parametrized reference URLs)
template-expand ──requires──► relative-peer resolved AFTER expansion (templated-unit links use ${name} in peer)
template-expand ──constrains──► one-template-one-unit (so ${name} peer substitution is well-defined)

relative-peer ──consumes──► template expansion output (peers inside templated units must resolve)
relative-peer ──orthogonal──► include (peer resolution runs on the merged model, file origin irrelevant)

reference ──orthogonal──► include, relative-peer (just a field on the merged unit)
reference ──consumes──► template param substitution (reference = "https://wiki/${name}" must substitute)

validate ──consumes──► all four (runs last, on the fully assembled + expanded + resolved model)
render  ──consumes──► reference field (emits GraphViz URL attribute)
```

### Concrete interaction cases the plan must test

1. **Template + reference parametrization.** A template with `reference = "https://wiki/${name}"` must substitute the param into the URL. This is the cross-feature case the templates todo explicitly flags. If the templates feature ships *before* reference, this is forward-compatible (substitution covers any string field); if they ship together, test it explicitly.
2. **Include + templates (the motivating use case).** `templates.toml` defines `[template.microservice]`; `model.toml` does `[[include]] path = "templates.toml" once = true` then `[[use]] template = "microservice"`. The `once = true` is what makes this safe — without it, two model files both including the template library would double-define the template (hard error per §1). **This is the single most important end-to-end test in the milestone.**
3. **Template + relative-peer.** A templated unit's inline link has `peer = "${name}Bus"` (substitutes to e.g. `authBus`); relative-peer resolution must then resolve `authBus` against the enclosing parent. **Template expansion must run *before* relative-peer resolution** or the peer string is still `${name}Bus` at peer-resolution time and fails. The pipeline order is not stylistic — it is required for correctness.
4. **Include + UnitOrder.** The merged model's `UnitOrder` (load-bearing for rendering per `captureDefinitionOrder` at `parser.go:100`) must concatenate in include-site order. A template instantiated via `[[use]]` slots in at the use-site's position, not the template-definition's position.
5. **Optional-name + templates.** A templated unit can omit `name`; humanization runs on the instantiation key (e.g. `[[use]] name = "auth"` → unit `auth` → humanized "Auth"). The instantiation key, not the template's segment, is the humanization source. Confirm this in the humanize pass.
6. **Reference + render format.** The 📖 marker and the SVG `<a>` anchor must both render in `-f svg` and `-f html`. The HTML JS shim (Safari workaround) must handle external `reference` URLs distinctly from internal drill-down URLs.

### Pipeline ordering (confirmed by cross-feature analysis)

```
1. include          (assemble multiple files into one Model; cycle-detect; dedup via once=)
2. template-expand  (deep-copy + substitute ${param} into all string fields incl. Link.Peer, reference)
3. relative-peer    (resolve short peers against enclosing parent, walk up, absolute fallback)
4. humanize-names   (fill empty Name from segment; runs after expand so templated units get humanized)
5. validate         (peer-existence, level checks, conflict checks — on the fully assembled model)
6. generate views   (auto-derive C1/C2/C3 — unchanged)
7. render           (emit GraphViz URL attribute for reference; 📖 marker in label)
```

This matches the ordering in the include todo; the cross-feature analysis adds step 4 (humanize) as an explicit named pass between relative-peer and validate.

---

## 6. SUMMARY — Classification at a glance

### Table stakes (ship in v1.10; absence feels broken)

- **Include:** directive, transitive includes, path-relative-to-including-file, cycle detection, hard-error on duplicate unit path.
- **Templates:** named params, trailing defaults, error on missing required param, substitution into all string fields.
- **Ergonomics:** (none — both items are differentiators, not table stakes, because no comparable tool does them).
- **Reference:** per-element URL field, clickable SVG via GraphViz `URL` attribute, `reference` name (collision-free), HTML-shim coverage.

### Differentiators (nice-to-have; adds polish; ship if cost is low)

- **Include:** `once = true` (include_once) — *essential* for the template-library use case, not optional in practice.
- **Templates:** one-template-one-unit (the deliberate, documented deviation from multi-output precedent); `${uid}` defer.
- **Ergonomics:** relative-peer resolution (novel; high payoff); optional-name humanization (more aggressive than peers; high payoff); walk-up precedence.
- **Reference:** 📖 visible affordance (matches repo's 🔍 precedent; matches D2's icon pattern).

### Anti-features (do NOT copy; tied to C4Drill's philosophy)

- **Include:** Terraform-override-style silent merge (breaks "TOML is source of truth"); partial import / namespacing (adds collision layer for marginal gain); URL includes (network dep + security surface for a single-shot CLI).
- **Templates:** multi-output / `for_each`-style (fights structured-TOML; revisit only if 5-`[[use]]` pattern proves unbearable); nesting (resolution-order complexity); forward references (requires multi-pass); typed params (needs a type system).
- **Ergonomics:** acronym preservation (Terraform `title()` proves this is an unsolved tar pit); compact-link string shorthand (TOML-grammar rigidity; defer pending relative-peer measurement).
- **Reference:** URL over-validation (LikeC4 supports any URI scheme); link-specific tooltip (duplicates `description`); per-relationship URL (scope creep); image/background property (contrary to `dev/shapes-no-icons`).

### Single most relevant borrowed idiom per feature

| Feature | Most relevant borrowed idiom | Source |
|---|---|---|
| Include | `!include_once` (dedup for shared definition files) | [PlantUML preprocessing](https://plantuml.com/preprocessing) |
| Templates | Named params + trailing defaults (the modern idiom) | [PlantUML `!procedure`](https://plantuml.com/preprocessing), [Jsonnet language ref](https://jsonnet.org/ref/language.html) |
| Ergonomics | (no direct precedent — C4Drill is more aggressive than peers; `title()` is the closest but insufficient) | [Terraform `title()`](https://developer.hashicorp.com/terraform/language/functions/title) |
| Reference | `URL` node attribute → native SVG `<a>` anchor (zero-cost via existing backend) | [GraphViz URL attribute](https://graphviz.org/docs/attrs/URL/) |

---

## Sources

### Include / multi-file
- [Structurizr DSL — Includes](https://docs.structurizr.com/dsl/includes)
- [Structurizr DSL — Language reference](https://docs.structurizr.com/dsl/language)
- [D2 — Imports syntax](https://d2lang.com/tour/imports/)
- [D2 — Imports use cases](https://d2lang.com/tour/imports-use-cases/)
- [Terraform — Override files](https://developer.hashicorp.com/terraform/language/files/override)
- [Terraform — for_each meta-argument](https://developer.hashicorp.com/terraform/language/meta-arguments/for_each)
- [Terraform duplicate resource address error](https://oneuptime.com/blog/post/2026-02-23-how-to-fix-duplicate-resource-address-error-in-terraform/view)
- [CUE — Modules](https://cuelang.org/docs/reference/modules/)
- [CUE — `load` package (import cycle errors)](https://pkg.go.dev/cuelang.org/go/cue/load)
- [CUE — structural cycle detection (issue #990)](https://github.com/cue-lang/cue/issues/990)
- [Jsonnet — Language reference](https://jsonnet.org/ref/language.html)
- [Jsonnet — Specification](https://jsonnet.org/ref/spec.html)
- [PlantUML — Preprocessing (`!include`, `!include_once`, `!includesub`)](https://plantuml.com/preprocessing)
- [Puppet — node inheritance deprecation](https://itsecureadmin.com/2014/11/19/puppet-node-inheritance-deprecation/comment-page-1/)

### Templates / parametrized definitions
- [PlantUML — Preprocessing (`!procedure`, `!function`)](https://plantuml.com/preprocessing)
- [Jsonnet — Tutorial (functions, named params, defaults)](https://jsonnet.org/learning/tutorial.html)
- [Jsonnet — Richer functions / named args (issue #147)](https://github.com/google/jsonnet/issues/147)
- [Helm — Named Templates (`define`, `template`, `include`)](https://helm.sh/docs/chart_template_guide/named_templates)
- [Use Named Templates Like Functions in Helm](https://itnext.io/use-named-templates-like-functions-in-helm-charts-641fbcec38da)
- [Dhall — Language Tour](https://docs.dhall-lang.org/tutorials/Language-Tour.html)
- [CUE — Specifying a default value for a field](https://cuelang.org/docs/howto/specify-a-default-value-for-a-field/)
- [CUE — Scopes and Visibility](https://cuetorials.com/overview/scope-and-visibility/)
- [Terraform — count meta-argument](https://developer.hashicorp.com/terraform/language/meta-arguments/count)
- [Ansible — `include_vars` module](https://docs.ansible.com/projects/ansible/latest/collections/ansible/builtin/include_vars_module.html)
- [Structurizr DSL — Language reference (no parametrized element creation; `softwareSystemInstance`/`containerInstance`/`archetypes`)](https://docs.structurizr.com/dsl/language)

### Ergonomic shorthand / humanization
- [Terraform — `title()` function](https://developer.hashicorp.com/terraform/language/functions/title)
- [Terraform casing pitfalls (Stack Overflow)](https://stackoverflow.com/questions/65112736/in-terraform-is-it-possible-to-change-the-casing-of-values)
- [Terraform — Functions to convert identifier cases (issue #25230)](https://github.com/hashicorp/terraform/issues/25230)
- [Terraform — Naming conventions best practices](https://developer.hashicorp.com/terraform/plugin/best-practices/naming)
- [How to Use the `title` Function in Terraform](https://oneuptime.com/blog/post/2026-02-23-how-to-use-title-function-in-terraform/view)
- [CUE — Scopes and Visibility](https://cuetorials.com/overview/scope-and-visibility/)

### Element hyperlinks
- [GraphViz — `URL` attribute](https://graphviz.org/docs/attrs/URL/)
- [GraphViz — Attributes reference (`URL`, `href`, `tooltip`, `target`)](https://graphviz.org/doc/info/attrs.html)
- [Structurizr DSL — Language reference (`url` keyword)](https://docs.structurizr.com/dsl/language)
- [LikeC4 — Model (`link` keyword, URI schemes, relative source links)](https://likec4.dev/dsl/model/)
- [LikeC4 — Add Documentation & ADR's (discussion #1458)](https://github.com/likec4/likec4/discussions/1458)
- [D2 — Interactive (tooltips & links)](https://d2lang.com/tour/interactive/)
- [D2 — Release 0.1.4 (link/tooltip icon indicator)](https://d2lang.com/releases/0.1.4/)
- [PlantUML — Links (`[[url]]`, `{tooltip}`, triple-bracket for methods)](https://plantuml.com/link)
- [Mermaid — Clickable links in nodes (GitHub issue #6314)](https://github.com/mermaid-js/mermaid/issues/6314)
- [Clicking Entities in Mermaid (SVG anchor wrapping)](https://monogon.net/mermaid-click-writeup/)
- [SVG `<a>` element (MDN)](https://developer.mozilla.org/en-US/docs/Web/SVG/Reference/Element/a)
