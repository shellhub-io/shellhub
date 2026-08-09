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

func (h *Handler) AuthDevice(c *gateway.Context) error {
	var req requests.DeviceAuth
	if err := c.Bind(&req); err != nil {
		return err
	}

	// NOTE: The previous version of the Agent in Connector mode could send the container's name without converting
	// the dot character to an underscore, which is not supported in ShellHub device naming. To prevent validation
	// errors with this old version, we are implementing a server-side change to handle this conversion.
	// TODO: This modification could be in the service layer.
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
