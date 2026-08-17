package include_test

import (
	"path/filepath"
	"testing"

	"github.com/Djarvur/c4drill/internal/c4d"
	"github.com/Djarvur/c4drill/internal/include"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parseEntryAny parses an entry file by extension (.c4d through the C4D
// front-end, everything else through the TOML front-end) and resolves its
// include graph — the D-27 dispatch the CLI will carry, mirrored here so
// resolver tests can enter through either format.
//
// Mixed-format include graphs (D-26, Plan 35-05 Task 3): resolveOne
// dispatches on the INCLUDED file's extension — .toml through the TOML
// front-end, .c4d through the C4D front-end, anything else fails closed
// naming the accepted extensions (T-35-05-01). Cycle detection, depth cap
// and visited-set dedup apply unchanged across the mixed graph.
//
//nolint:wrapcheck // the resolver returns *parser.ParseError; tests inspect it unwrapped
func parseEntryAny(t *testing.T, entryPath string) (*parser.Model, error) {
	t.Helper()

	var (
		entry *parser.Model
		err   error
	)

	switch filepath.Ext(entryPath) {
	case ".c4d":
		entry, err = c4d.ParseFile(entryPath)
	default:
		entry, err = parser.ParseFile(entryPath)
	}

	require.NoError(t, err, "entry parse(%s) should not error", entryPath)

	return include.Resolve(entry, filepath.Dir(entryPath), entryPath)
}

// TestResolveTomlIncludesC4d (D-26): a .toml entry including a .c4d file
// resolves — the included file parses through the C4D front-end and merges
// at Model level.
func TestResolveTomlIncludesC4d(t *testing.T) {
	t.Parallel()

	merged, err := parseEntryAny(t, fixturePath("mixed_main.toml"))
	require.NoError(t, err, "mixed .toml -> .c4d graph should resolve")

	assert.Contains(t, merged.Units, "entrySvc", "entry unit preserved")
	assert.Contains(t, merged.Units, "sharedDb", "C4D-included unit merged in")
	require.Len(t, merged.UnitOrder, 2, "UnitOrder")
	assert.Equal(t, []string{"entrySvc", "sharedDb"}, merged.UnitOrder,
		"entry units first, included units appended (D-09 append)")
	assert.Empty(t, merged.Includes, "Includes drained after resolution")

	shared := merged.Units["sharedDb"]
	require.NotNil(t, shared, "sharedDb")
	assert.Equal(t, "Shared DB", shared.Name, "sharedDb parsed by the C4D front-end")
}

// TestResolveC4dIncludesToml (D-26): a .c4d entry including a .toml file
// resolves — the mirror direction of TestResolveTomlIncludesC4d.
func TestResolveC4dIncludesToml(t *testing.T) {
	t.Parallel()

	merged, err := parseEntryAny(t, fixturePath("mixed_main.c4d"))
	require.NoError(t, err, "mixed .c4d -> .toml graph should resolve")

	assert.Contains(t, merged.Units, "entrySvc", "entry unit preserved")
	assert.Contains(t, merged.Units, "sharedDb", "TOML-included unit merged in")
	require.Len(t, merged.UnitOrder, 2, "UnitOrder")
	assert.Equal(t, []string{"entrySvc", "sharedDb"}, merged.UnitOrder,
		"entry units first, included units appended (D-09 append)")
	assert.Empty(t, merged.Includes, "Includes drained after resolution")
}

// TestResolveUnknownExtensionHardError (T-35-05-01): an include whose
// extension is neither .toml nor .c4d is a hard error naming the accepted
// extensions — no fallback parsing, no content sniffing.
func TestResolveUnknownExtensionHardError(t *testing.T) {
	t.Parallel()

	dir := writeFiles(t, map[string]string{
		"main.toml": "[[include]]\npath = \"data.json\"\n",
		"data.json": `{"not": "a diagram"}`,
	})

	_, err := parseEntryAny(t, filepath.Join(dir, "main.toml"))
	pe := requireParseError(t, err)

	assert.Contains(t, pe.Message, "unsupported", "message names the failure")
	assert.Contains(t, pe.Message, ".toml", "message names accepted extension .toml")
	assert.Contains(t, pe.Message, ".c4d", "message names accepted extension .c4d")
	assert.Contains(t, pe.Message, ".json", "message names the offending extension")
}

// TestResolveMixedFormatCycleFatal (INC-04 across formats, D-26): a cycle
// whose hops alternate .c4d and .toml is still fatal with the chain message
// naming both files.
func TestResolveMixedFormatCycleFatal(t *testing.T) {
	t.Parallel()

	dir := writeFiles(t, map[string]string{
		"cycle_a.c4d":  "include ./cycle_b.toml\n",
		"cycle_b.toml": "[[include]]\npath = \"./cycle_a.c4d\"\n",
	})

	_, err := parseEntryAny(t, filepath.Join(dir, "cycle_a.c4d"))
	pe := requireParseError(t, err)

	assert.Contains(t, pe.Message, "include cycle detected", "cycle message")
	assert.Contains(t, pe.Message, "cycle_a.c4d", "chain names the .c4d hop")
	assert.Contains(t, pe.Message, "cycle_b.toml", "chain names the .toml hop")
}
