package c4d_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/peer"
	"github.com/Djarvur/c4drill/internal/template"
	"github.com/Djarvur/c4drill/internal/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file is the D-22/D-26 parity contract suite (Plan 35-06). The
// fixture guards below pin the edge-case corpus under testdata/c4d/; the
// round-trip and render-equivalence suites extend this file in Task 3.

// c4dFixtureDir is the edge-case fixture corpus root.
const c4dFixtureDir = "../../testdata/c4d"

// expectedC4DFixtures is the pinned corpus manifest — every file must exist,
// parse, and pass the full pre-render pipeline. The list doubles as the
// walker's anti-shrinkage guard: dropping a fixture fails by name.
var expectedC4DFixtures = []string{
	"external-types.toml",
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
			if utf8.RuneCountInString(value) > len(value) {
				multibyte = true
			}

			longWord = longestWord(longWord, value)
		}
	})

	assert.True(t, multibyte, "multi-byte values present")
	assert.Greater(t, utf8.RuneCountInString(longWord), len(longWord)/2,
		"long single word is multi-byte: got %q", longWord)
}
