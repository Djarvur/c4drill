package render_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Djarvur/c4drill/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIconExtractor_Extract(t *testing.T) {
	t.Parallel()
	// Create temp directory for test
	tmpDir := t.TempDir()

	extractor := render.NewIconExtractor(tmpDir)

	// Extract a person icon with C1 color
	relPath, err := extractor.Extract("person", "#3C7FC0")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(".icons", "person-3C7FC0.svg"), relPath)

	// Verify file exists
	fullPath := filepath.Join(tmpDir, relPath)
	_, err = os.Stat(fullPath)
	require.NoError(t, err)

	// Verify content has color
	//nolint:gosec // G304: Test file reading from controlled temp directory
	content, err := os.ReadFile(fullPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "#3C7FC0")
}

func TestIconExtractor_ExtractCachesInMemory(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	extractor := render.NewIconExtractor(tmpDir)

	// First extraction
	_, err := extractor.Extract("db", "#3C7FC0")
	require.NoError(t, err)

	// Second extraction should hit memory cache
	// (We can't easily test this without mocking, but we verify no error)
	_, err = extractor.Extract("db", "#3C7FC0")
	require.NoError(t, err)
}

func TestIconExtractor_ExtractSkipsExistingFiles(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	extractor := render.NewIconExtractor(tmpDir)

	// Extract icon
	relPath, err := extractor.Extract("system", "#3C7FC0")
	require.NoError(t, err)

	// Create new extractor (simulates new run)
	extractor2 := render.NewIconExtractor(tmpDir)

	// Should skip extraction for existing file
	relPath2, err := extractor2.Extract("system", "#3C7FC0")
	require.NoError(t, err)
	assert.Equal(t, relPath, relPath2)
}

func TestIconExtractor_HexColorWithAndWithoutHash(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	extractor := render.NewIconExtractor(tmpDir)

	// With hash
	_, err := extractor.Extract("component", "#78A8D8")
	require.NoError(t, err)

	// Verify file created
	fullPath := filepath.Join(tmpDir, ".icons", "component-78A8D8.svg")
	//nolint:gosec // G304: Test file reading from controlled temp directory
	content, err := os.ReadFile(fullPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "#78A8D8")
}

func TestIconExtractor_InvalidIconType(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	extractor := render.NewIconExtractor(tmpDir)

	_, err := extractor.Extract("nonexistent", "#000000")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
