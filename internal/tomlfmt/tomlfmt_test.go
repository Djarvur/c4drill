package tomlfmt_test

// Tests for internal/tomlfmt (Plan 35-08 Task 1): the comment-preserving
// TOML formatter (D-31/D-32). The contract under test:
//
//   - comments survive formatting EXACTLY (text + count + order): file
//     headers, comments above tables, same-line trailing comments,
//     standalone comments between groups, and comments inside multi-line
//     values (which ride the verbatim value bytes) — D-32;
//   - formatting normalizes ONLY whitespace/indent/style: statements at
//     column 0, exactly one space around '=', tight table brackets,
//     blank-line runs collapse to one blank line; key ORDER inside tables
//     is the AUTHOR's order (deliberate contrast with convert's D-23
//     canonical order — fmt never reorders);
//   - idempotency: fmt(fmt(x)) == fmt(x) over the full TOML corpus;
//   - malformed TOML is a hard error with NO output bytes (fail closed);
//   - semantics never change: parser.Parse(Format(x)) == parser.Parse(x).

import (
	"cmp"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/tomlfmt"
	"github.com/pelletier/go-toml/v2/unstable"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// commentsFixture exercises every comment position the formatter must
// preserve: a file-header run, comments above tables, a same-line trailing
// comment, a standalone comment between key-values, a comment inside a
// multi-line array value (kept verbatim with the value bytes) and a trailing
// document comment.
const commentsFixture = `# Styling demo header.
# Second header line.

[properties]
name = "Styling Demo"  # trailing on name
# standalone after a key-value
color = "#FAFAFA"

# Person with custom styling.
[admin]
type = "person"
description = "System admin"

[[admin.link]]
peer = "dashboard"
tech = ["HTTPS", # inside array
	"gRPC"]

# Trailing document comment.
`

// messyFixture is deliberately misformatted: leading indentation, padded
// table brackets, stretched '=' spacing, a three-blank-line run, a padded
// trailing comment — with key order NOT in the D-23 canonical order (color
// before name, description after) to pin author-order preservation.
const messyFixture = `# Header comment.

   [ properties ]
	color   =  "#FAFAFA"
name = "Demo"


description = "A demo"   # trailing on description

# Person comment
[ admin ]
type = "person"
	name = "Administrator"

[[ admin.linkFrom ]]
peer = "dashboard"
technology = "SSH"
`

// normalizedGolden is the canonical rendering of messyFixture: column-0
// statements, tight brackets, single space around '=', one blank line per
// authored blank-line group, one space before a trailing comment, and the
// author's key order untouched.
const normalizedGolden = `# Header comment.
[properties]
color = "#FAFAFA"
name = "Demo"

description = "A demo" # trailing on description

# Person comment
[admin]
type = "person"
name = "Administrator"

[[admin.linkFrom]]
peer = "dashboard"
technology = "SSH"
`

func TestFormatPreservesComments(t *testing.T) {
	t.Parallel()

	data := []byte(commentsFixture)
	out, err := tomlfmt.Format(data)
	require.NoError(t, err, "comment-rich fixture formats")

	assert.Equal(t, commentTexts(t, data), commentTexts(t, out),
		"every comment survives formatting (exact text list, in order)")
	assert.Contains(t, string(out), "# inside array",
		"comments inside multi-line values ride the verbatim value bytes")
	assert.NotEmpty(t, commentTexts(t, data), "fixture actually contains comments")
}

func TestFormatNormalizesWhitespaceAndPreservesOrder(t *testing.T) {
	t.Parallel()

	out, err := tomlfmt.Format([]byte(messyFixture))
	require.NoError(t, err, "messy fixture formats")

	assert.Equal(t, normalizedGolden, string(out),
		"formatting normalizes whitespace/indent/style only — never key order")
}

func TestFormatIdempotentOnFixture(t *testing.T) {
	t.Parallel()

	for name, src := range map[string]string{
		"comments": commentsFixture,
		"messy":    messyFixture,
	} {
		once, err := tomlfmt.Format([]byte(src))
		require.NoError(t, err, "format %s fixture", name)

		twice, err := tomlfmt.Format(once)
		require.NoError(t, err, "re-format %s fixture", name)

		assert.Equal(t, string(once), string(twice),
			"fmt(fmt(x)) == fmt(x) on the %s fixture", name)
	}
}

func TestFormatMalformedFailsClosed(t *testing.T) {
	t.Parallel()

	for name, src := range map[string]string{
		"missing value":       "key = ",
		"unclosed table":      "[unclosed",
		"missing key":         "= 1",
		"unterminated string": `key = "unterminated`,
	} {
		out, err := tomlfmt.Format([]byte(src))
		require.Error(t, err, "malformed input (%s) must be a hard error", name)
		assert.Empty(t, out, "no rewritten bytes may escape a failed format (%s)", name)
	}
}

func TestFormatIdempotentOverCorpus(t *testing.T) {
	t.Parallel()

	paths := corpusTOMLPaths(t)
	assert.Greater(t, len(paths), 10, "corpus walk covers more than 10 fixtures")

	for _, p := range paths {
		data, err := os.ReadFile(p)
		require.NoError(t, err, "read corpus fixture %s", p)

		once, err := tomlfmt.Format(data)
		require.NoError(t, err, "format corpus fixture %s", p)

		twice, err := tomlfmt.Format(once)
		require.NoError(t, err, "re-format corpus fixture %s", p)

		assert.Equal(t, string(once), string(twice),
			"fmt is idempotent over corpus fixture %s", p)
	}
}

func TestFormatAlreadyFormattedStable(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(filepath.Join(repoRoot(t),
		"skill", "examples", "04-styling.toml"))
	require.NoError(t, err, "read the comment-rich styling fixture")

	once, err := tomlfmt.Format(data)
	require.NoError(t, err, "format styling fixture")

	twice, err := tomlfmt.Format(once)
	require.NoError(t, err, "re-format styling fixture")

	assert.Equal(t, string(once), string(twice),
		"a formatted document reports no diff on a second format")
	assert.Equal(t, commentTexts(t, data), commentTexts(t, once),
		"the styling fixture's comments all survive")
}

func TestFormatPreservesSemanticsOverCorpus(t *testing.T) {
	t.Parallel()

	for _, p := range corpusTOMLPaths(t) {
		data, err := os.ReadFile(p)
		require.NoError(t, err, "read corpus fixture %s", p)

		formatted, err := tomlfmt.Format(data)
		require.NoError(t, err, "format corpus fixture %s", p)

		m1, err1 := parser.Parse(data)
		m2, err2 := parser.Parse(formatted)

		if err1 != nil {
			// Formatting must preserve REFUSAL too (fail closed on both
			// sides of the hop), never produce output that parses "better".
			require.Error(t, err2,
				"format preserves parse refusal for %s", p)

			continue
		}

		require.NoError(t, err2, "formatted %s still parses", p)
		require.Equal(t, m1, m2,
			"formatting never changes the parsed model for %s", p)
	}
}

// commentTexts extracts every top-level and same-line trailing comment text
// from data, in document order — the same extraction on both sides of a
// format proves nothing was dropped, reordered or rewritten (D-32). The
// go-toml unstable API with KeepComments exposes comments as standalone
// expressions and as the next sibling of the statement they trail.
func commentTexts(t *testing.T, data []byte) []string {
	t.Helper()

	type located struct {
		offset int64
		text   string
	}

	p := &unstable.Parser{KeepComments: true}
	p.Reset(data)

	var found []located

	for p.NextExpression() {
		expr := p.Expression()
		require.NoError(t, p.Error(), "comment extraction parses")

		if expr.Kind == unstable.Comment {
			found = append(found, located{
				int64(expr.Raw.Offset),
				strings.TrimSpace(string(expr.Data)),
			})
		}

		if tail := expr.Next(); tail != nil && tail.Kind == unstable.Comment {
			found = append(found, located{
				int64(tail.Raw.Offset),
				strings.TrimSpace(string(tail.Data)),
			})
		}
	}

	require.NoError(t, p.Error(), "comment extraction parses")

	slices.SortStableFunc(found, func(a, b located) int {
		return cmp.Compare(a.offset, b.offset)
	})

	texts := make([]string, len(found))
	for i, c := range found {
		texts[i] = c.text
	}

	return texts
}

// corpusTOMLPaths lists the full TOML fixture corpus (the 35-06 parity
// walker roots): testdata/, testdata/c4d/ and cmd/c4drill/testdata/ flat,
// plus skill/examples/ recursive (the 09-composed include-graph files are
// included — fmt formats per file, include wiring is irrelevant to
// formatting). filepath.WalkDir never follows symlinked directories.
func corpusTOMLPaths(t *testing.T) []string {
	t.Helper()

	root := repoRoot(t)

	var out []string

	for _, dir := range []string{"testdata", "testdata/c4d", "cmd/c4drill/testdata"} {
		entries, err := os.ReadDir(filepath.Join(root, dir))
		require.NoError(t, err, "read corpus dir %s", dir)

		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".toml") {
				out = append(out, filepath.Join(root, dir, e.Name()))
			}
		}
	}

	err := filepath.WalkDir(filepath.Join(root, "skill", "examples"),
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() || !strings.HasSuffix(d.Name(), ".toml") {
				return nil
			}

			out = append(out, path)

			return nil
		})
	require.NoError(t, err, "walk skill/examples")

	slices.Sort(out)

	return out
}

// repoRoot resolves the repository root from this package's directory.
func repoRoot(t *testing.T) string {
	t.Helper()

	root, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err, "resolve repo root")

	return root
}
