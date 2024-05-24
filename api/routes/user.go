package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/errors"
)

const (
	UpdateUserDataURL     = "/users/:id/data"
	UpdateUserPasswordURL = "/users/:id/password" //nolint:gosec
)

const (
	// ParamUserName User's username.
	ParamUserName = "username"
)

func (h *Handler) UpdateUserData(c gateway.Context) error {
	req := new(requests.UserDataUpdate)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	// TODO: remove the user id from the params and use only the authenticated header.
	if c.Param("id") != c.ID().ID {
		return services.NewErrForbidden(nil, nil)
	}

	if fields, err := h.service.UpdateDataUser(c.Ctx(), c.ID().ID, req); err != nil {
		// FIXME: API compatibility.
		//
		// The UI uses the fields with error messages to identify if it is invalid or duplicated.
		var e errors.Error
		if ok := errors.As(err, &e); !ok {
			return err
		}

		switch e.Code {
		case services.ErrCodeInvalid:
			return c.JSON(http.StatusBadRequest, fields)
		case services.ErrCodeDuplicated:
			return c.JSON(http.StatusConflict, fields)
		default:
			return err
		}
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
