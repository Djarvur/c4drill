package render

import (
	"fmt"
	"html"
	"strings"

	"github.com/Djarvur/c4drill/internal/graph"
)

// BuildNavigationLabel creates a GraphViz HTML-like label for the navigation
// bar and returns it as a single-row <TABLE>.
//
// Gap 2 (03-04): GraphViz HTML-like labels do NOT support <a href> tags — the
// label is silently dropped at render time. Clickable links inside an HTML
// label are expressed via the HREF attribute on a <TD> element, which GraphViz
// renders as an <a xlink:href="..."> anchor in SVG output (and an imagemap/URL
// in other formats). The plain-text back-link label, breadcrumb names, AND the
// HREF URLs are HTML-escaped so a name or URL containing <, >, or & cannot
// break the label (WR-01: url.PathEscape does not escape '&').
//
// Format rendered: "Back to {parent} | Ancestor1 > Ancestor2 > Current"
// where the back-link and ancestor items are clickable TDs and the separators
// plus the current level are plain (non-clickable) TDs.
func BuildNavigationLabel(nav *graph.Navigation) string {
	if nav == nil {
		return ""
	}

	tds := navigationTDs(nav)
	if len(tds) == 0 {
		return ""
	}

	return "<TABLE BORDER=\"0\" CELLBORDER=\"0\" CELLSPACING=\"0\">" +
		"<TR>" + strings.Join(tds, "") + "</TR></TABLE>"
}

// navigationTDs returns the <TD> elements that make up the navigation bar:
// clickable back-link and breadcrumb ancestors carry an HREF attribute;
// separators and the current level are plain (non-clickable) cells. Exposed
// so the graph label can merge navigation with the title in a single multi-row table.
func navigationTDs(nav *graph.Navigation) []string {
	if nav == nil {
		return nil
	}

	var tds []string

	// Back-link
	if nav.BackLink != nil && nav.BackLink.URL != "" {
		label := html.EscapeString("Back to " + nav.BackLink.Name)
		// HTML-escape the URL too (WR-01): url.PathEscape does NOT escape '&',
		// and GraphViz's HTML-like label parser rejects a raw '&' that is not a
		// valid entity, silently dropping the navigation TD. html.EscapeString
		// escapes only <, >, &, ', " — already-encoded segments (%20, %28) are
		// unaffected, so legitimate URLs are left intact.
		href := html.EscapeString(nav.BackLink.URL)
		tds = append(tds, fmt.Sprintf(`<TD HREF="%s">%s</TD>`, href, label))
	}

	// Breadcrumbs
	tds = append(tds, breadcrumbTDs(nav.Breadcrumbs, len(tds) > 0)...)

	return tds
}

// breadcrumbTDs turns breadcrumb items into <TD> cells. hasBackLink controls
// the first separator: a "|" separates the back-link from the breadcrumb trail,
// while subsequent items are separated by ">".
func breadcrumbTDs(items []graph.BreadcrumbItem, hasBackLink bool) []string {
	if len(items) == 0 {
		return nil
	}

	var tds []string

	for i, item := range items {
		if i == 0 && hasBackLink {
			tds = append(tds, plainTD("|"))
		} else if i > 0 {
			tds = append(tds, plainTD("&gt;"))
		}

		tds = append(tds, breadcrumbItemTD(item))
	}

	return tds
}

// breadcrumbItemTD renders a single breadcrumb item as a clickable or plain TD.
func breadcrumbItemTD(item graph.BreadcrumbItem) string {
	escaped := html.EscapeString(item.Name)
	if item.URL == "" {
		return plainTD(escaped)
	}

	// HTML-escape the URL for the HREF attribute (WR-01): url.PathEscape does
	// not escape '&', which GraphViz's HTML-like label parser rejects, silently
	// dropping the breadcrumb anchor. See navigationTDs for the full rationale.
	return fmt.Sprintf(`<TD HREF="%s">%s</TD>`, html.EscapeString(item.URL), escaped)
}

// plainTD wraps a literal HTML fragment (already escaped or a known entity)
// in a non-clickable table cell.
func plainTD(content string) string {
	return "<TD>" + content + "</TD>"
}
