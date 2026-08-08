# Phase 31: Template expansion - Pattern Map

**Mapped:** 2026-08-08
**Source:** CONTEXT.md (D-01..D-08) + RESEARCH.md (Architecture Patterns, Code Examples)

Maps every file to be created/modified to its closest existing analog in the codebase, with concrete line references the executor copies from.

## File Classification

| File | Action | Role | Data Flow | Analog |
|------|--------|------|-----------|--------|
| `internal/parser/parser.go` | MODIFY | parser core | transform (TOML→Model) | itself — `properties` extraction + skip rule |
| `internal/model/unit.go` | MODIFY | model | transform (deep-copy) | (no analog — new method; sibling `Link` value-copy semantics) |
| `internal/model/unit_test.go` | NEW | test | unit | `internal/parser/parser_test.go` |
| `internal/template/expand.go` | NEW | service/transform | transform (Model→Model) | `internal/parser/parser.go` Parse + `internal/validator/validator.go` Validate |
| `internal/template/expand_test.go` | NEW | test | unit + integration | `internal/parser/parser_test.go` |
| `internal/parser/parser_test.go` | MODIFY (extend) | test | unit | itself |
| `cmd/c4drill/root.go` | MODIFY | pipeline orchestration | request-response (CLI) | itself — Parse/Validate staging |
| `testdata/template_*.toml` | NEW | fixture | file-I/O | `testdata/valid.toml`, `testdata/links.toml` |

## Pattern Assignments

### File: `internal/parser/parser.go` (MODIFY)

**Change 1 — BC-1 skip rule (D-08).** Analog: the existing `properties` skip at `parser.go:127-130`.

Copy-from excerpt (current):
```go
// internal/parser/parser.go:127-130
// Skip [properties] section
if len(parts) == 1 && parts[0] == "properties" {
    continue
}
```
Target: add two sibling conditions after this block — one for `use`/`include` (len==1), one for the `template` namespace (`parts[0]=="template"`). See RESEARCH.md Pattern 1 for exact target shape. The executor reads `parser.go:107-150` (the full `captureDefinitionOrder` walk) to see where the skip sits.

**Change 2 — BC-1 rawMap extraction (D-08).** Analog: the existing `properties` extraction at `parser.go:68-77`.

Copy-from excerpt (current):
```go
// internal/parser/parser.go:68-77
if props, ok := rawMap["properties"]; ok {
    propsData, err := toml.Marshal(props)
    if err != nil {
        return nil, &ParseError{Message: "failed to marshal properties", Cause: err}
    }
    if err := toml.Unmarshal(propsData, &m.Properties); err != nil {
        return nil, wrapDecodeError(err)
    }
}
```
Target: mirror this for `template` (→ `&m.Templates`) and `use` (→ `&m.Instantiations`). Do NOT extract `include` (reserved, Phase 32). The executor reads `parser.go:47-96` (full `Parse`) to place the new blocks between the properties extraction and the unit loop at `parser.go:80`.

**Change 3 — Model struct fields.** Analog: the existing `Model.Properties` field and `toml:"-"` fields like `UnitOrder` at `parser.go:38-39`.

Copy-from excerpt (current):
```go
// internal/parser/parser.go:35-42
type Model struct {
    Properties model.Properties `toml:"properties"`
    UnitOrder  []string         // tracks definition order (not from TOML)
    Units      map[string]*model.Unit
}
```
Target: add `Templates map[string]*TemplateDef` and `Instantiations []Instantiation`, both `toml:"-"`, plus the `TemplateDef`/`Instantiation` type definitions. `TemplateDef` should hold a parsed `*model.Unit` subtree (parsed via `parseUnitWithOrder` semantics so `[[template.svc.link]]` becomes `model.Unit.Links`). See RESEARCH.md Open Questions.

### File: `internal/model/unit.go` (MODIFY) + `internal/model/unit_test.go` (NEW)

**Analog for Clone:** none direct — this is a new method. But the correctness mechanism is the **value-copy of `Link`** (all-value-type struct). The executor reads `internal/model/link.go:43-68` to confirm `Link` has no pointer/map/slice fields, so `out[i] = links[i]` is a full copy preserving unexported `Mirror`.

Clone target shape is in RESEARCH.md Pattern 3 (~15-20 LOC). Uses `slices.Clone` (already used at `parser.go:310`) for slice headers, then element-wise value copy for `Link`, then map-iteration recursion for `Subunits`.

**Analog for unit_test.go:** `internal/parser/parser_test.go` — same package-external test idiom (`package parser_test`), `t.Parallel()`, testify `require`/`assert`. The new file uses `package model_test`.

Copy-from test idiom (current):
```go
// internal/parser/parser_test.go:1-13
package parser_test

import (
    "os"
    "testing"

    "github.com/Djarvur/c4drill/internal/model"
    "github.com/Djarvur/c4drill/internal/parser"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)
```

### File: `internal/template/expand.go` (NEW)

**Analogs (two):**
1. `internal/parser/parser.go` `Parse` (parser.go:47-96) — for the transform-function shape: takes input, builds/returns `*parser.Model`, returns error. `Expand(m *parser.Model) (*parser.Model, error)`.
2. `internal/validator/validator.go` `Validate` (validator.go:16) — for iterating a `*parser.Model` and producing structured errors. `Validate(m *parser.Model) ValidationErrors`.

Copy-from excerpt (validator signature):
```go
// internal/validator/validator.go:16
func Validate(m *parser.Model) ValidationErrors {
```

**Error reporting analog:** the phase needs hard errors for missing-param (TMPL-06) and duplicate-path (TMPL-07). Two viable idioms, both present in the codebase:
- `*parser.ParseError` (`internal/parser/errors.go:10-37`) — `Message`/`Line`/`Context`/`Cause` struct, `Error()` method. Best for naming the template+param+site.
- `validator.ValidationError` / `ValidationErrors` (internal/validator/errors.go) — if Expand returns validation-style errors that the CLI reports uniformly.

Recommendation: define a small `template.ExpandError` (or reuse `parser.ParseError` with `Context` naming the `[[use]]` site) so the CLI can report it via the existing `cmd/c4drill/root.go` error path. The executor reads `internal/parser/errors.go` (full file, ~80 LOC) and `internal/validator/errors.go` to choose.

**Substitution mechanism:** `strings.NewReplacer` (stdlib). No analog needed — see RESEARCH.md Pattern 4. The executor reads `internal/model/unit.go:41-72` and `internal/model/link.go:43-68` to enumerate every string field to substitute.

### File: `internal/template/expand_test.go` (NEW)

**Analog:** `internal/parser/parser_test.go` — fixture-driven tests reading `../../testdata/*.toml`, plus in-line `parser.Parse([]byte(...))` for small cases.

Copy-from excerpt (current test shape):
```go
// internal/parser/parser_test.go:172-194 (TestParseLinksOutgoing)
func TestParseLinksOutgoing(t *testing.T) {
    t.Parallel()
    data, err := os.ReadFile("../../testdata/links.toml")
    require.NoError(t, err, "failed to read test fixture")
    got, err := parser.Parse(data)
    require.NoError(t, err, "Parse() should not error")
    webapp, ok := got.Units[webappKey]
    require.True(t, ok, "missing 'webapp' unit")
    require.Len(t, webapp.Links, 1, "webapp should have 1 link")
    // ... assert.Equal on fields
}
```

**The HS-1 three-instantiation test (TMPL-08, load-bearing)** has no direct analog — it must: (1) build a model with one template + three `[[use]]` with distinct params, (2) call `Expand`, (3) call `validator.Validate`, (4) assert `LinksFrom` slices of the three instantiations are disjoint (no shared `Link` elements / pointers), (5) call `Expand` again on a fresh parse and assert idempotency. The executor reads `internal/validator/index.go:55-86` to understand exactly what gets appended to `LinksFrom`.

### File: `internal/parser/parser_test.go` (MODIFY — extend)

**Analog:** itself. Add cases mirroring `TestParseValidProperties` (parser_test.go:15-28) and `TestParseNestedContainers` (parser_test.go:110-135) for: reserved-table skip (TMPL-09/01), `[[use]]` extraction order (TMPL-02), forward-ref (TMPL-09).

### File: `cmd/c4drill/root.go` (MODIFY)

**Analog:** the existing Parse → Validate staging at `root.go:112-119`.

Copy-from excerpt (current):
```go
// cmd/c4drill/root.go:112-119
// Stage 1: Parse
m, err := parser.ParseFile(inputPath)
if err != nil {
    return fmt.Errorf("parse: %w", err)
}
// Stage 2: Validate
valErrors := validator.Validate(m)
```
Target: insert between these two:
```go
// Stage 1.5: Expand templates (Phase 31; runs before future relative-peer/humanize passes)
m, err = template.Expand(m)
if err != nil {
    return fmt.Errorf("expand: %w", err)
}
```
The executor reads `cmd/c4drill/root.go:100-135` to see the full staging and the error-return idiom (`fmt.Errorf("stage: %w", err)`).

### File: `testdata/template_*.toml` (NEW fixtures)

**Analog:** `testdata/valid.toml` (hand-authored reference) and `testdata/links.toml` (link fixtures). New fixtures: `template_basic.toml`, `template_subtree.toml`, `template_3x_instantiate.toml`, `template_missing_param.toml`, `template_duplicate_path.toml`, `template_forward_ref.toml`. Each mirrors the `[properties]`/`[unit]`/`[[link]]` TOML shape of existing fixtures plus the new `[template.*]`/`[[use]]` tables.

## Shared Patterns

### Error handling / reporting
- Parser errors: `*parser.ParseError` (`internal/parser/errors.go:10-37`) with `Message`/`Context`/`Cause`. Expand's missing-param and duplicate-path errors should follow this shape so the CLI's existing `fmt.Errorf("...: %w", err)` path handles them uniformly.
- Validator errors: `validator.ValidationErrors` + `ReportErrors` (`internal/validator/validator.go:16,47`). Expand does NOT use these (it runs before validate), but its output must be validatable.

### Non-strict TOML (load-bearing)
go-toml/v2 stays non-strict (STATE.md DI-1 / BC-3). The `toml:",inline"` subunit trick (`unit.go:71`) depends on unknown keys being accepted. Template extraction must NOT enable `DisallowUnknownFields()`. The executor reads `internal/model/unit.go:71` for the inline tag and keeps strict mode off.

### testify test idiom (all test files)
`t.Parallel()` + `require.NoError` for preconditions + `assert.Equal`/`require.Len`/`require.True` for assertions. Fixtures via `os.ReadFile("../../testdata/...")`. All new/extended test files follow this.

### Order preservation (load-bearing)
`captureDefinitionOrder` exists to preserve authoring order (`parser.go:100-157`). Expanded template instances must append to `Model.UnitOrder` in `[[use]]` document order (Instantiations is a slice, preserving array-of-tables order). The executor reads `parser.go:80-93` (unit loop keyed on UnitOrder) and ensures Expand appends, not sorts.
