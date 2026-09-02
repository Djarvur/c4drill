// uri.go converts between LSP document URIs and filesystem paths. Documents
// are modeled as a set keyed by URI, each carrying its canonical on-disk
// path: [[include]] paths resolve relative to that path's directory (INC-02)
// and the include graph keys on the canonical form.

package lsp

import (
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
)

// uriToPath converts a file:// document URI to a filesystem path. A URI that
// is not parseable as file:// falls back to its literal text (lenient: some
// clients hand over plain paths for untitled buffers).
func uriToPath(raw DocumentURI) string {
	u, err := url.Parse(string(raw))
	if err != nil || u.Scheme != "file" {
		return string(raw)
	}

	p, err := url.PathUnescape(u.Path)
	if err != nil {
		return u.Path
	}

	if runtime.GOOS == "windows" && strings.HasPrefix(p, "/") {
		p = strings.TrimPrefix(p, "/")
	}

	return p
}

// pathToURI converts a filesystem path to a file:// document URI.
func pathToURI(path string) DocumentURI {
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(path)}

	return DocumentURI(u.String())
}

// canonicalPath normalizes a filesystem path to the absolute, cleaned form
// the include graph and buffer overlay key on. It mirrors the resolver's own
// canonicalization (filepath.Clean + filepath.Abs, symlinks unresolved).
func canonicalPath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}

	return filepath.Clean(abs)
}
