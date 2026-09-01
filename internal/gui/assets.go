// Package gui carries the c4drill GUI's embedded frontend build.
//
// The embed lives here rather than in cmd/c4drill-gui because go:embed
// patterns cannot reach outside the package directory: the vite frontend is
// developed and built under internal/gui/frontend, and both transports —
// the Wails desktop window and the --serve HTTP fallback — serve the same
// frontend/dist from this single embedded snapshot.
//
// Without a prior `npm run build` the committed frontend/dist/.gitkeep keeps
// the embed valid and the app serves a placeholder page (tests and the
// backend API work; the UI needs the vite build).
package gui

import (
	"embed"
)

// Assets is the embedded frontend build (all:frontend/dist), passed raw to
// the Wails asset server and fs.Sub-ed down to dist/ by the HTTP fallback
// (see cmd/c4drill-gui/main.go).
//
//go:embed all:frontend/dist
var Assets embed.FS
