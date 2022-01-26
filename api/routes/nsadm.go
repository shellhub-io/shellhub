package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/authorizer"
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

	namespaces, count, err := h.service.ListNamespaces(c.Ctx(), query.Query, query.Filter, false)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, namespaces)
}

func (h *Handler) CreateNamespace(c gateway.Context) error {
	var req models.Namespace
	if err := c.Bind(&req); err != nil {
		return err
	}

	userID := ""
	if v := c.ID(); v != nil {
		userID = v.ID
	}

	namespace, err := h.service.CreateNamespace(c.Ctx(), &req, userID)
	if err != nil {
		switch err {
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
		case services.ErrConflictName:
			return c.NoContent(http.StatusConflict)
		case services.ErrInvalidFormat:
			return c.NoContent(http.StatusBadRequest)
		default:
			return err
		}
	}

	return c.JSON(http.StatusOK, namespace)
}

func (h *Handler) GetNamespace(c gateway.Context) error {
	var uid string
	if c.ID() != nil {
		uid = c.ID().ID
	}

	ns, err := h.service.GetNamespace(c.Ctx(), c.Param(ParamNamespaceTenant))
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
	var uid string
	if c.ID() != nil {
		uid = c.ID().ID
	}

	ns, err := h.service.GetNamespace(c.Ctx(), c.Param(ParamNamespaceTenant))
	if err != nil || ns == nil {
		return c.NoContent(http.StatusNotFound)
	}

	err = guard.EvaluateNamespace(ns, uid, authorizer.Actions.Namespace.Delete, func() error {
		err := h.service.DeleteNamespace(c.Ctx(), ns.TenantID)

		return err
	})
	if err != nil {
		switch err {
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
		case services.ErrNamespaceNotFound:
			return c.NoContent(http.StatusNotFound)
		default:
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) EditNamespace(c gateway.Context) error {
	var req struct {
		Name string `json:"name"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	var uid string
	if c.ID() != nil {
		uid = c.ID().ID
	}

	ns, err := h.service.GetNamespace(c.Ctx(), c.Param(ParamNamespaceTenant))
	if err != nil || ns == nil {
		return c.NoContent(http.StatusNotFound)
	}

	var nns *models.Namespace
	err = guard.EvaluateNamespace(ns, uid, authorizer.Actions.Namespace.Rename, func() error {
		var err error
		nns, err = h.service.EditNamespace(c.Ctx(), ns.TenantID, req.Name)

		return err
	})
	if err != nil {
		switch err {
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
		case services.ErrInvalidFormat:
			return c.NoContent(http.StatusBadRequest)
		case services.ErrNamespaceNotFound:
			return c.NoContent(http.StatusNotFound)
		default:
			return err
		}
	}

	return c.JSON(http.StatusOK, nns)
}

func (h *Handler) AddNamespaceUser(c gateway.Context) error {
	var member struct {
		Username string `json:"username"`
		Role     string `json:"role"`
	}

	var uid string
	if c.ID() != nil {
		uid = c.ID().ID
	}

	if err := c.Bind(&member); err != nil {
		return err
	}

	ns, err := h.service.GetNamespace(c.Ctx(), c.Param(ParamNamespaceTenant))
	if err != nil || ns == nil {
		return c.NoContent(http.StatusNotFound)
	}

	var namespace *models.Namespace
	err = guard.EvaluateNamespace(ns, uid, authorizer.Actions.Namespace.AddMember, func() error {
		var err error
		namespace, err = h.service.AddNamespaceUser(c.Ctx(), member.Username, member.Role, ns.TenantID, uid)

		return err
	})
	if err != nil {
		switch err {
		case services.ErrInvalidFormat:
			return c.NoContent(http.StatusBadRequest)
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
		case services.ErrUserNotFound:
			return c.NoContent(http.StatusNotFound)
		case services.ErrNamespaceNotFound:
			return c.NoContent(http.StatusNotFound)
		case services.ErrNamespaceDuplicatedMember:
			return c.NoContent(http.StatusConflict)
		default:
			return err
		}
	}

	return c.JSON(http.StatusOK, namespace)
}

func (h *Handler) RemoveNamespaceUser(c gateway.Context) error {
	var uid string
	if v := c.ID(); v != nil {
		uid = c.ID().ID
	}

	ns, err := h.service.GetNamespace(c.Ctx(), c.Param(ParamNamespaceTenant))
	if err != nil || ns == nil {
		return c.NoContent(http.StatusNotFound)
	}

	var nns *models.Namespace
	err = guard.EvaluateNamespace(ns, uid, authorizer.Actions.Namespace.RemoveMember, func() error {
		var err error
		nns, err = h.service.RemoveNamespaceUser(c.Ctx(), ns.TenantID, c.Param(ParamNamespaceMemberID), uid)

		return err
	})
	if err != nil {
		switch err {
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
		case services.ErrUserNotFound:
			return c.NoContent(http.StatusNotFound)
		case services.ErrNamespaceNotFound:
			return c.NoContent(http.StatusNotFound)
		case services.ErrNamespaceMemberNotFound:
			return c.NoContent(http.StatusNotFound)
		default:
			return err
		}
	}

	return c.JSON(http.StatusOK, nns)
}

func (h *Handler) EditNamespaceUser(c gateway.Context) error {
	var member struct {
		Role string `json:"role"`
	}

	if err := c.Bind(&member); err != nil {
		return err
	}

	var uid string
	if c.ID() != nil {
		uid = c.ID().ID
	}

	ns, err := h.service.GetNamespace(c.Ctx(), c.Param(ParamNamespaceTenant))
	if err != nil || ns == nil {
		return c.NoContent(http.StatusNotFound)
	}

	err = guard.EvaluateNamespace(ns, uid, authorizer.Actions.Namespace.EditMember, func() error {
		err := h.service.EditNamespaceUser(c.Ctx(), ns.TenantID, uid, c.Param(ParamNamespaceMemberID), member.Role)

		return err
	})
	if err != nil {
		switch err {
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
		case services.ErrUserNotFound:
			return c.NoContent(http.StatusNotFound)
		case services.ErrNamespaceNotFound:
			return c.NoContent(http.StatusNotFound)
		case services.ErrNamespaceMemberNotFound:
			return c.NoContent(http.StatusNotFound)
		default:
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) EditSessionRecordStatus(c gateway.Context) error {
	var req struct {
		SessionRecord bool `json:"session_record"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}

	var uid string
	if c.ID() != nil {
		uid = c.ID().ID
	}

	ns, err := h.service.GetNamespace(c.Ctx(), c.Param(ParamNamespaceTenant))
	if err != nil || ns == nil {
		return c.NoContent(http.StatusNotFound)
	}

	err = guard.EvaluateNamespace(ns, uid, authorizer.Actions.Namespace.EnableSessionRecord, func() error {
		err := h.service.EditSessionRecordStatus(c.Ctx(), req.SessionRecord, ns.TenantID)

		return err
	})
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) GetSessionRecord(c gateway.Context) error {
	tenantID := ""
	if v := c.Tenant(); v != nil {
		tenantID = v.ID
	}

	status, err := h.service.GetSessionRecord(c.Ctx(), tenantID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, status)
}
