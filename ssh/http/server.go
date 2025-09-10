package http

import (
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	"github.com/shellhub-io/shellhub/pkg/revdial"
	"github.com/shellhub-io/shellhub/pkg/validator"
	"github.com/shellhub-io/shellhub/ssh/metrics"
	"github.com/shellhub-io/shellhub/ssh/pkg/dialer"
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
	// Metrics enables the Prometheus metrics endpoint at /metrics.
	Metrics bool
	// WebEndpoints enables the web endpoints (HTTP proxy) feature.
	WebEndpoints bool
	// WebEndpointsDomain is the base domain used when constructing the
	// host header for tunneled HTTP requests (e.g. <address>.<domain>).
	WebEndpointsDomain string
}

// Server wires HTTP routes (connection upgrade, reverse dialing,
// web endpoint proxy, healthcheck) to the underlying dialer and
// handlers. It exposes both V1 (/ssh/connection + /ssh/revdial) and V2
// (/connection) endpoints during the transition period while agents
// upgrade.
type Server struct {
	Config   *Config
	Router   *echo.Echo
	Handlers *Handlers
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

// ListenAndServe starts the Echo HTTP server on the provided address.
func (s *Server) ListenAndServe(address string) error {
	return s.Router.Start(address)
}

type Binder struct{}

func NewBinder() *Binder {
	return &Binder{}
}

func (b *Binder) Bind(s any, c echo.Context) error {
	binder := new(echo.DefaultBinder)
	if err := binder.Bind(s, c); err != nil {
		err := err.(*echo.HTTPError) //nolint:forcetypeassert

		return err
	}

	if err := binder.BindHeaders(c, s); err != nil {
		err := err.(*echo.HTTPError) //nolint:forcetypeassert

		return err
	}

	return nil
}

type Validator struct {
	validator *validator.Validator
}

// NewValidator creates a new validator for the echo framework from the ShellHub validator.
func NewValidator() *Validator {
	return &Validator{validator: validator.New()}
}

// Validate is called by the echo framework to validate the request body.
// If the request body is invalid, it returns an error with the invalid fields.
func (v *Validator) Validate(structure any) error {
	if ok, err := v.validator.Struct(structure); !ok || err != nil {
		return err
	}

	return nil
}

func NewServer(d *dialer.Dialer, cli internalclient.Client, cfg *Config) *Server {
	r := echo.New()

	r.Binder = NewBinder()
	r.Validator = NewValidator()
	r.HideBanner = true
	r.HidePort = true

	handlers := &Handlers{
		Dialer: d,
		Client: cli,
		Config: cfg,
	}

	r.GET(HandleConnectionV1Path, handlers.HandleConnectionV1)
	r.GET(HandleConnectionV2Path, handlers.HandleConnectionV2)

	r.GET(HandleRevdialPath, echo.WrapHandler(revdial.ConnHandler(upgrader)))

	r.POST(HandleSSHClosePath, handlers.HandleSSHClose)
	r.GET(HandleHealthcheckPath, handlers.HandleHealthcheck)

	if cfg.Metrics {
		const HandleMetricsPath = "/metrics"

		p := metrics.NewMetrics()
		r.GET(HandleMetricsPath, echo.WrapHandler(
			promhttp.HandlerFor(p, promhttp.HandlerOpts{})),
		)
	}

	if cfg.WebEndpoints {
		// NOTE: The `/http/proxy` endpoint is invoked by the NGINX gateway when a tunnel URL is accessed. It processes
		// the `X-Address` and `X-Path` headers, which specify the tunnel's address and the target path on the server,
		// returning an error related to the connection to device or what was returned from the server inside the tunnel.
		r.Any(HandleHTTPProxyPath, handlers.HandleHTTPProxy)
	}

	return &Server{
		Config:   cfg,
		Router:   r,
		Handlers: handlers,
	}
}
