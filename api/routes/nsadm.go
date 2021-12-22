package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/apicontext"
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

func (h *Handler) GetNamespaceList(c apicontext.Context) error {
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

func (h *Handler) CreateNamespace(c apicontext.Context) error {
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

func (h *Handler) GetNamespace(c apicontext.Context) error {
	var userID string
	if c.ID() != nil {
		userID = c.ID().ID
	}

	namespace, err := h.service.GetNamespace(c.Ctx(), c.Param(ParamNamespaceTenant))
	if err != nil || namespace == nil {
		return c.NoContent(http.StatusNotFound)
	}

	_, ok := guard.CheckMember(namespace, userID)
	if !ok {
		return c.NoContent(http.StatusForbidden)
	}

	return c.JSON(http.StatusOK, namespace)
}

func (h *Handler) DeleteNamespace(c apicontext.Context) error {
	var userID string
	if c.ID() != nil {
		userID = c.ID().ID
	}

	namespaceToMember, err := h.service.GetNamespace(c.Ctx(), c.Param(ParamNamespaceTenant))
	if err != nil || namespaceToMember == nil {
		return c.NoContent(http.StatusNotFound)
	}

	memberFromNamespace, ok := guard.CheckMember(namespaceToMember, userID)
	if !ok {
		return c.NoContent(http.StatusForbidden)
	}

	err = guard.EvaluatePermission(memberFromNamespace.Role, authorizer.Actions.Namespace.Delete, func() error {
		err := h.service.DeleteNamespace(c.Ctx(), namespaceToMember.TenantID)

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

func (h *Handler) EditNamespace(c apicontext.Context) error {
	var req struct {
		Name string `json:"name"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	var userID string
	if c.ID() != nil {
		userID = c.ID().ID
	}

	namespaceToMember, err := h.service.GetNamespace(c.Ctx(), c.Param(ParamNamespaceTenant))
	if err != nil || namespaceToMember == nil {
		return c.NoContent(http.StatusNotFound)
	}

	memberFromNamespace, ok := guard.CheckMember(namespaceToMember, userID)
	if !ok {
		return c.NoContent(http.StatusForbidden)
	}

	var namespace *models.Namespace
	err = guard.EvaluatePermission(memberFromNamespace.Role, authorizer.Actions.Namespace.Rename, func() error {
		var err error
		namespace, err = h.service.EditNamespace(c.Ctx(), namespaceToMember.TenantID, req.Name)

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

	return c.JSON(http.StatusOK, namespace)
}

func (h *Handler) AddNamespaceUser(c apicontext.Context) error {
	var member struct {
		Username string `json:"username"`
		Role     string `json:"role"`
	}

	userID := ""
	if c.ID() != nil {
		userID = c.ID().ID
	}

	if err := c.Bind(&member); err != nil {
		return err
	}

	namespaceToMember, err := h.service.GetNamespace(c.Ctx(), c.Param(ParamNamespaceTenant))
	if err != nil || namespaceToMember == nil {
		return c.NoContent(http.StatusNotFound)
	}

	memberFromNamespace, ok := guard.CheckMember(namespaceToMember, userID)
	if !ok {
		return c.NoContent(http.StatusForbidden)
	}

	var namespace *models.Namespace
	err = guard.EvaluatePermission(memberFromNamespace.Role, authorizer.Actions.Namespace.AddMember, func() error {
		var err error
		namespace, err = h.service.AddNamespaceUser(c.Ctx(), member.Username, member.Role, namespaceToMember.TenantID, userID)

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

func (h *Handler) RemoveNamespaceUser(c apicontext.Context) error {
	userID := ""
	if v := c.ID(); v != nil {
		userID = c.ID().ID
	}

	namespaceToMember, err := h.service.GetNamespace(c.Ctx(), c.Param(ParamNamespaceTenant))
	if err != nil || namespaceToMember == nil {
		return c.NoContent(http.StatusNotFound)
	}

	memberFromNamespace, ok := guard.CheckMember(namespaceToMember, userID)
	if !ok {
		return c.NoContent(http.StatusForbidden)
	}

	var namespace *models.Namespace
	err = guard.EvaluatePermission(memberFromNamespace.Role, authorizer.Actions.Namespace.RemoveMember, func() error {
		var err error
		namespace, err = h.service.RemoveNamespaceUser(c.Ctx(), namespaceToMember.TenantID, c.Param(ParamNamespaceMemberID), userID)

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

	return c.JSON(http.StatusOK, namespace)
}

func (h *Handler) EditNamespaceUser(c apicontext.Context) error {
	var member struct {
		Role string `json:"role"`
	}

	if err := c.Bind(&member); err != nil {
		return err
	}

	userID := ""
	if c.ID() != nil {
		userID = c.ID().ID
	}

	namespaceToMember, err := h.service.GetNamespace(c.Ctx(), c.Param(ParamNamespaceTenant))
	if err != nil || namespaceToMember == nil {
		return c.NoContent(http.StatusNotFound)
	}

	memberFromNamespace, ok := guard.CheckMember(namespaceToMember, userID)
	if !ok {
		return c.NoContent(http.StatusForbidden)
	}

	err = guard.EvaluatePermission(memberFromNamespace.Role, authorizer.Actions.Namespace.EditMember, func() error {
		err := h.service.EditNamespaceUser(c.Ctx(), namespaceToMember.TenantID, userID, c.Param(ParamNamespaceMemberID), member.Role)

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

func (h *Handler) EditSessionRecordStatus(c apicontext.Context) error {
	var req struct {
		SessionRecord bool `json:"session_record"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}

	userID := ""
	if c.ID() != nil {
		userID = c.ID().ID
	}

	namespaceToMember, err := h.service.GetNamespace(c.Ctx(), c.Param(ParamNamespaceTenant))
	if err != nil || namespaceToMember == nil {
		return c.NoContent(http.StatusNotFound)
	}

	memberFromNamespace, ok := guard.CheckMember(namespaceToMember, userID)
	if !ok {
		return c.NoContent(http.StatusForbidden)
	}

	err = guard.EvaluatePermission(memberFromNamespace.Role, authorizer.Actions.Namespace.EnableSessionRecord, func() error {
		err := h.service.EditSessionRecordStatus(c.Ctx(), req.SessionRecord, namespaceToMember.TenantID)

		return err
	})
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) GetSessionRecord(c apicontext.Context) error {
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
