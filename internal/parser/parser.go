// Package parser provides TOML parsing for C4 architecture definitions.
package parser

import (
	"os"
	"slices"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/pelletier/go-toml/v2"
)

// Default type based on nesting level:
// - C1 (root level): system
// - C2 (inside system/box): container
// - C3 (inside container): component
const (
	defaultTypeC1 = model.TypeSystem
	defaultTypeC2 = model.TypeContainer
	defaultTypeC3 = model.TypeComponent
)

// Generic types that can be auto-transformed based on nesting level:
// - db at C1 -> db, at C2 -> containerDb, at C3 -> componentDb
// - queue at C1 -> queue, at C2 -> containerQueue, at C3 -> componentQueue
var genericDbTypes = map[model.UnitType]bool{
	model.TypeDb:    true,
	model.TypeQueue: true,
}

// Model represents the root of a parsed TOML document.
// It contains the properties section and all top-level units.
type Model struct {
	// Properties contains the global [properties] section.
	Properties model.Properties `toml:"properties"`
	// Units contains all top-level units keyed by section name.
	Units map[string]*model.Unit
}

// Parse parses TOML data into a Model.
// It unmarshals the TOML content, with links automatically parsed from [[link]] arrays.
func Parse(data []byte) (*Model, error) {
	// First pass: unmarshal the entire document to a raw map
	var rawMap map[string]any

	if err := toml.Unmarshal(data, &rawMap); err != nil {
		return nil, wrapDecodeError(err)
	}

	// Second pass: build the Model struct
	m := &Model{
		Units: make(map[string]*model.Unit),
	}

	// Extract properties if present
	if props, ok := rawMap["properties"]; ok {
		propsData, err := toml.Marshal(props)
		if err != nil {
			return nil, &ParseError{Message: "failed to marshal properties", Cause: err}
		}

		if err := toml.Unmarshal(propsData, &m.Properties); err != nil {
			return nil, wrapDecodeError(err)
		}
	}

	// Process all other sections as units (C1 level, no parent type)
	for name, value := range rawMap {
		if name == "properties" {
			continue
		}

		unit, err := parseUnit(name, value, "")
		if err != nil {
			return nil, err
		}

		m.Units[name] = unit
	}

	return m, nil
}

// parseUnit parses a raw map value into a Unit struct, including nested subunits.
// parentType is the type of the parent unit, used to determine default type.
func parseUnit(name string, value any, parentType model.UnitType) (*model.Unit, error) {
	unitMap, ok := value.(map[string]any)
	if !ok {
		return nil, &ParseError{
			Message: "invalid unit format",
			Context: name,
		}
	}

	// Re-marshal and unmarshal to get proper type conversion
	unitData, err := toml.Marshal(unitMap)
	if err != nil {
		return nil, &ParseError{Message: "failed to marshal unit", Context: name, Cause: err}
	}

	var unit model.Unit

	if err := toml.Unmarshal(unitData, &unit); err != nil {
		return nil, wrapDecodeError(err)
	}

	// Apply default type if not specified
	if unit.Type == "" {
		unit.Type = defaultTypeForParent(parentType)
	}

	// Infer level-specific type for generic types (db, queue)
	unit.Type = inferGenericType(unit.Type, parentType)

	// Handle nested subunits (sections like [parent.child])
	for key, val := range unitMap {
		// Skip fields that are already in the Unit struct
		if isBuiltinField(key) {
			continue
		}

		// This must be a subunit
		subunit, err := parseUnit(key, val, unit.Type)
		if err != nil {
			return nil, err
		}

		if unit.Subunits == nil {
			unit.Subunits = make(map[string]*model.Unit)
		}

		unit.Subunits[key] = subunit
	}

	return &unit, nil
}

// defaultTypeForParent returns the default unit type based on parent type.
// - No parent (C1 level): system
// - Parent is system/systemExternal/box (C2 level): container
// - Parent is container (C3 level): component
func defaultTypeForParent(parentType model.UnitType) model.UnitType {
	switch parentType {
	case "": // No parent = C1 level (root)
		return defaultTypeC1
	case model.TypeSystem,
		model.TypeSystemExternal,
		model.TypeBox:
		return defaultTypeC2 // C2 level default
	case model.TypeContainer:
		return defaultTypeC3 // C3 level default
	default:
		// For other parent types (db, queue, etc.), default to C2
		return defaultTypeC2
	}
}

// inferGenericType transforms generic types (db, queue) to level-specific types
// based on the nesting level determined by parent type.
// - C1 (no parent): db -> db, queue -> queue
// - C2 (inside system/systemExternal/box): db -> containerDb, queue -> containerQueue
// - C3 (inside container): db -> componentDb, queue -> componentQueue
func inferGenericType(unitType model.UnitType, parentType model.UnitType) model.UnitType {
	if !genericDbTypes[unitType] {
		return unitType // Not a generic type, return as-is
	}

	// Determine the nesting level and transform type accordingly
	switch parentType {
	case "": // C1 level (root)
		return unitType // db stays db, queue stays queue
	case model.TypeSystem,
		model.TypeSystemExternal,
		model.TypeBox: // C2 level
		switch unitType {
		case model.TypeDb:
			return model.TypeContainerDb
		case model.TypeQueue:
			return model.TypeContainerQueue
		}
	case model.TypeContainer: // C3 level
		switch unitType {
		case model.TypeDb:
			return model.TypeComponentDb
		case model.TypeQueue:
			return model.TypeComponentQueue
		}
	}

	// For other parent types, keep as-is (shouldn't happen in valid nesting)
	return unitType
}

// isBuiltinField returns true if the key is a known Unit struct field.
func isBuiltinField(key string) bool {
	return slices.Contains([]string{
		"type", "name", "description", "technology",
		"color", "style", "border", "edges",
		"width", "height", "expanded",
		"link", "linkFrom",
	}, key)
}

// ParseFile reads a TOML file and parses it into a Model.
// It returns an error if the file cannot be read or parsed.
//
//nolint:gosec // G304: Path is provided by caller, this is intentional for CLI tool
func ParseFile(path string) (*Model, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &ParseError{
			Message: "failed to read file",
			Context: path,
			Cause:   err,
		}
	}

	return Parse(data)
}
