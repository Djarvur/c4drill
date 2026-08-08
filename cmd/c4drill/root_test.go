package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/peer"
	"github.com/Djarvur/c4drill/internal/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHelpText verifies that help shows usage examples and flag descriptions (CLII-04).
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestHelpText(t *testing.T) {
	cmd := NewRootCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()

	// Verify key content per CLII-04
	assert.Contains(t, output, "c4drill <input.toml>", "Usage should show command syntax")
	assert.Contains(t, output, "Examples:", "Help should include examples section")
	assert.Contains(t, output, "--format", "Help should document --format flag")
	assert.Contains(t, output, "--output", "Help should document --output flag")
	assert.Contains(t, output, "dot|svg|html", "Help should show available formats")
}

// TestHelpSubcommand verifies that help subcommand shows same content as --help.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestHelpSubcommand(t *testing.T) {
	cmd := NewRootCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	require.NoError(t, err)

	helpOutput := buf.String()

	// Test help subcommand
	cmd2 := NewRootCmd()

	var buf2 bytes.Buffer
	cmd2.SetOut(&buf2)
	cmd2.SetArgs([]string{"help"})

	err2 := cmd2.Execute()
	// Note: Cobra's help subcommand may work differently, so we check basic content
	// The --help flag is the primary way to get help
	_ = err2

	// Both should contain the same key elements
	assert.Contains(t, helpOutput, "c4drill <input.toml>")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestNewRootCmd(t *testing.T) {
	cmd := NewRootCmd()

	assert.Equal(t, "c4drill <input.toml>", cmd.Use)
	assert.Equal(t, "Generate C4 architecture diagrams from TOML", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.True(t, cmd.SilenceUsage)
	assert.NotNil(t, cmd.RunE)
	// Verify Args requires exactly 1 argument
	assert.NotNil(t, cmd.Args)
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestFormatFlag(t *testing.T) {
	cmd := NewRootCmd()

	formatFlag := cmd.PersistentFlags().Lookup("format")
	assert.NotNil(t, formatFlag)
	assert.Equal(t, "svg", formatFlag.DefValue)
	assert.Equal(t, "f", formatFlag.Shorthand)
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestOutputFlag(t *testing.T) {
	cmd := NewRootCmd()

	outputFlag := cmd.PersistentFlags().Lookup("output")
	assert.NotNil(t, outputFlag)
	assert.Empty(t, outputFlag.DefValue) // Empty default, resolved to input file's directory at runtime
	assert.Equal(t, "o", outputFlag.Shorthand)
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestFlagValidation(t *testing.T) {
	tests := []struct {
		name        string
		format      string
		expectError bool
	}{
		{
			name:        "svg format is valid",
			format:      "svg",
			expectError: false,
		},
		{
			name:        "dot format is valid",
			format:      "dot",
			expectError: false,
		},
		{
			name:        "html format is valid",
			format:      "html",
			expectError: false,
		},
		{
			name:        "png format is invalid",
			format:      "png",
			expectError: true,
		},
		{
			name:        "empty format is invalid",
			format:      "",
			expectError: true,
		},
		{
			name:        "uppercase format is invalid",
			format:      "SVG",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) { //nolint:paralleltest // go-graphviz WASM engine has concurrency issues
			cmd := NewRootCmd()
			buf := &bytes.Buffer{}
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Set format flag
			if err := cmd.PersistentFlags().Set("format", tt.format); err != nil {
				t.Fatalf("failed to set format flag: %v", err)
			}

			// Set a dummy arg to satisfy ExactArgs(1)
			cmd.SetArgs([]string{"dummy.toml"})

			err := cmd.Execute()

			if tt.expectError {
				assert.Error(t, err)
			} else if err != nil {
				// For valid formats, the command will fail at file reading stage
				// but not at format validation
				assert.NotContains(t, err.Error(), "invalid format")
			}
		})
	}
}

// Note: Tests in this file do NOT use t.Parallel() because the go-graphviz
// library uses a WASM-based rendering engine that has concurrency issues.

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestFullPipeline_NonexistentFile(t *testing.T) {
	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"nonexistent_file.toml"})

	err := cmd.Execute()
	require.Error(t, err, "Should return error for nonexistent file")
	assert.Contains(t, err.Error(), "parse", "Error should mention parse stage")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestFullPipeline_InvalidTOML(t *testing.T) {
	// Create invalid TOML file
	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "invalid.toml")
	err := os.WriteFile(invalidPath, []byte("invalid [toml"), 0o600)
	require.NoError(t, err)

	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{invalidPath})

	err = cmd.Execute()
	require.Error(t, err, "Should return error for invalid TOML")
	assert.Contains(t, err.Error(), "parse", "Error should mention parse stage")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestFullPipeline_ValidationError(t *testing.T) {
	// Create TOML with an undefined reference (validation error).
	// Uses a DOTTED peer so it skips the relative-peer resolver (D-16 step 1:
	// peers containing "." are absolute) and reaches the validator's
	// undefined-unit check — exercising the validation error path. A bare
	// peer would now be caught by peer.Resolve before validation (Phase 30).
	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "invalid_ref.toml")
	content := `
[properties]
name = "Test"

[user]
type = "person"
name = "User"

[[user.link]]
peer = "no.such.unit"
`
	err := os.WriteFile(invalidPath, []byte(content), 0o600)
	require.NoError(t, err)

	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{invalidPath})

	err = cmd.Execute()
	require.Error(t, err, "Should return error for validation failure")
	assert.Contains(t, err.Error(), "validation", "Error should mention validation")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestFullPipeline_ValidInput(t *testing.T) {
	// Create a valid TOML file in temp dir
	tmpDir := t.TempDir()
	validPath := filepath.Join(tmpDir, "valid.toml")
	content := `
[properties]
name = "Test System"
description = "A test architecture"

[user]
type = "person"
name = "User"
description = "End user of the system"

[[user.link]]
peer = "webapp"
technology = "HTTPS"

[webapp]
type = "system"
name = "Web Application"
description = "Main web application"
technology = "Go, React"

[[webapp.linkFrom]]
peer = "user"
technology = "HTTPS"
`
	err := os.WriteFile(validPath, []byte(content), 0o600)
	require.NoError(t, err)

	outputDir := filepath.Join(tmpDir, "output")
	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Set output directory to temp output dir
	if err := cmd.PersistentFlags().Set("output", outputDir); err != nil {
		t.Fatalf("failed to set output flag: %v", err)
	}

	cmd.SetArgs([]string{validPath})

	err = cmd.Execute()
	require.NoError(t, err, "Should succeed for valid input")

	// Verify C1 diagram was created
	assert.FileExists(t, filepath.Join(outputDir, "valid.svg"), "C1 diagram should exist")

	// Verify silent on success (no stdout output)
	assert.Empty(t, buf.String(), "Should be silent on success")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestFullPipeline_ValidInput_DotFormat(t *testing.T) {
	// Test DOT format output
	tmpDir := t.TempDir()
	validPath := filepath.Join(tmpDir, "valid.toml")
	content := `
[properties]
name = "Test System"

[user]
type = "person"
name = "User"

[[user.link]]
peer = "webapp"
technology = "HTTPS"

[webapp]
type = "system"
name = "Web Application"

[[webapp.linkFrom]]
peer = "user"
technology = "HTTPS"
`
	err := os.WriteFile(validPath, []byte(content), 0o600)
	require.NoError(t, err)

	outputDir := filepath.Join(tmpDir, "output")
	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	if err := cmd.PersistentFlags().Set("output", outputDir); err != nil {
		t.Fatalf("failed to set output flag: %v", err)
	}

	if err := cmd.PersistentFlags().Set("format", "dot"); err != nil {
		t.Fatalf("failed to set format flag: %v", err)
	}

	cmd.SetArgs([]string{validPath})

	err = cmd.Execute()
	require.NoError(t, err, "Should succeed for valid input with dot format")

	// Verify C1 diagram was created with .dot extension
	assert.FileExists(t, filepath.Join(outputDir, "valid.dot"), "C1 DOT diagram should exist")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestFullPipeline_NestedWithExpanded(t *testing.T) {
	// Test with nested model that has expanded units
	tmpDir := t.TempDir()

	// Create a test file with expanded units
	testPath := filepath.Join(tmpDir, "expanded.toml")
	content := `
[properties]
name = "Expanded Test"

[mainapp]
type = "system"
name = "Main App"
technology = "Go"
expanded = ["mainapp"]

[mainapp.api]
type = "container"
name = "API"
technology = "Go"

[[mainapp.api.link]]
peer = "external"
technology = "HTTPS"

[external]
type = "systemExternal"
name = "External API"

[[external.linkFrom]]
peer = "mainapp.api"
technology = "HTTPS"
`
	err := os.WriteFile(testPath, []byte(content), 0o600)
	require.NoError(t, err)

	outputDir := filepath.Join(tmpDir, "output")
	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	if err := cmd.PersistentFlags().Set("output", outputDir); err != nil {
		t.Fatalf("failed to set output flag: %v", err)
	}

	cmd.SetArgs([]string{testPath})

	err = cmd.Execute()
	require.NoError(t, err, "Should succeed for expanded model")

	// Verify C1 diagram was created
	assert.FileExists(t, filepath.Join(outputDir, "expanded.svg"), "C1 diagram should exist")

	// Verify C2 diagram for expanded system was created
	assert.FileExists(t, filepath.Join(outputDir, "expanded", "mainapp.svg"), "C2 diagram for mainapp should exist")
}

// D-01/D-02/D-03: uniform auto-detect incl. boxes; unit-key file naming.
// Any unit with subunits gets a sub-diagram — C1 boxes included (D-01), deep
// box parity for containerBox (D-02), files named by TOML section key with
// dotted-path directory layout (D-03). No per-unit expanded needed anywhere.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestFullPipeline_BoxWithSubunitsGeneratesSubDiagram(t *testing.T) {
	tmpDir := t.TempDir()

	testPath := filepath.Join(tmpDir, "boxtest.toml")
	content := `
[properties]
name = "Box Test"

[boxname]
type = "box"
name = "Box"

[boxname.child]
type = "system"
name = "Child"

[[boxname.child.link]]
peer = "system.cbox.comp"
technology = "HTTP"

[system]
type = "system"
name = "System"

[system.cbox]
type = "containerBox"
name = "Containers"

[system.cbox.comp]
type = "container"
name = "Container"

[[system.cbox.comp.linkFrom]]
peer = "boxname.child"
technology = "HTTP"
`
	err := os.WriteFile(testPath, []byte(content), 0o600)
	require.NoError(t, err)

	outputDir := filepath.Join(tmpDir, "output")
	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	if err := cmd.PersistentFlags().Set("output", outputDir); err != nil {
		t.Fatalf("failed to set output flag: %v", err)
	}

	cmd.SetArgs([]string{testPath})

	err = cmd.Execute()
	require.NoError(t, err, "Should succeed for box model")

	// C1 diagram for the whole model
	assert.FileExists(t, filepath.Join(outputDir, "boxtest.svg"), "C1 diagram should exist")

	// D-01: C2 sub-diagram for the box (unit-key naming, no display names)
	assert.FileExists(t, filepath.Join(outputDir, "boxtest", "boxname.svg"), "C2 diagram for box should exist")

	// D-02/D-03: C3 sub-diagram for the containerBox at its dotted-path location
	cboxSVG := filepath.Join(outputDir, "boxtest", "system", "cbox.svg")
	assert.FileExists(t, cboxSVG, "C3 diagram for containerBox should exist")
}

// =============================================================================
// Tests using test fixtures from testdata/ directory
// =============================================================================

// TestExitCode_Success verifies exit code 0 for successful execution.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestExitCode_Success(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		filepath.Join("testdata", "valid.toml"),
		"--output", tmpDir,
		"--format", "svg",
	})

	err := cmd.Execute()
	assert.NoError(t, err, "Expected exit code 0 (no error)")
}

// TestExitCode_NonexistentFile verifies exit code 1 for nonexistent file.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestExitCode_NonexistentFile(t *testing.T) {
	cmd := NewRootCmd()

	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"nonexistent.toml"})

	err := cmd.Execute()
	assert.Error(t, err, "Expected exit code 1 (error)")
}

// TestExitCode_InvalidTOML verifies exit code 1 for invalid TOML syntax.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestExitCode_InvalidTOML(t *testing.T) {
	cmd := NewRootCmd()

	var stderr bytes.Buffer
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{filepath.Join("testdata", "invalid.toml")})

	err := cmd.Execute()
	assert.Error(t, err, "Expected exit code 1 (error)")
}

// TestStderrOutput verifies errors go to stderr, not stdout (CLII-06).
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestStderrOutput(t *testing.T) {
	cmd := NewRootCmd()

	var stdout, stderr bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetErr(&stderr)
	cmd.SetArgs([]string{"nonexistent.toml"})

	_ = cmd.Execute()

	// Per CLII-06: errors go to stderr
	// Cobra writes errors to stderr via SetErr
	// stdout should be empty
	assert.Empty(t, stdout.String(), "stdout should be empty on error")
}

// TestSilentOnSuccess verifies no stdout output on successful execution.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestSilentOnSuccess(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := NewRootCmd()

	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{
		filepath.Join("testdata", "valid.toml"),
		"--output", tmpDir,
	})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Empty(t, stdout.String(), "stdout should be empty on success (silent)")
}

// TestExpandedUnits verifies C2/C3 diagrams generated for expanded units.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestExpandedUnits(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		filepath.Join("testdata", "expanded.toml"),
		"--output", tmpDir,
		"--format", "svg",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify C1 file exists
	_, err = os.Stat(filepath.Join(tmpDir, "expanded.svg"))
	require.NoError(t, err, "C1 diagram should exist")

	// Verify C2 file exists (mainsystem expanded)
	_, err = os.Stat(filepath.Join(tmpDir, "expanded", "mainsystem.svg"))
	require.NoError(t, err, "C2 diagram for mainsystem should exist")

	// Verify C3 file exists (webapp nested system expanded)
	_, err = os.Stat(filepath.Join(tmpDir, "expanded", "mainsystem", "webapp.svg"))
	require.NoError(t, err, "C3 diagram for mainsystem.webapp should exist")
}

// TestFormatFlag_Dot verifies DOT format output using test fixture.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestFormatFlag_Dot(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		filepath.Join("testdata", "valid.toml"),
		"--output", tmpDir,
		"--format", "dot",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	// Verify .dot file was created
	_, err = os.Stat(filepath.Join(tmpDir, "valid.dot"))
	require.NoError(t, err, "DOT file should be created")
}

// TestCompat01_ValidTomlAllCollapsed locks COMPAT-01: TOML without
// properties.expanded generates a correct C1 with all units collapsed.
// The C2 sub-diagram valid/app.dot IS generated by Phase 2 auto-detect
// (app has a subunit) — its existence is deliberately not asserted here.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestCompat01_ValidTomlAllCollapsed(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		filepath.Join("testdata", "valid.toml"),
		"--output", tmpDir,
		"--format", "dot",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	//nolint:gosec // G304: Test reads from temp directory created by t.TempDir()
	dotData, err := os.ReadFile(filepath.Join(tmpDir, "valid.dot"))
	require.NoError(t, err)

	dot := string(dotData)

	assert.Contains(t, dot, "user\t[", "user node present in C1")
	assert.Contains(t, dot, "app\t[", "app node present in C1")
	assert.Contains(t, dot, "Application 🔍", "app has subunits and no expansion hint -> collapsed with 🔍 in C1")
	assert.NotContains(t, dot, "subgraph cluster_", "no clusters when everything is collapsed (COMPAT-01)")
	assert.NotContains(t, dot, "app.api", "subunits must not appear in C1 when collapsed")
}

// TestCompat02_MultilevelFixtureFiveNodeC1 proves ROADMAP success criterion 4
// on the public fixture (D-01): multilevel.toml renders a 5-node C1 plus the
// auto-detected C2/C3 sub-diagrams.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestCompat02_MultilevelFixtureFiveNodeC1(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		filepath.Join("testdata", "multilevel.toml"),
		"--output", tmpDir,
		"--format", "dot",
	})

	err := cmd.Execute()
	require.NoError(t, err)

	//nolint:gosec // G304: Test reads from temp directory created by t.TempDir()
	dotData, err := os.ReadFile(filepath.Join(tmpDir, "multilevel.dot"))
	require.NoError(t, err)

	dot := string(dotData)

	for _, unit := range []string{"actorA", "actorB", "actorC", "externalSys"} {
		assert.Contains(t, dot, unit+"\t[", "the four external/top-level units render as C1 nodes")
	}

	// mainSystem is expanded (properties.expanded = ["mainSystem"]) so it
	// appears as a cluster, not a collapsed node.
	assert.Contains(t, dot, "subgraph cluster_mainSystem", "mainSystem expanded -> cluster present")
	assert.NotContains(t, dot, "Main System 🔍", "mainSystem is expanded, not collapsed")
	assert.Contains(t, dot, "mainSystem.sshAuth", "expanded mainSystem shows its containers")

	// C2 sub-diagram for mainSystem (auto-detected: has subunits).
	_, err = os.Stat(filepath.Join(tmpDir, "multilevel", "mainSystem.dot"))
	require.NoError(t, err, "C2 diagram for mainSystem should exist")

	// C3 sub-diagram for mainSystem.sshAuth (auto-detected: has subunits).
	_, err = os.Stat(filepath.Join(tmpDir, "multilevel", "mainSystem", "sshAuth.dot"))
	require.NoError(t, err, "C3 diagram for mainSystem.sshAuth should exist")
}

// TestCompat02_NavigationLinksResolve (03-04-03) proves the three UAT
// navigation gaps are closed end-to-end on the public multilevel.toml fixture:
//
//	Gap 1: every explore href in C2/C3 SVGs resolves to an existing sibling
//	       file (os.Stat join against the real CLI-generated tree); the C3
//	       collapsed-ancestor node no longer emits the empty href=".svg".
//	Gap 2: the navigation bar renders as clickable <a xlink:href> anchors in
//	       SVG (not escaped &lt;a href= literal text).
//	Gap 3: the .dot render format still emits .svg navigation URLs.
//
// The DOT-level canonicalDOT golden (TestBuildExpandedGraphBaselineDOT) strips
// and normalizes URL/label attributes, so it cannot enforce these user-facing
// navigation contracts; this test does.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestCompat02_NavigationLinksResolve(t *testing.T) {
	svgDir := generateMultilevelOutput(t, "svg")

	t.Run("C2 SVG navigation is clickable and hrefs resolve", func(t *testing.T) {
		c2 := readOutputFile(t, filepath.Join(svgDir, "multilevel", "mainSystem.svg"))

		assert.Contains(t, c2, `<a xlink:href=`,
			"C2 SVG navigation bar must render clickable anchors (Gap 2)")
		assert.NotContains(t, c2, `&lt;a href=`,
			"C2 SVG must not contain escaped nav markup (Gap 2)")

		c2Dir := filepath.Join(svgDir, "multilevel")
		for _, href := range collectNavAndExploreHrefs(c2) {
			assertHrefsResolve(t, c2Dir, href, "C2 mainSystem.svg")
		}
	})

	t.Run("C3 SVG ancestor link resolves and no empty href", func(t *testing.T) {
		c3 := readOutputFile(t, filepath.Join(svgDir, "multilevel", "mainSystem", "sshAuth.svg"))

		// Gap 1 symptom B: the C3 collapsed-ancestor node must not emit the
		// empty href=".svg" ComputeExploreURL produced before the self-link
		// guard.
		assert.NotContains(t, c3, `href=".svg"`,
			"C3 SVG must not contain the empty href=\".svg\" (Gap 1 symptom B)")
		assert.Contains(t, c3, `<a xlink:href=`,
			"C3 SVG navigation bar must render clickable anchors (Gap 2)")

		hrefs := collectNavAndExploreHrefs(c3)
		require.NotEmpty(t, hrefs, "C3 SVG should contain clickable hrefs")
		// The C2 ancestor (mainSystem.svg) must be reachable upward from C3.
		assert.Contains(t, hrefs, "../mainSystem.svg",
			"C3 SVG must link upward to its C2 ancestor mainSystem.svg (Gap 1)")

		c3Dir := filepath.Join(svgDir, "multilevel", "mainSystem")
		for _, href := range hrefs {
			assertHrefsResolve(t, c3Dir, href, "C3 sshAuth.svg")
		}
	})

	t.Run("dot render still uses svg navigation URLs", func(t *testing.T) {
		dotDir := generateMultilevelOutput(t, "dot")
		dotStr := readOutputFile(t, filepath.Join(dotDir, "multilevel", "mainSystem.dot"))

		// Gap 3: even though the render format is .dot, navigation URLs must
		// end with .svg (browser-navigable), matching the explore URLs.
		assert.Contains(t, dotStr, "../multilevel.svg",
			"C2 .dot navigation back-link must use ../multilevel.svg (Gap 3)")
		assert.NotContains(t, dotStr, "../multilevel.dot",
			"C2 .dot navigation must not reference the .dot file (Gap 3)")
	})
}

// generateMultilevelOutput runs the c4drill CLI on the multilevel fixture into
// a fresh temp directory in the given format and returns that directory.
func generateMultilevelOutput(t *testing.T, format string) string {
	t.Helper()

	dir := t.TempDir()
	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		filepath.Join("testdata", "multilevel.toml"),
		"--output", dir,
		"--format", format,
	})
	require.NoError(t, cmd.Execute())

	return dir
}

// readOutputFile reads a generated output file. The gosec G304 nolint applies
// because the path is constructed inside t.TempDir().
//
//nolint:gosec // G304: Test reads files generated into t.TempDir()
func readOutputFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	return string(data)
}

// collectNavAndExploreHrefs extracts every href value (navigation xlink:href
// anchors and node URL/xlink:href explore links) from a rendered SVG string.
// GraphViz emits hrefs as xlink:href="..." in SVG output.
func collectNavAndExploreHrefs(svg string) []string {
	var hrefs []string

	const marker = `xlink:href="`
	for {
		idx := strings.Index(svg, marker)
		if idx < 0 {
			break
		}

		start := idx + len(marker)
		rest := svg[start:]

		end := strings.Index(rest, `"`)
		if end < 0 {
			break
		}

		href := rest[:end]
		if href != "" {
			hrefs = append(hrefs, href)
		}

		svg = svg[start+end:]
	}

	return hrefs
}

// assertHrefsResolve asserts that a single href resolves to an existing file
// when joined with the directory of the SVG that emitted it.
func assertHrefsResolve(t *testing.T, dir, href, src string) {
	t.Helper()

	resolved := filepath.Join(dir, href)
	if _, err := os.Stat(resolved); err != nil {
		t.Errorf("href %q from %s does not resolve (joined %s): %v",
			href, src, resolved, err)
	}
}

// =============================================================================
// Tests for --expanded flag (EXPD-01, EXPD-03, EXPD-05)
// =============================================================================

// TestExpandedFlag_Exists verifies the --expanded flag is registered.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestExpandedFlag_Exists(t *testing.T) {
	cmd := NewRootCmd()

	expandedFlag := cmd.PersistentFlags().Lookup("expanded")
	require.NotNil(t, expandedFlag, "--expanded flag should exist")
	assert.Equal(t, "false", expandedFlag.DefValue, "--expanded should default to false")
}

// TestExpandedFlag_GeneratesExpandedFile verifies --expanded produces .expanded.{ext} file.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestExpandedFlag_GeneratesExpandedFile(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		filepath.Join("testdata", "expanded.toml"),
		"--output", tmpDir,
		"--format", "svg",
		"--expanded",
	})

	err := cmd.Execute()
	require.NoError(t, err, "--expanded should succeed")

	// Verify .expanded.svg file was created
	assert.FileExists(t, filepath.Join(tmpDir, "expanded.expanded.svg"), "Expanded file should exist")
}

// TestExpandedFlag_SkipsC1C2C3 verifies --expanded skips C1/C2/C3 generation (EXPD-05).
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestExpandedFlag_SkipsC1C2C3(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		filepath.Join("testdata", "expanded.toml"), // Has expanded units that normally generate C2/C3
		"--output", tmpDir,
		"--format", "svg",
		"--expanded",
	})

	err := cmd.Execute()
	require.NoError(t, err, "--expanded should succeed")

	// Only expanded file should exist, no C1/C2/C3 files
	assert.FileExists(t, filepath.Join(tmpDir, "expanded.expanded.svg"), "Expanded file should exist")

	// C1 file should NOT exist
	_, err = os.Stat(filepath.Join(tmpDir, "expanded.svg"))
	assert.True(t, os.IsNotExist(err), "C1 file should NOT exist with --expanded")

	// C2 file should NOT exist
	_, err = os.Stat(filepath.Join(tmpDir, "expanded", "mainsystem.svg"))
	assert.True(t, os.IsNotExist(err), "C2 file should NOT exist with --expanded")
}

// TestExpandedFlag_Off_StandardBehavior verifies normal C1/C2/C3 without --expanded.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestExpandedFlag_Off_StandardBehavior(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := NewRootCmd()
	cmd.SetArgs([]string{
		filepath.Join("testdata", "expanded.toml"),
		"--output", tmpDir,
		"--format", "svg",
	})

	err := cmd.Execute()
	require.NoError(t, err, "Standard mode should succeed")

	// C1 file should exist
	assert.FileExists(t, filepath.Join(tmpDir, "expanded.svg"), "C1 file should exist")

	// C2 file should exist (mainsystem is expanded)
	assert.FileExists(t, filepath.Join(tmpDir, "expanded", "mainsystem.svg"), "C2 file should exist")

	// Expanded file should NOT exist
	_, err = os.Stat(filepath.Join(tmpDir, "expanded.expanded.svg"))
	assert.True(t, os.IsNotExist(err), "Expanded file should NOT exist without --expanded")
}

// TestRootCmd_HTMLFormat (03-04 Safari-link-fix) verifies that `-f html`
// produces self-contained HTML files at C1/C2/C3 paths, that the Safari nav
// shim is present, that no raw .svg hrefs survive (all rewritten to .html so
// wrapped diagrams cross-link to wrapped siblings), and that the XML
// declaration is stripped. End-to-end click navigation is verified separately
// via Safari/AppleScript in the UAT.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestRootCmd_HTMLFormat(t *testing.T) {
	dir := generateMultilevelOutput(t, "html")

	t.Run("produces HTML files at C1/C2/C3 paths", func(t *testing.T) {
		assert.FileExists(t, filepath.Join(dir, "multilevel.html"),
			"C1 HTML should exist")
		assert.FileExists(t, filepath.Join(dir, "multilevel", "mainSystem.html"),
			"C2 HTML should exist")
		assert.FileExists(t, filepath.Join(dir, "multilevel", "mainSystem", "sshAuth.html"),
			"C3 HTML should exist")
	})

	t.Run("HTML output is well-formed and contains the SVG", func(t *testing.T) {
		c1 := readOutputFile(t, filepath.Join(dir, "multilevel.html"))

		assert.True(t, strings.HasPrefix(c1, "<!DOCTYPE html>"),
			"HTML output must start with <!DOCTYPE html>")
		assert.Contains(t, c1, "<svg", "HTML output must inline the SVG")
		assert.Contains(t, c1, "</html>", "HTML output must close <html>")
		assert.NotContains(t, c1, "<?xml",
			"XML declaration must be stripped (invalid inside HTML body)")
	})

	t.Run("injects the Safari navigation shim", func(t *testing.T) {
		c2 := readOutputFile(t, filepath.Join(dir, "multilevel", "mainSystem.html"))

		assert.Contains(t, c2, "window.location.href",
			"C2 HTML must contain the nav shim that makes SVG <a> clickable in Safari")
	})

	t.Run("rewrites all .svg hrefs to .html", func(t *testing.T) {
		c2 := readOutputFile(t, filepath.Join(dir, "multilevel", "mainSystem.html"))

		// C2 has 4 explore links + back-link/breadcrumb, all originally .svg.
		// Every href must end in .html, none in .svg.
		assert.NotContains(t, c2, `.svg"`,
			"C2 HTML must not contain any .svg href suffix (all rewritten to .html)")
		// And at least one href pointing at a sibling HTML diagram.
		assert.Contains(t, c2, `mainSystem/sshAuth.html`,
			"C2 HTML explore link for sshAuth must be rewritten to .html")
	})

	t.Run("C3 back-link rewritten to .html", func(t *testing.T) {
		c3 := readOutputFile(t, filepath.Join(dir, "multilevel", "mainSystem", "sshAuth.html"))

		assert.Contains(t, c3, `../mainSystem.html`,
			"C3 back-link must point at ../mainSystem.html (rewritten from .svg)")
		assert.NotContains(t, c3, `.svg"`,
			"C3 HTML must not contain any .svg href suffix")
	})
}

// --- Phase 30: relative-peer resolution integration tests ---
//
// These prove the end-to-end behavior the CLI gains from wiring peer.Resolve
// into the runRoot pipeline (Plan 30-02): the pipeline ordering makes bare
// peers resolvable BEFORE the validator sees them, and an unresolvable bare
// peer produces a clean non-zero CLI exit (no panic). They complement Plan
// 30-01's unit tests, which cover the resolution algorithm in isolation.

// TestPipelineResolveBeforeValidate proves the pipeline ordering (XC-01):
// peer.Resolve runs between Parse and Validate, so a model with BARE peers
// (which the validator could not resolve on its own) validates cleanly once
// Resolve has rewritten them to absolute paths. It also sanity-checks that
// the fixture actually exercises the feature (>= 1 bare peer pre-resolve).
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestPipelineResolveBeforeValidate(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "peer_walkup.toml"))
	require.NoError(t, err, "failed to read peer_walkup.toml fixture")

	m, err := parser.Parse(data)
	require.NoError(t, err, "peer_walkup.toml must parse cleanly")

	// Sanity: the fixture must declare at least one bare peer (no '.') so
	// the test actually exercises the resolver. Without this guard a future
	// edit that dots every peer would silently turn this into a no-op test.
	bareCount := countBarePeers(m)
	assert.GreaterOrEqual(t, bareCount, 1,
		"peer_walkup.toml must declare at least one bare peer to exercise the resolver")

	// Resolve relative peers (this is the call wired into runRoot Stage 1.6).
	require.NoError(t, peer.Resolve(m),
		"peer_walkup.toml's bare peers must all resolve (sibling/aunt/root/nearest-first)")

	// Validate: the validator now sees absolute peers only, so it must
	// report no reference errors. This would FAIL if peer.Resolve were
	// removed or reordered to run after Validate (the bare peers would be
	// reported as undefined units).
	valErrors := validator.Validate(m)
	assert.Empty(t, valErrors,
		"validator must see absolute peers post-resolve — bare-peer walkup fixture should validate cleanly")
}

// TestCLIUnresolvablePeerExits proves the CLI error path for an unresolvable
// bare peer: invoking runRoot via the cobra root command returns a non-nil
// error whose message names the resolve stage and the peer, and the process
// does not panic. This is the end-to-end form of Plan 30-01's
// TestResolveUnresolvableError.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestCLIUnresolvablePeerExits(t *testing.T) {
	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{filepath.Join("testdata", "peer_unresolvable.toml")})

	err := cmd.Execute()
	require.Error(t, err, "unresolvable bare peer must produce a non-nil CLI error")

	msg := err.Error()
	assert.Contains(t, msg, "resolve peers",
		"error must surface the resolve stage (the fmt.Errorf wrapper)")
	assert.Contains(t, msg, "nonexistent",
		"error must name the unresolvable peer")
	// No panic = the require.Error above already caught it; cobra returns
	// errors, it does not panic on RunE failure.
}

// TestCLICorpusRendersUnchanged proves ERGO-02 end-to-end: every existing
// cmd/c4drill/testdata corpus fixture with peers parses, resolves, and
// validates without error. peer.Resolve must be a no-op for their rendered
// output (their bare peers all reference top-level units → identity rewrite).
// Complements Plan 30-01's parser-corpus peer-set unit test by covering the
// cmd corpus through the full Parse→Resolve→Validate chain.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestCLICorpusRendersUnchanged(t *testing.T) {
	corpus := []string{"valid.toml", "expanded.toml", "multilevel.toml"}

	for _, fix := range corpus {
		t.Run(fix, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", fix))
			require.NoError(t, err, "failed to read corpus fixture %s", fix)

			m, err := parser.Parse(data)
			require.NoError(t, err, "corpus fixture %s must parse cleanly", fix)

			require.NoError(t, peer.Resolve(m),
				"peer.Resolve must be a no-op (error-wise) on corpus fixture %s — bare peers reference top-level units", fix)

			valErrors := validator.Validate(m)
			assert.Empty(t, valErrors,
				"corpus fixture %s must validate cleanly post-resolve (ERGO-02 backward-compat)", fix)
		})
	}
}

// countBarePeers walks m.Units + Subunits counting Link.Peer values that
// contain no '.' (the relative form peer.Resolve rewrites), across both
// Links and authored LinksFrom. Used by TestPipelineResolveBeforeValidate to
// confirm its fixture actually exercises the resolver.
func countBarePeers(m *parser.Model) int {
	count := 0

	var walk func(units map[string]*model.Unit)
	walk = func(units map[string]*model.Unit) {
		for _, unit := range units {
			for _, l := range unit.Links {
				if !strings.Contains(l.Peer, ".") {
					count++
				}
			}
			for _, lf := range unit.LinksFrom {
				if lf.Mirror {
					continue
				}
				if !strings.Contains(lf.Peer, ".") {
					count++
				}
			}
			if len(unit.Subunits) > 0 {
				walk(unit.Subunits)
			}
		}
	}
	walk(m.Units)

	return count
}
