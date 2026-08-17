# Phase 35: C4D DSL + Converters - Pattern Map

**Mapped:** 2026-08-14
**Files analyzed:** 21 (new + modified)
**Analogs found:** 17 / 21 (4 with no in-repo analog — pigeon grammar, generated parser, emitters, tools.go)

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/c4d/c4d.peg` (new) | grammar | transform | none (pigeon format, RESEARCH.md §2) | none |
| `internal/c4d/parser_gen.go` (new, generated) | generated parser | transform | none (`-nolint` flag + header per Risk 1) | none |
| `internal/c4d/ast.go` (new) | model | transform | `internal/model/unit.go` | role-match |
| `internal/c4d/parse.go` (new) | service (front-end) | transform | `internal/parser/parser.go` (`Parse`/`ParseFile`) | exact |
| `internal/c4d/tomodel.go` (new) | service (AST→Model) | transform | `internal/parser/parser.go` (`parseUnitWithOrder`, `extract*`) | exact |
| `internal/c4d/errors.go` (new) | utility (errors) | transform | `internal/parser/errors.go` (`wrapDecodeError`) | exact |
| `internal/c4d/emit_c4d.go` (new) | emitter (Model/AST→C4D text) | transform | `internal/testutil/canonical/canonical.go` (serializer), `internal/render/converter.go` (builder) | partial |
| `internal/c4d/emit_toml.go` or `internal/tomlemit/` (new) | emitter (Model→TOML text) | transform | none (target style = `skill/examples/06-templates.toml`, `testdata/valid.toml`) | none |
| `internal/c4d/tomlfmt.go` (new) | formatter (TOML fmt, position-aware) | file-I/O | `internal/parser/parser.go` (`captureDefinitionOrder` — unstable API precedent) | partial |
| `internal/testutil/canonsrc/canonsrc.go` (new) | test utility (round-trip normalizer) | transform | `internal/testutil/canonical/canonical.go` | exact (DI-1 precedent) |
| `cmd/c4drill/convert.go` (new) | controller (cobra subcommand) | file-I/O | `cmd/c4drill/root.go` (`NewRootCmd`/`runRoot`) | role-match (first subcommand — no `AddCommand` precedent) |
| `cmd/c4drill/fmt.go` (new) | controller (cobra subcommand) | file-I/O | `cmd/c4drill/root.go` | role-match (same caveat) |
| `cmd/c4drill/root.go` (modified) | controller | request-response | itself — Stage 1 dispatch at lines 110-115 | exact |
| `internal/template/expand.go` (modified) | service (recursive expansion) | transform | itself + `internal/include/resolve.go` (cycle/depth pattern) | exact |
| `internal/parser/parser.go` (modified, optional `[[unit.use]]` sugar) | service | transform | itself — `extractInstantiations` lines 280-333 | exact |
| `go.mod` + `tools.go` (new) | config | batch | none (standard Go tools.go pattern; pigeon is build-time only, D-20) | none |
| `README.adoc` (modified — NOTE: repo has `README.adoc`, NOT `README.md`) | docs | n/a | itself (C4D section appends beside existing TOML docs) | exact |
| `skill/SKILL.md` (modified) + `plugins/c4drill/**` copies | docs | n/a | itself (extended in place, name `c4drill-toml` kept, D-35) | exact |
| `skill/examples/{06,07,08,09}*.c4d` twins (new) | config (fixtures) | n/a | `skill/examples/06-templates.toml`, `08-include/`, `09-composed/` | exact |
| New edge-case fixtures (`testdata/c4d/*.c4d` etc.) | config (fixtures) | n/a | `testdata/*.toml` (valid, template_*, invalid_*) | exact |
| `internal/c4d/parse_test.go`, `emit_test.go`, round-trip tests (new) | test | transform | `internal/parser/parser_test.go`, `cmd/c4drill/root_test.go` | exact |

---

## Pattern Assignments

### `internal/c4d/parse.go` (service front-end, transform)

**Analog:** `internal/parser/parser.go` — the C4D front-end must be a drop-in sibling of `Parse`/`ParseFile`, producing the same `*parser.Model` so `include.Resolve → template.Expand → peer.Resolve → Validate` run unchanged (D-21).

**Model hub to produce** (`internal/parser/parser.go` lines 36-57 — the struct the front-end fills):
```go
type Model struct {
	Properties     model.Properties          `toml:"properties"`
	UnitOrder      []string
	Units          map[string]*model.Unit
	Templates      map[string]*TemplateDef   `toml:"-"`
	Instantiations []Instantiation           `toml:"-"`
	Includes       []IncludeDirective        `toml:"-"`
}
```
C4D mapping: `properties { }` → `Properties`; brace-block order → `UnitOrder`/`SubunitOrder` (PEG actions append in match order — free, no `captureDefinitionOrder` needed); `template n(p) { }` → `TemplateDef{Params, Unit}` with literal `${param}` tokens (NOT substituted at parse); `use n(args)` → `Instantiation{Template, Parent, Params}`; `include path [once]` → `IncludeDirective{Path, Once}`.

**Entry-point pattern** (lines 102-110 + 757-769):
```go
// Parse parses TOML data into a Model.
func Parse(data []byte) (*Model, error) {
	unitOrder, subunitOrders, templateSubunitOrders, err := captureDefinitionOrder(data)
	if err != nil {
		return nil, wrapDecodeError(err)
	}
	...
}

// ParseFile reads a TOML file and parses it into a Model.
//
//nolint:gosec // G304: Path is provided by caller, this is intentional for CLI tool
func ParseFile(path string) (*Model, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &ParseError{Message: "failed to read file", Context: path, Cause: err}
	}
	return Parse(data)
}
```
Copy the `ParseFile` shape verbatim (same `//nolint:gosec` + explanation), wrapping pigeon's `Parse(g, data)` call and asserting the `any` result to the typed C4D AST.

### `internal/c4d/tomodel.go` (AST→Model conversion, transform)

**Analog:** `internal/parser/parser.go` `parseUnitWithOrder` (lines 572-664) — recursive unit building with type inference and name humanization. The C4D converter must apply THE SAME three post-parse hooks so both front-ends produce identical Models:

**Type inference + humanize hooks** (lines 599-615):
```go
	// Apply default type if not specified
	if unit.Type == "" {
		unit.Type = defaultTypeForParent(parentType)
	}
	// Infer level-specific type for generic types (db, queue)
	unit.Type = inferGenericType(unit.Type, parentType)
	// v1.10 ERGO-03/05: derive the display name from the identifier segment
	if unit.Name == "" {
		unit.Name = model.Humanize(name)
	}
```
**PLANNING LANDMINE:** `defaultTypeForParent` (line 673) and `inferGenericType` (line 706) are UNEXPORTED in `internal/parser` — `internal/c4d` cannot call them. Planner must pick: export them, move them to `internal/model`, or duplicate. Recommend export/move (D-02 "type omittable, inferred" depends on them).

**Reserved-keyword set for D-19** — reuse the field list already enumerated in `isBuiltinField` (lines 744-752) as the C4D reserved word set (plus `once`, `use`, `include`, `template`):
```go
func isBuiltinField(key string) bool {
	return slices.Contains([]string{
		"type", "name", "description", "technology",
		"reference", "color", "style", "border", "edges",
		"width", "height", "expanded", "link", "linkFrom",
	}, key)
}
```

**Reserved-table extraction pattern** (`extractInstantiations`, lines 296-330 — the shape `use`/`include` desugaring follows; also the analog for optional `[[unit.use]]` sugar):
```go
	for _, entry := range useArr {
		inst := Instantiation{Params: make(map[string]string)}
		for k, v := range useMap {
			...
			switch k {
			case "template":
				inst.Template = s
			case "parent":
				inst.Parent = s
			default:
				inst.Params[k] = s
			}
		}
		m.Instantiations = append(m.Instantiations, inst)
	}
```
C4D `use n(args)` inside a unit block desugars to `Instantiation{Template, Parent: enclosingPath, Params}` — per RESEARCH §4 this is the EXISTING mechanism (XC-03), zero Expand changes for basic nested use.

**Type keyword vocabulary** (`internal/model/unit.go` lines 9-35): the exact TOML type strings (`person`, `personExternal`, `system`, `systemExternal`, `db`, `dbExternal`, `queue`, `queueExternal`, `box`, `container`, `containerDb`, `containerQueue`, `containerBox`, `component`, `componentDb`, `componentQueue`, `componentBox`) — D-03 requires 1:1 keyword mapping; D-04 `external` modifier maps to the `*External` variants.

### `internal/c4d/errors.go` (error wrapping, transform)

**Analog:** `internal/parser/errors.go` — the error contract `cmd/c4drill` already handles. Pigeon's `errList` wraps exactly like go-toml errors do in `wrapDecodeError` (lines 45-75):
```go
func wrapDecodeError(err error) error {
	var pe *unstable.ParserError
	if errors.As(err, &pe) {
		return &ParseError{Message: pe.Message, Line: 1, Cause: pe}
	}
	if de, ok := errors.AsType[*toml.DecodeError](err); ok {
		line := extractLineFromDecodeError(de)
		return &ParseError{Message: de.Error(), Line: line, Cause: de}
	}
	return &ParseError{Message: err.Error(), Cause: err}
}
```
The C4D wrapper converts pigeon `*parserError` (has `Inner` + `pos.line`/`pos.col` via `c.pos` in actions) into `*parser.ParseError{Message, Line, Context, Cause}` — DSL-native line/col per D-21. Reuse the `ParseError.Error()` format-priority idiom (lines 25-35: line > context > plain) and `Unwrap()`.

**Hard-error types with Kind/Site/Detail** — for semantic errors (duplicate edge D-11, reserved keyword D-19), follow `internal/template/expand.go` `ExpandError` (lines 505-521):
```go
type ExpandError struct {
	Kind   string
	Site   string
	Detail string
}
func (e *ExpandError) Error() string {
	if e.Site != "" {
		return fmt.Sprintf("template expand: %s at %s: %s", e.Kind, e.Site, e.Detail)
	}
	return fmt.Sprintf("template expand: %s: %s", e.Kind, e.Detail)
}
```

**Levenshtein suggestions (D-19)** — `internal/validator/suggest.go` lines 52-58 (exported, directly callable):
```go
func FormatSuggestion(typo string, candidates []string) string {
	if suggestion := SuggestSimilar(typo, candidates); suggestion != "" {
		return fmt.Sprintf(` (did you mean "%s"?)`, suggestion)
	}
	return ""
}
```
Pass the reserved-keyword list as candidates; parse-level errors embed `FormatSuggestion` output in the message.

### `internal/template/expand.go` (modified — template-body `use`, D-17)

**Analog:** itself. The recursion point is `expandInstantiation` (lines 99-165): lookup → param check → `tmpl.Unit.Clone()` (HS-1) → `buildReplacer` → `applySubstitution` → `attachProduced`. Genuinely new work (RESEARCH §4): after substitution, if the clone's subtree contains `use` entries (template-body instantiations), recurse — params flow outer→inner — with cycle detection. Key existing pieces to keep reusing:

**Clone + substitute** (lines 129-144):
```go
	clone := tmpl.Unit.Clone()
	...
	replacer := buildReplacer(tmpl.Params, inst.Params)
	applySubstitution(clone, replacer)
```
**Nested attachment already exists** (`attachNested`, lines 283-314): `resolveUnitByPath(m, parent)` walks dotted paths and appends to `parentUnit.Subunits`/`SubunitOrder` — nested use (D-16 C4D form) needs NO new attach code.

**Cycle-detection pattern to copy for template recursion** — `internal/include/resolve.go` lines 32 + 71-76 + 118-125:
```go
const maxIncludeDepth = 100
...
	if len(stack) > maxIncludeDepth {
		return nil, &parser.ParseError{
			Message: fmt.Sprintf("include depth exceeded %d (cycle or pathological graph)", maxIncludeDepth),
			Context: includingFile,
		}
	}
...
	if slices.Contains(stack, absPath) {
		cycle := append(append([]string{}, stack...), absPath)
		return false, &parser.ParseError{
			Message: "include cycle detected: " + strings.Join(cycle, " -> "),
			Context: includingFile,
		}
	}
```
Template A→B→A cycle mirrors this: ancestor stack of template names + `slices.Contains` + depth cap; error names the chain.

**Post-loop guards already cover recursion** (RESEARCH §4): TMPL-07 full-path collision (`pathTracker.claim`, lines 487-500) and TMPL-06 residual-`${` scan (`assertNoResidualTokens`, lines 350-362) run after all expansion — keep them as the recursive-loop's exit checks.

### `cmd/c4drill/convert.go` + `cmd/c4drill/fmt.go` (cobra subcommands, file-I/O)

**Analog:** `cmd/c4drill/root.go` — command construction + pipeline composition. NOTE: these are the FIRST subcommands in the repo (root has none today); there is no `AddCommand` precedent — wire via `cmd.AddCommand(newConvertCmd())` in `NewRootCmd` (line 50-84) following the same builder shape:

```go
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "c4drill <input.toml>",
		Short: "Generate C4 architecture diagrams from TOML",
		...
		SilenceUsage: true,
	}
	cmd.PersistentFlags().StringVarP(&format, "format", "f", formatSVG, "Output format (dot|svg|html)")
	...
}
```

**Stage composition + error prefixes** (lines 110-155) — convert composes the same stages up to Validate (D-24 validates before emitting):
```go
	// Stage 1: Parse
	m, err := parser.ParseFile(inputPath)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	if m, err = include.Resolve(m, filepath.Dir(inputPath), inputPath); err != nil {
		return fmt.Errorf("include: %w", err)
	}
	m, err = template.Expand(m)
	if err != nil {
		return fmt.Errorf("expand: %w", err)
	}
	if err := peer.Resolve(m); err != nil {
		return fmt.Errorf("resolve peers: %w", err)
	}
	valErrors := validator.Validate(m)
	if len(valErrors) > 0 {
		validator.ReportErrors(valErrors, cmd.OutOrStderr())
		return errValidationFailed
	}
```
Copy the lowercase single-word stage prefix convention (`parse:`, `include:`, `convert:`, `fmt:`) and the sentinel `var` block with `// Static errors for better error handling.` comment (lines 25-31).

**Extension-swap output naming (D-30)** — extend `deriveBasename` (lines 192-199):
```go
func deriveBasename(inputPath string) string {
	basename := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	if basename == "" {
		return "diagram"
	}
	return basename
}
```

### `cmd/c4drill/root.go` (modified — extension dispatch, D-27)

**Analog:** itself, Stage 1 (lines 110-115). Replace the direct `parser.ParseFile` with dispatch:
```go
	// Stage 1: Parse
	m, err := parser.ParseFile(inputPath)
	if err != nil {
		return fmt.Errorf("parse: %w", err)
	}
```
becomes extension-branched (`.toml` → `parser.ParseFile`, `.c4d` → `internal/c4d.ParseFile`, unknown → hard error naming accepted extensions per D-27). Everything downstream (include → expand → peer → validate → views → render) is untouched — guarded by `TestXC01_PipelineOrdering`. Also update `Use`/`Long` help text (`c4drill <input.toml>` → mention `.c4d`).

### `internal/testutil/canonsrc/canonsrc.go` (round-trip normalizer, test utility)

**Analog:** `internal/testutil/canonical/canonical.go` — the DI-1 precedent, structurally identical problem (normalize away non-semantic variation, then compare). Copy its package-level conventions verbatim:

**Test-only importable helper pattern** (package doc, lines 1-34):
```go
// Package canonical provides an order-insensitive semantic comparator for DOT
// (Graphviz) output, realizing STATE.md decision DI-1 (the "canonicalDOT"
// contract) as a reusable helper importable from any _test.go file in the repo.
//
// This package is test-only by convention but lives outside a _test.go file so
// Go's toolchain exposes it on the importable package surface...
// Canonical is a pure function (no I/O, no global state), deterministic, and
// depends only on stdlib sort + strings. It uses testing.T only for t.Helper()
// marking; callers assert on the returned canonical form.
package canonical

func Canonical(t *testing.T, dot string) string {
	t.Helper()
	stmts, ok := parseDOTStatements(dot)
	if !ok {
		t.Fatalf("canonical: failed to parse DOT output for canonical comparison")
	}
	return serializeDOTStatements(stmts)
}
```
canonsrc does the same for source formats: parse (TOML via unstable API / C4D via its AST), drop trivia (comments, whitespace), normalize quoting + key order, serialize to a sorted canonical string. Explicit defaults may normalize away (D-22).

### `internal/c4d/emit_c4d.go` + `emit_toml.go` (emitters, transform)

**No direct analog** (nothing in the repo emits TOML/Diagram-DSL text today). Two precedents to blend:
1. **Serializer shape** — `canonical.go` `serializeDOTStatements` (lines 277-308): recursive `strings.Builder` serializer over a tree, deterministic output. The C4D emitter is this pattern applied to `*parser.Model`/C4D AST, honoring D-33 compact-leaf style (`db: db { description: cache }` one-liners, nested multi-line, fixed field order D-23).
2. **Target format for TOML emit** — fixture style is the spec: `skill/examples/06-templates.toml` lines 27-37 (template + params + link tables), lines 77-89 (`[[use]]` with `template`/`parent`/params), `skill/examples/08-include/entry.toml` (`[[include]]` with `path`/`once`), `testdata/valid.toml` (`[properties]`, `[unit]` tables, `[[unit.link]]`).

### `internal/c4d/tomlfmt.go` (TOML formatter, file-I/O)

**Partial analog:** `internal/parser/parser.go` `captureDefinitionOrder` (lines 439-447) — the repo's existing go-toml `unstable` API usage; TOML fmt's position/comment awareness builds on the same API (D-32; unstable nodes carry positions):
```go
	p := unstable.Parser{}
	p.Reset(data)
	for p.NextExpression() {
		expr := p.Expression()
		if expr.Kind != unstable.Table {
			continue
		}
		parts := extractKeyParts(expr.Key())
		...
	}
	if err := p.Error(); err != nil {
		//nolint:wrapcheck // unstable.Parser error is wrapped by Parse's caller via wrapDecodeError
		return nil, nil, nil, err
	}
```

### Test files (`internal/c4d/*_test.go`, round-trip tests, CLI tests)

**Analog:** `internal/parser/parser_test.go` (black-box style, lines 1-27):
```go
package parser_test

import (
	"os"
	"testing"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseValidProperties(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../testdata/valid.toml")
	require.NoError(t, err, "failed to read test fixture")
	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")
	assert.Equal(t, "Test System", got.Properties.Name, "Properties.Name")
}
```
Conventions: black-box `_test` package, testify `require` for preconditions + `assert` with per-field msg strings, fixtures read from `testdata/`. `t.Parallel()` is fine for parse/convert/fmt tests — but render tests NEVER parallel (`//nolint:paralleltest // go-graphviz WASM engine has concurrency issues`, 81 occurrences — any test that renders `.c4d` through the WASM engine must follow, per `cmd/c4drill/root_test.go` lines 23-25). Round-trip tests walk `testdata/` + `cmd/c4drill/testdata/` + `skill/examples/` via `filepath.WalkDir` (RESEARCH §8.2; 09-composed is the 4-file include graph).

### Fixture + docs files

**Analog:** `.c4d` twins copy the naming + comment-header style of `skill/examples/06-templates.toml` (explanatory `#` header block naming the decisions demonstrated: "Demonstrates ... (TMPL-01 through TMPL-10)") and `08-include/` multi-file layout (entry + included files). New edge-case fixtures follow `testdata/` naming: `valid`, `template_*`, `invalid_*` prefixes.

**DISCREPANCY FLAG:** CONTEXT D-35 says "README.md gains a C4D section" but the repo has `README.adoc` (AsciiDoc) — the C4D section lands in `README.adoc`. `skill/SKILL.md` (front-matter `name: c4drill-toml` kept for compat per D-35) and its packaged copies under `plugins/c4drill/` (`commands/c4drill-render.md`, `opencode/skills/c4drill-toml/`) are extended in place; the render command doc must accept `.c4d` input.

---

## Shared Patterns

### Error contract (all C4D code)
**Source:** `internal/parser/errors.go` lines 13-40 + `internal/template/expand.go` lines 505-521
**Apply to:** every new `internal/c4d` file, `convert.go`, `fmt.go`
All failures are hard errors (v1.10 stance): syntax → `*parser.ParseError{Message, Line, Context, Cause}` (DSL-native line/col from pigeon `c.pos`); semantic (duplicate edge D-11, reserved keyword D-19, unknown type) → structured error type with Kind/Site/Detail following `ExpandError`. Never silently accept.

### Stage-prefix error wrapping
**Source:** `cmd/c4drill/root.go` lines 112-155
**Apply to:** `convert.go`, `fmt.go`, root dispatch
`fmt.Errorf("parse: %w", err)` — lowercase single-word stage prefix; sentinel errors in a `var (...)` block commented `// Static errors for better error handling.`

### Definition-order preservation
**Source:** `internal/parser/parser.go` (UnitOrder/SubunitOrder appends), `internal/template/expand.go` `attachTopLevel`/`attachNested` (lines 274-275, 310-311)
**Apply to:** C4D parser (brace order is natural order — PEG actions append), both emitters (emit in `UnitOrder`/`SubunitOrder`)

### Cycle detection + depth cap
**Source:** `internal/include/resolve.go` lines 32, 71-76, 118-125
**Apply to:** recursive template expansion (D-17), `--follow-includes` graph walk (D-25) — reuse include's canonical-path visited-set and stack verbatim

### Lint gate compliance
**Source:** `.planning/codebase/CONVENTIONS.md` (golangci-lint v2, `default: all`)
**Apply to:** ALL new files
- Every `//nolint:` paired with an inline explanation (`//nolint:gochecknoglobals // ...`, `//nolint:gosec // G304: ...`, `//nolint:exhaustive // Default case handles all remaining types`)
- Generated `parser_gen.go` needs a blanket `//nolint` header (pigeon `-nolint` flag generates these)
- Blank line after every `if err != nil { ... }` block; import groups: stdlib → module → third-party; doc comment on every exported declaration
- Package-level state always `//nolint:gochecknoglobals` with reason

### Hard-error suggestion UX
**Source:** `internal/validator/suggest.go` `FormatSuggestion` (lines 52-58)
**Apply to:** C4D reserved-keyword collisions (D-19), unknown type keywords — parse-level ` (did you mean "..."? )` suffix

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `internal/c4d/c4d.peg` | grammar | transform | No PEG grammars in repo; follow pigeon syntax per RESEARCH.md §2 (rules, `{ action }` blocks on `*current` receiver, `#{ state }`, initializer helpers) |
| `internal/c4d/parser_gen.go` | generated parser | transform | No generated code in repo; commit it + `//go:generate pigeon -o parser_gen.go -nolint c4d.peg` + tools.go |
| `emit_toml.go` / `emit_c4d.go` emitters | emitter | transform | No text emitters exist; serializer shape from `canonical.go`, output style from D-33 + fixture formats |
| `tools.go` | config | batch | No build-time tool deps today; standard `//go:build tools` pattern, pigeon enters `go.mod` as build-time only (D-20) |

## Metadata

**Analog search scope:** `internal/` (all 12 packages), `cmd/c4drill/`, `skill/`, `plugins/c4drill/`, `testdata/`, `skill/examples/`, `go.mod`, `.planning/codebase/`
**Files scanned:** 30+ (parser, model, template, include, peer, validator, canonical, root.go + tests, fixtures, plugin docs)
**Key facts for planner:**
1. `defaultTypeForParent`/`inferGenericType` are unexported in `internal/parser` — must be exported/moved for C4D reuse (D-02 inference parity).
2. Nested use via `Instantiation.Parent` + `attachNested` ALREADY WORKS (Phase 31 XC-03) — C4D use-in-block desugars to it; `[[unit.use]]` TOML sugar is optional/redundant (RESEARCH §4).
3. `cmd/c4drill` has NO existing subcommands — `convert`/`fmt` are the first `AddCommand` users.
4. README is `README.adoc` (AsciiDoc), not README.md as CONTEXT implies.
5. pigeon is NOT in `go.mod` yet; no `go:generate`/tools.go exists anywhere in the repo.
**Pattern extraction date:** 2026-08-14
