# Phase 12: HTML Labels for All Unit Types - Context

**Gathered:** 2026-03-13
**Status:** Ready for planning

<domain>
## Phase Boundary

Change all unit type labels from record-style format to HTML table format with specific layouts per unit type. All units must still use `shape=record`.

**Current state:**
- Person: Uses `{icon}|{name|description}` record format
- Other units: Uses `{name|technology|description}` record format

**Target state:** All units use HTML tables with exact format per type.

</domain>

<decisions>
## Implementation Decisions

### Label Format by Unit Type

**Person (TypePerson, TypePersonExternal):**
- HTML table with icon on left, name/description on right
- Format:
```html
<table>
  <tr align=center>
    <td rowspan=2 valign=middle><font size="+4">👤</font></td>
    <td valign=bottom><b>Name</b></td>
  </tr>
  <tr align=center>
    <td valign=top>Description</td>
  </tr>
</table>
```

**Database (TypeDb, TypeDbExternal):**
- HTML table with icon, name, technology, description
- Format:
```html
<table>
  <tr align=center>
    <td rowspan=3 valign=middle><font size="+4">⛁</font></td>
    <td valign=bottom><b>DB</b></td>
  </tr>
  <tr align=center>
    <td valign=middle><i>[technology]</i></td>
  </tr>
  <tr align=center>
    <td valign=top>Description</td>
  </tr>
</table>
```

**Queue (TypeQueue, TypeQueueExternal):**
- HTML table with icon, name, technology, description
- Format:
```html
<table>
  <tr align=center>
    <td valign=middle>═╦╩═╦══</td>
  </tr>
  <tr align=center>
    <td valign=bottom><b>Queue</b></td>
  </tr>
  <tr align=center>
    <td valign=middle><i>[technology]</i></td>
  </tr>
  <tr align=center>
    <td valign=top>Description</td>
  </tr>
</table>
```

**System, Container, Component (TypeSystem, TypeContainer, TypeComponent, plus External variants):**
- HTML table with name, technology, description (NO icon)
- Format:
```html
<table>
  <tr align=center>
    <td valign=bottom><b>Unit</b></td>
  </tr>
  <tr align=center>
    <td valign=middle><i>[technology]</i></td>
  </tr>
  <tr align=center>
    <td valign=top>Description</td>
  </tr>
</table>
```

### Shape Requirement
- All units must use `shape=record` (NOT shape=none)
- HTML tables are embedded inside record labels

### Optional Fields
- If technology is empty, omit the technology row
- If description is empty, omit the description row

</decisions>

<specifics>
## Specific Ideas

User provided EXACT HTML format for each unit type:
- Person: Icon (👤) with rowspan=2, name bold, description below
- DB: Icon (⛁) with rowspan=3, name bold, [technology] italic, description
- Queue: Custom graphics (═╦╩═╦══), name bold, [technology] italic, description  
- System/Container/Component: No icon, name bold, [technology] italic, description

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
- Need to add Type field to Label struct if not present

</code_context>

<deferred>
## Deferred Ideas

None — all requirements specified.

</deferred>

---

*Phase: 12-html-labels-for-all-unit-types*
*Context gathered: 2026-03-13*
