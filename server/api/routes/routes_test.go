package routes

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRouter(tb testing.TB, service services.Service, opts ...Option) *echo.Echo {
	tb.Helper()

	router, err := NewRouter(service, opts...)
	require.NoError(tb, err)

	return router
}

func TestNewRouterReportsAFailingOption(t *testing.T) {
	errOption := errors.New("option failed")

	succeeding := func(*echo.Echo, *Handler) error { return nil }
	failing := func(*echo.Echo, *Handler) error { return errOption }

	t.Run("reports the error of a failing option", func(t *testing.T) {
		router, err := NewRouter(mocks.NewMockService(t), succeeding, failing)
		require.ErrorIs(t, err, errOption)
		assert.Nil(t, router)
	})

	t.Run("builds the router when every option succeeds", func(t *testing.T) {
		router, err := NewRouter(mocks.NewMockService(t), succeeding)
		require.NoError(t, err)
		assert.NotNil(t, router)
	})
}

// TestRequestIDReachesTheHandler pins that a request carries an ID even when
// nothing in front of the API supplied one. The agent's V2 handshake binds
// X-Request-ID as a required header, so a missing value is a failed connection
// rather than a missing log field.
func TestRequestIDReachesTheHandler(t *testing.T) {
	tests := []struct {
		description string
		sent        string
	}{
		{
			description: "generated when the client sends none",
		},
		{
			description: "preserved when it arrives from a proxy",
			sent:        "from-the-edge",
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			router, authn, _ := authenticatedRouter(t)

			const probe = "/api/request-id-probe"

			var seen string

			router.GET(probe, func(c *echo.Context) error {
				seen = c.Request().Header.Get(echo.HeaderXRequestID)

				return c.NoContent(http.StatusOK)
			})
			authn.AllowAnonymous(http.MethodGet, probe)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, probe, nil)
			if tc.sent != "" {
				req.Header.Set(echo.HeaderXRequestID, tc.sent)
			}

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
			assert.NotEmpty(t, seen)
			assert.Equal(t, seen, rec.Header().Get(echo.HeaderXRequestID))

			if tc.sent != "" {
				assert.Equal(t, tc.sent, seen)
			}
		})
	}
}
