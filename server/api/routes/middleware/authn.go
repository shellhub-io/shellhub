package middleware

import (
	"context"
	"crypto/rsa"
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/authctx"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	log "github.com/sirupsen/logrus"
)

// AuthnService is the slice of the service layer the authenticator needs to turn
// a credential into an identity.
type AuthnService interface {
	AuthAPIKey(ctx context.Context, key string) (*models.APIKey, error)
	ResolveNamespaceRole(ctx context.Context, tenantID, userID string) (*models.Namespace, string, error)
	GetUserAdmin(ctx context.Context, userID string) (bool, error)
	PublicKey() *rsa.PublicKey
}

// exemptPrefixes are route prefixes the authenticator ignores entirely.
//
// /metrics is scraped in-cluster and carries no credential. Nothing else is
// exempt: an exempt route keeps whatever identity headers the client sent,
// since the scrubbing happens on the authenticated path.
var exemptPrefixes = []string{"/metrics"}

// Authenticator resolves the caller's identity from the request credentials and
// stamps it onto the canonical headers that [gateway.Context] reads.
//
// It fails closed: every route requires a valid credential unless it was
// registered with [Authenticator.AllowAnonymous]. A new route is therefore
// protected by default, and a typo in the allowlist leaves a public route
// returning 401 rather than leaving a private one open.
type Authenticator struct {
	service AuthnService

	// anonymous holds the routes reachable without a credential, keyed by
	// method and the registered route pattern (not the request path), so
	// matching cannot be fooled by path traversal or a longer path that
	// happens to share a prefix.
	anonymous map[string]struct{}
}

func NewAuthenticator(service AuthnService) *Authenticator {
	return &Authenticator{
		service:   service,
		anonymous: make(map[string]struct{}),
	}
}

// AnyMethod marks a route anonymous for every HTTP method, for routes
// registered with echo's Any.
const AnyMethod = "*"

// AllowAnonymous marks a route as reachable without a credential. The path must
// be the full registered pattern, including the group prefix. Pass [AnyMethod]
// as the method to cover every verb.
//
// An anonymous route still gets its identity resolved when the caller happens
// to present a valid credential, and still has forged identity headers
// stripped; it simply does not demand one.
func (a *Authenticator) AllowAnonymous(method, path string) {
	a.anonymous[method+" "+path] = struct{}{}
}

// AnonymousRoutes returns the registered anonymous routes as "METHOD path"
// keys. Tests use it to assert the allowlist matches the router.
func (a *Authenticator) AnonymousRoutes() []string {
	routes := make([]string, 0, len(a.anonymous))
	for route := range a.anonymous {
		routes = append(routes, route)
	}

	return routes
}

// Middleware authenticates the request. Register it with echo's Use so it runs
// after routing, where the matched route pattern is available.
func (a *Authenticator) Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		if isExempt(c.Path()) {
			return next(c)
		}

		anonymous := a.isAnonymous(c)

		identity, err := a.Resolve(c)
		if err != nil {
			// A route that needs no credential must not fail because the caller
			// volunteered a bad one.
			if !anonymous {
				return err
			}

			identity = nil
		}

		if identity != nil && strings.HasPrefix(c.Path(), adminAPIPrefix) {
			scoped := identity.WithoutUserScope()
			identity = &scoped
		}

		// Stamped unconditionally: on the anonymous and rejected paths this is
		// what strips a forged identity header the edge forwarded verbatim.
		identity.WriteTo(c.Request().Header)

		if identity == nil && !anonymous {
			return c.NoContent(http.StatusUnauthorized)
		}

		return next(c)
	}
}

// adminAPIPrefix is the admin panel's surface. Identities there are narrowed by
// [gateway.Identity.WithoutUserScope].
const adminAPIPrefix = "/admin/api"

func isExempt(path string) bool {
	for _, prefix := range exemptPrefixes {
		if path == prefix || strings.HasPrefix(path, prefix+"/") {
			return true
		}
	}

	return false
}

func (a *Authenticator) isAnonymous(c *echo.Context) bool {
	if _, ok := a.anonymous[c.Request().Method+" "+c.Path()]; ok {
		return true
	}

	_, ok := a.anonymous[AnyMethod+" "+c.Path()]

	return ok
}

// Resolve turns the request's credential into an identity. It returns a nil
// identity when the request carries no usable credential, and an error only
// when a credential was presented and could not be honoured.
func (a *Authenticator) Resolve(c *echo.Context) (*gateway.Identity, error) {
	if key := c.Request().Header.Get("X-API-Key"); key != "" {
		apiKey, err := a.service.AuthAPIKey(c.Request().Context(), key)
		if err != nil {
			// Same rule as the bearer token below: a key that does not resolve
			// yields no identity rather than an error. Propagating it would
			// surface the store's own status -- 404 for an unknown key, 400 for
			// an expired one -- when the caller's problem is the credential.
			// The edge proxy turned every non-2xx from the authentication
			// subrequest into a 401, including when its store was down. Log the
			// error for the same reason the bearer path does: an outage is
			// otherwise indistinguishable from an unusable key. The key itself
			// is a credential and never goes to the log.
			log.WithError(err).Warn("failed to resolve the API key")

			return nil, nil //nolint:nilerr
		}

		return &gateway.Identity{
			TenantID: apiKey.TenantID,
			Role:     apiKey.Role,
			APIKey:   key,
		}, nil
	}

	bearer := c.Request().Header.Get("Authorization")
	if bearer == "" {
		return nil, nil
	}

	claims, err := jwttoken.ClaimsFromBearerToken(a.service.PublicKey(), bearer)
	if err != nil {
		// A token that does not verify yields no identity rather than an error:
		// the caller decides what that means. A protected route rejects it, and
		// an anonymous route proceeds without an identity — which is how the
		// edge proxy behaved when it skipped the authentication subrequest.
		return nil, nil //nolint:nilerr
	}

	switch claims := claims.(type) {
	case *authorizer.DeviceClaims:
		return &gateway.Identity{
			DeviceUID: claims.UID,
			TenantID:  claims.TenantID,
		}, nil
	case *authorizer.UserClaims:
		// Role and admin are dynamic attributes that can change while a JWT is
		// still valid, so we re-fetch them from the store on every request.
		//
		// Both lookups follow the same rule as the API key above: any failure
		// yields no identity rather than an error, so a token whose namespace,
		// membership or user is gone gets a 401 telling the caller to
		// re-authenticate instead of the store's own status as a 500. As with
		// the API key, that deliberately covers a store outage too, so both log
		// the error they swallow: otherwise an outage is indistinguishable from
		// a stale token and shows up only as a rise in 401s.
		if claims.TenantID != "" {
			ns, role, err := a.service.ResolveNamespaceRole(c.Request().Context(), claims.TenantID, claims.ID)
			if err != nil {
				log.WithError(err).
					WithFields(log.Fields{"user_id": claims.ID, "tenant_id": claims.TenantID}).
					Warn("failed to resolve the token's role")

				return nil, nil //nolint:nilerr
			}

			claims.Role = authorizer.RoleFromString(role)

			// Resolving the role already read the namespace, and the device-limit check in
			// the handler would otherwise read it a second time.
			c.SetRequest(c.Request().WithContext(
				authctx.WithNamespaceDeviceLimit(c.Request().Context(), ns.TenantID, ns.DeviceLimit()),
			))
		}

		admin, err := a.service.GetUserAdmin(c.Request().Context(), claims.ID)
		if err != nil {
			log.WithError(err).
				WithField("user_id", claims.ID).
				Warn("failed to resolve the token's admin status")

			return nil, nil //nolint:nilerr
		}

		return &gateway.Identity{
			ID:       claims.ID,
			Username: claims.Username,
			TenantID: claims.TenantID,
			Role:     claims.Role,
			Admin:    admin,
		}, nil
	}

	return nil, nil
}
