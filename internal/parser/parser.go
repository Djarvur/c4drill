// Package parser provides TOML parsing for C4 architecture definitions.
package parser

import (
	"os"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/pelletier/go-toml/v2"
)

// Model represents the root of a parsed TOML document.
// It contains the properties section and all top-level units.
type Model struct {
	// Properties contains the global [properties] section.
	Properties model.Properties `toml:"properties"`
	// Units contains all top-level units keyed by section name.
	Units map[string]*model.Unit
}

// Parse parses TOML data into a Model.
// It unmarshals the TOML content and populates Link.Target fields from map keys.
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

	// Process all other sections as units
	for name, value := range rawMap {
		if name == "properties" {
			continue
		}

		unit, err := parseUnit(name, value)
		if err != nil {
			return nil, err
		}

		m.Units[name] = unit
	}

	// Populate Link.Target from map keys for all units
	if m.Units != nil {
		populateLinkTargets(m.Units)
	}

	return m, nil
}

// parseUnit parses a raw map value into a Unit struct, including nested subunits.
func parseUnit(name string, value any) (*model.Unit, error) {
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

	// Handle nested subunits (sections like [parent.child])
	for key, val := range unitMap {
		// Skip fields that are already in the Unit struct
		if isBuiltinField(key) {
			continue
		}

		// This must be a subunit
		subunit, err := parseUnit(key, val)
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

// isBuiltinField returns true if the key is a known Unit struct field.
func isBuiltinField(key string) bool {
	builtinFields := map[string]bool{
		"type":        true,
		"name":        true,
		"description": true,
		"technology":  true,
		"color":       true,
		"style":       true,
		"border":      true,
		"edges":       true,
		"width":       true,
		"height":      true,
		"expanded":    true,
		"link":        true,
		"linkFrom":    true,
	}

	return builtinFields[key]
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

// populateLinkTargets recursively populates the Target field in Link structs
// from their map keys. This handles nested units at any depth.
func populateLinkTargets(units map[string]*model.Unit) {
	for _, unit := range units {
		if unit == nil {
			continue
		}

		// Populate Links map keys into Target field
		for target, link := range unit.Links {
			link.Target = target
			unit.Links[target] = link
		}

		// Populate LinksFrom map keys into Target field
		for source, link := range unit.LinksFrom {
			link.Target = source
			unit.LinksFrom[source] = link
		}

		// Recurse into subunits
		if unit.Subunits != nil {
			populateLinkTargets(unit.Subunits)
		}
	}
}
