package validator

import (
	"maps"

	"github.com/Djarvur/c4drill/internal/model"
)

// UnitInfo holds validation metadata for a unit.
// It provides the full path and parent reference for nested units.
type UnitInfo struct {
	// Unit is a pointer to the actual unit data.
	Unit *model.Unit
	// FullPath is the complete dotted path (e.g., "mainapp.api.handler").
	FullPath string
	// Parent is the parent's full path (empty for top-level units).
	Parent string
}

// BuildIndex creates a flat map of all units for O(1) lookup.
// It recursively traverses subunits, building dotted paths.
// The parentPath parameter is used for recursive calls to build nested paths.
func BuildIndex(units map[string]*model.Unit, parentPath string) map[string]*UnitInfo {
	index := make(map[string]*UnitInfo)

	for name, unit := range units {
		fullPath := name
		if parentPath != "" {
			fullPath = parentPath + "." + name
		}

		index[fullPath] = &UnitInfo{
			Unit:     unit,
			FullPath: fullPath,
			Parent:   parentPath,
		}

		// Recursively add subunits
		if len(unit.Subunits) > 0 {
			subIndex := BuildIndex(unit.Subunits, fullPath)
			maps.Copy(index, subIndex)
		}
	}

	return index
}

// populateIncomingLinks adds LinksFrom entries based on Links from other units.
// When unit A has a link to unit B, this adds a LinksFrom entry to B pointing to A.
// This ensures the orphan validation correctly identifies units that are linked to.
func populateIncomingLinks(index map[string]*UnitInfo) {
	for sourcePath, sourceInfo := range index {
		for _, link := range sourceInfo.Unit.Links {
			targetInfo, exists := index[link.Peer]
			if !exists {
				continue // Skip if target doesn't exist (reference validation handles this)
			}

			// Check if this incoming link already exists
			if _, found := model.FindLinkByPeer(targetInfo.Unit.LinksFrom, sourcePath); found {
				continue
			}

			// Add reverse link entry
			targetInfo.Unit.LinksFrom = append(targetInfo.Unit.LinksFrom, model.Link{
				Peer: sourcePath,
			})
		}
	}
}
