// text_internal_test.go pins the LSP position ⇄ offset conversion (UTF-16
// columns) and ranged-edit splicing. Internal-state test (same package) per
// the repo's testpackage skip-regexp convention.

package lsp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOffsetForPosition(t *testing.T) {
	t.Parallel()

	// Lines: 0 "ab", 1 "😀c", 2 "" (trailing newline starts an empty line).
	text := "ab\n\U0001F600c\n"

	cases := []struct {
		name string
		pos  Position
		want int
	}{
		{"start", Position{}, 0},
		{"end of line 0", Position{Line: 0, Character: 2}, 2},
		{"start of line 1", Position{Line: 1}, 3},
		// 😀 is one rune (4 bytes) but two UTF-16 units: character 1 is
		// inside it (clamps to its start), character 2 is after it.
		{"inside surrogate pair", Position{Line: 1, Character: 1}, 3},
		{"after surrogate pair", Position{Line: 1, Character: 2}, 7},
		{"end of line 1", Position{Line: 1, Character: 3}, 8},
		{"trailing empty line", Position{Line: 2}, 9},
		{"line past end clamps", Position{Line: 9, Character: 0}, 9},
		{"column past end clamps", Position{Line: 0, Character: 99}, 2},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, offsetForPosition(text, tc.pos))
		})
	}
}

func TestSpliceText(t *testing.T) {
	t.Parallel()

	text := []byte("hello\nworld\n")

	// Replace "world" with "there".
	got := spliceText(text, Range{
		Start: Position{Line: 1, Character: 0},
		End:   Position{Line: 1, Character: 5},
	}, "there")
	assert.Equal(t, "hello\nthere\n", string(got))

	// Insert at a collapsed range.
	got = spliceText(text, Range{
		Start: Position{Line: 0, Character: 5},
		End:   Position{Line: 0, Character: 5},
	}, "!")
	assert.Equal(t, "hello!\nworld\n", string(got))

	// Inverted range degrades to insertion at start.
	got = spliceText(text, Range{
		Start: Position{Line: 1, Character: 3},
		End:   Position{Line: 1, Character: 1},
	}, "X")
	assert.Equal(t, "hello\nworXld\n", string(got))
}
