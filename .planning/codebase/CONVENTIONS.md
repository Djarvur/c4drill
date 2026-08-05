# Coding Conventions

**Analysis Date:** 2026-08-05

## Overview

This is a Go 1.26 CLI project (`module github.com/Djarvur/c4drill`) that converts TOML C4 architecture definitions into DOT/SVG diagrams. It uses a strict linting setup (golangci-lint v2 with `default: all`) and a layered package architecture under `internal/`. All code is gofmt-formatted and every exported declaration carries a doc comment.

## Naming Patterns

**Files:**
- Source files: single lowercase word per package concern, e.g., `parser.go`, `builder.go`, `labels.go`, `scope.go`, `writer.go`, `wrap.go`
- Test files: `<name>_test.go` for black-box tests, `<name>_internal_test.go` for white-box (internal package) tests, `integration_test.go` for cross-package pipeline tests
- Configuration files: `.golangci.yml`, `.mise.toml` (no Makefile)

**Functions:**
- Exported: `PascalCase` — `Parse`, `Validate`, `BuildIndex`, `GenerateC1View`, `RenderSVG`, `NewWriter` (`cmd/c4drill/root.go`, `internal/validator/validator.go`, `internal/render/render.go`, `internal/output/writer.go`)
- Unexported: `camelCase` — `parseUnitWithOrder`, `wrapDecodeError`, `buildBoundaryCluster`, `collectExpandedPaths`, `isC2Path` (`internal/parser/parser.go`, `internal/graph/builder.go`, `cmd/c4drill/root.go`)
- Constructors named `New<Type>`: `NewRootCmd`, `NewWriter` (`cmd/c4drill/root.go`, `internal/output/writer.go`)

**Variables:**
- Standard Go `camelCase`, short names idiomatic (varnamelen linter explicitly disabled in `.golangci.yml`)
- Package-level state is rare and always annotated `//nolint:gochecknoglobals` with an explanation, e.g. the Cobra flag vars in `cmd/c4drill/root.go` and `render.LabelRatio` in `internal/render/wrap.go`

**Types:**
- Exported types are singular nouns: `Unit`, `Link`, `Model`, `View`, `Writer`, `Graph`, `Node`, `Edge`, `Cluster` (`internal/model/`, `internal/view/view.go`, `internal/graph/graph.go`)
- Error types end in `Error`/`Errors`: `ParseError`, `ValidationError`, `ValidationErrors` (`internal/parser/errors.go`, `internal/validator/errors.go`)
- String-based enums declared as `type X string` with typed constants: `UnitType`, `ArrowDirection`, `RankDirection`, `LabelPosition` (`internal/model/unit.go`, `internal/model/link.go`)
- Integer enums use `iota`: `LevelC1 Level = iota` (`internal/view/view.go`)

**Constants:**
- Exported typed constants: `PascalCase` — `TypePerson`, `ArrowForward`, `RankEqual`, `LabelHead` (`internal/model/`)
- Unexported constants: `camelCase` — `formatDot`, `formatSVG` (`cmd/c4drill/root.go`), `typicalErrorCount` (`internal/validator/validator.go`), `pointsPerChar`, `pointsPerRow` (`internal/render/wrap.go`)
- Constant groups are organized by semantic level in separate `const` blocks with a comment header, e.g. "C1 Context level unit types", "C2 Container level unit types" (`internal/model/unit.go`)

**Sentinel errors:**
- Exported: `Err` prefix — `ErrNilGraph`, `ErrUnsupportedFormat` (`internal/render/render.go`)
- Unexported: `err` prefix — `errInvalidFormat`, `errValidationFailed`, `errGenerateView`, `errBuildGraph` (`cmd/c4drill/root.go`), grouped in a `var (...)` block with the comment `// Static errors for better error handling.`

## Code Style

**Formatting:**
- `gofmt` (enforced by golangci-lint)
- Blank line separates every statement/logical block (visible throughout — e.g., each `if err != nil { return ... }` is followed by a blank line)
- Octal permissions written with the `0o` prefix: `0o750`, `0o600` (`internal/output/writer.go`)
- Imports grouped: standard library first, then third-party/module, blank line between groups

**Linting:**
- golangci-lint v2, `linters.default: all` (all linters enabled), `tests: true` (test files linted too) — `.golangci.yml`
- Disabled linters (with reasons in config): `wsl`, `exhaustruct`, `depguard`, `noinlineerr`, `ireturn`, `varnamelen`, `tagliatelle`, `dupl`, `cyclop`
- `gocognit` min-complexity raised to 15 (`.golangci.yml`)
- `//nolint:` directives are used and ALWAYS paired with an inline explanation comment:
  - `//nolint:paralleltest // go-graphviz WASM engine has concurrency issues` (81 occurrences — dominant)
  - `//nolint:gochecknoglobals // Cobra flags require package-level variables...` / `// ... immutable after init` / `// CLI configuration set before render calls`
  - `//nolint:gosec // G304: Path is provided by caller, this is intentional for CLI tool` (`internal/parser/parser.go`)
  - `//nolint:funlen // Test helper that creates multiple related units` (`internal/render/expanded_internal_test.go`)
  - `//nolint:exhaustive // Default case handles all remaining types` (`internal/parser/parser.go`)
  - `//nolint:revive // Function name matches plan specification (04-01-PLAN.md)` (`internal/render/render.go`)
- Lint commands (from `.mise.toml` tasks): `golangci-lint run ./...`, `golangci-lint run --fix ./...`

## Import Organization

**Order:**
1. Standard library (`os`, `fmt`, `strings`, `slices`, `maps`, `testing`, ...)
2. Module packages (`github.com/Djarvur/c4drill/internal/...`)
3. Third-party (`github.com/spf13/cobra`, `github.com/stretchr/testify`, `github.com/pelletier/go-toml/v2`, `github.com/goccy/go-graphviz`)

**Path Aliases:**
- None. All imports use the full module path `github.com/Djarvur/c4drill/internal/<pkg>`. No relative or dot imports.

## Error Handling

**Patterns:**
- **Custom error types** carry structured context and implement `Error()` plus `Unwrap()` for `errors.Is`/`errors.As` compatibility:
  - `parser.ParseError{Message, Line, Context, Cause}` with a human-readable `Error()` that includes line number when known (`internal/parser/errors.go`)
  - `validator.ValidationError{Message, Path, Line}` with format-priority `Error()` (line > path > plain) and a `ValidationErrors` slice type whose `Error()` returns a count summary (`internal/validator/errors.go`)
- **Sentinel errors** declared with `errors.New` at package level and wrapped at the call site:
  - `fmt.Errorf("parse: %w", err)`, `fmt.Errorf("render: %w", err)`, `fmt.Errorf("write: %w", err)` — lowercase single-word stage prefix (`cmd/c4drill/root.go`)
  - `fmt.Errorf("create graphviz instance: %w", err)`, `fmt.Errorf("render output: %w", err)` (`internal/render/render.go`)
  - `fmt.Errorf("%w: %q", errInvalidFormat, format)` — wrap sentinel with value (`cmd/c4drill/root.go`)
- **Validation is not fail-fast**: `Validate` runs all rule functions and appends every error to a preallocated slice (`make(ValidationErrors, 0, typicalErrorCount)`) (`internal/validator/validator.go`)
- **Nil guards at public API boundaries**: `ParseFile`, `Validate`, `BuildGraph`, `render`, `GenerateC1View` return `nil`/empty for nil input (`internal/validator/validator.go`, `internal/graph/builder.go`, `internal/view/scope.go`)
- **Wrapper functions normalize third-party errors**: `wrapDecodeError` converts `*toml.DecodeError`/`*unstable.ParserError` into `*ParseError` with extracted line numbers (`internal/parser/errors.go`)
- Error responses in the CLI layer are single-line, wrapped stage errors; `SilenceUsage: true` on the root command suppresses usage text on runtime errors (`cmd/c4drill/root.go`)
- Deferred `Close()` calls ignore errors — explicitly excluded via golangci-lint exclusion rule in `.golangci.yml`
- Type-assertion errors written with `errors.AsType[*toml.DecodeError](err)` (Go 1.22+ generic helper) (`internal/parser/errors.go`)

## Logging

**Framework:** None. No `log` package or third-party logger is used anywhere in source code. The project is intentionally silent:

**Patterns:**
- The CLI prints nothing on success — `return nil // Success - silent per spec` (`cmd/c4drill/root.go`)
- User-facing error output goes to stderr via the cobra writer: `validator.ReportErrors(valErrors, cmd.OutOrStderr())` (`cmd/c4drill/root.go`, `internal/validator/validator.go`)
- `ReportErrors(errors, w io.Writer)` writes each error line plus a count summary ("N errors found") and returns a process exit code (1 if errors, 0 if none) (`internal/validator/validator.go`)
- Test helpers may use `t.Logf` (single occurrence: synthetic fixture fallback in `internal/render/expanded_internal_test.go`)

## Comments

**When to Comment:**
- Every exported type, function, constant, and method has a doc comment starting with its identifier name (Go convention), e.g. `// Parse parses TOML data into a Model.` (`internal/parser/parser.go`)
- Package doc comment at the top of the primary file of each package: `// Package render provides functions to render graph structures to DOT and SVG formats.` (`internal/render/render.go`)
- Every struct field is commented (e.g., all fields of `Unit`, `Link`, `Properties`, `View`, `Entry`, error types)
- Multi-step algorithms carry a "// Stage N:" / "// Pass N:" comment sequence, e.g. `// Stage 1: Parse`, `// Stage 2: Validate` (`cmd/c4drill/root.go`), `// First pass: capture definition order...` (`internal/parser/parser.go`)
- Comments frequently document "why", not just "what": `// Exit 0 is implicit` (`cmd/c4drill/main.go`), `// Success - silent per spec`
- Requirement traceability comments reference plan/context docs: `// Per CONTEXT.md: Person has NO technology field.` (`internal/render/labels.go`), `// Per CLII-06: errors go to stderr` (`cmd/c4drill/root_test.go`), `// REFINED-02: HTML tables should include border=...` (`internal/render/expanded_internal_test.go`)
- `//nolint:` directives always have a trailing explanation comment
- Test functions documenting multiple checks use `// Test N:` numbered comments mapping to spec cases (`internal/graph/graph_test.go`, `internal/graph/builder_test.go`)

**JSDoc/TSDoc:** Not applicable (Go). Doc comments are plain `//` style, single-line or paragraph.

## Function Design

**Size:**
- Small focused functions with single responsibility. Complex logic split into helpers: `parseUnitWithOrder`, `captureDefinitionOrder`, `wrapDecodeError` in `internal/parser/`; `buildBoundaryCluster`, `buildCluster`, `buildNode`, `buildEdges` in `internal/graph/builder.go`
- The only >100-line functions are top-level orchestrators (`runRoot` in `cmd/c4drill/root.go`, `addUnitRecursive` family in `internal/view/scope.go`) and test helpers (allowed via `//nolint:funlen`)
- `gocognit` complexity limit set to 15 in `.golangci.yml`

**Parameters:**
- Multiple related parameters are kept, or grouped into a struct when they grow (e.g. `parseUnitWithOrder(name, value, parentType, subunitOrder, subunitOrders)` in `internal/parser/parser.go`)
- Interfaces accepted where relevant: `ReportErrors(errors ValidationErrors, w io.Writer)` (`internal/validator/validator.go`)

**Return Values:**
- Idiomatic `(T, error)` pairs
- `(T, bool)` for lookup helpers: `FindLinkByPeer(links []Link, peer string) (*Link, bool)` (`internal/model/link.go`)
- Validation APIs return an error *slice* (`ValidationErrors`), not a single error, because they aggregate (`internal/validator/validator.go`)
- Rendering returns `([]byte, error)` for all formats (`internal/render/render.go`)

## Module Design

**Exports:**
- Each `internal/` package exposes a small, deliberate surface (types + functions) and keeps unexported helpers in the same files. No `internal` sub-package visibility tricks except the standard Go `internal/` boundary (only `cmd/c4drill` and sibling packages may import them)
- Layered dependency direction (strict, no cycles): `parser → model`, `validator → parser/model`, `view → parser/model`, `graph → view/model`, `render → graph`, `output → (standalone)`, `cmd/c4drill → all`
- A single package-level mutable config point exists: `render.LabelRatio` set by the CLI before render (`internal/render/wrap.go`)

**Barrel Files:** None used (not idiomatic in Go). Types are consumed directly from their defining package, e.g. `model.TypeSystem`, `graph.ShapeHTML`, `view.LevelC1`.

## Testing Conventions (summary)

See TESTING.md for the full treatment. Key rules:

- Use `github.com/stretchr/testify/require` for fatal conditions (nil checks, error expectations) and `assert` for non-fatal value checks
- Prefer `require.NoError` + explicit `assert.Equal(t, expected, actual, "<Field>")` with a field-path message argument
- Use external test packages (`package parser_test`) for public APIs; `package <pkg>` in `*_internal_test.go` only when unexported identifiers must be tested
- Call `t.Parallel()` unless the test touches the go-graphviz WASM renderer, in which case annotate `//nolint:paralleltest // go-graphviz WASM engine has concurrency issues`
- Read shared fixtures from the repo-root `testdata/` directory via `os.ReadFile("../../testdata/<name>.toml")`
- Table-driven tests use `tests := []struct{ name string; ... }{...}` with `for _, tt := range tests { t.Run(tt.name, func(t *testing.T) { ... }) }`
- Test helpers take `t *testing.T` as first argument and call `t.Helper()`; name them `buildTestModel`, `testGraph`, `load<Fixture>Model`

---

*Convention analysis: 2026-08-05*
