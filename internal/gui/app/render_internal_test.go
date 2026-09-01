// render_internal_test.go pins the preview's target algebra (same package:
// the drill resolution helpers are private). The CLI output-layout inversion
// cases match the hrefs the real renderer emits (verified against
// examples/cloud-system C1/C2/C3 outputs).

package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveDrillTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		basename string
		current  string
		href     string
		want     string
	}{
		// C1 drill-down: full layout path from the root file.
		{"c1 to c3", "cloud-system", "", "cloud-system/amazon/lambdas.svg", "amazon.lambdas"},
		{"c1 to c2", "cloud-system", "", "cloud-system/cloud.svg", "cloud"},
		{"c1 to self-basename", "cloud-system", "", "cloud-system.svg", ""},
		// C2 (file effectively at {basename}/cloud.svg).
		{"c2 back to c1", "cloud-system", "cloud", "../cloud-system.svg", ""},
		{"c2 to sibling c2", "cloud-system", "cloud", "amazon.svg", "amazon"},
		// C3 (file effectively at {basename}/amazon/rds.svg).
		{"c3 back to c1", "cloud-system", "amazon.rds", "../../cloud-system.svg", ""},
		{"c3 up to c2", "cloud-system", "amazon.rds", "../amazon.svg", "amazon"},
		{"c3 to sibling c2", "cloud-system", "amazon.rds", "../cloud.svg", "cloud"},
		// Dotted segments map to nested directories.
		{"deep drill", "demo", "", "demo/a/b/c.svg", "a.b.c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := resolveDrillTarget(tt.basename, tt.current, tt.href)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolveDrillTargetErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		current string
		href    string
	}{
		{"escapes root", "", "../elsewhere.svg"},
		{"escapes from c2", "cloud", "../../elsewhere.svg"},
		{"outside basename", "", "other-model/unit.svg"},
		{"not svg", "", "cloud-system/notes.txt"},
		{"empty href", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := resolveDrillTarget("cloud-system", tt.current, tt.href)
			assert.ErrorIs(t, err, errBadNavigation)
		})
	}
}

func TestIsInternalHref(t *testing.T) {
	t.Parallel()

	assert.True(t, isInternalHref("cloud-system/amazon.svg"))
	assert.True(t, isInternalHref("../cloud-system.svg"))
	assert.False(t, isInternalHref("https://github.com/x/y#L1"))
	assert.False(t, isInternalHref("http://example.com/a.svg"))
	assert.False(t, isInternalHref("mailto:a@b.c"))
	assert.False(t, isInternalHref("#anchor"))
	assert.False(t, isInternalHref(""))
}

// The round trip target→file→target must be the identity.
func TestFileTargetRoundTrip(t *testing.T) {
	t.Parallel()

	for _, target := range []string{"", "cloud", "cloud.ui", "cloud.ui.api"} {
		file := fileOfTarget("demo", target)
		got, ok := targetOfFile("demo", file)
		require.True(t, ok, "file %s", file)
		assert.Equal(t, target, got)
	}
}

func TestBreadcrumbsFor(t *testing.T) {
	t.Parallel()

	crumbs := BreadcrumbsFor("demo", "cloud.ui.api", false)
	require.Len(t, crumbs, 4)
	assert.Equal(t, Breadcrumb{Name: "demo", Target: ""}, crumbs[0])
	assert.Equal(t, Breadcrumb{Name: "cloud", Target: "cloud"}, crumbs[1])
	assert.Equal(t, Breadcrumb{Name: "ui", Target: "cloud.ui"}, crumbs[2])
	assert.Equal(t, Breadcrumb{Name: "api", Target: "cloud.ui.api"}, crumbs[3])

	// Expanded mode and the C1 root both collapse to the single root crumb.
	for _, tc := range []struct {
		target      string
		allExpanded bool
	}{
		{"", false},
		{"cloud.ui", true},
	} {
		crumbs := BreadcrumbsFor("demo", tc.target, tc.allExpanded)
		require.Len(t, crumbs, 1)
		assert.Empty(t, crumbs[0].Target)
	}
}
