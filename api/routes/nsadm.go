package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
)

const (
	ListNamespaceURL           = "/namespaces"
	CreateNamespaceURL         = "/namespaces"
	GetNamespaceURL            = "/namespaces/:tenant"
	DeleteNamespaceURL         = "/namespaces/:tenant"
	EditNamespaceURL           = "/namespaces/:tenant"
	AddNamespaceMemberURL      = "/namespaces/:tenant/members"
	RemoveNamespaceMemberURL   = "/namespaces/:tenant/members/:uid"
	EditNamespaceMemberURL     = "/namespaces/:tenant/members/:uid"
	GetSessionRecordURL        = "/users/security"
	EditSessionRecordStatusURL = "/users/security/:tenant"
)

const (
	ParamNamespaceTenant   = "tenant"
	ParamNamespaceMemberID = "uid"
)

func (h *Handler) GetNamespaceList(c gateway.Context) error {
	req := new(requests.NamespaceList)

	if err := c.Bind(req); err != nil {
		return err
	}

	req.Paginator.Normalize()
	if err := req.Filters.Unmarshal(); err != nil {
		return err
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

func (h *Handler) CreateNamespace(c gateway.Context) error {
	var req requests.NamespaceCreate
	if err := c.Bind(&req); err != nil {
		return err
	}

	var userID string
	if v := c.ID(); v != nil {
		userID = v.ID
	}

	namespace, err := h.service.CreateNamespace(c.Ctx(), req, userID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, namespace)
}

func (h *Handler) GetNamespace(c gateway.Context) error {
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

func (h *Handler) DeleteNamespace(c gateway.Context) error {
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

func (h *Handler) EditNamespace(c gateway.Context) error {
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

func (h *Handler) AddNamespaceMember(c gateway.Context) error {
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

func (h *Handler) RemoveNamespaceMember(c gateway.Context) error {
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

func (h *Handler) EditNamespaceMember(c gateway.Context) error {
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

func (h *Handler) EditSessionRecordStatus(c gateway.Context) error {
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

func (h *Handler) GetSessionRecord(c gateway.Context) error {
	var tenant string
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}

	status, err := h.service.GetSessionRecord(c.Ctx(), tenant)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, status)
}
