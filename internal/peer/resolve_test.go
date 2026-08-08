// Package peer_test black-box tests internal/peer.Resolve (Phase 30).
//
// Resolve rewrites Link.Peer from relative bare names to absolute dotted
// paths per the locked decisions:
//   - D-13: enclosing parent is the immediate parent (a.b.c's peers search
//     a.b's children first)
//   - D-14: walk-up ancestry nearest-first (cross-depth shadowing is silent)
//   - D-15: the walk reaches root (top-level units are the outermost scope)
//   - D-16: unified gate — a peer containing "." is absolute; a bare peer
//     runs the walk-up; a miss at root is a hard *ResolveError
package peer_test

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResolveSibling proves D-13: a bare peer on a nested host resolves to a
// sibling under the same immediate parent.
func TestResolveSibling(t *testing.T) {
	t.Parallel()

	const toml = `
[mainSystem]
type = "system"

[mainSystem.localIDP]
type = "container"

[mainSystem.localIDP.sessionAPI]
type = "component"
link = [{peer = "sessionManager"}]

[mainSystem.localIDP.sessionManager]
type = "component"
`
	m := mustParse(t, toml)
	require.NoError(t, peer.Resolve(m))

	link := mustFindLink(t, m, "mainSystem.localIDP.sessionAPI", "sessionManager", "mainSystem.localIDP.sessionManager")
	assert.Equal(t, "mainSystem.localIDP.sessionManager", link.Peer,
		"sibling bare peer must rewrite to immediate-parent's child path")
}

// TestResolveWalkUpAunt proves D-14: when no sibling matches, the resolver
// walks up to ancestor scopes and the first match (aunt) wins.
func TestResolveWalkUpAunt(t *testing.T) {
	t.Parallel()

	const toml = `
[frontend]
type = "system"

[frontend.cache]
type = "container"

[frontend.api]
type = "container"

[frontend.api.handlers]
type = "component"

[frontend.api.handlers.auth]
type = "component"
link = [{peer = "cache"}]
`
	m := mustParse(t, toml)
	require.NoError(t, peer.Resolve(m))

	link := mustFindLink(t, m, "frontend.api.handlers.auth", "cache", "frontend.cache")
	assert.Equal(t, "frontend.cache", link.Peer,
		"aunt bare peer must rewrite to the nearest ancestor scope's matching child")
}

// TestResolveRoot proves D-15: the walk-up reaches the top-level (root) scope.
func TestResolveRoot(t *testing.T) {
	t.Parallel()

	const toml = `
[messageBus]
type = "queue"

[linuxSystem]
type = "system"

[linuxSystem.sshAuth]
type = "container"

[linuxSystem.sshAuth.sshd]
type = "component"
link = [{peer = "messageBus"}]
`
	m := mustParse(t, toml)
	require.NoError(t, peer.Resolve(m))

	link := mustFindLink(t, m, "linuxSystem.sshAuth.sshd", "messageBus", "messageBus")
	assert.Equal(t, "messageBus", link.Peer,
		"root bare peer must rewrite to the top-level unit's path (identity at root)")
}

// TestResolveNearestFirst proves D-14 cross-depth precedence: when a bare
// peer matches children at two ancestor depths, the nearer scope wins
// silently (no error).
func TestResolveNearestFirst(t *testing.T) {
	t.Parallel()

	const toml = `
[a]
type = "system"

[a.x]
type = "container"

[a.b]
type = "container"

[a.b.x]
type = "component"

[a.b.c]
type = "component"
link = [{peer = "x"}]
`
	m := mustParse(t, toml)
	require.NoError(t, peer.Resolve(m))

	link := mustFindLink(t, m, "a.b.c", "x", "a.b.x")
	assert.Equal(t, "a.b.x", link.Peer,
		"nearest-first: nearer ancestor's child (a.b.x) must win over farther (a.x), no error")
}

// TestResolveDottedUntouched proves D-16 step 1: a peer containing "." is
// absolute and left unchanged.
func TestResolveDottedUntouched(t *testing.T) {
	t.Parallel()

	const toml = `
[anything]
type = "system"
link = [{peer = "already.dotted.path"}]

[already]
type = "system"

[already.dotted]
type = "container"

[already.dotted.path]
type = "component"
`
	m := mustParse(t, toml)
	require.NoError(t, peer.Resolve(m))

	link := mustFindLink(t, m, "anything", "already.dotted.path", "already.dotted.path")
	assert.Equal(t, "already.dotted.path", link.Peer,
		"dotted peer must be left untouched (absolute, D-16 step 1)")
}

// TestResolveUnresolvableError proves the D-16 miss-at-root hard error: a
// bare peer matching no ancestor scope yields *ResolveError naming the peer
// and the host unit.
func TestResolveUnresolvableError(t *testing.T) {
	t.Parallel()

	const toml = `
[a]
type = "system"

[a.b]
type = "container"

[a.b.c]
type = "component"
link = [{peer = "x"}]
`
	m := mustParse(t, toml)
	err := peer.Resolve(m)
	require.Error(t, err)

	var resolveErr *peer.ResolveError
	require.ErrorAs(t, err, &resolveErr, "error must unwrap to *peer.ResolveError")
	assert.Equal(t, "x", resolveErr.Peer, "ResolveError must name the unresolvable peer")
	assert.Equal(t, "a.b.c", resolveErr.Host, "ResolveError must name the host unit")
}

// TestResolveLinksFrom proves authored linkFrom peers are rewritten just like
// outgoing links (ERGO-01/02 uniformity).
func TestResolveLinksFrom(t *testing.T) {
	t.Parallel()

	const toml = `
[host]
type = "system"

[host.sibling]
type = "container"

[host.target]
type = "container"

[[host.sibling.linkFrom]]
peer = "target"
`
	m := mustParse(t, toml)
	require.NoError(t, peer.Resolve(m))

	// host.sibling.linkFrom[0].peer "target" should rewrite to host.target
	found := false
	walkUnits(t, m, func(unitPath string, unit *model.Unit) {
		if unitPath != "host.sibling" {
			return
		}
		for _, lf := range unit.LinksFrom {
			if lf.Mirror {
				continue // skip validator-synthesized mirrors; Resolve runs before they exist anyway
			}
			if lf.Peer == "host.target" || lf.Peer == "target" {
				found = true
				assert.Equal(t, "host.target", lf.Peer,
					"authored linkFrom bare peer must rewrite like an outgoing link")
			}
		}
	})
	assert.True(t, found, "expected an authored linkFrom on host.sibling with peer target/host.target")
}

// TestResolveCorpusByteIdentical proves ERGO-02 backward-compat at the unit
// level: the parser-corpus fixtures (root testdata/) have a byte-identical
// (unitPath, link.Peer) set before and after Resolve. These fixtures' bare
// peers all reference top-level units, so the rewrite is an identity and the
// set is unchanged. (The cmd/c4drill/testdata corpus is covered end-to-end by
// Plan 02's TestCLICorpusRendersUnchanged.)
func TestResolveCorpusByteIdentical(t *testing.T) {
	t.Parallel()

	fixtures := []string{"valid.toml", "links.toml", "nested.toml"}

	for _, fix := range fixtures {
		fix := fix
		t.Run(fix, func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(filepath.Join("..", "..", "testdata", fix))
			require.NoError(t, err, "failed to read parser-corpus fixture %s", fix)

			m, err := parser.Parse(data)
			require.NoError(t, err, "parser-corpus fixture %s must parse cleanly", fix)

			before := collectPeerSet(t, m)
			require.NoError(t, peer.Resolve(m), "Resolve must be a no-op error-wise on %s", fix)
			after := collectPeerSet(t, m)

			assert.Equal(t, before, after,
				"corpus fixture %s: (unitPath, peer) set must be byte-identical pre/post Resolve (ERGO-02)", fix)
		})
	}
}

// TestResolveNilSafe proves the nil-model guard (defensive; resolves without panic).
func TestResolveNilSafe(t *testing.T) {
	t.Parallel()

	require.NoError(t, peer.Resolve(nil), "Resolve(nil) must return nil (nil-safe)")
}

// TestResolveErrorFormat checks the human-readable error wording names both the
// peer and the host (T-30-03 author diagnostic).
func TestResolveErrorFormat(t *testing.T) {
	t.Parallel()

	e := &peer.ResolveError{Peer: "x", Host: "a.b.c"}
	msg := e.Error()
	assert.Contains(t, msg, `"x"`, "Error() must quote the peer name")
	assert.Contains(t, msg, `"a.b.c"`, "Error() must quote the host unit path")
}

// --- helpers ---

// mustParse parses inline TOML via the real parser (no shortcuts).
func mustParse(t *testing.T, toml string) *parser.Model {
	t.Helper()

	m, err := parser.Parse([]byte(toml))
	require.NoError(t, err, "inline test fixture must parse cleanly")
	require.NotNil(t, m)

	return m
}

// mustFindLink walks the model, finds the link with the given original peer
// on the given host path, and returns it. The expectedAfter value is used
// only for the failure message — the caller asserts the rewritten value.
func mustFindLink(t *testing.T, m *parser.Model, hostPath, originalPeer, expectedAfter string) *model.Link {
	t.Helper()

	var found *model.Link
	walkUnits(t, m, func(unitPath string, unit *model.Unit) {
		if unitPath != hostPath || found != nil {
			return
		}
		for i := range unit.Links {
			p := unit.Links[i].Peer
			if p == originalPeer || p == expectedAfter {
				found = &unit.Links[i]
				return
			}
		}
	})
	if found == nil {
		t.Fatalf("expected a link on %s with peer %q or %q; none found", hostPath, originalPeer, expectedAfter)
	}

	return found
}

// walkUnits walks m.Units + Subunits, mirroring validator.BuildIndex's
// recursion (internal/validator/index.go:24-43). unitPath is the dotted path.
func walkUnits(t *testing.T, m *parser.Model, fn func(unitPath string, unit *model.Unit)) {
	t.Helper()

	var walk func(units map[string]*model.Unit, parentPath string)
	walk = func(units map[string]*model.Unit, parentPath string) {
		// Stable order so failure messages are reproducible.
		names := make([]string, 0, len(units))
		for name := range units {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			unit := units[name]
			fullPath := name
			if parentPath != "" {
				fullPath = parentPath + "." + name
			}
			fn(fullPath, unit)
			if len(unit.Subunits) > 0 {
				walk(unit.Subunits, fullPath)
			}
		}
	}
	walk(m.Units, "")
}

// collectPeerSet builds the (unitPath, link.Peer) set across Links and
// authored LinksFrom for every unit in the tree. Returns a deterministic
// slice (sorted) so assert.Equal gives a readable diff on mismatch.
func collectPeerSet(t *testing.T, m *parser.Model) []string {
	t.Helper()

	var pairs []string
	walkUnits(t, m, func(unitPath string, unit *model.Unit) {
		for _, l := range unit.Links {
			pairs = append(pairs, unitPath+"\tLinks\t"+l.Peer)
		}
		for _, lf := range unit.LinksFrom {
			if lf.Mirror {
				continue
			}
			pairs = append(pairs, unitPath+"\tLinksFrom\t"+lf.Peer)
		}
	})
	sort.Strings(pairs)

	return pairs
}
