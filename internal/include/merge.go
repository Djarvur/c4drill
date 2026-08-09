package include

import (
	"fmt"
	"reflect"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
)

// merge applies src onto dst per the D-09/D-10/D-11/INC-08 rules and returns
// the merged model (mutates dst in place). dstFile/srcFile are display names
// for cross-file collision attribution.
//
//	Units:        union (D-11 cross-file dup key → *ParseError naming both files)
//	UnitOrder:    append src's NEW top-level keys in include order (D-09)
//	Subunits:     deep-merge per parent (D-10); cross-file subunit dup → *ParseError
//	Properties:   root-wins with non-zero conflict → *ParseError (INC-08)
//	Templates:    union (XC-02); dup name → *ParseError
//	Instantiations: append (both files' [[use]] flow through for template.Expand)
//
// Units are moved by pointer — each included file's units are distinct map
// entries with no aliasing risk (no round-trip-copy of Links: that would reset
// Link.Mirror and re-break multiplicity counting, STATE.md HS-1). Non-strict
// toml is load-bearing (BC-3); do NOT enable DisallowUnknownFields.
func merge(dst, src *parser.Model, dstFile, srcFile string) (*parser.Model, error) {
	if err := mergeUnits(dst, src, dstFile, srcFile); err != nil {
		return nil, err
	}

	if err := mergeProperties(&dst.Properties, src.Properties, dstFile, srcFile); err != nil {
		return nil, err
	}

	if err := mergeTemplates(dst, src.Templates, dstFile, srcFile); err != nil {
		return nil, err
	}

	// Both files' [[use]] instantiations flow through the merge so
	// template.Expand (Phase 31) sees instantiations from included files too.
	dst.Instantiations = append(dst.Instantiations, src.Instantiations...)

	return dst, nil
}

// mergeUnits unions src.Units into dst.Units. A key present in both files is a
// cross-file collision (D-11) unless the two share an existing parent being
// extended via subunits (D-10 handled by mergeSubunits on overlap). New
// src keys append onto dst.UnitOrder in src's UnitOrder order (D-09 append).
func mergeUnits(dst, src *parser.Model, dstFile, srcFile string) error {
	if dst.Units == nil {
		dst.Units = map[string]*model.Unit{}
	}

	for _, name := range src.UnitOrder {
		srcUnit, ok := src.Units[name]
		if !ok {
			continue
		}

		if existing, ok := dst.Units[name]; ok {
			// D-10: a top-level unit defined in both files is a collision
			// UNLESS the src entry contributes only subunits (the parent itself
			// is structurally identical by name). Hand off to mergeSubunits —
			// if src's unit has subunits, they attach to the existing parent;
			// otherwise it is a genuine duplicate-definition error.
			if err := mergeSubunits(existing, srcUnit, name, dstFile, srcFile); err != nil {
				return err
			}

			// If the src unit also redeclares non-subunit fields (type/name/...)
			// that is a cross-file collision (D-11). The parent's own scalar
			// fields are authoritative; src's are ignored unless src contributes
			// subunits (handled above). This keeps the entry file's [parent]
			// authoritative while letting an included file add [parent.child].
			continue
		}

		dst.Units[name] = srcUnit
		dst.UnitOrder = append(dst.UnitOrder, name)
	}

	return nil
}

// mergeSubunits attaches srcU's subunits onto dstU (the existing parent),
// appending to SubunitOrder in srcU's subunit order (D-10). A subunit key
// present in both the entry parent and the included parent is a cross-file
// collision (D-11) and hard-errors.
func mergeSubunits(dstU, srcU *model.Unit, path, dstFile, srcFile string) error {
	if len(srcU.Subunits) == 0 {
		// src redeclares the parent without contributing subunits → collision.
		return &parser.ParseError{
			Message: fmt.Sprintf("unit %q defined in both %s and %s", path, dstFile, srcFile),
			Context: path,
		}
	}

	if dstU.Subunits == nil {
		dstU.Subunits = map[string]*model.Unit{}
	}

	for _, child := range srcU.SubunitOrder {
		childUnit, ok := srcU.Subunits[child]
		if !ok {
			continue
		}

		if _, exists := dstU.Subunits[child]; exists {
			return &parser.ParseError{
				Message: fmt.Sprintf("subunit %s.%s defined in both %s and %s", path, child, dstFile, srcFile),
				Context: path + "." + child,
			}
		}

		dstU.Subunits[child] = childUnit
		dstU.SubunitOrder = append(dstU.SubunitOrder, child)
	}

	return nil
}

// mergeProperties applies INC-08 root-wins: for each non-zero field in src, if
// the same field on dst is non-zero AND differs, it is a conflict (hard error
// naming both files); otherwise the src value copies into dst (dst is the
// root/entry model, so dst already wins by virtue of being the destination — we
// only copy in fields dst lacks). Uses reflection over model.Properties so all
// 8 fields are covered uniformly without per-field enumeration.
func mergeProperties(dst *model.Properties, src model.Properties, dstFile, srcFile string) error {
	dstVal := reflect.ValueOf(dst).Elem()
	srcVal := reflect.ValueOf(src)

	for i := range dstVal.NumField() {
		srcField := srcVal.Field(i)
		if !srcField.IsValid() {
			continue
		}

		// Skip zero-valued src fields (no opinion from the included file).
		if srcField.IsZero() {
			continue
		}

		dstField := dstVal.Field(i)
		if dstField.IsZero() {
			// dst has no opinion; src wins (copy in).
			dstField.Set(srcField)

			continue
		}

		// Both non-zero: must be equal, else conflict (INC-08).
		if !reflect.DeepEqual(dstField.Interface(), srcField.Interface()) {
			fieldName := dstVal.Type().Field(i).Name

			return &parser.ParseError{
				Message: fmt.Sprintf(
					"conflicting [properties].%s: %s defines %v, %s defines %v",
					fieldName, dstFile, dstField.Interface(), srcFile, srcField.Interface(),
				),
				Context: "properties." + fieldName,
			}
		}
	}

	return nil
}

// mergeTemplates unions src.Templates into dst.Templates (XC-02: templates
// defined in included files flow through to the entry's [[use]]). A duplicate
// template name is a cross-file collision (hard error). No-op if src has no
// templates.
func mergeTemplates(dst *parser.Model, srcTemplates map[string]*parser.TemplateDef, dstFile, srcFile string) error {
	if len(srcTemplates) == 0 {
		return nil
	}

	if dst.Templates == nil {
		dst.Templates = map[string]*parser.TemplateDef{}
	}

	for name, tmpl := range srcTemplates {
		if _, exists := dst.Templates[name]; exists {
			return &parser.ParseError{
				Message: fmt.Sprintf("template %q defined in both %s and %s", name, dstFile, srcFile),
				Context: "template." + name,
			}
		}

		dst.Templates[name] = tmpl
	}

	return nil
}
