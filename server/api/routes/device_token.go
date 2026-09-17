package routes

import (
	"net/http"

	routesmiddleware "github.com/shellhub-io/shellhub/server/api/routes/middleware"
)

func registerDeviceRoutes(authn *routesmiddleware.Authenticator) {
	allow := func(method, path string) {
		authn.AllowDevice(method, publicAPIPrefix+path)
	}

	allow(http.MethodPost, AuthPublicKeyURL)

	allow(http.MethodPost, CreateDeviceLoginCodeURL)
	allow(http.MethodGet, GetDeviceAuthStatusURL)
}
