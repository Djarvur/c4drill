# Phase 35: Add a simple DSL alternative to the TOML diagram definition (likec4/d2-style, less verbose syntax) with converters to and from TOML - Context

**Gathered:** 2026-08-14
**Status:** Ready for planning

<domain>
## Phase Boundary

This phase delivers the **C4D format** — a `.c4d` brace-block DSL (D2-inspired, less verbose than TOML) with **full feature parity** to the existing TOML authoring format — plus:

1. **C4D parsing + direct rendering** — `c4drill diagram.c4d` runs the full pipeline (DSL → `*parser.Model` → include → expand → peer → validate → views → render).
2. **Bidirectional converters** — `c4drill convert to-toml` / `c4drill convert to-c4d`, with canonical-equivalent round-trip as the parity contract.
3. **gofmt-style formatter** — `c4drill fmt` for both `.c4d` and `.toml`, in-place rewrite + `--check` CI mode, comment-preserving.
4. **TOML format extensions** — nested template instantiation in BOTH formats: `[[unit.use]]` in TOML, `use` inside unit blocks in C4D; `use` inside template bodies (template nesting — lifts the v1.10 deferral).
5. **Full docs & examples** — README C4D section, `.c4d` twins of key example fixtures, plugin SKILL.md covering both formats.

Out of scope: new diagram semantics (no custom kinds, tags, icons, user-authored views — the v1.10 "no LikeC4 feature adoption" stance holds; C4D is a SYNTAX alternative only).

</domain>

<decisions>
## Implementation Decisions

### DSL syntax style
- **D-01:** Brace blocks, D2-inspired — explicit `{ }` nesting, arrows inside blocks, whitespace-insensitive (NOT indentation-significant).
- **D-02:** Unit header form `id: type "Name" { ... }` — header carries id + type + display name; description/technology/reference/styling live inside the block. Type is omittable (inferred, per the existing hierarchy-position inference).
- **D-03:** Type keywords are the EXACT TOML type names (`system`, `person`, `db`, `queue`, `box`, ...) — 1:1 converter mapping, no `System`/`Database` prettification.
- **D-04:** External units via `external` modifier: `system external`, `person external` — scales to any type; maps to `personExternal`/`systemExternal`/etc.
- **D-05:** ASCII arrows only: `->` `<->` `<-` and `--` (no arrowhead). No Unicode glyphs.
- **D-06:** Literals: barewords when unambiguous (no `{ } : |`, no edge whitespace), double quotes otherwise, triple-quoted multi-line strings, `#` line comments. URLs with scheme prefix (`https://...`) are valid barewords (for `reference`).
- **D-07:** Identifiers = TOML bare-key character class (`[A-Za-z0-9_-]+`); dotted paths in peer refs use `.`.

### Edges and linkFrom
- **D-08:** Edge statements live INSIDE unit blocks only (no top-level/freestanding arrows). `-> peer` = outgoing (`link`); `<- peer` = incoming (`linkFrom`). User's rationale (verbatim): "the main idea behind linkFrom is to allow to specify edge on source OR on target. we have to keep this possibility."
- **D-09:** Edge shorthand: `-> peer: "tech | description"`. A single un-piped value is the **description** (user's explicit choice, not the tech-first alternative): `-> db: queries orders` = description only; `-> db: sql |` = tech only; `-> db: sql | queries` = both. Options (color, style, rank/equal, reverse, ...) via trailing `{ }` block: `-> peer: "sql | q" { color: red }`.
- **D-10:** Link targets use v1.10 relative-peer semantics: bare name resolves against the enclosing parent's children walking up ancestry; dotted path is absolute. Reuses `internal/peer.Resolve` unchanged.
- **D-11:** Duplicate edge statements for the same (unit, peer) pair in one block are a hard error — mirrors TOML's link-map single-edge-per-pair semantics.

### Root properties, templates, includes
- **D-12:** `properties { }` top-level block with the same keys as TOML `[properties]` (name, description, color, style, border, edges, expanded).
- **D-13:** Templates: function-like `template name(p1, p2) { ... }` declaration + `use name(v1, v2)` instantiation. `${param}` substitution semantics unchanged. Template bodies accept the FULL unit grammar (nesting, fields, edges with relative peers).
- **D-14:** Includes: `include path` statements (top level), optional `once` modifier: `include ./shared.c4d once`. Path relative to the including file, like TOML.
- **D-15:** List values (`expanded`, etc.): both inline `[a, b]` and one-per-line list forms accepted.

### Nested use (TOML format extension)
- **D-16:** Nested `use` lands in BOTH formats (user: "we need nested use in c4d AND in toml for sure"):
  - C4D: `use name(args)` legal inside any unit block — instantiation attaches under the enclosing parent.
  - TOML: `[[unit.use]]` array-of-tables nested under the unit section — consistent with how subunits nest in TOML today.
- **D-17:** `use` inside template bodies is ENABLED in this phase — template-instantiating-template. This LIFTS the v1.10 deferral recorded in STATE.md "Deferred Items" (template nesting); expansion and cycle detection must handle it. Update the STATE.md deferred entry at planning.

### Grammar details
- **D-18:** `;` allowed as a statement separator (one-line nested blocks): `api: container { technology: Go; db: db { } }`; empty blocks `x: system { }` fine.
- **D-19:** Field keywords (name, description, technology, reference, color, style, border, edges, expanded, once, ...) are RESERVED in the unit-body namespace — a unit id colliding with one is a hard error with a Levenshtein suggestion (reuse `internal/validator/suggest.go` machinery at parse level).

### Parser implementation
- **D-20:** PEG grammar with `github.com/mna/pigeon` (user-specified tool) — generated Go parser via `go generate`. This deliberately departs from the repo's zero-new-deps culture; pigeon is a build-time code generator (tools.go pattern), no runtime dependency.

### Converter architecture
- **D-21:** DSL parses DIRECTLY to `*parser.Model` (DSL-native line/col errors as `*parser.ParseError`). The Model is the hub: TOML→Model (existing parser), C4D→Model (new), Model→TOML emit, Model→C4D emit. Render path unchanged downstream of Model.
- **D-22:** Round-trip contract is CANONICAL-EQUIVALENT (not byte-identical): TOML→C4D→TOML and C4D→TOML→C4D must produce canonically-equal text under normalization of whitespace/quoting/key-order (the DI-1 canonicalDOT precedent applied to source formats). Explicit defaults may normalize away. Round-trip over all fixtures is the parity test strategy.
- **D-23:** Emitted units use a FIXED canonical field order (struct-defined; enumerate at planning) — deterministic diffs; source field order is not preserved (normalized away under D-22).
- **D-24:** `convert` VALIDATES the model first — an invalid model is a hard error, no output written (hard-error-everywhere stance).
- **D-25:** Convert scope: single file by default (includes untouched); `--follow-includes` converts the whole include graph — each file gets its format twin, include paths rewritten to the new extension.
- **D-26:** Include graphs may MIX `.c4d` and `.toml` files freely — include.Resolve merges at Model level; file-by-file migration works.

### CLI surface
- **D-27:** Input dispatch by extension: `.toml` → TOML parser, `.c4d` → C4D parser (user: ".toml is toml, .c4d is dsl"). Unknown extension = error.
- **D-28:** Conversion via subcommand: `c4drill convert to-toml <file.c4d>` / `c4drill convert to-c4d <file.toml>` (user named the directions "to-toml and to-c4d").
- **D-29:** `c4drill diagram.c4d` renders DIRECTLY through the full pipeline — C4D is a first-class authoring format from day one, not a converter curiosity.
- **D-30:** Converted output writes next to the input with swapped extension (`diagram.c4d` → `diagram.toml`); respects `-o` for a different directory.
- **D-31:** `c4drill fmt`: in-place rewrite (gofmt-style) + `--check` mode (reports misformatted files, exit 1 — CI gate). Formats BOTH `.c4d` and `.toml`. Accepts multiple files and directories (recursive walk over `*.c4d`/`*.toml`).

### Formatter
- **D-32:** `fmt` PRESERVES comments (and blank-line grouping) — requires a trivia-aware AST: the C4D parse tree keeps comments attached to statements; TOML fmt needs position-aware parsing (go-toml unstable API). fmt is NOT a Model-hub re-emit. Idempotency (fmt∘fmt = fmt) is an implied contract.
- **D-33:** Canonical emitter style: compact one-line leaf blocks — units WITHOUT subunits or edges emit on one line (`db: db { description: cache }`); nested units stay multi-line. Applies to convert output and fmt alike.

### Docs & naming
- **D-34:** The format is named **C4D** in all docs (matches the `.c4d` extension).
- **D-35:** Full docs in-phase: README gains a C4D syntax section + convert/fmt docs; `skill/examples/` gains `.c4d` twins of key fixtures; the portable plugin skill `c4drill-toml` is EXTENDED IN PLACE (name kept for compat) to cover both formats and the new convert/fmt commands; the plugin's render command accepts `.c4d` input.

### Claude's Discretion
- Exact enumeration of type keywords (from `internal/model`) and the fixed canonical field order (D-23).
- PEG grammar file organization, pigeon error-listener customization for quality line/col messages.
- Package layout (e.g., `internal/c4d/` lexer+parser+emitters, where TOML/C4D emitters live).
- Emitter spacing/blank-line rules beyond D-33 (indent width, blank lines between top-level units).
- Round-trip fixture corpus composition (testdata/, skill/examples/, new edge-case fixtures).
- fmt exit codes / `-l` style flags beyond `--check`; include-path quoting for paths with spaces.
- Whether `--follow-includes` also handles cycle-safe traversal details (it must — reuse include cycle detection).

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Pipeline and parser (the hub this phase plugs into)
- `internal/parser/parser.go` — `Parse`/`ParseFile`, `parser.Model` (Units + UnitOrder + Includes + Templates), `captureDefinitionOrder` (unstable API, BC-1 pattern the C4D parser must match for order preservation)
- `internal/parser/errors.go` — `*parser.ParseError` conventions (Message, Line, Context, Unwrap) the C4D parser must reuse
- `internal/model/unit.go`, `internal/model/link.go` — `Unit` tree, `Link`, unit types, ArrowDirection — the exact vocabulary C4D keywords map onto

### v1.10 composition passes (parity surface)
- `internal/template/` — `Expand`, param handling, `Unit.Clone()` (HS-1 unexported Mirror), where nested-use and template-nesting changes land
- `internal/include/` — `Resolve(entry, entryDir, entryFile)` Stage 1a, flat merge, `once` dedup, cycle detection — mixed .c4d/.toml graphs ride on this
- `internal/peer/` — `Resolve` relative-peer pass reused verbatim by D-10

### Validation and CLI
- `internal/validator/validator.go`, `rules.go`, `suggest.go` — validation rules (D-24 gate), Levenshtein suggestions (D-19 pattern)
- `cmd/c4drill/root.go` — cobra root command, flags, pipeline orchestration — where `.c4d` dispatch (D-27), `convert` (D-28), and `fmt` (D-31) subcommands attach

### Testing and docs targets
- `internal/testutil/canonical` — DI-1 order-insensitive comparison helper; the round-trip normalizer (D-22) follows this precedent for source formats
- `.planning/codebase/ARCHITECTURE.md`, `STACK.md`, `CONVENTIONS.md` — layering rules, lint gate (all-linters), error/comment conventions new code must satisfy
- `README.md` — gains C4D section (D-35)
- `skill/SKILL.md`, `skill/examples/` — plugin skill extended in place; example fixtures get .c4d twins (D-35)
- `go.mod` — where `github.com/mna/pigeon` enters (build-time generator, D-20)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `parser.Model` is the natural hub — C4D front-end produces it; both emitters consume it (D-21). No downstream stage (validator, view, graph, render) changes for basic parity.
- `internal/{include,peer,template}` passes run on `*parser.Model` post-parse — untouched by the new front-end as long as the Model is faithful (order, includes, templates, uses).
- `canonical.Canonical` (Phase 33) — precedent and machinery for the canonical-equivalent round-trip normalizer.
- `internal/validator/suggest.go` Levenshtein suggestions — reuse for reserved-keyword collisions (D-19).
- go-toml v2 `unstable` API — already used for definition-order capture; TOML fmt's position/comment awareness builds on it (D-32).

### Established Patterns
- Staged pipeline with strict one-directional dependencies; `cmd/c4drill` is the only composer — convert/fmt subcommands compose the same stages.
- Hard-error-everywhere (v1.10 stance) — applied to duplicate edges (D-11), reserved keywords (D-19), invalid convert input (D-24).
- Definition order is load-bearing (`UnitOrder`/`SubunitOrder`) — the C4D parser must capture order; emitters must emit in order.
- golangci-lint v2 with `default: all` — new code must pass the full gate; intentional globals need `//nolint:gochecknoglobals` + explanation.
- Render tests never `t.Parallel()` (WASM engine) — new integration tests that render must follow.

### Integration Points
- `cmd/c4drill/root.go` input dispatch — extension-based parser selection (D-27); new `convert` and `fmt` cobra subcommands beside the root render command.
- `internal/template.Expand` — nested-use attachment (D-16) and template-body `use` (D-17) extend this pass.
- Pipeline order `include → template.Expand → peer.Resolve → Validate` is load-bearing (guarded by `TestXC01_PipelineOrdering`) — C4D enters before include.

</code_context>

<specifics>
## Specific Ideas

User's exact words captured during discussion:
- On linkFrom: "the main idea behind linkFrom is to allow to specify edge on source OR on target. we have to keep this possibility."
- On extensions: ".toml is toml, .c4d is dsl"
- On conversion directions: "to-toml and to-c4d"
- On nested use: "we need nested use in c4d AND in toml for sure"
- On parser: "use PEG, with https://github.com/mna/pigeon"
- On formatter: "also we need formatter for c4d files, like the 'gofmt' one"
- Single-value edge label = DESCRIPTION (user overrode the tech-first recommendation).
- Style reference points named by the user at phase creation: "similar to the likec4, or d2" — the chosen brace-block form follows D2.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope. (Note: template nesting was previously deferred in STATE.md and is now promoted INTO this phase by D-17 — the Deferred Items table should drop/annotate that row at planning.)

</deferred>

---

*Phase: 35-Add a simple DSL alternative to the TOML diagram definition*
*Context gathered: 2026-08-14*
