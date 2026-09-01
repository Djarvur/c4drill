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

	"github.com/Djarvur/c4drill/internal/gui/ai"
)

// chatEventName is the event the streaming chat pushes under.
const chatEventName = "chat"

// ChatEvent is one chat progress frame (streamed deltas, completion, errors).
// Aborted marks a request the user cancelled (issue #36): the partial answer
// is kept, no proposals are parsed from it.
type ChatEvent struct {
	RequestID string        `json:"requestID"`
	Delta     string        `json:"delta,omitempty"`
	Done      bool          `json:"done,omitempty"`
	Error     string        `json:"error,omitempty"`
	Aborted   bool          `json:"aborted,omitempty"`
	Answer    string        `json:"answer,omitempty"`
	Proposals []ai.Proposal `json:"proposals,omitempty"`
}

// ApplyResult reports which proposals were written and which failed.
type ApplyResult struct {
	Applied []string `json:"applied"`
	Errors  []string `json:"errors"`
}

// chatConfigStore is the local persistence shape (0600 file, issue #31):
// one config slot per provider (issue #36) plus the active one. Chat is the
// pre-#36 single-provider payload, kept for one-read migration only.
type chatConfigStore struct {
	Active    string               `json:"active,omitempty"`
	Providers map[string]ai.Config `json:"providers,omitempty"`
	Chat      *ai.Config           `json:"chat,omitempty"`
}

// migrate folds the legacy single-provider payload into the per-provider
// store (pre-#36 files carry only `chat`).
func (s *chatConfigStore) migrate() {
	if len(s.Providers) == 0 && s.Chat != nil {
		provider := s.Chat.NormalizedProvider()
		s.Providers = map[string]ai.Config{provider: *s.Chat}
		s.Active = provider
	}

	if s.Providers == nil {
		s.Providers = map[string]ai.Config{}
	}
}

// chatState is the App's P1 state: provider config, persistence path, and
// the in-flight request cancels.
type chatState struct {
	mu      sync.Mutex
	store   chatConfigStore
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
	a.chat.store = chatConfigStore{}
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

// ChatConfig returns the stored provider configuration (every provider slot
// with the API keys masked, plus which one is active) and the effective
// default prompt for display.
func (a *App) ChatConfig() (*ChatConfigResult, error) {
	store, err := a.loadChatConfig()
	if err != nil {
		return nil, err
	}

	providers := make(map[string]ai.PublicConfig, len(store.Providers))

	for name, cfg := range store.Providers {
		providers[name] = cfg.Mask()
	}

	active := store.activeConfig()

	return &ChatConfigResult{
		Provider:    active.NormalizedProvider(),
		Config:      active.Mask(),
		Providers:   providers,
		DefaultFull: ai.SystemPrompt(""),
	}, nil
}

// ChatConfigResult is the chatConfig response.
type ChatConfigResult struct {
	// Provider is the active provider (used by the next Chat call).
	Provider string `json:"provider"`
	// Config is the active provider's masked config (pre-#36 UI shape).
	Config ai.PublicConfig `json:"config"`
	// Providers carries every stored provider slot, keys masked.
	Providers map[string]ai.PublicConfig `json:"providers"`
	// DefaultFull is the built-in system prompt (SKILL.md seed + protocol).
	DefaultFull string `json:"defaultFull"`
}

// SaveChatConfig stores cfg in its provider's slot (0600 local file) and
// makes that provider active (issue #36). An empty BaseURL clears the slot.
// Returns the masked view of the saved slot.
func (a *App) SaveChatConfig(cfg ai.Config) (*ai.PublicConfig, error) {
	if !ai.ValidProvider(cfg.Provider) {
		return nil, ai.ErrUnknownProvider
	}

	provider := cfg.NormalizedProvider()
	cfg.Provider = provider

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

	a.chat.mu.Lock()

	store := a.chat.store
	store.Active = provider

	if store.Providers == nil {
		store.Providers = map[string]ai.Config{}
	}

	store.Providers[provider] = cfg

	a.chat.mu.Unlock()

	if err := a.writeChatConfig(path, store); err != nil {
		return nil, err
	}

	a.chat.mu.Lock()
	a.chat.store = store
	a.chat.loaded = true
	a.chat.mu.Unlock()

	masked := cfg.Mask()

	return &masked, nil
}

// writeChatConfig persists the store (the caller passes a snapshot whose map
// is no longer mutated under the lock).
func (a *App) writeChatConfig(path string, store chatConfigStore) error {
	data, err := json.MarshalIndent(map[string]chatConfigStore{"chat-store": store}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode chat config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write chat config: %w", err)
	}

	return nil
}

// loadChatConfig reads the stored config (empty store when none stored).
// Caller-free locking: returns a snapshot copy safe to read without the lock.
func (a *App) loadChatConfig() (chatConfigStore, error) {
	a.chat.mu.Lock()
	defer a.chat.mu.Unlock()

	if a.chat.loaded {
		return a.chat.store, nil
	}

	var store chatConfigStore

	path, err := a.chatConfigPath()
	if err == nil {
		data, readErr := os.ReadFile(path) //nolint:gosec // fixed user-config path
		if readErr == nil {
			var stored struct {
				// Legacy pre-#36 shape first (single config under `chat`),
				// then the per-provider store (issue #36).
				Legacy *ai.Config      `json:"chat"`
				Store  chatConfigStore `json:"chat-store"`
			}

			if json.Unmarshal(data, &stored) == nil {
				if stored.Store.Providers != nil || stored.Store.Active != "" {
					store = stored.Store
				} else if stored.Legacy != nil {
					store = chatConfigStore{Chat: stored.Legacy}
				}
			}
		}
	}

	store.migrate()
	a.chat.store = store
	a.chat.loaded = true

	return store, nil
}

// activeConfig returns the active provider's stored config (zero Config
// when nothing is stored). Caller holds chat.mu or works on a snapshot.
func (s *chatConfigStore) activeConfig() ai.Config {
	if s.Active == "" {
		return ai.Config{}
	}

	return s.Providers[s.Active]
}

// Chat starts one streaming chat request and returns its request id
// immediately. Deltas and the final frame (proposals included) arrive as
// "chat" events. history is the prior conversation (already roles-tagged).
// The active provider's adapter (issue #36) serves the request.
func (a *App) Chat(history []ai.Message, userText string, ctx *ai.AuthoringContext) (string, error) {
	store, err := a.loadChatConfig()
	if err != nil {
		return "", err
	}

	cfg := store.activeConfig()

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

// ChatAbort cancels an in-flight request (no-op when already finished). The
// stream's context is cancelled end-to-end (the provider HTTP request is
// torn down) and the request finishes with an aborted terminal frame that
// keeps the partial answer (issue #36).
func (a *App) ChatAbort(requestID string) {
	a.chat.mu.Lock()
	defer a.chat.mu.Unlock()

	if cancel, ok := a.chat.cancels[requestID]; ok {
		cancel()
	}
}

// runChat streams one request: deltas flow to the event sink as they arrive;
// the terminal frame carries the full answer and the parsed edit proposals.
// A cancelled context ends the stream with an aborted terminal frame — the
// partial answer is kept, no proposals are parsed from it.
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

	client := ai.NewProviderClient(cfg, nil)

	messages := ai.BuildMessages(ai.SystemPrompt(cfg.SystemPrompt), history, userText, ctx)

	deltas := make(chan ai.StreamDelta, 64)

	streamErr := make(chan error, 1)

	go func() {
		streamErr <- client.Stream(reqCtx, messages, deltas)

		close(deltas)
	}()

	answer, ok := a.drainChatDeltas(id, deltas)
	if !ok {
		return
	}

	if err := <-streamErr; err != nil && !errors.Is(err, context.Canceled) {
		a.emitChat(ChatEvent{RequestID: id, Error: err.Error(), Done: true})

		return
	}

	// Aborted replies keep their partial text but never propose edits — a
	// half-streamed edit block is dangerous even when it happens to close.
	aborted := reqCtx.Err() != nil

	var proposals []ai.Proposal

	if !aborted {
		proposals = ai.ParseProposals(answer, a.readProjectFile)
	}

	a.emitChat(ChatEvent{
		RequestID: id,
		Done:      true,
		Aborted:   aborted,
		Answer:    answer,
		Proposals: proposals,
	})
}

// drainChatDeltas forwards streamed deltas to the event sink while
// accumulating the answer. ok=false means a delta carried a terminal error
// (already emitted to the sink).
func (a *App) drainChatDeltas(id string, deltas <-chan ai.StreamDelta) (string, bool) {
	var answer strings.Builder

	for delta := range deltas {
		if delta.Err != nil {
			a.emitChat(ChatEvent{RequestID: id, Error: delta.Err.Error(), Done: true})

			return "", false
		}

		answer.WriteString(delta.Content)
		a.emitChat(ChatEvent{RequestID: id, Delta: delta.Content})
	}

	return answer.String(), true
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
