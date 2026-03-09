package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidationError_WithLine(t *testing.T) {
	err := &ValidationError{
		Message: `undefined unit "db1" referenced from "api"`,
		Path:    "api",
		Line:    15,
	}

	expected := `error: undefined unit "db1" referenced from "api" at line 15`
	assert.Equal(t, expected, err.Error())
}

func TestValidationError_WithoutLine(t *testing.T) {
	err := &ValidationError{
		Message: `undefined unit "db1"`,
		Path:    "",
		Line:    0,
	}

	expected := `error: undefined unit "db1"`
	assert.Equal(t, expected, err.Error())
}

func TestValidationError_WithPath(t *testing.T) {
	err := &ValidationError{
		Message: `invalid type for subunit`,
		Path:    "mainapp.api.handler",
		Line:    0,
	}

	expected := `error: invalid type for subunit in mainapp.api.handler`
	assert.Equal(t, expected, err.Error())
}

func TestValidationError_WithLineAndPath(t *testing.T) {
	err := &ValidationError{
		Message: `unit cannot have subunits`,
		Path:    "mainapp.api",
		Line:    42,
	}

	// Line takes precedence over path in formatting
	expected := `error: unit cannot have subunits at line 42`
	assert.Equal(t, expected, err.Error())
}

func TestValidationError_EmptyPathZeroLine(t *testing.T) {
	err := &ValidationError{
		Message: "generic validation error",
		Path:    "",
		Line:    0,
	}

	expected := "error: generic validation error"
	assert.Equal(t, expected, err.Error())
}

func TestValidationErrors_Empty(t *testing.T) {
	var errors ValidationErrors

	assert.Equal(t, "no errors", errors.Error())
	assert.Nil(t, errors)
}

func TestValidationErrors_Single(t *testing.T) {
	errors := ValidationErrors{
		{Message: "first error", Path: "", Line: 0},
	}

	assert.Equal(t, "1 error found", errors.Error())
}

func TestValidationErrors_Multiple(t *testing.T) {
	errors := ValidationErrors{
		{Message: "first error", Path: "", Line: 0},
		{Message: "second error", Path: "", Line: 0},
		{Message: "third error", Path: "", Line: 0},
	}

	assert.Equal(t, "3 errors found", errors.Error())
}

func TestValidationErrors_Append(t *testing.T) {
	var errors ValidationErrors

	errors = append(errors, &ValidationError{Message: "error 1"})
	errors = append(errors, &ValidationError{Message: "error 2"})

	assert.Len(t, errors, 2)
	assert.Equal(t, "2 errors found", errors.Error())
}
