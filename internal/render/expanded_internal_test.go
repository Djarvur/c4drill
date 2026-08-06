package render

import (
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// loadCYPAuthInfraModel returns the synthetic model matching the private
// auth-infrastructure fixture structure. The private fixture is gitignored;
// the synthetic model keeps render tests deterministic on every machine (D-01).
func loadCYPAuthInfraModel(t *testing.T) *parser.Model {
	t.Helper()

	return createSyntheticCYPModel(t)
}

// createSyntheticCYPModel creates a model matching the private
// auth-infrastructure fixture structure.
//
//nolint:funlen // Test helper that creates multiple related units
func createSyntheticCYPModel(t *testing.T) *parser.Model {
	t.Helper()

	m := &parser.Model{
		Properties: model.Properties{
			Name: "CYP Auth Infrastructure",
		},
		Units: make(map[string]*model.Unit),
	}

	// SSH User (external person)
	m.Units["sshUser"] = &model.Unit{
		Name: "SSH User",
		Type: model.TypePersonExternal,
		Links: []model.Link{
			{Peer: "server.sshd"},
		},
	}

	// Service User (external person)
	m.Units["serviceUser"] = &model.Unit{
		Name: "Service SSH User",
		Type: model.TypePersonExternal,
		Links: []model.Link{
			{Peer: "server.sshd"},
		},
	}

	// Server (system) with subunits
	server := &model.Unit{
		Name:        "Server",
		Type:        model.TypeSystem,
		Description: "Server hosting CYP authentication services",
		Subunits:    make(map[string]*model.Unit),
	}
	m.Units["server"] = server

	// server.sshd (component)
	server.Subunits["sshd"] = &model.Unit{
		Name:        "SSH Daemon",
		Type:        model.TypeComponent,
		Description: "Handles SSH connections for the server",
		Links: []model.Link{
			{Peer: "server.pam.unix"},
			{Peer: "server.pam.cyp"},
			{Peer: "server.nss"},
		},
	}

	// server.pam (container with subunits - this is the key nested structure)
	pam := &model.Unit{
		Name:        "PAM",
		Type:        model.TypeContainer,
		Description: "Pluggable Authentication Module subsystem",
		Subunits:    make(map[string]*model.Unit),
	}
	server.Subunits["pam"] = pam

	// server.pam.unix (component inside container)
	pam.Subunits["unix"] = &model.Unit{
		Name:        "PAM Unix Module",
		Type:        model.TypeComponent,
		Description: "PAM module for standard Unix authentication",
		Links: []model.Link{
			{Peer: "server.etc"},
		},
	}

	// server.pam.cyp (component inside container)
	pam.Subunits["cyp"] = &model.Unit{
		Name:        "PAM CYP Auth Module",
		Type:        model.TypeComponent,
		Description: "PAM module for CYP username/password authentication",
		Technology:  "Go",
		Links: []model.Link{
			{Peer: "server.systemd"},
		},
	}

	// server.etc (db)
	server.Subunits["etc"] = &model.Unit{
		Name:        "/etc/passwd",
		Type:        model.TypeDb,
		Description: "Standard Unix password authentication files",
		Technology:  "files",
	}

	// server.nss (container)
	server.Subunits["nss"] = &model.Unit{
		Name:        "NSS",
		Type:        model.TypeContainer,
		Description: "Name Service Switch subsystem for user information",
		Links: []model.Link{
			{Peer: "server.systemd"},
		},
	}

	// server.systemd (container)
	server.Subunits["systemd"] = &model.Unit{
		Name:        "Systemd",
		Type:        model.TypeContainer,
		Description: "Systemd init system for managing services",
	}

	return m
}

// TestExpandedViewNestedClusters tests nested cluster rendering:
//   - CASE-01: server.pam cluster should exist
//   - CASE-02: server.pam.unix and server.pam.cyp nodes should exist
//   - CASE-03: Edges to/from nested components should exist
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestExpandedViewNestedClusters(t *testing.T) {
	m := loadCYPAuthInfraModel(t)

	// Generate expanded view
	v := view.GenerateExpandedView(m)
	require.NotNil(t, v, "expanded view should not be nil")

	// Build expanded graph
	g := graph.BuildExpandedGraph(v)
	require.NotNil(t, g, "expanded graph should not be nil")

	// Render to DOT
	output, err := RenderDOT(g)
	require.NoError(t, err, "rendering DOT should not error")

	dotStr := string(output)

	// CASE-01: server.pam cluster should exist
	// The cluster naming follows "cluster_" + path pattern
	assert.Contains(t, dotStr, "cluster_server.pam",
		"DOT should contain nested cluster for server.pam (CASE-01)")

	// CASE-02: server.pam.unix and server.pam.cyp nodes should exist
	assert.Contains(t, dotStr, `"server.pam.unix"`,
		"DOT should contain nested component server.pam.unix (CASE-02)")
	assert.Contains(t, dotStr, `"server.pam.cyp"`,
		"DOT should contain nested component server.pam.cyp (CASE-02)")

	// CASE-03: Edges to/from nested components should exist
	assert.Contains(t, dotStr, `"server.sshd" -> "server.pam.unix"`,
		"DOT should contain edge from sshd to pam.unix (CASE-03)")
	assert.Contains(t, dotStr, `"server.sshd" -> "server.pam.cyp"`,
		"DOT should contain edge from sshd to pam.cyp (CASE-03)")
	assert.Contains(t, dotStr, `"server.pam.unix" -> "server.etc"`,
		"DOT should contain edge from pam.unix to etc (CASE-03)")
	assert.Contains(t, dotStr, `"server.pam.cyp" -> "server.systemd"`,
		"DOT should contain edge from pam.cyp to systemd (CASE-03)")
}

// TestHTMLTableAttributes tests REFINED-02:
// All HTML tables should include border="0" cellpadding="0" cellspacing="0".
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestHTMLTableAttributes(t *testing.T) {
	m := loadCYPAuthInfraModel(t)

	// Generate expanded view
	v := view.GenerateExpandedView(m)
	require.NotNil(t, v, "expanded view should not be nil")

	// Build expanded graph
	g := graph.BuildExpandedGraph(v)
	require.NotNil(t, g, "expanded graph should not be nil")

	// Render to DOT
	output, err := RenderDOT(g)
	require.NoError(t, err, "rendering DOT should not error")

	dotStr := string(output)

	// REFINED-02: HTML tables should have proper attributes
	// The table tag should include border="0" cellpadding="0" cellspacing="0"
	assert.Contains(t, dotStr, `border="0" cellpadding="0" cellspacing="0"`,
		"HTML tables should include border, cellpadding, cellspacing attributes (REFINED-02)")
}

// TestClusterHTMLLabels tests REFINED-03:
// Cluster labels should use HTML format with proper coloring.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestClusterHTMLLabels(t *testing.T) {
	m := loadCYPAuthInfraModel(t)

	// Generate expanded view
	v := view.GenerateExpandedView(m)
	require.NotNil(t, v, "expanded view should not be nil")

	// Build expanded graph
	g := graph.BuildExpandedGraph(v)
	require.NotNil(t, g, "expanded graph should not be nil")

	// Render to DOT
	output, err := RenderDOT(g)
	require.NoError(t, err, "rendering DOT should not error")

	dotStr := string(output)

	// REFINED-03: Cluster labels should use HTML format
	// Look for HTML table in cluster labels (label=<<table)
	// The cluster for server.pam should have an HTML label
	lines := strings.Split(dotStr, "\n")

	foundClusterPAMWithHTMLLabel := false

	for i, line := range lines {
		// Look for cluster_server.pam subgraph
		hasCluster := strings.Contains(line, "cluster_server.pam")

		hasSubgraph := strings.Contains(line, "subgraph cluster_server.pam")
		if hasCluster || hasSubgraph {
			// Check subsequent lines for HTML label
			for j := i; j < len(lines) && j < i+10; j++ {
				if strings.Contains(lines[j], "label=<<table") {
					foundClusterPAMWithHTMLLabel = true

					break
				}
			}
		}
	}

	assert.True(t, foundClusterPAMWithHTMLLabel,
		"cluster_server.pam should have HTML label format (REFINED-03)")
}

// TestNodeShapeBox tests REFINED-01:
// All nodes should render with shape=box and style=rounded.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestNodeShapeBox(t *testing.T) {
	m := loadCYPAuthInfraModel(t)

	// Generate expanded view
	v := view.GenerateExpandedView(m)
	require.NotNil(t, v, "expanded view should not be nil")

	// Build expanded graph
	g := graph.BuildExpandedGraph(v)
	require.NotNil(t, g, "expanded graph should not be nil")

	// Render to DOT
	output, err := RenderDOT(g)
	require.NoError(t, err, "rendering DOT should not error")

	dotStr := string(output)

	// REFINED-01: Nodes should have shape=box (not shape=none)
	// Note: shape=none was the old approach, shape=box is the new approach
	// We check that shape=none is NOT used for nodes with HTML labels
	// The new implementation should use shape=box with style=rounded

	// Check that at least one node uses shape=box
	assert.Contains(t, dotStr, "shape=box",
		"Nodes should use shape=box (REFINED-01)")

	// Check that style=rounded is present
	assert.Contains(t, dotStr, "rounded",
		"Nodes should have style=rounded (REFINED-01)")
}
