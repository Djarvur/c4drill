package validator

import (
	"fmt"
	"maps"
	"slices"

	"github.com/Djarvur/c4drill/internal/model"
)

// ValidateReferences checks that all referenced units exist.
// It validates both Links (target references) and LinksFrom (source references).
// Returns all errors found (not fail-fast).
func ValidateReferences(units map[string]*model.Unit, index map[string]*UnitInfo) ValidationErrors {
	var errors ValidationErrors

	// Collect all unit names for suggestions
	allNames := slices.Collect(maps.Keys(index))

	for path, info := range index {
		// Check Links references
		for target := range info.Unit.Links {
			if _, exists := index[target]; !exists {
				suggestion := FormatSuggestion(target, allNames)
				errors = append(errors, &ValidationError{
					Message: fmt.Sprintf(`undefined unit "%s" referenced from "%s"%s`, target, path, suggestion),
					Path:    path,
				})
			}
		}

		// Check LinksFrom references
		for source := range info.Unit.LinksFrom {
			if _, exists := index[source]; !exists {
				suggestion := FormatSuggestion(source, allNames)
				errors = append(errors, &ValidationError{
					Message: fmt.Sprintf(`undefined unit "%s" referenced in linkFrom from "%s"%s`, source, path, suggestion),
					Path:    path,
				})
			}
		}
	}

	return errors
}

// ValidateSubunitRules checks that only allowed types have subunits.
// Only system, systemExternal, and box types can have subunits.
// Returns all errors found (not fail-fast).
func ValidateSubunitRules(units map[string]*model.Unit, index map[string]*UnitInfo) ValidationErrors {
	var errors ValidationErrors

	// Types that can have subunits
	allowedTypes := map[model.UnitType]bool{
		model.TypeSystem:         true,
		model.TypeSystemExternal: true,
		model.TypeBox:            true,
	}

	for path, info := range index {
		if len(info.Unit.Subunits) == 0 {
			continue // No subunits, rule doesn't apply
		}

		if !allowedTypes[info.Unit.Type] {
			errors = append(errors, &ValidationError{
				Message: fmt.Sprintf(`unit "%s" has type %s which cannot have subunits`, path, info.Unit.Type),
				Path:    path,
			})
		}
	}

	return errors
}

// ValidateLinkRules checks link restrictions on units with subunits.
// - Units with subunits cannot have Links (VALD-02)
// - Units with subunits cannot have LinksFrom (VALD-02)
// - Links cannot target units with subunits (VALD-03)
// Returns all errors found (not fail-fast).
func ValidateLinkRules(units map[string]*model.Unit, index map[string]*UnitInfo) ValidationErrors {
	var errors ValidationErrors

	// First pass: identify all units with subunits
	hasSubunits := make(map[string]bool)
	for path, info := range index {
		if len(info.Unit.Subunits) > 0 {
			hasSubunits[path] = true
		}
	}

	for path, info := range index {
		// Check if this unit has subunits
		if hasSubunits[path] {
			// Units with subunits cannot have Links or LinksFrom
			if len(info.Unit.Links) > 0 || len(info.Unit.LinksFrom) > 0 {
				errors = append(errors, &ValidationError{
					Message: fmt.Sprintf(`unit "%s" has subunits and cannot have direct links`, path),
					Path:    path,
				})
			}
		}

		// Check if links target units with subunits
		for target := range info.Unit.Links {
			if hasSubunits[target] {
				errors = append(errors, &ValidationError{
					Message: fmt.Sprintf(`unit "%s" has subunits and cannot be linked to directly`, target),
					Path:    path,
				})
			}
		}
	}

	return errors
}
