package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/nsadm"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	ListNamespaceURL       = "/namespaces"
	CreateNamespaceURL     = "/namespaces"
	GetNamespaceURL        = "/namespaces/:id"
	DeleteNamespaceURL     = "/namespaces/:id"
	EditNamespaceURL       = "/namespaces/:id"
	AddNamespaceUserURL    = "/namespaces/:id/add"
	RemoveNamespaceUserURL = "/namespaces/:id/del"
	UserSecurityURL        = "/users/security"
	UpdateUserSecurityURL  = "/users/security/:id"
)

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
		namespaces[count].Members, _ = svc.ListMembers(c.Ctx(), namespace.TenantID)
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, namespaces)
}

func CreateNamespace(c apicontext.Context) error {
	svc := nsadm.NewService(c.Store())
	var namespace models.Namespace
	if err := c.Bind(&namespace); err != nil {
		return err
	}

	username := ""
	if v := c.Username(); v != nil {
		username = v.ID
	}
	if _, err := svc.CreateNamespace(c.Ctx(), &namespace, username); err != nil {
		if err == nsadm.ErrUnauthorized {
			return c.NoContent(http.StatusForbidden)
		}
		return err
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
	namespace.Members = members

	return c.JSON(http.StatusOK, namespace)
}

func DeleteNamespace(c apicontext.Context) error {
	svc := nsadm.NewService(c.Store())

	username := ""
	if v := c.Username(); v != nil {
		username = v.ID
	}

	if err := svc.DeleteNamespace(c.Ctx(), c.Param("id"), username); err != nil {
		if err == nsadm.ErrUnauthorized {
			return c.NoContent(http.StatusForbidden)
		}

		if err == nsadm.ErrNamespaceNotFound {
			return c.String(http.StatusNotFound, err.Error())
		}

		return err
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
		if err == nsadm.ErrUnauthorized {
			return c.NoContent(http.StatusForbidden)
		}
		if err == nsadm.ErrNamespaceNotFound {
			return c.String(http.StatusNotFound, err.Error())
		}

		return err
	}

	return c.JSON(http.StatusOK, namespace)
}

func AddNamespaceUser(c apicontext.Context) error {
	svc := nsadm.NewService(c.Store())

	var req struct {
		Username string `json:"username"`
	}

	ownerUsername := ""
	if v := c.Username(); v != nil {
		ownerUsername = v.ID
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	namespace, err := svc.AddNamespaceUser(c.Ctx(), c.Param("id"), req.Username, ownerUsername)
	if err != nil {
		if err == nsadm.ErrUnauthorized {
			return c.NoContent(http.StatusForbidden)
		}
		if err == nsadm.ErrUserNotFound {
			return c.String(http.StatusNotFound, err.Error())
		}
		if err == nsadm.ErrNamespaceNotFound {
			return c.String(http.StatusNotFound, err.Error())
		}

		if err == nsadm.ErrDuplicateID {
			return c.String(http.StatusConflict, err.Error())
		}

		return err
	}

	return c.JSON(http.StatusOK, namespace)
}
func RemoveNamespaceUser(c apicontext.Context) error {
	svc := nsadm.NewService(c.Store())

	var req struct {
		Username string `json:"username"`
	}

	ownerUsername := ""
	if v := c.Username(); v != nil {
		ownerUsername = v.ID
	}

	if err := c.Bind(&req); err != nil {
		return err
	}
	namespace, err := svc.RemoveNamespaceUser(c.Ctx(), c.Param("id"), req.Username, ownerUsername)
	if err != nil {
		if err == nsadm.ErrUnauthorized {
			return c.NoContent(http.StatusForbidden)
		}
		if err == nsadm.ErrUserNotFound {
			return c.String(http.StatusNotFound, err.Error())
		}
		if err == nsadm.ErrNamespaceNotFound {
			return c.String(http.StatusNotFound, err.Error())
		}

		if err == nsadm.ErrDuplicateID {
			return c.String(http.StatusConflict, err.Error())
		}

		return err
	}

	return c.JSON(http.StatusOK, namespace)
}

func UpdateUserSecurity(c apicontext.Context) error {
	var req struct {
		SessionRecord bool `json:"session_record"`
	}
	if err := c.Bind(&req); err != nil {
		return err
	}

	tenant := c.Param("id")

	svc := nsadm.NewService(c.Store())

	err := svc.UpdateDataUserSecurity(c.Ctx(), req.SessionRecord, tenant)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, nil)
}

func GetUserSecurity(c apicontext.Context) error {
	tenant := ""
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}

	svc := nsadm.NewService(c.Store())

	status, err := svc.GetDataUserSecurity(c.Ctx(), tenant)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, status)
}
