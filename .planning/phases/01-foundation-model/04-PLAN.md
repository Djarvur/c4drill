---
phase: 01-foundation-model
plan: 04
type: tdd
wave: 3
depends_on: [02, 03]
files_modified:
  - internal/parser/parser_test.go
  - testdata/valid.toml
  - testdata/invalid.toml
  - testdata/nested.toml
autonomous: true
requirements:
  - QUAL-04
  - INPT-02
  - INPT-03
  - INPT-04
  - INPT-05
  - INPT-06
  - INPT-07
must_haves:
  truths:
    - "Parser correctly parses valid TOML files"
    - "Parser rejects invalid TOML with meaningful errors"
    - "Parser handles nested units at arbitrary depth"
    - "Parser extracts all link types correctly"
    - "Test coverage is at least 75%"
  artifacts:
    - path: "internal/parser/parser_test.go"
      provides: "Test coverage for parser"
      contains: "func Test"
    - path: "testdata/valid.toml"
      provides: "Sample valid TOML for testing"
    - path: "testdata/nested.toml"
      provides: "Sample with nested units"
  key_links:
    - from: "internal/parser/parser_test.go"
      to: "internal/parser/parser.go"
      via: "import"
      pattern: "parser\\.Parse"
---

<objective>
Create comprehensive tests for the parser to verify all parsing behaviors and achieve 75% coverage.

Purpose: Ensure the parser correctly handles all TOML structures and provides good error messages for invalid input.
Output: Test file with individual test functions (not table-driven per CONTEXT.md) covering all parsing scenarios.
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

From internal/model (Plan 01):
```go
type UnitType string
type Unit struct {
    Type        UnitType
    Name        string
    Description string
    Color       string
    Style       string
    Border      string
    Edges       string
    Expanded    []string
    Links       map[string]Link
    LinksFrom   map[string]Link
    Subunits    map[string]*Unit
}

type Link struct {
    Target string
    Arrow  ArrowDirection
    Rank   RankDirection
    Color  string
    Style  string
}

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
</interfaces>
</context>

<feature>
  <name>Parser Test Coverage</name>
  <files>internal/parser/parser_test.go, testdata/*.toml</files>
  <behavior>
    Test cases (individual functions per CONTEXT.md preference):

    1. TestParseEmpty: Empty TOML returns empty model (not nil)
    2. TestParseProperties: Properties section parsed correctly
    3. TestParseSingleUnit: Single unit at root level
    4. TestParseMultipleUnits: Multiple units at root level
    5. TestParseAllUnitTypes: All 14 unit types (C1, C2, C3) parsed correctly
    6. TestParseExternalVariants: personExternal, systemExternal, etc.
    7. TestParseNestedUnits: system containing containers, container containing components
    8. TestParseDeepNesting: 3+ levels of nesting (system -> container -> component)
    9. TestParseLinks: Link definitions with target, color, style
    10. TestParseLinksFrom: linkFrom definitions
    11. TestParseExpanded: expanded list on properties and units
    12. TestParseStyling: color, style, border, edges on units
    13. TestParseInvalidTOML: Invalid syntax returns error
    14. TestParseInvalidType: Unknown type value returns error
    15. TestParseFileNotFound: ParseFile with non-existent path returns error
    16. TestParseFileValid: ParseFile reads and parses valid file

    Each test creates TOML content as a string literal and calls Parse.
    Verify specific fields are populated correctly.
    For error cases, verify error is returned (not panic).
  </behavior>
  <implementation>
    Follow TDD RED-GREEN-REFACTOR:
    1. RED: Write test, run `go test ./internal/parser/...` - should FAIL (parser exists but test is new)
    2. GREEN: Tests should pass since parser is already implemented in Plan 02
    3. REFACTOR: Clean up test structure if needed

    Create testdata/ directory with sample files:
    - valid.toml: Simple valid architecture
    - nested.toml: Multi-level nesting example
    - invalid.toml: Invalid TOML syntax for error testing

    Use individual test functions (not table-driven) per CONTEXT.md decision.
    Name tests descriptively: TestParseProperties, TestParseNestedUnits, etc.
  </implementation>
</feature>

<tasks>

<task type="auto" tdd="true">
  <name>Task 1: Create testdata files</name>
  <files>testdata/valid.toml, testdata/nested.toml, testdata/invalid.toml</files>
  <behavior>
    - valid.toml: Contains [properties] and 2-3 units at root level
    - nested.toml: Contains system with containers, container with components
    - invalid.toml: Contains malformed TOML (e.g., unclosed string, bad syntax)
  </behavior>
  <action>
    Create testdata/ directory and three TOML files:

    1. testdata/valid.toml - Simple valid architecture:
       ```toml
       [properties]
       name = "Test System"
       description = "A test architecture"
       expanded = ["webapp"]

       [webapp]
       type = "system"
       name = "Web Application"
       description = "Main web system"

       [database]
       type = "db"
       name = "Database"
       description = "Primary database"
       ```

    2. testdata/nested.toml - Nested structure:
       ```toml
       [properties]
       name = "Nested Test"

       [webapp]
       type = "system"
       name = "Web App"

       [webapp.api]
       type = "container"
       name = "API"

       [webapp.api.handler]
       type = "component"
       name = "Handler"
       ```

    3. testdata/invalid.toml - Invalid syntax:
       ```toml
       [properties
       name = "unclosed bracket
       ```
  </action>
  <verify>
    <automated>test -f testdata/valid.toml && test -f testdata/nested.toml && test -f testdata/invalid.toml</automated>
  </verify>
  <done>testdata/ directory exists with valid.toml, nested.toml, invalid.toml</done>
</task>

<task type="auto" tdd="true">
  <name>Task 2: Create parser tests</name>
  <files>internal/parser/parser_test.go</files>
  <behavior>
    - TestParseEmpty: Empty input returns non-nil model with empty collections
    - TestParseProperties: Properties section fields extracted correctly
    - TestParseSingleUnit: Single unit with type, name, description
    - TestParseMultipleUnits: Multiple units at root level
    - TestParseAllUnitTypes: All 14 types recognized
    - TestParseExternalVariants: External type variants work
    - TestParseNestedUnits: Nested subunits populated
    - TestParseDeepNesting: 3+ levels work
    - TestParseLinks: Link map populated with correct fields
    - TestParseLinksFrom: linkFrom map populated
    - TestParseExpanded: expanded []string populated
    - TestParseStyling: color, style, border, edges fields
    - TestParseInvalidTOML: Returns error (not panic)
    - TestParseFileNotFound: Returns error for missing file
    - TestParseFileValid: ParseFile works with testdata/valid.toml
  </behavior>
  <action>
    Create internal/parser/parser_test.go with individual test functions.

    Follow this pattern for each test:
    ```go
    package parser_test

    import (
        "testing"
        "github.com/Djarvur/c4drill/internal/parser"
    )

    func TestParseEmpty(t *testing.T) {
        input := ""
        model, err := parser.Parse([]byte(input))
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if model == nil {
            t.Fatal("model is nil")
        }
        if len(model.Units) != 0 {
            t.Errorf("expected 0 units, got %d", len(model.Units))
        }
    }

    func TestParseProperties(t *testing.T) {
        input := `
    [properties]
    name = "Test"
    description = "Description"
    `
        model, err := parser.Parse([]byte(input))
        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if model.Properties.Name != "Test" {
            t.Errorf("expected name 'Test', got %q", model.Properties.Name)
        }
    }
    ```

    Write all 15+ test functions covering the behaviors listed above.
    Use the testdata files via parser.ParseFile for file-based tests.
  </action>
  <verify>
    <automated>go test ./internal/parser/... -v -count=1</automated>
  </verify>
  <done>All tests pass, coverage >= 75%, individual test functions (not table-driven)</done>
</task>

<task type="auto">
  <name>Task 3: Verify coverage threshold</name>
  <files>internal/parser/parser_test.go (may need additional tests)</files>
  <action>
    Run coverage check: `go test -cover ./internal/parser/...`

    If coverage < 75%, identify uncovered code paths and add tests:
    - Check coverage by function: `go test -coverprofile=coverage.out ./internal/parser/... && go tool cover -func=coverage.out`
    - Add tests for uncovered error paths, edge cases
    - Focus on Parse function branches and error handling

    Target: 75% minimum coverage as per QUAL-04 requirement.
  </action>
  <verify>
    <automated>go test -cover ./internal/parser/... | grep -E "coverage:.*[7-9][0-9]\.[0-9]%|coverage:.*100%"</automated>
  </verify>
  <done>Coverage report shows >= 75% for parser package</done>
</task>

</tasks>

<verification>
After completing all tasks:
1. Run `go test -cover ./internal/parser/...` - all tests pass, coverage >= 75%
2. Run `go test -cover ./...` - all tests pass
3. Run `mise lint` - no lint errors
</verification>

<success_criteria>
- [x] testdata/ directory with valid.toml, nested.toml, invalid.toml
- [x] parser_test.go with individual test functions
- [x] All tests pass
- [x] Coverage >= 75% for parser package
- [x] Tests cover: empty input, properties, units, all types, nesting, links, errors
- [x] No lint errors in test code
</success_criteria>

<output>
After completion, create `.planning/phases/01-foundation-model/04-SUMMARY.md`
</output>
