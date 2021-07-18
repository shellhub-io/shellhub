package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/nsadm"
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
	EditWebhookURL             = "/namespaces/:id/webhook"
	EditWebhookStatusURL       = "/namespaces/:id/webhook/activate"
	GetSessionRecordURL        = "/users/security"
	EditSessionRecordStatusURL = "/users/security/:id"
)

type Invalid struct {
	Field string
	Tag   string
}

func GetNamespaceList(c apicontext.Context) error {
	svc := nsadm.NewService(c.Store())

	query := filterQuery{}
	if err := c.Bind(&query); err != nil {
		return err
	}

	namespaces, count, err := svc.ListNamespaces(c.Ctx(), query.Query, query.Filter, false)
	if err != nil {
		return err
	}

	for count, namespace := range namespaces {
		members, err := svc.ListMembers(c.Ctx(), namespace.TenantID)
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

func CreateNamespace(c apicontext.Context) error {
	svc := nsadm.NewService(c.Store())
	var req models.Namespace
	if err := c.Bind(&req); err != nil {
		return err
	}

	id := ""
	if v := c.ID(); v != nil {
		id = v.ID
	}

	namespace, err := svc.CreateNamespace(c.Ctx(), &req, id)
	if err != nil {
		switch err {
		case nsadm.ErrUnauthorized:
			return c.NoContent(http.StatusForbidden)
		case nsadm.ErrConflictName:
			return c.NoContent(http.StatusConflict)
		case nsadm.ErrInvalidFormat:
			return c.NoContent(http.StatusBadRequest)
		default:
			return err
		}
	}

	return c.JSON(http.StatusOK, namespace)
}

func GetNamespace(c apicontext.Context) error {
	svc := nsadm.NewService(c.Store())

	namespace, err := svc.GetNamespace(c.Ctx(), c.Param("id"))
	if err != nil {
		return err
	}

	members, err := svc.ListMembers(c.Ctx(), c.Param("id"))
	if err != nil {
		return err
	}

	namespace.Members = make([]interface{}, 0)

	for _, member := range members {
		namespace.Members = append(namespace.Members, member)
	}

	return c.JSON(http.StatusOK, namespace)
}

func GetNamespaceByName(c apicontext.Context) error {
	svc := nsadm.NewService(c.Store())

	namespace, err := svc.GetNamespaceByName(c.Ctx(), c.Param("id"))
	if err != nil {
		return err
	}

	members, err := svc.ListMembers(c.Ctx(), namespace.TenantID)
	if err != nil {
		return err
	}

	namespace.Members = make([]interface{}, 0)

	for _, member := range members {
		namespace.Members = append(namespace.Members, member)
	}

	return c.JSON(http.StatusOK, namespace)
}

func DeleteNamespace(c apicontext.Context) error {
	svc := nsadm.NewService(c.Store())

	id := ""
	if v := c.ID(); v != nil {
		id = v.ID
	}

	err := svc.DeleteNamespace(c.Ctx(), c.Param("id"), id)
	if err != nil {
		switch err {
		case nsadm.ErrUnauthorized:
			return c.NoContent(http.StatusForbidden)
		case nsadm.ErrNamespaceNotFound:
			return c.String(http.StatusNotFound, err.Error())
		default:
			return err
		}
	}

	return nil
}

func EditNamespace(c apicontext.Context) error {
	svc := nsadm.NewService(c.Store())

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

	namespace, err := svc.EditNamespace(c.Ctx(), c.Param("id"), req.Name, id)
	if err != nil {
		switch err {
		case nsadm.ErrInvalidFormat:
			return c.NoContent(http.StatusBadRequest)
		case nsadm.ErrUnauthorized:
			return c.NoContent(http.StatusForbidden)
		case nsadm.ErrNamespaceNotFound:
			return c.String(http.StatusNotFound, err.Error())
		default:
			return err
		}
	}

	return c.JSON(http.StatusOK, namespace)
}

func AddNamespaceUser(c apicontext.Context) error {
	svc := nsadm.NewService(c.Store())

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

	namespace, err := svc.AddNamespaceUser(c.Ctx(), c.Param("id"), req.Username, id)
	if err != nil {
		switch err {
		case nsadm.ErrUnauthorized:
			return c.NoContent(http.StatusForbidden)
		case nsadm.ErrUserNotFound:
			return c.String(http.StatusNotFound, err.Error())
		case nsadm.ErrNamespaceNotFound:
			return c.String(http.StatusNotFound, err.Error())
		case nsadm.ErrDuplicateID:
			return c.String(http.StatusConflict, err.Error())
		default:
			return err
		}
	}

	return c.JSON(http.StatusOK, namespace)
}

func RemoveNamespaceUser(c apicontext.Context) error {
	svc := nsadm.NewService(c.Store())

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
	namespace, err := svc.RemoveNamespaceUser(c.Ctx(), c.Param("id"), req.Username, id)
	if err != nil {
		switch err {
		case nsadm.ErrUnauthorized:
			return c.NoContent(http.StatusForbidden)
		case nsadm.ErrNamespaceNotFound:
			return c.String(http.StatusNotFound, err.Error())
		case nsadm.ErrUserNotFound:
			return c.String(http.StatusNotFound, err.Error())
		case nsadm.ErrDuplicateID:
			return c.String(http.StatusConflict, err.Error())
		default:
			return err
		}
	}

	return c.JSON(http.StatusOK, namespace)
}

func UpdateWebhook(c apicontext.Context) error {
	svc := nsadm.NewService(c.Store())

	var req struct {
		URL string `json:"url" bson:"url"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	tenant := c.Param("id")

	id := ""
	if v := c.ID(); v != nil {
		id = v.ID
	}

	invalidFields, wh, err := svc.UpdateWebhook(c.Ctx(), req.URL, tenant, id)
	if err != nil {
		switch err {
		case nsadm.ErrUnauthorized:
			return c.NoContent(http.StatusForbidden)
		case nsadm.ErrUserNotFound:
			return c.String(http.StatusNotFound, err.Error())
		case nsadm.ErrNamespaceNotFound:
			return c.String(http.StatusNotFound, err.Error())
		case nsadm.ErrDuplicateID:
			return c.String(http.StatusConflict, err.Error())
		case nsadm.ErrBadRequest:
			return c.JSON(http.StatusBadRequest, invalidFields)

		default:
			return err
		}
	}

	return c.JSON(http.StatusOK, wh)
}

func SetWebhookStatus(c apicontext.Context) error {
	var req struct {
		Status bool `json:"status"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	id := ""
	if v := c.ID(); v != nil {
		id = v.ID
	}

	tenant := c.Param("id")

	svc := nsadm.NewService(c.Store())

	wh, err := svc.SetWebhookStatus(c.Ctx(), req.Status, tenant, id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, wh)
}

func EditSessionRecordStatus(c apicontext.Context) error {
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

	svc := nsadm.NewService(c.Store())

	err := svc.EditSessionRecordStatus(c.Ctx(), req.SessionRecord, tenant, id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, nil)
}

func GetSessionRecord(c apicontext.Context) error {
	tenant := ""
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}

	svc := nsadm.NewService(c.Store())

	status, err := svc.GetSessionRecord(c.Ctx(), tenant)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, status)
}
