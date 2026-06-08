package http

import (
	"testing"
)

func TestConfigWebEndpointsDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		description        string
		webEndpointsDomain string
		domain             string
		expected           string
	}{
		{
			description:        "WebEndpointsDomain takes priority over Domain",
			webEndpointsDomain: "cloud.example",
			domain:             "example",
			expected:           "cloud.example",
		},
		{
			description:        "falls back to Domain when WebEndpointsDomain is empty",
			webEndpointsDomain: "",
			domain:             "example",
			expected:           "example",
		},
		{
			description:        "returns empty string when both are empty",
			webEndpointsDomain: "",
			domain:             "",
			expected:           "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()

			cfg := &Config{
				WebEndpointsDomain: tt.webEndpointsDomain,
				Domain:             tt.domain,
			}

			got := cfg.webEndpointsDomain()
			if got != tt.expected {
				t.Errorf("webEndpointsDomain() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestConfigWebEndpointHost(t *testing.T) {
	t.Parallel()

	tests := []struct {
		description        string
		webEndpointsDomain string
		domain             string
		address            string
		expected           string
	}{
		{
			description:        "address joined with WebEndpointsDomain when it is set",
			webEndpointsDomain: "cloud.example",
			domain:             "",
			address:            "abc123",
			expected:           "abc123.cloud.example",
		},
		{
			description:        "address joined with Domain when WebEndpointsDomain is empty",
			webEndpointsDomain: "",
			domain:             "example",
			address:            "abc123",
			expected:           "abc123.example",
		},
		{
			description:        "regression guard: no trailing dot when both domain fields are empty",
			webEndpointsDomain: "",
			domain:             "",
			address:            "abc123",
			expected:           "abc123",
		},
		{
			description:        "empty address with both domain fields empty returns empty string",
			webEndpointsDomain: "",
			domain:             "",
			address:            "",
			expected:           "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			t.Parallel()

			cfg := &Config{
				WebEndpointsDomain: tt.webEndpointsDomain,
				Domain:             tt.domain,
			}

			got := cfg.webEndpointHost(tt.address)
			if got != tt.expected {
				t.Errorf("webEndpointHost(%q) = %q, want %q", tt.address, got, tt.expected)
			}
		})
	}
}
