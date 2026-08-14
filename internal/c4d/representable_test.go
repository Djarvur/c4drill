package c4d_test

// F-01 regression suite: the representability gate (CR-01/CR-03), the
// quote-terminated multiline literal fix (CR-05) and the quoted-key TOML
// emitter (CR-02). The defect classes came from the phase verification
// report: a LEGAL source document must never convert into a twin that is
// silently corrupted or does not parse — values C4D cannot express are loud
// hard errors, values the target format CAN express must round-trip.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Djarvur/c4drill/internal/c4d"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCheckC4DRepresentableRejects: every corrupting identifier/label class
// from the verification report is a hard error NAMING the offending value
// (CR-01: '|' in technology/description; CR-03: unit ids outside
// [A-Za-z0-9_-]+ at both levels, peer path segments, template names, use
// template names).
func TestCheckC4DRepresentableRejects(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		tomlSrc   string
		offending string
	}{
		{
			name: "pipe in technology (CR-01)",
			tomlSrc: `
[web]
type = "system"

[[web.link]]
peer = "api"
technology = "HTTP | REST"
description = "calls"

[api]
type = "queue"
`,
			offending: `"HTTP | REST"`,
		},
		{
			name: "pipe in description",
			tomlSrc: `
[web]
type = "system"

[[web.link]]
peer = "api"
description = "calls | proxies"

[api]
type = "queue"
`,
			offending: `"calls | proxies"`,
		},
		{
			name:      "unit id with space (CR-03)",
			tomlSrc:   `["my unit"]` + "\ntype = \"system\"\n",
			offending: `"my unit"`,
		},
		{
			name:      "unit id with dot (CR-03)",
			tomlSrc:   `["my.unit"]` + "\ntype = \"system\"\n",
			offending: `"my.unit"`,
		},
		{
			name:      "unit id with unicode (CR-03)",
			tomlSrc:   `["ünit"]` + "\ntype = \"system\"\n",
			offending: `"ünit"`,
		},
		{
			name: "subunit id with space (CR-03)",
			tomlSrc: `
[web]
type = "system"

[web."my sub"]
type = "container"
`,
			offending: `"my sub"`,
		},
		{
			name: "peer with space (CR-03)",
			tomlSrc: `
[web]
type = "system"

[[web.link]]
peer = "my peer"

[api]
type = "queue"
`,
			offending: `"my peer"`,
		},
		{
			name: "peer with unicode segment (CR-03)",
			tomlSrc: `
[web]
type = "system"

[[web.link]]
peer = "api.störe"

[api]
type = "queue"
`,
			offending: `"api.störe"`,
		},
		{
			name: "template name with space",
			tomlSrc: `
[template."my tpl"]
type = "container"
`,
			offending: `"my tpl"`,
		},
		{
			name: "use template name with space",
			tomlSrc: `
[template.legal]
type = "container"

[[use]]
template = "my tpl"
`,
			offending: `"my tpl"`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			m, err := parser.Parse([]byte(tc.tomlSrc))
			require.NoError(t, err, "source is a LEGAL TOML document")

			err = c4d.CheckC4DRepresentable(m)
			require.Error(t, err, "unrepresentable value must be a hard error")
			assert.Contains(t, err.Error(), tc.offending,
				"error must name the offending value")
			assert.Contains(t, err.Error(), "not representable",
				"error must say what is wrong")
		})
	}
}

// TestCheckC4DRepresentableAcceptsClean: bareword-legal models — including
// template-body ${param} peers and mixed identifier/param peer paths, which
// the grammar admits as PeerSeg tokens — pass the gate untouched.
func TestCheckC4DRepresentableAcceptsClean(t *testing.T) {
	t.Parallel()

	m, err := parser.Parse([]byte(`
[template.sidecar]
type = "container"

[[template.sidecar.link]]
peer = "${target}"

[web]
type = "system"

[[web.link]]
peer = "api.${target}"

[api]
type = "queue"
`))
	require.NoError(t, err)

	assert.NoError(t, c4d.CheckC4DRepresentable(m),
		"${param} peer segments are legal PeerRef segments (D-13)")
	assert.NoError(t, c4d.CheckC4DRepresentable(nil), "nil model is trivially representable")
}

// TestQuoteTerminatedMultilineRoundTrip pins the CR-05 fix: a multi-line
// value ending in '"' must NOT emit as a triple-quoted literal (the trailing
// quote merges with the closing delimiter into an ambiguous `""""` run that
// does not re-parse) — it emits as the escaped quoted form and round-trips
// to the canonically-equal model.
func TestQuoteTerminatedMultilineRoundTrip(t *testing.T) {
	t.Parallel()

	m1, err := parser.Parse([]byte(`
[web]
type = "system"
description = "line1\nline2\""
`))
	require.NoError(t, err, "source parses")

	c4dText := c4d.EmitC4D(c4d.FromModel(m1))
	assert.Contains(t, c4dText, `description: "line1\nline2\""`,
		"quote-terminated multiline value emits as the escaped quoted form:\n%s", c4dText)

	m2, err := c4d.Parse([]byte(c4dText))
	require.NoError(t, err, "emitted C4D re-parses (CR-05)")

	require.Equal(t, "line1\nline2\"", m2.Units["web"].Description,
		"the exact value survives the round trip")

	require.Equal(t, canonicalModel(m1), canonicalModel(m2),
		"models equal modulo the D-22 explicit-defaults list")
}

// TestTypeLedDisplayNameToTOMLRoundTrip pins the CR-02 fix: a type-led .c4d
// unit whose map key is the display name ("My App" — not a TOML bare key)
// must emit a QUOTED table key (["My App"]), and the twin must re-parse to
// the same model.
func TestTypeLedDisplayNameToTOMLRoundTrip(t *testing.T) {
	t.Parallel()

	m1, err := c4d.Parse([]byte(`system "My App" {
	description: serves traffic
}
`))
	require.NoError(t, err, "type-led .c4d source parses")
	require.Contains(t, m1.Units, "My App", "type-led unit keys on its display name")

	tomlText, err := c4d.EmitTOML(m1)
	require.NoError(t, err)
	assert.Contains(t, tomlText, `["My App"]`,
		"non-bare table keys must emit quoted (CR-02):\n%s", tomlText)

	m2, err := parser.Parse([]byte(tomlText))
	require.NoError(t, err, "quoted-key twin re-parses (CR-02)")

	require.Equal(t, canonicalModel(m1), canonicalModel(m2),
		"models equal modulo the D-22 explicit-defaults list")
}

// TestEmitTOMLQuotedKeysPreserveTemplateUses exercises the quoted-key path
// end to end for template subunits, template-body uses and use param keys: a
// subunit whose key is not a bare key keeps its
// [[template.<name>."key".use]] placement, a non-bare param key emits quoted,
// and the twin re-parses to the same model.
func TestEmitTOMLQuotedKeysPreserveTemplateUses(t *testing.T) {
	t.Parallel()

	m1, err := parser.Parse([]byte(`
[template.svc]
type = "container"

[template.svc."my sub"]
type = "system"

[[template.svc."my sub".use]]
template = "svc"
"my key" = "hi there"
`))
	require.NoError(t, err, "source parses")

	tomlText, err := c4d.EmitTOML(m1)
	require.NoError(t, err)
	assert.Contains(t, tomlText, `[template.svc."my sub"]`,
		"quoted template subunit key:\n%s", tomlText)
	assert.Contains(t, tomlText, `[[template.svc."my sub".use]]`,
		"use stays under its quoted parent path:\n%s", tomlText)
	assert.Contains(t, tomlText, `"my key" = "hi there"`,
		"non-bare param keys emit quoted:\n%s", tomlText)

	m2, err := parser.Parse([]byte(tomlText))
	require.NoError(t, err, "quoted-key twin re-parses")

	require.Equal(t, canonicalModel(m1), canonicalModel(m2),
		"models equal modulo the D-22 explicit-defaults list")
}

// TestConvertGatesTransparentOverCorpus proves the F-01 gates (the
// representability check plus the re-parse/canonical-equality write gate,
// applied to every corpus fixture's fresh parse) stay TRANSPARENT for the
// 29-fixture parity corpus: the corpus contains no corrupting inputs, so
// every fixture must still convert to a C4D twin that re-parses to the
// canonically-equal model.
func TestConvertGatesTransparentOverCorpus(t *testing.T) {
	t.Parallel()

	validPaths, _ := corpusTOMLPaths(t)
	require.Greater(t, len(validPaths), 15, "corpus walker must cover the parity fixtures")

	for _, path := range validPaths {
		rel, err := filepath.Rel(repoRoot(), path)
		require.NoError(t, err)

		t.Run(rel, func(t *testing.T) {
			t.Parallel()

			//nolint:gosec // G304: path from the pinned corpus walker, not user input
			data, err := os.ReadFile(path)
			require.NoError(t, err, "read corpus fixture")

			m, err := parser.Parse(data)
			require.NoError(t, err, "fixture parses")

			require.NoError(t, c4d.CheckC4DRepresentable(m),
				"representability gate is transparent for corpus fixtures")

			twin, err := c4d.Parse([]byte(c4d.EmitC4D(c4d.FromModel(m))))
			require.NoError(t, err, "emitted twin re-parses (write gate's re-parse half)")

			assert.True(t, c4d.CanonicalEqual(m, twin),
				"write gate's model-equality half is transparent for corpus fixtures")
		})
	}
}
