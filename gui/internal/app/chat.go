// chat.go is the P1 chat panel's App surface: provider configuration with
// local-only persistence, streaming chat requests (deltas pushed through the
// event sink), structured edit proposals parsed from the assistant reply,
// and confirmation-gated application.

package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/Djarvur/c4drill/gui/internal/ai"
)

// chatEventName is the event the streaming chat pushes under.
const chatEventName = "chat"

// ChatEvent is one chat progress frame (streamed deltas, completion, errors).
type ChatEvent struct {
	RequestID string        `json:"requestID"`
	Delta     string        `json:"delta,omitempty"`
	Done      bool          `json:"done,omitempty"`
	Error     string        `json:"error,omitempty"`
	Answer    string        `json:"answer,omitempty"`
	Proposals []ai.Proposal `json:"proposals,omitempty"`
}

// ApplyResult reports which proposals were written and which failed.
type ApplyResult struct {
	Applied []string `json:"applied"`
	Errors  []string `json:"errors"`
}

// chatState is the App's P1 state: provider config, persistence path, and
// the in-flight request cancels.
type chatState struct {
	mu      sync.Mutex
	config  ai.Config
	loaded  bool
	cancels map[string]context.CancelFunc
	seq     int

	// configPath overrides the default local config file (tests); nil means
	// the platform default. Atomic: read while holding mu (setChatConfigPath
	// is setup-time only, but stay race-clean regardless).
	configPath atomic.Pointer[string]
}

// SetChatConfigPath overrides where the chat settings persist. Empty
// restores the platform default (<UserConfigDir>/c4drill/gui.json). Tests
// use this to stay off the real user config; production code never calls it.
func (a *App) SetChatConfigPath(path string) {
	a.chat.configPath.Store(&path)

	a.chat.mu.Lock()
	defer a.chat.mu.Unlock()

	a.chat.loaded = false
	a.chat.config = ai.Config{}
}

// chatConfigPath is the local app-config file holding the chat settings
// (keys live here ONLY — never anywhere else, per issue #31).
func (a *App) chatConfigPath() (string, error) {
	if override := a.chat.configPath.Load(); override != nil && *override != "" {
		return *override, nil
	}

	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}

	return filepath.Join(base, "c4drill", "gui.json"), nil
}

// ChatConfig returns the stored provider configuration with the API key
// masked, plus the effective default prompt for display.
func (a *App) ChatConfig() (*ChatConfigResult, error) {
	cfg, err := a.loadChatConfig()
	if err != nil {
		return nil, err
	}

	return &ChatConfigResult{
		Config:      cfg.Mask(),
		DefaultFull: ai.SystemPrompt(""),
	}, nil
}

// ChatConfigResult is the chatConfig response.
type ChatConfigResult struct {
	Config ai.PublicConfig `json:"config"`
	// DefaultFull is the built-in system prompt (SKILL.md seed + protocol).
	DefaultFull string `json:"defaultFull"`
}

// SaveChatConfig stores the provider configuration locally (0600) and
// returns the masked view.
func (a *App) SaveChatConfig(cfg ai.Config) (*ai.PublicConfig, error) {
	if cfg.BaseURL != "" {
		if err := cfg.Validate(); err != nil {
			return nil, fmt.Errorf("chat provider: %w", err)
		}
	}

	path, err := a.chatConfigPath()
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}

	//nolint:gosec // G117: persisting the API key IS the contract — local-only secret store (issue #31)
	data, err := json.MarshalIndent(map[string]ai.Config{"chat": cfg}, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode chat config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return nil, fmt.Errorf("write chat config: %w", err)
	}

	a.chat.mu.Lock()
	a.chat.config = cfg
	a.chat.loaded = true
	a.chat.mu.Unlock()

	masked := cfg.Mask()

	return &masked, nil
}

// loadChatConfig reads the stored config (zero Config when none stored).
func (a *App) loadChatConfig() (ai.Config, error) {
	a.chat.mu.Lock()
	defer a.chat.mu.Unlock()

	if a.chat.loaded {
		return a.chat.config, nil
	}

	var cfg ai.Config

	path, err := a.chatConfigPath()
	if err == nil {
		data, readErr := os.ReadFile(path) //nolint:gosec // fixed user-config path
		if readErr == nil {
			var stored struct {
				Chat ai.Config `json:"chat"`
			}

			if json.Unmarshal(data, &stored) == nil {
				cfg = stored.Chat
			}
		}
	}

	a.chat.config = cfg
	a.chat.loaded = true

	return cfg, nil
}

// Chat starts one streaming chat request and returns its request id
// immediately. Deltas and the final frame (proposals included) arrive as
// "chat" events. history is the prior conversation (already roles-tagged).
func (a *App) Chat(history []ai.Message, userText string, ctx *ai.AuthoringContext) (string, error) {
	cfg, err := a.loadChatConfig()
	if err != nil {
		return "", err
	}

	if err := cfg.Validate(); err != nil {
		return "", fmt.Errorf("chat provider: %w", err)
	}

	a.chat.mu.Lock()
	a.chat.seq++

	id := "chat-" + strconv.Itoa(a.chat.seq)

	reqCtx, cancel := context.WithCancel(context.Background())
	a.chat.cancels[id] = cancel
	a.chat.mu.Unlock()

	go a.runChat(reqCtx, cancel, id, cfg, history, userText, ctx)

	return id, nil
}

// ChatAbort cancels an in-flight request (no-op when already finished).
func (a *App) ChatAbort(requestID string) {
	a.chat.mu.Lock()
	defer a.chat.mu.Unlock()

	if cancel, ok := a.chat.cancels[requestID]; ok {
		cancel()
	}
}

// runChat streams one request: deltas flow to the event sink as they arrive;
// the terminal frame carries the full answer and the parsed edit proposals.
func (a *App) runChat(
	reqCtx context.Context, cancel context.CancelFunc, id string,
	cfg ai.Config, history []ai.Message, userText string, ctx *ai.AuthoringContext,
) {
	defer func() {
		a.chat.mu.Lock()
		delete(a.chat.cancels, id)
		a.chat.mu.Unlock()

		cancel()
	}()

	client := ai.NewClient(cfg, nil)

	messages := ai.BuildMessages(ai.SystemPrompt(cfg.SystemPrompt), history, userText, ctx)

	deltas := make(chan ai.StreamDelta, 64)

	streamErr := make(chan error, 1)

	go func() {
		streamErr <- client.Stream(reqCtx, messages, deltas)

		close(deltas)
	}()

	var answer strings.Builder

	for delta := range deltas {
		if delta.Err != nil {
			a.emitChat(ChatEvent{RequestID: id, Error: delta.Err.Error(), Done: true})

			return
		}

		answer.WriteString(delta.Content)
		a.emitChat(ChatEvent{RequestID: id, Delta: delta.Content})
	}

	if err := <-streamErr; err != nil && !errors.Is(err, context.Canceled) {
		a.emitChat(ChatEvent{RequestID: id, Error: err.Error(), Done: true})

		return
	}

	full := answer.String()

	proposals := ai.ParseProposals(full, a.readProjectFile)

	a.emitChat(ChatEvent{
		RequestID: id,
		Done:      true,
		Answer:    full,
		Proposals: proposals,
	})
}

// emitChat pushes one chat frame through the event sink (lock-free read of
// the sink under a.mu).
func (a *App) emitChat(frame ChatEvent) {
	a.mu.Lock()
	sink := a.events
	a.mu.Unlock()

	sink(chatEventName, frame)
}

// readProjectFile is the proposal reader: current on-disk content of a
// project file, "" for files that do not exist yet (creations). Paths that
// escape the project are hard errors even though the ai package pre-checked
// them — the App's own rule wins.
func (a *App) readProjectFile(rel string) (string, error) {
	abs, err := a.absOf(rel)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(abs) //nolint:gosec // G304 is the product: the user picked this project file
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // new file creation
		}

		return "", fmt.Errorf("read %s: %w", rel, err)
	}

	return string(data), nil
}

// ApplyEdits writes the confirmed proposals. Every proposal is re-validated
// at this point (scope check against the CURRENT project) — confirmation UI
// state alone never authorizes a write.
func (a *App) ApplyEdits(proposals []ai.Proposal) (*ApplyResult, error) {
	res := &ApplyResult{Applied: []string{}, Errors: []string{}}

	for _, p := range proposals {
		if !p.Valid {
			res.Errors = append(res.Errors, p.Path+": proposal was not valid")

			continue
		}

		if _, err := a.readProjectFile(p.Path); err != nil {
			res.Errors = append(res.Errors, fmt.Sprintf("%s: %s", p.Path, err.Error()))

			continue
		}

		if err := a.WriteFile(p.Path, p.NewContent); err != nil {
			res.Errors = append(res.Errors, fmt.Sprintf("%s: %s", p.Path, err.Error()))

			continue
		}

		res.Applied = append(res.Applied, p.Path)
	}

	return res, nil
}
