package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/envs"
	routesmiddleware "github.com/shellhub-io/shellhub/server/api/routes/middleware"
)

// publicAPIPrefix is the group prefix the anonymous allowlist keys are built
// from; it must match the group the routes are registered on.
const publicAPIPrefix = "/api"

// registerAnonymousRoutes declares the core API's public surface: every /api
// route reachable without a credential. Anything absent requires one.
//
// This is the list a security review should read. It replaces the per-route
// `auth_request off` scattered through the edge proxy's configuration, where the
// same intent was expressed by prefix matching and was one shadowed location
// away from exposing a neighbouring route.
//
// The SSH and WebSocket transports declare their own entries where they
// register; Cloud and Enterprise do the same from their route extensions.
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
