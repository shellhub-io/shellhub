package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/echo/handlers"
	"github.com/shellhub-io/shellhub/server/api/services"
	servicemocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// The status codes on these two routes are a contract with agents already in the
// field, which do not upgrade alongside the server. They are asserted against the
// wired error handler rather than the handler's return value, because the status
// is decided by the handler chain and not by the route.
//
// Every case below fails before the WebSocket upgrade, so no upgrade is needed to
// exercise it.
func newConnectionTestServer(t *testing.T, service services.Service, cfg *Config) *echo.Echo {
	t.Helper()

	e := echo.New()
	e.Binder = handlers.NewBinder()
	e.Validator = handlers.NewValidator()
	e.HTTPErrorHandler = handlers.NewErrors(nil)

	h := &Handlers{Config: cfg, Dialer: nil, Service: service}

	e.GET(HandleConnectionV1Path, h.HandleConnectionV1)
	e.GET(HandleConnectionV2Path, h.HandleConnectionV2)

	return e
}

const testDeviceUID = "0000000000000000000000000000000000000000000000000000000000000000"

func TestHandleConnectionV1StatusCodes(t *testing.T) {
	tests := []struct {
		description    string
		config         *Config
		tenantHeader   string
		setupMock      func(*servicemocks.MockService)
		expectedStatus int
	}{
		{
			description:  "an unknown device is refused with a server error",
			config:       &Config{RequireAcceptedTunnel: true}, //nolint:exhaustruct
			tenantHeader: "00000000-0000-4000-0000-000000000000",
			setupMock: func(m *servicemocks.MockService) {
				m.EXPECT().
					GetDevice(mock.Anything, mock.Anything, models.UID(testDeviceUID)).
					Return(nil, services.NewErrDeviceNotFound(models.UID(testDeviceUID), store.ErrNoDocuments)).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			description:  "a device that is not accepted is refused",
			config:       &Config{RequireAcceptedTunnel: true}, //nolint:exhaustruct
			tenantHeader: "00000000-0000-4000-0000-000000000000",
			setupMock: func(m *servicemocks.MockService) {
				m.EXPECT().
					GetDevice(mock.Anything, mock.Anything, models.UID(testDeviceUID)).
					Return(&models.Device{UID: testDeviceUID, Status: models.DeviceStatusPending}, nil). //nolint:exhaustruct
					Once()
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			description:  "the pre-0.15 fallback refuses an unknown device with a server error",
			config:       &Config{RequireAcceptedTunnel: false}, //nolint:exhaustruct
			tenantHeader: "",
			setupMock: func(m *servicemocks.MockService) {
				m.EXPECT().
					GetDevice(mock.Anything, mock.Anything, models.UID(testDeviceUID)).
					Return(nil, services.NewErrDeviceNotFound(models.UID(testDeviceUID), store.ErrNoDocuments)).
					Once()
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			serviceMock := servicemocks.NewMockService(t)
			tt.setupMock(serviceMock)

			e := newConnectionTestServer(t, serviceMock, tt.config)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, HandleConnectionV1Path, nil)
			req.Header.Set("X-Device-UID", testDeviceUID)
			req.Header.Set("X-Request-ID", "request-id")

			if tt.tenantHeader != "" {
				req.Header.Set("X-Tenant-ID", tt.tenantHeader)
			}

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

// A device that does not exist is asked for once; anything else is retried,
// because the loopback client used to absorb a database blip inside the agent's
// request and the fleet reconnects in lockstep without it.
func TestHandleConnectionV1RetriesOnlyInfrastructureFailures(t *testing.T) {
	tests := []struct {
		description      string
		err              error
		expectedAttempts int
	}{
		{
			description:      "a missing device is not retried",
			err:              services.NewErrDeviceNotFound(models.UID(testDeviceUID), store.ErrNoDocuments),
			expectedAttempts: 1,
		},
		{
			description:      "an infrastructure failure is retried",
			err:              services.NewErrDeviceNotFound(models.UID(testDeviceUID), errors.New("connection refused")),
			expectedAttempts: deviceResolveAttempts,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			serviceMock := servicemocks.NewMockService(t)
			serviceMock.EXPECT().
				GetDevice(mock.Anything, mock.Anything, models.UID(testDeviceUID)).
				Return(nil, tt.err).
				Times(tt.expectedAttempts)

			e := newConnectionTestServer(t, serviceMock, &Config{RequireAcceptedTunnel: true}) //nolint:exhaustruct

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, HandleConnectionV1Path, nil)
			req.Header.Set("X-Device-UID", testDeviceUID)
			req.Header.Set("X-Request-ID", "request-id")
			req.Header.Set("X-Tenant-ID", "00000000-0000-4000-0000-000000000000")

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusInternalServerError, rec.Code)
			serviceMock.AssertExpectations(t)
		})
	}
}

func TestHandleConnectionV2StatusCodes(t *testing.T) {
	tests := []struct {
		description    string
		headers        map[string]string
		expectedStatus int
	}{
		{
			description: "a malformed tenant is rejected before anything else",
			headers: map[string]string{
				"X-Request-ID": "request-id",
				"X-Device-UID": testDeviceUID,
				"X-Tenant-ID":  "not-a-uuid",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			description: "a short device uid is rejected",
			headers: map[string]string{
				"X-Request-ID": "request-id",
				"X-Device-UID": "too-short",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			description: "a missing request id is rejected",
			headers: map[string]string{
				"X-Device-UID": testDeviceUID,
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			e := newConnectionTestServer(t, nil, &Config{}) //nolint:exhaustruct

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, HandleConnectionV2Path, nil)
			for name, value := range tt.headers {
				req.Header.Set(name, value)
			}

			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			require.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}
