# Architecture Research: v1.10 Model Composition (include / templates / relative-peer / reference)

**Scope:** How the four new v1.10 features integrate into the existing C4Drill pipeline.
**Researched:** 2026-08-08
**Confidence:** HIGH — every claim cited to `file:line` in the current tree (HEAD = `58e1832`).

This document is the **code-architecture** companion to the four pending todos under
`.planning/todos/pending/2026-08-08-*.md`. It is deliberately concrete: it does not
re-litigate the design options (those are settled in the todos) — it pins down *where in
the existing code each step hooks in* and *what struct changes are required*.

> Note: the pre-existing `.planning/research/ARCHITECTURE.md` is the original v1.0
> domain-research doc (C4-tool comparison, 2026-03-09) and is intentionally left
> untouched. This v1.10 doc follows the established `ARCHITECTURE-v1.1.md` versioning
> convention.

---

## Current Pipeline (file:line trace)

The pipeline is a straight line orchestrated in `cmd/c4drill/root.go:runRoot`
(`root.go:85`). There is exactly one entry file and one `*parser.Model` flows through
every stage.

| Stage | Call | Location | Consumes | Produces |
|---|---|---|---|---|
| 0. Flag guard | `format`/`outDir` setup | `root.go:96-109` | CLI args | — |
| 1. **Parse** | `parser.ParseFile(inputPath)` | `root.go:112` | TOML file path | `*parser.Model` |
| 2. **Validate** | `validator.Validate(m)` | `root.go:118` | `*parser.Model` | `ValidationErrors` |
| 3. Path collect | `collectExpandedPaths(m)` | `root.go:140` → `:156` | `m.Units` (recurses `Subunits`) | `[]string` unit paths |
| 4-6. Per-path | `processView(m, …)` | `root.go:144` → `:203` | `*parser.Model` + path | written file |
| 4. View gen | `view.GenerateC1View` / `C2` / `C3` | `root.go:209-213` | `m.Units`, `m.UnitOrder`, `m.Properties` | `*view.View` |
| 5. Graph build | `graph.BuildGraphWithPath` | `root.go:221` | `*view.View` | `*graph.Graph` |
| 6. **Render** | `render.RenderSVGWithOutput` / `RenderHTML` / `Render` | `root.go:235-239` | `*graph.Graph` | `[]byte` |

The `--expanded` branch (`root.go:135`, `processExpandedView` `:256`) bypasses stages 3-4
and calls `view.GenerateExpandedView` (`scope.go:14`) directly, but still runs stage 1
(Parse) and stage 2 (Validate) on the same `*parser.Model`.

### Parse internals (`internal/parser/parser.go`)

`ParseFile` (`parser.go:322`) reads bytes then calls `Parse` (`parser.go:47`), which is a
three-pass pipeline:

1. **`captureDefinitionOrder`** (`parser.go:100`, called at `:49`) — walks the
   `pelletier/go-toml/v2/unstable` AST to record, in file order: top-level unit names
   (`unitOrder`, `:132-137`) and nested `[parent.child]` subunit names
   (`subunitOrders`, `:139-147`). `[properties]` is explicitly skipped at `:128`. Deeper
   nesting (`len(parts) > 2`) is ignored (`:149`).
2. **`toml.Unmarshal` into `map[string]any`** (`parser.go:57`) — raw map.
3. **Build `Model`** (`parser.go:62-93`) — extract `properties` (`:68-77`), then walk
   `unitOrder` calling `parseUnitWithOrder` (`:87` → `:160`). Subunits discovered via
   the captured order, with a map-iteration fallback at `:217-237` gated by
   `isBuiltinField` (`:309-316`).

### Data shapes

- **`parser.Model`** (`parser.go:35-42`): `Properties model.Properties`, `UnitOrder []string`, `Units map[string]*model.Unit`.
- **`model.Unit`** (`unit.go:41-72`): `Type, Name, Description, Technology, Color, Style, Border, Edges, Width, Height, Expanded []string, Links []Link, LinksFrom []Link, SubunitOrder []string (toml:"-"), Subunits map[string]*Unit (toml:",inline")`.
- **`model.Link`** (`link.go:43-68`): `Peer, Arrow, Rank, Color, Style, Technology, Description, LabelPosition, Length, Mirror (toml:"-")`.
- **`model.Properties`** (`properties.go:4-21`): `Name, Description, Color, Style, Border, Edges, LineLength, Expanded`.
- **`parser.ParseError`** (`internal/parser/errors.go:13-23`): `Message, Line, Context, Cause` — the error type new passes should return (it carries line numbers via `wrapDecodeError`, `errors.go:45`).

### Consumer invariants (what each downstream stage actually reads)

- **Validator** (`validator.Validate`, `validator.go:16`) reads **only `m.Units`**
  (`BuildIndex(m.Units, "")` at `validator.go:22`; `index.go:23`). It never touches
  `UnitOrder` or `Properties`. Rules in `rules.go` operate on the flat index.
- **View generators** (`scope.go`) read `m.Units`, `m.UnitOrder` (with map-key fallback
  at `:34`, `:106`, `:244`, etc.), and `m.Properties.Name`/`.Expanded`/`.Edges`
  (`:22`, `:94`, `:184`).
- **Render** (`render/converter.go`) consumes **`*graph.Graph`**, never `*parser.Model`.
  The graph is built from the view, which is built from the model. So render is fully
  insulated from any `Model` struct change.

---

## Extended Pipeline

The four features form a **pre-processing chain inserted between Stage 1 (Parse) and
Stage 2 (Validate)**. They are pure passes over `*parser.Model`: each takes a model and
returns a model. The post-chain model is structurally identical to a hand-authored
single-file model, so Validate and the view generators consume it **unchanged**.

```
ParseFile ──▶ include.Resolve ──▶ template.Expand ──▶ peer.Resolve ──▶ Validate ──▶ views ──▶ render
  (root.go:112)   (NEW, ~:113)     (NEW, ~:114)       (NEW, ~:115)    (:118)      (:209)     (:235)
```

### Concrete insertion in `runRoot` (`root.go:111-118`)

Today:
```go
// Stage 1: Parse
m, err := parser.ParseFile(inputPath)   // root.go:112
if err != nil { return fmt.Errorf("parse: %w", err) }

// Stage 2: Validate
valErrors := validator.Validate(m)      // root.go:118
```

Extended (three new lines slotted between `:115` and `:117`):
```go
m, err := parser.ParseFile(inputPath)
if err != nil { return fmt.Errorf("parse: %w", err) }

// Stage 1a: Resolve includes (recursive, merges *parser.Model structs)
if m, err = include.Resolve(m, filepath.Dir(inputPath)); err != nil {
    return fmt.Errorf("include: %w", err)
}
// Stage 1b: Expand [[use]] instantiations (deep-copy + param substitution)
if m, err = template.Expand(m); err != nil {
    return fmt.Errorf("template: %w", err)
}
// Stage 1c: Resolve relative Link.Peer to absolute paths
if m, err = peer.Resolve(m); err != nil {
    return fmt.Errorf("peer: %w", err)
}

// Stage 2: Validate (UNCHANGED — consumes assembled+expanded model)
valErrors := validator.Validate(m)
```

**Ordering rationale** (matches `.planning/todos/pending/2026-08-08-include-directive-multi-file-diagrams.md:88-96`):

1. **include first** — so templates defined in an included file are visible to the root
   model's `[[use]]` instantiations (the "template isolation" use case).
2. **template-expand second** — so peers *inside* a templated unit exist before
   relative-peer resolution runs; otherwise a templated unit's relative links would
   resolve against a model missing the very unit being expanded.
3. **relative-peer-resolve third** — needs the fully assembled + expanded model to look
   up peer targets (peers can be forward references across files/templates).
4. **validate last** — peer-existence (`rules.go:14`) and level checks (`rules.go:187`)
   must see the final, concrete unit set.

**Must validate/view-gen consume the model UNCHANGED?** Yes — and this is the central
design constraint. `validator.Validate` (`validator.go:16`) and the three view generators
(`scope.go:86/382/454`) take `*parser.Model` and read `.Units`/`.UnitOrder`/`.Properties`.
As long as the three new passes produce a `*parser.Model` whose `.Units` contains only
real renderable units (no template skeletons, no `[[use]]` stubs, no include directives),
the downstream code needs **zero changes**. The passes achieve this by moving
directives/templates/instantiations into dedicated `Model` fields that downstream code
does not read, then draining those fields into `.Units`/`.UnitOrder` as concrete units.

---

## Model Extension (recommended approach)

**Recommendation: dedicated fields on `parser.Model`, filtered out of the unit namespace
inside `Parse`. Do not use sidecar structs.**

Sidecar structs (`IncludeGraph`, `TemplateLibrary`, …) would force changing the signature
of `validator.Validate`, the three `view.Generate*` functions, `collectExpandedPaths`
(`root.go:156`), and every caller — a wide blast radius for no gain. Dedicated fields on
`parser.Model` are invisible to any consumer that reads only `.Units`/`.UnitOrder`/
`.Properties`, which is exactly the consumer invariant established above.

### Proposed `parser.Model` extension (`parser.go:35-42`)

```go
type Model struct {
    Properties model.Properties `toml:"properties"`
    UnitOrder  []string
    Units      map[string]*model.Unit

    // --- v1.10 composition fields (not read by validator/view/render) ---
    // Includes is the list of [[include]] directives, captured in file order.
    // Consumed by include.Resolve; empty after resolution.
    Includes []IncludeDirective `toml:"-"`   // populated specially in Parse
    // Templates is the set of [template.<name>] definitions.
    // Consumed by template.Expand; emptied (or left) after expansion.
    Templates map[string]*TemplateDef `toml:"-"`
    // Instantiations is the list of [[use]] requests, in file order.
    // Consumed (drained) by template.Expand.
    Instantiations []Instantiation `toml:"-"`
}
```

All four new fields are `toml:"-"` because they are **not** unmarshalled by go-toml
directly into the struct — `Parse` extracts them from the raw map (`parser.go:57`) before
the unit loop (`parser.go:80`) runs, exactly as it already extracts `properties`
(`parser.go:68-77`). This keeps them out of the unit namespace.

The new types live in `internal/parser/` (or a small `internal/compose/` package the
parser populates):

```go
type IncludeDirective struct {
    Path     string // relative to including file's dir
    Once     bool
    Optional bool
}

type TemplateDef struct {
    Name   string
    Params map[string]string // name → default; key absent ⇒ required
    Unit   *model.Unit       // body, with ${param} placeholders in string fields + Link fields
}

type Instantiation struct {
    Template string            // template name
    As       string            // unit key to create (top-level) — or
    Parent   string            // optional parent path for placement
    Args     map[string]string // named arg values
}
```

### Why this keeps consumers unchanged

- `validator.Validate` reads `m.Units` only (`validator.go:22`) → new fields invisible.
- `view.Generate*` read `m.Units`/`m.UnitOrder`/`m.Properties` → new fields invisible.
- `collectExpandedPaths` reads `m.Units` (`root.go:158`) → invisible.
- `render` reads `*graph.Graph` → fully insulated.

The only existing function that must learn about the new tables is `Parse`
(`parser.go:47`) itself, plus `captureDefinitionOrder` (`parser.go:100`) so the new
top-level tables are not misclassified as units.

---

## Per-Feature Integration

### (2)/(3) Include directive — `internal/include/` (new package)

**Parse-side changes:**

- `captureDefinitionOrder` (`parser.go:100`): add skips alongside the `properties` skip
  at `:128` for `include`, `template`, `use` (any `parts[0]` in a reserved set). Without
  this, `[include]` becomes a top-level unit named "include".
- `Parse` (`parser.go:47`): after extracting `properties` (`:68-77`), extract the
  `include` table from `rawMap` into `m.Includes` and **delete it from `rawMap`** before
  the unit loop (`:80`). This guarantees the directive never enters `m.Units`.

**New package `internal/include/`:**

```go
func Resolve(m *parser.Model, baseDir string) (*parser.Model, error)
```

- Walks `m.Includes` in order; for each, resolve `Path` relative to `baseDir`
  (`filepath.Dir(inputPath)` passed from `root.go`), read+`parser.ParseFile` the included
  file (recursion), maintain an include-stack for cycle detection (go-metadot
  `@incChain` precedent; cap depth at 100), honor `Once` (PlantUML `!include_once`
  precedent) via a seen-set.
- **Merge** (`internal/include/merge.go`): merge two `*parser.Model` structs:

  | Field | Rule | Code site |
  |---|---|---|
  | `Units` (map) | Union; **same key in two files = hard error** (returns `*parser.ParseError` with file context) | new `mergeUnits` |
  | `UnitOrder` (slice) | **Concatenate in include order** (root first, then each included file's order appended at the include site) | `merged.UnitOrder = append(dst.UnitOrder, src.UnitOrder...)` |
  | `Properties` | **Root/first-seen wins**; included files may not override non-zero `Properties` fields (error, or warn) | new `mergeProperties` |
  | `Templates` | Union; same name in two files = hard error | new `mergeTemplates` |
  | `Includes` | Recursively resolved; drained to empty on the merged model | flatten then clear |

- Recurse: `Resolve` on the included model (with the included file's dir as the new
  `baseDir`) before merging, so transitive includes resolve depth-first.

**`captureDefinitionOrder` question (research item 3):** it must run **per-file**
(it already does — it is called inside `Parse` at `:49`, once per file). Merge then
**concatenates the already-ordered `UnitOrder` slices**. It does **not** re-run on the
merged model; that would lose per-file authoring order and is unnecessary because each
file's order is already captured. The merge's only ordering decision is the inter-file
sequence, which is the include-site order (appended).

**No changes** to `validator/rules.go`, `view/scope.go`, or `render/`.

### (4) Template expansion — `internal/template/` (new package)

**Parse-side changes:**

- `captureDefinitionOrder` (`parser.go:100`): skip `parts[0] == "template"` (add to the
  reserved set above) so `[template.microservice]` is not recorded as a unit named
  "template".
- `Parse` (`parser.go:47`): extract the `template.*` subtable from `rawMap` into
  `m.Templates` (keyed by the second path segment, e.g. `microservice`), extract `[[use]]`
  arrays into `m.Instantiations` (in file order), and **delete both from `rawMap`** before
  the unit loop (`:80`). This is the same pattern as `properties` extraction at `:68-77`.
- `isBuiltinField` (`parser.go:309-316`): **no change required** for templates — the
  `params`/body fields of a template Unit are already in the allowlist (`type`, `name`,
  `description`, …). A `parent` field on `[[use]]` is not a Unit field (it lives on
  `Instantiation`), so it never reaches `isBuiltinField`.

**New package `internal/template/`:**

```go
func Expand(m *parser.Model) (*parser.Model, error)
```

- For each `Instantiation` in `m.Instantiations` (file order):
  1. Look up `Template` in `m.Templates`; error (`*parser.ParseError`) if missing.
  2. Validate args: every required param (no default) supplied; unknown args error.
  3. **Deep-copy** `tmpl.Unit` (must deep-copy nested `Links`/`LinksFrom` slices and
     `Subunits` so instantiations don't alias). A `deepcopy` helper belongs in this
     package (or `internal/model/`).
  4. **Substitute** `${param}` → value in every string field of the copy —
     `Name, Description, Technology, Color, Style, Border, Edges` and, for each `Link`,
     `Peer, Description, Technology` (the ergonomic todo at
     `2026-08-08-toml-authoring-ergonomic-improvements.md` and the template todo at
     `:139` both call this out). Plain `strings.ReplaceAll` is sufficient (Option B in
     the template todo; deliberately *not* `text/template`).
  5. **Insert** the concrete unit: if `Parent` is set, place it under
     `m.Units[parent].Subunits[as]` and append `as` to that unit's `SubunitOrder`
     (mirroring `parseUnitWithOrder` at `parser.go:215-216`); otherwise top-level at
     `m.Units[as]` and append `as` to `m.UnitOrder`.
- Clear `m.Instantiations` at the end (they are consumed). Leave or clear
  `m.Templates` (clearing is cleaner; templates are not renderable units).

After `Expand`, the model's `.Units`/`.UnitOrder` contain only concrete units, so the
validator sees them as normal. The validator's nesting check (`rules.go:187`) naturally
applies to a templated unit placed inside a parent — exactly the desired behavior.

### (5) Relative-peer resolution — `internal/peer/` (new package)

**Hook:** after `template.Expand`, before `validator.Validate` (see Extended Pipeline).

**New package `internal/peer/`:**

```go
func Resolve(m *parser.Model) (*parser.Model, error)
```

- Build a set of all unit paths by walking `m.Units` recursively (reuse the walk pattern
  from `validator.BuildIndex`, `index.go:23`, or `view.findUnitByPath`, `scope.go:541`).
- For each unit (recursive), for each `Link` in `Links` and `LinksFrom`: if `link.Peer`
  is **not** already an existing absolute path, resolve relative-to-enclosing-parent:
  - Try `currentParent + "." + peer`.
  - Else walk up: `grandparent + "." + peer`, … to top level.
  - Else if `peer` exists as a top-level unit, use it.
  - Else leave `peer` unchanged (absolute fallback — full backward compat; the validator's
    undefined-reference rule at `rules.go:23` then reports it as before).
- **Requires the full assembled model** (yes — include + template must have run, because
  a relative peer may target a unit that came from an included file or a template
  instantiation). This is why it is stage 1c, not 1a.
- Mutates `link.Peer` in place; returns the same `*parser.Model`.

This pass is a strict superset of current behavior for absolute paths (they resolve
unchanged), so existing models are unaffected.

### (6) Reference field (📖) — touches `model`, `parser`, `graph`, `render`

This is the lowest-risk feature and has **no pipeline dependency** — it can ship in any
phase. Touch points, in order of dependency:

1. **`internal/model/unit.go:41-72`** — add field to `Unit` struct:
   ```go
   // Reference is an optional external documentation URL.
   Reference string `toml:"reference"`
   ```
   (placed near `Description`/`Technology`).

2. **`internal/parser/parser.go:309-316`** — `isBuiltinField` allowlist; add `"reference"`
   to the `slices.Contains` literal. **Without this**, the parser treats
   `reference = "…"` as a subunit named "reference" via the fallback at `:217-237`.

3. **`internal/graph/graph.go:55-71`** — `Node` struct already has `ExploreURL` (`:69`,
   used for drill-down via `converter.go:268-270` `cn.SetURL`). Add a sibling:
   ```go
   // ReferenceURL is the external docs URL (📖), empty if none.
   ReferenceURL string
   ```
   The `Label` struct (`graph.go:123-132`) can also gain a `Reference string` if the 📖
   glyph is rendered as a label cell (option B below).

4. **`internal/graph/builder.go:250-279`** — `buildNode`: populate `ReferenceURL` from
   `entry.Unit.Reference` (and/or carry into `Label`). This is the **direct precedent
   site** for the 🔍 magnifier (`builder.go:258-261`):
   ```go
   if entry.HasSubunits && !entry.IsExpanded {
       label.Name += " 🔍"           // existing, builder.go:260
   }
   // NEW:
   if entry.Unit.Reference != "" {
       label.Name += " 📖"           // simplest option A: glyph in name
       node.ReferenceURL = entry.Unit.Reference
   }
   ```
   `buildClusterLabel` (`builder.go:331`) needs the same treatment for expanded parents.

5. **`internal/render/converter.go:267-272`** — `createNode` currently sets `cn.SetURL`
   for drill-down. **Design nuance:** a GraphViz node has a single `URL` attribute, so a
   unit with *both* subunits (drill-down) and a reference cannot have two node-level
   links. Two options:
   - **Option A (simplest, matches 🔍):** append 📖 to the label name
     (`builder.go:260`-style) and leave drill-down as the node URL. The 📖 is then purely
     visual; the link is not clickable from the glyph. Ships fast; matches the
     "emoji-as-indicator is an established pattern" note in the reference todo.
   - **Option B (clickable, richer):** render a dedicated `<TD HREF="…">📖</TD>` cell
     inside the HTML label, following the navigation-TD precedent at
     `converter.go:202` (`navigationTDs`). This requires touching **every** HTML label
     builder in `internal/render/labels.go` (`buildSystemHTMLLabel:231`,
     `buildContainerHTMLLabel`, `buildPersonHTMLLabel:72`, `buildDbHTMLLabel:130`,
     `buildQueueHTMLLabel:178`, `buildComponentHTMLLabel`, `buildBoxHTMLLabel`) — a wider
     change but the only way to get a per-glyph clickable anchor in SVG.

   **Recommendation:** ship Option A first (Phase 28), upgrade to Option B in a later
   phase if the clickable anchor is wanted. The 🔍 precedent at `builder.go:260` is
   exactly Option A.

6. **Docs/examples** — `README.md` (Unit fields table), `skill/SKILL.md` (schema
   reference), and a fixture in `skill/examples/` / `cmd/c4drill/testdata/`. The
   reference todo notes the README link-syntax section is stale — leave it, only add
   `reference`.

---

## Suggested Phase Breakdown

v1.9 shipped as **Phase 27** (`ROADMAP.md:8`). The next phases continue that numbering.
Dependency graph for *correctness at runtime* is `include → template → peer → validate`,
but the **build/ship order** prioritizes low-risk independent work first.

| Phase | Name | Goal | What ships | Depends on |
|---|---|---|---|---|
| **28** | **Reference field (📖)** | External-docs URL on units | `Unit.Reference`; `isBuiltinField` update; `buildNode` glyph (Option A); `Node.ReferenceURL`; README + skill + example. | None — fully independent, parallelizable. |
| **29** | **Optional `name` humanization** *(from ergonomic todo, low-risk, parallelizable)* | Derive display name from last path segment when `name` omitted | Humanize function in parser/model; `Unit.Name` fallback at parse time; docs. | None — parallelizable with 28. |
| **30** | **Relative-peer resolution** | Short `peer` names resolve relative to enclosing parent | `internal/peer/Resolve`; hook in `runRoot` stage 1c; relative-first/absolute-fallback; tests + saira re-measure. | None (ships as a no-op pass for absolute-only models). |
| **31** | **Template expansion** | `[template.*]` define + `[[use]]` instantiate | `Model.Templates`/`.Instantiations` fields; `captureDefinitionOrder` skip for `template`/`use`; `internal/template/Expand` (deepcopy + `${param}` substitution into Unit + Link fields); drain instantiations into `Units`/`UnitOrder`. | Benefits from 30 (templated links can be relative); can ship with absolute peers if 30 slips. |
| **32** | **Include directive (multi-file)** | Assemble model from multiple files | `Model.Includes` field; `captureDefinitionOrder` skip for `include`; `internal/include/Resolve` (recursive merge, cycle detection, `Once`); merge rules (Units union/conflict-error, `UnitOrder` concatenate, Properties root-wins). | 31 — the "template isolation" use case (include a template library) is the primary motivation; include must carry `Templates` through the merge. |
| **33** | **Docs sweep + end-to-end golden tests** | Document composition features; golden comparisons | Update `README.md`/`skill/SKILL.md`; add `skill/examples/06-templates.toml`, `07-include/`; golden test: 3-file model produces same SVG as equivalent single-file model. | 30-32. |

**Parallelization note:** Phases 28 and 29 touch disjoint code (`model/unit.go` +
`graph/builder.go` + `render` vs. parser humanization) and can proceed in parallel. Phases
30-32 are sequential due to the runtime ordering constraint, but each is independently
shippable (each pass is a no-op for models that don't use the feature).

---

## Unchanged Consumers (confirmation)

Confirmed by reading each consumer's actual field accesses:

- **`validator.Validate`** (`validator.go:16`) → calls `BuildIndex(m.Units, "")`
  (`validator.go:22`). Reads **`m.Units` only**. All rules in `rules.go` operate on the
  index. **No change needed** for any of the four features; it naturally validates the
  assembled+expanded model.
- **`view.GenerateC1View`/`C2`/`C3`/`ExpandedView`** (`scope.go:86/382/454/14`) → read
  `m.Units`, `m.UnitOrder` (with map-key fallback), `m.Properties`. **No change needed.**
- **`render.*`** (`converter.go`, `labels.go`, `render.go`) → consume `*graph.Graph`,
  never `*parser.Model`. **No change needed** (except the reference-field glyph, which is
  a *new* feature, not a forced change).
- **`collectExpandedPaths`** (`root.go:156`) → reads `m.Units`. **No change needed.**

The only existing function whose behavior must change is **`parser.Parse`**
(`parser.go:47`) and **`captureDefinitionOrder`** (`parser.go:100`), to extract the new
top-level tables (`include`, `template`, `use`) so they do not pollute the unit
namespace — directly analogous to the existing `properties` extraction at `:68-77` and
the `properties` skip at `:128`.

**Net: the four features are pure pre-processing passes + a reference-field rendering
addition. Validator, view, and render are unchanged for include/template/relative-peer.**
