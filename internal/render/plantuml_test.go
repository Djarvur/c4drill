package render_test

import (
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stdlibBaseURL is the C4-PlantUML stdlib include base the generated files
// must reference (mirrors the renderer's package constant for assertions).
const stdlibBaseURL = "https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master"

// pumlNode builds a styled-free node of the given type with the given label.
func pumlNode(id string, t model.UnitType, name string) *graph.Node {
	return &graph.Node{
		ID:    id,
		Label: &graph.Label{Name: name},
		Shape: graph.ShapeHTML,
		Type:  t,
	}
}

// pumlEdge builds an edge between two node ids.
func pumlEdge(source, target string, arrow graph.ArrowDirection) *graph.Edge {
	return &graph.Edge{
		Source:    source,
		Target:    target,
		ArrowHead: arrow,
		Style:     "solid",
	}
}

// TestRenderPlantUML proves the document frame: nil-graph error, PlantUML
// start/end markers, the stdlib !include line, and dispatch parity for the
// "plantuml" format.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestRenderPlantUML(t *testing.T) {
	t.Run("returns error for nil graph", func(t *testing.T) {
		out, err := render.RenderPlantUML(nil)
		require.Error(t, err)
		assert.Nil(t, out)
	})

	t.Run("frames the document with startuml and the stdlib include", func(t *testing.T) {
		g := &graph.Graph{
			Title: "Doc",
			Nodes: []*graph.Node{pumlNode("a", model.TypeSystem, "A")},
		}

		out, err := render.RenderPlantUML(g)
		require.NoError(t, err)

		s := string(out)
		assert.True(t, strings.HasPrefix(s, "@startuml\n"), "document must open with @startuml")
		assert.Contains(t, s, "@enduml", "document must close with @enduml")
		assert.Contains(t, s,
			"!include https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master/C4_Context.puml",
			"document must include C4-PlantUML from the stdlib URL")
		assert.Contains(t, s, "title Doc", "the diagram title must be emitted")
	})

	t.Run("plantuml format dispatch returns same result as RenderPlantUML", func(t *testing.T) {
		g := &graph.Graph{
			Title: "Doc",
			Nodes: []*graph.Node{pumlNode("a", model.TypeSystem, "A")},
		}

		dispatched, err := render.Render(g, "plantuml")
		require.NoError(t, err)

		direct, err := render.RenderPlantUML(g)
		require.NoError(t, err)

		assert.Equal(t, direct, dispatched, "Render(g, 'plantuml') must equal RenderPlantUML(g)")
	})
}

// TestRenderPlantUML_UnitTypeMapping pins the issue #25 mapping table: every
// one of the 17 unit types serializes as its C4-PlantUML macro. Boxes map to
// boundaries when they render as clusters (the grouping-frame case) and to
// their level's element macro when collapsed to plain nodes.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestRenderPlantUML_UnitTypeMapping(t *testing.T) {
	t.Run("non-box types map to their element macro", func(t *testing.T) {
		cases := []struct {
			utype model.UnitType
			macro string
		}{
			{model.TypePerson, "Person"},
			{model.TypePersonExternal, "Person_Ext"},
			{model.TypeSystem, "System"},
			{model.TypeSystemExternal, "System_Ext"},
			{model.TypeDb, "SystemDb"},
			{model.TypeDbExternal, "SystemDb_Ext"},
			{model.TypeQueue, "SystemQueue"},
			{model.TypeQueueExternal, "SystemQueue_Ext"},
			{model.TypeContainer, "Container"},
			{model.TypeContainerDb, "ContainerDb"},
			{model.TypeContainerQueue, "ContainerQueue"},
			{model.TypeComponent, "Component"},
			{model.TypeComponentDb, "ComponentDb"},
			{model.TypeComponentQueue, "ComponentQueue"},
		}

		for _, tc := range cases {
			g := &graph.Graph{Nodes: []*graph.Node{pumlNode("u", tc.utype, "Unit")}}

			out, err := render.RenderPlantUML(g)
			require.NoError(t, err, "type %s", tc.utype)

			assert.Contains(t, string(out), tc.macro+"(u, \"Unit\"",
				"type %s must map to %s", tc.utype, tc.macro)
		}
	})

	t.Run("box types map to boundary macros when rendered as clusters", func(t *testing.T) {
		cases := []struct {
			utype  model.UnitType
			expect string
		}{
			{model.TypeBox, "System_Boundary(b, \"Box\""},
			{model.TypeContainerBox, "Container_Boundary(b, \"Box\""},
			// The C4-PlantUML stdlib defines no Component_Boundary; the
			// generic Boundary macro with type text "Component" stands in.
			{model.TypeComponentBox, "Boundary(b, \"Box\", \"Component\""},
		}

		for _, tc := range cases {
			g := &graph.Graph{
				Clusters: []*graph.Cluster{{
					ID:    "b",
					Type:  tc.utype,
					Label: &graph.Label{Name: "Box"},
				}},
			}

			out, err := render.RenderPlantUML(g)
			require.NoError(t, err, "type %s", tc.utype)

			assert.Contains(t, string(out), tc.expect,
				"cluster type %s must render as %s", tc.utype, tc.expect)
		}
	})

	t.Run("collapsed box nodes render as their level element macro", func(t *testing.T) {
		cases := []struct {
			utype model.UnitType
			macro string
		}{
			{model.TypeBox, "System"},
			{model.TypeContainerBox, "Container"},
			{model.TypeComponentBox, "Component"},
		}

		for _, tc := range cases {
			g := &graph.Graph{Nodes: []*graph.Node{pumlNode("b", tc.utype, "Box")}}

			out, err := render.RenderPlantUML(g)
			require.NoError(t, err, "type %s", tc.utype)

			assert.Contains(t, string(out), tc.macro+"(b, \"Box\"",
				"collapsed node type %s must render as %s", tc.utype, tc.macro)
		}
	})

	t.Run("system and container clusters frame with their boundary macro", func(t *testing.T) {
		g := &graph.Graph{
			Clusters: []*graph.Cluster{
				{ID: "sys", Type: model.TypeSystem, Label: &graph.Label{Name: "Sys"}},
				{ID: "ctr", Type: model.TypeContainer, Label: &graph.Label{Name: "Ctr"}},
			},
		}

		out, err := render.RenderPlantUML(g)
		require.NoError(t, err)

		s := string(out)
		assert.Contains(t, s, "System_Boundary(sys, \"Sys\") {", "system cluster must frame with System_Boundary")
		assert.Contains(t, s, "Container_Boundary(ctr, \"Ctr\") {", "container cluster must frame with Container_Boundary")
	})
}

// TestRenderPlantUML_IncludeSelection proves the single !include line targets
// the deepest level the diagram needs (each C4 file transitively includes the
// levels below it).
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestRenderPlantUML_IncludeSelection(t *testing.T) {
	cases := []struct {
		name    string
		utype   model.UnitType
		include string
	}{
		{"C1-only diagram includes C4_Context", model.TypePerson, "C4_Context.puml"},
		{"container diagram includes C4_Container", model.TypeContainer, "C4_Container.puml"},
		{"component diagram includes C4_Component", model.TypeComponent, "C4_Component.puml"},
		{"box cluster keeps C4_Context", model.TypeBox, "C4_Context.puml"},
		{"containerBox cluster raises to C4_Container", model.TypeContainerBox, "C4_Container.puml"},
		{"componentBox cluster raises to C4_Component", model.TypeComponentBox, "C4_Component.puml"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := &graph.Graph{Nodes: []*graph.Node{pumlNode("u", tc.utype, "Unit")}}

			out, err := render.RenderPlantUML(g)
			require.NoError(t, err)

			assert.Contains(t, string(out), stdlibBaseURL+"/"+tc.include,
				"deepest level must select the include")
			assert.NotContains(t, string(out), "!include ./", "no relative includes")
			assert.Equal(t, 1, strings.Count(string(out), "!include"), "exactly one include line")
		})
	}
}

// TestRenderPlantUML_DrillLinks proves the core requirement (issue #25):
// every drill-capable node/cluster carries its ComputeExploreURL in the
// element macro's link slot (expanded to [[url]] markup by C4-PlantUML, so
// the converted SVG is clickable), with the converter's single-URL-slot
// precedence — drill-down beats the external reference.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestRenderPlantUML_DrillLinks(t *testing.T) {
	t.Run("drill-capable node carries its explore URL", func(t *testing.T) {
		g := &graph.Graph{
			Nodes: []*graph.Node{pumlNode("sys.child", model.TypeSystem, "Child")},
		}
		g.Nodes[0].ExploreURL = "basename/sys/child.svg"

		out, err := render.RenderPlantUML(g)
		require.NoError(t, err)

		assert.Contains(t, string(out), `System(sys_child, "Child", "", "", "", "basename/sys/child.svg")`,
			"the explore URL must ride in the macro's link slot (after the empty tags slot)")
	})

	t.Run("drill-down wins the single URL slot over the reference", func(t *testing.T) {
		g := &graph.Graph{
			Nodes: []*graph.Node{pumlNode("sys.child", model.TypeSystem, "Child")},
		}
		g.Nodes[0].ExploreURL = "basename/sys/child.svg"
		g.Nodes[0].ReferenceURL = "https://example.com/docs"

		out, err := render.RenderPlantUML(g)
		require.NoError(t, err)

		s := string(out)
		assert.Contains(t, s, "basename/sys/child.svg", "drill-down must be the link")
		assert.NotContains(t, s, "https://example.com/docs",
			"the reference must not share the single URL slot")
	})

	t.Run("leaf with a reference carries the external URL", func(t *testing.T) {
		g := &graph.Graph{
			Nodes: []*graph.Node{pumlNode("leaf", model.TypeComponent, "Leaf")},
		}
		g.Nodes[0].ReferenceURL = "https://example.com/docs"

		out, err := render.RenderPlantUML(g)
		require.NoError(t, err)

		assert.Contains(t, string(out), `"https://example.com/docs"`,
			"external references stay clickable links")
	})

	t.Run("drill-capable cluster carries its explore URL on the boundary", func(t *testing.T) {
		g := &graph.Graph{
			Clusters: []*graph.Cluster{{
				ID:         "sys",
				Type:       model.TypeSystem,
				Label:      &graph.Label{Name: "Sys"},
				ExploreURL: "basename/sys.svg",
			}},
		}

		out, err := render.RenderPlantUML(g)
		require.NoError(t, err)

		assert.Contains(t, string(out), `System_Boundary(sys, "Sys", "", "basename/sys.svg") {`,
			"the boundary's link slot must carry the drill-down")
	})
}

// TestRenderPlantUML_EdgeDirections pins the four arrow directions onto the
// C4-PlantUML relationship macros (issue #25): forward -> Rel, reverse ->
// Rel_Back (arrowhead on the `from` end = the c4drill source), bidirectional
// -> BiRel, none -> the arrowless Rel_ variant.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestRenderPlantUML_EdgeDirections(t *testing.T) {
	cases := []struct {
		name   string
		arrow  graph.ArrowDirection
		expect string
	}{
		{"forward renders Rel", graph.ArrowForward, "Rel(a, b, \"\")"},
		{"reverse renders Rel_Back", graph.ArrowReverse, "Rel_Back(a, b, \"\")"},
		{"bidirectional renders BiRel", graph.ArrowBoth, "BiRel(a, b, \"\")"},
		{"none renders the arrowless Rel_", graph.ArrowNone, "Rel_(a, b, \"\", \"--\")"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := &graph.Graph{
				Nodes: []*graph.Node{
					pumlNode("a", model.TypeSystem, "A"),
					pumlNode("b", model.TypeSystem, "B"),
				},
				Edges: []*graph.Edge{pumlEdge("a", "b", tc.arrow)},
			}

			out, err := render.RenderPlantUML(g)
			require.NoError(t, err)

			assert.Contains(t, string(out), tc.expect, "arrow %q must map onto its macro", tc.arrow)
		})
	}
}

// TestRenderPlantUML_EdgeLabelsAndStyles proves technology/description ride
// the Rel label slots and colour/style/thickness are carried through
// AddRelTag declarations (issue #25: into the parameters the macros support).
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestRenderPlantUML_EdgeLabelsAndStyles(t *testing.T) {
	t.Run("description and technology ride the label slots", func(t *testing.T) {
		g := &graph.Graph{
			Nodes: []*graph.Node{
				pumlNode("a", model.TypeSystem, "A"),
				pumlNode("b", model.TypeSystem, "B"),
			},
			Edges: []*graph.Edge{{
				Source:    "a",
				Target:    "b",
				ArrowHead: graph.ArrowForward,
				Style:     "solid",
				Label:     &graph.EdgeLabel{Technology: "HTTP", Description: "calls"},
			}},
		}

		out, err := render.RenderPlantUML(g)
		require.NoError(t, err)

		assert.Contains(t, string(out), `Rel(a, b, "calls", "HTTP")`,
			"description must be the Rel label, technology the techn slot")
	})

	t.Run("coloured dashed edge gains an AddRelTag declaration", func(t *testing.T) {
		g := &graph.Graph{
			Nodes: []*graph.Node{
				pumlNode("a", model.TypeSystem, "A"),
				pumlNode("b", model.TypeSystem, "B"),
			},
			Edges: []*graph.Edge{{
				Source:    "a",
				Target:    "b",
				ArrowHead: graph.ArrowForward,
				Style:     "dashed",
				Color:     "#FF0000",
			}},
		}

		out, err := render.RenderPlantUML(g)
		require.NoError(t, err)

		s := string(out)
		assert.Contains(t, s, `AddRelTag("c4drillRel0", $textColor="#FF0000", $lineColor="#FF0000", $lineStyle="dashed")`,
			"the style tag must be declared before use")
		assert.Contains(t, s, `Rel(a, b, "", "", "", "", "c4drillRel0")`,
			"the edge must reference its tag")
		assert.Less(t, strings.Index(s, "AddRelTag"), strings.Index(s, "Rel(a, b"),
			"tag declarations must precede their first use")
	})

	t.Run("doubled thickness of collapsed pairs becomes lineThickness", func(t *testing.T) {
		g := &graph.Graph{
			Nodes: []*graph.Node{
				pumlNode("a", model.TypeSystem, "A"),
				pumlNode("b", model.TypeSystem, "B"),
			},
			Edges: []*graph.Edge{{
				Source:    "a",
				Target:    "b",
				ArrowHead: graph.ArrowForward,
				Style:     "solid",
				PenWidth:  2.0,
			}},
		}

		out, err := render.RenderPlantUML(g)
		require.NoError(t, err)

		assert.Contains(t, string(out), `$lineThickness="2"`, "penwidth 2 must become lineThickness 2")
	})

	t.Run("arrowless edges keep the plain line without tags", func(t *testing.T) {
		g := &graph.Graph{
			Nodes: []*graph.Node{
				pumlNode("a", model.TypeSystem, "A"),
				pumlNode("b", model.TypeSystem, "B"),
			},
			Edges: []*graph.Edge{{
				Source:    "a",
				Target:    "b",
				ArrowHead: graph.ArrowNone,
				Style:     "dashed",
				Color:     "#FF0000",
				Label:     &graph.EdgeLabel{Technology: "HTTP", Description: "calls"},
			}},
		}

		out, err := render.RenderPlantUML(g)
		require.NoError(t, err)

		s := string(out)
		assert.Contains(t, s, `Rel_(a, b, "calls", "HTTP", "--")`,
			"the arrowless variant must carry label and technology")
		assert.NotContains(t, s, "AddRelTag", "Rel_ has no tags slot — no tag may be declared")
	})

	t.Run("edges with missing endpoints are skipped", func(t *testing.T) {
		g := &graph.Graph{
			Nodes: []*graph.Node{pumlNode("a", model.TypeSystem, "A")},
			Edges: []*graph.Edge{pumlEdge("a", "ghost", graph.ArrowForward)},
		}

		out, err := render.RenderPlantUML(g)
		require.NoError(t, err)

		assert.NotContains(t, string(out), "Rel(", "edges to unrendered endpoints must be skipped")
	})
}

// TestRenderPlantUML_ElementStyles proves author colouring reaches the C4
// macros through AddElementTag declarations (the only colour channel the
// element macros support).
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestRenderPlantUML_ElementStyles(t *testing.T) {
	g := &graph.Graph{
		Nodes: []*graph.Node{{
			ID:    "a",
			Label: &graph.Label{Name: "A"},
			Shape: graph.ShapeHTML,
			Type:  model.TypeContainer,
			Style: &graph.NodeStyle{
				FillColor:   "#1168BD",
				BorderColor: "#0B4884",
				FontColor:   "#FFFFFF",
				BorderStyle: "dashed",
			},
		}},
	}

	out, err := render.RenderPlantUML(g)
	require.NoError(t, err)

	s := string(out)
	assert.Contains(t, s,
		`AddElementTag("c4drillElem0", $bgColor="#1168BD", $fontColor="#FFFFFF",`+
			` $borderColor="#0B4884", $borderStyle="dashed")`,
		"the element style must be declared as an AddElementTag")
	assert.Contains(t, s, `"c4drillElem0"`, "the element must reference its tag")
	assert.Less(t, strings.Index(s, "AddElementTag"), strings.Index(s, "Container("),
		"tag declarations must precede their first use")
}

// TestRenderPlantUML_Legend proves LAYOUT_WITH_LEGEND() is emitted only when
// the graph carries a legend (properties.legend on), never otherwise.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestRenderPlantUML_Legend(t *testing.T) {
	t.Run("legend flag on emits LAYOUT_WITH_LEGEND", func(t *testing.T) {
		g := &graph.Graph{
			Nodes:  []*graph.Node{pumlNode("a", model.TypeSystem, "A")},
			Legend: &graph.Legend{Entries: []graph.LegendEntry{{Label: "system"}}},
		}

		out, err := render.RenderPlantUML(g)
		require.NoError(t, err)

		assert.Contains(t, string(out), "LAYOUT_WITH_LEGEND()")
	})

	t.Run("legend flag off omits it", func(t *testing.T) {
		g := &graph.Graph{Nodes: []*graph.Node{pumlNode("a", model.TypeSystem, "A")}}

		out, err := render.RenderPlantUML(g)
		require.NoError(t, err)

		assert.NotContains(t, string(out), "LAYOUT_WITH_LEGEND()")
	})
}

// TestRenderPlantUML_Escaping proves author-controlled text cannot break the
// generated PlantUML: quotes are escaped and newlines folded, and the
// breadcrumb navigation bar stays omitted (issue #25 defers it).
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestRenderPlantUML_Escaping(t *testing.T) {
	g := &graph.Graph{
		Title: `The "quoting" system`,
		Nodes: []*graph.Node{{
			ID:    "a",
			Label: &graph.Label{Name: `Weird "name"`, Description: "line1\nline2"},
			Shape: graph.ShapeHTML,
			Type:  model.TypeSystem,
		}},
	}

	out, err := render.RenderPlantUML(g)
	require.NoError(t, err)

	s := string(out)
	assert.Contains(t, s, `title The \"quoting\" system`, "title quotes must be escaped")
	assert.Contains(t, s, `System(a, "Weird \"name\"", "line1\nline2")`,
		"label quotes must be escaped and newlines folded")
	assert.NotContains(t, s, "breadcrumb", "breadcrumb nav bar is deliberately omitted")
}
