package render

import (
	"fmt"
	"html"
	"strings"

	"github.com/Djarvur/c4drill/internal/graph"
)

// pngLink is one navigation anchor in the PNG's HTML wrapper doc: a
// drill-down target or an external reference.
type pngLink struct {
	// Name is the anchor text (the unit's display label).
	Name string
	// URL is the link target: a graph-layer relative doc path for
	// drill-downs (.svg, rewritten to .html here) or an absolute external
	// URL for references.
	URL string
}

// RenderHTMLForPNG builds the HTML navigation document that accompanies a
// rendered PNG diagram (issue #26). A bare PNG cannot carry hyperlinks, so
// each raster gains a sibling HTML doc — the interactive navigation layer —
// which:
//
//   - embeds the PNG via <img src="pngName"> (the doc sits next to its
//     image, so pngName is the bare file name, e.g. "model.png"),
//   - re-emits the graph's breadcrumb as real HTML anchors above the image
//     (the in-raster nav bar is baked, non-clickable pixels),
//   - lists one drill-down link per drill-capable unit (Node/Cluster
//     ExploreURL) to the child diagram's page,
//   - lists external reference URLs as anchors opening in a new tab.
//
// All internal link targets reuse the graph layer's computed relative URLs
// (ComputeExploreURL / breadcrumbs), rewritten from .svg to .html — the same
// svgHrefSuffix rewrite the SVG->HTML wrapper applies. The ".svg links
// regardless of rendered format" contract (path.go) stays intact; only this
// HTML layer rewrites. pngName is used verbatim as the img src; call it with
// the file name output.PNGImageName / output.ExpandedPNGImageName compute.
//
//nolint:revive // Function name matches the RenderSVG/RenderHTML pattern
func RenderHTMLForPNG(g *graph.Graph, pngName string) ([]byte, error) {
	if g == nil {
		return nil, ErrNilGraph
	}

	if pngName == "" {
		return nil, fmt.Errorf("%w: empty png file name", ErrNilGraph)
	}

	return []byte(wrapPNGInHTML(g, pngName)), nil
}

// wrapPNGInHTML assembles the wrapper document around the graph's navigation
// data. The image itself is referenced, not inlined — the sibling .png file
// is the raster this doc navigates.
func wrapPNGInHTML(g *graph.Graph, pngName string) string {
	title := g.Title
	if title == "" {
		title = pngName
	}

	var b strings.Builder
	b.WriteString(`<!DOCTYPE html>` + "\n")
	b.WriteString(`<html lang="en">` + "\n")
	b.WriteString(`<head><meta charset="utf-8">` + "\n")
	b.WriteString(`<title>` + html.EscapeString(title) + `</title>` + "\n")
	b.WriteString(`<style>` +
		`html,body{margin:0;padding:0;background:#fff;font-family:sans-serif}` +
		`img{display:block;max-width:100%;height:auto}` +
		`nav{margin:0.5em 1em}` +
		`nav.breadcrumb{color:#666666;font-size:0.9em}` +
		`nav a{color:#0366d6}` +
		`ul{margin:0.2em 0;padding-left:1.4em}` +
		`.current{font-weight:bold;color:#000}` +
		`</style>` + "\n")
	b.WriteString(`</head>` + "\n")
	b.WriteString(`<body>` + "\n")
	b.WriteString(pngBreadcrumbNav(g, title) + "\n")
	b.WriteString(`<img src="` + html.EscapeString(pngName) +
		`" alt="` + html.EscapeString(title) + `">` + "\n")
	b.WriteString(pngExploreNav(g) + "\n")
	b.WriteString(`</body></html>` + "\n")

	return b.String()
}

// pngBreadcrumbNav renders the clickable breadcrumb trail. Every ancestor is
// an <a> to its .html doc page (the graph layer's .svg URL rewritten); the
// current level is plain. C1 and --expanded graphs carry no Navigation, so
// the trail collapses to the current title alone.
func pngBreadcrumbNav(g *graph.Graph, title string) string {
	items := make([]string, 0, 4)

	if g.Navigation != nil {
		for _, crumb := range g.Navigation.Breadcrumbs {
			items = append(items, pngBreadcrumbItem(crumb))
		}
	}

	if len(items) == 0 {
		items = append(items, `<span class="current">`+html.EscapeString(title)+`</span>`)
	}

	return `<nav class="breadcrumb">` + strings.Join(items, " &#8594; ") + `</nav>`
}

// pngBreadcrumbItem renders one breadcrumb entry: ancestors as anchors to
// their .html doc page, the current level (empty URL) as plain text. Names
// and URLs are HTML-escaped (WR-01: a name or URL containing <, >, or &
// must not break the markup).
func pngBreadcrumbItem(item graph.BreadcrumbItem) string {
	name := html.EscapeString(item.Name)
	if item.URL == "" {
		return `<span class="current">` + name + `</span>`
	}

	return `<a href="` + html.EscapeString(pngDocURL(item.URL)) + `">` + name + `</a>`
}

// pngExploreNav renders the drill-down and reference link lists under the
// image. Omitted entirely when the graph carries neither — the C1 root with
// no drill-capable units and every --expanded diagram produce image-only
// docs.
func pngExploreNav(g *graph.Graph) string {
	drills, refs := pngNavLinks(g)
	if len(drills) == 0 && len(refs) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(`<nav class="links">`)

	if len(drills) > 0 {
		b.WriteString(`<ul class="drill">`)

		for _, l := range drills {
			b.WriteString(`<li><a href="` + html.EscapeString(pngDocURL(l.URL)) + `">` +
				html.EscapeString(l.Name) + `</a></li>`)
		}

		b.WriteString(`</ul>`)
	}

	if len(refs) > 0 {
		b.WriteString(`<ul class="refs">`)

		for _, l := range refs {
			b.WriteString(`<li>` + pngRefAnchor(l) + `</li>`)
		}

		b.WriteString(`</ul>`)
	}

	b.WriteString(`</nav>`)

	return b.String()
}

// pngRefAnchor renders one reference entry. Only http(s) and protocol-relative
// URLs become target="_blank" anchors — exactly the classes the SVG nav shim
// routes to window.open (REF-04). Any other scheme (javascript:, data:, ...)
// is untrusted per T-28-02 and renders as inert plain text: the static doc
// has no click shim to no-op it, so the anchor must not exist at all.
func pngRefAnchor(l pngLink) string {
	text := html.EscapeString(l.Name)

	u := strings.ToLower(l.URL)
	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") &&
		!strings.HasPrefix(u, "//") {
		return text
	}

	return `<a href="` + html.EscapeString(l.URL) + `" target="_blank" rel="noopener">` +
		text + `</a>`
}

// pngNavLinks collects the wrapper's navigation targets from the graph:
// drill-down links from every drill-capable node/cluster (ExploreURL) and
// external references from entries whose URL slot is not spent on the
// drill-down — mirroring the converter's single-URL-slot precedence
// (converter.go: a GraphViz node carries ONE URL; the drill-down wins, and
// the unit's docs render on its own child diagram). The walk recurses into
// cluster nodes and nested clusters (CTX-03) so nothing drill-capable is
// missed.
func pngNavLinks(g *graph.Graph) ([]pngLink, []pngLink) {
	drills, refs := pngLinksFromNodes(g.Nodes)

	return pngLinksFromClusters(g.Clusters, drills, refs)
}

// pngLinksFromNodes files every non-nil node under its single URL slot:
// an ExploreURL is a drill-down, otherwise a ReferenceURL is an external
// reference.
func pngLinksFromNodes(nodes []*graph.Node) ([]pngLink, []pngLink) {
	var drills, refs []pngLink

	for _, n := range nodes {
		if n == nil {
			continue
		}

		switch {
		case n.ExploreURL != "":
			drills = append(drills, pngLink{Name: pngLinkName(n.Label, n.ID), URL: n.ExploreURL})
		case n.ReferenceURL != "":
			refs = append(refs, pngLink{Name: pngLinkName(n.Label, n.ID), URL: n.ReferenceURL})
		}
	}

	return drills, refs
}

// pngLinksFromClusters files each cluster by its URL slot, then recurses into
// the cluster's nodes and nested clusters, appending to the accumulated
// drills/refs slices so the whole tree lands in two flat lists.
func pngLinksFromClusters(clusters []*graph.Cluster, drills, refs []pngLink) ([]pngLink, []pngLink) {
	for _, c := range clusters {
		if c == nil {
			continue
		}

		switch {
		case c.ExploreURL != "":
			drills = append(drills, pngLink{Name: pngLinkName(c.Label, c.ID), URL: c.ExploreURL})
		case c.ReferenceURL != "":
			refs = append(refs, pngLink{Name: pngLinkName(c.Label, c.ID), URL: c.ReferenceURL})
		}

		d, r := pngLinksFromNodes(c.Nodes)
		drills = append(drills, d...)
		refs = append(refs, r...)

		drills, refs = pngLinksFromClusters(c.Clusters, drills, refs)
	}

	return drills, refs
}

// pngLinkName picks the anchor text for a link: the unit's display label when
// present, otherwise the unit path.
func pngLinkName(label *graph.Label, id string) string {
	if label != nil && label.Name != "" {
		return label.Name
	}

	return id
}

// pngDocURL rewrites a graph-layer navigation URL to its .html doc page.
// The graph layer always emits .svg links regardless of the rendered format
// (path.go Gap-3 contract), so a .svg suffix is rewritten to .html; anything
// else (no suffix) passes through unchanged.
func pngDocURL(u string) string {
	if strings.HasSuffix(u, ".svg") {
		return strings.TrimSuffix(u, ".svg") + ".html"
	}

	return u
}
