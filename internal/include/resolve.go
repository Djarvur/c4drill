// Package include resolves [[include]] directives into a single merged
// *parser.Model. It runs as Stage 1a in the c4drill pipeline (the FIRST
// pre-processing pass), before template.Expand (Phase 31) and validator.Validate.
//
// Semantics (see .planning/phases/32-include-directive-multi-file/32-CONTEXT.md):
//   - INC-02: paths resolve relative to the INCLUDING file's directory.
//   - INC-03: transitive includes resolve recursively.
//   - INC-04: cycle (self or mutual) is a fatal *parser.ParseError naming the cycle.
//   - INC-06: once=true opts into the visited-set dedup.
//   - D-11: a same-file diamond (same canonical path reached via two paths) is
//     auto-deduped silently; a cross-file unit-path collision hard-errors.
//   - INC-10/D-12: a missing include is an unconditional hard error.
//   - D-26 (Plan 35-05): include graphs may MIX .toml and .c4d files freely —
//     dispatch is per included file's extension (.c4d -> C4D front-end, .toml
//     -> TOML front-end) and merging happens at Model level; any other
//     extension is a hard error naming the accepted ones (T-35-05-01).
//
// Canonical paths (filepath.Clean + filepath.Abs) are the key for BOTH the
// cycle stack and the visited-set. filepath.Abs does NOT resolve symlinks
// (T-32-05 accepted for v1.10 — C4Drill is author-controlled local tooling).
package include

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Djarvur/c4drill/internal/c4d"
	"github.com/Djarvur/c4drill/internal/parser"
)

// maxIncludeDepth caps the recursion depth as defense-in-depth against
// pathological acyclic-but-deep graphs (T-32-03). The visited-set already
// bounds total work to O(number of distinct files), but the cap guards against
// a deeply-branching graph that re-enters via aliased paths before dedup fires.
const maxIncludeDepth = 100

// Resolve walks entry.Includes and merges every transitively-included file
// into entry, returning the merged model. entryDir is the directory of the
// entry file (filepath.Dir of the entry path) so [[include]] paths resolve
// relative to the including file (INC-02), independent of the CLI cwd.
//
// entryFile is the display name of the entry file, used in error attribution
// for the entry's own [[include]] directives (so errors name the real entry
// path, not a placeholder). For nested directives the included file's path is
// used directly. If entryFile is empty, entryDir is used as the attribution
// name.
//
// On success, the returned model's .Includes is drained to nil and its
// .Units/.UnitOrder/.Properties/.Templates/.Instantiations reflect the union
// of the entry and all transitively-included files per D-09/D-10/D-11/INC-08.
// Any failure (cycle, missing file, cross-file collision, properties conflict)
// is a *parser.ParseError naming the relevant file(s).
func Resolve(entry *parser.Model, entryDir, entryFile string) (*parser.Model, error) {
	return ResolveWithReader(entry, entryDir, entryFile, nil)
}

// ResolveWithReader is Resolve with an injectable file reader: read is called
// for every transitively-included file's bytes. nil selects os.ReadFile —
// byte-identical behavior to Resolve. The LSP (issue #32) passes an overlay
// reader that serves unsaved editor buffers by canonical path, so including
// documents validate against what the author SEES, not what is on disk;
// error messages are unaffected (reader failures produce the same
// include-not-found attribution as missing files).
func ResolveWithReader(
	entry *parser.Model,
	entryDir, entryFile string,
	read func(path string) ([]byte, error),
) (*parser.Model, error) {
	visited := map[string]bool{}

	if entryFile == "" {
		entryFile = entryDir
	}

	return resolve(entry, entryDir, entryFile, nil, visited, read)
}

// resolve is the recursive worker. stack is the current ancestor chain (for
// cycle detection, INC-04); visited is the global already-included set (for
// once + same-file diamond dedup, INC-06/D-11). includingFile is the display
// name of the file whose Includes we are walking (used in missing-file errors).
func resolve(
	m *parser.Model,
	includingDir string,
	includingFile string,
	stack []string,
	visited map[string]bool,
	read func(path string) ([]byte, error),
) (*parser.Model, error) {
	if len(stack) > maxIncludeDepth {
		return nil, &parser.ParseError{
			Message: fmt.Sprintf("include depth exceeded %d (cycle or pathological graph)", maxIncludeDepth),
			Context: includingFile,
		}
	}

	for _, dir := range m.Includes {
		skip, err := resolveOne(&m, dir, includingDir, includingFile, stack, visited, read)
		if err != nil {
			return nil, err
		}

		_ = skip // 'skip' is only the loop-control signal; resolveOne mutates m in place
	}

	// Drain Includes: after resolution, the merged model has no further
	// includes to process (the pipeline's downstream stages never see them).
	m.Includes = nil

	return m, nil
}

// resolveOne processes a single [[include]] directive: canonicalize the path,
// cycle-check against the stack, dedup via the visited-set, ParseFile + recurse,
// then merge into the entry model. Returns skip=true when the directive is
// deduped (once/same-file diamond) so the caller knows no merge happened.
// Mutates *m in place on merge.
func resolveOne(
	m **parser.Model,
	dir parser.IncludeDirective,
	includingDir string,
	includingFile string,
	stack []string,
	visited map[string]bool,
	read func(path string) ([]byte, error),
) (bool, error) {
	absPath, err := canonicalize(dir.Path, includingDir)
	if err != nil {
		return false, &parser.ParseError{
			Message: "include path canonicalization failed",
			Context: fmt.Sprintf("%s (included from %s)", dir.Path, includingFile),
			Cause:   err,
		}
	}

	// Cycle check (INC-04): the current chain is the stack. A repeat of any
	// ancestor is a cycle, fatal.
	if slices.Contains(stack, absPath) {
		cycle := append(append([]string{}, stack...), absPath)

		return false, &parser.ParseError{
			Message: "include cycle detected: " + strings.Join(cycle, " -> "),
			Context: includingFile,
		}
	}

	// once + same-file diamond dedup (INC-06/D-11). once=true files are
	// skipped once visited. A same-file diamond (the same file reached via
	// two non-ancestral paths) is byte-identical by construction, so the
	// visited-set silently dedups it regardless of the once flag.
	if visited[absPath] {
		return true, nil
	}

	// Extension gate (T-35-05-01): dispatch is extension-based and fails
	// closed on unknown types — no fallback parsing, no content sniffing.
	if err := checkIncludeExtension(absPath, dir.Path, includingFile); err != nil {
		return false, err
	}

	// Parse the included model through the front-end its extension selects
	// (D-26: .toml -> TOML front-end, .c4d -> C4D front-end; merging happens
	// at Model level so graphs mix formats freely). Missing file →
	// *ParseError naming both the referenced path and the including file
	// (INC-10/D-12).
	included, err := parseIncludedFile(absPath, read)
	if err != nil {
		return false, &parser.ParseError{
			Message: "include not found: " + dir.Path,
			Context: fmt.Sprintf("%s (included from %s)", absPath, includingFile),
			Cause:   err,
		}
	}

	// Recurse: resolve the included file's own [[include]] directives with the
	// updated stack and the included file's directory as the new includingDir
	// (INC-02 relative-to-including-file, INC-03 transitive).
	newStack := append(append([]string{}, stack...), absPath)

	included, err = resolve(included, filepath.Dir(absPath), absPath, newStack, visited, read)
	if err != nil {
		return false, err
	}

	visited[absPath] = true

	// Merge the now-fully-resolved included model into the entry model.
	*m, err = merge(*m, included, includingFile, absPath)
	if err != nil {
		return false, err
	}

	return false, nil
}

// canonicalize resolves a relative include path against the including file's
// directory and produces an absolute, cleaned path. The result is the key for
// BOTH the cycle stack and the visited-set. filepath.Abs does NOT resolve
// symlinks (T-32-05 accepted for v1.10).
func canonicalize(path, includingDir string) (string, error) {
	joined := filepath.Join(includingDir, path)

	abs, err := filepath.Abs(joined)
	if err != nil {
		// filepath.Abs only fails when filepath.EvalSymlinks fails or the cwd
		// is unreadable; surface a wrapped error so callers can attribute it.
		return "", fmt.Errorf("canonicalize include path %q: %w", path, err)
	}

	return filepath.Clean(abs), nil
}

// Accepted include extensions (D-26): include graphs may mix these freely —
// dispatch is per file, merging is at Model level.
const (
	extToml = ".toml"
	extC4d  = ".c4d"
)

// checkIncludeExtension enforces the extension gate (T-35-05-01): only .toml
// and .c4d files may be included. Anything else is a hard *parser.ParseError
// naming the accepted extensions — no fallback parsing, no content sniffing,
// so a hostile or mistyped path cannot smuggle content through an unexpected
// decoder.
func checkIncludeExtension(absPath, displayPath, includingFile string) error {
	ext := filepath.Ext(absPath)

	switch ext {
	case extToml, extC4d:
		return nil
	default:
		return &parser.ParseError{
			Message: fmt.Sprintf("unsupported include extension %q (accepted: %s, %s)",
				ext, extToml, extC4d),
			Context: fmt.Sprintf("%s (included from %s)", displayPath, includingFile),
		}
	}
}

// parseIncludedFile parses an included file through the front-end its
// extension selects (D-26): .c4d -> the C4D front-end, .toml -> the TOML
// front-end. Bytes come through read (nil = os.ReadFile) so embedders can
// overlay open-editor buffers; error messages are identical either way —
// read failures produce the same "failed to read file" ParseError ParseFile
// would, and c4d errors carry the file path via ParseNamed's attribution.
// The extension gate ran before this call, so the default branch is
// defensive only.
//
//nolint:wrapcheck // resolveOne wraps the returned error with include-not-found attribution
func parseIncludedFile(path string, read func(path string) ([]byte, error)) (*parser.Model, error) {
	if read == nil {
		read = os.ReadFile
	}

	data, err := read(path)
	if err != nil {
		return nil, &parser.ParseError{Message: "failed to read file", Context: path, Cause: err}
	}

	switch filepath.Ext(path) {
	case extC4d:
		return c4d.ParseNamed(path, data)
	default:
		return parser.Parse(data)
	}
}
