package canonical_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/Djarvur/c4drill/internal/testutil/canonical"
)

func TestCanonicalDOTPreservesLastAttribute(t *testing.T) {
	t.Parallel()

	// WR-01 regression: the attribute block was sliced with an off-by-one that
	// dropped the final character of the last attribute, canonicalizing
	// style=solid as style=soli. The full value must survive.
	dot := "digraph G {\n" +
		"\t\"a\" -> \"b\" [key=\"a_to_b\",\n" +
		"\t\tstyle=solid];\n" +
		"}\n"

	require.Equal(t, "attr\x00\"a\" -> \"b\"\x00key=\"a_to_b\"\x00style=solid",
		canonical.Canonical(t, dot))
}

func TestCanonicalDOTFinalAttributeDriftDetected(t *testing.T) {
	t.Parallel()

	// WR-01 false-pass window: penwidth=1 vs penwidth=2 as the FINAL attribute
	// must canonicalize differently, otherwise a renderer drift confined to
	// that position would silently pass the COMPAT-02 comparison.
	dotPenwidth1 := "digraph G {\n" +
		"\t\"a\" -> \"b\" [key=\"a_to_b\",\n" +
		"\t\tpenwidth=1];\n" +
		"}\n"
	dotPenwidth2 := "digraph G {\n" +
		"\t\"a\" -> \"b\" [key=\"a_to_b\",\n" +
		"\t\tpenwidth=2];\n" +
		"}\n"

	require.NotEqual(t, canonical.Canonical(t, dotPenwidth1), canonical.Canonical(t, dotPenwidth2))
	require.Contains(t, canonical.Canonical(t, dotPenwidth2), "penwidth=2")
}

func TestCanonicalDOTQuotedValuesDoNotTruncate(t *testing.T) {
	t.Parallel()

	// WR-02 regression: `];` and braces inside quoted attribute values must not
	// terminate the statement early — the full value and every following
	// statement must survive canonicalization.
	dot := "digraph G {\n" +
		"\t\"a\" -> \"b\" [description=\"SSH [session]; admin\",\n" +
		"\t\tminlen=3];\n" +
		"\t\"b\" -> \"c\" [label=\"uses {braces}\"];\n" +
		"}\n"

	canon := canonical.Canonical(t, dot)
	require.Contains(t, canon, "description=\"SSH [session]; admin\"")
	require.Contains(t, canon, "minlen=3")
	require.Contains(t, canon, "label=\"uses {braces}\"")
	require.Contains(t, canon, "\"b\" -> \"c\"")
}

func TestCanonicalDOTHTMLLabelDoesNotTruncate(t *testing.T) {
	t.Parallel()

	// WR-02 regression: `];` inside an HTML label (<...>) must not terminate
	// the statement early; HTML labels contain entities and arbitrary text.
	dot := "digraph G {\n" +
		"\t\"a\" [label=<<b>SSH [session]; admin</b>>];\n" +
		"}\n"

	require.Contains(t, canonical.Canonical(t, dot), "label=<<b>SSH [session]; admin</b>>")
}
