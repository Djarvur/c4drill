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
			name:     "over-budget word stays unsplit",
			text:     "abcdefghij",
			maxChars: 5,
			expected: "abcdefghij",
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
			expected: "日本語テスト文字列",
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
		{
			name:     "hyphen break",
			text:     "Multi-Consumer Broadcast",
			maxChars: 10,
			expected: "Multi-<BR/>Consumer<BR/>Broadcast",
		},
		{
			name:     "arrow and colon break",
			text:     "YUV420->EXTERNAL:TOKEN",
			maxChars: 12,
			expected: "YUV420-><BR/>EXTERNAL:<BR/>TOKEN",
		},
		{
			name:     "underscore break",
			text:     "IMAGE_NATIVE_PROCESSED",
			maxChars: 9,
			expected: "IMAGE_<BR/>NATIVE_<BR/>PROCESSED",
		},
		{
			name:     "bracket attaches to following word",
			text:     "[CGF Channel]",
			maxChars: 9,
			expected: "[CGF<BR/>Channel]",
		},
		{
			name:     "punctuation-rejoin no space",
			text:     "foo-bar",
			maxChars: 20,
			expected: "foo-bar",
		},
		{
			name:     "over-budget letter run still unsplit",
			text:     "abcdefghijkl",
			maxChars: 5,
			expected: "abcdefghijkl",
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
func TestLabelMaxCharsForText(t *testing.T) {
	tests := []struct {
		name      string
		fixedRows int
		textLen   int
		ratio     float64
		expected  int
	}{
		{
			name:      "long description widens the label",
			fixedRows: 1,
			textLen:   110,
			ratio:     1.6,
			expected:  21, // width 174.25 → 21.78 chars
		},
		{
			name:      "medium description, name+tech rows",
			fixedRows: 2,
			textLen:   38,
			ratio:     1.6,
			expected:  15, // width 126.7 → 15.8 chars
		},
		{
			name:      "short description, no fixed rows",
			fixedRows: 0,
			textLen:   14,
			ratio:     1.6,
			expected:  7, // width 56.8 → 7.1 chars
		},
		{
			name:      "empty text keeps a floor",
			fixedRows: 1,
			textLen:   0,
			ratio:     1.6,
			expected:  6, // width 28.8 → 3.6 chars → floor 6
		},
		{
			name:      "longer text widens monotonically",
			fixedRows: 1,
			textLen:   200,
			ratio:     1.6,
			expected:  28, // width 229.6 → 28.7 chars — must exceed the 50-char case
		},
		{
			name:      "higher ratio widens",
			fixedRows: 1,
			textLen:   110,
			ratio:     2.0,
			expected:  24, // width 194.9 → 24.4 chars — must exceed the 1.6 case
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := labelMaxCharsForText(tt.fixedRows, tt.textLen, tt.ratio)
			assert.Equal(t, tt.expected, result)
		})
	}
}

//nolint:paralleltest // Consistent with project test patterns
func TestLabelMaxCharsForPerson(t *testing.T) {
	tests := []struct {
		name      string
		fixedRows int
		textLen   int
		expected  int
	}{
		{
			name:      "long actor description widens the text column",
			fixedRows: 0,
			textLen:   70,
			expected:  17, // icon-aware textCol 138.6 → 17.3 chars
		},
		{
			name:      "short actor name keeps a floor",
			fixedRows: 0,
			textLen:   11,
			expected:  10, // textCol 46.2 → 5.8 chars → floor 10 ("Actor A" stays one line)
		},
		{
			name:      "medium description",
			fixedRows: 0,
			textLen:   30,
			expected:  10, // textCol 85.4 → 10.7 chars
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := labelMaxCharsForPerson(tt.fixedRows, tt.textLen)
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
