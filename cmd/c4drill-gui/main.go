// Command c4drill-gui is the c4drill desktop app (issue #31): a Wails v2
// shell whose backend binds the in-process pipeline — the shared LSP server
// core (internal/lsp) and the render/export packages — to a CodeMirror 6 +
// SVG frontend.
//
// Two transports, one backend:
//
//	c4drill-gui          native desktop window (Wails v2, system webview)
//	c4drill-gui --serve  HTTP fallback on 127.0.0.1 — same UI in a regular
//	                     browser (no webview required; also used by the smoke
//	                     e2e test)
//
// Both speak the same Dispatch(method, params) JSON protocol implemented in
// internal/gui/app, so behavior cannot drift between them.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/Djarvur/c4drill/internal/gui"
	"github.com/Djarvur/c4drill/internal/gui/app"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// version is stamped at build time (-ldflags "-X main.version=...").
var version = "dev"

func main() {
	serve := flag.Bool("serve", false, "serve the UI over HTTP instead of opening a native window")
	addr := flag.String("addr", "127.0.0.1:5278", "listen address for --serve")
	dir := flag.String("dir", "", "project directory to open at startup")

	flag.Parse()

	app.SetVersion(version)

	backend := app.New(nil)

	if *dir != "" {
		if _, err := backend.OpenProject(*dir); err != nil {
			log.Fatalf("gui: open project: %v", err)
		}
	}

	if *serve {
		if err := runHTTP(backend, *addr); err != nil {
			log.Fatal(err)
		}

		return
	}

	if err := runWails(backend); err != nil {
		log.Fatal(err)
	}
}

// desktop is the Wails-bound struct. A single Dispatch method carries the
// whole JSON protocol (see internal/gui/app), keeping the frontend identical
// across the desktop and browser transports.
type desktop struct {
	backend *app.App

	//nolint:containedctx // the Wails lifecycle hands its context via OnStartup
	ctx context.Context
}

func newDesktop(backend *app.App) *desktop {
	return &desktop{backend: backend}
}

// Dispatch routes one backend call: method + JSON params → JSON result.
func (g *desktop) Dispatch(method, params string) (string, error) {
	res, err := g.backend.Dispatch(method, json.RawMessage(params))
	if res == nil {
		res = json.RawMessage("null")
	}

	return string(res), err
}

func (g *desktop) startup(ctx context.Context) {
	g.ctx = ctx
}

func runWails(backend *app.App) error {
	g := newDesktop(backend)

	// Backend → frontend events (diagnostics, P1 chat streaming) ride Wails
	// runtime events under the single "backend" channel; the frontend's
	// rpc.ts unwraps {event, payload} frames.
	backend.SetEventSink(func(event string, payload any) {
		if g.ctx == nil {
			return
		}

		wailsruntime.EventsEmit(g.ctx, "backend", map[string]any{"event": event, "payload": payload})
	})

	//nolint:wrapcheck // main maps this to log.Fatal
	return wails.Run(&options.App{
		Title:     "c4drill",
		Width:     1440,
		Height:    900,
		MinWidth:  960,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: gui.Assets,
		},
		OnStartup: g.startup,
		Bind:      []interface{}{g},
		Mac: &mac.Options{
			About: &mac.AboutInfo{
				Title:   "c4drill " + version,
				Message: "C4 architecture authoring: editor + live diagram preview",
			},
		},
	})
}

// runHTTP is the webview-less fallback: the same embedded frontend and the
// same Dispatch protocol over localhost HTTP. Events stream as SSE.
func runHTTP(backend *app.App, addr string) error {
	handler := newHandler(backend)

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("c4drill GUI (HTTP mode) serving on http://%s — Ctrl-C to stop", addr)

	err := srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err //nolint:wrapcheck // main maps this to log.Fatal
}

// newHandler builds the HTTP transport: static frontend assets, the
// /api/dispatch RPC endpoint, and the /api/events SSE stream.
func newHandler(backend *app.App) http.Handler {
	hub := newEventHub()

	backend.SetEventSink(func(event string, payload any) {
		hub.broadcast(eventFrame{Event: event, Payload: payload})
	})

	dist, err := fs.Sub(gui.Assets, "frontend/dist")
	if err != nil {
		panic(fmt.Sprintf("gui: embed frontend/dist: %v", err)) // build-time invariant
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(dist)))
	mux.HandleFunc("/api/dispatch", dispatchHandler(backend))
	mux.HandleFunc("/api/events", hub.eventsHandler)

	return mux
}

// dispatchHandler adapts the JSON request body {method, params} to the
// backend's Dispatch.
func dispatchHandler(backend *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			Method string          `json:"method"`
			Params json.RawMessage `json:"params"`
		}

		body, err := io.ReadAll(io.LimitReader(r.Body, 32<<20))
		if err == nil {
			err = json.Unmarshal(body, &req)
		}

		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad request: " + err.Error()})
			return
		}

		res, err := backend.Dispatch(req.Method, req.Params)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return
		}

		if res == nil {
			res = json.RawMessage("null")
		}

		writeJSON(w, http.StatusOK, map[string]json.RawMessage{"result": res})
	}
}

// eventsHandler streams backend events to the browser as Server-Sent Events.
func (h *eventHub) eventsHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	ch := h.subscribe()
	defer h.unsubscribe(ch)

	for {
		select {
		case <-r.Context().Done():
			return
		case frame, ok := <-ch:
			if !ok {
				return
			}

			data, _ := json.Marshal(frame) //nolint:errchkjson // eventFrame is JSON-safe by construction

			if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
				return
			}

			flusher.Flush()
		}
	}
}

// eventFrame is one SSE payload: {event, payload} — the same shape the Wails
// transport emits under the "backend" channel.
type eventFrame struct {
	Event   string `json:"event"`
	Payload any    `json:"payload"`
}

// eventHub fans backend events out to subscribed SSE clients. Delivery is
// best-effort: a slow client's channel is dropped rather than blocking the
// backend.
type eventHub struct {
	chans map[chan eventFrame]struct{}
}

func newEventHub() *eventHub {
	return &eventHub{chans: make(map[chan eventFrame]struct{})}
}

func (h *eventHub) subscribe() chan eventFrame {
	ch := make(chan eventFrame, 64)

	h.chans[ch] = struct{}{}

	return ch
}

func (h *eventHub) unsubscribe(ch chan eventFrame) {
	if _, ok := h.chans[ch]; ok {
		delete(h.chans, ch)
		close(ch)
	}
}

func (h *eventHub) broadcast(frame eventFrame) {
	for ch := range h.chans {
		select {
		case ch <- frame:
		default: // slow client: drop the frame
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(body) //nolint:errchkjson // best-effort encode
}
