package render

import (
	"fmt"
	"strings"

	"github.com/Djarvur/c4drill/internal/graph"
)

// BuildNavigationLabel creates an HTML-like label for the navigation bar.
// Format: "Back to {parent} | Ancestor1 > Ancestor2 > Current"
// The back-link and ancestor items are clickable; current level is plain text.
func BuildNavigationLabel(nav *graph.Navigation) string {
	if nav == nil {
		return ""
	}

	var parts []string

	// Back-link
	if nav.BackLink != nil && nav.BackLink.URL != "" {
		backLink := fmt.Sprintf("<a href=\"%s\">%s</a>",
			nav.BackLink.URL,
			"Back to "+nav.BackLink.Name)
		parts = append(parts, backLink)
	}

	// Breadcrumbs
	if len(nav.Breadcrumbs) > 0 {
		var crumbs []string

		for _, item := range nav.Breadcrumbs {
			if item.URL != "" {
				// Clickable ancestor
				crumbs = append(crumbs, fmt.Sprintf("<a href=\"%s\">%s</a>",
					item.URL, item.Name))
			} else {
				// Current level - plain text
				crumbs = append(crumbs, item.Name)
			}
		}

		if len(crumbs) > 0 {
			parts = append(parts, strings.Join(crumbs, " > "))
		}
	}

	return strings.Join(parts, " | ")
}
