package parser_test

import (
	"os"
	"testing"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const webappKey = "webapp"

func TestParseValidProperties(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/valid.toml")
	require.NoError(t, err, "failed to read test fixture")

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	assert.Equal(t, "Test System", got.Properties.Name, "Properties.Name")
	assert.Equal(t, "A test architecture", got.Properties.Description, "Properties.Description")
	assert.Equal(t, 40, got.Properties.LineLength, "Properties.LineLength")
	assert.Equal(t, []string{webappKey}, got.Properties.Expanded, "Properties.Expanded")
}

func TestParseValidUserUnit(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/valid.toml")
	require.NoError(t, err, "failed to read test fixture")

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")
	require.Len(t, got.Units, 2, "should have 2 units")

	user, ok := got.Units["user"]
	require.True(t, ok, "missing 'user' unit")

	assert.Equal(t, model.TypePerson, user.Type, "user.Type")
	assert.Equal(t, "User", user.Name, "user.Name")
	assert.Equal(t, "End user of the system", user.Description, "user.Description")
}

func TestParseValidWebappUnit(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/valid.toml")
	require.NoError(t, err, "failed to read test fixture")

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	webapp, ok := got.Units[webappKey]
	require.True(t, ok, "missing 'webapp' unit")

	assert.Equal(t, model.TypeSystem, webapp.Type, "webapp.Type")
	assert.Equal(t, "Web Application", webapp.Name, "webapp.Name")
	assert.Equal(t, "Go, React", webapp.Technology, "webapp.Technology")
}

func TestParseNestedProperties(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/nested.toml")
	require.NoError(t, err, "failed to read test fixture")

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	assert.Equal(t, "Nested Test", got.Properties.Name, "Properties.Name")
}

func TestParseNestedExternalUnit(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/nested.toml")
	require.NoError(t, err, "failed to read test fixture")

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")
	require.Len(t, got.Units, 2, "should have 2 units")

	externals, ok := got.Units["externals"]
	require.True(t, ok, "missing 'externals' unit")

	assert.Equal(t, model.TypeSystemExternal, externals.Type, "externals.Type")
}

func TestParseNestedMainappUnit(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/nested.toml")
	require.NoError(t, err, "failed to read test fixture")

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	mainapp, ok := got.Units["mainapp"]
	require.True(t, ok, "missing 'mainapp' unit")

	assert.Equal(t, model.TypeSystem, mainapp.Type, "mainapp.Type")
	assert.Equal(t, "Main Application", mainapp.Name, "mainapp.Name")
	assert.Equal(t, "Go", mainapp.Technology, "mainapp.Technology")
}

func TestParseNestedContainers(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/nested.toml")
	require.NoError(t, err, "failed to read test fixture")

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	mainapp, ok := got.Units["mainapp"]
	require.True(t, ok, "missing 'mainapp' unit")
	require.Len(t, mainapp.Subunits, 2, "mainapp should have 2 subunits")

	api, ok := mainapp.Subunits["api"]
	require.True(t, ok, "missing 'mainapp.api' subunit")

	assert.Equal(t, model.TypeContainer, api.Type, "api.Type")
	assert.Equal(t, "API Server", api.Name, "api.Name")

	db, ok := mainapp.Subunits["db"]
	require.True(t, ok, "missing 'mainapp.db' subunit")

	assert.Equal(t, model.TypeContainerDb, db.Type, "db.Type")
	assert.Equal(t, "Database", db.Name, "db.Name")
}

func TestParseNestedComponents(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/nested.toml")
	require.NoError(t, err, "failed to read test fixture")

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	mainapp := got.Units["mainapp"]
	require.NotNil(t, mainapp, "missing 'mainapp' unit")

	api := mainapp.Subunits["api"]
	require.NotNil(t, api, "missing 'mainapp.api' subunit")
	require.Len(t, api.Subunits, 1, "api should have 1 subunit")

	handler, ok := api.Subunits["handler"]
	require.True(t, ok, "missing 'mainapp.api.handler' subunit")

	assert.Equal(t, model.TypeComponent, handler.Type, "handler.Type")
	assert.Equal(t, "Request Handler", handler.Name, "handler.Name")
}

func TestParseLinksProperties(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/links.toml")
	require.NoError(t, err, "failed to read test fixture")

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	assert.Equal(t, "Links Test", got.Properties.Name, "Properties.Name")
	assert.Equal(t, "spline", got.Properties.Edges, "Properties.Edges")
}

func TestParseLinksOutgoing(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/links.toml")
	require.NoError(t, err, "failed to read test fixture")

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	webapp, ok := got.Units[webappKey]
	require.True(t, ok, "missing 'webapp' unit")
	require.Len(t, webapp.Links, 1, "webapp should have 1 link")

	link, ok := model.FindLinkByPeer(webapp.Links, "user")
	require.True(t, ok, "missing 'user' link in webapp")

	assert.Equal(t, "user", link.Peer, "link.Peer")
	assert.Equal(t, model.ArrowForward, link.Arrow, "link.Arrow")
	assert.Equal(t, model.RankForward, link.Rank, "link.Rank")
	assert.Equal(t, "HTTPS", link.Technology, "link.Technology")
	assert.Equal(t, "Uses", link.Description, "link.Description")
}

func TestParseLinksIncoming(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/links.toml")
	require.NoError(t, err, "failed to read test fixture")

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	api, ok := got.Units["api"]
	require.True(t, ok, "missing 'api' unit")
	require.Len(t, api.LinksFrom, 1, "api should have 1 linkFrom")

	linkFrom, ok := model.FindLinkByPeer(api.LinksFrom, webappKey)
	require.True(t, ok, "missing 'webapp' linkFrom in api")

	assert.Equal(t, webappKey, linkFrom.Peer, "linkFrom.Peer")
	assert.Equal(t, model.ArrowForward, linkFrom.Arrow, "linkFrom.Arrow")
	assert.Equal(t, "HTTP/JSON", linkFrom.Technology, "linkFrom.Technology")
}

func TestParseInvalidTOML(t *testing.T) {
	t.Parallel()

	invalidData := []byte("invalid [[[")
	_, err := parser.Parse(invalidData)
	require.Error(t, err, "Parse() should error for invalid TOML")
	assert.Contains(t, err.Error(), "parse error", "error message should contain 'parse error'")

	var parseErr *parser.ParseError
	require.ErrorAs(t, err, &parseErr, "error should be *ParseError")
	assert.NotZero(t, parseErr.Line, "ParseError.Line should be non-zero for invalid TOML")
}

func TestParseMissingFile(t *testing.T) {
	t.Parallel()

	_, err := parser.ParseFile("nonexistent.toml")
	require.Error(t, err, "ParseFile() should error for missing file")

	var parseErr *parser.ParseError
	require.ErrorAs(t, err, &parseErr, "error should be *ParseError")
	assert.Equal(t, "failed to read file", parseErr.Message, "ParseError.Message")
	assert.Equal(t, "nonexistent.toml", parseErr.Context, "ParseError.Context")
}

func TestParsePropertiesFields(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/valid.toml")
	require.NoError(t, err, "failed to read test fixture")

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	assert.Equal(t, "transparent", got.Properties.Color, "Properties.Color")
	assert.Equal(t, "straight", got.Properties.Edges, "Properties.Edges")
	assert.Equal(t, 40, got.Properties.LineLength, "Properties.LineLength")
	assert.Len(t, got.Properties.Expanded, 1, "len(Properties.Expanded)")
	assert.Equal(t, webappKey, got.Properties.Expanded[0], "Properties.Expanded[0]")
}

func TestParseUnitFields(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/valid.toml")
	require.NoError(t, err, "failed to read test fixture")

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	webapp, ok := got.Units[webappKey]
	require.True(t, ok, "missing 'webapp' unit")

	assert.Equal(t, model.TypeSystem, webapp.Type, "webapp.Type")
	assert.Equal(t, "Web Application", webapp.Name, "webapp.Name")
	assert.Equal(t, "Main web application", webapp.Description, "webapp.Description")
	assert.Equal(t, "Go, React", webapp.Technology, "webapp.Technology")
}

func TestParseEmptyFile(t *testing.T) {
	t.Parallel()

	got, err := parser.Parse([]byte(""))
	require.NoError(t, err, "Parse() should not error for empty file")

	assert.Empty(t, got.Properties.Name, "Properties.Name should be empty")
	assert.Empty(t, got.Units, "Units should be empty")
}

func TestParseOnlyProperties(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Only Properties Test"
description = "Test with only properties"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	assert.Equal(t, "Only Properties Test", got.Properties.Name, "Properties.Name")
	assert.Empty(t, got.Units, "Units should be empty")
}

func TestParseExternalTypes(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "External Types Test"

[personext]
type = "personExternal"
name = "External User"

[systemext]
type = "systemExternal"
name = "External System"

[dbext]
type = "dbExternal"
name = "External Database"

[queueext]
type = "queueExternal"
name = "External Queue"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	assert.Equal(t, model.TypePersonExternal, got.Units["personext"].Type, "personext.Type")
	assert.Equal(t, model.TypeSystemExternal, got.Units["systemext"].Type, "systemext.Type")
	assert.Equal(t, model.TypeDbExternal, got.Units["dbext"].Type, "dbext.Type")
	assert.Equal(t, model.TypeQueueExternal, got.Units["queueext"].Type, "queueext.Type")
}

func TestParseLinkFrom(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "LinkFrom Test"

[a]
type = "system"
name = "System A"

[[a.linkFrom]]
peer = "b"
arrow = "reverse"
technology = "TCP"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	unitA, ok := got.Units["a"]
	require.True(t, ok, "missing 'a' unit")
	require.Len(t, unitA.LinksFrom, 1, "unitA should have 1 linkFrom")

	linkFrom, ok := model.FindLinkByPeer(unitA.LinksFrom, "b")
	require.True(t, ok, "missing 'b' linkFrom in a")

	assert.Equal(t, "b", linkFrom.Peer, "linkFrom.Peer")
	assert.Equal(t, model.ArrowReverse, linkFrom.Arrow, "linkFrom.Arrow")
	assert.Equal(t, "TCP", linkFrom.Technology, "linkFrom.Technology")
}

func TestParseAllUnitTypesC1(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "C1 Types Test"

[p1]
type = "person"

[s1]
type = "system"

[d1]
type = "db"

[q1]
type = "queue"

[b1]
type = "box"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	expectedTypes := map[string]model.UnitType{
		"p1": model.TypePerson,
		"s1": model.TypeSystem,
		"d1": model.TypeDb,
		"q1": model.TypeQueue,
		"b1": model.TypeBox,
	}

	for name, expectedType := range expectedTypes {
		unit, ok := got.Units[name]
		require.True(t, ok, "missing unit %q", name)
		assert.Equal(t, expectedType, unit.Type, "%s.Type", name)
	}
}

func TestParseAllUnitTypesC2(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "C2 Types Test"

[c1]
type = "container"

[cd1]
type = "containerDb"

[cq1]
type = "containerQueue"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	expectedTypes := map[string]model.UnitType{
		"c1":  model.TypeContainer,
		"cd1": model.TypeContainerDb,
		"cq1": model.TypeContainerQueue,
	}

	for name, expectedType := range expectedTypes {
		unit, ok := got.Units[name]
		require.True(t, ok, "missing unit %q", name)
		assert.Equal(t, expectedType, unit.Type, "%s.Type", name)
	}
}

func TestParseAllUnitTypesC3(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "C3 Types Test"

[cmp1]
type = "component"

[cmpd1]
type = "componentDb"

[cmpq1]
type = "componentQueue"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	expectedTypes := map[string]model.UnitType{
		"cmp1":  model.TypeComponent,
		"cmpd1": model.TypeComponentDb,
		"cmpq1": model.TypeComponentQueue,
	}

	for name, expectedType := range expectedTypes {
		unit, ok := got.Units[name]
		require.True(t, ok, "missing unit %q", name)
		assert.Equal(t, expectedType, unit.Type, "%s.Type", name)
	}
}

func TestParseLinkAllFields(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Link Fields Test"

[a]
type = "system"
name = "A"

[[a.link]]
peer = "b"
arrow = "bidirectional"
rank = "equal"
color = "red"
style = "dashed"
technology = "gRPC"
description = "syncs data"
labelPosition = "head"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	unitA, ok := got.Units["a"]
	require.True(t, ok, "missing 'a' unit")

	link, ok := model.FindLinkByPeer(unitA.Links, "b")
	require.True(t, ok, "missing 'b' link in a")

	assert.Equal(t, "b", link.Peer, "link.Peer")
	assert.Equal(t, model.ArrowBidirectional, link.Arrow, "link.Arrow")
	assert.Equal(t, model.RankEqual, link.Rank, "link.Rank")
	assert.Equal(t, "red", link.Color, "link.Color")
	assert.Equal(t, "dashed", link.Style, "link.Style")
}

func TestParseLinkAllFieldsMetadata(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Link Fields Test"

[a]
type = "system"
name = "A"

[[a.link]]
peer = "b"
technology = "gRPC"
description = "syncs data"
labelPosition = "head"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	link, ok := model.FindLinkByPeer(got.Units["a"].Links, "b")
	require.True(t, ok, "missing 'b' link")

	assert.Equal(t, "gRPC", link.Technology, "link.Technology")
	assert.Equal(t, "syncs data", link.Description, "link.Description")
	assert.Equal(t, model.LabelHead, link.LabelPosition, "link.LabelPosition")
}

func TestParseNestedLink(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Nested Link Test"

[parent]
type = "system"
name = "Parent"

[parent.child]
type = "container"
name = "Child"

[[parent.child.link]]
peer = "other"
technology = "HTTP"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	parent, ok := got.Units["parent"]
	require.True(t, ok, "missing 'parent' unit")

	child, ok := parent.Subunits["child"]
	require.True(t, ok, "missing 'parent.child' subunit")
	require.Len(t, child.Links, 1, "child should have 1 link")

	link, ok := model.FindLinkByPeer(child.Links, "other")
	require.True(t, ok, "missing 'other' link in child")

	assert.Equal(t, "other", link.Peer, "link.Peer")
}

func TestParseFile(t *testing.T) {
	t.Parallel()

	got, err := parser.ParseFile("../../testdata/valid.toml")
	require.NoError(t, err, "ParseFile() should not error")

	assert.Equal(t, "Test System", got.Properties.Name, "Properties.Name")
}
