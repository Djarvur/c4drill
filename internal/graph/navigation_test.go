package graph_test

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/stretchr/testify/assert"
)

func TestNavigationStruct(t *testing.T) {
	t.Parallel()

	// Test 1: Navigation struct with BackLink and Breadcrumbs can be created
	t.Run("Navigation with BackLink and Breadcrumbs", func(t *testing.T) {
		t.Parallel()

		nav := &graph.Navigation{
			BackLink: &graph.BackLink{
				Name: "Main System",
				URL:  "../main.svg",
			},
			Breadcrumbs: []graph.BreadcrumbItem{
				{Name: "Main System", URL: "../main.svg"},
				{Name: "API", URL: ""},
			},
		}

		assert.NotNil(t, nav)
		assert.Equal(t, "Main System", nav.BackLink.Name)
		assert.Equal(t, "../main.svg", nav.BackLink.URL)
		assert.Len(t, nav.Breadcrumbs, 2)
	})
}

func TestBackLinkStruct(t *testing.T) {
	t.Parallel()

	// Test 2: BackLink with Name and URL fields
	t.Run("BackLink with Name and URL", func(t *testing.T) {
		t.Parallel()

		backLink := &graph.BackLink{
			Name: "Parent System",
			URL:  "../parent.svg",
		}

		assert.Equal(t, "Parent System", backLink.Name)
		assert.Equal(t, "../parent.svg", backLink.URL)
	})
}

func TestBreadcrumbItemStruct(t *testing.T) {
	t.Parallel()

	// Test 3: BreadcrumbItem with Name and URL (empty URL for current level)
	t.Run("BreadcrumbItem with Name and URL", func(t *testing.T) {
		t.Parallel()

		// Ancestor item with URL
		ancestor := graph.BreadcrumbItem{
			Name: "Main System",
			URL:  "../main.svg",
		}
		assert.Equal(t, "Main System", ancestor.Name)
		assert.Equal(t, "../main.svg", ancestor.URL)

		// Current level item (empty URL)
		current := graph.BreadcrumbItem{
			Name: "API",
			URL:  "",
		}
		assert.Equal(t, "API", current.Name)
		assert.Empty(t, current.URL)
	})
}

func TestNodeHasExploreURL(t *testing.T) {
	t.Parallel()

	// Test 4: Node has ExploreURL field that defaults to empty string
	t.Run("Node ExploreURL defaults to empty", func(t *testing.T) {
		t.Parallel()

		node := &graph.Node{
			ID:    "test",
			Label: &graph.Label{Name: "Test"},
		}

		assert.Empty(t, node.ExploreURL)
	})

	t.Run("Node ExploreURL can be set", func(t *testing.T) {
		t.Parallel()

		node := &graph.Node{
			ID:         "test",
			Label:      &graph.Label{Name: "Test"},
			ExploreURL: "./subdir/test.svg",
		}

		assert.Equal(t, "./subdir/test.svg", node.ExploreURL)
	})
}

func TestGraphHasNavigation(t *testing.T) {
	t.Parallel()

	// Test 5: Graph has Navigation field that can be nil (for C1)
	t.Run("Graph Navigation can be nil for C1", func(t *testing.T) {
		t.Parallel()

		g := &graph.Graph{
			Title: "C1 Diagram",
		}

		assert.Nil(t, g.Navigation)
	})

	t.Run("Graph Navigation can be set for C2/C3", func(t *testing.T) {
		t.Parallel()

		g := &graph.Graph{
			Title: "C2 Diagram",
			Navigation: &graph.Navigation{
				BackLink: &graph.BackLink{
					Name: "Main",
					URL:  "../main.svg",
				},
			},
		}

		assert.NotNil(t, g.Navigation)
		assert.Equal(t, "Main", g.Navigation.BackLink.Name)
	})
}
