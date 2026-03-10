package graph_test

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeExploreURL(t *testing.T) {
	t.Parallel()

	// Test 1: ComputeExploreURL for C1->C2: currentPath="", targetPath="mainapp" returns "./mainapp.svg"
	t.Run("C1 to C2 navigation", func(t *testing.T) {
		t.Parallel()

		url := graph.ComputeExploreURL("", "mainapp", "diagram", "svg")
		assert.Equal(t, "./mainapp.svg", url)
	})

	// Test 2: ComputeExploreURL for C2->C3: currentPath="mainapp", targetPath="mainapp.api" returns "./mainapp/api.svg"
	t.Run("C2 to C3 navigation", func(t *testing.T) {
		t.Parallel()

		url := graph.ComputeExploreURL("mainapp", "mainapp.api", "diagram", "svg")
		assert.Equal(t, "./mainapp/api.svg", url)
	})

	// Test 3: ComputeExploreURL with special chars: targetPath="api (v2)" returns URL-encoded path
	t.Run("URL encoding for special characters", func(t *testing.T) {
		t.Parallel()

		url := graph.ComputeExploreURL("", "api (v2)", "diagram", "svg")
		assert.Equal(t, "./api%20%28v2%29.svg", url)
	})

	t.Run("nested path with multiple levels", func(t *testing.T) {
		t.Parallel()

		url := graph.ComputeExploreURL("mainapp.api", "mainapp.api.auth", "diagram", "svg")
		assert.Equal(t, "./mainapp/api/auth.svg", url)
	})

	t.Run("different format", func(t *testing.T) {
		t.Parallel()

		url := graph.ComputeExploreURL("", "mainapp", "diagram", "png")
		assert.Equal(t, "./mainapp.png", url)
	})
}

func TestComputeBackLinkURL(t *testing.T) {
	t.Parallel()

	// Test 4: ComputeBackLinkURL for C2->C1: currentPath="mainapp", format="svg" returns "../basename.svg"
	t.Run("C2 to C1 back-link", func(t *testing.T) {
		t.Parallel()

		url := graph.ComputeBackLinkURL("mainapp", "diagram", "svg")
		assert.Equal(t, "../diagram.svg", url)
	})

	// Test 5: ComputeBackLinkURL for C3->C2: currentPath="mainapp.api", format="svg" returns "../mainapp.svg"
	// File structure: C3 at diagram/mainapp/api.svg, C2 at diagram/mainapp.svg
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
