package validator

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestBuildIndex_SingleTopLevel(t *testing.T) {
	units := map[string]*model.Unit{
		"api": {
			Type:        model.TypeSystem,
			Name:        "API System",
			Description: "Main API",
		},
	}

	index := BuildIndex(units, "")

	assert.Len(t, index, 1)
	assert.Contains(t, index, "api")
	assert.Equal(t, "api", index["api"].FullPath)
	assert.Equal(t, "", index["api"].Parent)
	assert.Equal(t, model.TypeSystem, index["api"].Unit.Type)
}

func TestBuildIndex_MultipleTopLevel(t *testing.T) {
	units := map[string]*model.Unit{
		"api": {
			Type: model.TypeSystem,
			Name: "API",
		},
		"db": {
			Type: model.TypeDb,
			Name: "Database",
		},
		"user": {
			Type: model.TypePerson,
			Name: "User",
		},
	}

	index := BuildIndex(units, "")

	assert.Len(t, index, 3)
	assert.Contains(t, index, "api")
	assert.Contains(t, index, "db")
	assert.Contains(t, index, "user")
}

func TestBuildIndex_NestedUnits(t *testing.T) {
	units := map[string]*model.Unit{
		"mainapp": {
			Type: model.TypeSystem,
			Name: "Main App",
			Subunits: map[string]*model.Unit{
				"api": {
					Type: model.TypeContainer,
					Name: "API Container",
				},
			},
		},
	}

	index := BuildIndex(units, "")

	assert.Len(t, index, 2)
	assert.Contains(t, index, "mainapp")
	assert.Contains(t, index, "mainapp.api")
	assert.Equal(t, "mainapp", index["mainapp"].FullPath)
	assert.Equal(t, "mainapp.api", index["mainapp.api"].FullPath)
}

func TestBuildIndex_DeepNesting(t *testing.T) {
	units := map[string]*model.Unit{
		"mainapp": {
			Type: model.TypeSystem,
			Name: "Main App",
			Subunits: map[string]*model.Unit{
				"api": {
					Type: model.TypeContainer,
					Name: "API Container",
					Subunits: map[string]*model.Unit{
						"handler": {
							Type: model.TypeComponent,
							Name: "Handler",
						},
					},
				},
			},
		},
	}

	index := BuildIndex(units, "")

	assert.Len(t, index, 3)
	assert.Contains(t, index, "mainapp")
	assert.Contains(t, index, "mainapp.api")
	assert.Contains(t, index, "mainapp.api.handler")
}

func TestBuildIndex_ParentPaths(t *testing.T) {
	units := map[string]*model.Unit{
		"mainapp": {
			Type: model.TypeSystem,
			Name: "Main App",
			Subunits: map[string]*model.Unit{
				"api": {
					Type: model.TypeContainer,
					Name: "API",
					Subunits: map[string]*model.Unit{
						"handler": {
							Type: model.TypeComponent,
							Name: "Handler",
						},
					},
				},
			},
		},
	}

	index := BuildIndex(units, "")

	// Top-level has no parent
	assert.Equal(t, "", index["mainapp"].Parent)

	// Container has system as parent
	assert.Equal(t, "mainapp", index["mainapp.api"].Parent)

	// Component has container as parent
	assert.Equal(t, "mainapp.api", index["mainapp.api.handler"].Parent)
}

func TestBuildIndex_EmptyUnits(t *testing.T) {
	units := map[string]*model.Unit{}

	index := BuildIndex(units, "")

	assert.Empty(t, index)
}

func TestBuildIndex_WithParentPath(t *testing.T) {
	units := map[string]*model.Unit{
		"handler": {
			Type: model.TypeComponent,
			Name: "Handler",
		},
	}

	// Simulate being called from a parent context
	index := BuildIndex(units, "mainapp.api")

	assert.Len(t, index, 1)
	assert.Contains(t, index, "mainapp.api.handler")
	assert.Equal(t, "mainapp.api.handler", index["mainapp.api.handler"].FullPath)
	assert.Equal(t, "mainapp.api", index["mainapp.api.handler"].Parent)
}

func TestBuildIndex_MultipleBranches(t *testing.T) {
	units := map[string]*model.Unit{
		"mainapp": {
			Type: model.TypeSystem,
			Name: "Main App",
			Subunits: map[string]*model.Unit{
				"api": {
					Type: model.TypeContainer,
					Name: "API",
				},
				"web": {
					Type: model.TypeContainer,
					Name: "Web",
				},
				"db": {
					Type: model.TypeContainerDb,
					Name: "Database",
				},
			},
		},
	}

	index := BuildIndex(units, "")

	assert.Len(t, index, 4)
	assert.Contains(t, index, "mainapp")
	assert.Contains(t, index, "mainapp.api")
	assert.Contains(t, index, "mainapp.web")
	assert.Contains(t, index, "mainapp.db")
}
