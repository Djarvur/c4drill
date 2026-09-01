// Package ai is the P1 chat panel's engine (issue #31): a provider-agnostic
// OpenAI-compatible streaming client, the authoring context assembly, and
// the structured edit-proposal pipeline (parse → validate → diff → apply on
// explicit confirmation only).
package ai

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Exported client errors (providers/tests match with errors.Is).
var (
	ErrNoAPIKey     = errors.New("no API key configured")
	ErrEmptyBaseURL = errors.New("no base URL configured")
	ErrEmptyModel   = errors.New("no model configured")

	// ErrStreamEnded wraps provider-side stream terminations.
	ErrStreamEnded = errors.New("stream ended without a [DONE] sentinel")

	// errChatRequest wraps non-200 provider responses.
	errChatRequest = errors.New("chat request failed")
)

// Message is one chat-transcript message (OpenAI wire shape).
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Config is the provider configuration. The API key lives in local app
// config only (issue #31) — never in the model output back to the UI.
type Config struct {
	BaseURL      string `json:"baseURL"`
	APIKey       string `json:"apiKey"`
	Model        string `json:"model"`
	SystemPrompt string `json:"systemPrompt"`
}

// PublicConfig is what the UI may see: the key reduced to presence.
type PublicConfig struct {
	BaseURL      string `json:"baseURL"`
	Model        string `json:"model"`
	HasAPIKey    bool   `json:"hasAPIKey"`
	SystemPrompt string `json:"systemPrompt"`
}

// Mask reduces a Config to its public form.
func (c Config) Mask() PublicConfig {
	return PublicConfig{
		BaseURL:      c.BaseURL,
		Model:        c.Model,
		HasAPIKey:    c.APIKey != "",
		SystemPrompt: c.SystemPrompt,
	}
}

// Validate checks the config is usable for a chat request.
func (c Config) Validate() error {
	switch {
	case c.BaseURL == "":
		return ErrEmptyBaseURL
	case c.Model == "":
		return ErrEmptyModel
	case c.APIKey == "":
		return ErrNoAPIKey
	default:
		return nil
	}
}

// StreamDelta is one streamed chunk: an assistant text fragment or the
// terminal error.
type StreamDelta struct {
	Content string
	Err     error
}

// chatRequest is the OpenAI /chat/completions request body (the subset used).
type chatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// streamEvent is one SSE event body of the streamed response.
type streamEvent struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// Client is an OpenAI-compatible streaming chat client. baseURL includes the
// version prefix if the provider needs one (e.g.
// https://api.openai.com/v1, http://localhost:11434/v1 for Ollama).
type Client struct {
	baseURL string
	apiKey  string
	model   string

	httpClient *http.Client
}

// NewClient builds a client; httpClient may be nil (sane defaults used).
func NewClient(cfg Config, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Minute}
	}

	return &Client{
		baseURL: strings.TrimRight(cfg.BaseURL, "/"),
		apiKey:  cfg.APIKey,
		model:   cfg.Model,

		httpClient: httpClient,
	}
}

// Stream posts the chat completion and feeds text deltas to ch until the
// provider's [DONE] sentinel or an error. It returns when the stream ends;
// the caller owns ch.
func (c *Client) Stream(ctx context.Context, messages []Message, ch chan<- StreamDelta) error {
	body, err := json.Marshal(chatRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   true,
	})
	if err != nil {
		return fmt.Errorf("marshal chat request: %w", err)
	}

	url := c.baseURL + "/chat/completions"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("build chat request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("chat request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		limited, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))

		return fmt.Errorf("%w: %s: %s", errChatRequest, resp.Status, strings.TrimSpace(string(limited)))
	}

	return c.consume(ctx, resp, ch)
}

// consume scans the SSE stream line by line, dispatching data frames to
// handleFrame until [DONE] or EOF.
func (c *Client) consume(ctx context.Context, resp *http.Response, ch chan<- StreamDelta) error {
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64<<10), 1<<20)

	done := false

	for scanner.Scan() && !done {
		if ctx.Err() != nil {
			return ctx.Err() //nolint:wrapcheck // the caller is the canceller
		}

		payload, ok := dataFrame(scanner.Text())
		if !ok {
			continue // SSE comment, separator, or non-data line
		}

		end, err := c.handleFrame(payload, ch)
		if err != nil {
			return err
		}

		done = end
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read chat stream: %w", err)
	}

	return nil // EOF without [DONE]: providers do this; the stream is over
}

// dataFrame extracts the payload of a `data:` SSE line (ok=false for
// separators, comments, and other fields).
func dataFrame(line string) (string, bool) {
	if line == "" || !strings.HasPrefix(line, "data:") {
		return "", false
	}

	payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))

	return payload, payload != ""
}

// handleFrame decodes one data frame and forwards its deltas. The first
// return reports the [DONE] sentinel (or a provider error — both end the
// stream).
func (c *Client) handleFrame(payload string, ch chan<- StreamDelta) (bool, error) {
	if payload == "[DONE]" {
		return true, nil
	}

	var ev streamEvent
	if err := json.Unmarshal([]byte(payload), &ev); err != nil {
		return false, nil //nolint:nilerr // tolerate keep-alives and provider quirks: skip the frame
	}

	if ev.Error != nil {
		return true, fmt.Errorf("%w: %s", ErrStreamEnded, ev.Error.Message)
	}

	for _, choice := range ev.Choices {
		if choice.Delta.Content != "" {
			ch <- StreamDelta{Content: choice.Delta.Content}
		}
	}

	return false, nil
}
