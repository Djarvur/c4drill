// semantictokens.go implements textDocument/semanticTokens/full for the
// c4drill TOML dialect — the piece of #32 explicitly deferred to #33. The
// tokens carry the c4drill-SPECIFIC semantics a generic TOML grammar cannot
// know:
//
//   - `type` keys in unit sections and template subtrees, plus their
//     unit-type keyword values (property / class) — the value token sits
//     inside the quotes;
//   - link tables: the `link`/`linkFrom` header segment and the link field
//     keys inside [[...link]] / [[...linkFrom]] sections (property);
//   - enum values wherever the known keys take them — edges, arrow, rank,
//     kind, style, labelPosition (enumMember).
//
// Like the other language features, the analysis runs on the CURRENT buffer
// via tomlctx.go's line analyzer, so tokens survive mid-edit states.

package lsp

import (
	"path/filepath"
	"sort"
	"strings"
)

// semTokenTypes is the legend the server advertises; indexes are wire
// token-type ids.
//
//nolint:gochecknoglobals // pinned legend (protocol constant table)
var semTokenTypes = []string{
	SemTokenTypeProperty,
	SemTokenTypeClass,
	SemTokenTypeEnumMember,
}

// legend indexes into semTokenTypes.
const (
	semPropIdx = iota
	semClassIdx
	semEnumIdx
)

// semToken is one unencoded token: absolute line and character plus the
// UTF-16 length and legend index.
type semToken struct {
	line    uint32
	char    uint32
	length  uint32
	typeIdx uint32
}

// semEnumValues are the enum vocabularies keyed by field key.
//
//nolint:gochecknoglobals // pinned closed vocabularies (model enums)
var semEnumValues = map[string]map[string]bool{
	"edges":         setOf("straight", "spline", "square", "ortho"),
	"arrow":         setOf("forward", "reverse", "bidirectional", "none"),
	"rank":          setOf("forward", "reverse", "equal"),
	"kind":          setOf("read", "write", "read-write"),
	"style":         setOf("solid", "dashed", "dotted"),
	"labelPosition": setOf("middle", "tail", "head"),
}

// setOf builds a literal set.
func setOf(values ...string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, v := range values {
		set[v] = true
	}

	return set
}

// semanticTokens is the textDocument/semanticTokens/full feature entry.
func (s *Server) semanticTokens(doc *document) SemanticTokens {
	if filepath.Ext(doc.Path) != extToml {
		return SemanticTokens{Data: []uint32{}}
	}

	tokens := semTokensOf(string(doc.Text))
	if len(tokens) == 0 {
		return SemanticTokens{Data: []uint32{}}
	}

	return SemanticTokens{Data: encodeSemTokens(tokens)}
}

// semTokensOf walks the TOML buffer line by line and collects the semantic
// tokens in document order.
func semTokensOf(text string) []semToken {
	var tokens []semToken

	current := ""    // enclosing table path
	isArray := false // [[array-of-tables]] header

	for lineNo, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)

		switch {
		case isTableHeader(trimmed):
			current = headerTablePath(trimmed)
			isArray = strings.HasPrefix(trimmed, "[[")

			if t, ok := linkHeaderToken(trimmed, line, lineNo); ok {
				tokens = append(tokens, t)
			}

			continue
		case trimmed == "" || strings.HasPrefix(trimmed, "#"):
			continue
		}

		eq := indexOutsideQuotes(line, '=')
		if eq < 0 {
			continue
		}

		key := strings.TrimSpace(line[:eq])

		tokens = append(tokens, semValueTokens(current, isArray, key, line, eq, lineNo)...)
	}

	return tokens
}

// semValueTokens emits the tokens for one `key = value` line against its
// enclosing section.
func semValueTokens(tablePath string, isArray bool, key, line string, eq, lineNo int) []semToken {
	var tokens []semToken

	unitTable := !isArray && tablePath != "" && isUnitTablePath(tablePath)
	_, isTemplate := templateOfTable(tablePath)
	templateTable := !isArray && isTemplate
	linkTable := isArray && isLinkTablePath(tablePath)

	typeKey := key == "type" && (unitTable || templateTable)

	if typeKey {
		tokens = append(tokens, keyToken(key, line, lineNo, semPropIdx))
	}

	if linkTable && linkFieldKey(key) {
		tokens = append(tokens, keyToken(key, line, lineNo, semPropIdx))
	}

	if content, char, good := quotedValueSpan(line, eq); good {
		switch {
		case typeKey && c4dTypeKeyword(content):
			tokens = append(tokens, spanToken(lineNo, line, char, content, semClassIdx))
		case semEnumValues[key][content]:
			tokens = append(tokens, spanToken(lineNo, line, char, content, semEnumIdx))
		}
	}

	return tokens
}

// keyToken emits a property token spanning the key text.
func keyToken(key, line string, lineNo int, typeIdx uint32) semToken {
	char := 0

	if i := strings.Index(line, key); i >= 0 {
		char = i
	}

	return semToken{
		line:    uint32(lineNo),                  //nolint:gosec // line indices are non-negative
		char:    uint32(utf16Units(line[:char])), //nolint:gosec // bounded by line length
		length:  uint32(utf16Units(key)),         //nolint:gosec // bounded by line length
		typeIdx: typeIdx,
	}
}

// spanToken emits a token spanning value content at character offset char.
func spanToken(lineNo int, line string, char int, content string, typeIdx uint32) semToken {
	return semToken{
		line:    uint32(lineNo),                  //nolint:gosec // line indices are non-negative
		char:    uint32(utf16Units(line[:char])), //nolint:gosec // bounded by line length
		length:  uint32(utf16Units(content)),     //nolint:gosec // bounded by line length
		typeIdx: typeIdx,
	}
}

// linkHeaderToken emits a property token over the link/linkFrom suffix of a
// [[...]] header line.
func linkHeaderToken(trimmed, line string, lineNo int) (semToken, bool) {
	if !strings.HasPrefix(trimmed, "[[") {
		return semToken{}, false
	}

	inner := strings.TrimSuffix(strings.TrimPrefix(trimmed, "[["), "]]")
	if !isLinkTablePath(normalizeTablePath(inner)) {
		return semToken{}, false
	}

	seg := inner
	if dot := strings.LastIndexByte(inner, '.'); dot >= 0 {
		seg = inner[dot+1:]
	}

	if seg != tblLink && seg != tblLinkFrom {
		return semToken{}, false
	}

	idx := strings.LastIndex(line, seg)
	if idx < 0 {
		return semToken{}, false
	}

	return semToken{
		line:    uint32(lineNo),                 //nolint:gosec // line indices are non-negative
		char:    uint32(utf16Units(line[:idx])), //nolint:gosec // bounded by line length
		length:  uint32(utf16Units(seg)),        //nolint:gosec // bounded by line length
		typeIdx: semPropIdx,
	}, true
}

// isLinkTablePath reports an [[...link]]/[[...linkFrom]] table path.
func isLinkTablePath(path string) bool {
	dot := strings.LastIndexByte(path, '.')
	if dot < 0 {
		return false
	}

	switch path[dot+1:] {
	case tblLink, tblLinkFrom:
		return true
	default:
		return false
	}
}

// linkFieldKey reports the [[link]] entry key set.
func linkFieldKey(key string) bool {
	switch key {
	case "peer", "arrow", "rank", "kind", "color", "style",
		"technology", "description", "labelPosition", "length":
		return true
	default:
		return false
	}
}

// quotedValueSpan extracts the content and character offset of a quoted
// string value after the '=' at eq; ok is false when the value is not a
// complete quoted string (barewords and mid-edit quotes get no tokens).
func quotedValueSpan(line string, eq int) (string, int, bool) {
	rest := line[eq+1:]

	i := 0

	for i < len(rest) && (rest[i] == ' ' || rest[i] == '\t') {
		i++
	}

	if i >= len(rest) || rest[i] != '"' {
		return "", 0, false
	}

	start := i + 1

	for j := start; j < len(rest); j++ {
		switch {
		case rest[j] == '\\' && j+1 < len(rest):
			j++
		case rest[j] == '"':
			return rest[start:j], eq + 1 + start, true
		case rest[j] == '\n':
			return "", 0, false
		}
	}

	return "", 0, false // unterminated mid-edit: no token
}

// encodeSemTokens sorts the tokens in document order and delta-encodes them
// into the LSP quintuple array (lengths in UTF-16 code units).
func encodeSemTokens(tokens []semToken) []uint32 {
	sorted := make([]semToken, len(tokens))
	copy(sorted, tokens)

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].line != sorted[j].line {
			return sorted[i].line < sorted[j].line
		}

		return sorted[i].char < sorted[j].char
	})

	data := make([]uint32, 0, len(sorted)*5)

	var line, char uint32

	for _, t := range sorted {
		deltaLine := t.line - line
		deltaChar := t.char

		if deltaLine == 0 {
			deltaChar = t.char - char
		}

		data = append(data, deltaLine, deltaChar, t.length, t.typeIdx, 0)

		line, char = t.line, t.char
	}

	return data
}
