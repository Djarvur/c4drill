---
phase: 01-foundation-model
plan: 01
type: execute
wave: 1
depends_on: []
files_modified:
  - go.mod
  - .mise.toml
  - internal/model/unit.go
  - internal/model/link.go
  - internal/model/properties.go
  - internal/model/colors.go
autonomous: true
requirements:
  - DEVI-01
  - DEVI-02
  - DEVI-03
  - DEVI-04
  - TYPE-01
  - TYPE-02
  - TYPE-03
  - TYPE-04
  - TYPE-05
  - TYPE-06
  - TYPE-07
  - TYPE-08
must_haves:
  truths:
    - "Developer can run `mise test` and see tests execute"
    - "Developer can run `mise lint` and see linting execute"
    - "All 14 unit types are defined as Go constants (person, personExternal, system, systemExternal, db, dbExternal, queue, queueExternal, box, container, containerDb, containerQueue, component, componentDb, componentQueue)"
    - "Unit struct has all required fields including Technology, Width, Height"
    - "Link struct has all required fields including Technology, Description, LabelPosition"
    - "Properties struct has LineLength field"
  artifacts:
    - path: "go.mod"
      provides: "Go 1.26.1 module definition"
      contains: "go 1.26.1"
    - path: ".mise.toml"
      provides: "Development tasks"
      exports: ["test", "lint"]
    - path: "internal/model/unit.go"
      provides: "UnitType enum and Unit struct"
      min_lines: 50
    - path: "internal/model/link.go"
      provides: "Link struct and direction types"
      min_lines: 30
    - path: "internal/model/properties.go"
      provides: "Properties struct"
      min_lines: 20
    - path: "internal/model/colors.go"
      provides: "C4-PlantUML color constants"
      min_lines: 30
  key_links:
    - from: ".mise.toml"
      to: "golangci-lint binary"
      via: "mise tools"
      pattern: "golangci-lint"
    - from: ".mise.toml"
      to: "go test"
      via: "task run"
      pattern: "go test.*-cover"
---

<objective>
Set up development environment with mise tasks and define all domain types for the C4 model.

Purpose: Establish reproducible development workflow and type definitions that all downstream phases will use.
Output: Working mise tasks (test, lint) and complete domain model types.
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
<!-- Key decisions from CONTEXT.md that define the contracts -->

**Unit struct (flat, all fields top-level):**
```go
type Unit struct {
    Type         UnitType          // Discriminator
    Name         string
    Description  string
    Technology   string            // NOT for person types
    Color        string
    Style        string
    Border       string
    Edges        string            // Cascades from parent
    Width        float64           // 0 = auto
    Height       float64           // 0 = auto
    Expanded     []string
    Links        map[string]Link
    LinksFrom    map[string]Link
    Subunits     map[string]*Unit  // Recursive, needs pointer
}
```

**Link struct:**
```go
type LabelPosition string  // "tail", "head", "middle"

type Link struct {
    Target        string
    Arrow         ArrowDirection  // Forward, Reverse, Bidirectional, None
    Rank          RankDirection   // Forward, Reverse, Equal
    Color         string
    Style         string
    Technology    string          // C4: protocol/technology
    Description   string          // C4: relationship description
    LabelPosition LabelPosition   // Default: "middle"
}
```

**Properties struct:**
```go
type Properties struct {
    Name        string
    Description string
    Color       string
    Style       string
    Border      string
    Edges       string
    LineLength  int       // 0 = auto wrap
    Expanded    []string
}
```

**UnitType values by level:**
- C1: person, personExternal, system, systemExternal, db, dbExternal, queue, queueExternal, box
- C2: container, containerDb, containerQueue
- C3: component, componentDb, componentQueue
</interfaces>
</context>

<tasks>

<task type="auto">
  <name>Task 1: Update Go version and create mise configuration</name>
  <files>go.mod, .mise.toml</files>
  <action>
    1. Update go.mod: change `go 1.25.1` to `go 1.26.1`
    2. Add dependencies: `go get github.com/pelletier/go-toml/v2@latest` and `go get github.com/mitchellh/go-wordwrap@latest`
    3. Create .mise.toml with:
       - [tools] section: go = "1.26", golangci-lint = "2"
       - [tasks.test] description="Run Go tests with coverage", run="go test -cover ./..."
       - [tasks.lint] description="Run golangci-lint", run="golangci-lint run ./..."
       - [tasks.lint-fix] description="Run golangci-lint with auto-fix", run="golangci-lint run --fix ./..."

    The golangci-lint install is sandboxed by mise (not global) per DEVI-04.
  </action>
  <verify>
    <automated>mise tasks ls | grep -E "test|lint"</automated>
  </verify>
  <done>
    - go.mod contains "go 1.26.1"
    - go.sum exists with pelletier/go-toml/v2 and mitchellh/go-wordwrap
    - `mise tasks ls` shows test and lint tasks
    - `mise test` runs without error (no tests yet is OK)
    - `mise lint` runs without error (no Go files yet is OK)
  </done>
</task>

<task type="auto">
  <name>Task 2: Define UnitType enum and Unit struct</name>
  <files>internal/model/unit.go</files>
  <action>
    Create internal/model/unit.go with:

    1. UnitType type (string-based) with String() method
    2. All 14 type constants organized by level:
       - C1 Context: TypePerson, TypePersonExternal, TypeSystem, TypeSystemExternal, TypeDb, TypeDbExternal, TypeQueue, TypeQueueExternal, TypeBox
       - C2 Containers: TypeContainer, TypeContainerDb, TypeContainerQueue
       - C3 Components: TypeComponent, TypeComponentDb, TypeComponentQueue
    3. Unit struct with TOML tags for all fields:
       - Type UnitType `toml:"type"`
       - Name, Description, Technology (string, NOT for person types - validation in Phase 2)
       - Color, Style, Border, Edges (string)
       - Width, Height (float64, 0 = auto)
       - Expanded []string `toml:"expanded"`
       - Links, LinksFrom map[string]Link (use Link type - will be defined in next task)
       - Subunits map[string]*Unit (recursive, needs pointer)

    Use TOML struct tags matching the field names (lowercase). Add `toml:",inline"` for Subunits if needed for nested table unmarshaling.

    Note: Technology field is on Unit but should not be populated for person types (validation is Phase 2).
  </action>
  <verify>
    <automated>go build ./internal/model/...</automated>
  </verify>
  <done>
    - UnitType has all 14 constants
    - Unit struct has all required fields with correct types
    - TOML tags present on all fields
    - Code compiles without errors
  </done>
</task>

<task type="auto">
  <name>Task 3: Define Link struct and direction types</name>
  <files>internal/model/link.go</files>
  <action>
    Create internal/model/link.go with:

    1. ArrowDirection type (string) with constants:
       - ArrowForward (default, arrow at target)
       - ArrowReverse (arrow at source)
       - ArrowBidirectional (arrows at both ends)
       - ArrowNone (no arrow)

    2. RankDirection type (string) with constants:
       - RankForward (target ranks after source)
       - RankReverse (target ranks before source)
       - RankEqual (same rank)

    3. LabelPosition type (string) with constants:
       - LabelMiddle (default)
       - LabelTail
       - LabelHead

    4. Link struct with TOML tags:
       - Target string `toml:"-"` (set from map key, not TOML field)
       - Arrow ArrowDirection
       - Rank RankDirection
       - Color, Style string
       - Technology string (protocol/technology for the relationship)
       - Description string (relationship description)
       - LabelPosition LabelPosition

    Note: Target is populated from the map key in Links/LinksFrom maps, not from a TOML field. Consider using `toml:"-"` tag.
  </action>
  <verify>
    <automated>go build ./internal/model/...</automated>
  </verify>
  <done>
    - ArrowDirection has Forward, Reverse, Bidirectional, None constants
    - RankDirection has Forward, Reverse, Equal constants
    - LabelPosition has Middle, Tail, Head constants
    - Link struct has all required fields
    - Code compiles without errors
  </done>
</task>

<task type="auto">
  <name>Task 4: Define Properties struct</name>
  <files>internal/model/properties.go</files>
  <action>
    Create internal/model/properties.go with:

    Properties struct for root-level TOML [properties] section:
    - Name string `toml:"name"`
    - Description string `toml:"description"`
    - Color string `toml:"color"`
    - Style string `toml:"style"`
    - Border string `toml:"border"`
    - Edges string `toml:"edges"`
    - LineLength int `toml:"lineLength"` (0 = auto wrap)
    - Expanded []string `toml:"expanded"`
  </action>
  <verify>
    <automated>go build ./internal/model/...</automated>
  </verify>
  <done>
    - Properties struct has all required fields
    - LineLength field present (int type)
    - Code compiles without errors
  </done>
</task>

<task type="auto">
  <name>Task 5: Define C4-PlantUML color constants</name>
  <files>internal/model/colors.go</files>
  <action>
    Create internal/model/colors.go with color constants from C4-PlantUML:

    Base colors:
    - ElementFontColor = "#FFFFFF"
    - ArrowColor = "#666666"
    - BoundaryColor = "#444444"

    C1 Context level (background, border):
    - Person: "#08427B", "#073B6F"
    - PersonExternal: "#686868", "#8A8A8A"
    - System: "#1168BD", "#3C7FC0"
    - SystemExternal: "#999999", "#8A8A8A"

    C2 Container level:
    - Container: "#438DD5", "#3C7FC0"
    - ContainerExternal: "#B3B3B3", "#A6A6A6"

    C3 Component level:
    - Component: "#85BBF0", "#78A8D8"
    - ComponentExternal: "#CCCCCC", "#BFBFBF"

    Font colors:
    - C1/C2: "#FFFFFF" (white)
    - C3: "#000000" (black)

    Note: Db and Queue types use same colors as their parent level.
    Export all constants (capital first letter) so they're accessible from other packages.
  </action>
  <verify>
    <automated>go build ./internal/model/...</automated>
  </verify>
  <done>
    - All base colors defined
    - All type-specific colors defined with Background and Border variants
    - Constants are exported (capitalized)
    - Code compiles without errors
  </done>
</task>

<task type="auto">
  <name>Task 6: Run lint and verify model package</name>
  <files>internal/model/*.go</files>
  <action>
    Run `mise lint` and fix any lint errors in the model package.

    Common fixes that may be needed:
    - Add comments for exported types and constants
    - Fix any import issues

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
Phase 1 Plan 01 verification:

1. **Environment check:**
   ```bash
   mise tasks ls | grep -E "test|lint"
   # Should show test and lint tasks

   grep "go 1.26" go.mod
   # Should find go 1.26.1
   ```

2. **Model types check:**
   ```bash
   go build ./internal/model/...
   # Should compile without errors

   grep -c "Type.*UnitType" internal/model/unit.go
   # Should find 14+ type constants
   ```

3. **Lint check:**
   ```bash
   mise lint
   # Should pass with no errors
   ```
</verification>

<success_criteria>
- Developer can run `mise test` and `mise lint` to verify code quality
- All unit types (person, system, db, queue, box, external variants) are defined as Go types
- Link objects with target, arrow, rank, color, style, technology, description, labelPosition attributes are defined
- Properties struct includes LineLength field
- Unit struct includes Technology, Width, Height fields
</success_criteria>

<output>
After completion, create `.planning/phases/01-foundation-model/01-01-SUMMARY.md`
</output>
