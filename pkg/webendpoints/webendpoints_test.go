package webendpoints

import (
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
