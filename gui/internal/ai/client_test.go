// client_test.go exercises the OpenAI-compatible streaming client against a
// mocked provider (httptest): delta parsing, the [DONE] sentinel, provider
// error frames, HTTP failures, and keep-alive tolerance.

package ai_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Djarvur/c4drill/gui/internal/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sseServer spins up a provider mock answering with the given raw SSE body.
func sseServer(t *testing.T, status int, body string, seen *map[string]string) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if seen != nil {
			(*seen)["status"] = "hit"
		}

		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
}

// collect drains the delta channel for up to a second.
func collect(t *testing.T, ch <-chan ai.StreamDelta) ([]string, error) {
	t.Helper()

	var (
		texts []string
		err   error
	)

	for {
		select {
		case d, ok := <-ch:
			if !ok {
				return texts, err
			}

			if d.Err != nil {
				err = d.Err

				continue
			}

			texts = append(texts, d.Content)
		case <-time.After(time.Second):
			return texts, err
		}
	}
}

func TestStreamParsesDeltasUntilDone(t *testing.T) {
	t.Parallel()

	frames := []map[string]any{
		{"choices": []map[string]any{{"delta": map[string]any{"role": "assistant"}}}},
		{"choices": []map[string]any{{"delta": map[string]any{"content": "Hello"}}}},
		{"choices": []map[string]any{{"delta": map[string]any{"content": ", world"}}}},
		{"choices": []map[string]any{{"delta": map[string]any{}, "finish_reason": "stop"}}},
	}

	var sb strings.Builder

	sb.WriteString(": keep-alive comment\n\n")

	for _, f := range frames {
		raw, err := json.Marshal(f)
		require.NoError(t, err)
		sb.WriteString("data: " + string(raw) + "\n\n")
	}

	sb.WriteString("data: [DONE]\n\n")

	server := sseServer(t, http.StatusOK, sb.String(), nil)
	defer server.Close()

	client := ai.NewClient(ai.Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Model:   "test-model",
	}, server.Client())

	ch := make(chan ai.StreamDelta, 16)

	err := client.Stream(context.Background(), []ai.Message{{Role: "user", Content: "hi"}}, ch)
	require.NoError(t, err)

	texts, cerr := collect(t, ch)
	require.NoError(t, cerr)
	assert.Equal(t, []string{"Hello", ", world"}, texts)
}

func TestStreamReportsProviderErrorFrame(t *testing.T) {
	t.Parallel()

	body := "data: {\"error\": {\"message\": \"quota exceeded\"}}\n\n"

	server := sseServer(t, http.StatusOK, body, nil)

	defer server.Close()

	client := ai.NewClient(ai.Config{BaseURL: server.URL, APIKey: "k", Model: "m"}, server.Client())

	ch := make(chan ai.StreamDelta, 8)

	err := client.Stream(context.Background(), nil, ch)
	require.ErrorContains(t, err, "quota exceeded")
}

func TestStreamNon200CarriesStatus(t *testing.T) {
	t.Parallel()

	server := sseServer(t, http.StatusUnauthorized, `{"error":"bad key"}`, nil)
	defer server.Close()

	client := ai.NewClient(ai.Config{BaseURL: server.URL, APIKey: "wrong", Model: "m"}, server.Client())

	ch := make(chan ai.StreamDelta, 8)

	err := client.Stream(context.Background(), nil, ch)
	require.ErrorContains(t, err, "401")
}

func TestConfigValidationAndMask(t *testing.T) {
	t.Parallel()

	empty := ai.Config{}
	require.ErrorIs(t, empty.Validate(), ai.ErrEmptyBaseURL)
	require.ErrorIs(t, ai.Config{BaseURL: "http://x"}.Validate(), ai.ErrEmptyModel)
	require.ErrorIs(t, ai.Config{BaseURL: "http://x", Model: "m"}.Validate(), ai.ErrNoAPIKey)
	require.NoError(t, ai.Config{BaseURL: "http://x", Model: "m", APIKey: "k"}.Validate())

	masked := ai.Config{BaseURL: "http://x", Model: "m", APIKey: "secret"}.Mask()
	assert.True(t, masked.HasAPIKey)

	raw, err := json.Marshal(masked)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "secret")
}
