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
			Links: []model.Link{
								{Peer: "db"},
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
			Links: []model.Link{
								{Peer: "undefined"},
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
			LinksFrom: []model.Link{
								{Peer: "undefined"},
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
			Links: []model.Link{
								{Peer: "dbb"}, // typo for "db"
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
					Links: []model.Link{
										{Peer: "undefined"},
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
			Links: []model.Link{
								{Peer: "undef1"},
								{Peer: "undef2"},
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
		{"systemExternal", model.TypeSystemExternal},
		{"db", model.TypeDb},
		{"dbExternal", model.TypeDbExternal},
		{"queue", model.TypeQueue},
		{"queueExternal", model.TypeQueueExternal},
		{"containerDb", model.TypeContainerDb},
		{"containerQueue", model.TypeContainerQueue},
		{"component", model.TypeComponent},
		{"componentDb", model.TypeComponentDb},
		{"componentQueue", model.TypeComponentQueue},
		{"componentBox", model.TypeComponentBox},
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

	// Types that can have subunits: system, box, container, containerBox
	validTypes := []struct {
		name     string
		unitType model.UnitType
	}{
		{"system", model.TypeSystem},
		{"box", model.TypeBox},
		{"container", model.TypeContainer},
		{"containerBox", model.TypeContainerBox},
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
			Links: []model.Link{
				{Peer: "other"},
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
			LinksFrom: []model.Link{
				{Peer: "other"},
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
			Links: []model.Link{
				{Peer: "system"},
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
			Links: []model.Link{
								{Peer: "db"},
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
			Links: []model.Link{
				{Peer: "other"},
			},
			Subunits: map[string]*model.Unit{
				"api": {Type: model.TypeContainer},
			},
		},
		"other": {
			Type: model.TypeSystem,
			Links: []model.Link{
				{Peer: "system"},
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

func TestValidateOrphanUnits_NoOrphans(t *testing.T) {
	t.Parallel()

	units := map[string]*model.Unit{
		"api": {
			Type: model.TypeSystem,
			Links: []model.Link{
								{Peer: "db"},
			},
		},
		"db": {
			Type: model.TypeDb,
			LinksFrom: []model.Link{
				{Peer: "api"},
			},
		},
	}

	index := validator.BuildIndex(units, "")
	errors := validator.ValidateOrphanUnits(index)

	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errors), errors)
	}
}

func TestValidateOrphanUnits_SingleOrphan(t *testing.T) {
	t.Parallel()

	units := map[string]*model.Unit{
		"orphan": {Type: model.TypeSystem}, // No Links, LinksFrom, or Subunits
		"connected": {
			Type: model.TypeSystem,
			Links: []model.Link{
				{Peer: "other"},
			},
		},
		"other": {
			Type: model.TypeSystem,
			LinksFrom: []model.Link{
				{Peer: "connected"},
			},
		},
	}

	index := validator.BuildIndex(units, "")
	errors := validator.ValidateOrphanUnits(index)

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	expectedMsg := `unit "orphan" has no incoming or outgoing links`
	if errors[0].Message != expectedMsg {
		t.Errorf("expected message %q, got %q", expectedMsg, errors[0].Message)
	}
}

func TestValidateOrphanUnits_MultipleOrphans(t *testing.T) {
	t.Parallel()

	units := map[string]*model.Unit{
		"orphan1": {Type: model.TypeSystem},
		"orphan2": {Type: model.TypeDb},
		"connected": {
			Type: model.TypeSystem,
			Links: []model.Link{
				{Peer: "other"},
			},
		},
		"other": {
			Type: model.TypeSystem,
			LinksFrom: []model.Link{
				{Peer: "connected"},
			},
		},
	}

	index := validator.BuildIndex(units, "")
	errors := validator.ValidateOrphanUnits(index)

	if len(errors) != 2 {
		t.Errorf("expected 2 errors, got %d: %v", len(errors), errors)
	}
}

func TestValidateOrphanUnits_UnitWithSubunits(t *testing.T) {
	t.Parallel()

	units := map[string]*model.Unit{
		"system": {
			Type: model.TypeSystem,
			Subunits: map[string]*model.Unit{
				"api": {
					Type: model.TypeContainer,
					Links: []model.Link{
				{Peer: "db"},
					},
				},
			},
		},
		"db": {
			Type: model.TypeDb,
			LinksFrom: []model.Link{
				{Peer: "system.api"},
			},
		},
	}

	index := validator.BuildIndex(units, "")
	errors := validator.ValidateOrphanUnits(index)

	// System has subunits, so it's not an orphan
	// api has links, so it's not an orphan
	// db has LinksFrom, so it's not an orphan
	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errors), errors)
	}
}

func TestValidateOrphanUnits_UnitWithLinksFrom(t *testing.T) {
	t.Parallel()

	units := map[string]*model.Unit{
		"api": {
			Type: model.TypeSystem,
			Links: []model.Link{
								{Peer: "db"},
			},
		},
		"db": {
			Type: model.TypeDb,
			LinksFrom: []model.Link{
				{Peer: "api"},
			},
		},
	}

	index := validator.BuildIndex(units, "")
	errors := validator.ValidateOrphanUnits(index)

	// db has LinksFrom, so it's not an orphan
	if len(errors) != 0 {
		t.Errorf("expected no errors, got %d: %v", len(errors), errors)
	}
}

func TestValidateOrphanUnits_NestedOrphan(t *testing.T) {
	t.Parallel()

	units := map[string]*model.Unit{
		"system": {
			Type: model.TypeSystem,
			Subunits: map[string]*model.Unit{
				"api": {
					Type: model.TypeContainer,
					Links: []model.Link{
				{Peer: "db"},
					},
				},
				"orphan": {Type: model.TypeContainer}, // No links
			},
		},
		"db": {
			Type: model.TypeDb,
			LinksFrom: []model.Link{
				{Peer: "system.api"},
			},
		},
	}

	index := validator.BuildIndex(units, "")
	errors := validator.ValidateOrphanUnits(index)

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	expectedMsg := `unit "system.orphan" has no incoming or outgoing links`
	if errors[0].Message != expectedMsg {
		t.Errorf("expected message %q, got %q", expectedMsg, errors[0].Message)
	}
}

// TestValidateNestingHierarchy tests the C4 nesting hierarchy validation.
// C1 types (person, system, db, queue, box + external variants) are top-level only.
// C2 types (container, containerDb, containerQueue) belong inside system/box.
// C3 types (component, componentDb, componentQueue) belong inside container.

func TestValidateNestingHierarchy_RejectsC2AtTopLevel(t *testing.T) {
	t.Parallel()

	c2Types := []struct {
		name     string
		unitType model.UnitType
	}{
		{"container", model.TypeContainer},
		{"containerDb", model.TypeContainerDb},
		{"containerQueue", model.TypeContainerQueue},
	}

	for _, tc := range c2Types {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			units := map[string]*model.Unit{
				"unit": {
					Type: tc.unitType,
					Links: []model.Link{
						{Peer: "other"},
					},
				},
				"other": {Type: model.TypeSystem},
			}

			index := validator.BuildIndex(units, "")
			errors := validator.ValidateNestingHierarchy(index)

			if len(errors) != 1 {
				t.Fatalf("expected 1 error, got %d", len(errors))
			}

			expectedMsg := `unit "unit" has type ` + string(tc.unitType) + ` which is not allowed at top level (C1 types only)`
			if errors[0].Message != expectedMsg {
				t.Errorf("expected message %q, got %q", expectedMsg, errors[0].Message)
			}
		})
	}
}

func TestValidateNestingHierarchy_RejectsC3AtTopLevel(t *testing.T) {
	t.Parallel()

	c3Types := []struct {
		name     string
		unitType model.UnitType
	}{
		{"component", model.TypeComponent},
		{"componentDb", model.TypeComponentDb},
		{"componentQueue", model.TypeComponentQueue},
		{"componentBox", model.TypeComponentBox},
	}

	for _, tc := range c3Types {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			units := map[string]*model.Unit{
				"unit": {
					Type: tc.unitType,
					Links: []model.Link{
						{Peer: "other"},
					},
				},
				"other": {Type: model.TypeSystem},
			}

			index := validator.BuildIndex(units, "")
			errors := validator.ValidateNestingHierarchy(index)

			if len(errors) != 1 {
				t.Fatalf("expected 1 error, got %d", len(errors))
			}

			expectedMsg := `unit "unit" has type ` + string(tc.unitType) + ` which is not allowed at top level (C1 types only)`
			if errors[0].Message != expectedMsg {
				t.Errorf("expected message %q, got %q", expectedMsg, errors[0].Message)
			}
		})
	}
}

func TestValidateNestingHierarchy_AllowsC1AtTopLevel(t *testing.T) {
	t.Parallel()

	c1Types := []struct {
		name     string
		unitType model.UnitType
	}{
		{"person", model.TypePerson},
		{"personExternal", model.TypePersonExternal},
		{"system", model.TypeSystem},
		{"systemExternal", model.TypeSystemExternal},
		{"db", model.TypeDb},
		{"dbExternal", model.TypeDbExternal},
		{"queue", model.TypeQueue},
		{"queueExternal", model.TypeQueueExternal},
		{"box", model.TypeBox},
	}

	for _, tc := range c1Types {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			units := map[string]*model.Unit{
				"unit": {
					Type: tc.unitType,
					Links: []model.Link{
						{Peer: "other"},
					},
				},
				"other": {Type: model.TypeSystem},
			}

			index := validator.BuildIndex(units, "")
			errors := validator.ValidateNestingHierarchy(index)

			if len(errors) != 0 {
				t.Errorf("expected no errors for C1 type %s at top level, got %d: %v", tc.name, len(errors), errors)
			}
		})
	}
}

func TestValidateNestingHierarchy_RejectsC3InSystem(t *testing.T) {
	t.Parallel()

	c3Types := []struct {
		name     string
		unitType model.UnitType
	}{
		{"component", model.TypeComponent},
		{"componentDb", model.TypeComponentDb},
		{"componentQueue", model.TypeComponentQueue},
		{"componentBox", model.TypeComponentBox},
	}

	for _, tc := range c3Types {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			units := map[string]*model.Unit{
				"system": {
					Type: model.TypeSystem,
					Subunits: map[string]*model.Unit{
						"child": {
							Type: tc.unitType,
							Links: []model.Link{
								{Peer: "other"},
							},
						},
					},
				},
				"other": {Type: model.TypeDb},
			}

			index := validator.BuildIndex(units, "")
			errors := validator.ValidateNestingHierarchy(index)

			if len(errors) != 1 {
				t.Fatalf("expected 1 error, got %d", len(errors))
			}

			expectedMsg := `unit "system.child" has type ` + string(tc.unitType) +
				` which must be C2 type (inside system)`
			if errors[0].Message != expectedMsg {
				t.Errorf("expected message %q, got %q", expectedMsg, errors[0].Message)
			}
		})
	}
}

func TestValidateNestingHierarchy_AllowsC2InSystem(t *testing.T) {
	t.Parallel()

	c2Types := []struct {
		name     string
		unitType model.UnitType
	}{
		{"container", model.TypeContainer},
		{"containerDb", model.TypeContainerDb},
		{"containerQueue", model.TypeContainerQueue},
		{"containerBox", model.TypeContainerBox},
	}

	// Only system can contain C2 types
	for _, child := range c2Types {
		t.Run("system_"+child.name, func(t *testing.T) {
			t.Parallel()

			units := map[string]*model.Unit{
				"system": {
					Type: model.TypeSystem,
					Subunits: map[string]*model.Unit{
						"child": {
							Type: child.unitType,
							Links: []model.Link{
								{Peer: "other"},
							},
						},
					},
				},
				"other": {Type: model.TypeDb},
			}

			index := validator.BuildIndex(units, "")
			errors := validator.ValidateNestingHierarchy(index)

			if len(errors) != 0 {
				t.Errorf("expected no errors for C2 type %s inside system, got %d: %v", child.name, len(errors), errors)
			}
		})
	}
}

func TestValidateNestingHierarchy_RejectsC2InContainer(t *testing.T) {
	t.Parallel()

	c2Types := []struct {
		name     string
		unitType model.UnitType
	}{
		{"container", model.TypeContainer},
		{"containerDb", model.TypeContainerDb},
		{"containerQueue", model.TypeContainerQueue},
		{"containerBox", model.TypeContainerBox},
	}

	for _, tc := range c2Types {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			units := map[string]*model.Unit{
				"system": {
					Type: model.TypeSystem,
					Subunits: map[string]*model.Unit{
						"api": {
							Type: model.TypeContainer,
							Subunits: map[string]*model.Unit{
								"child": {
									Type: tc.unitType,
									Links: []model.Link{
										{Peer: "other"},
									},
								},
							},
						},
					},
				},
				"other": {Type: model.TypeDb},
			}

			index := validator.BuildIndex(units, "")
			errors := validator.ValidateNestingHierarchy(index)

			if len(errors) != 1 {
				t.Fatalf("expected 1 error, got %d", len(errors))
			}

			expectedMsg := `unit "system.api.child" has type ` + string(tc.unitType) +
				` which must be C3 type (inside container)`
			if errors[0].Message != expectedMsg {
				t.Errorf("expected message %q, got %q", expectedMsg, errors[0].Message)
			}
		})
	}
}

func TestValidateNestingHierarchy_AllowsC3InContainer(t *testing.T) {
	t.Parallel()

	c3Types := []struct {
		name     string
		unitType model.UnitType
	}{
		{"component", model.TypeComponent},
		{"componentDb", model.TypeComponentDb},
		{"componentQueue", model.TypeComponentQueue},
		{"componentBox", model.TypeComponentBox},
	}

	for _, tc := range c3Types {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			units := map[string]*model.Unit{
				"system": {
					Type: model.TypeSystem,
					Subunits: map[string]*model.Unit{
						"api": {
							Type: model.TypeContainer,
							Subunits: map[string]*model.Unit{
								"child": {
									Type: tc.unitType,
									Links: []model.Link{
										{Peer: "other"},
									},
								},
							},
						},
					},
				},
				"other": {Type: model.TypeDb},
			}

			index := validator.BuildIndex(units, "")
			errors := validator.ValidateNestingHierarchy(index)

			if len(errors) != 0 {
				t.Errorf("expected no errors for C3 type %s inside container, got %d: %v", tc.name, len(errors), errors)
			}
		})
	}
}

func TestValidateNestingHierarchy_ValidChain(t *testing.T) {
	t.Parallel()

	// Valid nesting: system -> container -> component
	units := map[string]*model.Unit{
		"system": {
			Type: model.TypeSystem,
			Subunits: map[string]*model.Unit{
				"api": {
					Type: model.TypeContainer,
					Subunits: map[string]*model.Unit{
						"handler": {
							Type: model.TypeComponent,
							Links: []model.Link{
								{Peer: "db"},
							},
						},
					},
				},
			},
		},
		"db": {Type: model.TypeDb},
	}

	index := validator.BuildIndex(units, "")
	errors := validator.ValidateNestingHierarchy(index)

	if len(errors) != 0 {
		t.Errorf("expected no errors for valid nesting chain, got %d: %v", len(errors), errors)
	}
}

func TestValidateNestingHierarchy_CollectsAllErrors(t *testing.T) {
	t.Parallel()

	// Multiple violations: C2 at top level, C3 at top level, C3 in system
	units := map[string]*model.Unit{
		"container": {
			Type: model.TypeContainer,
			Links: []model.Link{
				{Peer: "db"},
			},
		},
		"component": {
			Type: model.TypeComponent,
			Links: []model.Link{
				{Peer: "db"},
			},
		},
		"system": {
			Type: model.TypeSystem,
			Subunits: map[string]*model.Unit{
				"comp": {
					Type: model.TypeComponent,
					Links: []model.Link{
						{Peer: "db"},
					},
				},
			},
		},
		"db": {Type: model.TypeDb},
	}

	index := validator.BuildIndex(units, "")
	errors := validator.ValidateNestingHierarchy(index)

	if len(errors) != 3 {
		t.Errorf("expected 3 errors for multiple violations, got %d: %v", len(errors), errors)
	}
}
