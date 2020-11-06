package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/nsadm"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	ListNamespaceURL       = "/namespace"
	CreateNamespaceURL     = "/namespace"
	GetNamespaceURL        = "/namespace/:id"
	DeleteNamespaceURL     = "/namespace/:id"
	EditNamespaceURL       = "/namespace/:id"
	AddNamespaceUserURL    = "/namespace/:id/add"
	RemoveNamespaceUserURL = "/namespace/:id/del"
)

func GetNamespaceList(c apicontext.Context) error {
	svc := nsadm.NewService(c.Store())
	query := paginator.NewQuery()
	c.Bind(query)

	query.Normalize()
	namespaces, count, err := svc.ListNamespaces(c.Ctx(), *query)
	if err != nil {
		return err
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

	username := ""
	if v := c.Username(); v != nil {
		username = v.ID
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	namespace, err := svc.EditNamespace(c.Ctx(), c.Param("id"), req.Name, username)
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

		return err
	}

	return c.JSON(http.StatusOK, namespace)
}
