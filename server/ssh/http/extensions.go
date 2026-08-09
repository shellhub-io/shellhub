package http

import (
	"github.com/labstack/echo/v5"
	routesmiddleware "github.com/shellhub-io/shellhub/server/api/routes/middleware"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
	log "github.com/sirupsen/logrus"
)

// TunnelExtension registers routes that carry traffic into a device over the
// reverse tunnel. Cloud and Enterprise builds register one in init() to serve
// web endpoints.
//
// The dialer is the reason this exists as its own extension point rather than
// as a route registered through routes.RegisterRouteExtension: it holds the
// live agent connections and is built during server setup, so a feature that
// proxies into a device has to be handed it. Everything else such a feature
// needs, it already owns.
type TunnelExtension func(router *echo.Echo, authn *routesmiddleware.Authenticator, d *dialer.Dialer) error

var tunnelExtensions []TunnelExtension

// RegisterTunnelExtension registers an extension. It must be called before
// Register runs; extensions are applied in registration order.
func RegisterTunnelExtension(ext TunnelExtension) {
	tunnelExtensions = append(tunnelExtensions, ext)
}

func applyTunnelExtensions(router *echo.Echo, authn *routesmiddleware.Authenticator, d *dialer.Dialer) error {
	for _, ext := range tunnelExtensions {
		if err := ext(router, authn, d); err != nil {
			log.WithError(err).Error("failed to apply tunnel extension")

			return err
		}
	}

	return nil
}
