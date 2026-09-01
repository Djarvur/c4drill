// errors.go is the app package's static error set (the repo's error style:
// sentinel errors wrapped with stage context, never bare dynamic errors).

package app

import "errors"

// Exported sentinels (part of the package contract; transports and tests
// match on them with errors.Is).
var (
	// ErrPathOutsideProject rejects paths that escape the project root —
	// the GUI's edit scope is the opened project (issue #31).
	ErrPathOutsideProject = errors.New("path outside project")

	// ErrInvalidExportFormat mirrors the CLI's -f validation message.
	ErrInvalidExportFormat = errors.New("invalid format: must be dot, svg, html, png, or plantuml")

	// ErrUnknownMethod is the dispatch table's miss.
	ErrUnknownMethod = errors.New("unknown method")

	// ErrNotModelFile rejects operations on files the pipeline cannot parse.
	ErrNotModelFile = errors.New("not a .toml/.c4d model file")
)

// internal sentinels (wrapped with context by the callers).
var (
	// errNoProjectOpen guards every call that needs an opened project.
	errNoProjectOpen = errors.New("no project open")

	// errEmptyPath rejects empty project-relative paths.
	errEmptyPath = errors.New("empty path")

	// errGenerateView / errBuildGraph are the pipeline stage failures, named
	// like the CLI's.
	errGenerateView = errors.New("failed to generate view")
	errBuildGraph   = errors.New("failed to build graph")

	// errValidationFailed wraps the validator's error list at export time.
	errValidationFailed = errors.New("validation failed")

	// errNotInternalLink marks hrefs that are not diagram navigation links.
	errNotInternalLink = errors.New("not an internal navigation link")

	// errBadNavigation marks drill hrefs that resolve outside the diagram
	// set.
	errBadNavigation = errors.New("navigation link outside the diagram set")

	// errEmptyBasename guards resolveDrillTarget against a degenerate call.
	errEmptyBasename = errors.New("empty basename")

	// errLSPNoResponse / errLSPResponse surface in-memory transport-level
	// failures of the shared LSP server core.
	errLSPNoResponse    = errors.New("lsp: no response")
	errLSPRequestFailed = errors.New("lsp: request failed")
)
