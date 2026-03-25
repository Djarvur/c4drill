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
		for _, link := range info.Unit.Links {
			if _, exists := index[link.Peer]; !exists {
				suggestion := FormatSuggestion(link.Peer, allNames)
				errors = append(errors, &ValidationError{
					Message: fmt.Sprintf(`undefined unit "%s" referenced from "%s"%s`, link.Peer, path, suggestion),
					Path:    path,
				})
			}
		}

		// Check LinksFrom references
		for _, link := range info.Unit.LinksFrom {
			if _, exists := index[link.Peer]; !exists {
				suggestion := FormatSuggestion(link.Peer, allNames)
				errors = append(errors, &ValidationError{
					Message: fmt.Sprintf(`undefined unit "%s" referenced in linkFrom from "%s"%s`, link.Peer, path, suggestion),
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
		model.TypeSystem:        true,
		model.TypeBox:           true, // C1 box can contain C1 types
		model.TypeContainer:     true,
		model.TypeContainerBox:  true, // C2 box can contain C2 types
		model.TypeComponentBox:  true, // C3 box can contain C3 types
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
func ValidateLinkRules(index map[string]*UnitInfo) ValidationErrors {
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
		for _, link := range info.Unit.Links {
			if hasSubunits[link.Peer] {
				errors = append(errors, &ValidationError{
					Message: fmt.Sprintf(`unit "%s" has subunits and cannot be linked to directly`, link.Peer),
					Path:    path,
				})
			}
		}
	}

	return errors
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

// C1 types - top-level context types.
//
//nolint:gochecknoglobals // Lookup map for O(1) type checking, immutable after init
var c1Types = map[model.UnitType]bool{
	model.TypePerson:         true,
	model.TypePersonExternal: true,
	model.TypeSystem:         true,
	model.TypeSystemExternal: true,
	model.TypeDb:             true,
	model.TypeDbExternal:     true,
	model.TypeQueue:          true,
	model.TypeQueueExternal:  true,
	model.TypeBox:            true,
}

// C2 types - container-level types (inside system/box).
//
//nolint:gochecknoglobals // Lookup map for O(1) type checking, immutable after init
var c2Types = map[model.UnitType]bool{
	model.TypeContainer:      true,
	model.TypeContainerDb:    true,
	model.TypeContainerQueue: true,
	model.TypeContainerBox:   true,
}

// C3 types - component-level types (inside container).
//
//nolint:gochecknoglobals // Lookup map for O(1) type checking, immutable after init
var c3Types = map[model.UnitType]bool{
	model.TypeComponent:      true,
	model.TypeComponentDb:    true,
	model.TypeComponentQueue: true,
	model.TypeComponentBox:   true,
}

// ValidateNestingHierarchy checks that units are placed at the correct C4 level.
// - C1 types (person, system, db, queue, box + external variants) are top-level only
// - Inside system: C2 types (container variants)
// - Inside box (C1): C1 types only (same-level grouping)
// - Inside container: C3 types (component variants)
// - Inside containerBox (C2): C2 types only (same-level grouping)
// - Inside componentBox (C3): C3 types only (same-level grouping)
// Returns all errors found (not fail-fast).
func ValidateNestingHierarchy(index map[string]*UnitInfo) ValidationErrors {
	errors := make(ValidationErrors, 0, len(index))

	for path, info := range index {
		errors = append(errors, validateUnitNesting(path, info, index)...)
	}

	return errors
}

// validateUnitNesting checks that a single unit is at the correct C4 level.
func validateUnitNesting(path string, info *UnitInfo, index map[string]*UnitInfo) ValidationErrors {
	unitType := info.Unit.Type

	// Top-level units must be C1 types
	if info.Parent == "" {
		if !c1Types[unitType] {
			return ValidationErrors{&ValidationError{
				Message: fmt.Sprintf(`unit "%s" has type %s which is not allowed at top level (C1 types only)`, path, unitType),
				Path:    path,
			}}
		}

		return nil
	}

	// Nested units: check parent type to determine allowed child types
	parentInfo, parentExists := index[info.Parent]
	if !parentExists {
		// Orphan references are handled by ValidateReferences
		return nil
	}

	parentType := parentInfo.Unit.Type

	// If parent is system, children must be C2
	if parentType == model.TypeSystem {
		if !c2Types[unitType] {
			return ValidationErrors{&ValidationError{
				Message: fmt.Sprintf(`unit "%s" has type %s which must be C2 type (inside system)`,
					path, unitType),
				Path:    path,
			}}
		}

		return nil
	}

	// If parent is C1 box, children must be C1 (same-level grouping)
	if parentType == model.TypeBox {
		if !c1Types[unitType] {
			return ValidationErrors{&ValidationError{
				Message: fmt.Sprintf(`unit "%s" has type %s which must be C1 type (inside C1 box)`,
					path, unitType),
				Path:    path,
			}}
		}

		return nil
	}

	// If parent is container, children must be C3
	if parentType == model.TypeContainer {
		if !c3Types[unitType] {
			return ValidationErrors{&ValidationError{
				Message: fmt.Sprintf(`unit "%s" has type %s which must be C3 type (inside container)`,
					path, unitType),
				Path:    path,
			}}
		}

		return nil
	}

	// If parent is containerBox (C2), children must be C2 (same-level grouping)
	if parentType == model.TypeContainerBox {
		if !c2Types[unitType] {
			return ValidationErrors{&ValidationError{
				Message: fmt.Sprintf(`unit "%s" has type %s which must be C2 type (inside containerBox)`,
					path, unitType),
				Path:    path,
			}}
		}

		return nil
	}

	// If parent is componentBox (C3), children must be C3 (same-level grouping)
	if parentType == model.TypeComponentBox {
		if !c3Types[unitType] {
			return ValidationErrors{&ValidationError{
				Message: fmt.Sprintf(`unit "%s" has type %s which must be C3 type (inside componentBox)`,
					path, unitType),
				Path:    path,
			}}
		}

		return nil
	}

	// Other parent types (person, systemExternal, db, queue, C2/C3 leaf variants) cannot have children
	// This is handled by ValidateSubunitRules, so we skip here
	return nil
}

// externalTypes maps all C1 external types for O(1) lookup.
//
//nolint:gochecknoglobals // Lookup map for O(1) type checking, immutable after init
var externalTypes = map[model.UnitType]bool{
	model.TypePersonExternal:  true,
	model.TypeSystemExternal:  true,
	model.TypeDbExternal:      true,
	model.TypeQueueExternal:   true,
}

// ValidateBoxMixedContents checks that C1 boxes contain only external or only non-external units.
// Mixing external and non-external units in the same C1 box is not allowed.
// ContainerBox (C2) and ComponentBox (C3) are not validated by this rule.
// Returns all errors found (not fail-fast).
func ValidateBoxMixedContents(index map[string]*UnitInfo) ValidationErrors {
	var errors ValidationErrors

	for path, info := range index {
		// Only validate C1 TypeBox units
		if info.Unit.Type != model.TypeBox {
			continue
		}

		// Skip boxes without subunits
		if len(info.Unit.Subunits) == 0 {
			continue
		}

		// Check if box has both external and non-external subunits
		hasExternal := false
		hasNonExternal := false

		for _, subunit := range info.Unit.Subunits {
			if externalTypes[subunit.Type] {
				hasExternal = true
			} else if c1Types[subunit.Type] {
				// Only count C1 non-external types (person, system, db, queue)
				hasNonExternal = true
			}
		}

		// If both external and non-external found, it's an error
		if hasExternal && hasNonExternal {
			errors = append(errors, &ValidationError{
				Message: fmt.Sprintf(`box "%s" cannot contain both external and non-external units`, path),
				Path:    path,
			})
		}
	}

	return errors
}
