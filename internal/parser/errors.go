package parser

import (
	"errors"
	"fmt"

	"github.com/pelletier/go-toml/v2"
)

// ParseError represents an error that occurred during TOML parsing.
// It includes line number information and context for human-readable output.
type ParseError struct {
	// Message is a brief description of the error.
	Message string
	// Line is the line number where the error occurred (1-indexed, 0 if unknown).
	Line int
	// Context provides additional context about the error location.
	Context string
	// Cause is the underlying error that caused this parse error.
	Cause error
}

// Error returns a human-readable error message.
func (e *ParseError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("parse error at line %d: %s", e.Line, e.Message)
	}
	if e.Context != "" {
		return fmt.Sprintf("parse error: %s (%s)", e.Message, e.Context)
	}
	return fmt.Sprintf("parse error: %s", e.Message)
}

// Unwrap returns the underlying cause of the error.
func (e *ParseError) Unwrap() error {
	return e.Cause
}

// wrapDecodeError wraps a go-toml DecodeError to extract line number information.
// If the error is not a DecodeError, it returns a generic ParseError.
func wrapDecodeError(err error) error {
	var de *toml.DecodeError
	if errors.As(err, &de) {
		// DecodeError.String() already provides a nicely formatted error
		// with line number and context. We extract the line for our ParseError.
		line := extractLineFromDecodeError(de)
		return &ParseError{
			Message: de.Error(),
			Line:    line,
			Cause:   de,
		}
	}
	return &ParseError{
		Message: err.Error(),
		Cause:   err,
	}
}

// extractLineFromDecodeError extracts the line number from a DecodeError.
// go-toml v2 DecodeError has a Position() method that returns row and column.
func extractLineFromDecodeError(de *toml.DecodeError) int {
	row, _ := de.Position()
	return row
}
