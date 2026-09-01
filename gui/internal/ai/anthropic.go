// anthropic.go is the Anthropic-native chat adapter (issue #36): the
// Messages-API twin of the OpenAI-compatible client. Same Streamer contract,
// same StreamDelta pipeline, same edit-proposal handling downstream — only
// the wire shape differs:
//
//   - POST {baseURL}/v1/messages with x-api-key + anthropic-version headers
//     (no Authorization: Bearer);
//   - the system prompt rides the top-level `system` field, never a message;
//   - max_tokens is mandatory;
//   - the SSE stream carries typed events (content_block_delta carries the
//     text, message_stop is the sentinel, `error` events fail the stream).

package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// anthropicVersion is the Messages-API version header value.
const anthropicVersion = "2023-06-01"

// anthropicMaxTokens is the mandatory reply budget. c4drill replies are
// architecture explanations plus fenced edit blocks — 8k tokens is ample and
// accepted by every current Anthropic model.
const anthropicMaxTokens = 8192

// anthropicRequest is the Messages-API request body (the subset used).
type anthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system,omitempty"`
	Messages  []Message `json:"messages"`
	Stream    bool      `json:"stream"`
}

// anthropicEvent is one SSE data frame of the streamed Messages-API response.
type anthropicEvent struct {
	Type string `json:"type"`
	// Delta carries the text for content_block_delta frames.
	Delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta"`
	// Error appears on terminal error events (type "error").
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// AnthropicClient is a native Anthropic Messages-API streaming client
// (issue #36). baseURL is the API root without the version path (e.g.
// https://api.anthropic.com); the client posts to {baseURL}/v1/messages.
type AnthropicClient struct {
	baseURL string
	apiKey  string
	model   string

	httpClient *http.Client
}

// NewAnthropicClient builds an Anthropic Messages-API client; httpClient may
// be nil (sane defaults used).
func NewAnthropicClient(cfg Config, httpClient *http.Client) *AnthropicClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Minute}
	}

	return &AnthropicClient{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:  cfg.APIKey,
		model:   cfg.Model,

		httpClient: httpClient,
	}
}

// Stream posts the Messages-API request and feeds text deltas to ch until
// message_stop, an error event, or an error. It returns when the stream
// ends; the caller owns ch.
func (c *AnthropicClient) Stream(ctx context.Context, messages []Message, ch chan<- StreamDelta) error {
	// The transcript is assembled in OpenAI shape (system message first);
	// the Messages API wants the system prompt hoisted to a top-level field.
	var system string

	rest := make([]Message, 0, len(messages))

	for _, m := range messages {
		if m.Role == "system" {
			system = m.Content

			continue
		}

		rest = append(rest, m)
	}

	body, err := json.Marshal(anthropicRequest{
		Model:     c.model,
		MaxTokens: anthropicMaxTokens,
		System:    system,
		Messages:  rest,
		Stream:    true,
	})
	if err != nil {
		return fmt.Errorf("marshal chat request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/v1/messages", strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("build chat request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", c.apiKey)
	req.Header.Set("Anthropic-Version", anthropicVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("chat request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		limited, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))

		return fmt.Errorf("%w: %s: %s", errChatRequest, resp.Status, strings.TrimSpace(string(limited)))
	}

	return consumeSSE(ctx, resp.Body, func(payload string) (bool, error) {
		return handleAnthropicFrame(payload, ch)
	})
}

// handleAnthropicFrame decodes one Messages-API data frame and forwards its
// deltas. The first return reports message_stop / provider errors (both end
// the stream); unknown event types (ping, message_start, block boundaries)
// pass through silently.
func handleAnthropicFrame(payload string, ch chan<- StreamDelta) (bool, error) {
	var ev anthropicEvent
	if err := json.Unmarshal([]byte(payload), &ev); err != nil {
		return false, nil //nolint:nilerr // tolerate keep-alives and provider quirks: skip the frame
	}

	switch ev.Type {
	case "content_block_delta":
		if ev.Delta.Text != "" {
			ch <- StreamDelta{Content: ev.Delta.Text}
		}

		return false, nil
	case "error":
		msg := "unknown provider error"
		if ev.Error != nil {
			msg = ev.Error.Message
		}

		return true, fmt.Errorf("%w: %s", ErrStreamEnded, msg)
	case "message_stop":
		return true, nil
	default:
		return false, nil
	}
}
