package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	UpdateUserDataURL     = "/users/:id/data"
	UpdateUserPasswordURL = "/users/:id/password" //nolint:gosec
)

const (
	// ParamUserName User's username.
	ParamUserName = "username"
)

func (h *Handler) updateUser() *Route {
	return &Route{
		endpoint:              "/users/:id/data",
		method:                MethodPatch,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           true,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var req requests.UserDataUpdate
			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			if fields, err := h.service.UpdateDataUser(c.Ctx(), req.ID, models.UserData{
				Name:     req.Name,
				Username: req.Username,
				Email:    req.Email,
			}); err != nil {
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
		},
	}
}

func (h *Handler) updateUserPassword() *Route {
	return &Route{
		endpoint:              "/users/:id/password",
		method:                MethodPatch,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           true,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
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
		},
	}
}
