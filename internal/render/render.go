// Package render provides functions to render graph structures to DOT, SVG, and HTML formats.
package render

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/goccy/go-graphviz"
)

// wasmMutex protects all WASM graphviz operations from concurrent access.
// The go-graphviz library uses a WASM engine that is not thread-safe,
// so all render operations must be serialized to prevent memory corruption.
//
//nolint:gochecknoglobals // Required for thread-safe WASM operations across all render calls
var wasmMutex sync.Mutex

// Static errors for error checking.
var (
	ErrNilGraph          = errors.New("graph is nil")
	ErrUnsupportedFormat = errors.New("unsupported format")
)

// RenderDOT renders a graph to DOT format (XDOT).
//
//nolint:revive // Function name matches plan specification (04-01-PLAN.md)
func RenderDOT(g *graph.Graph) ([]byte, error) {
	return render(g, graphviz.XDOT)
}

// RenderSVG renders a graph to SVG format.
//
//nolint:revive // Function name matches plan specification (04-01-PLAN.md)
func RenderSVG(g *graph.Graph) ([]byte, error) {
	return render(g, graphviz.SVG)
}

// RenderHTML renders a graph to a self-contained HTML document that inlines
// the SVG and injects a small JS shim so diagram links work in Safari/WebKit.
//
// Safari silently ignores <a> navigation inside SVG (both xlink:href and plain
// href, whether standalone or inline). The HTML wrapper restores navigation by
// attaching click listeners that call window.location.href. Without this, the
// drill-down links that c4drill emits are unclickable in Safari.
//
// The wrapper also rewrites .svg hrefs to .html so cross-diagram navigation
// stays within the HTML format. The SVG bytes are produced by the standard
// graphviz SVG renderer and then post-processed.
//
//nolint:revive // Function name matches plan specification
func RenderHTML(g *graph.Graph) ([]byte, error) {
	svgBytes, err := render(g, graphviz.SVG)
	if err != nil {
		return nil, err
	}

	return wrapSVGInHTML(svgBytes), nil
}

// RenderSVGWithOutput renders a graph to SVG format.
// The outputDir parameter is kept for API compatibility but is no longer used
// since icons are now handled via native GraphViz shapes and emoji.
//
//nolint:revive // Function name matches plan specification (04-01-PLAN.md)
func RenderSVGWithOutput(g *graph.Graph, _ string) ([]byte, error) {
	return render(g, graphviz.SVG)
}

// Render renders a graph to the specified format ("dot", "svg", or "html").
// Returns an error for unsupported formats.
func Render(g *graph.Graph, format string) ([]byte, error) {
	switch format {
	case "dot":
		return RenderDOT(g)
	case "svg":
		return RenderSVG(g)
	case "html":
		return RenderHTML(g)
	default:
		return nil, fmt.Errorf("%w: %q (supported: dot, svg, html)", ErrUnsupportedFormat, format)
	}
}

// RenderWithOutput renders a graph to the specified format.
// The outputDir parameter is kept for API compatibility but is no longer used
// since icons are now handled via native GraphViz shapes and emoji.
//
//nolint:revive // Function name matches plan specification (04-01-PLAN.md)
func RenderWithOutput(g *graph.Graph, format, _ string) ([]byte, error) {
	switch format {
	case "dot":
		return render(g, graphviz.XDOT)
	case "svg":
		return render(g, graphviz.SVG)
	case "html":
		return RenderHTML(g)
	default:
		return nil, fmt.Errorf("%w: %q (supported: dot, svg, html)", ErrUnsupportedFormat, format)
	}
}

// render is the internal render function that handles all formats.
func render(g *graph.Graph, format graphviz.Format) ([]byte, error) {
	if g == nil {
		return nil, ErrNilGraph
	}

	// Lock to prevent concurrent WASM access (go-graphviz uses WASM engine)
	wasmMutex.Lock()
	defer wasmMutex.Unlock()

	ctx := context.Background()

	gv, err := graphviz.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("create graphviz instance: %w", err)
	}
	defer gv.Close()

	cg, err := buildCgraph(gv, g)
	if err != nil {
		return nil, fmt.Errorf("build graph: %w", err)
	}
	defer cg.Close()

	var buf bytes.Buffer
	if err := gv.Render(ctx, cg, format, &buf); err != nil {
		return nil, fmt.Errorf("render output: %w", err)
	}

	return buf.Bytes(), nil
}

// xmlDeclOrDoctype matches the XML declaration (<?xml ...?>) and the SVG
// DOCTYPE declaration that graphviz emits. Both are invalid inside an HTML
// body and must be stripped before inlining the SVG.
//
//nolint:gochecknoglobals // Compiled once at init; regexp.Compile at every call would be wasteful
var xmlDeclOrDoctype = regexp.MustCompile(`(?s)^\s*(<\?xml[^>]*\?>\s*)?(<!DOCTYPE[^>]*>\s*)?`)

// svgHrefSuffix matches the .svg extension inside an href/xlink:href attribute
// value, capturing the quote so we can rewrite it to .html. GraphViz always
// emits double-quoted attribute values.
//
//nolint:gochecknoglobals // Compiled once at init
var svgHrefSuffix = regexp.MustCompile(`((?:xlink:)?href="[^"]+)\.svg(")`)

// htmlNavShim is the JS injected into every HTML-wrapped diagram. Safari
// ignores default navigation on SVG <a> elements; this shim attaches click
// listeners that navigate via window.location.href, restoring drill-down
// navigation in WebKit browsers. Also sets pointer cursor on links.
//
// REF-04: external reference URLs (📖, http(s):// or protocol-relative //) open
// in a new tab via window.open, distinct from internal drill-down navigation
// (.html after svgHrefSuffix rewrite) which stays in the same tab. T-28-02
// hardening: any href whose scheme is NOT http(s) and which is not a relative
// internal path is treated as untrusted (e.g. javascript:/data:) and the click
// is a no-op (preventDefault without navigation).
const htmlNavShim = `<script>(function(){
function go(e){var a=e.currentTarget;var h=a.getAttribute("href")||a.getAttributeNS("http://www.w3.org/1999/xlink","href");if(!h){return;}e.preventDefault();if(/^https?:\/\//.test(h)||h.indexOf("//")===0){window.open(h,"_blank");}else if(/^(?:[a-z]+:)?\/\//i.test(h)){return;}else{window.location.href=h;}}
function init(){var l=document.querySelectorAll("svg a");for(var i=0;i<l.length;i++){l[i].style.cursor="pointer";l[i].addEventListener("click",go);}}
if(document.readyState==="loading"){document.addEventListener("DOMContentLoaded",init);}else{init();}
})();</script>`

// wrapSVGInHTML post-processes raw graphviz SVG bytes into a self-contained
// HTML document with working cross-diagram navigation in all browsers
// (including Safari, which silently ignores SVG <a> navigation).
//
// Transformations applied:
//  1. Strip the XML declaration and DOCTYPE (invalid inside HTML body).
//  2. Rewrite every href="X.svg" / xlink:href="X.svg" to X.html so wrapped
//     diagrams link to wrapped siblings, not standalone .svg files.
//  3. Wrap in an HTML document with a minimal style and the nav shim.
//
// titleBytes is the optional document title derived from the graph; pass nil
// to omit (the renderer currently does not surface a title here).
func wrapSVGInHTML(svgBytes []byte) []byte {
	svg := string(svgBytes)

	// 1. Strip XML declaration and DOCTYPE
	svg = xmlDeclOrDoctype.ReplaceAllString(svg, "")

	// 2. Rewrite .svg hrefs → .html
	svg = svgHrefSuffix.ReplaceAllString(svg, "${1}.html${2}")

	// 3. Wrap in HTML with the nav shim
	var b strings.Builder
	b.WriteString(`<!DOCTYPE html>` + "\n")
	b.WriteString(`<html lang="en">` + "\n")
	b.WriteString(`<head><meta charset="utf-8">` + "\n")
	b.WriteString(`<style>html,body{margin:0;padding:0;background:#fff}` +
		`svg{display:block;max-width:100%;height:auto}` +
		`svg a{cursor:pointer}</style>` + "\n")
	b.WriteString(`</head>` + "\n")
	b.WriteString(`<body>` + "\n")
	b.WriteString(svg)
	b.WriteString("\n")
	b.WriteString(htmlNavShim)
	b.WriteString("\n")
	b.WriteString(`</body></html>` + "\n")

	return []byte(b.String())
}
