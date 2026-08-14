// Package parser provides TOML parsing for C4 architecture definitions.
package parser

import (
	"os"
	"slices"
	"strings"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/pelletier/go-toml/v2"
	"github.com/pelletier/go-toml/v2/unstable"
)

// Default type based on nesting level:
// - C1 (root level): system
// - C2 (inside system/box): container
// - C3 (inside container): component.
const (
	defaultTypeC1 = model.TypeSystem
	defaultTypeC2 = model.TypeContainer
	defaultTypeC3 = model.TypeComponent
)

// Generic types that can be auto-transformed based on nesting level:
// - db at C1 -> db, at C2 -> containerDb, at C3 -> componentDb
// - queue at C1 -> queue, at C2 -> containerQueue, at C3 -> componentQueue.
//
//nolint:gochecknoglobals // Lookup map for O(1) type checking, immutable after init
var genericDbTypes = map[model.UnitType]bool{
	model.TypeDb:    true,
	model.TypeQueue: true,
}

// Model represents the root of a parsed TOML document.
// It contains the properties section and all top-level units.
type Model struct {
	// Properties contains the global [properties] section.
	Properties model.Properties `toml:"properties"`
	// UnitOrder tracks the definition order of unit names.
	UnitOrder []string
	// Units contains all top-level units keyed by section name.
	Units map[string]*model.Unit
	// Templates holds parsed [template.<name>] tables keyed by <name>.
	// Populated by Parse from rawMap (not direct unmarshal) — see Plan 31-01.
	// Consumed by internal/template.Expand (Plan 31-02). Empty for hand-authored
	// models with no [template.*] tables.
	Templates map[string]*TemplateDef `toml:"-"`
	// Instantiations holds parsed [[use]] array-of-tables entries in document
	// order. Populated by Parse from rawMap. Consumed by internal/template.Expand.
	// Empty for hand-authored models with no [[use]] tables.
	Instantiations []Instantiation `toml:"-"`
	// Includes holds parsed [[include]] array-of-tables entries in document
	// order. Populated by Parse from rawMap (Plan 32-01). Consumed by
	// internal/include.Resolve (Plan 32-02). Empty for single-file models with
	// no [[include]] tables.
	Includes []IncludeDirective `toml:"-"`
}

// IncludeDirective captures a single [[include]] array-of-tables entry: the
// referenced file path (relative to the including file's directory, resolved
// by internal/include.Resolve) and the optional once flag (INC-06: a once=true
// file is included at most once even when reached via multiple paths).
type IncludeDirective struct {
	// Path is the TOML file to merge in, relative to the including file's dir.
	Path string `toml:"path"`
	// Once opts into the visited-set dedup: an already-included file is skipped
	// on subsequent [[include]] directives (INC-06).
	Once bool `toml:"once"`
}

// TemplateDef holds a parsed [template.<name>] table: its declared named
// parameters and the parsed *model.Unit subtree (including any declared
// [[template.<name>.link]] arrays parsed into Unit.Links and subunit subtrees
// parsed into Unit.Subunits). Params are NOT substituted at parse time — that
// is internal/template.Expand's job (Plan 31-02).
type TemplateDef struct {
	// Params lists the declared parameter names from the `params = [...]` array.
	Params []string
	// Unit is the parsed template subtree. Its fields carry literal ${param}
	// tokens awaiting substitution. Populated manually via parseUnitWithOrder
	// (NOT direct toml unmarshal), so no toml tag is needed here.
	Unit *model.Unit
	// Instantiations holds [[template.<name>.<path>.use]] array-of-tables
	// entries parsed from the template body (D-17, Plan 35-02):
	// template-instantiating-template. Parent paths are RELATIVE to the
	// template's unit root (empty = the produced unit attaches as a direct
	// subunit of the instantiated root); an explicit `parent` key in an entry
	// is relative to the site's path. NOT expanded at parse time —
	// internal/template.Expand recurses through them when the outer template
	// is instantiated (params flow outer-to-inner). Equivalent authoring
	// form: a top-level [[use]] with parent = the enclosing unit's dotted
	// path (D-16) — both normalize to the same Instantiation mechanism.
	Instantiations []Instantiation
}

// UseSite records one [[use]]-family array-of-tables site in document order:
// a top-level [[use]], a unit-nested [[unit.<path>.use]] (D-16), or a
// template-body [[template.<name>.<path>.use]] (D-17). captureDefinitionOrder
// records the sites (Go maps lose document order) so the extraction pass can
// interleave all three forms in authoring order; entry VALUES are read from
// the raw map afterwards.
type UseSite struct {
	// Parent is the dotted unit path the site's uses attach under. Empty for
	// a top-level [[use]]. For template sites it is relative to the template's
	// unit root.
	Parent string
	// Template is the owning [template.<name>] when the site lives inside a
	// template body ([[template.<name>...use]]); empty otherwise.
	Template string
}

// Instantiation captures a single [[use]] array-of-tables entry: the template
// name to instantiate, the optional parent placement path, and the supplied
// named parameters (including "name" as a regular parameter per D-04).
type Instantiation struct {
	// Template is the name of the [template.<name>] table to instantiate.
	Template string `toml:"template"`
	// Parent is the optional dotted placement path (empty = top-level).
	// The produced unit's full path = Parent + "." + Name-param (or just the
	// Name-param if Parent is empty) per D-04.
	Parent string `toml:"parent"`
	// Params carries all supplied named parameters, including "name".
	Params map[string]string `toml:"params"`
}

// Parse parses TOML data into a Model.
// It unmarshals the TOML content, with links automatically parsed from [[link]] arrays.
// Definition order of units and subunits is preserved.
func Parse(data []byte) (*Model, error) {
	// First pass: capture definition order using unstable API
	unitOrder, subunitOrders, templateSubunitOrders, useSites, err := captureDefinitionOrder(data)
	if err != nil {
		return nil, wrapDecodeError(err)
	}

	// Second pass: unmarshal the entire document to a raw map
	var rawMap map[string]any

	if err := toml.Unmarshal(data, &rawMap); err != nil {
		return nil, wrapDecodeError(err)
	}

	// Third pass: build the Model struct
	m := &Model{
		UnitOrder: unitOrder,
		Units:     make(map[string]*model.Unit),
	}

	// Extract properties if present
	if props, ok := rawMap["properties"]; ok {
		propsData, err := toml.Marshal(props)
		if err != nil {
			return nil, &ParseError{Message: "failed to marshal properties", Cause: err}
		}

		if err := toml.Unmarshal(propsData, &m.Properties); err != nil {
			return nil, wrapDecodeError(err)
		}
	}

	// BC-1 (D-08, Plan 31-01): extract reserved tables BEFORE the unit loop so
	// they neither register as phantom units nor get re-parsed as units. These
	// mirror the properties extraction above: pull the raw value, re-marshal,
	// unmarshal into the dedicated Model field.
	if err := extractTemplates(rawMap, templateSubunitOrders, m); err != nil {
		return nil, err
	}

	if err := extractUses(rawMap, useSites, m); err != nil {
		return nil, err
	}

	// Phase 32 (Plan 32-01): extract [[include]] array-of-tables into m.Includes.
	// The BC-1 skip in captureDefinitionOrder already prevents [[include]] from
	// registering a phantom 'include' unit; extractIncludes also deletes the
	// key from rawMap as belt-and-suspenders so the directive cannot reach the
	// unit loop even if the skip were ever removed. Non-strict toml stays
	// (BC-3) — go-toml/v2 silently accepts the bare [[include]] key (D-06).
	if err := extractIncludes(rawMap, m); err != nil {
		return nil, err
	}

	// Process units in the captured order (not rawMap iteration order)
	for _, name := range unitOrder {
		value, ok := rawMap[name]
		if !ok {
			continue // Should not happen if captureDefinitionOrder is correct
		}

		subunitOrder := subunitOrders[name]

		unit, err := parseUnitWithOrder(name, value, "", subunitOrder, subunitOrders)
		if err != nil {
			return nil, err
		}

		m.Units[name] = unit
	}

	return m, nil
}

// extractTemplates pulls rawMap["template"] into m.Templates. Each template
// name's value is re-parsed via parseUnitWithOrder so its declared
// [[template.<name>.link]] arrays become model.Unit.Links and any subunit
// subtree ([template.<name>.<child>]) becomes model.Unit.Subunits, identical
// to hand-authored unit semantics. Params are captured separately. Substitution
// is NOT applied here (Plan 31-02's job).
func extractTemplates(
	rawMap map[string]any,
	templateSubunitOrders map[string][]string,
	m *Model,
) error {
	tmplRoot, ok := rawMap["template"]
	if !ok {
		return nil
	}

	tmplMap, ok := tmplRoot.(map[string]any)
	if !ok {
		return &ParseError{
			Message: "invalid [template] format: expected a table",
			Context: "template",
		}
	}

	m.Templates = make(map[string]*TemplateDef, len(tmplMap))

	for name, val := range tmplMap {
		tmpl, err := parseTemplateDef(name, val, templateSubunitOrders)
		if err != nil {
			return err
		}

		m.Templates[name] = tmpl
	}

	return nil
}

// parseTemplateDef parses a single [template.<name>] table into a TemplateDef.
// It separates the declared `params` array from the unit subtree (the unit
// parser treats unknown keys as subunits, and `params` is not a Unit field),
// then re-parses the remainder into a *model.Unit with subunit order preserved
// relative to the template namespace. Substitution is NOT applied.
func parseTemplateDef(
	name string,
	val any,
	templateSubunitOrders map[string][]string,
) (*TemplateDef, error) {
	tmplUnitMap, ok := val.(map[string]any)
	if !ok {
		return nil, &ParseError{
			Message: "invalid template format: expected a table",
			Context: "template." + name,
		}
	}

	// Strip `params` from a copy of the map so the subtree parser sees only
	// Unit fields + declared subunits.
	var params []string

	unitMapCopy := make(map[string]any, len(tmplUnitMap))

	for k, v := range tmplUnitMap {
		if k == "params" {
			if rawParams, ok := v.([]any); ok {
				params = toStringSlice(rawParams)
			}

			continue
		}

		unitMapCopy[k] = v
	}

	// Re-parse the subtree into a *model.Unit. The subunit order is
	// captured relative to the template's own namespace (keyed by `name`
	// and `name.child`) so template subunits preserve authoring order.
	subOrder := templateSubunitOrders[name]

	unit, err := parseUnitWithOrder(name, unitMapCopy, "", subOrder, templateSubunitOrders)
	if err != nil {
		return nil, err
	}

	return &TemplateDef{Params: params, Unit: unit}, nil
}

// toStringSlice coerces a raw TOML array into a []string. Non-string elements
// are skipped. Used for the template `params = [...]` array.
func toStringSlice(arr []any) []string {
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}

	return out
}

// extractUses desugars every [[use]]-family site into Instantiation entries
// in document order (Plan 35-02). Top-level [[use]] and unit-nested
// [[unit.<path>.use]] entries (D-16) land in m.Instantiations — the nested
// form's Parent is the enclosing unit's dotted path, an explicit `parent` key
// being relative to it. Template-body [[template.<name>.<path>.use]] entries
// (D-17) land in m.Templates[<name>].Instantiations with Parents relative to
// the template's unit root. The two unit-level forms are the SAME
// Instantiation mechanism the expansion pass consumes; no parallel semantics.
func extractUses(rawMap map[string]any, useSites []UseSite, m *Model) error {
	// cursors tracks how many entries of each site path's array have been
	// consumed: the Nth site with the same path pairs with the Nth element of
	// the array at that path (every [[...use]] header appends exactly one
	// element), preserving document order across repeated blocks.
	cursors := make(map[string]int)

	for _, site := range useSites {
		if err := extractSiteUses(rawMap, site, cursors, m); err != nil {
			return err
		}
	}

	return nil
}

// extractSiteUses desugars the single site's paired array element (see
// extractUses for the cursor pairing) into either m.Instantiations (top-level
// and unit-nested sites) or the owning template's Instantiations.
func extractSiteUses(
	rawMap map[string]any,
	site UseSite,
	cursors map[string]int,
	m *Model,
) error {
	context := useContext(site)

	raw := lookupUseArray(rawMap, site)

	useArr, ok := raw.([]any)
	if !ok {
		// Malformed site (e.g. a bare [x.use] table): parseUseEntries renders
		// the standard format error.
		_, err := parseUseEntries(raw, context)

		return err
	}

	idx := cursors[context]
	cursors[context]++

	if idx >= len(useArr) {
		return &ParseError{
			Message: "invalid [[use]] site: no entry matches the recorded site",
			Context: context,
		}
	}

	entries, err := parseUseEntries([]any{useArr[idx]}, context)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// An explicit `parent` key resolves RELATIVE to the site's path
		// (for a top-level site the site path is "" — absolute, exactly
		// the Phase 31 semantics).
		entry.Parent = joinPath(site.Parent, entry.Parent)

		if site.Template != "" {
			if err := appendTemplateUse(m, site, entry, context); err != nil {
				return err
			}

			continue
		}

		m.Instantiations = append(m.Instantiations, entry)
	}

	return nil
}

// appendTemplateUse routes a template-body use entry into the owning
// template's Instantiations (D-17). Hard error when the owning template is
// not defined.
func appendTemplateUse(m *Model, site UseSite, entry Instantiation, context string) error {
	tmpl := m.Templates[site.Template]
	if tmpl == nil {
		return &ParseError{
			Message: "invalid [[use]] site: template is not defined",
			Context: context,
		}
	}

	tmpl.Instantiations = append(tmpl.Instantiations, entry)

	return nil
}

// parseUseEntries decodes one [[use]]-family array into Instantiation entries,
// mirroring the Phase 31 top-level strictness: every value must be a string
// (all template params are string-substituted); `template` and `parent` are
// the fixed fields; every other key is a supplied named parameter (including
// "name" per D-04). context names the site in error messages.
func parseUseEntries(raw any, context string) ([]Instantiation, error) {
	useArr, ok := raw.([]any)
	if !ok {
		return nil, &ParseError{
			Message: "invalid [[use]] format: expected an array of tables",
			Context: context,
		}
	}

	out := make([]Instantiation, 0, len(useArr))

	for _, entry := range useArr {
		useMap, ok := entry.(map[string]any)
		if !ok {
			return nil, &ParseError{
				Message: "invalid [[use]] entry: expected a table",
				Context: context,
			}
		}

		inst := Instantiation{Params: make(map[string]string)}

		for k, v := range useMap {
			s, ok := v.(string)
			if !ok {
				// Non-string values are not supported for [[use]] params
				// (all template params are string-substituted). Surface as
				// a parse error naming the offending key.
				return nil, &ParseError{
					Message: "invalid [[use]] field: expected a string value",
					Context: context + "." + k,
				}
			}

			switch k {
			case "template":
				inst.Template = s
			case "parent":
				inst.Parent = s
			default:
				inst.Params[k] = s
			}
		}

		out = append(out, inst)
	}

	return out, nil
}

// lookupUseArray walks rawMap along the site's path and returns the raw "use"
// value at that site (top-level sites read rawMap["use"] directly). Returns
// nil when any segment is missing — impossible for a site recorded from the
// same document, so nil flows into parseUseEntries as the malformed-format
// error (defensive).
func lookupUseArray(rawMap map[string]any, site UseSite) any {
	var segments []string
	if site.Template != "" {
		segments = append(segments, reservedTemplate, site.Template)
	}

	if site.Parent != "" {
		segments = append(segments, strings.Split(site.Parent, ".")...)
	}

	current := any(rawMap)

	for _, seg := range segments {
		curMap, ok := current.(map[string]any)
		if !ok {
			return nil
		}

		current, ok = curMap[seg]
		if !ok {
			return nil
		}
	}

	unitMap, ok := current.(map[string]any)
	if !ok {
		return nil
	}

	return unitMap["use"]
}

// useContext renders the error-attribution name for a use site: "use" for a
// top-level site (identical to the Phase 31 error contexts), the dotted unit
// path + ".use" for unit-nested sites, "template.<name>[.<path>].use" for
// template-body sites.
func useContext(site UseSite) string {
	if site.Template != "" {
		return joinPath(joinPath(reservedTemplate, site.Template), site.Parent) + "." + reservedUse
	}

	return joinPath(site.Parent, reservedUse)
}

// joinPath joins two dotted path segments, treating "" as identity
// (joinPath("", "a") = "a", joinPath("a", "") = "a", joinPath("", "") = "").
func joinPath(base, rel string) string {
	switch {
	case base == "":
		return rel
	case rel == "":
		return base
	default:
		return base + "." + rel
	}
}

// extractIncludes pulls rawMap["include"] into m.Includes, preserving the
// array-of-tables document order. Each [[include]] table has two fields:
// `path` (required) and `once` (optional, defaults false). Mirrors
// extractInstantiations (NOT the properties marshal/unmarshal pattern, which
// does not work for top-level arrays). Deletes rawMap["include"] afterwards as
// belt-and-suspenders against the unit loop — captureDefinitionOrder's BC-1
// skip already keeps [[include]] out of unitOrder.
func extractIncludes(rawMap map[string]any, m *Model) error {
	incRoot, ok := rawMap["include"]
	if !ok {
		return nil
	}

	incArr, ok := incRoot.([]any)
	if !ok {
		return &ParseError{
			Message: "invalid [[include]] format: expected an array of tables",
			Context: "include",
		}
	}

	m.Includes = make([]IncludeDirective, 0, len(incArr))

	for _, entry := range incArr {
		directive, err := parseIncludeDirective(entry)
		if err != nil {
			return err
		}

		m.Includes = append(m.Includes, directive)
	}

	delete(rawMap, "include")

	return nil
}

// parseIncludeDirective decodes a single [[include]] table into an
// IncludeDirective. `path` must be a string; `once` must be a boolean. Unknown
// keys are accepted (non-strict toml, BC-3) to avoid coupling the parser to
// future fields.
func parseIncludeDirective(entry any) (IncludeDirective, error) {
	incMap, ok := entry.(map[string]any)
	if !ok {
		return IncludeDirective{}, &ParseError{
			Message: "invalid [[include]] entry: expected a table",
			Context: "include",
		}
	}

	directive := IncludeDirective{}

	for k, v := range incMap {
		switch k {
		case "path":
			s, ok := v.(string)
			if !ok {
				return IncludeDirective{}, &ParseError{
					Message: "invalid [[include]] field: path must be a string",
					Context: "include.path",
				}
			}

			directive.Path = s
		case "once":
			b, ok := v.(bool)
			if !ok {
				return IncludeDirective{}, &ParseError{
					Message: "invalid [[include]] field: once must be a boolean",
					Context: "include.once",
				}
			}

			directive.Once = b
		default:
			// Unknown keys accepted (BC-3 non-strict toml); nothing to do.
		}
	}

	return directive, nil
}

// Reserved top-level table names (BC-1, D-08). Bare names — no namespace
// prefix (D-06). template is a parent namespace ([template.<name>]); use and
// include are array-of-tables ([[use]] / [[include]]).
const (
	reservedProperties = "properties"
	reservedTemplate   = "template"
	reservedUse        = "use"
	reservedInclude    = "include"
)

// captureDefinitionOrder uses the unstable API to capture the order of units
// and subunits. Returns: unitOrder (top-level), subunitOrders (nested,
// hand-authored), templateSubunitOrders (nested, WITHIN the [template.*]
// namespace), useSites ([[use]]-family array-of-tables in document order,
// Plan 35-02), error.
func captureDefinitionOrder(data []byte) ([]string, map[string][]string, map[string][]string, []UseSite, error) {
	unitOrder := make([]string, 0)
	subunitOrders := make(map[string][]string)
	templateSubunitOrders := make(map[string][]string)
	useSites := make([]UseSite, 0)

	seenUnits := make(map[string]bool)
	seenSubunits := make(map[string]bool)
	seenTemplateSubunits := make(map[string]bool)

	p := unstable.Parser{}
	p.Reset(data)

	for p.NextExpression() {
		expr := p.Expression()

		if !isTableExpression(expr) {
			continue // Skip key-value / comment expressions
		}

		parts := extractKeyParts(expr.Key())

		if len(parts) == 0 {
			continue
		}

		// [[use]]-family sites (D-16/D-17, Plan 35-02): top-level [[use]],
		// unit-nested [[unit.<path>.use]], template-body
		// [[template.<name>.<path>.use]]. Recorded in document order for the
		// extraction pass (Go maps lose authoring order); never units or
		// subunits. A bare [x.use] table also records so extraction can reject
		// the malformed form with the standard error.
		if parts[len(parts)-1] == reservedUse {
			recordUseSite(parts, &useSites)

			continue
		}

		if expr.Kind == unstable.Table {
			recordUnitTable(parts, &unitOrder, subunitOrders, templateSubunitOrders,
				seenUnits, seenSubunits, seenTemplateSubunits)
		}
		// Other array-of-tables ([[a.link]] etc.) are not units — skip.
	}

	if err := p.Error(); err != nil {
		//nolint:wrapcheck // unstable.Parser error is wrapped by Parse's caller via wrapDecodeError
		return nil, nil, nil, nil, err
	}

	return unitOrder, subunitOrders, templateSubunitOrders, useSites, nil
}

// isTableExpression reports whether expr is a [table] or [[array-of-tables]]
// header — the only expression kinds this pass records from. Array-of-tables
// expressions pass ONLY on their way to the [[...use]] site check; every other
// array-of-tables is skipped exactly as before use sites existed.
func isTableExpression(expr *unstable.Node) bool {
	return expr.Kind == unstable.Table || expr.Kind == unstable.ArrayTable
}

// recordUnitTable records a regular [table] header into the order structures:
// [properties]/[[include]] are skipped (reserved), [template.*] subtrees feed
// the template-namespace orders, everything else is a hand-authored unit or
// subunit.
func recordUnitTable(
	parts []string,
	unitOrder *[]string,
	subunitOrders map[string][]string,
	templateSubunitOrders map[string][]string,
	seenUnits, seenSubunits, seenTemplateSubunits map[string]bool,
) {
	// Skip [properties] section
	if len(parts) == 1 && parts[0] == reservedProperties {
		return
	}

	// BC-1 (D-08, Plan 31-01): skip reserved top-level tables so they neither
	// register phantom units nor leak into the unit loop. [[include]] is an
	// array-of-tables skipped by the Kind filter in the caller; this catches a
	// bare [include] table. [[use]] is handled by the site check in the caller.
	if len(parts) == 1 && parts[0] == reservedInclude {
		return
	}

	// [template.*] subtrees: do NOT enter unitOrder/subunitOrders, but DO
	// capture the subunit structure within the template namespace so the
	// extraction pass can preserve authoring order. Paths of the form
	// [template.<name>] define a template root; [template.<name>.<child>]
	// and deeper define its subunit subtree.
	if parts[0] == reservedTemplate {
		recordTemplateSubunit(parts, templateSubunitOrders, seenTemplateSubunits)

		return
	}

	recordHandAuthored(parts, unitOrder, subunitOrders, seenUnits, seenSubunits)
}

// recordUseSite records one [[use]]-family site in document order (Plan
// 35-02): a top-level [[use]] (empty Parent), a unit-nested
// [[unit.<path>.use]] whose Parent is the enclosing unit's dotted path (D-16),
// or a template-body [[template.<name>.<path>.use]] whose Template is the
// owning template and whose Parent is relative to the template's unit root
// (D-17).
func recordUseSite(parts []string, useSites *[]UseSite) {
	// minTemplateUseParts: template + name + use — a shorter template-path
	// form ([[template.use]]) names a template "use" as an array, which
	// extractTemplates rejects; do not record it as a site.
	const minTemplateUseParts = 3

	if len(parts) == 1 {
		// Top-level [[use]] / [use]: no enclosing path.
		*useSites = append(*useSites, UseSite{})

		return
	}

	if parts[0] == reservedTemplate {
		if len(parts) < minTemplateUseParts {
			return
		}

		*useSites = append(*useSites, UseSite{
			Template: parts[1],
			Parent:   strings.Join(parts[2:len(parts)-1], "."),
		})

		return
	}

	*useSites = append(*useSites, UseSite{
		Parent: strings.Join(parts[:len(parts)-1], "."),
	})
}

// extractKeyParts reads the dotted key segments from an unstable iterator.
// Returns nil for an empty key.
func extractKeyParts(keyIter unstable.Iterator) []string {
	var parts []string

	for keyIter.Next() {
		parts = append(parts, string(keyIter.Node().Data))
	}

	return parts
}

// recordTemplateSubunit records the subunit structure WITHIN a [template.*]
// namespace. For [template.svc.api] (parts: template, svc, api), records
// svc -> [api]. For deeper paths like [template.svc.api.handler], records
// svc.api -> [handler]. The template root itself ([template.svc], len==2) has
// no subunit to record — it is the root the extraction pass parses directly.
// Only paths with at least one segment after the template name are recorded.
func recordTemplateSubunit(
	parts []string,
	templateSubunitOrders map[string][]string,
	seen map[string]bool,
) {
	const minTemplateSubunitParts = 3 // template + name + child
	if len(parts) < minTemplateSubunitParts {
		return
	}

	// parts[0] == "template"; parts[1] is the template name; parts[2:] is the
	// subunit path within the template. parent is the path RELATIVE to the
	// template namespace (e.g. "svc" for [template.svc.api], "svc.api" for
	// [template.svc.api.handler]).
	parent := strings.Join(parts[1:len(parts)-1], ".")
	child := parts[len(parts)-1]
	key := parent + "." + child

	if !seen[key] {
		templateSubunitOrders[parent] = append(templateSubunitOrders[parent], child)
		seen[key] = true
	}
}

// recordHandAuthored records a hand-authored top-level unit ([name], len==1)
// or subunit ([parent.child], len==2) into the appropriate order slice.
// Deeper nesting (len > 2) is ignored — not supported for hand-authored units.
// unitOrder is passed by pointer because append may reallocate the slice.
func recordHandAuthored(
	parts []string,
	unitOrder *[]string,
	subunitOrders map[string][]string,
	seenUnits, seenSubunits map[string]bool,
) {
	const subunitParts = 2 // parent + child

	if len(parts) == 1 {
		// Top-level unit [name]
		name := parts[0]
		if !seenUnits[name] {
			*unitOrder = append(*unitOrder, name)
			seenUnits[name] = true
		}

		return
	}

	if len(parts) == subunitParts {
		// Subunit [parent.child]
		parent := parts[0]
		child := parts[1]
		key := parent + "." + child

		if !seenSubunits[key] {
			subunitOrders[parent] = append(subunitOrders[parent], child)
			seenSubunits[key] = true
		}
	}
	// Ignore deeper nesting (len(parts) > 2) - not supported
}

// parseUnitWithOrder parses a unit with explicit subunit order.
//
//nolint:gocognit,nestif,funlen // pre-existing; metrics surface only after Plan 31-01 grew the package
func parseUnitWithOrder(
	name string,
	value any,
	parentType model.UnitType,
	subunitOrder []string,
	subunitOrders map[string][]string,
) (*model.Unit, error) {
	unitMap, ok := value.(map[string]any)
	if !ok {
		return nil, &ParseError{
			Message: "invalid unit format",
			Context: name,
		}
	}

	// Re-marshal and unmarshal to get proper type conversion
	unitData, err := toml.Marshal(unitMap)
	if err != nil {
		return nil, &ParseError{Message: "failed to marshal unit", Context: name, Cause: err}
	}

	var unit model.Unit

	if err := toml.Unmarshal(unitData, &unit); err != nil {
		return nil, wrapDecodeError(err)
	}

	// Apply default type if not specified
	if unit.Type == "" {
		unit.Type = defaultTypeForParent(parentType)
	}

	// Infer level-specific type for generic types (db, queue)
	unit.Type = inferGenericType(unit.Type, parentType)

	// v1.10 ERGO-03/05: derive the display name from the identifier segment
	// when the author omits `name`. Explicit name = always wins (backward
	// compat). `name` is parseUnitWithOrder's first arg — already the last
	// path segment for both top-level units and nested subunits, so this one
	// hook covers both. Phase 31's XC-04 relocates this call to a
	// post-template-expansion pass so templated units humanize from their
	// substituted key, not the literal "${...}" segment.
	if unit.Name == "" {
		unit.Name = model.Humanize(name)
	}

	// Process subunits in the provided order
	if len(subunitOrder) > 0 {
		unit.Subunits = make(map[string]*model.Unit)
		unit.SubunitOrder = subunitOrder

		for _, subName := range subunitOrder {
			subVal, ok := unitMap[subName]
			if !ok {
				continue
			}

			// Get the subunit's own subunit order
			fullPath := name + "." + subName
			subSubunitOrder := subunitOrders[fullPath]

			subunit, err := parseUnitWithOrder(subName, subVal, unit.Type, subSubunitOrder, subunitOrders)
			if err != nil {
				return nil, err
			}

			unit.Subunits[subName] = subunit
		}
	} else {
		// Fallback: process subunits from raw map (no order guarantee, but maintains backward compatibility)
		for key, val := range unitMap {
			// Skip fields that are already in the Unit struct
			if isBuiltinField(key) {
				continue
			}

			// This must be a subunit
			subunit, err := parseUnitWithOrder(key, val, unit.Type, nil, subunitOrders)
			if err != nil {
				return nil, err
			}

			if unit.Subunits == nil {
				unit.Subunits = make(map[string]*model.Unit)
			}

			unit.Subunits[key] = subunit
			unit.SubunitOrder = append(unit.SubunitOrder, key)
		}
	}

	return &unit, nil
}

// defaultTypeForParent returns the default unit type based on parent type.
// - No parent (C1 level): system
// - Parent is system (C2 level): container
// - Parent is box (C1 level): system (C1 same-level grouping)
// - Parent is container (C3 level): component
// - Parent is containerBox (C2 level): container (C2 same-level grouping)
// - Parent is componentBox (C3 level): component (C3 same-level grouping).
func defaultTypeForParent(parentType model.UnitType) model.UnitType {
	//nolint:exhaustive // Default case handles all remaining types
	switch parentType {
	case "": // No parent = C1 level (root)
		return defaultTypeC1
	case model.TypeSystem:
		return defaultTypeC2 // C2 level default
	case model.TypeBox:
		return defaultTypeC1 // C1 same-level grouping
	case model.TypeContainer:
		return defaultTypeC3 // C3 level default
	case model.TypeContainerBox:
		return defaultTypeC2 // C2 same-level grouping
	case model.TypeComponentBox:
		return defaultTypeC3 // C3 same-level grouping
	default:
		// For other parent types (db, queue, etc.), default to C1
		return defaultTypeC1
	}
}

// inferGenericType transforms generic (level-agnostic) types to level-specific
// types based on the nesting level determined by parent type.
// Generic types are db, queue, and box; each resolves by parent:
//   - C1 (no parent or inside C1 box): db -> db, queue -> queue, box -> box
//   - C2 (inside system or containerBox): db -> containerDb, queue -> containerQueue,
//     box -> containerBox
//   - C3 (inside container or componentBox): db -> componentDb, queue -> componentQueue,
//     box -> componentBox.
//
// box is a universal grouping shorthand: writing type = "box" anywhere promotes
// it to the level-appropriate variant. Explicit containerBox/componentBox pass
// through unchanged (they are not TypeBox).
func inferGenericType(unitType model.UnitType, parentType model.UnitType) model.UnitType {
	// Only db, queue, and box are generic (level-agnostic) types.
	if !genericDbTypes[unitType] && unitType != model.TypeBox {
		return unitType // Not a generic type, return as-is
	}

	// Determine the nesting level and transform type accordingly
	//nolint:exhaustive // Default case handles all remaining types
	switch parentType {
	case "", model.TypeBox: // C1 level (root or C1 box)
		return unitType // db/queue/box unchanged at C1
	case model.TypeSystem, model.TypeContainerBox: // C2 level
		//nolint:exhaustive // Only db/queue/box are generic types, handled explicitly
		switch unitType {
		case model.TypeDb:
			return model.TypeContainerDb
		case model.TypeQueue:
			return model.TypeContainerQueue
		case model.TypeBox:
			return model.TypeContainerBox
		}
	case model.TypeContainer, model.TypeComponentBox: // C3 level
		//nolint:exhaustive // Only db/queue/box are generic types, handled explicitly
		switch unitType {
		case model.TypeDb:
			return model.TypeComponentDb
		case model.TypeQueue:
			return model.TypeComponentQueue
		case model.TypeBox:
			return model.TypeComponentBox
		}
	}

	// For other parent types, keep as-is (shouldn't happen in valid nesting)
	return unitType
}

// isBuiltinField returns true if the key is a known Unit struct field or a
// reserved unit-section key (the [[...use]] syntax extracted before unit
// parsing, Plan 35-02) — either way, never a subunit.
func isBuiltinField(key string) bool {
	return slices.Contains([]string{
		"type", "name", "description", "technology",
		"reference",
		"color", "style", "border", "edges",
		"width", "height", "expanded",
		"link", "linkFrom",
		"use",
	}, key)
}

// ParseFile reads a TOML file and parses it into a Model.
// It returns an error if the file cannot be read or parsed.
//
//nolint:gosec // G304: Path is provided by caller, this is intentional for CLI tool
func ParseFile(path string) (*Model, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &ParseError{
			Message: "failed to read file",
			Context: path,
			Cause:   err,
		}
	}

	return Parse(data)
}
