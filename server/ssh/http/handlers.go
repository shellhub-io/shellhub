package http

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/wsconnadapter"
	"github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/api/store"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
	log "github.com/sirupsen/logrus"
)

// Handlers serves the HTTP endpoints agents use to open and hold their reverse tunnels. It
// is a sidecar to the SSH server proper, sharing its dialer.
type Handlers struct {
	Config  *Config
	Dialer  *dialer.Dialer
	Service services.Service
}

const (
	deviceResolveAttempts = 3
	deviceResolveBackoff  = time.Second
)

func (h *Handlers) resolveDevice(ctx context.Context, uid string) (*models.Device, error) {
	var err error

	for attempt := 1; attempt <= deviceResolveAttempts; attempt++ {
		var device *models.Device

		device, err = h.Service.GetDevice(ctx, scope.NewUnbounded(reasonAgentDeviceResolve), models.UID(uid))
		if err == nil {
			return device, nil
		}

		if errors.Is(err, store.ErrNoDocuments) {
			break
		}

		if attempt == deviceResolveAttempts {
			break
		}

		select {
		case <-ctx.Done():
			return nil, echo.ErrInternalServerError
		case <-time.After(time.Duration(attempt) * deviceResolveBackoff):
		}
	}

	log.WithError(err).WithField("uid", uid).Error("unable to retrieve device for connection")

	return nil, echo.ErrInternalServerError
}

const (
	// HandleSSHClosePath receives a request to close an existing SSH session.
	HandleSSHClosePath = "/api/sessions/:uid/close"
)

const (
	// HandleConnectionV1Path is the connection endpoint where agents using revdial connects to establish
	// a WebSocket connection. Each new logical session requires an extra reverse dial handshake.
	HandleConnectionV1Path = "/ssh/connection"
	// HandleConnectionV2Path is the connection endpoint where agents using yamux/multistream connects to
	// establish a WebSocket connection. Subsequent logical streams are opened without additional HTTP
	// handshakes and are protocol-negotiated via multistream-select.
	HandleConnectionV2Path = "/agent/connection"
)

const (
	// HandleRevdialPath is the reverse dial endpoint where agents using revdial requests a new logical
	// session.
	HandleRevdialPath = "/ssh/revdial"
)

// HandleSSHClose receives a notification from the agent that an SSH
// session should be closed. It dials the device (choosing the correct
// transport version) and then performs the version-specific close
// sequence: HTTP GET for V1 or multistream + JSON payload for V2.
func (h *Handlers) HandleSSHClose(c *echo.Context) error {
	var data struct {
		UID    string `param:"uid"`
		Device string `json:"device"`
	}

	if err := c.Bind(&data); err != nil {
		return err
	}

	if role := authorizer.RoleFromString(c.Request().Header.Get("X-Role")); !role.HasPermission(authorizer.SessionClose) {
		return c.NoContent(http.StatusForbidden)
	}

	ctx := c.Request().Context()

	tenant := c.Request().Header.Get("X-Tenant-ID")

	if _, err := h.Dialer.DialTo(ctx, tenant, data.Device, dialer.SSHCloseTarget{SessionID: data.UID}); err != nil {
		log.WithError(err).Error("failed to send ssh close message")

		return ErrDeviceTunnelDial
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handlers) requireAcceptedDevice(ctx context.Context, uid string) (*models.Device, error) {
	device, err := h.resolveDevice(ctx, uid)
	if err != nil {
		return nil, err
	}

	if device.Status != models.DeviceStatusAccepted {
		log.WithFields(log.Fields{"uid": uid, "status": device.Status}).
			Debug("refusing reverse tunnel for a non-accepted device")

		return nil, echo.NewHTTPError(http.StatusForbidden, "device is not accepted")
	}

	return device, nil
}

// HandleConnectionV1 upgrades the HTTP connection to WebSocket and
// registers a legacy (V1) reverse dialer for the agent. Each new logical
// session requires an extra reverse dial handshake.
func (h *Handlers) HandleConnectionV1(c *echo.Context) error {
	requestID := c.Request().Header.Get("X-Request-ID")

	tenant := c.Request().Header.Get("X-Tenant-ID")
	uid := c.Request().Header.Get("X-Device-UID")

	if h.Config.RequireAcceptedTunnel {
		device, err := h.requireAcceptedDevice(c.Request().Context(), uid)
		if err != nil {
			return err
		}

		if tenant == "" {
			tenant = device.TenantID
		}
	} else if tenant == "" {
		device, err := h.resolveDevice(c.Request().Context(), uid)
		if err != nil {
			return err
		}

		tenant = device.TenantID
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	h.Dialer.Manager.Set(
		dialer.NewKey(tenant, uid),
		wsconnadapter.New(
			conn,
			wsconnadapter.WithID(requestID),
			wsconnadapter.WithDevice(tenant, uid),
		),
		HandleRevdialPath,
	)

	return nil
}

// HandleConnectionV2Data is the identifying header set an agent sends when opening a V2
// tunnel. The device UID and tenant are validated here rather than trusted, because the
// connection is not yet authenticated at that point.
type HandleConnectionV2Data struct {
	RequestID string `header:"x-request-id" validate:"required"`
	UID       string `header:"x-device-uid" validate:"required,len=64"`
	Tenant    string `header:"x-tenant-id" validate:"required,uuid"`
}

// HandleConnectionV2 upgrades the HTTP connection to WebSocket and
// binds it to a yamux session (V2). Subsequent logical streams are
// opened without additional HTTP handshakes and are protocol-negotiated
// via multistream-select.
func (h *Handlers) HandleConnectionV2(c *echo.Context) error {
	log.Trace("handling v2 connection")
	defer log.Trace("v2 connection handle closed")

	var data HandleConnectionV2Data

	if err := c.Bind(&data); err != nil {
		log.WithError(err).Error("failed to bind the request")

		return err
	}

	if err := c.Validate(&data); err != nil {
		log.WithError(err).Error("failed to validate the request")

		return err
	}

	if h.Config.RequireAcceptedTunnel {
		if _, err := h.requireAcceptedDevice(c.Request().Context(), data.UID); err != nil {
			return err
		}
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	logger := log.WithFields(log.Fields{
		"request-id": data.RequestID,
		"tenant":     data.Tenant,
		"uid":        data.UID,
	})

	logger.Info("v2 connection established")

	if err := h.Dialer.Manager.Bind(
		data.Tenant,
		data.UID,
		wsconnadapter.New(
			conn,
			wsconnadapter.WithID(data.RequestID),
			wsconnadapter.WithDevice(data.Tenant, data.UID),
		),
	); err != nil {
		logger.WithError(err).Error("failed to bind the connection")

		return err
	}

	logger.Info("v2 connection bound")

	return nil
}
