package host

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHost(t *testing.T) {
	cases := []struct {
		description string
		address     string
		expected    *Host
	}{
		{
			description: "fails when address is empty",
			address:     "",
			expected:    nil,
		},
		{
			description: "fails when address does not contain any port",
			address:     "192.168.0.1",
			expected:    nil,
		},
		{
			description: "succeeds when address contains an IPv4 and port",
			address:     "192.168.0.1:8080",
			expected:    &Host{"192.168.0.1"},
		},
		{
			description: "fails when IPv6 address does not contain any port",
			address:     "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			expected:    nil,
		},
		{
			description: "succeeds when address contains an IPv6 and port",
			address:     "[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:8080",
			expected:    &Host{"2001:0db8:85a3:0000:0000:8a2e:0370:7334"},
		},
		{
			description: "fails when the address is neither IPv4 nor IPv6 and does not contain any port",
			address:     "example.com",
			expected:    nil,
		},
		{
			description: "succeeds when the address is a domain name with port",
			address:     "example.com:8080",
			expected:    &Host{"example.com"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			host, _ := NewHost(tc.address)
			assert.Equal(t, tc.expected, host)
		})
	}
}

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
