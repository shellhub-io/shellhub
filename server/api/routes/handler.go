package routes

import (
	"github.com/shellhub-io/shellhub/pkg/websocket"
	routesmiddleware "github.com/shellhub-io/shellhub/server/api/routes/middleware"
	svc "github.com/shellhub-io/shellhub/server/api/services"
)

// Handler holds what every route needs: the service layer to call, the authenticator that
// guards them, and the upgrader for the routes that become WebSockets.
type Handler struct {
	service svc.Service
	// WebSocketUpgrader is used to turns a HTTP request into WebSocketUpgrader connection.
	WebSocketUpgrader websocket.Upgrader

	authn *routesmiddleware.Authenticator
}

// NewHandler returns a handler serving over s, upgrading WebSocket routes with w.
func NewHandler(s svc.Service, w websocket.Upgrader) *Handler {
	return &Handler{
		service:           s,
		WebSocketUpgrader: w,
	}
}

// WithAuthenticator hands the authenticator to the router so the anonymous
// allowlist can be declared against it, by the core and by each extension.
func (h *Handler) WithAuthenticator(authn *routesmiddleware.Authenticator) *Handler {
	h.authn = authn

	return h
}
