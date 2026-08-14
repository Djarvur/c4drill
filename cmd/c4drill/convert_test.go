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
	"github.com/Djarvur/c4drill/internal/include"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/peer"
	"github.com/Djarvur/c4drill/internal/template"
	"github.com/Djarvur/c4drill/internal/validator"
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
	-> api.cache: HTTPS | drives traffic
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
peer = "api.cache"
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

// execConvert executes the cobra root command with the given convert args
// and returns the resulting error.
func execConvert(t *testing.T, args ...string) error {
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

	require.NoError(t, execConvert(t, "to-toml", entry), "convert to-toml must succeed")

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

	require.NoError(t, execConvert(t, "to-c4d", entry), "convert to-c4d must succeed")

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
	require.NoError(t, execConvert(t, "to-c4d", entry, "--output", outDir),
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
	//nolint:gosec // G703: committed fixture copied into t.TempDir()
	require.NoError(t, os.WriteFile(entry, src, 0o600), "write invalid fixture copy")

	err = execConvert(t, "to-c4d", entry)
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

	err := execConvert(t, "to-c4d", entry)
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
	err := execConvert(t, "to-toml", entry)
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

// --- Plan 35-07 Task 3: convert --follow-includes (D-25 graph mode) ---

// copyComposedGraph copies the 09-composed include graph (entry.toml ->
// templates.toml + domains/auth.toml, plus single-file-equivalent.toml which
// is documentation OUTSIDE the graph) into a temp dir so graph-mode twins
// never touch committed example files. Returns the copied entry path.
func copyComposedGraph(t *testing.T) string {
	t.Helper()

	srcRoot := filepath.Join("..", "..", "skill", "examples", "09-composed")
	dstRoot := t.TempDir()

	for _, rel := range []string{
		"entry.toml",
		"templates.toml",
		filepath.Join("domains", "auth.toml"),
		"single-file-equivalent.toml",
	} {
		data, err := os.ReadFile(filepath.Join(srcRoot, rel)) //nolint:gosec // G304: committed fixture path
		require.NoError(t, err, "read %s", rel)

		dst := filepath.Join(dstRoot, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(dst), 0o750), "mkdir for %s", rel)
		//nolint:gosec // G703: committed fixture copied into t.TempDir()
		require.NoError(t, os.WriteFile(dst, data, 0o600), "write %s", rel)
	}

	return filepath.Join(dstRoot, "entry.toml")
}

// runGraphPipeline runs the load-bearing pipeline stages (include.Resolve ->
// template.Expand -> peer.Resolve -> Validate) on a graph entry and asserts
// zero validation errors — proving a converted graph still composes.
func runGraphPipeline(t *testing.T, entryPath string) {
	t.Helper()

	m, err := parseInput(entryPath)
	require.NoError(t, err, "parse graph entry")

	m, err = include.Resolve(m, filepath.Dir(entryPath), entryPath)
	require.NoError(t, err, "include.Resolve on converted graph")

	m, err = template.Expand(m)
	require.NoError(t, err, "template.Expand on converted graph")

	require.NoError(t, peer.Resolve(m), "peer.Resolve on converted graph")
	assert.Empty(t, validator.Validate(m),
		"converted include graph must validate clean end to end (D-24/D-25)")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestConvertFollowIncludesGraph(t *testing.T) {
	entry := copyComposedGraph(t)
	root := filepath.Dir(entry)

	require.NoError(t, execConvert(t, "to-c4d", entry, "--follow-includes"),
		"whole-graph conversion must succeed")

	// Exactly the 3 graph twins exist (D-25): entry, templates, domains/auth.
	for _, twin := range []string{
		filepath.Join(root, "entry.c4d"),
		filepath.Join(root, "templates.c4d"),
		filepath.Join(root, "domains", "auth.c4d"),
	} {
		assert.FileExists(t, twin, "graph twin must exist: %s", twin)
	}

	// single-file-equivalent.toml is NOT in the include graph — no twin.
	_, err := os.Stat(filepath.Join(root, "single-file-equivalent.c4d"))
	assert.True(t, os.IsNotExist(err),
		"files outside the include graph must not get twins")

	// Include paths inside entry.c4d are rewritten to the twin extension,
	// and the `once` flag survives the rewrite (D-25).
	m, err := c4d.ParseFile(filepath.Join(root, "entry.c4d"))
	require.NoError(t, err, "converted entry must re-parse")

	require.Len(t, m.Includes, 2, "entry keeps both include directives")
	assert.Equal(t, parser.IncludeDirective{Path: "templates.c4d", Once: true},
		m.Includes[0], "first include rewritten to .c4d with once preserved")
	assert.Equal(t, parser.IncludeDirective{Path: filepath.Join("domains", "auth.c4d"), Once: false},
		m.Includes[1], "second include rewritten to .c4d, relative form preserved")

	// The converted graph still composes: full pipeline on the .c4d entry
	// (mixed-format include dispatch, D-26) validates clean.
	runGraphPipeline(t, filepath.Join(root, "entry.c4d"))
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestConvertFollowIncludesOutputDir(t *testing.T) {
	entry := copyComposedGraph(t)
	outDir := filepath.Join(t.TempDir(), "twins")

	require.NoError(t, execConvert(t, "to-c4d", entry, "--follow-includes", "--output", outDir),
		"graph conversion with -o must succeed")

	// Twins land under -o preserving the graph's directory structure
	// relative to the entry's dir (domains/auth.toml -> domains/auth.c4d).
	for _, rel := range []string{
		"entry.c4d",
		"templates.c4d",
		filepath.Join("domains", "auth.c4d"),
	} {
		assert.FileExists(t, filepath.Join(outDir, rel), "twin under -o: %s", rel)
	}

	_, err := os.Stat(filepath.Join(filepath.Dir(entry), "entry.c4d"))
	assert.True(t, os.IsNotExist(err), "no twin next to the input when -o redirects")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestConvertFollowIncludesOutputDirRelativeEntry(t *testing.T) {
	entry := copyComposedGraph(t)
	root := filepath.Dir(entry)
	outDir := filepath.Join(t.TempDir(), "twins")

	// F-03 regression: invoke convert with a RELATIVE entry path (cwd-relative
	// name, not the absolute temp-dir path). t.Chdir makes the graph root the
	// working directory for the duration of the test and restores it after
	// (it also forbids parallelism, which these tests already renounce).
	t.Chdir(root)

	require.NoError(t,
		execConvert(t, "to-c4d", "entry.toml", "--follow-includes", "--output", outDir),
		"graph conversion with -o must succeed from a relative entry path")

	// Twins land under -o preserving the graph's directory structure relative
	// to the entry's dir — the SAME layout as the absolute-path invocation
	// (domains/auth.toml -> domains/auth.c4d).
	for _, rel := range []string{
		"entry.c4d",
		"templates.c4d",
		filepath.Join("domains", "auth.c4d"),
	} {
		assert.FileExists(t, filepath.Join(outDir, rel), "twin under -o: %s", rel)
	}

	// The domains subdirectory must NOT be flattened away by the relative
	// entry invocation.
	_, err := os.Stat(filepath.Join(outDir, "auth.c4d"))
	assert.True(t, os.IsNotExist(err),
		"relative entry path must not flatten domains/ under -o")
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestConvertGraphCycle(t *testing.T) {
	dir := t.TempDir()

	for name, other := range map[string]string{
		"a.toml": "b.toml",
		"b.toml": "a.toml",
	} {
		require.NoError(t,
			os.WriteFile(filepath.Join(dir, name), []byte(`
[properties]
name = "Cycle `+name+`"

[user]
type = "person"
name = "User"

[[include]]
path = "`+other+`"
`), 0o600), "write %s", name)
	}

	entry := filepath.Join(dir, "a.toml")

	// The D-24 gate's include.Resolve detects the cycle before any walk; the
	// traversal's own stack/visited guard (T-35-07-02) is defense-in-depth.
	// A hang here fails the suite timeout, pinning traversal termination.
	err := execConvert(t, "to-c4d", entry, "--follow-includes")
	require.Error(t, err, "include cycle must be a hard error")
	assert.Contains(t, err.Error(), "cycle", "error must name the cycle")

	for _, leaked := range []string{"a.c4d", "b.c4d"} {
		_, statErr := os.Stat(filepath.Join(dir, leaked))
		assert.True(t, os.IsNotExist(statErr), "no partial output %s on cycle", leaked)
	}
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestConvertFollowIncludesDefaultOff(t *testing.T) {
	entry := copyComposedGraph(t)
	root := filepath.Dir(entry)

	// Without --follow-includes exactly one file converts (D-25 default).
	require.NoError(t, execConvert(t, "to-c4d", entry), "single-file conversion must succeed")

	assert.FileExists(t, filepath.Join(root, "entry.c4d"), "entry twin exists")

	for _, notConverted := range []string{
		filepath.Join(root, "templates.c4d"),
		filepath.Join(root, "domains", "auth.c4d"),
	} {
		_, err := os.Stat(notConverted)
		assert.True(t, os.IsNotExist(err), "graph files stay unconverted: %s", notConverted)
	}

	// Include directives are preserved verbatim — untouched .toml paths.
	m, err := c4d.ParseFile(filepath.Join(root, "entry.c4d"))
	require.NoError(t, err, "single-file twin must re-parse")

	require.Len(t, m.Includes, 2, "both include directives preserved")
	assert.Equal(t, "templates.toml", m.Includes[0].Path,
		"include path stays verbatim without --follow-includes (D-25 default)")
	assert.True(t, m.Includes[0].Once, "once flag preserved verbatim")
	assert.Equal(t, filepath.Join("domains", "auth.toml"), m.Includes[1].Path,
		"second include path stays verbatim")
}

// mixedGraphTOMLEntry is a gate-valid mixed-format graph (D-26): a .toml
// entry including a .c4d fragment.
const mixedGraphTOMLEntry = `[properties]
name = "Mixed Graph"

[user]
type = "person"
name = "User"

[[user.link]]
peer = "svc.store"
technology = "HTTPS"

[[include]]
path = "./shared.c4d"
`

// mixedGraphSharedC4D is the .c4d fragment of the TOML-entry mixed graph.
const mixedGraphSharedC4D = `svc: system "Service" {
	store: containerDb "Store" {
		-> user
	}
}
`

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestConvertGraphMixedTOMLEntry(t *testing.T) {
	dir := t.TempDir()
	entry := filepath.Join(dir, "main.toml")

	require.NoError(t, os.WriteFile(entry, []byte(mixedGraphTOMLEntry), 0o600), "write entry")
	require.NoError(t,
		os.WriteFile(filepath.Join(dir, "shared.c4d"), []byte(mixedGraphSharedC4D), 0o600),
		"write shared fragment")

	require.NoError(t, execConvert(t, "to-c4d", entry, "--follow-includes"),
		"mixed-graph conversion must succeed")

	// The .toml entry gets its twin; shared.c4d is ALREADY in the target
	// format, so it is left untouched (conversion is additive — originals
	// stay, so the untouched file keeps resolving).
	assert.FileExists(t, filepath.Join(dir, "main.c4d"), "entry twin exists")

	_, err := os.Stat(filepath.Join(dir, "shared.toml"))
	assert.True(t, os.IsNotExist(err),
		"a file already in the target format gets no twin")

	// The include path inside main.c4d stays ./shared.c4d (identity rewrite).
	m, err := c4d.ParseFile(filepath.Join(dir, "main.c4d"))
	require.NoError(t, err, "converted mixed entry must re-parse")

	require.Len(t, m.Includes, 1, "include directive preserved")
	assert.Equal(t, "./shared.c4d", m.Includes[0].Path,
		"include of an already-target-format file keeps its path")

	runGraphPipeline(t, filepath.Join(dir, "main.c4d"))
}

//nolint:paralleltest // cobra flags bind package-level vars; serial execution only
func TestConvertGraphMixedC4DEntry(t *testing.T) {
	dir := t.TempDir()
	entry := filepath.Join(dir, "main.c4d")

	require.NoError(t, os.WriteFile(entry, []byte(`properties { name: Mixed Graph C4D }

user: person "User" {
	-> svc.store: HTTPS
}

include ./frag.toml
`), 0o600), "write entry")

	require.NoError(t,
		os.WriteFile(filepath.Join(dir, "frag.toml"), []byte(`[svc]
type = "system"
name = "Service"

[svc.store]
type = "containerDb"
name = "Store"

[[svc.store.link]]
peer = "user"
`), 0o600), "write TOML fragment")

	require.NoError(t, execConvert(t, "to-toml", entry, "--follow-includes"),
		"mixed-graph conversion to TOML must succeed")

	assert.FileExists(t, filepath.Join(dir, "main.toml"), "entry twin exists")

	_, err := os.Stat(filepath.Join(dir, "frag.c4d"))
	assert.True(t, os.IsNotExist(err), "fragment already in target format gets no twin")

	m, err := parser.ParseFile(filepath.Join(dir, "main.toml"))
	require.NoError(t, err, "converted mixed entry must re-parse")

	require.Len(t, m.Includes, 1, "include directive preserved")
	assert.Equal(t, "./frag.toml", m.Includes[0].Path,
		"include of an already-target-format file keeps its path")

	runGraphPipeline(t, filepath.Join(dir, "main.toml"))
}
