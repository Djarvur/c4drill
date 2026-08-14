// Package tomlfmt implements the TOML half of `c4drill fmt` (D-31/D-32): a
// comment-preserving, gofmt-style formatter for c4drill TOML documents.
//
// Formatting normalizes ONLY whitespace, indentation and blank-line
// grouping:
//
//   - every statement starts at column 0 (leading indent stripped);
//   - exactly one space around '=' in key/value pairs;
//   - table headers render tightly bracketed ([key] / [[key]]);
//   - a run of blank lines collapses to exactly one blank line;
//   - a same-line trailing comment renders one space after the statement.
//
// Everything else is preserved VERBATIM: key order inside tables stays the
// AUTHOR's order (deliberate contrast with convert's D-23 canonical field
// order — canonical ordering applies to emitted conversions, never to fmt),
// and value spelling rides the raw source bytes (quoting style, multi-line
// strings/arrays and any comments inside them pass through untouched).
//
// The formatter walks the go-toml/v2 unstable API with KeepComments enabled
// — the same API internal/parser uses for definition-order capture, now
// reading the comment nodes it exposes with source positions. Malformed
// input is a hard error and Format returns no bytes for it: the formatter
// fails closed, never rewriting what it cannot fully re-render (D-31).
package tomlfmt

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/pelletier/go-toml/v2/unstable"
)

// Static errors for better error handling.
var (
	// errMalformedTOML wraps every parse failure — fmt never rewrites
	// input it cannot fully understand.
	errMalformedTOML = errors.New("malformed TOML")
	// errUnsupportedExpr guards the top-level expression-kind switch.
	errUnsupportedExpr = errors.New("unsupported TOML expression kind")
)

// fmtEvent is one rendered statement in document order: the source byte
// span it occupies (used to detect the author's blank-line grouping) and
// its rendered text (never containing a newline terminator).
type fmtEvent struct {
	start int
	end   int
	text  string
}

// minBlankGapNewlines is the newline count in a source gap between two
// events that renders as a blank line: the first newline closes the
// previous statement's line, so two or more mean the author left at least
// one blank line (longer runs collapse to one — that is the normalization).
const minBlankGapNewlines = 2

// Format normalizes data as c4drill TOML (D-31/D-32): comments, key order
// and value spelling survive exactly; whitespace, indentation and
// blank-line grouping normalize deterministically. Format(Format(x)) ==
// Format(x) for every input, and parser.Parse(Format(x)) equals
// parser.Parse(x) — formatting never changes semantics.
func Format(data []byte) ([]byte, error) {
	events, err := collectEvents(data)
	if err != nil {
		return nil, err
	}

	return render(data, events), nil
}

// collectEvents parses data into rendered statements in document order.
// The unstable arena resets on every NextExpression call, so every string
// and offset is copied out before advancing — no node pointers escape.
func collectEvents(data []byte) ([]fmtEvent, error) {
	p := &unstable.Parser{KeepComments: true}
	p.Reset(data)

	events := make([]fmtEvent, 0, 64)

	for p.NextExpression() {
		expr := p.Expression()

		var (
			ev  fmtEvent
			err error
		)

		switch expr.Kind { //nolint:exhaustive // default fail-closes every non-top-level kind
		case unstable.Comment:
			ev = commentEvent(expr)
		case unstable.Table, unstable.ArrayTable:
			ev, err = tableEvent(p, expr)
		case unstable.KeyValue:
			ev, err = keyValueEvent(p, expr, data)
		default:
			return nil, fmt.Errorf("%w: %s", errUnsupportedExpr, expr.Kind)
		}

		if err != nil {
			return nil, err
		}

		// The parser's finishLine attaches a same-line trailing comment as
		// the expression root's next sibling — at most one, always on the
		// statement's own line.
		if tail := expr.Next(); tail != nil && tail.Kind == unstable.Comment {
			ev.text += " " + strings.TrimSpace(string(tail.Data))
			ev.end = rangeEnd(tail.Raw)
		}

		events = append(events, ev)
	}

	if err := p.Error(); err != nil {
		return nil, fmt.Errorf("%w: %w", errMalformedTOML, err)
	}

	slices.SortStableFunc(events, func(a, b fmtEvent) int {
		return cmp.Compare(a.start, b.start)
	})

	return events, nil
}

// commentEvent renders a standalone comment expression (its own line).
func commentEvent(expr *unstable.Node) fmtEvent {
	return fmtEvent{
		start: int(expr.Raw.Offset),
		end:   rangeEnd(expr.Raw),
		text:  strings.TrimSpace(string(expr.Data)),
	}
}

// tableEvent renders a [table] or [[array-of-tables]] header from its key
// segments' raw bytes (quoted segments keep their quotes verbatim). The
// header span is measured over the key segments: the brackets live on the
// same line, so blank-line detection is unaffected.
func tableEvent(p *unstable.Parser, expr *unstable.Node) (fmtEvent, error) {
	var (
		segments []string
		start    = -1
		end      int
	)

	for it := expr.Children(); it.Next(); {
		n := it.Node()
		if n.Kind != unstable.Key {
			continue
		}

		segments = append(segments, string(p.Raw(n.Raw)))

		if start < 0 {
			start = int(n.Raw.Offset)
		}

		end = rangeEnd(n.Raw)
	}

	if len(segments) == 0 {
		return fmtEvent{}, fmt.Errorf("%w: table header without a key", errMalformedTOML)
	}

	text := "[" + strings.Join(segments, ".") + "]"
	if expr.Kind == unstable.ArrayTable {
		text = "[" + text + "]"
	}

	return fmtEvent{start: start, end: end, text: text}, nil
}

// keyValueEvent renders `dotted.key = value`: key segments from their raw
// bytes, exactly one space around '=', and the value exactly as authored
// (raw bytes from after the '=' to the end of the expression's span).
func keyValueEvent(p *unstable.Parser, expr *unstable.Node, data []byte) (fmtEvent, error) {
	value := expr.Value()
	if !value.Valid() {
		return fmtEvent{}, fmt.Errorf("%w: key-value without a value", errMalformedTOML)
	}

	var (
		segments []string
		keyStart = -1
		keyEnd   int
	)

	for it := expr.Key(); it.Next(); {
		n := it.Node()
		if n.Kind != unstable.Key {
			continue
		}

		segments = append(segments, string(p.Raw(n.Raw)))

		if keyStart < 0 {
			keyStart = int(n.Raw.Offset)
		}

		keyEnd = rangeEnd(n.Raw)
	}

	if len(segments) == 0 {
		return fmtEvent{}, fmt.Errorf("%w: key-value without a key", errMalformedTOML)
	}

	raw, err := rawValue(data, expr, keyEnd)
	if err != nil {
		return fmtEvent{}, err
	}

	return fmtEvent{
		start: keyStart,
		end:   rangeEnd(expr.Raw),
		text:  strings.Join(segments, ".") + " = " + string(raw),
	}, nil
}

// rawValue slices the value bytes exactly as authored: everything from
// after the '=' to the end of the expression's raw span (which the parser
// closes right after the value token). Quoting style, multi-line
// strings/arrays and any comments inside them are inside this span and pass
// through VERBATIM — the formatter normalizes statement layout, never value
// spelling.
func rawValue(data []byte, expr *unstable.Node, keyEnd int) ([]byte, error) {
	end := min(rangeEnd(expr.Raw), len(data))

	i := keyEnd
	for i < end && (data[i] == ' ' || data[i] == '\t') {
		i++
	}

	if i >= end || data[i] != '=' {
		return nil, fmt.Errorf("%w: missing '=' in key-value", errMalformedTOML)
	}

	i++
	for i < end && (data[i] == ' ' || data[i] == '\t') {
		i++
	}

	if i >= end {
		return nil, fmt.Errorf("%w: key-value without a value", errMalformedTOML)
	}

	return data[i:end], nil
}

// render writes the events in document order, preserving the author's
// blank-line grouping: a source gap of two or more newlines between
// consecutive events renders as exactly one blank line. The output always
// ends with exactly one newline.
func render(data []byte, events []fmtEvent) []byte {
	var b strings.Builder

	for i, ev := range events {
		if i > 0 && blankBetween(data, events[i-1].end, ev.start) {
			b.WriteByte('\n')
		}

		b.WriteString(ev.text)
		b.WriteByte('\n')
	}

	return []byte(b.String())
}

// blankBetween reports whether the source gap between two event spans
// contains a blank line. Overlapping or out-of-order spans (never produced
// by collectEvents) count as adjacent.
func blankBetween(data []byte, prevEnd, currStart int) bool {
	if prevEnd < 0 || prevEnd >= currStart || currStart > len(data) {
		return false
	}

	return bytes.Count(data[prevEnd:currStart], []byte{'\n'}) >= minBlankGapNewlines
}

// rangeEnd returns the exclusive end offset of an unstable byte range.
func rangeEnd(r unstable.Range) int {
	return int(r.Offset) + int(r.Length)
}
