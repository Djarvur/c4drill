package output_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/output"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/render"
	"github.com/Djarvur/c4drill/internal/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: Tests in this file do NOT use t.Parallel() because the go-graphviz
// library uses a WASM-based rendering engine that has concurrency issues.

// buildTestModelForOutput creates a simple test model for output tests.
func buildTestModelForOutput() *parser.Model {
	return &parser.Model{
		Properties: model.Properties{
			Name: "Test System",
		},
		Units: map[string]*model.Unit{
			"user": {
				Type:        model.TypePerson,
				Name:        "User",
				Description: "A system user",
			},
			"system": {
				Type:        model.TypeSystem,
				Name:        "Main System",
				Description: "The main software system",
				Technology:  "Go",
			},
		},
	}
}

// buildExpandedTestModel creates a test model with expanded subunits.
func buildExpandedTestModel() *parser.Model {
	return &parser.Model{
		Properties: model.Properties{
			Name: "Test System with Subunits",
		},
		Units: map[string]*model.Unit{
			"mainapp": {
				Type:        model.TypeSystem,
				Name:        "Main App",
				Description: "The main application",
				Technology:  "Go",
				Expanded:    []string{"mainapp"},
				Subunits: map[string]*model.Unit{
					"api": {
						Type:        model.TypeContainer,
						Name:        "API",
						Description: "REST API",
						Technology:  "Go",
					},
					"db": {
						Type:        model.TypeContainerDb,
						Name:        "Database",
						Description: "PostgreSQL",
						Technology:  "PostgreSQL",
					},
				},
			},
		},
	}
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegrationWriteRenderedDOTToC1Path(t *testing.T) {
	// Test 1: Write rendered DOT to C1 path and read back matches
	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)

	// Build full pipeline
	m := buildTestModelForOutput()
	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)

	// Render to DOT
	dotData, err := render.RenderDOT(g)
	require.NoError(t, err, "RenderDOT should not return error")

	// Write to C1 path (empty unitPath for context level)
	err = w.Write("diagram", "", "dot", dotData)
	require.NoError(t, err, "Write should not return error")

	// Verify file exists at expected path
	expectedPath := filepath.Join(tmpDir, "diagram.dot")
	assert.FileExists(t, expectedPath, "DOT file should exist at C1 path")

	// Read file and verify content matches rendered output
	content, err := os.ReadFile(expectedPath)
	require.NoError(t, err, "Should be able to read written file")
	assert.Equal(t, dotData, content, "Written content should match rendered output")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegrationWriteRenderedSVGToC2Path(t *testing.T) {
	// Test 2: Write rendered SVG to C2 path and read back matches
	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)

	// Build full pipeline
	m := buildTestModelForOutput()
	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)

	// Render to SVG
	svgData, err := render.RenderSVG(g)
	require.NoError(t, err, "RenderSVG should not return error")

	// Write to C2 path (dotted unitPath for expanded unit)
	err = w.Write("diagram", "mainapp.api", "svg", svgData)
	require.NoError(t, err, "Write should not return error")

	// Verify file exists at expected nested path
	expectedPath := filepath.Join(tmpDir, "diagram", "mainapp", "api.svg")
	assert.FileExists(t, expectedPath, "SVG file should exist at C2/C3 nested path")

	// Read file and verify content matches rendered output
	content, err := os.ReadFile(expectedPath)
	require.NoError(t, err, "Should be able to read written file")
	assert.Equal(t, svgData, content, "Written content should match rendered output")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegrationWriteMultipleFilesDifferentPaths(t *testing.T) {
	// Test 3: Write multiple files with different unit paths
	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)

	// Build full pipeline
	m := buildTestModelForOutput()
	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)

	// Render once
	dotData, err := render.RenderDOT(g)
	require.NoError(t, err, "RenderDOT should not return error")

	// Write to multiple paths
	err = w.Write("diagram", "", "dot", dotData)
	require.NoError(t, err, "Write to root should not error")

	err = w.Write("diagram", "api", "dot", dotData)
	require.NoError(t, err, "Write to api path should not error")

	err = w.Write("diagram", "db", "dot", dotData)
	require.NoError(t, err, "Write to db path should not error")

	// Verify all files exist
	assert.FileExists(t, filepath.Join(tmpDir, "diagram.dot"))
	assert.FileExists(t, filepath.Join(tmpDir, "diagram", "api.dot"))
	assert.FileExists(t, filepath.Join(tmpDir, "diagram", "db.dot"))
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegrationNestedDirectoryStructure(t *testing.T) {
	// Test 4: Nested directory structure is created correctly
	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)

	// Build full pipeline
	m := buildTestModelForOutput()
	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)

	svgData, err := render.RenderSVG(g)
	require.NoError(t, err, "RenderSVG should not return error")

	// Write to deeply nested path
	err = w.Write("diagram", "mainapp.api.handlers.auth", "svg", svgData)
	require.NoError(t, err, "Write to deeply nested path should not error")

	// Verify directory hierarchy created
	expectedPath := filepath.Join(tmpDir, "diagram", "mainapp", "api", "handlers", "auth.svg")
	assert.FileExists(t, expectedPath, "File should exist at deeply nested path")

	// Verify all intermediate directories exist
	assert.DirExists(t, filepath.Join(tmpDir, "diagram"))
	assert.DirExists(t, filepath.Join(tmpDir, "diagram", "mainapp"))
	assert.DirExists(t, filepath.Join(tmpDir, "diagram", "mainapp", "api"))
	assert.DirExists(t, filepath.Join(tmpDir, "diagram", "mainapp", "api", "handlers"))
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegrationWrittenFilesContainValidDOT(t *testing.T) {
	// Test 5: Written files contain valid DOT content
	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)

	// Build full pipeline
	m := buildTestModelForOutput()
	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)

	dotData, err := render.RenderDOT(g)
	require.NoError(t, err, "RenderDOT should not return error")

	err = w.Write("diagram", "", "dot", dotData)
	require.NoError(t, err, "Write should not error")

	// Read back and verify DOT structure
	content, err := os.ReadFile(filepath.Join(tmpDir, "diagram.dot"))
	require.NoError(t, err, "Should be able to read written file")

	contentStr := string(content)
	assert.Contains(t, contentStr, "digraph", "Written file should contain valid DOT structure")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegrationWrittenFilesContainValidSVG(t *testing.T) {
	// Test 5b: Written files contain valid SVG content
	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)

	// Build full pipeline
	m := buildTestModelForOutput()
	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)

	svgData, err := render.RenderSVG(g)
	require.NoError(t, err, "RenderSVG should not return error")

	err = w.Write("diagram", "", "svg", svgData)
	require.NoError(t, err, "Write should not error")

	// Read back and verify SVG structure
	content, err := os.ReadFile(filepath.Join(tmpDir, "diagram.svg"))
	require.NoError(t, err, "Should be able to read written file")

	contentStr := string(content)
	assert.Contains(t, contentStr, "<?xml", "Written file should contain XML declaration")
	assert.Contains(t, contentStr, "<svg", "Written file should contain SVG element")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegrationFullEndToEndPipeline(t *testing.T) {
	// Full end-to-end test: model -> view -> graph -> render -> write -> read
	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)

	// Create model
	m := buildExpandedTestModel()

	// Generate C1 view and write
	c1View := view.GenerateC1View(m)
	c1Graph := graph.BuildGraph(c1View)
	c1SVG, err := render.RenderSVG(c1Graph)
	require.NoError(t, err, "C1 render should not error")

	err = w.Write("architecture", "", "svg", c1SVG)
	require.NoError(t, err, "C1 write should not error")

	// Generate C2 view and write
	c2View := view.GenerateC2View(m, "mainapp")
	c2Graph := graph.BuildGraph(c2View)
	c2SVG, err := render.RenderSVG(c2Graph)
	require.NoError(t, err, "C2 render should not error")

	err = w.Write("architecture", "mainapp", "svg", c2SVG)
	require.NoError(t, err, "C2 write should not error")

	// Verify both files exist
	assert.FileExists(t, filepath.Join(tmpDir, "architecture.svg"), "C1 file should exist")
	assert.FileExists(t, filepath.Join(tmpDir, "architecture", "mainapp.svg"), "C2 file should exist")

	// Verify content is valid SVG
	c1Content, err := os.ReadFile(filepath.Join(tmpDir, "architecture.svg"))
	require.NoError(t, err)
	assert.Contains(t, string(c1Content), "<svg", "C1 should be valid SVG")

	c2Content, err := os.ReadFile(filepath.Join(tmpDir, "architecture", "mainapp.svg"))
	require.NoError(t, err)
	assert.Contains(t, string(c2Content), "<svg", "C2 should be valid SVG")
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegrationBothFormatsSameDiagram(t *testing.T) {
	// Write same diagram in both DOT and SVG formats
	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)

	m := buildTestModelForOutput()
	v := view.GenerateC1View(m)
	g := graph.BuildGraph(v)

	dotData, err := render.RenderDOT(g)
	require.NoError(t, err)

	svgData, err := render.RenderSVG(g)
	require.NoError(t, err)

	// Write both formats
	err = w.Write("diagram", "", "dot", dotData)
	require.NoError(t, err)

	err = w.Write("diagram", "", "svg", svgData)
	require.NoError(t, err)

	// Verify both files exist with correct content
	dotContent, err := os.ReadFile(filepath.Join(tmpDir, "diagram.dot"))
	require.NoError(t, err)
	assert.Equal(t, dotData, dotContent)

	svgContent, err := os.ReadFile(filepath.Join(tmpDir, "diagram.svg"))
	require.NoError(t, err)
	assert.Equal(t, svgData, svgContent)
}

//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestIntegrationC3NestedPath(t *testing.T) {
	// Test C3 level deeply nested path
	tmpDir := t.TempDir()
	w := output.NewWriter(tmpDir)

	// Create model with nested containers
	m := &parser.Model{
		Properties: model.Properties{Name: "C3 Test"},
		Units: map[string]*model.Unit{
			"system": {
				Type:     model.TypeSystem,
				Name:     "System",
				Expanded: []string{"api"},
				Subunits: map[string]*model.Unit{
					"api": {
						Type:     model.TypeContainer,
						Name:     "API",
						Expanded: []string{"handlers"},
						Subunits: map[string]*model.Unit{
							"handlers": {
								Type:        model.TypeComponent,
								Name:        "Handlers",
								Description: "Request handlers",
							},
						},
					},
				},
			},
		},
	}

	// Generate C3 view
	v := view.GenerateC3View(m, "system.api")
	require.NotNil(t, v)

	g := graph.BuildGraph(v)
	require.NotNil(t, g)

	svgData, err := render.RenderSVG(g)
	require.NoError(t, err)

	// Write to C3 path
	err = w.Write("diagram", "system.api.handlers", "svg", svgData)
	require.NoError(t, err)

	// Verify deeply nested path
	expectedPath := filepath.Join(tmpDir, "diagram", "system", "api", "handlers.svg")
	assert.FileExists(t, expectedPath)
}
