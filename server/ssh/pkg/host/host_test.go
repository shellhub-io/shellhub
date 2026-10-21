package host

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsLocalHost(t *testing.T) {
	cases := []struct {
		description string
		host        *Host
		expected    bool
	}{
		{
			description: "returns false for a generic hostname",
			host:        &Host{"host"},
			expected:    false,
		},
		{
			description: "returns true for IPv4 localhost address",
			host:        &Host{"127.0.0.1"},
			expected:    true,
		},
		{
			description: "returns true for IPv6 localhost address",
			host:        &Host{"::1"},
			expected:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			isLocalHost := tc.host.IsLocalhost()
			assert.Equal(t, tc.expected, isLocalHost)
		})
	}
}
