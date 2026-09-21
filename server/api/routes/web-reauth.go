package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	errs "github.com/shellhub-io/shellhub/server/api/routes/errors"
)

// WebReauthURL is where the browser submits a step-up factor. The enterprise
// overlay re-registers this exact path to add TOTP, so it is part of that
// contract.
const WebReauthURL = "/web-terminal/reauth"

// WebReauthVerify validates the logged-in user's step-up factor and, on success,
// refreshes the presented identity's re-auth window. Identity comes from the
// gateway-injected X-ID/X-Tenant-ID, never the body.
func (h *Handler) WebReauthVerify(c *gateway.Context) error {
	req := new(requests.WebReauthVerify)
	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	userID, ok := c.GetID()
	if !ok {
		return errs.NewErrUnauthorized(nil)
	}

	req.UserID = userID
	if c.Tenant() != nil {
		req.TenantID = c.Tenant().ID
	}

	confirmationCode, err := h.service.WebReauthVerify(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.SSHApprovalConfirmation{ConfirmationCode: confirmationCode})
}
