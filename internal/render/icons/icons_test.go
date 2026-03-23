package icons_test

import (
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/render/icons"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTemplate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		iconType     string
		wantErr      bool
		wantContains string
	}{
		{"person icon", icons.TypePerson, false, "currentColor"},
		{"db icon", icons.TypeDb, false, "currentColor"},
		{"pipe icon", icons.TypePipe, false, "currentColor"},
		{"system icon", icons.TypeSystem, false, "currentColor"},
		{"container icon", icons.TypeContainer, false, "currentColor"},
		{"component icon", icons.TypeComponent, false, "currentColor"},
		{"unknown icon", "unknown", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := icons.GetTemplate(tt.iconType)
			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Contains(t, got, tt.wantContains)
		})
	}
}

func TestColorize(t *testing.T) {
	t.Parallel()

	template := `<svg><circle stroke="currentColor"/></svg>`

	result := icons.Colorize(template, "#3C7FC0")

	assert.Contains(t, result, "#3C7FC0")
	assert.NotContains(t, result, "currentColor")
}

func TestColorizePreservesStructure(t *testing.T) {
	t.Parallel()

	template, err := icons.GetTemplate(icons.TypePerson)
	require.NoError(t, err)

	colored := icons.Colorize(template, "#073B6F")

	// Count elements - should be same before and after
	originalCount := strings.Count(template, "stroke")
	coloredCount := strings.Count(colored, "stroke")
	assert.Equal(t, originalCount, coloredCount)
}
