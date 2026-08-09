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
// targets uniformly. Returns the empty string when targetPath is empty or
// equals currentPath (self-link / empty-target guard — see Gap 1 symptom B,
// WR-02) so the renderer omits the URL rather than emitting a broken ".svg".
func ComputeExploreURL(currentPath, targetPath, basename, _ string) string {
	// Always use SVG for clickable links (browser navigation)
	const linkFormat = "svg"

	// Self-link / empty-target guard (Gap 1 symptom B, generalized — WR-02): a
	// target equal to the current path, or an empty target, cannot yield a valid
	// exploration URL — without this, an empty target collapses to a broken
	// ".svg" (C2 branch) or "basename/.svg" (C1 branch). Return empty so the
	// caller (createNode) skips SetURL entirely. The guard is symmetric: it must
	// fire for empty targets regardless of whether currentPath is empty.
	if targetPath == "" || targetPath == currentPath {
		return ""
	}

	// Split dotted paths into segments and URL-encode each segment.
	currentParts := strings.Split(currentPath, ".")
	targetParts := strings.Split(targetPath, ".")
	encodedTarget := encodePathSegments(targetParts)

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
	commonDirLen := commonDirectoryPrefixLength(curFileDir, tgtFileDir)

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
// format is the output format being generated; it is IGNORED — clickable
// navigation URLs always use ".svg" for browser navigation regardless of the
// render format (matching ComputeExploreURL), so a .dot diagram still links to
// the browser-navigable .svg sibling. The parameter is retained to avoid
// rippling signature changes through the builder call sites.
func ComputeBackLinkURL(currentPath, basename, _ string) string {
	// Always use SVG for clickable links (browser navigation) — Gap 3.
	const linkFormat = "svg"

	// C1 has no parent
	if currentPath == "" {
		return ""
	}

	parts := strings.Split(currentPath, ".")
	if len(parts) == 1 {
		// C2 level - go back to C1
		return "../" + basename + "." + linkFormat
	}
	// C3+ level - go back one level to parent directory
	// The parent file is named after the parent unit (last but one part)
	parentName := parts[len(parts)-2]

	return "../" + URLEncodePath(parentName) + "." + linkFormat
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
// format is the output format being generated; it is IGNORED — clickable
// navigation URLs always use ".svg" for browser navigation regardless of the
// render format (matching ComputeExploreURL), so a .dot diagram still links to
// the browser-navigable .svg sibling. The parameter is retained to avoid
// rippling signature changes through the builder call sites.
//
// The relative path from the current diagram's directory to the ancestor file
// is computed with the same bidirectional algorithm as ComputeExploreURL so
// that intermediate ancestors (e.g. the C2 "mainSystem" ancestor inside a C3
// "mainSystem.sshAuth" diagram) resolve correctly. The pre-fix code assumed
// the index-0 segment was always the C1 root, which 404'd for C3+ breadcrumbs.
func computeBreadcrumbURL(parts []string, ancestorIndex int, basename, _ string) string {
	// Reconstruct the ancestor's full dotted path and reuse ComputeExploreURL's
	// bidirectional relative-path logic. parts is the current diagram's path
	// split into segments, so the ancestor path is parts[:ancestorIndex+1].
	currentPath := strings.Join(parts, ".")
	ancestorPath := strings.Join(parts[:ancestorIndex+1], ".")

	return ComputeExploreURL(currentPath, ancestorPath, basename, "svg")
}

// URLEncodePath URL-encodes a path segment for use in URLs.
func URLEncodePath(path string) string {
	return url.PathEscape(path)
}

// encodePathSegments URL-encodes each segment of a dotted-path split.
func encodePathSegments(parts []string) []string {
	encoded := make([]string, len(parts))
	for i, part := range parts {
		encoded[i] = URLEncodePath(part)
	}

	return encoded
}

// commonDirectoryPrefixLength returns the length of the shared leading segment
// prefix of two directory-segment slices.
func commonDirectoryPrefixLength(a, b []string) int {
	commonDirLen := 0

	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] != b[i] {
			break
		}

		commonDirLen++
	}

	return commonDirLen
}
