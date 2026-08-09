# Testing Patterns

**Analysis Date:** 2026-08-05

## Test Framework

**Runner:**
- Go standard `testing` package (`go 1.26.1` per `go.mod`, toolchain 1.26.5 detected)
- No third-party test runner. No `make test` — tasks are defined in `.mise.toml` under `[tasks]`.

**Assertion Library:**
- `github.com/stretchr/testify v1.11.1` (`go.mod`) — both `assert` and `require` imported in nearly every test file. Usage is heavily skewed toward `assert` (`assert.Equal` 340, `assert.Contains` 165, `assert.Empty` 51) with `require` reserved for fatal/precondition checks (`require.NoError` 187, `require.Error` 13, `require.ErrorAs` 2, `require.ErrorContains` 2).

**Run Commands:**
```bash
go test ./...                        # Run all tests
go test -v -race -cover ./...        # Standard task (mise task "test", .mise.toml)
golangci-lint run ./...              # Lint (lints tests too: tests: true in .golangci.yml)
golangci-lint run --fix ./...        # Lint with autofix (mise task "lint-fix")
```

## Test File Organization

**Location:**
- Co-located with source in the same package directory. Every `internal/<pkg>/` directory and `cmd/c4drill/` contains its own `*_test.go` files.
- **Black-box tests are the default**: `package parser_test`, `package validator_test`, `package graph_test`, `package render_test`, `package output_test`, `package view_test`, `package main` (in `cmd/c4drill/root_test.go` the package is `main` because the root command is in `package main`)
- **White-box tests** (internal package, for unexported functions) use the `_internal_test.go` suffix with `package render`:
  - `internal/render/expanded_internal_test.go`
  - `internal/render/html_labels_internal_test.go`
  - `internal/render/multirender_internal_test.go`
- **Integration tests** live in `integration_test.go` files: `internal/graph/integration_test.go`, `internal/output/integration_test.go`, `internal/render/integration_test.go`, `internal/view/integration_test.go`

**Naming:**
- `Test<PascalCase>` with underscore separating scenario groups: `TestParseGenericTypeInference_DbAtC1`, `TestValidateReferences_WithSuggestion`, `TestFullPipeline_ValidationError`
- Table-driven subtest names are descriptive lowercase: `"svg format is valid"`, `"returns error for nil graph input"`

**Structure:**
```
internal/<pkg>/
├── <file>.go              # source
├── <file>_test.go         # black-box unit tests (package <pkg>_test)
├── <file>_internal_test.go # white-box tests (package <pkg>)
└── integration_test.go    # cross-package pipeline tests
```

## Test Structure

**Suite Organization:**
```go
// internal/parser/parser_test.go
package parser_test

func TestParseValidProperties(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/valid.toml")
	require.NoError(t, err, "failed to read test fixture")

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	assert.Equal(t, "Test System", got.Properties.Name, "Properties.Name")
	assert.Equal(t, 40, got.Properties.LineLength, "Properties.LineLength")
}
```

**Patterns:**
- Every standalone test calls `t.Parallel()` first thing, except render-touching tests (see below)
- Fatal preconditions via `require.*` (fixture reads, nil-ness, error presence) then value assertions via `assert.*`
- `assert` calls carry a third "message" argument naming the field, e.g. `assert.Equal(t, "b", link.Peer, "link.Peer")` — this becomes the failure label
- `require.True(t, ok, "missing 'webapp' unit")` for map lookups; `require.Len(t, got.Units, 2, "should have 2 units")` for counts
- Table-driven tests use `t.Run(tt.name, ...)` subtests: `TestFlagValidation` (`cmd/c4drill/root_test.go`), `TestRenderDOT`/`TestRenderSVG`/`TestRenderFormatDispatch` (`internal/render/render_test.go`), `TestValidate*` in `internal/validator/rules_test.go`, `TestNavigation*` in `internal/graph/navigation_test.go`
- Tests asserting struct shape use numbered `// Test N:` comments mapping to spec items (`internal/graph/graph_test.go`, `internal/graph/builder_test.go`)
- CLI tests build the command, buffer stdout/stderr with `bytes.Buffer` + `cmd.SetOut(buf)`/`cmd.SetErr(buf)`, set args, and assert on output and produced files (`cmd/c4drill/root_test.go`)

**Parallelism rule (important):**
- The go-graphviz library uses a WASM render engine that is not concurrency-safe. Any test that calls `render.*` (directly or via the full pipeline) must NOT call `t.Parallel()` and carries `//nolint:paralleltest // go-graphviz WASM engine has concurrency issues`. There are 91 such annotations. Affected files: `internal/render/*` (render_test.go, integration_test.go, labels_test.go, wrap_test.go, navigation_test.go, converter_test.go, all `*_internal_test.go`), `internal/output/integration_test.go`, `internal/view/integration_test.go`, `internal/graph/builder_test.go` (imports render for some tests), and all of `cmd/c4drill/root_test.go`.
- Tests that only build `parser.Model` → `view.View` → `graph.Graph` (no rendering) are safe to parallelize: `internal/parser/parser_test.go`, `internal/validator/*_test.go`, `internal/graph/graph_test.go`, `internal/graph/navigation_test.go`, `internal/graph/path_test.go`, `internal/graph/shapes_test.go`, `internal/view/scope_test.go`, `internal/view/view_test.go`, `internal/output/writer_test.go`.

## Mocking

**Framework:** None. No mock framework (no gomock, no testify/mock) and no interface fakes are defined. Test doubles are avoided entirely by constructing real domain structs inline.

**Patterns:**
```go
// internal/validator/validator_test.go — real structs, no mocks
m := &parser.Model{
	Units: map[string]*model.Unit{
		"api": {
			Type:  model.TypeSystem,
			Links: []model.Link{{Peer: "db"}},
		},
		"db": {
			Type:      model.TypeDb,
			LinksFrom: []model.Link{{Peer: "api"}},
		},
	},
}
errors := validator.Validate(m)
assert.Nil(t, errors)
```

```go
// internal/render/render_test.go — real graph struct
func testGraph(nodeID string) *graph.Graph {
	return &graph.Graph{
		Title:     "Test Diagram",
		Direction: "TB",
		Nodes: []*graph.Node{
			{ID: nodeID, Label: &graph.Label{Name: "Test Node"}, Shape: graph.ShapeHTML},
		},
	}
}
```

**What to Mock:**
- Nothing is mocked. Side effects are contained instead:
  - **Filesystem writes** → `t.TempDir()` (`os.WriteFile`, then `assert.FileExists`)
  - **Stdout/stderr** → `bytes.Buffer` passed to `cmd.SetOut`/`cmd.SetErr` or as `io.Writer` argument
  - **go-graphviz** → invoked for real (guarded by the WASM mutex in `internal/render/render.go`); concurrency is avoided by not using `t.Parallel()`

**What NOT to Mock:**
- Do not introduce mocks for `graph`, `parser.Model`, or `validator` — the codebase convention is integration-by-construction (build real models, assert on real output). See `internal/graph/integration_test.go` and `internal/render/integration_test.go` for the canonical pipeline tests.

## Fixtures and Factories

**Test Data:**
- **Shared TOML fixtures** at the repo root: `testdata/valid.toml`, `testdata/nested.toml`, `testdata/links.toml`, `testdata/invalid_links.toml`, `testdata/invalid_references.toml`, `testdata/invalid_subunits.toml`, plus `testdata/links.dot` (expected DOT output). Read with `os.ReadFile("../../testdata/valid.toml")` from `internal/<pkg>` tests.
- **CLI fixtures**: `cmd/c4drill/testdata/valid.toml`, `invalid.toml`, `expanded.toml`, plus expected rendered outputs `expanded.dot`, `expanded/mainsystem.dot`, `expanded/mainsystem/webapp.dot`
- **Inline TOML strings**: many parser/view tests embed TOML directly as `data := []byte(...)` when a fixture file is overkill (e.g., type-inference and link tests in `internal/parser/parser_test.go`)
- **Programmatic model factories**: `buildTestModel()`, `buildTestModelWithLinks()`, `buildTestModelWithExpanded()` (`internal/render/integration_test.go`, `internal/graph/integration_test.go`), `testGraph(nodeID)` (`internal/render/render_test.go`)
- **Synthetic fallback fixture**: `loadCYPAuthInfraModel(t)` tries to load `cyp-auth-infra/cyp-auth-infra.toml` from several paths and, if missing, builds a synthetic equivalent via `createSyntheticCYPModel(t)` so tests still run (`internal/render/expanded_internal_test.go`)

**Location:**
- Shared parser fixtures: `testdata/` (repo root)
- CLI fixtures: `cmd/c4drill/testdata/`
- Live architecture example used as fixture: `cyp-auth-infra/cyp-auth-infra.toml`
- Example inputs validated in CI: `skill/examples/*.toml` (see `.github/workflows/validate-skill-examples.yml`)

## Coverage

**Requirements:** No enforced threshold (no `-covermode` gates, no CI coverage job). Coverage is measured by the standard mise task `go test -v -race -cover ./...`.

**View Coverage:**
```bash
go test -cover ./...          # per-package statement coverage
go test -coverprofile=/tmp/c.out ./... && go tool cover -html=/tmp/c.out
```

**Current coverage (measured 2026-08-05, `go test -cover ./...`, all tests pass):**

| Package | Coverage |
|---------|----------|
| `internal/validator` | 93.1% |
| `internal/graph` | 89.3% |
| `internal/render` | 88.4% |
| `internal/output` | 86.4% |
| `cmd/c4drill` | 77.6% |
| `internal/parser` | 76.2% |
| `internal/view` | 73.1% |
| `internal/model` | **0.0%** — no test files exist for `internal/model` (unit.go, link.go, properties.go, colors.go); its logic is exercised only indirectly through parser/validator tests |

## Test Types

**Unit Tests:**
- One test file per concern, testing a single package's public API (and internals via `*_internal_test.go`): parser tests parse TOML strings and assert field values; validator tests build models and assert error counts/messages; graph tests assert node/edge/cluster structure; render tests assert DOT/SVG string contents; output tests assert file layout.

**Integration Tests:**
- `integration_test.go` files exercise the cross-package pipeline without a real binary:
  - `internal/graph/integration_test.go`: `parser.ParseFile`/inline model → `view.GenerateC1View` → `graph.BuildGraph` → assert nodes/edges/shapes/icons
  - `internal/render/integration_test.go`: model → view → graph → `render.Render`/`RenderDOT`/`RenderSVG` → assert DOT/SVG content (subgraphs, HTML labels, shape attributes)
  - `internal/view/integration_test.go`: parser → view generation for nested/expanded hierarchies
  - `internal/output/integration_test.go`: writer + rendered output → file assertions
- **CLI-level tests** in `cmd/c4drill/root_test.go` run `NewRootCmd().Execute()` end-to-end (help text, flag validation, full pipeline parse→validate→render→write, exit-code semantics, stderr routing, silent-on-success, `--expanded` behavior)

**E2E Tests:**
- No dedicated E2E harness. CI (`validate-skill-examples.yml`) builds the binary and runs it against every `skill/examples/*.toml` with `-f dot` to validate examples on push/PR.

## Common Patterns

**Async/Parallel Testing:**
```go
// Safe (no graphviz): internal/parser/parser_test.go
func TestParseEmptyFile(t *testing.T) {
	t.Parallel()

	got, err := parser.Parse([]byte(""))
	require.NoError(t, err, "Parse() should not error for empty file")
	assert.Empty(t, got.Properties.Name, "Properties.Name should be empty")
}

// NOT safe (graphviz WASM): internal/render/render_test.go
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestRenderDOT(t *testing.T) {
	t.Run("returns valid DOT bytes for a simple graph with one node", func(t *testing.T) {
		...
	})
}
```

**Error Testing:**
```go
// internal/parser/parser_test.go
func TestParseInvalidTOML(t *testing.T) {
	t.Parallel()

	invalidData := []byte("invalid [[[")
	_, err := parser.Parse(invalidData)
	require.Error(t, err, "Parse() should error for invalid TOML")
	assert.Contains(t, err.Error(), "parse error", "error message should contain 'parse error'")

	var parseErr *parser.ParseError
	require.ErrorAs(t, err, &parseErr, "error should be *ParseError")
	assert.NotZero(t, parseErr.Line, "ParseError.Line should be non-zero for invalid TOML")
}
```
- Error type checks via `require.ErrorAs(t, err, &target, "message")`; message content via `assert.Contains` or `require.ErrorContains`
- String-exact error assertions used in `internal/validator/rules_test.go` with plain `t.Errorf`/`t.Fatalf` (this file is the exception; all other test files use testify)

**Table-Driven Testing:**
```go
// cmd/c4drill/root_test.go
tests := []struct {
	name        string
	format      string
	expectError bool
}{
	{name: "svg format is valid", format: "svg", expectError: false},
	{name: "png format is invalid", format: "png", expectError: true},
}
for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) { //nolint:paralleltest // go-graphviz WASM engine has concurrency issues
		...
	})
}
```

**Test Helpers:**
- Helpers receive `t *testing.T` as first argument and call `t.Helper()`: `loadCYPAuthInfraModel(t)`, `createSyntheticCYPModel(t)` (`internal/render/expanded_internal_test.go`)
- Non-helper factory functions return plain values: `testGraph(nodeID)`, `buildTestModel()`

---

*Testing analysis: 2026-08-05*
