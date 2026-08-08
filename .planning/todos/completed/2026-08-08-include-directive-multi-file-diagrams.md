---
created: 2026-08-08T00:53:13.640Z
title: Include directive — build diagrams from multiple input files
area: tooling
resolves_phase: 32
status: completed
resolved_at: 2026-08-08
resolution: "Shipped in Phase 32 (v1.10). internal/include/Resolve + merge implemented; INC-01..10 + XC-02 all validated. Decisions D-09 (append ordering), D-10 (cross-file subunits), D-11 (same-file diamond dedup), D-12 (hard-error on missing include) all realized."
files:
  - internal/parser/parser.go
  - cmd/c4drill/root.go
  - internal/model/unit.go
  - internal/validator/rules.go
---

## Problem

C4Drill currently takes a single input TOML file (`cmd/c4drill/root.go` single-path CLI; `internal/parser/parser.go:322` `ParseFile`). For large architectures this becomes unwieldy — hundreds of units in one file. The user wants an `include` directive so a diagram can be assembled from multiple input files. Two stated motivations:
1. **Large diagrams** — split a big model across files (e.g. one file per system/domain), assemble at build time.
2. **Template isolation** — keep reusable unit templates (see `2026-08-08-unit-templates-parametrized-definitions.md`) in their own file(s), included by the models that use them, so model files stay clean.

### Reference implementations

Both reference tools have mature include mechanisms; both are line-oriented DSLs, so their syntax is suggestive but the C4Drill adaptation must account for structured TOML.

**go-metadot** (`metadot.pl:839-864` `registerInclude`):
- `include <path>` directive at top level.
- Recursively calls `readFile` on the included file — included content is processed inline at the include site, in order.
- Runs `doSubst` (the `${N}` macro substitution) on each included line, so included files inherit the caller's current `$params` (CLI args). Relevant if templates land.
- **Cycle detection** via `@incChain` (a stack of files currently being included); re-including a file in its own ancestry is a fatal error.
- Paths resolved relative to the including file's directory.

**PlantUML** ([preprocessing docs](https://plantuml.com/preprocessing)):
- `!include foo.puml` — splice file content. `!include_many` / `!include_once` variants control dedup.
- `!includesub`, `!includeurl`, `!include <searchdir:file>` extensions for namespacing, URLs, search paths.
- Resolves name conflicts with `!includesub` (strip/keep/rename sub-diagram prefixes).
- The `!include_once` idiom is the key lesson: template/definition files must be safely includable from multiple model files without double-definition errors.

## Solution

### Open design question (resolve before implementing)

C4Drill parses TOML into Go structs via `pelletier/go-toml/v2`. An include directive can be realized at one of two layers. **A decision is required before implementation.**

**Option A — Pre-parse byte concatenation:** process includes by reading the referenced files and concatenating their raw bytes into a single TOML document before handing to go-toml. Pros: simplest mental model (one logical file); templates, relative peers, everything sees one unified namespace. Cons: **TOML does not naively concatenate** — two files each starting with `[properties]` collide; dotted-table paths `[a.b]` in different files merge correctly (TOML spec allows), but duplicate explicit tables (`[a]` in two files) error. Would need a merge strategy (deep-merge tables, error on scalar conflicts) that essentially reinvents part of a TOML parser. Fragile.

**Option B — Parse-then-merge structs (RECOMMENDED):** parse each included file independently into its own `parser.Model`, then deep-merge the `Model` structs: union `Units` maps, concatenate `UnitOrder` slices (in include order), merge `[properties]` with explicit conflict resolution. Pros: type-safe; each file is independently valid TOML; merge logic is a small, testable Go function; preserves error fidelity (errors point to the originating file). Cons: must define merge semantics for every field — especially `UnitOrder` (authoring order is load-bearing for rendering, per `captureDefinitionOrder` at `parser.go:100`) and `properties.name` (which file wins?). 

**Recommendation: Option B.** Define clear merge rules (see below) and implement the merge as a post-parse pass. Captures the "multiple files assemble into one model" semantics without fighting TOML's own table-merge rules.

### Merge semantics to pin down (Option B)

| Field | Merge rule |
|---|---|
| `Units` (map) | Union. **Same unit path in two files = hard error** (avoid silent shadowing), UNLESS one is a template and templates are namespaced — decide. |
| `UnitOrder` (slice) | Concatenate in include order: root file's units first, then each included file's units in the order they appear. The directive's position in the root file determines where included units slot in the order (or always appended — decide). Recommend: **appended in include-site order** (simplest; matches go-metadot inline-splice behavior). |
| `properties` | First-seen wins for `name`/`description` (root file is authoritative); or error on conflict. Recommend: **root file wins, included files may not override `[properties]`** (clear rule, avoids surprises). |
| Template defs (`[template.*]` from the templates todo) | Union; same template name in two files = hard error (like go-metadot's duplicate-define) OR last-wins with a warning. Recommend hard error for safety. |

### Directive syntax (TOML-compatible candidates)

TOML has no native include. The directive is a conventional field/table C4Drill interprets. Options:

```toml
# Option 1: dedicated table, evaluated pre-parse
[include]
files = ["templates/common.toml", "domains/auth.toml"]

# Option 2: array-of-tables with per-include options (PlantUML-inspired)
[[include]]
path = "templates/common.toml"
once = true   # include_once semantics — safe to re-include

[[include]]
path = "domains/auth.toml"

# Option 3: top-level array shorthand
include = ["templates/common.toml", "domains/auth.toml"]
```

**Lean toward Option 2** (array-of-tables) — it scales to `once`/`optional`/`namespace` flags later without re-designing, and PlantUML's include variants are precedent that those flags are wanted in practice. Resolve exact syntax in the plan.

### Touch points

- **`cmd/c4drill/root.go`** — CLI still takes one entry file; the include directive is resolved from that entry, transitively. (Alternative: accept multiple files on the CLI as an implicit top-level include — consider but don't require.)
- **`internal/parser/parser.go`** — new pre-parse or post-parse orchestration: read entry file → detect includes → recursively parse included files → merge `Model` structs. `Parse` (`:47`) / `ParseFile` (`:322`) signatures may gain a variant that returns the merged model. `captureDefinitionOrder` (`:100`) must produce a coherent `UnitOrder` across the merge.
- **`internal/validator/rules.go`** — runs after merge, so peer-existence (`:14`) and level checks (`:187`) naturally apply to the assembled model. New validation: include-path existence, cycle detection, unit-path conflict detection across files.
- **`internal/model/unit.go`** — no struct change needed for include itself (merge happens at the `Model` level).
- **README.md / skill/SKILL.md** — document the directive, merge rules, and the "one file per domain" / "template library" patterns.

### Resolution pipeline ordering (critical — interacts with other todos)

Includes must resolve at the right point relative to the other pending features:

1. **Include** (this todo) — assemble multiple files into one `Model`. **Runs first.**
2. **Template expansion** (`2026-08-08-unit-templates-parametrized-definitions.md`) — deep-copy + substitute params. **Runs after include**, so templates defined in an included file are visible to the root model's instantiations. This is the "template isolation" use case.
3. **Relative-peer resolution** (`2026-08-08-toml-authoring-ergonomic-improvements.md`) — resolve short `peer` names. **Runs after template expansion**, so peers inside a templated unit resolve correctly.
4. **Validation** (`rules.go`) — **runs last**, after the model is fully assembled and expanded.

If all four land, the pipeline is: `include → template-expand → relative-peer-resolve → validate → generate views → render`. Capture this ordering in the eventual plan.

### Cycle detection & safety

- Maintain an include stack (list of absolute file paths in the current include chain). Re-entering a file on the stack = fatal error (go-metadot `@incChain` precedent).
- `once = true` (include_once): a file included once is skipped on subsequent includes (PlantUML `!include_once`). Essential for template-library files that many model files include.
- Path resolution: relative to the including file's directory (both references do this). Support absolute paths too.
- Max include depth: cap (e.g. 100, matching go-metadot's recursion cap) to catch pathological include graphs.

### Verification

- Unit tests: two-file include produces merged model with correct `UnitOrder`; cycle detection errors; `once` dedup; unit-path conflict across files errors; missing file errors (or `optional=true` skips).
- End-to-end: a model split across 3 files (root + 2 includes) produces the same SVG as the equivalent single-file model (golden comparison, order-insensitive per existing canonicalDOT approach).
- Template-isolation test: `templates.toml` (defines a template) + `model.toml` (includes it, instantiates) produces the same output as a single-file version with the template inline.

## Related

- **Unit templates** (`2026-08-08-unit-templates-parametrized-definitions.md`) — the primary motivation for include. Template files must be includable; the `once` flag prevents double-definition when shared. Include resolves before template expansion.
- **TOML ergonomic improvements** (`2026-08-08-toml-authoring-ergonomic-improvements.md`) — relative-peer resolution runs after include + template expansion in the pipeline ordering above.
