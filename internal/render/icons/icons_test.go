package icons

import (
	"strings"
	"testing"

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
		{"person icon", TypePerson, false, "currentColor"},
		{"db icon", TypeDb, false, "currentColor"},
		{"pipe icon", TypePipe, false, "currentColor"},
		{"system icon", TypeSystem, false, "currentColor"},
		{"container icon", TypeContainer, false, "currentColor"},
		{"component icon", TypeComponent, false, "currentColor"},
		{"unknown icon", "unknown", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := GetTemplate(tt.iconType)
			if tt.wantErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Contains(t, got, tt.wantContains)
			assert.Contains(t, got, "svg")
		})
	}
}

func TestColorize(t *testing.T) {
	t.Parallel()

	template := `<svg><circle stroke="currentColor"/></svg>`

	result := Colorize(template, "#3C7FC0")

	assert.Contains(t, result, "#3C7FC0")
	assert.NotContains(t, result, "currentColor")
}

func TestColorizePreservesStructure(t *testing.T) {
	t.Parallel()

	template, err := GetTemplate(TypePerson)
	require.NoError(t, err)

	colored := Colorize(template, "#073B6F")

	// Count elements - should be same before and after
	originalCount := strings.Count(template, "stroke=")
	coloredCount := strings.Count(colored, "stroke=")
	assert.Equal(t, originalCount, coloredCount)
}
