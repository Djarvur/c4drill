// chat_test.go covers the App's chat orchestration end to end against a
// mocked OpenAI-compatible provider: config persistence/masking, streamed
// deltas through the event sink, proposal parsing from the reply, and
// confirmation-gated application with scope re-checks.

package app_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Djarvur/c4drill/gui/internal/ai"
	"github.com/Djarvur/c4drill/gui/internal/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// chatFixture builds an app with the demo project and a mocked provider
// answering with the given assistant text (as SSE deltas).
func chatFixture(t *testing.T, reply string) (*app.App, *[]app.ChatEvent, *sync.Mutex, string) {
	t.Helper()

	var frames strings.Builder

	for _, word := range strings.Split(reply, "\n") {
		delta, err := json.Marshal(map[string]any{
			"choices": []map[string]any{{"delta": map[string]any{"content": word + "\n"}}},
		})
		require.NoError(t, err)

		frames.WriteString("data: " + string(delta) + "\n\n")
	}

	frames.WriteString("data: [DONE]\n\n")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(frames.String()))
	}))
	t.Cleanup(server.Close)

	a, _ := newTestApp(t)
	a.SetChatConfigPath(filepath.Join(t.TempDir(), "gui.json"))

	_, err := a.SaveChatConfig(ai.Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Model:   "test-model",
	})
	require.NoError(t, err)

	var (
		mu     sync.Mutex
		events []app.ChatEvent
	)

	// The sink is called from the chat goroutine; guard the slice.
	a.SetEventSink(func(event string, payload any) {
		if event != "chat" {
			return
		}

		ce, ok := payload.(app.ChatEvent)
		if !ok {
			return
		}

		mu.Lock()

		events = append(events, ce)

		mu.Unlock()
	})

	return a, &events, &mu, server.URL
}

// waitChatDone returns the terminal chat frame, failing after a timeout.
func waitChatDone(t *testing.T, events *[]app.ChatEvent, mu *sync.Mutex) app.ChatEvent {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		mu.Lock()

		for _, ev := range *events {
			if ev.Done {
				mu.Unlock()

				return ev
			}
		}

		mu.Unlock()

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("chat never finished")

	return app.ChatEvent{}
}

func TestChatConfigMaskingAndPersistence(t *testing.T) {
	t.Parallel()

	a, _ := newTestApp(t)

	cfgFile := filepath.Join(t.TempDir(), "gui.json")
	a.SetChatConfigPath(cfgFile)

	// Nothing stored yet: empty, key-free view.
	res, err := a.ChatConfig()
	require.NoError(t, err)
	assert.False(t, res.Config.HasAPIKey)

	masked, err := a.SaveChatConfig(ai.Config{
		BaseURL:      "http://127.0.0.1:1/v1",
		APIKey:       "super-secret",
		Model:        "m",
		SystemPrompt: "custom",
	})
	require.NoError(t, err)
	assert.True(t, masked.HasAPIKey)
	assert.Equal(t, "m", masked.Model)

	// The public config struct must never carry the key…
	raw := mustJSON(t, masked)
	assert.NotContains(t, string(raw), "super-secret")

	// …but it must be persisted locally (local-only keys, per the issue).
	stored, err := os.ReadFile(cfgFile) //nolint:gosec // test-managed config path
	require.NoError(t, err)
	assert.Contains(t, string(stored), "super-secret")

	again, err := a.ChatConfig()
	require.NoError(t, err)
	assert.True(t, again.Config.HasAPIKey)
	assert.Equal(t, "http://127.0.0.1:1/v1", again.Config.BaseURL)
	assert.Contains(t, again.DefaultFull, "c4drill-edit path=")

	// A second App instance reading the same local store sees the key.
	fresh := app.New(nil)
	fresh.SetChatConfigPath(cfgFile)

	freshInfo, err := fresh.ChatConfig()
	require.NoError(t, err)
	assert.True(t, freshInfo.Config.HasAPIKey)

	// Invalid configs are refused.
	_, err = a.SaveChatConfig(ai.Config{BaseURL: "http://x"}) // no model, no key
	require.ErrorIs(t, err, ai.ErrEmptyModel)
}

func TestChatStreamsAndProposesEdits(t *testing.T) {
	t.Parallel()

	reply := "I will rename the model.\n" +
		"```c4drill-edit path=demo.toml\n" +
		"[properties]\nname = \"Renamed Demo\"\n" +
		"```\n" +
		"Done."

	a, events, mu, _ := chatFixture(t, reply)

	id, err := a.Chat(
		[]ai.Message{{Role: "user", Content: "hello"}},
		"rename the model",
		&ai.AuthoringContext{ActiveFile: "demo.toml", ActiveContent: "[properties]\n"},
	)
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	done := waitChatDone(t, events, mu)

	require.Empty(t, done.Error)
	assert.Contains(t, done.Answer, "Renamed Demo")
	require.Len(t, done.Proposals, 1)

	p := done.Proposals[0]
	assert.True(t, p.Valid)
	assert.Equal(t, "demo.toml", p.Path)
	assert.Contains(t, p.OldContent, "GUI Demo")

	// Deltas were streamed (more than one chat frame arrived).
	mu.Lock()
	defer mu.Unlock()

	deltas := 0

	for _, ev := range *events {
		if ev.Delta != "" {
			deltas++
		}
	}

	assert.Greater(t, deltas, 1)

	// Apply only after confirmation — the UI's Apply button drives this.
	res, err := a.ApplyEdits(done.Proposals)
	require.NoError(t, err)
	assert.Equal(t, []string{"demo.toml"}, res.Applied)
	assert.Empty(t, res.Errors)

	content, err := a.ReadFile("demo.toml")
	require.NoError(t, err)
	assert.Contains(t, content.Text, "Renamed Demo")
}

func TestApplyEditsRejectsInvalidProposals(t *testing.T) {
	t.Parallel()

	a, _ := newTestApp(t)

	res, err := a.ApplyEdits([]ai.Proposal{
		{Path: "../escape.toml", Valid: true, NewContent: "nope"},
		{Path: "demo.toml", Valid: false, NewContent: "nope"},
	})
	require.NoError(t, err)
	assert.Empty(t, res.Applied)
	require.Len(t, res.Errors, 2)
}

func TestChatWithoutProviderConfigFails(t *testing.T) {
	t.Parallel()

	// Fresh app, nothing persisted locally: Chat must fail with a clear
	// provider error (CI environments have no provider).
	a, _ := newTestApp(t)

	cfg, err := a.ChatConfig()
	require.NoError(t, err)

	if cfg.Config.HasAPIKey {
		t.Skip("a local chat config exists; cannot test the unconfigured path")
	}

	_, err = a.Chat(nil, "hi", nil)
	require.ErrorIs(t, err, ai.ErrEmptyBaseURL)
}
