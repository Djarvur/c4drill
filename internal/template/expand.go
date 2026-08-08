// Package template expands [template.<name>] + [[use]] instantiations into
// concrete, parametrized unit subtrees drained into parser.Model.Units /
// UnitOrder, producing a model structurally indistinguishable from a
// hand-authored one.
//
// Plan 31-02 (TMPL-01..10, XC-03, XC-04). The expansion pass runs in the
// pipeline AFTER parser.Parse and BEFORE peer.Resolve + validator.Validate:
//
//	Parse -> template.Expand -> peer.Resolve -> Validate
//
// The load-bearing correctness concern is HS-1: the validator mutates
// Unit.LinksFrom in place (internal/validator/index.go:70-81), so each
// instantiation MUST deep-copy its template via model.Unit.Clone (which
// preserves the unexported Link.Mirror field). A shallow copy corrupts the
// Nth instantiation. The three-instantiation regression test
// (TestExpandThreeInstantiationsHS1) is the gate.
package template

import (
	"fmt"
	"strings"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
)

// substitutionTokenPrefix is the literal starting marker of a ${param} token.
// Its presence in any string field AFTER substitution is a hard error
// (TMPL-06 belt-and-suspenders: missing params should already have been
// caught by the pre-check).
const substitutionTokenPrefix = "${"

// nameParam is the special parameter name that fills the produced unit's last
// path segment (D-04). It is always required even if the template does not
// list it in its declared params, because every produced unit needs a path.
const nameParam = "name"

// replacerPairsPerParam is the number of strings.NewReplacer pairs (old, new)
// per parameter: "${param}" and the value.
const replacerPairsPerParam = 2

// Expand turns every [[use]] instantiation in m into a concrete, parametrized
// unit subtree drained into m.Units / m.UnitOrder. It mutates m in place and
// returns it (the same pointer) on success. Expand owns the model between
// Parse and peer.Resolve/Validate.
//
// Per instantiation (in document order):
//  1. Look up the named [template.<name>] table; missing template = hard error.
//  2. Verify every declared param is present; missing param = hard error
//     naming template + param + instantiation site (TMPL-06).
//  3. Clone the template's *model.Unit (HS-1) and substitute ${param} -> value
//     into every string field of the clone + subunits + links (TMPL-03).
//  4. Attach the produced unit under the [[use]] parent (XC-03) or top-level.
//  5. Track produced full paths for collision detection.
//
// After the loop, a full-path collision check (hand-authored + all instances)
// surfaces any duplicate as a single hard error naming both sources (TMPL-07),
// and a residual-"${" scan catches any unsatisfied token (TMPL-06).
//
// Expand is a no-op (returns m unchanged, nil error) when m has no templates
// and no instantiations — guaranteeing no regression for hand-authored models.
func Expand(m *parser.Model) (*parser.Model, error) {
	if m == nil {
		//nolint:nilnil // nil input -> nil output is the identity contract; no error to report
		return nil, nil
	}

	// Fast path: nothing to expand. Preserve the no-regression contract for
	// hand-authored-only models (testdata/valid.toml etc.).
	if len(m.Templates) == 0 && len(m.Instantiations) == 0 {
		return m, nil
	}

	// produced tracks the full path of every expanded instance for the
	// post-loop collision check (TMPL-07). Pre-seeded with hand-authored
	// top-level paths so a templated unit cannot silently overwrite one.
	produced := newPathTracker(m)

	for i := range m.Instantiations {
		inst := m.Instantiations[i]

		if err := expandInstantiation(m, inst, i, produced); err != nil {
			return nil, err
		}
	}

	// TMPL-06 belt-and-suspenders: scan every produced (and hand-authored)
	// unit's string fields for any residual "${" — a missing param should
	// already have failed the pre-check, but this catches substitution bugs.
	if err := assertNoResidualTokens(m); err != nil {
		return nil, err
	}

	return m, nil
}

// expandInstantiation expands a single [[use]] entry: resolves the template,
// checks params, clones, substitutes, and attaches the produced subtree.
func expandInstantiation(
	m *parser.Model,
	inst parser.Instantiation,
	index int,
	produced *pathTracker,
) error {
	tmpl, ok := m.Templates[inst.Template]
	if !ok {
		return &ExpandError{
			Kind:    "unknown template",
			Site:    siteLabel(inst, index),
			Detail:  fmt.Sprintf("template %q is not defined", inst.Template),
		}
	}

	// TMPL-06: every declared param must be supplied. name is always required
	// (it is the produced unit's path segment per D-04) even if the template
	// did not list it explicitly — enforce it here so a template without a
	// declared 'name' param still produces a path-named unit.
	for _, p := range tmpl.Params {
		if _, present := inst.Params[p]; !present {
			return &ExpandError{
				Kind:   "missing parameter",
				Site:   siteLabel(inst, index),
				Detail: fmt.Sprintf("template %q requires parameter %q which is not supplied", inst.Template, p),
			}
		}
	}

	// Deep-copy the template subtree (HS-1). The clone's LinksFrom is
	// independent, so the validator's in-place appends stay isolated per
	// instantiation.
	clone := tmpl.Unit.Clone()
	if clone == nil {
		return &ExpandError{
			Kind:   "empty template",
			Site:   siteLabel(inst, index),
			Detail: fmt.Sprintf("template %q has no unit subtree", inst.Template),
		}
	}

	// Build the replacer from declared params -> supplied values. Include the
	// 'name' param even if undeclared so authors can reference ${name} in any
	// template (D-04 makes name the path segment; it is always supplied).
	replacer := buildReplacer(tmpl.Params, inst.Params)
	applySubstitution(clone, replacer)

	// D-04: produced path = parent + "." + name (or just name if no parent).
	name := inst.Params[nameParam]
	if name == "" {
		return &ExpandError{
			Kind:   "missing parameter",
			Site:   siteLabel(inst, index),
			Detail: fmt.Sprintf(
				"template %q instantiation is missing the %q parameter "+
					"(produces the unit path segment)",
				inst.Template, nameParam,
			),
		}
	}

	if err := attachProduced(m, inst.Parent, name, clone, produced, inst, index); err != nil {
		return err
	}

	return nil
}

// buildReplacer constructs a strings.NewReplacer from ${param} -> value pairs
// for every declared param, plus 'name' if present in supplied (defensive —
// 'name' is always enforced by expandInstantiation, but templates may reference
// it without declaring it).
func buildReplacer(declared []string, supplied map[string]string) *strings.Replacer {
	seen := make(map[string]bool, len(declared)+1)
	pairs := make([]string, 0, (len(declared)+1)*replacerPairsPerParam)

	add := func(param string) {
		if seen[param] {
			return
		}

		seen[param] = true
		pairs = append(pairs, "${"+param+"}", supplied[param])
	}

	for _, p := range declared {
		add(p)
	}

	// 'name' is special (D-04) — always include it if supplied so authors can
	// reference ${name} without listing it in params.
	if _, hasName := supplied[nameParam]; hasName {
		add(nameParam)
	}

	return strings.NewReplacer(pairs...)
}

// applySubstitution walks a unit tree applying replacer.Replace to every string
// field of the unit and its subunits and links. Recurses into Subunits.
func applySubstitution(u *model.Unit, r *strings.Replacer) {
	if u == nil {
		return
	}

	u.Name = r.Replace(u.Name)
	u.Description = r.Replace(u.Description)
	u.Technology = r.Replace(u.Technology)
	u.Reference = r.Replace(u.Reference)
	u.Color = r.Replace(u.Color)
	u.Style = r.Replace(u.Style)
	u.Border = r.Replace(u.Border)
	u.Edges = r.Replace(u.Edges)

	for i := range u.Links {
		applySubstitutionLink(&u.Links[i], r)
	}

	// LinksFrom: authored entries are substituted too (a template may declare
	// an authored linkFrom with a parametrized peer/description). Validator-
	// synthesized mirrors do not exist yet at Expand time (they are appended
	// later by populateIncomingLinks), so every entry here is authored.
	for i := range u.LinksFrom {
		applySubstitutionLink(&u.LinksFrom[i], r)
	}

	for _, child := range u.Subunits {
		applySubstitution(child, r)
	}
}

// applySubstitutionLink substitutes the string fields of a single Link.
func applySubstitutionLink(l *model.Link, r *strings.Replacer) {
	l.Peer = r.Replace(l.Peer)
	l.Description = r.Replace(l.Description)
	l.Technology = r.Replace(l.Technology)
	l.Color = r.Replace(l.Color)
	l.Style = r.Replace(l.Style)
}

// attachProduced places the produced unit under its parent (XC-03) and records
// the full path for collision detection. parent="" means top-level.
func attachProduced(
	m *parser.Model,
	parent, name string,
	unit *model.Unit,
	produced *pathTracker,
	inst parser.Instantiation,
	index int,
) error {
	if parent == "" {
		return attachTopLevel(m, name, unit, produced, inst, index)
	}

	return attachNested(m, parent, name, unit, produced, inst, index)
}

// attachTopLevel places the produced unit at m.Units[name] and appends to
// m.UnitOrder. Records the path for collision detection.
func attachTopLevel(
	m *parser.Model,
	name string,
	unit *model.Unit,
	produced *pathTracker,
	inst parser.Instantiation,
	index int,
) error {
	if err := produced.claim(name, inst, index); err != nil {
		return err
	}

	if m.Units == nil {
		m.Units = make(map[string]*model.Unit)
	}

	m.Units[name] = unit
	m.UnitOrder = append(m.UnitOrder, name)

	return nil
}

// attachNested places the produced unit under the parent's Subunits. The parent
// is resolved by walking the tree along the dotted path (parent may itself be
// top-level or nested). Records parent+"."+name for collision detection.
func attachNested(
	m *parser.Model,
	parent, name string,
	unit *model.Unit,
	produced *pathTracker,
	inst parser.Instantiation,
	index int,
) error {
	fullPath := parent + "." + name

	if err := produced.claim(fullPath, inst, index); err != nil {
		return err
	}

	parentUnit, ok := resolveUnitByPath(m, parent)
	if !ok {
		return &ExpandError{
			Kind:   "unknown parent",
			Site:   siteLabel(inst, index),
			Detail: fmt.Sprintf("parent %q does not exist in the model", parent),
		}
	}

	if parentUnit.Subunits == nil {
		parentUnit.Subunits = make(map[string]*model.Unit)
	}

	parentUnit.Subunits[name] = unit
	parentUnit.SubunitOrder = append(parentUnit.SubunitOrder, name)

	return nil
}

// resolveUnitByPath walks m.Units along a dotted path (e.g. "a.b.c") and
// returns the deepest unit, mirroring validator.BuildIndex / peer.Resolve path
// construction. Returns (nil, false) if any segment is missing.
func resolveUnitByPath(m *parser.Model, path string) (*model.Unit, bool) {
	if path == "" {
		return nil, false
	}

	segments := strings.Split(path, ".")

	current, ok := m.Units[segments[0]]
	if !ok {
		return nil, false
	}

	for _, seg := range segments[1:] {
		if current.Subunits == nil {
			return nil, false
		}

		next, ok := current.Subunits[seg]
		if !ok {
			return nil, false
		}

		current = next
	}

	return current, true
}

// assertNoResidualTokens scans every unit (top-level + nested) for any "${"
// remaining in a string field after substitution. A residual token means a
// param was referenced but not supplied (TMPL-06) or a substitution bug.
func assertNoResidualTokens(m *parser.Model) error {
	return scanUnits(m.Units, func(path, field, value string) error {
		if strings.Contains(value, substitutionTokenPrefix) {
			return &ExpandError{
				Kind:   "unresolved parameter",
				Site:   path,
				Detail: fmt.Sprintf("field %q still contains %q after expansion: %q", field, substitutionTokenPrefix, value),
			}
		}

		return nil
	})
}

// scanUnits walks the unit tree and calls visit on every string field of every
// unit. path is the dotted unit path; field is the field name; value is the
// current string value. visit returning a non-nil error aborts the walk.
func scanUnits(
	units map[string]*model.Unit,
	visit func(path, field, value string) error,
) error {
	return scanUnitsPath(units, "", visit)
}

func scanUnitsPath(
	units map[string]*model.Unit,
	parentPath string,
	visit func(path, field, value string) error,
) error {
	for name, unit := range units {
		fullPath := name
		if parentPath != "" {
			fullPath = parentPath + "." + name
		}

		if err := scanUnitFields(unit, fullPath, visit); err != nil {
			return err
		}

		if len(unit.Subunits) > 0 {
			if err := scanUnitsPath(unit.Subunits, fullPath, visit); err != nil {
				return err
			}
		}
	}

	return nil
}

// scanUnitFields calls visit on every string field of unit + its links.
func scanUnitFields(
	unit *model.Unit,
	path string,
	visit func(path, field, value string) error,
) error {
	fields := []struct {
		name  string
		value string
	}{
		{"name", unit.Name},
		{"description", unit.Description},
		{"technology", unit.Technology},
		{"reference", unit.Reference},
		{"color", unit.Color},
		{"style", unit.Style},
		{"border", unit.Border},
		{"edges", unit.Edges},
	}

	for _, f := range fields {
		if err := visit(path, f.name, f.value); err != nil {
			return err
		}
	}

	for i := range unit.Links {
		linkPath := fmt.Sprintf("%s.link[%d]", path, i)
		if err := visit(linkPath, "peer", unit.Links[i].Peer); err != nil {
			return err
		}

		if err := visit(linkPath, "description", unit.Links[i].Description); err != nil {
			return err
		}

		if err := visit(linkPath, "technology", unit.Links[i].Technology); err != nil {
			return err
		}
	}

	return nil
}

// siteLabel renders a human-readable instantiation-site identifier for error
// messages: the [[use]]'s name param if present, else its document index.
func siteLabel(inst parser.Instantiation, index int) string {
	if name, ok := inst.Params[nameParam]; ok && name != "" {
		return fmt.Sprintf("[[use]] #%d (name=%q, template=%q)", index+1, name, inst.Template)
	}

	return fmt.Sprintf("[[use]] #%d (template=%q)", index+1, inst.Template)
}

// pathTracker records the full paths claimed by hand-authored units and
// expanded instances so a duplicate is caught as a single hard error (TMPL-07).
type pathTracker struct {
	owner  map[string]string // path -> source label
}

func newPathTracker(m *parser.Model) *pathTracker {
	pt := &pathTracker{owner: make(map[string]string)}
	// Seed with hand-authored top-level paths; recurse for nested.
	seedAuthored(m.Units, "", pt)

	return pt
}

// seedAuthored walks hand-authored units recording their full paths as
// "hand-authored" sources so an expanded instance colliding with them surfaces
// a clear conflict.
func seedAuthored(units map[string]*model.Unit, parentPath string, pt *pathTracker) {
	for name, unit := range units {
		fullPath := name
		if parentPath != "" {
			fullPath = parentPath + "." + name
		}

		pt.owner[fullPath] = "hand-authored unit " + fullPath

		if len(unit.Subunits) > 0 {
			seedAuthored(unit.Subunits, fullPath, pt)
		}
	}
}

// claim records a path for an expanded instance. A duplicate path yields a hard
// error naming both the prior owner and the current instantiation.
func (pt *pathTracker) claim(path string, inst parser.Instantiation, index int) error {
	prior, exists := pt.owner[path]
	if exists {
		return &ExpandError{
			Kind:   "duplicate unit path",
			Site:   siteLabel(inst, index),
			Detail: fmt.Sprintf("path %q is already claimed by %s", path, prior),
		}
	}

	pt.owner[path] = "instantiation " + siteLabel(inst, index)

	return nil
}

// ExpandError describes a template-expansion failure (unknown template /
// missing parameter / duplicate path / unresolved token / unknown parent).
// It follows the *parser.ParseError struct-with-Error idiom.
type ExpandError struct {
	// Kind categorizes the failure (e.g. "missing parameter", "duplicate unit path").
	Kind string
	// Site is a human-readable instantiation-site identifier.
	Site string
	// Detail is a human-readable explanation including the conflicting names.
	Detail string
}

// Error returns a human-readable diagnostic.
func (e *ExpandError) Error() string {
	if e.Site != "" {
		return fmt.Sprintf("template expand: %s at %s: %s", e.Kind, e.Site, e.Detail)
	}

	return fmt.Sprintf("template expand: %s: %s", e.Kind, e.Detail)
}
