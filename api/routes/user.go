package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/api/pkg/apicontext"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	UpdateUserDataURL     = "/users/:id/data"
	UpdateUserPasswordURL = "/users/:id/password" //nolint:gosec
)

func (h *Handler) UpdateUserData(c apicontext.Context) error {
	const UserID = "id"

	var user models.User
	if err := c.Bind(&user); err != nil {
		return err
	}

	if fields, err := h.service.UpdateDataUser(c.Ctx(), &user, c.Param(UserID)); err != nil {
		switch {
		case err == services.ErrBadRequest:
			return c.JSON(http.StatusBadRequest, fields)
		case err == services.ErrConflict:
			return c.JSON(http.StatusConflict, fields)
		default:
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) UpdateUserPassword(c apicontext.Context) error {
	const UserID = "id"

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := h.service.UpdatePasswordUser(c.Ctx(), req.CurrentPassword, req.NewPassword, c.Param(UserID)); err != nil {
		switch {
		case err == services.ErrBadRequest:
			return c.NoContent(http.StatusBadRequest)
		case err == services.ErrUnauthorized:
			return c.NoContent(http.StatusUnauthorized)
		default:
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}
