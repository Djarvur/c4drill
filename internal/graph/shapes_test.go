package graph_test

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestShapeForType(t *testing.T) {
	t.Parallel()

	// Test 1: ShapeForType returns ShapeRecord for all types (collapsed units)
	types := []model.UnitType{
		model.TypePerson, model.TypePersonExternal,
		model.TypeSystem, model.TypeSystemExternal,
		model.TypeDb, model.TypeDbExternal,
		model.TypeQueue, model.TypeQueueExternal,
		model.TypeBox,
		model.TypeContainer, model.TypeContainerDb, model.TypeContainerQueue, model.TypeContainerBox,
		model.TypeComponent, model.TypeComponentDb, model.TypeComponentQueue, model.TypeComponentBox,
	}

	for _, typ := range types {
		assert.Equal(t, graph.ShapeRecord, graph.ShapeForType(typ), "ShapeForType(%s)", typ)
	}
}

func TestIconForType(t *testing.T) {
	t.Parallel()

	// Test 2: IconForType returns correct emoji for person types
	assert.Equal(t, "\U0001F464", graph.IconForType(model.TypePerson))
	assert.Equal(t, "\U0001F464", graph.IconForType(model.TypePersonExternal))

	// Test 3: IconForType returns correct emoji for db types
	assert.Equal(t, "\u26C1", graph.IconForType(model.TypeDb))
	assert.Equal(t, "\u26C1", graph.IconForType(model.TypeDbExternal))
	assert.Equal(t, "\u26C1", graph.IconForType(model.TypeContainerDb))
	assert.Equal(t, "\u26C1", graph.IconForType(model.TypeComponentDb))

	// Test 4: IconForType returns correct emoji for queue types
	assert.Equal(t, "\u255F\n\u2562", graph.IconForType(model.TypeQueue))
	assert.Equal(t, "\u255F\n\u2562", graph.IconForType(model.TypeQueueExternal))
	assert.Equal(t, "\u255F\n\u2562", graph.IconForType(model.TypeContainerQueue))
	assert.Equal(t, "\u255F\n\u2562", graph.IconForType(model.TypeComponentQueue))

	// Test 5: IconForType returns empty string for system/container/component/box types
	assert.Empty(t, graph.IconForType(model.TypeSystem))
	assert.Empty(t, graph.IconForType(model.TypeSystemExternal))
	assert.Empty(t, graph.IconForType(model.TypeBox))
	assert.Empty(t, graph.IconForType(model.TypeContainer))
	assert.Empty(t, graph.IconForType(model.TypeContainerBox))
	assert.Empty(t, graph.IconForType(model.TypeComponent))
	assert.Empty(t, graph.IconForType(model.TypeComponentBox))
}

func TestIsExternalType(t *testing.T) {
	t.Parallel()

	// Test 6: IsExternalType correctly identifies external variants
	externalTypes := []model.UnitType{
		model.TypePersonExternal,
		model.TypeSystemExternal,
		model.TypeDbExternal,
		model.TypeQueueExternal,
	}
	for _, typ := range externalTypes {
		assert.True(t, graph.IsExternalType(typ), "IsExternalType(%s) should be true", typ)
	}

	internalTypes := []model.UnitType{
		model.TypePerson, model.TypeSystem, model.TypeDb, model.TypeQueue, model.TypeBox,
		model.TypeContainer, model.TypeContainerDb, model.TypeContainerQueue, model.TypeContainerBox,
		model.TypeComponent, model.TypeComponentDb, model.TypeComponentQueue, model.TypeComponentBox,
	}
	for _, typ := range internalTypes {
		assert.False(t, graph.IsExternalType(typ), "IsExternalType(%s) should be false", typ)
	}
}

func TestLevelForType(t *testing.T) {
	t.Parallel()

	// Test 7: LevelForType returns 1 for C1 types, 2 for C2 types, 3 for C3 types
	c1Types := []model.UnitType{
		model.TypePerson, model.TypePersonExternal,
		model.TypeSystem, model.TypeSystemExternal,
		model.TypeDb, model.TypeDbExternal,
		model.TypeQueue, model.TypeQueueExternal,
		model.TypeBox,
	}
	for _, typ := range c1Types {
		assert.Equal(t, 1, graph.LevelForType(typ), "LevelForType(%s)", typ)
	}

	c2Types := []model.UnitType{
		model.TypeContainer, model.TypeContainerDb, model.TypeContainerQueue, model.TypeContainerBox,
	}
	for _, typ := range c2Types {
		assert.Equal(t, 2, graph.LevelForType(typ), "LevelForType(%s)", typ)
	}

	c3Types := []model.UnitType{
		model.TypeComponent, model.TypeComponentDb, model.TypeComponentQueue, model.TypeComponentBox,
	}
	for _, typ := range c3Types {
		assert.Equal(t, 3, graph.LevelForType(typ), "LevelForType(%s)", typ)
	}
}

func TestGetStyleForType(t *testing.T) {
	t.Parallel()

	// Test 8: GetStyleForType returns transparent fill for C1 level internal types
	// C1 uses dark blue (PersonBorder) for all unit types
	c1Style := graph.GetStyleForType(model.TypeSystem, false)
	assert.Empty(t, c1Style.FillColor) // Transparent background
	assert.Equal(t, model.PersonBorder, c1Style.BorderColor) // Dark blue for C1
	assert.Equal(t, model.PersonBorder, c1Style.FontColor)   // Font color matches border color
	assert.Equal(t, "solid", c1Style.BorderStyle)

	// C2 level - blue
	c2Style := graph.GetStyleForType(model.TypeContainer, false)
	assert.Empty(t, c2Style.FillColor) // Transparent background
	assert.Equal(t, model.ContainerBorder, c2Style.BorderColor)
	assert.Equal(t, model.ContainerBorder, c2Style.FontColor) // Font color matches border color
	assert.Equal(t, "solid", c2Style.BorderStyle)

	// C3 level - light blue
	c3Style := graph.GetStyleForType(model.TypeComponent, false)
	assert.Empty(t, c3Style.FillColor) // Transparent background
	assert.Equal(t, model.ComponentBorder, c3Style.BorderColor)
	assert.Equal(t, model.ComponentBorder, c3Style.FontColor) // Font color matches border color
	assert.Equal(t, "solid", c3Style.BorderStyle)

	// Test 9: GetStyleForType returns transparent fill for external types
	// C1 external uses dark gray (PersonExternalBorder)
	extStyle := graph.GetStyleForType(model.TypeSystemExternal, true)
	assert.Empty(t, extStyle.FillColor) // Transparent background
	assert.Equal(t, model.PersonExternalBorder, extStyle.BorderColor) // Dark gray for C1 external
	assert.Equal(t, model.PersonExternalBorder, extStyle.FontColor)   // Font color matches border color

	// Test 10: GetStyleForType returns solid border style for external nodes
	assert.Equal(t, "solid", extStyle.BorderStyle)
}

func TestGetStyleForType_BoxDashedBorders(t *testing.T) {
	t.Parallel()

	// Test: TypeBox returns dashed border style
	boxStyle := graph.GetStyleForType(model.TypeBox, false)
	assert.Equal(t, "dashed", boxStyle.BorderStyle, "TypeBox should have dashed border")

	// Test: TypeContainerBox returns dashed border style
	containerBoxStyle := graph.GetStyleForType(model.TypeContainerBox, false)
	assert.Equal(t, "dashed", containerBoxStyle.BorderStyle, "TypeContainerBox should have dashed border")

	// Test: TypeComponentBox returns dashed border style
	componentBoxStyle := graph.GetStyleForType(model.TypeComponentBox, false)
	assert.Equal(t, "dashed", componentBoxStyle.BorderStyle, "TypeComponentBox should have dashed border")

	// Test: TypeSystem returns solid border style (unchanged)
	systemStyle := graph.GetStyleForType(model.TypeSystem, false)
	assert.Equal(t, "solid", systemStyle.BorderStyle, "TypeSystem should have solid border")

	// Test: TypeContainer returns solid border style (unchanged)
	containerStyle := graph.GetStyleForType(model.TypeContainer, false)
	assert.Equal(t, "solid", containerStyle.BorderStyle, "TypeContainer should have solid border")

	// Test: TypeComponent returns solid border style (unchanged)
	componentStyle := graph.GetStyleForType(model.TypeComponent, false)
	assert.Equal(t, "solid", componentStyle.BorderStyle, "TypeComponent should have solid border")
}

func TestHasExternalSubunits(t *testing.T) {
	t.Parallel()

	// Test: Returns true for box with TypePersonExternal subunit
	boxWithPersonExt := &model.Unit{
		Type: model.TypeBox,
		Subunits: map[string]*model.Unit{
			"extPerson": {Type: model.TypePersonExternal},
		},
	}
	assert.True(t, graph.HasExternalSubunits(boxWithPersonExt), "box with TypePersonExternal subunit should return true")

	// Test: Returns true for box with TypeSystemExternal subunit
	boxWithSystemExt := &model.Unit{
		Type: model.TypeBox,
		Subunits: map[string]*model.Unit{
			"extSystem": {Type: model.TypeSystemExternal},
		},
	}
	assert.True(t, graph.HasExternalSubunits(boxWithSystemExt), "box with TypeSystemExternal subunit should return true")

	// Test: Returns true for box with TypeDbExternal subunit
	boxWithDbExt := &model.Unit{
		Type: model.TypeBox,
		Subunits: map[string]*model.Unit{
			"extDb": {Type: model.TypeDbExternal},
		},
	}
	assert.True(t, graph.HasExternalSubunits(boxWithDbExt), "box with TypeDbExternal subunit should return true")

	// Test: Returns true for box with TypeQueueExternal subunit
	boxWithQueueExt := &model.Unit{
		Type: model.TypeBox,
		Subunits: map[string]*model.Unit{
			"extQueue": {Type: model.TypeQueueExternal},
		},
	}
	assert.True(t, graph.HasExternalSubunits(boxWithQueueExt), "box with TypeQueueExternal subunit should return true")

	// Test: Returns false for box with only TypePerson subunit
	boxWithPerson := &model.Unit{
		Type: model.TypeBox,
		Subunits: map[string]*model.Unit{
			"person": {Type: model.TypePerson},
		},
	}
	assert.False(t, graph.HasExternalSubunits(boxWithPerson), "box with only TypePerson subunit should return false")

	// Test: Returns false for box with only TypeSystem subunit
	boxWithSystem := &model.Unit{
		Type: model.TypeBox,
		Subunits: map[string]*model.Unit{
			"system": {Type: model.TypeSystem},
		},
	}
	assert.False(t, graph.HasExternalSubunits(boxWithSystem), "box with only TypeSystem subunit should return false")

	// Test: Returns false for box with no subunits
	emptyBox := &model.Unit{
		Type:     model.TypeBox,
		Subunits: nil,
	}
	assert.False(t, graph.HasExternalSubunits(emptyBox), "box with no subunits should return false")

	// Test: Returns false for nil unit
	assert.False(t, graph.HasExternalSubunits(nil), "nil unit should return false")
}

func TestGetBoxStyleByContents(t *testing.T) {
	t.Parallel()

	// Test: Returns grey border for box with external subunits
	boxWithExternal := &model.Unit{
		Type: model.TypeBox,
		Subunits: map[string]*model.Unit{
			"extPerson": {Type: model.TypePersonExternal},
			"extSystem": {Type: model.TypeSystemExternal},
		},
	}
	styleExternal := graph.GetBoxStyleByContents(boxWithExternal)
	assert.Equal(t, model.PersonExternalBorder, styleExternal.BorderColor,
		"box with external subunits should have grey border")
	assert.Equal(t, model.PersonExternalBorder, styleExternal.FontColor,
		"box with external subunits should have grey font color")
	assert.Equal(t, "dashed", styleExternal.BorderStyle, "box should have dashed border")
	assert.Empty(t, styleExternal.FillColor, "box should have transparent fill")

	// Test: Returns dark blue border for box with only non-external subunits
	boxWithInternal := &model.Unit{
		Type: model.TypeBox,
		Subunits: map[string]*model.Unit{
			"person": {Type: model.TypePerson},
			"system": {Type: model.TypeSystem},
		},
	}
	styleInternal := graph.GetBoxStyleByContents(boxWithInternal)
	assert.Equal(t, model.PersonBorder, styleInternal.BorderColor,
		"box with only non-external subunits should have dark blue border")
	assert.Equal(t, model.PersonBorder, styleInternal.FontColor,
		"box with only non-external subunits should have dark blue font color")
	assert.Equal(t, "dashed", styleInternal.BorderStyle, "box should have dashed border")
	assert.Empty(t, styleInternal.FillColor, "box should have transparent fill")

	// Test: Returns dark blue border for empty box (no external subunits)
	emptyBox := &model.Unit{
		Type:     model.TypeBox,
		Subunits: nil,
	}
	styleEmpty := graph.GetBoxStyleByContents(emptyBox)
	assert.Equal(t, model.PersonBorder, styleEmpty.BorderColor, "empty box should have dark blue border")
	assert.Equal(t, "dashed", styleEmpty.BorderStyle, "box should have dashed border")

	// Test: Returns dark blue border for nil unit
	styleNil := graph.GetBoxStyleByContents(nil)
	assert.Equal(t, model.PersonBorder, styleNil.BorderColor, "nil unit should have dark blue border")
	assert.Equal(t, "dashed", styleNil.BorderStyle, "box should have dashed border")
}
