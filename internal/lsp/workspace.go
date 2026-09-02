// workspace.go is the cross-file awareness layer: the include graph over the
// open-document set (each document carrying its on-disk directory, INC-02).
// Editing an included document republishes diagnostics for every open
// including document; closing or externally changing one does the same.

package lsp

import (
	"os"
	"path/filepath"
)

// revalidate republishes diagnostics for the document at changedPath (when
// open) plus every other open document whose [[include]] closure reaches it.
// It runs while the server lock is held — notification delivery may not call
// back into the server.
func (s *Server) revalidate(changedPath string) {
	if doc, ok := s.docByPath(changedPath); ok {
		s.publishDocument(doc)
	}

	for _, doc := range s.docs {
		if doc.Path == changedPath {
			continue // already republished above
		}

		if s.includesTransitively(doc, changedPath) {
			s.publishDocument(doc)
		}
	}
}

// publishDocument computes a document's diagnostics and publishes them with
// the document's current version.
func (s *Server) publishDocument(doc *document) {
	version := doc.Version
	s.publish(doc.URI, &version, s.computeDiagnostics(doc))
}

// includesTransitively reports whether entry's [[include]] closure — through
// open buffers and disk alike — reaches target (both canonical paths).
func (s *Server) includesTransitively(entry *document, target string) bool {
	return s.includesFrom(entry.Path, string(entry.Text), target, make(map[string]bool))
}

// includesFrom walks the [[include]] chain of the document (path, text)
// looking for target. visited guards against cycles so a broken include
// graph cannot spin the walk; a document whose own parse fails has no
// discoverable includes.
func (s *Server) includesFrom(path, text, target string, visited map[string]bool) bool {
	if visited[path] {
		return false
	}

	visited[path] = true

	m, err := parseByExt(path, []byte(text))
	if err != nil {
		return false
	}

	dir := filepath.Dir(path)
	for _, inc := range m.Includes {
		p := canonicalPath(filepath.Join(dir, inc.Path))
		if p == target {
			return true
		}

		data, err := s.readForWalk(p)
		if err != nil {
			continue // unreadable branch of the graph cannot reach target
		}

		if s.includesFrom(p, string(data), target, visited) {
			return true
		}
	}

	return false
}

// readForWalk fetches a file's bytes for include-graph walks: open buffers
// first (mirroring validation), then disk.
func (s *Server) readForWalk(path string) ([]byte, error) {
	if doc, ok := s.docByPath(path); ok {
		return doc.Text, nil
	}

	//nolint:gosec // G304: include-graph walks read the resolver-controlled paths by design
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err //nolint:wrapcheck // walk callers only branch on it
	}

	return data, nil
}

// docByPath finds the open document whose canonical path matches.
func (s *Server) docByPath(path string) (*document, bool) {
	for _, doc := range s.docs {
		if doc.Path == path {
			return doc, true
		}
	}

	return nil, false
}
