package routes

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	errs "github.com/shellhub-io/shellhub/server/api/routes/errors"
	svc "github.com/shellhub-io/shellhub/server/api/services"
)

// The authentication routes. Each has a V2 spelling under /auth; the older paths remain
// because deployed agents and scripts still call them.
const (
	AuthDeviceURL            = "/devices/auth"
	AuthDeviceURLV2          = "/auth/device"
	AuthLocalUserURL         = "/login"
	AuthLocalUserURLV2       = "/auth/user"
	AuthUserTokenInternalURL = "/auth/token/:id"     //nolint:gosec
	AuthUserTokenPublicURL   = "/auth/token/:tenant" //nolint:gosec
	AuthPublicKeyURL         = "/auth/ssh"
	AuthMFAURL               = "/auth/mfa"
)

// AuthDevice authenticates an agent and enrols its device if the namespace has not seen it
// before, returning the token the agent then holds.
func (h *Handler) AuthDevice(c *gateway.Context) error {
	var req requests.DeviceAuth
	if err := c.Bind(&req); err != nil {
		return err
	}

	if strings.Contains(req.Hostname, ".") {
		req.Hostname = strings.ReplaceAll(req.Hostname, ".", "_")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	res, err := h.service.AuthDevice(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// AuthLocalUser authenticates a user by password, returning a token or, when the account has
// a second factor, the challenge to complete first.
func (h *Handler) AuthLocalUser(c *gateway.Context) error {
	req := new(requests.AuthLocalUser)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	res, lockout, mfaToken, err := h.service.AuthLocalUser(c.Ctx(), req, c.RealIP())
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

// CreateUserToken issues a token for a user without their password. It is reachable only
// internally, so that the SSH gateway can act on a user's behalf.
func (h *Handler) CreateUserToken(c *gateway.Context) error {
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

// AuthPublicKey authenticates by SSH key, which is how the SSH gateway checks a key the
// client offered before opening a session.
func (h *Handler) AuthPublicKey(c *gateway.Context) error {
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
