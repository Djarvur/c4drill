package graph_test

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/render"
	"github.com/Djarvur/c4drill/internal/validator"
	"github.com/Djarvur/c4drill/internal/view"
	"github.com/stretchr/testify/require"
)

func TestWrapSmokeC3RendersValidDOT(t *testing.T) {
	m, err := parser.ParseFile("../../cmd/c4drill/testdata/multilevel.toml")
	require.NoError(t, err)
	require.Empty(t, validator.Validate(m))

	for _, path := range []string{"mainSystem", "mainSystem.storages.localStorage", "mainSystem.sshAuth"} {
		var v *view.View
		if len([]rune(path)) > 0 {
			v = view.GenerateC3View(m, path)
			if v == nil {
				v = view.GenerateC2View(m, path)
			}
		}

		g := graph.BuildGraphWithPath(v, path, "ml", "svg")
		dot, err := render.RenderDOT(g)
		require.NoError(t, err, path)
		require.NotEmpty(t, dot, path)
	}
}
