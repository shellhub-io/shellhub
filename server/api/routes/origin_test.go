package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/server/api/pkg/origin"
	"github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouterResolvesTheClientThroughTheTrustedProxies(t *testing.T) {
	trusted, err := origin.ParseTrust("10.0.0.0/8")
	require.NoError(t, err)

	cases := []struct {
		description string
		opts        []Option
		peer        string
		expected    string
	}{
		{
			description: "with no trusted proxy, a private peer is the client",
			peer:        "172.18.0.4:54321",
			expected:    "172.18.0.4",
		},
		{
			description: "with no trusted proxy, a loopback peer speaks for the client",
			peer:        "127.0.0.1:54321",
			expected:    "203.0.113.7",
		},
		{
			description: "a listed peer speaks for the client",
			opts:        []Option{WithTrustedProxies(trusted)},
			peer:        "10.0.0.5:54321",
			expected:    "203.0.113.7",
		},
		{
			description: "a peer outside the list is the client",
			opts:        []Option{WithTrustedProxies(trusted)},
			peer:        "172.18.0.4:54321",
			expected:    "172.18.0.4",
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			router := newTestRouter(t, mocks.NewMockService(t), tc.opts...)

			var fromContext origin.Origin
			var realIP string
			router.GET("/probe", func(c *echo.Context) error {
				fromContext, _ = origin.FromContext(c.Request().Context())
				realIP = c.RealIP()

				return c.NoContent(http.StatusNoContent)
			})

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/probe", nil)
			req.RemoteAddr = tc.peer
			req.Header.Set("X-Real-IP", "203.0.113.7")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			require.Equal(t, http.StatusNoContent, rec.Code)
			assert.Equal(t, tc.expected, realIP)
			assert.Equal(t, tc.expected, fromContext.IP)
		})
	}
}
