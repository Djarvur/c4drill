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
    - "Developer can run mise test to execute tests"
    - "Developer can run mise lint to check code quality"
    - "All C4 unit types are defined as Go constants"
    - "Link struct defines all required attributes"
  artifacts:
    - path: ".mise.toml"
      provides: "Development tasks and tool installation"
      contains: "[tasks.test]"
    - path: "internal/model/unit.go"
      provides: "UnitType enum and Unit struct"
      exports: ["UnitType", "Unit", "TypePerson", "TypeSystem"]
    - path: "internal/model/link.go"
      provides: "Link struct and direction types"
      exports: ["Link", "ArrowDirection", "RankDirection"]
    - path: "internal/model/properties.go"
      provides: "Properties struct for root-level config"
      exports: ["Properties"]
    - path: "internal/model/colors.go"
      provides: "C4-PlantUML color constants"
      exports: ["ColorPerson", "ColorSystem", "ColorArrow"]
  key_links:
    - from: ".mise.toml"
      to: "go test"
      via: "task definition"
      pattern: "go test.*-cover"
---

<objective>
Set up the development environment with mise tasks and define all domain model types for the C4 architecture.

Purpose: Establish the foundation for the entire project - development tooling and type definitions that all subsequent phases depend on.
Output: Working mise tasks (test, lint) and complete model package with all C4 unit types, link types, and color constants.
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
<!-- Key patterns from research - executor should use these directly -->

From 01-RESEARCH.md - UnitType discriminator pattern:
```go
type UnitType string

const (
    // C1 Context level
    TypePerson         UnitType = "person"
    TypePersonExternal UnitType = "personExternal"
    TypeSystem         UnitType = "system"
    TypeSystemExternal UnitType = "systemExternal"
    TypeDb             UnitType = "db"
    TypeDbExternal     UnitType = "dbExternal"
    TypeQueue          UnitType = "queue"
    TypeQueueExternal  UnitType = "queueExternal"
    TypeBox            UnitType = "box"

    // C2 Container level
    TypeContainer      UnitType = "container"
    TypeContainerDb    UnitType = "containerDb"
    TypeContainerQueue UnitType = "containerQueue"

    // C3 Component level
    TypeComponent      UnitType = "component"
    TypeComponentDb    UnitType = "componentDb"
    TypeComponentQueue UnitType = "componentQueue"
)
```

From 01-RESEARCH.md - Link struct pattern:
```go
type Link struct {
    Target string
    Arrow  ArrowDirection  // Forward, Reverse, Bidirectional, None
    Rank   RankDirection   // Forward, Reverse, Equal
    Color  string
    Style  string
}
```

From 01-RESEARCH.md - Mise task configuration:
```toml
[tools]
go = "1.26"
golangci-lint = "2"

[tasks.test]
description = "Run Go tests with coverage"
run = "go test -cover ./..."

[tasks.lint]
description = "Run golangci-lint"
run = "golangci-lint run ./..."
```

From 01-RESEARCH.md - C4-PlantUML colors:
- ELEMENT_FONT_COLOR: #FFFFFF
- ARROW_COLOR: #666666
- BOUNDARY_COLOR: #444444
- Person background: #08427B, border: #073B6F
- External Person background: #686868, border: #8A8A8A
- System background: #1168BD, border: #3C7FC0
- External System background: #999999, border: #8A8A8A
- Container background: #438DD5, border: #3C7FC0
- Component background: #85BBF0, border: #78A8D8
</interfaces>
</context>

<tasks>

<task type="auto">
  <name>Task 1: Update Go version and create mise configuration</name>
  <files>go.mod, .mise.toml</files>
  <action>
    1. Update go.mod: change `go 1.25.1` to `go 1.26.1`
    2. Create .mise.toml with:
       - [tools] section: go = "1.26", golangci-lint = "2"
       - [tasks.test] with description and "go test -cover ./..."
       - [tasks.lint] with description and "golangci-lint run ./..."
       - [tasks.lint-fix] with description and "golangci-lint run --fix ./..."

    This installs golangci-lint v2 into mise sandbox (not globally), satisfying DEVI-04.
    The tasks provide mise test and mise lint commands (DEVI-02, DEVI-03).
  </action>
  <verify>
    <automated>grep "go 1.26.1" go.mod && grep "\[tasks.test\]" .mise.toml && grep "\[tasks.lint\]" .mise.toml</automated>
  </verify>
  <done>go.mod shows 1.26.1, .mise.toml exists with test and lint tasks, golangci-lint v2 configured in sandbox</done>
</task>

<task type="auto">
  <name>Task 2: Create domain model types</name>
  <files>internal/model/unit.go, internal/model/link.go, internal/model/properties.go, internal/model/colors.go</files>
  <action>
    Create internal/model/ directory and four files:

    1. unit.go:
       - Define UnitType as `type UnitType string`
       - Define all 14 type constants (C1: person, personExternal, system, systemExternal, db, dbExternal, queue, queueExternal, box; C2: container, containerDb, containerQueue; C3: component, componentDb, componentQueue)
       - Add `func (t UnitType) String() string { return string(t) }`
       - Define Unit struct with TOML tags matching schema:
         - Type UnitType `toml:"type"`
         - Name, Description string
         - Color, Style, Border, Edges string
         - Expanded []string
         - Links, LinksFrom map[string]Link (use `toml:"link"` and `toml:"linkFrom"`)
         - Subunits map[string]*Unit (use `toml:",inline"` for arbitrary nested tables)
       - FLAT struct - all fields at top level per CONTEXT.md decision

    2. link.go:
       - Define ArrowDirection as `type ArrowDirection string`
       - Constants: ArrowForward, ArrowReverse, ArrowBidirectional, ArrowNone
       - Define RankDirection as `type RankDirection string`
       - Constants: RankForward, RankReverse, RankEqual
       - Define Link struct with Target, Arrow, Rank, Color, Style fields

    3. properties.go:
       - Define Properties struct with Name, Description, Color, Style, Border, Edges, Expanded fields
       - Matches [properties] section in TOML schema

    4. colors.go:
       - Define exported constants for C4-PlantUML colors:
         - ColorFont = "#FFFFFF"
         - ColorArrow = "#666666"
         - ColorBoundary = "#444444"
         - ColorPersonBg = "#08427B", ColorPersonBorder = "#073B6F"
         - ColorPersonExtBg = "#686868", ColorPersonExtBorder = "#8A8A8A"
         - ColorSystemBg = "#1168BD", ColorSystemBorder = "#3C7FC0"
         - ColorSystemExtBg = "#999999", ColorSystemExtBorder = "#8A8A8A"
         - ColorContainerBg = "#438DD5", ColorContainerBorder = "#3C7FC0"
         - ColorComponentBg = "#85BBF0", ColorComponentBorder = "#78A8D8"

    DO NOT use interface-based polymorphism - use discriminator pattern per CONTEXT.md.
    DO NOT create nested style objects - flat struct per CONTEXT.md.
  </action>
  <verify>
    <automated>go build ./internal/model/...</automated>
  </verify>
  <done>All 14 UnitType constants defined, Unit struct with TOML tags compiles, Link struct with Arrow/Rank directions compiles, Properties struct compiles, color constants exported</done>
</task>

</tasks>

<verification>
After completing all tasks:
1. Run `mise install` to verify tools install correctly
2. Run `go build ./...` to verify all code compiles
3. Run `mise lint` to verify no lint errors in new code
</verification>

<success_criteria>
- [x] go.mod updated to Go 1.26.1
- [x] .mise.toml exists with test and lint tasks
- [x] golangci-lint v2 configured for sandbox install
- [x] UnitType enum defines all 14 C4 types (C1, C2, C3 levels)
- [x] Unit struct has flat design with TOML tags
- [x] Link struct has Arrow and Rank direction fields
- [x] Properties struct matches TOML schema
- [x] Color constants match C4-PlantUML values
- [x] All code compiles without errors
- [x] No lint errors in new code
</success_criteria>

<output>
After completion, create `.planning/phases/01-foundation-model/01-SUMMARY.md`
</output>
