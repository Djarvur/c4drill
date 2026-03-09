package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraphStruct(t *testing.T) {
	t.Parallel()

	// Test 1: Graph struct contains Title, Direction, EdgeStyle, Nodes, Edges, Clusters, Legend
	g := &Graph{
		Title:     "Test Diagram",
		Direction: "TB",
		EdgeStyle: "spline",
		Nodes:     []*Node{},
		Edges:     []*Edge{},
		Clusters:  []*Cluster{},
		Legend:    &Legend{},
	}

	assert.Equal(t, "Test Diagram", g.Title)
	assert.Equal(t, "TB", g.Direction)
	assert.Equal(t, "spline", g.EdgeStyle)
	assert.NotNil(t, g.Nodes)
	assert.NotNil(t, g.Edges)
	assert.NotNil(t, g.Clusters)
	assert.NotNil(t, g.Legend)
}

func TestNodeStruct(t *testing.T) {
	t.Parallel()

	// Test 2: Node struct contains ID, Label, Shape, Style, IsExternal, IsInCluster
	n := &Node{
		ID:          "mainapp.api",
		Label:       &Label{Name: "API"},
		Shape:       ShapeHTML,
		Style:       &NodeStyle{FillColor: "#1168BD"},
		IsExternal:  false,
		IsInCluster: true,
	}

	assert.Equal(t, "mainapp.api", n.ID)
	assert.NotNil(t, n.Label)
	assert.Equal(t, ShapeHTML, n.Shape)
	assert.NotNil(t, n.Style)
	assert.False(t, n.IsExternal)
	assert.True(t, n.IsInCluster)
}

func TestEdgeStruct(t *testing.T) {
	t.Parallel()

	// Test 3: Edge struct contains Source, Target, Label, Style, ArrowHead
	e := &Edge{
		Source:    "mainapp.api",
		Target:    "mainapp.db",
		Label:     &EdgeLabel{Technology: "SQL"},
		Style:     "solid",
		ArrowHead: ArrowForward,
	}

	assert.Equal(t, "mainapp.api", e.Source)
	assert.Equal(t, "mainapp.db", e.Target)
	assert.NotNil(t, e.Label)
	assert.Equal(t, "solid", e.Style)
	assert.Equal(t, ArrowForward, e.ArrowHead)
}

func TestClusterStruct(t *testing.T) {
	t.Parallel()

	// Test 4: Cluster struct contains ID, Label, Nodes, Style
	c := &Cluster{
		ID:    "cluster_mainapp",
		Label: &Label{Name: "Main App"},
		Nodes: []*Node{},
		Style: &NodeStyle{FillColor: "#438DD5"},
	}

	assert.Equal(t, "cluster_mainapp", c.ID)
	assert.NotNil(t, c.Label)
	assert.NotNil(t, c.Nodes)
	assert.NotNil(t, c.Style)
}

func TestLabelStruct(t *testing.T) {
	t.Parallel()

	// Test 5: Label struct contains Name, Technology, Description, Icon
	l := &Label{
		Name:        "API Server",
		Technology:  "Go",
		Description: "Handles HTTP requests",
		Icon:        "\U0001F464",
	}

	assert.Equal(t, "API Server", l.Name)
	assert.Equal(t, "Go", l.Technology)
	assert.Equal(t, "Handles HTTP requests", l.Description)
	assert.Equal(t, "\U0001F464", l.Icon)
}

func TestNodeStyleStruct(t *testing.T) {
	t.Parallel()

	// Test 6: NodeStyle struct contains FillColor, BorderColor, FontColor, BorderStyle
	s := &NodeStyle{
		FillColor:   "#1168BD",
		BorderColor: "#3C7FC0",
		FontColor:   "#FFFFFF",
		BorderStyle: "solid",
	}

	assert.Equal(t, "#1168BD", s.FillColor)
	assert.Equal(t, "#3C7FC0", s.BorderColor)
	assert.Equal(t, "#FFFFFF", s.FontColor)
	assert.Equal(t, "solid", s.BorderStyle)
}

func TestShapeConstants(t *testing.T) {
	t.Parallel()

	// Test 7: Shape constants (ShapeRecord, ShapeHTML, ShapeCluster) are defined
	assert.Equal(t, Shape("record"), ShapeRecord)
	assert.Equal(t, Shape("html"), ShapeHTML)
	assert.Equal(t, Shape("cluster"), ShapeCluster)
}

func TestArrowDirectionConstants(t *testing.T) {
	t.Parallel()

	// Test 8: ArrowDirection constants (ArrowForward, ArrowReverse, ArrowBoth, ArrowNone) are defined
	assert.Equal(t, ArrowDirection("forward"), ArrowForward)
	assert.Equal(t, ArrowDirection("reverse"), ArrowReverse)
	assert.Equal(t, ArrowDirection("both"), ArrowBoth)
	assert.Equal(t, ArrowDirection("none"), ArrowNone)
}
