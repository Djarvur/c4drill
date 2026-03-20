// Package render provides functions to render graph structures to DOT and SVG formats.
package render

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/goccy/go-graphviz"
)

// Static errors for error checking.
var (
	ErrNilGraph         = errors.New("graph is nil")
	ErrUnsupportedFormat = errors.New("unsupported format")
)

// RenderDOT renders a graph to DOT format (XDOT).
//
//nolint:revive // Function name matches plan specification (04-01-PLAN.md)
func RenderDOT(g *graph.Graph) ([]byte, error) {
	return render(g, graphviz.XDOT, "")
}

// RenderSVG renders a graph to SVG format.
//
//nolint:revive // Function name matches plan specification (04-01-PLAN.md)
func RenderSVG(g *graph.Graph) ([]byte, error) {
	return render(g, graphviz.SVG, "")
}

// RenderSVGWithOutput renders a graph to SVG format with icon extraction.
// The outputDir is used as the base directory for the .icons/ subdirectory.
func RenderSVGWithOutput(g *graph.Graph, outputDir string) ([]byte, error) {
	return render(g, graphviz.SVG, outputDir)
}

// Render renders a graph to the specified format ("dot" or "svg").
// Returns an error for unsupported formats.
func Render(g *graph.Graph, format string) ([]byte, error) {
	switch format {
	case "dot":
		return RenderDOT(g)
	case "svg":
		return RenderSVG(g)
	default:
		return nil, fmt.Errorf("%w: %q (supported: dot, svg)", ErrUnsupportedFormat, format)
	}
}

// render is the internal render function that handles all formats.
func render(g *graph.Graph, format graphviz.Format, outputDir string) ([]byte, error) {
	if g == nil {
		return nil, ErrNilGraph
	}

	ctx := context.Background()

	gv, err := graphviz.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("create graphviz instance: %w", err)
	}
	defer gv.Close()

	cg, err := buildCgraph(gv, g, outputDir)
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
