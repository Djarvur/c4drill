package main

// Tests for the convert subcommand (Plan 35-07 Task 2): to-toml/to-c4d with
// validate-first semantics (D-24), swapped-extension output placement (D-30)
// and lossless single-file structure preservation (D-25) — include
// directives, template declarations, use instantiations and authored bare
// peers survive conversion because the emission model is a FRESH source
// parse that never sees the pipeline's mutating stages.
//
// NOTE: no t.Parallel — cobra flags bind package-level vars (root.go
// precedent), so concurrent Execute calls would race on flag state even
// though these tests never render (the WASM constraint does not apply).

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/Djarvur/c4drill/internal/c4d"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// convertSourceC4D is a gate-valid .c4d document exercising every structure
// the D-25 preservation contract must survive: an include directive (with
// once), a template declaration, a unit-nested use instantiation (D-16) and
// authored bare relative peers (api.cache -> bus).
const convertSourceC4D = `properties {
	name: Convert Source
}

user: person "User" {
	-> api: HTTPS | drives traffic
}

api: system "API" {
	cache: containerDb "Cache" {
		-> bus: Redis
	}
	use sidecar(name: metrics)
}

bus: queue "Bus" {
	<- api.cache: feeds
}

template sidecar(name) {
	type: container
	technology: ${name}
	-> bus
}

include ./extra.c4d once
`

// convertExtraC4D is the include target of convertSourceC4D — a linked
// top-level unit so the merged model passes the D-24 validation gate.
const convertExtraC4D = `metricsDb: db "Metrics DB" {
	<- bus: reports
}
`

// convertSourceTOML is the TOML mirror of convertSourceC4D (same structures,
// same semantics) used for the to-c4d direction.
const convertSourceTOML = `[properties]
name = "Convert Source TOML"

[user]
type = "person"
name = "User"

[[user.link]]
peer = "api"
technology = "HTTPS"
description = "drives traffic"

[api]
type = "system"
name = "API"

[api.cache]
type = "containerDb"
name = "Cache"

[[api.cache.link]]
peer = "bus"
technology = "Redis"

[bus]
type = "queue"
name = "Bus"

[[bus.linkFrom]]
peer = "api.cache"
description = "feeds"

[template.sidecar]
params = ["name"]
type = "container"
technology = "${name}"

[[template.sidecar.link]]
peer = "bus"

[[use]]
template = "sidecar"
parent = "api"
name = "metrics"

[[include]]
path = "./extra.toml"
once = true
`

// convertExtraTOML is the include target of convertSourceTOML.
const convertExtraTOML = `[metricsDb]
type = "db"
name = "Metrics DB"

[[metricsDb.linkFrom]]
peer = "bus"
description = "reports"
`

// writeConvertFixture writes name + extraName (the include target) into a
// fresh temp dir and returns the entry file path.
func writeConvertFixture(t *testing.T, name, entry, extraName, extra string) string {
	t.Helper()

	dir := t.TempDir()
	entryPath := filepath.Join(dir, name)

	require.NoError(t, os.WriteFile(entryPath, []byte(entry), 0o600), "write %s", name)
	require.NoError(t,
		os.WriteFile(filepath.Join(dir, extraName), []byte(extra), 0o600),
		"write include target %s", extraName)

	return entryPath
}

// runConvert executes the cobra root command with the given convert args and
// returns the resulting error.
//
//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func runConvert(t *testing.T, args ...string) error {
	t.Helper()

	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs(append([]string{"convert"}, args...))

	return cmd.Execute() //nolint:wrapcheck // test returns the command error verbatim
}

// convertCanonicalModel clones m and fills the D-22 explicit-defaults list
// (arrow=forward, rank=forward, labelPosition=middle) on every link, so
// models parsed from implicitly-defaulted sources compare fairly. Mirrors
// internal/c4d's parity-test helper (Compare-before-Validate also applies:
// Validate mutates LinksFrom in map-iteration order).
func convertCanonicalModel(m *parser.Model) *parser.Model {
	clone := &parser.Model{
		Properties: m.Properties,
		UnitOrder:  cloneStringSlice(m.UnitOrder),
		Units:      make(map[string]*model.Unit, len(m.Units)),
	}

	for name, unit := range m.Units {
		clone.Units[name] = convertCanonicalUnit(unit)
	}

	// Templates, instantiations and includes are deep-compared verbatim:
	// the preservation contract (D-25) requires them to survive EXACTLY.
	clone.Templates = m.Templates
	clone.Instantiations = m.Instantiations
	clone.Includes = m.Includes

	return clone
}

// convertCanonicalUnit deep-clones one unit subtree with link defaults filled.
func convertCanonicalUnit(unit *model.Unit) *model.Unit {
	if unit == nil {
		return nil
	}

	clone := unit.Clone()
	fillConvertLinkDefaults(clone)

	return clone
}

// fillConvertLinkDefaults fills the D-22 defaults in place on a cloned unit.
func fillConvertLinkDefaults(unit *model.Unit) {
	for i := range unit.Links {
		convertLinkDefaults(&unit.Links[i])
	}

	for i := range unit.LinksFrom {
		convertLinkDefaults(&unit.LinksFrom[i])
	}

	for _, sub := range unit.Subunits {
		fillConvertLinkDefaults(sub)
	}
}

// convertLinkDefaults fills one link's D-22 defaults.
func convertLinkDefaults(link *model.Link) {
	if link.Arrow == "" {
		link.Arrow = model.ArrowForward
	}

	if link.Rank == "" {
		link.Rank = model.RankForward
	}

	if link.LabelPosition == "" {
		link.LabelPosition = model.LabelMiddle
	}
}

// cloneStringSlice clones a string slice (nil-preserving).
func cloneStringSlice(s []string) []string {
	if s == nil {
		return nil
	}

	out := make([]string, len(s))
	copy(out, s)

	return out
}

// requirePreserved asserts the D-25 single-file preservation contract on a
// converted twin's re-parsed model, against the FRESH source parse:
// include directives verbatim (path + once), template declarations with
// their bare template-body peers, unit-nested use instantiations, and the
// authored bare relative peers non-absolutized.
func requirePreserved(t *testing.T, twin *parser.Model, src *parser.Model) {
	t.Helper()

	// Include directives survive verbatim (D-25 default: the emission model
	// never ran include.Resolve, which would drain Includes to nil).
	require.NotEmpty(t, src.Includes, "sanity: source must have include directives")
	require.Equal(t, src.Includes, twin.Includes,
		"include directives must survive conversion verbatim (path + once)")

	// Template declarations survive (the emission model never ran
	// template.Expand, which consumes Templates).
	require.Contains(t, src.Templates, "sidecar", "sanity: source declares the template")
	require.Equal(t, src.Templates["sidecar"].Params, twin.Templates["sidecar"].Params,
		"template params survive")

	require.Len(t, twin.Templates["sidecar"].Unit.Links, 1,
		"template body link survives")
	require.Equal(t, "bus", twin.Templates["sidecar"].Unit.Links[0].Peer,
		"template-body bare peer stays bare (no peer.Resolve on the emission model)")

	// Use instantiations survive (no Expand).
	require.Len(t, twin.Instantiations, 1, "use instantiation survives")
	require.Equal(t, "sidecar", twin.Instantiations[0].Template, "use names the template")

	// Authored bare relative peers stay bare (no peer.Resolve absolutization).
	cache := twin.Units["api"].Subunits["cache"]
	require.NotNil(t, cache, "converted twin keeps the api.cache subunit")
	require.Len(t, cache.Links, 1, "api.cache keeps its link")

	require.NotContains(t, cache.Links[0].Peer, ".",
		"authored bare peer must stay non-absolutized (D-22 round-trip parity)")
	require.Equal(t, "bus", cache.Links[0].Peer, "authored bare peer value preserved")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestConvertToTOMLRoundTrip(t *testing.T) {
	entry := writeConvertFixture(t, "diagram.c4d", convertSourceC4D, "extra.c4d", convertExtraC4D)

	require.NoError(t, runConvert(t, "to-toml", entry), "convert to-toml must succeed")

	// D-30: output lands next to the input with the extension swapped.
	outPath := filepath.Join(filepath.Dir(entry), "diagram.toml")
	assert.FileExists(t, outPath, "twin must be written next to the input")

	// D-28: the twin re-parses to the same Model as a FRESH source parse
	// (NOT the pipeline-mutated gate model).
	data, err := os.ReadFile(outPath) //nolint:gosec // G304: path inside t.TempDir()
	require.NoError(t, err, "read converted twin")

	twin, err := parser.Parse(data)
	require.NoError(t, err, "converted TOML must re-parse")

	src, err := c4d.ParseFile(entry)
	require.NoError(t, err, "fresh source parse")

	require.Equal(t, convertCanonicalModel(src), convertCanonicalModel(twin),
		"converted twin re-parses to the fresh-source Model (D-28)")

	requirePreserved(t, twin, src)
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestConvertToC4DRoundTrip(t *testing.T) {
	entry := writeConvertFixture(t, "diagram.toml", convertSourceTOML, "extra.toml", convertExtraTOML)

	require.NoError(t, runConvert(t, "to-c4d", entry), "convert to-c4d must succeed")

	outPath := filepath.Join(filepath.Dir(entry), "diagram.c4d")
	assert.FileExists(t, outPath, "twin must be written next to the input")

	data, err := os.ReadFile(outPath) //nolint:gosec // G304: path inside t.TempDir()
	require.NoError(t, err, "read converted twin")

	twin, err := c4d.Parse(data)
	require.NoError(t, err, "converted C4D must re-parse")

	src, err := parser.ParseFile(entry)
	require.NoError(t, err, "fresh source parse")

	require.Equal(t, convertCanonicalModel(src), convertCanonicalModel(twin),
		"converted twin re-parses to the fresh-source Model (D-28)")

	requirePreserved(t, twin, src)
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestConvertOutputDirFlag(t *testing.T) {
	entry := writeConvertFixture(t, "diagram.toml", convertSourceTOML, "extra.toml", convertExtraTOML)

	outDir := filepath.Join(t.TempDir(), "nested", "twins")
	require.NoError(t, runConvert(t, "to-c4d", entry, "--output", outDir),
		"convert with -o must succeed and create the target directory")

	assert.FileExists(t, filepath.Join(outDir, "diagram.c4d"),
		"twin must exist in the -o directory (D-30)")

	_, err := os.Stat(filepath.Join(filepath.Dir(entry), "diagram.c4d"))
	assert.True(t, os.IsNotExist(err),
		"twin must NOT be written next to the input when -o redirects (D-30)")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestConvertInvalidInputNoOutput(t *testing.T) {
	// invalid_links.toml is invalid by design (parity corpus refusal list):
	// a parent unit carrying a direct link. Copy it into a temp dir so a
	// broken implementation cannot scribble into committed testdata.
	src, err := os.ReadFile(filepath.Join("..", "..", "testdata", "invalid_links.toml"))
	require.NoError(t, err, "read invalid_links.toml fixture")

	dir := t.TempDir()
	entry := filepath.Join(dir, "invalid_links.toml")
	require.NoError(t, os.WriteFile(entry, src, 0o600), "write invalid fixture copy")

	err = runConvert(t, "to-c4d", entry)
	require.Error(t, err, "invalid input must be a hard error (D-24)")
	assert.Contains(t, err.Error(), "validation", "error must surface the validation gate")

	_, statErr := os.Stat(filepath.Join(dir, "invalid_links.c4d"))
	assert.True(t, os.IsNotExist(statErr),
		"NO output file may be written when the validation gate fails (D-24)")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestConvertUnresolvableInclude(t *testing.T) {
	dir := t.TempDir()
	entry := filepath.Join(dir, "ghosty.toml")
	require.NoError(t, os.WriteFile(entry, []byte(`
[properties]
name = "Ghost Include"

[user]
type = "person"
name = "User"

[[include]]
path = "ghost.toml"
`), 0o600), "write entry with missing include")

	err := runConvert(t, "to-c4d", entry)
	require.Error(t, err, "missing include must be a hard error")

	msg := err.Error()
	assert.Contains(t, msg, "include", "error must surface the include stage prefix")
	assert.Contains(t, msg, "ghost.toml", "error must name the missing path")

	_, statErr := os.Stat(filepath.Join(dir, "ghosty.c4d"))
	assert.True(t, os.IsNotExist(statErr), "no twin on stage failure")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestConvertWrongDirection(t *testing.T) {
	entry := writeConvertFixture(t, "diagram.toml", convertSourceTOML, "extra.toml", convertExtraTOML)

	// to-toml needs a .c4d input; feeding it a .toml must hard-error and
	// explain the expected extension.
	err := runConvert(t, "to-toml", entry)
	require.Error(t, err, "wrong-direction extension must be a hard error")

	msg := err.Error()
	assert.Contains(t, msg, "to-toml", "error must name the direction")
	assert.Contains(t, msg, ".c4d", "error must name the expected input extension")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestConvertHelp(t *testing.T) {
	cmd := NewRootCmd()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"convert", "--help"})

	require.NoError(t, cmd.Execute(), "convert --help must succeed")

	help := buf.String()
	assert.Contains(t, help, "to-toml", "help must document the to-toml direction")
	assert.Contains(t, help, "to-c4d", "help must document the to-c4d direction")
}
