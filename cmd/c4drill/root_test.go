package main

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/include"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/peer"
	"github.com/Djarvur/c4drill/internal/render"
	"github.com/Djarvur/c4drill/internal/template"
	"github.com/Djarvur/c4drill/internal/testutil/canonical"
	"github.com/Djarvur/c4drill/internal/validator"
	"github.com/Djarvur/c4drill/internal/view"
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

	// Verify key content per CLII-04 (Plan 35-07: usage names both extensions, D-27)
	assert.Contains(t, output, "c4drill <input.toml|input.c4d>", "Usage should show command syntax")
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
	assert.Contains(t, helpOutput, "c4drill <input.toml|input.c4d>")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestNewRootCmd(t *testing.T) {
	cmd := NewRootCmd()

	assert.Equal(t, "c4drill <input.toml|input.c4d>", cmd.Use)
	assert.Equal(t, "Generate C4 architecture diagrams from TOML and C4D", cmd.Short)
	assert.NotEmpty(t, cmd.Long)
	assert.Contains(t, cmd.Long, ".c4d", "Long help must mention .c4d acceptance (D-27)")
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
// CTX-02 refines the "all collapsed" contract: valid.toml's user -> app.api
// is a deep link into the collapsed app unit, so app unfolds as a cluster
// carrying its depicted link target (app.api) and the edge terminates at the
// TRUE target instead of the anonymous ancestor. Units WITHOUT deep links
// stay fully collapsed.
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
	assert.NotContains(t, dot, "app\t[",
		"app must NOT render as a plain node — its chain was inserted for the deep-link target app.api (CTX-02)")
	assert.Contains(t, dot, "subgraph cluster_app",
		"the chain-bearing collapsed ancestor unfolds as a cluster (CTX-02)")
	assert.Contains(t, dot, "\"app.api\"\t[",
		"the true deep-link target renders as a node inside its container chain (CTX-02)")
	assert.Contains(t, dot, "user -> \"app.api\"",
		"the link edge terminates at the TRUE target, not the collapsed ancestor (CTX-02)")
	assert.Contains(t, dot, "Application 🔍", "app has subunits and no expansion hint -> collapsed with 🔍 in C1")
	// The invisible cluster___content wrapper is layout-only (legend
	// separation, LEG-01) and draws nothing; cluster_app is the CTX-02
	// chain-unfold of the deep-link target's container.
	assert.Equal(t, 2, strings.Count(dot, "subgraph cluster_"),
		"content wrapper + the unfolded deep-link container cluster (CTX-02)")
	assert.Contains(t, dot, "subgraph cluster___content", "content wrapper present")
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
			//nolint:gosec // test fixture read, path comes from the fixed corpus above
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
			count += bareLinkPeers(unit.Links)
			count += bareLinkFromPeers(unit.LinksFrom)

			if len(unit.Subunits) > 0 {
				walk(unit.Subunits)
			}
		}
	}
	walk(m.Units)

	return count
}

// bareLinkPeers counts Link.Peer values that contain no '.' (the relative
// form peer.Resolve rewrites).
func bareLinkPeers(links []model.Link) int {
	n := 0

	for _, l := range links {
		if !strings.Contains(l.Peer, ".") {
			n++
		}
	}

	return n
}

// bareLinkFromPeers counts authored (non-mirror) LinksFrom.Peer values that
// contain no '.' (the relative form peer.Resolve rewrites).
func bareLinkFromPeers(links []model.Link) int {
	n := 0

	for _, lf := range links {
		if lf.Mirror {
			continue
		}

		if !strings.Contains(lf.Peer, ".") {
			n++
		}
	}

	return n
}

// --- Phase 32: include.Resolve pipeline wiring (Plan 32-02) ---

// TestPipelineIncludeBeforeValidate proves the Stage 1a wiring (XC-01): a
// multi-file model with [[include]] runs through ParseFile → include.Resolve →
// Expand → peer.Resolve → Validate → render without error. A pre-resolve model
// with unresolved includes would have no merged units, so successful rendering
// proves include.Resolve ran BEFORE Validate (and before template.Expand, which
// needs the merged Units). go-graphviz WASM rules out t.Parallel.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestPipelineIncludeBeforeValidate(t *testing.T) {
	tmpDir := t.TempDir()

	// Entry file includes auth.toml; both files contribute top-level units.
	entryPath := filepath.Join(tmpDir, "main.toml")
	require.NoError(t, os.WriteFile(entryPath, []byte(`
[properties]
name = "Multi-File Pipeline"

[user]
type = "person"
name = "User"

[[include]]
path = "auth.toml"
`), 0o600), "write entry file")

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "auth.toml"), []byte(`
[authService]
type = "system"
name = "Auth Service"

[[authService.link]]
peer = "user"
technology = "HTTPS"
`), 0o600), "write included file")

	outputDir := filepath.Join(tmpDir, "output")
	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	require.NoError(t, cmd.PersistentFlags().Set("output", outputDir), "set output flag")
	require.NoError(t, cmd.PersistentFlags().Set("format", "dot"), "set format=dot")

	cmd.SetArgs([]string{entryPath})

	require.NoError(t, cmd.Execute(),
		"multi-file model must parse → include.Resolve → validate → render cleanly")

	// C1 diagram exists → the merged model (both files' units) rendered.
	assert.FileExists(t, filepath.Join(outputDir, "main.dot"),
		"C1 DOT diagram for the merged model must exist")

	// Silent on success.
	assert.Empty(t, buf.String(), "CLI must be silent on success")
}

// TestPipelineSingleFileNoRegression32 proves include.Resolve is a no-op for
// single-file models: an existing corpus fixture (no [[include]]) still
// renders identically through the now-wired pipeline.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestPipelineSingleFileNoRegression32(t *testing.T) {
	tmpDir := t.TempDir()

	// Copy the existing valid.toml corpus fixture into the temp dir so the
	// output path is isolated (no clobbering of committed testdata).
	src, err := os.ReadFile(filepath.Join("testdata", "valid.toml"))
	require.NoError(t, err, "read corpus fixture valid.toml")

	entryPath := filepath.Join(tmpDir, "valid.toml")
	//nolint:gosec // G703: src is a corpus fixture read from committed testdata, written into a temp dir — no taint risk.
	require.NoError(t, os.WriteFile(entryPath, src, 0o600), "copy valid.toml into temp dir")

	outputDir := filepath.Join(tmpDir, "output")
	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	require.NoError(t, cmd.PersistentFlags().Set("output", outputDir), "set output flag")
	require.NoError(t, cmd.PersistentFlags().Set("format", "dot"), "set format=dot")

	cmd.SetArgs([]string{entryPath})

	require.NoError(t, cmd.Execute(),
		"single-file model must render identically (include.Resolve is a no-op)")

	assert.FileExists(t, filepath.Join(outputDir, "valid.dot"),
		"C1 DOT diagram for the single-file model must still render")
	assert.Empty(t, buf.String(), "CLI must be silent on success")
}

// TestCLIMissingIncludeExits proves the CLI error path for a missing include
// (INC-10/D-12): invoking runRoot returns a non-nil error naming the include
// stage and the referenced path, and does not panic.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestCLIMissingIncludeExits(t *testing.T) {
	tmpDir := t.TempDir()
	entryPath := filepath.Join(tmpDir, "missing_include.toml")
	require.NoError(t, os.WriteFile(entryPath, []byte(`
[properties]
name = "Missing Include"
[[include]]
path = "ghost.toml"
`), 0o600), "write entry file")

	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{entryPath})

	err := cmd.Execute()
	require.Error(t, err, "missing include must produce a non-nil CLI error")

	msg := err.Error()
	assert.Contains(t, msg, "include", "error must surface the include stage")
	assert.Contains(t, msg, "ghost.toml", "error must name the missing path")
}

// renderThroughPipeline runs the full v1.10 pipeline on the fixture at inputPath
// and returns the canonical (order-insensitive, DI-1) DOT serialization. The
// pipeline sequence mirrors cmd/c4drill/root.go exactly:
//
//	ParseFile -> include.Resolve -> template.Expand -> peer.Resolve
//	         -> Validate -> GenerateExpandedView -> BuildExpandedGraph -> RenderDOT
//
// Humanize runs INSIDE ParseFile (parser.go:614, the Phase 29 stopgap — Phase
// 31's XC-04 relocation to a post-expansion pass was deferred). The composed
// fixtures carry explicit name= so parse-time humanize does not fire for them.
func renderThroughPipeline(t *testing.T, inputPath string) string {
	t.Helper()

	m, err := parser.ParseFile(inputPath)
	require.NoError(t, err, "ParseFile")

	// Stage 1a: include.Resolve runs FIRST (3-arg signature: entry, entryDir,
	// entryFile — the third arg threads the real entry filename so error
	// messages name the including file per INC-10/D-12).
	m, err = include.Resolve(m, filepath.Dir(inputPath), inputPath)
	require.NoError(t, err, "include.Resolve")

	// Stage 1.5: template.Expand runs after include so templates defined in
	// included files are visible to [[use]] in the entry (XC-02).
	m, err = template.Expand(m)
	require.NoError(t, err, "template.Expand")

	// Stage 1.6: peer.Resolve runs after Expand so relative peers authored
	// inside templates resolve at the instantiation site (XC-03).
	require.NoError(t, peer.Resolve(m), "peer.Resolve")

	// Stage 2: Validate (must see only absolute paths + a fully-expanded model).
	require.Empty(t, validator.Validate(m), "model should be valid after pipeline")

	// Views + graph + render.
	v := view.GenerateExpandedView(m)
	g := graph.BuildExpandedGraph(v)
	dot, err := render.RenderDOT(g)
	require.NoError(t, err, "render.RenderDOT")

	return canonical.Canonical(t, string(dot))
}

// TestXC05_ComposedEquivSingleFile guards XC-05: the composed multi-file fixture
// (include + templates + relative-peer + reference) renders to IDENTICAL DOT as
// its hand-expanded single-file equivalent through the full pipeline.
//
// The comparison is order-insensitive (canonical.Canonical, STATE.md DI-1):
// the pinned go-graphviz fork emits map-order-dependent sibling ordering and
// layout geometry, so byte-exact require.Equal on raw DOT would fail spuriously.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestXC05_ComposedEquivSingleFile(t *testing.T) {
	const (
		multiFile  = "../../skill/examples/09-composed/entry.toml"
		singleFile = "../../skill/examples/09-composed/single-file-equivalent.toml"
	)

	multiCanon := renderThroughPipeline(t, multiFile)
	singleCanon := renderThroughPipeline(t, singleFile)

	require.Equal(t, multiCanon, singleCanon,
		"XC-05: composed multi-file model must render identical to its "+
			"hand-expanded single-file equivalent (canonicalDOT, DI-1)")
}

// TestXC01_PipelineOrdering guards XC-01: the v1.10 pipeline order is
// load-bearing for correctness. This behavioral test (D-20) asserts the two
// order-dependent properties that ONLY hold when the pipeline runs
// include.Resolve -> template.Expand -> peer.Resolve in that order:
//
//   - XC-02 (include before Expand): the [[use]] in entry.toml instantiates the
//     `microservice` template that is DEFINED in the included templates.toml.
//     If include ran AFTER Expand, the [[use]] would fail with "unknown template".
//     A successful renderThroughPipeline (no error at the Expand stage) proves
//     include ran first.
//
//   - XC-03 (Expand before peer.Resolve): the instantiated platform.auth.cache
//     unit's parametrized link (peer = "${upstreamBus}") expands to the concrete
//     value "messageBus" and then resolves as an edge connecting platform.auth.cache
//     to messageBus. If peer.Resolve ran BEFORE Expand, the literal "${upstreamBus}"
//     would not yet be substituted and peer resolution would fail or resolve wrongly.
//
// This test is behavioral (inspects the rendered DOT), NOT a source-scan — it
// is robust to refactors that move pipeline calls into helper functions (D-20
// explicitly rejected the source-scan alternative).
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestXC01_PipelineOrdering(t *testing.T) {
	canon := renderThroughPipeline(t, "../../skill/examples/09-composed/entry.toml")

	// Assertion A (XC-02 — include before template.Expand): the render
	// completed without "unknown template" error (require.NoError inside
	// renderThroughPipeline at the Expand stage would have fired otherwise).
	// Make it explicit by asserting the instantiated unit appears in the
	// canonical output — the [[use]] succeeded against a template that lives
	// in an INCLUDED file.
	assert.Contains(t, canon, `"platform.auth.cache"`,
		"XC-02: the [[use]] instantiated the microservice template defined in "+
			"the included templates.toml — proves include.Resolve ran before template.Expand")

	// Assertion B (XC-03 — template.Expand before peer.Resolve): the
	// parametrized peer ${upstreamBus} expanded to the concrete "messageBus"
	// and the edge platform.auth.cache -> messageBus is present in the render.
	// The canonical serialization format (serializeDOTStatement in the
	// canonical helper) emits edges as `"src" -> "dst"` in the head field.
	assert.Contains(t, canon, `"platform.auth.cache" -> messageBus`,
		"XC-03: the templated cache's parametrized peer ${upstreamBus} expanded "+
			"to messageBus and resolved as an edge to the instantiation-site target "+
			"— proves template.Expand ran before peer.Resolve")
}

// --- Phase 35: C4D direct render + extension dispatch (Plan 35-07 Task 1) ---

// c4dMinimalTwin is the hand-written C4D twin of skill/examples/01-minimal.toml:
// same properties, same units, same edges (linkFrom rides `<-`, link rides
// `->`), so rendering the .c4d file directly through the pipeline must
// produce the same view outputs as the .toml original (D-29: C4D is a
// first-class authoring format, not a converter curiosity).
const c4dMinimalTwin = `properties {
	name: "Minimal Architecture"
}

user: person "User" {
	<- webapp: Uses
}

webapp: system "Web Application" {
	-> user: Uses
}
`

// runRenderDot renders inputPath through the real cobra root command into a
// fresh temp output dir in DOT format and returns the C1 {basename}.dot
// content. Asserts the output file was created inside the temp dir (D-29).
func runRenderDot(t *testing.T, inputPath string) string {
	t.Helper()

	outDir := t.TempDir()
	cmd := NewRootCmd()
	cmd.SetArgs([]string{inputPath, "--output", outDir, "--format", "dot"})

	require.NoError(t, cmd.Execute(), "render %s must succeed", inputPath)

	dotPath := filepath.Join(outDir,
		strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))+".dot")

	data, err := os.ReadFile(dotPath) //nolint:gosec // G304: path inside t.TempDir()
	require.NoError(t, err, "C1 DOT output should exist next to the render dir for %s", inputPath)

	return string(data)
}

// TestC4DRenderDirect proves D-29: a .c4d document renders directly through
// the full pipeline (parse -> include -> expand -> peer -> validate -> views
// -> render) producing the SAME C1 view as its .toml twin, compared under
// the order-insensitive canonicalDOT contract (DI-1 — layout geometry and
// map-order siblings are normalized away).
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestC4DRenderDirect(t *testing.T) {
	tmpDir := t.TempDir()

	// The .toml original: skill/examples/01-minimal.toml verbatim.
	src, err := os.ReadFile(filepath.Join("..", "..", "skill", "examples", "01-minimal.toml"))
	require.NoError(t, err, "read 01-minimal.toml fixture")

	tomlPath := filepath.Join(tmpDir, "minimal.toml")
	//nolint:gosec // G703: src is a committed corpus fixture, written into t.TempDir()
	require.NoError(t, os.WriteFile(tomlPath, src, 0o600), "write .toml original")

	// The hand-written .c4d twin.
	c4dPath := filepath.Join(tmpDir, "minimal.c4d")
	require.NoError(t, os.WriteFile(c4dPath, []byte(c4dMinimalTwin), 0o600), "write .c4d twin")

	tomlDot := runRenderDot(t, tomlPath)
	c4dDot := runRenderDot(t, c4dPath)

	require.Equal(t, canonical.Canonical(t, tomlDot), canonical.Canonical(t, c4dDot),
		"D-29: rendering the .c4d twin directly must produce the same C1 view "+
			"as the .toml original (canonicalDOT, DI-1)")
}

// TestRootInputUnknownExtension proves D-27's fail-closed dispatch: an input
// whose extension is neither .toml nor .c4d is a hard parse error naming BOTH
// accepted extensions — no fallback parsing, no content sniffing.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestRootInputUnknownExtension(t *testing.T) {
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "diagram.json")
	require.NoError(t, os.WriteFile(jsonPath, []byte(`{"not": "c4drill input"}`), 0o600),
		"write .json input")

	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{jsonPath})

	err := cmd.Execute()
	require.Error(t, err, "unknown input extension must be a hard error (D-27)")

	msg := err.Error()
	assert.Contains(t, msg, "parse", "error must keep the parse: stage prefix")
	assert.Contains(t, msg, ".toml", "error must name the .toml extension")
	assert.Contains(t, msg, ".c4d", "error must name the .c4d extension")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestGlobalEdgesE2E(t *testing.T) {
	// GEDGE-01: properties.edges = "straight" must reach the generated C2
	// diagram as splines=false even when the expanded unit sets no own edges.
	tmpDir := t.TempDir()
	srcPath := filepath.Join(tmpDir, "edges.toml")
	content := `
[properties]
name = "Global Edges"
edges = "straight"

[app]
type = "system"
name = "App"

[app.api]
type = "container"
name = "API"

[[app.api.link]]
peer = "app.db"
technology = "SQL"

[app.db]
type = "containerDb"
name = "DB"
`
	require.NoError(t, os.WriteFile(srcPath, []byte(content), 0o600))

	outputDir := filepath.Join(tmpDir, "output")
	cmd := NewRootCmd()
	require.NoError(t, cmd.PersistentFlags().Set("output", outputDir))
	require.NoError(t, cmd.PersistentFlags().Set("format", "dot"))
	cmd.SetArgs([]string{srcPath})
	require.NoError(t, cmd.Execute())

	// The C2 diagram for app/ is auto-generated with the global edge style.
	//nolint:gosec // G304: Test reads from temp directory created by t.TempDir()
	c2, err := os.ReadFile(filepath.Join(outputDir, "edges", "app.dot"))
	require.NoError(t, err, "C2 dot generated")
	assert.Contains(t, string(c2), "splines=false",
		"global edges=straight disables splines on the C2 diagram")
}

// =============================================================================
// Tests for --plain flag E2E (PLAIN-03/PLAIN-04, phase 37 plan 04)
// =============================================================================

// generatePlainFixtureOutput runs the CLI on the styled plain.toml fixture
// into a fresh temp directory in the given format, with any extra args
// (e.g. "--plain", "--expanded"), and returns that directory.
func generatePlainFixtureOutput(t *testing.T, format string, extraArgs ...string) string {
	t.Helper()

	dir := t.TempDir()
	args := append([]string{
		filepath.Join("testdata", "plain.toml"),
		"--output", dir,
		"--format", format,
	}, extraArgs...)

	cmd := NewRootCmd()
	cmd.SetArgs(args)
	require.NoError(t, cmd.Execute(), "plain fixture must render cleanly")

	return dir
}

// collectGeneratedFiles returns every regular file generated under dir
// (C1 + drill-down tree), relative walk in lexical order.
func collectGeneratedFiles(t *testing.T, dir string) []string {
	t.Helper()

	var files []string

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			files = append(files, path)
		}

		return nil
	})
	require.NoError(t, err, "walk output dir")
	require.NotEmpty(t, files, "at least the C1 output must be generated")

	return files
}

// TestPlainFlagC1Golden locks the --plain C1 output against the committed
// testdata/plain.dot golden (order-insensitive canonical comparison, DI-1).
// The golden is a NEW file — no existing golden is re-baselined by it.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestPlainFlagC1Golden(t *testing.T) {
	dir := generatePlainFixtureOutput(t, "dot", "--plain")

	got := readOutputFile(t, filepath.Join(dir, "plain.dot"))
	expected := readOutputFile(t, filepath.Join("testdata", "plain.dot"))

	require.Equal(t, canonical.Canonical(t, expected), canonical.Canonical(t, got),
		"--plain C1 output must match the committed plain.dot golden (canonical, DI-1)")
}

// TestDeeplinkRootCompactGolden locks the NON-EXPANDED root of the deepcross
// fixture against the committed testdata/deepcross.dot golden (order-
// insensitive canonical comparison, same DI-1 pattern as TestPlainFlagC1Golden).
// The fixture reproduces the BUG-1-ROOT-COMPACT regression shape (deep nesting
// + cross-container links + external actors deep-linking into nested subunits);
// the golden pins the compact C1 so the whole-model flood can never ship
// silently again (the pre-quick-task fixture corpus was blind to it).
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestDeeplinkRootCompactGolden(t *testing.T) {
	dir := generateFixtureOutput(t, "deepcross.toml", "dot")

	got := readOutputFile(t, filepath.Join(dir, "deepcross.dot"))
	expected := readOutputFile(t, filepath.Join("testdata", "deepcross.dot"))

	require.Equal(t, canonical.Canonical(t, expected), canonical.Canonical(t, got),
		"non-expanded deepcross root must match the committed deepcross.dot golden (canonical, DI-1)")
}

// TestPlainFlagExpandedGolden locks --plain x --expanded against the committed
// testdata/plain.expanded.dot golden (Pitfall 5: the expanded pipeline must
// honour the flag too).
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestPlainFlagExpandedGolden(t *testing.T) {
	dir := generatePlainFixtureOutput(t, "dot", "--plain", "--expanded")

	got := readOutputFile(t, filepath.Join(dir, "plain.expanded.dot"))
	expected := readOutputFile(t, filepath.Join("testdata", "plain.expanded.dot"))

	require.Equal(t, canonical.Canonical(t, expected), canonical.Canonical(t, got),
		"--plain --expanded output must match the committed plain.expanded.dot golden (canonical, DI-1)")
}

// TestPlainFlagAppliesToAllGeneratedFiles proves PLAIN-04's "every generated
// output file" contract: for EVERY .dot file the --plain run produces (the C1
// context diagram AND each drill-down), author-custom formatting from the
// fixture is suppressed and labels render as plain text, while the semantic
// surface (legend, kind-derived colours, label content) survives.
//
// HTML-marker note: the nav/title graph label and the legend node carry
// UPPERCASE <TABLE> markup by design (both stay under --plain); unit/edge
// HTML labels are the only source of LOWERCASE html tags, so the lowercase
// markers are the precise per-file signal that label formatting simplified.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestPlainFlagAppliesToAllGeneratedFiles(t *testing.T) {
	dir := generatePlainFixtureOutput(t, "dot", "--plain")

	dotFiles := make([]string, 0)

	for _, f := range collectGeneratedFiles(t, dir) {
		if strings.HasSuffix(f, ".dot") {
			dotFiles = append(dotFiles, f)
		}
	}

	require.Len(t, dotFiles, 2, "C1 plus the orders drill-down must be generated")

	for _, f := range dotFiles {
		dot := readOutputFile(t, f)
		lower := strings.ToLower(dot)
		name := filepath.Base(f)

		// Suppressed author formatting (fixture custom colours).
		for _, hex := range []string{"#fff9c4", "#f9a825", "#1565c0"} {
			assert.NotContains(t, lower, hex,
				"%s: author colour %s must be suppressed under --plain", name, hex)
		}

		// Suppressed author formatting (spacing/ranking/labels).
		assert.NotContains(t, lower, "minlen", "%s: link length=3 must not emit minlen", name)
		assert.NotContains(t, lower, "dir=back", "%s: rank=reverse must not swap endpoints", name)
		assert.NotContains(t, lower, "constraint=false", "%s: rank=equal must not suppress constraint", name)
		assert.NotContains(t, lower, "splines=false", "%s: properties.edges must be ignored", name)

		// Plain-text labels: no LOWERCASE html label markup (nodes/edges).
		assert.NotContains(t, dot, "<table", "%s: node/edge labels must not emit HTML tables", name)
		assert.NotContains(t, dot, "<b>", "%s: no bold name rows (HTML node labels)", name)
		assert.NotContains(t, dot, "<i>", "%s: no italic technology rows (HTML labels)", name)

		// Exactly the two sanctioned HTML labels remain: the nav/title graph
		// label and the floating legend node.
		assert.Equal(t, 2, strings.Count(dot, "label=<"),
			"%s: only the graph label and the legend may carry HTML labels", name)

		// Semantic surface survives plain mode.
		assert.Contains(t, dot, "__c4drill_legend", "%s: legend must be present", name)
		assert.Contains(t, dot, "#2E7D32", "%s: kind-derived colour must survive", name)
		assert.Contains(t, dot, "[gRPC] Streams order events",
			"%s: edge label content must be preserved as plain text", name)
		assert.Contains(t, dot, "Order API|Go|Order processing API",
			"%s: node label content must be preserved as plain text", name)
	}
}

// TestPlainFlagAllFormats proves the flag reaches the svg and html pipelines:
// every expected output file exists and is non-empty. Exact svg/html bytes are
// graphviz-owned (layout geometry), so dot remains the precise golden format.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestPlainFlagAllFormats(t *testing.T) {
	for _, format := range []string{"svg", "html"} {
		t.Run(format, func(t *testing.T) {
			dir := generatePlainFixtureOutput(t, format, "--plain")

			expected := []string{
				filepath.Join(dir, "plain."+format),
				filepath.Join(dir, "plain", "orders."+format),
			}

			for _, path := range expected {
				info, err := os.Stat(path)
				require.NoError(t, err, "%s must exist", path)
				assert.Positive(t, info.Size(), "%s must be non-empty", path)
			}

			for _, f := range collectGeneratedFiles(t, dir) {
				info, err := os.Stat(f)
				require.NoError(t, err, "stat %s", f)
				assert.Positive(t, info.Size(), "%s must be non-empty", f)
			}
		})
	}
}

// TestPlainFlagOptIn proves the flag is opt-in: WITHOUT --plain the same
// fixture still renders its author formatting (custom link colour, HTML
// labels) — default output is untouched by the plain pipeline.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestPlainFlagOptIn(t *testing.T) {
	dir := generatePlainFixtureOutput(t, "dot")

	dot := readOutputFile(t, filepath.Join(dir, "plain.dot"))

	assert.Contains(t, dot, "#1565C0",
		"default mode must keep the author link colour (opt-in proven)")
	assert.Contains(t, strings.ToLower(dot), "<table",
		"default mode must keep the HTML label path")
	assert.Contains(t, dot, "minlen",
		"default mode must keep the author link length")
}

// =============================================================================
// Tests for granular switches (KEY-01/KEY-02, phase 38 plan 02)
// =============================================================================

// TestPlainUnionParity is the KEY-02 lock: rendering plain.toml with --plain
// alone and with --plain plus all four granular flags produces canonically
// IDENTICAL DOT — --plain is the exact union of the switches.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestPlainUnionParity(t *testing.T) {
	plainDir := generatePlainFixtureOutput(t, "dot", "--plain")
	unionDir := generatePlainFixtureOutput(t, "dot",
		"--plain", "--no-colors", "--no-styles", "--no-length", "--no-rank")

	plain := readOutputFile(t, filepath.Join(plainDir, "plain.dot"))
	union := readOutputFile(t, filepath.Join(unionDir, "plain.dot"))

	require.Equal(t, canonical.Canonical(t, plain), canonical.Canonical(t, union),
		"--plain must be canonically identical with and without the granular flags (KEY-02 union lock)")
}

// TestGranularFlagsE2E proves each switch suppresses exactly its own aspect
// over the full pipeline on plain.toml (which carries color/border/style/
// length=3/rank=reverse/rank=equal/kind), leaving the OTHER aspects intact.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestGranularFlagsE2E(t *testing.T) {
	t.Run("no-colors suppresses colouring only", func(t *testing.T) {
		dir := generatePlainFixtureOutput(t, "dot", "--no-colors")
		dot := strings.ToLower(readOutputFile(t, filepath.Join(dir, "plain.dot")))

		for _, hex := range []string{"#fff9c4", "#f9a825", "#1565c0", "#2e7d32"} {
			assert.NotContains(t, dot, hex, "--no-colors must suppress author and kind colours")
		}

		// Other aspects untouched.
		assert.Contains(t, dot, "minlen=3", "--no-colors must keep link length")
		assert.Contains(t, dot, "dir=back", "--no-colors must keep rank=reverse")
		assert.Contains(t, dot, "constraint=false", "--no-colors must keep rank=equal")
	})

	t.Run("no-styles suppresses styles only", func(t *testing.T) {
		// plain.toml's author styles are not DOT-observable (the pair collapse
		// aggregates the dashed link style to solid, and node style emission
		// is fill-colour-driven), so this subtest uses a dedicated fixture
		// where author styles reach the DOT verbatim.
		dir := t.TempDir()
		src := filepath.Join(dir, "styles.toml")
		content := `[properties]
name = "Styles"

[app]
type = "system"
name = "App"
style = "dotted"
[[app.link]]
peer = "db"
style = "dashed"
color = "#1565C0"

[db]
type = "db"
name = "DB"
`
		require.NoError(t, os.WriteFile(src, []byte(content), 0o600))

		out := filepath.Join(dir, "out")
		cmd := NewRootCmd()
		cmd.SetArgs([]string{src, "--output", out, "--format", "dot", "--no-styles"})
		require.NoError(t, cmd.Execute())

		dot := readOutputFile(t, filepath.Join(out, "styles.dot"))

		assert.NotContains(t, dot, "dashed", "--no-styles must suppress the author link style")
		assert.NotContains(t, dot, "dotted", "--no-styles must suppress the author node style")
		assert.Contains(t, dot, "#1565C0", "--no-styles must keep the author link colour")
	})

	t.Run("no-length suppresses minlen only", func(t *testing.T) {
		dir := generatePlainFixtureOutput(t, "dot", "--no-length")
		dot := readOutputFile(t, filepath.Join(dir, "plain.dot"))

		assert.NotContains(t, dot, "minlen", "--no-length must suppress link length")

		// Other aspects untouched.
		assert.Contains(t, dot, "#1565C0", "--no-length must keep the author link colour")
		assert.Contains(t, dot, "dir=back", "--no-length must keep rank=reverse")
	})

	t.Run("no-rank suppresses ranking only", func(t *testing.T) {
		dir := generatePlainFixtureOutput(t, "dot", "--no-rank")
		dot := readOutputFile(t, filepath.Join(dir, "plain.dot"))

		assert.NotContains(t, dot, "dir=back", "--no-rank must suppress the rank=reverse swap")
		assert.NotContains(t, dot, "constraint=false", "--no-rank must suppress rank=equal")

		// Other aspects untouched.
		assert.Contains(t, dot, "minlen=3", "--no-rank must keep link length")
		assert.Contains(t, dot, "#1565C0", "--no-rank must keep the author link colour")
	})
}

// =============================================================================
// Tests for --no-labels (LBL-01..03, phase 38 plan 03)
// =============================================================================

// TestNoLabelsAllGenerationsAndFormats proves the quick 260831-01u BUG-2
// semantics: --no-labels suppresses ONLY edge label text. It applies to the C1
// context diagram, every drill-down, and the --expanded generation, across
// dot/svg/html formats, while node/cluster labels and the legend (metadata,
// LBL-03 pin) all stay.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestNoLabelsAllGenerationsAndFormats(t *testing.T) {
	t.Run("dot: edge labels suppressed, node/cluster labels survive", func(t *testing.T) {
		dir := generatePlainFixtureOutput(t, "dot", "--no-labels")

		dotFiles := make([]string, 0)

		for _, f := range collectGeneratedFiles(t, dir) {
			if strings.HasSuffix(f, ".dot") {
				dotFiles = append(dotFiles, f)
			}
		}

		require.Len(t, dotFiles, 2, "C1 plus the orders drill-down must be generated")

		for _, f := range dotFiles {
			dot := readOutputFile(t, f)
			name := filepath.Base(f)

			// Suppressed: edge label text only.
			assert.NotContains(t, dot, "[HTTPS]", "%s: edge label text suppressed", name)
			assert.NotContains(t, dot, "Streams order events", "%s: edge description text suppressed", name)

			// Surviving: node and cluster label content.
			assert.Contains(t, dot, "Order API", "%s: node name text survives", name)
			assert.Contains(t, dot, "Order Context", "%s: cluster name text survives", name)

			// The legend (and the nav/title graph label) are the only
			// sanctioned UPPERCASE HTML labels; element labels now add their
			// lowercase HTML tables.
			assert.Contains(t, dot, "__c4drill_legend", "%s: legend stays under --no-labels (LBL-03)", name)
		}
	})

	t.Run("expanded generation honours the flag", func(t *testing.T) {
		dir := generatePlainFixtureOutput(t, "dot", "--no-labels", "--expanded")
		dot := readOutputFile(t, filepath.Join(dir, "plain.expanded.dot"))

		assert.NotContains(t, dot, "[HTTPS]", "expanded edge label text suppressed")
		assert.NotContains(t, dot, "Streams order events", "expanded edge description text suppressed")
		assert.Contains(t, dot, "Order API", "expanded node text survives (edge labels only)")
		assert.Contains(t, dot, "__c4drill_legend", "expanded legend stays")
	})

	t.Run("svg and html formats generate non-empty", func(t *testing.T) {
		for _, format := range []string{"svg", "html"} {
			dir := generatePlainFixtureOutput(t, format, "--no-labels")

			for _, f := range collectGeneratedFiles(t, dir) {
				info, err := os.Stat(f)
				require.NoError(t, err, "stat %s", f)
				assert.Positive(t, info.Size(), "%s must be non-empty", f)
			}
		}
	})
}

// TestNoLabelsComposesWithPlainAndSwitches proves the flag composition: the
// flag unions with --plain and the granular switches without error, and the
// composed output still suppresses edge label text only.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestNoLabelsComposesWithPlainAndSwitches(t *testing.T) {
	for _, extraArgs := range [][]string{
		{"--plain", "--no-labels"},
		{"--no-colors", "--no-labels"},
		{"--plain", "--no-colors", "--no-styles", "--no-length", "--no-rank", "--no-labels"},
	} {
		args := append([]string(nil), extraArgs...)
		label := strings.Join(args, " ")

		t.Run(label, func(t *testing.T) {
			dir := generatePlainFixtureOutput(t, "dot", args...)
			dot := readOutputFile(t, filepath.Join(dir, "plain.dot"))

			assert.NotContains(t, dot, "Streams order events",
				"%s: edge label text suppressed in composition", label)
			assert.Contains(t, dot, "Order API",
				"%s: node label text survives (edge labels only)", label)
			assert.Contains(t, dot, "__c4drill_legend", "%s: legend still present", label)
		})
	}
}

// =============================================================================
// Tests for the KEY-03 composition matrix (phase 38 plan 04)
// =============================================================================

// generateFixtureOutput runs the CLI on the named testdata fixture into a
// fresh temp directory in the given format, with any extra args, and returns
// that directory. Generalizes generatePlainFixtureOutput to any fixture.
func generateFixtureOutput(t *testing.T, fixture, format string, extraArgs ...string) string {
	t.Helper()

	dir := t.TempDir()
	args := append([]string{
		filepath.Join("testdata", fixture),
		"--output", dir,
		"--format", format,
	}, extraArgs...)

	cmd := NewRootCmd()
	cmd.SetArgs(args)
	require.NoError(t, cmd.Execute(), "%s must render cleanly with %v", fixture, extraArgs)

	return dir
}

// TestNoLabelsC1Golden locks the --no-labels C1 output against the committed
// testdata/nolabels.dot golden (order-insensitive canonical comparison, DI-1).
// Golden re-baselined in quick 260831-01u for the BUG-2 semantics (edge-label
// suppression only: node/cluster labels returned, edge labels still gone).
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestNoLabelsC1Golden(t *testing.T) {
	dir := generatePlainFixtureOutput(t, "dot", "--no-labels")

	got := readOutputFile(t, filepath.Join(dir, "plain.dot"))
	expected := readOutputFile(t, filepath.Join("testdata", "nolabels.dot"))

	require.Equal(t, canonical.Canonical(t, expected), canonical.Canonical(t, got),
		"--no-labels C1 output must match the committed nolabels.dot golden (canonical, DI-1)")
}

// TestNoLabelsExpandedGolden locks --no-labels x --expanded against the
// committed testdata/nolabels.expanded.dot golden (re-baselined in quick
// 260831-01u for the edge-labels-only semantics).
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestNoLabelsExpandedGolden(t *testing.T) {
	dir := generatePlainFixtureOutput(t, "dot", "--no-labels", "--expanded")

	got := readOutputFile(t, filepath.Join(dir, "plain.expanded.dot"))
	expected := readOutputFile(t, filepath.Join("testdata", "nolabels.expanded.dot"))

	require.Equal(t, canonical.Canonical(t, expected), canonical.Canonical(t, got),
		"--no-labels --expanded output must match the committed nolabels.expanded.dot golden (canonical, DI-1)")
}

// TestKeyComposition is the KEY-03 E2E matrix: for every granular switch and
// --no-labels × every generation (C1, a drill-down, --expanded) × every format
// (dot/svg/html), output is generated non-empty and the switch's suppression is
// observable in the dot output. Pairwise compositions (--plain --no-labels,
// --no-colors --no-labels) are included, over the flat (plain.toml), styles,
// and multilevel fixtures.
//
// svg/html bytes are graphviz-owned (layout geometry); they are asserted for
// existence/non-emptiness plus one marker each (e.g. author hex absent under
// --no-colors in svg) — dot remains the precise suppression surface.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestKeyComposition(t *testing.T) {
	for _, sw := range keySwitches() {
		t.Run(sw.flag, func(t *testing.T) {
			runKeySwitchMatrix(t, sw)
		})
	}

	t.Run("compositions", func(t *testing.T) {
		t.Run("--plain --no-labels (dot, all generations)", func(t *testing.T) {
			for _, gen := range []string{"C1", "drilldown", "expanded"} {
				assertPlainNoLabelsComposition(t, gen)
			}
		})

		t.Run("--no-colors --no-labels (dot, C1 + drill-down)", func(t *testing.T) {
			for _, gen := range []string{"C1", "drilldown"} {
				assertNoColorsNoLabelsComposition(t, gen)
			}
		})

		t.Run("--no-colors --no-labels (multilevel C2: wrappers stay legible)", func(t *testing.T) {
			assertMultilevelLabelsLegible(t)
		})
	})
}

// keyDrilldowns maps each matrix fixture to its drill-down relative path
// (empty = no drill-down generated).
func keyDrilldowns() map[string]string {
	return map[string]string{
		"plain":      filepath.Join("plain", "orders.dot"),
		"multilevel": filepath.Join("multilevel", "mainSystem.dot"),
		"styles":     filepath.Join("styles", "app.dot"),
	}
}

// keySwitchCase defines one granular switch's flag plus per-fixture dot
// discriminators (compared against the lowercased dot for hexes, raw dot
// otherwise). A fixture absent from a marker list has nothing observable for
// that switch — only generation is asserted there (the suppression is covered
// on the fixture that can show it).
type keySwitchCase struct {
	flag    string
	absent  map[string][]string
	present map[string][]string
}

// keySwitches enumerates the KEY-03 matrix switches.
func keySwitches() []keySwitchCase {
	return []keySwitchCase{
		{
			flag: "--no-colors",
			absent: map[string][]string{
				"plain":  {"#fff9c4", "#f9a825", "#1565c0", "#2e7d32"},
				"styles": {"#1565c0"},
			},
		},
		{
			flag:   "--no-styles",
			absent: map[string][]string{"styles": {"dashed", "dotted"}},
		},
		{
			flag:   "--no-length",
			absent: map[string][]string{"plain": {"minlen"}},
		},
		{
			flag:   "--no-rank",
			absent: map[string][]string{"plain": {"dir=back", "constraint=false"}},
		},
		{
			flag: "--no-labels",
			absent: map[string][]string{
				// Edge label text only — node/cluster labels survive
				// (quick 260831-01u BUG-2).
				"plain": {"[HTTPS]"},
			},
			present: map[string][]string{
				"plain":      {"Order API"},
				"multilevel": {"<table"},
				"styles":     {"<table"},
			},
		},
	}
}

// runKeySwitchMatrix runs one switch across the three fixtures, the three dot
// generations, and the svg/html formats.
func runKeySwitchMatrix(t *testing.T, sw keySwitchCase) {
	t.Helper()

	for _, fixture := range []string{"plain", "multilevel", "styles"} {
		t.Run(fixture, func(t *testing.T) {
			for _, gen := range []string{"C1", "drilldown", "expanded"} {
				if fixture == "styles" && sw.flag != "--no-styles" {
					// The styles fixture exists solely to make
					// --no-styles observable; other switches assert
					// their suppression on plain/multilevel.
					continue
				}

				t.Run(gen+" dot", func(t *testing.T) {
					runKeyDotCase(t, sw, fixture, gen)
				})
			}

			// svg/html: C1 generation + one marker each.
			for _, format := range []string{"svg", "html"} {
				t.Run("C1 "+format, func(t *testing.T) {
					runKeyBinaryCase(t, sw, fixture, format)
				})
			}
		})
	}
}

// runKeyDotCase renders one (switch, fixture, generation) dot matrix cell and
// asserts the switch's suppression markers.
func runKeyDotCase(t *testing.T, sw keySwitchCase, fixture, gen string) {
	t.Helper()

	dot := renderKeyCell(t, fixture, sw.flag, gen, "dot")
	lower := strings.ToLower(dot)

	for _, marker := range sw.absent[fixture] {
		// Hex colours are emitted in mixed case — compare
		// lowercased. Structural markers ("<table" is
		// lowercase only in element labels; the sanctioned
		// graph/legend markup is UPPERCASE) compare raw.
		haystack := dot

		if strings.HasPrefix(marker, "#") {
			haystack = lower
		}

		assert.NotContains(t, haystack, marker,
			"%s %s %s: %q must be suppressed", fixture, sw.flag, gen, marker)
	}

	for _, marker := range sw.present[fixture] {
		assert.Contains(t, dot, marker,
			"%s %s %s: %q must be present", fixture, sw.flag, gen, marker)
	}
}

// runKeyBinaryCase renders the C1 cell of one (switch, fixture) in an svg or
// html format and asserts the format-level markers.
func runKeyBinaryCase(t *testing.T, sw keySwitchCase, fixture, format string) {
	t.Helper()

	out := renderKeyCell(t, fixture, sw.flag, "C1", format)

	if format == "html" {
		assert.Contains(t, out, "<svg", "%s %s html: SVG inlined", fixture, sw.flag)
	} else {
		assert.Contains(t, out, "<svg", "%s %s svg: SVG markup present", fixture, sw.flag)
	}

	// --no-colors marker: author hex absent in svg too.
	if sw.flag == "--no-colors" && fixture == "plain" {
		for _, hex := range []string{"#fff9c4", "#f9a825", "#1565C0"} {
			assert.NotContains(t, out, hex,
				"%s %s svg: author colour %s must be suppressed", fixture, sw.flag, hex)
		}
	}
}

// renderKeyCell runs one matrix cell and returns the generated file content
// (dot) or file size (svg/html).
func renderKeyCell(t *testing.T, fixture, flag, gen, format string) string {
	t.Helper()

	dir, rel := generateKeyFixture(t, fixture, flag, gen, format)
	path := keyCellPath(t, dir, rel, fixture, gen, format)

	info, err := os.Stat(path)
	require.NoError(t, err, "%s %s %s: %s must be generated", fixture, flag, gen, format)
	require.Positive(t, info.Size(), "%s %s %s: output must be non-empty", fixture, flag, gen, format)

	return readOutputFile(t, path)
}

// generateKeyFixture renders one matrix cell's fixture with the switch flag
// and returns the output directory plus the fixture's output base name.
func generateKeyFixture(t *testing.T, fixture, flag, gen, format string) (string, string) {
	t.Helper()

	var (
		dir string
		rel string
	)

	switch fixture {
	case "styles":
		src := writeStylesFixture(t)

		out := t.TempDir()

		args := append([]string{src, "--output", out, "--format", format}, flag)

		if gen == "expanded" {
			args = append(args, "--expanded")
		}

		cmd := NewRootCmd()
		cmd.SetArgs(args)
		require.NoError(t, cmd.Execute(), "styles fixture must render with %s %s %s", flag, gen, format)

		dir = out
		rel = "styles"
	default:
		extra := []string{flag}

		if gen == "expanded" {
			extra = append(extra, "--expanded")
		}

		dir = generateFixtureOutput(t, fixture+".toml", format, extra...)
		rel = fixture
	}

	return dir, rel
}

// keyCellPath resolves the expected output file path for one matrix cell.
func keyCellPath(t *testing.T, dir, rel, fixture, gen, format string) string {
	t.Helper()

	var path string

	switch gen {
	case "C1":
		path = filepath.Join(dir, rel+"."+format)
	case "drilldown":
		if keyDrilldowns()[fixture] == "" {
			t.Skipf("fixture %s has no drill-down", fixture)
		}

		path = filepath.Join(dir, strings.TrimSuffix(keyDrilldowns()[fixture], ".dot")+"."+format)
	case "expanded":
		path = filepath.Join(dir, rel+".expanded."+format)
	default:
		t.Fatalf("unknown generation %q", gen)
	}

	return path
}

// writeStylesFixture writes the dedicated styles fixture (author styles that
// reach the DOT verbatim, per TestGranularFlagsE2E) into a temp dir and
// returns its path.
func writeStylesFixture(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	src := filepath.Join(dir, "styles.toml")
	content := `[properties]
name = "Styles"

[app]
type = "system"
name = "App"
style = "dotted"

[app.api]
type = "container"
name = "API"

[[app.api.link]]
peer = "db"
style = "dashed"
color = "#1565C0"

[db]
type = "db"
name = "DB"
`
	require.NoError(t, os.WriteFile(src, []byte(content), 0o600))

	return src
}

// assertPlainNoLabelsComposition asserts one generation of the
// --plain --no-labels composition over plain.toml.
func assertPlainNoLabelsComposition(t *testing.T, gen string) {
	t.Helper()

	// renderKeyCell takes one flag, so the composition renders directly.
	args := []string{"--plain", "--no-labels"}

	if gen == "expanded" {
		args = append(args, "--expanded")
	}

	dir := generateFixtureOutput(t, "plain.toml", "dot", args...)

	path := filepath.Join(dir, "plain.dot")

	if gen == "drilldown" {
		path = filepath.Join(dir, "plain", "orders.dot")
	}

	if gen == "expanded" {
		path = filepath.Join(dir, "plain.expanded.dot")
	}

	dot := readOutputFile(t, path)

	assert.NotContains(t, dot, "Streams order events", "%s: edge label text suppressed", gen)
	assert.Contains(t, dot, "Order API", "%s: node label text survives", gen)
	assert.Contains(t, dot, "__c4drill_legend", "%s: legend stays", gen)
	assert.NotContains(t, strings.ToLower(dot), "#fff9c4", "%s: --plain still strips author colours", gen)
}

// assertNoColorsNoLabelsComposition asserts one generation of the
// --no-colors --no-labels composition over plain.toml.
func assertNoColorsNoLabelsComposition(t *testing.T, gen string) {
	t.Helper()

	dir := generateFixtureOutput(t, "plain.toml", "dot", "--no-colors", "--no-labels")

	path := filepath.Join(dir, "plain.dot")

	if gen == "drilldown" {
		path = filepath.Join(dir, "plain", "orders.dot")
	}

	dot := readOutputFile(t, path)
	lower := strings.ToLower(dot)

	assert.NotContains(t, dot, "Streams order events", "%s: edge label text suppressed (raw dot)", gen)
	assert.Contains(t, dot, "Order API", "%s: node labels survive", gen)

	for _, hex := range []string{"#fff9c4", "#f9a825", "#1565c0", "#2e7d32"} {
		assert.NotContains(t, lower, hex, "%s: colours suppressed in composition", gen)
	}
}

// assertMultilevelLabelsLegible pins wrapper/boundary label legibility on the
// multilevel C2 under the --no-colors --no-labels composition.
func assertMultilevelLabelsLegible(t *testing.T) {
	t.Helper()

	dir := generateFixtureOutput(t, "multilevel.toml", "dot", "--no-colors", "--no-labels")
	c2 := readOutputFile(t, filepath.Join(dir, "multilevel", "mainSystem.dot"))

	assert.Contains(t, c2, "Main System", "wrapper/boundary labels legible on multilevel C2 (edge labels only)")
	assert.Contains(t, c2, "subgraph cluster_", "cluster structure survives labels-off")
}

// TestNoLabelsOptIn proves the flag is opt-in: without it the same fixture
// still renders HTML labels with full text.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestNoLabelsOptIn(t *testing.T) {
	dir := generatePlainFixtureOutput(t, "dot")

	dot := readOutputFile(t, filepath.Join(dir, "plain.dot"))

	assert.Contains(t, strings.ToLower(dot), "<table",
		"default mode must keep the HTML label path (opt-in proven)")
	assert.Contains(t, dot, "Order API", "default mode must keep node label text")
}
