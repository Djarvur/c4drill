package parser

import (
	"os"
	"strings"
	"testing"

	"github.com/Djarvur/c4drill/internal/model"
)

func TestParseValid(t *testing.T) {
	data, err := os.ReadFile("../../testdata/valid.toml")
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	got, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	if got.Properties.Name != "Test System" {
		t.Errorf("Properties.Name = %q, want %q", got.Properties.Name, "Test System")
	}

	if got.Properties.Description != "A test architecture" {
		t.Errorf("Properties.Description = %q, want %q", got.Properties.Description, "A test architecture")
	}

	if got.Properties.LineLength != 40 {
		t.Errorf("Properties.LineLength = %d, want 40", got.Properties.LineLength)
	}

	if len(got.Properties.Expanded) != 1 || got.Properties.Expanded[0] != "webapp" {
		t.Errorf("Properties.Expanded = %v, want [webapp]", got.Properties.Expanded)
	}

	if len(got.Units) != 2 {
		t.Fatalf("len(Units) = %d, want 2", len(got.Units))
	}

	user, ok := got.Units["user"]
	if !ok {
		t.Fatal("missing 'user' unit")
	}
	if user.Type != model.TypePerson {
		t.Errorf("user.Type = %q, want %q", user.Type, model.TypePerson)
	}
	if user.Name != "User" {
		t.Errorf("user.Name = %q, want %q", user.Name, "User")
	}
	if user.Description != "End user of the system" {
		t.Errorf("user.Description = %q, want %q", user.Description, "End user of the system")
	}

	webapp, ok := got.Units["webapp"]
	if !ok {
		t.Fatal("missing 'webapp' unit")
	}
	if webapp.Type != model.TypeSystem {
		t.Errorf("webapp.Type = %q, want %q", webapp.Type, model.TypeSystem)
	}
	if webapp.Name != "Web Application" {
		t.Errorf("webapp.Name = %q, want %q", webapp.Name, "Web Application")
	}
	if webapp.Technology != "Go, React" {
		t.Errorf("webapp.Technology = %q, want %q", webapp.Technology, "Go, React")
	}
}

func TestParseNestedUnits(t *testing.T) {
	data, err := os.ReadFile("../../testdata/nested.toml")
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	got, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	if got.Properties.Name != "Nested Test" {
		t.Errorf("Properties.Name = %q, want %q", got.Properties.Name, "Nested Test")
	}

	if len(got.Units) != 2 {
		t.Fatalf("len(Units) = %d, want 2", len(got.Units))
	}

	externals, ok := got.Units["externals"]
	if !ok {
		t.Fatal("missing 'externals' unit")
	}
	if externals.Type != model.TypeSystemExternal {
		t.Errorf("externals.Type = %q, want %q", externals.Type, model.TypeSystemExternal)
	}

	mainapp, ok := got.Units["mainapp"]
	if !ok {
		t.Fatal("missing 'mainapp' unit")
	}
	if mainapp.Type != model.TypeSystem {
		t.Errorf("mainapp.Type = %q, want %q", mainapp.Type, model.TypeSystem)
	}
	if mainapp.Name != "Main Application" {
		t.Errorf("mainapp.Name = %q, want %q", mainapp.Name, "Main Application")
	}
	if mainapp.Technology != "Go" {
		t.Errorf("mainapp.Technology = %q, want %q", mainapp.Technology, "Go")
	}

	if len(mainapp.Subunits) != 2 {
		t.Fatalf("len(mainapp.Subunits) = %d, want 2", len(mainapp.Subunits))
	}

	api, ok := mainapp.Subunits["api"]
	if !ok {
		t.Fatal("missing 'mainapp.api' subunit")
	}
	if api.Type != model.TypeContainer {
		t.Errorf("api.Type = %q, want %q", api.Type, model.TypeContainer)
	}
	if api.Name != "API Server" {
		t.Errorf("api.Name = %q, want %q", api.Name, "API Server")
	}

	db, ok := mainapp.Subunits["db"]
	if !ok {
		t.Fatal("missing 'mainapp.db' subunit")
	}
	if db.Type != model.TypeContainerDb {
		t.Errorf("db.Type = %q, want %q", db.Type, model.TypeContainerDb)
	}
	if db.Name != "Database" {
		t.Errorf("db.Name = %q, want %q", db.Name, "Database")
	}

	if len(api.Subunits) != 1 {
		t.Fatalf("len(api.Subunits) = %d, want 1", len(api.Subunits))
	}

	handler, ok := api.Subunits["handler"]
	if !ok {
		t.Fatal("missing 'mainapp.api.handler' subunit")
	}
	if handler.Type != model.TypeComponent {
		t.Errorf("handler.Type = %q, want %q", handler.Type, model.TypeComponent)
	}
	if handler.Name != "Request Handler" {
		t.Errorf("handler.Name = %q, want %q", handler.Name, "Request Handler")
	}
}

func TestParseLinks(t *testing.T) {
	data, err := os.ReadFile("../../testdata/links.toml")
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	got, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	if got.Properties.Name != "Links Test" {
		t.Errorf("Properties.Name = %q, want %q", got.Properties.Name, "Links Test")
	}

	if got.Properties.Edges != "spline" {
		t.Errorf("Properties.Edges = %q, want %q", got.Properties.Edges, "spline")
	}

	webapp, ok := got.Units["webapp"]
	if !ok {
		t.Fatal("missing 'webapp' unit")
	}

	if len(webapp.Links) != 1 {
		t.Fatalf("len(webapp.Links) = %d, want 1", len(webapp.Links))
	}

	link, ok := webapp.Links["user"]
	if !ok {
		t.Fatal("missing 'user' link in webapp")
	}

	if link.Target != "user" {
		t.Errorf("link.Target = %q, want %q", link.Target, "user")
	}
	if link.Arrow != model.ArrowForward {
		t.Errorf("link.Arrow = %q, want %q", link.Arrow, model.ArrowForward)
	}
	if link.Rank != model.RankForward {
		t.Errorf("link.Rank = %q, want %q", link.Rank, model.RankForward)
	}
	if link.Technology != "HTTPS" {
		t.Errorf("link.Technology = %q, want %q", link.Technology, "HTTPS")
	}
	if link.Description != "Uses" {
		t.Errorf("link.Description = %q, want %q", link.Description, "Uses")
	}

	api, ok := got.Units["api"]
	if !ok {
		t.Fatal("missing 'api' unit")
	}

	if len(api.LinksFrom) != 1 {
		t.Fatalf("len(api.LinksFrom) = %d, want 1", len(api.LinksFrom))
	}

	linkFrom, ok := api.LinksFrom["webapp"]
	if !ok {
		t.Fatal("missing 'webapp' linkFrom in api")
	}

	if linkFrom.Target != "webapp" {
		t.Errorf("linkFrom.Target = %q, want %q", linkFrom.Target, "webapp")
	}
	if linkFrom.Arrow != model.ArrowForward {
		t.Errorf("linkFrom.Arrow = %q, want %q", linkFrom.Arrow, model.ArrowForward)
	}
	if linkFrom.Technology != "HTTP/JSON" {
		t.Errorf("linkFrom.Technology = %q, want %q", linkFrom.Technology, "HTTP/JSON")
	}
}

func TestParseInvalidTOML(t *testing.T) {
	invalidData := []byte("invalid [[[")

	_, err := Parse(invalidData)
	if err == nil {
		t.Fatal("Parse() error = nil, want error for invalid TOML")
	}

	var parseErr *ParseError
	if !strings.Contains(err.Error(), "parse error") {
		t.Errorf("error message should contain 'parse error', got %q", err.Error())
	}

	// Check that the error is a ParseError
	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("error type = %T, want *ParseError", err)
	}

	// Line should be extracted from DecodeError
	if parseErr.Line == 0 {
		t.Error("ParseError.Line = 0, want non-zero for invalid TOML")
	}
}

func TestParseMissingFile(t *testing.T) {
	_, err := ParseFile("nonexistent.toml")
	if err == nil {
		t.Fatal("ParseFile() error = nil, want error for missing file")
	}

	var parseErr *ParseError
	parseErr, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("error type = %T, want *ParseError", err)
	}

	if parseErr.Message != "failed to read file" {
		t.Errorf("ParseError.Message = %q, want %q", parseErr.Message, "failed to read file")
	}

	if parseErr.Context != "nonexistent.toml" {
		t.Errorf("ParseError.Context = %q, want %q", parseErr.Context, "nonexistent.toml")
	}
}

func TestParsePropertiesFields(t *testing.T) {
	data, err := os.ReadFile("../../testdata/valid.toml")
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	got, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	if got.Properties.Color != "transparent" {
		t.Errorf("Properties.Color = %q, want %q", got.Properties.Color, "transparent")
	}

	if got.Properties.Edges != "straight" {
		t.Errorf("Properties.Edges = %q, want %q", got.Properties.Edges, "straight")
	}

	if got.Properties.LineLength != 40 {
		t.Errorf("Properties.LineLength = %d, want 40", got.Properties.LineLength)
	}

	if len(got.Properties.Expanded) != 1 {
		t.Errorf("len(Properties.Expanded) = %d, want 1", len(got.Properties.Expanded))
	}

	if got.Properties.Expanded[0] != "webapp" {
		t.Errorf("Properties.Expanded[0] = %q, want %q", got.Properties.Expanded[0], "webapp")
	}
}

func TestParseUnitFields(t *testing.T) {
	data, err := os.ReadFile("../../testdata/valid.toml")
	if err != nil {
		t.Fatalf("failed to read test fixture: %v", err)
	}

	got, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	webapp, ok := got.Units["webapp"]
	if !ok {
		t.Fatal("missing 'webapp' unit")
	}

	if webapp.Type != model.TypeSystem {
		t.Errorf("webapp.Type = %q, want %q", webapp.Type, model.TypeSystem)
	}

	if webapp.Name != "Web Application" {
		t.Errorf("webapp.Name = %q, want %q", webapp.Name, "Web Application")
	}

	if webapp.Description != "Main web application" {
		t.Errorf("webapp.Description = %q, want %q", webapp.Description, "Main web application")
	}

	if webapp.Technology != "Go, React" {
		t.Errorf("webapp.Technology = %q, want %q", webapp.Technology, "Go, React")
	}
}

func TestParseEmptyFile(t *testing.T) {
	got, err := Parse([]byte(""))
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil for empty file", err)
	}

	if got.Properties.Name != "" {
		t.Errorf("Properties.Name = %q, want empty string", got.Properties.Name)
	}

	if len(got.Units) != 0 {
		t.Errorf("len(Units) = %d, want 0", len(got.Units))
	}
}

func TestParseOnlyProperties(t *testing.T) {
	data := []byte(`
[properties]
name = "Only Properties Test"
description = "Test with only properties"
`)

	got, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	if got.Properties.Name != "Only Properties Test" {
		t.Errorf("Properties.Name = %q, want %q", got.Properties.Name, "Only Properties Test")
	}

	if len(got.Units) != 0 {
		t.Errorf("len(Units) = %d, want 0", len(got.Units))
	}
}

func TestParseExternalTypes(t *testing.T) {
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

	got, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	if got.Units["personext"].Type != model.TypePersonExternal {
		t.Errorf("personext.Type = %q, want %q", got.Units["personext"].Type, model.TypePersonExternal)
	}

	if got.Units["systemext"].Type != model.TypeSystemExternal {
		t.Errorf("systemext.Type = %q, want %q", got.Units["systemext"].Type, model.TypeSystemExternal)
	}

	if got.Units["dbext"].Type != model.TypeDbExternal {
		t.Errorf("dbext.Type = %q, want %q", got.Units["dbext"].Type, model.TypeDbExternal)
	}

	if got.Units["queueext"].Type != model.TypeQueueExternal {
		t.Errorf("queueext.Type = %q, want %q", got.Units["queueext"].Type, model.TypeQueueExternal)
	}
}

func TestParseLinkFrom(t *testing.T) {
	data := []byte(`
[properties]
name = "LinkFrom Test"

[a]
type = "system"
name = "System A"
linkFrom = { "b" = { arrow = "reverse", technology = "TCP" } }
`)

	got, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	unitA, ok := got.Units["a"]
	if !ok {
		t.Fatal("missing 'a' unit")
	}

	if len(unitA.LinksFrom) != 1 {
		t.Fatalf("len(unitA.LinksFrom) = %d, want 1", len(unitA.LinksFrom))
	}

	linkFrom, ok := unitA.LinksFrom["b"]
	if !ok {
		t.Fatal("missing 'b' linkFrom in a")
	}

	if linkFrom.Target != "b" {
		t.Errorf("linkFrom.Target = %q, want %q", linkFrom.Target, "b")
	}

	if linkFrom.Arrow != model.ArrowReverse {
		t.Errorf("linkFrom.Arrow = %q, want %q", linkFrom.Arrow, model.ArrowReverse)
	}

	if linkFrom.Technology != "TCP" {
		t.Errorf("linkFrom.Technology = %q, want %q", linkFrom.Technology, "TCP")
	}
}

func TestParseAllUnitTypes(t *testing.T) {
	data := []byte(`
[properties]
name = "All Types Test"

# C1 types
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

# C2 types
[c1]
type = "container"
[cd1]
type = "containerDb"
[cq1]
type = "containerQueue"

# C3 types
[cmp1]
type = "component"
[cmpd1]
type = "componentDb"
[cmpq1]
type = "componentQueue"
`)

	got, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	expectedTypes := map[string]model.UnitType{
		"p1":    model.TypePerson,
		"s1":    model.TypeSystem,
		"d1":    model.TypeDb,
		"q1":    model.TypeQueue,
		"b1":    model.TypeBox,
		"c1":    model.TypeContainer,
		"cd1":   model.TypeContainerDb,
		"cq1":   model.TypeContainerQueue,
		"cmp1":  model.TypeComponent,
		"cmpd1": model.TypeComponentDb,
		"cmpq1": model.TypeComponentQueue,
	}

	for name, expectedType := range expectedTypes {
		unit, ok := got.Units[name]
		if !ok {
			t.Errorf("missing unit %q", name)
			continue
		}
		if unit.Type != expectedType {
			t.Errorf("%s.Type = %q, want %q", name, unit.Type, expectedType)
		}
	}
}

func TestParseLinkAllFields(t *testing.T) {
	data := []byte(`
[properties]
name = "Link Fields Test"

[a]
type = "system"
name = "A"
link = { "b" = { arrow = "bidirectional", rank = "equal", color = "red", style = "dashed", technology = "gRPC", description = "syncs data", labelPosition = "head" } }
`)

	got, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	unitA, ok := got.Units["a"]
	if !ok {
		t.Fatal("missing 'a' unit")
	}

	link, ok := unitA.Links["b"]
	if !ok {
		t.Fatal("missing 'b' link in a")
	}

	if link.Target != "b" {
		t.Errorf("link.Target = %q, want %q", link.Target, "b")
	}
	if link.Arrow != model.ArrowBidirectional {
		t.Errorf("link.Arrow = %q, want %q", link.Arrow, model.ArrowBidirectional)
	}
	if link.Rank != model.RankEqual {
		t.Errorf("link.Rank = %q, want %q", link.Rank, model.RankEqual)
	}
	if link.Color != "red" {
		t.Errorf("link.Color = %q, want %q", link.Color, "red")
	}
	if link.Style != "dashed" {
		t.Errorf("link.Style = %q, want %q", link.Style, "dashed")
	}
	if link.Technology != "gRPC" {
		t.Errorf("link.Technology = %q, want %q", link.Technology, "gRPC")
	}
	if link.Description != "syncs data" {
		t.Errorf("link.Description = %q, want %q", link.Description, "syncs data")
	}
	if link.LabelPosition != model.LabelHead {
		t.Errorf("link.LabelPosition = %q, want %q", link.LabelPosition, model.LabelHead)
	}
}

func TestParseNestedLink(t *testing.T) {
	data := []byte(`
[properties]
name = "Nested Link Test"

[parent]
type = "system"
name = "Parent"

[parent.child]
type = "container"
name = "Child"
link = { "other" = { technology = "HTTP" } }
`)

	got, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	parent, ok := got.Units["parent"]
	if !ok {
		t.Fatal("missing 'parent' unit")
	}

	child, ok := parent.Subunits["child"]
	if !ok {
		t.Fatal("missing 'parent.child' subunit")
	}

	if len(child.Links) != 1 {
		t.Fatalf("len(child.Links) = %d, want 1", len(child.Links))
	}

	link, ok := child.Links["other"]
	if !ok {
		t.Fatal("missing 'other' link in child")
	}

	if link.Target != "other" {
		t.Errorf("link.Target = %q, want %q", link.Target, "other")
	}
}

func TestParseFile(t *testing.T) {
	got, err := ParseFile("../../testdata/valid.toml")
	if err != nil {
		t.Fatalf("ParseFile() error = %v, want nil", err)
	}

	if got.Properties.Name != "Test System" {
		t.Errorf("Properties.Name = %q, want %q", got.Properties.Name, "Test System")
	}
}
