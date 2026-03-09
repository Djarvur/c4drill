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

func TestValidate_NilModel(t *testing.T) {
	t.Parallel()

	errors := validator.Validate(nil)
	assert.Nil(t, errors)
}

func TestValidate_ValidModel(t *testing.T) {
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
}

func TestValidate_InvalidModel(t *testing.T) {
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
}

func TestValidate_EmptyUnits(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Units: map[string]*model.Unit{},
	}

	errors := validator.Validate(m)
	assert.Nil(t, errors)
}

func TestValidate_NestedUnits(t *testing.T) {
	t.Parallel()

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
}

func TestReportErrors_PrintsAllErrors(t *testing.T) {
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
}

func TestReportErrors_SingularSummary(t *testing.T) {
	t.Parallel()

	errors := validator.ValidationErrors{
		{Message: "single error", Path: "unit"},
	}

	var buf bytes.Buffer

	exitCode := validator.ReportErrors(errors, &buf)

	output := buf.String()
	assert.Contains(t, output, "1 error found")
	assert.Equal(t, 1, exitCode)
}

func TestReportErrors_PluralSummary(t *testing.T) {
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
}

func TestReportErrors_EmptyList(t *testing.T) {
	t.Parallel()

	var errors validator.ValidationErrors

	var buf bytes.Buffer

	exitCode := validator.ReportErrors(errors, &buf)

	output := buf.String()
	assert.Empty(t, output)
	assert.Equal(t, 0, exitCode)
}

func TestReportErrors_UsesWriter(t *testing.T) {
	t.Parallel()

	errors := validator.ValidationErrors{
		{Message: "test error", Path: "unit"},
	}

	var buf bytes.Buffer

	_ = validator.ReportErrors(errors, &buf)

	assert.Contains(t, buf.String(), "test error")
}

func TestReportErrors_WithLineNumber(t *testing.T) {
	t.Parallel()

	errors := validator.ValidationErrors{
		{Message: "test error", Path: "unit", Line: 42},
	}

	var buf bytes.Buffer

	_ = validator.ReportErrors(errors, &buf)

	output := buf.String()
	assert.Contains(t, output, "at line 42")
}

func TestReportErrors_WithPathOnly(t *testing.T) {
	t.Parallel()

	errors := validator.ValidationErrors{
		{Message: "test error", Path: "unit.subunit"},
	}

	var buf bytes.Buffer

	_ = validator.ReportErrors(errors, &buf)

	output := buf.String()
	assert.Contains(t, output, "in unit.subunit")
}

// Test type alias for io.Writer usage.
func TestReportErrorsAcceptsIoWriter(t *testing.T) {
	t.Parallel()

	// This test verifies the function signature accepts any io.Writer
	var writer io.Writer = &bytes.Buffer{}

	errors := validator.ValidationErrors{{Message: "test"}}

	// Should compile and work
	exitCode := validator.ReportErrors(errors, writer)
	assert.Equal(t, 1, exitCode)
}
