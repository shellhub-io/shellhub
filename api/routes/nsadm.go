package routes

import (
	"net/http"
	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/nsadm"
	"github.com/shellhub-io/shellhub/pkg/models"
	"gopkg.in/go-playground/validator.v9"
)

const (
	ListNamespaceURL       = "/namespaces"
	CreateNamespaceURL     = "/namespaces"
	GetNamespaceURL        = "/namespaces/:id"
	DeleteNamespaceURL     = "/namespaces/:id"
	EditWebhookURL         = "/namespaces/:id/webhook"
	EditWebhookStatusURL   = "/namespaces/:id/webhook/activate"
	EditNamespaceURL       = "/namespaces/:id"
	AddNamespaceUserURL    = "/namespaces/:id/add"
	RemoveNamespaceUserURL = "/namespaces/:id/del"
	UserSecurityURL        = "/users/security"
	UpdateUserSecurityURL  = "/users/security/:id"
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
	var namespace models.Namespace
	if err := c.Bind(&namespace); err != nil {
		return err
	}

	id := ""
	if v := c.ID(); v != nil {
		id = v.ID
	}

	if _, err := svc.CreateNamespace(c.Ctx(), &namespace, id); err != nil {
		switch {
		case err == nsadm.ErrUnauthorized:
			return c.NoContent(http.StatusForbidden)
		case err == nsadm.ErrConflictName:
			return c.NoContent(http.StatusConflict)
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

	if err := svc.DeleteNamespace(c.Ctx(), c.Param("id"), id); err != nil {
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

	id := ""
	if v := c.ID(); v != nil {
		id = v.ID
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	namespace, err := svc.AddNamespaceUser(c.Ctx(), c.Param("id"), req.Username, id)
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

	id := ""
	if v := c.ID(); v != nil {
		id = v.ID
	}

	if err := c.Bind(&req); err != nil {
		return err
	}
	namespace, err := svc.RemoveNamespaceUser(c.Ctx(), c.Param("id"), req.Username, id)
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
func UpdateWebhook(c apicontext.Context) error {
	svc := nsadm.NewService(c.Store())

	var req struct {
		URL string `json:"url" bson:"url" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	tenant := c.Param("id")

	id := ""
	if v := c.ID(); v != nil {
		id = v.ID
	}

	validURL := govalidator.IsURL(req.URL)

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		var I []Invalid
		for _, verr := range err.(validator.ValidationErrors) {
			var invalid Invalid
			invalid.Tag = verr.Tag()
			invalid.Field = verr.Field()
			I = append(I, invalid)
		}

		return c.JSON(http.StatusBadRequest, I)
	}

	if !validURL {
		var I []Invalid
		var invalid Invalid
		invalid.Tag = "invalid url"
		invalid.Field = "URL"
		I = append(I, invalid)

		return c.JSON(http.StatusBadRequest, I)
	}

	wh, err := svc.UpdateWebhook(c.Ctx(), req.URL, tenant, id)
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
