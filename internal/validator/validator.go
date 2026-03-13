package validator

import (
	"fmt"
	"io"

	"github.com/Djarvur/c4drill/internal/parser"
)

// typicalErrorCount is the preallocation size for error slices.
const typicalErrorCount = 4

// Validate checks a parsed model for semantic errors.
// It runs all validation rules and collects all errors (not fail-fast).
// Returns nil if the model is valid, or a slice of ValidationErrors.
func Validate(m *parser.Model) ValidationErrors {
	if m == nil {
		return nil
	}

	// Build index for O(1) lookup
	index := BuildIndex(m.Units, "")

	// Populate incoming links from other units' outgoing links
	populateIncomingLinks(index)

	// Collect all errors from all rules (preallocate for typical case)
	errors := make(ValidationErrors, 0, typicalErrorCount)

	errors = append(errors, ValidateReferences(index)...)
	errors = append(errors, ValidateSubunitRules(index)...)
	errors = append(errors, ValidateLinkRules(index)...)
	errors = append(errors, ValidateOrphanUnits(index)...)

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
		_, _ = fmt.Fprintln(w, err.Error())
	}

	// Print summary
	count := len(errors)
	if count == 1 {
		_, _ = fmt.Fprintln(w, "1 error found")
	} else {
		_, _ = fmt.Fprintf(w, "%d errors found\n", count)
	}

	return 1
}
