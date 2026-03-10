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

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestBuildNavigationLabel(t *testing.T) {
	t.Run("empty navigation returns empty string", func(t *testing.T) {
		result := render.BuildNavigationLabel(nil)
		assert.Equal(t, "", result)
	})

	t.Run("navigation with nil backlink and empty breadcrumbs returns empty string", func(t *testing.T) {
		nav := &graph.Navigation{
			BackLink:    nil,
			Breadcrumbs: []graph.BreadcrumbItem{},
		}
		result := render.BuildNavigationLabel(nav)
		assert.Equal(t, "", result)
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
		assert.Contains(t, result, "<a href=")
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
		// Ancestor should be clickable
		assert.Contains(t, result, "<a href=\"../mainsystem.svg\">Main System</a>")
		// Current level (empty URL) should be plain text
		assert.Contains(t, result, "API Container")
		assert.Contains(t, result, " > ")
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
		// Should have separator between backlink and breadcrumbs
		assert.Contains(t, result, " | ")
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
		// Ancestor should have link
		assert.Contains(t, result, "<a href=\"../ancestor.svg\">Ancestor</a>")
		// Current should NOT have link (plain text only)
		assert.NotContains(t, result, "<a href=\"\">Current</a>")
		assert.Contains(t, result, "Current")
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
		// Both ancestors should be clickable
		assert.Contains(t, result, "<a href=\"../../root.svg\">Root</a>")
		assert.Contains(t, result, "<a href=\"../parent.svg\">Parent</a>")
		// Should have separator between items
		assert.Contains(t, result, " > ")
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
		assert.Equal(t, "", result)
	})
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestNavigationInGraphOutput(t *testing.T) {
	t.Run("C2/C3 diagram with navigation has xlabel in output", func(t *testing.T) {
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
		// Navigation should appear as xlabel in DOT output
		assert.Contains(t, string(output), "xlabel")
		assert.Contains(t, string(output), "Back to Main System")
	})

	t.Run("C1 diagram without navigation has no xlabel", func(t *testing.T) {
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
		// C1 should not have navigation-related xlabel
		dotStr := string(output)
		assert.NotContains(t, dotStr, "Back to")
	})
}
