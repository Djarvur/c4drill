// app_test.go covers the P0 binding logic end to end against the real
// in-process pipeline: workspace listing/reading/writing, the in-memory LSP
// round trip (diagnostics, completion, formatting), live rendering, drill
// navigation, and the export menu.

package app_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/gui/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestApp opens the demo fixture project and returns the app plus the
// collected diagnostic events.
func newTestApp(t *testing.T) (*app.App, *[]app.DiagnosticsEvent) {
	t.Helper()

	var events []app.DiagnosticsEvent

	a := app.New(nil)
	a.SetEventSink(func(event string, payload any) {
		if event == "diagnostics" {
			if de, ok := payload.(app.DiagnosticsEvent); ok {
				events = append(events, de)
			}
		}
	})

	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "nested"), 0o750))

	// Copy the demo fixture into the temp project (it gets edited).
	src, err := os.ReadFile(filepath.Join("testdata", "demo", "demo.toml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(root, "demo.toml"), src, 0o600)) //nolint:gosec // temp-dir fixture

	_, err = a.OpenProject(root)
	require.NoError(t, err)

	return a, &events
}

func TestOpenProjectListsModelFiles(t *testing.T) {
	t.Parallel()

	a, _ := newTestApp(t)

	info, err := a.ListFiles()
	require.NoError(t, err)
	require.NotEmpty(t, info.Files)

	paths := make([]string, 0, len(info.Files))
	for _, f := range info.Files {
		paths = append(paths, f.Path)
	}

	assert.Contains(t, paths, "demo.toml")
	assert.DirExists(t, info.Dir)
}

func TestReadAndWriteFile(t *testing.T) {
	t.Parallel()

	a, _ := newTestApp(t)

	content, err := a.ReadFile("demo.toml")
	require.NoError(t, err)
	assert.Contains(t, content.Text, "[properties]")

	_, err = a.ReadFile("../outside.toml")
	require.ErrorIs(t, err, app.ErrPathOutsideProject)

	require.NoError(t, a.WriteFile("demo.toml", content.Text+"\n# touched\n"))

	again, err := a.ReadFile("demo.toml")
	require.NoError(t, err)
	assert.Contains(t, again.Text, "# touched")

	_, err = a.ReadFile("missing.toml")
	require.Error(t, err)
}

// TestLSPRoundTripDiagnostics is the acceptance-critical in-memory LSP
// round trip: didOpen → clean diagnostics; a broken buffer → error
// diagnostics with the CLI's wording; didClose → cleared.
func TestLSPRoundTripDiagnostics(t *testing.T) {
	t.Parallel()

	a, events := newTestApp(t)

	require.NoError(t, a.DidOpen("demo.toml", readFixture(t, a, "demo.toml")))
	requireDiagEvent(t, events, "demo.toml", 0)

	// Break the file: an unknown unit type must surface a validation error.
	broken := strings.Replace(readFixture(t, a, "demo.toml"), `type = "person"`, `type = "personX"`, 1)
	require.NoError(t, a.DidChange("demo.toml", broken, 2))

	de := requireDiagEvent(t, events, "demo.toml", 1)
	require.NotEmpty(t, de.Diagnostics)
	assert.Equal(t, "c4drill", de.Diagnostics[0].Source)
	assert.Contains(t, de.Diagnostics[0].Message, "personX")

	// Fix it again: diagnostics clear.
	fixed := strings.Replace(broken, `type = "personX"`, `type = "person"`, 1)
	require.NoError(t, a.DidChange("demo.toml", fixed, 3))
	requireDiagEvent(t, events, "demo.toml", 0)

	require.NoError(t, a.DidClose("demo.toml"))
}

// TestCrossFileDiagnostics is acceptance criterion 4: opening an included
// file and breaking it republishes diagnostics for the including document.
func TestCrossFileDiagnostics(t *testing.T) {
	t.Parallel()

	var events []app.DiagnosticsEvent

	a := app.New(nil)
	a.SetEventSink(func(_ string, payload any) {
		if de, ok := payload.(app.DiagnosticsEvent); ok {
			events = append(events, de)
		}
	})

	root := t.TempDir()

	for _, name := range []string{"main.toml", "shared_half.toml"} {
		src, err := os.ReadFile(filepath.Join("testdata", "crossfile", name)) //nolint:gosec // fixture path constant
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(root, name), src, 0o600)) //nolint:gosec // temp-dir fixture
	}

	_, err := a.OpenProject(root)
	require.NoError(t, err)

	require.NoError(t, a.DidOpen("main.toml", readFixture(t, a, "main.toml")))
	requireDiagEvent(t, &events, "main.toml", 0)

	// Break the included file (unknown type); the includer must republish.
	shared, err := a.ReadFile("shared_half.toml")
	require.NoError(t, err)

	broken := strings.Replace(shared.Text, `type = "system"`, `type = "systemX"`, 1)
	require.NoError(t, a.DidOpen("shared_half.toml", shared.Text))
	require.NoError(t, a.DidChange("shared_half.toml", broken, 2))

	mainDiags := requireDiagEvent(t, &events, "main.toml", 1)
	require.NotEmpty(t, mainDiags.Diagnostics)
	assert.Contains(t, mainDiags.Diagnostics[0].Message, "systemX")
}

func TestCompletionAndFormat(t *testing.T) {
	t.Parallel()

	a, _ := newTestApp(t)

	text := readFixture(t, a, "demo.toml")
	require.NoError(t, a.DidOpen("demo.toml", text))

	// Completion at a fresh `type = "` value line: unit types come back.
	text += "\n[newUnit]\ntype = \""
	require.NoError(t, a.DidChange("demo.toml", text, 2))

	lines := strings.Split(text, "\n")
	last := uint32(len(lines) - 1) //nolint:gosec // fixture line count

	res, err := a.Completion("demo.toml", last, uint32(len(`type = "`)))
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotEmpty(t, res.Items, "unit-type completions expected at type = \"")

	labels := make([]string, 0, len(res.Items))
	for _, it := range res.Items {
		labels = append(labels, it.Label)
	}

	assert.Contains(t, labels, "system")
	assert.Contains(t, labels, "person")

	// Formatting: mangle whitespace, format, expect canonical output.
	mangled := strings.Replace(text+("db\"\n"), `peer = "mail"`, "peer   =   \"mail\"", 1)
	require.NoError(t, a.DidChange("demo.toml", mangled, 3))

	formatted, err := a.Format("demo.toml")
	require.NoError(t, err)
	require.NotNil(t, formatted)
	require.NotEmpty(t, formatted.Text, "formatter should produce a replacement")
	assert.Contains(t, formatted.Text, `peer = "mail"`)
	assert.NotContains(t, formatted.Text, "peer   =")
}

func TestHover(t *testing.T) {
	t.Parallel()

	a, _ := newTestApp(t)

	text := readFixture(t, a, "demo.toml")
	require.NoError(t, a.DidOpen("demo.toml", text))

	// Hover hovers peer reference values (the resolved path, level, type).
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, `peer = "cloud.ui.api"`) {
			h, err := a.Hover("demo.toml", uint32(i), 10)
			require.NoError(t, err)
			require.NotNil(t, h, "hover expected over a peer value")
			assert.Contains(t, h.Contents.Value, "cloud.ui.api")

			return
		}
	}

	t.Fatal("fixture changed: peer link not found")
}

// TestRenderLive covers the preview contract: valid model → SVG; broken
// model → empty SVG + diagnostics (no stale diagram); expanded override and
// legend controls flow into the renderer.
func TestRenderLive(t *testing.T) {
	t.Parallel()

	a, _ := newTestApp(t)

	res, err := a.Render("demo.toml", app.RenderOptions{})
	require.NoError(t, err)
	assert.Contains(t, res.SVG, "<svg")
	require.Empty(t, res.Diagnostics)
	require.Empty(t, res.Target)
	require.Len(t, res.Breadcrumbs, 1)

	// Drill target: C3 of cloud.ui.api.
	res, err = a.Render("demo.toml", app.RenderOptions{Target: "cloud.ui.api"})
	require.NoError(t, err)
	assert.Contains(t, res.SVG, "<svg")
	require.Len(t, res.Breadcrumbs, 4)

	// Expanded override ("collapse all" — distinct from model default).
	res, err = a.Render("demo.toml", app.RenderOptions{Expanded: []string{}})
	require.NoError(t, err)
	assert.Contains(t, res.SVG, "<svg")

	// Legend off.
	legendOff := false
	res, err = a.Render("demo.toml", app.RenderOptions{Legend: &legendOff})
	require.NoError(t, err)
	assert.NotContains(t, res.SVG, "LEGEND")

	// Broken model: empty SVG + pipeline diagnostics, CLI-parity message.
	text := readFixture(t, a, "demo.toml")
	require.NoError(t, a.DidOpen("demo.toml", text))

	orphan := text + "\n[orphan]\ntype = \"db\"\nname = \"Orphan\"\ndescription = \"dangling\"\n"
	require.NoError(t, a.DidChange("demo.toml", orphan, 2))

	res, err = a.Render("demo.toml", app.RenderOptions{})
	require.NoError(t, err, "renderDiagram reports failures as diagnostics, not transport errors")
	require.Empty(t, res.SVG, "no stale diagram on invalid model")
	require.NotEmpty(t, res.Diagnostics)
	assert.Equal(t, 1, res.Diagnostics[0].Severity) // LSP error severity
}

// TestRenderAutoOpensUnopenedFile: rendering a file the editor never opened
// must work (open-from-disk path).
func TestRenderAutoOpensUnopenedFile(t *testing.T) {
	t.Parallel()

	a, _ := newTestApp(t)

	res, err := a.Render("demo.toml", app.RenderOptions{})
	require.NoError(t, err)
	assert.Contains(t, res.SVG, "<svg")
}

// TestExportAllFormats runs the export menu through the CLI's five formats
// and checks the writer-convention layout.
func TestExportAllFormats(t *testing.T) {
	t.Parallel()

	a, _ := newTestApp(t)

	outDir := t.TempDir()

	for _, format := range []string{app.FormatSVG, app.FormatHTML, app.FormatDOT, app.FormatPNG, app.FormatPlantUML} {
		res, err := a.Export("demo.toml", app.RenderOptions{}, format, outDir)
		require.NoError(t, err, "format %s", format)
		require.NotEmpty(t, res.Files, "format %s", format)
		assert.Equal(t, "demo."+format, res.Files[0], "format %s", format)

		data, err := os.ReadFile(filepath.Join(outDir, "demo."+format)) //nolint:gosec // export-produced path
		require.NoError(t, err, "format %s", format)
		require.NotEmpty(t, data, "format %s", format)
	}

	// PNG also writes its HTML navigation sibling (CLI convention).
	_, err := os.Stat(filepath.Join(outDir, "demo.html"))
	require.NoError(t, err)

	// Exported SVG must equal the preview SVG for the same view (acceptance
	// criterion 6: byte-identical to the CLI/preview pipeline).
	preview, err := a.Render("demo.toml", app.RenderOptions{})
	require.NoError(t, err)

	exported, err := os.ReadFile(filepath.Join(outDir, "demo.svg")) //nolint:gosec // export-produced path
	require.NoError(t, err)
	assert.Equal(t, preview.SVG, string(exported))
}

func TestExportTargetLayoutAndExpanded(t *testing.T) {
	t.Parallel()

	a, _ := newTestApp(t)

	outDir := t.TempDir()

	res, err := a.Export("demo.toml", app.RenderOptions{Target: "cloud.ui.api"}, app.FormatSVG, outDir)
	require.NoError(t, err)
	assert.Equal(t, []string{"demo/cloud/ui/api.svg"}, res.Files)

	res, err = a.Export("demo.toml", app.RenderOptions{AllExpanded: true}, app.FormatSVG, outDir)
	require.NoError(t, err)
	assert.Equal(t, []string{"demo.expanded.svg"}, res.Files)
}

func TestExportInvalidModelFails(t *testing.T) {
	t.Parallel()

	a, _ := newTestApp(t)

	text := readFixture(t, a, "demo.toml")
	require.NoError(t, a.WriteFile("demo.toml", strings.Replace(text, `type = "person"`, `type = "personX"`, 1)))

	_, err := a.Export("demo.toml", app.RenderOptions{}, app.FormatSVG, t.TempDir())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "personX")

	_, err = a.Export("demo.toml", app.RenderOptions{}, "pdf", t.TempDir())
	require.ErrorIs(t, err, app.ErrInvalidExportFormat)
}

// TestDispatchRoundTrip drives the same JSON surface the transports use.
func TestDispatchRoundTrip(t *testing.T) {
	t.Parallel()

	a, _ := newTestApp(t)

	res, err := a.Dispatch("listFiles", nil)
	require.NoError(t, err)

	var info app.ProjectInfo
	require.NoError(t, json.Unmarshal(res, &info))
	require.NotEmpty(t, info.Files)

	res, err = a.Dispatch("appInfo", nil)
	require.NoError(t, err)

	var boot map[string]any
	require.NoError(t, json.Unmarshal(res, &boot))
	assert.Contains(t, boot, "initialDir")

	res, err = a.Dispatch("render", mustJSON(t, map[string]any{
		"path": "demo.toml",
		"opts": map[string]any{"target": "cloud", "allExpanded": false},
	}))
	require.NoError(t, err)

	var rr app.RenderResult
	require.NoError(t, json.Unmarshal(res, &rr))
	assert.Contains(t, rr.SVG, "<svg")

	_, err = a.Dispatch("nope", nil)
	require.ErrorIs(t, err, app.ErrUnknownMethod)

	_, err = a.Dispatch("readFile", mustJSON(t, map[string]string{"path": "../../etc/passwd"}))
	require.ErrorIs(t, err, app.ErrPathOutsideProject)
}

// --- helpers --------------------------------------------------------------

func readFixture(t *testing.T, a *app.App, rel string) string {
	t.Helper()

	content, err := a.ReadFile(rel)
	require.NoError(t, err)

	return content.Text
}

// requireDiagEvent waits for the latest diagnostics event for path and
// asserts its diagnostic count.
func requireDiagEvent(t *testing.T, events *[]app.DiagnosticsEvent, path string, wantCount int) *app.DiagnosticsEvent {
	t.Helper()

	require.NotEmpty(t, *events, "no diagnostics events for %s", path)

	for i := len(*events) - 1; i >= 0; i-- {
		de := (*events)[i]
		if de.Path == path {
			require.Len(t, de.Diagnostics, wantCount, "diagnostics for %s", path)

			return &de
		}
	}

	t.Fatalf("no diagnostics event for %s in %d events", path, len(*events))

	return nil
}

func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()

	raw, err := json.Marshal(v)
	require.NoError(t, err)

	return raw
}
