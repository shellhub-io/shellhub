package origin

import (
	"context"
	"crypto/tls"
	"maps"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTrust(t *testing.T) {
	t.Run("an empty list trusts loopback alone", func(t *testing.T) {
		trust, err := ParseTrust("")
		require.NoError(t, err)

		assert.True(t, trust.Trusts("127.0.0.1"))
		assert.True(t, trust.Trusts("::1"))
		assert.False(t, trust.Trusts("10.0.0.1"))
		assert.False(t, trust.Trusts("172.18.0.4"))
		assert.False(t, trust.Trusts("203.0.113.7"))
	})

	t.Run("a list adds its prefixes to loopback", func(t *testing.T) {
		trust, err := ParseTrust(" 10.0.0.0/8\t2001:db8::/32 ")
		require.NoError(t, err)

		assert.True(t, trust.Trusts("127.0.0.1"))
		assert.True(t, trust.Trusts("10.1.2.3"))
		assert.True(t, trust.Trusts("2001:db8::1"))
		assert.False(t, trust.Trusts("192.168.0.1"))
	})

	t.Run("an IPv4 address mapped into IPv6 matches its IPv4 prefix", func(t *testing.T) {
		trust, err := ParseTrust("10.0.0.0/8")
		require.NoError(t, err)

		assert.True(t, trust.Trusts("::ffff:10.0.0.1"))
	})

	t.Run("an IPv4 prefix written mapped into IPv6 matches its IPv4 addresses", func(t *testing.T) {
		trust, err := ParseTrust("::ffff:10.0.0.0/104")
		require.NoError(t, err)

		assert.True(t, trust.Trusts("10.0.0.5"))
		assert.True(t, trust.Trusts("::ffff:10.0.0.5"))
		assert.False(t, trust.Trusts("192.168.0.1"))
	})

	t.Run("loopback trusts what an empty list trusts", func(t *testing.T) {
		trust, err := ParseTrust("")
		require.NoError(t, err)

		for _, addr := range []string{"127.0.0.1", "::1", "10.0.0.1", "203.0.113.7", "not-an-ip"} {
			assert.Equal(t, trust.Trusts(addr), Loopback().Trusts(addr), addr)
		}
	})

	t.Run("an entry that is not a CIDR prefix is refused", func(t *testing.T) {
		for _, list := range []string{"10.0.0.1", "private_ranges", "10.0.0.0/33", "10.0.0.0/8,192.168.0.0/16"} {
			_, err := ParseTrust(list)
			require.ErrorIs(t, err, ErrInvalidTrust, list)
		}
	})

	t.Run("an address that does not parse is not trusted", func(t *testing.T) {
		trust, err := ParseTrust("0.0.0.0/0")
		require.NoError(t, err)

		assert.False(t, trust.Trusts("not-an-ip"))
	})
}

func TestResolve(t *testing.T) {
	proxied, err := ParseTrust("10.0.0.0/8")
	require.NoError(t, err)

	forwarded := http.Header{
		"X-Real-Ip":         {"203.0.113.7"},
		"X-Forwarded-Host":  {"shellhub.example.com"},
		"X-Forwarded-Proto": {"https"},
		"X-Forwarded-Port":  {"443"},
	}

	cases := []struct {
		description string
		remoteAddr  string
		host        string
		tls         bool
		localPort   int
		header      http.Header
		expected    Origin
	}{
		{
			description: "a peer outside the list is the client, whatever its headers say",
			remoteAddr:  "203.0.113.9:54321",
			host:        "shellhub.example.com:8443",
			tls:         true,
			localPort:   8443,
			header:      forwarded,
			expected:    Origin{IP: "203.0.113.9", Host: "shellhub.example.com:8443", Scheme: "https", Port: "8443"},
		},
		{
			description: "a peer outside the list on a Host without a port reached the local port",
			remoteAddr:  "203.0.113.9:54321",
			host:        "shellhub.example.com",
			localPort:   80,
			expected:    Origin{IP: "203.0.113.9", Host: "shellhub.example.com", Scheme: "http", Port: "80"},
		},
		{
			description: "a private peer outside the list is the client",
			remoteAddr:  "172.18.0.4:54321",
			host:        "server:8080",
			localPort:   8080,
			header:      forwarded,
			expected:    Origin{IP: "172.18.0.4", Host: "server:8080", Scheme: "http", Port: "8080"},
		},
		{
			description: "a peer in the list speaks for the client",
			remoteAddr:  "10.0.0.5:54321",
			host:        "server:8080",
			localPort:   8080,
			header:      forwarded,
			expected:    Origin{IP: "203.0.113.7", Host: "shellhub.example.com", Scheme: "https", Port: "443"},
		},
		{
			description: "a loopback peer speaks for the client",
			remoteAddr:  "127.0.0.1:54321",
			host:        "localhost:8080",
			localPort:   8080,
			header:      forwarded,
			expected:    Origin{IP: "203.0.113.7", Host: "shellhub.example.com", Scheme: "https", Port: "443"},
		},
		{
			description: "a peer in the list that names nothing leaves the connection's own values, without its port",
			remoteAddr:  "10.0.0.5:54321",
			host:        "shellhub.example.com",
			localPort:   8080,
			expected:    Origin{IP: "10.0.0.5", Host: "shellhub.example.com", Scheme: "http"},
		},
		{
			description: "a bracketed IPv6 client address is unwrapped",
			remoteAddr:  "10.0.0.5:54321",
			host:        "shellhub.example.com",
			header:      http.Header{"X-Real-Ip": {"[2001:db8::1]"}},
			expected:    Origin{IP: "2001:db8::1", Host: "shellhub.example.com", Scheme: "http"},
		},
		{
			description: "values that do not parse are ignored",
			remoteAddr:  "10.0.0.5:54321",
			host:        "shellhub.example.com",
			header: http.Header{
				"X-Real-Ip":         {"not-an-ip"},
				"X-Forwarded-Proto": {"gopher"},
				"X-Forwarded-Port":  {"http"},
			},
			expected: Origin{IP: "10.0.0.5", Host: "shellhub.example.com", Scheme: "http"},
		},
		{
			description: "a remote address without a port is the client's address",
			remoteAddr:  "203.0.113.9",
			host:        "shellhub.example.com",
			expected:    Origin{IP: "203.0.113.9", Host: "shellhub.example.com", Scheme: "http"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := t.Context()
			if tc.localPort != 0 {
				ctx = context.WithValue(ctx, http.LocalAddrContextKey, &net.TCPAddr{IP: net.IPv4(10, 0, 0, 1), Port: tc.localPort})
			}

			req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
			req.RemoteAddr = tc.remoteAddr
			req.Host = tc.host
			req.TLS = nil
			if tc.tls {
				req.TLS = &tls.ConnectionState{}
			}

			maps.Copy(req.Header, tc.header)

			assert.Equal(t, tc.expected, proxied.Resolve(req))
			assert.Equal(t, tc.expected.IP, proxied.ClientIP(req))
		})
	}
}

func TestContext(t *testing.T) {
	_, ok := FromContext(t.Context())
	assert.False(t, ok)

	want := Origin{IP: "203.0.113.7", Host: "shellhub.example.com", Scheme: "https", Port: "443"}
	got, ok := FromContext(NewContext(t.Context(), want))
	require.True(t, ok)
	assert.Equal(t, want, got)
}
