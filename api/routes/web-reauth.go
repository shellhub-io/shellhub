package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
)

// WebReauthURL is where the browser submits a step-up factor. The enterprise
// overlay re-registers this exact path to add TOTP, so it is part of that
// contract.
const WebReauthURL = "/web-terminal/reauth"

// WebReauthVerify validates the logged-in user's step-up factor and, on success,
// refreshes the presented identity's re-auth window. Identity comes from the
// gateway-injected X-ID/X-Tenant-ID, never the body.
func (h *Handler) WebReauthVerify(c gateway.Context) error {
	req := new(requests.WebReauthVerify)
	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	userID, ok := c.GetID()
	if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}

	req.UserID = userID
	if c.Tenant() != nil {
		req.TenantID = c.Tenant().ID
	}

	if err := h.service.WebReauthVerify(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
