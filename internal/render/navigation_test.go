package render_test

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: Tests in this file do NOT use t.Parallel() because the go-graphviz
// library uses a WASM-based rendering engine that has concurrency issues.

//nolint:paralleltest,funlen // go-graphviz WASM engine has concurrency issues
func TestBuildNavigationLabel(t *testing.T) {
	t.Run("empty navigation returns empty string", func(t *testing.T) {
		result := render.BuildNavigationLabel(nil)
		assert.Empty(t, result)
	})

	t.Run("navigation with nil backlink and empty breadcrumbs returns empty string", func(t *testing.T) {
		nav := &graph.Navigation{
			BackLink:    nil,
			Breadcrumbs: []graph.BreadcrumbItem{},
		}
		result := render.BuildNavigationLabel(nav)
		assert.Empty(t, result)
	})

	t.Run("backlink only produces correct format", func(t *testing.T) {
		nav := &graph.Navigation{
			BackLink: &graph.BackLink{
				Name: "Main System",
				URL:  "../mainsystem.svg",
			},
			Breadcrumbs: nil,
		}
		result := render.BuildNavigationLabel(nav)
		assert.Contains(t, result, "Back to Main System")
		assert.Contains(t, result, "../mainsystem.svg")
		// Gap 2 (03-04): clickable items use <TD HREF="..."> (GraphViz HTML-label
		// idiom); <a href> tags are not supported and are dropped at render time.
		assert.Contains(t, result, `<TD HREF="../mainsystem.svg">Back to Main System</TD>`)
		assert.Contains(t, result, "<TABLE")
	})

	t.Run("breadcrumbs only produces correct format", func(t *testing.T) {
		nav := &graph.Navigation{
			BackLink: nil,
			Breadcrumbs: []graph.BreadcrumbItem{
				{Name: "Main System", URL: "../mainsystem.svg"},
				{Name: "API Container", URL: ""},
			},
		}
		result := render.BuildNavigationLabel(nav)
		// Ancestor should be clickable via TD HREF
		assert.Contains(t, result, `<TD HREF="../mainsystem.svg">Main System</TD>`)
		// Current level (empty URL) should be a plain TD, not a clickable TD
		assert.Contains(t, result, "<TD>API Container</TD>")
		// Breadcrumb separator is a separate cell with the &gt; entity
		assert.Contains(t, result, "<TD>&gt;</TD>")
	})

	t.Run("backlink and breadcrumbs combined with separator", func(t *testing.T) {
		nav := &graph.Navigation{
			BackLink: &graph.BackLink{
				Name: "Main System",
				URL:  "../mainsystem.svg",
			},
			Breadcrumbs: []graph.BreadcrumbItem{
				{Name: "Main System", URL: "../../mainsystem.svg"},
				{Name: "API Container", URL: "../api.svg"},
				{Name: "Auth Service", URL: ""},
			},
		}
		result := render.BuildNavigationLabel(nav)
		// Should have backlink
		assert.Contains(t, result, "Back to Main System")
		// Should have breadcrumb trail
		assert.Contains(t, result, "Main System")
		assert.Contains(t, result, "API Container")
		assert.Contains(t, result, "Auth Service")
		// Separator between backlink and breadcrumbs is a "|" cell
		assert.Contains(t, result, "<TD>|</TD>")
	})

	t.Run("current level with empty URL is plain text not link", func(t *testing.T) {
		nav := &graph.Navigation{
			BackLink: nil,
			Breadcrumbs: []graph.BreadcrumbItem{
				{Name: "Ancestor", URL: "../ancestor.svg"},
				{Name: "Current", URL: ""},
			},
		}
		result := render.BuildNavigationLabel(nav)
		// Ancestor should have a clickable TD
		assert.Contains(t, result, `<TD HREF="../ancestor.svg">Ancestor</TD>`)
		// Current should be a plain TD (no HREF)
		assert.Contains(t, result, "<TD>Current</TD>")
		assert.NotContains(t, result, `<TD HREF="">Current</TD>`)
	})

	t.Run("multiple ancestors all clickable", func(t *testing.T) {
		nav := &graph.Navigation{
			BackLink: nil,
			Breadcrumbs: []graph.BreadcrumbItem{
				{Name: "Root", URL: "../../root.svg"},
				{Name: "Parent", URL: "../parent.svg"},
				{Name: "Current", URL: ""},
			},
		}
		result := render.BuildNavigationLabel(nav)
		// Both ancestors should be clickable via TD HREF
		assert.Contains(t, result, `<TD HREF="../../root.svg">Root</TD>`)
		assert.Contains(t, result, `<TD HREF="../parent.svg">Parent</TD>`)
		// Breadcrumb separator cell between items
		assert.Contains(t, result, "<TD>&gt;</TD>")
	})

	t.Run("backlink with empty URL produces no link", func(t *testing.T) {
		nav := &graph.Navigation{
			BackLink: &graph.BackLink{
				Name: "Main System",
				URL:  "",
			},
			Breadcrumbs: nil,
		}
		result := render.BuildNavigationLabel(nav)
		// Empty URL backlink should not produce output
		assert.Empty(t, result)
	})

	// WR-01 (03-04): a URL containing '&' must be HTML-escaped when embedded in
	// the HREF="..." attribute. url.PathEscape does NOT escape '&' (it is a
	// valid path character), and GraphViz's HTML-like label parser rejects a raw
	// '&' that is not part of a valid entity, silently dropping the navigation
	// anchor. The back-link TD HREF site must therefore emit '&amp;'.
	t.Run("backlink URL with ampersand is HTML-escaped in HREF (WR-01)", func(t *testing.T) {
		nav := &graph.Navigation{
			BackLink: &graph.BackLink{
				Name: "R&D",
				URL:  "../r&d.svg",
			},
			Breadcrumbs: nil,
		}
		result := render.BuildNavigationLabel(nav)
		// The HREF attribute must contain the HTML-escaped URL.
		assert.Contains(t, result, `<TD HREF="../r&amp;d.svg">`)
		// The raw, unescaped '&' must NOT appear inside the HREF attribute —
		// GraphViz would drop the TD. (The display label 'Back to R&D' is a
		// separate concern; here we assert only the HREF.)
		assert.NotContains(t, result, `HREF="../r&d.svg"`)
	})

	// WR-01 (03-04): the breadcrumb TD HREF site must also HTML-escape '&'.
	t.Run("breadcrumb URL with ampersand is HTML-escaped in HREF (WR-01)", func(t *testing.T) {
		nav := &graph.Navigation{
			BackLink: nil,
			Breadcrumbs: []graph.BreadcrumbItem{
				{Name: "R&D", URL: "../r&d.svg"},
				{Name: "Current", URL: ""},
			},
		}
		result := render.BuildNavigationLabel(nav)
		assert.Contains(t, result, `<TD HREF="../r&amp;d.svg">`)
		assert.NotContains(t, result, `HREF="../r&d.svg"`)
	})

	// WR-01 (03-04): already URL-encoded segments (%20, %28) must be left
	// untouched by the HTML-escape — html.EscapeString alters only <, >, &, ',
	// ", so a legitimately encoded URL is unaffected.
	t.Run("already-encoded URL is unaffected by HTML-escape (WR-01)", func(t *testing.T) {
		nav := &graph.Navigation{
			BackLink: &graph.BackLink{
				Name: "API",
				URL:  "api%20(v2).svg",
			},
			Breadcrumbs: nil,
		}
		result := render.BuildNavigationLabel(nav)
		// Parentheses are valid in HTML attributes and not altered by
		// html.EscapeString; %20 stays literal.
		assert.Contains(t, result, `<TD HREF="api%20(v2).svg">`)
	})
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestNavigationInGraphOutput(t *testing.T) {
	t.Run("C2/C3 diagram with navigation has label in output", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "API Container",
			Direction: "TB",
			Navigation: &graph.Navigation{
				BackLink: &graph.BackLink{
					Name: "Main System",
					URL:  "../mainsystem.svg",
				},
				Breadcrumbs: []graph.BreadcrumbItem{
					{Name: "Main System", URL: "../mainsystem.svg"},
					{Name: "API Container", URL: ""},
				},
			},
			Nodes: []*graph.Node{
				{ID: "api", Label: &graph.Label{Name: "API"}, Shape: graph.ShapeHTML},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)
		// Navigation should appear in the label attribute in DOT output
		assert.Contains(t, string(output), "label")
		assert.Contains(t, string(output), "Back to Main System")
	})

	t.Run("C1 diagram without navigation has no navigation label", func(t *testing.T) {
		g := &graph.Graph{
			Title:      "System Context",
			Direction:  "TB",
			Navigation: nil, // C1 has no navigation
			Nodes: []*graph.Node{
				{ID: "system", Label: &graph.Label{Name: "System"}, Shape: graph.ShapeHTML},
			},
		}

		output, err := render.RenderDOT(g)
		require.NoError(t, err)
		// C1 should not have navigation-related label
		dotStr := string(output)
		assert.NotContains(t, dotStr, "Back to")
	})
}
