package graph

import (
	"net/url"
	"strings"
)

// ComputeExploreURL calculates the relative path from current diagram to target.
// currentPath is the dotted path of the current diagram (empty for C1).
// targetPath is the dotted path of the target unit (a FULL dotted path such
// as "mainSystem.sshAuth" because BuildGraphWithPath passes node.ID).
// basename is the output base filename (from TOML file).
// format is the output format (e.g., "svg") - ignored; URLs always use "svg"
// for browser navigation regardless of the format being generated.
//
// File layout under {basename}/:
//
//	C1 for ""               -> {basename}.svg           (dir: .)
//	C2 for "mainapp"        -> mainapp.svg              (dir: .)
//	C3 for "mainapp.api"    -> mainapp/api.svg          (dir: mainapp/)
//	C4 for "mainapp.api.v2" -> mainapp/api/v2.svg       (dir: mainapp/api/)
//
// The current file's directory (all segments but the last) and the target
// file's directory (all segments but the last) are compared to compute a
// bidirectional relative URL that handles ancestor, sibling, and descendant
// targets uniformly. Returns the empty string when targetPath == currentPath
// (self-link guard — see Gap 1 symptom B) so the renderer omits the URL.
func ComputeExploreURL(currentPath, targetPath, basename, _ string) string {
	// Always use SVG for clickable links (browser navigation)
	const linkFormat = "svg"

	// Self-link guard (Gap 1 symptom B): a node whose ID equals the current
	// diagram's path would otherwise emit a broken href=".svg". Return empty so
	// the caller (createNode) skips SetURL entirely.
	if currentPath != "" && targetPath == currentPath {
		return ""
	}

	// Split dotted paths into segments and URL-encode each segment.
	currentParts := strings.Split(currentPath, ".")
	targetParts := strings.Split(targetPath, ".")

	encodeAll := func(parts []string) []string {
		encoded := make([]string, len(parts))
		for i, part := range parts {
			encoded[i] = URLEncodePath(part)
		}
		return encoded
	}

	encodedTarget := encodeAll(targetParts)

	// C1 case: current file is {basename}.svg at the root; target file is the
	// full encoded target path under {basename}/.
	if currentPath == "" {
		return basename + "/" + strings.Join(encodedTarget, "/") + "." + linkFormat
	}

	// For C2/C3+, compute a bidirectional relative URL from the current file's
	// directory to the target file. Both are treated as files whose directory is
	// all segments except the last (the filename segment).
	curFileDir := currentParts[:len(currentParts)-1] // directory containing current file
	tgtFileDir := targetParts[:len(targetParts)-1]   // directory containing target file

	// Length of the common directory prefix shared by the two directories.
	commonDirLen := 0
	for i := 0; i < len(curFileDir) && i < len(tgtFileDir); i++ {
		if curFileDir[i] == tgtFileDir[i] {
			commonDirLen++
		} else {
			break
		}
	}

	// Levels to climb from the current file's directory to the common ancestor.
	levelsUp := len(curFileDir) - commonDirLen

	upPath := strings.Repeat("../", levelsUp)

	// Remaining directory segments to descend from the common ancestor into the
	// target file's directory, then the URL-encoded filename segment.
	downDirSegs := encodedTarget[commonDirLen : len(targetParts)-1]
	lastTarget := encodedTarget[len(targetParts)-1]

	var downPath string
	if len(downDirSegs) > 0 {
		downPath = strings.Join(downDirSegs, "/") + "/" + lastTarget + "." + linkFormat
	} else {
		downPath = lastTarget + "." + linkFormat
	}

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
