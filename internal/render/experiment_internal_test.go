package render

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/render/icons"
	"github.com/goccy/go-graphviz"
)

// createTestBase64Icon creates a base64-encoded SVG icon for testing.
func createTestBase64Icon(t *testing.T) string {
	t.Helper()

	tmpl, err := icons.GetTemplate(icons.TypeSystem)
	if err != nil {
		t.Fatal(err)
	}

	colored := icons.Colorize(tmpl, "#3C7FC0")

	return "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(colored))
}

// createTestGraphWithIcon creates a test graph with an icon-capable node.
func createTestGraphWithIcon() *graph.Graph {
	return &graph.Graph{
		Direction: "TB",
		EdgeStyle: "spline",
		Title:     "Test Diagram",
		Nodes: []*graph.Node{
			{
				ID:   "webapp",
				Type: model.TypeContainer,
				Label: &graph.Label{
					Name:        "Web Application",
					Technology:  "Go",
					Description: "Serves web pages",
				},
				Style: &graph.NodeStyle{
					BorderColor: "#3C7FC0",
					FillColor:   "#438DD5",
					FontColor:   "#FFFFFF",
				},
			},
			{
				ID:   "user",
				Type: model.TypePerson,
				Label: &graph.Label{
					Name:        "User",
					Description: "End user",
				},
				Style: &graph.NodeStyle{
					BorderColor: "#8A8A8A",
					FontColor:   "#8A8A8A",
				},
			},
		},
		Edges: []*graph.Edge{
			{
				Source: "user",
				Target: "webapp",
				Label: &graph.EdgeLabel{
					Description: "Uses",
				},
			},
		},
	}
}

// renderTestNode renders a node with the given HTML label and returns SVG output.
func renderTestNode(t *testing.T, htmlLabel string) string {
	t.Helper()

	ctx := context.Background()

	gv, err := graphviz.New(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer gv.Close()

	cg, err := gv.Graph()
	if err != nil {
		t.Fatal(err)
	}
	defer cg.Close()

	cn, err := cg.CreateNodeByName("webapp")
	if err != nil {
		t.Fatal(err)
	}

	cn.SetShape("box")
	_ = cn.SafeSet("style", "rounded,filled", "")
	cn.SetFillColor("#438DD5")
	cn.SetColor("#3C7FC0")

	htmlStr, err := cg.StrdupHTML(htmlLabel)
	if err != nil {
		t.Fatal(err)
	}

	cn.SetLabel(htmlStr)

	var svgBuf bytes.Buffer
	if err := gv.Render(ctx, cg, graphviz.SVG, &svgBuf); err != nil {
		t.Fatal(err)
	}

	return svgBuf.String()
}

// TestExperimentBase64InSVG tests that base64 images are rendered in SVG.
//
//nolint:paralleltest
func TestExperimentBase64InSVG(t *testing.T) {
	b64 := createTestBase64Icon(t)

	htmlLabel := fmt.Sprintf(`<table border="0" cellpadding="0" cellspacing="0">
<tr align="center">
<td rowspan="3" valign="middle"><img src="%s" width="32" height="32"/></td>
<td valign="bottom"><b>Web App</b></td>
</tr>
<tr align="center"><td valign="middle"><i>[Go]</i></td></tr>
<tr align="center"><td valign="top">A web application</td></tr>
</table>`, b64)

	svgOutput := renderTestNode(t, htmlLabel)

	t.Log("Contains <image>:", strings.Contains(svgOutput, "<image"))
	t.Log("Contains data:image:", strings.Contains(svgOutput, "data:image"))
	t.Log("Contains base64:", strings.Contains(svgOutput, "base64"))
}

// TestExperimentNoIcon tests rendering without an icon for comparison.
//
//nolint:paralleltest
func TestExperimentNoIcon(t *testing.T) {
	htmlLabel := `<table border="0" cellpadding="0" cellspacing="0">
<tr align="center">
<td valign="bottom"><b>Web App</b></td>
</tr>
<tr align="center"><td valign="middle"><i>[Go]</i></td></tr>
<tr align="center"><td valign="top">A web application</td></tr>
</table>`

	svgOutput := renderTestNode(t, htmlLabel)

	t.Log("SVG output length:", len(svgOutput))
}

// TestExperimentSpacerTD tests using a spacer TD for icon layout.
//
//nolint:paralleltest
func TestExperimentSpacerTD(t *testing.T) {
	htmlLabel := `<table border="0" cellpadding="0" cellspacing="0">
<tr align="center">
<td rowspan="3" width="36" fixedsize="true" valign="middle"> </td>
<td valign="bottom"><b>Web App</b></td>
</tr>
<tr align="center"><td valign="middle"><i>[Go]</i></td></tr>
<tr align="center"><td valign="top">A web application</td></tr>
</table>`

	svgOutput := renderTestNode(t, htmlLabel)

	t.Log("SVG output length:", len(svgOutput))
}

// TestExperimentMarkerTD tests using a marker in TD for icon injection.
//
//nolint:paralleltest
func TestExperimentMarkerTD(t *testing.T) {
	htmlLabel := `<table border="0" cellpadding="0" cellspacing="0">
<tr align="center">
<td rowspan="3" width="36" height="32" fixedsize="true" valign="middle">&#x2063;ICON:webapp&#x2063;</td>
<td valign="bottom"><b>Web App</b></td>
</tr>
<tr align="center"><td valign="middle"><i>[Go]</i></td></tr>
<tr align="center"><td valign="top">A web application</td></tr>
</table>`

	svgOutput := renderTestNode(t, htmlLabel)

	t.Log("Contains ICON marker:", strings.Contains(svgOutput, "ICON:webapp"))
}

// TestExperimentIconWithWidth tests icon with explicit width.
//
//nolint:paralleltest
func TestExperimentIconWithWidth(t *testing.T) {
	b64 := createTestBase64Icon(t)

	htmlLabel := fmt.Sprintf(`<table border="0" cellpadding="0" cellspacing="0">
<tr align="center">
<td rowspan="3" width="36" valign="middle"><img src="%s" width="32" height="32"/></td>
<td valign="bottom"><b>Web App</b></td>
</tr>
<tr align="center"><td valign="middle"><i>[Go]</i></td></tr>
<tr align="center"><td valign="top">A web application</td></tr>
</table>`, b64)

	svgOutput := renderTestNode(t, htmlLabel)

	t.Log("Contains <image>:", strings.Contains(svgOutput, "<image"))
}

// TestExperimentCurrentApproach generates SVG with the current approach (inject post-render).
//
//nolint:paralleltest
func TestExperimentCurrentApproach(t *testing.T) {
	tmpDir := t.TempDir()

	g := createTestGraphWithIcon()

	svg, err := RenderSVGWithOutput(g, tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("=== CURRENT SVG WITH INJECTED ICONS ===")
	t.Log(string(svg))
	t.Log("=== END ===")
}
