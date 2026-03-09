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
	// The toml:",inline" tag captures all sections except [properties].
	Units map[string]*model.Unit `toml:",inline"`
}

// Parse parses TOML data into a Model.
// It unmarshals the TOML content and populates Link.Target fields from map keys.
func Parse(data []byte) (*Model, error) {
	var m Model
	if err := toml.Unmarshal(data, &m); err != nil {
		return nil, wrapDecodeError(err)
	}

	// Populate Link.Target from map keys for all units
	if m.Units != nil {
		populateLinkTargets(m.Units)
	}

	return &m, nil
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
