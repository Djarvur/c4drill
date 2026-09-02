package output_test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// windowsOS is the runtime.GOOS value for Windows systems.
	windowsOS = "windows"
	// dirPermission is the permission for created directories.
	dirPermission = 0o750
	// filePermission is the permission for created files.
	filePermission = 0o600
)

func TestWriterC1FlatPath(t *testing.T) {
	t.Parallel()

	// Test 1: Writer creates C1 flat file at {basename}.{format}
	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)

	data := []byte("test diagram content")
	err := w.Write("system", "", "svg", data)
	require.NoError(t, err)

	// Verify file exists at expected path
	expectedPath := filepath.Join(tmpDir, "system.svg")
	assert.FileExists(t, expectedPath)

	// Verify content
	//nolint:gosec // G304: Test reads from temp directory created by t.TempDir()
	content, err := os.ReadFile(expectedPath)
	require.NoError(t, err)
	assert.Equal(t, data, content)
}

func TestWriterC2C3NestedPath(t *testing.T) {
	t.Parallel()

	// Test 2: Writer creates C2/C3 file at {basename}/{unit-path}.{format}
	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)

	data := []byte("expanded diagram content")
	err := w.Write("system", "mainapp.api", "svg", data)
	require.NoError(t, err)

	// Verify file exists at expected nested path
	expectedPath := filepath.Join(tmpDir, "system", "mainapp", "api.svg")
	assert.FileExists(t, expectedPath)

	// Verify content
	//nolint:gosec // G304: Test reads from temp directory created by t.TempDir()
	content, err := os.ReadFile(expectedPath)
	require.NoError(t, err)
	assert.Equal(t, data, content)
}

func TestWriterDottedPathConversion(t *testing.T) {
	t.Parallel()

	// Test 3: Writer converts dotted path to directory hierarchy
	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)

	data := []byte("deeply nested content")
	err := w.Write("system", "mainapp.api.handlers", "dot", data)
	require.NoError(t, err)

	// Verify dotted path converted to directory hierarchy
	// mainapp.api.handlers -> mainapp/api/handlers
	expectedPath := filepath.Join(tmpDir, "system", "mainapp", "api", "handlers.dot")
	assert.FileExists(t, expectedPath)
}

func TestWriterDirectoryCreation(t *testing.T) {
	t.Parallel()

	// Test 4: Writer creates parent directories with os.MkdirAll
	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)

	// Write to deeply nested path that requires multiple directory levels
	data := []byte("content")
	err := w.Write("system", "a.b.c.d.e", "svg", data)
	require.NoError(t, err)

	// Verify all intermediate directories were created
	expectedPath := filepath.Join(tmpDir, "system", "a", "b", "c", "d", "e.svg")
	assert.FileExists(t, expectedPath)

	// Verify parent directories exist
	assert.DirExists(t, filepath.Join(tmpDir, "system"))
	assert.DirExists(t, filepath.Join(tmpDir, "system", "a", "b", "c", "d"))
}

func TestWriterDirectoryCreationError(t *testing.T) {
	t.Parallel()

	// Test 5: Writer returns error when directory creation fails
	if runtime.GOOS == windowsOS {
		t.Skip("Unix-specific error test")
	}

	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)

	// Create a file where we need a directory to force mkdir error
	filePath := filepath.Join(tmpDir, "blocked")
	err := os.WriteFile(filePath, []byte("blocker"), filePermission)
	require.NoError(t, err)

	// Try to write a file that would require "blocked" to be a directory
	data := []byte("content")
	err = w.Write("blocked", "subfile", "svg", data)
	require.Error(t, err)

	// Error should mention directory creation
	assert.Contains(t, err.Error(), "create output directory")
}

func TestWriterFileWriteError(t *testing.T) {
	t.Parallel()

	// Test 6: Writer returns error when file write fails
	if runtime.GOOS == windowsOS {
		t.Skip("Unix-specific error test")
	}

	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)

	// Create a directory where we need a file to force write error
	dirPath := filepath.Join(tmpDir, "system.svg")
	err := os.MkdirAll(dirPath, dirPermission)
	require.NoError(t, err)

	// Try to write to a path that's a directory
	data := []byte("content")
	err = w.Write("system", "", "svg", data)
	require.Error(t, err)

	// Error should mention file write
	assert.Contains(t, err.Error(), "write output file")
}

func TestWriterMultipleFiles(t *testing.T) {
	t.Parallel()

	// Additional test: Verify multiple files can be written
	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)

	// Write C1 diagram
	err := w.Write("system", "", "svg", []byte("c1 content"))
	require.NoError(t, err)

	// Write multiple C2/C3 diagrams
	err = w.Write("system", "api", "svg", []byte("api content"))
	require.NoError(t, err)

	err = w.Write("system", "db", "svg", []byte("db content"))
	require.NoError(t, err)

	// Verify all files exist
	assert.FileExists(t, filepath.Join(tmpDir, "system.svg"))
	assert.FileExists(t, filepath.Join(tmpDir, "system", "api.svg"))
	assert.FileExists(t, filepath.Join(tmpDir, "system", "db.svg"))
}

func TestWriterDifferentFormats(t *testing.T) {
	t.Parallel()

	// Additional test: Verify different output formats
	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)

	// Write same diagram in different formats
	dotData := []byte("digraph { }")
	svgData := []byte("<svg></svg>")

	err := w.Write("system", "", "dot", dotData)
	require.NoError(t, err)

	err = w.Write("system", "", "svg", svgData)
	require.NoError(t, err)

	// Verify both files exist with correct extensions
	//nolint:gosec // G304: Test reads from temp directory created by t.TempDir()
	dotContent, err := os.ReadFile(filepath.Join(tmpDir, "system.dot"))
	require.NoError(t, err)
	assert.Equal(t, dotData, dotContent)

	//nolint:gosec // G304: Test reads from temp directory created by t.TempDir()
	svgContent, err := os.ReadFile(filepath.Join(tmpDir, "system.svg"))
	require.NoError(t, err)
	assert.Equal(t, svgData, svgContent)
}

func TestWriterEmptyBasename(t *testing.T) {
	t.Parallel()

	// Edge case: Empty basename should still work
	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)

	data := []byte("content")
	err := w.Write("", "unit", "svg", data)
	require.NoError(t, err)

	// File should be at tmpDir/unit.svg
	expectedPath := filepath.Join(tmpDir, "unit.svg")
	assert.FileExists(t, expectedPath)
}

func TestWriterNewWriterReturnsNil(t *testing.T) {
	t.Parallel()

	// Verify NewWriter returns a valid Writer pointer
	w := output.NewWriter("/tmp")
	require.NotNil(t, w)
}

func TestWriterErrorWrapping(t *testing.T) {
	t.Parallel()

	// Verify errors are properly wrapped with context
	if runtime.GOOS == windowsOS {
		t.Skip("Unix-specific error test")
	}

	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)

	// Create a file where directory should be
	blockerPath := filepath.Join(tmpDir, "test")
	err := os.WriteFile(blockerPath, []byte("blocker"), filePermission)
	require.NoError(t, err)

	// This should fail with wrapped error
	err = w.Write("test", "nested.path", "svg", []byte("content"))
	require.Error(t, err)

	// Verify error has useful context
	assert.True(t,
		errors.Is(err, os.ErrExist) || strings.Contains(err.Error(), "not a directory"),
		"error should wrap underlying OS error")
}

// Tests for WriteExpanded

func TestWriteExpanded_CreatesFileWithExpandedExtension(t *testing.T) {
	t.Parallel()

	// Test: WriteExpanded creates file at {basename}.expanded.{format}
	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)

	data := []byte("expanded diagram content")
	err := w.WriteExpanded("system", "svg", data)
	require.NoError(t, err)

	// Verify file exists at expected path
	expectedPath := filepath.Join(tmpDir, "system.expanded.svg")
	assert.FileExists(t, expectedPath)

	// Verify content
	//nolint:gosec // G304: Test reads from temp directory created by t.TempDir()
	content, err := os.ReadFile(expectedPath)
	require.NoError(t, err)
	assert.Equal(t, data, content)
}

func TestWriteExpanded_DifferentFormats(t *testing.T) {
	t.Parallel()

	// Test: WriteExpanded works with different formats
	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)

	dotData := []byte("digraph { }")
	svgData := []byte("<svg></svg>")

	err := w.WriteExpanded("system", "dot", dotData)
	require.NoError(t, err)

	err = w.WriteExpanded("system", "svg", svgData)
	require.NoError(t, err)

	// Verify both files exist with correct extensions
	assert.FileExists(t, filepath.Join(tmpDir, "system.expanded.dot"))
	assert.FileExists(t, filepath.Join(tmpDir, "system.expanded.svg"))
}

func TestWriteExpanded_CreatesParentDirectory(t *testing.T) {
	t.Parallel()

	// Test: WriteExpanded creates parent directories if needed
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "nested", "deep", "path")
	w := output.NewWriter(nestedDir)

	data := []byte("content")
	err := w.WriteExpanded("system", "svg", data)
	require.NoError(t, err)

	// Verify file was created in nested directory
	expectedPath := filepath.Join(nestedDir, "system.expanded.svg")
	assert.FileExists(t, expectedPath)
}

func TestWriteExpanded_EmptyBasename(t *testing.T) {
	t.Parallel()

	// Edge case: Empty basename should still work
	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)

	data := []byte("content")
	err := w.WriteExpanded("", "svg", data)
	require.NoError(t, err)

	// File should be at tmpDir/.expanded.svg
	expectedPath := filepath.Join(tmpDir, ".expanded.svg")
	assert.FileExists(t, expectedPath)
}

func TestWriteExpanded_OverwritesExistingFile(t *testing.T) {
	t.Parallel()

	// Test: WriteExpanded overwrites existing file
	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)

	// Write initial content
	err := w.WriteExpanded("system", "svg", []byte("original content"))
	require.NoError(t, err)

	// Overwrite with new content
	err = w.WriteExpanded("system", "svg", []byte("new content"))
	require.NoError(t, err)

	// Verify new content
	//nolint:gosec // G304: Test reads from temp directory created by t.TempDir()
	content, err := os.ReadFile(filepath.Join(tmpDir, "system.expanded.svg"))
	require.NoError(t, err)
	assert.Equal(t, []byte("new content"), content)
}

// Tests for PNG image-name helpers (issue #26): the HTML navigation doc that
// accompanies each PNG embeds its raster sibling by bare file name, so the
// helper must mirror the Write/WriteExpanded layout exactly.

func TestPNGImageName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		basename   string
		unitPath   string
		wantImage  string
		wantDocDir string // directory the doc/PNG pair lands in, relative to base
	}{
		{
			name:       "C1 flat layout",
			basename:   "model",
			unitPath:   "",
			wantImage:  "model.png",
			wantDocDir: ".",
		},
		{
			name:       "C2 single segment",
			basename:   "model",
			unitPath:   "mainapp",
			wantImage:  "mainapp.png",
			wantDocDir: "model",
		},
		{
			name:       "C3 dotted path",
			basename:   "model",
			unitPath:   "mainapp.api",
			wantImage:  "api.png",
			wantDocDir: filepath.Join("model", "mainapp"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantImage, output.PNGImageName(tt.basename, tt.unitPath))

			// Consistency lock: the name helper must reference exactly the
			// PNG file Write produces for the same basename/unitPath.
			tmpDir := t.TempDir()
			w := output.NewWriter(tmpDir)
			require.NoError(t, w.Write(tt.basename, tt.unitPath, "png", []byte("raster")))

			docDir := tmpDir
			if tt.wantDocDir != "." {
				docDir = filepath.Join(tmpDir, tt.wantDocDir)
			}

			assert.FileExists(t, filepath.Join(docDir, tt.wantImage),
				"PNGImageName must name the PNG Write creates")
		})
	}
}

func TestExpandedPNGImageName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "model.expanded.png", output.ExpandedPNGImageName("model"))

	// Consistency lock against WriteExpanded's layout.
	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)
	require.NoError(t, w.WriteExpanded("model", "png", []byte("raster")))
	assert.FileExists(t, filepath.Join(tmpDir, output.ExpandedPNGImageName("model")))
}
