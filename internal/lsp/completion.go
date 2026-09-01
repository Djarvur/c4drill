// completion.go implements textDocument/completion for the c4drill TOML
// dialect (the .c4d counterpart lives in c4dlang.go). Context classes, per
// issue #32:
//
//   - unit `type =` — all 17 unit types, context-aware: the nesting default
//     (parser.DefaultTypeForParent) sorts first, generic db/queue/box show
//     their level promotion (parser.InferGenericType), everything else after;
//   - per-unit fields, per-link fields, reserved-table fields;
//   - enum values: edges, arrow, rank, kind, style, labelPosition;
//   - `peer =` — bare walk-up names (D-13/D-14/D-15 scopes) plus absolute
//     dotted paths from the current document;
//   - `template =` names and ${param} names in template bodies;
//   - `params = {` keys of the template chosen by the enclosing [[use]];
//   - `path =` in [[include]] — a filesystem scan relative to the including
//     file (INC-02).
//
// The analyzer works on the live buffer, so completion survives mid-edit
// states that no longer parse as a model.

package lsp

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Djarvur/c4drill/internal/model"
	"github.com/Djarvur/c4drill/internal/parser"
)

// completionAt is the textDocument/completion feature entry.
func (s *Server) completionAt(doc *document, pos Position) any {
	switch ext := filepath.Ext(doc.Path); ext {
	case extC4d:
		return s.c4dCompletion(doc, pos)
	case extToml:
		text := string(doc.Text)
		ctx := analyzeLine(text, pos)

		items := s.completionItems(doc, text, ctx)
		if len(items) == 0 {
			return CompletionList{Items: []CompletionItem{}}
		}

		return CompletionList{Items: items}
	default:
		return CompletionList{Items: []CompletionItem{}}
	}
}

// completionItems dispatches on the cursor context.
func (s *Server) completionItems(doc *document, text string, ctx lineContext) []CompletionItem {
	switch {
	case ctx.inTemplateRef:
		return templateParamItems(text, ctx, ctx.templateRefPrefix)
	case ctx.inValue:
		return s.valueCompletion(doc, text, ctx)
	case ctx.kind() == kindOther:
		return topLevelItems()
	default:
		return fieldItems(text, ctx)
	}
}

// valueCompletion produces items for the value position at the cursor.
func (s *Server) valueCompletion(doc *document, text string, ctx lineContext) []CompletionItem {
	switch ctx.key {
	case "type":
		return unitTypeItems(text, ctx)
	case "edges":
		return enumItems("edge routing", "straight", "spline", "square", "ortho")
	case "arrow":
		return enumItems("arrow direction", "forward", "reverse", "bidirectional", "none")
	case "rank":
		return enumItems("rank direction", "forward", "reverse", "equal")
	case "kind":
		return enumItems("link kind", "read", "write", "read-write")
	case "style":
		return enumItems("line style", "solid", "dashed", "dotted")
	case "labelPosition":
		return enumItems("label position", "middle", "tail", "head")
	case "peer":
		return peerItems(text, ctx)
	case "template":
		return templateNameItems(text)
	case "path":
		return s.includePathItems(doc, ctx.valuePrefix)
	case "params":
		return s.useParamKeyItems(text, ctx)
	default:
		return nil
	}
}

// unitFieldItems are the [unit] section keys in authoring order (model.Unit).
//
//nolint:gochecknoglobals // static completion table (root.go `version` precedent)
var unitFieldItems = []CompletionItem{
	{Label: "type", Kind: KindField, Documentation: "Unit type — 17 C1/C2/C3 variants"},
	{Label: "name", Kind: KindField, Documentation: "Display name"},
	{Label: "description", Kind: KindField, Documentation: "Short description"},
	{Label: "technology", Kind: KindField, Documentation: "Technology (not for person types)"},
	{Label: "reference", Kind: KindField, Documentation: "External documentation URL"},
	{Label: "color", Kind: KindField, Documentation: "Background color override"},
	{Label: "style", Kind: KindField, Documentation: "Line style: solid | dashed | dotted"},
	{Label: "border", Kind: KindField, Documentation: "Border color override"},
	{Label: "edges", Kind: KindField, Documentation: "Edge routing: straight | spline | square | ortho"},
	{Label: "width", Kind: KindField, Documentation: "Explicit width (0 = auto)"},
	{Label: "height", Kind: KindField, Documentation: "Explicit height (0 = auto)"},
	{Label: "expanded", Kind: KindField, Documentation: "Subunits expanded by default"},
	{Label: "link", Kind: KindField, InsertText: "link",
		Documentation: "Outgoing links ([[link]] table)"},
	{Label: "linkFrom", Kind: KindField, InsertText: "linkFrom",
		Documentation: "Incoming links ([[linkFrom]] table)"},
}

// linkFieldItems are the [[link]]/[[linkFrom]] entry keys (model.Link).
//
//nolint:gochecknoglobals // static completion table
var linkFieldItems = []CompletionItem{
	{Label: "peer", Kind: KindField, Documentation: "Target unit (bare name or dotted path)"},
	{Label: "arrow", Kind: KindField, Documentation: "forward | reverse | bidirectional | none"},
	{Label: "rank", Kind: KindField, Documentation: "forward | reverse | equal"},
	{Label: "kind", Kind: KindField, Documentation: "read | write | read-write"},
	{Label: "color", Kind: KindField, Documentation: "Link line color"},
	{Label: "style", Kind: KindField, Documentation: "solid | dashed | dotted"},
	{Label: "technology", Kind: KindField, Documentation: "Protocol or technology"},
	{Label: "description", Kind: KindField, Documentation: "Relationship description"},
	{Label: "labelPosition", Kind: KindField, Documentation: "middle | tail | head"},
	{Label: "length", Kind: KindField, Documentation: "Minimum edge length (rank spacing)"},
}

// propertiesFieldItems are the [properties] keys (model.Properties).
//
//nolint:gochecknoglobals // static completion table
var propertiesFieldItems = []CompletionItem{
	{Label: "name", Kind: KindField, Documentation: "Diagram name"},
	{Label: "description", Kind: KindField, Documentation: "Diagram description"},
	{Label: "color", Kind: KindField, Documentation: "Default background color"},
	{Label: "style", Kind: KindField, Documentation: "Default line style"},
	{Label: "border", Kind: KindField, Documentation: "Default border color"},
	{Label: "edges", Kind: KindField, Documentation: "Default edge routing"},
	{Label: "lineLength", Kind: KindField, Documentation: "Max label line length (0 = auto)"},
	{Label: "expanded", Kind: KindField, Documentation: "Units expanded by default"},
	{Label: "legend", Kind: KindField, Documentation: "Show the legend (default true)"},
	{Label: "legendLine", Kind: KindField, InsertText: "legendLine",
		Documentation: "Custom legend row ([[legendLine]]: label, color, style)"},
}

// includeFieldItems and useFieldItems are the [[include]]/[[use]] keys.
var (
	//nolint:gochecknoglobals // static completion table
	includeFieldItems = []CompletionItem{
		{Label: "path", Kind: KindField, Documentation: "File to merge, relative to this file's directory"},
		{Label: "once", Kind: KindField, Documentation: "Include at most once (INC-06)"},
	}

	//nolint:gochecknoglobals // static completion table
	useFieldItems = []CompletionItem{
		{Label: "template", Kind: KindField, Documentation: "Name of the [template.<name>] to instantiate"},
		{Label: "parent", Kind: KindField, Documentation: "Dotted placement path (empty = top level)"},
		{Label: "params", Kind: KindField, Documentation: "Named parameters, including name"},
	}
)

// templateFieldItems are the keys allowed inside [template.<name>]: the
// params array plus every unit field (the template body IS a unit subtree).
//
//nolint:gochecknoglobals // static completion table
var templateFieldItems = append(
	[]CompletionItem{{Label: "params", Kind: KindField, Documentation: "Declared parameter names"}},
	unitFieldItems...)

// fieldItems picks the field list for the cursor's table kind, filtering out
// scalar keys already authored in the same section.
func fieldItems(text string, ctx lineContext) []CompletionItem {
	var items []CompletionItem

	switch ctx.kind() {
	case kindUnit:
		items = unitFieldItems
	case kindLink:
		items = linkFieldItems
	case kindProperties:
		items = propertiesFieldItems
	case kindInclude:
		items = includeFieldItems
	case kindUse:
		items = useFieldItems
	case kindTemplate:
		items = templateFieldItems
	case kindOther:
		return nil
	}

	authored := sectionKeys(text, ctx)

	out := make([]CompletionItem, 0, len(items))
	for _, it := range items {
		if authored[it.Label] {
			continue
		}

		out = append(out, it)
	}

	return out
}

// sectionKeys collects the keys already authored in the cursor's section
// (excluding the cursor's own line — the author may be retyping it).
func sectionKeys(text string, ctx lineContext) map[string]bool {
	keys := map[string]bool{}

	current := ""

	for i, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)

		if isTableHeader(trimmed) {
			current = headerTablePath(trimmed)

			continue
		}

		if i == ctx.lineNo || current != ctx.tablePath {
			continue
		}

		if k, _, found := cutKeyValue(trimmed); found {
			keys[k] = true
		}
	}

	return keys
}

// topLevelItems suggests the reserved entry tables before any header.
func topLevelItems() []CompletionItem {
	return []CompletionItem{
		{Label: "[properties]", Kind: KindModule, InsertText: "properties",
			Documentation: "Global diagram properties"},
		{Label: "[[include]]", Kind: KindKeyword, InsertText: "include",
			Documentation: "Merge another file into this model"},
		{Label: "[[use]]", Kind: KindKeyword, InsertText: "use",
			Documentation: "Instantiate a template"},
		{Label: "[template.name]", Kind: KindKeyword, InsertText: "template.",
			Documentation: "Declare a unit template"},
	}
}

// enumItems builds keyword items for an enum value list.
func enumItems(detail string, values ...string) []CompletionItem {
	items := make([]CompletionItem, 0, len(values))
	for _, v := range values {
		items = append(items, CompletionItem{
			Label: v, Kind: KindEnumMember, InsertText: v, Detail: detail,
		})
	}

	return items
}

// allUnitTypes is the full 17-type C1/C2/C3 surface.
func allUnitTypes() []model.UnitType {
	return []model.UnitType{
		model.TypePerson, model.TypePersonExternal,
		model.TypeSystem, model.TypeSystemExternal,
		model.TypeDb, model.TypeDbExternal,
		model.TypeQueue, model.TypeQueueExternal,
		model.TypeBox,
		model.TypeContainer, model.TypeContainerDb,
		model.TypeContainerQueue, model.TypeContainerBox,
		model.TypeComponent, model.TypeComponentDb,
		model.TypeComponentQueue, model.TypeComponentBox,
	}
}

// unitTypeItems produces the 17 unit types, context-aware: the level default
// sorts first, generics note their promotion at this nesting, rest after.
func unitTypeItems(text string, ctx lineContext) []CompletionItem {
	return unitTypeItemsForParent(model.UnitType(declaredParentType(text, ctx)))
}

// unitTypeItemsForParent is the shared type-slot list for both dialects: the
// level default sorts first, generics note their promotion at this nesting.
func unitTypeItemsForParent(parentType model.UnitType) []CompletionItem {
	defaultType := parser.DefaultTypeForParent(parentType)

	items := make([]CompletionItem, 0, len(allUnitTypes()))

	for _, t := range allUnitTypes() {
		promoted := parser.InferGenericType(t, parentType)
		level := typeLevel(promoted)

		item := CompletionItem{
			Label:      string(t),
			Kind:       KindClass,
			InsertText: string(t),
			SortText:   "1" + string(t),
			Detail:     level + " unit",
		}

		switch {
		case promoted != t:
			item.Detail = level + " unit — promotes to " + string(promoted) + " here"
			item.SortText = "2" + string(t)
		case t == defaultType:
			item.Detail = level + " unit — default at this nesting level"
			item.SortText = "0" + string(t)
		}

		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool { return items[i].SortText < items[j].SortText })

	return items
}

// declaredParentType finds the declared type of the parent of the unit being
// typed: for table [a.b.c] that is the `type =` value in [a.b] ("" at root).
func declaredParentType(text string, ctx lineContext) string {
	host := ctx.hostUnitPath()
	if host == "" {
		return ""
	}

	dot := strings.LastIndexByte(host, '.')
	if dot < 0 {
		return "" // top-level unit: no parent
	}

	return declaredTypes(text)[host[:dot]]
}

// typeLevel labels a unit type's C4 level from its name family. The
// default branch covers every remaining (C1) type deliberately.
//
//nolint:exhaustive // C1 types fall through to the default by design
func typeLevel(t model.UnitType) string {
	switch t {
	case model.TypeContainer, model.TypeContainerDb, model.TypeContainerQueue, model.TypeContainerBox:
		return "C2"
	case model.TypeComponent, model.TypeComponentDb, model.TypeComponentQueue, model.TypeComponentBox:
		return "C3"
	default:
		return "C1"
	}
}

// peerItems completes `peer =`: bare walk-up names across the host's
// ancestor scopes (D-13/D-14/D-15) plus every absolute dotted path.
func peerItems(text string, ctx lineContext) []CompletionItem {
	return peerItemsFromPaths(unitPaths(scanHeaders(text)), ctx.hostUnitPath())
}

// peerItemsFromPaths is the shared peer-target list for both dialects: bare
// candidates are the direct children of each ancestor scope of the host —
// precisely the scopes peer.Resolve searches, nearest-first — plus every
// absolute path (the host itself excluded).
func peerItemsFromPaths(paths []string, host string) []CompletionItem {
	if len(paths) == 0 {
		return nil
	}

	hostSegments := strings.Split(host, ".")

	bare := map[string]bool{}

	for i := len(hostSegments) - 1; i >= 0; i-- {
		for _, p := range paths {
			if scopeChildName(p, hostSegments[:i]) != "" {
				bare[scopeChildName(p, hostSegments[:i])] = true
			}
		}
	}

	items := make([]CompletionItem, 0, len(bare)+len(paths))

	for _, name := range sortedSet(bare) {
		items = append(items, CompletionItem{
			Label: name, Kind: KindValue, InsertText: name,
			Detail: "bare peer (walk-up resolution)",
		})
	}

	for _, p := range paths {
		if p == host {
			continue
		}

		items = append(items, CompletionItem{
			Label: p, Kind: KindValue, InsertText: p, FilterText: p,
			Detail: "absolute peer path",
		})
	}

	sort.Slice(items, func(i, j int) bool { return items[i].Label < items[j].Label })

	return items
}

// scopeChildName returns path's last segment when path is a DIRECT child of
// the scope segments, else "" (a unit name usable as a bare peer there).
func scopeChildName(path string, scope []string) string {
	parts := strings.Split(path, ".")
	if len(parts) != len(scope)+1 {
		return ""
	}

	for i, s := range scope {
		if parts[i] != s {
			return ""
		}
	}

	return parts[len(scope)]
}

// templateNameItems completes `template =` with the declared template names.
func templateNameItems(text string) []CompletionItem {
	seen := map[string]bool{}

	for _, h := range scanHeaders(text) {
		if h.isArray {
			continue
		}

		segments := strings.Split(h.path, ".")
		if len(segments) >= 2 && segments[0] == tblTemplate {
			seen[segments[1]] = true
		}
	}

	names := sortedSet(seen)
	if len(names) == 0 {
		return nil
	}

	items := make([]CompletionItem, 0, len(names))
	for _, n := range names {
		items = append(items, CompletionItem{Label: n, Kind: KindClass, InsertText: n, Detail: "template"})
	}

	sort.Slice(items, func(i, j int) bool { return items[i].Label < items[j].Label })

	return items
}

// templateParams returns the declared params of template name — the
// `params = [...]` value authored directly in [template.<name>].
func templateParams(text, name string) []string {
	target := tblTemplate + "." + name

	current := ""
	params := []string{}

	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]") {
			current = normalizeTablePath(strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]"))

			continue
		}

		if current != target {
			continue
		}

		if k, v, found := cutKeyValue(trimmed); found && k == "params" {
			params = parseStringArray(v)

			break
		}
	}

	return params
}

// templateParamItems completes inside a ${...} token: the union of every
// template's declared parameters (each item names its owning template).
func templateParamItems(text string, ctx lineContext, prefix string) []CompletionItem {
	found := declaredTemplateParams(text, prefix)

	items := make([]CompletionItem, 0, len(found))
	for _, p := range found {
		detail := "parameter of template " + p.template
		if enclosingTemplate(ctx) == p.template {
			detail += " (this template)"
		}

		items = append(items, CompletionItem{
			Label: p.name, Kind: KindValue, InsertText: p.name, Detail: detail,
		})
	}

	sort.Slice(items, func(i, j int) bool { return items[i].Label < items[j].Label })

	return items
}

// templateParam names one declared template parameter.
type templateParam struct {
	name     string
	template string
}

// declaredTemplateParams collects the document's declared template
// parameters matching prefix, walking [template.<name>] sections.
func declaredTemplateParams(text, prefix string) []templateParam {
	var found []templateParam

	seen := map[string]bool{}

	current := ""

	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)

		if isTableHeader(trimmed) {
			current = headerTablePath(trimmed)

			continue
		}

		name, ok := templateOfTable(current)
		if !ok {
			continue
		}

		found = appendTemplateParams(found, seen, name, trimmed, prefix)
	}

	return found
}

// isTableHeader reports whether the trimmed line opens a table section.
func isTableHeader(trimmed string) bool {
	return strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")
}

// headerTablePath extracts a trimmed table header line's dotted path,
// accepting both [table] and [[array-of-tables]] spellings.
func headerTablePath(trimmed string) string {
	inner := strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]")
	inner = strings.TrimSuffix(strings.TrimPrefix(inner, "["), "]")

	return normalizeTablePath(inner)
}

// appendTemplateParams appends a params array's new prefix-matching entries.
func appendTemplateParams(
	found []templateParam,
	seen map[string]bool,
	name, trimmed, prefix string,
) []templateParam {
	k, v, isKV := cutKeyValue(trimmed)
	if !isKV || k != "params" {
		return found
	}

	for _, p := range parseStringArray(v) {
		if seen[name+":"+p] || !strings.HasPrefix(p, prefix) {
			continue
		}

		seen[name+":"+p] = true

		found = append(found, templateParam{name: p, template: name})
	}

	return found
}

// templateOfTable returns the <name> when the table path is a
// [template.<name>] subtree root or member.
func templateOfTable(tablePath string) (string, bool) {
	segments := strings.Split(tablePath, ".")
	if len(segments) >= 2 && segments[0] == tblTemplate {
		return segments[1], true
	}

	return "", false
}

// enclosingTemplate names the [template.<x>] the cursor sits in, if any.
func enclosingTemplate(ctx lineContext) string {
	name, _ := templateOfTable(ctx.tablePath)

	return name
}

// useParamKeyItems completes the keys inside a [[use]] entry's inline
// `params = {` table: the chosen template's declared parameters.
func (s *Server) useParamKeyItems(text string, ctx lineContext) []CompletionItem {
	if ctx.kind() != kindUse {
		return nil
	}

	name := nearestKeyValue(text, ctx, "template")
	if name == "" {
		return nil
	}

	params := templateParams(text, name)
	if len(params) == 0 {
		return nil
	}

	items := make([]CompletionItem, 0, len(params))
	for _, p := range params {
		items = append(items, CompletionItem{
			Label: p, Kind: KindField, InsertText: p,
			Detail: "parameter of template " + name,
		})
	}

	sort.Slice(items, func(i, j int) bool { return items[i].Label < items[j].Label })

	return items
}

// nearestKeyValue scans upward from the cursor for `key = "value"` within
// the enclosing table section (finds a use entry's template name).
func nearestKeyValue(text string, ctx lineContext, key string) string {
	lines := strings.Split(text, "\n")

	if ctx.lineNo >= len(lines) {
		return ""
	}

	for i := ctx.lineNo; i >= 0; i-- {
		trimmed := strings.TrimSpace(lines[i])

		if isTableHeader(trimmed) && i != ctx.lineNo {
			return "" // left the section
		}

		if k, v, found := cutKeyValue(trimmed); found && k == key {
			return unquote(v)
		}
	}

	return ""
}

// includePathItems completes `path =` in [[include]] by scanning the
// filesystem relative to the including document's directory (INC-02):
// diagram files (.toml/.c4d) and directories, one hop per request.
func (s *Server) includePathItems(doc *document, typed string) []CompletionItem {
	dir := filepath.Dir(doc.Path)
	sub := ""

	if i := strings.LastIndexByte(typed, '/'); i >= 0 {
		sub = typed[:i+1]
	}

	entries, err := os.ReadDir(filepath.Join(dir, sub))
	if err != nil {
		return nil
	}

	items := make([]CompletionItem, 0, len(entries))
	for _, e := range entries {
		name := e.Name()

		switch {
		case e.IsDir():
			items = append(items, CompletionItem{
				Label: name + "/", Kind: KindFolder,
				InsertText: sub + name + "/",
				FilterText: sub + name + "/",
				Detail:     "directory",
			})
		case filepath.Ext(name) == extToml || filepath.Ext(name) == extC4d:
			items = append(items, CompletionItem{
				Label: name, Kind: KindFile,
				InsertText: sub + name,
				FilterText: sub + name,
				Detail:     "diagram file",
			})
		}
	}

	sort.Slice(items, func(i, j int) bool { return items[i].Label < items[j].Label })

	return items
}

// parseStringArray parses a one-line TOML string array literal (the
// `params = ["a", "b"]` shape — the c4drill authoring norm).
func parseStringArray(v string) []string {
	v = strings.TrimSpace(v)
	if !strings.HasPrefix(v, "[") || !strings.HasSuffix(v, "]") {
		return nil
	}

	inner := strings.TrimSuffix(strings.TrimPrefix(v, "["), "]")

	out := []string{}

	for _, part := range strings.Split(inner, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, unquote(part))
		}
	}

	return out
}

// sortedSet returns the set's members sorted.
func sortedSet(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}

	sort.Strings(out)

	return out
}
