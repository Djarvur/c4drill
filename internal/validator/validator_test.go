package validator_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	t.Run("returns nil for nil model", func(t *testing.T) {
		t.Parallel()

		errors := validator.Validate(nil)
		assert.Nil(t, errors)
	})

	t.Run("returns nil for valid model", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Units: map[string]*model.Unit{
				"api": {
					Type: model.TypeSystem,
					Links: map[string]model.Link{
						"db": {Target: "db"},
					},
				},
				"db": {
					Type: model.TypeDb,
				},
			},
		}

		errors := validator.Validate(m)
		assert.Nil(t, errors)
	})

	t.Run("returns all errors for invalid model", func(t *testing.T) {
		t.Parallel()

		// Create model with multiple violations:
		// 1. Undefined reference
		// 2. Person with subunits
		// 3. System with subunits has links
		m := &parser.Model{
			Units: map[string]*model.Unit{
				"person": {
					Type: model.TypePerson,
					Subunits: map[string]*model.Unit{
						"child": {Type: model.TypeComponent},
					},
				},
				"system": {
					Type: model.TypeSystem,
					Links: map[string]model.Link{
						"undefined": {Target: "undefined"},
					},
					Subunits: map[string]*model.Unit{
						"api": {Type: model.TypeContainer},
					},
				},
			},
		}

		errors := validator.Validate(m)
		require.NotNil(t, errors)
		// Should have at least 3 errors: undefined reference, person with subunits, system with links
		assert.GreaterOrEqual(t, len(errors), 3)
	})

	t.Run("handles empty units map", func(t *testing.T) {
		t.Parallel()

		m := &parser.Model{
			Units: map[string]*model.Unit{},
		}

		errors := validator.Validate(m)
		assert.Nil(t, errors)
	})

	t.Run("processes nested units correctly", func(t *testing.T) {
		t.Parallel()

		// Nested unit with invalid reference
		m := &parser.Model{
			Units: map[string]*model.Unit{
				"system": {
					Type: model.TypeSystem,
					Subunits: map[string]*model.Unit{
						"api": {
							Type: model.TypeContainer,
							Links: map[string]model.Link{
								"undefined": {Target: "undefined"},
							},
						},
					},
				},
			},
		}

		errors := validator.Validate(m)
		require.NotNil(t, errors)
		require.Len(t, errors, 1)
		assert.Contains(t, errors[0].Message, "system.api")
	})
}

func TestReportErrors(t *testing.T) {
	t.Parallel()

	t.Run("prints all errors", func(t *testing.T) {
		t.Parallel()

		errors := validator.ValidationErrors{
			{Message: "first error", Path: "unit1"},
			{Message: "second error", Path: "unit2"},
		}

		var buf bytes.Buffer
		exitCode := validator.ReportErrors(errors, &buf)

		output := buf.String()
		assert.Contains(t, output, "first error")
		assert.Contains(t, output, "second error")
		assert.Equal(t, 1, exitCode)
	})

	t.Run("prints singular summary for one error", func(t *testing.T) {
		t.Parallel()

		errors := validator.ValidationErrors{
			{Message: "single error", Path: "unit"},
		}

		var buf bytes.Buffer
		exitCode := validator.ReportErrors(errors, &buf)

		output := buf.String()
		assert.Contains(t, output, "1 error found")
		assert.Equal(t, 1, exitCode)
	})

	t.Run("prints plural summary for multiple errors", func(t *testing.T) {
		t.Parallel()

		errors := validator.ValidationErrors{
			{Message: "error 1", Path: "unit1"},
			{Message: "error 2", Path: "unit2"},
			{Message: "error 3", Path: "unit3"},
		}

		var buf bytes.Buffer
		exitCode := validator.ReportErrors(errors, &buf)

		output := buf.String()
		assert.Contains(t, output, "3 errors found")
		assert.Equal(t, 1, exitCode)
	})

	t.Run("returns 0 for empty error list", func(t *testing.T) {
		t.Parallel()

		var errors validator.ValidationErrors

		var buf bytes.Buffer
		exitCode := validator.ReportErrors(errors, &buf)

		output := buf.String()
		assert.Empty(t, output)
		assert.Equal(t, 0, exitCode)
	})

	t.Run("prints to given writer", func(t *testing.T) {
		t.Parallel()

		errors := validator.ValidationErrors{
			{Message: "test error", Path: "unit"},
		}

		var buf bytes.Buffer
		_ = validator.ReportErrors(errors, &buf)

		// Verify output went to the buffer (not stdout/stderr)
		assert.Contains(t, buf.String(), "test error")
	})

	t.Run("formats error with line number", func(t *testing.T) {
		t.Parallel()

		errors := validator.ValidationErrors{
			{Message: "test error", Path: "unit", Line: 42},
		}

		var buf bytes.Buffer
		_ = validator.ReportErrors(errors, &buf)

		output := buf.String()
		assert.Contains(t, output, "at line 42")
	})

	t.Run("formats error with path only", func(t *testing.T) {
		t.Parallel()

		errors := validator.ValidationErrors{
			{Message: "test error", Path: "unit.subunit"},
		}

		var buf bytes.Buffer
		_ = validator.ReportErrors(errors, &buf)

		output := buf.String()
		assert.Contains(t, output, "in unit.subunit")
	})
}

func TestValidationErrorsInterface(t *testing.T) {
	t.Parallel()

	t.Run("Error returns no errors for empty slice", func(t *testing.T) {
		var errors validator.ValidationErrors
		assert.Equal(t, "no errors", errors.Error())
	})

	t.Run("Error returns singular for one error", func(t *testing.T) {
		errors := validator.ValidationErrors{{Message: "test"}}
		assert.Equal(t, "1 error found", errors.Error())
	})

	t.Run("Error returns plural for multiple errors", func(t *testing.T) {
		errors := validator.ValidationErrors{
			{Message: "test1"},
			{Message: "test2"},
		}
		assert.Equal(t, "2 errors found", errors.Error())
	})
}

func TestValidationErrorFormatting(t *testing.T) {
	t.Parallel()

	t.Run("formats with line number", func(t *testing.T) {
		err := &validator.ValidationError{
			Message: "undefined unit \"foo\"",
			Path:    "bar",
			Line:    15,
		}

		expected := `error: undefined unit "foo" at line 15`
		assert.Equal(t, expected, err.Error())
	})

	t.Run("formats with path only", func(t *testing.T) {
		err := &validator.ValidationError{
			Message: "undefined unit \"foo\"",
			Path:    "bar",
		}

		expected := `error: undefined unit "foo" in bar`
		assert.Equal(t, expected, err.Error())
	})

	t.Run("formats with message only", func(t *testing.T) {
		err := &validator.ValidationError{
			Message: "undefined unit \"foo\"",
		}

		expected := `error: undefined unit "foo"`
		assert.Equal(t, expected, err.Error())
	})
}

// Test type alias for io.Writer usage
func TestReportErrorsAcceptsIoWriter(t *testing.T) {
	t.Parallel()

	// This test verifies the function signature accepts any io.Writer
	var writer io.Writer = &bytes.Buffer{}
	errors := validator.ValidationErrors{{Message: "test"}}

	// Should compile and work
	exitCode := validator.ReportErrors(errors, writer)
	assert.Equal(t, 1, exitCode)
}
