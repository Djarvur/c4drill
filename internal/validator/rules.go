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
func ValidateReferences(index map[string]*UnitInfo) ValidationErrors {
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
// System-level types (system, systemExternal, box) and containers can have subunits.
// Container variants (containerDb, containerQueue) cannot have subunits.
// Returns all errors found (not fail-fast).
func ValidateSubunitRules(index map[string]*UnitInfo) ValidationErrors {
	var errors ValidationErrors

	// Types that can have subunits
	allowedTypes := map[model.UnitType]bool{
		model.TypeSystem:         true,
		model.TypeSystemExternal: true,
		model.TypeBox:            true,
		model.TypeContainer:      true,
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

// ValidateLinkRules validates link-related rules.
// Currently a no-op placeholder for future link validation rules.
// Previous restrictions on links for units with subunits have been removed
// to allow more flexible architecture diagrams.
func ValidateLinkRules(index map[string]*UnitInfo) ValidationErrors {
	return nil
}

// ValidateOrphanUnits checks that all units have connectivity.
// A unit is an orphan if it has no Links, no LinksFrom, and no Subunits.
// Returns all errors found (not fail-fast).
func ValidateOrphanUnits(index map[string]*UnitInfo) ValidationErrors {
	var errors ValidationErrors

	for path, info := range index {
		hasLinks := len(info.Unit.Links) > 0
		hasLinksFrom := len(info.Unit.LinksFrom) > 0
		hasSubunits := len(info.Unit.Subunits) > 0

		if !hasLinks && !hasLinksFrom && !hasSubunits {
			errors = append(errors, &ValidationError{
				Message: fmt.Sprintf(`unit "%s" has no incoming or outgoing links`, path),
				Path:    path,
			})
		}
	}

	return errors
}
