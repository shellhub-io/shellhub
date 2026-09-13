package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/server/api/pkg/echo/handlers"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer/dialertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCloseRequest(t *testing.T, role string) (*echo.Context, *httptest.ResponseRecorder) {
	t.Helper()

	e := echo.New()
	e.Binder = handlers.NewBinder()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/sessions/session-uid/close", strings.NewReader(`{"device":"device-uid"}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set("X-Role", role)
	req.Header.Set("X-Tenant-ID", "tenant-id")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPathValues(echo.PathValues{{Name: "uid", Value: "session-uid"}})

	return c, rec
}

// TestHandleSSHCloseAsksTheTunnelToCloseTheSession pins what the handler states at the call
// site: the device the session is on, and that the connection is for closing that session.
func TestHandleSSHCloseAsksTheTunnelToCloseTheSession(t *testing.T) {
	c, rec := newCloseRequest(t, "administrator")

	agent := dialertest.NewAgent(t)
	h := &Handlers{Dialer: agent} //nolint:exhaustruct // the close path reaches neither the service nor the tunnel registry

	require.NoError(t, h.HandleSSHClose(c))
	assert.Equal(t, http.StatusOK, rec.Code)

	assert.Equal(t, []dialertest.Dial{{
		Tenant: "tenant-id",
		UID:    "device-uid",
		Target: dialer.SSHCloseTarget{SessionID: "session-uid"},
	}}, agent.Dials())
}

// TestHandleSSHCloseReportsAnUnreachableDevice covers the dial failing however it fails: the
// agent asked for the close is told the device could not be reached.
func TestHandleSSHCloseReportsAnUnreachableDevice(t *testing.T) {
	for _, failure := range []error{dialer.ErrNoConnection, dialer.ErrUnreachable, dialer.ErrInvalidArgument} {
		t.Run(failure.Error(), func(t *testing.T) {
			c, _ := newCloseRequest(t, "administrator")

			stub := &dialertest.Stub{Err: failure} //nolint:exhaustruct // the recording field starts empty and is appended to under the mutex
			h := &Handlers{Dialer: stub}           //nolint:exhaustruct // the close path reaches neither the service nor the tunnel registry

			require.ErrorIs(t, h.HandleSSHClose(c), ErrDeviceTunnelDial)

			assert.Equal(t, []dialertest.Dial{{
				Tenant: "tenant-id",
				UID:    "device-uid",
				Target: dialer.SSHCloseTarget{SessionID: "session-uid"},
			}}, stub.Dials(), "the failure path must hand the tunnel the same close target")
		})
	}
}

func TestHandleSSHCloseAuthorization(t *testing.T) {
	forbiddenRoles := []string{"observer", "operator", "", "invalid"}

	for _, role := range forbiddenRoles {
		t.Run("rejects role "+role, func(t *testing.T) {
			c, rec := newCloseRequest(t, role)

			h := &Handlers{} //nolint:exhaustruct // Dialer must not be reached for forbidden roles.

			require.NoError(t, h.HandleSSHClose(c))
			assert.Equal(t, http.StatusForbidden, rec.Code)
		})
	}
}
