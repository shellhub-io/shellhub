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

// The install key routes, relative to the API's base path.
const (
	CreateInstallKeyURL  = "/namespaces/install-key"
	ListInstallKeysURL   = "/namespaces/install-key"
	UpdateInstallKeyURL  = "/namespaces/install-key/:name"
	RevealInstallKeyURL  = "/namespaces/install-key/:name/reveal"
	HistoryInstallKeyURL = "/namespaces/install-key/:id/history"

	// EnrollmentCallbackURL is the public, token-authenticated endpoint a webhook integrator POSTs its
	// deferred decision to. The token in the path is the credential (no API key/JWT).
	EnrollmentCallbackURL = "/devices/enroll/callback/:token"
)

// CreateInstallKey mints a key an agent can enrol with, returning its plaintext.
func (h *Handler) CreateInstallKey(c *gateway.Context) error {
	req := new(requests.CreateInstallKey)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	res, err := h.service.CreateInstallKey(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// ListInstallKeys serves the namespace's install keys, without their plaintext.
func (h *Handler) ListInstallKeys(ctx context.Context, _ scope.Scope, _ gateway.Actor, req *requests.ListInstallKey) ([]models.InstallKey, int, error) {
	return h.service.ListInstallKeys(ctx, req)
}

// UpdateInstallKey changes a key's name, expiry or the device attributes it pre-assigns.
func (h *Handler) UpdateInstallKey(c *gateway.Context) error {
	req := new(requests.UpdateInstallKey)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.service.UpdateInstallKey(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// RevealInstallKey returns a key's plaintext, decrypting the copy kept for this purpose.
func (h *Handler) RevealInstallKey(c *gateway.Context) error {
	req := new(requests.RevealInstallKey)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	key, err := h.service.RevealInstallKey(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &responses.RevealInstallKey{Key: key})
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

// HistoryInstallKey serves the record of what a key has been used for.
func (h *Handler) HistoryInstallKey(ctx context.Context, _ scope.Scope, _ gateway.Actor, req *requests.ListInstallKeyEvents) ([]models.InstallKeyEvent, int, error) {
	return h.service.ListInstallKeyEvents(ctx, req)
}
