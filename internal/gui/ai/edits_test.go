// edits_test.go pins the structured edit-proposal pipeline: fence parsing
// (single, multiple, invalid, unterminated), scope validation, and the LCS
// line diff.

package ai_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/gui/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errNotFound is the fake reader's miss error (static: lint-visible).
var errNotFound = errors.New("not found")

// reader over a fake project: demo.toml exists, nested/extra.toml exists.
func fakeReader(t *testing.T) func(string) (string, error) {
	t.Helper()

	files := map[string]string{
		"demo.toml":         "[properties]\nname = \"Old\"\n",
		"nested/extra.toml": "[extra]\nname = \"Extra\"\n",
	}

	return func(rel string) (string, error) {
		if content, ok := files[rel]; ok {
			return content, nil
		}

		if rel == "created.toml" {
			return "", nil // new file creation
		}

		return "", errNotFound
	}
}

func TestParseProposalsSingle(t *testing.T) {
	t.Parallel()

	reply := "Here is my proposal:\n" +
		"```c4drill-edit path=demo.toml\n" +
		"[properties]\nname = \"New\"\n" +
		"```\n" +
		"I renamed the model."

	proposals := ai.ParseProposals(reply, fakeReader(t))
	require.Len(t, proposals, 1)

	p := proposals[0]
	assert.True(t, p.Valid)
	assert.Empty(t, p.Error)
	assert.Equal(t, "demo.toml", p.Path)
	assert.Equal(t, "[properties]\nname = \"New\"", p.NewContent)
	assert.Contains(t, p.OldContent, "Old")
}

func TestParseProposalsMultipleAndCreation(t *testing.T) {
	t.Parallel()

	reply := strings.Join([]string{
		"```c4drill-edit path=demo.toml",
		"name = \"A\"",
		"```",
		"prose in between",
		"```c4drill-edit path=created.toml",
		"[newUnit]",
		"type = \"system\"",
		"```",
	}, "\n")

	proposals := ai.ParseProposals(reply, fakeReader(t))
	require.Len(t, proposals, 2)

	assert.Equal(t, "demo.toml", proposals[0].Path)
	assert.Equal(t, "created.toml", proposals[1].Path)
	assert.Empty(t, proposals[1].OldContent, "creation proposals have no old content")
	assert.True(t, proposals[1].Valid)
}

func TestParseProposalsRejectsBadScopes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
	}{
		{"traversal", "path=../outside.toml"},
		{"absolute", "path=/etc/passwd"},
		{"not a model file", "path=notes.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reply := "```c4drill-edit " + tt.path + "\ncontent\n```"

			proposals := ai.ParseProposals(reply, fakeReader(t))
			require.Len(t, proposals, 1)
			assert.False(t, proposals[0].Valid)
			assert.NotEmpty(t, proposals[0].Error)
		})
	}
}

func TestParseProposalsIgnoresProse(t *testing.T) {
	t.Parallel()

	reply := "No edits here.\n```toml\nname = \"x\"\n```\nand an unterminated one:\n```c4drill-edit path=demo.toml\noops"

	assert.Empty(t, ai.ParseProposals(reply, fakeReader(t)))
}

func TestLineDiff(t *testing.T) {
	t.Parallel()

	oldText := "a\nb\nc\n"
	newText := "a\nB\nc\nd\n"

	diff := ai.LineDiff(oldText, newText)

	kinds := make([]string, 0, len(diff))
	for _, l := range diff {
		kinds = append(kinds, l.Kind)
	}

	assert.Equal(t, []string{"ctx", "del", "add", "ctx", "add"}, kinds)

	texts := make([]string, 0, len(diff))
	for _, l := range diff {
		texts = append(texts, l.Text)
	}

	assert.Equal(t, []string{"a", "b", "B", "c", "d"}, texts)
	assert.True(t, ai.HasChanges(diff))

	assert.False(t, ai.HasChanges(ai.LineDiff("same\n", "same\n")))
	assert.True(t, ai.HasChanges(ai.LineDiff("", "brand new file\n")))
}
