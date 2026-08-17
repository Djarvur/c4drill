# Phase 35 Research — C4D DSL alternative + bidirectional converters

**Date:** 2026-08-14
**Scope anchor:** `35-CONTEXT.md` (D-01..D-35). This research answers "what do I need to know to PLAN this phase well?"

> Note: research was performed in the orchestrator context (the `gsd-phase-researcher` agent could not spawn in this session — its `exa` MCP dependency is not connected). Facts below are verified against the live codebase and the pigeon godoc/pkg.go.dev.

---

## 1. Pipeline integration (verified against code)

The pipeline today (`cmd/c4drill/root.go`, guarded by `TestXC01_PipelineOrdering`):

```
Parse (TOML → *parser.Model) → include.Resolve → template.Expand → peer.Resolve → Validate → views → graph → render
```

**`parser.Model` (internal/parser/parser.go:36)** — the hub the C4D front-end must produce:

| Field | Type | C4D front-end duty |
|---|---|---|
| `Properties` | `model.Properties` | `properties { }` block |
| `UnitOrder` | `[]string` | capture statement order (brace DSL order is natural) |
| `Units` | `map[string]*model.Unit` | nested `{ }` blocks → recursive tree |
| `Templates` | `map[string]*TemplateDef` | `template name(p) { }` — Params + `*model.Unit` subtree with literal `${param}` tokens |
| `Instantiations` | `[]Instantiation` | `use name(args)` entries in document order |
| `Includes` | `[]IncludeDirective` | `include path [once]` statements (Path + Once) |

- `TemplateDef{Params []string; Unit *model.Unit}` — params NOT substituted at parse time (that's `template.Expand`'s job).
- `Instantiation{Template string; Parent string; Params map[string]string}` — see §4 (nested-use reality check).
- `model.Unit` fields (internal/model/unit.go:43): Type, Name, Description, Technology, Reference, Color, Style, Border, Edges, Expanded, Links, LinksFrom, Subunits, SubunitOrder.
- `model.Link` (internal/model/link.go:43): Peer, Arrow, Rank, Color, Style, Technology, Description (+ unexported Mirror).
- Definition order: `captureDefinitionOrder` (parser.go:100) uses the go-toml `unstable` API. The C4D parser captures order for free — PEG rule actions append in match order.
- Error contract: `*parser.ParseError{Message, Line, Context, Cause}` (internal/parser/errors.go:13) — C4D parse errors must wrap into this type so `cmd/c4drill` error handling is unchanged.

## 2. mna/pigeon integration (D-20)

Verified via pkg.go.dev/github.com/mna/pigeon:

- **API:** generated parser exports `Parse(string, []byte, ...Option) (any, error)`, `ParseFile`, `ParseReader`. Options include `Recover(bool)`, `Entrypoint(string)`, `GlobalStore(string, any)`, `Memoize(bool)`, `MaxExpressions(uint64)`.
- **Error accumulation:** errors are always `errList` (`[]error` of `*parserError` with `Inner`); by default the parser continues after an error and accumulates all — good for multi-error UX parity with the validator's collect-all stance.
- **Positions in actions:** code blocks run on a `*current` receiver with `c.pos` (line 1-based, col 1-based, offset), `c.text` ([]byte matched), `c.state` (backtrack-rolled-back) and `c.globalStore` (not rolled back). Wrap failures as `*parser.ParseError` with `c.pos.line` — matches D-"DSL-native line/col errors".
- **Grammar:** PEG `Rule <- expr`, labeled matches (`id:Ident`), `{ action }` per expression, top initializer block for helpers, `#{ state }` blocks, failure labels `%{Label}` + `//' {Labels}` recovery for quality messages.
- **Generate workflow:** `go install github.com/mna/pigeon@latest` → `//go:generate pigeon -o parser_gen.go c4d.peg` (nolint flag `-nolint` exists — useful against the repo's all-linters gate). Decide at planning: commit generated `parser_gen.go` (no build-time dep for consumers) + tools.go pattern for regeneration. BSD-3, Go-security-policy versions, `any`-based codegen. v1.0.0 tagged; v2 WIP — pin @latest v1.
- **Comments as trivia (D-32):** pigeon has no built-in trivia — the GRAMMAR captures it. Make `Comment` a token rule matched at statement level, attach to the statement's AST node in the action; the C4D AST (not `parser.Model`) is what fmt consumes. Design the front-end as two layers: `c4d.peg` → typed C4D AST (with comments + positions) → `toModel()` producing `*parser.Model`.

## 3. Emitters + TOML fmt

- **Model→TOML emit (convert to-toml):** plain text/template or builder emitting `[properties]`, `[unit]` tables in UnitOrder, fixed field order (D-23 planner enumerates; suggest: type, name, description, technology, reference, color, style, border, edges, expanded), inline link tables `link = { peer = { ... } }`, `[[use]]` array-of-tables with `template`/`parent`/`[use.params]`, `[[include]]` with path/once, `[template.<name>]` with `params = [...]` + unit body. Match existing fixture style (see `skill/examples/06-templates.toml`, `08-include/`).
- **Model→C4D emit (convert to-c4d, fmt):** compact-leaf style (D-33): leaf units one-line, nested multi-line, 2-space indent, blank line between top-level units, `id: type "Name" {` headers, `-> peer: "tech | desc" { opts }` edges, `template n(p) { }`, `use n(args)`, `include path [once]`. Deterministic — fmt is this emitter applied to the trivia-aware AST (comments re-attached), NOT a Model round-trip (comments would be lost — D-32).
- **TOML fmt comment preservation:** go-toml v2 `unstable` API already used in parser.go (`captureDefinitionOrder`) — unstable nodes carry positions; scanning the unstable tree preserves comments/order. Verify exact comment-node availability at execution time (fallback: line-based TOML rewriter that only normalizes whitespace/indent while touching nothing else).
- **Round-trip normalization (D-22):** build `internal/testutil/canonsrc` (or similar) — for TOML: parse via unstable API to ordered event stream, drop trivia, normalize quoting; for C4D: parse to C4D AST minus comments. Precedent: `internal/testutil/canonical` (DI-1). Corpus: `testdata/`, `cmd/c4drill/testdata/`, `skill/examples/` (09-composed is the 4-file include graph), plus new edge-case fixtures (external types, linkFrom, rank=equal, template+use+nested-use, long/unicode strings).

## 4. Nested use — REALITY CHECK (affects D-16/D-17 planning)

- **TOML nested use ALREADY EXISTS** since Phase 31: `Instantiation.Parent` (`parent = "mainapp"` in `[[use]]`, XC-03) + `attachProduced(m, inst.Parent, ...)` (internal/template/expand.go:160) attaches the produced unit under any dotted parent path. Empty Parent = top level.
  - ⇒ C4D `use svc(args)` inside a block desugars to `Instantiation{Template, Parent: enclosingPath, Params}` — **zero Expand changes needed for basic nested use**.
  - ⇒ D-16's `[[unit.use]]` TOML syntax is REDUNDANT with the existing `parent` field. Planner recommendation: treat C4D use-in-block as the authoring form; for TOML keep/document the existing `[[use]] parent` mechanism (D-16's intent — "specify nested instantiation in TOML" — is already satisfied). Adding `[[unit.use]]` as sugar is optional; if added, it must normalize to the same `Instantiation` form. Decide in plan; do not blindly add parallel syntax.
- **Genuinely NEW Expand work (D-17):** `use` inside template bodies (template-instantiating-template). Expansion must recurse: template body instantiations expand when the OUTER template is instantiated (params flow outer→inner), with cycle detection (template A using B using A = hard error). This is the v1.10 deferred item now in scope. HS-1 constraint applies at every level (deep-copy via `Unit.Clone()` — reflection/gob drops unexported `Link.Mirror`).
- Full-path collision detection (TMPL-07) and residual-`${` scan (TMPL-06) already run post-loop — they naturally cover nested+recursive expansion once recursion lands.

## 5. CLI wiring (D-27..D-31)

- Cobra root at `cmd/c4drill/root.go` (`NewRootCmd`, flags: `-f/-o/--expanded/--label-ratio`). Add:
  - `convert` subcommand: `convert to-toml|to-c4d <file> [--follow-includes] [-o dir]`. Parse (by extension) → validate (D-24) → emit. `--follow-includes`: walk the include graph (reuse include.Resolve's traversal + cycle detection), convert each file, rewrite `include` statements to the new extension.
  - `fmt` subcommand: file/dir args, recursive `*.c4d` + `*.toml`, in-place rewrite, `--check` exits 1 listing misformatted files.
- Render dispatch: in `runRoot`, branch on `strings.HasSuffix(input, ".c4d")` → C4D front-end → same `*parser.Model` downstream. Unknown extension → hard error naming accepted extensions.
- Output convention: converted twin = same dir, swapped extension (D-30); `-o` overrides dir.

## 6. Docs & examples (D-34/D-35)

- README.md: new "C4D format" section (syntax summary + convert/fmt usage).
- `skill/SKILL.md` + `skill/examples/`: `.c4d` twins of key fixtures (06-templates, 07-relative-peer, 08-include, 09-composed). Plugin skill `c4drill-toml` extended in place; render command gains `.c4d` acceptance.
- Naming: "C4D" everywhere.

## 7. Risks / landmines

1. **golangci-lint all-linters vs generated code** — use pigeon's `-nolint` flag and/or `//nolint` headers on `parser_gen.go`; generated code must pass the repo lint gate.
2. **Bareword vs reserved-key grammar ambiguity** — the PEG must encode the exact disambiguation order (D-19: field keywords reserved; suggestion errors on collision). Table-driven lexer tests.
3. **Edge-label shorthand `desc` vs `tech |` (D-09)** — single value = description; `tech |` trailing pipe = tech-only; easy to get backwards — pin with unit tests before emitters.
4. **TOML fmt scope creep** — canonical TOML emit reorders fields (D-23); fmt on .toml will rewrite user files aggressively (gofmt analogy holds — document it).
5. **go-toml unstable API** is explicitly unstable — pin behavior with round-trip tests; the parser already depends on it (precedent).
6. **Recursive template expansion + HS-1** — every level deep-copies; test 3-level nesting with links (LinksFrom disjointness gate, like TestExpandThreeInstantiationsHS1).
7. **Existing tests must stay green** — no TOML-path behavior changes except [[unit.use]] sugar (if added) and template-body use (additive). 12/12 packages green is the regression bar.

## 8. Validation Architecture

Verification dimensions for this phase (feeds VALIDATION.md / Dimension 8):

1. **Parser correctness:** table-driven C4D parse tests (every D-01..D-19 construct, every error case: reserved-keyword collision, duplicate edge, unknown type, bad arrow).
2. **Parity / round-trip:** property-style tests over the full fixture corpus — `TOML→C4D→TOML` and `C4D→TOML→C4D` canonically-equal (D-22 normalizer); fixture corpus = testdata/ + cmd/c4drill/testdata/ + skill/examples/ + new edge-case fixtures.
3. **Render equivalence:** for each fixture, canonicalDOT of `.toml` render == canonicalDOT of `.c4d` render (proves Model-level equivalence end-to-end, DI-1 machinery).
4. **Convert CLI behavior:** validate-first refusal on invalid models; `--follow-includes` graph conversion with rewritten include paths; output placement.
5. **fmt behavior:** idempotency (fmt∘fmt == fmt), comment preservation (comment-count/type assertions before/after), `--check` exit codes.
6. **Nested/recursive templates:** expansion depth tests incl. cycle hard-error, HS-1 disjointness under 3-level nesting.
7. **Regression:** `go test ./...` green (12/12); existing TOML goldens untouched.
