package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
)

// The Access Policy routes, relative to the API's base path.
const (
	ListAccessPoliciesURL = "/access-policies"
	GetAccessPolicyURL    = "/access-policies/:id"
	CreateAccessPolicyURL = "/access-policies"
	UpdateAccessPolicyURL = "/access-policies/:id"
	DeleteAccessPolicyURL = "/access-policies/:id"
)

// ListAccessPolicies serves the policy list for the caller's namespace.
func (h *Handler) ListAccessPolicies(c *gateway.Context) error {
	var tenant string
	if c.Tenant() != nil {
		tenant = c.Tenant().ID
	}

	list, err := h.service.ListAccessPolicies(c.Ctx(), tenant)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(len(list)))

	return c.JSON(http.StatusOK, list)
}

// GetAccessPolicy serves a single policy by id.
func (h *Handler) GetAccessPolicy(c *gateway.Context) error {
	var req requests.AccessPolicyGet
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if c.Tenant() != nil {
		req.TenantID = c.Tenant().ID
	}

	policy, err := h.service.GetAccessPolicy(c.Ctx(), &req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, policy)
}

// CreateAccessPolicy adds a policy to the caller's namespace.
func (h *Handler) CreateAccessPolicy(c *gateway.Context) error {
	var req requests.AccessPolicyCreate
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if c.Tenant() != nil {
		req.TenantID = c.Tenant().ID
	}

	policy, err := h.service.CreateAccessPolicy(c.Ctx(), &req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, policy)
}

// UpdateAccessPolicy replaces a policy's rules.
func (h *Handler) UpdateAccessPolicy(c *gateway.Context) error {
	var req requests.AccessPolicyUpdate
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if c.Tenant() != nil {
		req.TenantID = c.Tenant().ID
	}

	policy, err := h.service.UpdateAccessPolicy(c.Ctx(), &req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, policy)
}

// DeleteAccessPolicy removes a policy, after which it no longer grants any access.
func (h *Handler) DeleteAccessPolicy(c *gateway.Context) error {
	var req requests.AccessPolicyDelete
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if c.Tenant() != nil {
		req.TenantID = c.Tenant().ID
	}

	if err := h.service.DeleteAccessPolicy(c.Ctx(), &req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
