package c4d_test

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Djarvur/c4drill/internal/c4d"
	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/include"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/peer"
	"github.com/Djarvur/c4drill/internal/render"
	"github.com/Djarvur/c4drill/internal/template"
	"github.com/Djarvur/c4drill/internal/testutil/canonical"
	"github.com/Djarvur/c4drill/internal/testutil/canonsrc"
	"github.com/Djarvur/c4drill/internal/validator"
	"github.com/Djarvur/c4drill/internal/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file is the D-22/D-26 parity contract suite (Plan 35-06). The
// fixture guards below pin the edge-case corpus under testdata/c4d/; the
// round-trip and render-equivalence suites extend this file in Task 3.

// c4dFixtureDir is the edge-case fixture corpus root.
const c4dFixtureDir = "../../testdata/c4d"

// repoRoot is the repository root seen from internal/c4d tests.
func repoRoot() string {
	return filepath.Clean("../..")
}

// expectedC4DFixtures is the pinned corpus manifest — every file must exist,
// parse, and pass the full pre-render pipeline. The list doubles as the
// walker's anti-shrinkage guard: dropping a fixture fails by name.
//
//nolint:gochecknoglobals // immutable fixture manifest, the test corpus contract
var expectedC4DFixtures = []string{
	"external-types.toml",
	"kind.toml",
	"linkfrom.toml",
	"multiline-strings.toml",
	"rank-equal.toml",
	"template-nested-use.toml",
	"unicode-strings.toml",
}

// parseFixture parses one fixture file (fresh model — Validate mutates
// LinksFrom in place, so every guard parses its own copy).
func parseFixture(t *testing.T, name string) *parser.Model {
	t.Helper()

	//nolint:gosec // G304: fixture path built from the pinned manifest, not user input
	data, err := os.ReadFile(filepath.Join(c4dFixtureDir, name))
	require.NoError(t, err, "read fixture %s", name)

	m, err := parser.Parse(data)
	require.NoError(t, err, "fixture %s parses (fixtures are VALID documents)", name)

	return m
}

// assertFixturePipelineClean runs the D-24 convert gate mirrored at model
// level: a fixture must survive expand -> peer.Resolve -> validate with zero
// errors, exactly what `convert` will require before writing any output.
func assertFixturePipelineClean(t *testing.T, name string, m *parser.Model) {
	t.Helper()

	m, err := template.Expand(m)
	require.NoError(t, err, "fixture %s expands", name)

	require.NoError(t, peer.Resolve(m), "fixture %s peer-resolves", name)
	require.Empty(t, validator.Validate(m), "fixture %s validates clean", name)
}

// TestFixturesParseValidate: every corpus fixture exists, parses, and passes
// the full pre-render pipeline (D-24 gate at model level) — invalid
// documents live under the invalid_* names, never here.
func TestFixturesParseValidate(t *testing.T) {
	t.Parallel()

	for _, name := range expectedC4DFixtures {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			m := parseFixture(t, name)
			assertFixturePipelineClean(t, name, m)
		})
	}
}

// walkFixtureUnits visits every unit (and subunit) of a model with its
// dotted path.
func walkFixtureUnits(m *parser.Model, fn func(path string, u *model.Unit)) {
	for _, name := range m.UnitOrder {
		walkFixtureUnit(name, m.Units[name], fn)
	}
}

func walkFixtureUnit(path string, u *model.Unit, fn func(path string, u *model.Unit)) {
	if u == nil {
		return
	}

	fn(path, u)

	for _, sub := range u.SubunitOrder {
		walkFixtureUnit(path+"."+sub, u.Subunits[sub], fn)
	}
}

// fixtureUnitTypes collects the type set of every unit in the model.
func fixtureUnitTypes(m *parser.Model) map[model.UnitType]bool {
	types := make(map[model.UnitType]bool)

	walkFixtureUnits(m, func(_ string, u *model.Unit) { types[u.Type] = true })

	return types
}

// TestFixtureExternalTypesCoverage: external-types.toml covers all four C1
// external variants (D-04).
func TestFixtureExternalTypesCoverage(t *testing.T) {
	t.Parallel()

	types := fixtureUnitTypes(parseFixture(t, "external-types.toml"))

	for _, want := range []model.UnitType{
		model.TypePersonExternal, model.TypeSystemExternal,
		model.TypeDbExternal, model.TypeQueueExternal,
	} {
		assert.True(t, types[want], "fixture covers %s", want)
	}
}

// TestFixtureLinkFromCoverage: linkfrom.toml exercises an AUTHORED linkFrom
// plus a bidirectional and a none-arrow link (D-08/D-10 arrow set).
func TestFixtureLinkFromCoverage(t *testing.T) {
	t.Parallel()

	m := parseFixture(t, "linkfrom.toml")

	var (
		authoredLinkFrom bool
		bidirectional    bool
		noneArrow        bool
	)

	walkFixtureUnits(m, func(_ string, u *model.Unit) {
		if len(u.LinksFrom) > 0 {
			authoredLinkFrom = true
		}

		for _, link := range u.Links {
			switch link.Arrow {
			case model.ArrowBidirectional:
				bidirectional = true
			case model.ArrowNone:
				noneArrow = true
			case model.ArrowForward, model.ArrowReverse, "":
			}
		}
	})

	assert.True(t, authoredLinkFrom, "authored [[...linkFrom]] present")
	assert.True(t, bidirectional, "arrow = \"bidirectional\" link present")
	assert.True(t, noneArrow, "arrow = \"none\" link present")
}

// TestFixtureMultilineCoverage: multiline-strings.toml carries at least two
// values with embedded newlines (one TOML multi-line basic string, one
// escaped-\n single-line string) plus a long single word — the D-06 newline
// emission paths and the wrap-relevant long-word case.
func TestFixtureMultilineCoverage(t *testing.T) {
	t.Parallel()

	m := parseFixture(t, "multiline-strings.toml")

	newlineValues := 0
	longWord := ""

	walkFixtureUnits(m, func(_ string, u *model.Unit) {
		for _, value := range []string{u.Name, u.Description, u.Technology} {
			if strings.Contains(value, "\n") {
				newlineValues++
			}

			longWord = longestWord(longWord, value)
		}

		for _, link := range u.Links {
			for _, value := range []string{link.Technology, link.Description} {
				if strings.Contains(value, "\n") {
					newlineValues++
				}

				longWord = longestWord(longWord, value)
			}
		}
	})

	if strings.Contains(m.Properties.Description, "\n") {
		newlineValues++
	}

	assert.GreaterOrEqual(t, newlineValues, 2, "at least two embedded-newline values")
	assert.GreaterOrEqual(t, utf8.RuneCountInString(longWord), 20,
		"a long single word (wrap-relevant): got %q", longWord)
}

// longestWord returns the longer of prev and the longest whitespace-free
// run in s.
func longestWord(prev, s string) string {
	for _, word := range strings.Fields(s) {
		if utf8.RuneCountInString(word) > utf8.RuneCountInString(prev) {
			prev = word
		}
	}

	return prev
}

// TestFixtureRankEqualCoverage: rank-equal.toml sets rank = "equal" on at
// least one link.
func TestFixtureRankEqualCoverage(t *testing.T) {
	t.Parallel()

	found := false

	walkFixtureUnits(parseFixture(t, "rank-equal.toml"), func(_ string, u *model.Unit) {
		for _, link := range u.Links {
			if link.Rank == model.RankEqual {
				found = true
			}
		}
	})

	assert.True(t, found, `rank = "equal" link present`)
}

// TestFixtureKindCoverage: kind.toml carries all three kinds and an explicit
// colour override on a kind-carrying link (KIND-02).
func TestFixtureKindCoverage(t *testing.T) {
	t.Parallel()

	kinds := map[model.LinkKind]bool{}
	explicitColourOnKind := false

	walkFixtureUnits(parseFixture(t, "kind.toml"), func(_ string, u *model.Unit) {
		for _, link := range u.Links {
			if link.Kind != "" {
				kinds[link.Kind] = true
			}

			if link.Kind != "" && link.Color != "" {
				explicitColourOnKind = true
			}
		}
	})

	assert.True(t, kinds[model.KindRead], `kind = "read" present`)
	assert.True(t, kinds[model.KindWrite], `kind = "write" present`)
	assert.True(t, kinds[model.KindReadWrite], `kind = "read-write" present`)
	assert.True(t, explicitColourOnKind, "a kind link carries an explicit colour (KIND-02)")
}

// TestFixtureTemplateNestedUseCoverage: template-nested-use.toml combines
// [template.X] with the unit-nested [[unit.Y.use]] form (D-16) AND the
// template-body [[template.X.use]] form (D-17).
func TestFixtureTemplateNestedUseCoverage(t *testing.T) {
	t.Parallel()

	m := parseFixture(t, "template-nested-use.toml")

	require.NotEmpty(t, m.Templates, "template declarations present")

	bodyUse := false

	for _, tmpl := range m.Templates {
		if len(tmpl.Instantiations) > 0 {
			bodyUse = true
		}
	}

	assert.True(t, bodyUse, "[[template.X.use]] body use present (D-17)")

	unitNested := false

	for _, inst := range m.Instantiations {
		if inst.Parent != "" {
			unitNested = true
		}
	}

	assert.True(t, unitNested, "[[unit.Y.use]] unit-nested use present (D-16)")
}

// TestFixtureUnicodeCoverage: unicode-strings.toml carries multi-byte
// names/descriptions and a long single multi-byte word.
func TestFixtureUnicodeCoverage(t *testing.T) {
	t.Parallel()

	m := parseFixture(t, "unicode-strings.toml")

	multibyte := false
	longWord := ""

	walkFixtureUnits(m, func(_ string, u *model.Unit) {
		for _, value := range []string{u.Name, u.Description, u.Technology} {
			// Multi-byte content: fewer runes than bytes.
			if utf8.RuneCountInString(value) < len(value) {
				multibyte = true
			}

			longWord = longestWord(longWord, value)
		}
	})

	assert.True(t, multibyte, "multi-byte values present")

	require.NotEmpty(t, longWord, "a long word value exists")
	assert.Less(t, utf8.RuneCountInString(longWord), len(longWord),
		"long single word carries multi-byte runes: got %q", longWord)
	assert.GreaterOrEqual(t, utf8.RuneCountInString(longWord), 10,
		"long single word is long enough to stress wrapping: got %q", longWord)
}

// invalidCorpusFixtures lists corpus fixtures that are INVALID by design —
// the parity loop asserts conversion REFUSAL for them (the D-24 convert
// gate mirrored at model level: parse/expand/peer/validate must error), not
// round-trip output. Paths are repo-root-relative.
//
//nolint:gochecknoglobals // immutable invalid-fixture set (the refusal list)
var invalidCorpusFixtures = map[string]bool{
	"testdata/invalid_links.toml":                 true,
	"testdata/invalid_references.toml":            true,
	"testdata/invalid_subunits.toml":              true,
	"testdata/template_duplicate_path.toml":       true,
	"testdata/template_missing_param.toml":        true,
	"cmd/c4drill/testdata/invalid.toml":           true,
	"cmd/c4drill/testdata/peer_unresolvable.toml": true,
}

// corpusTOMLPaths walks the full fixture corpus (35-RESEARCH §3): testdata/,
// testdata/c4d/, cmd/c4drill/testdata/ (top level only) and skill/examples/
// (recursive, minus the 09-composed include graph which the composed-graph
// test converts as a whole). filepath.WalkDir never follows symlinked
// directories and only *.toml files pass the filter (T-35-06-01).
func corpusTOMLPaths(t *testing.T) ([]string, []string) {
	t.Helper()

	var validPaths, invalidPaths []string

	collect := func(paths []string) {
		for _, p := range paths {
			rel, err := filepath.Rel(repoRoot(), p)
			require.NoError(t, err, "corpus path under repo root")

			if invalidCorpusFixtures[rel] {
				invalidPaths = append(invalidPaths, p)

				continue
			}

			validPaths = append(validPaths, p)
		}
	}

	collect(flatCorpusTOMLs(t))
	collect(walkedExamplesTOMLs(t))
	slices.Sort(validPaths)
	slices.Sort(invalidPaths)

	return validPaths, invalidPaths
}

// flatCorpusTOMLs lists the *.toml files of the three top-level-only corpus
// directories.
func flatCorpusTOMLs(t *testing.T) []string {
	t.Helper()

	var out []string

	for _, dir := range []string{"testdata", "testdata/c4d", "cmd/c4drill/testdata"} {
		entries, err := os.ReadDir(filepath.Join(repoRoot(), dir))
		require.NoError(t, err, "read corpus dir %s", dir)

		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".toml") {
				out = append(out, filepath.Join(repoRoot(), dir, e.Name()))
			}
		}
	}

	return out
}

// walkedExamplesTOMLs recursively lists skill/examples/*.toml, skipping the
// 09-composed include graph (converted as a whole graph separately).
func walkedExamplesTOMLs(t *testing.T) []string {
	t.Helper()

	var out []string

	err := filepath.WalkDir(filepath.Join(repoRoot(), "skill", "examples"),
		func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			switch {
			case d.IsDir() && d.Name() == "09-composed":
				return filepath.SkipDir
			case d.IsDir() || !strings.HasSuffix(d.Name(), ".toml"):
				return nil
			}

			out = append(out, path)

			return nil
		})
	require.NoError(t, err, "walk skill/examples")

	return out
}

// canonicalModel returns a deep copy of m with the D-22 explicit-default
// list filled in (arrow "" == "forward", rank "" == "forward",
// labelPosition "" == "middle") so require.Equal can compare models whose
// only difference is defaults made explicit by a round-trip through the DSL
// (the `->` glyph IS the arrow spec).
func canonicalModel(m *parser.Model) *parser.Model {
	clone := &parser.Model{
		Properties:     m.Properties,
		UnitOrder:      slices.Clone(m.UnitOrder),
		Units:          make(map[string]*model.Unit, len(m.Units)),
		Templates:      make(map[string]*parser.TemplateDef, len(m.Templates)),
		Instantiations: slices.Clone(m.Instantiations),
		Includes:       slices.Clone(m.Includes),
	}

	for name, unit := range m.Units {
		clone.Units[name] = canonicalUnitForCompare(unit)
	}

	for name, tmpl := range m.Templates {
		if tmpl == nil {
			continue
		}

		t := *tmpl
		t.Params = slices.Clone(tmpl.Params)
		t.Unit = canonicalUnitForCompare(tmpl.Unit)
		t.Instantiations = slices.Clone(tmpl.Instantiations)
		clone.Templates[name] = &t
	}

	return clone
}

// canonicalUnitForCompare clones a unit and fills link defaults recursively.
func canonicalUnitForCompare(unit *model.Unit) *model.Unit {
	clone := unit.Clone()
	if clone == nil {
		return nil
	}

	fillLinkDefaults(clone)

	return clone
}

// fillLinkDefaults applies the D-22 default list to every link in the
// subtree, in place, on an already-cloned unit.
func fillLinkDefaults(unit *model.Unit) {
	for i := range unit.Links {
		canonicalLinkDefaults(&unit.Links[i])
	}

	for i := range unit.LinksFrom {
		canonicalLinkDefaults(&unit.LinksFrom[i])
	}

	for _, sub := range unit.Subunits {
		fillLinkDefaults(sub)
	}
}

// canonicalLinkDefaults fills one link's D-22 defaults.
func canonicalLinkDefaults(link *model.Link) {
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

// TestRoundTripTOMLToC4DToTOML is the D-22 forward contract over the FULL
// corpus: parser.Parse -> EmitC4D(FromModel(m)) -> c4d.Parse -> EmitTOML ->
// parser.Parse must produce canonically-equal TOML text (canonsrc) and
// require.Equal Models — modulo the explicit-defaults list — so UnitOrder
// and every authored value survive (order-preservation contract).
func TestRoundTripTOMLToC4DToTOML(t *testing.T) {
	t.Parallel()

	validPaths, _ := corpusTOMLPaths(t)
	require.Greater(t, len(validPaths), 15, "corpus walker must cover >15 fixtures (anti-shrinkage)")

	t.Logf("round-trip corpus tally: %d fixtures", len(validPaths))

	for _, path := range validPaths {
		rel, err := filepath.Rel(repoRoot(), path)
		require.NoError(t, err)

		t.Run(rel, func(t *testing.T) {
			t.Parallel()

			//nolint:gosec // G304: path from the pinned corpus walker, not user input
			data, err := os.ReadFile(path)
			require.NoError(t, err, "read corpus fixture")

			m1, err := parser.Parse(data)
			require.NoError(t, err, "fixture parses")

			c4dText := c4d.EmitC4D(c4d.FromModel(m1))

			m2, err := c4d.Parse([]byte(c4dText))
			require.NoError(t, err, "emitted C4D re-parses")

			secondEmit, err := c4d.EmitTOML(m2)
			require.NoError(t, err, "second TOML emission")

			m3, err := parser.Parse([]byte(secondEmit))
			require.NoError(t, err, "second TOML parses")

			firstEmit, err := c4d.EmitTOML(m1)
			require.NoError(t, err, "first TOML emission")

			require.Equal(t,
				canonsrc.NormalizeTOML(t, firstEmit),
				canonsrc.NormalizeTOML(t, secondEmit),
				"TOML -> C4D -> TOML canonical text equality (D-22)")

			require.Equal(t, m2, m3, "post-DSL models exactly equal")

			require.Equal(t, canonicalModel(m1), canonicalModel(m3),
				"models equal modulo the D-22 explicit-defaults list")
		})
	}
}

// TestRoundTripC4DToTOMLToC4D is the D-22 reverse contract over the full
// corpus: emit C4D, parse it, emit TOML, parse, emit C4D again — the C4D
// text must be canonically stable across the loop.
func TestRoundTripC4DToTOMLToC4D(t *testing.T) {
	t.Parallel()

	validPaths, _ := corpusTOMLPaths(t)

	for _, path := range validPaths {
		rel, err := filepath.Rel(repoRoot(), path)
		require.NoError(t, err)

		t.Run(rel, func(t *testing.T) {
			t.Parallel()

			//nolint:gosec // G304: path from the pinned corpus walker, not user input
			//nolint:gosec // G304: path from the pinned corpus walker, not user input
			data, err := os.ReadFile(path)
			require.NoError(t, err, "read corpus fixture")

			m1, err := parser.Parse(data)
			require.NoError(t, err, "fixture parses")

			c4dFirst := c4d.EmitC4D(c4d.FromModel(m1))

			m2, err := c4d.Parse([]byte(c4dFirst))
			require.NoError(t, err, "first C4D emission re-parses")

			tomlText, err := c4d.EmitTOML(m2)
			require.NoError(t, err, "TOML emission")

			m3, err := parser.Parse([]byte(tomlText))
			require.NoError(t, err, "TOML re-parses")

			c4dSecond := c4d.EmitC4D(c4d.FromModel(m3))

			require.Equal(t,
				canonsrc.NormalizeC4D(t, c4dFirst),
				canonsrc.NormalizeC4D(t, c4dSecond),
				"C4D -> TOML -> C4D canonical stability (D-22)")
		})
	}
}

// TestUnitWidthHeightRoundTrip pins the F-02 parity fix: TOML-authored
// width/height styling (README 'Styling' — width = 300, height = 200) must
// survive the full TOML -> C4D -> TOML loop. The C4D side carries them as
// body fields, the TOML side re-serializes them, and both the canonical-text
// and model comparisons would catch a silent drop back to 0.
func TestUnitWidthHeightRoundTrip(t *testing.T) {
	t.Parallel()

	m1, err := parser.Parse([]byte(`[webapp]
type = "system"
name = "Web App"
width = 300
height = 200
`))
	require.NoError(t, err, "TOML with width/height parses")

	require.Contains(t, m1.Units, "webapp", "unit present")
	require.InDelta(t, 300.0, m1.Units["webapp"].Width, 0.0001, "TOML width parses")
	require.InDelta(t, 200.0, m1.Units["webapp"].Height, 0.0001, "TOML height parses")

	c4dText := c4d.EmitC4D(c4d.FromModel(m1))
	assert.Contains(t, c4dText, "width: 300", "C4D emission carries width")
	assert.Contains(t, c4dText, "height: 200", "C4D emission carries height")

	m2, err := c4d.Parse([]byte(c4dText))
	require.NoError(t, err, "emitted C4D re-parses")
	require.InDelta(t, 300.0, m2.Units["webapp"].Width, 0.0001, "C4D width maps into the Model")
	require.InDelta(t, 200.0, m2.Units["webapp"].Height, 0.0001, "C4D height maps into the Model")

	tomlText, err := c4d.EmitTOML(m2)
	require.NoError(t, err, "TOML emission")
	assert.Contains(t, tomlText, "width = 300", "TOML emission carries width")
	assert.Contains(t, tomlText, "height = 200", "TOML emission carries height")

	m3, err := parser.Parse([]byte(tomlText))
	require.NoError(t, err, "round-tripped TOML parses")

	firstEmit, err := c4d.EmitTOML(m1)
	require.NoError(t, err, "first TOML emission")

	require.Equal(t,
		canonsrc.NormalizeTOML(t, firstEmit),
		canonsrc.NormalizeTOML(t, tomlText),
		"width/height survive the TOML -> C4D -> TOML loop (F-02)")

	require.Equal(t, canonicalModel(m1), canonicalModel(m3),
		"models equal after the width/height round trip")
}

// TestInvalidFixturesRefuseConversion: invalid corpus fixtures assert
// conversion REFUSAL — some stage of the pre-render pipeline (parse,
// expand, peer.Resolve, validate) must error, mirroring D-24's convert gate
// at model level. No output is ever produced for them.
func TestInvalidFixturesRefuseConversion(t *testing.T) {
	t.Parallel()

	_, invalidPaths := corpusTOMLPaths(t)
	require.NotEmpty(t, invalidPaths, "invalid corpus fixtures present")

	for _, path := range invalidPaths {
		rel, err := filepath.Rel(repoRoot(), path)
		require.NoError(t, err)

		t.Run(rel, func(t *testing.T) {
			t.Parallel()

			//nolint:gosec // G304: path from the pinned corpus walker, not user input
			data, err := os.ReadFile(path)
			require.NoError(t, err, "read invalid fixture")

			m, perr := parser.Parse(data)
			if perr != nil {
				return // parse refusal IS refusal
			}

			if m, eerr := template.Expand(m); eerr == nil {
				if eerr := peer.Resolve(m); eerr == nil {
					require.NotEmpty(t, validator.Validate(m),
						"invalid fixture must fail somewhere in the D-24 gate")

					return
				}
			}
		})
	}
}

// twinExtension rewrites a .toml path to its .c4d twin (the include-graph
// conversion D-26 groundwork for --follow-includes).
func twinExtension(path string) string {
	return strings.TrimSuffix(path, ".toml") + ".c4d"
}

// convertGraphFileToC4D parses one TOML file of an include graph, rewrites
// its include paths to the twin extension, and writes the converted .c4d
// twin under dstRoot at the same relative location.
func convertGraphFileToC4D(t *testing.T, srcRoot, relPath, dstRoot string) {
	t.Helper()

	//nolint:gosec // G304/G703: path from the pinned example graph, not user input
	src, err := os.ReadFile(filepath.Join(srcRoot, relPath))
	require.NoError(t, err, "read graph file")

	m, err := parser.Parse(src)
	require.NoError(t, err, "graph file parses")

	for i := range m.Includes {
		m.Includes[i].Path = twinExtension(m.Includes[i].Path)
	}

	dst := filepath.Join(dstRoot, twinExtension(relPath))
	require.NoError(t, os.MkdirAll(filepath.Dir(dst), 0o750), "mkdir for twin")

	require.NoError(t,
		os.WriteFile( //nolint:gosec // G304/G703: paths from the pinned example graph
			dst, []byte(c4d.EmitC4D(c4d.FromModel(m))), 0o600),
		"write converted twin")
}

// composedPipeline runs the load-bearing stage order (include.Resolve ->
// template.Expand -> peer.Resolve) and returns the expanded model, ready
// for comparison.
func composedPipeline(t *testing.T, path string) *parser.Model {
	t.Helper()

	m, err := parser.ParseFile(path)
	require.NoError(t, err, "parse graph entry")

	dir := filepath.Dir(path)

	m, err = include.Resolve(m, dir, path)
	require.NoError(t, err, "include.Resolve")

	m, err = template.Expand(m)
	require.NoError(t, err, "template.Expand")

	require.NoError(t, peer.Resolve(m), "peer.Resolve")

	return m
}

// composedPipelineC4D is composedPipeline's C4D-entry twin (extension
// dispatch through include.Resolve, D-26).
func composedPipelineC4D(t *testing.T, path string) *parser.Model {
	t.Helper()

	m, err := c4d.ParseFile(path)
	require.NoError(t, err, "parse .c4d graph entry")

	dir := filepath.Dir(path)

	m, err = include.Resolve(m, dir, path)
	require.NoError(t, err, "include.Resolve")

	m, err = template.Expand(m)
	require.NoError(t, err, "template.Expand")

	require.NoError(t, peer.Resolve(m), "peer.Resolve")

	return m
}

// TestComposedGraphRoundTrip converts the 09-composed include graph (entry
// + 3 includes) file-by-file to .c4d twins with include paths rewritten to
// the twin extension (the in-test graph conversion mirroring what
// convert --follow-includes will do; D-26), then proves the converted
// graph expands and validates to the SAME model as the original.
func TestComposedGraphRoundTrip(t *testing.T) {
	t.Parallel()

	srcRoot := filepath.Join(repoRoot(), "skill", "examples", "09-composed")
	dstRoot := t.TempDir()

	// The graph: entry plus its transitive includes (templates.toml,
	// domains/auth.toml). single-file-equivalent.toml is documentation, not
	// part of the graph.
	for _, rel := range []string{"entry.toml", "templates.toml", filepath.Join("domains", "auth.toml")} {
		convertGraphFileToC4D(t, srcRoot, rel, dstRoot)
	}

	// Compare BEFORE validating: Validate mutates the models in place
	// (populateIncomingLinks appends mirrors in map-iteration order), which
	// would make the comparison itself nondeterministic.
	original := composedPipeline(t, filepath.Join(srcRoot, "entry.toml"))
	converted := composedPipelineC4D(t, filepath.Join(dstRoot, "entry.c4d"))

	require.Equal(t, canonicalModel(original), canonicalModel(converted),
		"converted include graph expands to the same model (D-22/D-26)")

	// D-24 gate on fresh parses: both graphs validate clean end to end.
	require.Empty(t, validator.Validate(composedPipeline(t, filepath.Join(srcRoot, "entry.toml"))),
		"original graph validates")
	require.Empty(t, validator.Validate(composedPipelineC4D(t, filepath.Join(dstRoot, "entry.c4d"))),
		"converted graph validates (D-26)")
}

// renderEquivalenceFixtures is the pinned render set (DI-1): the four
// classic corpus fixtures plus the six testdata/c4d edge cases (paths are
// repo-root-relative; the loop joins repoRoot()).
//
//nolint:gochecknoglobals // immutable pinned fixture set
var renderEquivalenceFixtures = []string{
	"testdata/valid.toml",
	"testdata/nested.toml",
	"testdata/links.toml",
	"testdata/template_basic.toml",
	"testdata/c4d/external-types.toml",
	"testdata/c4d/linkfrom.toml",
	"testdata/c4d/multiline-strings.toml",
	"testdata/c4d/rank-equal.toml",
	"testdata/c4d/template-nested-use.toml",
	"testdata/c4d/unicode-strings.toml",
}

// expandedPathsOf replicates cmd/c4drill's collectExpandedPaths: C1 plus
// every unit with subunits, recursively.
func expandedPathsOf(m *parser.Model) []string {
	paths := []string{""}

	walkFixtureUnits(m, func(path string, u *model.Unit) {
		if len(u.Subunits) > 0 {
			paths = append(paths, path)
		}
	})

	return paths
}

// renderAllViewsDOT renders every auto-detected view of m to DOT through
// the real pipeline composition (views -> graph -> render, format=dot) and
// returns the per-view canonicalDOT forms joined in view order — DI-1.
func renderAllViewsDOT(t *testing.T, m *parser.Model) string {
	t.Helper()

	// The CLI default ratio; both sides render under it.
	render.LabelRatio = 1.6

	parts := make([]string, 0, 4)

	for _, path := range expandedPathsOf(m) {
		var v *view.View

		switch {
		case path == "":
			v = view.GenerateC1View(m)
		case !strings.Contains(path, "."):
			v = view.GenerateC2View(m, path)
		default:
			v = view.GenerateC3View(m, path)
		}

		require.NotNil(t, v, "view for %q", path)

		g := graph.BuildGraphWithPath(v, path, "parity", "dot")
		require.NotNil(t, g, "graph for %q", path)

		data, err := render.Render(g, "dot")
		require.NoError(t, err, "render view %q", path)

		parts = append(parts, canonical.Canonical(t, string(data)))
	}

	return strings.Join(parts, "\x00view\x00")
}

// pipelineForRender runs expand + peer.Resolve (validation is NOT required:
// the classic fixtures include by-design orphan units like testdata/valid's
// unlinked pair, which the render path itself does not depend on).
func pipelineForRender(t *testing.T, m *parser.Model) *parser.Model {
	t.Helper()

	m, err := template.Expand(m)
	require.NoError(t, err, "expand for render")

	require.NoError(t, peer.Resolve(m), "peer.Resolve for render")

	return m
}

// expectedExampleTwins is the pinned manifest of the skill/examples .c4d
// twins (D-35): every listed .toml must have a fmt-clean .c4d twin on disk
// that renders identically to its source. Paths are relative to
// skill/examples/ with slash separators (OS-normalized at compare time).
//
//nolint:gochecknoglobals // immutable twin manifest, the examples contract
var expectedExampleTwins = []string{
	"02-nested.toml",
	"03-links.toml",
	"04-styling.toml",
	"06-templates.toml",
	"07-relative-peer.toml",
	"08-include/auth.toml",
	"08-include/billing.toml",
	"08-include/entry.toml",
	"09-composed/domains/auth.toml",
	"09-composed/entry.toml",
	"09-composed/single-file-equivalent.toml",
	"09-composed/templates.toml",
	"10-edge-kinds.toml",
}

// exampleTwin is one .toml/.c4d pair found under skill/examples/.
type exampleTwin struct {
	tomlPath string // absolute .toml source
	c4dPath  string // absolute .c4d twin (same path, extension swapped)
	rel      string // slash-separated path relative to skill/examples/
}

// walkExampleTwins collects every .toml under skill/examples/ whose .c4d
// twin exists on disk, pinned against expectedExampleTwins (anti-shrinkage:
// dropping a twin fails by name).
func walkExampleTwins(t *testing.T) []exampleTwin {
	t.Helper()

	root := filepath.Join(repoRoot(), "skill", "examples")

	var rels []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(d.Name(), ".toml") {
			return nil
		}

		twin := strings.TrimSuffix(path, ".toml") + ".c4d"
		if _, statErr := os.Stat(twin); statErr == nil {
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return fmt.Errorf("example twin rel: %w", relErr)
			}

			rels = append(rels, filepath.ToSlash(rel))
		}

		// A .toml without a twin on disk (01-minimal, 05-ecommerce) is not a
		// pair — skip it.
		return nil
	})
	require.NoError(t, err, "walk skill/examples for twin pairs")

	slices.Sort(rels)
	require.Equal(t, expectedExampleTwins, rels,
		"the shipped .c4d twin set must match the pinned manifest exactly")

	twins := make([]exampleTwin, 0, len(rels))
	for _, rel := range rels {
		tomlPath := filepath.Join(root, filepath.FromSlash(rel))
		twins = append(twins, exampleTwin{
			tomlPath: tomlPath,
			c4dPath:  strings.TrimSuffix(tomlPath, ".toml") + ".c4d",
			rel:      rel,
		})
	}

	return twins
}

// readExampleFile reads one example pair member.
func readExampleFile(t *testing.T, path string) []byte {
	t.Helper()

	//nolint:gosec // G304: path from the pinned twins walker, not user input
	data, err := os.ReadFile(path)
	require.NoError(t, err, "read example %s", path)

	return data
}

// assertTwinModelParity: the two sides parse to canonically-equal models
// (the D-22 explicit-defaults list applied to both sides).
func assertTwinModelParity(t *testing.T, twin exampleTwin) {
	t.Helper()

	mToml, err := parser.Parse(readExampleFile(t, twin.tomlPath))
	require.NoError(t, err, "%s parses", twin.rel)

	mC4D, err := c4d.Parse(readExampleFile(t, twin.c4dPath))
	require.NoError(t, err, "%s twin parses", twin.rel)

	require.Equal(t, canonicalModel(mToml), canonicalModel(mC4D),
		"twin parses to the same model as its .toml source (D-22): %s", twin.rel)
}

// assertTwinRenderParity: both sides render to canonically-equal DOT through
// the real view/graph/render composition (DI-1). For standalone files only.
func assertTwinRenderParity(t *testing.T, twin exampleTwin) {
	t.Helper()

	mToml, err := parser.Parse(readExampleFile(t, twin.tomlPath))
	require.NoError(t, err, "%s parses", twin.rel)

	mC4D, err := c4d.Parse(readExampleFile(t, twin.c4dPath))
	require.NoError(t, err, "%s twin parses", twin.rel)

	require.Equal(t,
		renderAllViewsDOT(t, pipelineForRender(t, mToml)),
		renderAllViewsDOT(t, pipelineForRender(t, mC4D)),
		"twin renders identically to its .toml source (DI-1): %s", twin.rel)
}

// assertTwinGraphParity: for include-graph entries — the twin graph must be
// self-contained (every include path references a .c4d twin), resolve to the
// same post-include model, validate clean, and render identically.
func assertTwinGraphParity(t *testing.T, twin exampleTwin) {
	t.Helper()

	mC4DRaw, err := c4d.Parse(readExampleFile(t, twin.c4dPath))
	require.NoError(t, err, "%s twin parses", twin.rel)

	for _, inc := range mC4DRaw.Includes {
		require.Equal(t, extC4D, filepath.Ext(inc.Path),
			"twin include graph is self-contained (.c4d references only): %s", twin.rel)
	}

	original := composedPipeline(t, twin.tomlPath)
	converted := composedPipelineC4D(t, twin.c4dPath)

	require.Equal(t, canonicalModel(original), canonicalModel(converted),
		"twin graph resolves to the same model as the .toml graph: %s", twin.rel)

	require.Equal(t,
		renderAllViewsDOT(t, original),
		renderAllViewsDOT(t, converted),
		"twin graph renders identically to the .toml graph (DI-1): %s", twin.rel)

	require.Empty(t, validator.Validate(composedPipeline(t, twin.tomlPath)),
		"original graph validates: %s", twin.rel)
	require.Empty(t, validator.Validate(composedPipelineC4D(t, twin.c4dPath)),
		"twin graph validates: %s", twin.rel)
}

// extC4D names the C4D file extension in assertions.
const extC4D = ".c4d"

// TestExampleTwins is the D-35 twins contract: every .toml/.c4d pair under
// skill/examples/ parses to the same model; standalone files and include
// graphs additionally render to canonically-equal DOT (DI-1). Include-graph
// fragments (targets of some entry's include) get model parity — the entry
// graph render covers them end to end. Rendering makes the test serial.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestExampleTwins(t *testing.T) {
	twins := walkExampleTwins(t)

	// Classify: entries declare includes; fragments are entry include targets.
	includeTargets := make(map[string]bool)

	for _, twin := range twins {
		m, err := parser.Parse(readExampleFile(t, twin.tomlPath))
		require.NoError(t, err, "%s parses", twin.rel)

		for _, inc := range m.Includes {
			target := filepath.Join(filepath.Dir(twin.tomlPath), inc.Path)
			includeTargets[filepath.Clean(target)] = true
		}
	}

	for _, twin := range twins {
		t.Run(twin.rel, func(t *testing.T) {
			m, err := parser.Parse(readExampleFile(t, twin.tomlPath))
			require.NoError(t, err, "%s parses", twin.rel)

			isFragment := includeTargets[filepath.Clean(twin.tomlPath)]

			switch {
			case len(m.Includes) > 0:
				assertTwinGraphParity(t, twin)
			case isFragment:
				assertTwinModelParity(t, twin)
			default:
				assertTwinModelParity(t, twin)
				assertTwinRenderParity(t, twin)
			}
		})
	}
}

// TestRenderEquivalence proves Model-level equivalence end to end (DI-1):
// for the pinned set, the .toml Model and the round-tripped .c4d Model
// render to canonically-equal DOT through the real view/graph/render
// composition.
//
//nolint:paralleltest // go-graphviz WASM engine has concurrency issues
func TestRenderEquivalence(t *testing.T) {
	for _, rel := range renderEquivalenceFixtures {
		path := filepath.Join(repoRoot(), rel)

		//nolint:gosec // G304: path from the pinned render set, not user input
		data, err := os.ReadFile(path)
		require.NoError(t, err, "read render fixture %s", rel)

		t.Run(rel, func(t *testing.T) {
			m1, err := parser.Parse(data)
			require.NoError(t, err, "fixture parses")

			m2, err := c4d.Parse([]byte(c4d.EmitC4D(c4d.FromModel(m1))))
			require.NoError(t, err, "round-tripped .c4d parses")

			dotTOML := renderAllViewsDOT(t, pipelineForRender(t, m1))
			dotC4D := renderAllViewsDOT(t, pipelineForRender(t, m2))

			require.Equal(t, dotTOML, dotC4D,
				"canonicalDOT equality of .toml and round-tripped .c4d renders (DI-1)")
		})
	}
}
