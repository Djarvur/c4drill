package c4d

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Djarvur/c4drill/internal/c4d/ast"
	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
	"github.com/Djarvur/c4drill/internal/validator"
)

// ToModel converts a typed C4D AST document into a *parser.Model — the same
// Model the TOML front-end produces for an equivalent document (D-02/D-21
// parity hub). The conversion applies the IDENTICAL post-parse hooks
// parseUnitWithOrder applies, in the same order:
//
//  1. empty Type -> parser.DefaultTypeForParent(parentType)
//  2. Type = parser.InferGenericType(Type, parentType)
//  3. empty Name -> model.Humanize(id)
//
// Semantic checks that fail closed with a *parser.ParseError carrying the
// AST statement's line: duplicate unit paths, duplicate edges per
// (unit, peer) pair in one list (D-11), duplicate field/property keys,
// types without an external variant (D-04), non-numeric length/lineLength,
// and positional `use` args the document cannot resolve.
//
// Peer strings ride verbatim (bare or dotted) — relative-peer resolution is
// internal/peer.Resolve's pass, never this one (D-10).
func ToModel(doc *ast.Document) (*parser.Model, error) {
	if doc == nil {
		return nil, &parser.ParseError{Message: "cannot convert a nil document"}
	}

	// UnitOrder starts as an EMPTY (non-nil) slice — parser.Parse always
	// returns make([]string, 0) from captureDefinitionOrder, and the parity
	// contract compares models with require.Equal (nil != empty).
	m := &parser.Model{
		UnitOrder: make([]string, 0),
		Units:     make(map[string]*model.Unit),
	}

	if err := applyProperties(m, doc.Properties); err != nil {
		return nil, err
	}

	// Templates build FIRST: positional use args pair against the declared
	// params at conversion time, so the registry must be complete before any
	// use statement converts (forward template references work — same as
	// TOML, where extractUses runs after extractTemplates).
	templates, pending, err := buildTemplates(doc)
	if err != nil {
		return nil, err
	}

	// Units (with their nested use statements collected for the second
	// phase). The AST presents top-level statements in canonical sections —
	// units before uses — so unit-nested instantiations surface in
	// m.Instantiations before top-level ones.
	if err := buildUnits(m, doc.Units, &pending); err != nil {
		return nil, err
	}

	// Top-level use statements (Parent "" — the Phase 31 semantics).
	for _, us := range doc.UseStmts {
		if us == nil {
			continue
		}

		pending = append(pending, pendingUse{stmt: us, parent: ""})
	}

	// Second phase: convert every collected use statement now that the
	// template registry is complete. Template-body uses route into their
	// owning TemplateDef.Instantiations (D-17); document-level uses into
	// m.Instantiations.
	if err := flushUses(m, pending, templates); err != nil {
		return nil, err
	}

	// Includes (D-14): path text as written, once flag verbatim — resolution
	// relative to the including file is include.Resolve's job.
	for _, inc := range doc.Includes {
		if inc == nil {
			continue
		}

		m.Includes = append(m.Includes, parser.IncludeDirective{Path: inc.Path, Once: inc.Once})
	}

	// Templates stay nil for template-less documents — the TOML front-end's
	// nil-ness parity (require.Equal twin tests compare the whole field).
	if len(templates) > 0 {
		m.Templates = templates
	}

	return m, nil
}

// pendingUse is a use statement collected during the unit/template walk,
// awaiting conversion once the template registry is complete. template is
// non-empty for template-body uses (D-17) — they route into the owning
// TemplateDef.Instantiations instead of m.Instantiations; parent is then
// relative to the template's unit root.
type pendingUse struct {
	stmt     *ast.UseStmt
	parent   string
	template string
}

// externalizableTypes are the base type keywords with an *External variant
// (D-04): the C1-level person/system/db/queue. Every other keyword — box,
// the container*/component* level types, and the *External keywords
// themselves — has no external form and hard-errors on the modifier.
//
//nolint:gochecknoglobals // Pinned closed vocabulary, immutable after init
var externalizableTypes = []string{"person", "system", "db", "queue"}

// isExternalizable reports whether the base type keyword has an *External
// variant.
func isExternalizable(base string) bool {
	for _, t := range externalizableTypes {
		if t == base {
			return true
		}
	}

	return false
}

// buildTemplates converts every template declaration (D-13) into the
// registry and collects the bodies' use statements as pending conversions.
// Duplicate template names are a hard error (the TOML front-end rejects the
// duplicate [template.x] table twin).
func buildTemplates(doc *ast.Document) (map[string]*parser.TemplateDef, []pendingUse, error) {
	templates := make(map[string]*parser.TemplateDef, len(doc.Templates))

	var pending []pendingUse

	for _, decl := range doc.Templates {
		if decl == nil {
			continue
		}

		if _, dup := templates[decl.Name]; dup {
			return nil, nil, &parser.ParseError{
				Message: fmt.Sprintf("duplicate template %q", decl.Name),
				Line:    decl.Pos,
			}
		}

		tmpl, bodyUses, err := templateDefFromAST(decl)
		if err != nil {
			return nil, nil, err
		}

		templates[decl.Name] = tmpl

		pending = append(pending, bodyUses...)
	}

	return templates, pending, nil
}

// buildUnits converts the document's top-level units in statement order,
// collecting their nested use statements into pending.
func buildUnits(m *parser.Model, nodes []*ast.UnitNode, pending *[]pendingUse) error {
	seen := make(map[string]bool, len(nodes))

	for _, node := range nodes {
		if node == nil {
			continue
		}

		id, err := unitKey(node)
		if err != nil {
			return err
		}

		if seen[id] {
			return &parser.ParseError{
				Message: fmt.Sprintf("duplicate unit path %q", id),
				Line:    node.Pos,
			}
		}

		seen[id] = true

		unit, uses, err := buildUnit(id, node, "", id)
		if err != nil {
			return err
		}

		m.Units[id] = unit
		m.UnitOrder = append(m.UnitOrder, id)

		*pending = append(*pending, uses...)
	}

	return nil
}

// flushUses converts every pending use statement. Template-body uses route
// into their owning TemplateDef.Instantiations (D-17); document-level uses
// (top-level + unit-nested) into m.Instantiations.
func flushUses(
	m *parser.Model,
	pending []pendingUse,
	templates map[string]*parser.TemplateDef,
) error {
	for _, p := range pending {
		inst, err := instantiationFromUse(p.stmt, p.parent, templates)
		if err != nil {
			return err
		}

		if p.template != "" {
			templates[p.template].Instantiations = append(templates[p.template].Instantiations, inst)

			continue
		}

		m.Instantiations = append(m.Instantiations, inst)
	}

	return nil
}

// unitKey derives the unit's map key from its header. The id-led form uses
// the id; a type-led header without id uses the quoted display name (the
// TOML twin is the quoted table ["Name"]); a type-led header without either
// uses the type keyword. The derived key feeds the Humanize hook exactly
// like a TOML table name would.
func unitKey(node *ast.UnitNode) (string, error) {
	switch {
	case node.ID != "":
		return node.ID, nil
	case node.Name != "":
		return node.Name, nil
	case node.Type != "":
		return node.Type, nil
	default:
		return "", &parser.ParseError{
			Message: "unit header carries no identifier, name, or type",
			Line:    node.Pos,
		}
	}
}

// applyProperties maps the `properties { }` block (D-12) onto
// model.Properties. Duplicate keys and malformed values (non-numeric
// lineLength, non-list expanded) are hard errors — the TOML front-end's
// unmarshal rejects the same shapes.
func applyProperties(m *parser.Model, block *ast.PropertiesBlock) error {
	if block == nil {
		return nil
	}

	seen := make(map[string]bool, len(block.Fields))

	for _, f := range block.Fields {
		if f == nil {
			continue
		}

		if seen[f.Key] {
			return &parser.ParseError{
				Message: fmt.Sprintf("duplicate properties key %q", f.Key),
				Line:    f.Pos,
			}
		}

		seen[f.Key] = true

		if err := applyPropertyField(&m.Properties, f); err != nil {
			return err
		}
	}

	return nil
}

// applyPropertyField maps one properties field onto the target struct.
func applyPropertyField(props *model.Properties, f *ast.FieldStmt) error {
	switch f.Key {
	case "lineLength":
		n, err := intValue(f)
		if err != nil {
			return err
		}

		props.LineLength = n
	case "expanded":
		list, err := listValue(f)
		if err != nil {
			return err
		}

		props.Expanded = list
	case "legend":
		// C4D has no boolean literals: legend rides a bareword true/false.
		switch f.Value.Str {
		case "true":
			truly := true

			props.Legend = &truly
		case "false":
			falsy := false

			props.Legend = &falsy
		default:
			return &parser.ParseError{
				Message: fmt.Sprintf("property %q must be true or false, got %q", f.Key, f.Value.Str),
				Line:    f.Pos,
			}
		}
	case "legendLine":
		list, err := listValue(f)
		if err != nil {
			return err
		}

		lines := make([]model.LegendLine, 0, len(list))

		for _, raw := range list {
			parts := strings.SplitN(raw, "|", 3)
			if len(parts) < 2 {
				return &parser.ParseError{
					Message: fmt.Sprintf(
						"property %q entries need \"label|color\" or \"label|color|style\", got %q",
						f.Key, raw),
					Line: f.Pos,
				}
			}

			line := model.LegendLine{Label: parts[0], Color: parts[1]}
			if len(parts) == 3 {
				line.Style = parts[2]
			}

			lines = append(lines, line)
		}

		props.LegendLines = lines
	default:
		if dst := propertyStringField(props, f.Key); dst != nil {
			if f.Value.Kind == ast.KindList {
				return listRejected(f, "properties")
			}

			*dst = f.Value.Str
		}
	}

	return nil
}

// propertyStringField returns the string field pointer for a scalar
// properties key, nil for keys that are not scalar properties (unknown keys
// are rejected by the grammar, D-19).
func propertyStringField(props *model.Properties, key string) *string {
	switch key {
	case "name":
		return &props.Name
	case "description":
		return &props.Description
	case "color":
		return &props.Color
	case "style":
		return &props.Style
	case "border":
		return &props.Border
	case "edges":
		return &props.Edges
	default:
		return nil
	}
}

// buildUnit converts one UnitNode into a *model.Unit, mirroring
// parseUnitWithOrder's hook order exactly: type inference (default ->
// generic promotion), external mapping (D-04), humanized name, fields,
// edges (D-11 duplicate check), subunits in statement order, and nested use
// collection (D-16). path is the unit's dotted path — error context and the
// parent path of its use statements.
func buildUnit(
	id string,
	node *ast.UnitNode,
	parentType model.UnitType,
	path string,
) (*model.Unit, []pendingUse, error) {
	unitType, err := resolveUnitType(node, parentType)
	if err != nil {
		return nil, nil, err
	}

	unit := &model.Unit{Type: unitType, Name: node.Name}

	if err := applyUnitFields(unit, node, path); err != nil {
		return nil, nil, err
	}

	// Humanize hook (ERGO-03/05): only when neither the header nor a body
	// `name` field supplied a display name — the same input shape as the
	// TOML hook (the identifier segment).
	if unit.Name == "" {
		unit.Name = model.Humanize(id)
	}

	if err := applyEdges(unit, node.Edges, path); err != nil {
		return nil, nil, err
	}

	uses, err := buildSubunits(unit, node.Subunits, path)
	if err != nil {
		return nil, nil, err
	}

	// D-16: a use inside a unit block attaches under the enclosing unit.
	for _, us := range node.UseStmts {
		if us == nil {
			continue
		}

		uses = append(uses, pendingUse{stmt: us, parent: path})
	}

	return unit, uses, nil
}

// applyUnitFields maps the body field statements in source order onto the
// unit. Duplicate keys are a hard error (the TOML front-end rejects the
// duplicate table key twin).
func applyUnitFields(unit *model.Unit, node *ast.UnitNode, path string) error {
	seen := make(map[string]bool, len(node.Fields))

	for _, f := range node.Fields {
		if f == nil {
			continue
		}

		if seen[f.Key] {
			return &parser.ParseError{
				Message: fmt.Sprintf("duplicate field %q in unit %q", f.Key, path),
				Line:    f.Pos,
			}
		}

		seen[f.Key] = true

		if err := applyUnitField(unit, f, path); err != nil {
			return err
		}
	}

	return nil
}

// applyUnitField maps one body field onto the unit. Scalar fields reject
// list literals; `expanded` rejects scalars; `width`/`height` reject
// non-numeric values — the same shapes the TOML unmarshal rejects.
func applyUnitField(unit *model.Unit, f *ast.FieldStmt, path string) error {
	if f.Key == "expanded" {
		list, err := listValue(f)
		if err != nil {
			return err
		}

		unit.Expanded = list

		return nil
	}

	if f.Key == "width" || f.Key == "height" {
		if f.Value.Kind == ast.KindList {
			return listRejected(f, "unit "+path)
		}

		n, err := floatValue(f)
		if err != nil {
			return err
		}

		if f.Key == "width" {
			unit.Width = n
		} else {
			unit.Height = n
		}

		return nil
	}

	dst := unitStringField(unit, f.Key)
	if dst == nil {
		return nil // unknown keys are rejected by the grammar (D-19)
	}

	if f.Value.Kind == ast.KindList {
		return listRejected(f, "unit "+path)
	}

	*dst = f.Value.Str

	return nil
}

// unitStringField returns the string field pointer for a scalar unit field
// key, nil for keys that are not scalar unit fields.
func unitStringField(unit *model.Unit, key string) *string {
	switch key {
	case "name":
		return &unit.Name
	case "description":
		return &unit.Description
	case "technology":
		return &unit.Technology
	case "reference":
		return &unit.Reference
	case "color":
		return &unit.Color
	case "style":
		return &unit.Style
	case "border":
		return &unit.Border
	case "edges":
		return &unit.Edges
	default:
		return nil
	}
}

// buildSubunits converts the nested units in statement order, recording
// SubunitOrder and collecting their use statements. Duplicate subunit keys
// within one block are a hard error (TOML rejects the duplicate table twin).
func buildSubunits(unit *model.Unit, nodes []*ast.UnitNode, path string) ([]pendingUse, error) {
	var uses []pendingUse

	seen := make(map[string]bool, len(nodes))

	for _, sub := range nodes {
		if sub == nil {
			continue
		}

		subID, err := unitKey(sub)
		if err != nil {
			return nil, err
		}

		if seen[subID] {
			return nil, &parser.ParseError{
				Message: fmt.Sprintf("duplicate unit path %q", joinDotted(path, subID)),
				Line:    sub.Pos,
			}
		}

		seen[subID] = true

		child, childUses, err := buildUnit(subID, sub, unit.Type, joinDotted(path, subID))
		if err != nil {
			return nil, err
		}

		if unit.Subunits == nil {
			unit.Subunits = make(map[string]*model.Unit)
		}

		unit.Subunits[subID] = child
		unit.SubunitOrder = append(unit.SubunitOrder, subID)
		uses = append(uses, childUses...)
	}

	return uses, nil
}

// resolveUnitType applies the type hooks in parseUnitWithOrder's order:
// omitted type -> DefaultTypeForParent; `external` modifier -> the
// *External variant (validated against the type vocabulary, D-04); then
// InferGenericType's level promotion.
func resolveUnitType(node *ast.UnitNode, parentType model.UnitType) (model.UnitType, error) {
	typeStr := node.Type
	if typeStr == "" {
		typeStr = string(parser.DefaultTypeForParent(parentType))
	}

	if node.External {
		var err error

		if typeStr, err = externalVariant(typeStr, node.Pos); err != nil {
			return "", err
		}
	}

	return parser.InferGenericType(model.UnitType(typeStr), parentType), nil
}

// externalVariant maps a base type keyword + the external modifier to the
// *External variant, hard-erroring on keywords without one (D-04) and on
// the redundant modifier after an already-external keyword.
func externalVariant(typeStr string, pos int) (string, error) {
	if strings.HasSuffix(typeStr, "External") {
		return "", &parser.ParseError{
			Message: fmt.Sprintf("type %q is already external — drop the external modifier", typeStr),
			Line:    pos,
		}
	}

	if !isExternalizable(typeStr) {
		return "", &parser.ParseError{
			Message: fmt.Sprintf("type %q has no external variant%s",
				typeStr, validator.FormatSuggestion(typeStr, externalizableTypes)),
			Line: pos,
		}
	}

	return typeStr + "External", nil
}

// applyEdges maps edge statements onto unit.Links / unit.LinksFrom (D-08):
// `->`, `<->` and `--` are outgoing (Links) with the glyph's ArrowDirection;
// `<-` is incoming (LinksFrom, arrowhead at the owner — the exact shape the
// validator's mirror of `peer -> owner` carries). Duplicate (unit, peer)
// pairs within one list are a hard error (D-11).
func applyEdges(unit *model.Unit, edges []*ast.EdgeStmt, path string) error {
	seenLinks := make(map[string]bool, len(edges))
	seenFrom := make(map[string]bool, len(edges))

	for _, e := range edges {
		if e == nil {
			continue
		}

		link, incoming, err := linkFromEdge(e)
		if err != nil {
			return err
		}

		seen, listName := seenLinks, "link"

		if incoming {
			seen, listName = seenFrom, "linkFrom"
		}

		if seen[link.Peer] {
			return &parser.ParseError{
				Message: fmt.Sprintf(
					"duplicate edge: unit %q already has a %s statement for peer %q",
					path, listName, link.Peer),
				Line: e.Pos,
			}
		}

		seen[link.Peer] = true

		if incoming {
			unit.LinksFrom = append(unit.LinksFrom, link)
		} else {
			unit.Links = append(unit.Links, link)
		}
	}

	return nil
}

// linkFromEdge builds a model.Link from an edge statement. incoming reports
// the `<-` form (D-08).
func linkFromEdge(e *ast.EdgeStmt) (model.Link, bool, error) {
	link := model.Link{
		Peer:        e.Peer,
		Technology:  e.Technology,
		Description: e.Description,
	}

	incoming, err := arrowFromGlyph(e, &link)
	if err != nil {
		return model.Link{}, false, err
	}

	if err := applyEdgeOptions(e, &link); err != nil {
		return model.Link{}, false, err
	}

	return link, incoming, nil
}

// arrowFromGlyph maps the statement's glyph onto the link's default arrow
// direction, reporting whether the `<-` form targets LinksFrom.
func arrowFromGlyph(e *ast.EdgeStmt, link *model.Link) (bool, error) {
	switch e.ArrowGlyph {
	case "->":
		// The DEFAULT arrow: the same Link value the TOML front-end produces
		// for an omitted `arrow` key — "" (D-22 explicit-defaults rule at
		// Model level; the render layer distinguishes "" from "forward" via
		// the dir attribute, so exact parity requires the omitted form).
		// `arrow: forward` in the option block states the default explicitly.
		link.Arrow = ""
	case "<->":
		link.Arrow = model.ArrowBidirectional
	case "--":
		link.Arrow = model.ArrowNone
	case "<-":
		// The glyph's arrowhead sits at the owning unit, i.e. at the TARGET
		// of the peer->owner edge — the default (omitted) arrow in LinksFrom
		// orientation, the same value the TOML twin's bare [[...linkFrom]]
		// entry carries.
		link.Arrow = ""

		return true, nil
	default:
		return false, &parser.ParseError{
			Message: fmt.Sprintf("unknown arrow glyph %q", e.ArrowGlyph),
			Line:    e.Pos,
		}
	}

	return false, nil
}

// applyEdgeOptions maps the trailing option block onto the link. The
// `arrow` option overrides the glyph default; enum-ish values ride verbatim
// (TOML accepts arbitrary strings there — the render layer applies
// defaults). `length` must parse as an integer, like the TOML field.
func applyEdgeOptions(e *ast.EdgeStmt, link *model.Link) error {
	for _, opt := range e.Options {
		if opt == nil {
			continue
		}

		if opt.Value.Kind == ast.KindList {
			return &parser.ParseError{
				Message: fmt.Sprintf("edge option %q does not accept a list value", opt.Key),
				Line:    opt.Pos,
			}
		}

		if err := applyEdgeOption(opt, link); err != nil {
			return err
		}
	}

	return nil
}

// applyEdgeOption maps a single edge option onto the link.
func applyEdgeOption(opt *ast.FieldStmt, link *model.Link) error {
	v := opt.Value.Str

	switch opt.Key {
	case "arrow":
		link.Arrow = model.ArrowDirection(v)
	case "rank":
		link.Rank = model.RankDirection(v)
	case "kind":
		link.Kind = model.LinkKind(v)
	case "labelPosition":
		link.LabelPosition = model.LabelPosition(v)
	case "color":
		link.Color = v
	case "style":
		link.Style = v
	case "length":
		n, err := strconv.Atoi(v)
		if err != nil {
			return &parser.ParseError{
				Message: fmt.Sprintf("edge option %q must be an integer, got %q", opt.Key, v),
				Line:    opt.Pos,
			}
		}

		link.Length = n
	}

	return nil
}

// templateDefFromAST converts a template declaration (D-13) into a
// TemplateDef: declared params, the body as the template root unit (same
// inference hooks as a top-level unit — the TOML front-end parses the root
// with parentType ""), and the body's use statements collected with
// root-relative parents (D-17). Body.Type/Body.External are honored when
// set — FromModel records a non-default template root type there (the
// grammar itself has no template-root-type syntax; this closes the 35-04
// Model->AST->Model gap).
func templateDefFromAST(decl *ast.TemplateDecl) (*parser.TemplateDef, []pendingUse, error) {
	unit, uses, err := buildUnit(decl.Name, decl.Body, "", "")
	if err != nil {
		return nil, nil, err
	}

	// Tag the collected uses with the owning template so the second phase
	// routes them into TemplateDef.Instantiations.
	for i := range uses {
		uses[i].template = decl.Name
	}

	return &parser.TemplateDef{Params: decl.Params, Unit: unit}, uses, nil
}

// instantiationFromUse converts a use statement into an Instantiation.
// Named args map directly; positional args pair with the template's
// declared params in order — which requires the template to be declared in
// the same document (pairing is a conversion-time decision; named args work
// for templates arriving via includes).
func instantiationFromUse(
	us *ast.UseStmt,
	parent string,
	templates map[string]*parser.TemplateDef,
) (parser.Instantiation, error) {
	inst := parser.Instantiation{
		Template: us.Template,
		Parent:   parent,
		Params:   make(map[string]string, len(us.Args)),
	}

	positional := 0

	for _, arg := range us.Args {
		if arg.Value.Kind == ast.KindList {
			return parser.Instantiation{}, &parser.ParseError{
				Message: "use arguments do not accept list values",
				Line:    us.Pos,
			}
		}

		if arg.Name != "" {
			if err := setNamedArg(&inst, arg.Name, arg.Value.Str, us.Pos); err != nil {
				return parser.Instantiation{}, err
			}

			continue
		}

		var err error

		if positional, err = setPositionalArg(&inst, templates, positional, arg.Value.Str, us.Pos); err != nil {
			return parser.Instantiation{}, err
		}
	}

	return inst, nil
}

// setNamedArg records a named use argument, rejecting duplicate keys.
func setNamedArg(inst *parser.Instantiation, name, value string, pos int) error {
	if _, dup := inst.Params[name]; dup {
		return &parser.ParseError{
			Message: fmt.Sprintf("duplicate use argument %q", name),
			Line:    pos,
		}
	}

	inst.Params[name] = value

	return nil
}

// setPositionalArg pairs the next positional argument with the template's
// declared params, returning the advanced cursor.
func setPositionalArg(
	inst *parser.Instantiation,
	templates map[string]*parser.TemplateDef,
	positional int,
	value string,
	pos int,
) (int, error) {
	tmpl, ok := templates[inst.Template]
	if !ok {
		return 0, &parser.ParseError{
			Message: fmt.Sprintf(
				"positional use arguments require template %q to be declared in the same file — use named arguments",
				inst.Template),
			Line: pos,
		}
	}

	if positional >= len(tmpl.Params) {
		return 0, &parser.ParseError{
			Message: fmt.Sprintf(
				"too many positional use arguments: template %q declares %d params",
				inst.Template, len(tmpl.Params)),
			Line: pos,
		}
	}

	key := tmpl.Params[positional]

	if err := setNamedArg(inst, key, value, pos); err != nil {
		return 0, err
	}

	return positional + 1, nil
}

// intValue parses a numeric field value (lineLength, length).
func intValue(f *ast.FieldStmt) (int, error) {
	n, err := strconv.Atoi(f.Value.Str)
	if err != nil {
		return 0, &parser.ParseError{
			Message: fmt.Sprintf("field %q must be an integer, got %q", f.Key, f.Value.Str),
			Line:    f.Pos,
		}
	}

	return n, nil
}

// floatValue parses a numeric field value (width, height) — the Model carries
// them as float64, so integer and decimal spellings both parse.
func floatValue(f *ast.FieldStmt) (float64, error) {
	n, err := strconv.ParseFloat(f.Value.Str, 64)
	if err != nil {
		return 0, &parser.ParseError{
			Message: fmt.Sprintf("field %q must be a number, got %q", f.Key, f.Value.Str),
			Line:    f.Pos,
		}
	}

	return n, nil
}

// listValue extracts the list items of a field value, rejecting non-list
// literals.
func listValue(f *ast.FieldStmt) ([]string, error) {
	if f.Value.Kind != ast.KindList {
		return nil, &parser.ParseError{
			Message: fmt.Sprintf("field %q requires a list value", f.Key),
			Line:    f.Pos,
		}
	}

	return f.Value.List, nil
}

// listRejected renders the scalar-expected error for a list literal.
func listRejected(f *ast.FieldStmt, where string) error {
	return &parser.ParseError{
		Message: fmt.Sprintf("field %q in %s does not accept a list value", f.Key, where),
		Line:    f.Pos,
	}
}
