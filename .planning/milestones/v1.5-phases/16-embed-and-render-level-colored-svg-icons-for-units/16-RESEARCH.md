# Phase 16: Embed and render level-colored SVG icons for units - Research

**Researched:** 2026-03-20
**Domain:** Go embed.FS for SVG assets, GraphViz HTML label IMG tags, dynamic SVG color manipulation
**Confidence:** HIGH

## Summary

Phase 16 replaces Unicode emoji icons with embedded SVG images that dynamically match each unit's C4 level border color. The implementation uses Go's `embed.FS` to bundle SVG templates with the executable, extracts colored variants on-demand during rendering to a `.icons/` directory, and references them via `<IMG>` tags in GraphViz HTML table labels.

**Primary recommendation:** Create a new `internal/render/icons` package with embedded SVGs. Use `strings.ReplaceAll()` to substitute `currentColor` with the unit's border color hex code. Extract only icons for types present in the current diagram to `{output}/.icons/`. Reference icons via relative paths in `<IMG SRC=".icons/person-3C7FC0.svg">` tags within HTML table cells.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01:** Use Go's `embed.FS` to embed SVG icons in the renderer package
- Icons travel with the executable for single-binary distribution
- Source SVGs moved from `data/` to `internal/render/icons/`

- **D-02:** Extract icons on-demand per diagram type
- Only extract icons for types that appear in the current diagram
- Extract to `{output-base}/.icons/` directory (e.g., `diagram/.icons/`)
- Keeps output clean, no unused icon files

- **D-03:** Dynamic `currentColor` replacement at render time
- Template SVGs use `currentColor` placeholder
- Replace `currentColor` with actual hex color based on unit's C4 level
- **Consistent naming:** Always `type-{hexcolor}.svg` format
  - Standard colors: `person-3C7FC0.svg`, `db-3C7FC0.svg`, `pipe-78A8D8.svg`
  - Custom colors: `person-FF0000.svg`, `db-00FF00.svg`
  - No special naming for standard vs custom - always use hex color

- **D-04:** Use `<img src='...'>` tags in HTML table labels
- Path is relative to rendered SVG file: `.icons/person-3C7FC0.svg`
- GraphViz supports `<IMG>` in HTML labels

- **D-05:** Icons rendered at 32x32 pixels

- **D-06:** Icon column with rowspan (same pattern as SYS/CONT/COMP text labels)
- Icon spans all content rows (name, technology, description)

- **D-07:** All 6 unit types display icons: person, db, pipe, system, container, component
- External variants use external border colors (gray tones)
- Box type uses container icon

### Claude's Discretion

- Exact icon column width and padding
- Fallback behavior if icon extraction fails
- Whether to cache extracted icons across multiple renders

### Deferred Ideas (OUT OF SCOPE)

None - discussion stayed within phase scope.

</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| ICON-01 | Embed SVG templates in renderer package | Go `//go:embed` directive with `embed.FS` |
| ICON-02 | Replace currentColor with C4 level hex codes | `strings.ReplaceAll()` for simple string substitution |
| ICON-03 | Extract colored icons to output directory | `os.WriteFile()` with directory creation |
| ICON-04 | Reference icons via IMG tags in HTML labels | GraphViz `<IMG SRC="...">` syntax |
| ICON-05 | Icon column with rowspan in HTML tables | Existing rowspan pattern in `labels.go` |
| ICON-06 | On-demand extraction only for present types | Type tracking during graph traversal |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| embed | stdlib (Go 1.16+) | Embed SVG files in binary | Native Go, zero dependencies, compile-time inclusion |
| strings | stdlib | Replace currentColor with hex | Simple string substitution, no regex needed |
| os | stdlib | File I/O for icon extraction | Standard file operations |
| path/filepath | stdlib | Path construction for icons | Cross-platform path handling |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| fmt | stdlib | Sprintf for icon filenames | Building `type-{hex}.svg` names |
| sync | stdlib | `sync.Once` for lazy loading | Optional: cache embedded SVG templates |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `strings.ReplaceAll` | `regexp.ReplaceAllString` | Regex overkill for simple currentColor replacement |
| On-demand extraction | Pre-extract all icons | Pre-extract creates unused files, wasteful |
| Relative path `.icons/` | Absolute path | Absolute paths break portability |
| `<IMG>` tag in HTML label | `image` attribute on node | `image` attribute doesn't work with HTML table layout |

**Installation:**
```bash
# No external dependencies required - uses Go standard library
```

**Version verification:** Go 1.26.1 (verified via `go version`)

## Architecture Patterns

### Recommended Project Structure
```
internal/render/
├── icons/                    # NEW: Embedded SVG icons
│   ├── embed.go              # //go:embed directive and accessor functions
│   ├── person.svg            # Moved from data/
│   ├── db.svg
│   ├── pipe.svg
│   ├── system.svg
│   ├── container.svg
│   └── component.svg
├── labels.go                 # MODIFY: Add icon column to HTML labels
├── converter.go              # MODIFY: Track types, trigger icon extraction
└── render.go                 # Unchanged
```

### Pattern 1: Embed.FS for SVG Assets
**What:** Use `//go:embed` directive to embed SVG files at compile time
**When to use:** Packaging static assets with the executable
**Example:**
```go
// Source: Go embed package documentation
package icons

import "embed"

//go:embed person.svg db.svg pipe.svg system.svg container.svg component.svg
var svgFiles embed.FS

// GetTemplate returns the raw SVG template content for the given icon type.
func GetTemplate(iconType string) (string, error) {
    filename := iconType + ".svg"
    data, err := svgFiles.ReadFile(filename)
    if err != nil {
        return "", fmt.Errorf("read icon template %s: %w", filename, err)
    }
    return string(data), nil
}
```

### Pattern 2: currentColor Replacement
**What:** Simple string replacement to inject C4 level colors
**When to use:** Converting template SVGs to colored variants
**Example:**
```go
// Source: Go strings package documentation
func ColorizeIcon(svgTemplate, hexColor string) string {
    // Remove # from hex color for filename (e.g., "#3C7FC0" -> "3C7FC0")
    hexWithoutHash := strings.TrimPrefix(hexColor, "#")

    // Replace all currentColor occurrences with the hex color
    return strings.ReplaceAll(svgTemplate, "currentColor", "#"+hexWithoutHash)
}
```

### Pattern 3: Icon Extraction to Filesystem
**What:** Write colored SVG to `.icons/` directory during rendering
**When to use:** Before creating nodes that need icons
**Example:**
```go
// Source: Standard Go file I/O patterns
func ExtractIcon(outputDir, iconType, hexColor string) (string, error) {
    // Build path: {output}/.icons/{type}-{hex}.svg
    hexWithoutHash := strings.TrimPrefix(hexColor, "#")
    iconFilename := fmt.Sprintf("%s-%s.svg", iconType, hexWithoutHash)
    iconDir := filepath.Join(outputDir, ".icons")
    iconPath := filepath.Join(iconDir, iconFilename)

    // Create .icons directory if needed
    if err := os.MkdirAll(iconDir, 0755); err != nil {
        return "", fmt.Errorf("create icons directory: %w", err)
    }

    // Check if already exists (skip extraction for repeated types)
    if _, err := os.Stat(iconPath); err == nil {
        return iconPath, nil // Already extracted
    }

    // Get template and colorize
    template, err := icons.GetTemplate(iconType)
    if err != nil {
        return "", err
    }
    coloredSVG := strings.ReplaceAll(template, "currentColor", hexColor)

    // Write to file
    if err := os.WriteFile(iconPath, []byte(coloredSVG), 0644); err != nil {
        return "", fmt.Errorf("write icon file: %w", err)
    }

    return iconPath, nil
}
```

### Pattern 4: IMG Tag in HTML Label
**What:** Use GraphViz `<IMG>` tag within HTML table cell
**When to use:** Building HTML labels for nodes with icons
**Example:**
```go
// Source: GraphViz documentation - https://graphviz.org/doc/info/shapes.html
func buildPersonHTMLLabelWithIcon(label *graph.Label, iconPath string) string {
    var sb strings.Builder
    sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0">`)

    // Calculate rowspan for icon
    rowspan := 1
    if label.Description != "" {
        rowspan = 2
    }

    // Row 1: Icon (rowspan) + Name
    sb.WriteString(`<tr align="center">`)
    sb.WriteString(`<td rowspan="`)
    sb.WriteString(strconv.Itoa(rowspan))
    sb.WriteString(`" valign="middle">`)
    // IMG tag with relative path
    sb.WriteString(`<img src="`)
    sb.WriteString(iconPath) // e.g., ".icons/person-3C7FC0.svg"
    sb.WriteString(`" width="32" height="32"/>`)
    sb.WriteString(`</td>`)
    sb.WriteString(`<td valign="bottom"><b>`)
    sb.WriteString(html.EscapeString(label.Name))
    sb.WriteString(`</b></td>`)
    sb.WriteString(`</tr>`)

    // ... rest of label rows

    sb.WriteString(`</table>`)
    return sb.String()
}
```

### Pattern 5: Type-to-Icon Mapping
**What:** Map unit types to icon names, accounting for variants
**When to use:** Determining which icon to use for a unit
**Example:**
```go
// IconTypeForUnit returns the icon type name for a unit type.
// External variants use the same icon as internal variants.
// Box type uses container icon.
func IconTypeForUnit(t model.UnitType) string {
    switch {
    case graph.IsPersonType(t):
        return "person"
    case graph.IsDbType(t):
        return "db"
    case graph.IsQueueType(t):
        return "pipe"
    case graph.IsSystemType(t):
        return "system"
    case graph.IsContainerType(t), t == model.TypeBox:
        return "container"
    case graph.IsComponentType(t):
        return "component"
    default:
        return "container" // Fallback
    }
}
```

### Anti-Patterns to Avoid
- **Absolute paths in IMG SRC:** Breaks when diagrams moved to different locations
- **Pre-extracting all 6 icons × N colors:** Creates unused files, wasteful
- **Embedding colored SVGs:** Would need 6 icons × 6+ colors = 36+ files
- **Using emoji in IMG tags:** Emoji are text, not images - won't work
- **Forgetting rowspan:** Icon would only span first row, look broken

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| File embedding | Custom base64 encoding | `embed.FS` | Native Go, compile-time, type-safe |
| SVG color replacement | XML parsing | `strings.ReplaceAll` | Simple text substitution sufficient |
| Path construction | String concatenation | `filepath.Join()` | Cross-platform compatibility |
| Icon caching | Custom cache system | Let OS handle file cache | OS file cache is efficient |

**Key insight:** The `currentColor` approach in the existing SVG templates is ideal - it's a single, predictable string that can be replaced with simple string operations. No XML parsing needed.

## Common Pitfalls

### Pitfall 1: Relative Path Resolution
**What goes wrong:** IMG SRC paths don't resolve correctly when SVG is opened from different directories
**Why it happens:** GraphViz renders the IMG tag into the output SVG, but the browser/viewer resolves the path relative to the SVG file location
**How to avoid:** Always use paths relative to the output SVG file location: `.icons/person-3C7FC0.svg`
**Warning signs:** Broken image icons in rendered SVGs

### Pitfall 2: Hash in Filename
**What goes wrong:** Creating files with `#` in the name (e.g., `person-#3C7FC0.svg`) fails on some filesystems or URL encodes incorrectly
**Why it happens:** `#` has special meaning in URLs and is problematic in filenames
**How to avoid:** Strip the `#` from hex colors before using in filenames: `person-3C7FC0.svg`
**Warning signs:** File creation errors, broken image links

### Pitfall 3: Missing .icons Directory
**What goes wrong:** Icon extraction fails because parent directory doesn't exist
**Why it happens:** `os.WriteFile` doesn't create parent directories
**How to avoid:** Always call `os.MkdirAll(filepath.Dir(iconPath), 0755)` before writing
**Warning signs:** "no such file or directory" errors during icon extraction

### Pitfall 4: Icon Extraction Race Condition
**What goes wrong:** Multiple units with same type/color try to extract the same icon simultaneously
**Why it happens:** Parallel node creation without coordination
**How to avoid:** Check if file exists before extracting, or use file existence as natural lock
**Warning signs:** Intermittent file write errors

### Pitfall 5: IMG Tag Not Supported in All GraphViz Outputs
**What goes wrong:** Images appear in SVG output but not PNG/JPG
**Why it happens:** GraphViz IMG support varies by output format
**How to avoid:** This phase targets SVG output only (per scope); document limitation
**Warning signs:** Missing icons in non-SVG formats

## Code Examples

Verified patterns from official sources:

### Complete Icons Package
```go
// Source: Go embed package documentation
package icons

import (
    "embed"
    "fmt"
)

//go:embed person.svg db.svg pipe.svg system.svg container.svg component.svg
var svgFiles embed.FS

// Icon types supported by the renderer.
const (
    TypePerson    = "person"
    TypeDb        = "db"
    TypePipe      = "pipe"
    TypeSystem    = "system"
    TypeContainer = "container"
    TypeComponent = "component"
)

// GetTemplate returns the raw SVG template for the given icon type.
// The template uses "currentColor" as a placeholder for the stroke color.
func GetTemplate(iconType string) (string, error) {
    filename := iconType + ".svg"
    data, err := svgFiles.ReadFile(filename)
    if err != nil {
        return "", fmt.Errorf("icon template %s not found: %w", filename, err)
    }
    return string(data), nil
}

// Colorize replaces "currentColor" in the SVG with the provided hex color.
func Colorize(svgTemplate, hexColor string) string {
    return strings.ReplaceAll(svgTemplate, "currentColor", hexColor)
}
```

### Icon Extractor for Renderer
```go
// Source: Standard Go patterns + project conventions
package render

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "github.com/Djarvur/c4drill/internal/render/icons"
)

// IconExtractor manages extraction of colored icons to the output directory.
type IconExtractor struct {
    outputDir string
    extracted map[string]bool // Track which icons have been extracted
}

// NewIconExtractor creates an extractor for the given output directory.
func NewIconExtractor(outputDir string) *IconExtractor {
    return &IconExtractor{
        outputDir: outputDir,
        extracted: make(map[string]bool),
    }
}

// Extract extracts a colored icon if not already extracted.
// Returns the relative path to use in IMG SRC (e.g., ".icons/person-3C7FC0.svg").
func (e *IconExtractor) Extract(iconType, hexColor string) (string, error) {
    // Build filename: type-{hex}.svg (without # in hex)
    hexClean := strings.TrimPrefix(hexColor, "#")
    iconFilename := fmt.Sprintf("%s-%s.svg", iconType, hexClean)

    // Check cache
    if e.extracted[iconFilename] {
        return filepath.Join(".icons", iconFilename), nil
    }

    // Build full path
    iconDir := filepath.Join(e.outputDir, ".icons")
    iconPath := filepath.Join(iconDir, iconFilename)

    // Create directory
    if err := os.MkdirAll(iconDir, 0755); err != nil {
        return "", fmt.Errorf("create icons directory: %w", err)
    }

    // Check if file already exists (from previous run)
    if _, err := os.Stat(iconPath); err == nil {
        e.extracted[iconFilename] = true
        return filepath.Join(".icons", iconFilename), nil
    }

    // Get template
    template, err := icons.GetTemplate(iconType)
    if err != nil {
        return "", err
    }

    // Colorize
    coloredSVG := icons.Colorize(template, hexColor)

    // Write file
    if err := os.WriteFile(iconPath, []byte(coloredSVG), 0644); err != nil {
        return "", fmt.Errorf("write icon %s: %w", iconFilename, err)
    }

    e.extracted[iconFilename] = true
    return filepath.Join(".icons", iconFilename), nil
}
```

### Modified HTML Label Builder
```go
// Source: Existing labels.go pattern + GraphViz IMG syntax
func buildPersonHTMLLabelWithIcon(label *graph.Label, iconRelPath string) string {
    if label == nil {
        return ""
    }

    var sb strings.Builder
    sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0">`)

    rowspan := 1
    if label.Description != "" {
        rowspan = 2
    }

    // Icon column (rowspan)
    sb.WriteString(`<tr align="center">`)
    sb.WriteString(`<td rowspan="`)
    sb.WriteString(strconv.Itoa(rowspan))
    sb.WriteString(`" valign="middle" width="40">`)
    sb.WriteString(`<img src="`)
    sb.WriteString(iconRelPath)
    sb.WriteString(`" width="32" height="32"/>`)
    sb.WriteString(`</td>`)

    // Name column
    sb.WriteString(`<td valign="bottom" align="left"><b>`)
    sb.WriteString(html.EscapeString(label.Name))
    sb.WriteString(`</b></td>`)
    sb.WriteString(`</tr>`)

    if label.Description != "" {
        sb.WriteString(`<tr align="center">`)
        sb.WriteString(`<td valign="top" align="left">`)
        sb.WriteString(html.EscapeString(label.Description))
        sb.WriteString(`</td>`)
        sb.WriteString(`</tr>`)
    }

    sb.WriteString(`</table>`)
    return sb.String()
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Unicode emoji icons | SVG image icons | Phase 16 | Consistent cross-platform rendering, color-matched to C4 levels |
| Static icon color | Dynamic `currentColor` replacement | Phase 16 | Icons match unit's border color automatically |
| Embedded emoji text | Extracted SVG files | Phase 16 | Requires `.icons/` directory, but enables color customization |

**Deprecated/outdated:**
- Unicode emoji icons: Inconsistent rendering across platforms, no color control
- `IconForType()` returning emoji strings: Replaced by icon extraction system

## Open Questions

1. **Icon Column Width**
   - What we know: Icons are 32x32 pixels
   - What's unclear: Optimal cell width/padding for visual balance
   - Recommendation: Start with 40px cell width (32px icon + 8px padding), adjust based on visual testing

2. **Fallback Behavior**
   - What we know: Icon extraction could fail (disk full, permissions, etc.)
   - What's unclear: Should node render without icon or show placeholder?
   - Recommendation: Log warning, render node without icon column (graceful degradation)

3. **Icon Cache Across Renders**
   - What we know: `.icons/` directory persists between renders
   - What's unclear: Should we skip extraction for existing files or always overwrite?
   - Recommendation: Check file existence, skip if present (faster re-renders)

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + stretchr/testify v1.11.1 |
| Config file | none - standard go test |
| Quick run command | `go test ./internal/render/... -v -short` |
| Full suite command | `go test ./... -cover -race` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| ICON-01 | Embed SVG templates | unit | `go test ./internal/render/icons/... -run TestGetTemplate -v` | Wave 0 |
| ICON-02 | currentColor replacement | unit | `go test ./internal/render/... -run TestColorize -v` | Wave 0 |
| ICON-03 | Icon extraction to filesystem | unit | `go test ./internal/render/... -run TestExtractIcon -v` | Wave 0 |
| ICON-04 | IMG tag in HTML labels | unit | `go test ./internal/render/... -run TestHTMLLabelWithIcon -v` | Wave 0 |
| ICON-05 | Icon rowspan in tables | unit | `go test ./internal/render/... -run TestIconRowspan -v` | Wave 0 |
| ICON-06 | On-demand extraction only | integration | `go test ./internal/render/... -run TestOnDemandExtraction -v` | Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/render/... -v -short`
- **Per wave merge:** `go test ./... -cover -race`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/render/icons/embed.go` - Embed directive and accessor functions
- [ ] `internal/render/icons/*.svg` - Move SVG files from `data/` to `internal/render/icons/`
- [ ] `internal/render/icons/icons_test.go` - Tests for template loading and colorization
- [ ] `internal/render/icon_extractor.go` - IconExtractor type with Extract method
- [ ] `internal/render/icon_extractor_test.go` - Tests for icon extraction
- [ ] Update `internal/render/labels.go` - Add icon column to HTML label builders
- [ ] Update `internal/render/converter.go` - Integrate icon extraction with node creation

## Sources

### Primary (HIGH confidence)
- [Go embed package documentation](https://pkg.go.dev/embed) - `//go:embed` directive and `embed.FS` type
- [GraphViz HTML-Like Labels](https://graphviz.org/doc/info/shapes.html#html) - `<IMG>` tag syntax and attributes
- [goccy/go-graphviz GitHub](https://github.com/goccy/go-graphviz) - Library usage patterns

### Secondary (MEDIUM confidence)
- Project source files: `internal/model/colors.go` - C4 level color constants
- Project source files: `internal/graph/shapes.go` - Existing `IconForType()` and type helpers
- Project source files: `internal/render/labels.go` - Existing HTML label builders with rowspan pattern

### Tertiary (LOW confidence)
- N/A - All critical information from primary/secondary sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Go embed.FS is well-documented stdlib feature since Go 1.16
- Architecture: HIGH - Clear patterns from GraphViz documentation and existing codebase
- Pitfalls: HIGH - Common file I/O and path issues, well-understood solutions

**Research date:** 2026-03-20
**Valid until:** 30 days - Go embed API stable, GraphViz IMG support unchanged
