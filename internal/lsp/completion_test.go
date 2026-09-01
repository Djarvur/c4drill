// completion_test.go covers textDocument/completion (issue #32 M2): one test
// class per spec item — the 17 context-aware unit types with generic
// promotion, per-unit/per-link fields, enum values, peer (bare + absolute),
// template names and ${param} names, and include paths relative to the
// including file.

package lsp_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Djarvur/c4drill/internal/lsp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// completeAt runs completion on text at (line, char) and returns the labels.
func completeAt(t *testing.T, text string, line, char uint32) []lsp.CompletionItem {
	t.Helper()

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	uri := lsp.DocumentURI("file:///ws/model.toml")
	h.openDoc(uri, text)

	resp := h.request("textDocument/completion", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: uri},
		Position:     lsp.Position{Line: line, Character: char},
	})

	require.Nil(t, resp.Error, "completion must not error: %v", resp.Error)

	if resp.Result == nil {
		return nil
	}

	var list lsp.CompletionList
	require.NoError(t, json.Unmarshal(resp.Result, &list))

	return list.Items
}

func labels(items []lsp.CompletionItem) []string {
	out := make([]string, 0, len(items))
	for _, it := range items {
		out = append(out, it.Label)
	}

	return out
}

// modelWithSections is a fixture exercising the nesting/type context.
const completionModel = `[properties]
name = "Completion Fixture"

[user]
type = "person"
name = "User"
description = "u"

[cloud]
type = "system"
name = "Cloud"
description = "c"

[cloud.db1]
type = "containerDb"
name = "DB"
description = "d"

[cloud.queue1]
type = "containerQueue"
name = "Q"
description = "q"

[cloud.api]
type = "container"
name = "API"
description = "a"

[[cloud.api.link]]
peer = "db1"
description = "uses"

[[include]]
path = "shared.toml"

[template.svc]
params = ["name", "tech"]

[[template.svc.link]]
peer = "db1"
description = "t"

[[use]]
template = "svc"
parent = "cloud"
params = { name = "auth" }
`

func TestCompletionUnitTypesAreLevelAware(t *testing.T) {
	t.Parallel()

	// At C1 (top-level [user] section) the default is "system"; db stays db.
	items := completeAt(t, "[user]\ntype = \"\"\n", 1, 8)

	require.NotEmpty(t, items)
	assert.Len(t, items, 17, "all 17 unit types are offered")

	var first lsp.CompletionItem

	for _, it := range items {
		if it.SortText < first.SortText || first.Label == "" {
			first = it
		}
	}

	assert.Equal(t, "system", first.Label, "C1 default sorts first")
	assert.Contains(t, first.Detail, "default")

	// Inside [cloud] (a system) the default is "container" and db promotes.
	text := "[cloud]\ntype = \"system\"\n\n[cloud.db1]\ntype = \"\"\n"
	items = completeAt(t, text, 4, 8)

	var (
		dbItem        *lsp.CompletionItem
		containerItem *lsp.CompletionItem
	)

	for i := range items {
		switch items[i].Label {
		case "db":
			dbItem = &items[i]
		case "container":
			containerItem = &items[i]
		}
	}

	require.NotNil(t, dbItem)
	require.NotNil(t, containerItem)

	assert.Contains(t, dbItem.Detail, "promotes to containerDb",
		"generic db shows its promotion inside a system")
	assert.Contains(t, containerItem.Detail, "default")

	// Inside a container (C3) db promotes to componentDb.
	text = "[cloud]\ntype = \"container\"\n\n[cloud.svc]\ntype = \"\"\n"
	items = completeAt(t, text, 4, 8)

	for i := range items {
		if items[i].Label == "db" {
			assert.Contains(t, items[i].Detail, "promotes to componentDb")
		}
	}
}

func TestCompletionUnitFields(t *testing.T) {
	t.Parallel()

	// Typing a fresh key inside a unit section: field names, no enums.
	items := completeAt(t, "[web]\ntype = \"system\"\nna\n", 2, 2)

	assert.Contains(t, labels(items), "name")
	assert.Contains(t, labels(items), "technology")
	assert.Contains(t, labels(items), "link")
	assert.NotContains(t, labels(items), "peer", "link fields stay out of unit sections")
}

func TestCompletionLinkFields(t *testing.T) {
	t.Parallel()

	items := completeAt(t, "[web]\ntype = \"system\"\n\n[[web.link]]\npeer = \"a\"\n\n", 5, 0)

	assert.Contains(t, labels(items), "arrow")
	assert.Contains(t, labels(items), "rank")
	assert.Contains(t, labels(items), "kind")
	assert.Contains(t, labels(items), "labelPosition")
	assert.Contains(t, labels(items), "technology")
	assert.NotContains(t, labels(items), "peer", "already authored keys are not re-offered")
}

func TestCompletionEnumValues(t *testing.T) {
	t.Parallel()

	cases := []struct {
		key  string
		want []string
	}{
		{"edges", []string{"straight", "spline", "square", "ortho"}},
		{"arrow", []string{"forward", "reverse", "bidirectional", "none"}},
		{"kind", []string{"read", "write", "read-write"}},
		{"style", []string{"solid", "dashed", "dotted"}},
		{"labelPosition", []string{"middle", "tail", "head"}},
		{"rank", []string{"forward", "reverse", "equal"}},
	}

	for _, tc := range cases {
		t.Run(tc.key, func(t *testing.T) {
			t.Parallel()

			text := "[web]\ntype = \"system\"\n\n[[web.link]]\n" + tc.key + " = \"\"\n"
			items := completeAt(t, text, 4, uint32(len(tc.key)+5)) //nolint:gosec // test-local bounded length

			assert.Equal(t, tc.want, labels(items))
		})
	}
}

func TestCompletionPeerBareAndAbsolute(t *testing.T) {
	t.Parallel()

	text := completionModel + "\n[[cloud.api.link]]\npeer = \"\"\n"
	lineIdx := stringsCount(text) - 1 // the peer line, before the trailing newline
	line := uint32(lineIdx)           //nolint:gosec // test-local bounded index

	items := completeAt(t, text, line, 7)

	got := labels(items)

	// Bare walk-up candidates: cloud's children (db1, queue1, api) and the
	// root scope (user, cloud, properties-free unit names).
	for _, bare := range []string{"db1", "queue1", "user", "cloud"} {
		assert.Contains(t, got, bare, "bare walk-up candidate %s", bare)
	}

	// Absolute dotted paths from the whole document (the host itself is
	// excluded — a unit cannot link to itself).
	for _, abs := range []string{"cloud.db1", "cloud.queue1", "user"} {
		assert.Contains(t, got, abs, "absolute peer path %s", abs)
	}

	assert.NotContains(t, got, "cloud.api", "the host unit is not offered as its own peer")

	// Every item is annotated with its resolution style.
	for _, it := range items {
		if it.Detail != "" {
			assert.True(t,
				it.Detail == "bare peer (walk-up resolution)" || it.Detail == "absolute peer path",
				"unexpected detail %q", it.Detail)
		}
	}
}

// stringsCount is a local line counter (avoids importing strings for one use).
func stringsCount(s string) int {
	n := 0

	for i := range len(s) {
		if s[i] == '\n' {
			n++
		}
	}

	return n
}

func TestCompletionTemplateNamesAndParams(t *testing.T) {
	t.Parallel()

	// template = " inside a [[use]]: the declared template names.
	text := "[template.svc]\nparams = [\"name\", \"tech\"]\n\n[template.other]\nparams = [\"x\"]\n\n" +
		"[[use]]\ntemplate = \"\"\n"
	items := completeAt(t, text, 7, 11)
	assert.Equal(t, []string{"other", "svc"}, labels(items))

	// Inside a ${...} in the template body: the declared params.
	text = "[template.svc]\nparams = [\"name\", \"tech\"]\n\ntechnology = \"${}\"\n"
	items = completeAt(t, text, 3, 16)
	assert.Contains(t, labels(items), "name")
	assert.Contains(t, labels(items), "tech")

	// params = { in a [[use]]: the CHOSEN template's param keys.
	text = "[template.svc]\nparams = [\"name\", \"tech\"]\n\n[template.other]\nparams = [\"x\"]\n\n" +
		"[[use]]\ntemplate = \"svc\"\nparams = {}\n"
	items = completeAt(t, text, 8, 10)

	assert.Equal(t, []string{"name", "tech"}, labels(items),
		"only the instantiated template's params")
}

func TestCompletionIncludePathFromDisk(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "svc.toml"), []byte("[a]\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("x"), 0o600))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "sub"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "sub", "leaf.c4d"), []byte(""), 0o600))

	h := newHarness(t)
	h.request("initialize", lsp.InitializeResult{})
	h.notify("initialized", lsp.InitializedParams{})

	uri := lsp.DocumentURI("file://" + filepath.Join(dir, "main.toml"))
	h.openDoc(uri, "[[include]]\npath = \"\"\n")

	resp := h.request("textDocument/completion", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: uri},
		Position:     lsp.Position{Line: 1, Character: 7},
	})

	require.Nil(t, resp.Error)

	var list lsp.CompletionList
	require.NoError(t, json.Unmarshal(resp.Result, &list))

	got := labels(list.Items)
	assert.Contains(t, got, "svc.toml", "diagram files are offered")
	assert.Contains(t, got, "sub/", "directories are offered with a trailing slash")
	assert.NotContains(t, got, "notes.txt", "non-diagram files are filtered")
	assert.NotContains(t, got, "main.toml", "the including file itself is still listed (harmless)")

	// The sub/ directory's content completes after the slash.
	h.notify("textDocument/didChange", lsp.DidChangeTextDocumentParams{
		TextDocument:   lsp.VersionedTextDocumentIdentifier{URI: uri, Version: 2},
		ContentChanges: []lsp.TextDocumentContentChangeEvent{{Text: "[[include]]\npath = \"sub/\"\n"}},
	})

	resp = h.request("textDocument/completion", lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: uri},
		Position:     lsp.Position{Line: 1, Character: 12},
	})

	require.Nil(t, resp.Error)
	require.NoError(t, json.Unmarshal(resp.Result, &list))
	assert.Contains(t, labels(list.Items), "leaf.c4d", "paths resolve relative to the including file")
}

func TestCompletionTopLevelTables(t *testing.T) {
	t.Parallel()

	items := completeAt(t, "\n", 0, 0)

	assert.Contains(t, labels(items), "[properties]")
	assert.Contains(t, labels(items), "[[include]]")
	assert.Contains(t, labels(items), "[[use]]")
}
