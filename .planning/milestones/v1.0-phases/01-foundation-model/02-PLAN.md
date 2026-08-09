---
phase: 01-foundation-model
plan: 02
type: execute
wave: 2
depends_on:
  - 01
files_modified:
  - internal/parser/parser.go
  - internal/parser/errors.go
autonomous: true
requirements:
  - INPT-01
  - INPT-02
  - INPT-03
  - INPT-04
  - INPT-05
  - INPT-06
  - INPT-07
must_haves:
  truths:
    - "Parser accepts TOML file path and returns populated Model struct"
    - "Parser handles nested unit definitions at arbitrary depth"
    - "Parser extracts properties section with name, description, styling, lineLength"
    - "Parser extracts all context-level unit types"
    - "Parser handles external variants (personExternal, systemExternal, etc.)"
    - "Parser extracts link and linkFrom definitions with styling attributes"
    - "Parser handles expanded list from properties and unit-level"
    - "Parser errors include line numbers from go-toml DecodeError"
  artifacts:
    - path: "internal/parser/parser.go"
      provides: "Parse function and Model struct"
      exports: ["Parse", "ParseFile", "Model"]
      min_lines: 80
    - path: "internal/parser/errors.go"
      provides: "Error wrapping utilities"
      min_lines: 20
  key_links:
    - from: "internal/parser/parser.go"
      to: "internal/model"
      via: "import"
      pattern: "github.com/Djarvur/c4drill/internal/model"
    - from: "internal/parser/parser.go"
      to: "go-toml v2"
      via: "import"
      pattern: "github.com/pelletier/go-toml/v2"
---

<objective>
Implement TOML parser using pelletier/go-toml v2 that unmarshals C4 architecture definitions into model structs.

Purpose: Transform TOML input files into typed Go structures that downstream phases can validate and render.
Output: Working parser that handles nested units, links, and properties.
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
@.planning/phases/01-foundation-model/01-PLAN.md

<interfaces>
<!-- Model types from Plan 01 - executor should use these directly -->

From internal/model/unit.go:
```go
type UnitType string

const (
    TypePerson, TypePersonExternal UnitType = "person", "personExternal"
    TypeSystem, TypeSystemExternal UnitType = "system", "systemExternal"
    TypeDb, TypeDbExternal         UnitType = "db", "dbExternal"
    TypeQueue, TypeQueueExternal   UnitType = "queue", "queueExternal"
    TypeBox                        UnitType = "box"
    TypeContainer, TypeContainerDb, TypeContainerQueue UnitType
    TypeComponent, TypeComponentDb, TypeComponentQueue UnitType
)

type Unit struct {
    Type         UnitType          `toml:"type"`
    Name         string            `toml:"name"`
    Description  string            `toml:"description"`
    Technology   string            `toml:"technology"`
    Color        string            `toml:"color"`
    Style        string            `toml:"style"`
    Border       string            `toml:"border"`
    Edges        string            `toml:"edges"`
    Width        float64           `toml:"width"`
    Height       float64           `toml:"height"`
    Expanded     []string          `toml:"expanded"`
    Links        map[string]Link   `toml:"link"`
    LinksFrom    map[string]Link   `toml:"linkFrom"`
    Subunits     map[string]*Unit  `toml:",inline"`
}
```

From internal/model/link.go:
```go
type ArrowDirection string  // Forward, Reverse, Bidirectional, None
type RankDirection string   // Forward, Reverse, Equal
type LabelPosition string   // Middle, Tail, Head

type Link struct {
    Target        string          `toml:"-"`  // Set from map key
    Arrow         ArrowDirection  `toml:"arrow"`
    Rank          RankDirection   `toml:"rank"`
    Color         string          `toml:"color"`
    Style         string          `toml:"style"`
    Technology    string          `toml:"technology"`
    Description   string          `toml:"description"`
    LabelPosition LabelPosition   `toml:"labelPosition"`
}
```

From internal/model/properties.go:
```go
type Properties struct {
    Name        string   `toml:"name"`
    Description string   `toml:"description"`
    Color       string   `toml:"color"`
    Style       string   `toml:"style"`
    Border      string   `toml:"border"`
    Edges       string   `toml:"edges"`
    LineLength  int      `toml:"lineLength"`
    Expanded    []string `toml:"expanded"`
}
```

**TOML Schema Reference (from CONTEXT.md):**
```toml
[properties]
name = "Project Name"
description = "Project description"
color = "transparent"
style = "none"
border = "transparent"
edges = "straight"
lineLength = 0
expanded = ["system1", "box1"]

[section_name]
type = "system"
name = "Display Name"
description = "Description"
technology = "Go, PostgreSQL"
color = "transparent"
style = "none"
border = "transparent"
edges = "straight"
width = 0
height = 0
expanded = ["subunit1"]
link = { "target" = { arrow = "forward", rank = "forward", color = "black", style = "solid", technology = "HTTP", description = "API calls", labelPosition = "middle" } }
linkFrom = { "source" = { ... } }

[section_name.subunit]
# Same structure, nested
```
</interfaces>
</context>

<tasks>

<task type="auto">
  <name>Task 1: Create parser Model struct and Parse functions</name>
  <files>internal/parser/parser.go</files>
  <action>
    Create internal/parser/parser.go with:

    1. Model struct that wraps the TOML document:
       ```go
       type Model struct {
           Properties model.Properties   `toml:"properties"`
           Units      map[string]*model.Unit `toml:",inline"`
       }
       ```
       The `toml:",inline"` tag captures all top-level sections (except [properties]) as units.

    2. Parse function that accepts []byte:
       ```go
       func Parse(data []byte) (*Model, error)
       ```
       - Use toml.Unmarshal from go-toml v2
       - Handle DecodeError to extract line numbers
       - After unmarshaling, populate Link.Target from map keys

    3. ParseFile function that reads from file path:
       ```go
       func ParseFile(path string) (*Model, error)
       ```
       - Use os.ReadFile to get content
       - Wrap file errors with context
       - Call Parse with the content

    4. Link target population:
       After unmarshaling, iterate over each unit's Links and LinksFrom maps to set the Target field from the map key:
       ```go
       for _, unit := range model.Units {
           for target, link := range unit.Links {
               link.Target = target
           }
           for source, link := range unit.LinksFrom {
               link.Target = source
           }
           // Also handle Subunits recursively
       }
       ```

    Key implementation notes:
    - go-toml v2 uses `toml.Unmarshal`, not `toml.Unmarshal` with Tree
    - DecodeError has a String() method that returns formatted error with line numbers
    - Must handle nested Subunits recursively to populate Link.Target
  </action>
  <verify>
    <automated>go build ./internal/parser/...</automated>
  </verify>
  <done>
    - Model struct has Properties and Units fields
    - Parse([]byte) function exists and returns (*Model, error)
    - ParseFile(string) function exists and returns (*Model, error)
    - Code compiles without errors
  </done>
</task>

<task type="auto">
  <name>Task 2: Implement error handling with line numbers</name>
  <files>internal/parser/errors.go</files>
  <action>
    Create internal/parser/errors.go with error handling utilities:

    1. ParseError struct that wraps go-toml errors:
       ```go
       type ParseError struct {
           Message string
           Line    int
           Context string
           Cause   error
       }
       ```
       Implement Error() method that returns human-readable format.

    2. wrapDecodeError function:
       ```go
       func wrapDecodeError(err error) error
       ```
       - Check if err is *toml.DecodeError using errors.As
       - Extract line number and context from DecodeError
       - Return ParseError with formatted message

    3. In parser.go, use the wrapper:
       ```go
       if err := toml.Unmarshal(data, &model); err != nil {
           return nil, wrapDecodeError(err)
       }
       ```

    The go-toml v2 DecodeError already provides formatted output with:
    - Line number
    - Surrounding context
    - Caret pointing to error position

    Use DecodeError.String() directly if it provides sufficient context, or wrap it for consistency.
  </action>
  <verify>
    <automated>go build ./internal/parser/...</automated>
  </verify>
  <done>
    - ParseError type exists with Line field
    - wrapDecodeError function extracts line numbers from DecodeError
    - Error messages are human-readable (not JSON)
  </done>
</task>

<task type="auto">
  <name>Task 3: Handle recursive Link.Target population</name>
  <files>internal/parser/parser.go</files>
  <action>
    Add a recursive helper function to populate Link.Target in all nested units:

    ```go
    func populateLinkTargets(units map[string]*model.Unit) {
        for _, unit := range units {
            // Populate Links map keys into Target field
            for target, link := range unit.Links {
                link.Target = target
            }
            // Populate LinksFrom map keys into Target field
            for source, link := range unit.LinksFrom {
                link.Target = source
            }
            // Recurse into subunits
            if unit.Subunits != nil {
                populateLinkTargets(unit.Subunits)
            }
        }
    }
    ```

    Call this from Parse() after unmarshaling:
    ```go
    populateLinkTargets(model.Units)
    ```

    This handles arbitrary nesting depth because it recursively processes all Subunits maps.
  </action>
  <verify>
    <automated>go build ./internal/parser/...</automated>
  </verify>
  <done>
    - populateLinkTargets function exists and is recursive
    - Called from Parse after unmarshaling
    - Handles nested units at any depth
  </done>
</task>

<task type="auto">
  <name>Task 4: Run lint and fix parser issues</name>
  <files>internal/parser/*.go</files>
  <action>
    Run `mise lint` and fix any lint errors in the parser package.

    Common fixes that may be needed:
    - Add comments for exported types and functions
    - Fix any import issues
    - Handle all error returns properly

    DO NOT modify .golangci.yml to silence errors per QUAL-02.
  </action>
  <verify>
    <automated>mise lint</automated>
  </verify>
  <done>
    - `mise lint` passes with no errors
    - No modifications to .golangci.yml
  </done>
</task>

</tasks>

<verification>
Phase 1 Plan 02 verification:

1. **Parser package compiles:**
   ```bash
   go build ./internal/parser/...
   # Should compile without errors
   ```

2. **Parser exports correct functions:**
   ```bash
   go doc ./internal/parser | grep -E "Parse|Model"
   # Should show Parse, ParseFile, Model
   ```

3. **Lint passes:**
   ```bash
   mise lint
   # Should pass with no errors
   ```
</verification>

<success_criteria>
- Parser accepts path to TOML input file (INPT-01)
- Parser handles nested unit definitions with arbitrary depth (INPT-02)
- Parser extracts properties section with name, description, styling (INPT-03)
- Parser extracts context-level units (INPT-04)
- Parser handles external variants (INPT-05)
- Parser extracts link and linkFrom definitions (INPT-06)
- Parser handles expanded list (INPT-07)
- Error messages include line numbers (VALD-05 preview)
</success_criteria>

<output>
After completion, create `.planning/phases/01-foundation-model/01-02-SUMMARY.md`
</output>
