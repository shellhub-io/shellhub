package routes

import (
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/server/api/pkg/origin"
	routesmiddleware "github.com/shellhub-io/shellhub/server/api/routes/middleware"
	svc "github.com/shellhub-io/shellhub/server/api/services"
)

// Handler holds what every route needs: the service layer to call and the authenticator that
// guards them.
type Handler struct {
	service svc.Service

	authn *routesmiddleware.Authenticator

	edition envs.Edition

	deviceListMaxPageSize int64
	geoipSource           string

	trust *origin.Trust
}

// NewHandler returns a handler over s with no authenticator. Until WithAuthenticator sets one,
// NewRouter declares no route reachable anonymously or with a device token.
func NewHandler(s svc.Service) *Handler {
	return &Handler{
		service: s,
		edition: envs.Community,
		trust:   origin.Loopback(),
	}
}

// WithAuthenticator hands the authenticator to the router so the anonymous
// allowlist can be declared against it, by the core and by each extension.
func (h *Handler) WithAuthenticator(authn *routesmiddleware.Authenticator) *Handler {
	h.authn = authn

	return h
}
