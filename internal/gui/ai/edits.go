// edits.go is the structured edit-proposal pipeline: the assistant's reply
// is scanned for ```c4drill-edit path=... fenced blocks, each proposal is
// validated against the opened project, and a line diff is computed for the
// confirmation UI. Application happens ONLY on the user's explicit confirm
// (ApplyEdits), and re-checks the scope at that moment.

package ai

import (
	"errors"
	"fmt"
	"path"
	"strings"
)

// edit scope errors (wrapped into the Proposal.Error text).
var (
	errEditEmptyPath    = errors.New("edit proposal: empty path")
	errEditPathScope    = errors.New("edit proposal path must stay inside the project")
	errEditNotModelFile = errors.New("edit proposal: not a .toml/.c4d model file")
)

// EditBlock fences one proposed edit inside the assistant reply.
const (
	editFenceOpen  = "```c4drill-edit"
	editFenceClose = "```"
)

// Proposal is one parsed edit proposal.
type Proposal struct {
	// Path is the project-relative target file.
	Path string `json:"path"`
	// NewContent is the proposed full file content.
	NewContent string `json:"newContent"`
	// OldContent is the current on-disk content ([]byte as string) captured
	// at parse time for the diff display.
	OldContent string `json:"oldContent"`
	// Valid reports whether the proposal passed validation.
	Valid bool `json:"valid"`
	// Error explains why the proposal is invalid ("" when valid).
	Error string `json:"error,omitempty"`
}

// ParseProposals scans an assistant reply for c4drill-edit blocks. validate
// receives each proposed path and reports whether it may be written (scope
// check: files in the opened project only — the ai package itself does not
// know the project). Files with no on-disk content yet read as "" (creation).
func ParseProposals(reply string, read func(relPath string) (string, error)) []Proposal {
	var proposals []Proposal

	lines := strings.Split(reply, "\n")

	for i := 0; i < len(lines); i++ {
		prefix, ok := editOpenLine(lines[i])
		if !ok {
			continue
		}

		target := strings.TrimSpace(strings.TrimPrefix(prefix, editFenceOpen))
		target = strings.Trim(target, "=") // tolerate "path= x" spacing quirks
		target = strings.TrimSpace(strings.TrimPrefix(target, "path="))
		target = strings.Trim(target, `"`)

		var content []string

		closed := false

		for j := i + 1; j < len(lines); j++ {
			if strings.TrimSpace(lines[j]) == editFenceClose {
				closed = true
				i = j

				break
			}

			content = append(content, lines[j])
		}

		if !closed || target == "" {
			continue // unterminated block or no path: not a proposal
		}

		proposals = append(proposals, buildProposal(target, strings.Join(content, "\n"), read))
	}

	return proposals
}

// editOpenLine reports a c4drill-edit fence opener, tolerating a language
// suffix after the path argument.
func editOpenLine(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, editFenceOpen) {
		return "", false
	}

	return trimmed, true
}

// buildProposal validates one proposal and captures the diff base. The
// returned Proposal is safe to hand to the UI even when invalid.
func buildProposal(target, newContent string, read func(relPath string) (string, error)) Proposal {
	p := Proposal{Path: target, NewContent: newContent}

	if err := checkPath(target); err != nil {
		p.Error = err.Error()

		return p
	}

	old, err := read(target)
	if err != nil {
		p.Error = err.Error()

		return p
	}

	p.OldContent = old
	p.Valid = true

	return p
}

// checkPath enforces the edit scope: project-relative, inside the project,
// a model file. It mirrors the backend's own path rules without importing
// them (path-shaped only — no filesystem access).
func checkPath(target string) error {
	if target == "" {
		return errEditEmptyPath
	}

	if strings.Contains(target, "..") || path.IsAbs(target) {
		return fmt.Errorf("%w: %q", errEditPathScope, target)
	}

	switch ext := path.Ext(target); strings.ToLower(ext) {
	case ".toml", ".c4d":
		return nil
	default:
		return fmt.Errorf("%w: %q", errEditNotModelFile, target)
	}
}

// DiffLine is one rendered diff row for the confirmation UI.
type DiffLine struct {
	Kind string `json:"kind"` // "ctx", "add", "del", "hunk"
	Text string `json:"text"`
}

// LineDiff computes a classic LCS line diff between old and new content.
// Pure and dependency-free; model files are small enough for O(n·m).
func LineDiff(oldText, newText string) []DiffLine {
	oldLines := splitLines(oldText)
	newLines := splitLines(newText)

	ops := lcsOps(oldLines, newLines)

	out := make([]DiffLine, 0, len(ops))

	for _, op := range ops {
		switch op.kind {
		case opKeep:
			out = append(out, DiffLine{Kind: "ctx", Text: op.text})
		case opDelete:
			out = append(out, DiffLine{Kind: "del", Text: op.text})
		case opInsert:
			out = append(out, DiffLine{Kind: "add", Text: op.text})
		}
	}

	return out
}

// HasChanges reports whether the diff is non-empty (identical content
// proposals are valid but pointless to apply).
func HasChanges(diff []DiffLine) bool {
	for _, l := range diff {
		if l.Kind == "add" || l.Kind == "del" {
			return true
		}
	}

	return false
}

// --- LCS machinery --------------------------------------------------------

type (
	diffOpKind uint8
	diffOp     struct {
		kind diffOpKind
		text string
	}
)

const (
	opKeep diffOpKind = iota
	opDelete
	opInsert
)

// splitLines splits into lines, dropping the trailing empty element of a
// final newline.
func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")

	lines := strings.Split(s, "\n")
	if len(lines) > 1 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return lines
}

// lcsOps walks the LCS backtrack table emitting keep/delete/insert runs.
func lcsOps(oldLines, newLines []string) []diffOp {
	n, m := len(oldLines), len(newLines)

	// lcs[i][j] = LCS length of oldLines[i:] vs newLines[j:].
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}

	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			switch {
			case oldLines[i] == newLines[j]:
				lcs[i][j] = lcs[i+1][j+1] + 1
			case lcs[i+1][j] >= lcs[i][j+1]:
				lcs[i][j] = lcs[i+1][j]
			default:
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}

	ops := make([]diffOp, 0, n+m)

	i, j := 0, 0
	for i < n && j < m {
		switch {
		case oldLines[i] == newLines[j]:
			ops = append(ops, diffOp{opKeep, oldLines[i]})
			i++
			j++
		case lcs[i+1][j] >= lcs[i][j+1]:
			ops = append(ops, diffOp{opDelete, oldLines[i]})
			i++
		default:
			ops = append(ops, diffOp{opInsert, newLines[j]})
			j++
		}
	}

	for ; i < n; i++ {
		ops = append(ops, diffOp{opDelete, oldLines[i]})
	}

	for ; j < m; j++ {
		ops = append(ops, diffOp{opInsert, newLines[j]})
	}

	return ops
}
