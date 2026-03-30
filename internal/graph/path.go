package graph

import (
	"net/url"
	"strings"
)

// ComputeExploreURL calculates the relative path from current diagram to target.
// currentPath is the dotted path of the current diagram (empty for C1).
// targetPath is the dotted path of the target unit.
// basename is the output base filename (from TOML file).
// format is the output format (e.g., "svg") - ignored; URLs always use "svg"
// for browser navigation regardless of the format being generated.
func ComputeExploreURL(currentPath, targetPath, basename, _ string) string {
	// Always use SVG for clickable links (browser navigation)
	const linkFormat = "svg"

	// Convert dotted target path to URL-encoded directory/file segments
	targetParts := strings.Split(targetPath, ".")
	encodedTarget := make([]string, len(targetParts))
	for i, part := range targetParts {
		encodedTarget[i] = URLEncodePath(part)
	}

	// C1 case: file is at {basename}.{fmt}, target is at {basename}/{path}.svg
	if currentPath == "" {
		return basename + "/" + strings.Join(encodedTarget, "/") + "." + linkFormat
	}

	// For C2/C3, compute proper relative path from current file's directory to target file.
	//
	// File layout under {basename}/:
	//   C2 for "mainapp"        -> mainapp.dot              (dir: .)
	//   C3 for "mainapp.api"    -> mainapp/api.dot          (dir: mainapp/)
	//   C4 for "mainapp.api.v2" -> mainapp/api/v2.dot       (dir: mainapp/api/)
	//
	// The current file's directory is all parts except the last.
	// The target file's path is the full encoded target path.
	currentParts := strings.Split(currentPath, ".")

	// Current file's directory parts (all but last which is filename)
	currentDirParts := currentParts[:len(currentParts)-1]

	// Find common directory prefix length
	commonDirLen := 0
	for i := 0; i < len(currentDirParts) && i < len(encodedTarget); i++ {
		if currentDirParts[i] == targetParts[i] {
			commonDirLen++
		} else {
			break
		}
	}

	// Levels to go up from current directory to common ancestor
	levelsUp := len(currentDirParts) - commonDirLen

	upPath := ""
	for range levelsUp {
		upPath += "../"
	}

	// Path down from common ancestor to target (all target parts after common prefix)
	downPath := strings.Join(encodedTarget[commonDirLen:], "/") + "." + linkFormat

	return upPath + downPath
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

	up := ""
	for range levelsUp {
		up += "../"
	}

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
