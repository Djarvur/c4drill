// Package output provides file output functionality for rendered diagrams.
package output

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// dirPermission is the permission for created directories.
	dirPermission = 0o750
	// filePermission is the permission for created files.
	filePermission = 0o600
	// pngExt is the PNG raster file extension.
	pngExt = ".png"
)

// PNGImageName returns the bare file name (no directory) of the PNG that the
// sibling HTML navigation doc embeds via <img src> (issue #26). It mirrors
// the Write layout so the doc always references the image sitting next to it:
// {basename}.png for C1, {last-path-segment}.png for C2/C3.
func PNGImageName(basename, unitPath string) string {
	if unitPath == "" {
		return basename + pngExt
	}

	parts := strings.Split(unitPath, ".")

	return parts[len(parts)-1] + pngExt
}

// ExpandedPNGImageName returns the bare file name of the PNG for the
// --expanded diagram's HTML navigation doc: {basename}.expanded.png,
// mirroring WriteExpanded's layout.
func ExpandedPNGImageName(basename string) string {
	return basename + ".expanded" + pngExt
}

// Writer handles writing rendered diagrams to files with proper directory structure.
type Writer struct {
	baseDir string
}

// NewWriter creates a new Writer that outputs files to the specified base directory.
func NewWriter(baseDir string) *Writer {
	return &Writer{baseDir: baseDir}
}

// BaseDir returns the base output directory.
func (w *Writer) BaseDir() string {
	return w.baseDir
}

// Write writes rendered data to the appropriate output path.
// For C1 (Context level, empty unitPath): {basename}.{format}
// For C2/C3 (expanded units): {basename}/{unit-path}.{format}
// Dotted paths are converted to directory hierarchies (e.g., "mainapp.api" -> "mainapp/api").
func (w *Writer) Write(basename, unitPath, format string, data []byte) error {
	var relPath string
	if unitPath == "" {
		// C1: flat file at {basename}.{format}
		relPath = fmt.Sprintf("%s.%s", basename, format)
	} else {
		// C2/C3: directory hierarchy from dotted path
		dirPath := strings.ReplaceAll(unitPath, ".", string(filepath.Separator))
		relPath = filepath.Join(basename, dirPath+"."+format)
	}

	fullPath := filepath.Join(w.baseDir, relPath)

	// Create parent directories (fail fast on error)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, dirPermission); err != nil {
		return fmt.Errorf("create output directory %s: %w", dir, err)
	}

	// Write file (fail fast on error)
	if err := os.WriteFile(fullPath, data, filePermission); err != nil {
		return fmt.Errorf("write output file %s: %w", fullPath, err)
	}

	return nil
}

// WriteExpanded writes expanded diagram data to {basename}.expanded.{format}.
// This is used for the --expanded mode that generates a single diagram with all units.
func (w *Writer) WriteExpanded(basename, format string, data []byte) error {
	relPath := fmt.Sprintf("%s.expanded.%s", basename, format)
	fullPath := filepath.Join(w.baseDir, relPath)

	// Create parent directories (fail fast on error)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, dirPermission); err != nil {
		return fmt.Errorf("create output directory %s: %w", dir, err)
	}

	// Write file (fail fast on error)
	if err := os.WriteFile(fullPath, data, filePermission); err != nil {
		return fmt.Errorf("write output file %s: %w", fullPath, err)
	}

	return nil
}
