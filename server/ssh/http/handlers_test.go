package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/server/api/pkg/echo/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleSSHCloseAuthorization(t *testing.T) {
	forbiddenRoles := []string{"observer", "operator", "", "invalid"}

	for _, role := range forbiddenRoles {
		t.Run("rejects role "+role, func(t *testing.T) {
			e := echo.New()
			e.Binder = handlers.NewBinder()

			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/sessions/session-uid/close", strings.NewReader(`{"device":"device-uid"}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.Header.Set("X-Role", role)
			req.Header.Set("X-Tenant-ID", "tenant-id")

			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPathValues(echo.PathValues{{Name: "uid", Value: "session-uid"}})

			h := &Handlers{} //nolint:exhaustruct // Dialer must not be reached for forbidden roles.

			require.NoError(t, h.HandleSSHClose(c))
			assert.Equal(t, http.StatusForbidden, rec.Code)
		})
	}
}
