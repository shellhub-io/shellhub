package routes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	errs "github.com/shellhub-io/shellhub/api/routes/errors"
	svc "github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/jwttoken"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
)

const (
	AuthRequestURL           = "/auth"
	AuthDeviceURL            = "/devices/auth"
	AuthDeviceURLV2          = "/auth/device"
	AuthUserURL              = "/login"
	AuthUserURLV2            = "/auth/user"
	AuthUserTokenInternalURL = "/auth/token/:id"     //nolint:gosec
	AuthUserTokenPublicURL   = "/auth/token/:tenant" //nolint:gosec
	AuthPublicKeyURL         = "/auth/ssh"
	AuthMFAURL               = "/auth/mfa"
)

// AuthRequest is a proxy-level authentication middleware. It decodes a specified
// authentication hash (e.g. JWT tokens and API keys), sets the credentials in
// headers, and redirects to the original endpoint.
//
// The following sequential diagram represents the authentication pipeline:
//
//	+------+       +----------------+        +----------+
//	| User |       | /internal/auth |        | /api/... |
//	+------+       +----------------+        +----------+
//	   |              |                         |
//	   | Send Request |                         |
//	   |------------->|                         |
//	   |              | Extract and decode hash |
//	   |              | Set auth headers        |
//	   |              |------------------------>|
//	   |              |                         | Execute the target endpoint
//	   |                                        |
//	   | Send response back to the user         |
//	   |<---------------------------------------|
//
// If the authentication fails for any reason, it must return the failed status
// without redirecting the request. A token can be use to authenticate either a
// device or a user.
func (h *Handler) AuthRequest(c gateway.Context) error {
	if key := c.Request().Header.Get("X-API-Key"); key != "" {
		apiKey, err := h.service.AuthAPIKey(c.Ctx(), key)
		if err != nil {
			return err
		}

		c.Response().Header().Set("X-Tenant-ID", apiKey.TenantID)
		c.Response().Header().Set("X-Role", apiKey.Role.String())
		c.Response().Header().Set("X-API-KEY", key)

		return c.NoContent(http.StatusOK)
	}

	bearerToken := c.Request().Header.Get("Authorization")
	claims, err := jwttoken.ClaimsFromBearerToken(h.service.PublicKey(), bearerToken)
	if err != nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	switch claims := claims.(type) {
	case *authorizer.DeviceClaims:
		c.Response().Header().Set("X-Device-UID", claims.UID)
		c.Response().Header().Set("X-Tenant-ID", claims.TenantID)
	case *authorizer.UserClaims:
		// As the role is a dynamic attribute, and a JWT token must be stateless, we need to retrieve the role
		// every time this middleware is invoked (generally from the cache).
		if claims.TenantID != "" {
			role, err := h.service.GetUserRole(c.Ctx(), claims.TenantID, claims.ID)
			if err != nil {
				return err
			}

			claims.Role = authorizer.RoleFromString(role)
		}

		c.Response().Header().Set("X-ID", claims.ID)
		c.Response().Header().Set("X-Username", claims.Username)
		c.Response().Header().Set("X-Tenant-ID", claims.TenantID)
		c.Response().Header().Set("X-Role", claims.Role.String())
	default:
		return c.NoContent(http.StatusUnauthorized)
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) AuthDevice(c gateway.Context) error {
	var req requests.DeviceAuth
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	ip := c.Request().Header.Get("X-Real-IP")
	res, err := h.service.AuthDevice(c.Ctx(), req, ip)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) AuthUser(c gateway.Context) error {
	req := new(requests.UserAuth)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	res, lockout, mfaToken, err := h.service.AuthUser(c.Ctx(), req, c.RealIP())
	c.Response().Header().Set("X-Account-Lockout", strconv.FormatInt(lockout, 10))
	c.Response().Header().Set("X-MFA-Token", mfaToken)

	if lockout > 0 {
		return c.NoContent(http.StatusTooManyRequests)
	}

	if mfaToken != "" {
		return c.NoContent(http.StatusUnauthorized)
	}

	if err != nil {
		switch {
		case errors.Is(err, svc.ErrUserNotFound):
			return errs.NewErrUnauthorized(err)
		default:
			return err
		}
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) CreateUserToken(c gateway.Context) error {
	req := new(requests.CreateUserToken)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	res, err := h.service.CreateUserToken(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) AuthPublicKey(c gateway.Context) error {
	var req requests.PublicKeyAuth
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	res, err := h.service.AuthPublicKey(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}
