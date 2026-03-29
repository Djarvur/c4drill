# Phase 17: Word-Wrapped Labels for Credit Card Proportions - Research

**Researched:** 2026-03-23
**Domain:** Go text wrapping, GraphViz HTML labels, CLI configuration
**Confidence:** HIGH

## Summary

This phase requires implementing word-wrapping for HTML label cells to achieve credit card proportions (1.6:1 width:height ratio). The core challenge is that GraphViz controls final rendering, so we cannot directly set pixel dimensions. Instead, we must constrain text column width via GraphViz's HTML table `WIDTH` attribute (in points) and use word-wrapping to break long text into multiple lines.

**Primary recommendation:** Implement a custom word-wrapping function with character fallback (no external dependency needed for this simple case), and use GraphViz HTML table `WIDTH` attribute on text cells to constrain width. Calculate target width dynamically based on row count and configurable ratio.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **D-01:** Default ratio is **8/5 = 1.6:1** (width:height)
- **D-02:** Dynamic width calculation based on content height
- **D-03:** Hybrid wrapping: word-based with forced character break for long words
- **D-04:** All label fields wrapped: name, technology, description
- **D-05:** Ratio configurable via CLI flag `--label-ratio` and env var `C4DRILL_LABEL_RATIO`

### Claude's Discretion
- Exact pixel calculations for height estimation
- Maximum line length before forced character break
- Whether to expose max line length as configurable parameter

### Deferred Ideas (OUT OF SCOPE)
None - discussion stayed within phase scope.
</user_constraints>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| stdlib `strings` | Go 1.26 | Word splitting, string manipulation | No dependency, sufficient for this use case |
| stdlib `os` | Go 1.26 | Environment variable reading | Standard Go pattern |
| spf13/cobra | v1.10.2 | CLI flag handling | Already in project, established pattern |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| github.com/muesli/reflow/wordwrap | v0.3.0 | ANSI-aware word wrapping | If complex wrapping needed (optional) |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Custom word wrap | muesli/reflow | reflow is overkill for simple word wrap; custom is ~20 lines |
| GraphViz WIDTH | CSS styling | GraphViz HTML doesn't support CSS; use WIDTH attribute |

**Installation:**
```bash
# No new dependencies needed for core implementation
# Optional: go get github.com/muesli/reflow@v0.3.0
```

**Version verification:**
```
github.com/muesli/reflow v0.3.0 (latest as of 2026-03-23)
```

## Architecture Patterns

### Recommended Implementation Structure
```
internal/render/
├── labels.go          # Existing label builders - add wrapText() calls
├── wrap.go            # NEW: word-wrap implementation
└── wrap_test.go       # NEW: unit tests for wrapping

cmd/c4drill/
└── root.go            # Add --label-ratio flag, env var handling
```

### Pattern 1: Word Wrapping with Character Fallback
**What:** Split text into lines that fit within a maximum character width. Break at word boundaries first, force character break for long words.
**When to use:** All label text fields (name, technology, description)
**Example:**
```go
// wrapText wraps text to maxChars per line, breaking at word boundaries.
// If a single word exceeds maxChars, it is forcibly broken.
func wrapText(text string, maxChars int) string {
    words := strings.Fields(text)
    var lines []string
    var currentLine strings.Builder

    for _, word := range words {
        // If word alone exceeds max, force break it
        if len(word) > maxChars {
            if currentLine.Len() > 0 {
                lines = append(lines, currentLine.String())
                currentLine.Reset()
            }
            // Break word into chunks of maxChars
            for i := 0; i < len(word); i += maxChars {
                end := i + maxChars
                if end > len(word) {
                    end = len(word)
                }
                lines = append(lines, word[i:end])
            }
            continue
        }

        // Check if adding word would exceed limit
        if currentLine.Len() > 0 && currentLine.Len()+1+len(word) > maxChars {
            lines = append(lines, currentLine.String())
            currentLine.Reset()
        }

        if currentLine.Len() > 0 {
            currentLine.WriteString(" ")
        }
        currentLine.WriteString(word)
    }

    if currentLine.Len() > 0 {
        lines = append(lines, currentLine.String())
    }

    return strings.Join(lines, "<BR/>")
}
```

### Pattern 2: Dynamic Width Calculation
**What:** Calculate target text column width based on number of content rows and desired ratio.
**When to use:** Before generating HTML label
**Example:**
```go
const (
    iconColumnWidth  = 36  // Fixed icon column width in points
    pointsPerRow     = 18  // Approximate row height in points (font + padding)
    defaultRatio     = 1.6 // 8/5 credit card ratio
)

// calculateTextWidth estimates text column width to achieve target ratio.
// totalHeight = rowCount * pointsPerRow
// totalWidth = totalHeight * ratio
// textWidth = totalWidth - iconColumnWidth
func calculateTextWidth(rowCount int, ratio float64) int {
    totalHeight := float64(rowCount * pointsPerRow)
    totalWidth := totalHeight * ratio
    textWidth := int(totalWidth) - iconColumnWidth
    if textWidth < 50 {
        textWidth = 50 // Minimum readable width
    }
    return textWidth
}
```

### Pattern 3: CLI Flag with Environment Variable Fallback
**What:** Allow ratio configuration via CLI flag (takes precedence) or environment variable.
**When to use:** In root.go for flag registration
**Example:**
```go
var labelRatio float64

func init() {
    rootCmd.PersistentFlags().Float64Var(&labelRatio, "label-ratio", 0,
        "Width:height ratio for unit labels (default: 1.6)")

    // Set default from environment if flag not provided
    if labelRatio == 0 {
        if envRatio := os.Getenv("C4DRILL_LABEL_RATIO"); envRatio != "" {
            if parsed, err := strconv.ParseFloat(envRatio, 64); err == nil {
                labelRatio = parsed
            }
        }
    }
    if labelRatio == 0 {
        labelRatio = 1.6 // Default credit card ratio
    }
}
```

### Anti-Patterns to Avoid
- **Using `<br>` instead of `<BR/>`:** GraphViz HTML requires uppercase `<BR/>` for line breaks
- **Setting WIDTH on TABLE:** Set WIDTH on individual `<TD>` cells for precise control
- **Forgetting to escape HTML:** Always use `html.EscapeString()` before wrapping to avoid breaking HTML entities

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Complex ANSI-aware wrapping | Custom ANSI parser | muesli/reflow/wordwrap | Handles terminal color codes correctly |
| Font metrics calculation | Custom measurement | GraphViz's WIDTH attribute | GraphViz handles font rendering; we just constrain |

**Key insight:** GraphViz renders the final output, so we cannot control exact pixel dimensions. We can only influence the layout via WIDTH constraints and line breaks. The ratio will be approximate, not exact.

## Common Pitfalls

### Pitfall 1: GraphViz WIDTH is Minimum, Not Maximum
**What goes wrong:** Setting `WIDTH="100"` doesn't prevent content from expanding beyond 100 points.
**Why it happens:** GraphViz WIDTH is a minimum width constraint unless `FIXEDSIZE="TRUE"` is set.
**How to avoid:** Word-wrap text to fit within the target width; don't rely on WIDTH alone. The wrapping ensures content fits.
**Warning signs:** Labels still appear stretched despite WIDTH attribute.

### Pitfall 2: Unicode Character Count vs Byte Count
**What goes wrong:** Using `len(word)` counts bytes, not characters, causing incorrect breaks for non-ASCII text.
**Why it happens:** Go strings are UTF-8 encoded; `len()` returns byte count.
**How to avoid:** Use `utf8.RuneCountInString(word)` for character count, or convert to `[]rune` for indexing.
**Warning signs:** Multi-byte characters (emoji, CJK) break at wrong positions.

### Pitfall 3: HTML Escaping Before Wrapping
**What goes wrong:** Escaping text before wrapping breaks word boundary detection.
**Why it happens:** `html.EscapeString("a & b")` becomes `a &amp; b`, adding characters.
**How to avoid:** Wrap first, then escape. Or account for escaped length in calculations.
**Warning signs:** Ampersands, quotes, or angle brackets cause wrapping errors.

### Pitfall 4: Forcing Ratio Creates Tiny Labels
**What goes wrong:** Labels with single-word names become too narrow to be readable.
**Why it happens:** Mathematically correct ratio may result in impractically small widths.
**How to avoid:** Set minimum width threshold (e.g., 50 points) regardless of ratio calculation.
**Warning signs:** Short labels appear cramped or text overflows.

## Code Examples

### Word Wrap Implementation (Verified Pattern)
```go
// Source: Custom implementation based on standard algorithm
package render

import (
    "strings"
    "unicode/utf8"
)

// wrapText wraps text to maxChars per line using word boundaries.
// Long words are forcibly broken at maxChars.
// Returns HTML with <BR/> line breaks for GraphViz.
func wrapText(text string, maxChars int) string {
    if maxChars <= 0 {
        return text
    }

    words := strings.Fields(text)
    if len(words) == 0 {
        return text
    }

    var lines []string
    var currentLine strings.Builder
    currentLen := 0

    for _, word := range words {
        wordLen := utf8.RuneCountInString(word)

        // Force break long words
        if wordLen > maxChars {
            if currentLine.Len() > 0 {
                lines = append(lines, currentLine.String())
                currentLine.Reset()
                currentLen = 0
            }
            // Break word into chunks
            runes := []rune(word)
            for i := 0; i < len(runes); i += maxChars {
                end := i + maxChars
                if end > len(runes) {
                    end = len(runes)
                }
                lines = append(lines, string(runes[i:end]))
            }
            continue
        }

        // Check if word fits on current line
        needed := wordLen
        if currentLen > 0 {
            needed++ // space before word
        }

        if currentLen+needed > maxChars && currentLen > 0 {
            lines = append(lines, currentLine.String())
            currentLine.Reset()
            currentLen = 0
        }

        if currentLen > 0 {
            currentLine.WriteString(" ")
            currentLen++
        }
        currentLine.WriteString(word)
        currentLen += wordLen
    }

    if currentLine.Len() > 0 {
        lines = append(lines, currentLine.String())
    }

    return strings.Join(lines, "<BR/>")
}
```

### HTML Label with Width Constraint (GraphViz Pattern)
```go
// Source: GraphViz documentation - https://graphviz.org/doc/info/shapes.html
// WIDTH attribute is in points (1/72 inch)
func buildContainerHTMLLabelWithWrap(label *graph.Label, iconRelPath string, textWidth int) string {
    var sb strings.Builder
    sb.WriteString(`<table border="0" cellpadding="0" cellspacing="0">`)

    // Row 1: Icon + Name
    sb.WriteString(`<tr align="center">`)
    if iconRelPath != "" {
        sb.WriteString(`<td width="36" valign="middle">`)
        sb.WriteString(`<img src="`)
        sb.WriteString(iconRelPath)
        sb.WriteString(`" width="32" height="32"/>`)
        sb.WriteString(`</td>`)
    }
    // Text cell with constrained width
    sb.WriteString(`<td width="`)
    sb.WriteString(strconv.Itoa(textWidth))
    sb.WriteString(`" valign="bottom"><b>`)
    // Wrap and escape the name
    wrappedName := wrapText(label.Name, estimateCharsFromWidth(textWidth))
    sb.WriteString(wrappedName) // Already contains <BR/> for breaks
    sb.WriteString(`</b></td>`)
    sb.WriteString(`</tr>`)

    // ... additional rows for technology, description

    sb.WriteString(`</table>`)
    return sb.String()
}
```

### CLI Flag Pattern (Existing Code Style)
```go
// Source: cmd/c4drill/root.go - existing pattern
var (
    format     string
    outputDir  string
    expanded   bool
    labelRatio float64 // NEW
    version    = "dev"
)

func NewRootCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "c4drill <input.toml>",
        // ...
    }

    // Existing flags
    cmd.PersistentFlags().StringVarP(&format, "format", "f", formatSVG, ...)
    cmd.PersistentFlags().StringVarP(&outputDir, "output", "o", "", ...)
    cmd.PersistentFlags().BoolVar(&expanded, "expanded", false, ...)

    // NEW: Label ratio flag
    cmd.PersistentFlags().Float64Var(&labelRatio, "label-ratio", 0,
        "Width:height ratio for unit labels (default: 1.6, credit card proportions)")

    return cmd
}

// In runRoot or initialization:
func getLabelRatio() float64 {
    if labelRatio > 0 {
        return labelRatio
    }
    if envVal := os.Getenv("C4DRILL_LABEL_RATIO"); envVal != "" {
        if parsed, err := strconv.ParseFloat(envVal, 64); err == nil && parsed > 0 {
            return parsed
        }
    }
    return 1.6 // Default
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Record-based labels | HTML-like labels | Phase 12 | Allows table formatting with icons |
| Static label widths | Dynamic width calculation | This phase | Achieves target proportions |

**Deprecated/outdated:**
- Using `shape=record` with `|` separators: Replaced by HTML labels for better control

## Open Questions

1. **Character width estimation accuracy**
   - What we know: GraphViz uses proportional fonts, so character count is approximate
   - What's unclear: Exact conversion from points to character count
   - Recommendation: Use ~8 points per character as rough estimate; allow user adjustment via ratio

2. **Maximum line length for forced breaks**
   - What we know: Long URLs/identifiers should break at some point
   - What's unclear: Optimal threshold (30 chars? 40? depends on width)
   - Recommendation: Calculate from text width: `maxChars = textWidthPoints / 8`

3. **Row height estimation**
   - What we know: Need to estimate height to calculate target width
   - What's unclear: Exact points per row (font size + padding + cell spacing)
   - Recommendation: Use 18 points per row (12pt font + 6pt padding/margins)

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go testing + stretchr/testify v1.11.1 |
| Config file | None - Go standard |
| Quick run command | `go test ./internal/render/... -run TestWrap -v` |
| Full suite command | `go test ./internal/render/... -v` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|--------------|
| D-03 | Word wrapping with word boundaries | unit | `go test ./internal/render/... -run TestWrapText -v` | No - Wave 0 |
| D-03 | Character fallback for long words | unit | `go test ./internal/render/... -run TestWrapText/long_word -v` | No - Wave 0 |
| D-04 | All fields wrapped | unit | `go test ./internal/render/... -run Test.*HTMLLabel.* -v` | Partial - needs update |
| D-05 | CLI flag --label-ratio | unit | `go test ./cmd/c4drill/... -run TestLabelRatio -v` | No - Wave 0 |
| D-05 | Env var C4DRILL_LABEL_RATIO | unit | `go test ./cmd/c4drill/... -run TestLabelRatioEnv -v` | No - Wave 0 |

### Sampling Rate
- **Per task commit:** `go test ./internal/render/... -v`
- **Per wave merge:** `go test ./... -v`
- **Phase gate:** Full suite green before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/render/wrap.go` - word wrapping implementation
- [ ] `internal/render/wrap_test.go` - wrapping unit tests
- [ ] Update `cmd/c4drill/root.go` - add --label-ratio flag
- [ ] `cmd/c4drill/root_test.go` (if exists) - flag/env tests
- [ ] Update `internal/render/html_labels_internal_test.go` - test wrapped labels

## Sources

### Primary (HIGH confidence)
- GraphViz HTML Labels Documentation - https://graphviz.org/doc/info/shapes.html
- Project source: `internal/render/labels.go` - existing label builders
- Project source: `cmd/c4drill/root.go` - existing CLI flag patterns
- Project source: `internal/graph/graph.go` - Label struct definition

### Secondary (MEDIUM confidence)
- muesli/reflow GitHub - https://github.com/muesli/reflow (optional library for word wrapping)
- Go standard library `strings` and `unicode/utf8` packages

### Tertiary (LOW confidence)
- None - implementation is straightforward based on standard patterns

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - No new dependencies needed; existing patterns sufficient
- Architecture: HIGH - GraphViz HTML label patterns well-documented; existing code provides clear structure
- Pitfalls: HIGH - Common issues identified from GraphViz documentation and Unicode handling

**Research date:** 2026-03-23
**Valid until:** 30 days - GraphViz HTML label syntax is stable; Go patterns are stable
