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
	"github.com/shellhub-io/shellhub/server/ssh/session"
	log "github.com/sirupsen/logrus"
)

// Handlers serves the HTTP endpoints agents use to open and hold their reverse tunnels. It
// is a sidecar to the SSH server proper, sharing its dialer.
//
// Dialer reaches a device that already holds a tunnel; Tunnels is the registry the agent
// endpoints put one into. They are separate fields so that the dial can be substituted in a
// test without standing up a registry. What stops an out-of-tree caller registering a tunnel
// is [TunnelExtension] taking the interface, not this split.
type Handlers struct {
	Config  *Config
	Dialer  dialer.TunnelDialer
	Tunnels *dialer.Manager
	Service services.Service
	// Sessions is the set of sessions this process owns. The close endpoint reads it to tell a
	// session whose teardown will run from one that has been orphaned by a restart.
	Sessions *session.Registry
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

// HandleSSHClose asks for an SSH session to be closed. It dials the device to end the session
// there, choosing the correct transport version, and then answers according to whether this
// process still owns the session.
//
// It returns 200 when the session is closed, which is the case for a session no gateway holds:
// nothing else would ever deactivate it, so this call does. It returns 202 when a gateway still
// holds the session, because that gateway's teardown is what closes it, and ErrDeviceTunnelDial
// when the device backing such a session cannot be reached. It returns 404 for a session no
// gateway holds that is not in the caller's namespace, which it leaves untouched.
//
// The device's answer never decides this. A device ignores a close for a session it does not
// have, so the dial succeeding says nothing about whether the session existed.
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

	owned := h.Sessions.Has(data.UID)

	if _, err := h.Dialer.DialTo(ctx, tenant, data.Device, dialer.SSHCloseTarget{SessionID: data.UID}); err != nil {
		logger := log.WithError(err).
			WithFields(log.Fields{"session": data.UID, "device": data.Device})

		if errors.Is(err, dialer.ErrNoConnection) {
			logger.Warning("failed to send the ssh close message: " + dialer.Describe(err))
		} else {
			logger.Error("failed to send the ssh close message: " + dialer.Describe(err))
		}

		if owned {
			return ErrDeviceTunnelDial
		}
	}

	if owned {
		return c.NoContent(http.StatusAccepted)
	}

	sc, err := scope.NewBounded(tenant)
	if err != nil {
		return c.NoContent(http.StatusForbidden)
	}

	if _, err := h.Service.GetSession(ctx, sc, models.UID(data.UID)); err != nil {
		return c.NoContent(http.StatusNotFound)
	}

	if err := h.Service.DeactivateSession(ctx, models.UID(data.UID)); err != nil {
		log.WithError(err).
			WithField("session", data.UID).
			Error("failed to deactivate a session no gateway owns")

		return err
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

	h.Tunnels.Set(
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

	if err := h.Tunnels.Bind(
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
