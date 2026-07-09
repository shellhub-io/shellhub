package routes

import (
	"net/http"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/errors"
)

const (
	URLUpdateUser                   = "/users"
	URLDeprecatedUpdateUser         = "/users/:id/data"
	URLDeprecatedUpdateUserPassword = "/users/:id/password"   //nolint:gosec
	URLCreateUserActivationToken    = "/users/:id/activation" //nolint:gosec
	URLActivateUser                 = "/users/:id/activate"
)

const (
	// ParamUserName User's username.
	ParamUserName = "username"
)

func (h *Handler) UpdateUser(c gateway.Context) error {
	req := new(requests.UpdateUser)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if fields, err := h.service.UpdateUser(c.Ctx(), req); err != nil {
		// FIXME: API compatibility.
		//
		// The UI uses the fields with error messages to identify if it is invalid or duplicated.
		var e errors.Error
		if ok := errors.As(err, &e); !ok {
			return err
		}

		switch e.Code {
		case services.ErrCodeInvalid:
			if len(fields) > 1 {
				return c.JSON(http.StatusBadRequest, fields)
			}

			return c.NoContent(http.StatusBadRequest)
		case services.ErrCodeDuplicated:
			return c.JSON(http.StatusConflict, fields)
		default:
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}

// CreateUserActivationToken mints a one-time activation link token for a provisioned account.
// It is restricted to admins (enforced in the service on the resolved actor): the admin
// provisions the account and hands the resulting link to the user out of band, so the admin
// never learns or sets the password.
func (h *Handler) CreateUserActivationToken(c gateway.Context) error {
	req := new(requests.CreateUserActivation)
	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	token, expiresAt, err := h.service.CreateUserActivationToken(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}{Token: token, ExpiresAt: expiresAt})
}

// ActivateUser completes a provisioned account from the activation link: it is public because
// the one-time token in the body is the credential (the user has no password yet).
func (h *Handler) ActivateUser(c gateway.Context) error {
	req := new(requests.ActivateUser)
	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.service.ActivateUser(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) UpdateUserPassword(c gateway.Context) error {
	var req requests.UserPasswordUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.service.UpdatePasswordUser(c.Ctx(), req.ID, req.CurrentPassword, req.NewPassword); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
