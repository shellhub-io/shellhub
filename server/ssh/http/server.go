package http

import (
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/revdial"
	routesmiddleware "github.com/shellhub-io/shellhub/server/api/routes/middleware"
	"github.com/shellhub-io/shellhub/server/api/services"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
	log "github.com/sirupsen/logrus"
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
type Config struct {
	// RequireAcceptedTunnel refuses the reverse tunnel for a device that is not accepted (pending or
	// rejected). Off by default; opt-in per instance.
	RequireAcceptedTunnel bool
}

var ErrDeviceTunnelDial = errors.New("failed to connect to device")

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
//
// authn is the API's authenticator, used to declare which of these routes are
// reachable without a credential. It may be nil in tests.
func Register(router *echo.Echo, authn *routesmiddleware.Authenticator, d *dialer.Dialer, service services.Service, cfg *Config) *Handlers {
	handlers := &Handlers{
		Dialer:  d,
		Service: service,
		Config:  cfg,
	}

	allowAnonymous := func(method, path string) {
		if authn != nil {
			authn.AllowAnonymous(method, path)
		}
	}

	// Both agent transports authenticate with the device's own token.
	router.GET(HandleConnectionV1Path, handlers.HandleConnectionV1)
	router.GET(HandleConnectionV2Path, handlers.HandleConnectionV2)

	// The revdial handshake carries no credential of its own: it opens a further
	// logical session over a connection already authenticated on
	// HandleConnectionV1Path.
	router.GET(HandleRevdialPath, echo.WrapHandler(revdial.ConnHandler(upgrader)))
	allowAnonymous(http.MethodGet, HandleRevdialPath)

	// Registered at the root rather than under the API group: the group carries
	// license enforcement in Enterprise, which this route does not have today.
	// HandleSSHClose enforces the SessionClose permission itself, from the role
	// the authenticator resolves.
	router.POST(HandleSSHClosePath, handlers.HandleSSHClose)

	if err := applyTunnelExtensions(router, authn, d); err != nil {
		log.WithError(err).Error("failed to register the tunnel extensions")
	}

	return handlers
}
