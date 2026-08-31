package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/envs"
	routesmiddleware "github.com/shellhub-io/shellhub/server/api/routes/middleware"
)

const publicAPIPrefix = "/api"

func registerAnonymousRoutes(authn *routesmiddleware.Authenticator) {
	allow := func(method, path string) {
		authn.AllowAnonymous(method, publicAPIPrefix+path)
	}

	allow(http.MethodPost, AuthLocalUserURL)
	allow(http.MethodPost, RegisterUserURL)

	allow(http.MethodPost, AuthDeviceURL)

	allow(http.MethodPost, CreateDevicePairingURL)
	allow(http.MethodGet, GetDevicePairingStatusURL)

	allow(http.MethodPost, EnrollmentCallbackURL)

	allow(http.MethodGet, URLResolveInvitation)

	allow(http.MethodGet, HealthCheckURL)
	allow(http.MethodGet, GetSystemInfoURL)
	allow(http.MethodGet, GetSystemDownloadInstallScriptURL)

	if !envs.IsCloud() {
		allow(http.MethodPost, SetupEndpoint)
	}
}
