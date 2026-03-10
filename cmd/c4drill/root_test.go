package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestFormatFlag(t *testing.T) {
	cmd := NewRootCmd()

	formatFlag := cmd.PersistentFlags().Lookup("format")
	assert.NotNil(t, formatFlag)
	assert.Equal(t, "svg", formatFlag.DefValue)
	assert.Equal(t, "f", formatFlag.Shorthand)
}

func TestOutputFlag(t *testing.T) {
	cmd := NewRootCmd()

	outputFlag := cmd.PersistentFlags().Lookup("output")
	assert.NotNil(t, outputFlag)
	assert.Equal(t, ".", outputFlag.DefValue)
	assert.Equal(t, "o", outputFlag.Shorthand)
}

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
		t.Run(tt.name, func(t *testing.T) {
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
			} else {
				// For valid formats, the command will fail at file reading stage
				// but not at format validation
				if err != nil {
					assert.NotContains(t, err.Error(), "invalid format")
				}
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
link = { "nonexistent" = {} }
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

[webapp]
type = "system"
name = "Web Application"
description = "Main web application"
technology = "Go, React"
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

[webapp]
type = "system"
name = "Web Application"
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
