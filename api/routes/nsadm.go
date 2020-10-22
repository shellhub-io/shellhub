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
	ListNamespaceURL   = "/namespace"
	CreateNamespaceURL = "/namespace"
	GetNamespaceURL    = "/namespace/:id"
	DeleteNamespaceURL = "/namespace/:id"
	EditNamespaceURL   = "/namespace/:id"
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

	if _, err := svc.CreateNamespace(c.Ctx(), &namespace); err != nil {
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

	if err := svc.DeleteNamespace(c.Ctx(), c.Param("id")); err != nil {
		if err == nsadm.ErrUnauthorized {
			return c.NoContent(http.StatusForbidden)
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

	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := svc.EditNamespace(c.Ctx(), c.Param("id"), req.Name); err != nil {
		if err == nsadm.ErrUnauthorized {
			return c.NoContent(http.StatusForbidden)
		}

		return err
	}

	return nil
}
