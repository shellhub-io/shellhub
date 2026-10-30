package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/server/api/pkg/authctx"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	errs "github.com/shellhub-io/shellhub/server/api/routes/errors"
)

// The user routes, relative to the API's base path. The deprecated spellings take the user's
// ID in the path; the current ones act on the authenticated caller.
const (
	URLUpdateUser                   = "/users"
	URLDeprecatedUpdateUser         = "/users/:id/data"
	URLDeprecatedUpdateUserPassword = "/users/:id/password" //nolint:gosec
	URLRevokeUserTokens             = "/users/tokens"       //nolint:gosec // a route path, not a credential
	URLRecordOnboardingSurvey       = "/users/onboarding-survey"
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

// RevokeUserTokens revokes every token the caller holds, the one it calls with included.
func (h *Handler) RevokeUserTokens(c *gateway.Context) error {
	var req requests.RevokeUserTokens
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.service.RevokeUserTokens(c.Ctx(), req.UserID); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// RecordOnboardingSurvey records whether the caller answered or skipped the onboarding survey. It
// answers 401 without a user token, 403 to an API key and to an administrator signed in as the
// user, 400 for any other outcome, pending included, 404 when the caller's account no longer
// exists, and 500 when the outcome cannot be saved.
func (h *Handler) RecordOnboardingSurvey(c *gateway.Context) error {
	var req requests.RecordOnboardingSurvey
	if err := c.Bind(&req); err != nil {
		return err
	}

	if authctx.UserImpersonated(c.Ctx(), req.UserID) {
		return errs.NewErrForbidden(nil)
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.service.RecordOnboardingSurvey(c.Ctx(), &req); err != nil {
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
