package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/echo/handlers"
	"github.com/shellhub-io/shellhub/server/api/services"
	servicemocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer/dialertest"
	"github.com/shellhub-io/shellhub/server/ssh/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

	registry := session.NewRegistry()
	registry.Add("session-uid")

	agent := dialertest.NewAgent(t)
	h := &Handlers{Dialer: agent, Sessions: registry} //nolint:exhaustruct // an owned session never reaches the service

	require.NoError(t, h.HandleSSHClose(c))
	assert.Equal(t, http.StatusAccepted, rec.Code)

	assert.Equal(t, []dialertest.Dial{{
		Tenant: "tenant-id",
		UID:    "device-uid",
		Target: dialer.SSHCloseTarget{SessionID: "session-uid"},
	}}, agent.Dials())
}

func TestHandleSSHCloseReportsAnUnreachableDevice(t *testing.T) {
	for _, failure := range []error{dialer.ErrNoConnection, dialer.ErrUnreachable, dialer.ErrInvalidArgument} {
		t.Run(failure.Error(), func(t *testing.T) {
			c, _ := newCloseRequest(t, "administrator")

			registry := session.NewRegistry()
			registry.Add("session-uid")

			stub := &dialertest.Stub{Err: failure}           //nolint:exhaustruct // the recording field starts empty and is appended to under the mutex
			h := &Handlers{Dialer: stub, Sessions: registry} //nolint:exhaustruct // an owned session never reaches the service

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

func TestHandleSSHCloseRetiresASessionNobodyOwns(t *testing.T) {
	c, rec := newCloseRequest(t, "administrator")

	service := servicemocks.NewMockService(t)
	service.On("GetSession", mock.Anything, scope.MustBounded("tenant-id"), models.UID("session-uid")).
		Return(&models.Session{UID: "session-uid", TenantID: "tenant-id"}, nil).Once()
	service.On("DeactivateSession", mock.Anything, models.UID("session-uid")).Return(nil).Once()

	h := &Handlers{ //nolint:exhaustruct // the close path reaches neither the tunnel registry nor the config
		Dialer:   dialertest.NewAgent(t),
		Service:  service,
		Sessions: session.NewRegistry(),
	}

	require.NoError(t, h.HandleSSHClose(c))
	assert.Equal(t, http.StatusOK, rec.Code, "200 means the session is closed, and here the handler closed it")

	service.AssertExpectations(t)
}

func TestHandleSSHCloseLeavesAnotherNamespacesSessionAlone(t *testing.T) {
	c, rec := newCloseRequest(t, "administrator")

	service := servicemocks.NewMockService(t)
	service.On("GetSession", mock.Anything, scope.MustBounded("tenant-id"), models.UID("session-uid")).
		Return(nil, services.NewErrSessionNotFound("session-uid", store.ErrNoDocuments)).Once()

	h := &Handlers{ //nolint:exhaustruct // the close path reaches neither the tunnel registry nor the config
		Dialer:   dialertest.NewAgent(t),
		Service:  service,
		Sessions: session.NewRegistry(),
	}

	require.NoError(t, h.HandleSSHClose(c))
	assert.Equal(t, http.StatusNotFound, rec.Code, "a session outside the caller's namespace is not the caller's to retire")

	service.AssertNotCalled(t, "DeactivateSession", mock.Anything, mock.Anything)
}

func TestHandleSSHCloseDelegatesASessionItOwns(t *testing.T) {
	c, rec := newCloseRequest(t, "administrator")

	registry := session.NewRegistry()
	registry.Add("session-uid")

	h := &Handlers{ //nolint:exhaustruct // the close path reaches neither the tunnel registry nor the config
		Dialer:   dialertest.NewAgent(t),
		Service:  servicemocks.NewMockService(t),
		Sessions: registry,
	}

	require.NoError(t, h.HandleSSHClose(c))
	assert.Equal(t, http.StatusAccepted, rec.Code, "202 means the request was delivered, not that the session closed")
}

func TestHandleSSHCloseKeepsADialFailureAnErrorForAnOwnedSession(t *testing.T) {
	c, _ := newCloseRequest(t, "administrator")

	registry := session.NewRegistry()
	registry.Add("session-uid")

	stub := &dialertest.Stub{Err: dialer.ErrUnreachable} //nolint:exhaustruct // the recording field starts empty and is appended to under the mutex

	h := &Handlers{ //nolint:exhaustruct // the close path reaches neither the tunnel registry nor the config
		Dialer:   stub,
		Service:  servicemocks.NewMockService(t),
		Sessions: registry,
	}

	assert.ErrorIs(t, h.HandleSSHClose(c), ErrDeviceTunnelDial)
}
