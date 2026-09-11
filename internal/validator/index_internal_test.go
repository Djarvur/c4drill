package validator

// Internal test file: TestMirrorOrderDeterministicAcrossRuns exercises the
// unexported populateIncomingLinks directly, so it stays in package validator
// per the testpackage skip-regexp convention in .golangci.yml. Exported-API
// tests for BuildIndex live in index_test.go (package validator_test).

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/stretchr/testify/require"
)

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
