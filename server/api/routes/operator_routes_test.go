package routes

import (
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
)

func assertNoOperatorRoute(t *testing.T, router *echo.Echo) {
	t.Helper()

	for _, route := range router.Router().Routes() {
		for _, prefix := range []string{"/internal", "/metrics"} {
			assert.False(t, route.Path == prefix || strings.HasPrefix(route.Path, prefix+"/"),
				"%s %s is mounted on the router every client reaches", route.Method, route.Path)
		}
	}
}

func TestRouterMountsNoOperatorRoute(t *testing.T) {
	for _, edition := range editions {
		t.Run(string(edition), func(t *testing.T) {
			assertNoOperatorRoute(t, newTestRouter(t, mocks.NewMockService(t), WithEdition(edition)))
		})
	}

	t.Run("with request metrics recorded", func(t *testing.T) {
		assertNoOperatorRoute(t, newTestRouter(t, mocks.NewMockService(t), WithMetrics()))
	})
}
