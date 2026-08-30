package view_test

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/render"
	"github.com/Djarvur/c4drill/internal/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntegrationC1ViewMultiLevelModel tests that GenerateC1View with a multi-level
// model produces the correct top-level view showing only top-level units.
func TestIntegrationC1ViewMultiLevelModel(t *testing.T) {
	t.Parallel()

	// Create a model with nested units (system -> container -> component)
	m := &parser.Model{
		Properties: model.Properties{
			Name: "Multi-Level System",
		},
		Units: map[string]*model.Unit{
			"mainsystem": {
				Type:        model.TypeSystem,
				Name:        "Main System",
				Description: "The main system",
				Technology:  "Go",
				Subunits: map[string]*model.Unit{
					"api": {
						Type:        model.TypeContainer,
						Name:        "API",
						Description: "API container",
						Subunits: map[string]*model.Unit{
							"handler": {
								Type:        model.TypeComponent,
								Name:        "Handler",
								Description: "Request handler",
							},
						},
					},
					"web": {
						Type:        model.TypeContainer,
						Name:        "Web",
						Description: "Web frontend",
					},
				},
			},
			"external": {
				Type:        model.TypeSystemExternal,
				Name:        "External System",
				Description: "An external system",
			},
		},
	}

	v := view.GenerateC1View(m)

	require.NotNil(t, v)
	// Test 1: C1 view only shows top-level units (not nested containers/components)
	assert.Equal(t, view.LevelC1, v.Level)
	assert.Contains(t, v.Units, "mainsystem")
	assert.Contains(t, v.Units, "external")

	// Nested units should NOT appear at C1 level
	assert.NotContains(t, v.Units, "mainsystem.api")
	assert.NotContains(t, v.Units, "mainsystem.api.handler")

	// HasSubunits should be true for system with subunits
	assert.True(t, v.Units["mainsystem"].HasSubunits)
}

// TestIntegrationC2ViewNestedSystem tests that GenerateC2View with a nested system
// produces the correct container view showing subunits.
func TestIntegrationC2ViewNestedSystem(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"mainsystem": {
				Type:        model.TypeSystem,
				Name:        "Main System",
				Description: "The main system",
				Technology:  "Go",
				Edges:       "spline",
				Subunits: map[string]*model.Unit{
					"api": {
						Type:        model.TypeContainer,
						Name:        "API",
						Description: "API container",
						Technology:  "Go",
					},
					"web": {
						Type:        model.TypeContainer,
						Name:        "Web",
						Description: "Web frontend",
						Technology:  "React",
					},
					"db": {
						Type:        model.TypeContainerDb,
						Name:        "Database",
						Description: "Data store",
						Technology:  "PostgreSQL",
					},
				},
			},
		},
	}

	v := view.GenerateC2View(m, "mainsystem")

	require.NotNil(t, v)
	// Test 2: C2 view shows containers of expanded system
	assert.Equal(t, view.LevelC2, v.Level)
	assert.Contains(t, v.Units, "mainsystem.api")
	assert.Contains(t, v.Units, "mainsystem.web")
	assert.Contains(t, v.Units, "mainsystem.db")

	// Parent and ExpandedUnit should be set
	assert.Equal(t, "mainsystem", v.Parent)
	assert.Equal(t, "mainsystem", v.ExpandedUnit)

	// Edges should come from expanded unit
	assert.Equal(t, "spline", v.Edges)

	// FullPath should use dotted notation
	assert.Equal(t, "mainsystem.api", v.Units["mainsystem.api"].FullPath)
}

// TestIntegrationC3ViewNestedContainer tests that GenerateC3View with a nested container
// produces the correct component view.
func TestIntegrationC3ViewNestedContainer(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"mainsystem": {
				Type: model.TypeSystem,
				Name: "Main System",
				Subunits: map[string]*model.Unit{
					"api": {
						Type:        model.TypeContainer,
						Name:        "API",
						Description: "API container",
						Technology:  "Go",
						Edges:       "square",
						Subunits: map[string]*model.Unit{
							"handler": {
								Type:        model.TypeComponent,
								Name:        "Handler",
								Description: "Request handler",
								Technology:  "Go",
							},
							"service": {
								Type:        model.TypeComponent,
								Name:        "Service",
								Description: "Business logic",
								Technology:  "Go",
							},
							"repo": {
								Type:        model.TypeComponentDb,
								Name:        "Repository",
								Description: "Data access",
								Technology:  "Go",
							},
						},
					},
				},
			},
		},
	}

	v := view.GenerateC3View(m, "mainsystem.api")

	require.NotNil(t, v)
	// Test 3: C3 view shows components of expanded container
	assert.Equal(t, view.LevelC3, v.Level)
	assert.Contains(t, v.Units, "mainsystem.api.handler")
	assert.Contains(t, v.Units, "mainsystem.api.service")
	assert.Contains(t, v.Units, "mainsystem.api.repo")

	// Edges should come from expanded container
	assert.Equal(t, "square", v.Edges)

	// FullPath uses dotted notation for nested units
	assert.Equal(t, "mainsystem.api.handler", v.Units["mainsystem.api.handler"].FullPath)
}

// TestIntegrationExternalBoundaryNodes tests that external boundary nodes appear
// when links reference out-of-scope units.
func TestIntegrationExternalBoundaryNodes(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"internal": {
				Type: model.TypeSystem,
				Name: "Internal System",
				Links: []model.Link{
					{
						Peer:        "externalapi",
						Technology:  "HTTP",
						Description: "Calls external API",
					},
				},
			},
		},
	}

	v := view.GenerateC1View(m)

	require.NotNil(t, v)
	// Test 4: External boundary nodes appear when links reference out-of-scope units
	assert.Contains(t, v.Units, "externalapi")
	assert.True(t, v.Units["externalapi"].IsExternal)
	assert.Equal(t, model.TypeSystemExternal, v.Units["externalapi"].Unit.Type)
}

// TestIntegrationViewRespectsExpandedAttribute tests that the view respects
// the per-unit expanded attribute.
func TestIntegrationViewRespectsExpandedAttribute(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"expanded_system": {
				Type:     model.TypeSystem,
				Name:     "Expanded System",
				Expanded: []string{"expanded_system"}, // This unit is expanded
				Subunits: map[string]*model.Unit{
					"container": {
						Type: model.TypeContainer,
						Name: "Container",
					},
				},
			},
			"collapsed_system": {
				Type: model.TypeSystem,
				Name: "Collapsed System",
				// No Expanded attribute - should be collapsed
				Subunits: map[string]*model.Unit{
					"container": {
						Type: model.TypeContainer,
						Name: "Container",
					},
				},
			},
		},
	}

	v := view.GenerateC1View(m)

	require.NotNil(t, v)
	// Test 5: View respects per-unit expanded attribute
	assert.True(t, v.Units["expanded_system"].IsExpanded)
	assert.False(t, v.Units["collapsed_system"].IsExpanded)
}

// TestIntegrationFullPathDottedNotation tests that FullPath correctly uses
// dotted notation for nested units at all levels.
func TestIntegrationFullPathDottedNotation(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"system": {
				Type: model.TypeSystem,
				Name: "System",
				Subunits: map[string]*model.Unit{
					"container": {
						Type: model.TypeContainer,
						Name: "Container",
						Subunits: map[string]*model.Unit{
							"component": {
								Type: model.TypeComponent,
								Name: "Component",
							},
						},
					},
				},
			},
		},
	}

	// Test 6: FullPath correctly uses dotted notation for nested units

	// C1: top-level units have simple paths
	c1 := view.GenerateC1View(m)
	require.NotNil(t, c1)
	assert.Equal(t, "system", c1.Units["system"].FullPath)

	// C2: containers have "system.container" paths
	c2 := view.GenerateC2View(m, "system")
	require.NotNil(t, c2)
	assert.Equal(t, "system.container", c2.Units["system.container"].FullPath)

	// C3: components have "system.container.component" paths
	c3 := view.GenerateC3View(m, "system.container")
	require.NotNil(t, c3)
	assert.Equal(t, "system.container.component", c3.Units["system.container.component"].FullPath)
}

// TestIntegrationLinksWithExternalBoundary tests that links to external units
// create proper external boundary nodes.
func TestIntegrationLinksWithExternalBoundary(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"system": {
				Type: model.TypeSystem,
				Name: "System",
				Subunits: map[string]*model.Unit{
					"container": {
						Type: model.TypeContainer,
						Name: "Container",
						Links: []model.Link{
							{
								Peer:        "externalservice",
								Technology:  "HTTP",
								Description: "External call",
							},
						},
					},
				},
			},
		},
	}

	// At C2 level, external boundary nodes should be created for container's links
	v := view.GenerateC2View(m, "system")
	require.NotNil(t, v)
	assert.Contains(t, v.Units, "externalservice")
	assert.True(t, v.Units["externalservice"].IsExternal)
}

// TestIntegrationC1ViewNoNestedBoundaryPollution verifies that deeply nested
// subunit links do not pollute C1 with arbitrary boundary nodes: only the
// ancestor chains of DEPICTED LINK TARGETS are added (CTX-02) — internal links
// (sshd -> nss) still depict nothing, and no other nested unit appears.
// This is the regression test for the bug where C1 showed ~100 nodes instead of ~5.
func TestIntegrationC1ViewNoNestedBoundaryPollution(t *testing.T) {
	t.Parallel()

	// Model mimics the real saira TOML: top-level users, keycloak, and a deeply nested system.
	// Definition orders are explicit (the parser always produces them) so the
	// CTX-02 chain set below is deterministic.
	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		UnitOrder:  []string{"webUser", "sshUser", "adminUser", "keycloak", "linuxSystem"},
		Units: map[string]*model.Unit{
			"webUser": {
				Type: model.TypePersonExternal,
				Name: "Web User",
				Links: []model.Link{
					{Peer: "linuxSystem.localIDP.grpcAPIs.authAPI"},
					{Peer: "linuxSystem.localIDP.grpcAPIs.sessionAPI"},
				},
			},
			"sshUser": {
				Type: model.TypePersonExternal,
				Name: "SSH User",
				Links: []model.Link{
					{Peer: "linuxSystem.sshAuth.sshd"},
				},
			},
			"adminUser": {
				Type: model.TypePersonExternal,
				Name: "Admin User",
				Links: []model.Link{
					{Peer: "linuxSystem.rbac"},
				},
			},
			"keycloak": {
				Type:       model.TypeSystemExternal,
				Name:       "Keycloak",
				Technology: "External IDP",
			},
			"linuxSystem": {
				Type:         model.TypeSystem,
				Name:         "Linux System",
				Description:  "Linux server",
				SubunitOrder: []string{"sshAuth", "localIDP", "rbac"},
				Subunits: map[string]*model.Unit{
					"sshAuth": {
						Type:         model.TypeContainerBox,
						Name:         "SSH Auth",
						SubunitOrder: []string{"sshd", "nss"},
						Subunits: map[string]*model.Unit{
							"sshd": {
								Type: model.TypeContainer,
								Name: "SSHD",
								Links: []model.Link{
									{Peer: "linuxSystem.sshAuth.nss"},
								},
							},
							"nss": {
								Type: model.TypeContainer,
								Name: "NSS",
							},
						},
					},
					"localIDP": {
						Type:         model.TypeContainerBox,
						Name:         "Local IDP",
						SubunitOrder: []string{"grpcAPIs"},
						Subunits: map[string]*model.Unit{
							"grpcAPIs": {
								Type:         model.TypeComponentBox,
								Name:         "gRPC APIs",
								SubunitOrder: []string{"authAPI", "sessionAPI"},
								Subunits: map[string]*model.Unit{
									"authAPI": {
										Type: model.TypeComponent,
										Name: "Auth API",
										Links: []model.Link{
											{Peer: "keycloak"},
										},
									},
									"sessionAPI": {
										Type: model.TypeComponent,
										Name: "Session API",
									},
								},
							},
						},
					},
					"rbac": {
						Type: model.TypeContainer,
						Name: "RBAC",
					},
				},
			},
		},
	}

	v := view.GenerateC1View(m)
	require.NotNil(t, v)

	// The 5 top-level units plus exactly 8 CTX-02 chain entries for depicted
	// link targets: sshAuth+sshd (sshUser), localIDP+grpcAPIs+authAPI+sessionAPI
	// (webUser), rbac (adminUser), and — transitively, once sshd is itself
	// depicted — nss (sshd's own link resolves at its true endpoints).
	assert.Len(t, v.Units, 13, "C1 should have 5 top-level units + 8 chain entries, got %d: %v",
		len(v.Units), keys(v.Units))

	// Verify each expected top-level node is present
	assert.Contains(t, v.Units, "webUser")
	assert.Contains(t, v.Units, "sshUser")
	assert.Contains(t, v.Units, "adminUser")
	assert.Contains(t, v.Units, "keycloak")
	assert.Contains(t, v.Units, "linuxSystem")

	// CTX-02: the ancestor chains of depicted link targets are visible
	// entries (Units + VisiblePaths), staying in scope (no boundary flags).
	for _, chainPath := range []string{
		"linuxSystem.sshAuth", "linuxSystem.sshAuth.sshd", "linuxSystem.sshAuth.nss",
		"linuxSystem.localIDP", "linuxSystem.localIDP.grpcAPIs",
		"linuxSystem.localIDP.grpcAPIs.authAPI", "linuxSystem.localIDP.grpcAPIs.sessionAPI",
		"linuxSystem.rbac",
	} {
		entry := v.Units[chainPath]
		require.NotNil(t, entry, "chain entry %s must exist", chainPath)
		assert.True(t, v.VisiblePaths[chainPath], "chain entry %s must be a visible path", chainPath)
		assert.False(t, entry.IsBoundary, "chain entry %s stays in scope", chainPath)
		assert.False(t, entry.IsExternal, "chain entry %s stays in scope", chainPath)
	}

	// A depicted chain entry carries its own links at their true endpoints:
	// the depicted sshd's link to nss resolves sshd -> nss (no collapse to a
	// remote ancestor), while units never scanned as sources keep nil.
	if assert.NotNil(t, v.Units["linuxSystem.sshAuth.sshd"].ResolvedLinks) {
		require.Len(t, v.Units["linuxSystem.sshAuth.sshd"].ResolvedLinks, 1)
		assert.Equal(t, "linuxSystem.sshAuth.nss", v.Units["linuxSystem.sshAuth.sshd"].ResolvedLinks[0].Peer)
	}

	assert.Nil(t, v.Units["linuxSystem"].ResolvedLinks)

	// linuxSystem should have [+] indicator and the chain-unfold marker
	assert.True(t, v.Units["linuxSystem"].HasSubunits)
	assert.True(t, v.Units["linuxSystem"].UnfoldChain)
}

// TestBuildGraphExpandedC1VisibleSubunitEdges verifies at graph level that
// visible subunits of an expanded C1 unit render inside the parent cluster
// (skipped as top-level nodes — duplicate node IDs would break DOT), and that
// resolved edges terminate at the TRUE target (CTX-02) — the chain entry
// inside the visible subunit's nested cluster. CTX-03: a visible subunit WITH
// subunits (sshAuth) renders as a NESTED cluster inside the parent cluster —
// not as a flat node.
func TestBuildGraphExpandedC1VisibleSubunitEdges(t *testing.T) {
	t.Parallel()

	v := view.GenerateC1View(expandedC1Model(nil))
	require.NotNil(t, v)

	g := graph.BuildGraph(v)
	require.NotNil(t, g)

	// Skip logic: the visible subunit is NOT a top-level node
	for _, node := range g.Nodes {
		assert.NotEqual(t, sshAuthPath, node.ID, "visible subunit must not render as top-level node")
	}

	// The cluster renders the visible subunit — as a nested cluster (CTX-03),
	// since sshAuth itself has subunits.
	require.Len(t, g.Clusters, 1)
	cluster := g.Clusters[0]
	assert.Equal(t, linuxSystemPath, cluster.ID)

	var sshAuthCluster *graph.Cluster

	for _, nested := range cluster.Clusters {
		if nested.ID == sshAuthPath {
			sshAuthCluster = nested
		}
	}

	require.NotNil(t, sshAuthCluster, "cluster must contain nested cluster "+sshAuthPath)
	assert.NotEmpty(t, sshAuthCluster.Nodes, "the nested sshAuth cluster unfolds its children")

	// D-07 as refined by CTX-02: the edge terminates at the TRUE target —
	// the chain entry inside the nested sshAuth cluster — not the parent.
	edgeFound := false

	for _, edge := range g.Edges {
		if edge.Source == webUserPath && edge.Target == sshAuthPath+".sshd" {
			edgeFound = true
		}
	}

	assert.True(t, edgeFound, "expected edge webUser -> "+sshAuthPath+".sshd")

	// RenderDOT must succeed — duplicate node IDs would break rendering
	dot, err := render.RenderDOT(g)
	require.NoError(t, err)
	assert.NotEmpty(t, dot)
}

// TestBuildGraphExpandedC1VisibleSubunitExploreLink documents the side effect
// of visible-subunit entries in v.Units: BuildGraphWithPath can now look up
// the entry when deciding explore links, so a visible subunit WITH subunits
// (sshAuth) gets an ExploreURL for drill-down to its C2 diagram (which
// auto-generation produces since it has subunits). CTX-03: the subunit renders
// as a nested cluster, so the explore URL lives on Cluster.ExploreURL.
func TestBuildGraphExpandedC1VisibleSubunitExploreLink(t *testing.T) {
	t.Parallel()

	v := view.GenerateC1View(expandedC1Model(nil))
	require.NotNil(t, v)

	g := graph.BuildGraphWithPath(v, "", "diagram", "svg")
	require.NotNil(t, g)
	require.Len(t, g.Clusters, 1)

	var sshAuthCluster *graph.Cluster

	for _, nested := range g.Clusters[0].Clusters {
		if nested.ID == sshAuthPath {
			sshAuthCluster = nested
		}
	}

	require.NotNil(t, sshAuthCluster, "cluster must contain nested cluster "+sshAuthPath)
	assert.NotEmpty(t, sshAuthCluster.ExploreURL,
		"the nested sshAuth cluster keeps its drill-down explore URL")
}

// keys returns the keys of a map for error messages.
func keys(m map[string]*view.Entry) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}

	return result
}

// TestIntegrationC1EdgeResolution verifies that edges in C1 are resolved to top-level ancestors.
func TestIntegrationC1EdgeResolution(t *testing.T) {
	t.Parallel()

	m := &parser.Model{
		Properties: model.Properties{Name: "Test"},
		Units: map[string]*model.Unit{
			"webUser": {
				Type: model.TypePersonExternal,
				Name: "Web User",
				// Link targets a deeply nested component
				Links: []model.Link{
					{Peer: "system.api.handler", Technology: "HTTP"},
				},
			},
			"system": {
				Type: model.TypeSystem,
				Name: "System",
				Subunits: map[string]*model.Unit{
					"api": {
						Type: model.TypeContainer,
						Name: "API",
						Subunits: map[string]*model.Unit{
							"handler": {
								Type: model.TypeComponent,
								Name: "Handler",
								Links: []model.Link{
									{Peer: "externaldb"},
								},
							},
						},
					},
				},
			},
		},
	}

	v := view.GenerateC1View(m)
	require.NotNil(t, v)

	// 2 top-level units (webUser, system), the external boundary node
	// (externaldb), and the CTX-02 chain entries (system.api,
	// system.api.handler) for the depicted deep-link target.
	assert.Len(t, v.Units, 5)

	// webUser's link terminates at the TRUE target (CTX-02) — the chain entry
	// added beneath the collapsed system ancestor.
	if assert.NotNil(t, v.Units["webUser"].ResolvedLinks) {
		assert.Equal(t, "system.api.handler", v.Units["webUser"].ResolvedLinks[0].Peer)
	}

	// handler is now a depicted chain entry, so ITS externaldb link is
	// recorded on the handler entry (true source), not the collapsed system.
	if assert.NotNil(t, v.Units["system.api.handler"].ResolvedLinks) {
		assert.Equal(t, "externaldb", v.Units["system.api.handler"].ResolvedLinks[0].Peer)
	}

	assert.Nil(t, v.Units["system"].ResolvedLinks)
}
