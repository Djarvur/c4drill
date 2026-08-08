package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHumanize is the ERGO-04 contract test. The cases below are the
// D-01 reference table (phase 29 CONTEXT.md) — all 9 rows are mandatory and
// must pass verbatim. Acronym preservation is an ANTI-feature (ERGO-04): the
// gRPC → "Grpc" and IDPToken → "Idp Token" rows pin the dumb-split behavior.
func TestHumanize(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want string
	}{
		// D-01 reference table (9 mandatory rows) — DO NOT alter or remove.
		{name: "linuxSystem", in: "linuxSystem", want: "Linux System"},
		{name: "localIDP", in: "localIDP", want: "Local IDP"},
		{name: "sessionManager", in: "sessionManager", want: "Session Manager"},
		{name: "sessionAPI", in: "sessionAPI", want: "Session API"},
		{name: "gRPC", in: "gRPC", want: "Grpc"},
		{name: "grpcAPIs", in: "grpcAPIs", want: "Grpc Apis"},
		{name: "webapp", in: "webapp", want: "Webapp"},
		{name: "IDPToken", in: "IDPToken", want: "Idp Token"},
		{name: "empty", in: "", want: ""},

		// Edge cases discovered during implementation — additional coverage.
		{name: "allCapsAPI", in: "API", want: "Api"},
		{name: "singleLowerA", in: "a", want: "A"},
		{name: "singleUpperZ", in: "Z", want: "Z"},
		{name: "alreadySpaced", in: "Already Space", want: "Already Space"},
		{name: "pascalCaseTwoWords", in: "WebServer", want: "Web Server"},
		{name: "consecutiveCapsRun", in: "HTTPSClient", want: "Https Client"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, c.want, Humanize(c.in), "Humanize(%q)", c.in)
		})
	}
}
