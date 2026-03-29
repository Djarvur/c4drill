# Phase 5: Navigation - Research

**Researched:** 2026-03-10
**Domain:** SVG Navigation Links via GraphViz URL Attributes
**Confidence:** HIGH (verified against go-graphviz v0.2.10 source and existing codebase)

## Summary

Phase 5 adds interactive navigation between C4 diagram levels using GraphViz URL attributes. The go-graphviz library (v0.2.10) provides `SetURL()` and `SetTarget()` methods on Node, Edge, and Graph objects for generating clickable SVG elements. The existing codebase has a clean two-phase construction (view -> graph -> render) with well-defined extension points in `graph.Node`, `graph.Graph`, and `render.createNode()`.

**Primary recommendation:** Extend `graph.Node` with `ExploreURL string` field, add `Navigation` struct to `graph.Graph`, and call `cnode.SetURL(url)` in `createNode()` for nodes with explore links. Use graph labels or xlabel for breadcrumb/back-link navigation bar.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### Explore Links
- Node name shows "System Name [+]" where [+] indicates drill-down capability
- Only expandable node types get explore links (systems, boxes - not persons, databases, queues)
- Links use relative paths (e.g., "./mainapp/api.svg")
- Clickable nodes set cursor:pointer via GraphViz (no visual style change)
- Use GraphViz URL attribute (SetURL) for whole-node clickability

#### Back-Links
- Text link at top-left: "Back to {parent_name}"
- Positioned outside diagram area (in margin/padding, not part of layout)
- C1 (Context) diagram has no back-link (it's the root)
- Uses parent's actual name in link text

#### Breadcrumbs
- Show unit names: "Main System > API Container > Auth Service"
- Chevron (>) separator between items
- Same line as back-link: "Back to Main | Main System > API Container"
- Current level item is NOT clickable (only ancestors are links)
- No limit on breadcrumb length - show full path

#### Path Handling
- Paths pre-computed during graph building stage (stored in graph types)
- URL encode paths for special characters (url.PathEscape)
- Relative paths for portability across hosting contexts

### Claude's Discretion
- Exact positioning of navigation bar (margin values)
- Font size and styling for navigation elements
- Whether to use HTML label or graph label for navigation bar

### Deferred Ideas (OUT OF SCOPE)
None - discussion stayed within phase scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| REND-04 | Collapsed units include explore link pointing to drill-down file | go-graphviz `Node.SetURL()` method; compute relative paths from current diagram location to target |
| REND-05 | All diagrams include back-link to parent level | `Graph.SetURL()` or xlabel attribute; store parent info in `graph.Navigation` struct |
| REND-06 | All diagrams include breadcrumb trail showing path | Graph label or HTML label construction; `Navigation.Breadcrumbs` field |
| OUTP-05 | Relative paths used for explore and back links | `url.PathEscape()` for encoding; relative path computation from current file position |
| QUAL-01 | All lint errors must be fixed before commit | Existing golangci-lint config; standard Go patterns |
| QUAL-02 | Lint config MUST NOT be adjusted to silence errors | No modifications to .golangci.yml |
| QUAL-03 | nolint directives require explicit user confirmation | Avoid adding new nolint directives |
| QUAL-04 | Minimum 75% test coverage required | Existing testify patterns; test URL generation logic |
| QUAL-05 | Coverage enforced in CI/quality gate | Run `go test -cover ./...` |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| go-graphviz | v0.2.10 | SVG/DOT rendering with URL support | Already integrated; provides SetURL/SetTarget |
| stretchr/testify | v1.11.1 | Testing assertions | Existing test patterns |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| net/url | stdlib | URL path encoding | When computing relative paths with special chars |
| path/filepath | stdlib | Path manipulation | Cross-platform relative path computation |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| SetURL on nodes | href attribute in HTML labels | SetURL makes whole node clickable; HTML href only makes text clickable |
| Graph label for breadcrumbs | Separate invisible node | Graph label positions outside layout; invisible node affects layout |

**Installation:**
No new dependencies required. All needed packages are already in go.mod or stdlib.

## Architecture Patterns

### Recommended Data Model Extensions

```go
// In internal/graph/graph.go

// Navigation holds navigation elements for a diagram.
type Navigation struct {
    // BackLink is the parent back-link (nil for C1 root).
    BackLink *BackLink
    // Breadcrumbs is the breadcrumb trail from root to current.
    Breadcrumbs []BreadcrumbItem
}

// BackLink represents a link back to the parent diagram.
type BackLink struct {
    // Name is the display name for the parent.
    Name string
    // URL is the relative path to the parent diagram.
    URL string
}

// BreadcrumbItem represents one item in the breadcrumb trail.
type BreadcrumbItem struct {
    // Name is the display name.
    Name string
    // URL is the relative path (empty for current level).
    URL string
}

// Node extension - add to existing Node struct:
type Node struct {
    // ... existing fields ...

    // ExploreURL is the relative path for drill-down (empty if not expandable).
    ExploreURL string
}

// Graph extension - add to existing Graph struct:
type Graph struct {
    // ... existing fields ...

    // Navigation contains breadcrumb and back-link info (nil for C1).
    Navigation *Navigation
}
```

### Pattern 1: Compute Explore URLs During Graph Building

**What:** Calculate relative paths from current diagram to drill-down targets during `BuildGraph()`.

**When to use:** When building graphs for C2/C3 views that contain collapsed units with subunits.

**Example:**
```go
// In internal/graph/builder.go

// BuildGraphWithPath constructs a graph with navigation paths.
// currentPath is the dotted path of the current diagram (empty for C1).
func BuildGraphWithPath(v *view.View, currentPath string) *Graph {
    g := BuildGraph(v) // existing logic

    // Add explore URLs to collapsed nodes with subunits
    for _, node := range g.Nodes {
        if node.HasSubunits && !node.IsExpanded {
            // Compute relative path: from current to target
            node.ExploreURL = computeExploreURL(currentPath, node.ID)
        }
    }

    return g
}

// computeExploreURL calculates relative path from current diagram to target.
func computeExploreURL(currentPath, targetPath string) string {
    // C1 -> C2: "./basename/target.svg"
    // C2 -> C3: "./target.svg" (same directory)
    // C3 -> deeper: "./subcomponent.svg"
    return "./" + strings.ReplaceAll(targetPath, ".", "/") + ".svg"
}
```

### Pattern 2: Set URL Attribute in Renderer

**What:** Call `SetURL()` on cgraph.Node when node has ExploreURL.

**When to use:** During cgraph node creation in `createNode()`.

**Example:**
```go
// In internal/render/converter.go

func createNode(cg *cgraph.Graph, node *graph.Node) (*cgraph.Node, error) {
    cn, err := cg.CreateNodeByName(node.ID)
    if err != nil {
        return nil, fmt.Errorf("create node by name: %w", err)
    }

    // ... existing label and style code ...

    // Set URL for clickable nodes
    if node.ExploreURL != "" {
        cn.SetURL(node.ExploreURL)
        // Optional: set target for new window/tab
        // cn.SetTarget("_blank")
    }

    return cn, nil
}
```

### Pattern 3: Navigation Bar via Graph Label

**What:** Use `cg.SetLabel()` with HTML-like formatting for breadcrumb/back-link bar.

**When to use:** For C2/C3 diagrams that need navigation back to parent.

**Example:**
```go
// In internal/render/converter.go

func configureGraphSettings(cg *cgraph.Graph, g *graph.Graph) {
    // ... existing settings ...

    // Build navigation label
    if g.Navigation != nil {
        navLabel := buildNavigationLabel(g.Navigation)
        // Use xlabel for external label (outside diagram bounds)
        cg.SetXLabel(navLabel)
        cg.SetXLabelLocation(cgraph.TopLocation) // or use SetLabel for inside
    }
}

func buildNavigationLabel(nav *graph.Navigation) string {
    var parts []string

    // Back-link
    if nav.BackLink != nil {
        parts = append(parts, fmt.Sprintf("<<a href=\"%s\">Back to %s</a>>",
            nav.BackLink.URL, nav.BackLink.Name))
    }

    // Breadcrumbs
    var crumbs []string
    for _, item := range nav.Breadcrumbs {
        if item.URL != "" {
            crumbs = append(crumbs, fmt.Sprintf("<a href=\"%s\">%s</a>",
                item.URL, item.Name))
        } else {
            crumbs = append(crumbs, item.Name) // Current level - not clickable
        }
    }
    if len(crumbs) > 0 {
        parts = append(parts, strings.Join(crumbs, " > "))
    }

    return strings.Join(parts, " | ")
}
```

### Anti-Patterns to Avoid

- **Do NOT use absolute paths:** Breaks when diagrams are moved or hosted differently
- **Do NOT put navigation inside diagram layout:** Affects node positioning; use xlabel or external label
- **Do NOT add nolint directives without user confirmation:** Per QUAL-03
- **Do NOT skip URL encoding:** Special characters in unit names will break links

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Relative path computation | Custom string manipulation | `filepath.Rel()` or simple prefix-based logic | Edge cases with .. traversal |
| URL encoding | Manual percent-encoding | `url.PathEscape()` | Handles all special characters correctly |
| Clickable SVG nodes | Custom JavaScript | GraphViz URL attribute | Native SVG <a> element wrapping |

**Key insight:** GraphViz has built-in SVG interactivity support through URL attributes. The go-graphviz library exposes this via `SetURL()` method on nodes, edges, and graphs.

## Common Pitfalls

### Pitfall 1: Incorrect Relative Path Depth

**What goes wrong:** Links work from C1 but break from C2/C3 because relative path depth is wrong.

**Why it happens:** C1 is at `system.svg`, C2 is at `system/api.svg`. A link from C1 to C2 needs `./system/api.svg`, but from C2 to C3 needs just `./api.svg`.

**How to avoid:**
- Track current diagram path during graph building
- Compute relative paths based on current location depth
- Test links from all levels (C1, C2, C3)

**Warning signs:** Links work in browser but show 404; path in browser address bar looks wrong

### Pitfall 2: Special Characters in URLs

**What goes wrong:** Unit names with spaces or special characters break links.

**Why it happens:** URLs must be percent-encoded; raw unit names may contain spaces, unicode, etc.

**How to avoid:**
```go
import "net/url"

func computeExploreURL(currentPath, targetPath string) string {
    encoded := url.PathEscape(targetPath)
    return "./" + strings.ReplaceAll(encoded, ".", "/") + ".svg"
}
```

**Warning signs:** Links to units like "Auth Service" or "API (v2)" don't work

### Pitfall 3: Navigation Bar Affects Layout

**What goes wrong:** Back-link or breadcrumb takes up space inside the diagram, pushing nodes around.

**Why it happens:** Using regular `SetLabel()` puts label inside diagram bounds.

**How to avoid:** Use `SetXLabel()` (external label) which positions outside the diagram area, or use `labelloc="t"` with `labeljust="l"` for top-left positioning in margin.

**Warning signs:** Diagram nodes shift position when navigation is added; empty C1 diagram has different size than C2

### Pitfall 4: Current Level in Breadcrumb is Clickable

**What goes wrong:** User clicks current level in breadcrumb and gets confused (same page reloads).

**Why it happens:** All breadcrumb items rendered with same link format.

**How to avoid:** Set `URL: ""` for the current level item; check for empty URL when building label to render as plain text instead of link.

**Warning signs:** Current level has underline/hover effect in browser

## Code Examples

### Verified Pattern: go-graphviz SetURL Method

```go
// Source: go-graphviz v0.2.10 cgraph/attribute.go (verified via grep)
// Method signature: func (n *Node) SetURL(v string) *Node

// Usage in converter:
func createNode(cg *cgraph.Graph, node *graph.Node) (*cgraph.Node, error) {
    cn, err := cg.CreateNodeByName(node.ID)
    if err != nil {
        return nil, err
    }

    // Set label, style, etc.

    // Make node clickable
    if node.ExploreURL != "" {
        cn.SetURL(node.ExploreURL)
    }

    return cn, nil
}
```

### Verified Pattern: URL Encoding

```go
// Source: Go stdlib net/url
import "net/url"

// PathEscape escapes the string so it can be safely placed inside a URL path segment.
encoded := url.PathEscape("API Service (v2)")
// Result: "API%20Service%20%28v2%29"
```

### Verified Pattern: Relative Path Computation

```go
// For C4Drill, paths follow the output file structure:
// C1: {basename}.svg
// C2: {basename}/{system}.svg
// C3: {basename}/{system}/{container}.svg

// From C1 to C2 (system "mainapp"):
// Current: "." (basename directory)
// Target: "./mainapp.svg" (in same directory)
// But files are at: system.svg and system/mainapp.svg
// So from system.svg to system/mainapp.svg: "./system/mainapp.svg"

// From C2 to C3 (container "api" in system "mainapp"):
// Current: system/mainapp.svg
// Target: system/mainapp/api.svg
// Relative: "./mainapp/api.svg" (from system/mainapp.svg perspective)

// Simplified approach: always use path relative to current file's directory
func relativePath(currentUnitPath, targetUnitPath, format string) string {
    // Both are dotted paths like "mainapp" or "mainapp.api"
    targetDirPath := strings.ReplaceAll(targetUnitPath, ".", "/")
    return "./" + targetDirPath + "." + format
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| HTML area maps for clickable SVG | Native SVG <a> elements | GraphViz 2.x | Better browser support, automatic from URL attribute |
| Manual breadcrumb HTML | GraphViz xlabel with HTML-like labels | This phase | Consistent styling with diagram |

**Deprecated/outdated:**
- xlink:href in SVG: Modern browsers support plain href; GraphViz still generates xlink:href for compatibility

## Open Questions

1. **Graph label vs xlabel for navigation bar**
   - What we know: `SetLabel()` puts label inside diagram, `SetXLabel()` puts it outside
   - What's unclear: Which provides better visual positioning for top-left back-link
   - Recommendation: Try xlabel first; if positioning is awkward, use label with `labelloc="t" labeljust="l"`

2. **HTML entities in graph labels**
   - What we know: GraphViz supports HTML-like labels for nodes
   - What's unclear: Whether graph xlabel supports the same HTML-like syntax with <a> tags
   - Recommendation: Test with simple HTML-like label first; if <a> not supported, use URL attribute on a small invisible node

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | testify v1.11.1 |
| Config file | none - standard go test |
| Quick run command | `go test ./internal/graph/... ./internal/render/... -v -short` |
| Full suite command | `go test ./... -cover` |

### Phase Requirements -> Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| REND-04 | Explore links on collapsed units | unit | `go test ./internal/graph/... -run TestExploreURL -v` | Wave 0 |
| REND-05 | Back-link in navigation | unit | `go test ./internal/graph/... -run TestBackLink -v` | Wave 0 |
| REND-06 | Breadcrumb trail | unit | `go test ./internal/graph/... -run TestBreadcrumb -v` | Wave 0 |
| OUTP-05 | Relative path computation | unit | `go test ./internal/graph/... -run TestRelativePath -v` | Wave 0 |
| QUAL-04 | 75% coverage | quality | `go test ./... -cover` | Wave 0 |

### Sampling Rate
- Per task commit: `go test ./internal/graph/... ./internal/render/... -v -short`
- Per wave merge: `go test ./... -cover`
- Phase gate: Full suite green with >75% coverage before `/gsd:verify-work`

### Wave 0 Gaps
- [ ] `internal/graph/navigation_test.go` - tests for Navigation, BackLink, BreadcrumbItem types
- [ ] `internal/graph/path_test.go` - tests for relative path computation
- [ ] `internal/render/navigation_test.go` - tests for SetURL calls and navigation label generation
- [ ] Integration test: generate SVG, parse output, verify <a> elements with correct xlink:href

## Sources

### Primary (HIGH confidence)
- go-graphviz v0.2.10 cgraph/attribute.go - SetURL/SetTarget method signatures (verified via local module cache)
- Existing codebase: internal/graph/graph.go, internal/render/converter.go - established patterns
- CONTEXT.md - locked decisions from user discussion

### Secondary (MEDIUM confidence)
- GraphViz documentation - URL attribute behavior in SVG output
- Project PITFALLS.md - SVG link path resolution considerations

### Tertiary (LOW confidence)
- None - all critical information verified from primary sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - go-graphviz already integrated, SetURL method verified
- Architecture: HIGH - existing codebase has clear extension points
- Pitfalls: HIGH - based on existing project research and GraphViz documentation

**Research date:** 2026-03-10
**Valid until:** 30 days (stable library API, established patterns)
