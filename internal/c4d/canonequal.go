package c4d

import (
	"reflect"
	"slices"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
)

// CanonicalEqual reports whether a and b are equal modulo the D-22
// explicit-defaults list: an omitted link default (arrow "", rank "",
// labelPosition "") compares equal to its explicit spelling ("forward",
// "forward", "middle") because a round-trip through either front-end may
// state or omit exactly those. Everything else — UnitOrder, every unit field
// including width/height, templates, instantiations, includes — must match
// EXACTLY, so the comparison hides no legitimate difference (F-01 caution:
// dropping width/height or reshuffling a label makes models unequal).
//
// This is the parity contract's canonicalModel comparison (parity_test.go)
// exported as a predicate for the convert write gate (cmd/c4drill): a
// converted twin must re-parse to a canonically-equal model before it is
// written to disk.
func CanonicalEqual(a, b *parser.Model) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}

	return reflect.DeepEqual(canonicalModelForCompare(a), canonicalModelForCompare(b))
}

// canonicalModelForCompare deep-copies m with the D-22 explicit-default list
// filled in on every link, so reflect.DeepEqual compares fairly. The clone
// never mutates the source (the fill runs on Clone'd units only).
func canonicalModelForCompare(m *parser.Model) *parser.Model {
	clone := &parser.Model{
		Properties:     m.Properties,
		UnitOrder:      slices.Clone(m.UnitOrder),
		Units:          make(map[string]*model.Unit, len(m.Units)),
		Templates:      make(map[string]*parser.TemplateDef, len(m.Templates)),
		Instantiations: slices.Clone(m.Instantiations),
		Includes:       slices.Clone(m.Includes),
	}

	for name, unit := range m.Units {
		clone.Units[name] = canonicalUnitForCompare(unit)
	}

	for name, tmpl := range m.Templates {
		if tmpl == nil {
			continue
		}

		t := *tmpl
		t.Params = slices.Clone(tmpl.Params)
		t.Unit = canonicalUnitForCompare(tmpl.Unit)
		t.Instantiations = slices.Clone(tmpl.Instantiations)
		clone.Templates[name] = &t
	}

	return clone
}

// canonicalUnitForCompare clones a unit and fills link defaults recursively.
func canonicalUnitForCompare(unit *model.Unit) *model.Unit {
	clone := unit.Clone()
	if clone == nil {
		return nil
	}

	fillCanonicalLinkDefaults(clone)

	return clone
}

// fillCanonicalLinkDefaults applies the D-22 default list to every link in
// the subtree, in place, on an already-cloned unit.
func fillCanonicalLinkDefaults(unit *model.Unit) {
	for i := range unit.Links {
		canonicalLinkDefaults(&unit.Links[i])
	}

	for i := range unit.LinksFrom {
		canonicalLinkDefaults(&unit.LinksFrom[i])
	}

	for _, sub := range unit.Subunits {
		fillCanonicalLinkDefaults(sub)
	}
}

// canonicalLinkDefaults fills one link's D-22 defaults.
func canonicalLinkDefaults(link *model.Link) {
	if link.Arrow == "" {
		link.Arrow = model.ArrowForward
	}

	if link.Rank == "" {
		link.Rank = model.RankForward
	}

	if link.LabelPosition == "" {
		link.LabelPosition = model.LabelMiddle
	}
}
