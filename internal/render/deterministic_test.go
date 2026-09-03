package render_test

import (
	"fmt"
	"testing"

	"github.com/Djarvur/c4drill/internal/c4d"
	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/peer"
	"github.com/Djarvur/c4drill/internal/render"
	"github.com/Djarvur/c4drill/internal/validator"
	"github.com/Djarvur/c4drill/internal/view"
	"github.com/stretchr/testify/require"
)

// Note: no t.Parallel() — the go-graphviz library uses a WASM-based rendering
// engine that has concurrency issues (same convention as integration_test.go).

// issue42Model is the issue #42 reproducer model, translated deterministically
// from the issue's Python generator (D-05, REPRO-01/REPRO-03): 10 containers
// c0..c9 inside sys: system "S", where container ci links -> sys.cj for every
// j != i with (i+j)%3 == 0, plus ext: personExternal "User" linking -> sys.c0.
// The many shared targets per container force validator-synthesized LinksFrom
// mirrors whose order used to follow Go map iteration, permuting edge insertion
// order and therefore GraphViz's generated id="edge<N>" group ids in SVG.
// Inline const content — deterministic, no randomness, no testdata file.
const issue42Model = `sys: system "S" {
  c0: container "C0" {
    -> sys.c3
    -> sys.c6
    -> sys.c9
  }
  c1: container "C1" {
    -> sys.c2
    -> sys.c5
    -> sys.c8
  }
  c2: container "C2" {
    -> sys.c1
    -> sys.c4
    -> sys.c7
  }
  c3: container "C3" {
    -> sys.c0
    -> sys.c6
    -> sys.c9
  }
  c4: container "C4" {
    -> sys.c2
    -> sys.c5
    -> sys.c8
  }
  c5: container "C5" {
    -> sys.c1
    -> sys.c4
    -> sys.c7
  }
  c6: container "C6" {
    -> sys.c0
    -> sys.c3
    -> sys.c9
  }
  c7: container "C7" {
    -> sys.c2
    -> sys.c5
    -> sys.c8
  }
  c8: container "C8" {
    -> sys.c1
    -> sys.c4
    -> sys.c7
  }
  c9: container "C9" {
    -> sys.c0
    -> sys.c3
    -> sys.c6
  }
}
ext: personExternal "User" {
  -> sys.c0
}
`

// runIssue42Pipeline performs one full CLI-equivalent pipeline run over the
// issue #42 reproducer: parse -> peer.Resolve -> validator.Validate ->
// GenerateC2View -> BuildGraphWithPath -> render.Render for the given format.
// The BuildGraphWithPath arguments mirror cmd/c4drill/root.go processView for
// the top-level drill-down render (unitPath "sys" selects the C2 view of the
// top system, whose visible containers c0..c9 are where the permuted
// validator-synthesized LinksFrom mirrors reach edge insertion order; the
// collapsed C1 view hides them behind link resolution and is deterministic
// even pre-fix). basename "issue42" is the output basename the CLI would
// derive for the entry file.
func runIssue42Pipeline(t *testing.T, format string) []byte {
	t.Helper()

	m, err := c4d.Parse([]byte(issue42Model))
	require.NoError(t, err, "c4d.Parse on issue #42 reproducer model")

	require.NoError(t, peer.Resolve(m), "peer.Resolve on issue #42 reproducer model")
	require.Empty(t, validator.Validate(m), "issue #42 reproducer model must validate clean")

	v := view.GenerateC2View(m, "sys")
	require.NotNil(t, v, "GenerateC2View should return a view")

	g := graph.BuildGraphWithPath(v, "sys", "issue42", format)
	require.NotNil(t, g, "BuildGraphWithPath should return a graph")

	data, err := render.Render(g, format)
	require.NoError(t, err, "render.Render(%q) should not return error", format)
	require.NotEmpty(t, data, "render.Render(%q) should return non-empty bytes", format)

	return data
}

// firstDiffByte returns the index of the first differing byte between a and b,
// plus the total lengths, for failure diagnostics.
func firstDiffByte(a, b []byte) string {
	n := len(a)

	if len(b) < n {
		n = len(b)
	}

	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return fmt.Sprintf("first differing byte at index %d: got %q want %q (len got=%d want=%d)",
				i, b[i], a[i], len(b), len(a))
		}
	}

	return fmt.Sprintf("common prefix of %d bytes, lengths differ: got=%d want=%d", n, len(b), len(a))
}

// TestDeterministicByteIdenticalOutput is the D-05 regression test for issue
// #42 (REPRO-01/REPRO-02/REPRO-03): rendering the same model twice in-process
// must produce byte-identical output for svg, dot, and html — including the
// GraphViz-generated id="edge<N>" group ids in SVG, which follow cgraph edge
// insertion order. The svg assertion is the RED carrier (dot may already be
// stable — that asymmetry is the documented issue #42 evidence).
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestDeterministicByteIdenticalOutput(t *testing.T) {
	for _, format := range []string{"svg", "dot", "html"} {
		t.Run(format, func(t *testing.T) {
			first := runIssue42Pipeline(t, format)
			second := runIssue42Pipeline(t, format)

			if format == "svg" && string(first) != string(second) {
				// Annotate the RED/GREEN diagnosis with the exact divergence point.
				t.Logf("svg outputs differ across runs: %s", firstDiffByte(first, second))
			}

			require.Equal(t, first, second,
				"rendering the issue #42 model twice in-process must produce byte-identical %s output", format)
		})
	}
}
