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
		model.TypeContainer, model.TypeContainerDb, model.TypeContainerQueue,
		model.TypeComponent, model.TypeComponentDb, model.TypeComponentQueue,
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

	// Test 5: IconForType returns empty string for system/container/component/box
	assert.Empty(t, graph.IconForType(model.TypeSystem))
	assert.Empty(t, graph.IconForType(model.TypeSystemExternal))
	assert.Empty(t, graph.IconForType(model.TypeBox))
	assert.Empty(t, graph.IconForType(model.TypeContainer))
	assert.Empty(t, graph.IconForType(model.TypeComponent))
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
		model.TypeContainer, model.TypeContainerDb, model.TypeContainerQueue,
		model.TypeComponent, model.TypeComponentDb, model.TypeComponentQueue,
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
		model.TypeContainer, model.TypeContainerDb, model.TypeContainerQueue,
	}
	for _, typ := range c2Types {
		assert.Equal(t, 2, graph.LevelForType(typ), "LevelForType(%s)", typ)
	}

	c3Types := []model.UnitType{
		model.TypeComponent, model.TypeComponentDb, model.TypeComponentQueue,
	}
	for _, typ := range c3Types {
		assert.Equal(t, 3, graph.LevelForType(typ), "LevelForType(%s)", typ)
	}
}

func TestGetStyleForType(t *testing.T) {
	t.Parallel()

	// Test 8: GetStyleForType returns transparent fill for C1 level internal types
	c1Style := graph.GetStyleForType(model.TypeSystem, false)
	assert.Empty(t, c1Style.FillColor) // Transparent background
	assert.Equal(t, model.SystemBorder, c1Style.BorderColor)
	assert.Equal(t, model.FontColorC1C2, c1Style.FontColor)
	assert.Equal(t, "solid", c1Style.BorderStyle)

	// C2 level
	c2Style := graph.GetStyleForType(model.TypeContainer, false)
	assert.Empty(t, c2Style.FillColor) // Transparent background
	assert.Equal(t, model.ContainerBorder, c2Style.BorderColor)
	assert.Equal(t, model.FontColorC1C2, c2Style.FontColor)
	assert.Equal(t, "solid", c2Style.BorderStyle)

	// C3 level
	c3Style := graph.GetStyleForType(model.TypeComponent, false)
	assert.Empty(t, c3Style.FillColor) // Transparent background
	assert.Equal(t, model.ComponentBorder, c3Style.BorderColor)
	assert.Equal(t, model.FontColorC3, c3Style.FontColor)
	assert.Equal(t, "solid", c3Style.BorderStyle)

	// Test 9: GetStyleForType returns transparent fill for external types
	extStyle := graph.GetStyleForType(model.TypeSystemExternal, true)
	assert.Empty(t, extStyle.FillColor) // Transparent background
	assert.Equal(t, model.SystemExternalBorder, extStyle.BorderColor)

	// Test 10: GetStyleForType returns dashed border style for external nodes
	assert.Equal(t, "dashed", extStyle.BorderStyle)
}
