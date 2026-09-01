package render

import (
	"strings"

	"github.com/Djarvur/c4drill/internal/graph"
	"github.com/Djarvur/c4drill/internal/model"
)

// plantUMLStdlibBase is the base URL the generated .puml files !include the
// C4-PlantUML standard library from (issue #25). It is a package constant —
// not a [properties] field — because threading a config value through
// parser -> view -> graph -> render would touch every layer for a URL that
// only changes when the whole toolchain moves; anyone needing a local copy
// can post-process the single !include line.
const plantUMLStdlibBase = "https://raw.githubusercontent.com/plantuml-stdlib/C4-PlantUML/master"

// The three C4-PlantUML entry points. Each include transitively pulls in the
// levels below it (C4_Component -> C4_Container -> C4_Context -> C4), so the
// renderer only ever emits ONE include line: the deepest level the diagram
// needs.
const (
	pumlContextInclude   = plantUMLStdlibBase + "/C4_Context.puml"
	pumlContainerInclude = plantUMLStdlibBase + "/C4_Container.puml"
	pumlComponentInclude = plantUMLStdlibBase + "/C4_Component.puml"
)

// Element macro names per c4drill unit type (issue #25 mapping table).
// Person/System-level macros take (alias, label, descr, sprite, tags, link);
// Container/Component-level macros take (alias, label, techn, descr, sprite,
// tags, link) — hasTechn records which dialect a type uses. The generic
// default is System so hand-built or unknown-typed nodes still serialize.
func plantUMLElementMacro(t model.UnitType) (string, bool) {
	switch t {
	case model.TypePerson:
		return "Person", false
	case model.TypePersonExternal:
		return "Person_Ext", false
	case model.TypeSystem:
		return "System", false
	case model.TypeSystemExternal:
		return "System_Ext", false
	case model.TypeDb:
		return "SystemDb", false
	case model.TypeDbExternal:
		return "SystemDb_Ext", false
	case model.TypeQueue:
		return "SystemQueue", false
	case model.TypeQueueExternal:
		return "SystemQueue_Ext", false
	case model.TypeContainer:
		return "Container", true
	case model.TypeContainerDb:
		return "ContainerDb", true
	case model.TypeContainerQueue:
		return "ContainerQueue", true
	case model.TypeComponent:
		return "Component", true
	case model.TypeComponentDb:
		return "ComponentDb", true
	case model.TypeComponentQueue:
		return "ComponentQueue", true
	case model.TypeBox:
		// A COLLAPSED box renders as an element node (like the DOT
		// renderer's record node), not as a boundary frame — boundaries
		// are only emitted for clusters (see plantUMLBoundaryMacro).
		return "System", false
	case model.TypeContainerBox:
		return "Container", true
	case model.TypeComponentBox:
		return "Component", true
	default:
		return "System", false
	}
}

// plantUMLBoundaryMacro maps a cluster's unit type onto a boundary macro.
// Systems (and C1 boxes) frame with System_Boundary, containers (and
// containerBox) with Container_Boundary; the stdlib has NO Component_Boundary,
// so componentBox (and the type-less ancestor wrapper clusters) fall back to
// the generic Boundary(alias, label, type, ...) macro — "Component" for
// component-level frames so the boundary still reads as one, "" for wrappers.
func plantUMLBoundaryMacro(t model.UnitType) (string, string) {
	switch t {
	case model.TypeSystem, model.TypeBox:
		return "System_Boundary", ""
	case model.TypeContainer, model.TypeContainerBox:
		return "Container_Boundary", ""
	case model.TypeComponent, model.TypeComponentBox:
		return "Boundary", "Component"
	case model.TypePerson, model.TypePersonExternal,
		model.TypeSystemExternal, model.TypeDb, model.TypeDbExternal,
		model.TypeQueue, model.TypeQueueExternal,
		model.TypeContainerDb, model.TypeContainerQueue,
		model.TypeComponentDb, model.TypeComponentQueue:
		// Leaf entities are never grouping frames; a cluster carrying one
		// (hand-built graphs only) still falls back to the generic boundary.
		return "Boundary", ""
	default:
		return "Boundary", ""
	}
}

// RenderPlantUML renders a graph as a C4-PlantUML source document (issue #25):
// the standard C4_Context/C4_Container/C4_Component macros for units,
// Rel/Rel_Back/BiRel (and the arrowless Rel_ "--" variant) for edges, and
// clickable drill-down links carried in every element macro's $link parameter
// — C4-PlantUML expands those into `[[url]]` markup, so `plantuml -tsvg`
// produces SVGs whose elements are real anchors. Link targets are the
// sibling .svg files ComputeExploreURL computes (the path.go Gap-3 contract:
// navigation always points at the browser-navigable .svg layout), and the
// single-URL-slot precedence carries over: drill-down beats the external
// reference, exactly like the converter's SetURL.
//
// The breadcrumb nav bar is deliberately omitted: GraphViz HTML-like labels
// have no PlantUML equivalent (issue #25 defers it). The legend is emitted as
// LAYOUT_WITH_LEGEND() only when the graph carries one (properties.legend).
//
// Output is pure text — no PlantUML tooling is invoked.
//
//nolint:revive // Function name matches the RenderSVG/RenderPNG pattern
func RenderPlantUML(g *graph.Graph) ([]byte, error) {
	if g == nil {
		return nil, ErrNilGraph
	}

	w := newPlantUMLWriter(g)

	return w.render(), nil
}

// plantUMLWriter serializes one graph into C4-PlantUML source. Elements and
// relationships are rendered into separate buffers because C4-PlantUML
// requires AddElementTag/AddRelTag declarations to appear BEFORE their first
// use, while the tags themselves are only discovered while serializing the
// elements they style.
type plantUMLWriter struct {
	g *graph.Graph

	aliases     map[string]string // unit path -> PlantUML alias
	usedAliases map[string]bool   // sanitized aliases already taken

	elemTags map[string]string // style key -> registered element tag name
	relTags  map[string]string // style key -> registered relationship tag name
	tagOrder []string          // registration order of both tag kinds ("elem:"/"rel:" prefixed keys)
}

// newPlantUMLWriter allocates a writer for the graph.
func newPlantUMLWriter(g *graph.Graph) *plantUMLWriter {
	return &plantUMLWriter{
		g:           g,
		aliases:     make(map[string]string),
		usedAliases: make(map[string]bool),
		elemTags:    make(map[string]string),
		relTags:     make(map[string]string),
	}
}

// render assembles the document: include line, title, legend switch, tag
// declarations, elements (boundaries included), then relationships.
func (w *plantUMLWriter) render() []byte {
	var b strings.Builder

	b.WriteString("@startuml\n")
	b.WriteString("!include " + w.include() + "\n")

	if w.g.Title != "" {
		b.WriteString("\ntitle " + pumlEscape(w.g.Title) + "\n")
	}

	if w.g.Legend != nil && len(w.g.Legend.Entries) > 0 {
		b.WriteString("\nLAYOUT_WITH_LEGEND()\n")
	}

	elements, rels := w.body()

	if len(w.tagOrder) > 0 {
		b.WriteString("\n")

		for _, key := range w.tagOrder {
			b.WriteString(w.tagDecl(key) + "\n")
		}
	}

	if elements != "" {
		b.WriteString("\n" + elements)
	}

	if rels != "" {
		b.WriteString("\n" + rels)
	}

	b.WriteString("\n@enduml\n")

	return []byte(b.String())
}

// include returns the include line for the deepest C4 level the diagram
// needs: C4_Component pulls in C4_Container and C4_Context transitively, so
// one line always suffices.
func (w *plantUMLWriter) include() string {
	switch deepestLevelOf(w.g) {
	case 3:
		return pumlComponentInclude
	case 2:
		return pumlContainerInclude
	default:
		return pumlContextInclude
	}
}

// deepestLevelOf returns the deepest C4 level (1..3) among the graph's node
// and cluster unit types, walking the cluster tree.
func deepestLevelOf(g *graph.Graph) int {
	deepest := 1

	consider := func(t model.UnitType) {
		if l := graph.LevelForType(t); l > deepest {
			deepest = l
		}
	}

	var walk func(nodes []*graph.Node, clusters []*graph.Cluster)

	walk = func(nodes []*graph.Node, clusters []*graph.Cluster) {
		for _, n := range nodes {
			if n != nil {
				consider(n.Type)
			}
		}

		for _, c := range clusters {
			if c == nil {
				continue
			}

			consider(c.Type)
			walk(c.Nodes, c.Clusters)
		}
	}

	walk(g.Nodes, g.Clusters)

	return deepest
}

// body serializes all elements and relationships, returning them as two
// text blocks. Nodes outside clusters come first, then the cluster tree in
// declaration order (boundaries opening/closing around their contents).
func (w *plantUMLWriter) body() (string, string) {
	var elems, relLines strings.Builder

	for _, n := range w.g.Nodes {
		if n == nil || n.IsInCluster {
			continue
		}

		elems.WriteString(w.nodeStmt(n) + "\n")
	}

	for _, c := range w.g.Clusters {
		if c != nil {
			w.boundaryStmt(&elems, c)
		}
	}

	for _, e := range w.g.Edges {
		if e == nil {
			continue
		}

		if stmt := w.relStmt(e); stmt != "" {
			relLines.WriteString(stmt + "\n")
		}
	}

	return strings.TrimRight(elems.String(), "\n"), strings.TrimRight(relLines.String(), "\n")
}

// boundaryStmt emits one boundary macro block for the cluster — open line,
// its nodes, nested clusters, close — recursing so nested boundaries nest in
// the PlantUML source exactly as they nest in the graph.
func (w *plantUMLWriter) boundaryStmt(elems *strings.Builder, c *graph.Cluster) {
	macro, typeText := plantUMLBoundaryMacro(c.Type)
	alias := w.aliasFor(c.ID)

	// System_Boundary(alias, label, tags, link) vs generic
	// Boundary(alias, label, type, tags, link): the type text shifts the
	// trailing slots by one, so fill them by index.
	args := []string{alias, "\"" + pumlEscape(clusterName(c)) + "\""}

	setArg := func(idx int, val string) {
		for len(args) <= idx {
			args = append(args, "\"\"")
		}

		if val != "" {
			args[idx] = "\"" + val + "\""
		}
	}

	if typeText != "" {
		setArg(2, typeText)
	}

	setArg(len(args), w.elementTag(c.Style)) // tags slot

	if link := nodeURL(c.ExploreURL, c.ReferenceURL); link != "" {
		setArg(len(args), pumlEscape(link)) // link slot
	}

	elems.WriteString(macro + "(" + strings.Join(trimTrailingEmptyArgs(args, 2), ", ") + ") {\n")

	for _, n := range c.Nodes {
		elems.WriteString("\t" + w.nodeStmt(n) + "\n")
	}

	for _, nested := range c.Clusters {
		w.boundaryStmt(elems, nested)
	}

	elems.WriteString("}\n")
}

// nodeStmt builds one element macro call for the node: macro chosen by the
// resolved unit type, label/technology/description from the graph label, the
// element tag for any non-default styling, and the drill-down/reference link
// in the macro's $link slot (drill-down wins the single URL slot — same
// precedence as the converter's SetURL). C4-PlantUML expands $link into
// `[[url]]` markup, so the converted SVG carries a clickable anchor with the
// ComputeExploreURL href.
func (w *plantUMLWriter) nodeStmt(n *graph.Node) string {
	macro, hasTechn := plantUMLElementMacro(n.Type)
	alias := w.aliasFor(n.ID)

	// Slot layout per macro family (1-based in the docs, 0-based here):
	// Person/System family: (alias, label, descr, sprite, tags, link);
	// Container/Component family: (alias, label, techn, descr, sprite,
	// tags, link). Arguments are filled by index so a value in a later
	// slot pads every earlier optional slot with "".
	descrIdx, tagsIdx, linkIdx := 2, 4, 5

	args := []string{alias, "\"" + pumlEscape(nodeName(n)) + "\""}

	setArg := func(idx int, val string) {
		for len(args) <= idx {
			args = append(args, "\"\"")
		}

		if val != "" {
			args[idx] = "\"" + val + "\""
		}
	}

	label := n.Label

	if label != nil {
		if hasTechn && label.Technology != "" {
			setArg(2, pumlEscape(label.Technology))

			descrIdx = 3
		}

		if label.Description != "" {
			// Person/System macros have no techn slot: technology is
			// skipped there (the stdlib macro has no parameter for it).
			setArg(descrIdx, pumlEscape(label.Description))
		}
	}

	// The tags slot must be addressed (even if empty) whenever a link
	// follows it.
	tags := w.elementTag(n.Style)
	link := nodeURL(n.ExploreURL, n.ReferenceURL)

	if tags != "" || link != "" {
		setArg(tagsIdx, tags)
	}

	if link != "" {
		setArg(linkIdx, pumlEscape(link))
	}

	return macro + "(" + strings.Join(trimTrailingEmptyArgs(args, 2), ", ") + ")"
}

// trimTrailingEmptyArgs drops trailing "" placeholders from a macro call's
// argument list, keeping at least keep arguments — empty optional slots at
// the end of a macro call are noise.
func trimTrailingEmptyArgs(args []string, keep int) []string {
	for len(args) > keep && args[len(args)-1] == "\"\"" {
		args = args[:len(args)-1]
	}

	return args
}

// nodeName picks the label text for an element: the graph label's name
// (which already carries the 🔍/📖 affordance glyphs) falling back to the
// unit path.
func nodeName(n *graph.Node) string {
	if n.Label != nil && n.Label.Name != "" {
		return n.Label.Name
	}

	return n.ID
}

// clusterName picks the boundary title: the cluster label's name falling
// back to the cluster path.
func clusterName(c *graph.Cluster) string {
	if c.Label != nil && c.Label.Name != "" {
		return c.Label.Name
	}

	return c.ID
}

// nodeURL applies the single-URL-slot precedence: the drill-down wins, the
// external reference takes the slot only when no drill-down applies.
func nodeURL(exploreURL, referenceURL string) string {
	if exploreURL != "" {
		return exploreURL
	}

	return referenceURL
}

// relStmt builds one relationship macro call for the edge, or "" when either
// endpoint is not rendered in this diagram (the converter's skip-missing rule).
// Arrow direction maps onto the macro: forward -> Rel, reverse -> Rel_Back
// (C4-PlantUML puts the arrowhead on the `from` end, which is the c4drill
// source — exactly the c4drill semantics), bidirectional -> BiRel, none -> the
// arrowless Rel_ variant with the "--" direction. Technology and description
// ride in the macro's label/techn slots; colour, line style and thickness are
// carried by an AddRelTag-declared tag (except for the arrowless variant,
// whose macro has no tags slot).
func (w *plantUMLWriter) relStmt(e *graph.Edge) string {
	from, okFrom := w.aliases[e.Source]
	to, okTo := w.aliases[e.Target]

	if !okFrom || !okTo {
		return ""
	}

	var label, techn string

	if e.Label != nil {
		label, techn = e.Label.Description, e.Label.Technology
	}

	switch e.ArrowHead {
	case graph.ArrowForward:
		return relCall("Rel", from, to, label, techn, w.relTag(e))
	case graph.ArrowReverse:
		return relCall("Rel_Back", from, to, label, techn, w.relTag(e))
	case graph.ArrowBoth:
		return relCall("BiRel", from, to, label, techn, w.relTag(e))
	case graph.ArrowNone:
		// Rel_'s arrowless form: Rel_(from, to, label, techn, "--") with
		// techn, else the 4-argument shape. No tags slot exists, so
		// colour/style/thickness have nowhere to ride — the arrowless
		// variant stays an undecorated plain line.
		if techn != "" {
			return "Rel_(" + from + ", " + to + ", \"" + pumlEscape(label) + "\", \"" +
				pumlEscape(techn) + "\", \"--\")"
		}

		return "Rel_(" + from + ", " + to + ", \"" + pumlEscape(label) + "\", \"--\")"
	default:
		// Unknown arrow directions draw forward, matching the converter's
		// setEdgeDir default.
		return relCall("Rel", from, to, label, techn, w.relTag(e))
	}
}

// relCall renders a Rel/Rel_Back/BiRel call. The tags argument is positional
// slot 7 (after label, techn, descr, sprite), so the empty slots are padded.
func relCall(macro, from, to, label, techn, tags string) string {
	args := []string{from, to, "\"" + pumlEscape(label) + "\""}

	if techn != "" {
		args = append(args, "\""+pumlEscape(techn)+"\"")
	}

	if tags != "" {
		for len(args) < 6 {
			args = append(args, "\"\"")
		}

		args = append(args, "\""+tags+"\"")
	}

	return macro + "(" + strings.Join(trimTrailingEmptyArgs(args, 3), ", ") + ")"
}

// aliasFor returns the stable PlantUML alias for a unit path: the path with
// every non-[A-Za-z0-9_] character folded to '_' (dots are not safe in
// aliases), uniquified with a numeric suffix on collision.
func (w *plantUMLWriter) aliasFor(path string) string {
	if a, ok := w.aliases[path]; ok {
		return a
	}

	base := sanitizeForName(path)
	if base == "" {
		base = "element"
	}

	alias := base

	for n := 2; w.usedAliases[alias]; n++ {
		alias = base + "_" + itoa(n)
	}

	w.usedAliases[alias] = true
	w.aliases[path] = alias

	return alias
}

// elementTag returns the AddElementTag tag name styling the given node
// style, registering a new tag on first use. Empty styles (and styles that
// carry nothing beyond defaults) yield "" — the element keeps the stock
// C4-PlantUML palette.
func (w *plantUMLWriter) elementTag(style *graph.NodeStyle) string {
	if style == nil {
		return ""
	}

	if style.FillColor == "" && style.FontColor == "" && style.BorderColor == "" &&
		style.BorderStyle != borderStyleDashed && style.BorderStyle != borderStyleDotted {
		return ""
	}

	key := "elem:" + style.FillColor + "|" + style.FontColor + "|" +
		style.BorderColor + "|" + style.BorderStyle

	if tag, ok := w.elemTags[key]; ok {
		return tag
	}

	tag := "c4drillElem" + itoa(len(w.elemTags))
	w.elemTags[key] = tag
	w.tagOrder = append(w.tagOrder, key)

	return tag
}

// relTag returns the AddRelTag tag name styling the given edge (line/label
// colour, dashed/dotted line style, doubled thickness for collapsed pairs),
// registering a new tag on first use. Plain edges yield "".
func (w *plantUMLWriter) relTag(e *graph.Edge) string {
	if e.Color == "" && e.Style != borderStyleDashed && e.Style != borderStyleDotted && e.PenWidth < 2 {
		return ""
	}

	key := "rel:" + e.Color + "|" + e.Style + "|" + itoa(int(e.PenWidth))

	if tag, ok := w.relTags[key]; ok {
		return tag
	}

	tag := "c4drillRel" + itoa(len(w.relTags))
	w.relTags[key] = tag
	w.tagOrder = append(w.tagOrder, key)

	return tag
}

// tagDecl renders the AddElementTag/AddRelTag declaration a previously
// registered tag key stands for. Named arguments (the C4-PlantUML README
// style) keep the declarations readable without padding ten positional
// slots.
func (w *plantUMLWriter) tagDecl(key string) string {
	if rest, ok := strings.CutPrefix(key, "elem:"); ok {
		parts := strings.SplitN(rest, "|", 4)

		var b strings.Builder
		b.WriteString("AddElementTag(\"" + w.elemTags[key] + "\"")

		if parts[0] != "" {
			b.WriteString(", $bgColor=\"" + parts[0] + "\"")
		}

		if parts[1] != "" {
			b.WriteString(", $fontColor=\"" + parts[1] + "\"")
		}

		if parts[2] != "" {
			b.WriteString(", $borderColor=\"" + parts[2] + "\"")
		}

		// "solid" is the C4-PlantUML default — only the deviations are
		// worth a declaration.
		if parts[3] == borderStyleDashed || parts[3] == borderStyleDotted {
			b.WriteString(", $borderStyle=\"" + parts[3] + "\"")
		}

		b.WriteString(")")

		return b.String()
	}

	rest, _ := strings.CutPrefix(key, "rel:")
	parts := strings.SplitN(rest, "|", 3)

	color, style, thick := parts[0], parts[1], parts[2]

	var b strings.Builder
	b.WriteString("AddRelTag(\"" + w.relTags[key] + "\"")

	if color != "" {
		b.WriteString(", $textColor=\"" + color + "\", $lineColor=\"" + color + "\"")
	}

	if style == borderStyleDashed || style == borderStyleDotted {
		b.WriteString(", $lineStyle=\"" + style + "\"")
	}

	if thick != "0" {
		b.WriteString(", $lineThickness=\"" + thick + "\"")
	}

	b.WriteString(")")

	return b.String()
}

// pumlEscape makes author-controlled text safe inside a PlantUML double
// quoted string: quotes are backslash-escaped and newlines become the \n
// escape so a multi-line description cannot break the macro call.
func pumlEscape(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, `"`, `\"`)

	return s
}

// itoa formats a small non-negative integer (tag/alias suffix counters).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	var digits [20]byte

	i := len(digits)

	for n > 0 {
		i--
		digits[i] = byte('0' + n%10)
		n /= 10
	}

	return string(digits[i:])
}
