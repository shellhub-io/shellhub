package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	ListNamespaceURL           = "/namespaces"
	CreateNamespaceURL         = "/namespaces"
	GetNamespaceURL            = "/namespaces/:id"
	DeleteNamespaceURL         = "/namespaces/:id"
	EditNamespaceURL           = "/namespaces/:id"
	AddNamespaceUserURL        = "/namespaces/:id/add"
	RemoveNamespaceUserURL     = "/namespaces/:id/del"
	GetSessionRecordURL        = "/users/security"
	EditSessionRecordStatusURL = "/users/security/:id"
)

func (h *handler) GetNamespaceList(c apicontext.Context) error {
	query := filterQuery{}
	if err := c.Bind(&query); err != nil {
		return err
	}

	namespaces, count, err := h.service.ListNamespaces(c.Ctx(), query.Query, query.Filter, false)
	if err != nil {
		return err
	}

	for count, namespace := range namespaces {
		members, err := h.service.ListMembers(c.Ctx(), namespace.TenantID)
		if err != nil {
			return err
		}

		namespaces[count].Members = make([]interface{}, 0)

		for _, member := range members {
			namespaces[count].Members = append(namespaces[count].Members, member)
		}
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, namespaces)
}

func (h *handler) CreateNamespace(c apicontext.Context) error {
	var req models.Namespace
	if err := c.Bind(&req); err != nil {
		return err
	}

	id := ""
	if v := c.ID(); v != nil {
		id = v.ID
	}

	namespace, err := h.service.CreateNamespace(c.Ctx(), &req, id)
	if err != nil {
		switch err {
		case services.ErrUnauthorized:
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

func (h *handler) GetNamespace(c apicontext.Context) error {
	namespace, err := h.service.GetNamespace(c.Ctx(), c.Param("id"))
	if err != nil {
		return err
	}

	members, err := h.service.ListMembers(c.Ctx(), c.Param("id"))
	if err != nil {
		return err
	}

	namespace.Members = make([]interface{}, 0)

	for _, member := range members {
		namespace.Members = append(namespace.Members, member)
	}

	return c.JSON(http.StatusOK, namespace)
}

func (h *handler) DeleteNamespace(c apicontext.Context) error {
	id := ""
	if v := c.ID(); v != nil {
		id = v.ID
	}

	err := h.service.DeleteNamespace(c.Ctx(), c.Param("id"), id)
	if err != nil {
		switch err {
		case services.ErrUnauthorized:
			return c.NoContent(http.StatusForbidden)
		case services.ErrNamespaceNotFound:
			return c.String(http.StatusNotFound, err.Error())
		default:
			return err
		}
	}

	return nil
}

func (h *handler) EditNamespace(c apicontext.Context) error {
	var req struct {
		Name string `json:"name"`
	}

	id := ""
	if v := c.ID(); v != nil {
		id = v.ID
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	namespace, err := h.service.EditNamespace(c.Ctx(), c.Param("id"), req.Name, id)
	if err != nil {
		switch err {
		case services.ErrInvalidFormat:
			return c.NoContent(http.StatusBadRequest)
		case services.ErrUnauthorized:
			return c.NoContent(http.StatusForbidden)
		case services.ErrNamespaceNotFound:
			return c.String(http.StatusNotFound, err.Error())
		default:
			return err
		}
	}

	return c.JSON(http.StatusOK, namespace)
}

func (h *handler) AddNamespaceUser(c apicontext.Context) error {
	var req struct {
		Username string `json:"username"`
	}

	id := ""
	if v := c.ID(); v != nil {
		id = v.ID
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	namespace, err := h.service.AddNamespaceUser(c.Ctx(), c.Param("id"), req.Username, id)
	if err != nil {
		switch err {
		case services.ErrUnauthorized:
			return c.NoContent(http.StatusForbidden)
		case services.ErrUserNotFound:
			return c.String(http.StatusNotFound, err.Error())
		case services.ErrNamespaceNotFound:
			return c.String(http.StatusNotFound, err.Error())
		case services.ErrDuplicateID:
			return c.String(http.StatusConflict, err.Error())
		default:
			return err
		}
	}

	return c.JSON(http.StatusOK, namespace)
}

func (h *handler) RemoveNamespaceUser(c apicontext.Context) error {
	var req struct {
		Username string `json:"username"`
	}

	id := ""
	if v := c.ID(); v != nil {
		id = v.ID
	}

	if err := c.Bind(&req); err != nil {
		return err
	}
	namespace, err := h.service.RemoveNamespaceUser(c.Ctx(), c.Param("id"), req.Username, id)
	if err != nil {
		switch err {
		case services.ErrUnauthorized:
			return c.NoContent(http.StatusForbidden)
		case services.ErrNamespaceNotFound:
			return c.String(http.StatusNotFound, err.Error())
		case services.ErrUserNotFound:
			return c.String(http.StatusNotFound, err.Error())
		case services.ErrDuplicateID:
			return c.String(http.StatusConflict, err.Error())
		default:
			return err
		}
	}

	return c.JSON(http.StatusOK, namespace)
}

func (h *handler) EditSessionRecordStatus(c apicontext.Context) error {
	var req struct {
		SessionRecord bool `json:"session_record"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}

	id := ""
	if v := c.ID(); v != nil {
		id = v.ID
	}

	tenant := c.Param("id")

	err := h.service.EditSessionRecordStatus(c.Ctx(), req.SessionRecord, tenant, id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, nil)
}

func (h *handler) GetSessionRecord(c apicontext.Context) error {
	tenant := ""
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}

	status, err := h.service.GetSessionRecord(c.Ctx(), tenant)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, status)
}
