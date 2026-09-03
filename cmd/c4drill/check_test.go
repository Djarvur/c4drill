package main

// Tests for the check subcommand (Phase 41, issue #41): render-free model
// validation. check runs the SAME pipeline front-half as render (parse ->
// includes -> templates -> peers -> validation, D-02) through the shared
// parseValidatedModel helper and reports byte-identical validation output
// (D-03), then stops — no views, no graphviz, no files written (CHECK-01).
// Pins: silent exit-0 success (D-04), render-identical VAL errors including
// the orphan-unit rule (CHECK-02), composed multi-file parity (CHECK-03),
// fail-closed extension dispatch (D-27), and a render-flag-free help surface
// (D-05).
//
// NOTE: no t.Parallel — cobra flags bind package-level vars (convert_test.go
// precedent); these tests never render except the parity baselines, which use
// -o into t.TempDir and stop at validation for the invalid fixture.

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// execCheck executes the cobra root command with the given check args,
// returning the captured output buffer and the resulting error.
func execCheck(t *testing.T, args ...string) (*bytes.Buffer, error) {
	t.Helper()

	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(append([]string{"check"}, args...))

	return buf, cmd.Execute() //nolint:wrapcheck // test returns the command error verbatim
}

// execRender executes the root (render) command — the parity baseline check
// must match. Output is forced into a caller-owned directory so a successful
// render never writes next to the fixtures.
func execRender(t *testing.T, args ...string) (*bytes.Buffer, error) {
	t.Helper()

	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(args)

	return buf, cmd.Execute() //nolint:wrapcheck // test returns the command error verbatim
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestCheckValidModelSilentSuccess(t *testing.T) {
	buf, err := execCheck(t, filepath.Join("testdata", "valid.toml"))

	require.NoError(t, err, "check exits 0 on a valid model")
	assert.Empty(t, buf.String(), "check prints nothing on success (D-04 silent success)")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestCheckOrphanUnitFailsWithValErrors(t *testing.T) {
	buf, err := execCheck(t, filepath.Join("testdata", "check_orphan.toml"))

	require.Error(t, err, "check exits 1 on a validation failure (CHECK-02)")
	assert.True(t, errors.Is(err, errValidationFailed),
		"failure rides the errValidationFailed sentinel — the render exit path")
	assert.Contains(t, buf.String(),
		`error: unit "orphan" has no incoming or outgoing links`,
		"the orphan-unit VAL rule is reported (internal/validator/rules.go)")
	assert.Contains(t, buf.String(), "1 error found",
		"the validator summary line is included (exactly one orphan in the fixture)")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestCheckOrphanRenderParity(t *testing.T) {
	checkBuf, checkErr := execCheck(t, filepath.Join("testdata", "check_orphan.toml"))
	require.Error(t, checkErr, "check exits 1 on the orphan fixture")

	// Render of the SAME fixture fails at the same stage — it stops at
	// validation, writes nothing, and runs no graphviz (stages 3-6 never run).
	renderBuf, renderErr := execRender(t, "-o", t.TempDir(),
		filepath.Join("testdata", "check_orphan.toml"))
	require.Error(t, renderErr, "render fails validation on the same fixture")

	assert.Equal(t, renderBuf.String(), checkBuf.String(),
		"check reports byte-identical validation output to render (CHECK-02/D-03)")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestCheckParseFailureStagePrefixed(t *testing.T) {
	_, err := execCheck(t, filepath.Join("testdata", "invalid.toml"))

	require.Error(t, err, "check exits 1 on a parse failure")
	assert.Contains(t, err.Error(), "parse",
		"stage-prefixed error, render-identical (D-03)")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestCheckComposedFixtureValid(t *testing.T) {
	buf, err := execCheck(t, filepath.Join("testdata", "check_composed_main.toml"))

	require.NoError(t, err,
		"cross-file peer authService resolves only because include.Resolve ran before peer.Resolve (CHECK-03)")
	assert.Empty(t, buf.String(), "silent success (D-04)")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestCheckComposedRenderParity(t *testing.T) {
	checkBuf, checkErr := execCheck(t, filepath.Join("testdata", "check_composed_main.toml"))
	require.NoError(t, checkErr, "check validates the composed model")

	renderBuf, renderErr := execRender(t, "-o", t.TempDir(),
		filepath.Join("testdata", "check_composed_main.toml"))
	require.NoError(t, renderErr, "render succeeds on the same composed model")

	assert.Equal(t, renderBuf.String(), checkBuf.String(),
		"check treats the composed fixture exactly as render does (CHECK-03)")
}

// testdataSnapshot walks cmd/c4drill/testdata and returns a path -> bytes map
// of every file, for before/after no-write comparisons (CHECK-01).
func testdataSnapshot(t *testing.T) map[string]string {
	t.Helper()

	snap := map[string]string{}

	walkErr := filepath.WalkDir("testdata", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		//nolint:gosec // G304: test-relative fixture path, not user input
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		snap[path] = string(data)

		return nil
	})
	require.NoError(t, walkErr, "walk testdata")

	return snap
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestCheckWritesNothing(t *testing.T) {
	before := testdataSnapshot(t)

	for _, fixture := range []string{
		filepath.Join("testdata", "valid.toml"),
		filepath.Join("testdata", "check_orphan.toml"),
		filepath.Join("testdata", "check_composed_main.toml"),
	} {
		_, _ = execCheck(t, fixture) // outcome asserted by the dedicated tests
	}

	assert.Equal(t, before, testdataSnapshot(t),
		"check leaves the testdata tree byte-identical — no output files, no output dir required (CHECK-01)")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestCheckUnknownExtensionFailsClosed(t *testing.T) {
	_, err := execCheck(t, "path.with.dots.unknownext")

	require.Error(t, err, "check fails closed on an unknown extension (D-27)")
	assert.Contains(t, err.Error(), "unsupported input extension",
		"the parse-stage dispatch names the failure, matching render")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestCheckRequiresExactlyOneArg(t *testing.T) {
	_, err := execCheck(t)

	require.Error(t, err, "check with no args is an argument error")
	assert.Contains(t, err.Error(), "arg",
		"the failure is cobra's ExactArgs(1) count error, not a file lookup")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestCheckHelpHidesRenderFlags(t *testing.T) {
	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"check", "--help"})

	require.NoError(t, cmd.Execute())

	out := buf.String()
	assert.Contains(t, out, "check", "help documents the check command")
	assert.Contains(t, out, ".toml", "usage names both accepted authoring formats")
	assert.Contains(t, out, ".c4d", "usage names both accepted authoring formats")

	for _, flag := range []string{
		"-o", "--format", "--expanded", "--plain", "--edges",
		"--no-colors", "--no-styles", "--no-length", "--no-rank",
		"--no-labels", "--label-ratio",
	} {
		assert.NotContains(t, out, flag,
			"render flag %s is hidden from check help (D-05)", flag)
	}
}
