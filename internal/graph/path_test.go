package graph_test

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeExploreURL(t *testing.T) {
	t.Parallel()

	// C1->C2: C1 file is at {basename}.dot, C2 target is at {basename}/mainapp.svg
	t.Run("C1 to C2 navigation", func(t *testing.T) {
		t.Parallel()

		url := graph.ComputeExploreURL("", "mainapp", "diagram", "svg")
		assert.Equal(t, "diagram/mainapp.svg", url)
	})

	// C2->C3: C2 file is at {basename}/mainapp.dot, C3 target is at {basename}/mainapp/api.svg
	t.Run("C2 to C3 navigation", func(t *testing.T) {
		t.Parallel()

		url := graph.ComputeExploreURL("mainapp", "mainapp.api", "diagram", "svg")
		assert.Equal(t, "mainapp/api.svg", url)
	})

	// Special chars in target path are URL-encoded
	t.Run("URL encoding for special characters", func(t *testing.T) {
		t.Parallel()

		url := graph.ComputeExploreURL("", "api (v2)", "diagram", "svg")
		assert.Equal(t, "diagram/api%20%28v2%29.svg", url)
	})

	// C3->C4: mainapp.api file at {basename}/mainapp/api.dot,
	// mainapp.api.auth target at {basename}/mainapp/api/auth.svg
	// Common directory prefix: mainapp, so relative is api/auth.svg
	t.Run("nested path with multiple levels", func(t *testing.T) {
		t.Parallel()

		url := graph.ComputeExploreURL("mainapp.api", "mainapp.api.auth", "diagram", "svg")
		assert.Equal(t, "api/auth.svg", url)
	})

	// Format parameter is ignored; URLs always use SVG for browser navigation
	t.Run("different format still produces svg URLs", func(t *testing.T) {
		t.Parallel()

		url := graph.ComputeExploreURL("", "mainapp", "diagram", "dot")
		assert.Equal(t, "diagram/mainapp.svg", url)
	})

	// C2->C2 sibling: from mainapp.dot to otherunit.dot (both at same depth under basename/)
	t.Run("C2 to sibling C2 navigation", func(t *testing.T) {
		t.Parallel()

		url := graph.ComputeExploreURL("mainapp", "otherunit", "diagram", "svg")
		assert.Equal(t, "otherunit.svg", url)
	})

	// C3->C3 sibling: from mainapp/api.dot to mainapp/web.svg (both in same directory)
	t.Run("C3 to C3 sibling navigation", func(t *testing.T) {
		t.Parallel()

		url := graph.ComputeExploreURL("mainapp.api", "mainapp.web", "diagram", "svg")
		assert.Equal(t, "web.svg", url)
	})
}

func TestComputeBackLinkURL(t *testing.T) {
	t.Parallel()

	// C2->C1: from {basename}/mainapp.dot back to {basename}.dot
	t.Run("C2 to C1 back-link", func(t *testing.T) {
		t.Parallel()

		url := graph.ComputeBackLinkURL("mainapp", "diagram", "svg")
		assert.Equal(t, "../diagram.svg", url)
	})

	// C3->C2: from {basename}/mainapp/api.dot back to {basename}/mainapp.dot
	t.Run("C3 to C2 back-link", func(t *testing.T) {
		t.Parallel()

		url := graph.ComputeBackLinkURL("mainapp.api", "diagram", "svg")
		assert.Equal(t, "../mainapp.svg", url)
	})

	t.Run("C1 has no back-link", func(t *testing.T) {
		t.Parallel()

		url := graph.ComputeBackLinkURL("", "diagram", "svg")
		assert.Empty(t, url)
	})

	t.Run("back-link with special characters", func(t *testing.T) {
		t.Parallel()

		// From "mainapp.api (v2)" back to parent "mainapp"
		url := graph.ComputeBackLinkURL("mainapp.api (v2)", "diagram", "svg")
		assert.Equal(t, "../mainapp.svg", url)
	})
}

func TestBuildBreadcrumbPath(t *testing.T) {
	t.Parallel()

	t.Run("empty path returns nil", func(t *testing.T) {
		t.Parallel()

		items := graph.BuildBreadcrumbPath("", "diagram", "svg")
		assert.Nil(t, items)
	})

	t.Run("single level breadcrumb", func(t *testing.T) {
		t.Parallel()

		items := graph.BuildBreadcrumbPath("mainapp", "diagram", "svg")
		require.Len(t, items, 1)
		assert.Equal(t, "mainapp", items[0].Name)
		assert.Empty(t, items[0].URL) // Current level has no URL
	})

	t.Run("multi-level breadcrumb", func(t *testing.T) {
		t.Parallel()

		items := graph.BuildBreadcrumbPath("mainapp.api.auth", "diagram", "svg")
		require.Len(t, items, 3)

		// First level (ancestor) - has URL
		assert.Equal(t, "mainapp", items[0].Name)
		assert.NotEmpty(t, items[0].URL)

		// Second level (ancestor) - has URL
		assert.Equal(t, "api", items[1].Name)
		assert.NotEmpty(t, items[1].URL)

		// Third level (current) - no URL
		assert.Equal(t, "auth", items[2].Name)
		assert.Empty(t, items[2].URL)
	})
}

func TestURLEncodePath(t *testing.T) {
	t.Parallel()

	t.Run("simple path unchanged", func(t *testing.T) {
		t.Parallel()

		encoded := graph.URLEncodePath("mainapp")
		assert.Equal(t, "mainapp", encoded)
	})

	t.Run("spaces encoded", func(t *testing.T) {
		t.Parallel()

		encoded := graph.URLEncodePath("api service")
		assert.Equal(t, "api%20service", encoded)
	})

	t.Run("parentheses encoded", func(t *testing.T) {
		t.Parallel()

		encoded := graph.URLEncodePath("api (v2)")
		assert.Equal(t, "api%20%28v2%29", encoded)
	})
}
