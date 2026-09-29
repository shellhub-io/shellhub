package routes

import (
	"context"
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
)

// The provisioning key routes, relative to the API's base path.
const (
	CreateProvisioningKeyURL  = "/namespaces/provisioning-key"
	ListProvisioningKeysURL   = "/namespaces/provisioning-key"
	UpdateProvisioningKeyURL  = "/namespaces/provisioning-key/:name"
	RevealProvisioningKeyURL  = "/namespaces/provisioning-key/:name/reveal"
	HistoryProvisioningKeyURL = "/namespaces/provisioning-key/:id/history"

	// EnrollmentCallbackURL is the public, token-authenticated endpoint a webhook integrator POSTs its
	// deferred decision to. The token in the path is the credential (no API key/JWT).
	EnrollmentCallbackURL = "/devices/enroll/callback/:token"
)

// CreateProvisioningKey mints a key an agent can enrol with, returning its plaintext.
func (h *Handler) CreateProvisioningKey(c *gateway.Context) error {
	req := new(requests.CreateProvisioningKey)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	res, err := h.service.CreateProvisioningKey(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// ListProvisioningKeys serves the namespace's provisioning keys, without their plaintext.
func (h *Handler) ListProvisioningKeys(ctx context.Context, _ scope.Scope, _ gateway.Actor, req *requests.ListProvisioningKey) ([]models.ProvisioningKey, int, error) {
	return h.service.ListProvisioningKeys(ctx, req)
}

// UpdateProvisioningKey changes a key's name, expiry or the device attributes it pre-assigns.
func (h *Handler) UpdateProvisioningKey(c *gateway.Context) error {
	req := new(requests.UpdateProvisioningKey)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.service.UpdateProvisioningKey(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// RevealProvisioningKey returns a key's plaintext, decrypting the copy kept for this purpose.
func (h *Handler) RevealProvisioningKey(c *gateway.Context) error {
	req := new(requests.RevealProvisioningKey)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	key, err := h.service.RevealProvisioningKey(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &responses.RevealProvisioningKey{Key: key})
}

// EnrollmentCallback is called by an agent once it has enrolled, so the key's history records
// which devices it produced.
func (h *Handler) EnrollmentCallback(c *gateway.Context) error {
	req := new(requests.EnrollmentCallback)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.service.ResolveEnrollmentCallback(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// HistoryProvisioningKey serves the record of what a key has been used for.
func (h *Handler) HistoryProvisioningKey(ctx context.Context, _ scope.Scope, _ gateway.Actor, req *requests.ListProvisioningKeyEvents) ([]models.ProvisioningKeyEvent, int, error) {
	return h.service.ListProvisioningKeyEvents(ctx, req)
}
