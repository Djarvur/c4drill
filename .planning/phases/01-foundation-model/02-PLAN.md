---
phase: 01-foundation-model
plan: 02
type: execute
wave: 2
depends_on: [01]
files_modified:
  - internal/parser/parser.go
  - internal/parser/errors.go
autonomous: true
requirements:
  - INPT-02
  - INPT-03
  - INPT-04
  - INPT-05
  - INPT-06
  - INPT-07
must_haves:
  truths:
    - "Parser accepts valid TOML content and returns populated Model struct"
    - "Parser handles nested unit definitions at arbitrary depth"
    - "Parser extracts properties section correctly"
    - "Parser extracts all unit types including external variants"
    - "Parser handles link and linkFrom definitions"
    - "Parser returns rich error context with line numbers"
  artifacts:
    - path: "internal/parser/parser.go"
      provides: "Parse function and Model struct"
      exports: ["Parse", "Model", "ParseFile"]
    - path: "internal/parser/errors.go"
      provides: "Error formatting with context"
      exports: ["ParseError"]
  key_links:
    - from: "internal/parser/parser.go"
      to: "github.com/pelletier/go-toml/v2"
      via: "import"
      pattern: "toml.Unmarshal"
    - from: "internal/parser/parser.go"
      to: "internal/model"
      via: "import"
      pattern: "model\\.Unit|model\\.Properties"
---

<objective>
Implement the TOML parser that unmarshals architecture definitions into the domain model.

Purpose: Transform TOML text into structured Go data that downstream phases can validate and render.
Output: Parser package with Parse function that handles nested units, properties, and links with rich error context.
</objective>

<execution_context>
@/Users/nil/.claude/get-shit-done/workflows/execute-plan.md
@/Users/nil/.claude/get-shit-done/templates/summary.md
</execution_context>

<context>
@.planning/PROJECT.md
@.planning/ROADMAP.md
@.planning/STATE.md
@.planning/1-CONTEXT.md
@.planning/phases/01-foundation-model/01-RESEARCH.md

<interfaces>
<!-- Types from Plan 01 that parser must use -->

From internal/model/unit.go (Plan 01):
```go
type UnitType string
type Unit struct {
    Type        UnitType          `toml:"type"`
    Name        string            `toml:"name"`
    Description string            `toml:"description"`
    Color       string            `toml:"color"`
    Style       string            `toml:"style"`
    Border      string            `toml:"border"`
    Edges       string            `toml:"edges"`
    Expanded    []string          `toml:"expanded"`
    Links       map[string]Link   `toml:"link"`
    LinksFrom   map[string]Link   `toml:"linkFrom"`
    Subunits    map[string]*Unit  `toml:",inline"`
}
```

From internal/model/link.go (Plan 01):
```go
type Link struct {
    Target string
    Arrow  ArrowDirection
    Rank   RankDirection
    Color  string
    Style  string
}
```

From internal/model/properties.go (Plan 01):
```go
type Properties struct {
    Name        string
    Description string
    Color       string
    Style       string
    Border      string
    Edges       string
    Expanded    []string
}
```

From 01-RESEARCH.md - go-toml v2 patterns:
```go
import "github.com/pelletier/go-toml/v2"

// DecodeError provides rich context
var decodeErr *toml.DecodeError
if errors.As(err, &decodeErr) {
    // decodeErr.String() returns formatted error with line numbers
    return fmt.Errorf("parse error:\n%s", decodeErr)
}

// Model struct for unmarshaling
type Model struct {
    Properties Properties           `toml:"properties"`
    Units      map[string]*Unit     `toml:",inline"` // Top-level units
}
```
</interfaces>
</context>

<tasks>

<task type="auto">
  <name>Task 1: Create parser with TOML unmarshaling</name>
  <files>internal/parser/parser.go, go.mod</files>
  <action>
    1. Add dependency: `go get github.com/pelletier/go-toml/v2@latest`
       This updates go.mod and go.sum with the TOML library.

    2. Create internal/parser/parser.go:
       - Define Model struct containing:
         - Properties model.Properties `toml:"properties"`
         - Units map[string]*model.Unit `toml:",inline"` (captures arbitrary top-level units)
       - Define `func Parse(data []byte) (*Model, error)`:
         - Create empty Model struct
         - Call toml.Unmarshal(data, &model)
         - Handle *toml.DecodeError specially - wrap with fmt.Errorf("TOML parsing error:\n%s", decodeErr)
         - Return populated model on success
       - Define `func ParseFile(path string) (*Model, error)`:
         - Read file with os.ReadFile
         - Delegate to Parse
         - Wrap file errors with context

    Key implementation notes from RESEARCH.md:
    - Use `toml:",inline"` tag on Units field to capture arbitrary top-level tables
    - Use map[string]*Unit (pointer) for recursive type definition
    - go-toml v2 DecodeError.String() already provides formatted output with line numbers
    - DO NOT use toml.Tree API (v1 only) or toml.Unmarshaler interface (dropped in v2)
  </action>
  <verify>
    <automated>go build ./internal/parser/...</automated>
  </verify>
  <done>Parse function unmarshals TOML to Model struct, handles DecodeError with line context, ParseFile reads from filesystem, go.mod updated with go-toml v2 dependency</done>
</task>

<task type="auto">
  <name>Task 2: Create error formatting helper</name>
  <files>internal/parser/errors.go</files>
  <action>
    Create internal/parser/errors.go:
    - Define ParseError struct with:
      - Message string
      - Line int (0 if unknown)
      - Context string (surrounding text from DecodeError)
    - Implement error interface: `func (e *ParseError) Error() string`
      - Format as "line N: message" if line known
      - Format as "message" if line unknown
    - Implement `func wrapDecodeError(err error) error`:
      - Type assert to *toml.DecodeError
      - Extract line number and context from DecodeError
      - Return ParseError with rich information

    This provides a clean error type that downstream code can inspect while preserving the rich context from go-toml.

    Note: go-toml v2's DecodeError.String() already formats nicely. The ParseError wraps this for consistent error handling across the codebase.
  </action>
  <verify>
    <automated>go build ./internal/parser/...</automated>
  </verify>
  <done>ParseError struct defined with Line and Context fields, Error() method formats appropriately, wrapDecodeError extracts context from go-toml DecodeError</done>
</task>

</tasks>

<verification>
After completing all tasks:
1. Run `go build ./...` to verify compilation
2. Run `mise lint` to verify no lint errors
3. Verify go.mod contains github.com/pelletier/go-toml/v2
</verification>

<success_criteria>
- [x] go-toml v2 dependency added to go.mod
- [x] Parse function accepts []byte and returns (*Model, error)
- [x] ParseFile function accepts path and returns (*Model, error)
- [x] Model struct contains Properties and Units fields
- [x] Nested units handled via map[string]*Unit with inline tag
- [x] DecodeError wrapped with line number context
- [x] ParseError type defined for consistent error handling
- [x] All code compiles without errors
- [x] No lint errors in new code
</success_criteria>

<output>
After completion, create `.planning/phases/01-foundation-model/02-SUMMARY.md`
</output>
