package validator_test

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/validator"
)

func TestValidateReferences(t *testing.T) {
	t.Parallel()

	t.Run("valid model returns no errors", func(t *testing.T) {
		t.Parallel()

		// Create a model with valid references
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
		errors := validator.ValidateReferences(units, index)

		if len(errors) != 0 {
			t.Errorf("expected no errors, got %d: %v", len(errors), errors)
		}
	})

	t.Run("detects undefined Links target", func(t *testing.T) {
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
		errors := validator.ValidateReferences(units, index)

		if len(errors) != 1 {
			t.Fatalf("expected 1 error, got %d", len(errors))
		}

		expectedMsg := `undefined unit "undefined" referenced from "api"`
		if errors[0].Message != expectedMsg {
			t.Errorf("expected message %q, got %q", expectedMsg, errors[0].Message)
		}
	})

	t.Run("detects undefined LinksFrom source", func(t *testing.T) {
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
		errors := validator.ValidateReferences(units, index)

		if len(errors) != 1 {
			t.Fatalf("expected 1 error, got %d", len(errors))
		}

		expectedMsg := `undefined unit "undefined" referenced in linkFrom from "api"`
		if errors[0].Message != expectedMsg {
			t.Errorf("expected message %q, got %q", expectedMsg, errors[0].Message)
		}
	})

	t.Run("includes suggestion for similar names", func(t *testing.T) {
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
		errors := validator.ValidateReferences(units, index)

		if len(errors) != 1 {
			t.Fatalf("expected 1 error, got %d", len(errors))
		}

		// Should include suggestion
		expectedSuffix := ` (did you mean "db"?)`
		if errors[0].Message != `undefined unit "dbb" referenced from "api"`+expectedSuffix {
			t.Errorf("expected suggestion in message, got %q", errors[0].Message)
		}
	})

	t.Run("handles nested units", func(t *testing.T) {
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
		errors := validator.ValidateReferences(units, index)

		if len(errors) != 1 {
			t.Fatalf("expected 1 error, got %d", len(errors))
		}

		expectedMsg := `undefined unit "undefined" referenced from "system.api"`
		if errors[0].Message != expectedMsg {
			t.Errorf("expected message %q, got %q", expectedMsg, errors[0].Message)
		}
	})

	t.Run("collects all errors", func(t *testing.T) {
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
		errors := validator.ValidateReferences(units, index)

		if len(errors) != 2 {
			t.Errorf("expected 2 errors, got %d", len(errors))
		}
	})
}

func TestValidateSubunitRules(t *testing.T) {
	t.Parallel()

	// Types that CANNOT have subunits
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
		t.Run("rejects subunits on "+tc.name, func(t *testing.T) {
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
			errors := validator.ValidateSubunitRules(units, index)

			if len(errors) != 1 {
				t.Fatalf("expected 1 error, got %d", len(errors))
			}

			expectedMsg := `unit "unit" has type ` + string(tc.unitType) + ` which cannot have subunits`
			if errors[0].Message != expectedMsg {
				t.Errorf("expected message %q, got %q", expectedMsg, errors[0].Message)
			}
		})
	}

	// Types that CAN have subunits
	validTypes := []struct {
		name     string
		unitType model.UnitType
	}{
		{"system", model.TypeSystem},
		{"systemExternal", model.TypeSystemExternal},
		{"box", model.TypeBox},
	}

	for _, tc := range validTypes {
		t.Run("allows subunits on "+tc.name, func(t *testing.T) {
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
			errors := validator.ValidateSubunitRules(units, index)

			if len(errors) != 0 {
				t.Errorf("expected no errors for %s with subunits, got %d: %v", tc.name, len(errors), errors)
			}
		})
	}

	t.Run("no error for unit without subunits", func(t *testing.T) {
		t.Parallel()

		units := map[string]*model.Unit{
			"person": {Type: model.TypePerson},
		}

		index := validator.BuildIndex(units, "")
		errors := validator.ValidateSubunitRules(units, index)

		if len(errors) != 0 {
			t.Errorf("expected no errors, got %d: %v", len(errors), errors)
		}
	})
}

func TestValidateLinkRules(t *testing.T) {
	t.Parallel()

	t.Run("rejects Links on units with subunits", func(t *testing.T) {
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
		errors := validator.ValidateLinkRules(units, index)

		if len(errors) != 1 {
			t.Fatalf("expected 1 error, got %d", len(errors))
		}

		expectedMsg := `unit "system" has subunits and cannot have direct links`
		if errors[0].Message != expectedMsg {
			t.Errorf("expected message %q, got %q", expectedMsg, errors[0].Message)
		}
	})

	t.Run("rejects LinksFrom on units with subunits", func(t *testing.T) {
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
		errors := validator.ValidateLinkRules(units, index)

		if len(errors) != 1 {
			t.Fatalf("expected 1 error, got %d", len(errors))
		}

		expectedMsg := `unit "system" has subunits and cannot have direct links`
		if errors[0].Message != expectedMsg {
			t.Errorf("expected message %q, got %q", expectedMsg, errors[0].Message)
		}
	})

	t.Run("rejects links targeting units with subunits", func(t *testing.T) {
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
		errors := validator.ValidateLinkRules(units, index)

		if len(errors) != 1 {
			t.Fatalf("expected 1 error, got %d", len(errors))
		}

		expectedMsg := `unit "system" has subunits and cannot be linked to directly`
		if errors[0].Message != expectedMsg {
			t.Errorf("expected message %q, got %q", expectedMsg, errors[0].Message)
		}
	})

	t.Run("allows links on units without subunits", func(t *testing.T) {
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
		errors := validator.ValidateLinkRules(units, index)

		if len(errors) != 0 {
			t.Errorf("expected no errors, got %d: %v", len(errors), errors)
		}
	})

	t.Run("allows links targeting units without subunits", func(t *testing.T) {
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
		errors := validator.ValidateLinkRules(units, index)

		if len(errors) != 0 {
			t.Errorf("expected no errors, got %d: %v", len(errors), errors)
		}
	})

	t.Run("collects all link rule violations", func(t *testing.T) {
		t.Parallel()

		// Parent has links AND is targeted by links
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
		errors := validator.ValidateLinkRules(units, index)

		// Should have 2 errors: system has links, system is targeted
		if len(errors) != 2 {
			t.Errorf("expected 2 errors, got %d: %v", len(errors), errors)
		}
	})
}
