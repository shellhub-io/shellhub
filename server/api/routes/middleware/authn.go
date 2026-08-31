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
			if !anonymous {
				return err
			}

			identity = nil
		}

		if identity != nil && strings.HasPrefix(c.Path(), adminAPIPrefix) {
			scoped := identity.WithoutUserScope()
			identity = &scoped
		}

		identity.WriteTo(c.Request().Header)

		if identity == nil && !anonymous {
			return c.NoContent(http.StatusUnauthorized)
		}

		return next(c)
	}
}

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
			log.WithError(err).Warn("failed to resolve the API key")

			return nil, nil
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
		return nil, nil //nolint:nilerr
	}

	switch claims := claims.(type) {
	case *authorizer.DeviceClaims:
		return &gateway.Identity{
			DeviceUID: claims.UID,
			TenantID:  claims.TenantID,
		}, nil
	case *authorizer.UserClaims:
		if claims.TenantID != "" {
			ns, role, err := a.service.ResolveNamespaceRole(c.Request().Context(), claims.TenantID, claims.ID)
			if err != nil {
				log.WithError(err).
					WithFields(log.Fields{"user_id": claims.ID, "tenant_id": claims.TenantID}).
					Warn("failed to resolve the token's role")

				return nil, nil
			}

			claims.Role = authorizer.RoleFromString(role)

			c.SetRequest(c.Request().WithContext(
				authctx.WithNamespaceDeviceLimit(c.Request().Context(), ns.TenantID, ns.DeviceLimit()),
			))
		}

		admin, err := a.service.GetUserAdmin(c.Request().Context(), claims.ID)
		if err != nil {
			log.WithError(err).
				WithField("user_id", claims.ID).
				Warn("failed to resolve the token's admin status")

			return nil, nil
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
