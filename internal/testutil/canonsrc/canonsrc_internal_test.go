// canonsrc internal tests: the error contract behind the t.Fatalf wrappers.
// NormalizeTOML/NormalizeC4D fail the test on malformed input (the
// canonical.Canonical contract); the underlying helpers return the error so
// the failure path is unit-testable without a failing *testing.T.
package canonsrc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeHelpersRejectMalformedInput(t *testing.T) {
	t.Parallel()

	t.Run("toml", func(t *testing.T) {
		t.Parallel()

		out, err := normalizeTOML("[unclosed")
		require.Error(t, err, "malformed TOML errors")
		assert.Empty(t, out, "no canonical output on error")
	})

	t.Run("c4d", func(t *testing.T) {
		t.Parallel()

		out, err := normalizeC4D("a: system {")
		require.Error(t, err, "malformed C4D errors")
		assert.Empty(t, out, "no canonical output on error")
	})
}
