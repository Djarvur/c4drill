package c4d

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/Djarvur/c4drill/internal/parser"
)

// pigeonErrorPrefix matches the position prefix pigeon puts on every error
// line: `line:col (offset)` optionally followed by `: rule X`. A named
// parse (ParseFile/ParseNamed pass the caller's path into grammar.Parse)
// renders the name in front — wrapPigeonError strips that exact prefix
// first, so the positioned form matches for both named and unnamed parses.
var pigeonErrorPrefix = regexp.MustCompile(`^(\d+):(\d+) \(\d+\)(: rule [^:]+)?: (.*)$`)

// nestedParseErrorPrefix matches the rendering a *parser.ParseError adds
// when an action-level check (grammar/reserved.go, D-19) is recorded inside
// pigeon's error list — stripped so the final message is not doubled.
var nestedParseErrorPrefix = regexp.MustCompile(`^parse error at line \d+: `)

// maxExpressionsMsg is pigeon's errMaxExprCnt text — the expression-budget
// failure. It is a resource cap tripped by volume, not a syntax verdict, so
// the wrapped message spells out the current limit and the way out: to a
// user it otherwise reads as a syntax error on the file's last line (the
// budget trips at the furthest position reached).
const maxExpressionsMsg = "max number of expressions parsed"

// wrapPigeonError converts the generated parser's error (an errList whose
// entries render as `line:col (offset)[: rule X]: message`) into the repo's
// standard *parser.ParseError with the DSL-native line number extracted
// from the first positioned entry (D-21). The original error stays
// reachable via ParseError.Cause/Unwrap.
func wrapPigeonError(err error, context string) error {
	line := 0

	var messages []string

	for _, l := range strings.Split(err.Error(), "\n") {
		// A named parse prefixes every pigeon line with `name:` (the
		// caller's path); strip it so the positioned-prefix regex matches.
		// An empty name renders no prefix (ParseAST).
		if context != "" {
			l = strings.TrimPrefix(l, context+":")
		}

		m := pigeonErrorPrefix.FindStringSubmatch(l)
		if m == nil {
			messages = append(messages, l)

			continue
		}

		if line == 0 {
			line, _ = strconv.Atoi(m[1])
		}

		messages = append(messages, nestedParseErrorPrefix.ReplaceAllString(m[4], ""))
	}

	for i, msg := range messages {
		if msg == maxExpressionsMsg {
			messages[i] = fmt.Sprintf(
				"%s: the expression budget (%d) is a resource cap against runaway parses, "+
					"not a syntax limit — raise it with --max-expressions (0 disables the cap)",
				maxExpressionsMsg, MaxExpressions)
		}
	}

	return &parser.ParseError{
		Message: strings.Join(messages, "; "),
		Line:    line,
		Context: context,
		Cause:   err,
	}
}
