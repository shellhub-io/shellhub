package http

import (
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/revdial"
	"github.com/shellhub-io/shellhub/pkg/webendpoints"
	"github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
)

type Message struct {
	Message string `json:"message"`
}

func NewMessageFromError(err error) Message {
	return Message{
		Message: err.Error(),
	}
}

// Config controls optional features for the SSH HTTP sidecar server.
//
// When WebEndpoints is enabled the server exposes an HTTP proxy entry
// point (/http/proxy) that allows externally accessible per-device
// subdomains to be resolved and forwarded through the reverse tunnel
// transport (supporting both legacy V1 and yamux/multistream V2).
type Config struct {
	// WebEndpoints enables the web endpoints (HTTP proxy) feature.
	WebEndpoints bool
	// WebEndpointsDomain is the base domain used when constructing the
	// host header for tunneled HTTP requests (e.g. <address>.<domain>).
	// When non-empty it takes precedence over Domain.
	WebEndpointsDomain string
	// Domain is the fallback base domain when WebEndpointsDomain is not
	// set.
	Domain string
	// RequireAcceptedTunnel refuses the reverse tunnel for a device that is not accepted (pending or
	// rejected). Off by default; opt-in per instance.
	RequireAcceptedTunnel bool
}

// webEndpointHost builds the full host value for a tunneled HTTP
// request by joining address and the effective domain with a dot.
// When the effective domain is empty the address is returned as-is
// (no trailing dot), acting as a regression guard against malformed
// Host headers.
func (c *Config) webEndpointHost(address string) string {
	return webendpoints.Host(address, webendpoints.Domain(c.WebEndpointsDomain, c.Domain))
}

var (
	ErrWebEndpointForbidden      = errors.New("web endpoint not found")
	ErrDeviceTunnelDial          = errors.New("failed to connect to device")
	ErrDeviceTunnelWriteRequest  = errors.New("failed to send data to the device")
	ErrDeviceTunnelReadResponse  = errors.New("failed to write the response back to the client")
	ErrDeviceTunnelHijackRequest = errors.New("failed to capture the request")
	ErrDeviceTunnelParsePath     = errors.New("failed to parse the path")
	ErrDeviceTunnelConnect       = errors.New("failed to connect to the port on device")
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	Subprotocols:    []string{"binary"},
	CheckOrigin: func(_ *http.Request) bool {
		return true
	},
}

// Register adds the SSH routes to the API's router. The binder, validator and
// error handler are the API's; they are a superset of what these handlers need.
func Register(router *echo.Echo, d *dialer.Dialer, service services.Service, cli internalclient.Client, cfg *Config) *Handlers {
	handlers := &Handlers{
		Dialer:  d,
		Service: service,
		Client:  cli,
		Config:  cfg,
	}

	router.GET(HandleConnectionV1Path, handlers.HandleConnectionV1)
	router.GET(HandleConnectionV2Path, handlers.HandleConnectionV2)

	router.GET(HandleRevdialPath, echo.WrapHandler(revdial.ConnHandler(upgrader)))

	// Registered at the root rather than under the API group: the group carries
	// license enforcement in Enterprise, which this route does not have today.
	router.POST(HandleSSHClosePath, handlers.HandleSSHClose)

	if cfg.WebEndpoints {
		// NOTE: The `/http/proxy` endpoint is invoked by the NGINX gateway when a tunnel URL is accessed. It processes
		// the `X-Address` and `X-Path` headers, which specify the tunnel's address and the target path on the server,
		// returning an error related to the connection to device or what was returned from the server inside the tunnel.
		router.Any(HandleHTTPProxyPath, handlers.HandleHTTPProxy)
	}

	return handlers
}
