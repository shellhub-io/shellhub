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
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
		require.NoError(t, err)

		_, err = NewClient(nil).Do(req) //nolint:bodyclose // the SSRF guard rejects the dial before any request is sent, so there is no response
		require.Error(t, err, "a webhook to loopback must be refused so it can't reach the host's own services")
	})

	t.Run("allows a target the operator explicitly permits", func(t *testing.T) {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
		require.NoError(t, err)

		resp, err := NewClient(ParseCIDRs("127.0.0.0/8,::1/128")).Do(req)
		require.NoError(t, err)
		require.NoError(t, resp.Body.Close())
	})
}

func TestSign(t *testing.T) {
	t.Parallel()

	require.Equal(t, "22b0f73be999eaa51329b3d06e101dc4cd11f330c696b9000c0f7c05b9f65dc2", Sign("s3cr3t", []byte("body")))
}

func TestSignV1(t *testing.T) {
	t.Parallel()

	require.Equal(t, "v1=44179dce73147efe36b07b18dbe5bc64d2883e10e723dec9ee95914290939de2", SignV1("s3cr3t", 1700000000, []byte("body")))
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
