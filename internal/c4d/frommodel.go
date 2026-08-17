package c4d

import (
	"fmt"
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

	if unit.Width != 0 {
		node.Fields = append(node.Fields, &ast.FieldStmt{
			Key:   "width",
			Value: ast.Literal{Kind: ast.KindBareword, Str: formatFloat(unit.Width)},
		})
	}

	if unit.Height != 0 {
		node.Fields = append(node.Fields, &ast.FieldStmt{
			Key:   "height",
			Value: ast.Literal{Kind: ast.KindBareword, Str: formatFloat(unit.Height)},
		})
	}

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

// edgeStmtFromLink maps a model link to an edge statement with EXACT
// round-trip arrow forms (D-22): the glyph carries only non-default arrow
// semantics (`->` IS the omitted default, `<->` bidirectional, `--` none),
// so forward and reverse ride as the `arrow` option; LinksFrom entries emit
// the `<-` form with any non-default arrow as the option. Options follow
// the canonical option order: arrow, rank, color, style, labelPosition,
// length.
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

	if from {
		// `<-` alone is the omitted default; every explicit arrow value
		// rides the option verbatim.
		addOpt("arrow", string(link.Arrow))
	} else {
		addOpt("arrow", arrowOptionValue(link.Arrow))
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

// arrowGlyphFor maps the model arrow direction to the C4D glyph. Only
// non-default semantics ride the glyph: `->` is the default's exact twin
// (the omitted `arrow` key), `<->` and `--` carry bidirectional/none, and
// reverse pairs `->` with an `arrow: reverse` option — the glyph set has no
// reverse form of its own (D-05).
func arrowGlyphFor(a model.ArrowDirection) string {
	switch a { //nolint:exhaustive // default covers forward, reverse and the omitted ""
	case model.ArrowBidirectional:
		return "<->"
	case model.ArrowNone:
		return "--"
	default:
		// forward (with an `arrow: forward` option), reverse (option), and
		// the omitted default all ride `->`.
		return "->"
	}
}

// arrowOptionValue returns the explicit arrow option for an outgoing link —
// forward states the default explicitly, reverse cannot ride the glyph;
// the omitted default, bidirectional and none are glyph-exact and carry no
// option.
func arrowOptionValue(a model.ArrowDirection) string {
	switch a { //nolint:exhaustive // default covers "", bidirectional and none
	case model.ArrowForward, model.ArrowReverse:
		return string(a)
	default:
		return ""
	}
}

// templateDeclFromModel builds a template declaration. The body carries the
// template unit's fields (name included — `name` is a legal body field,
// unlike the header-only unit form), edges and subunits; subunits build
// BEFORE uses so parented uses can find their host block. The body's
// Type/External slots record the template root's type; EmitC4D renders them
// as the body's `type:` statement (the D-22 text form — the grammar admits
// it in template bodies since 35-06, closing the 35-04 deferred gap).
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
// delimiter or ENDS in a quote — a trailing quote merges with the closing
// delimiter into an ambiguous `""""` run the grammar cannot split (CR-05) —
// and both cases fall back to the escaped quoted form, which represents
// newlines as \n; bareword-safe values stay barewords; everything else is
// quoted.
func literalFor(s string) ast.Literal {
	switch {
	case strings.ContainsAny(s, "\n\r"):
		if !strings.Contains(s, `"""`) && !strings.HasSuffix(s, `"`) {
			return ast.Literal{Kind: ast.KindTriple, Str: s}
		}

		return ast.Literal{Kind: ast.KindQuoted, Str: s}
	case c4dBarewordSafe(s):
		return ast.Literal{Kind: ast.KindBareword, Str: s}
	default:
		return ast.Literal{Kind: ast.KindQuoted, Str: s}
	}
}

// CheckC4DRepresentable verifies that every Model value the C4D grammar
// expresses through a FIXED-CHARSET token survives emission and re-parse
// with its identity intact (F-01/CR-01/CR-03): unit ids at every nesting
// level, template names and use template names ([A-Za-z0-9_-]+ identifiers,
// D-07), link peer path segments (identifiers or ${param} tokens), and link
// technology/description ('|' is unrepresentable — the D-09 pipe shorthand
// splits labels on the first pipe after unquoting). It returns a
// *parser.ParseError naming the first offending value, nil when m converts
// losslessly. The convert write gate (cmd/c4drill) catches anything outside
// this pinned list; this check exists so the common corrupting classes get a
// LOUD error naming the value instead of a generic re-parse failure.
func CheckC4DRepresentable(m *parser.Model) error {
	if m == nil {
		return nil
	}

	if err := checkTopUnitsRepresentable(m); err != nil {
		return err
	}

	if err := checkTemplatesRepresentable(m); err != nil {
		return err
	}

	return checkUsesRepresentable(m.Instantiations)
}

// checkTopUnitsRepresentable checks every top-level unit subtree (ids, peers,
// link labels) in UnitOrder.
func checkTopUnitsRepresentable(m *parser.Model) error {
	for _, name := range m.UnitOrder {
		if unit, ok := m.Units[name]; ok {
			if err := checkC4DIdent("unit id "+strconv.Quote(name), name); err != nil {
				return err
			}

			if err := checkUnitRepresentable(name, unit); err != nil {
				return err
			}
		}
	}

	return nil
}

// checkTemplatesRepresentable checks template names, their bodies and their
// body-use template names, in sorted-name order (deterministic errors).
func checkTemplatesRepresentable(m *parser.Model) error {
	for _, name := range sortedTemplateNames(m.Templates) {
		tmpl := m.Templates[name]
		if tmpl == nil || tmpl.Unit == nil {
			continue
		}

		if err := checkC4DIdent("template name "+strconv.Quote(name), name); err != nil {
			return err
		}

		if err := checkUnitRepresentable(name, tmpl.Unit); err != nil {
			return err
		}

		if err := checkUsesRepresentable(tmpl.Instantiations); err != nil {
			return err
		}
	}

	return nil
}

// checkUsesRepresentable checks the use-template names of a use list.
func checkUsesRepresentable(insts []parser.Instantiation) error {
	for _, inst := range insts {
		if err := checkC4DIdent(
			"use template name "+strconv.Quote(inst.Template), inst.Template); err != nil {
			return err
		}
	}

	return nil
}

// checkUnitRepresentable checks one unit subtree's edges (peer paths, pipe
// labels) and subunit ids. path is the unit's dotted path, used for error
// context only.
func checkUnitRepresentable(path string, unit *model.Unit) error {
	if unit == nil {
		return nil
	}

	if err := checkLinksRepresentable(path, unit.Links); err != nil {
		return err
	}

	if err := checkLinksRepresentable(path, unit.LinksFrom); err != nil {
		return err
	}

	return checkSubunitsRepresentable(path, unit)
}

// checkSubunitsRepresentable checks every declared subunit's id and subtree.
func checkSubunitsRepresentable(path string, unit *model.Unit) error {
	for _, subName := range unit.SubunitOrder {
		sub, ok := unit.Subunits[subName]
		if !ok {
			continue
		}

		if err := checkC4DIdent(
			"unit id "+strconv.Quote(subName)+" (subunit of "+strconv.Quote(path)+")", subName); err != nil {
			return err
		}

		if err := checkUnitRepresentable(joinDotted(path, subName), sub); err != nil {
			return err
		}
	}

	return nil
}

// checkLinksRepresentable checks every link of one list.
func checkLinksRepresentable(path string, links []model.Link) error {
	for _, link := range links {
		if err := checkLinkRepresentable(path, link); err != nil {
			return err
		}
	}

	return nil
}

// checkLinkRepresentable checks one link: the peer path's charset and the
// D-09 pipe-shorthand representability of the label halves.
func checkLinkRepresentable(path string, link model.Link) error {
	if !c4dPeerSafe(link.Peer) {
		return &parser.ParseError{Message: fmt.Sprintf(
			"link peer %q on unit %q is not representable in C4D: "+
				"peer path segments must match [A-Za-z0-9_-]+ or be ${param} tokens",
			link.Peer, path)}
	}

	for _, half := range []struct{ field, value string }{
		{"technology", link.Technology},
		{"description", link.Description},
	} {
		if strings.Contains(half.value, "|") {
			return &parser.ParseError{Message: fmt.Sprintf(
				"link %s %q on unit %q is not representable in C4D: "+
					"labels must not contain '|' (the pipe shorthand splits labels on the first pipe)",
				half.field, half.value, path)}
		}
	}

	return nil
}

// checkC4DIdent verifies one identifier-position value against the grammar's
// identifier charset (D-07). what names the value for the error message.
func checkC4DIdent(what, value string) error {
	if c4dIdentSafe(value) {
		return nil
	}

	return &parser.ParseError{Message: what +
		" is not representable in C4D: identifiers must match [A-Za-z0-9_-]+"}
}

// c4dIdentSafe reports whether s matches the grammar's identifier charset
// ([A-Za-z0-9_-]+, D-07) — the charset unit ids, template names and peer path
// segments must carry to re-parse.
func c4dIdentSafe(s string) bool {
	if s == "" {
		return false
	}

	for _, c := range s {
		switch {
		case c >= '0' && c <= '9', c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z':
		case c == '_' || c == '-':
		default:
			return false
		}
	}

	return true
}

// c4dPeerSafe reports whether peer re-parses as a PeerRef (D-07/D-13): every
// dotted segment matches the identifier charset or is a ${param} token
// (template bodies parametrize link peers, so tokens ride verbatim).
func c4dPeerSafe(peer string) bool {
	for _, seg := range strings.Split(peer, ".") {
		if c4dIdentSafe(seg) {
			continue
		}

		return strings.HasPrefix(seg, "${") && strings.HasSuffix(seg, "}") &&
			c4dIdentSafe(seg[2:len(seg)-1])
	}

	return true
}

// formatFloat renders a float64 unit field (width, height) in its shortest
// exact decimal form — 300 stays "300" (the bareword/TOML integer spelling),
// 300.5 keeps its fraction.
func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
