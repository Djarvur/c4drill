// c4dctx.go is the C4D-side counterpart of tomlctx.go: the block-structure
// scanner the .c4d language features share. One pass over the CURRENT buffer
// tracks brace nesting with C4D's real line structure — '#' comments,
// double-quoted strings with escapes, triple-quoted multi-line strings, ';'
// separators, one-line nested blocks — and classifies the cursor's statement
// context (header type slot, field value, edge peer, use arguments, include
// path, statement start).
//
// This is deliberately NOT a second C4D parser: DSL semantics (type
// inference, Model building, peer/template resolution) stay with
// internal/c4d's PEG grammar, and hover/definition resolve through the
// c4d.ParseNamed *parser.Model exactly like the TOML features resolve through
// parser.Parse. The scanner exists because completion and the outline must
// survive mid-edit buffers that no longer parse (the contract tomlctx.go
// serves for TOML), and because the AST carries open lines only — the brace
// block ranges the outline needs come from this walk.

package lsp

import (
	"strings"

	"github.com/Djarvur/c4drill/internal/c4d/grammar"
)

// statement keywords and modifiers of the DSL (c4d.peg TopPart/BodyStmt).
const (
	kwExternal = "external"
	kwOnce     = "once"
)

// c4dArrowGlyphs are the four edge arrows, longest first so prefix matching
// picks "<->" over "->".
//
//nolint:gochecknoglobals // pinned closed vocabulary (c4d.peg ArrowGlyph), immutable
var c4dArrowGlyphs = []string{"<->", "->", "<-", "--"}

// c4dBlockKind classifies an open brace block.
type c4dBlockKind int

const (
	c4dBlockUnit        c4dBlockKind = iota // a unit header's brace block
	c4dBlockProperties                      // properties { }
	c4dBlockTemplate                        // template name(params) { }
	c4dBlockEdgeOptions                     // an edge statement's option block
	c4dBlockOther                           // unrecognized (mid-edit junk)
)

// c4dBlock is one brace block found by the scan. Unit blocks carry their
// dotted path and header slots; template declarations carry the param list.
type c4dBlock struct {
	kind     c4dBlockKind
	line     int      // 0-based line the '{' sits on
	endLine  int      // 0-based line the block closed on (-1 while open)
	path     string   // dotted unit path (unit blocks; template-relative inside template bodies)
	parent   string   // parent unit path ("" at document/template root)
	key      string   // the unit's own key segment (id, display name, or type)
	id       string   // authored header id ("" for type-led headers)
	typ      string   // declared type keyword ("" when omitted)
	external bool     // the `external` modifier (D-04)
	name     string   // quoted display name from the header ("" when omitted)
	params   []string // template declaration params
	tmpl     string   // owning template for units inside template bodies ("" = document unit)
	fields   []string // field keys authored in this block, source order, deduped
}

// c4dInclude is one `include path [once]` directive (D-14).
type c4dInclude struct {
	line int
	path string
	once bool
}

// c4dScan is the scanner state and the resulting document inventory.
type c4dScan struct {
	text string

	pos      int // byte offset
	line     int // 0-based line
	stack    []*c4dBlock
	units    []*c4dBlock // every unit block, source order (open ones included)
	tmpls    []*c4dBlock // template declarations, source order
	includes []c4dInclude

	// current statement accumulation. The statement buffer keeps RAW text
	// (quotes included) because header classification needs the display
	// name; newlines and ';' end a statement unless inside parentheses
	// (multi-line use/template args) or brackets (multi-line list values).
	stmt     strings.Builder
	stmtLine int
	paren    int
	bracket  int

	// use-statement context of the current statement: `use name(args)`.
	useName  string
	useParen bool

	// cursor snapshot: taken once pos crosses the scan's stop offset.
	stop    int
	snapped bool
	snap    c4dCursorState

	lastLine int
}

// c4dCursorState is the scanner state at the cursor: the open block stack,
// the partial statement, and the use context.
type c4dCursorState struct {
	stack    []*c4dBlock
	stmt     string
	stmtLine int
	useName  string
	useParen bool
}

// c4dScanDocument scans the whole document and returns the inventory.
func c4dScanDocument(text string) *c4dScan {
	sc := newC4DScan(text)
	sc.stop = len(text)
	sc.run()

	return sc
}

// newC4DScan builds a scanner over text.
func newC4DScan(text string) *c4dScan {
	return &c4dScan{
		text:     text,
		stmtLine: 0,
		stop:     len(text),
	}
}

// run scans to the end of text, snapshotting the cursor state at stop.
func (sc *c4dScan) run() {
	text := sc.text

	for sc.pos < len(text) {
		if !sc.snapped && sc.pos >= sc.stop {
			sc.takeSnapshot()
		}

		sc.step(text[sc.pos])
	}

	sc.finish()
}

// step consumes one structural byte.
func (sc *c4dScan) step(b byte) {
	switch b {
	case '#':
		sc.skipComment()
	case '"':
		sc.skipString()
	case '\n':
		sc.line++
		sc.pos++
		sc.endStatement()
	case ';':
		sc.pos++
		sc.endStatement()
	case '{':
		sc.openBrace()
	case '}':
		sc.closeBrace()
	case '(':
		sc.paren++
		sc.appendByte('(')
		sc.noteOpenParen()
	case ')':
		if sc.paren > 0 {
			sc.paren--
		}

		sc.appendByte(')')
		sc.useParen = false
	case '[', ']':
		sc.stepBracket(b)
	default:
		sc.appendByte(b)
	}
}

// stepBracket tracks list-literal brackets so multi-line list values do not
// split the statement.
func (sc *c4dScan) stepBracket(b byte) {
	if b == '[' {
		sc.bracket++
	} else if sc.bracket > 0 {
		sc.bracket--
	}

	sc.appendByte(b)
}

// finish closes the scan: snapshot, flush the trailing statement, and bound
// unterminated blocks at the last line.
func (sc *c4dScan) finish() {
	if !sc.snapped {
		sc.takeSnapshot()
	}

	sc.recordStatement() // a trailing statement without a final newline

	sc.lastLine = sc.line

	for _, blk := range sc.stack {
		if blk.endLine < 0 {
			blk.endLine = sc.lastLine // unterminated block runs to EOF
		}
	}
}

// takeSnapshot freezes the cursor state (once).
func (sc *c4dScan) takeSnapshot() {
	sc.snapped = true
	sc.snap = c4dCursorState{
		stack:    append([]*c4dBlock(nil), sc.stack...),
		stmt:     sc.stmt.String(),
		stmtLine: sc.stmtLine,
		useName:  sc.useName,
		useParen: sc.useParen,
	}
}

// skipComment consumes a '#' comment to (not including) the newline.
func (sc *c4dScan) skipComment() {
	for sc.pos < len(sc.text) && sc.text[sc.pos] != '\n' {
		sc.pos++
	}
}

// skipString consumes a double-quoted (escape-aware) or triple-quoted
// string, appending its RAW text to the statement buffer — header
// classification needs the quoted display name. When the string spans the
// scan's stop offset (the cursor sits inside it), only the text up to the
// stop is appended so the cursor snapshot stays exact.
func (sc *c4dScan) skipString() {
	start := sc.pos

	if strings.HasPrefix(sc.text[start:], `"""`) {
		end := start + 3

		if i := strings.Index(sc.text[start+3:], `"""`); i >= 0 {
			end = start + 3 + i + 3
		}

		sc.appendRawUntil(start, end)

		return
	}

	sc.appendRawUntil(start, sc.doubleQuotedEnd())
}

// doubleQuotedEnd returns the end offset of the double-quoted string opening
// at sc.pos: escape-aware, terminated by a closing quote or the line's end
// (the grammar's DoubleQuoted never spans lines).
func (sc *c4dScan) doubleQuotedEnd() int {
	pos := sc.pos + 1 // opening quote

	for pos < len(sc.text) {
		switch {
		case sc.text[pos] == '\\' && pos+1 < len(sc.text):
			pos += 2
		case sc.text[pos] == '"':
			pos++

			return pos
		case sc.text[pos] == '\n':
			return pos // unterminated: the newline re-enters the scan
		default:
			pos++
		}
	}

	return pos
}

// appendRawUntil appends text[start:limit) to the statement (counting
// lines), where limit respects the cursor stop, and jumps the scan to end.
func (sc *c4dScan) appendRawUntil(start, end int) {
	limit := end

	if !sc.snapped && sc.stop > start && sc.stop < end {
		limit = sc.stop // the cursor sits inside this string
	}

	raw := sc.text[start:limit]
	sc.line += strings.Count(raw, "\n")
	sc.stmt.WriteString(raw)
	sc.pos = end
}

// appendByte appends one statement byte.
func (sc *c4dScan) appendByte(b byte) {
	sc.stmt.WriteByte(b)
	sc.pos++
}

// noteOpenParen captures the `use name(` context when the statement's opening
// paren arrives (D-13/D-16: use arguments may span lines).
func (sc *c4dScan) noteOpenParen() {
	sc.useParen = true

	head := sc.stmtHead()
	if name, ok := c4dUseHeadName(head); ok {
		sc.useName = name
	}
}

// stmtHead returns the current statement text, trimmed.
func (sc *c4dScan) stmtHead() string {
	return strings.TrimSpace(sc.stmt.String())
}

// endStatement records the finished statement into its enclosing block and
// resets the buffer.
func (sc *c4dScan) endStatement() {
	sc.recordStatement()
	sc.resetStmt()
}

// resetStmt starts a fresh statement.
func (sc *c4dScan) resetStmt() {
	sc.stmt.Reset()
	sc.stmtLine = sc.line
	sc.paren = 0
	sc.bracket = 0
	sc.useName = ""
	sc.useParen = false
}

// recordStatement folds a non-block statement into the inventory: include
// directives at top level, field keys into the innermost recordable block.
func (sc *c4dScan) recordStatement() {
	stmt := sc.stmtHead()
	if stmt == "" {
		return
	}

	if path, once, ok := c4dIncludeOf(stmt); ok {
		sc.includes = append(sc.includes, c4dInclude{line: sc.stmtLine, path: path, once: once})

		return
	}

	if sc.paren > 0 || sc.bracket > 0 {
		return // a still-open argument list or list literal: not a statement yet
	}

	blk := sc.recordableBlock()
	if blk == nil {
		return
	}

	if colon := indexOutsideQuotes(stmt, ':'); colon >= 0 {
		// Only identifier keys record — edge statements (`-> peer: label`)
		// and use statements are not field statements.
		if key := strings.TrimSpace(stmt[:colon]); c4dIsIdent(key) {
			blk.fields = appendUniqueField(blk.fields, key)
		}
	}
}

// recordableBlock returns the innermost block that owns field statements.
func (sc *c4dScan) recordableBlock() *c4dBlock {
	if len(sc.stack) == 0 {
		return nil
	}

	blk := sc.stack[len(sc.stack)-1]
	if blk.kind == c4dBlockUnit || blk.kind == c4dBlockProperties ||
		blk.kind == c4dBlockTemplate || blk.kind == c4dBlockEdgeOptions {
		return blk
	}

	return nil
}

// appendUniqueField appends key when non-empty and not yet recorded.
func appendUniqueField(fields []string, key string) []string {
	if key == "" {
		return fields
	}

	for _, f := range fields {
		if f == key {
			return fields
		}
	}

	return append(fields, key)
}

// openBrace classifies the accumulated statement and pushes the block.
func (sc *c4dScan) openBrace() {
	if sc.paren > 0 || sc.bracket > 0 {
		sc.appendByte('{') // structural braces cannot occur inside parens/lists

		return
	}

	blk := classifyC4DBlock(sc.stmtHead())
	blk.line = sc.line
	sc.pos++

	sc.placeUnit(blk)

	sc.stack = append(sc.stack, blk)

	if blk.kind == c4dBlockUnit {
		sc.units = append(sc.units, blk)
	}

	if blk.kind == c4dBlockTemplate {
		sc.tmpls = append(sc.tmpls, blk)
	}

	sc.resetStmt()
}

// placeUnit derives the unit's dotted path and owning template from the open
// block stack, mirroring internal/c4d's unitKey: the authored id wins, then
// the quoted display name, then the type keyword. Key and parent are kept
// verbatim — dotted display names are legal keys and must not be re-split.
func (sc *c4dScan) placeUnit(blk *c4dBlock) {
	if blk.kind != c4dBlockUnit {
		return
	}

	parent := ""
	blk.tmpl = ""

	for i := len(sc.stack) - 1; i >= 0; i-- {
		enclosing := sc.stack[i]

		switch enclosing.kind {
		case c4dBlockUnit:
			parent = enclosing.path
			blk.tmpl = enclosing.tmpl
		case c4dBlockTemplate:
			blk.tmpl = enclosing.id
		case c4dBlockProperties, c4dBlockEdgeOptions, c4dBlockOther:
			continue
		}

		break
	}

	blk.parent = parent

	blk.key = blk.id
	switch {
	case blk.key != "": // id-led header
	case blk.name != "":
		blk.key = blk.name
	default:
		blk.key = blk.typ
	}

	blk.path = c4dJoinPath(parent, blk.key)
}

// closeBrace pops the innermost block, bounding its range.
func (sc *c4dScan) closeBrace() {
	if sc.paren > 0 || sc.bracket > 0 {
		sc.appendByte('}')

		return
	}

	sc.pos++

	if n := len(sc.stack); n > 0 {
		blk := sc.stack[n-1]
		blk.endLine = sc.line
		sc.stack = sc.stack[:n-1]
	}

	sc.resetStmt()
}

// classifyC4DBlock builds a block from the statement head that opened a
// brace, in the grammar's own preference order (c4d.peg): edge option block,
// properties, template declaration, then the unit header forms.
func classifyC4DBlock(stmt string) *c4dBlock {
	blk := &c4dBlock{endLine: -1}

	switch {
	case stmt == "":
		blk.kind = c4dBlockOther
	case c4dArrowPrefix(stmt) != "":
		blk.kind = c4dBlockEdgeOptions // `-> peer: label { ... }`
	case stmt == tblProperties:
		blk.kind = c4dBlockProperties
	default:
		classifyC4DBlockHead(blk, stmt)
	}

	return blk
}

// classifyC4DBlockHead resolves template declarations and the three unit
// header forms (D-02): id-led `id: type [external] ["Name"]`, type-led
// `type [external] ["Name"]`, and bare-id `id`.
func classifyC4DBlockHead(blk *c4dBlock, stmt string) {
	if name, params, ok := c4dTemplateHead(stmt); ok {
		blk.kind = c4dBlockTemplate
		blk.id = name
		blk.params = params

		return
	}

	id := ""

	rest := stmt

	if colon := indexOutsideQuotes(stmt, ':'); colon >= 0 {
		head := strings.TrimSpace(stmt[:colon])
		if c4dReservedWord(head) {
			blk.kind = c4dBlockOther // field-shaped head (or ReservedUnitId junk)

			return
		}

		if !c4dIsIdent(head) {
			blk.kind = c4dBlockOther

			return
		}

		id = head
		rest = stmt[colon+1:]
	}

	typ, external, name, ok := c4dHeaderSlots(rest)
	if !ok {
		blk.kind = c4dBlockOther

		return
	}

	switch {
	case id != "" || typ != "" || name != "":
		blk.kind = c4dBlockUnit
		blk.id, blk.typ, blk.external, blk.name = id, typ, external, name
	case c4dIsIdent(stmt) && !c4dReservedWord(stmt):
		blk.kind = c4dBlockUnit // bare-id header: `id {`
		blk.id = stmt
	default:
		blk.kind = c4dBlockOther
	}
}

// c4dHeaderSlots parses a header remainder `[type [external] ["Name"]]` and
// reports whether it is well-formed. All slots are optional; the order is
// fixed. An empty remainder is well-formed (a mid-typing id-led header).
func c4dHeaderSlots(rest string) (string, bool, string, bool) {
	tokens := c4dSplitTokens(rest)

	typ := ""
	external := false
	name := ""

	i := 0

	if i < len(tokens) && c4dTypeKeyword(tokens[i]) {
		typ = tokens[i]
		i++
	}

	if i < len(tokens) && tokens[i] == kwExternal {
		external = true
		i++
	}

	if i < len(tokens) {
		if i != len(tokens)-1 || !strings.HasPrefix(tokens[i], `"`) {
			return "", false, "", false
		}

		name = unquote(tokens[i])
	}

	return typ, external, name, true
}

// c4dSplitTokens splits on whitespace outside double-quoted spans, keeping
// the quotes on quoted tokens.
func c4dSplitTokens(s string) []string {
	var tokens []string

	for s = strings.TrimLeft(s, " \t\r\n"); s != ""; s = strings.TrimLeft(s, " \t\r\n") {
		end := strings.IndexAny(s, " \t\r\n")
		if s[0] == '"' {
			end = quotedTokenEnd(s)
		}

		if end < 0 {
			end = len(s)
		}

		tokens = append(tokens, s[:end])
		s = s[end:]
	}

	return tokens
}

// quotedTokenEnd returns the offset just past a double-quoted token's
// closing quote (escape-aware), or len(s) when unterminated.
func quotedTokenEnd(s string) int {
	for i := 1; i < len(s); i++ {
		switch {
		case s[i] == '\\' && i+1 < len(s):
			i++
		case s[i] == '"':
			return i + 1
		}
	}

	return len(s)
}

// c4dTemplateHead matches `template name(p1, p2)` and returns the name and
// declared params.
func c4dTemplateHead(stmt string) (string, []string, bool) {
	rest, ok := strings.CutPrefix(stmt, tblTemplate+" ")
	if !ok {
		return "", nil, false
	}

	name, tail, found := strings.Cut(rest, "(")
	if !found {
		return "", nil, false
	}

	name, tail = strings.TrimSpace(name), strings.TrimSpace(tail)
	if !c4dIsIdent(name) || !strings.HasSuffix(tail, ")") {
		return "", nil, false
	}

	inner := strings.TrimSuffix(tail, ")")

	var params []string

	for _, p := range strings.Split(inner, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			params = append(params, p)
		}
	}

	return name, params, true
}

// c4dUseHeadName extracts the template name of a `use name(` head.
func c4dUseHeadName(head string) (string, bool) {
	rest, ok := strings.CutPrefix(head, tblUse+" ")
	if !ok {
		return "", false
	}

	rest = strings.TrimSpace(rest)

	if i := strings.IndexByte(rest, '('); i >= 0 {
		rest = rest[:i]
	}

	rest = strings.TrimSpace(rest)
	if !c4dIsIdent(rest) {
		return "", false
	}

	return rest, true
}

// c4dIncludeOf parses an `include path [once]` statement (D-14).
func c4dIncludeOf(stmt string) (string, bool, bool) {
	rest, ok := strings.CutPrefix(stmt, tblInclude+" ")
	if !ok {
		if stmt == tblInclude {
			return "", false, true // bare directive, path still being typed
		}

		return "", false, false
	}

	rest = strings.TrimSpace(rest)

	once := false

	if unquoted := !strings.HasPrefix(rest, `"`); unquoted {
		if trimmed, found := strings.CutSuffix(rest, " "+kwOnce); found {
			rest, once = strings.TrimSpace(trimmed), true
		}
	}

	path := unquote(rest)
	if path == "" {
		return "", false, true // path still being typed
	}

	return path, once, true
}

// c4dJoinPath joins a dotted parent with a child segment.
func c4dJoinPath(parent, id string) string {
	switch {
	case parent == "":
		return id
	case id == "":
		return parent
	default:
		return parent + "." + id
	}
}

// c4dIsIdent reports the DSL identifier shape ([A-Za-z0-9_-]+, D-07).
func c4dIsIdent(s string) bool {
	if s == "" {
		return false
	}

	for i := range len(s) {
		b := s[i]

		switch {
		case b >= 'a' && b <= 'z', b >= 'A' && b <= 'Z', b >= '0' && b <= '9':
		case b == '_' || b == '-':
		default:
			return false
		}
	}

	return true
}

// c4dTypeKeyword reports whether s is one of the 17 exact type keywords
// (c4d.peg TypeKeyword).
func c4dTypeKeyword(s string) bool {
	for _, t := range allUnitTypes() {
		if string(t) == s {
			return true
		}
	}

	return false
}

// c4dReservedWords is the closed reserved-word set (c4d/grammar/reserved.go,
// D-19): the ids a unit header cannot use, which therefore mark field- and
// statement-shaped heads.
//
//nolint:gochecknoglobals // pinned closed set per D-19, immutable after init
var c4dReservedWords = func() map[string]bool {
	set := make(map[string]bool)

	for _, kw := range grammar.ReservedKeywords() {
		set[kw] = true
	}

	return set
}()

// c4dReservedWord reports the reserved-word set membership.
func c4dReservedWord(s string) bool {
	return c4dReservedWords[s]
}

// c4dArrowPrefix returns the edge glyph heading s, "" when none.
func c4dArrowPrefix(s string) string {
	for _, g := range c4dArrowGlyphs {
		if strings.HasPrefix(s, g) {
			return g
		}
	}

	return ""
}

// --- cursor context ------------------------------------------------------

// c4dStmtKind classifies the cursor's statement context.
type c4dStmtKind int

const (
	c4dStmtTemplateRef c4dStmtKind = iota // inside a ${param} token
	c4dStmtTypeSlot                       // unit header type slot / template root type value
	c4dStmtFieldValue                     // `key: value` in a unit, properties, or option block
	c4dStmtEdgePeer                       // `-> peer` target position
	c4dStmtUseName                        // `use name(` template-name position
	c4dStmtUseArgKey                      // `use t(key` argument-key position
	c4dStmtUseArgValue                    // `use t(key: value` argument-value position
	c4dStmtIncludePath                    // `include path` position
	c4dStmtStart                          // statement start (block keyword lists)
	c4dStmtOther                          // anything else
)

// c4dCursor is the language-feature view of the cursor location in a .c4d
// document: the full-document inventory plus the statement context.
type c4dCursor struct {
	inventory *c4dScan

	stack    []*c4dBlock // open blocks at the cursor, innermost last
	stmt     string      // statement head text before the cursor (raw)
	stmtLine int         // 0-based line the statement started on
	line     string      // the cursor's line
	lineNo   int         // 0-based cursor line
	byteIdx  int         // byte index within the line

	kind          c4dStmtKind
	key           string // field/option key for c4dStmtFieldValue
	valuePrefix   string // typed value text, quotes stripped (also the ${...} prefix)
	fullPeer      string // the complete peer token of the cursor's edge
	useName       string // template named by the cursor's use statement
	useArgPrefix  string // partial argument key inside use (...)
	includePrefix string // typed include path (quotes stripped)

	hostPath string // innermost document unit's dotted path ("" when none)
	tmplName string // enclosing template declaration name ("" when outside)
}

// c4dAnalyze scans text and classifies the cursor position.
func c4dAnalyze(text string, pos Position) *c4dCursor {
	sc := newC4DScan(text)
	sc.stop = offsetForPosition(text, pos)
	sc.run()

	cur := &c4dCursor{
		inventory: sc,
		stack:     sc.snap.stack,
		stmt:      sc.snap.stmt,
		stmtLine:  sc.snap.stmtLine,
		useName:   sc.snap.useName,
	}

	starts := lineStarts(text)
	if int(pos.Line) < len(starts) {
		cur.line = text[starts[pos.Line]:]
		if nl := strings.IndexByte(cur.line, '\n'); nl >= 0 {
			cur.line = cur.line[:nl]
		}
	}

	cur.lineNo = int(pos.Line)

	cur.byteIdx = sc.stop - starts[min(cur.lineNo, len(starts)-1)]

	if cur.byteIdx < 0 {
		cur.byteIdx = 0
	}

	cur.hostPath, cur.tmplName = c4dScopeOf(cur.stack)
	cur.classify()

	return cur
}

// c4dScopeOf derives the innermost document unit path and the enclosing
// template (if any) from a block stack.
func c4dScopeOf(stack []*c4dBlock) (string, string) {
	hostPath := ""
	tmplName := ""

	for i := len(stack) - 1; i >= 0; i-- {
		switch stack[i].kind {
		case c4dBlockUnit:
			if stack[i].tmpl == "" {
				return stack[i].path, tmplName
			}

			if hostPath == "" {
				hostPath = stack[i].path // template-local path
			}
		case c4dBlockTemplate:
			return hostPath, stack[i].id
		case c4dBlockProperties, c4dBlockEdgeOptions, c4dBlockOther:
		}
	}

	return hostPath, ""
}

// innermost returns the top of the block stack (nil when at document level).
func (c *c4dCursor) innermost() *c4dBlock {
	if len(c.stack) == 0 {
		return nil
	}

	return c.stack[len(c.stack)-1]
}

// classify determines the cursor's statement kind and its derived fields.
func (c *c4dCursor) classify() {
	if m := templateRefRe.FindStringSubmatch(c.stmt); m != nil {
		c.kind = c4dStmtTemplateRef
		c.valuePrefix = m[1] // the param text already typed inside the ${...}

		return
	}

	stmt := strings.TrimSpace(c.stmt)
	if stmt == "" {
		c.kind = c4dStmtStart

		return
	}

	if glyph := c4dArrowPrefix(stmt); glyph != "" {
		c.classifyEdge(glyph, stmt)

		return
	}

	if c.classifyUse(stmt) {
		return
	}

	if prefix, ok := strings.CutPrefix(stmt, tblInclude); ok && (stmt == tblInclude || prefix[0] == ' ') {
		c.kind = c4dStmtIncludePath
		c.includePrefix = stripValueQuotes(strings.TrimSpace(prefix))

		return
	}

	if _, _, isDecl := c4dTemplateHead(stmt); isDecl {
		c.kind = c4dStmtOther // declaration head: params are being typed

		return
	}

	if c.classifyValue(stmt) {
		return
	}

	// Statement start with a partial first token — the client filters the
	// block's keyword list by it.
	if _, rest := c4dFirstToken(stmt); rest == "" {
		c.kind = c4dStmtStart

		return
	}

	c.kind = c4dStmtOther
}

// classifyEdge resolves an edge statement's cursor zone: the peer target
// before the label colon, else the freeform label (where only ${param}
// tokens complete).
func (c *c4dCursor) classifyEdge(glyph, stmt string) {
	rest := strings.TrimSpace(stmt[len(glyph):])

	if colon := indexOutsideQuotes(rest, ':'); colon >= 0 {
		c.kind = c4dStmtOther // label zone

		return
	}

	c.kind = c4dStmtEdgePeer
	c.fullPeer = c.edgePeerToken()
}

// edgePeerToken extracts the complete peer token of the edge on the
// statement's first line (D-07/D-10: identifier segments and ${param} tokens).
func (c *c4dCursor) edgePeerToken() string {
	lineNo := c.stmtLine
	if lineNo < 0 || lineNo != c.lineNo {
		return "" // hover/definition anchor to the statement's own line
	}

	_, rest := c4dFirstToken(strings.TrimSpace(c.line)) // drop the glyph
	rest = strings.TrimLeft(rest, " \t")

	return c4dScanPeerToken(rest)
}

// c4dScanPeerToken reads a leading peer reference: identifier characters,
// dotted, with embedded ${param} tokens.
func c4dScanPeerToken(s string) string {
	var b strings.Builder

	for i := 0; i < len(s); i++ {
		if strings.HasPrefix(s[i:], "${") {
			end := strings.IndexByte(s[i:], '}')
			if end < 0 {
				break
			}

			b.WriteString(s[i : i+end+1])
			i += end + 1 - 1 // advance; the loop's i++ finishes the step

			continue
		}

		ch := s[i]
		if ch == '.' || ch == '_' || ch == '-' ||
			ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' {
			b.WriteByte(ch)

			continue
		}

		break
	}

	return b.String()
}

// classifyUse resolves `use` statement zones (D-13/D-16): the template-name
// slot, then argument keys and values inside the parens.
func (c *c4dCursor) classifyUse(stmt string) bool {
	if stmt != tblUse && !strings.HasPrefix(stmt, tblUse+" ") && !strings.HasPrefix(stmt, tblUse+"(") {
		return false
	}

	rest := strings.TrimPrefix(strings.TrimPrefix(stmt, tblUse), " ")

	if idx := strings.IndexByte(rest, '('); idx < 0 {
		c.kind = c4dStmtUseName // the full name resolves from the line token

		return true
	}

	// Inside the argument list: the segment after the last top-level comma
	// is the argument being typed.
	seg := c.stmt[strings.LastIndexByte(c.stmt, '(')+1:]
	if i := strings.LastIndexByte(seg, ','); i >= 0 {
		seg = seg[i+1:]
	}

	seg = strings.TrimSpace(seg)

	if colon := indexOutsideQuotes(seg, ':'); colon >= 0 {
		c.kind = c4dStmtUseArgValue
		c.key = strings.TrimSpace(seg[:colon])
		c.valuePrefix = stripValueQuotes(strings.TrimSpace(seg[colon+1:]))

		return true
	}

	c.kind = c4dStmtUseArgKey
	c.useArgPrefix = seg

	return true
}

// classifyValue resolves a `key: value` statement: known field keys take
// their value completion; a non-reserved key where unit headers are legal is
// the header type slot (the same word-set disambiguation the PEG and the
// tree-sitter grammar make).
func (c *c4dCursor) classifyValue(stmt string) bool {
	colon := indexOutsideQuotes(stmt, ':')
	if colon < 0 {
		return false
	}

	key := strings.TrimSpace(stmt[:colon])
	c.key = key
	c.valuePrefix = stripValueQuotes(strings.TrimSpace(stmt[colon+1:]))

	switch {
	case c4dFieldKey(key):
		c.kind = c4dStmtFieldValue
	case !c4dReservedWord(key) && c4dIsIdent(key) && c4dHeaderAllowed(c.stack):
		c.kind = c4dStmtTypeSlot
	default:
		c.kind = c4dStmtOther
	}

	return true
}

// c4dFirstToken splits the leading non-whitespace token off s.
func c4dFirstToken(s string) (string, string) {
	s = strings.TrimLeft(s, " \t")

	i := 0

	for i < len(s) && s[i] != ' ' && s[i] != '\t' {
		i++
	}

	return s[:i], strings.TrimLeft(s[i:], " \t")
}

// c4dFieldKey reports whether key is a known field/option/property key (the
// value-position vocabulary).
func c4dFieldKey(key string) bool {
	switch key {
	case "description", "technology", "reference", "expanded", "name", "color",
		"style", "border", "edges", "width", "height",
		"lineLength", "legend", "legendLine",
		"arrow", "rank", "kind", "labelPosition", "length":
		return true
	default:
		return false
	}
}

// c4dHeaderAllowed reports whether a unit header may open at the cursor's
// block level (document level, unit bodies, and template bodies — the
// BodyPart/TopPart statement positions).
func c4dHeaderAllowed(stack []*c4dBlock) bool {
	if len(stack) == 0 {
		return true
	}

	blk := stack[len(stack)-1]

	return blk.kind == c4dBlockUnit || blk.kind == c4dBlockTemplate || blk.kind == c4dBlockOther
}

// hostUnitType is the declared type of the innermost enclosing unit — the
// parent type the header type slot's promotion is computed against.
func (c *c4dCursor) hostUnitType() string {
	for i := len(c.stack) - 1; i >= 0; i-- {
		if c.stack[i].kind == c4dBlockUnit {
			return c.stack[i].typ
		}
	}

	return ""
}

// scopeUnitPaths lists the unit paths the cursor's peer references resolve
// against: document units, or the enclosing template's own tree (peers inside
// template bodies resolve against the expanded instance, whose shape is the
// template tree).
func (c *c4dCursor) scopeUnitPaths() []string {
	paths := make([]string, 0, len(c.inventory.units))

	for _, u := range c.inventory.units {
		if u.tmpl == c.tmplName {
			paths = append(paths, u.path)
		}
	}

	return paths
}

// templateParams returns a template's declared parameters (nil when the
// template is undeclared).
func (c *c4dCursor) templateParams(name string) []string {
	for _, t := range c.inventory.tmpls {
		if t.id == name {
			return t.params
		}
	}

	return nil
}

// unitTargetLine finds the document unit at path and returns its header line.
func (sc *c4dScan) unitTargetLine(path string) int {
	for _, u := range sc.units {
		if u.tmpl == "" && u.path == path {
			return u.line
		}
	}

	return -1
}

// templateTargetLine finds a template declaration by name.
func (sc *c4dScan) templateTargetLine(name string) int {
	for _, t := range sc.tmpls {
		if t.id == name {
			return t.line
		}
	}

	return -1
}
