// Package canonsrc_test pins the D-22 canonical-source normalizers:
// NormalizeTOML and NormalizeC4D must make representation differences
// (whitespace, comments, quoting, key order, explicit defaults, newline
// spellings) normalize away while keeping genuinely different documents
// distinct — the anti-vacuity guard every equality test below pairs with a
// negative case.
package canonsrc_test

import (
	"testing"

	"github.com/Djarvur/c4drill/internal/testutil/canonsrc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNormalizeTOMLKeyOrderQuotingAndTrivia: two TOML spellings of the same
// document — reordered keys, single- vs double-quoted strings, comments and
// blank lines — normalize to ONE canonical form (D-22). The plan's schematic
// "a = x / b = 2" pair is expressed over real c4drill documents: top-level
// bare scalars are outside the model (parser.Parse skips non-table
// expressions), so a vacuous pair would prove nothing.
func TestNormalizeTOMLKeyOrderQuotingAndTrivia(t *testing.T) {
	t.Parallel()

	left := `[properties]
name = "Demo"
description = "Two units"

[a]
type = "system"
name = "A"

[[a.link]]
peer = "b"
technology = "HTTP"
`

	right := `[properties]
description = 'Two units'   # reordered, single-quoted, tailed
name = "Demo"

[a]
name = "A"
type = 'system'

[[a.link]]
technology = "HTTP"
peer = "b"
`

	gotLeft := canonsrc.NormalizeTOML(t, left)
	gotRight := canonsrc.NormalizeTOML(t, right)

	require.Equal(t, gotLeft, gotRight, "key order, quoting, comments and blank lines normalize away")

	assert.Contains(t, gotLeft, `name = "Demo"`, "canonical form carries properties.name")
	assert.Contains(t, gotLeft, `peer = "b"`, "canonical form carries link peer")
}

// TestNormalizeTOMLDistinctDocumentsDiffer is the anti-vacuity guard: a
// different type, name or link target must produce a DIFFERENT canonical
// form (equality tests elsewhere are only meaningful because of this).
func TestNormalizeTOMLDistinctDocumentsDiffer(t *testing.T) {
	t.Parallel()

	base := canonsrc.NormalizeTOML(t, `[a]
type = "system"
name = "A"

[[a.link]]
peer = "b"
`)

	for name, doc := range map[string]string{
		"type": `[a]
type = "db"
name = "A"

[[a.link]]
peer = "b"
`,
		"name": `[a]
type = "system"
name = "Different"

[[a.link]]
peer = "b"
`,
		"peer": `[a]
type = "system"
name = "A"

[[a.link]]
peer = "c"
`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.NotEqual(t, base, canonsrc.NormalizeTOML(t, doc),
				"different %s must normalize to a different form", name)
		})
	}
}

// TestNormalizeTOMLExplicitDefaultsNormalizeAway pins the D-22 defaults list:
// a link field set to its default representation equals the omitted field —
// arrow = "forward", rank = "forward", labelPosition = "middle" — on both
// [[...link]] and [[...linkFrom]] tables.
func TestNormalizeTOMLExplicitDefaultsNormalizeAway(t *testing.T) {
	t.Parallel()

	omitted := `[a]
type = "system"

[[a.link]]
peer = "b"

[[a.linkFrom]]
peer = "c"
`

	explicit := `[a]
type = "system"

[[a.link]]
peer = "b"
arrow = "forward"
rank = "forward"
labelPosition = "middle"

[[a.linkFrom]]
peer = "c"
arrow = "forward"
rank = "forward"
labelPosition = "middle"
`

	require.Equal(t, canonsrc.NormalizeTOML(t, omitted), canonsrc.NormalizeTOML(t, explicit),
		"explicit defaults equal omitted fields (D-22 defaults list)")

	// Non-default values stay distinct.
	assert.NotEqual(t,
		canonsrc.NormalizeTOML(t, omitted),
		canonsrc.NormalizeTOML(t, `[a]
type = "system"

[[a.link]]
peer = "b"
arrow = "reverse"
`),
		"arrow = reverse is NOT the default and must not normalize away")
}

// TestNormalizeTOMLNewlineRepresentations: the SAME value authored as a
// single-line string with escaped \n and as a TOML multi-line basic string
// (""") normalizes to one canonical form (D-06/D-22 newline rule).
func TestNormalizeTOMLNewlineRepresentations(t *testing.T) {
	t.Parallel()

	escaped := `[a]
type = "system"
description = "line one\nline two"
`

	multiline := `[a]
type = "system"
description = """
line one
line two"""
`

	require.Equal(t, canonsrc.NormalizeTOML(t, escaped), canonsrc.NormalizeTOML(t, multiline),
		"escaped-\\n and multi-line basic string are one canonical value")

	assert.NotEqual(t,
		canonsrc.NormalizeTOML(t, escaped),
		canonsrc.NormalizeTOML(t, `[a]
type = "system"
description = "line one X line two"
`),
		"a different value must normalize differently")
}

// TestNormalizeTOMLFixpoint: normalize(normalize(x)) == normalize(x) — the
// canonical form is valid TOML that re-parses to the same canonical form.
func TestNormalizeTOMLFixpoint(t *testing.T) {
	t.Parallel()

	for name, doc := range map[string]string{
		"units and links": `[properties]
name = "Demo"
expanded = ["a"]

[a]
type = "system"

[[a.link]]
peer = "b"
technology = "HTTP"

[b]
type = "db"
`,
		"templates and uses": `[template.svc]
params = ["name", "tech"]
type = "container"
technology = "${tech}"

[[use]]
template = "svc"
name = "auth"
tech = "Go"

[[include]]
path = "other.toml"
once = true
`,
		"newlines": `[a]
type = "system"
description = "one\ntwo"
`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			once := canonsrc.NormalizeTOML(t, doc)
			assert.Equal(t, once, canonsrc.NormalizeTOML(t, once), "normalization is a fixpoint")
		})
	}
}

// TestNormalizeC4DWhitespaceAndComments: whitespace- and comment-insensitive
// canonical form — the plan's `db:db{description:cache}` pair — plus the
// fixpoint property over the canonical form itself.
func TestNormalizeC4DWhitespaceAndComments(t *testing.T) {
	t.Parallel()

	left := "properties{name:Demo}\ndb:db{description:cache}"

	right := "# lead\nproperties { name: Demo } # tail\n\ndb: db { description: cache }"

	gotLeft := canonsrc.NormalizeC4D(t, left)
	gotRight := canonsrc.NormalizeC4D(t, right)

	require.Equal(t, gotLeft, gotRight, "comments and whitespace normalize away")

	assert.Contains(t, gotLeft, "description", "canonical form carries the field")
	assert.Equal(t, gotLeft, canonsrc.NormalizeC4D(t, gotLeft), "normalization is a fixpoint")
}

// TestNormalizeC4DDocumentShape: the canonical form is a deterministic
// per-statement serialization — distinct documents differ, and the shape
// (header, edge, template, use, include) survives normalization.
func TestNormalizeC4DDocumentShape(t *testing.T) {
	t.Parallel()

	base := canonsrc.NormalizeC4D(t, `properties { name: Demo }

api: system "API" {
	-> db: sql | queries { rank: equal length: 2 }
	<- web
}

template svc(name) {
	technology: ${name}
	use other(name: x)
}

use svc(name: auth)

include ./other.c4d once
`)

	for name, needle := range map[string]string{
		"unit header":  `api: system "API"`,
		"edge":         `-> db: "sql | queries" { length: "2" rank: "equal" }`,
		"incoming":     `<- web`,
		"template":     `template svc(name) {`,
		"template use": `use other(name: "x")`,
		"top use":      `use svc(name: "auth")`,
		"include":      `include "./other.c4d" once`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Contains(t, base, needle, "canonical C4D form carries the statement")
		})
	}

	assert.NotEqual(t, base, canonsrc.NormalizeC4D(t, `properties { name: Other }`),
		"anti-vacuity: different documents normalize differently")
}

// TestNormalizeC4DNewlineRepresentations: a value authored as a C4D
// triple-quoted string and as an escaped-\n double-quoted string normalizes
// to one canonical form (D-06/D-22).
func TestNormalizeC4DNewlineRepresentations(t *testing.T) {
	t.Parallel()

	triple := "a: system {\n\tdescription: \"\"\"\nline one\nline two\"\"\"\n}"

	escaped := `a: system {
	description: "line one\nline two"
}`

	require.Equal(t, canonsrc.NormalizeC4D(t, triple), canonsrc.NormalizeC4D(t, escaped),
		"triple-quoted and escaped-\\n are one canonical value")
}

// TestNormalizeC4DQuotingKindsNormalizeAway: bareword vs quoted literal
// kinds drop away — the VALUE is canonical, not the spelling.
func TestNormalizeC4DQuotingKindsNormalizeAway(t *testing.T) {
	t.Parallel()

	bare := `a: system { description: serves traffic }`

	quoted := `a: system { description: "serves traffic" }`

	require.Equal(t, canonsrc.NormalizeC4D(t, bare), canonsrc.NormalizeC4D(t, quoted),
		"literal kind normalizes away, the value remains")
}
