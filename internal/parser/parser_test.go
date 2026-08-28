package parser_test

import (
	"fmt"
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
kind = "read"
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
	assert.Equal(t, model.KindRead, link.Kind, "link.Kind")
	assert.Equal(t, "red", link.Color, "link.Color")
	assert.Equal(t, "dashed", link.Style, "link.Style")
}

func TestParseLinkKind(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kind string
		want model.LinkKind
	}{
		{"read", "read", model.KindRead},
		{"write", "write", model.KindWrite},
		{"read-write", "read-write", model.KindReadWrite},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data := fmt.Sprintf(`
[properties]
name = "Kind Test"

[a]
type = "system"

[[a.link]]
peer = "b"
kind = "%s"
`, tt.kind)

			got, err := parser.Parse([]byte(data))
			require.NoError(t, err, "Parse() should not error")

			unitA, ok := got.Units["a"]
			require.True(t, ok, "missing 'a' unit")

			link, ok := model.FindLinkByPeer(unitA.Links, "b")
			require.True(t, ok, "missing 'b' link in a")
			assert.Equal(t, tt.want, link.Kind, "link.Kind")
		})
	}
}

func TestParseLinkKindAbsent(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Kind Absent Test"

[a]
type = "system"

[[a.link]]
peer = "b"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	unitA, ok := got.Units["a"]
	require.True(t, ok, "missing 'a' unit")

	link, ok := model.FindLinkByPeer(unitA.Links, "b")
	require.True(t, ok, "missing 'b' link in a")
	assert.Equal(t, model.LinkKind(""), link.Kind, "absent kind should be the zero value")
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

// TestParseGenericTypeInference tests that generic types (db, queue) are
// automatically transformed to level-specific types based on nesting depth.
func TestParseGenericTypeInference_DbAtC1(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Generic Type Test"

[mydb]
type = "db"
name = "Database"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	mydb, ok := got.Units["mydb"]
	require.True(t, ok, "missing 'mydb' unit")
	assert.Equal(t, model.TypeDb, mydb.Type, "db at C1 should stay as db")
}

func TestParseGenericTypeInference_DbAtC2(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Generic Type Test"

[system]
type = "system"
name = "System"

[system.mydb]
type = "db"
name = "Database"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	system, ok := got.Units["system"]
	require.True(t, ok, "missing 'system' unit")

	mydb, ok := system.Subunits["mydb"]
	require.True(t, ok, "missing 'system.mydb' subunit")
	assert.Equal(t, model.TypeContainerDb, mydb.Type, "db at C2 should become containerDb")
}

func TestParseGenericTypeInference_DbAtC3(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Generic Type Test"

[system]
type = "system"
name = "System"

[system.api]
type = "container"
name = "API"

[system.api.mydb]
type = "db"
name = "Database"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	system := got.Units["system"]
	require.NotNil(t, system, "missing 'system' unit")

	api := system.Subunits["api"]
	require.NotNil(t, api, "missing 'system.api' subunit")

	mydb, ok := api.Subunits["mydb"]
	require.True(t, ok, "missing 'system.api.mydb' subunit")
	assert.Equal(t, model.TypeComponentDb, mydb.Type, "db at C3 should become componentDb")
}

func TestParseGenericTypeInference_QueueAtC1(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Generic Type Test"

[myqueue]
type = "queue"
name = "Queue"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	myqueue, ok := got.Units["myqueue"]
	require.True(t, ok, "missing 'myqueue' unit")
	assert.Equal(t, model.TypeQueue, myqueue.Type, "queue at C1 should stay as queue")
}

func TestParseGenericTypeInference_QueueAtC2(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Generic Type Test"

[system]
type = "system"
name = "System"

[system.myqueue]
type = "queue"
name = "Queue"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	system, ok := got.Units["system"]
	require.True(t, ok, "missing 'system' unit")

	myqueue, ok := system.Subunits["myqueue"]
	require.True(t, ok, "missing 'system.myqueue' subunit")
	assert.Equal(t, model.TypeContainerQueue, myqueue.Type, "queue at C2 should become containerQueue")
}

func TestParseGenericTypeInference_QueueAtC3(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Generic Type Test"

[system]
type = "system"
name = "System"

[system.api]
type = "container"
name = "API"

[system.api.myqueue]
type = "queue"
name = "Queue"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	system := got.Units["system"]
	require.NotNil(t, system, "missing 'system' unit")

	api := system.Subunits["api"]
	require.NotNil(t, api, "missing 'system.api' subunit")

	myqueue, ok := api.Subunits["myqueue"]
	require.True(t, ok, "missing 'system.api.myqueue' subunit")
	assert.Equal(t, model.TypeComponentQueue, myqueue.Type, "queue at C3 should become componentQueue")
}

func TestParseGenericTypeInference_InBox(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Generic Type Test"

[mybox]
type = "box"
name = "Box"

[mybox.mydb]
type = "db"
name = "Database"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	mybox, ok := got.Units["mybox"]
	require.True(t, ok, "missing 'mybox' unit")

	mydb, ok := mybox.Subunits["mydb"]
	require.True(t, ok, "missing 'mybox.mydb' subunit")
	// C1 box can only contain C1 types, so db stays as db (not converted to containerDb)
	assert.Equal(t, model.TypeDb, mydb.Type, "db inside C1 box should stay as db (same-level grouping)")
}

func TestParseGenericTypeInference_BoxAtC1(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Generic Type Test"

[mybox]
type = "box"
name = "Box"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	mybox, ok := got.Units["mybox"]
	require.True(t, ok, "missing 'mybox' unit")
	assert.Equal(t, model.TypeBox, mybox.Type, "box at C1 (root) should stay as box")
}

func TestParseGenericTypeInference_BoxAtC2(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Generic Type Test"

[system]
type = "system"
name = "System"

[system.mybox]
type = "box"
name = "Group"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	system, ok := got.Units["system"]
	require.True(t, ok, "missing 'system' unit")

	mybox, ok := system.Subunits["mybox"]
	require.True(t, ok, "missing 'system.mybox' subunit")
	assert.Equal(t, model.TypeContainerBox, mybox.Type, "box inside system should promote to containerBox")
}

func TestParseGenericTypeInference_BoxAtC3(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Generic Type Test"

[system]
type = "system"
name = "System"

[system.api]
type = "container"
name = "API"

[system.api.mybox]
type = "box"
name = "Group"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	system := got.Units["system"]
	require.NotNil(t, system, "missing 'system' unit")

	api := system.Subunits["api"]
	require.NotNil(t, api, "missing 'system.api' subunit")

	mybox, ok := api.Subunits["mybox"]
	require.True(t, ok, "missing 'system.api.mybox' subunit")
	assert.Equal(t, model.TypeComponentBox, mybox.Type, "box inside container should promote to componentBox")
}

func TestParseGenericTypeInference_BoxInBoxStaysBox(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Generic Type Test"

[outer]
type = "box"
name = "Outer"

[outer.inner]
type = "box"
name = "Inner"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	outer, ok := got.Units["outer"]
	require.True(t, ok, "missing 'outer' unit")

	inner, ok := outer.Subunits["inner"]
	require.True(t, ok, "missing 'outer.inner' subunit")
	// A box inside a C1 box is same-level C1 grouping, so it stays box.
	assert.Equal(t, model.TypeBox, inner.Type, "box inside C1 box should stay as box")
}

func TestParseGenericTypeInference_PromotedBoxChildrenDefaultToContainer(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Generic Type Test"

[system]
type = "system"
name = "System"

[system.group]
type = "box"
name = "Group"

[system.group.svc]
name = "Service"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	system := got.Units["system"]
	require.NotNil(t, system, "missing 'system' unit")

	group := system.Subunits["group"]
	require.NotNil(t, group, "missing 'system.group' subunit")
	assert.Equal(t, model.TypeContainerBox, group.Type, "box inside system should promote to containerBox")

	svc, ok := group.Subunits["svc"]
	require.True(t, ok, "missing 'system.group.svc' subunit")
	// Once promoted to containerBox, children default to container (C2), not system.
	assert.Equal(t, model.TypeContainer, svc.Type,
		"child of promoted box should default to container, not system")
}

func TestParseGenericTypeInference_ExplicitBoxVariantsUnchanged(t *testing.T) {
	t.Parallel()

	// Explicit containerBox/componentBox are not TypeBox, so inference must not touch them.
	data := []byte(`
[properties]
name = "Explicit Box Test"

[system]
type = "system"

[system.cbox]
type = "containerBox"

[system.api]
type = "container"

[system.api.mpbox]
type = "componentBox"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	system := got.Units["system"]
	require.NotNil(t, system, "missing 'system' unit")

	cbox := system.Subunits["cbox"]
	require.NotNil(t, cbox, "missing 'system.cbox' subunit")
	assert.Equal(t, model.TypeContainerBox, cbox.Type, "explicit containerBox should stay as containerBox")

	api := system.Subunits["api"]
	require.NotNil(t, api, "missing 'system.api' subunit")

	mpbox := api.Subunits["mpbox"]
	require.NotNil(t, mpbox, "missing 'system.api.mpbox' subunit")
	assert.Equal(t, model.TypeComponentBox, mpbox.Type, "explicit componentBox should stay as componentBox")
}

func TestParseGenericTypeInference_InSystemExternal(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Generic Type Test"

[external]
type = "systemExternal"
name = "External"

[external.myqueue]
type = "queue"
name = "Queue"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	external, ok := got.Units["external"]
	require.True(t, ok, "missing 'external' unit")

	myqueue, ok := external.Subunits["myqueue"]
	require.True(t, ok, "missing 'external.myqueue' subunit")
	// systemExternal cannot contain subunits (validator will reject), but parser keeps type as-is
	assert.Equal(t, model.TypeQueue, myqueue.Type,
		"queue inside systemExternal stays as queue (systemExternal cannot contain subunits)")
}

func TestParseGenericTypeInference_ExplicitTypesUnchanged(t *testing.T) {
	t.Parallel()

	// Explicit level-specific types should remain unchanged
	data := []byte(`
[properties]
name = "Explicit Type Test"

[system]
type = "system"

[system.containerdb]
type = "containerDb"

[system.containerqueue]
type = "containerQueue"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	system := got.Units["system"]
	require.NotNil(t, system, "missing 'system' unit")

	containerdb := system.Subunits["containerdb"]
	require.NotNil(t, containerdb, "missing 'system.containerdb' subunit")
	assert.Equal(t, model.TypeContainerDb, containerdb.Type, "explicit containerDb should stay as containerDb")

	containerqueue := system.Subunits["containerqueue"]
	require.NotNil(t, containerqueue, "missing 'system.containerqueue' subunit")
	assert.Equal(t, model.TypeContainerQueue, containerqueue.Type, "explicit containerQueue should stay as containerQueue")
}

func TestParseDefinitionOrder(t *testing.T) {
	t.Parallel()

	// Define units in non-alphabetical order: zulu, alpha, gamma
	data := []byte(`
[properties]
name = "Order Test"

[zulu]
type = "system"
name = "Zulu System"

[alpha]
type = "system"
name = "Alpha System"

[gamma]
type = "db"
name = "Gamma Database"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	// UnitOrder should preserve TOML definition order: zulu, alpha, gamma
	require.Len(t, got.UnitOrder, 3, "UnitOrder should have 3 entries")
	assert.Equal(t, "zulu", got.UnitOrder[0], "first unit should be zulu (definition order)")
	assert.Equal(t, "alpha", got.UnitOrder[1], "second unit should be alpha (definition order)")
	assert.Equal(t, "gamma", got.UnitOrder[2], "third unit should be gamma (definition order)")
}

func TestParseSubunitDefinitionOrder(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Subunit Order Test"

[system]
type = "system"
name = "System"

[system.zeta]
type = "container"
name = "Zeta Container"

[system.alpha]
type = "container"
name = "Alpha Container"

[system.gamma]
type = "container"
name = "Gamma Container"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	system, ok := got.Units["system"]
	require.True(t, ok, "missing 'system' unit")

	// SubunitOrder should preserve TOML definition order: zeta, alpha, gamma
	require.Len(t, system.SubunitOrder, 3, "SubunitOrder should have 3 entries")
	assert.Equal(t, "zeta", system.SubunitOrder[0], "first subunit should be zeta")
	assert.Equal(t, "alpha", system.SubunitOrder[1], "second subunit should be alpha")
	assert.Equal(t, "gamma", system.SubunitOrder[2], "third subunit should be gamma")
}

// TestReferenceField_PreservesURL exercises REF-01: a TOML unit authored with a
// `reference` key parses into Unit.Reference carrying the exact string (no scheme
// enforcement, preserved verbatim).
func TestReferenceField_PreservesURL(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Reference Test"

[webapp]
type = "system"
name = "Web Application"
reference = "https://example.com/runbook"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	webapp, ok := got.Units["webapp"]
	require.True(t, ok, "missing 'webapp' unit")
	assert.Equal(t, "https://example.com/runbook", webapp.Reference, "webapp.Reference should preserve the URL exactly")
}

// TestReferenceField_EmptyEqualsOmitted exercises REF-01 empty-equals-omitted:
// both an absent `reference` key and `reference = ""` parse into Unit.Reference == "".
func TestReferenceField_EmptyEqualsOmitted(t *testing.T) {
	t.Parallel()

	// Case 1: no `reference` key at all.
	dataNoKey := []byte(`
[properties]
name = "Reference Test"

[without]
type = "system"
name = "Without Reference"
`)
	gotNoKey, err := parser.Parse(dataNoKey)
	require.NoError(t, err, "Parse() should not error for the no-key case")

	without, ok := gotNoKey.Units["without"]
	require.True(t, ok, "missing 'without' unit")
	assert.Empty(t, without.Reference, "absent reference key should parse to empty string")

	// Case 2: explicit empty `reference = ""`.
	dataEmpty := []byte(`
[properties]
name = "Reference Test"

[empty]
type = "system"
name = "Empty Reference"
reference = ""
`)
	gotEmpty, err := parser.Parse(dataEmpty)
	require.NoError(t, err, "Parse() should not error for the empty case")

	empty, ok := gotEmpty.Units["empty"]
	require.True(t, ok, "missing 'empty' unit")
	assert.Empty(t, empty.Reference, "empty reference value should parse to empty string")
}

// TestReferenceField_NoPhantomSubunit exercises REF-05 / BC-1: a unit authored
// with `reference = "..."` must NOT create a phantom subunit named "reference"
// in Unit.Subunits (the bug that fires without the isBuiltinField entry).
func TestReferenceField_NoPhantomSubunit(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Reference Test"

[webapp]
type = "system"
name = "Web Application"
reference = "https://example.com/runbook"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	webapp, ok := got.Units["webapp"]
	require.True(t, ok, "missing 'webapp' unit")

	assert.Empty(t, webapp.Subunits, "reference must not create a phantom subunit (BC-1)")
	assert.NotContains(t, webapp.SubunitOrder, "reference", "SubunitOrder must not contain 'reference' (BC-1)")
}

// TestReferenceField_AnyType exercises REF-01/REF-05: the field is accepted on
// multiple distinct unit types in the same model, each carrying a distinct URL
// that round-trips exactly.
func TestReferenceField_AnyType(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Reference Any-Type Test"

[srv]
type = "system"
name = "System"
reference = "https://example.com/docs/system"

[store]
type = "db"
name = "Database"
reference = "https://example.com/docs/db"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	srv, ok := got.Units["srv"]
	require.True(t, ok, "missing 'srv' unit")
	assert.Equal(t, model.TypeSystem, srv.Type, "srv.Type")
	assert.Equal(t, "https://example.com/docs/system", srv.Reference, "srv.Reference")

	store, ok := got.Units["store"]
	require.True(t, ok, "missing 'store' unit")
	assert.Equal(t, model.TypeDb, store.Type, "store.Type")
	assert.Equal(t, "https://example.com/docs/db", store.Reference, "store.Reference")
}

// TestParseOmittedNameTopLevel exercises ERGO-03: a top-level unit defined
// with no `name` field derives its display name from the identifier segment
// via model.Humanize ("linuxSystem" → "Linux System").
func TestParseOmittedNameTopLevel(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Omitted Name Test"

[linuxSystem]
type = "system"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	unit, ok := got.Units["linuxSystem"]
	require.True(t, ok, "missing 'linuxSystem' unit")
	assert.Equal(t, "Linux System", unit.Name,
		"top-level unit with omitted name should humanize from the identifier segment")
}

// TestParseOmittedNameNestedSegment exercises ERGO-03 + D-02: a nested
// subunit defined with no `name` humanizes from its LAST path segment only
// (not the full dotted path). [linuxSystem.localIDP] → "Local IDP".
func TestParseOmittedNameNestedSegment(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Omitted Nested Name Test"

[linuxSystem]
type = "system"
name = "Linux System"

[linuxSystem.localIDP]
type = "container"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	parent, ok := got.Units["linuxSystem"]
	require.True(t, ok, "missing 'linuxSystem' unit")

	child, ok := parent.Subunits["localIDP"]
	require.True(t, ok, "missing 'linuxSystem.localIDP' subunit")
	assert.Equal(t, "Local IDP", child.Name,
		"nested unit with omitted name should humanize from the last segment only (not 'Linux System Local IDP')")
}

// TestParseExplicitNameWins exercises ERGO-05: an explicit `name =` is NEVER
// overwritten by humanization, even when the segment would humanize
// (gRPC → "Grpc"). The authored value renders verbatim.
func TestParseExplicitNameWins(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Explicit Name Wins Test"

[gRPC]
type = "system"
name = "My gRPC Service"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	unit, ok := got.Units["gRPC"]
	require.True(t, ok, "missing 'gRPC' unit")
	assert.Equal(t, "My gRPC Service", unit.Name, "explicit name = must win over humanization (ERGO-05)")
}

// TestParseOmittedNameNoRegression exercises the ERGO-05 backward-compat hard
// contract: every existing fixture already carries explicit `name =` values, so
// the humanize fallback MUST be a no-op for them — every unit's Name must
// remain byte-identical to the authored value.
func TestParseOmittedNameNoRegression(t *testing.T) {
	t.Parallel()

	t.Run("valid.toml", func(t *testing.T) {
		t.Parallel()

		data, err := os.ReadFile("../../testdata/valid.toml")
		require.NoError(t, err, "failed to read testdata/valid.toml")

		got, err := parser.Parse(data)
		require.NoError(t, err, "Parse() should not error")

		expected := map[string]string{
			"user":   "User",
			"webapp": "Web Application",
		}
		for key, wantName := range expected {
			unit, ok := got.Units[key]
			require.True(t, ok, "missing %q unit", key)
			assert.Equal(t, wantName, unit.Name, "%q unit Name must equal its explicit name= value (no regression)", key)
		}
	})

	t.Run("nested.toml", func(t *testing.T) {
		t.Parallel()

		data, err := os.ReadFile("../../testdata/nested.toml")
		require.NoError(t, err, "failed to read testdata/nested.toml")

		got, err := parser.Parse(data)
		require.NoError(t, err, "Parse() should not error")

		externals, ok := got.Units["externals"]
		require.True(t, ok, "missing 'externals' unit")
		assert.Equal(t, "External System", externals.Name,
			"externals.Name must equal its explicit name= value (no regression)")

		mainapp, ok := got.Units["mainapp"]
		require.True(t, ok, "missing 'mainapp' unit")
		assert.Equal(t, "Main Application", mainapp.Name, "mainapp.Name must equal its explicit name= value (no regression)")

		api, ok := mainapp.Subunits["api"]
		require.True(t, ok, "missing 'mainapp.api' subunit")
		assert.Equal(t, "API Server", api.Name, "mainapp.api.Name must equal its explicit name= value (no regression)")

		db, ok := mainapp.Subunits["db"]
		require.True(t, ok, "missing 'mainapp.db' subunit")
		assert.Equal(t, "Database", db.Name, "mainapp.db.Name must equal its explicit name= value (no regression)")

		handler, ok := api.Subunits["handler"]
		require.True(t, ok, "missing 'mainapp.api.handler' subunit")
		assert.Equal(t, "Request Handler", handler.Name,
			"mainapp.api.handler.Name must equal its explicit name= value (no regression)")
	})
}

// --- Phase 31: BC-1 reserved-table parser tests (Plan 31-01) ---
//
// These tests exercise the BC-1 prerequisite (D-08): the parser must skip
// [template.*] / [[use]] / [[include]] tables so they neither create phantom
// units nor leak into the unit loop, and route template/use into new
// Model.Templates / Model.Instantiations fields consumed by Plan 02's
// expansion pass.

// TestParseReservedTablesSkipped exercises the BC-1 skip rule: a document
// containing [template.svc], [[use]], and [[include]] alongside a hand-authored
// [user] unit produces a Model whose UnitOrder/Units contain ONLY the
// hand-authored unit. None of "template", "svc", "use", "include" may leak.
func TestParseReservedTablesSkipped(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/template_reserved.toml")
	require.NoError(t, err, "failed to read template_reserved.toml")

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	// UnitOrder contains ONLY the hand-authored unit, in order.
	assert.Equal(t, []string{"user"}, got.UnitOrder, "UnitOrder must exclude all reserved tables")

	// Units contains ONLY the hand-authored unit.
	assert.Len(t, got.Units, 1, "Units must contain only the hand-authored unit")
	assert.Contains(t, got.Units, "user", "Units must contain 'user'")
	assert.NotContains(t, got.Units, "template", "phantom 'template' parent must not exist")
	assert.NotContains(t, got.Units, "svc", "template name must not leak as a unit")
	assert.NotContains(t, got.Units, "use", "[[use]] must not leak as a unit")
	assert.NotContains(t, got.Units, "include", "[[include]] must not leak as a unit")

	// Model.Templates has the 'svc' key.
	require.NotNil(t, got.Templates, "Templates map must be non-nil when [template.*] present")
	assert.Contains(t, got.Templates, "svc", "Templates must contain 'svc'")

	// Model.Instantiations has exactly one entry (document order).
	require.Len(t, got.Instantiations, 1, "Instantiations must contain the single [[use]] entry")
}

// TestParseTemplateTableExtractsSubtree verifies the [template.<name>] table is
// extracted into Model.Templates with its full parsed *model.Unit subtree,
// INCLUDING [[template.<name>.link]] arrays becoming model.Unit.Links. Params
// are captured but NOT substituted at parse time (Plan 02's job).
func TestParseTemplateTableExtractsSubtree(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/template_reserved.toml")
	require.NoError(t, err, "failed to read template_reserved.toml")

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	require.Contains(t, got.Templates, "svc", "Templates must contain 'svc'")
	tmpl := got.Templates["svc"]
	require.NotNil(t, tmpl, "Templates['svc'] must be non-nil")
	require.NotNil(t, tmpl.Unit, "TemplateDef.Unit must be the parsed subtree")

	assert.Equal(t, model.TypeContainer, tmpl.Unit.Type, "template svc.Type")
	assert.Equal(t, "${name} Service", tmpl.Unit.Name, "template svc.Name (unsubstituted at parse time)")
	assert.Equal(t, "${tech}", tmpl.Unit.Technology, "template svc.Technology (unsubstituted)")

	// Declared params captured.
	assert.ElementsMatch(t, []string{"name", "tech"}, tmpl.Params, "template svc declared params")

	// [[template.svc.link]] parsed into Unit.Links with Peer unsubstituted.
	require.Len(t, tmpl.Unit.Links, 1, "template svc.Links must contain the declared link")
	assert.Equal(t, "${bus}", tmpl.Unit.Links[0].Peer, "link Peer (unsubstituted)")
	assert.Equal(t, "Publishes events", tmpl.Unit.Links[0].Description, "link Description")
}

// TestParseUseArrayPreservesOrder verifies [[use]] array-of-tables entries are
// captured into Model.Instantiations in document order, each carrying the
// template name + supplied named params.
func TestParseUseArrayPreservesOrder(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/template_use_array.toml")
	require.NoError(t, err, "failed to read template_use_array.toml")

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	require.Len(t, got.Instantiations, 2, "two [[use]] blocks must yield two Instantiations")

	assert.Equal(t, "svc", got.Instantiations[0].Template, "Instantiations[0].Template")
	assert.Equal(t, "svc", got.Instantiations[1].Template, "Instantiations[1].Template")

	// Document order: 'alpha' before 'beta'.
	require.NotNil(t, got.Instantiations[0].Params, "Instantiations[0].Params")
	assert.Equal(t, "alpha", got.Instantiations[0].Params["name"], "Instantiations[0] name param (document order)")
	require.NotNil(t, got.Instantiations[1].Params, "Instantiations[1].Params")
	assert.Equal(t, "beta", got.Instantiations[1].Params["name"], "Instantiations[1] name param (document order)")
}

// TestParseUseBeforeTemplate exercises TMPL-09: a [[use]] appearing textually
// BEFORE its [template.<name>] definition parses without error because
// extraction is structured (rawMap post-parse), not textual.
func TestParseUseBeforeTemplate(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/template_forward_ref.toml")
	require.NoError(t, err, "failed to read template_forward_ref.toml")

	got, err := parser.Parse(data)
	require.NoError(t, err, "forward-reference [[use]] before [template.*] must not error (TMPL-09)")

	// Both fields populated despite textual ordering.
	assert.Contains(t, got.Templates, "svc", "Templates['svc'] populated regardless of textual order")
	require.Len(t, got.Instantiations, 1, "forward-ref [[use]] captured")
	assert.Equal(t, "svc", got.Instantiations[0].Template, "forward-ref Instantiation.Template")
}

// TestParseIncludeReservedSkipped verifies [[include]] tables are skipped in
// captureDefinitionOrder (no phantom 'include' unit) and NOT extracted into any
// Model field (reserved for Phase 32).
func TestParseIncludeReservedSkipped(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Include Skip Test"

[[include]]
path = "other.toml"

[user]
type = "person"
name = "User"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	assert.NotContains(t, got.UnitOrder, "include", "[[include]] must not leak into UnitOrder")
	assert.NotContains(t, got.Units, "include", "[[include]] must not leak into Units")
	assert.Equal(t, []string{"user"}, got.UnitOrder, "UnitOrder contains only the hand-authored unit")
}

// TestParseNoRegressionOnHandAuthoredTemplates verifies that existing
// hand-authored fixtures (no reserved tables) parse identically: no Templates,
// no Instantiations, Units/UnitOrder unchanged.
func TestParseNoRegressionOnHandAuthoredTemplates(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/valid.toml")
	require.NoError(t, err, "failed to read testdata/valid.toml")

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	// No reserved tables in valid.toml → Templates empty, Instantiations empty.
	assert.Empty(t, got.Templates, "Templates must be empty for hand-authored-only models")
	assert.Empty(t, got.Instantiations, "Instantiations must be empty for hand-authored-only models")

	// Units/UnitOrder unchanged.
	assert.Equal(t, []string{"user", "webapp"}, got.UnitOrder, "UnitOrder unchanged")
	assert.Len(t, got.Units, 2, "Units count unchanged")
}

// --- Phase 32: IncludeDirective extraction (Plan 32-01) ---
//
// These tests exercise the parser-side half of the include feature (INC-01):
// [[include]] array-of-tables are extracted into Model.Includes in document
// order, each carrying Path and Once; no phantom 'include' unit leaks into
// UnitOrder/Units. They CONSUME the BC-1 skip landed in Plan 31-01.

// TestParseIncludesExtracted exercises INC-01 extraction: two [[include]]
// blocks (path="a.toml" once=true then path="b.toml") plus a hand-authored
// [user] unit populate Model.Includes in document order with the right Path/Once
// values, and produce ZERO phantom 'include' units.
func TestParseIncludesExtracted(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Includes Test"

[[include]]
path = "a.toml"
once = true

[[include]]
path = "b.toml"

[user]
type = "person"
name = "User"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	// Includes has two entries in document order with the right Path/Once.
	require.Len(t, got.Includes, 2, "Includes must contain the two [[include]] blocks in document order")
	assert.Equal(t, "a.toml", got.Includes[0].Path, "Includes[0].Path")
	assert.True(t, got.Includes[0].Once, "Includes[0].Once")
	assert.Equal(t, "b.toml", got.Includes[1].Path, "Includes[1].Path")
	assert.False(t, got.Includes[1].Once, "Includes[1].Once defaults to false when omitted")

	// No phantom 'include' unit.
	assert.Equal(t, []string{"user"}, got.UnitOrder, "UnitOrder must exclude [[include]]")
	assert.Len(t, got.Units, 1, "Units must contain only the hand-authored unit")
	assert.Contains(t, got.Units, "user", "Units must contain 'user'")
	assert.NotContains(t, got.Units, "include", "[[include]] must not leak as a unit")
}

// TestParseIncludesOnceDefaultsFalse exercises the Once zero-value default: a
// single [[include]] path="x.toml" with no `once` key parses into
// Model.Includes[0].Once == false.
func TestParseIncludesOnceDefaultsFalse(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Include Once Default Test"

[[include]]
path = "x.toml"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	require.Len(t, got.Includes, 1, "Includes must contain the single [[include]] block")
	assert.Equal(t, "x.toml", got.Includes[0].Path, "Includes[0].Path")
	assert.False(t, got.Includes[0].Once, "Includes[0].Once must default to false when the key is omitted")
	assert.NotContains(t, got.UnitOrder, "include", "[[include]] must not leak into UnitOrder")
}

// TestParseNoIncludes exercises the no-regression contract: an existing
// single-file fixture (valid.toml) with no [[include]] tables has nil/empty
// Model.Includes and unchanged Properties/UnitOrder/Units.
func TestParseNoIncludes(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../testdata/valid.toml")
	require.NoError(t, err)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	// No [[include]] in valid.toml → Includes empty.
	assert.Empty(t, got.Includes, "Includes must be empty for single-file models")

	// Properties/UnitOrder/Units unchanged from existing behavior.
	assert.Equal(t, "Test System", got.Properties.Name, "Properties.Name unchanged")
	assert.Equal(t, []string{"user", "webapp"}, got.UnitOrder, "UnitOrder unchanged")
	assert.Len(t, got.Units, 2, "Units count unchanged")
}

// --- Phase 35 Plan 02: [[unit.X.use]] sugar + [[template.X.use]] body use ---
//
// D-16: [[unit.<name>.use]] array-of-tables nested under a unit section
// desugars to the SAME Instantiation the top-level [[use]] + parent form
// produces (Parent = enclosing unit dotted path). D-17: [[template.<name>.use]]
// entries populate TemplateDef.Instantiations with Parents relative to the
// template's unit root. Both authoring forms normalize to the single
// Instantiation mechanism the expansion pass consumes.

// TestParseUnitUseSugarEquivalentToParentField exercises D-16: a document
// using [[mainapp.use]] parses to Model.Instantiations identical (deep-equal)
// to the equivalent document using a top-level [[use]] with parent="mainapp".
func TestParseUnitUseSugarEquivalentToParentField(t *testing.T) {
	t.Parallel()

	sugar := []byte(`
[properties]
name = "Unit Use Sugar Test"

[mainapp]
type = "system"
name = "Main App"

[template.svc]
params = ["name", "tech"]
name = "${name} Service"
type = "container"

[[mainapp.use]]
template = "svc"
name = "api"
tech = "Go"
`)

	topLevel := []byte(`
[properties]
name = "Unit Use Sugar Test"

[mainapp]
type = "system"
name = "Main App"

[template.svc]
params = ["name", "tech"]
name = "${name} Service"
type = "container"

[[use]]
template = "svc"
parent = "mainapp"
name = "api"
tech = "Go"
`)

	gotSugar, err := parser.Parse(sugar)
	require.NoError(t, err, "[[mainapp.use]] sugar must parse")

	gotTop, err := parser.Parse(topLevel)
	require.NoError(t, err, "top-level [[use]] parent form must parse")

	// The acceptance criterion: byte-identical Model.Instantiations.
	require.Equal(t, gotTop.Instantiations, gotSugar.Instantiations,
		"[[unit.X.use]] must desugar to the identical Instantiation the [[use]] parent form produces")

	// And the shape of the single desugared entry.
	require.Len(t, gotSugar.Instantiations, 1, "one [[mainapp.use]] entry")
	assert.Equal(t, "svc", gotSugar.Instantiations[0].Template, "Instantiations[0].Template")
	assert.Equal(t, "mainapp", gotSugar.Instantiations[0].Parent, "Instantiations[0].Parent (enclosing unit path)")
	require.NotNil(t, gotSugar.Instantiations[0].Params, "Instantiations[0].Params")
	assert.Equal(t, "api", gotSugar.Instantiations[0].Params["name"], "name param")
	assert.Equal(t, "Go", gotSugar.Instantiations[0].Params["tech"], "tech param")

	// The sugar must not create a phantom 'use' subunit under the unit.
	mainapp := gotSugar.Units["mainapp"]
	require.NotNil(t, mainapp, "mainapp unit present")
	assert.NotContains(t, mainapp.Subunits, "use", "'use' must not leak as a phantom subunit (BC-1)")
	assert.NotContains(t, mainapp.SubunitOrder, "use", "SubunitOrder must not contain 'use'")

	// UnitOrder unchanged: only the hand-authored unit.
	assert.Equal(t, []string{"mainapp"}, gotSugar.UnitOrder, "UnitOrder excludes the use site")
}

// TestParseUnitUseNestedPath exercises D-16 at depth: [[a.b.use]] (use under a
// nested subunit section) desugars to Parent "a.b".
func TestParseUnitUseNestedPath(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Nested Unit Use Test"

[a]
type = "system"
name = "A"

[a.b]
type = "container"
name = "B"

[template.svc]
params = ["name"]
name = "${name} Service"
type = "component"

[[a.b.use]]
template = "svc"
name = "handler"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "[[a.b.use]] must parse")

	require.Len(t, got.Instantiations, 1, "one nested use entry")
	assert.Equal(t, "svc", got.Instantiations[0].Template, "Template")
	assert.Equal(t, "a.b", got.Instantiations[0].Parent, "Parent is the full dotted enclosing path")

	// No phantom 'use' subunit under a.b (the fallback parse must skip it).
	b := got.Units["a"].Subunits["b"]
	require.NotNil(t, b, "a.b subunit present")
	assert.NotContains(t, b.Subunits, "use", "'use' must not leak as a phantom subunit of a.b")
}

// TestParseUnitUseInterleavesDocumentOrder exercises D-16 ordering: top-level
// [[use]] entries and [[unit.X.use]] entries interleave in Model.Instantiations
// in AUTHORING order, not grouped by form.
func TestParseUnitUseInterleavesDocumentOrder(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Interleave Test"

[template.svc]
params = ["name"]
name = "${name} Service"
type = "system"

[beta]
type = "system"
name = "Beta"

[[beta.use]]
template = "svc"
name = "first"

[[use]]
template = "svc"
name = "second"

[alpha]
type = "system"
name = "Alpha"

[[alpha.use]]
template = "svc"
name = "third"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "Parse() should not error")

	require.Len(t, got.Instantiations, 3, "three use entries across both forms")

	wantOrder := []struct{ name, parent string }{
		{"first", "beta"},
		{"second", ""},
		{"third", "alpha"},
	}

	for i, want := range wantOrder {
		assert.Equal(t, "svc", got.Instantiations[i].Template, "Instantiations[%d].Template", i)
		assert.Equal(t, want.parent, got.Instantiations[i].Parent, "Instantiations[%d].Parent (document order)", i)
		assert.Equal(t, want.name, got.Instantiations[i].Params["name"], "Instantiations[%d] name param (document order)", i)
	}
}

// TestParseTemplateBodyUseExtracted exercises D-17 parsing: [[template.svc.use]]
// entries land in Templates["svc"].Instantiations with Parent relative to the
// template's unit root (empty for the root; nested [[template.svc.api.use]]
// yields Parent "api"). Body uses do NOT leak into Model.Instantiations.
func TestParseTemplateBodyUseExtracted(t *testing.T) {
	t.Parallel()

	data := []byte(`
[properties]
name = "Template Body Use Test"

[template.leaf]
params = ["name"]
name = "${name} Leaf"
type = "db"

[template.svc]
params = ["name"]
name = "${name} Service"
type = "box"

[template.svc.api]
type = "db"
name = "API"

[[template.svc.use]]
template = "leaf"
name = "root-cache"

[[template.svc.api.use]]
template = "leaf"
name = "api-cache"
`)

	got, err := parser.Parse(data)
	require.NoError(t, err, "[[template.svc.use]] must parse")

	require.Contains(t, got.Templates, "svc", "Templates must contain 'svc'")
	svc := got.Templates["svc"]
	require.NotNil(t, svc, "Templates['svc'] non-nil")

	// Two body uses, in document order, Parents relative to the template root.
	require.Len(t, svc.Instantiations, 2, "svc.Instantiations has the two body uses")

	assert.Equal(t, "leaf", svc.Instantiations[0].Template, "svc.Instantiations[0].Template")
	assert.Empty(t, svc.Instantiations[0].Parent, "svc.Instantiations[0].Parent (empty = template root)")
	assert.Equal(t, "root-cache", svc.Instantiations[0].Params["name"], "svc.Instantiations[0] name param")

	assert.Equal(t, "leaf", svc.Instantiations[1].Template, "svc.Instantiations[1].Template")
	assert.Equal(t, "api", svc.Instantiations[1].Parent, "svc.Instantiations[1].Parent (relative to template root)")
	assert.Equal(t, "api-cache", svc.Instantiations[1].Params["name"], "svc.Instantiations[1] name param")

	// Body uses never leak into Model.Instantiations.
	assert.Empty(t, got.Instantiations, "template body uses stay out of Model.Instantiations")

	// The template's own subtree parse is unaffected by the use arrays.
	assert.Contains(t, svc.Unit.Subunits, "api", "declared subunit api still parsed")
	assert.NotContains(t, svc.Unit.Subunits, "use", "'use' must not leak as a phantom template subunit")
}

// TestParseUnitUseNonStringValueRejected mirrors extractInstantiations'
// strictness into the nested forms: a [[unit.X.use]] (and [[template.X.use]])
// entry carrying a non-string value is a hard *parser.ParseError.
func TestParseUnitUseNonStringValueRejected(t *testing.T) {
	t.Parallel()

	t.Run("unit use", func(t *testing.T) {
		t.Parallel()

		data := []byte(`
[properties]
name = "Bad Unit Use Test"

[mainapp]
type = "system"

[template.svc]
params = ["name"]
name = "${name} Service"
type = "container"

[[mainapp.use]]
template = "svc"
name = 123
`)

		_, err := parser.Parse(data)
		require.Error(t, err, "non-string value in [[unit.X.use]] must be a hard error")

		var parseErr *parser.ParseError
		require.ErrorAs(t, err, &parseErr, "error should be *ParseError")
	})

	t.Run("template body use", func(t *testing.T) {
		t.Parallel()

		data := []byte(`
[properties]
name = "Bad Template Use Test"

[template.leaf]
params = ["name"]
name = "${name} Leaf"
type = "db"

[template.svc]
params = ["name"]
name = "${name} Service"
type = "box"

[[template.svc.use]]
template = "leaf"
name = ["not", "a", "string"]
`)

		_, err := parser.Parse(data)
		require.Error(t, err, "non-string value in [[template.X.use]] must be a hard error")

		var parseErr *parser.ParseError
		require.ErrorAs(t, err, &parseErr, "error should be *ParseError")
	})
}

func TestParseLegendProperties(t *testing.T) {
	t.Parallel()

	t.Run("legend absent defaults to nil (enabled)", func(t *testing.T) {
		t.Parallel()

		got, err := parser.Parse([]byte("[properties]\nname = \"T\"\n"))
		require.NoError(t, err)
		assert.Nil(t, got.Properties.Legend, "absent legend is nil — default-on semantics")
	})

	t.Run("legend false parses to false", func(t *testing.T) {
		t.Parallel()

		got, err := parser.Parse([]byte("[properties]\nname = \"T\"\nlegend = false\n"))
		require.NoError(t, err)
		require.NotNil(t, got.Properties.Legend)
		assert.False(t, *got.Properties.Legend)
	})

	t.Run("legend true parses to true", func(t *testing.T) {
		t.Parallel()

		got, err := parser.Parse([]byte("[properties]\nname = \"T\"\nlegend = true\n"))
		require.NoError(t, err)
		require.NotNil(t, got.Properties.Legend)
		assert.True(t, *got.Properties.Legend)
	})

	t.Run("legendLine rows parse all fields", func(t *testing.T) {
		t.Parallel()

		data := []byte(`
[properties]
name = "T"

[[properties.legendLine]]
label = "Batch import"
color = "#C0392B"
style = "dashed"

[[properties.legendLine]]
label = "Custom flow"
color = "blue"
`)

		got, err := parser.Parse(data)
		require.NoError(t, err)
		require.Len(t, got.Properties.LegendLines, 2)

		first := got.Properties.LegendLines[0]
		assert.Equal(t, "Batch import", first.Label)
		assert.Equal(t, "#C0392B", first.Color)
		assert.Equal(t, "dashed", first.Style)

		assert.Equal(t, "Custom flow", got.Properties.LegendLines[1].Label)
	})

	t.Run("legendLine does not register a phantom properties unit", func(t *testing.T) {
		t.Parallel()

		data := []byte(`
[properties]
name = "T"

[[properties.legendLine]]
label = "X"
color = "red"

[app]
type = "system"
name = "App"
`)

		got, err := parser.Parse(data)
		require.NoError(t, err)
		assert.Len(t, got.Units, 1, "only the app unit exists")
	})
}
