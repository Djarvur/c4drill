package c4d

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/Djarvur/c4drill/internal/parser"
)

// pigeonErrorPrefix matches the position prefix pigeon puts on every error
// line: `line:col (offset)` optionally followed by `: rule X`. The filename
// segment is absent because parse always invokes the generated parser with
// an empty name — the caller-supplied path travels in ParseError.Context
// instead (T-35-01-02: no content dumps, caller-controlled context only).
var pigeonErrorPrefix = regexp.MustCompile(`^(\d+):(\d+) \(\d+\)(: rule [^:]+)?: (.*)$`)

// wrapPigeonError converts the generated parser's error (an errList whose
// entries render as `line:col (offset)[: rule X]: message`) into the repo's
// standard *parser.ParseError with the DSL-native line number extracted
// from the first positioned entry (D-21). The original error stays
// reachable via ParseError.Cause/Unwrap.
func wrapPigeonError(err error, context string) error {
	line := 0

	var messages []string

	for _, l := range strings.Split(err.Error(), "\n") {
		m := pigeonErrorPrefix.FindStringSubmatch(l)
		if m == nil {
			messages = append(messages, l)

			continue
		}

		if line == 0 {
			line, _ = strconv.Atoi(m[1])
		}

		messages = append(messages, m[4])
	}

	return &parser.ParseError{
		Message: strings.Join(messages, "; "),
		Line:    line,
		Context: context,
		Cause:   err,
	}
}
