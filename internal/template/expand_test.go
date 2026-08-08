package template_test

import (
	"os"
	"testing"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/template"
	"github.com/Djarvur/c4drill/internal/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parseFixture reads and parses a testdata fixture, failing the test on error.
func parseFixture(t *testing.T, name string) *parser.Model {
	t.Helper()

	//nolint:gosec // G304: name is a hard-coded fixture filename in testdata/, not user input
	data, err := os.ReadFile("../../testdata/" + name)
	require.NoError(t, err, "failed to read %s", name)

	m, err := parser.Parse(data)
	require.NoError(t, err, "Parse(%s) should not error", name)

	return m
}

// TestExpandBasic exercises TMPL-01/02/03/05: instantiating a template with all
// declared params produces a concrete unit subtree that passes validator.Validate.
// Substitution applies to Name/Technology/Description/Color and Link fields.
func TestExpandBasic(t *testing.T) {
	t.Parallel()

	m := parseFixture(t, "template_basic.toml")

	expanded, err := template.Expand(m)
	require.NoError(t, err, "Expand should not error")

	// Produced unit "auth" present at top level (no parent).
	require.Contains(t, expanded.Units, "auth", "expanded Units must contain 'auth'")
	assert.Contains(t, expanded.UnitOrder, "auth", "expanded UnitOrder must contain 'auth'")

	auth := expanded.Units["auth"]
	assert.Equal(t, "auth Service", auth.Name, "auth.Name (substituted)")
	assert.Equal(t, "Go", auth.Technology, "auth.Technology (substituted)")
	assert.Equal(t, "auth handles requests", auth.Description, "auth.Description (substituted)")
	assert.Equal(t, "auth-color", auth.Color, "auth.Color (substituted)")

	// Link peer + description + technology substituted.
	require.Len(t, auth.Links, 1, "auth.Links length")
	assert.Equal(t, "messageBus", auth.Links[0].Peer, "auth.Links[0].Peer (substituted)")
	assert.Equal(t, "Publishes auth events", auth.Links[0].Description, "auth.Links[0].Description (substituted)")
	assert.Equal(t, "Go", auth.Links[0].Technology, "auth.Links[0].Technology (substituted)")

	// Expanded model passes validation.
	valErrors := validator.Validate(expanded)
	assert.Empty(t, valErrors, "expanded model should validate cleanly")
}

// TestExpandSubstitutionAllFields (TMPL-03): no "${" remains anywhere in any
// string field of any produced unit (root + links).
func TestExpandSubstitutionAllFields(t *testing.T) {
	t.Parallel()

	m := parseFixture(t, "template_basic.toml")

	expanded, err := template.Expand(m)
	require.NoError(t, err, "Expand should not error")

	auth := expanded.Units["auth"]
	require.NotNil(t, auth, "auth unit present")

	assertNotContainsSubst(t, auth)
}

// assertNotContainsSubst asserts no "${" remains in any string field of u or its
// links/subunits. Recurses into subunits.
func assertNotContainsSubst(t *testing.T, u *model.Unit) {
	t.Helper()

	assert.NotContains(t, u.Name, "${", "Name has no residual substitution token")
	assert.NotContains(t, u.Description, "${", "Description has no residual substitution token")
	assert.NotContains(t, u.Technology, "${", "Technology has no residual substitution token")
	assert.NotContains(t, u.Reference, "${", "Reference has no residual substitution token")
	assert.NotContains(t, u.Color, "${", "Color has no residual substitution token")
	assert.NotContains(t, u.Style, "${", "Style has no residual substitution token")
	assert.NotContains(t, u.Border, "${", "Border has no residual substitution token")
	assert.NotContains(t, u.Edges, "${", "Edges has no residual substitution token")

	for i := range u.Links {
		assert.NotContains(t, u.Links[i].Peer, "${", "Links[%d].Peer has no residual token", i)
		assert.NotContains(t, u.Links[i].Description, "${", "Links[%d].Description has no residual token", i)
		assert.NotContains(t, u.Links[i].Technology, "${", "Links[%d].Technology has no residual token", i)
	}

	for _, sub := range u.Subunits {
		assertNotContainsSubst(t, sub)
	}
}

// TestExpandSubtree (TMPL-04): a template declaring a subunit subtree expands
// whole, subunit keys verbatim, fields substituted.
func TestExpandSubtree(t *testing.T) {
	t.Parallel()

	m := parseFixture(t, "template_subtree.toml")

	expanded, err := template.Expand(m)
	require.NoError(t, err, "Expand should not error")

	require.Contains(t, expanded.Units, "auth", "expanded Units must contain 'auth'")
	auth := expanded.Units["auth"]

	// Subunits present with verbatim keys.
	require.Contains(t, auth.Subunits, "api", "auth has 'api' subunit (verbatim key)")
	require.Contains(t, auth.Subunits, "db", "auth has 'db' subunit (verbatim key)")

	api := auth.Subunits["api"]
	assert.Equal(t, "auth API", api.Name, "api.Name (substituted)")
	assert.Equal(t, "Go", api.Technology, "api.Technology (substituted)")

	db := auth.Subunits["db"]
	assert.Equal(t, "auth DB", db.Name, "db.Name (substituted)")
	assert.Equal(t, "PostgreSQL", db.Technology, "db.Technology (NOT substituted — no ${tech} token)")

	// SubunitOrder preserved.
	assert.Equal(t, []string{"api", "db"}, auth.SubunitOrder, "auth.SubunitOrder (verbatim)")

	// No residual tokens anywhere.
	assertNotContainsSubst(t, auth)
}

// TestExpandThreeInstantiationsHS1 is THE load-bearing regression test
// (TMPL-08 / HS-1). Instantiating one template 3x with distinct params must
// yield 3 INDEPENDENT unit subtrees. After validator.Validate, the three
// instantiations' LinksFrom slices MUST be disjoint (no shared Link elements) —
// the validator mutates LinksFrom in place (index.go:70-81), so a shallow Clone
// would corrupt the 2nd/3rd instantiation. Re-expand on a fresh Parse must be
// idempotent (deeply equal output).
func TestExpandThreeInstantiationsHS1(t *testing.T) {
	t.Parallel()

	m := parseFixture(t, "template_3x_instantiate.toml")

	expanded, err := template.Expand(m)
	require.NoError(t, err, "Expand should not error")

	// Three distinct top-level units.
	require.Contains(t, expanded.Units, "alpha", "alpha present")
	require.Contains(t, expanded.Units, "beta", "beta present")
	require.Contains(t, expanded.Units, "gamma", "gamma present")

	alpha := expanded.Units["alpha"]
	beta := expanded.Units["beta"]
	gamma := expanded.Units["gamma"]

	// Each has an outgoing link to messageBus with a distinct description.
	assertLinkPeer := func(u *model.Unit, wantDesc string) {
		t.Helper()
		require.Len(t, u.Links, 1, "unit has one outgoing link")
		assert.Equal(t, "messageBus", u.Links[0].Peer, "link Peer")
		assert.Equal(t, wantDesc, u.Links[0].Description, "link Description (substituted per instantiation)")
	}

	assertLinkPeer(alpha, "Publishes alpha events")
	assertLinkPeer(beta, "Publishes beta events")
	assertLinkPeer(gamma, "Publishes gamma events")

	// Run the validator — it populates LinksFrom on messageBus in place, and
	// (critically) each of alpha/beta/gamma must have its OWN independent
	// LinksFrom slice (no shared backing array).
	valErrors := validator.Validate(expanded)
	assert.Empty(t, valErrors, "expanded model should validate cleanly")

	// HS-1: collect the validator-synthesized LinksFrom mirror entries pointing
	// at each instantiation. messageBus should have THREE distinct mirror
	// LinksFrom entries (one from each of alpha/beta/gamma). If Clone were
	// shallow, the mirror entries would alias across instantiations.
	bus := expanded.Units["messageBus"]
	require.NotNil(t, bus, "messageBus present")

	var mirrorSources []string

	for i := range bus.LinksFrom {
		if bus.LinksFrom[i].IsMirror() {
			mirrorSources = append(mirrorSources, bus.LinksFrom[i].Peer)
		}
	}

	assert.ElementsMatch(t,
		[]string{"alpha", "beta", "gamma"},
		mirrorSources,
		"messageBus.LinksFrom must carry one mirror per instantiation (3 disjoint entries, HS-1)",
	)

	// Idempotency: re-Parse + re-Expand the same input; the two expanded Models
	// must be deeply equal (UnitOrder + Units structure).
	m2 := parseFixture(t, "template_3x_instantiate.toml")
	expanded2, err := template.Expand(m2)
	require.NoError(t, err, "second Expand should not error")

	assert.ElementsMatch(t, expanded.UnitOrder, expanded2.UnitOrder, "idempotent UnitOrder")
	assertExpandedEqual(t, expanded, expanded2)
}

// assertExpandedEqual asserts two expanded models have equal Units maps by
// comparing each unit's exported fields. Used for idempotency checks.
func assertExpandedEqual(t *testing.T, a, b *parser.Model) {
	t.Helper()

	assert.Len(t, b.Units, len(a.Units), "unit count equal")

	for name, ua := range a.Units {
		ub, ok := b.Units[name]
		require.True(t, ok, "unit %q present in both", name)
		assertUnitEqual(t, ua, ub, name)
	}
}

// assertUnitEqual asserts two units are field-equal (recursing into subunits).
func assertUnitEqual(t *testing.T, want, got *model.Unit, path string) {
	t.Helper()

	assert.Equal(t, want.Type, got.Type, "%s.Type", path)
	assert.Equal(t, want.Name, got.Name, "%s.Name", path)
	assert.Equal(t, want.Description, got.Description, "%s.Description", path)
	assert.Equal(t, want.Technology, got.Technology, "%s.Technology", path)
	assert.Equal(t, want.Reference, got.Reference, "%s.Reference", path)
	assert.Equal(t, want.Color, got.Color, "%s.Color", path)
	assert.ElementsMatch(t, want.SubunitOrder, got.SubunitOrder, "%s.SubunitOrder", path)
	assert.Len(t, got.Subunits, len(want.Subunits), "%s.Subunits count", path)

	for subName, wSub := range want.Subunits {
		gSub, ok := got.Subunits[subName]
		require.True(t, ok, "%s.%s present", path, subName)
		assertUnitEqual(t, wSub, gSub, path+"."+subName)
	}

	assert.Len(t, got.Links, len(want.Links), "%s.Links count", path)

	for i := range want.Links {
		if i < len(got.Links) {
			assert.Equal(t, want.Links[i].Peer, got.Links[i].Peer, "%s.Links[%d].Peer", path, i)
			assert.Equal(t, want.Links[i].Description, got.Links[i].Description, "%s.Links[%d].Description", path, i)
			assert.Equal(t, want.Links[i].Technology, got.Links[i].Technology, "%s.Links[%d].Technology", path, i)
		}
	}
}

// TestExpandMissingParamNames (TMPL-06): a [[use]] missing a declared param is
// a hard error whose message contains the template name, the param name, and an
// instantiation-site identifier.
func TestExpandMissingParamNames(t *testing.T) {
	t.Parallel()

	m := parseFixture(t, "template_missing_param.toml")

	_, err := template.Expand(m)
	require.Error(t, err, "missing declared param must be a hard error")

	msg := err.Error()
	assert.Contains(t, msg, "svc", "error names the template")
	assert.Contains(t, msg, "tech", "error names the missing param")
	// Instantiation-site identifier: the [[use]]'s name param 'auth' or its index.
	siteContains := contains(msg, "auth") || contains(msg, "[[use]]") || contains(msg, "use")
	assert.True(t, siteContains, "error identifies the instantiation site (got: %q)", msg)
}

// contains reports whether s contains sub. A tiny helper avoiding
// strings.Contains import noise in a test file.
func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}

	return false
}

// TestExpandDuplicatePath (TMPL-07): two [[use]] producing the same path is a
// hard error naming both sources.
func TestExpandDuplicatePath(t *testing.T) {
	t.Parallel()

	m := parseFixture(t, "template_duplicate_path.toml")

	_, err := template.Expand(m)
	require.Error(t, err, "duplicate path must be a hard error")

	msg := err.Error()
	assert.Contains(t, msg, "auth", "error names the conflicting path")
}

// TestExpandParentPlacement (XC-03): a [[use]] with parent places the produced
// unit as a CHILD of the parent, so the future relative-peer pass resolves its
// peers against the instantiation site.
func TestExpandParentPlacement(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Parent Placement Test"

[linuxSystem]
type = "system"
name = "Linux System"

[template.svc]
params = ["name"]
name = "${name} Service"
type = "container"

[[use]]
template = "svc"
parent = "linuxSystem"
name = "auth"
`)

	m, err := parser.Parse(data)
	require.NoError(t, err, "Parse should not error")

	expanded, err := template.Expand(m)
	require.NoError(t, err, "Expand should not error")

	// 'auth' is NOT top-level — it's a child of linuxSystem.
	assert.NotContains(t, expanded.Units, "auth", "auth must NOT be top-level (placed under parent)")
	assert.Contains(t, expanded.Units, "linuxSystem", "linuxSystem present")

	linuxSys := expanded.Units["linuxSystem"]
	require.Contains(t, linuxSys.Subunits, "auth", "linuxSystem.Subunits contains 'auth'")
	assert.Equal(t, "auth Service", linuxSys.Subunits["auth"].Name, "auth.Name (substituted)")
	assert.Contains(t, linuxSys.SubunitOrder, "auth", "linuxSystem.SubunitOrder contains 'auth'")
}

// TestExpandForwardRef (TMPL-09): a [[use]] before [template.*] expands
// successfully (Plan 01 already parses this; Expand consumes it).
func TestExpandForwardRef(t *testing.T) {
	t.Parallel()

	m := parseFixture(t, "template_forward_ref.toml")

	expanded, err := template.Expand(m)
	require.NoError(t, err, "forward-ref Expand should not error")
	require.Contains(t, expanded.Units, "auth", "forward-ref produced 'auth'")
	assert.Equal(t, "auth Service", expanded.Units["auth"].Name, "auth.Name (substituted)")
	assert.Equal(t, "Go", expanded.Units["auth"].Technology, "auth.Technology (substituted)")
}

// TestExpandNoOpOnHandAuthored verifies Expand is a no-op (returns the model
// unchanged, nil error) when there are no templates/instantiations — guarantee
// of no regression for hand-authored-only models.
func TestExpandNoOpOnHandAuthored(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/valid.toml")
	require.NoError(t, err)

	m, err := parser.Parse(data)
	require.NoError(t, err)

	expanded, err := template.Expand(m)
	require.NoError(t, err, "Expand on hand-authored model must not error")

	// No templates → no new units; structure identical.
	assert.Equal(t, []string{"user", "webapp"}, expanded.UnitOrder, "UnitOrder unchanged")
	assert.Len(t, expanded.Units, 2, "Units count unchanged")
}

// TestPipelineExpandBeforeValidate (XC-04): an end-to-end test exercising the
// Parse → Expand → Validate sequence on a templated fixture. This proves the
// pipeline wiring in cmd/c4drill/root.go places Expand between Parse and
// Validate. (The literal wiring is in root.go; this test exercises the same
// sequence the pipeline runs.)
func TestPipelineExpandBeforeValidate(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/template_basic.toml")
	require.NoError(t, err)

	// Stage 1: Parse
	m, err := parser.Parse(data)
	require.NoError(t, err, "Parse should not error")

	// Stage 1.5: Expand (the new pipeline stage)
	m, err = template.Expand(m)
	require.NoError(t, err, "Expand should not error")

	// Stage 2: Validate (post-expand)
	valErrors := validator.Validate(m)
	assert.Empty(t, valErrors, "expanded model validates cleanly (pipeline order correct)")

	// Produced unit is present and rendered-ready.
	require.Contains(t, m.Units, "auth", "auth present after pipeline")
	assert.Equal(t, "auth Service", m.Units["auth"].Name, "auth.Name substituted")
}
