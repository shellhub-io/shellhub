package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/server/api/services"
)

// The API key routes, relative to the API's base path.
const (
	CreateAPIKeyURL = "/namespaces/api-key"
	ListAPIKeysURL  = "/namespaces/api-key"
	UpdateAPIKeyURL = "/namespaces/api-key/:name"
	DeleteAPIKeyURL = "/namespaces/api-key/:name"

	CreateAPIKeySSHIdentityURL = "/namespaces/api-key/:name/ssh-identities"
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
func (h *Handler) ListAPIKeys(c *gateway.Context) error {
	req := new(requests.ListAPIKey)

	if err := c.Bind(req); err != nil {
		return err
	}

	req.Paginator.Normalize()

	if req.Sorter.By == "" {
		req.Sorter.By = "expires_in"
	}

	if req.Sorter.Order == "" {
		req.Sorter.Order = "desc"
	}

	if err := query.ValidateSorter(&req.Sorter, services.APIKeySortFields); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	res, count, err := h.service.ListAPIKeys(c.Ctx(), req)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, res)
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

// CreateAPIKeySSHIdentity enrolls an SSH public key an API key owns, which is how an automation
// is given a way to reach a device. The route is gated by SSHIdentityManage rather than
// SSHIdentityAdd: adding is about one's own key, and this adds one to something else.
func (h *Handler) CreateAPIKeySSHIdentity(c *gateway.Context) error {
	req := new(requests.APIKeySSHIdentityCreate)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if c.Tenant() != nil {
		req.TenantID = c.Tenant().ID
	}

	identity, err := h.service.CreateAPIKeySSHIdentity(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, identity)
}
