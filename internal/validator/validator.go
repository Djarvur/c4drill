package validator

import (
	"fmt"
	"io"

	"github.com/Djarvur/c4drill/internal/parser"
)

// Validate checks a parsed model for semantic errors.
// It runs all validation rules and collects all errors (not fail-fast).
// Returns nil if the model is valid, or a slice of ValidationErrors.
func Validate(m *parser.Model) ValidationErrors {
	if m == nil {
		return nil
	}

	// Build index for O(1) lookup
	index := BuildIndex(m.Units, "")

	// Collect all errors from all rules
	var errors ValidationErrors
	errors = append(errors, ValidateReferences(m.Units, index)...)
	errors = append(errors, ValidateSubunitRules(m.Units, index)...)
	errors = append(errors, ValidateLinkRules(m.Units, index)...)

	if len(errors) == 0 {
		return nil
	}

	return errors
}

// ReportErrors prints validation errors to the writer and returns exit code.
// It prints each error on a separate line, followed by a summary.
// Returns 1 if errors exist, 0 if empty.
func ReportErrors(errors ValidationErrors, w io.Writer) int {
	if len(errors) == 0 {
		return 0
	}

	for _, err := range errors {
		fmt.Fprintln(w, err.Error())
	}

	// Print summary
	count := len(errors)
	if count == 1 {
		fmt.Fprintln(w, "1 error found")
	} else {
		fmt.Fprintf(w, "%d errors found\n", count)
	}

	return 1
}
