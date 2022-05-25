package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	UpdateUserDataURL     = "/users/:id/data"
	UpdateUserPasswordURL = "/users/:id/password" //nolint:gosec
)

const (
	ParamUserID   = "id"
	ParamUserName = "username"
)

func (h *Handler) UpdateUserData(c gateway.Context) error {
	var user models.User

	if err := c.Bind(&user); err != nil {
		return err
	}

	if fields, err := h.service.UpdateDataUser(c.Ctx(), &user, c.Param(ParamUserID)); err != nil {
		// FIXME: API compatibility
		//
		// The UI uses the fields with error messages to identify if it is invalid or duplicated.
		e, ok := err.(errors.Error)
		if !ok {
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
	var req struct {
		// NIST 800-53.
		// https://davintechgroup.com/toolkit/password-requirements-gdpr-iso-27001-27002-pci-dss-nist-800-53/
		// https://pages.nist.gov/800-63-3/sp800-63b.html

		// The current should contains 8 to 64 characters,be an alphanumeric/ascii and contains, at least, one special character.
		CurrentPassword string `json:"current_password" validate:"required,min=8,max=64,nefield=NewPassword,alphanum,ascii,containsany=!@#$%^&*()_+-=[]{}<>?"`
		// The new password should contains 8 to 64 characters, don't be equal to the current password, be an alphanumeric/ascii and contains, at least, one special character.
		NewPassword string `json:"new_password" validate:"required,min=8,max=64,nefield=CurrentPassword,alphanum,ascii,containsany=!@#$%^&*()_+-=[]{}<>?"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.service.UpdatePasswordUser(c.Ctx(), req.CurrentPassword, req.NewPassword, c.Param(ParamUserID)); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
