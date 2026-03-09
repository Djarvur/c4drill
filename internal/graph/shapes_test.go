package graph

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestShapeForType(t *testing.T) {
	t.Parallel()

	// Test 1: ShapeForType returns ShapeHTML for all types (all use HTML-like labels)
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
		assert.Equal(t, ShapeHTML, ShapeForType(typ), "ShapeForType(%s)", typ)
	}
}

func TestIconForType(t *testing.T) {
	t.Parallel()

	// Test 2: IconForType returns correct emoji for person types
	assert.Equal(t, "\U0001F464", IconForType(model.TypePerson))
	assert.Equal(t, "\U0001F464", IconForType(model.TypePersonExternal))

	// Test 3: IconForType returns correct emoji for db types
	assert.Equal(t, "\u26C1", IconForType(model.TypeDb))
	assert.Equal(t, "\u26C1", IconForType(model.TypeDbExternal))
	assert.Equal(t, "\u26C1", IconForType(model.TypeContainerDb))
	assert.Equal(t, "\u26C1", IconForType(model.TypeComponentDb))

	// Test 4: IconForType returns correct emoji for queue types
	assert.Equal(t, "\u255F\n\u2562", IconForType(model.TypeQueue))
	assert.Equal(t, "\u255F\n\u2562", IconForType(model.TypeQueueExternal))
	assert.Equal(t, "\u255F\n\u2562", IconForType(model.TypeContainerQueue))
	assert.Equal(t, "\u255F\n\u2562", IconForType(model.TypeComponentQueue))

	// Test 5: IconForType returns empty string for system/container/component/box
	assert.Equal(t, "", IconForType(model.TypeSystem))
	assert.Equal(t, "", IconForType(model.TypeSystemExternal))
	assert.Equal(t, "", IconForType(model.TypeBox))
	assert.Equal(t, "", IconForType(model.TypeContainer))
	assert.Equal(t, "", IconForType(model.TypeComponent))
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
		assert.True(t, IsExternalType(typ), "IsExternalType(%s) should be true", typ)
	}

	internalTypes := []model.UnitType{
		model.TypePerson, model.TypeSystem, model.TypeDb, model.TypeQueue, model.TypeBox,
		model.TypeContainer, model.TypeContainerDb, model.TypeContainerQueue,
		model.TypeComponent, model.TypeComponentDb, model.TypeComponentQueue,
	}
	for _, typ := range internalTypes {
		assert.False(t, IsExternalType(typ), "IsExternalType(%s) should be false", typ)
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
		assert.Equal(t, 1, LevelForType(typ), "LevelForType(%s)", typ)
	}

	c2Types := []model.UnitType{
		model.TypeContainer, model.TypeContainerDb, model.TypeContainerQueue,
	}
	for _, typ := range c2Types {
		assert.Equal(t, 2, LevelForType(typ), "LevelForType(%s)", typ)
	}

	c3Types := []model.UnitType{
		model.TypeComponent, model.TypeComponentDb, model.TypeComponentQueue,
	}
	for _, typ := range c3Types {
		assert.Equal(t, 3, LevelForType(typ), "LevelForType(%s)", typ)
	}
}

func TestGetStyleForType(t *testing.T) {
	t.Parallel()

	// Test 8: GetStyleForType returns correct colors for C1 level internal types
	c1Style := GetStyleForType(model.TypeSystem, false)
	assert.Equal(t, model.SystemBackground, c1Style.FillColor)
	assert.Equal(t, model.SystemBorder, c1Style.BorderColor)
	assert.Equal(t, model.FontColorC1C2, c1Style.FontColor)
	assert.Equal(t, "solid", c1Style.BorderStyle)

	// C2 level
	c2Style := GetStyleForType(model.TypeContainer, false)
	assert.Equal(t, model.ContainerBackground, c2Style.FillColor)
	assert.Equal(t, model.ContainerBorder, c2Style.BorderColor)
	assert.Equal(t, model.FontColorC1C2, c2Style.FontColor)
	assert.Equal(t, "solid", c2Style.BorderStyle)

	// C3 level
	c3Style := GetStyleForType(model.TypeComponent, false)
	assert.Equal(t, model.ComponentBackground, c3Style.FillColor)
	assert.Equal(t, model.ComponentBorder, c3Style.BorderColor)
	assert.Equal(t, model.FontColorC3, c3Style.FontColor)
	assert.Equal(t, "solid", c3Style.BorderStyle)

	// Test 9: GetStyleForType returns external palette colors for external types
	extStyle := GetStyleForType(model.TypeSystemExternal, true)
	assert.Equal(t, model.SystemExternalBackground, extStyle.FillColor)
	assert.Equal(t, model.SystemExternalBorder, extStyle.BorderColor)

	// Test 10: GetStyleForType returns dashed border style for external nodes
	assert.Equal(t, "dashed", extStyle.BorderStyle)
}
