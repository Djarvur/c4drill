// c4dlang.go implements the c4drill C4D language features over the c4dctx
// scanner (issue #33): textDocument/completion, hover, definition, and
// documentSymbol for .c4d documents. The classes mirror the TOML-dialect
// providers completion.go/hover.go/definition.go/symbols.go implement:
//
//   - completion — unit header type slots with nesting-aware promotion,
//     unit/properties/edge-option field keywords, enum values, edge peer
//     targets (bare walk-up names + absolute dotted paths, the TOML
//     semantics), `properties`/`template`/`use`/`include` statement
//     keywords, use-argument keys of the named template, and ${param}
//     tokens in template bodies;
//   - hover — peer references (resolved absolute path, level, promoted
//     type, through the merged model), template references, ${param} tokens;
//   - definition — peer references to the unit header (this document or the
//     include closure, both formats), use references to the template
//     declaration, include paths to the target file;
//   - documentSymbol — the hierarchical unit outline with brace-block ranges.
//
// Structural analysis is the scanner's (c4dctx.go); every semantic question
// — peer resolution, promoted types, merged models — goes through the
// internal/c4d front-end and the shared pipeline, never re-implemented here.

package lsp

import (
	"path/filepath"
	"strings"

	"github.com/Djarvur/c4drill/internal/model"
)

// --- completion ----------------------------------------------------------

// c4dCompletion is the textDocument/completion feature entry for .c4d.
func (s *Server) c4dCompletion(doc *document, pos Position) any {
	cur := c4dAnalyze(string(doc.Text), pos)

	items := s.c4dCompletionItems(doc, cur)
	if items == nil {
		items = []CompletionItem{}
	}

	return CompletionList{Items: items}
}

// c4dCompletionItems dispatches on the cursor's statement kind.
func (s *Server) c4dCompletionItems(doc *document, cur *c4dCursor) []CompletionItem {
	switch cur.kind {
	case c4dStmtTemplateRef:
		return c4dTemplateParamItems(cur)
	case c4dStmtTypeSlot:
		return unitTypeItemsForParent(model.UnitType(cur.hostUnitType()))
	case c4dStmtEdgePeer:
		return peerItemsFromPaths(cur.scopeUnitPaths(), cur.hostPath)
	case c4dStmtUseName:
		return c4dTemplateNameItems(cur)
	case c4dStmtUseArgKey:
		return c4dUseParamKeyItems(cur)
	case c4dStmtIncludePath:
		return s.includePathItems(doc, cur.includePrefix)
	case c4dStmtFieldValue:
		return c4dFieldValueItems(cur)
	case c4dStmtStart:
		return c4dStatementStartItems(cur)
	case c4dStmtUseArgValue, c4dStmtOther:
		return nil
	}

	return nil
}

// c4dStatementStartItems suggests the statements legal at the cursor's block
// level: reserved statement keywords at document level, field keywords /
// edge arrows / nested unit headers inside bodies, and per-block key lists.
func c4dStatementStartItems(cur *c4dCursor) []CompletionItem {
	blk := cur.innermost()

	if blk == nil {
		return c4dTopLevelItems()
	}

	switch blk.kind {
	case c4dBlockProperties:
		return filterAuthored(c4dPropertyKeyItems, blk.fields)
	case c4dBlockEdgeOptions:
		return filterAuthored(c4dOptionKeyItems, blk.fields)
	case c4dBlockUnit, c4dBlockTemplate, c4dBlockOther:
		return c4dBodyStatementItems(cur, blk)
	}

	return nil
}

// c4dTopLevelItems suggests the reserved top-level statements plus the
// type-led unit headers (nesting-aware: no parent at document level).
func c4dTopLevelItems() []CompletionItem {
	items := make([]CompletionItem, 0, 21)

	items = append(items,
		CompletionItem{Label: "properties {", Kind: KindModule, InsertText: "properties {",
			Documentation: "Global diagram properties"},
		CompletionItem{Label: "include", Kind: KindKeyword, InsertText: "include ",
			Documentation: "Merge another file into this model"},
		CompletionItem{Label: "template", Kind: KindKeyword, InsertText: "template ",
			Documentation: "Declare a parametrized unit template"},
		CompletionItem{Label: "use", Kind: KindKeyword, InsertText: "use ",
			Documentation: "Instantiate a template"},
	)

	return append(items, unitTypeItemsForParent("")...)
}

// c4dBodyStatementItems suggests what a unit or template body accepts: the
// unit field keywords (minus the authored ones), the edge arrows, `use`,
// and type-led nested unit headers.
func c4dBodyStatementItems(cur *c4dCursor, blk *c4dBlock) []CompletionItem {
	items := filterAuthored(c4dUnitFieldItems, blk.fields)

	items = append(items,
		CompletionItem{Label: "->", Kind: KindKeyword, InsertText: "-> ",
			Documentation: "Outgoing edge"},
		CompletionItem{Label: "<->", Kind: KindKeyword, InsertText: "<-> ",
			Documentation: "Bidirectional edge"},
		CompletionItem{Label: "<-", Kind: KindKeyword, InsertText: "<- ",
			Documentation: "Incoming edge (linkFrom)"},
		CompletionItem{Label: "--", Kind: KindKeyword, InsertText: "-- ",
			Documentation: "Plain edge (no arrowhead)"},
		CompletionItem{Label: "use", Kind: KindKeyword, InsertText: "use ",
			Documentation: "Instantiate a template"},
	)

	if blk.kind == c4dBlockTemplate {
		items = append(items, CompletionItem{Label: "type:", Kind: KindField, InsertText: "type: ",
			Documentation: "Template root type (the TOML [template.x] type key twin)"})
	}

	// Type-led nested unit headers: promotion is computed against the
	// enclosing unit's declared type.
	items = append(items, unitTypeItemsForParent(model.UnitType(cur.hostUnitType()))...)

	return items
}

// c4dUnitFieldItems are the unit body field keywords in canonical (D-23)
// authoring order.
//
//nolint:gochecknoglobals // static completion table (unitFieldItems precedent)
var c4dUnitFieldItems = []CompletionItem{
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
}

// c4dPropertyKeyItems are the properties { } keys in canonical order.
//
//nolint:gochecknoglobals // static completion table
var c4dPropertyKeyItems = []CompletionItem{
	{Label: "name", Kind: KindField, Documentation: "Diagram name"},
	{Label: "description", Kind: KindField, Documentation: "Diagram description"},
	{Label: "color", Kind: KindField, Documentation: "Default background color"},
	{Label: "style", Kind: KindField, Documentation: "Default line style"},
	{Label: "border", Kind: KindField, Documentation: "Default border color"},
	{Label: "edges", Kind: KindField, Documentation: "Default edge routing"},
	{Label: "lineLength", Kind: KindField, Documentation: "Max label line length (0 = auto)"},
	{Label: "expanded", Kind: KindField, Documentation: "Units expanded by default"},
	{Label: "legend", Kind: KindField, Documentation: "Show the legend: true | false (default true)"},
	{Label: "legendLine", Kind: KindField, Documentation: "Custom legend row: \"label|color[|style]\""},
}

// c4dOptionKeyItems are the edge option-block keys (c4d.peg OptionKey).
//
//nolint:gochecknoglobals // static completion table
var c4dOptionKeyItems = []CompletionItem{
	{Label: "arrow", Kind: KindField, Documentation: "forward | reverse | bidirectional | none"},
	{Label: "rank", Kind: KindField, Documentation: "forward | reverse | equal"},
	{Label: "kind", Kind: KindField, Documentation: "read | write | read-write"},
	{Label: "color", Kind: KindField, Documentation: "Link line color"},
	{Label: "style", Kind: KindField, Documentation: "solid | dashed | dotted"},
	{Label: "labelPosition", Kind: KindField, Documentation: "middle | tail | head"},
	{Label: "length", Kind: KindField, Documentation: "Minimum edge length (rank spacing)"},
}

// filterAuthored drops already-authored keys (duplicate field keys are hard
// errors in both dialects) and returns a fresh slice.
func filterAuthored(items []CompletionItem, authored []string) []CompletionItem {
	out := make([]CompletionItem, 0, len(items))

	for _, it := range items {
		taken := false

		for _, k := range authored {
			if k == it.Label {
				taken = true

				break
			}
		}

		if !taken {
			out = append(out, it)
		}
	}

	return out
}

// c4dFieldValueItems completes the enum-ish field values; freeform keys
// complete nothing (the same shape as the TOML valueCompletion).
func c4dFieldValueItems(cur *c4dCursor) []CompletionItem {
	switch cur.key {
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
	case "legend":
		return enumItems("show the legend", "true", "false")
	default:
		return nil
	}
}

// c4dTemplateNameItems completes the `use <cursor>(` slot with the declared
// template names.
func c4dTemplateNameItems(cur *c4dCursor) []CompletionItem {
	items := make([]CompletionItem, 0, len(cur.inventory.tmpls))

	for _, t := range cur.inventory.tmpls {
		items = append(items, CompletionItem{
			Label: t.id, Kind: KindClass, InsertText: t.id,
			Detail: "template (" + strings.Join(t.params, ", ") + ")",
		})
	}

	return items
}

// c4dUseParamKeyItems completes a use statement's argument keys with the
// named template's declared parameters.
func c4dUseParamKeyItems(cur *c4dCursor) []CompletionItem {
	if cur.useName == "" {
		return nil
	}

	params := cur.templateParams(cur.useName)
	if len(params) == 0 {
		return nil
	}

	items := make([]CompletionItem, 0, len(params))
	for _, p := range params {
		items = append(items, CompletionItem{
			Label: p, Kind: KindField, InsertText: p,
			Detail: "parameter of template " + cur.useName,
		})
	}

	return items
}

// c4dTemplateParamItems completes inside a ${...} token: the union of every
// template's declared parameters, each naming its owning template.
func c4dTemplateParamItems(cur *c4dCursor) []CompletionItem {
	items := make([]CompletionItem, 0)

	for _, t := range cur.inventory.tmpls {
		for _, p := range t.params {
			if !strings.HasPrefix(p, cur.valuePrefix) {
				continue
			}

			detail := "parameter of template " + t.id
			if t.id == cur.tmplName {
				detail += " (this template)"
			}

			items = append(items, CompletionItem{
				Label: p, Kind: KindValue, InsertText: p, Detail: detail,
			})
		}
	}

	return items
}

// --- hover ---------------------------------------------------------------

// c4dHover is the textDocument/hover feature entry for .c4d.
func (s *Server) c4dHover(doc *document, pos Position) any {
	cur := c4dAnalyze(string(doc.Text), pos)

	switch cur.kind { //nolint:exhaustive // the remaining kinds hover nothing
	case c4dStmtTemplateRef:
		if cur.tmplName == "" {
			return nil
		}

		return templateParamListHover(cur.tmplName, cur.templateParams(cur.tmplName))
	case c4dStmtEdgePeer:
		return s.c4dPeerHover(doc, cur, pos)
	case c4dStmtUseName:
		return c4dTemplateRefHover(cur)
	default:
		return nil
	}
}

// c4dPeerHover resolves the edge's peer the way peer.Resolve would
// (read-only) and reports the target's absolute path, level, and type.
func (s *Server) c4dPeerHover(doc *document, cur *c4dCursor, pos Position) *Hover {
	if cur.tmplName != "" || cur.fullPeer == "" {
		return nil // peers inside template bodies resolve at expansion time
	}

	m := s.mergedModel(doc)
	if m == nil {
		return nil // a buffer that does not parse has nothing resolvable
	}

	target := resolvePeerReadOnly(m, cur.hostPath, cur.fullPeer)
	if target == "" {
		return nil
	}

	unit := findUnit(m, target)
	if unit == nil {
		return nil
	}

	text := string(doc.Text)

	return &Hover{
		Contents: MarkupContent{Kind: "markdown", Value: peerHoverMarkdown(target, unit)},
		Range:    wordRangeAt(text, pos),
	}
}

// c4dTemplateRefHover reports the referenced template's parameter info at a
// use statement's template-name slot.
func c4dTemplateRefHover(cur *c4dCursor) *Hover {
	name := c4dUseNameOnLine(cur.line)
	if name == "" {
		return nil
	}

	params := cur.templateParams(name)
	if params == nil {
		return nil // undeclared template: nothing to report
	}

	return templateRefListHover(name, params)
}

// c4dUseNameOnLine extracts the template-name token of a `use name(...)`
// statement from the statement's line (the cursor may sit mid-token, so the
// partial prefix is not enough to resolve).
func c4dUseNameOnLine(line string) string {
	_, rest := c4dFirstToken(line) // drop `use`
	rest = strings.TrimSpace(rest)

	if i := strings.IndexByte(rest, '('); i >= 0 {
		rest = rest[:i]
	}

	rest = strings.TrimSpace(rest)
	if !c4dIsIdent(rest) {
		return ""
	}

	return rest
}

// --- definition ----------------------------------------------------------

// c4dDefinition is the textDocument/definition feature entry for .c4d.
func (s *Server) c4dDefinition(doc *document, pos Position) any {
	cur := c4dAnalyze(string(doc.Text), pos)

	var loc *Location

	switch cur.kind { //nolint:exhaustive // the remaining kinds define nothing
	case c4dStmtEdgePeer:
		loc = s.c4dPeerDefinition(doc, cur)
	case c4dStmtUseName:
		loc = s.c4dTemplateDefinition(doc, cur)
	case c4dStmtIncludePath:
		loc = s.c4dIncludeDefinition(doc, cur)
	default:
		return nil
	}

	if loc == nil {
		return nil
	}

	return []Location{*loc}
}

// c4dPeerDefinition resolves the edge peer and points at the unit header —
// in this document first, then through the include closure (both formats).
func (s *Server) c4dPeerDefinition(doc *document, cur *c4dCursor) *Location {
	if cur.tmplName != "" || cur.fullPeer == "" {
		return nil
	}

	m := s.mergedModel(doc)
	if m == nil {
		return nil
	}

	target := resolvePeerReadOnly(m, cur.hostPath, cur.fullPeer)
	if target == "" {
		return nil
	}

	return s.searchIncludesFrom(doc.Path, string(doc.Text), defTarget{unitPath: target}, map[string]bool{})
}

// c4dTemplateDefinition jumps from a use statement to its template
// declaration.
func (s *Server) c4dTemplateDefinition(doc *document, cur *c4dCursor) *Location {
	name := c4dUseNameOnLine(cur.line)
	if name == "" {
		return nil
	}

	return s.searchIncludesFrom(doc.Path, string(doc.Text), defTarget{template: name}, map[string]bool{})
}

// c4dIncludeDefinition jumps from an include directive to the target file.
func (s *Server) c4dIncludeDefinition(doc *document, cur *c4dCursor) *Location {
	path := c4dIncludePathOnLine(cur.line)
	if path == "" {
		return nil
	}

	target := filepath.Join(filepath.Dir(doc.Path), path)
	if !fileExists(target) {
		return nil
	}

	return &Location{URI: pathToURI(target), Range: Range{}}
}

// c4dIncludePathOnLine extracts the full include path from the directive's
// line: quoted paths ride verbatim, barewords end at whitespace (the `once`
// modifier is a separate token).
func c4dIncludePathOnLine(line string) string {
	_, rest := c4dFirstToken(line) // drop `include`
	rest = strings.TrimSpace(rest)

	if strings.HasPrefix(rest, `"`) {
		if end := strings.Index(rest[1:], `"`); end >= 0 {
			return rest[1 : 1+end]
		}

		return stripValueQuotes(rest) // still-open quote mid-edit
	}

	path, _ := c4dFirstToken(rest)

	return path
}

// --- documentSymbol ------------------------------------------------------

// c4dDocumentSymbols is the documentSymbol feature entry for .c4d: the unit
// blocks as a nested outline, closed at each block's own brace.
func (s *Server) c4dDocumentSymbols(doc *document) []DocumentSymbol {
	sc := c4dScanDocument(string(doc.Text))

	type frame struct {
		path string
		sym  *DocumentSymbol
	}

	symbols := []DocumentSymbol{}

	var stack []frame

	for _, u := range sc.units {
		if u.tmpl != "" {
			continue // template bodies are declarations, not document units
		}

		// The parent chain is exact-path matched: unit keys may contain dots
		// (a quoted display name becomes the key, per the parser's unitKey),
		// so nesting is never derived by splitting on '.'.
		for len(stack) > 0 && stack[len(stack)-1].path != u.parent {
			stack = stack[:len(stack)-1]
		}

		sym := c4dSymbolFor(u)

		switch {
		case len(stack) == 0:
			symbols = append(symbols, sym)
			stack = append(stack, frame{path: u.path, sym: &symbols[len(symbols)-1]})
		default:
			parent := stack[len(stack)-1].sym
			parent.Children = append(parent.Children, sym)
			stack = append(stack, frame{
				path: u.path,
				sym:  &parent.Children[len(parent.Children)-1],
			})
		}
	}

	if len(symbols) == 0 {
		return nil
	}

	return symbols
}

// c4dSymbolFor builds one unit symbol with its detail line.
func c4dSymbolFor(u *c4dBlock) DocumentSymbol {
	detail := u.typ
	if u.external {
		detail = strings.TrimSpace(detail + " " + kwExternal)
	}

	if u.name != "" {
		detail = u.name + " (" + detail + ")"
	}

	ln := uint32(u.line) //nolint:gosec // line indices are non-negative

	end := u.endLine
	if end < 0 {
		end = u.line
	}

	return DocumentSymbol{
		Name:           u.key,
		Detail:         detail,
		Kind:           KindObject,
		Range:          Range{Start: Position{Line: ln}, End: Position{Line: uint32(end)}}, //nolint:gosec,lll // line indices are non-negative
		SelectionRange: Range{Start: Position{Line: ln}, End: Position{Line: ln}},
	}
}

// --- defTarget: format-aware definition lookup ---------------------------

// defTarget is a definition lookup: a unit path or a template name.
type defTarget struct {
	unitPath string
	template string
}

// findInToml locates the target's header line in TOML text.
func (t defTarget) findInToml(text string) *Range {
	path := t.unitPath
	if path == "" {
		path = tblTemplate + "." + t.template
	}

	return headerRange(scanHeaders(text), path, false)
}

// findInC4D locates the target's line in C4D text.
func (t defTarget) findInC4D(text string) *Range {
	sc := c4dScanDocument(text)

	line := -1

	switch {
	case t.unitPath != "":
		line = sc.unitTargetLine(t.unitPath)
	case t.template != "":
		line = sc.templateTargetLine(t.template)
	}

	if line < 0 {
		return nil
	}

	r := Range{
		Start: Position{Line: uint32(line)},
		End:   Position{Line: uint32(line)},
	}

	r.End.Character = 1 // represent a non-empty selection span

	return &r
}
