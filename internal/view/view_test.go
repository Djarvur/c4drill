package view_test

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestView_StructContainsFields(t *testing.T) {
	t.Parallel()

	v := &view.View{
		Level:        view.LevelC1,
		Title:        "Test Diagram",
		Units:        make(map[string]*view.ViewUnit),
		Edges:        "spline",
		Parent:       "",
		ExpandedUnit: "",
	}

	assert.Equal(t, view.LevelC1, v.Level)
	assert.Equal(t, "Test Diagram", v.Title)
	assert.NotNil(t, v.Units)
	assert.Equal(t, "spline", v.Edges)
	assert.Empty(t, v.Parent)
	assert.Empty(t, v.ExpandedUnit)
}

func TestViewUnit_StructContainsFields(t *testing.T) {
	t.Parallel()

	unit := &model.Unit{
		Type:        model.TypeSystem,
		Name:        "Test System",
		Description: "A test system",
		Technology:  "Go",
	}

	vu := &view.ViewUnit{
		Unit:        unit,
		FullPath:    "system.api",
		IsExpanded:  true,
		IsExternal:  false,
		HasSubunits: true,
	}

	require.NotNil(t, vu.Unit)
	assert.Equal(t, model.TypeSystem, vu.Unit.Type)
	assert.Equal(t, "system.api", vu.FullPath)
	assert.True(t, vu.IsExpanded)
	assert.False(t, vu.IsExternal)
	assert.True(t, vu.HasSubunits)
}

func TestLevel_Constants(t *testing.T) {
	t.Parallel()

	// Verify Level constants exist and have distinct values
	assert.Equal(t, view.Level(0), view.LevelC1)
	assert.Equal(t, view.Level(1), view.LevelC2)
	assert.Equal(t, view.Level(2), view.LevelC3)

	// Verify they are distinct
	assert.NotEqual(t, view.LevelC1, view.LevelC2)
	assert.NotEqual(t, view.LevelC2, view.LevelC3)
	assert.NotEqual(t, view.LevelC1, view.LevelC3)
}

func TestIsExternalType_CorrectlyIdentifiesExternalTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		unitType model.UnitType
		expected bool
	}{
		// External types
		{"personExternal", model.TypePersonExternal, true},
		{"systemExternal", model.TypeSystemExternal, true},
		{"dbExternal", model.TypeDbExternal, true},
		{"queueExternal", model.TypeQueueExternal, true},

		// Non-external types
		{"person", model.TypePerson, false},
		{"system", model.TypeSystem, false},
		{"db", model.TypeDb, false},
		{"queue", model.TypeQueue, false},
		{"box", model.TypeBox, false},
		{"container", model.TypeContainer, false},
		{"containerDb", model.TypeContainerDb, false},
		{"containerQueue", model.TypeContainerQueue, false},
		{"component", model.TypeComponent, false},
		{"componentDb", model.TypeComponentDb, false},
		{"componentQueue", model.TypeComponentQueue, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := view.IsExternalType(tt.unitType)
			assert.Equal(t, tt.expected, result, "IsExternalType(%s)", tt.unitType)
		})
	}
}
