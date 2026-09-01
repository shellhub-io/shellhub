package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
)

// The user routes, relative to the API's base path. The deprecated spellings take the user's
// ID in the path; the current ones act on the authenticated caller.
const (
	URLUpdateUser                   = "/users"
	URLDeprecatedUpdateUser         = "/users/:id/data"
	URLDeprecatedUpdateUserPassword = "/users/:id/password" //nolint:gosec
)

const (
	// ParamUserName User's username.
	ParamUserName = "username"
)

// UpdateUser changes the caller's profile.
func (h *Handler) UpdateUser(c *gateway.Context) error {
	req := new(requests.UpdateUser)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.service.UpdateUser(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// UpdateUserPassword changes the caller's password, requiring the current one.
func (h *Handler) UpdateUserPassword(c *gateway.Context) error {
	var req requests.UserPasswordUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.service.UpdatePasswordUser(c.Ctx(), req.UserID, req.CurrentPassword, req.NewPassword); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
