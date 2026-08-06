package render_test

import (
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: Tests in this file do NOT use t.Parallel() because the go-graphviz
// library uses a WASM-based rendering engine that concurrency issues.

// navFontTag is the muted-font wrapper applied to every nav cell, asserted as
// a substring in the styled TDs below. Kept in sync with navigation.go.
const navFontOpen = `<FONT POINT-SIZE="10" COLOR="#666666">`

//nolint:paralleltest,funlen // go-graphviz WASM engine has concurrency issues
func TestBuildNavigationLabel(t *testing.T) {
	t.Run("nil navigation returns empty string", func(t *testing.T) {
		result := render.BuildNavigationLabel(nil)
		assert.Empty(t, result)
	})

	t.Run("empty breadcrumbs returns empty string", func(t *testing.T) {
		nav := &graph.Navigation{
			Breadcrumbs: []graph.BreadcrumbItem{},
		}
		result := render.BuildNavigationLabel(nav)
		assert.Empty(t, result)
	})

	t.Run("single breadcrumb (current level only) renders plain", func(t *testing.T) {
		nav := &graph.Navigation{
			Breadcrumbs: []graph.BreadcrumbItem{
				{Name: "Current", URL: ""},
			},
		}
		result := render.BuildNavigationLabel(nav)
		assert.Contains(t, result, "Current")
		// Current level is plain (no HREF, no <U>)
		assert.Contains(t, result, navFontOpen+"Current</FONT>")
		assert.NotContains(t, result, "HREF=")
		assert.NotContains(t, result, "<U>")
	})

	t.Run("ancestor + current: ancestor clickable and underlined", func(t *testing.T) {
		nav := &graph.Navigation{
			Breadcrumbs: []graph.BreadcrumbItem{
				{Name: "Main System", URL: "../mainsystem.svg"},
				{Name: "API Container", URL: ""},
			},
		}
		result := render.BuildNavigationLabel(nav)
		// Ancestor is clickable via TD HREF, underlined, with trailing " >" separator
		// merged into the same cell (avoids GraphViz column-stretching gaps).
		assert.Contains(t, result, `<TD HREF="../mainsystem.svg">`+navFontOpen+`<U>Main System</U> &gt;</FONT></TD>`)
		// Current level is plain (no HREF, no underline, no trailing separator)
		assert.Contains(t, result, "<TD>"+navFontOpen+"API Container</FONT></TD>")
		assert.NotContains(t, result, "<U>API Container</U>")
	})

	t.Run("multiple ancestors all clickable and underlined", func(t *testing.T) {
		nav := &graph.Navigation{
			Breadcrumbs: []graph.BreadcrumbItem{
				{Name: "Root", URL: "../../root.svg"},
				{Name: "Parent", URL: "../parent.svg"},
				{Name: "Current", URL: ""},
			},
		}
		result := render.BuildNavigationLabel(nav)
		// Both ancestors carry trailing " >" inside their cells
		assert.Contains(t, result, `<TD HREF="../../root.svg">`+navFontOpen+`<U>Root</U> &gt;</FONT></TD>`)
		assert.Contains(t, result, `<TD HREF="../parent.svg">`+navFontOpen+`<U>Parent</U> &gt;</FONT></TD>`)
		// Current is plain, no trailing separator
		assert.Contains(t, result, "<TD>"+navFontOpen+"Current</FONT></TD>")
	})

	t.Run("no back-link (breadcrumb-only)", func(t *testing.T) {
		nav := &graph.Navigation{
			Breadcrumbs: []graph.BreadcrumbItem{
				{Name: "Main System", URL: "../mainsystem.svg"},
				{Name: "Current", URL: ""},
			},
		}
		result := render.BuildNavigationLabel(nav)
		// The old "Back to X" affordance is gone; breadcrumb alone covers nav.
		assert.NotContains(t, result, "Back to")
		assert.NotContains(t, result, "|", "no '|' separator without back-link")
	})

	t.Run("all nav cells carry muted font styling", func(t *testing.T) {
		nav := &graph.Navigation{
			Breadcrumbs: []graph.BreadcrumbItem{
				{Name: "Root", URL: "../root.svg"},
				{Name: "Current", URL: ""},
			},
		}
		result := render.BuildNavigationLabel(nav)
		// Every nav cell (both <TD HREF=...> and <TD>) should contain the muted
		// <FONT> wrapper. Count cells by "<TD" (matches both forms) and verify
		// each is followed by a <FONT> tag before its </TD>.
		cellOpenCount := strings.Count(result, "<TD")
		fontCount := strings.Count(result, navFontOpen)
		assert.Equal(t, cellOpenCount, fontCount,
			"every nav cell (%d) should have exactly one muted <FONT> wrapper (%d)", cellOpenCount, fontCount)
	})

	// WR-01 (03-04): a URL containing '&' must be HTML-escaped when embedded in
	// the HREF="..." attribute. url.PathEscape does NOT escape '&' (it is a
	// valid path character), and GraphViz's HTML-like label parser rejects a
	// raw '&' that is not part of a valid entity, silently dropping the nav
	// anchor. The breadcrumb TD HREF must therefore emit '&amp;'.
	t.Run("breadcrumb URL with ampersand is HTML-escaped in HREF (WR-01)", func(t *testing.T) {
		nav := &graph.Navigation{
			Breadcrumbs: []graph.BreadcrumbItem{
				{Name: "R&D", URL: "../r&d.svg"},
				{Name: "Current", URL: ""},
			},
		}
		result := render.BuildNavigationLabel(nav)
		assert.Contains(t, result, `HREF="../r&amp;d.svg"`,
			"the HREF attribute value must be HTML-escaped (ALIGN may follow)")
		assert.NotContains(t, result, `HREF="../r&d.svg"`)
	})

	// WR-01 (03-04): already URL-encoded segments (%20, %28) must be left
	// untouched by the HTML-escape — html.EscapeString alters only <, >, &, ',
	// ", so a legitimately encoded URL is unaffected.
	t.Run("already-encoded URL is unaffected by HTML-escape (WR-01)", func(t *testing.T) {
		nav := &graph.Navigation{
			Breadcrumbs: []graph.BreadcrumbItem{
				{Name: "API", URL: "api%20(v2).svg"},
				{Name: "Current", URL: ""},
			},
		}
		result := render.BuildNavigationLabel(nav)
		assert.Contains(t, result, `HREF="api%20(v2).svg"`,
			"already-encoded URL segments must survive HTML-escape (ALIGN may follow)")
	})
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestNavigationInGraphOutput(t *testing.T) {
	t.Run("C2/C3 diagram with navigation has breadcrumb in label", func(t *testing.T) {
		g := &graph.Graph{
			Title:     "API Container",
			Direction: "TB",
			Navigation: &graph.Navigation{
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
		// Breadcrumb ancestor name is present (no more "Back to X")
		assert.Contains(t, string(output), "Main System")
		assert.NotContains(t, string(output), "Back to")
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
		assert.NotContains(t, dotStr, navFontOpen)
	})
}
