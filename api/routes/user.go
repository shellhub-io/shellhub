package routes

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	UpdateUserDataURL     = "/users/:id/data"
	UpdateUserPasswordURL = "/users/:id/password" //nolint:gosec
)

func (h *Handler) UpdateUserData(c apicontext.Context) error {
	var user models.User

	if err := c.Bind(&user); err != nil {
		return err
	}

	ID := c.Param("id")

	if fields, err := h.service.UpdateDataUser(c.Ctx(), &user, ID); err != nil {
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
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}

	ID := c.Param("id")

	if req.CurrentPassword != "" {
		sum := sha256.Sum256([]byte(req.CurrentPassword))
		sumByte := sum[:]
		req.CurrentPassword = hex.EncodeToString(sumByte)
	}

	if req.NewPassword != "" {
		sum := sha256.Sum256([]byte(req.NewPassword))
		sumByte := sum[:]
		req.NewPassword = hex.EncodeToString(sumByte)
	}

	if err := h.service.UpdatePasswordUser(c.Ctx(), req.CurrentPassword, req.NewPassword, ID); err != nil {
		switch {
		case err == services.ErrUnauthorized:
			return c.NoContent(http.StatusForbidden)
		default:
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}
