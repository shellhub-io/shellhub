package origin

import (
	"context"
	"crypto/tls"
	"maps"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
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

func TestParsePrefixes(t *testing.T) {
	t.Run("the list's own prefixes, masked and with IPv4 mapped into IPv6 written as IPv4", func(t *testing.T) {
		prefixes, err := ParsePrefixes(" ::ffff:10.0.0.0/104\n192.168.1.7/16\t2001:db8::/32 ")
		require.NoError(t, err)

		assert.Equal(t, []netip.Prefix{
			netip.MustParsePrefix("10.0.0.0/8"),
			netip.MustParsePrefix("192.168.0.0/16"),
			netip.MustParsePrefix("2001:db8::/32"),
		}, prefixes)
	})

	t.Run("an empty list has no prefixes, not even loopback", func(t *testing.T) {
		prefixes, err := ParsePrefixes("")
		require.NoError(t, err)

		assert.Empty(t, prefixes)
	})

	t.Run("an entry that is not a CIDR prefix is refused", func(t *testing.T) {
		_, err := ParsePrefixes("10.0.0.0/8 10.0.0.1")
		require.ErrorIs(t, err, ErrInvalidTrust)
	})
}

func TestResolve(t *testing.T) {
	proxied, err := ParseTrust("10.0.0.0/8")
	require.NoError(t, err)

	forwarded := http.Header{
		"X-Forwarded-For":   {"203.0.113.7"},
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
			description: "the client is the right-most hop outside the list, whatever the client prepended",
			remoteAddr:  "10.0.0.5:54321",
			host:        "shellhub.example.com",
			header:      http.Header{"X-Forwarded-For": {"198.51.100.66, 203.0.113.7, 10.0.0.9"}},
			expected:    Origin{IP: "203.0.113.7", Host: "shellhub.example.com", Scheme: "http"},
		},
		{
			description: "hops spread over several X-Forwarded-For fields are read in order",
			remoteAddr:  "10.0.0.5:54321",
			host:        "shellhub.example.com",
			header:      http.Header{"X-Forwarded-For": {"198.51.100.66", "203.0.113.7", "10.0.0.9"}},
			expected:    Origin{IP: "203.0.113.7", Host: "shellhub.example.com", Scheme: "http"},
		},
		{
			description: "a chain of trusted hops alone leaves the peer as the client",
			remoteAddr:  "10.0.0.5:54321",
			host:        "shellhub.example.com",
			header:      http.Header{"X-Forwarded-For": {"10.0.0.8, 127.0.0.1"}},
			expected:    Origin{IP: "10.0.0.5", Host: "shellhub.example.com", Scheme: "http"},
		},
		{
			description: "X-Real-IP does not name the client, even from a peer in the list",
			remoteAddr:  "10.0.0.5:54321",
			host:        "shellhub.example.com",
			header:      http.Header{"X-Real-Ip": {"203.0.113.7"}},
			expected:    Origin{IP: "10.0.0.5", Host: "shellhub.example.com", Scheme: "http"},
		},
		{
			description: "a hop with a port and brackets is reduced to its address",
			remoteAddr:  "10.0.0.5:54321",
			host:        "shellhub.example.com",
			header:      http.Header{"X-Forwarded-For": {"[2001:db8::1]:443, 10.0.0.9:80"}},
			expected:    Origin{IP: "2001:db8::1", Host: "shellhub.example.com", Scheme: "http"},
		},
		{
			description: "a hop with a zone is reduced to its address",
			remoteAddr:  "10.0.0.5:54321",
			host:        "shellhub.example.com",
			header:      http.Header{"X-Forwarded-For": {"fe80::1%eth0, 10.0.0.9"}},
			expected:    Origin{IP: "fe80::1", Host: "shellhub.example.com", Scheme: "http"},
		},
		{
			description: "a bracketed IPv6 hop without a port is unwrapped",
			remoteAddr:  "10.0.0.5:54321",
			host:        "shellhub.example.com",
			header:      http.Header{"X-Forwarded-For": {"[2001:db8::1]"}},
			expected:    Origin{IP: "2001:db8::1", Host: "shellhub.example.com", Scheme: "http"},
		},
		{
			description: "an IPv4 hop mapped into IPv6 is the IPv4 client",
			remoteAddr:  "10.0.0.5:54321",
			host:        "shellhub.example.com",
			header:      http.Header{"X-Forwarded-For": {"::ffff:203.0.113.7"}},
			expected:    Origin{IP: "203.0.113.7", Host: "shellhub.example.com", Scheme: "http"},
		},
		{
			description: "a hop that does not parse ends the walk at the peer, so no hop to its left counts",
			remoteAddr:  "10.0.0.5:54321",
			host:        "shellhub.example.com",
			header:      http.Header{"X-Forwarded-For": {"203.0.113.66, unknown, 10.0.0.9"}},
			expected:    Origin{IP: "10.0.0.5", Host: "shellhub.example.com", Scheme: "http"},
		},
		{
			description: "values that do not parse are ignored",
			remoteAddr:  "10.0.0.5:54321",
			host:        "shellhub.example.com",
			header: http.Header{
				"X-Forwarded-For":   {"not-an-ip, 10.0.0.9"},
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
