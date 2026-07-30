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

	// Obtaining a credential cannot itself require one.
	allow(http.MethodPost, AuthLocalUserURL)
	allow(http.MethodPost, RegisterUserURL)

	// An agent authenticates with its install key or tenant, carried in the body.
	allow(http.MethodPost, AuthDeviceURL)

	// Tenant-less pairing: the short-lived code is the credential. Minting a
	// pre-authorized code (/pairing/prepare) and accepting one both act on behalf
	// of a user and stay authenticated.
	allow(http.MethodPost, CreateDevicePairingURL)
	allow(http.MethodGet, GetDevicePairingStatusURL)

	// Deferred enrollment: the signed token in the path is the credential.
	allow(http.MethodPost, EnrollmentCallbackURL)

	// A logged-out invitee resolves their invite by its code before any account
	// exists.
	allow(http.MethodGet, URLResolveInvitation)

	// Instance metadata and the install script, served to unauthenticated
	// clients by design.
	allow(http.MethodGet, HealthCheckURL)
	allow(http.MethodGet, GetSystemInfoURL)
	allow(http.MethodGet, GetSystemDownloadInstallScriptURL)

	// First-boot setup, registered only outside Cloud — mirror that here so the
	// allowlist never names a route the router does not have.
	if !envs.IsCloud() {
		allow(http.MethodPost, SetupEndpoint)
	}
}
