package graph

import (
	"net/url"
	"strings"
)

// ComputeExploreURL calculates the relative path from current diagram to target.
// currentPath is the dotted path of the current diagram (empty for C1).
// targetPath is the dotted path of the target unit.
// basename is the output base filename (from TOML file).
// format is the output format (e.g., "svg").
func ComputeExploreURL(_ string, targetPath, _ string, format string) string {
	// Convert dotted path to directory structure with URL-encoded segments
	parts := strings.Split(targetPath, ".")

	encodedParts := make([]string, len(parts))
	for i, part := range parts {
		encodedParts[i] = URLEncodePath(part)
	}

	return "./" + strings.Join(encodedParts, "/") + "." + format
}

// ComputeBackLinkURL calculates the relative path from current diagram back to parent.
// currentPath is the dotted path of the current diagram.
// basename is the output base filename.
// format is the output format.
func ComputeBackLinkURL(currentPath, basename, format string) string {
	// C1 has no parent
	if currentPath == "" {
		return ""
	}

	parts := strings.Split(currentPath, ".")
	if len(parts) == 1 {
		// C2 level - go back to C1
		return "../" + basename + "." + format
	}
	// C3+ level - go back one level to parent directory
	// The parent file is named after the parent unit (last but one part)
	parentName := parts[len(parts)-2]

	return "../" + URLEncodePath(parentName) + "." + format
}

// BuildBreadcrumbPath creates breadcrumb items from a dotted path.
// The last item has empty URL (current level).
func BuildBreadcrumbPath(dottedPath, basename, format string) []BreadcrumbItem {
	if dottedPath == "" {
		return nil
	}

	parts := strings.Split(dottedPath, ".")
	items := make([]BreadcrumbItem, len(parts))

	for i, part := range parts {
		items[i] = BreadcrumbItem{
			Name: part,
		}

		// Last item is current level (no URL)
		if i < len(parts)-1 {
			// Build URL to this ancestor
			// For breadcrumbs, we need to compute relative path from current to ancestor
			items[i].URL = computeBreadcrumbURL(parts, i, basename, format)
		}
	}

	return items
}

// computeBreadcrumbURL computes the URL for a breadcrumb ancestor at the given index.
func computeBreadcrumbURL(parts []string, ancestorIndex int, basename, format string) string {
	// Number of levels to go up = total depth - ancestor level
	levelsUp := len(parts) - ancestorIndex - 1

	var (
		up     string
		upSb76 strings.Builder
	)

	for range levelsUp {
		upSb76.WriteString("../")
	}

	up += upSb76.String()

	if ancestorIndex == 0 {
		// First level ancestor - link to C1 basename
		return up + basename + "." + format
	}

	// Build path to ancestor
	ancestorPath := strings.Join(parts[:ancestorIndex+1], "/")

	return up + URLEncodePath(ancestorPath) + "." + format
}

// URLEncodePath URL-encodes a path segment for use in URLs.
func URLEncodePath(path string) string {
	return url.PathEscape(path)
}
