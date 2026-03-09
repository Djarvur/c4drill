---
phase: 01-foundation-model
plan: 03
type: execute
wave: 2
depends_on: [02]
files_modified:
  - cmd/c4drill/main.go
autonomous: true
requirements:
  - INPT-01
must_haves:
  truths:
    - "CLI accepts a path argument to a TOML file"
    - "CLI parses the TOML file and prints the resulting model"
    - "CLI reports errors to stderr with context"
  artifacts:
    - path: "cmd/c4drill/main.go"
      provides: "CLI entry point for development testing"
      contains: "func main()"
  key_links:
    - from: "cmd/c4drill/main.go"
      to: "internal/parser"
      via: "import"
      pattern: "parser\\.ParseFile"
---

<objective>
Create a minimal CLI entry point that exercises the parser for development testing.

Purpose: Provide a way to verify the parser works end-to-end before building the full CLI in Phase 6.
Output: Simple main.go that accepts a TOML file path, parses it, and prints the model.
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
<!-- Parser interface from Plan 02 -->

From internal/parser/parser.go (Plan 02):
```go
type Model struct {
    Properties model.Properties
    Units      map[string]*model.Unit
}

func Parse(data []byte) (*Model, error)
func ParseFile(path string) (*Model, error)
```

From internal/model/unit.go (Plan 01):
```go
type Unit struct {
    Type        UnitType
    Name        string
    Description string
    // ... other fields
    Subunits    map[string]*Unit
}
```
</interfaces>
</context>

<tasks>

<task type="auto">
  <name>Task 1: Create CLI entry point</name>
  <files>cmd/c4drill/main.go</files>
  <action>
    Create cmd/c4drill/main.go:
    - Package main with func main()
    - Check os.Args length - expect exactly one argument (path)
    - If wrong args: print usage to stderr, exit with code 1
      - Usage: "c4drill <input.toml>"
    - Call parser.ParseFile with the path argument
    - On error: print error to stderr with fmt.Fprintf(os.Stderr, "error: %v\n", err), exit with code 1
    - On success: print model using fmt.Printf with %#v or custom formatting
      - Print Properties.Name and Properties.Description
      - Print count of top-level units
      - For each unit, print its Type, Name, and Description

    This is a DEVELOPMENT CLI only - full CLI with flags comes in Phase 6.
    Keep it simple: just enough to verify the parser works.

    DO NOT add flag parsing - that's Phase 6.
    DO NOT add sophisticated output formatting - just enough to see the model.
  </action>
  <verify>
    <automated>go build ./cmd/c4drill/...</automated>
  </verify>
  <done>cmd/c4drill/main.go compiles, accepts TOML path argument, parses file, prints model or error, exits with appropriate codes</done>
</task>

</tasks>

<verification>
After completing:
1. Build the CLI: `go build ./cmd/c4drill/...`
2. Run `mise lint` to verify no lint errors
3. Test with a sample TOML file (create minimal test file if needed)
</verification>

<success_criteria>
- [x] cmd/c4drill/main.go exists and compiles
- [x] CLI accepts path argument from os.Args
- [x] CLI calls parser.ParseFile
- [x] CLI prints model on success
- [x] CLI prints error to stderr on failure
- [x] CLI exits with code 1 on error, 0 on success
- [x] No lint errors in new code
</success_criteria>

<output>
After completion, create `.planning/phases/01-foundation-model/03-SUMMARY.md`
</output>
