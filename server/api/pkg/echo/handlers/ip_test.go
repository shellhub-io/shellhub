package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestRealIPExtractor(t *testing.T) {
	cases := []struct {
		description string
		remoteAddr  string
		realIP      string
		expected    string
	}{
		{
			description: "takes the header when the gateway forwards a public client",
			remoteAddr:  "172.18.0.4:54321",
			realIP:      "203.0.113.7",
			expected:    "203.0.113.7",
		},
		{
			description: "takes the header when the peer is loopback",
			remoteAddr:  "127.0.0.1:54321",
			realIP:      "203.0.113.7",
			expected:    "203.0.113.7",
		},
		{
			description: "unwraps a bracketed IPv6 header",
			remoteAddr:  "172.18.0.4:54321",
			realIP:      "[2001:db8::1]",
			expected:    "2001:db8::1",
		},
		{
			description: "ignores the header when the peer is not trusted",
			remoteAddr:  "203.0.113.9:54321",
			realIP:      "198.51.100.1",
			expected:    "203.0.113.9",
		},
		{
			description: "falls back to the peer when the header is absent",
			remoteAddr:  "172.18.0.4:54321",
			expected:    "172.18.0.4",
		},
		{
			description: "falls back to the peer when the header is not an address",
			remoteAddr:  "172.18.0.4:54321",
			realIP:      "not-an-ip",
			expected:    "172.18.0.4",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			e := echo.New()
			e.IPExtractor = RealIPExtractor()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tc.remoteAddr

			if tc.realIP != "" {
				req.Header.Set(echo.HeaderXRealIP, tc.realIP)
			}

			assert.Equal(t, tc.expected, e.NewContext(req, httptest.NewRecorder()).RealIP())
		})
	}
}
