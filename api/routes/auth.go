package routes

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	errs "github.com/shellhub-io/shellhub/api/routes/errors"
	svc "github.com/shellhub-io/shellhub/api/services"
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

const (
	// AuthRequestUserToken is the type of the token used to authenticate a user.
	AuthRequestUserToken = "user"
	// AuthRequestDeviceToken is the type of the token used to authenticate a device.
	AuthRequestDeviceToken = "device"
)

// AuthRequest is a special handler that works as an authentication middleware at proxy level. It passes
// the authentication attributes in headers to the next request or returns a [http.StatusUnauthorized] error.
//
// TODO: explain why
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

	claims := jwttoken.ClaimsFromBearer(h.service.PublicKey(), c.Request().Header.Get("Authorization"))

	switch claims.Kind {
	case jwttoken.KindDeviceClaims:
		break
	case jwttoken.KindUserClaims:
		// As the role is a dynamic attribute, and a JWT token must be stateless, we need to retrieve the role
		// every time this middleware is invoked (generally from the cache).
		if claims.UserClaims.TenantID != "" {
			if err := h.service.FillClaimsRole(c.Ctx(), &claims.UserClaims); err != nil {
				return err
			}
		}
	case jwttoken.KindUnknownClaims:
		return c.NoContent(http.StatusUnauthorized)
	}

	for k, v := range claims.Headers() {
		c.Response().Header().Set(k, v)
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

	// If the tenant id is not in the parameters, we get it from the header.
	if req.TenantID == "" {
		req.TenantID = c.Request().Header.Get("X-Tenant-ID")
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
