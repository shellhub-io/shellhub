package routes

import (
	"context"
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
)

// The API key routes, relative to the API's base path.
const (
	CreateAPIKeyURL = "/namespaces/api-key"
	ListAPIKeysURL  = "/namespaces/api-key"
	UpdateAPIKeyURL = "/namespaces/api-key/:name"
	DeleteAPIKeyURL = "/namespaces/api-key/:name"
)

// CreateAPIKey mints a key for the caller's namespace and returns its plaintext, which is the
// only time the plaintext is available.
func (h *Handler) CreateAPIKey(c *gateway.Context) error {
	req := new(requests.CreateAPIKey)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	res, err := h.service.CreateAPIKey(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// ListAPIKeys serves the namespace's keys, without their plaintext.
func (h *Handler) ListAPIKeys(ctx context.Context, _ scope.Scope, _ gateway.Actor, req *requests.ListAPIKey) ([]models.APIKey, int, error) {
	return h.service.ListAPIKeys(ctx, req)
}

// UpdateAPIKey renames a key or changes the role it acts with.
func (h *Handler) UpdateAPIKey(c *gateway.Context) error {
	req := new(requests.UpdateAPIKey)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.service.UpdateAPIKey(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// DeleteAPIKey revokes a key.
func (h *Handler) DeleteAPIKey(c *gateway.Context) error {
	req := new(requests.DeleteAPIKey)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.service.DeleteAPIKey(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
