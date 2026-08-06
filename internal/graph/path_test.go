package graph_test

import (
	"path/filepath"
	"strings"
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

	// 03-04 Gap 1: C3 to C2 ancestor target must yield an UPWARD relative path
	// (the current code emits a downward ".svg"-style path that does NOT resolve
	// to the C2 file from the C3 file's directory). currentPath="mainSystem.sshAuth"
	// lives at {basename}/mainSystem/sshAuth.svg (dir {basename}/mainSystem/);
	// the C2 mainSystem.svg lives at {basename}/mainSystem.svg, so the relative
	// URL must be "../mainSystem.svg".
	t.Run("C3 to C2 ancestor target yields upward relative path", func(t *testing.T) {
		t.Parallel()

		url := graph.ComputeExploreURL("mainSystem.sshAuth", "mainSystem", "multilevel", "svg")
		assert.Equal(t, "../mainSystem.svg", url)
	})

	// 03-04 Gap 1 symptom B: a node whose targetPath equals currentPath (the C3
	// collapsed-ancestor node "mainSystem" rendered inside the C3
	// "mainSystem.sshAuth" diagram) currently emits a broken href=".svg". The
	// self-link guard returns empty so the renderer omits the URL attribute.
	t.Run("self-link target returns empty (Gap 1 symptom B guard)", func(t *testing.T) {
		t.Parallel()

		url := graph.ComputeExploreURL("mainSystem", "mainSystem", "multilevel", "svg")
		assert.Empty(t, url)
	})

	// 03-04 Gap 1: a deeper self-link (e.g. C4 ancestor node equal to currentPath)
	// must also return empty.
	t.Run("self-link returns empty for nested path", func(t *testing.T) {
		t.Parallel()

		url := graph.ComputeExploreURL("mainapp.api.auth", "mainapp.api.auth", "diagram", "svg")
		assert.Empty(t, url)
	})

	// 03-04 Gap 1 regression: C2->C3 descendant href must RESOLVE against the real
	// file tree (multilevel/mainSystem.svg -> multilevel/mainSystem/sshAuth.svg).
	// We assert the emitted url joined with the C2 file's directory points at an
	// existing file — NOT a hardcoded exact string, because the descendant
	// convention is already locked by nested_path_with_multiple_levels above.
	t.Run("C2 to C3 descendant href resolves against real tree", func(t *testing.T) {
		t.Parallel()

		// C2 file mainSystem.svg lives at {basename}/mainSystem.svg -> dir {basename}/.
		// Compute the relative href the renderer would emit.
		href := graph.ComputeExploreURL("mainSystem", "mainSystem.sshAuth", "multilevel", "svg")
		require.NotEmpty(t, href)

		// Reconstruct the expected target file path under {basename}/ and assert
		// that dir(C2 file) + href equals it (file-resolution, no hardcoded string).
		c2FileDir := "multilevel" // {basename}/
		expectedTarget := "multilevel/mainSystem/sshAuth.svg"
		resolved := filepath.Join(c2FileDir, href)
		assert.Equal(t, expectedTarget, filepath.Clean(resolved),
			"C2->C3 descendant href must resolve to the C3 sibling file")
	})

	// 03-04 Gap 1 regression: C3->C3 sibling descendant href must keep resolving
	// after the ancestor fix (guards against the bidirectional fix regressing
	// sibling resolution).
	t.Run("C3 to C3 sibling descendant href resolves", func(t *testing.T) {
		t.Parallel()

		// C3 file sshAuth.svg lives at {basename}/mainSystem/sshAuth.svg
		// -> dir {basename}/mainSystem/.
		href := graph.ComputeExploreURL("mainSystem.sshAuth", "mainSystem.web", "multilevel", "svg")
		require.NotEmpty(t, href)

		c3FileDir := "multilevel/mainSystem"
		expectedTarget := "multilevel/mainSystem/web.svg"
		resolved := filepath.Join(c3FileDir, href)
		assert.Equal(t, expectedTarget, filepath.Clean(resolved),
			"C3->C3 sibling href must resolve to the sibling file")
	})

	// 03-04 Gap 1 regression: C4->C2 ancestor (two levels up) must yield ../../.
	t.Run("C4 to C2 ancestor yields two levels up", func(t *testing.T) {
		t.Parallel()

		// currentPath="mainapp.api.auth" -> file at {basename}/mainapp/api/auth.svg
		// -> dir {basename}/mainapp/api/. Target mainapp.svg at {basename}/mainapp.svg.
		// Relative URL: ../../mainapp.svg.
		url := graph.ComputeExploreURL("mainapp.api.auth", "mainapp", "diagram", "svg")
		assert.Equal(t, "../../mainapp.svg", url)
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

	// 03-04 Gap 3: navigation URLs must always use .svg regardless of the render
	// format being generated, so a .dot diagram still produces browser-navigable
	// .svg links in its breadcrumb/back-link bar (matching ComputeExploreURL).
	t.Run("back-link always uses svg regardless of format", func(t *testing.T) {
		t.Parallel()

		url := graph.ComputeBackLinkURL("mainapp", "diagram", "dot")
		assert.Equal(t, "../diagram.svg", url)
	})

	t.Run("C3 back-link always uses svg regardless of format", func(t *testing.T) {
		t.Parallel()

		url := graph.ComputeBackLinkURL("mainapp.api", "diagram", "dot")
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

	// 03-04 Gap 3: breadcrumb ancestor URLs must always end with .svg regardless
	// of the render format parameter (matching ComputeExploreURL).
	t.Run("breadcrumb URL always uses svg regardless of format", func(t *testing.T) {
		t.Parallel()

		items := graph.BuildBreadcrumbPath("mainapp.api", "diagram", "dot")
		require.Len(t, items, 2)

		// Ancestor (mainapp) has a URL; it must end with .svg even though the
		// format parameter is "dot".
		assert.True(t, strings.HasSuffix(items[0].URL, ".svg"),
			"breadcrumb ancestor URL must end with .svg regardless of format; got %q", items[0].URL)
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
