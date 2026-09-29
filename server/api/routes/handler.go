package routes

import (
	routesmiddleware "github.com/shellhub-io/shellhub/server/api/routes/middleware"
	svc "github.com/shellhub-io/shellhub/server/api/services"
)

// Handler holds what every route needs: the service layer to call and the authenticator that
// guards them.
type Handler struct {
	service svc.Service

	authn *routesmiddleware.Authenticator
}

// NewHandler returns a handler serving over s.
func NewHandler(s svc.Service) *Handler {
	return &Handler{
		service: s,
	}
}

// WithAuthenticator hands the authenticator to the router so the anonymous
// allowlist can be declared against it, by the core and by each extension.
func (h *Handler) WithAuthenticator(authn *routesmiddleware.Authenticator) *Handler {
	h.authn = authn

	return h
}
