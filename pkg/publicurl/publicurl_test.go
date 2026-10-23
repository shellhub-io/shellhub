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
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()

			got, err := Resolve(tc.domain, tc.autoSSL)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, got.String())
		})
	}
}

func TestResolveRejectsADomainThatIsNotAURL(t *testing.T) {
	t.Parallel()

	for _, domain := range []string{"shellhub.example.com:port", "https://"} {
		_, err := Resolve(domain, false)
		assert.True(t, errors.Is(err, ErrInvalidDomain), domain)
	}
}
