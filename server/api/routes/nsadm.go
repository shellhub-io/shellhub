package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/server/api/services"
	log "github.com/sirupsen/logrus"
)

// The namespace and membership routes, relative to the API's base path.
const (
	ListNamespaceURL           = "/namespaces"
	CreateNamespaceURL         = "/namespaces"
	GetNamespaceURL            = "/namespaces/:tenant"
	DeleteNamespaceURL         = "/namespaces/:tenant"
	EditNamespaceURL           = "/namespaces/:tenant"
	ListNamespaceMembersURL    = "/namespaces/:tenant/members"
	LeaveNamespaceURL          = "/namespaces/:tenant/members"
	AddNamespaceMemberURL      = "/namespaces/:tenant/members"
	RemoveNamespaceMemberURL   = "/namespaces/:tenant/members/:uid"
	EditNamespaceMemberURL     = "/namespaces/:tenant/members/:uid"
	EditSessionRecordStatusURL = "/users/security/:tenant"
	EditSSHAccessModeURL       = "/namespaces/ssh-access-mode/:tenant"
)

// The path parameter names these routes bind by.
const (
	ParamNamespaceTenant   = "tenant"
	ParamNamespaceMemberID = "uid"
)

// GetNamespaceList serves the namespaces the caller belongs to.
func (h *Handler) GetNamespaceList(c *gateway.Context) error {
	req := new(requests.NamespaceList)

	if err := c.Bind(req); err != nil {
		return err
	}

	req.Normalize()

	if err := req.Unmarshal(); err != nil {
		log.WithError(err).WithField("filter", req.Filters.Raw).Warn("failed to decode namespace list filter")

		return c.NoContent(http.StatusBadRequest)
	}

	if err := query.ValidateFilters(&req.Filters, services.NamespaceFilterFields); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	namespaces, count, err := h.service.ListNamespaces(c.Ctx(), req)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, namespaces)
}

// CreateNamespace creates a namespace owned by the caller.
func (h *Handler) CreateNamespace(c *gateway.Context) error {
	req := new(requests.NamespaceCreate)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	namespace, err := h.service.CreateNamespace(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, namespace)
}

// GetNamespace serves one namespace by tenant ID.
func (h *Handler) GetNamespace(c *gateway.Context) error {
	var req requests.NamespaceGet
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	var uid string
	if c.ID() != nil {
		uid = c.ID().ID
	}

	ns, err := h.service.GetNamespace(c.Ctx(), req.Tenant)
	if err != nil || ns == nil {
		return c.NoContent(http.StatusNotFound)
	}

	if uid != "" {
		if _, ok := ns.FindMember(uid); !ok {
			return c.NoContent(http.StatusForbidden)
		}
	}

	return c.JSON(http.StatusOK, ns)
}

// ListNamespaceMembers serves who belongs to a namespace and in what role.
func (h *Handler) ListNamespaceMembers(c *gateway.Context) error {
	req := new(requests.MemberList)

	if err := c.Bind(req); err != nil {
		return err
	}

	req.Paginator.Normalize()

	if err := c.Validate(req); err != nil {
		return err
	}

	members, count, err := h.service.ListNamespaceMembers(c.Ctx(), req)
	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, members)
}

// DeleteNamespace removes a namespace and everything scoped to it.
func (h *Handler) DeleteNamespace(c *gateway.Context) error {
	var req requests.NamespaceDelete
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.service.DeleteNamespace(c.Ctx(), req.Tenant); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// EditNamespace changes a namespace's own attributes, not its membership.
func (h *Handler) EditNamespace(c *gateway.Context) error {
	req := new(requests.NamespaceEdit)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	res, err := h.service.EditNamespace(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// AddNamespaceMember invites or adds a member in the requested role.
func (h *Handler) AddNamespaceMember(c *gateway.Context) error {
	req := new(requests.NamespaceAddMember)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	res, err := h.service.AddNamespaceMember(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// RemoveNamespaceMember removes another member from the namespace.
func (h *Handler) RemoveNamespaceMember(c *gateway.Context) error {
	req := new(requests.NamespaceRemoveMember)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	res, err := h.service.RemoveNamespaceMember(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// LeaveNamespace removes the caller from the namespace. It is separate from removing a member
// because leaving needs no permission over others.
func (h *Handler) LeaveNamespace(c *gateway.Context) error {
	req := new(requests.LeaveNamespace)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	res, err := h.service.LeaveNamespace(c.Ctx(), req)
	switch {
	case err != nil:
		return err
	case res != nil:
		return c.JSON(http.StatusOK, res)
	default:
		return c.NoContent(http.StatusOK)
	}
}

// EditNamespaceMember changes a member's role.
func (h *Handler) EditNamespaceMember(c *gateway.Context) error {
	req := new(requests.NamespaceUpdateMember)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.service.UpdateNamespaceMember(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// EditSSHAccessMode switches the namespace between key-based and identity-based SSH access.
func (h *Handler) EditSSHAccessMode(c *gateway.Context) error {
	var req requests.EditSSHAccessMode
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.service.EditSSHAccessMode(c.Ctx(), req.SSHAccessMode, req.Tenant); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// EditSessionRecordStatus turns session recording on or off for the namespace.
func (h *Handler) EditSessionRecordStatus(c *gateway.Context) error {
	var req requests.SessionEditRecordStatus
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.service.EditSessionRecordStatus(c.Ctx(), req.SessionRecord, req.Tenant); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
