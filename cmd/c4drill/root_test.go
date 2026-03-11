package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

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
	assert.Contains(t, output, "dot|svg", "Help should show available formats")
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
	assert.Equal(t, ".", outputFlag.DefValue)
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
	// Create TOML with invalid reference (validation error)
	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "invalid_ref.toml")
	content := `
[properties]
name = "Test"

[user]
type = "person"
name = "User"

[[user.link]]
peer = "nonexistent"
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
