// Package icons provides embedded SVG icon templates for C4 diagram units.
package icons

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed person.svg db.svg pipe.svg system.svg container.svg component.svg
var svgFiles embed.FS

// Icon type constants for type-safe icon lookups.
const (
	TypePerson    = "person"
	TypeDb        = "db"
	TypePipe      = "pipe"
	TypeSystem    = "system"
	TypeContainer = "container"
	TypeComponent = "component"
)

// GetTemplate returns the raw SVG template for the given icon type.
// The template uses "currentColor" as a placeholder for the stroke color.
func GetTemplate(iconType string) (string, error) {
	filename := iconType + ".svg"
	data, err := svgFiles.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("icon template %s not found: %w", filename, err)
	}
	return string(data), nil
}

// Colorize replaces "currentColor" in the SVG with the provided hex color.
// The hexColor should include the # prefix (e.g., "#3C7FC0").
func Colorize(svgTemplate, hexColor string) string {
	return strings.ReplaceAll(svgTemplate, "currentColor", hexColor)
}
