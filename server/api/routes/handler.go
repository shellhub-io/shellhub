package routes

import (
	"github.com/shellhub-io/shellhub/pkg/websocket"
	routesmiddleware "github.com/shellhub-io/shellhub/server/api/routes/middleware"
	svc "github.com/shellhub-io/shellhub/server/api/services"
)

type Handler struct {
	service svc.Service
	// WebSocketUpgrader is used to turns a HTTP request into WebSocketUpgrader connection.
	WebSocketUpgrader websocket.Upgrader

	authn *routesmiddleware.Authenticator
}

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
