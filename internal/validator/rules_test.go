package validator_test

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/validator"
)

func TestValidateReferences_ValidModel(t *testing.T) {
	t.Parallel()

	units := map[string]*model.Unit{
		"api": {
			Type: model.TypeSystem,
			Links: map[string]model.Link{
				"db": {Target: "db"},
			},
		},
		"db": {
			Type: model.TypeDb,
		},
	}

	index := validator.BuildIndex(units, "")
	errors := validator.ValidateReferences(index)

	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errors), errors)
	}
}

func TestValidateReferences_UndefinedLinksTarget(t *testing.T) {
	t.Parallel()

	units := map[string]*model.Unit{
		"api": {
			Type: model.TypeSystem,
			Links: map[string]model.Link{
				"undefined": {Target: "undefined"},
			},
		},
	}

	index := validator.BuildIndex(units, "")
	errors := validator.ValidateReferences(index)

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	expectedMsg := `undefined unit "undefined" referenced from "api"`
	if errors[0].Message != expectedMsg {
		t.Errorf("expected message %q, got %q", expectedMsg, errors[0].Message)
	}
}

func TestValidateReferences_UndefinedLinksFromSource(t *testing.T) {
	t.Parallel()

	units := map[string]*model.Unit{
		"api": {
			Type: model.TypeSystem,
			LinksFrom: map[string]model.Link{
				"undefined": {Target: "undefined"},
			},
		},
	}

	index := validator.BuildIndex(units, "")
	errors := validator.ValidateReferences(index)

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	expectedMsg := `undefined unit "undefined" referenced in linkFrom from "api"`
	if errors[0].Message != expectedMsg {
		t.Errorf("expected message %q, got %q", expectedMsg, errors[0].Message)
	}
}

func TestValidateReferences_WithSuggestion(t *testing.T) {
	t.Parallel()

	units := map[string]*model.Unit{
		"api": {
			Type: model.TypeSystem,
			Links: map[string]model.Link{
				"dbb": {Target: "dbb"}, // typo for "db"
			},
		},
		"db": {
			Type: model.TypeDb,
		},
	}

	index := validator.BuildIndex(units, "")
	errors := validator.ValidateReferences(index)

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	expectedSuffix := ` (did you mean "db"?)`
	if errors[0].Message != `undefined unit "dbb" referenced from "api"`+expectedSuffix {
		t.Errorf("expected suggestion in message, got %q", errors[0].Message)
	}
}

func TestValidateReferences_NestedUnits(t *testing.T) {
	t.Parallel()

	units := map[string]*model.Unit{
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
	}

	index := validator.BuildIndex(units, "")
	errors := validator.ValidateReferences(index)

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	expectedMsg := `undefined unit "undefined" referenced from "system.api"`
	if errors[0].Message != expectedMsg {
		t.Errorf("expected message %q, got %q", expectedMsg, errors[0].Message)
	}
}

func TestValidateReferences_CollectsAllErrors(t *testing.T) {
	t.Parallel()

	units := map[string]*model.Unit{
		"api": {
			Type: model.TypeSystem,
			Links: map[string]model.Link{
				"undef1": {Target: "undef1"},
				"undef2": {Target: "undef2"},
			},
		},
	}

	index := validator.BuildIndex(units, "")
	errors := validator.ValidateReferences(index)

	if len(errors) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errors))
	}
}

func TestValidateSubunitRules_RejectsInvalidTypes(t *testing.T) {
	t.Parallel()

	invalidTypes := []struct {
		name     string
		unitType model.UnitType
	}{
		{"person", model.TypePerson},
		{"personExternal", model.TypePersonExternal},
		{"db", model.TypeDb},
		{"dbExternal", model.TypeDbExternal},
		{"queue", model.TypeQueue},
		{"queueExternal", model.TypeQueueExternal},
		{"container", model.TypeContainer},
		{"containerDb", model.TypeContainerDb},
		{"containerQueue", model.TypeContainerQueue},
		{"component", model.TypeComponent},
		{"componentDb", model.TypeComponentDb},
		{"componentQueue", model.TypeComponentQueue},
	}

	for _, tc := range invalidTypes {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			units := map[string]*model.Unit{
				"unit": {
					Type: tc.unitType,
					Subunits: map[string]*model.Unit{
						"child": {Type: model.TypeComponent},
					},
				},
			}

			index := validator.BuildIndex(units, "")
			errors := validator.ValidateSubunitRules(index)

			if len(errors) != 1 {
				t.Fatalf("expected 1 error, got %d", len(errors))
			}

			expectedMsg := `unit "unit" has type ` + string(tc.unitType) + ` which cannot have subunits`
			if errors[0].Message != expectedMsg {
				t.Errorf("expected message %q, got %q", expectedMsg, errors[0].Message)
			}
		})
	}
}

func TestValidateSubunitRules_AllowsValidTypes(t *testing.T) {
	t.Parallel()

	validTypes := []struct {
		name     string
		unitType model.UnitType
	}{
		{"system", model.TypeSystem},
		{"systemExternal", model.TypeSystemExternal},
		{"box", model.TypeBox},
	}

	for _, tc := range validTypes {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			units := map[string]*model.Unit{
				"unit": {
					Type: tc.unitType,
					Subunits: map[string]*model.Unit{
						"child": {Type: model.TypeComponent},
					},
				},
			}

			index := validator.BuildIndex(units, "")
			errors := validator.ValidateSubunitRules(index)

			if len(errors) != 0 {
				t.Errorf("expected no errors for %s with subunits, got %d: %v", tc.name, len(errors), errors)
			}
		})
	}
}

func TestValidateSubunitRules_NoSubunits(t *testing.T) {
	t.Parallel()

	units := map[string]*model.Unit{
		"person": {Type: model.TypePerson},
	}

	index := validator.BuildIndex(units, "")
	errors := validator.ValidateSubunitRules(index)

	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errors), errors)
	}
}

func TestValidateLinkRules_RejectsLinksOnParent(t *testing.T) {
	t.Parallel()

	units := map[string]*model.Unit{
		"system": {
			Type: model.TypeSystem,
			Links: map[string]model.Link{
				"other": {Target: "other"},
			},
			Subunits: map[string]*model.Unit{
				"api": {Type: model.TypeContainer},
			},
		},
		"other": {Type: model.TypeSystem},
	}

	index := validator.BuildIndex(units, "")
	errors := validator.ValidateLinkRules(index)

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	expectedMsg := `unit "system" has subunits and cannot have direct links`
	if errors[0].Message != expectedMsg {
		t.Errorf("expected message %q, got %q", expectedMsg, errors[0].Message)
	}
}

func TestValidateLinkRules_RejectsLinksFromOnParent(t *testing.T) {
	t.Parallel()

	units := map[string]*model.Unit{
		"system": {
			Type: model.TypeSystem,
			LinksFrom: map[string]model.Link{
				"other": {Target: "other"},
			},
			Subunits: map[string]*model.Unit{
				"api": {Type: model.TypeContainer},
			},
		},
		"other": {Type: model.TypeSystem},
	}

	index := validator.BuildIndex(units, "")
	errors := validator.ValidateLinkRules(index)

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	expectedMsg := `unit "system" has subunits and cannot have direct links`
	if errors[0].Message != expectedMsg {
		t.Errorf("expected message %q, got %q", expectedMsg, errors[0].Message)
	}
}

func TestValidateLinkRules_RejectsTargetWithSubunits(t *testing.T) {
	t.Parallel()

	units := map[string]*model.Unit{
		"system": {
			Type: model.TypeSystem,
			Subunits: map[string]*model.Unit{
				"api": {Type: model.TypeContainer},
			},
		},
		"other": {
			Type: model.TypeSystem,
			Links: map[string]model.Link{
				"system": {Target: "system"},
			},
		},
	}

	index := validator.BuildIndex(units, "")
	errors := validator.ValidateLinkRules(index)

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	expectedMsg := `unit "system" has subunits and cannot be linked to directly`
	if errors[0].Message != expectedMsg {
		t.Errorf("expected message %q, got %q", expectedMsg, errors[0].Message)
	}
}

func TestValidateLinkRules_AllowsValidLinks(t *testing.T) {
	t.Parallel()

	units := map[string]*model.Unit{
		"api": {
			Type: model.TypeSystem,
			Links: map[string]model.Link{
				"db": {Target: "db"},
			},
		},
		"db": {Type: model.TypeDb},
	}

	index := validator.BuildIndex(units, "")
	errors := validator.ValidateLinkRules(index)

	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errors), errors)
	}
}

func TestValidateLinkRules_CollectsAllViolations(t *testing.T) {
	t.Parallel()

	units := map[string]*model.Unit{
		"system": {
			Type: model.TypeSystem,
			Links: map[string]model.Link{
				"other": {Target: "other"},
			},
			Subunits: map[string]*model.Unit{
				"api": {Type: model.TypeContainer},
			},
		},
		"other": {
			Type: model.TypeSystem,
			Links: map[string]model.Link{
				"system": {Target: "system"},
			},
		},
	}

	index := validator.BuildIndex(units, "")
	errors := validator.ValidateLinkRules(index)

	// Should have 2 errors: system has links, system is targeted
	if len(errors) != 2 {
		t.Errorf("expected 2 errors, got %d: %v", len(errors), errors)
	}
}
