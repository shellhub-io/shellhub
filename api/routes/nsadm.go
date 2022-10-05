package routes

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/pkg/api/request"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	ListNamespaceURL           = "/namespaces"
	CreateNamespaceURL         = "/namespaces"
	GetNamespaceURL            = "/namespaces/:tenant"
	DeleteNamespaceURL         = "/namespaces/:tenant"
	EditNamespaceURL           = "/namespaces/:tenant"
	AddNamespaceUserURL        = "/namespaces/:tenant/members"
	RemoveNamespaceUserURL     = "/namespaces/:tenant/members/:uid"
	EditNamespaceUserURL       = "/namespaces/:tenant/members/:uid"
	GetSessionRecordURL        = "/users/security"
	EditSessionRecordStatusURL = "/users/security/:tenant"
)

const (
	ParamNamespaceTenant   = "tenant"
	ParamNamespaceMemberID = "uid"
)

func (h *Handler) GetNamespaceList(c gateway.Context) error {
	query := filterQuery{}
	if err := c.Bind(&query); err != nil {
		return err
	}

	raw, err := base64.StdEncoding.DecodeString(query.Filter)
	if err != nil {
		return err
	}

	var filter []models.Filter
	if err := json.Unmarshal(raw, &filter); len(raw) > 0 && err != nil {
		return err
	}

	namespaces, count, err := h.service.ListNamespaces(c.Ctx(), query.Query, filter, false)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, namespaces)
}

func (h *Handler) CreateNamespace(c gateway.Context) error {
	var req request.NamespaceCreate
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
	var req request.NamespaceGet
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
		_, ok := guard.CheckMember(ns, uid)
		if !ok {
			return c.NoContent(http.StatusForbidden)
		}
	}

	return c.JSON(http.StatusOK, ns)
}

func (h *Handler) DeleteNamespace(c gateway.Context) error {
	var req request.NamespaceDelete
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

	err = guard.EvaluateNamespace(ns, uid, guard.Actions.Namespace.Delete, func() error {
		err := h.service.DeleteNamespace(c.Ctx(), ns.TenantID)

		return err
	})
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) EditNamespace(c gateway.Context) error {
	var req request.NamespaceEdit
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

	namespace, err := h.service.GetNamespace(c.Ctx(), req.Tenant)
	if err != nil || namespace == nil {
		return c.NoContent(http.StatusNotFound)
	}

	var nns *models.Namespace
	err = guard.EvaluateNamespace(namespace, uid, guard.Actions.Namespace.Rename, func() error {
		var err error
		nns, err = h.service.EditNamespace(c.Ctx(), namespace.TenantID, req.Name)

		return err
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, nns)
}

func (h *Handler) AddNamespaceUser(c gateway.Context) error {
	var req request.NamespaceAddUser
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

	var namespace *models.Namespace
	err = guard.EvaluateNamespace(ns, uid, guard.Actions.Namespace.AddMember, func() error {
		var err error
		namespace, err = h.service.AddNamespaceUser(c.Ctx(), req.Username, req.Role, ns.TenantID, uid)

		return err
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, namespace)
}

func (h *Handler) RemoveNamespaceUser(c gateway.Context) error {
	var req request.NamespaceRemoveUser
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	var uid string
	if v := c.ID(); v != nil {
		uid = c.ID().ID
	}

	ns, err := h.service.GetNamespace(c.Ctx(), req.Tenant)
	if err != nil || ns == nil {
		return c.NoContent(http.StatusNotFound)
	}

	var nns *models.Namespace
	err = guard.EvaluateNamespace(ns, uid, guard.Actions.Namespace.RemoveMember, func() error {
		var err error
		nns, err = h.service.RemoveNamespaceUser(c.Ctx(), ns.TenantID, req.MemberUID, uid)

		return err
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, nns)
}

func (h *Handler) EditNamespaceUser(c gateway.Context) error {
	var req request.NamespaceEditUser
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

	err = guard.EvaluateNamespace(ns, uid, guard.Actions.Namespace.EditMember, func() error {
		err := h.service.EditNamespaceUser(c.Ctx(), ns.TenantID, uid, req.MemberUID, req.Role)

		return err
	})
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) EditSessionRecordStatus(c gateway.Context) error {
	var req request.SessionEditRecordStatus
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

	err = guard.EvaluateNamespace(ns, uid, guard.Actions.Namespace.EnableSessionRecord, func() error {
		err := h.service.EditSessionRecordStatus(c.Ctx(), req.SessionRecord, ns.TenantID)

		return err
	})
	if err != nil {
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
