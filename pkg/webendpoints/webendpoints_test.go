package webendpoints

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDomain(t *testing.T) {
	cases := []struct {
		description string
		preferred   string
		fallback    string
		expected    string
	}{
		{
			description: "returns preferred when preferred is set",
			preferred:   "cloud.example",
			fallback:    "example",
			expected:    "cloud.example",
		},
		{
			description: "returns fallback when preferred is empty",
			preferred:   "",
			fallback:    "example",
			expected:    "example",
		},
		{
			description: "returns empty string when both are empty",
			preferred:   "",
			fallback:    "",
			expected:    "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expected, Domain(tc.preferred, tc.fallback))
		})
	}
}

func TestAddressFromHost(t *testing.T) {
	const valid = "0123456789abcdef0123456789abcdef"

	cases := []struct {
		description string
		host        string
		domain      string
		expected    string
		found       bool
	}{
		{
			description: "extracts the address from a web endpoint host",
			host:        valid + ".cloud.example",
			domain:      "cloud.example",
			expected:    valid,
			found:       true,
		},
		{
			description: "ignores the port the client connected to",
			host:        valid + ".cloud.example:8080",
			domain:      "cloud.example",
			expected:    valid,
			found:       true,
		},
		{
			description: "accepts a fully qualified host and a different case",
			host:        strings.ToUpper(valid) + ".Cloud.Example.",
			domain:      "cloud.example",
			expected:    valid,
			found:       true,
		},
		{
			description: "rejects the domain itself",
			host:        "cloud.example",
			domain:      "cloud.example",
		},
		{
			description: "rejects a label that is not an address",
			host:        "console.cloud.example",
			domain:      "cloud.example",
		},
		{
			description: "rejects an address of the wrong length",
			host:        "abc123.cloud.example",
			domain:      "cloud.example",
		},
		{
			description: "rejects a deeper subdomain",
			host:        "extra." + valid + ".cloud.example",
			domain:      "cloud.example",
		},
		{
			description: "rejects a domain that only shares a suffix",
			host:        valid + ".evilcloud.example",
			domain:      "cloud.example",
		},
		{
			description: "rejects every host when no domain is configured",
			host:        valid,
			domain:      "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			addr, found := AddressFromHost(tc.host, tc.domain)

			assert.Equal(t, tc.found, found)
			assert.Equal(t, tc.expected, addr)
		})
	}
}

func TestHost(t *testing.T) {
	cases := []struct {
		description string
		address     string
		domain      string
		expected    string
	}{
		{
			description: "returns address.domain when domain is set",
			address:     "abc123",
			domain:      "cloud.example",
			expected:    "abc123.cloud.example",
		},
		{
			description: "returns address.domain with simple domain",
			address:     "abc123",
			domain:      "example",
			expected:    "abc123.example",
		},
		{
			description: "returns address only when domain is empty (no trailing dot)",
			address:     "abc123",
			domain:      "",
			expected:    "abc123",
		},
		{
			description: "returns empty string when both address and domain are empty",
			address:     "",
			domain:      "",
			expected:    "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expected, Host(tc.address, tc.domain))
		})
	}
}
