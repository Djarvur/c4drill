// Internal test file (package validator): TestMirrorOrderDeterministicAcrossRuns
// exercises the unexported populateIncomingLinks directly. The BuildIndex tests
// below use only exported API and were kept in this file when it switched from
// the external validator_test package.
package validator

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildIndex_SingleTopLevel(t *testing.T) {
	t.Parallel()

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
	assert.Empty(t, index["api"].Parent)
	assert.Equal(t, model.TypeSystem, index["api"].Unit.Type)
}

func TestBuildIndex_MultipleTopLevel(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

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
	assert.Empty(t, index["mainapp"].Parent)

	// Container has system as parent
	assert.Equal(t, "mainapp", index["mainapp.api"].Parent)

	// Component has container as parent
	assert.Equal(t, "mainapp.api", index["mainapp.api.handler"].Parent)
}

func TestBuildIndex_EmptyUnits(t *testing.T) {
	t.Parallel()

	units := map[string]*model.Unit{}

	index := BuildIndex(units, "")

	assert.Empty(t, index)
}

func TestBuildIndex_WithParentPath(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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

// TestMirrorOrderDeterministicAcrossRuns pins the per-target LinksFrom mirror
// order as a pure function of model content (issue #42, decision D-02).
//
// Three distinct source units (a, b, z — deliberately not inserted in sorted
// order) each link to the same target t. populateIncomingLinks ranges over the
// BuildIndex map, so on pre-fix code the mirror append order permutes between
// runs, which downstream permutes global edge insertion order and therefore
// GraphViz's edge<N> SVG group ids. The test repeats BuildIndex +
// populateIncomingLinks from scratch 20 times and requires every run to record
// the same mirror Peer sequence. Only Mirror-flagged entries are asserted —
// the authored linkFrom entry on the target is excluded from the order check.
func TestMirrorOrderDeterministicAcrossRuns(t *testing.T) {
	t.Parallel()

	// buildUnits returns a fresh model.Unit tree on every call:
	// populateIncomingLinks appends to Unit.LinksFrom, so runs must never
	// share Unit pointers or mirrors would accumulate across iterations.
	buildUnits := func() map[string]*model.Unit {
		return map[string]*model.Unit{
			// Keys deliberately NOT in sorted order in the literal — the
			// deterministic order must come from sorted-key iteration, not
			// from incidental map layout.
			"z": {Type: model.TypeContainer, Name: "Z", Links: []model.Link{{Peer: "t"}}},
			"b": {Type: model.TypeContainer, Name: "B", Links: []model.Link{{Peer: "t"}}},
			"a": {Type: model.TypeContainer, Name: "A", Links: []model.Link{{Peer: "t"}}},
			"t": {
				Type: model.TypeContainer,
				Name: "T",
				// Authored (non-mirror) incoming link — excluded from the
				// mirror-order assertion below.
				LinksFrom: []model.Link{{Peer: "ext"}},
			},
		}
	}

	mirrorPeers := func() []string {
		index := BuildIndex(buildUnits(), "")
		populateIncomingLinks(index)

		var peers []string
		for _, link := range index["t"].Unit.LinksFrom {
			if link.Mirror {
				peers = append(peers, link.Peer)
			}
		}
		return peers
	}

	first := mirrorPeers()
	require.Len(t, first, 3, "expected three mirror LinksFrom entries on target t, got %v", first)

	for run := 1; run < 20; run++ {
		got := mirrorPeers()
		require.Equal(t, first, got,
			"mirror Peer order differs on run %d: first run recorded %v, run %d recorded %v",
			run, first, run, got)
	}
}
