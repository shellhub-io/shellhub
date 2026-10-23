package publicurl

import (
	"testing"

	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolve(t *testing.T) {
	t.Parallel()

	cases := []struct {
		description string
		domain      string
		autoSSL     bool
		override    string
		expected    string
	}{
		{
			description: "uses http when automatic SSL is off",
			domain:      "shellhub.example.com",
			expected:    "http://shellhub.example.com",
		},
		{
			description: "uses https when automatic SSL is on",
			domain:      "shellhub.example.com",
			autoSSL:     true,
			expected:    "https://shellhub.example.com",
		},
		{
			description: "keeps the scheme the domain carries",
			domain:      "https://shellhub.example.com",
			expected:    "https://shellhub.example.com",
		},
		{
			description: "drops the domain's trailing slash",
			domain:      "https://shellhub.example.com/",
			expected:    "https://shellhub.example.com",
		},
		{
			description: "prefers the override when TLS ends before the gateway",
			domain:      "shellhub.example.com",
			override:    "https://shellhub.example.com",
			expected:    "https://shellhub.example.com",
		},
		{
			description: "drops the override's trailing slash",
			domain:      "shellhub.example.com",
			autoSSL:     true,
			override:    "https://shellhub.example.com:8443/",
			expected:    "https://shellhub.example.com:8443",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()

			got, err := Resolve(tc.domain, tc.autoSSL, tc.override)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, got.String())
		})
	}
}

func TestResolveRejectsADomainThatIsNotAURL(t *testing.T) {
	t.Parallel()

	for _, domain := range []string{"shellhub.example.com:port", "https://"} {
		_, err := Resolve(domain, false, "")
		assert.True(t, errors.Is(err, ErrInvalidDomain), domain)
	}
}

func TestResolveRejectsAnOverrideThatIsNotAnOrigin(t *testing.T) {
	t.Parallel()

	cases := []struct {
		description string
		override    string
	}{
		{description: "no scheme", override: "shellhub.example.com"},
		{description: "not http", override: "ftp://shellhub.example.com"},
		{description: "no host", override: "https://:8443"},
		{description: "empty port", override: "https://shellhub.example.com:"},
		{description: "bad port", override: "https://shellhub.example.com:port"},
		{description: "port zero", override: "https://shellhub.example.com:0"},
		{description: "port out of range", override: "https://shellhub.example.com:99999"},
		{description: "userinfo", override: "https://user@shellhub.example.com"},
		{description: "path", override: "https://shellhub.example.com/prefix"},
		{description: "query", override: "https://shellhub.example.com?a=b"},
		{description: "fragment", override: "https://shellhub.example.com#a"},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()

			_, err := Resolve("shellhub.example.com", false, tc.override)
			assert.True(t, errors.Is(err, ErrInvalidOverride))
		})
	}
}
