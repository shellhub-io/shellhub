package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
)

// The SSH identity routes, relative to the API's base path.
const (
	ListSSHIdentitiesURL = "/ssh-identities"
	CreateSSHIdentityURL = "/ssh-identities"
	UpdateSSHIdentityURL = "/ssh-identities/:id"
	DeleteSSHIdentityURL = "/ssh-identities/:id"
)

// ListSSHIdentities returns the caller's enrolled SSH identities in the current
// namespace. With ?all=true (and the manage permission) it returns every
// member's, for offboarding.
func (h *Handler) ListSSHIdentities(c *gateway.Context) error {
	req := new(requests.SSHIdentityList)
	if err := c.Bind(req); err != nil {
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

	if req.All && !c.Role().HasPermission(authorizer.SSHIdentityManage) {
		req.All = false
	}

	list, err := h.service.ListSSHIdentities(c.Ctx(), req)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(len(list)))

	return c.JSON(http.StatusOK, list)
}

// CreateSSHIdentity manually enrolls a pasted OpenSSH public key for the caller.
func (h *Handler) CreateSSHIdentity(c *gateway.Context) error {
	req := new(requests.SSHIdentityCreate)
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

	identity, err := h.service.CreateSSHIdentity(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, identity)
}

// UpdateSSHIdentity renames one of the caller's own identities.
func (h *Handler) UpdateSSHIdentity(c *gateway.Context) error {
	req := new(requests.SSHIdentityUpdate)
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

	identity, err := h.service.RenameSSHIdentity(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, identity)
}

// DeleteSSHIdentity revokes an identity. Revoking one's own needs the enroll
// permission; revoking another member's needs the manage permission.
func (h *Handler) DeleteSSHIdentity(c *gateway.Context) error {
	req := new(requests.SSHIdentityDelete)
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

	manage := c.Role().HasPermission(authorizer.SSHIdentityManage)
	if !manage && !c.Role().HasPermission(authorizer.SSHIdentityAdd) {
		return c.NoContent(http.StatusForbidden)
	}

	req.UserID = userID
	req.Manage = manage
	if c.Tenant() != nil {
		req.TenantID = c.Tenant().ID
	}

	if err := h.service.DeleteSSHIdentity(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
