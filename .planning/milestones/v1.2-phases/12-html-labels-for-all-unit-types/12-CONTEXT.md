# Phase 12: HTML Labels for All Unit Types - Context

**Gathered:** 2026-03-13
**Status:** Ready for planning

<domain>
## Phase Boundary

Change all unit type labels from record-style format to HTML table format with specific layouts per unit type. All units must still use `shape=record`.

**Current state:**
- Person: Uses `{icon}|{name|description}` record format
- Other units: Uses `{name|technology|description}` record format

**Target state:** All units use HTML tables embedded in record labels, with exact format per type.

</domain>

<decisions>
## Implementation Decisions

### Label Format by Unit Type

**Person (TypePerson, TypePersonExternal):**
- HTML table with icon on left (rowspan=2), name/description on right
- Exact format:
```html
<table>
  <tr align=center>
    <td rowspan=2 valign=middle><font size="+4">👤</font></td>
    <td valign=bottom><b>User name</b></td>
  </tr>
  <tr align=center>
    <td valign=top>Description</td>
  </tr>
</table>
```

**Database (TypeDb, TypeDbExternal, TypeContainerDb, TypeComponentDb):**
- HTML table with icon (rowspan=3), name bold, technology italic, description
- Exact format:
```html
<table>
  <tr align=center>
    <td rowspan=3 valign=middle><font size="+4">⛁</font></td>
    <td valign=bottom><b>DB name</b></td>
  </tr>
  <tr align=center>
    <td valign=middle><i>[technology]</i></td>
  </tr>
  <tr align=center>
    <td valign=top>Description</td>
  </tr>
</table>
```

**Queue (TypeQueue, TypeQueueExternal, TypeContainerQueue, TypeComponentQueue):**
- HTML table with 4 rows (no rowspan): graphics, name, technology, description
- Exact format:
```html
<table>
  <tr align=center>
    <td valign=middle>═╦╩═╦══</td>
  </tr>
  <tr align=center>
    <td valign=bottom><b>Queue name</b></td>
  </tr>
  <tr align=center>
    <td valign=middle><i>[technology]</i></td>
  </tr>
  <tr align=center>
    <td valign=top>Description</td>
  </tr>
</table>
```

**System (TypeSystem, TypeSystemExternal):**
- HTML table with `SYS` label (rowspan=3), name bold, technology italic, description
- Exact format:
```html
<table>
  <tr align=center>
    <td rowspan=3 valign=middle><tt>SYS</tt></td>
    <td valign=bottom><b>System name</b></td>
  </tr>
  <tr align=center>
    <td valign=middle><i>[technology]</i></td>
  </tr>
  <tr align=center>
    <td valign=top>Description</td>
  </tr>
</table>
```

**Container (TypeContainer):**
- HTML table with `CONT` label (rowspan=3), name bold, technology italic, description
- Exact format:
```html
<table>
  <tr align=center>
    <td rowspan=3 valign=middle><tt>CONT</tt></td>
    <td valign=bottom><b>Container name</b></td>
  </tr>
  <tr align=center>
    <td valign=middle><i>[technology]</i></td>
  </tr>
  <tr align=center>
    <td valign=top>Description</td>
  </tr>
</table>
```

**Component (TypeComponent):**
- HTML table with `COMP` label (rowspan=3), name bold, technology italic, description
- Exact format:
```html
<table>
  <tr align=center>
    <td rowspan=3 valign=middle><tt>COMP</tt></td>
    <td valign=bottom><b>Component name</b></td>
  </tr>
  <tr align=center>
    <td valign=middle><i>[technology]</i></td>
  </tr>
  <tr align=center>
    <td valign=top>Description</td>
  </tr>
</table>
```

**Box (TypeBox):**
- Same format as Container (use `CONT` label)
- Box is a grouping container, visually same as Container

### Shape Requirement
- All units must use `shape=record` (NOT shape=none)
- HTML tables are embedded inside record labels via `<...>` syntax

### Optional Fields
- If technology is empty, omit the technology row
- If description is empty, omit the description row

</decisions>

<specifics>
## Specific Ideas

User provided EXACT HTML format for each unit type:
- **Person**: Icon (👤) with rowspan=2, name bold, description below
- **DB**: Icon (⛁) with rowspan=3, name bold, [technology] italic, description
- **Queue**: Custom graphics (═╦╩═╦══) on row 1, name bold on row 2, [technology] italic on row 3, description on row 4 — NO rowspan
- **System**: `SYS` label (monospace) with rowspan=3, name bold, [technology] italic, description
- **Container**: `CONT` label (monospace) with rowspan=3, name bold, [technology] italic, description
- **Component**: `COMP` label (monospace) with rowspan=3, name bold, [technology] italic, description

</specifics>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/render/labels.go`: Contains `buildHTMLLabel()` (deprecated but has HTML table logic), `buildRecordLabel()`, `buildPersonRecordLabel()`
- `internal/graph/shapes.go`: `IconForType()` returns icons for each unit type
- `internal/render/converter.go`: `createNode()` applies labels via `cn.SetLabel()`

### Established Patterns
- Labels are built from `graph.Label` struct (Name, Technology, Description, Icon fields)
- Converter applies shape and label to nodes
- Person labels already have special handling

### Integration Points
- `internal/render/converter.go`: Needs to call new label builders based on unit type
- `internal/graph/builder.go`: Sets label fields when building nodes
- Need to add Type field to Label struct if not present, or pass UnitType to label builder

### Key Implementation Notes
- Current `IconForType()` returns empty string for System/Container/Component — will need new logic for `SYS`/`CONT`/`COMP` labels
- Queue currently uses `\u255F\n\u2562` as icon — need to change to `═╦╩═╦══` graphics on separate row
- External variants (TypePersonExternal, etc.) use same label format as internal types

</code_context>

<deferred>
## Deferred Ideas

None — all requirements specified with exact HTML templates.

</deferred>

---

*Phase: 12-html-labels-for-all-unit-types*
*Context gathered: 2026-03-13*
*Context updated: 2026-03-13*
