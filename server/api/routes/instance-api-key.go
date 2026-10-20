package routes

import (
	"context"
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
)

// The instance API key routes, relative to the admin API's base path. They are mounted by the
// enterprise admin surface, which is what keeps them out of community builds and behind the
// admin authorization check.
const (
	CreateInstanceAPIKeyURL = "/instance-api-keys"       //nolint:gosec // G101: a route path, not a credential
	ListInstanceAPIKeysURL  = "/instance-api-keys"       //nolint:gosec // G101: a route path, not a credential
	DeleteInstanceAPIKeyURL = "/instance-api-keys/:name" //nolint:gosec // G101: a route path, not a credential
)

// CreateInstanceAPIKey mints an instance API key and returns its plaintext, which is the only
// time the plaintext is available.
func (h *Handler) CreateInstanceAPIKey(c *gateway.Context) error {
	req := new(requests.CreateInstanceAPIKey)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	res, err := h.service.CreateInstanceAPIKey(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// ListInstanceAPIKeys serves every instance API key, without their plaintext.
func (h *Handler) ListInstanceAPIKeys(ctx context.Context, _ scope.Scope, _ gateway.Actor, req *requests.ListInstanceAPIKey) ([]models.InstanceAPIKey, int64, error) {
	return h.service.ListInstanceAPIKeys(ctx, req)
}

// DeleteInstanceAPIKey revokes an instance API key, taking effect immediately.
func (h *Handler) DeleteInstanceAPIKey(c *gateway.Context) error {
	req := new(requests.DeleteInstanceAPIKey)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.service.DeleteInstanceAPIKey(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
