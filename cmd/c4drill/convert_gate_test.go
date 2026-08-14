package main

// F-01 regression suite for the convert command: the representability gate
// (CR-01/CR-03 — loud hard errors instead of silently corrupted twins), the
// quoted-key TOML twin (CR-02), the quote-terminated multiline literal
// (CR-05) and the WRITE GATE — a twin that would not re-parse or would
// re-parse to a different model is NEVER written to disk (fmt's T-35-08-01
// safety-gate pattern). Every scenario drives the real cobra command path
// (execConvert) with real files — no mocking.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Djarvur/c4drill/internal/c4d"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeGateFixture writes one source file into a fresh temp dir and returns
// its path plus the dir (for twin-absence assertions).
func writeGateFixture(t *testing.T, name, src string) (string, string) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(src), 0o600), "write %s", name)

	return path, dir
}

// requireNoTwin asserts the swapped-extension twin of the standard
// "diagram.*" fixture was never written.
func requireNoTwin(t *testing.T, dir string) {
	t.Helper()

	twin := filepath.Join(dir, "diagram.c4d")
	_, err := os.Stat(twin)
	assert.True(t, os.IsNotExist(err), "no corrupt twin may be written: %s", twin)
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestConvertPipeLabelHardError(t *testing.T) {
	// CR-01: a LEGAL TOML document whose link technology contains '|' — the
	// D-09 first-pipe split would reshuffle tech/desc in the twin. Conversion
	// must hard-error naming the value, not corrupt.
	entry, dir := writeGateFixture(t, "diagram.toml", `
[web]
type = "system"

[[web.link]]
peer = "api"
technology = "HTTP | REST"
description = "calls"

[api]
type = "queue"
`)

	err := execConvert(t, "to-c4d", entry)
	require.Error(t, err, "pipe in a link label must be a hard error (CR-01)")

	msg := err.Error()
	assert.Contains(t, msg, "not representable", "error names the representability gate")
	assert.Contains(t, msg, "HTTP | REST", "error names the offending value")
	assert.Contains(t, msg, entry, "error names the source file")

	requireNoTwin(t, dir)
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestConvertNonIdentifierUnitIDHardError(t *testing.T) {
	// CR-03: TOML unit ids outside [A-Za-z0-9_-]+ (spaces, dots, unicode via
	// quoted table keys) cannot ride a C4D unit header — hard error, no twin.
	for name, src := range map[string]string{
		"space id": `["my unit"]
type = "system"

[["my unit".link]]
peer = "api"

[api]
type = "queue"
`,
		"dot id": `["my.unit"]
type = "system"

[["my.unit".link]]
peer = "api"

[api]
type = "queue"
`,
		"unicode id": `["ünit"]
type = "system"

[["ünit".link]]
peer = "api"

[api]
type = "queue"
`,
	} {
		entry, dir := writeGateFixture(t, "diagram.toml", src)

		err := execConvert(t, "to-c4d", entry)
		require.Error(t, err, "%s: non-identifier unit id must be a hard error (CR-03)", name)

		assert.Contains(t, err.Error(), "not representable",
			"%s: error names the representability gate", name)

		requireNoTwin(t, dir)
	}
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestConvertTypeLedDisplayNameToTOMLRoundTrip(t *testing.T) {
	// CR-02: a type-led .c4d unit with a display name converts to a QUOTED
	// TOML table key (["My App"]) that re-parses to the same model.
	entry, dir := writeGateFixture(t, "diagram.c4d", `system "My App" {
	-> api: serves
}

api: queue "Api" { }
`)

	require.NoError(t, execConvert(t, "to-toml", entry),
		"display-name unit key is representable via a quoted TOML key (CR-02)")

	twinPath := filepath.Join(dir, "diagram.toml")
	assert.FileExists(t, twinPath, "twin must be written")

	data, err := os.ReadFile(twinPath) //nolint:gosec // path inside t.TempDir()
	require.NoError(t, err, "read twin")

	assert.Contains(t, string(data), `["My App"]`,
		"non-bare table keys must emit quoted (CR-02):\n%s", data)

	twin, err := parser.Parse(data)
	require.NoError(t, err, "quoted-key twin re-parses (CR-02)")

	src, err := c4d.ParseFile(entry)
	require.NoError(t, err, "fresh source parse")

	require.Equal(t, convertCanonicalModel(src), convertCanonicalModel(twin),
		"twin re-parses to the source model")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestConvertQuoteTerminatedMultilineToC4D(t *testing.T) {
	// CR-05: a multi-line value ending in '"' must emit a re-parseable
	// literal (the escaped quoted form) and round-trip to the same model.
	entry, dir := writeGateFixture(t, "diagram.toml", `
[web]
type = "system"
description = "line1\nline2\""

[[web.link]]
peer = "api"

[api]
type = "queue"
`)

	require.NoError(t, execConvert(t, "to-c4d", entry),
		"quote-terminated multiline values are representable (CR-05)")

	twinPath := filepath.Join(dir, "diagram.c4d")
	assert.FileExists(t, twinPath, "twin must be written")

	data, err := os.ReadFile(twinPath) //nolint:gosec // path inside t.TempDir()
	require.NoError(t, err, "read twin")

	twin, err := c4d.Parse(data)
	require.NoError(t, err, "twin re-parses (CR-05)")

	src, err := parser.ParseFile(entry)
	require.NoError(t, err, "fresh source parse")

	require.Equal(t, convertCanonicalModel(src), convertCanonicalModel(twin),
		"twin re-parses to the source model")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestConvertWriteGateRefusesUnparseableTwin(t *testing.T) {
	// A reserved word as a subunit id passes the pinned representability
	// list (it matches the identifier charset) but the emitted twin cannot
	// re-parse (`include:` inside a unit body is the D-19 reserved-id error).
	// The write gate must abort with a loud error and write NOTHING.
	entry, dir := writeGateFixture(t, "diagram.toml", `
[web]
type = "system"

[web.include]
type = "container"

[[web.include.link]]
peer = "api"

[api]
type = "queue"
`)

	err := execConvert(t, "to-c4d", entry)
	require.Error(t, err, "unparseable twin must abort the conversion (write gate)")

	msg := err.Error()
	assert.Contains(t, msg, "safety gate", "error names the write gate")
	assert.Contains(t, msg, "does not re-parse", "error names the failure mode")
	assert.Contains(t, msg, entry, "error names the source file")

	requireNoTwin(t, dir)
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestConvertWriteGateRefusesUnequalTwin(t *testing.T) {
	// Padded link technology: the twin re-parses but the D-09 label split
	// TRIMS the padding, so the twin's model differs from the source's —
	// silent corruption the model-equality half of the gate must catch.
	entry, dir := writeGateFixture(t, "diagram.toml", `
[web]
type = "system"

[[web.link]]
peer = "api"
technology = " HTTPS "

[api]
type = "queue"
`)

	err := execConvert(t, "to-c4d", entry)
	require.Error(t, err, "semantically unequal twin must abort the conversion (write gate)")

	msg := err.Error()
	assert.Contains(t, msg, "safety gate", "error names the write gate")
	assert.Contains(t, msg, "different model", "error names the failure mode")
	assert.Contains(t, msg, entry, "error names the source file")

	requireNoTwin(t, dir)
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestConvertGraphWriteGateRefusesCorruptFragment(t *testing.T) {
	// Graph mode runs the same gates per file: a fragment whose link label
	// contains a pipe hard-errors and its twin is never written.
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "entry.toml"), []byte(`
[entry]
type = "system"

[[entry.link]]
peer = "fragunit"

[[include]]
path = "./frag.toml"
`), 0o600), "write entry")

	require.NoError(t, os.WriteFile(filepath.Join(dir, "frag.toml"), []byte(`
[fragunit]
type = "queue"

[[fragunit.link]]
peer = "entry"
technology = "HTTP | REST"
`), 0o600), "write fragment")

	entry := filepath.Join(dir, "entry.toml")

	err := execConvert(t, "to-c4d", entry, "--follow-includes")
	require.Error(t, err, "corrupting fragment must abort the graph conversion")
	assert.Contains(t, err.Error(), "not representable",
		"error names the representability gate")
	assert.Contains(t, err.Error(), "HTTP | REST", "error names the offending value")

	_, statErr := os.Stat(filepath.Join(dir, "frag.c4d"))
	assert.True(t, os.IsNotExist(statErr),
		"the corrupt fragment twin must NOT be written (write gate)")
}
