package c4d

import (
	"strconv"
	"strings"

	"github.com/Djarvur/c4drill/internal/c4d/ast"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
)

// FromModel converts m into the typed C4D AST — the canonical Model-to-AST
// inverse of toModel (35-05). Units land in UnitOrder, links become edge
// statements with the D-05 arrow-glyph inverse, templates become
// declarations, instantiations become use statements placed at their parent
// (D-16: a parented use lives inside that unit's block; a parentless use at
// top level), and includes become include statements.
//
// Comments stay nil on this path: the Model carries no trivia, so conversion
// loses comments by design — only fmt preserves them (D-32). Emission walks
// order slices only, never map iteration (T-35-04-01); template names sort
// deterministically and [[use]] params order by sorted key.
func FromModel(m *parser.Model) *ast.Document {
	doc := &ast.Document{}
	if m == nil {
		return doc
	}

	doc.Properties = propertiesBlockFromModel(&m.Properties)

	for _, name := range m.UnitOrder {
		if unit, ok := m.Units[name]; ok {
			doc.Units = append(doc.Units, unitNodeFromModel(name, unit))
		}
	}

	for _, name := range sortedTemplateNames(m.Templates) {
		if tmpl := m.Templates[name]; tmpl != nil && tmpl.Unit != nil {
			doc.Templates = append(doc.Templates, templateDeclFromModel(name, tmpl))
		}
	}

	for _, inst := range m.Instantiations {
		us := useStmtFromModel(inst)

		if host := findUnitByPath(doc.Units, inst.Parent); host != nil {
			host.UseStmts = append(host.UseStmts, us)
		} else {
			doc.UseStmts = append(doc.UseStmts, us)
		}
	}

	for _, inc := range m.Includes {
		doc.Includes = append(doc.Includes, &ast.IncludeStmt{Path: inc.Path, Once: inc.Once})
	}

	return doc
}

// propertiesBlockFromModel builds the properties block in the D-23 property
// field order: name, description, color, style, border, edges, lineLength,
// expanded. Returns nil when no field is set (an empty block emits nothing).
func propertiesBlockFromModel(props *model.Properties) *ast.PropertiesBlock {
	block := &ast.PropertiesBlock{}

	add := func(key, value string) {
		if value != "" {
			block.Fields = append(block.Fields, &ast.FieldStmt{Key: key, Value: literalFor(value)})
		}
	}

	add("name", props.Name)
	add("description", props.Description)
	add("color", props.Color)
	add("style", props.Style)
	add("border", props.Border)
	add("edges", props.Edges)

	if props.LineLength != 0 {
		block.Fields = append(block.Fields, &ast.FieldStmt{
			Key:   "lineLength",
			Value: ast.Literal{Kind: ast.KindBareword, Str: strconv.Itoa(props.LineLength)},
		})
	}

	if len(props.Expanded) > 0 {
		block.Fields = append(block.Fields, &ast.FieldStmt{
			Key:   "expanded",
			Value: ast.Literal{Kind: ast.KindList, List: props.Expanded},
		})
	}

	if len(block.Fields) == 0 {
		return nil
	}

	return block
}

// unitNodeFromModel builds a UnitNode: ID from the map key, the type split
// into its base keyword plus the external modifier (D-04), Name in the
// header slot, then body fields (D-23 order), edges from Links/LinksFrom and
// subunits in SubunitOrder.
func unitNodeFromModel(id string, unit *model.Unit) *ast.UnitNode {
	if unit == nil {
		return nil
	}

	base, external := splitExternalType(unit.Type)
	node := &ast.UnitNode{ID: id, Type: base, External: external, Name: unit.Name}

	appendUnitBody(node, unit)

	for _, subName := range unit.SubunitOrder {
		if sub, ok := unit.Subunits[subName]; ok {
			node.Subunits = append(node.Subunits, unitNodeFromModel(subName, sub))
		}
	}

	return node
}

// appendUnitBody fills a unit node's fields and edges from a model unit in
// the D-23 body field order. Mirror links are validator bookkeeping and
// never map to statements.
func appendUnitBody(node *ast.UnitNode, unit *model.Unit) {
	add := func(key, value string) {
		if value != "" {
			node.Fields = append(node.Fields, &ast.FieldStmt{Key: key, Value: literalFor(value)})
		}
	}

	add("description", unit.Description)
	add("technology", unit.Technology)
	add("reference", unit.Reference)
	add("color", unit.Color)
	add("style", unit.Style)
	add("border", unit.Border)
	add("edges", unit.Edges)

	if len(unit.Expanded) > 0 {
		node.Fields = append(node.Fields, &ast.FieldStmt{
			Key:   "expanded",
			Value: ast.Literal{Kind: ast.KindList, List: unit.Expanded},
		})
	}

	for _, link := range unit.Links {
		if !link.IsMirror() {
			node.Edges = append(node.Edges, edgeStmtFromLink(link, false))
		}
	}

	for _, link := range unit.LinksFrom {
		if !link.IsMirror() {
			node.Edges = append(node.Edges, edgeStmtFromLink(link, true))
		}
	}
}

// edgeStmtFromLink maps a model link to an edge statement. Outgoing links
// use the D-05 arrow-glyph inverse (ArrowReverse renders through the
// linkFrom statement form "<-"); LinksFrom entries always emit "<-".
// Options follow the canonical option order: rank, color, style,
// labelPosition, length.
func edgeStmtFromLink(link model.Link, from bool) *ast.EdgeStmt {
	glyph := "<-"
	if !from {
		glyph = arrowGlyphFor(link.Arrow)
	}

	edge := &ast.EdgeStmt{
		ArrowGlyph:  glyph,
		Peer:        link.Peer,
		Technology:  link.Technology,
		Description: link.Description,
	}

	addOpt := func(key, value string) {
		if value != "" {
			edge.Options = append(edge.Options, &ast.FieldStmt{Key: key, Value: literalFor(value)})
		}
	}

	addOpt("rank", string(link.Rank))
	addOpt("color", link.Color)
	addOpt("style", link.Style)
	addOpt("labelPosition", string(link.LabelPosition))

	if link.Length != 0 {
		edge.Options = append(edge.Options, &ast.FieldStmt{
			Key:   "length",
			Value: ast.Literal{Kind: ast.KindBareword, Str: strconv.Itoa(link.Length)},
		})
	}

	return edge
}

// arrowGlyphFor maps the model arrow direction to the C4D glyph — the
// inverse of the Model conversion's glyph mapping.
func arrowGlyphFor(a model.ArrowDirection) string {
	switch a {
	case model.ArrowBidirectional:
		return "<->"
	case model.ArrowNone:
		return "--"
	case model.ArrowReverse:
		return "<-"
	case model.ArrowForward:
		return "->"
	default:
		return "->"
	}
}

// templateDeclFromModel builds a template declaration. The body carries the
// template unit's fields (name included — `name` is a legal body field,
// unlike the header-only unit form), edges and subunits; subunits build
// BEFORE uses so parented uses can find their host block. The body's
// Type/External slots record the template root's type, but the C4D grammar
// has no template-root-type syntax (D-13 bodies are typeless at the root),
// so EmitC4D does not render them — see the 35-04 SUMMARY known-stub note.
func templateDeclFromModel(name string, tmpl *parser.TemplateDef) *ast.TemplateDecl {
	decl := &ast.TemplateDecl{Name: name, Params: tmpl.Params}

	body := &ast.UnitNode{}
	body.Type, body.External = splitExternalType(tmpl.Unit.Type)

	if tmpl.Unit.Name != "" {
		body.Fields = append(body.Fields, &ast.FieldStmt{Key: "name", Value: literalFor(tmpl.Unit.Name)})
	}

	appendUnitBody(body, tmpl.Unit)

	for _, subName := range tmpl.Unit.SubunitOrder {
		if sub, ok := tmpl.Unit.Subunits[subName]; ok {
			body.Subunits = append(body.Subunits, unitNodeFromModel(subName, sub))
		}
	}

	// D-17: a template-body use whose root-relative parent names a subunit
	// path lands inside that subunit's block; unresolvable parents fall back
	// to the body root so the instantiation is never dropped.
	for _, inst := range tmpl.Instantiations {
		us := useStmtFromModel(inst)

		if host := findUnitByPath(body.Subunits, inst.Parent); host != nil {
			host.UseStmts = append(host.UseStmts, us)
		} else {
			body.UseStmts = append(body.UseStmts, us)
		}
	}

	decl.Body = body

	return decl
}

// useStmtFromModel converts an Instantiation into a use statement with named
// args. Params is a map in the Model, so args order by sorted key — the only
// deterministic order available (D-23 pins field order, not arg order).
func useStmtFromModel(inst parser.Instantiation) *ast.UseStmt {
	us := &ast.UseStmt{Template: inst.Template}

	for _, key := range sortedParamKeys(inst.Params) {
		us.Args = append(us.Args, ast.Arg{Name: key, Value: literalFor(inst.Params[key])})
	}

	return us
}

// splitExternalType splits a level-specific external type keyword (D-04)
// into its base keyword and the external modifier flag: systemExternal ->
// ("system", true); db -> ("db", false).
func splitExternalType(t model.UnitType) (string, bool) {
	s := string(t)
	if strings.HasSuffix(s, "External") && len(s) > len("External") {
		return strings.TrimSuffix(s, "External"), true
	}

	return s, false
}

// findUnitByPath walks nested UnitNodes along a dotted path and returns the
// deepest match, nil when any segment misses ("" never matches).
func findUnitByPath(units []*ast.UnitNode, path string) *ast.UnitNode {
	if path == "" {
		return nil
	}

	var found *ast.UnitNode

	for _, seg := range strings.Split(path, ".") {
		found = nil

		for _, u := range units {
			if u != nil && u.ID == seg {
				found = u

				break
			}
		}

		if found == nil {
			return nil
		}

		units = found.Subunits
	}

	return found
}

// literalFor picks the canonical C4D literal form for a string value (D-06):
// newline-containing values become triple-quoted (the D-06 inverse of Task
// 1's escaped-\n TOML rule) unless the value itself contains the triple-quote
// delimiter, which cannot ride raw and falls back to the escaped quoted
// form; bareword-safe values stay barewords; everything else is quoted.
func literalFor(s string) ast.Literal {
	switch {
	case strings.ContainsAny(s, "\n\r"):
		if !strings.Contains(s, `"""`) {
			return ast.Literal{Kind: ast.KindTriple, Str: s}
		}

		return ast.Literal{Kind: ast.KindQuoted, Str: s}
	case c4dBarewordSafe(s):
		return ast.Literal{Kind: ast.KindBareword, Str: s}
	default:
		return ast.Literal{Kind: ast.KindQuoted, Str: s}
	}
}
