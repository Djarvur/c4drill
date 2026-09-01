// chat_test.go covers the App's chat orchestration end to end against a
// mocked OpenAI-compatible provider: config persistence/masking, streamed
// deltas through the event sink, proposal parsing from the reply, and
// confirmation-gated application with scope re-checks.

package app_test

import (
	"encoding/json"
	"io"
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

// anthropicFixtureServer mocks the Anthropic Messages API: it asserts the
// wire shape (endpoint, headers, hoisted system prompt, stream flag) and
// answers with content_block_delta SSE events.
func anthropicFixtureServer(t *testing.T, reply string) *httptest.Server {
	t.Helper()

	var frames strings.Builder

	for _, word := range strings.Split(reply, "\n") {
		delta, err := json.Marshal(map[string]any{
			"type":  "content_block_delta",
			"index": 0,
			"delta": map[string]any{"type": "text_delta", "text": word + "\n"},
		})
		require.NoError(t, err)

		frames.WriteString("data: " + string(delta) + "\n\n")
	}

	frames.WriteString("data: {\"type\":\"message_stop\"}\n\n")

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/messages", r.URL.Path)
		assert.Equal(t, "test-key", r.Header.Get("X-Api-Key"))
		assert.NotEmpty(t, r.Header.Get("Anthropic-Version"))

		raw, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		var req struct {
			System   string `json:"system"`
			Messages []struct {
				Role string `json:"role"`
			} `json:"messages"`
			Stream bool `json:"stream"`
		}

		assert.NoError(t, json.Unmarshal(raw, &req))

		assert.True(t, req.Stream)
		assert.NotEmpty(t, req.System, "system prompt rides the top-level field")

		for _, m := range req.Messages {
			assert.NotEqual(t, "system", m.Role)
		}

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(frames.String()))
	}))
}

// TestChatConfigPerProviderSlots covers the #36 per-provider store: both
// providers keep independent settings, SaveChatConfig switches the active
// one, keys stay masked to the UI but persist locally.
func TestChatConfigPerProviderSlots(t *testing.T) {
	t.Parallel()

	cfgFile := filepath.Join(t.TempDir(), "gui.json")

	a, _ := newTestApp(t)
	a.SetChatConfigPath(cfgFile)

	// Save the OpenAI-compatible slot.
	_, err := a.SaveChatConfig(ai.Config{
		BaseURL: "http://127.0.0.1:1/v1",
		APIKey:  "test-key",
		Model:   "gpt-test",
	})
	require.NoError(t, err)

	// Save the Anthropic slot — the provider becomes active.
	masked, err := a.SaveChatConfig(ai.Config{
		Provider: ai.ProviderAnthropic,
		BaseURL:  "https://api.anthropic.com",
		APIKey:   "test-key-2",
		Model:    "claude-test",
	})
	require.NoError(t, err)
	require.NotNil(t, masked)
	assert.Equal(t, ai.ProviderAnthropic, masked.Provider)

	res, err := a.ChatConfig()
	require.NoError(t, err)
	assert.Equal(t, ai.ProviderAnthropic, res.Provider)
	assert.Equal(t, "claude-test", res.Config.Model)
	require.Len(t, res.Providers, 2)
	assert.Equal(t, "gpt-test", res.Providers[ai.ProviderOpenAI].Model)
	assert.True(t, res.Providers[ai.ProviderOpenAI].HasAPIKey)
	assert.True(t, res.Providers[ai.ProviderAnthropic].HasAPIKey)

	// The public view never carries either key…
	raw := mustJSON(t, res)
	assert.NotContains(t, string(raw), "test-key")
	assert.NotContains(t, string(raw), "test-key-2")

	// …but the local store does, with both slots and the active marker.
	stored, err := os.ReadFile(cfgFile) //nolint:gosec // test-managed config path
	require.NoError(t, err)
	assert.Contains(t, string(stored), "test-key")
	assert.Contains(t, string(stored), "test-key-2")
	assert.Contains(t, string(stored), string(ai.ProviderAnthropic))

	// A fresh App sees both slots with Anthropic still active.
	fresh := app.New(nil)
	fresh.SetChatConfigPath(cfgFile)

	freshRes, err := fresh.ChatConfig()
	require.NoError(t, err)
	assert.Equal(t, ai.ProviderAnthropic, freshRes.Provider)
	assert.Equal(t, "gpt-test", freshRes.Providers[ai.ProviderOpenAI].Model)

	// Re-saving the OpenAI slot switches the active provider back.
	_, err = a.SaveChatConfig(ai.Config{
		BaseURL: "http://127.0.0.1:1/v1",
		APIKey:  "openai-secret",
		Model:   "gpt-test-2",
	})
	require.NoError(t, err)

	res, err = a.ChatConfig()
	require.NoError(t, err)
	assert.Equal(t, ai.ProviderOpenAI, res.Provider)
	assert.Equal(t, "gpt-test-2", res.Config.Model)
	assert.Equal(t, "claude-test", res.Providers[ai.ProviderAnthropic].Model, "sibling slot untouched")

	// Unknown providers are refused.
	_, err = a.SaveChatConfig(ai.Config{
		Provider: "azure", BaseURL: "http://x", APIKey: "k", Model: "m",
	})
	require.ErrorIs(t, err, ai.ErrUnknownProvider)
}

// TestChatAnthropicProviderEndToEnd drives a full chat request through the
// Anthropic adapter selected in settings: the Messages-API wire shape is
// asserted by the mock, deltas stream through the event sink, and the
// reply's edit proposal parses as usual.
func TestChatAnthropicProviderEndToEnd(t *testing.T) {
	t.Parallel()

	reply := "Here is the change.\n" +
		"```c4drill-edit path=demo.toml\n" +
		"[properties]\nname = \"Anthropic Demo\"\n" +
		"```\n" +
		"Done."

	server := anthropicFixtureServer(t, reply)
	t.Cleanup(server.Close)

	a, _ := newTestApp(t)
	a.SetChatConfigPath(filepath.Join(t.TempDir(), "gui.json"))

	_, err := a.SaveChatConfig(ai.Config{
		Provider: ai.ProviderAnthropic,
		BaseURL:  server.URL,
		APIKey:   "test-key",
		Model:    "claude-test",
	})
	require.NoError(t, err)

	var (
		mu     sync.Mutex
		events []app.ChatEvent
	)

	a.SetEventSink(func(event string, payload any) {
		if event != "chat" {
			return
		}

		if ce, ok := payload.(app.ChatEvent); ok {
			mu.Lock()

			events = append(events, ce)

			mu.Unlock()
		}
	})

	id, err := a.Chat(
		[]ai.Message{{Role: "user", Content: "hello"}},
		"rename the model",
		&ai.AuthoringContext{ActiveFile: "demo.toml", ActiveContent: "[properties]\n"},
	)
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	done := waitChatDone(t, &events, &mu)

	require.Empty(t, done.Error)
	assert.Contains(t, done.Answer, "Anthropic Demo")
	require.Len(t, done.Proposals, 1)
	assert.True(t, done.Proposals[0].Valid)
}

// TestChatAbortCancelsStream is the #36 cancellation path: a provider that
// streams one delta and then hangs; ChatAbort cancels the request context
// (the mock observes its request context being torn down) and the terminal
// frame keeps the partial answer, marked aborted, without proposals.
func TestChatAbortCancelsStream(t *testing.T) {
	t.Parallel()

	var (
		mu         sync.Mutex
		clientGone bool
		release    = make(chan struct{}) // closed when the client disconnects
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		delta, err := json.Marshal(map[string]any{
			"choices": []map[string]any{{"delta": map[string]any{"content": "partial "}}},
		})
		assert.NoError(t, err)

		_, _ = w.Write([]byte("data: " + string(delta) + "\n\n"))

		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		// Hold the stream open until the abort tears the request down.
		<-r.Context().Done()

		mu.Lock()
		clientGone = true
		mu.Unlock()

		close(release)
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
		evMu   sync.Mutex
		events []app.ChatEvent

		sawDelta = make(chan struct{})
		once     sync.Once
	)

	a.SetEventSink(func(event string, payload any) {
		if event != "chat" {
			return
		}

		if ce, ok := payload.(app.ChatEvent); ok {
			if ce.Delta != "" {
				once.Do(func() { close(sawDelta) })
			}

			evMu.Lock()

			events = append(events, ce)

			evMu.Unlock()
		}
	})

	id, err := a.Chat(
		[]ai.Message{{Role: "user", Content: "hello"}},
		"write something long",
		nil,
	)
	require.NoError(t, err)

	// Abort only after the first delta was delivered to the sink — the
	// partial answer must surface in the terminal frame.
	select {
	case <-sawDelta:
	case <-time.After(5 * time.Second):
		t.Fatal("no delta ever streamed to the sink")
	}

	a.ChatAbort(id)

	done := waitChatDone(t, &events, &evMu)

	assert.True(t, done.Aborted, "terminal frame must be marked aborted")
	assert.Empty(t, done.Error)
	assert.Equal(t, "partial ", done.Answer, "the partial answer is kept")
	assert.Empty(t, done.Proposals, "aborted answers never propose edits")

	// The cancellation propagated to the provider HTTP request.
	select {
	case <-release:
	case <-time.After(5 * time.Second):
		t.Fatal("provider request context was never cancelled")
	}

	mu.Lock()
	defer mu.Unlock()

	assert.True(t, clientGone, "the provider must observe the client disconnect")

	// Aborting again is a no-op (the request is finished and unregistered).
	a.ChatAbort(id)
}
