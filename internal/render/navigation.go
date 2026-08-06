package render

import (
	"fmt"
	"html"
	"strings"

	"github.com/Djarvur/c4drill/internal/graph"
)

// Nav styling constants. The nav line is rendered in a smaller, muted font so
// it reads as secondary to the diagram title; clickable breadcrumb items are
// underlined so they are recognisable as links (Safari/Chrome/Firefox all
// render <U> inside GraphViz HTML labels as text-decoration: underline).
const (
	navFontPoint = "10"     // smaller than the 14pt title
	navFontColor = "#666666" // muted gray
)

// navFontOpen is the opening <FONT ...> tag wrapping every nav cell so the
// entire breadcrumb trail shares the secondary styling.
const navFontOpen = `<FONT POINT-SIZE="` + navFontPoint + `" COLOR="` + navFontColor + `">`

// BuildNavigationLabel creates a GraphViz HTML-like label for the navigation
// bar and returns it as a single-row <TABLE>.
//
// The navigation is breadcrumb-only: "Ancestor1 > Ancestor2 > Current", where
// each ancestor is a clickable TD (rendered underlined in a muted font) and the
// current level is plain. The earlier "Back to {parent}" affordance was dropped
// because it always duplicated the breadcrumb's nearest-ancestor link (same
// destination, same label) — the breadcrumb alone covers the up-navigation.
//
// Gap 2 (03-04): GraphViz HTML-like labels do NOT support <a href> tags — the
// label is silently dropped at render time. Clickable links inside an HTML
// label are expressed via the HREF attribute on a <TD> element, which GraphViz
// renders as an <a xlink:href="..."> anchor in SVG output. The plain-text
// breadcrumb names AND the HREF URLs are HTML-escaped so a name or URL
// containing <, >, or & cannot break the label (WR-01: url.PathEscape does not
// escape '&').
func BuildNavigationLabel(nav *graph.Navigation) string {
	if nav == nil {
		return ""
	}

	tds := navigationTDs(nav)
	if len(tds) == 0 {
		return ""
	}

	return "<TABLE BORDER=\"0\" CELLBORDER=\"0\" CELLSPACING=\"0\" CELLPADDING=\"0\" ALIGN=\"CENTER\">" +
		"<TR>" + strings.Join(tds, "") + "</TR></TABLE>"
}

// navigationTDs returns the <TD> elements that make up the breadcrumb bar:
// clickable ancestor items carry an HREF attribute; separators and the current
// level are plain (non-clickable) cells. Exposed so the graph label can merge
// navigation with the title in a single multi-row table.
func navigationTDs(nav *graph.Navigation) []string {
	if nav == nil {
		return nil
	}

	return breadcrumbTDs(nav.Breadcrumbs)
}

// breadcrumbTDs turns breadcrumb items into <TD> cells. The "→" separator
// between items is merged INTO the preceding item's cell (as trailing inline
// text) rather than emitted as a separate cell. This avoids GraphViz's column-
// sizing behavior: separate separator cells stretch to the column width and
// create large visual gaps between breadcrumb items. Merging the separator
// into the item cell packs the breadcrumb tightly.
//
// The last item (current level) has no trailing separator.
func breadcrumbTDs(items []graph.BreadcrumbItem) []string {
	if len(items) == 0 {
		return nil
	}

	var tds []string

	for i, item := range items {
		// Leading separator on all items except the first (merged into the
		// item's cell as a prefix, avoiding inter-cell gaps that GraphViz
		// adds between trailing separators and the next cell's content).
		isFirst := i == 0
		prefix := ""
		if !isFirst {
			prefix = "&#8594;"
		}

		tds = append(tds, breadcrumbItemTD(item, prefix))
	}

	return tds
}

// breadcrumbItemTD renders a single breadcrumb item as a clickable or plain TD.
// Clickable items are wrapped in <U> (underline) inside the muted <FONT> so
// they are recognisable as links. The leading separator (→) is prepended
// inside the same cell to keep the breadcrumb compact.
func breadcrumbItemTD(item graph.BreadcrumbItem, prefix string) string {
	escaped := html.EscapeString(item.Name)
	if item.URL == "" {
		return plainNavTD(prefix + escaped)
	}

	// HTML-escape the URL for the HREF attribute (WR-01): url.PathEscape does
	// not escape '&', which GraphViz's HTML-like label parser rejects, silently
	// dropping the breadcrumb anchor.
	return fmt.Sprintf(`<TD HREF="%s">%s%s<U>%s</U>%s</TD>`,
		html.EscapeString(item.URL), navFontOpen, prefix, escaped, "</FONT>")
}

// plainNavTD wraps a literal HTML fragment (already escaped or a known entity)
// in a non-clickable table cell with the muted nav font styling.
func plainNavTD(content string) string {
	return "<TD>" + navFontOpen + content + "</FONT></TD>"
}

// plainTD wraps a literal HTML fragment (already escaped) in a non-clickable,
// unstyled table cell. Used for the diagram title, which should render at the
// default font size/color (not the muted nav styling).
func plainTD(content string) string {
	return "<TD>" + content + "</TD>"
}
