// Package parser provides TOML parsing for C4 architecture definitions.
package parser

import (
	"os"
	"slices"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/pelletier/go-toml/v2"
	"github.com/pelletier/go-toml/v2/unstable"
)

// Default type based on nesting level:
// - C1 (root level): system
// - C2 (inside system/box): container
// - C3 (inside container): component.
const (
	defaultTypeC1 = model.TypeSystem
	defaultTypeC2 = model.TypeContainer
	defaultTypeC3 = model.TypeComponent
)

// Generic types that can be auto-transformed based on nesting level:
// - db at C1 -> db, at C2 -> containerDb, at C3 -> componentDb
// - queue at C1 -> queue, at C2 -> containerQueue, at C3 -> componentQueue.
//
//nolint:gochecknoglobals // Lookup map for O(1) type checking, immutable after init
var genericDbTypes = map[model.UnitType]bool{
	model.TypeDb:    true,
	model.TypeQueue: true,
}

// Model represents the root of a parsed TOML document.
// It contains the properties section and all top-level units.
type Model struct {
	// Properties contains the global [properties] section.
	Properties model.Properties `toml:"properties"`
	// UnitOrder tracks the definition order of unit names.
	UnitOrder []string
	// Units contains all top-level units keyed by section name.
	Units map[string]*model.Unit
}

// Parse parses TOML data into a Model.
// It unmarshals the TOML content, with links automatically parsed from [[link]] arrays.
// Definition order of units and subunits is preserved.
func Parse(data []byte) (*Model, error) {
	// First pass: capture definition order using unstable API
	unitOrder, subunitOrders, err := captureDefinitionOrder(data)
	if err != nil {
		return nil, wrapDecodeError(err)
	}

	// Second pass: unmarshal the entire document to a raw map
	var rawMap map[string]any

	if err := toml.Unmarshal(data, &rawMap); err != nil {
		return nil, wrapDecodeError(err)
	}

	// Third pass: build the Model struct
	m := &Model{
		UnitOrder: unitOrder,
		Units:     make(map[string]*model.Unit),
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

	// Process units in the captured order (not rawMap iteration order)
	for _, name := range unitOrder {
		value, ok := rawMap[name]
		if !ok {
			continue // Should not happen if captureDefinitionOrder is correct
		}

		subunitOrder := subunitOrders[name]
		unit, err := parseUnitWithOrder(name, value, "", subunitOrder, subunitOrders)
		if err != nil {
			return nil, err
		}

		m.Units[name] = unit
	}

	return m, nil
}

// captureDefinitionOrder uses the unstable API to capture the order of units and subunits.
// Returns: unitOrder (top-level), subunitOrders (nested), error.
func captureDefinitionOrder(data []byte) ([]string, map[string][]string, error) {
	unitOrder := make([]string, 0)
	subunitOrders := make(map[string][]string)

	seenUnits := make(map[string]bool)
	seenSubunits := make(map[string]bool)

	p := unstable.Parser{}
	p.Reset(data)

	for p.NextExpression() {
		expr := p.Expression()
		if expr.Kind != unstable.Table {
			continue // Skip non-table expressions
		}

		// Extract the table key parts
		keyIter := expr.Key()
		var parts []string
		for keyIter.Next() {
			parts = append(parts, string(keyIter.Node().Data))
		}

		if len(parts) == 0 {
			continue
		}

		// Skip [properties] section
		if len(parts) == 1 && parts[0] == "properties" {
			continue
		}

		if len(parts) == 1 {
			// Top-level unit [name]
			name := parts[0]
			if !seenUnits[name] {
				unitOrder = append(unitOrder, name)
				seenUnits[name] = true
			}
		} else if len(parts) == 2 {
			// Subunit [parent.child]
			parent := parts[0]
			child := parts[1]
			key := parent + "." + child
			if !seenSubunits[key] {
				subunitOrders[parent] = append(subunitOrders[parent], child)
				seenSubunits[key] = true
			}
		}
		// Ignore deeper nesting (len(parts) > 2) - not supported
	}

	if err := p.Error(); err != nil {
		return nil, nil, err
	}

	return unitOrder, subunitOrders, nil
}

// parseUnitWithOrder parses a unit with explicit subunit order.
func parseUnitWithOrder(
	name string,
	value any,
	parentType model.UnitType,
	subunitOrder []string,
	subunitOrders map[string][]string,
) (*model.Unit, error) {
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

	// Process subunits in the provided order
	if len(subunitOrder) > 0 {
		unit.Subunits = make(map[string]*model.Unit)
		unit.SubunitOrder = subunitOrder

		for _, subName := range subunitOrder {
			subVal, ok := unitMap[subName]
			if !ok {
				continue
			}

			// Get the subunit's own subunit order
			fullPath := name + "." + subName
			subSubunitOrder := subunitOrders[fullPath]

			subunit, err := parseUnitWithOrder(subName, subVal, unit.Type, subSubunitOrder, subunitOrders)
			if err != nil {
				return nil, err
			}

			unit.Subunits[subName] = subunit
		}
	} else {
		// Fallback: process subunits from raw map (no order guarantee, but maintains backward compatibility)
		for key, val := range unitMap {
			// Skip fields that are already in the Unit struct
			if isBuiltinField(key) {
				continue
			}

			// This must be a subunit
			subunit, err := parseUnitWithOrder(key, val, unit.Type, nil, subunitOrders)
			if err != nil {
				return nil, err
			}

			if unit.Subunits == nil {
				unit.Subunits = make(map[string]*model.Unit)
			}

			unit.Subunits[key] = subunit
			unit.SubunitOrder = append(unit.SubunitOrder, key)
		}
	}

	return &unit, nil
}

// defaultTypeForParent returns the default unit type based on parent type.
// - No parent (C1 level): system
// - Parent is system (C2 level): container
// - Parent is box (C1 level): system (C1 same-level grouping)
// - Parent is container (C3 level): component
// - Parent is containerBox (C2 level): container (C2 same-level grouping)
// - Parent is componentBox (C3 level): component (C3 same-level grouping).
func defaultTypeForParent(parentType model.UnitType) model.UnitType {
	//nolint:exhaustive // Default case handles all remaining types
	switch parentType {
	case "": // No parent = C1 level (root)
		return defaultTypeC1
	case model.TypeSystem:
		return defaultTypeC2 // C2 level default
	case model.TypeBox:
		return defaultTypeC1 // C1 same-level grouping
	case model.TypeContainer:
		return defaultTypeC3 // C3 level default
	case model.TypeContainerBox:
		return defaultTypeC2 // C2 same-level grouping
	case model.TypeComponentBox:
		return defaultTypeC3 // C3 same-level grouping
	default:
		// For other parent types (db, queue, etc.), default to C1
		return defaultTypeC1
	}
}

// inferGenericType transforms generic types (db, queue) to level-specific types
// based on the nesting level determined by parent type.
// - C1 (no parent or inside C1 box): db -> db, queue -> queue
// - C2 (inside system or containerBox): db -> containerDb, queue -> containerQueue
// - C3 (inside container or componentBox): db -> componentDb, queue -> componentQueue.
func inferGenericType(unitType model.UnitType, parentType model.UnitType) model.UnitType {
	if !genericDbTypes[unitType] {
		return unitType // Not a generic type, return as-is
	}

	// Determine the nesting level and transform type accordingly
	//nolint:exhaustive // Default case handles all remaining types
	switch parentType {
	case "", model.TypeBox: // C1 level (root or C1 box)
		return unitType // db stays db, queue stays queue
	case model.TypeSystem, model.TypeContainerBox: // C2 level
		//nolint:exhaustive // Only db/queue are generic types, handled explicitly
		switch unitType {
		case model.TypeDb:
			return model.TypeContainerDb
		case model.TypeQueue:
			return model.TypeContainerQueue
		}
	case model.TypeContainer, model.TypeComponentBox: // C3 level
		//nolint:exhaustive // Only db/queue are generic types, handled explicitly
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
