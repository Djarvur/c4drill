package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//nolint:paralleltest // Consistent with project test patterns
func TestWrapText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxChars int
		expected string
	}{
		{
			name:     "fits in one line",
			text:     "hello world",
			maxChars: 20,
			expected: "hello world",
		},
		{
			name:     "word break",
			text:     "hello world",
			maxChars: 5,
			expected: "hello<BR/>world",
		},
		{
			name:     "forced character break",
			text:     "abcdefghij",
			maxChars: 5,
			expected: "abcde<BR/>fghij",
		},
		{
			name:     "short text unchanged",
			text:     "a b c",
			maxChars: 10,
			expected: "a b c",
		},
		{
			name:     "empty string",
			text:     "",
			maxChars: 10,
			expected: "",
		},
		{
			name:     "multi-byte unicode",
			text:     "日本語テスト文字列",
			maxChars: 4,
			expected: "日本語テ<BR/>スト文字<BR/>列",
		},
		{
			name:     "multiple words wrapping",
			text:     "one two three four",
			maxChars: 9,
			expected: "one two<BR/>three<BR/>four",
		},
		{
			name:     "single word exactly maxChars",
			text:     "hello",
			maxChars: 5,
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wrapText(tt.text, tt.maxChars)
			assert.Equal(t, tt.expected, result)
		})
	}
}

//nolint:paralleltest // Consistent with project test patterns
func TestEstimateCharsFromWidth(t *testing.T) {
	tests := []struct {
		name        string
		widthPoints int
		expected    int
	}{
		{
			name:        "100 points",
			widthPoints: 100,
			expected:    12,
		},
		{
			name:        "50 points",
			widthPoints: 50,
			expected:    6,
		},
		{
			name:        "zero points",
			widthPoints: 0,
			expected:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := estimateCharsFromWidth(tt.widthPoints)
			assert.Equal(t, tt.expected, result)
		})
	}
}

//nolint:paralleltest // Consistent with project test patterns
func TestCalculateTextWidth(t *testing.T) {
	tests := []struct {
		name     string
		rowCount int
		ratio    float64
		expected int
	}{
		{
			name:     "3 rows default ratio",
			rowCount: 3,
			ratio:    1.6,
			expected: 50, // 3*18*1.6 - 36 = 86.4 - 36 = 50.4 → 50
		},
		{
			name:     "1 row default ratio",
			rowCount: 1,
			ratio:    1.6,
			expected: 50, // 1*18*1.6 - 36 = 28.8 - 36 = -7.2 → clamped to 50
		},
		{
			name:     "5 rows wide ratio",
			rowCount: 5,
			ratio:    2.0,
			expected: 144, // 5*18*2.0 - 36 = 180 - 36 = 144
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateTextWidth(tt.rowCount, tt.ratio)
			assert.Equal(t, tt.expected, result)
		})
	}
}
