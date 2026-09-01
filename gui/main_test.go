// main_test.go is the P0 smoke e2e: the full HTTP transport (the webview-less
// mode real users can run with `go run ./gui --serve`) driven end to end —
// open project → didOpen → live render → drill navigation → export.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/gui/internal/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHTTPSmokeE2E drives the same endpoints a browser session would hit,
// against the demo fixture project.
func TestHTTPSmokeE2E(t *testing.T) {
	t.Parallel()

	// Assemble the project directory.
	root := t.TempDir()

	fixture := filepath.Join("internal", "app", "testdata", "demo", "demo.toml")
	src, err := os.ReadFile(fixture) //nolint:gosec // fixture path constant
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(root, "demo.toml"), src, 0o600)) //nolint:gosec // temp-dir fixture

	backend := app.New(nil)
	_, err = backend.OpenProject(root)
	require.NoError(t, err)

	server := httptest.NewServer(newHandler(backend))
	t.Cleanup(server.Close)

	post := func(t *testing.T, method string, params any, wantStatus int) map[string]any {
		t.Helper()

		body, err := json.Marshal(map[string]any{"method": method, "params": params})
		require.NoError(t, err)

		req, err := http.NewRequestWithContext(context.Background(), http.MethodPost,
			server.URL+"/api/dispatch", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		defer resp.Body.Close()

		require.Equal(t, wantStatus, resp.StatusCode, "POST %s", method)

		var out map[string]any
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&out))

		return out
	}

	// index.html serves (the embedded frontend build).
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL+"/", nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	// listFiles
	res := post(t, "listFiles", nil, http.StatusOK)
	files := resultSlice(t, res, "files")
	require.NotEmpty(t, files)

	demo, ok := files[0].(map[string]any)["path"].(string)
	require.True(t, ok, "file path must be a string")

	// didOpen + render (the live preview call)
	content, err := os.ReadFile(filepath.Join(root, demo)) //nolint:gosec // listed project file
	require.NoError(t, err)

	post(t, "didOpen", map[string]any{"path": demo, "text": string(content)}, http.StatusOK)

	res = post(t, "render", map[string]any{
		"path": demo,
		"opts": map[string]any{"target": "", "allExpanded": false},
	}, http.StatusOK)

	rendered, ok := res["result"].(map[string]any)
	require.True(t, ok, "render must return an object")
	assert.Contains(t, rendered["svg"], "<svg")

	// resolveDrill (a drill-down click on a C2 link)
	basename := strings.TrimSuffix(demo, ".toml")
	res = post(t, "resolveDrill", map[string]any{
		"path":   demo,
		"target": "",
		"href":   basename + "/cloud.svg",
	}, http.StatusOK)

	drilled, ok := res["result"].(map[string]any)["target"].(string)
	require.True(t, ok)
	assert.Equal(t, "cloud", drilled)

	// render the drilled target (C3)
	res = post(t, "render", map[string]any{
		"path": demo,
		"opts": map[string]any{"target": "cloud.ui.api", "allExpanded": false},
	}, http.StatusOK)
	drilledRender, ok := res["result"].(map[string]any)
	require.True(t, ok, "render must return an object")
	assert.Contains(t, drilledRender["svg"], "<svg")

	// export the current view
	res = post(t, "export", map[string]any{
		"path":   demo,
		"opts":   map[string]any{"target": "cloud.ui.api", "allExpanded": false},
		"format": "svg",
		"outDir": t.TempDir(),
	}, http.StatusOK)
	exported, ok := res["result"].(map[string]any)
	require.True(t, ok, "export must return an object")
	assert.NotEmpty(t, exported["files"])

	// unknown methods are rejected, not swallowed
	post(t, "bogus", nil, http.StatusBadRequest)

	// malformed JSON bodies are rejected
	req, err = http.NewRequestWithContext(context.Background(), http.MethodPost,
		server.URL+"/api/dispatch", strings.NewReader("{nope"))
	require.NoError(t, err)

	resp2, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer resp2.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp2.StatusCode)
}

// resultSlice fetches a []any field from the decoded result envelope.
func resultSlice(t *testing.T, envelope map[string]any, field string) []any {
	t.Helper()

	result, ok := envelope["result"].(map[string]any)
	require.True(t, ok, "result must be an object")

	list, ok := result[field].([]any)
	require.True(t, ok, "result.%s must be a list", field)

	return list
}

// TestEventHubBroadcast keeps the SSE fan-out honest: broadcast reaches a
// subscriber, slow clients do not block the hub.
func TestEventHubBroadcast(t *testing.T) {
	t.Parallel()

	hub := newEventHub()

	ch := hub.subscribe()
	hub.broadcast(eventFrame{Event: "diagnostics", Payload: "x"})

	frame := <-ch
	assert.Equal(t, "diagnostics", frame.Event)

	// A full channel must not wedge broadcast.
	slow := hub.subscribe()
	for range 128 {
		hub.broadcast(eventFrame{Event: "flood"})
	}

	hub.unsubscribe(slow)
	hub.unsubscribe(ch)
}
