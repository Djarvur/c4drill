// Package peer rewrites relative (bare) Link.Peer values into absolute
// dotted paths on an already-parsed model, before validation.
//
// Phase 30 implements ERGO-01 (relative resolution) and ERGO-02 (absolute
// fallback + backward-compat). The four load-bearing decisions are:
//
//   - D-13: the "enclosing parent" searched first for a bare peer is the
//     host unit's IMMEDIATE parent. For a unit at path a.b.c, peers search
//     a.b's children first (c's siblings), NOT c's own children.
//   - D-14: when no sibling matches, the resolver walks UP the ancestry
//     nearest-first (a.b's children, then a's children, then root). The
//     first depth with a match wins. Cross-depth shadowing is silent.
//   - D-15: the walk-up REACHES ROOT — top-level units are the outermost
//     scope, so a bare peer resolves to a top-level unit if one exists,
//     regardless of depth.
//   - D-16: unified gate. A peer containing "." is absolute (untouched);
//     a bare peer runs the walk-up. A miss at root is a hard *ResolveError.
//     There is NO separate top-level short-circuit: the root step of the
//     walk-up produces identical results for bare peers matching top-level
//     units, preserving the backward-compat hard contract.
//
// The pass is a pure function: no I/O, no globals, stdlib only. It mutates
// Link.Peer in place and changes no model structure. It runs AFTER Parse
// (and, in future milestones, AFTER template.Expand) and BEFORE Validate,
// so the validator's absolute-path logic is untouched and LinksFrom still
// holds only authored entries (validator-synthesized mirrors are added
// later by populateIncomingLinks).
package peer

import (
	"fmt"
	"strings"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
)

// Resolve rewrites every Link.Peer from a relative bare name to an absolute
// dotted path across all units (top-level + nested subunits), for both
// outgoing Links and authored LinksFrom. It is nil-safe (Resolve(nil) is a
// no-op) and fails fast on the first unresolvable peer.
//
// Returns *ResolveError naming the peer and host unit when a bare peer
// matches no ancestor scope's children all the way to root.
func Resolve(m *parser.Model) error {
	if m == nil {
		return nil
	}

	return resolveUnits(m.Units, "", m)
}

// resolveUnits walks the unit tree mirroring validator.BuildIndex's recursion
// (internal/validator/index.go:24-43): same fullPath construction, same
// Subunits descent. At each unit it rewrites that unit's links before
// descending, so a peer on a parent never depends on a child's resolution.
func resolveUnits(units map[string]*model.Unit, parentPath string, m *parser.Model) error {
	for name, unit := range units {
		fullPath := name
		if parentPath != "" {
			fullPath = parentPath + "." + name
		}

		if err := resolveUnitLinks(unit, fullPath, m); err != nil {
			return err
		}

		if len(unit.Subunits) > 0 {
			if err := resolveUnits(unit.Subunits, fullPath, m); err != nil {
				return err
			}
		}
	}

	return nil
}

// resolveUnitLinks rewrites the Peer field of every Link in unit.Links and
// unit.LinksFrom in place. Index assignment (unit.Links[i].Peer = ...) is
// required rather than a range variable because a Go range copies the slice
// element — `for _, l := range unit.Links { l.Peer = ... }` would mutate the
// copy and silently drop every rewrite.
func resolveUnitLinks(unit *model.Unit, fullPath string, m *parser.Model) error {
	for i := range unit.Links {
		resolved, err := resolvePeer(fullPath, unit.Links[i].Peer, m)
		if err != nil {
			return err
		}

		unit.Links[i].Peer = resolved
	}

	for i := range unit.LinksFrom {
		resolved, err := resolvePeer(fullPath, unit.LinksFrom[i].Peer, m)
		if err != nil {
			return err
		}

		unit.LinksFrom[i].Peer = resolved
	}

	return nil
}

// resolvePeer applies the D-16 unified gate. A dotted peer is absolute and
// returned untouched. A bare peer is resolved by walking the host's ancestor
// scopes nearest-first (D-13/D-14/D-15); the first scope whose children-map
// contains the peer wins, and the absolute path is constructed from that
// scope's parent path. A miss at root yields *ResolveError.
//
// Note on ROADMAP criterion 3 (same-depth ambiguity = hard error): under the
// walk-up model this case is structurally UNREACHABLE. Each scope is a single
// parent's children-map (model.Unit.Subunits is map[string]*Unit, or the
// top-level m.Units of the same shape), and map keys are unique per parent.
// A bare name therefore matches at most one child per scope. The defensive
// ambiguity branch is omitted as dead code (Plan 30 RESEARCH.md Pitfall 3);
// cross-depth "ambiguity" is not an error — nearest-first handles it silently.
func resolvePeer(hostPath, peer string, m *parser.Model) (string, error) {
	// D-16 step 1: absolute peers (containing ".") are untouched.
	if strings.Contains(peer, ".") {
		return peer, nil
	}

	// D-13/D-14/D-15: walk ancestor scopes nearest-first.
	for _, sc := range ancestorScopes(hostPath, m) {
		if sc.children == nil {
			continue
		}

		if _, ok := sc.children[peer]; ok {
			// Top-level match (root scope, empty parent path) is an identity
			// rewrite — the peer already equals its absolute form.
			if sc.parentPath == "" {
				return peer, nil
			}

			return sc.parentPath + "." + peer, nil
		}
	}

	return "", &ResolveError{Peer: peer, Host: hostPath}
}

// scope pairs a children-map with the dotted path of its parent, so
// resolvePeer can construct the absolute peer path when a match is found.
type scope struct {
	parentPath string
	children   map[string]*model.Unit
}

// ancestorScopes returns the host's ancestor scopes nearest-first, each as
// the children-map of an ancestor plus that ancestor's dotted path. The list
// always ends with the root scope (m.Units, empty parent path) per D-15.
//
// For hostPath "a.b.c" the result is:
//
//	[{a.b, a.b.Subunits}, {a, a.Subunits}, {"", m.Units}]
//
// i.e. immediate parent's children first, then grandparent's, ..., then root.
// The host's OWN Subunits are never included (D-13: start at immediate parent).
// For a top-level host ("x") the result is just [{"", m.Units}].
//
// If the hostPath does not resolve in the tree (should not happen for a model
// produced by parser.Parse, which only creates units at paths it parsed), the
// descent stops early and returns the scopes collected so far plus root.
func ancestorScopes(hostPath string, m *parser.Model) []scope {
	segments := strings.Split(hostPath, ".")
	if hostPath == "" {
		return []scope{{parentPath: "", children: m.Units}}
	}

	// Descend from root toward the host's parent, collecting each ancestor's
	// children-map. descended is grandparent-first; we reverse below.
	descended := make([]scope, 0, len(segments))

	current := m.Units
	accPath := ""

	// All segments except the last name ancestors of the host.
	for i := range len(segments) - 1 {
		seg := segments[i]
		unit, ok := current[seg]
		if !ok {
			break // path not in tree; stop descent, keep what we have
		}

		ancestorPath := seg
		if accPath != "" {
			ancestorPath = accPath + "." + seg
		}

		descended = append(descended, scope{parentPath: ancestorPath, children: unit.Subunits})
		accPath = ancestorPath
		current = unit.Subunits
	}

	// Nearest-first: reverse the descended scopes, then append root.
	out := make([]scope, 0, len(descended)+1)
	for i := len(descended) - 1; i >= 0; i-- {
		out = append(out, descended[i])
	}

	out = append(out, scope{parentPath: "", children: m.Units})

	return out
}

// ResolveError names a bare peer that could not be resolved against any
// ancestor scope of its host unit, and the host unit's dotted path. It
// follows the *parser.ParseError struct-with-Error idiom
// (internal/parser/errors.go). Author-facing diagnostic data only — no
// secrets cross this boundary (T-30-03).
type ResolveError struct {
	// Peer is the unresolvable bare peer value.
	Peer string
	// Host is the dotted path of the unit that declared the link.
	Host string
}

// Error returns a human-readable diagnostic naming the peer and host.
func (e *ResolveError) Error() string {
	return fmt.Sprintf("cannot resolve peer %q from unit %q", e.Peer, e.Host)
}
