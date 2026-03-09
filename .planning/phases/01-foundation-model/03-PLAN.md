---
phase: 01-foundation-model
plan: 03
type: execute
wave: 3
depends_on:
  - 02
files_modified:
  - cmd/c4drill/main.go
  - internal/parser/parser_test.go
  - testdata/valid.toml
  - testdata/nested.toml
  - testdata/links.toml
autonomous: false
requirements:
  - INPT-01
  - QUAL-01
  - QUAL-04
must_haves:
  truths:
    - "CLI entry point accepts TOML file path as argument"
    - "CLI parses TOML and prints model to stdout for verification"
    - "Parser tests cover basic parsing, nested units, and links"
    - "Parser tests achieve 75% code coverage"
    - "All tests pass with `mise test`"
  artifacts:
    - path: "cmd/c4drill/main.go"
      provides: "CLI entry point"
      min_lines: 30
    - path: "internal/parser/parser_test.go"
      provides: "Parser unit tests"
      min_lines: 150
    - path: "testdata/valid.toml"
      provides: "Simple valid TOML test fixture"
    - path: "testdata/nested.toml"
      provides: "Nested units test fixture"
    - path: "testdata/links.toml"
      provides: "Link definitions test fixture"
  key_links:
    - from: "cmd/c4drill/main.go"
      to: "internal/parser"
      via: "import and call"
      pattern: "parser.ParseFile"
    - from: "internal/parser/parser_test.go"
      to: "testdata/*.toml"
      via: "os.ReadFile"
      pattern: "testdata/"
---

<objective>
Create CLI entry point for development testing and comprehensive parser tests achieving 75% coverage.

Purpose: Verify parser works correctly with real TOML files and establish test patterns for the project.
Output: Working CLI that parses and prints models, parser tests with 75%+ coverage.
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
@.planning/phases/01-foundation-model/02-PLAN.md

<interfaces>
<!-- Parser interface from Plan 02 -->

From internal/parser/parser.go:
```go
type Model struct {
    Properties model.Properties      `toml:"properties"`
    Units      map[string]*model.Unit `toml:",inline"`
}

func Parse(data []byte) (*Model, error)
func ParseFile(path string) (*Model, error)
```

**Test style preference (from CONTEXT.md):**
- Individual test functions (not table-driven)
- User preference per RESEARCH.md

**Coverage requirement (QUAL-04):**
- Minimum 75% test coverage
- Enforced via `go test -cover ./...`
</interfaces>
</context>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Create test fixtures</name>
  <files>testdata/valid.toml, testdata/nested.toml, testdata/links.toml</files>
  <behavior>
    - Test fixture: valid.toml with properties and basic units
    - Test fixture: nested.toml with system containing subunits
    - Test fixture: links.toml with link and linkFrom definitions
  </behavior>
  <action>
    Create testdata directory and TOML fixtures:

    1. testdata/valid.toml - Simple valid TOML:
       ```toml
       [properties]
       name = "Test System"
       description = "A test architecture"
       color = "transparent"
       edges = "straight"
       lineLength = 40
       expanded = ["webapp"]

       [user]
       type = "person"
       name = "User"
       description = "End user of the system"

       [webapp]
       type = "system"
       name = "Web Application"
       description = "Main web application"
       technology = "Go, React"
       ```

    2. testdata/nested.toml - Nested units:
       ```toml
       [properties]
       name = "Nested Test"
       edges = "straight"

       [externals]
       type = "systemExternal"
       name = "External System"
       description = "An external dependency"

       [mainapp]
       type = "system"
       name = "Main Application"
       technology = "Go"

       [mainapp.api]
       type = "container"
       name = "API Server"
       technology = "Go"

       [mainapp.db]
       type = "containerDb"
       name = "Database"
       technology = "PostgreSQL"

       [mainapp.api.handler]
       type = "component"
       name = "Request Handler"
       technology = "Go"
       ```

    3. testdata/links.toml - Link definitions:
       ```toml
       [properties]
       name = "Links Test"
       edges = "spline"

       [user]
       type = "person"
       name = "User"

       [webapp]
       type = "system"
       name = "Web App"
       link = { "user" = { arrow = "forward", rank = "forward", technology = "HTTPS", description = "Uses" } }

       [api]
       type = "system"
       name = "API"
       linkFrom = { "webapp" = { arrow = "forward", technology = "HTTP/JSON" } }
       ```
  </action>
  <verify>
    <automated>test -f testdata/valid.toml && test -f testdata/nested.toml && test -f testdata/links.toml</automated>
  </verify>
  <done>
    - testdata/ directory exists
    - valid.toml, nested.toml, links.toml files exist
    - Files contain valid TOML syntax
  </done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Write parser unit tests</name>
  <files>internal/parser/parser_test.go</files>
  <behavior>
    - Test ParseFile with valid.toml returns non-nil model
    - Test ParseFile with nested.toml returns nested units
    - Test ParseFile with links.toml populates Link.Target
    - Test Parse with invalid TOML returns error with line info
    - Test Parse with missing file returns error
  </behavior>
  <action>
    Create internal/parser/parser_test.go with individual test functions (not table-driven per user preference):

    1. TestParseValid - Parse valid.toml:
       - Read testdata/valid.toml
       - Call ParseFile
       - Assert no error
       - Assert model.Properties.Name == "Test System"
       - Assert model.Units contains "user" and "webapp"
       - Assert user unit has TypePerson type

    2. TestParseNestedUnits - Parse nested.toml:
       - Read testdata/nested.toml
       - Call ParseFile
       - Assert mainapp unit exists
       - Assert mainapp.Subunits contains "api" and "db"
       - Assert api.Subunits contains "handler" (3 levels deep)

    3. TestParseLinks - Parse links.toml:
       - Read testdata/links.toml
       - Call ParseFile
       - Assert webapp.Links contains "user" entry
       - Assert Link.Target is populated with "user"
       - Assert Link.Technology == "HTTPS"

    4. TestParseInvalidTOML - Parse invalid TOML:
       - Call Parse with "invalid [[[" bytes
       - Assert error is not nil
       - Assert error message contains line info (ParseError or DecodeError)

    5. TestParseMissingFile - Parse non-existent file:
       - Call ParseFile with "nonexistent.toml"
       - Assert error is not nil

    6. TestParsePropertiesFields - Verify all properties fields:
       - Parse valid.toml
       - Assert Properties.LineLength == 40
       - Assert Properties.Expanded contains "webapp"

    7. TestParseUnitFields - Verify all unit fields:
       - Parse valid.toml
       - Assert webapp.Technology == "Go, React"
       - Assert webapp.Type == TypeSystem

    Each test should use testing.T and t.Fatal/t.Errorf for assertions.
    Use individual test functions, NOT table-driven tests (per user preference in CONTEXT.md).
  </action>
  <verify>
    <automated>go test ./internal/parser/... -v -count=1</automated>
  </verify>
  <done>
    - All tests pass
    - Coverage for parser package >= 75%
    - Tests use individual functions (not table-driven)
  </done>
</task>

<task type="auto">
  <name>Task 3: Create CLI entry point</name>
  <files>cmd/c4drill/main.go</files>
  <action>
    Create cmd/c4drill/main.go with simple CLI for development testing:

    ```go
    package main

    import (
        "fmt"
        "os"

        "github.com/Djarvur/c4drill/internal/parser"
    )

    func main() {
        if len(os.Args) < 2 {
            fmt.Fprintln(os.Stderr, "Usage: c4drill <input.toml>")
            os.Exit(1)
        }

        path := os.Args[1]
        model, err := parser.ParseFile(path)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", path, err)
            os.Exit(1)
        }

        // Print model for verification (simple debug output)
        fmt.Printf("Properties: %s\n", model.Properties.Name)
        fmt.Printf("Units: %d\n", len(model.Units))
        for name, unit := range model.Units {
            fmt.Printf("  - %s: %s (%s)\n", name, unit.Name, unit.Type)
            if len(unit.Subunits) > 0 {
                for subname, subunit := range unit.Subunits {
                    fmt.Printf("    - %s: %s (%s)\n", subname, subunit.Name, subunit.Type)
                }
            }
        }
    }
    ```

    This is a simple main.go for Phase 1 testing. Full CLI with flags comes in Phase 6.
    Errors go to stderr, exit code is non-zero on failure (per CLII-05, CLII-06).
  </action>
  <verify>
    <automated>go build ./cmd/c4drill/...</automated>
  </verify>
  <done>
    - cmd/c4drill/main.go exists and compiles
    - Accepts file path argument
    - Prints parsed model to stdout
    - Prints errors to stderr
    - Exits with non-zero code on error
  </done>
</task>

<task type="auto">
  <name>Task 4: Verify coverage and run all tests</name>
  <files>internal/parser/parser_test.go</files>
  <action>
    Run full test suite with coverage:

    1. Run `mise test` to verify all tests pass
    2. Check coverage: `go test -cover ./internal/parser/...`
    3. If coverage < 75%, add more tests:
       - Test all unit types are parsed correctly
       - Test all link fields are populated
       - Test edge cases (empty file, only properties, etc.)

    Add additional tests if needed:
    - TestParseEmptyFile - empty TOML
    - TestParseOnlyProperties - only [properties] section
    - TestParseExternalTypes - all external type variants
    - TestParseLinkFrom - verify linkFrom direction

    Ensure at least 75% coverage before proceeding.
  </action>
  <verify>
    <automated>go test -cover ./internal/parser/... | grep -E "coverage:.*%.*75"</automated>
  </verify>
  <done>
    - All tests pass with `mise test`
    - Parser coverage >= 75%
    - `mise lint` passes
  </done>
</task>

<task type="checkpoint:human-verify">
  <name>Task 5: Verify complete Phase 1 foundation</name>
  <files>N/A</files>
  <action>
    Human verification checkpoint for Phase 1 completion.
  </action>
  <what-built>
    Complete Phase 1 foundation:
    - Development environment (mise test, mise lint)
    - Domain types (Unit, Link, Properties, colors)
    - TOML parser with error handling
    - CLI entry point
    - Parser tests with 75%+ coverage
  </what-built>
  <how-to-verify>
    1. Run `mise test` - should pass with all tests green
    2. Run `mise lint` - should pass with no errors
    3. Build CLI: `go build ./cmd/c4drill`
    4. Run CLI on test file: `./c4drill testdata/nested.toml`
    5. Verify output shows nested units correctly
    6. Check coverage: `go test -cover ./...`
  </how-to-verify>
  <verify>
    <automated>echo "Checkpoint: human verification required"</automated>
  </verify>
  <resume-signal>Type "approved" or describe issues found</resume-signal>
  <done>
    - User confirms all verification steps pass
    - Phase 1 complete and ready for Phase 2
  </done>
</task>

</tasks>

<verification>
Phase 1 Plan 03 verification:

1. **All tests pass:**
   ```bash
   mise test
   # Should show PASS for all tests
   ```

2. **Coverage meets threshold:**
   ```bash
   go test -cover ./...
   # Should show >= 75% coverage for parser package
   ```

3. **Lint passes:**
   ```bash
   mise lint
   # Should pass with no errors
   ```

4. **CLI works:**
   ```bash
   go build ./cmd/c4drill && ./c4drill testdata/valid.toml
   # Should print model structure
   ```

5. **CLI handles errors:**
   ```bash
   ./c4drill nonexistent.toml
  # Should print error to stderr and exit non-zero
   ```
</verification>

<success_criteria>
- CLI accepts path to TOML input file (INPT-01)
- All lint errors fixed before commit (QUAL-01)
- Minimum 75% test coverage achieved (QUAL-04)
- Tests use individual functions per user preference
- CLI prints parsed model for verification
- Error output goes to stderr
</success_criteria>

<output>
After completion, create `.planning/phases/01-foundation-model/01-03-SUMMARY.md`
</output>
