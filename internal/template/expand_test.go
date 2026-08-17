package template_test

import (
	"fmt"
	"os"
	"strings"
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

// TestExpandReferenceParamSubstitution (TMPL-10): the reference field
// substitutes params correctly so reference URLs can be parameterized
// (e.g. reference = "https://wiki/${name}" -> "https://wiki/auth").
func TestExpandReferenceParamSubstitution(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Reference Param Test"

[template.svc]
params = ["name"]
name = "${name} Service"
type = "system"
reference = "https://wiki.example.com/${name}"

[[use]]
template = "svc"
name = "auth"
`)

	m, err := parser.Parse(data)
	require.NoError(t, err, "Parse should not error")

	expanded, err := template.Expand(m)
	require.NoError(t, err, "Expand should not error")

	require.Contains(t, expanded.Units, "auth", "auth present")
	auth := expanded.Units["auth"]
	assert.Equal(t,
		"https://wiki.example.com/auth",
		auth.Reference,
		"Reference field substituted (TMPL-10)",
	)
	assert.NotContains(t, auth.Reference, "${", "no residual token in Reference")
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

// --- Phase 35 Plan 02: recursive template-body expansion (D-17) ---
//
// [[template.<name>.use]] entries expand when the OUTER template is
// instantiated: params flow outer-to-inner, cycles and depth overruns are hard
// errors naming the chain, HS-1 deep-copy holds at every level.

// TestExpandTemplateBodyUseParamFlow exercises D-17 outer-to-inner param flow:
// template outer(p) whose body uses inner(q = ${p}) — instantiating outer with
// p="payload" expands inner with q substituted from the outer param, the
// produced inner unit attaching inside the outer clone's subtree.
func TestExpandTemplateBodyUseParamFlow(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Body Use Param Flow Test"

[messageBus]
type = "queue"
name = "Message Bus"

[template.inner]
params = ["name", "q"]
name = "${name} Inner"
type = "container"
description = "q=${q}"

[[template.inner.link]]
peer = "messageBus"
description = "from ${name} with ${q}"

[template.outer]
params = ["name", "p"]
name = "${name} Outer"
type = "system"

[[template.outer.use]]
template = "inner"
name = "${name}-inner"
q = "${p}"

[[use]]
template = "outer"
name = "app"
p = "payload"
`)

	m, err := parser.Parse(data)
	require.NoError(t, err, "Parse should not error")

	expanded, err := template.Expand(m)
	require.NoError(t, err, "Expand with template-body use should not error")

	// The outer instantiation produced 'app' at top level.
	app, ok := expanded.Units["app"]
	require.True(t, ok, "app present at top level")
	assert.Equal(t, "app Outer", app.Name, "app.Name (outer substitution)")

	// The body use produced app-inner INSIDE app's subtree (empty body-use
	// parent = direct subunit of the outer clone root).
	inner, ok := app.Subunits["app-inner"]
	require.True(t, ok, "app.Subunits contains 'app-inner' (body use attached in the clone)")
	assert.Contains(t, app.SubunitOrder, "app-inner", "app.SubunitOrder contains 'app-inner'")

	// Outer-to-inner param flow: q carried ${p}, substituted with the OUTER
	// instantiation's p="payload" BEFORE the inner expansion ran.
	assert.Equal(t, "app-inner Inner", inner.Name, "inner.Name (nested substitution)")
	assert.Equal(t, "q=payload", inner.Description, "inner.Description (outer param flowed inner)")

	require.Len(t, inner.Links, 1, "inner has its template's link")
	assert.Equal(t, "messageBus", inner.Links[0].Peer, "inner.Links[0].Peer")
	assert.Equal(t, "from app-inner with payload", inner.Links[0].Description,
		"inner.Links[0].Description (both name and q substituted)")

	// No residual tokens anywhere in the nested subtree.
	assertNotContainsSubst(t, app)
}

// TestExpandTemplateBodyUseThreeLevels exercises a 3-level chain (outer -> mid
// -> inner): every level expands fully, produced units land at their Parent
// paths INSIDE the outer clone's subtree — including a body use whose parent
// targets a subunit declared in the intermediate template's body.
func TestExpandTemplateBodyUseThreeLevels(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Three-Level Body Use Test"

[template.leaf]
params = ["name"]
name = "${name} Leaf"
type = "queue"

[template.mid]
params = ["name"]
name = "${name} Mid"
type = "box"

[template.mid.storage]
type = "db"
name = "Storage"

[[template.mid.use]]
template = "leaf"
name = "cache"
parent = "storage"

[template.outer]
params = ["name"]
name = "${name} Outer"
type = "system"

[[template.outer.use]]
template = "mid"
name = "svc"

[[use]]
template = "outer"
name = "app"
`)

	m, err := parser.Parse(data)
	require.NoError(t, err, "Parse should not error")

	expanded, err := template.Expand(m)
	require.NoError(t, err, "3-level chain should expand without error")

	app := expanded.Units["app"]
	require.NotNil(t, app, "app present")

	// Level 1: outer clone at top level; level 2: mid clone inside it.
	svc, ok := app.Subunits["svc"]
	require.True(t, ok, "app.svc (mid clone) present")
	assert.Equal(t, "svc Mid", svc.Name, "svc.Name (substituted)")
	assert.Contains(t, svc.SubunitOrder, "storage", "svc.SubunitOrder keeps the declared subunit")

	// mid's declared subunit survived the expansion...
	storage, ok := svc.Subunits["storage"]
	require.True(t, ok, "app.svc.storage (declared in mid) present")
	assert.Equal(t, "Storage", storage.Name, "storage.Name")

	// ...and mid's body use attached the leaf UNDER that declared subunit
	// (parent relative to the mid clone root).
	cache, ok := storage.Subunits["cache"]
	require.True(t, ok, "app.svc.storage.cache (leaf via 3-level recursion) present")
	assert.Equal(t, "cache Leaf", cache.Name, "cache.Name (substituted)")
	assert.Contains(t, storage.SubunitOrder, "cache", "storage.SubunitOrder contains 'cache'")

	// No residual tokens anywhere in the nested subtree.
	assertNotContainsSubst(t, app)
}

// TestExpandTemplateCycle exercises D-17 cycle detection: template a whose
// body uses b, and b's body uses a — Expand returns a hard error whose message
// names the cycle chain "a -> b -> a".
func TestExpandTemplateCycle(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Template Cycle Test"

[template.a]
params = ["name"]
name = "${name} A"
type = "box"

[[template.a.use]]
template = "b"
name = "x"

[template.b]
params = ["name"]
name = "${name} B"
type = "box"

[[template.b.use]]
template = "a"
name = "y"

[[use]]
template = "a"
name = "start"
`)

	m, err := parser.Parse(data)
	require.NoError(t, err, "Parse should not error")

	_, err = template.Expand(m)
	require.Error(t, err, "mutual template cycle must be a hard error")

	msg := err.Error()
	assert.Contains(t, msg, "cycle", "error names the cycle kind")
	assert.Contains(t, msg, "a -> b -> a", "error names the full cycle chain")
}

// TestExpandTemplateSelfCycle exercises the direct self-loop: a template whose
// body uses itself — the cycle error names the self-loop chain.
func TestExpandTemplateSelfCycle(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Template Self Cycle Test"

[template.loop]
params = ["name"]
name = "${name} Loop"
type = "box"

[[template.loop.use]]
template = "loop"
name = "again"

[[use]]
template = "loop"
name = "start"
`)

	m, err := parser.Parse(data)
	require.NoError(t, err, "Parse should not error")

	_, err = template.Expand(m)
	require.Error(t, err, "self-referencing template must be a hard error")

	msg := err.Error()
	assert.Contains(t, msg, "cycle", "error names the cycle kind")
	assert.Contains(t, msg, "loop -> loop", "error names the self-loop chain")
}

// TestExpandTemplateDepthCap exercises the depth cap (mirrors
// include.maxIncludeDepth = 100): an ACYCLIC chain of 150 distinct templates
// each using the next hits the cap as a hard error — never a stack-overflow
// panic, never a hang.
func TestExpandTemplateDepthCap(t *testing.T) {
	t.Parallel()

	const chainLen = 150

	var sb strings.Builder
	sb.WriteString("[properties]\nname = \"Depth Cap Test\"\n\n")

	for i := range chainLen {
		fmt.Fprintf(&sb, "[template.t%d]\nparams = [\"name\"]\nname = \"t%d\"\ntype = \"box\"\n\n", i, i)

		if i+1 < chainLen {
			fmt.Fprintf(&sb, "[[template.t%d.use]]\ntemplate = \"t%d\"\nname = \"n%d\"\n\n", i, i+1, i)
		}
	}

	sb.WriteString("[[use]]\ntemplate = \"t0\"\nname = \"root\"\n")

	m, err := parser.Parse([]byte(sb.String()))
	require.NoError(t, err, "generated chain should parse")

	_, err = template.Expand(m)
	require.Error(t, err, "150-deep acyclic chain must hit the depth cap")

	assert.Contains(t, err.Error(), "depth", "error names the depth cap")
}

// TestExpandNestedResidualToken exercises TMPL-06 post-recursion: a ${token}
// in a nested body that no param fills is caught by the residual scan AFTER
// the recursion completes (exit gate).
func TestExpandNestedResidualToken(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Nested Residual Token Test"

[template.leaf]
params = ["name"]
name = "${name} Leaf"
description = "needs ${unfilled}"
type = "db"

[template.outer]
params = ["name"]
name = "${name} Outer"
type = "box"

[[template.outer.use]]
template = "leaf"
name = "cache"

[[use]]
template = "outer"
name = "app"
`)

	m, err := parser.Parse(data)
	require.NoError(t, err, "Parse should not error")

	_, err = template.Expand(m)
	require.Error(t, err, "residual ${unfilled} in a nested body must be a hard error")

	msg := err.Error()
	assert.Contains(t, msg, "unresolved parameter", "error kind is the residual-token scan")
	assert.Contains(t, msg, "${unfilled", "error names the unfilled token")
}

// TestExpandNestedPathCollision exercises TMPL-07 with nested uses: a
// template-body use producing a path that collides with an authored unit is a
// hard duplicate-path error — never a silent overwrite. "Authored" covers both
// a subunit declared in the template body and a sibling body use's product.
func TestExpandNestedPathCollision(t *testing.T) {
	t.Parallel()

	t.Run("body use vs template-declared subunit", func(t *testing.T) {
		t.Parallel()

		data := []byte(`
[properties]
name = "Nested Collision Test"

[template.leaf]
params = ["name"]
name = "${name} Leaf"
type = "db"

[template.outer]
params = ["name"]
name = "${name} Outer"
type = "box"

[template.outer.api]
type = "db"
name = "Declared API"

[[template.outer.use]]
template = "leaf"
name = "api"

[[use]]
template = "outer"
name = "app"
`)

		m, err := parser.Parse(data)
		require.NoError(t, err, "Parse should not error")

		_, err = template.Expand(m)
		require.Error(t, err, "body use colliding with a template-declared subunit must be a hard error")

		msg := err.Error()
		assert.Contains(t, msg, "duplicate unit path", "error kind is the path collision")
		assert.Contains(t, msg, "app.api", "error names the colliding path")
	})

	t.Run("two body uses colliding", func(t *testing.T) {
		t.Parallel()

		data := []byte(`
[properties]
name = "Nested Collision Test 2"

[template.leaf]
params = ["name"]
name = "${name} Leaf"
type = "db"

[template.outer]
params = ["name"]
name = "${name} Outer"
type = "box"

[[template.outer.use]]
template = "leaf"
name = "dup"

[[template.outer.use]]
template = "leaf"
name = "dup"

[[use]]
template = "outer"
name = "app2"
`)

		m, err := parser.Parse(data)
		require.NoError(t, err, "Parse should not error")

		_, err = template.Expand(m)
		require.Error(t, err, "two body uses producing the same path must be a hard error")

		assert.Contains(t, err.Error(), "app2.dup", "error names the colliding path")
	})
}

// TestExpandThreeLevelNestingHS1 extends TestExpandThreeInstantiationsHS1 to
// 3-level nesting (D-17): the outer template instantiates 3x; each expansion
// recursively produces mid and leaf levels whose links carry per-instantiation
// params. After validator.Validate (which mutates LinksFrom in place), the
// messageBus mirrors must be DISJOINT — one per leaf, no shared Mirror state
// at any recursion level (HS-1).
func TestExpandThreeLevelNestingHS1(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "3-Level Nesting HS-1 Test"

[messageBus]
type = "queue"
name = "Message Bus"

[template.leaf]
params = ["name"]
name = "${name} leaf"
type = "containerDb"

[[template.leaf.link]]
peer = "messageBus"
description = "leaf ${name} events"

[template.mid]
params = ["name"]
name = "${name} mid"
type = "containerBox"

[[template.mid.use]]
template = "leaf"
name = "${name}-leaf"

[template.outer]
params = ["name"]
name = "${name} outer"
type = "system"

[[template.outer.use]]
template = "mid"
name = "${name}-mid"

[[use]]
template = "outer"
name = "alpha"

[[use]]
template = "outer"
name = "beta"

[[use]]
template = "outer"
name = "gamma"
`)

	m, err := parser.Parse(data)
	require.NoError(t, err, "Parse should not error")

	expanded, err := template.Expand(m)
	require.NoError(t, err, "3-level nesting x3 should expand without error")

	// Each instantiation produced the full 3-level subtree with distinct
	// substituted values at every level.
	for _, instName := range []string{"alpha", "beta", "gamma"} {
		root := expanded.Units[instName]
		require.NotNil(t, root, "%s present", instName)
		assert.Equal(t, instName+" outer", root.Name, "%s.Name", instName)

		mid := root.Subunits[instName+"-mid"]
		require.NotNil(t, mid, "%s.%s-mid present", instName, instName)
		assert.Equal(t, instName+"-mid mid", mid.Name, "%s-mid.Name", instName)

		leaf := mid.Subunits[instName+"-mid-leaf"]
		require.NotNil(t, leaf, "%s leaf present", instName)
		assert.Equal(t, instName+"-mid-leaf leaf", leaf.Name, "%s leaf.Name", instName)

		require.Len(t, leaf.Links, 1, "%s leaf has its link", instName)
		assert.Equal(t, "leaf "+instName+"-mid-leaf events", leaf.Links[0].Description,
			"%s leaf link description (substituted at every level)", instName)
	}

	// The validator mutates LinksFrom in place — the 3 leaves must yield three
	// DISTINCT mirror entries on messageBus (HS-1 at every recursion level).
	valErrors := validator.Validate(expanded)
	assert.Empty(t, valErrors, "expanded model should validate cleanly")

	bus := expanded.Units["messageBus"]
	require.NotNil(t, bus, "messageBus present")

	var mirrorSources []string

	for i := range bus.LinksFrom {
		if bus.LinksFrom[i].IsMirror() {
			mirrorSources = append(mirrorSources, bus.LinksFrom[i].Peer)
		}
	}

	assert.ElementsMatch(t,
		[]string{
			"alpha.alpha-mid.alpha-mid-leaf",
			"beta.beta-mid.beta-mid-leaf",
			"gamma.gamma-mid.gamma-mid-leaf",
		},
		mirrorSources,
		"messageBus.LinksFrom must carry one DISJOINT mirror per 3-level "+
			"instantiation (HS-1; nested peers mirror by full dotted path)",
	)

	// Idempotency: re-Parse + re-Expand produces a deeply equal model.
	m2, err := parser.Parse(data)
	require.NoError(t, err, "second Parse should not error")

	expanded2, err := template.Expand(m2)
	require.NoError(t, err, "second Expand should not error")

	assert.ElementsMatch(t, expanded.UnitOrder, expanded2.UnitOrder, "idempotent UnitOrder")
	assertExpandedEqual(t, expanded, expanded2)
}
