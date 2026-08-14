package main

// Tests for the fmt subcommand (Plan 35-08 Task 2): gofmt-style in-place
// formatting of BOTH authoring formats (D-31) — .c4d through the trivia-aware
// AST + EmitC4D canonical printer (D-32/D-33), .toml through internal/tomlfmt
// — with a --check CI mode that writes nothing and exits 1 listing offenders,
// and the T-35-08-01 semantic safety gate: a candidate rewrite that fails to
// re-parse to the original file's Model is a hard error and the file is left
// untouched. All fixtures run in t.TempDir copies — the real corpus is never
// formatted in place by tests.
//
// NOTE: no t.Parallel — cobra flags bind package-level vars (convert_test.go
// precedent); these tests never render (the WASM constraint does not apply).

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/Djarvur/c4drill/internal/c4d"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/tomlfmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fmtC4DMessy is a deliberately misformatted .c4d document: leading
// indentation, a multi-line-authored leaf, non-canonical field order and no
// blank lines between units — with a lead comment and a same-line tail
// comment that must survive verbatim (D-32).
const fmtC4DMessy = `# Lead comment on user.
		user: person "User" {
	-> api: HTTPS | drives traffic
}
api: system "API" {
			cache: containerDb "Cache" {
				-> bus: Redis
			}
	<- user: requests # tail comment
}
bus: queue "Bus" {
	technology: Kafka
	description: events
}
`

// fmtTOMLMessy is a deliberately misformatted TOML document (the Task 1
// messy style) with comments that must survive (D-32).
const fmtTOMLMessy = `# Header comment.

   [ properties ]
	name   =  "Demo"
color = "#FAFAFA"  # trailing on color

[ user ]
type = "person"
description = "End user"
`

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestFMTFormatsC4DInPlace(t *testing.T) {
	path := writeTempFile(t, "messy.c4d", fmtC4DMessy)

	_, err := execFMT(t, path)
	require.NoError(t, err, "fmt rewrites a misformatted .c4d in place")

	doc, docErr := c4d.ParseAST([]byte(fmtC4DMessy))
	require.NoError(t, docErr, "fixture parses to an AST")

	assert.Equal(t, c4d.EmitC4D(doc), readFixture(t, path),
		".c4d output equals the EmitC4D canonical text (D-33)")
	assert.Contains(t, readFixture(t, path), "# Lead comment on user.",
		"lead comments survive .c4d formatting (D-32)")
	assert.Contains(t, readFixture(t, path), "# tail comment",
		"same-line tail comments survive .c4d formatting (D-32)")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestFMTFormatsTOMLInPlace(t *testing.T) {
	path := writeTempFile(t, "messy.toml", fmtTOMLMessy)

	_, err := execFMT(t, path)
	require.NoError(t, err, "fmt rewrites a misformatted .toml in place")

	want, fmtErr := tomlfmt.Format([]byte(fmtTOMLMessy))
	require.NoError(t, fmtErr, "fixture formats")

	assert.Equal(t, string(want), readFixture(t, path),
		".toml output equals the tomlfmt canonical text")
	assert.Contains(t, readFixture(t, path), "# Header comment.",
		"header comments survive .toml formatting (D-32)")
	assert.Contains(t, readFixture(t, path), "# trailing on color",
		"trailing comments survive .toml formatting (D-32)")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestFMTWalksDirectories(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "sub")
	require.NoError(t, os.MkdirAll(sub, 0o750), "create nested dir")

	c4dPath := filepath.Join(dir, "a.c4d")
	tomlPath := filepath.Join(sub, "b.toml")
	txtPath := filepath.Join(sub, "c.txt")
	mdPath := filepath.Join(dir, "d.md")

	require.NoError(t, os.WriteFile(c4dPath, []byte(fmtC4DMessy), 0o600), "write a.c4d")
	require.NoError(t, os.WriteFile(tomlPath, []byte(fmtTOMLMessy), 0o600), "write b.toml")
	require.NoError(t, os.WriteFile(txtPath, []byte("not a diagram\n"), 0o600), "write c.txt")
	require.NoError(t, os.WriteFile(mdPath, []byte("not a diagram either\n"), 0o600), "write d.md")

	_, err := execFMT(t, dir)
	require.NoError(t, err, "fmt dir/ walks recursively formatting both formats")

	doc, docErr := c4d.ParseAST([]byte(fmtC4DMessy))
	require.NoError(t, docErr)

	wantTOML, fmtErr := tomlfmt.Format([]byte(fmtTOMLMessy))
	require.NoError(t, fmtErr)

	assert.Equal(t, c4d.EmitC4D(doc), readFixture(t, c4dPath),
		"nested-dir .c4d formatted")
	assert.Equal(t, string(wantTOML), readFixture(t, tomlPath),
		"nested .toml formatted")
	assert.Equal(t, "not a diagram\n", readFixture(t, txtPath),
		"non-matching extensions are untouched")
	assert.Equal(t, "not a diagram either\n", readFixture(t, mdPath),
		"non-matching extensions are untouched")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestFMTCheckMisformattedExitsOne(t *testing.T) {
	path := writeTempFile(t, "messy.c4d", fmtC4DMessy)

	before := readFixture(t, path)

	buf, runErr := execFMT(t, "--check", path)
	require.Error(t, runErr, "--check exits 1 on a misformatted file (D-31)")
	assert.Contains(t, buf.String(), path,
		"--check prints the offending file path (T-35-08-04)")

	assert.Equal(t, before, readFixture(t, path),
		"--check writes NOTHING (zero byte change)")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestFMTCheckFormattedExitsZero(t *testing.T) {
	path := writeTempFile(t, "clean.c4d", fmtC4DMessy)

	_, err := execFMT(t, path)
	require.NoError(t, err, "format for real first")

	buf, runErr := execFMT(t, "--check", path)
	require.NoError(t, runErr, "--check exits 0 on a formatted file")
	assert.Empty(t, buf.String(), "--check is silent when clean")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestFMTMultipleArgs(t *testing.T) {
	dir := t.TempDir()
	c4dPath := filepath.Join(dir, "a.c4d")
	tomlPath := filepath.Join(dir, "b.toml")

	require.NoError(t, os.WriteFile(c4dPath, []byte(fmtC4DMessy), 0o600), "write a.c4d")
	require.NoError(t, os.WriteFile(tomlPath, []byte(fmtTOMLMessy), 0o600), "write b.toml")

	_, err := execFMT(t, c4dPath, tomlPath)
	require.NoError(t, err, "multiple file args accepted in one invocation")

	doc, docErr := c4d.ParseAST([]byte(fmtC4DMessy))
	require.NoError(t, docErr)

	wantTOML, fmtErr := tomlfmt.Format([]byte(fmtTOMLMessy))
	require.NoError(t, fmtErr)

	assert.Equal(t, c4d.EmitC4D(doc), readFixture(t, c4dPath), "first arg formatted")
	assert.Equal(t, string(wantTOML), readFixture(t, tomlPath), "second arg formatted")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestFMTUnknownExtensionDirectFile(t *testing.T) {
	path := writeTempFile(t, "notes.txt", "not a diagram\n")

	before := readFixture(t, path)

	_, runErr := execFMT(t, path)
	require.Error(t, runErr,
		"fmt cannot be pointed at arbitrary file types (T-35-08-02)")
	assert.Contains(t, runErr.Error(), ".toml", "error names the accepted extensions")
	assert.Contains(t, runErr.Error(), ".c4d", "error names the accepted extensions")

	assert.Equal(t, before, readFixture(t, path), "file untouched")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestFMTRefusesModelBrokenC4D(t *testing.T) {
	// Grammar-valid but Model-refused (duplicate unit path): fmt's safety
	// gate runs on the ORIGINAL parse — a file that does not survive its own
	// Model parse is a hard error, never a rewrite candidate.
	const src = `a: person "A" {
	-> b
}
a: system "Again" {
}
`

	path := writeTempFile(t, "dup.c4d", src)

	before := readFixture(t, path)

	_, runErr := execFMT(t, path)
	require.Error(t, runErr, "fmt refuses a file that fails its own Model parse")

	assert.Equal(t, before, readFixture(t, path), "file untouched by the refusal")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestFMTGateBlocksBrokenRewrite(t *testing.T) {
	path := writeTempFile(t, "gate.toml", fmtTOMLMessy)

	before := readFixture(t, path)

	orig, parseErr := parser.Parse([]byte(before))
	require.NoError(t, parseErr, "gate baseline parses")

	// A candidate output that does not re-parse: the gate must refuse the
	// rewrite before any byte touches the file (T-35-08-01).
	_, gateErr := applyFormatted(path, []byte(before), []byte("key = "), orig, parser.Parse)
	require.Error(t, gateErr, "gate refuses a candidate that fails to re-parse")

	assert.Equal(t, before, readFixture(t, path),
		"file untouched by the refused rewrite")

	// A candidate that parses to a DIFFERENT Model: also refused — fmt can
	// never change semantics, only layout.
	changed := []byte("[other]\nname = \"Different\"\n")
	_, gateErr = applyFormatted(path, []byte(before), changed, orig, parser.Parse)
	require.Error(t, gateErr, "gate refuses a candidate whose Model differs")

	assert.Equal(t, before, readFixture(t, path),
		"file untouched by the refused rewrite")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestFMTNoMatchingTargets(t *testing.T) {
	_, err := execFMT(t, t.TempDir())
	require.Error(t, err,
		"fmt on a directory with no diagram files errors loudly")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestFMTHelp(t *testing.T) {
	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"fmt", "--help"})

	require.NoError(t, cmd.Execute())
	assert.Contains(t, buf.String(), "fmt")
	assert.Contains(t, buf.String(), "--check")
}

// writeTempFile writes content into a fresh t.TempDir file named name and
// returns its path — tests never format the real corpus fixtures in place.
func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600),
		"write fixture %s", name)

	return path
}

// readFixture reads a formatted fixture back.
func readFixture(t *testing.T, path string) string {
	t.Helper()

	//nolint:gosec // G304: test-created temp path, not user input
	data, err := os.ReadFile(path)
	require.NoError(t, err, "read %s", path)

	return string(data)
}

// execFMT executes the cobra root command with the given fmt args, returning
// the captured output buffer and the resulting error.
func execFMT(t *testing.T, args ...string) (*bytes.Buffer, error) {
	t.Helper()

	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(append([]string{"fmt"}, args...))

	return buf, cmd.Execute() //nolint:wrapcheck // test returns the command error verbatim
}
