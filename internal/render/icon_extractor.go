package render

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Djarvur/c4drill/internal/render/icons"
)

const (
	dirPermission  = 0o750 // Directory permissions (gosec G301)
	filePermission = 0o600 // File permissions (gosec G306)
)

// IconExtractor manages extraction of colored icons to the output directory.
// Per D-02: Icons are extracted on-demand to {output}/.icons/
// For SVG output, icons are embedded as base64 data URIs (WASM graphviz can't load external files).
type IconExtractor struct {
	outputDir string            // Base output directory (where SVG files are written)
	extracted map[string]bool   // Cache of already-extracted icon filenames
	svgCache  map[string]string // Cache of SVG data URIs
}

// NewIconExtractor creates an extractor for the given output directory.
func NewIconExtractor(outputDir string) *IconExtractor {
	return &IconExtractor{
		outputDir: outputDir,
		extracted: make(map[string]bool),
		svgCache:  make(map[string]string),
	}
}

// Extract extracts a colored icon if not already extracted.
// Per D-03: Filename format is type-{hexcolor}.svg (hex without #)
// Per D-04: Returns relative path for use in IMG SRC (e.g., ".icons/person-3C7FC0.svg").
func (e *IconExtractor) Extract(iconType, hexColor string) (string, error) {
	// Clean hex color: remove # if present for filename
	hexClean := strings.TrimPrefix(hexColor, "#")
	iconFilename := fmt.Sprintf("%s-%s.svg", iconType, hexClean)

	// Check memory cache first
	if e.extracted[iconFilename] {
		return filepath.Join(".icons", iconFilename), nil
	}

	// Build paths
	iconDir := filepath.Join(e.outputDir, ".icons")
	iconPath := filepath.Join(iconDir, iconFilename)

	// Create .icons directory if needed
	if err := os.MkdirAll(iconDir, dirPermission); err != nil {
		return "", fmt.Errorf("create icons directory: %w", err)
	}

	// Check if file already exists (from previous run)
	if _, err := os.Stat(iconPath); err == nil {
		e.extracted[iconFilename] = true

		return filepath.Join(".icons", iconFilename), nil
	}

	// Get template
	template, err := icons.GetTemplate(iconType)
	if err != nil {
		return "", fmt.Errorf("get icon template: %w", err)
	}

	// Colorize (ensure hexColor has # prefix for SVG)
	if !strings.HasPrefix(hexColor, "#") {
		hexColor = "#" + hexColor
	}

	coloredSVG := icons.Colorize(template, hexColor)

	// Write file
	if err := os.WriteFile(iconPath, []byte(coloredSVG), filePermission); err != nil {
		return "", fmt.Errorf("write icon file: %w", err)
	}

	e.extracted[iconFilename] = true

	return filepath.Join(".icons", iconFilename), nil
}

// ExtractSVGBase64 extracts a colored icon and returns it as an SVG data URI.
// This is needed for WASM-based graphviz which cannot load external image files.
// Returns data URI like "data:image/svg+xml;base64,...".
func (e *IconExtractor) ExtractSVGBase64(iconType, hexColor string) (string, error) {
	// Clean hex color: remove # if present for cache key
	hexClean := strings.TrimPrefix(hexColor, "#")
	cacheKey := fmt.Sprintf("%s-%s", iconType, hexClean)

	// Check cache first
	if dataURI, ok := e.svgCache[cacheKey]; ok {
		return dataURI, nil
	}

	// Get template
	template, err := icons.GetTemplate(iconType)
	if err != nil {
		return "", fmt.Errorf("get icon template: %w", err)
	}

	// Colorize (ensure hexColor has # prefix for SVG)
	if !strings.HasPrefix(hexColor, "#") {
		hexColor = "#" + hexColor
	}

	coloredSVG := icons.Colorize(template, hexColor)

	// Encode as base64 data URI
	dataURI := "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(coloredSVG))
	e.svgCache[cacheKey] = dataURI

	return dataURI, nil
}
