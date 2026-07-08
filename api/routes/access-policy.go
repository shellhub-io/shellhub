package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	ListAccessPoliciesURL = "/access-policies"
	GetAccessPolicyURL    = "/access-policies/:id"
	CreateAccessPolicyURL = "/access-policies"
	UpdateAccessPolicyURL = "/access-policies/:id"
	DeleteAccessPolicyURL = "/access-policies/:id"

	AuthorizeSSHAccessURL = "/access-policies/authorize"
	HasAccessPoliciesURL  = "/access-policies/exists"
)

// HasAccessPolicies is the internal endpoint the SSH gateway calls before minting
// an approval, to short-circuit an identity-mode login when the namespace has no
// policies at all (every login would be default-denied, so there is no point
// asking the user to approve).
func (h *Handler) HasAccessPolicies(c gateway.Context) error {
	exists, err := h.service.NamespaceHasAccessPolicies(c.Ctx(), c.QueryParam("tenant"))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]bool{"exists": exists})
}

// AuthorizeSSHAccess is the internal endpoint the SSH gateway calls at the
// ephemeral-key mint point to decide whether an approved identity may reach a
// device as a given login under the namespace's Access Policies.
func (h *Handler) AuthorizeSSHAccess(c gateway.Context) error {
	tenant := c.QueryParam("tenant")
	userID := c.QueryParam("user_id")
	device := c.QueryParam("device")
	login := c.QueryParam("login")

	decision, err := h.service.Authorize(c.Ctx(), tenant, userID, &models.Device{UID: device}, login)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, decision)
}

func (h *Handler) ListAccessPolicies(c gateway.Context) error {
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

func (h *Handler) GetAccessPolicy(c gateway.Context) error {
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

func (h *Handler) CreateAccessPolicy(c gateway.Context) error {
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

func (h *Handler) UpdateAccessPolicy(c gateway.Context) error {
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

func (h *Handler) DeleteAccessPolicy(c gateway.Context) error {
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
