package http

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/wsconnadapter"
	"github.com/shellhub-io/shellhub/ssh/pkg/dialer"
	log "github.com/sirupsen/logrus"
)

type Handlers struct {
	Config *Config
	Dialer *dialer.Dialer
	Client internalclient.Client
}

const (
	// HandleSSHClosePath receives a request to close an existing SSH session.
	HandleSSHClosePath = "/api/sessions/:uid/close"
	// HandleHTTPProxyPath proxies an inbound HTTP request to a device's HTTP server.
	HandleHTTPProxyPath = "/http/proxy"
	// HandleHealthcheckPath is used for readiness/liveness checks.
	HandleHealthcheckPath = "/healthcheck"
)

const (
	// HandleConnectionV1Path is the connection endpoint where agents using revdial connects to establish
	// a WebSocket connection. Each new logical session requires an extra reverse dial handshake.
	HandleConnectionV1Path = "/ssh/connection"
	// HandleConnectionV2Path is the connection endpoint where agents using yamux/multistream connects to
	// establish a WebSocket connection. Subsequent logical streams are opened without additional HTTP
	// handshakes and are protocol-negotiated via multistream-select.
	HandleConnectionV2Path = "/connection"
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
func (h *Handlers) HandleSSHClose(c echo.Context) error {
	var data struct {
		UID    string `param:"uid"`
		Device string `json:"device"`
	}

	if err := c.Bind(&data); err != nil {
		return err
	}

	ctx := c.Request().Context()

	tenant := c.Request().Header.Get("X-Tenant-ID")

	if _, err := h.Dialer.DialTo(ctx, tenant, data.Device, dialer.SSHCloseTarget{SessionID: data.UID}); err != nil {
		log.WithError(err).Error("failed to send ssh close message")

		return ErrDeviceTunnelDial
	}

	return c.NoContent(http.StatusOK)
}

// HandleHTTPProxy proxies an inbound HTTP request to a device's HTTP
// service exposed through the reverse tunnel/web endpoint feature. It
// supports both transport versions:
//   - V1: issues a CONNECT prelude then performs a standard HTTP request over the established raw tunnel.
//   - V2: negotiates the /http/proxy multistream protocol and exchanges a JSON envelope to set up the target host/port.
//
// The handler then hijacks the Echo response writer to stream data
// bidirectionally between client and device.
func (h *Handlers) HandleHTTPProxy(c echo.Context) error {
	requestID := c.Request().Header.Get("X-Request-ID")

	address := c.Request().Header.Get("X-Address")
	log.WithFields(log.Fields{
		"request-id": requestID,
		"address":    address,
	}).Debug("address value")

	path := c.Request().Header.Get("X-Path")
	log.WithFields(log.Fields{
		"request-id": requestID,
		"address":    address,
	}).Debug("path")

	endpoint, err := h.Client.LookupWebEndpoints(c.Request().Context(), address)
	if err != nil {
		log.WithError(err).Error("failed to get the web endpoint")

		return c.JSON(http.StatusForbidden, NewMessageFromError(ErrWebEndpointForbidden))
	}

	logger := log.WithFields(log.Fields{
		"request-id": requestID,
		"namespace":  endpoint.Namespace,
		"device":     endpoint.DeviceUID,
	})

	// Prepare V1 CONNECT handshake request (only used if version=V1 inside target implementation)
	handshakeReq, _ := http.NewRequest(http.MethodConnect, fmt.Sprintf("/http/proxy/%s:%d", endpoint.Host, endpoint.Port), nil)
	conn, err := h.Dialer.DialTo(c.Request().Context(), endpoint.Namespace, endpoint.DeviceUID, dialer.HTTPProxyTarget{
		RequestID:        requestID,
		Host:             endpoint.Host,
		Port:             endpoint.Port,
		HandshakeRequest: handshakeReq,
	})
	if err != nil {
		logger.WithError(err).Error("failed to dial to device")

		return c.JSON(http.StatusForbidden, NewMessageFromError(ErrDeviceTunnelDial))
	}
	defer conn.Close()

	logger.Trace("new web endpoint connection initialized")
	defer logger.Trace("web endpoint connection doned")

	req := c.Request()
	req.Host = strings.Join([]string{address, h.Config.WebEndpointsDomain}, ".")
	req.URL, err = url.Parse(path)
	if err != nil {
		logger.WithError(err).Error("failed to parse the path")

		return c.JSON(http.StatusInternalServerError, NewMessageFromError(ErrDeviceTunnelReadResponse))
	}

	if err := req.Write(conn); err != nil {
		logger.WithError(err).Error("failed to write the request to the agent")

		return c.JSON(http.StatusInternalServerError, NewMessageFromError(ErrDeviceTunnelWriteRequest))
	}

	log.WithFields(log.Fields{
		"request-id": requestID,
		"method":     req.Method,
		"url":        req.URL.String(),
		"host":       req.Host,
		"headers":    req.Header,
	}).Debug("request to device")

	ctr := http.NewResponseController(c.Response())
	out, _, err := ctr.Hijack()
	if err != nil {
		logger.WithError(err).Error("failed to hijack the http request")

		return c.JSON(http.StatusInternalServerError, NewMessageFromError(ErrDeviceTunnelHijackRequest))
	}

	defer out.Close()

	// Bidirectional copy between the client and the device.
	var wg sync.WaitGroup
	wg.Add(2)

	done := sync.OnceFunc(func() {
		// underlying connection closed by done() after copy
		defer out.Close()

		logger.Trace("close called on in and out connections")
	})

	go func() {
		defer done()
		defer wg.Done()

		if _, err := io.Copy(conn, out); err != nil {
			logger.WithError(err).Debug("in and out done returned a error")
		}

		logger.Trace("in and out done")
	}()

	go func() {
		defer done()
		defer wg.Done()

		if _, err := io.Copy(out, conn); err != nil {
			logger.WithError(err).Debug("out and in done returned a error")
		}

		logger.Trace("out and in done")
	}()

	wg.Wait()

	logger.Debug("http proxy is done")

	return nil
}

// HandleHealthcheck returns a simple 200 OK used for readiness/liveness
// checks.
func (h *Handlers) HandleHealthcheck(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

// HandleConnectionV1 upgrades the HTTP connection to WebSocket and
// registers a legacy (V1) reverse dialer for the agent. Each new logical
// session requires an extra reverse dial handshake.
func (h *Handlers) HandleConnectionV1(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	requestID := c.Request().Header.Get("X-Request-ID")

	tenant := c.Request().Header.Get("X-Tenant-ID")
	uid := c.Request().Header.Get("X-Device-UID")

	// WARN: In versions before 0.15, the agent's authentication may not provide the "X-Tenant-ID" header.
	// This can cause issues with establishing sessions and tracking online devices. To solve this,
	// we retrieve the tenant ID by querying the API. Maybe this can be removed in a future release.
	if tenant == "" {
		device, err := h.Client.GetDevice(c.Request().Context(), uid)
		if err != nil {
			log.WithError(err).
				WithField("uid", uid).
				Error("unable to retrieve device's tenant id")

			return err
		}

		tenant = device.TenantID
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

type HandleConnectionV2Data struct {
	RequestID string `header:"x-request-id" validate:"required"`
	UID       string `header:"x-device-uid" validate:"required,len=64"`
	Tenant    string `header:"x-tenant-id" validate:"required,uuid"`
}

// HandleConnectionV2 upgrades the HTTP connection to WebSocket and
// binds it to a yamux session (V2). Subsequent logical streams are
// opened without additional HTTP handshakes and are protocol-negotiated
// via multistream-select.
func (h *Handlers) HandleConnectionV2(c echo.Context) error {
	log.Trace("handling v2 connection")
	defer log.Trace("v2 connection handle closed")

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	var data HandleConnectionV2Data

	if err := c.Bind(&data); err != nil {
		log.WithError(err).Error("failed to bind the request")

		return err
	}

	if err := c.Validate(&data); err != nil {
		log.WithError(err).Error("failed to validate the request")

		return err
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
