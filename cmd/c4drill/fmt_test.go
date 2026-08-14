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

func TestFMTFormatsC4DInPlace(t *testing.T) {
	path := writeTempFile(t, "messy.c4d", fmtC4DMessy)

	err, _ := execFMT(t, path)
	require.NoError(t, err, "fmt rewrites a misformatted .c4d in place")

	doc, docErr := c4d.ParseAST([]byte(fmtC4DMessy))
	require.NoError(t, docErr, "fixture parses to an AST")

	want := c4d.EmitC4D(doc)

	got, readErr := os.ReadFile(path)
	require.NoError(t, readErr, "read the formatted file back")

	assert.Equal(t, want, string(got),
		".c4d output equals the EmitC4D canonical text (D-33)")
	assert.Contains(t, string(got), "# Lead comment on user.",
		"lead comments survive .c4d formatting (D-32)")
	assert.Contains(t, string(got), "# tail comment",
		"same-line tail comments survive .c4d formatting (D-32)")
}

func TestFMTFormatsTOMLInPlace(t *testing.T) {
	path := writeTempFile(t, "messy.toml", fmtTOMLMessy)

	err, _ := execFMT(t, path)
	require.NoError(t, err, "fmt rewrites a misformatted .toml in place")

	want, fmtErr := tomlfmt.Format([]byte(fmtTOMLMessy))
	require.NoError(t, fmtErr, "fixture formats")

	got, readErr := os.ReadFile(path)
	require.NoError(t, readErr, "read the formatted file back")

	assert.Equal(t, string(want), string(got),
		".toml output equals the tomlfmt canonical text")
	assert.Contains(t, string(got), "# Header comment.",
		"header comments survive .toml formatting (D-32)")
	assert.Contains(t, string(got), "# trailing on color",
		"trailing comments survive .toml formatting (D-32)")
}

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

	err, _ := execFMT(t, dir)
	require.NoError(t, err, "fmt dir/ walks recursively formatting both formats")

	doc, docErr := c4d.ParseAST([]byte(fmtC4DMessy))
	require.NoError(t, docErr)

	gotC4D, readErr := os.ReadFile(c4dPath)
	require.NoError(t, readErr)
	assert.Equal(t, c4d.EmitC4D(doc), string(gotC4D), "nested-dir .c4d formatted")

	wantTOML, fmtErr := tomlfmt.Format([]byte(fmtTOMLMessy))
	require.NoError(t, fmtErr)

	gotTOML, readErr := os.ReadFile(tomlPath)
	require.NoError(t, readErr)
	assert.Equal(t, string(wantTOML), string(gotTOML), "nested .toml formatted")

	gotTxt, readErr := os.ReadFile(txtPath)
	require.NoError(t, readErr)
	assert.Equal(t, "not a diagram\n", string(gotTxt),
		"non-matching extensions are untouched")

	gotMd, readErr := os.ReadFile(mdPath)
	require.NoError(t, readErr)
	assert.Equal(t, "not a diagram either\n", string(gotMd),
		"non-matching extensions are untouched")
}

func TestFMTCheckMisformattedExitsOne(t *testing.T) {
	path := writeTempFile(t, "messy.c4d", fmtC4DMessy)

	before, err := os.ReadFile(path)
	require.NoError(t, err, "read pre-check bytes")

	runErr, buf := execFMT(t, "--check", path)
	require.Error(t, runErr, "--check exits 1 on a misformatted file (D-31)")
	assert.Contains(t, buf.String(), path,
		"--check prints the offending file path (T-35-08-04)")

	after, err := os.ReadFile(path)
	require.NoError(t, err, "read post-check bytes")
	assert.Equal(t, string(before), string(after),
		"--check writes NOTHING (zero byte change)")
}

func TestFMTCheckFormattedExitsZero(t *testing.T) {
	path := writeTempFile(t, "clean.c4d", fmtC4DMessy)

	err, _ := execFMT(t, path)
	require.NoError(t, err, "format for real first")

	runErr, buf := execFMT(t, "--check", path)
	require.NoError(t, runErr, "--check exits 0 on a formatted file")
	assert.Empty(t, buf.String(), "--check is silent when clean")
}

func TestFMTMultipleArgs(t *testing.T) {
	dir := t.TempDir()
	c4dPath := filepath.Join(dir, "a.c4d")
	tomlPath := filepath.Join(dir, "b.toml")

	require.NoError(t, os.WriteFile(c4dPath, []byte(fmtC4DMessy), 0o600), "write a.c4d")
	require.NoError(t, os.WriteFile(tomlPath, []byte(fmtTOMLMessy), 0o600), "write b.toml")

	err, _ := execFMT(t, c4dPath, tomlPath)
	require.NoError(t, err, "multiple file args accepted in one invocation")

	doc, docErr := c4d.ParseAST([]byte(fmtC4DMessy))
	require.NoError(t, docErr)

	gotC4D, readErr := os.ReadFile(c4dPath)
	require.NoError(t, readErr)
	assert.Equal(t, c4d.EmitC4D(doc), string(gotC4D), "first arg formatted")

	wantTOML, fmtErr := tomlfmt.Format([]byte(fmtTOMLMessy))
	require.NoError(t, fmtErr)

	gotTOML, readErr := os.ReadFile(tomlPath)
	require.NoError(t, readErr)
	assert.Equal(t, string(wantTOML), string(gotTOML), "second arg formatted")
}

func TestFMTUnknownExtensionDirectFile(t *testing.T) {
	path := writeTempFile(t, "notes.txt", "not a diagram\n")

	before, err := os.ReadFile(path)
	require.NoError(t, err)

	runErr, _ := execFMT(t, path)
	require.Error(t, runErr,
		"fmt cannot be pointed at arbitrary file types (T-35-08-02)")
	assert.Contains(t, runErr.Error(), ".toml", "error names the accepted extensions")
	assert.Contains(t, runErr.Error(), ".c4d", "error names the accepted extensions")

	after, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, string(before), string(after), "file untouched")
}

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

	before, err := os.ReadFile(path)
	require.NoError(t, err)

	runErr, _ := execFMT(t, path)
	require.Error(t, runErr, "fmt refuses a file that fails its own Model parse")

	after, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, string(before), string(after), "file untouched by the refusal")
}

func TestFMTGateBlocksBrokenRewrite(t *testing.T) {
	path := writeTempFile(t, "gate.toml", fmtTOMLMessy)

	before, err := os.ReadFile(path)
	require.NoError(t, err)

	orig, parseErr := parser.Parse(before)
	require.NoError(t, parseErr, "gate baseline parses")

	// A candidate output that does not re-parse: the gate must refuse the
	// rewrite before any byte touches the file (T-35-08-01).
	_, gateErr := applyFormatted(path, before, []byte("key = "), orig, parser.Parse)
	require.Error(t, gateErr, "gate refuses a candidate that fails to re-parse")

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, string(before), string(got), "file untouched by the refused rewrite")

	// A candidate that parses to a DIFFERENT Model: also refused — fmt can
	// never change semantics, only layout.
	changed := []byte("[other]\nname = \"Different\"\n")
	_, gateErr = applyFormatted(path, before, changed, orig, parser.Parse)
	require.Error(t, gateErr, "gate refuses a candidate whose Model differs")

	got, err = os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, string(before), string(got), "file untouched by the refused rewrite")
}

func TestFMTNoMatchingTargets(t *testing.T) {
	err, _ := execFMT(t, t.TempDir())
	require.Error(t, err,
		"fmt on a directory with no diagram files errors loudly")
}

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

// execFMT executes the cobra root command with the given fmt args, returning
// the resulting error and the captured output buffer.
func execFMT(t *testing.T, args ...string) (error, *bytes.Buffer) {
	t.Helper()

	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(append([]string{"fmt"}, args...))

	return cmd.Execute(), buf //nolint:wrapcheck // test returns the command error verbatim
}
