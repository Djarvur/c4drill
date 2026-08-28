package c4d

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
)

// EmitTOML renders m as TOML text in the fixed canonical field order (D-23):
// unit fields type, name, description, technology, reference, color, style,
// border, edges, width, height, expanded; link fields peer, arrow, rank,
// color, style, technology, description, labelPosition, length; properties
// fields name, description, color, style, border, edges, lineLength,
// expanded. Sections
// emit in the fixed order [properties], unit tables (UnitOrder/SubunitOrder,
// recursively), [template.<name>], [[use]], [[template.<name>.use]] and
// [[include]] — mirroring the fixture files, which are the de-facto style
// spec (skill/examples/06-templates.toml, skill/examples/08-include/entry.toml,
// testdata/valid.toml).
//
// Emission walks the order slices only, never map iteration, so the output is
// deterministic (T-35-04-01). The error return is reserved for future
// validation failures — today only a nil model errors.
func EmitTOML(m *parser.Model) (string, error) {
	if m == nil {
		return "", &parser.ParseError{Message: "cannot emit TOML: model is nil"}
	}

	var b strings.Builder

	emitPropertiesTOML(&b, &m.Properties)

	for _, name := range m.UnitOrder {
		if unit, ok := m.Units[name]; ok {
			emitUnitTOML(&b, name, unit, "")
		}
	}

	for _, name := range sortedTemplateNames(m.Templates) {
		emitTemplateTOML(&b, name, m.Templates[name])
	}

	for _, inst := range m.Instantiations {
		emitInstantiationTOML(&b, "use", inst, true)
	}

	for _, inc := range m.Includes {
		b.WriteString("[[include]]\n")
		fmt.Fprintf(&b, "path = %s\n", quoteTOML(inc.Path))

		if inc.Once {
			b.WriteString("once = true\n")
		}
	}

	return b.String(), nil
}

// emitPropertiesTOML writes the [properties] table when any field is set
// (a fully-zero properties section is omitted entirely).
func emitPropertiesTOML(b *strings.Builder, props *model.Properties) {
	var body strings.Builder

	if props.Name != "" {
		fmt.Fprintf(&body, "name = %s\n", quoteTOML(props.Name))
	}

	if props.Description != "" {
		fmt.Fprintf(&body, "description = %s\n", quoteTOML(props.Description))
	}

	if props.Color != "" {
		fmt.Fprintf(&body, "color = %s\n", quoteTOML(props.Color))
	}

	if props.Style != "" {
		fmt.Fprintf(&body, "style = %s\n", quoteTOML(props.Style))
	}

	if props.Border != "" {
		fmt.Fprintf(&body, "border = %s\n", quoteTOML(props.Border))
	}

	if props.Edges != "" {
		fmt.Fprintf(&body, "edges = %s\n", quoteTOML(props.Edges))
	}

	if props.LineLength != 0 {
		fmt.Fprintf(&body, "lineLength = %d\n", props.LineLength)
	}

	if len(props.Expanded) > 0 {
		fmt.Fprintf(&body, "expanded = %s\n", quoteTOMLArray(props.Expanded))
	}

	emitLegendPropertiesTOML(b, &body, props)

	if body.Len() == 0 {
		return
	}

	b.WriteString("[properties]\n")
	b.WriteString(body.String())
}

// emitLegendPropertiesTOML writes the legend surface: `legend = false` inside
// the scalar block (emitted ONLY when explicitly false — nil/true omitted, so
// LEG-01 default-on keeps source-level byte stability for models not using
// the feature) and one [[properties.legendLine]] array table per custom row
// after it. Array-table headers cannot live inside the [properties] body —
// they are separate top-level tables.
func emitLegendPropertiesTOML(b *strings.Builder, body *strings.Builder, props *model.Properties) {
	if props.Legend != nil && !*props.Legend {
		body.WriteString("legend = false\n")
	}

	if len(props.LegendLines) == 0 {
		return
	}

	if body.Len() > 0 {
		b.WriteString("[properties]\n")
		b.WriteString(body.String())
		body.Reset()
	}

	for _, line := range props.LegendLines {
		fmt.Fprintf(b, "[[properties.legendLine]]\n")
		fmt.Fprintf(b, "label = %s\n", quoteTOML(line.Label))
		fmt.Fprintf(b, "color = %s\n", quoteTOML(line.Color))

		if line.Style != "" {
			fmt.Fprintf(b, "style = %s\n", quoteTOML(line.Style))
		}
	}
}

// emitUnitTOML writes the [path] table for one unit, then its link/linkFrom
// array tables, then its subunit tables recursively. Links MUST precede the
// subunit tables: a TOML array-table header attaches to the most recent
// table, so [[path.link]] after [path.sub] would wrongly land on the subunit.
// Key segments render bare-or-quoted (CR-02/F-01) so ids the C4D side derives
// from display names ("My App") round-trip as ["My App"].
func emitUnitTOML(b *strings.Builder, name string, unit *model.Unit, prefix string) {
	if unit == nil {
		return
	}

	path := joinKeyPath(prefix, name)
	b.WriteString("[" + path + "]\n")

	emitUnitFieldsTOML(b, unit)
	emitLinksTOML(b, path, unit)

	for _, subName := range unit.SubunitOrder {
		if sub, ok := unit.Subunits[subName]; ok {
			emitUnitTOML(b, subName, sub, path)
		}
	}
}

// emitUnitFieldsTOML writes the canonical unit field set (D-23) of unit in
// order, omitting empty/zero values.
func emitUnitFieldsTOML(b *strings.Builder, unit *model.Unit) {
	if unit.Type != "" {
		fmt.Fprintf(b, "type = %s\n", quoteTOML(string(unit.Type)))
	}

	if unit.Name != "" {
		fmt.Fprintf(b, "name = %s\n", quoteTOML(unit.Name))
	}

	if unit.Description != "" {
		fmt.Fprintf(b, "description = %s\n", quoteTOML(unit.Description))
	}

	if unit.Technology != "" {
		fmt.Fprintf(b, "technology = %s\n", quoteTOML(unit.Technology))
	}

	if unit.Reference != "" {
		fmt.Fprintf(b, "reference = %s\n", quoteTOML(unit.Reference))
	}

	if unit.Color != "" {
		fmt.Fprintf(b, "color = %s\n", quoteTOML(unit.Color))
	}

	if unit.Style != "" {
		fmt.Fprintf(b, "style = %s\n", quoteTOML(unit.Style))
	}

	if unit.Border != "" {
		fmt.Fprintf(b, "border = %s\n", quoteTOML(unit.Border))
	}

	if unit.Edges != "" {
		fmt.Fprintf(b, "edges = %s\n", quoteTOML(unit.Edges))
	}

	if unit.Width != 0 {
		fmt.Fprintf(b, "width = %s\n", formatFloat(unit.Width))
	}

	if unit.Height != 0 {
		fmt.Fprintf(b, "height = %s\n", formatFloat(unit.Height))
	}

	if len(unit.Expanded) > 0 {
		fmt.Fprintf(b, "expanded = %s\n", quoteTOMLArray(unit.Expanded))
	}
}

// emitLinksTOML writes the unit's [[path.link]] and [[path.linkFrom]] array
// tables. Validator-synthesized mirror links are bookkeeping, never authored
// content — they are skipped (Link.Mirror is never serialized).
func emitLinksTOML(b *strings.Builder, path string, unit *model.Unit) {
	for _, link := range unit.Links {
		if !link.IsMirror() {
			emitLinkTOML(b, joinKeyPath(path, "link"), link)
		}
	}

	for _, link := range unit.LinksFrom {
		if !link.IsMirror() {
			emitLinkTOML(b, joinKeyPath(path, "linkFrom"), link)
		}
	}
}

// emitLinkTOML writes one [[header]] array table in the D-23 link field
// order: peer, arrow, rank, kind, color, style, technology, description,
// labelPosition, length.
func emitLinkTOML(b *strings.Builder, header string, link model.Link) {
	b.WriteString("[[" + header + "]]\n")

	if link.Peer != "" {
		fmt.Fprintf(b, "peer = %s\n", quoteTOML(link.Peer))
	}

	if link.Arrow != "" {
		fmt.Fprintf(b, "arrow = %s\n", quoteTOML(string(link.Arrow)))
	}

	if link.Rank != "" {
		fmt.Fprintf(b, "rank = %s\n", quoteTOML(string(link.Rank)))
	}

	if link.Kind != "" {
		fmt.Fprintf(b, "kind = %s\n", quoteTOML(string(link.Kind)))
	}

	if link.Color != "" {
		fmt.Fprintf(b, "color = %s\n", quoteTOML(link.Color))
	}

	if link.Style != "" {
		fmt.Fprintf(b, "style = %s\n", quoteTOML(link.Style))
	}

	if link.Technology != "" {
		fmt.Fprintf(b, "technology = %s\n", quoteTOML(link.Technology))
	}

	if link.Description != "" {
		fmt.Fprintf(b, "description = %s\n", quoteTOML(link.Description))
	}

	if link.LabelPosition != "" {
		fmt.Fprintf(b, "labelPosition = %s\n", quoteTOML(string(link.LabelPosition)))
	}

	if link.Length != 0 {
		fmt.Fprintf(b, "length = %d\n", link.Length)
	}
}

// emitTemplateTOML writes one [template.<name>] table: declared params, the
// template unit fields, its links, its root-level body uses, then its
// subunit tables recursively (uses nested under a subunit path emit inside
// that recursion, mirroring the [[template.<name>.<path>.use]] authoring
// form, D-17).
func emitTemplateTOML(b *strings.Builder, name string, tmpl *parser.TemplateDef) {
	if tmpl == nil || tmpl.Unit == nil {
		return
	}

	header := joinKeyPath("template", name)
	b.WriteString("[" + header + "]\n")

	if len(tmpl.Params) > 0 {
		fmt.Fprintf(b, "params = %s\n", quoteTOMLArray(tmpl.Params))
	}

	emitUnitFieldsTOML(b, tmpl.Unit)
	emitLinksTOML(b, header, tmpl.Unit)

	for _, inst := range tmpl.Instantiations {
		if inst.Parent == "" {
			emitInstantiationTOML(b, joinKeyPath(header, "use"), inst, false)
		}
	}

	emitTemplateSubunitsTOML(b, name, nil, tmpl.Unit, tmpl.Instantiations)
}

// emitTemplateSubunitsTOML walks the template's subunit subtree, writing each
// [template.<name>.<path>] table and the [[template.<name>.<path>.use]]
// entries whose root-relative parent matches that path. relSegs carries the
// RAW subunit path segments: Parent matching compares raw names, while table
// headers render each segment bare-or-quoted (CR-02/F-01).
func emitTemplateSubunitsTOML(
	b *strings.Builder,
	tmplName string,
	relSegs []string,
	unit *model.Unit,
	insts []parser.Instantiation,
) {
	for _, subName := range unit.SubunitOrder {
		sub, ok := unit.Subunits[subName]
		if !ok {
			continue
		}

		subSegs := append(slices.Clone(relSegs), subName)
		subRel := strings.Join(subSegs, ".")
		header := keyPathTOML(append([]string{"template", tmplName}, subSegs...))
		b.WriteString("[" + header + "]\n")

		emitUnitFieldsTOML(b, sub)
		emitLinksTOML(b, header, sub)

		for _, inst := range insts {
			if inst.Parent == subRel {
				emitInstantiationTOML(b, joinKeyPath(header, "use"), inst, false)
			}
		}

		emitTemplateSubunitsTOML(b, tmplName, subSegs, sub, insts)
	}
}

// emitInstantiationTOML writes one [[header]] array table: template, the
// optional parent key (top-level [[use]] only — a template-body use's parent
// is encoded in the header path itself), then the supplied params sorted by
// key (the Model stores params as a map, so sorted keys are the only
// deterministic order; D-23 pins field order, not arg order). Param keys
// render bare-or-quoted like every other TOML key (CR-02/F-01).
func emitInstantiationTOML(b *strings.Builder, header string, inst parser.Instantiation, withParent bool) {
	b.WriteString("[[" + header + "]]\n")
	fmt.Fprintf(b, "template = %s\n", quoteTOML(inst.Template))

	if withParent && inst.Parent != "" {
		fmt.Fprintf(b, "parent = %s\n", quoteTOML(inst.Parent))
	}

	for _, key := range sortedParamKeys(inst.Params) {
		fmt.Fprintf(b, "%s = %s\n", tomlKeySegment(key), quoteTOML(inst.Params[key]))
	}
}

// quoteTOML renders s as a TOML basic string. Control characters (including
// newlines — pinned to the escaped single-line form under D-06) always emit
// as escapes, quotes and backslashes are escaped, and every other byte
// (UTF-8 sequences included) passes through verbatim, so no value can inject
// raw structure into the output (T-35-04-02).
func quoteTOML(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 2)
	b.WriteByte('"')

	for i := range len(s) {
		c := s[i]

		switch c {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if c < 0x20 || c == 0x7f {
				const hexDigits = "0123456789ABCDEF"

				b.WriteString(`\u00`)
				b.WriteByte(hexDigits[c>>4])
				b.WriteByte(hexDigits[c&0x0f])
			} else {
				b.WriteByte(c)
			}
		}
	}

	b.WriteByte('"')

	return b.String()
}

// quoteTOMLArray renders items as a TOML array of basic strings.
func quoteTOMLArray(items []string) string {
	parts := make([]string, len(items))
	for i, item := range items {
		parts[i] = quoteTOML(item)
	}

	return "[" + strings.Join(parts, ", ") + "]"
}

// sortedTemplateNames returns the template names in sorted order — the only
// deterministic order available for a map (T-35-04-01).
func sortedTemplateNames(templates map[string]*parser.TemplateDef) []string {
	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}

	slices.Sort(names)

	return names
}

// sortedParamKeys returns the param map keys in sorted order (deterministic
// emission of [[use]] params).
func sortedParamKeys(params map[string]string) []string {
	keys := make([]string, 0, len(params))
	for key := range params {
		keys = append(keys, key)
	}

	slices.Sort(keys)

	return keys
}

// joinDotted joins two dotted path segments, treating "" as identity. Both
// arguments must ALREADY be rendered (or raw-but-bare) — it never quotes; use
// joinKeyPath when the second segment is a raw name that may need quoting.
func joinDotted(base, rel string) string {
	if base == "" {
		return rel
	}

	return base + "." + rel
}

// joinKeyPath joins a rendered key path with one more RAW segment, rendering
// the new segment bare-or-quoted (CR-02/F-01): table headers stay parseable
// when a name is not a bare key.
func joinKeyPath(base, seg string) string {
	if base == "" {
		return tomlKeySegment(seg)
	}

	return base + "." + tomlKeySegment(seg)
}

// keyPathTOML renders raw key segments as a dotted table-header path, quoting
// every segment that is not a bare key: ["My App", sub] -> ["My App".sub].
func keyPathTOML(segs []string) string {
	parts := make([]string, len(segs))
	for i, seg := range segs {
		parts[i] = tomlKeySegment(seg)
	}

	return strings.Join(parts, ".")
}

// tomlKeySegment renders one key segment: TOML bare keys ([A-Za-z0-9_-]+)
// pass through verbatim, everything else becomes a quoted basic string the
// parser reads back as the identical segment (["My App"], CR-02/F-01).
func tomlKeySegment(seg string) string {
	if tomlBareKeySafe(seg) {
		return seg
	}

	return quoteTOML(seg)
}

// tomlBareKeySafe reports whether s is a TOML bare key ([A-Za-z0-9_-]+).
func tomlBareKeySafe(s string) bool {
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
