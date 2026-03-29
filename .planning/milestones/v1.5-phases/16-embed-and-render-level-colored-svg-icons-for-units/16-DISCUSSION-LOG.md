# Phase 16: Embed and render level-colored SVG icons for units - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-03-20
**Phase:** 16-embed-and-render-level-colored-svg-icons-for-units
**Areas discussed:** Storage, Extraction, Coloring, HTML Labels, Size, Reference, Icon Set, Naming, Placement, Scope

---

## Icon Storage Location

| Option | Description | Selected |
|--------|-------------|----------|
| embed.FS in renderer | Embed SVGs into binary using Go's embed.FS. Single-binary distribution. | ✓ |
| Load from data/ at runtime | Keep SVGs in data/ directory and read at runtime. Requires data/ present. | |
| Hardcoded string constants | Store SVGs as string constants in a Go file. Compiled but no embed.FS. | |

**User's choice:** embed.FS in renderer (Recommended)
**Notes:** Single-binary distribution is preferred for portability.

---

## Icon Extraction Strategy

| Option | Description | Selected |
|--------|-------------|----------|
| On-demand per diagram | Only extract icons for types that appear in current diagram. | ✓ |
| Extract all icons at render start | Extract all icon variants up front regardless of usage. | |
| Inline base64 data URIs | Don't extract to files - embed SVGs inline as base64. | |

**User's choice:** On-demand per diagram (Recommended)
**Notes:** Keeps .icons/ clean, no unused files.

---

## Icon Color Mapping

| Option | Description | Selected |
|--------|-------------|----------|
| Dynamic currentColor replacement | At render time, replace currentColor with actual hex color for unit's level. | ✓ |
| Pre-generated color variants | Pre-generate 6 colored variants (db-c1.svg, db-c2.svg, etc.). | |
| Keep currentColor (no modification) | Embed SVGs with currentColor and let renderer handle inheritance. | |

**User's choice:** Dynamic currentColor replacement (Recommended)
**Notes:** One template SVG per icon type, flexible color application.

---

## Image Reference in HTML Labels

| Option | Description | Selected |
|--------|-------------|----------|
| <img> tags with SVG files | Use SVG <img> tags directly. GraphViz supports SVG images. | ✓ |
| <img> tags in HTML labels | Add icon column with <img src='.icons/db-c1.png' />. Requires PNG conversion. | |
| Inline base64 in <img> tags | Embed base64-encoded SVG directly in <img src='data:image/svg+xml;base64,...'>. | |

**User's choice:** <img> tags with SVG files (Recommended)
**Notes:** GraphViz supports SVG images directly, no PNG conversion needed.

---

## Icon Size

| Option | Description | Selected |
|--------|-------------|----------|
| 24x24 pixels | Compact icon that fits nicely without dominating label. | |
| 32x32 pixels | Medium size clearly visible but doesn't overwhelm text. | ✓ |
| 48x48 pixels | Large icon for maximum visibility, may require larger cells. | |

**User's choice:** 32x32 pixels (Recommended)

---

## Icon File Naming

| Option | Description | Selected |
|--------|-------------|----------|
| type-level.svg for standard, type-hexcolor.svg for custom | Standard: person-c1.svg, db-c2.svg. Custom: person-FF0000.svg. | ✓ |
| type-level.svg only | Include color in filename: person-c1.svg, db-c2.svg. | |
| Just type.svg | Simple names: person.svg, db.svg. Color applied dynamically. | |

**User's choice:** type-level.svg for standard, type-hexcolor.svg for custom
**Notes:** Standard level colors use type-cN.svg, custom colors use type-{hexcolor}.svg.

---

## Icon Placement in HTML Table

| Option | Description | Selected |
|--------|-------------|----------|
| Icon column with rowspan | Add new icon column at start, using rowspan to span all rows. | ✓ |
| Icon row above content | Add icon as separate row above name. More vertical space. | |
| Inline with name in same cell | Place icon inline with name: <td><img> <b>name</b></td>. | |

**User's choice:** Icon column with rowspan (Recommended)
**Notes:** Consistent with current SYS/CONT/COMP text label pattern.

---

## Icon Scope

| Option | Description | Selected |
|--------|-------------|----------|
| All 6 unit types get icons | Replace SYS/CONT/COMP text with SVG icons. All types: person, db, pipe, system, container, component. | ✓ |
| Only person/db/pipe get icons | Special types get icons. System/container/component keep text labels. | |
| Configurable per type | Let user opt-in via TOML config which types show icons. | |

**User's choice:** All 6 unit types get icons (Recommended)

---

## Claude's Discretion

- Exact icon column width and padding
- Fallback behavior if icon extraction fails
- Whether to cache extracted icons across multiple renders

## Deferred Ideas

None — discussion stayed within phase scope.
