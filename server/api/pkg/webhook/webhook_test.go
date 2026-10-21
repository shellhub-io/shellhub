package webhook

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewClientSSRF(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	t.Run("blocks a loopback webhook target by default", func(t *testing.T) {
		//nolint:noctx,bodyclose // the SSRF guard rejects the dial before any request is sent, so there is no response
		_, err := NewClient(nil).Get(srv.URL)
		require.Error(t, err, "a webhook to loopback must be refused so it can't reach the host's own services")
	})

	t.Run("allows a target the operator explicitly permits", func(t *testing.T) {
		//nolint:noctx // exercising the dial path in a test
		resp, err := NewClient(ParseCIDRs("127.0.0.0/8,::1/128")).Get(srv.URL)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
	})
}

func TestParseCIDRs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		description string
		raw         string
		expected    []netip.Prefix
	}{
		{description: "empty is no prefix", raw: "", expected: nil},
		{
			description: "trims and splits on commas",
			raw:         " 10.0.0.0/8 , ::1/128 ",
			expected:    []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8"), netip.MustParsePrefix("::1/128")},
		},
		{
			description: "skips empty and invalid entries",
			raw:         "10.0.0.0/8,,not-a-cidr,192.168.0.0/16",
			expected:    []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8"), netip.MustParsePrefix("192.168.0.0/16")},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tc.expected, ParseCIDRs(tc.raw))
		})
	}
}
