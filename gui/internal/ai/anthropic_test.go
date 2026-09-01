// anthropic_test.go exercises the Anthropic-native Messages-API client
// against a mocked provider (httptest): request wire shape (system prompt,
// messages, mandatory fields, headers), SSE event parsing (text deltas,
// message_stop, error events, ping tolerance), HTTP failures, and the
// provider factory dispatch (issue #36).

package ai_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/gui/internal/ai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// anthropicFixture is a minimal well-formed Messages-API SSE stream.
const anthropicFixture = "event: message_start\n" +
	`data: {"type":"message_start","message":{"id":"msg_1","role":"assistant"}}` + "\n\n" +
	": keep-alive\n\n" +
	`data: {"type":"content_block_start","index":0,"content_block":{"type":"text","text":""}}` + "\n\n" +
	"event: ping\n" +
	`data: {"type":"ping"}` + "\n\n" +
	`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}` + "\n\n" +
	`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":", world"}}` + "\n\n" +
	`data: {"type":"content_block_stop","index":0}` + "\n\n" +
	"event: message_delta\n" +
	`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"}}` + "\n\n" +
	"event: message_stop\n" +
	`data: {"type":"message_stop"}` + "\n\n"

func TestAnthropicStreamWireShapeAndDeltas(t *testing.T) {
	t.Parallel()

	var (
		gotMethod string
		gotPath   string
		gotKey    string
		gotVer    string
		body      []byte
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotKey = r.Header.Get("X-Api-Key")
		gotVer = r.Header.Get("Anthropic-Version")

		raw, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		body = raw

		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte(anthropicFixture))
	}))
	defer server.Close()

	client := ai.NewProviderClient(ai.Config{
		Provider: ai.ProviderAnthropic,
		BaseURL:  server.URL,
		APIKey:   "test-key",
		Model:    "claude-sonnet",
	}, server.Client())

	ch := make(chan ai.StreamDelta, 16)

	err := client.Stream(context.Background(),
		[]ai.Message{
			{Role: "system", Content: "You are the c4drill assistant."},
			{Role: "user", Content: "hi"},
			{Role: "assistant", Content: "hello"},
			{Role: "user", Content: "explain the model"},
		}, ch)
	require.NoError(t, err)

	texts, cerr := collect(t, ch)
	require.NoError(t, cerr)
	assert.Equal(t, []string{"Hello", ", world"}, texts)

	// Wire shape: endpoint, auth headers.
	assert.Equal(t, http.MethodPost, gotMethod)
	assert.Equal(t, "/v1/messages", gotPath)
	assert.Equal(t, "test-key", gotKey)
	assert.NotEmpty(t, gotVer)

	// Body: model, mandatory max_tokens, stream flag, hoisted system prompt,
	// non-system messages only.
	var req struct {
		Model     string `json:"model"`
		MaxTokens int    `json:"max_tokens"`
		System    string `json:"system"`
		Messages  []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
		Stream bool `json:"stream"`
	}
	require.NoError(t, json.Unmarshal(body, &req))

	assert.Equal(t, "claude-sonnet", req.Model)
	assert.Positive(t, req.MaxTokens)
	assert.True(t, req.Stream)
	assert.Equal(t, "You are the c4drill assistant.", req.System)
	require.Len(t, req.Messages, 3)

	for _, m := range req.Messages {
		assert.NotEqual(t, "system", m.Role, "system prompt must ride the top-level field")
	}

	assert.Equal(t, "hi", req.Messages[0].Content)
	assert.Equal(t, "explain the model", req.Messages[2].Content)
}

func TestAnthropicStreamReportsErrorEvent(t *testing.T) {
	t.Parallel()

	body := `data: {"type":"error","error":{"type":"overloaded_error","message":"Overloaded"}}` + "\n\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	client := ai.NewProviderClient(ai.Config{
		Provider: ai.ProviderAnthropic, BaseURL: server.URL, APIKey: "k", Model: "m",
	}, server.Client())

	ch := make(chan ai.StreamDelta, 8)

	err := client.Stream(context.Background(), nil, ch)
	require.ErrorIs(t, err, ai.ErrStreamEnded)
	require.ErrorContains(t, err, "Overloaded")
}

func TestAnthropicStreamStopsAtMessageStop(t *testing.T) {
	t.Parallel()

	// A frame after message_stop must never surface.
	body := `data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"kept"}}` + "\n\n" +
		`data: {"type":"message_stop"}` + "\n\n" +
		`data: {"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"dropped"}}` + "\n\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer server.Close()

	client := ai.NewProviderClient(ai.Config{
		Provider: ai.ProviderAnthropic, BaseURL: server.URL, APIKey: "k", Model: "m",
	}, server.Client())

	ch := make(chan ai.StreamDelta, 8)

	err := client.Stream(context.Background(), nil, ch)
	require.NoError(t, err)

	texts, cerr := collect(t, ch)
	require.NoError(t, cerr)
	assert.Equal(t, []string{"kept"}, texts)
}

func TestAnthropicNon200CarriesStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"type":"error","error":{"message":"invalid x-api-key"}}`))
	}))
	defer server.Close()

	client := ai.NewProviderClient(ai.Config{
		Provider: ai.ProviderAnthropic, BaseURL: server.URL, APIKey: "wrong", Model: "m",
	}, server.Client())

	ch := make(chan ai.StreamDelta, 8)

	err := client.Stream(context.Background(), nil, ch)
	require.ErrorContains(t, err, "401")
}

func TestProviderFactoryDispatchesByConfig(t *testing.T) {
	t.Parallel()

	// One mock that records which endpoint shape was hit; the factory must
	// route by cfg.Provider.
	var paths []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)

		if strings.HasSuffix(r.URL.Path, "/v1/messages") {
			_, _ = w.Write([]byte(`data: {"type":"message_stop"}` + "\n\n"))

			return
		}

		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	for _, tc := range []struct {
		provider string
		wantPath string
	}{
		{provider: ai.ProviderOpenAI, wantPath: "/chat/completions"},
		{provider: "", wantPath: "/chat/completions"}, // legacy empty = openai-compatible
		{provider: ai.ProviderAnthropic, wantPath: "/v1/messages"},
	} {
		client := ai.NewProviderClient(ai.Config{
			Provider: tc.provider, BaseURL: server.URL, APIKey: "k", Model: "m",
		}, server.Client())

		ch := make(chan ai.StreamDelta, 4)

		require.NoError(t, client.Stream(context.Background(), []ai.Message{{Role: "user", Content: "hi"}}, ch))
		require.NotEmpty(t, paths)
		assert.Equal(t, tc.wantPath, paths[len(paths)-1], "provider %q", tc.provider)
	}
}

func TestConfigProviderValidation(t *testing.T) {
	t.Parallel()

	require.ErrorIs(t, ai.Config{Provider: "azure", BaseURL: "http://x", Model: "m", APIKey: "k"}.Validate(),
		ai.ErrUnknownProvider)
	require.NoError(t, ai.Config{Provider: ai.ProviderAnthropic, BaseURL: "http://x", Model: "m", APIKey: "k"}.Validate())

	// The provider rides the masked (public) view; the key never does.
	masked := ai.Config{Provider: ai.ProviderAnthropic, BaseURL: "http://x", Model: "m", APIKey: "secret"}.Mask()
	assert.Equal(t, ai.ProviderAnthropic, masked.Provider)
	assert.True(t, masked.HasAPIKey)

	raw, err := json.Marshal(masked)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "secret")
}
