---
phase: 13
slug: refined-html-labels
created: 2026-03-13
status: draft
---

# Phase 13 — Context

> Refined HTML labels with shape=box style=rounded and updated table attributes

---

## Goal

Refine HTML labels from Phase 12 with:
1. All units: `shape=box, style=rounded` (instead of `shape=none`)
2. Table attributes: `border="0" cellpadding="0" cellspacing="0"`
3. Consistent label format per unit type

---

## Prior Decisions (Phase 12)

Phase 12 implemented HTML labels with:
- `shape=none` for HTML labels
- `graph.StrdupHTML()` for HTML label strings
- `<font face="monospace">` instead of `<tt>` (GraphViz doesn't support `<tt>`)
- Type-specific label builders: Person, DB, Queue, System, Container, Component

---

## New Requirements

### REFINED-01: Shape and Style

All units MUST render with:
- `shape=box`
- `style=rounded`

This replaces the current `shape=none` approach.

### REFINED-02: Table Attributes

All HTML tables MUST include:
- `border="0"`
- `cellpadding="0"`
- `cellspacing="0"`

---

## HTML Label Specifications

### Person Label

```html
<table border="0" cellpadding="0" cellspacing="0">
  <tr align=center>
    <td rowspan=2 valign=middle><font size="+4">👤</font></td>
    <td valign=bottom><b>User name</b></td>
  </tr>
  <tr align=center>
    <td valign=top>Description</td>
  </tr>
</table>
```

**Notes:**
- Person has NO technology field
- rowspan=2 (name + description)

### DB Label

```html
<table border="0" cellpadding="0" cellspacing="0">
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

**Notes:**
- Icon: ⛳ (U+26C1 golf flag, used as cylinder proxy)
- rowspan=3 (name + technology + description)

### Queue Label

```html
<table border="0" cellpadding="0" cellspacing="0">
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

**Notes:**
- Queue has NO rowspan - 4 separate rows
- Graphics: Unicode box drawing characters

### System Label

```html
<table border="0" cellpadding="0" cellspacing="0">
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

**Notes:**
- Use `<font face="monospace">SYS</font>` (not `<tt>`)
- rowspan=3 (name + technology + description)

### Container Label

```html
<table border="0" cellpadding="0" cellspacing="0">
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

**Notes:**
- Use `<font face="monospace">CONT</font>` (not `<tt>`)
- Also used for Box type
- rowspan=3 (name + technology + description)

### Component Label

```html
<table border="0" cellpadding="0" cellspacing="0">
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

**Notes:**
- Use `<font face="monospace">COMP</font>` (not `<tt>`)
- rowspan=3 (name + technology + description)

---

## Implementation Notes

1. **Shape change**: Change `ShapeNone` to `ShapeBox` with `style=rounded` in converter
2. **Table attributes**: Add `border="0" cellpadding="0" cellspacing="0"` to all `<table>` tags
3. **Monospace**: Continue using `<font face="monospace">` instead of `<tt>`
4. **Tests**: Update existing tests to match new output format

---

## Files to Modify

- `internal/render/labels.go` - Add table attributes to all HTML label builders
- `internal/render/converter.go` - Change shape from `none` to `box` with `style=rounded`
- `internal/render/labels_test.go` - Update expected outputs

---

## Out of Scope

- Changes to record-style labels (deprecated)
- Edge label changes
- Cluster/subgraph rendering changes
