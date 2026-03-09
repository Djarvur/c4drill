// Package validator provides validation infrastructure for C4 model integrity checks.
// It includes error types with human-readable formatting, unit index building for O(1) lookup,
// and Levenshtein-based suggestion helpers for typo detection.
package validator

import "fmt"

// ValidationError represents a single validation failure.
// It provides human-readable error messages following the user's specified format.
type ValidationError struct {
	// Message is a concise description of the validation error.
	Message string
	// Path is the full dotted path to the unit (e.g., "mainapp.api.handler").
	Path string
	// Line is the line number where the error occurred (best-effort, 0 if unknown).
	Line int
}

// Error returns a human-readable single-line error message.
// Format priority:
//   - With line: "error: {message} at line {N}"
//   - With path only: "error: {message} in {path}"
//   - Neither: "error: {message}"
func (e *ValidationError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("error: %s at line %d", e.Message, e.Line)
	}

	if e.Path != "" {
		return fmt.Sprintf("error: %s in %s", e.Message, e.Path)
	}

	return "error: " + e.Message
}

// ValidationErrors is a collection of validation errors.
// It implements the error interface and provides a count summary.
type ValidationErrors []*ValidationError

// Error returns a summary of the validation errors.
// Returns "no errors" if empty, or "N error(s) found" otherwise.
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "no errors"
	}

	count := len(ve)
	if count == 1 {
		return "1 error found"
	}

	return fmt.Sprintf("%d errors found", count)
}
