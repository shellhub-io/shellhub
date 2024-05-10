package routes

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
)

func (h *Handler) createAPIKey() *Route {
	return &Route{
		endpoint:              "/namespaces/:tenant/api-key",
		method:                MethodPost,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			return c.NoContent(200)

			var req requests.CreateAPIKey

			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			var userID string
			if v := c.ID(); v != nil {
				userID = v.ID
			}

			key := c.Request().Header.Get("X-API-KEY")
			tenant := c.Request().Header.Get("X-Tenant-ID")

			var uid string

			if err := guard.EvaluatePermission(c.Role(), guard.Actions.APIKey.Create, func() error {
				var err error

				uid, err = h.service.CreateAPIKey(c.Ctx(), userID, tenant, key, c.Role(), &req)

				return err
			}); err != nil {
				return err
			}

			return c.JSON(http.StatusOK, uid)
		},
	}
}

func (h *Handler) listAPIKey() *Route {
	return &Route{
		endpoint:              "/namespaces/:tenant/api-key",
		method:                MethodGet,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			return c.NoContent(200)

			var req requests.APIKeyList

			if err := c.Bind(&req); err != nil {
				return err
			}

			req.Paginator.Normalize()
			req.Sorter.Normalize()

			if req.Sorter.By == "" {
				req.Sorter.By = "expires_in"
				req.Sorter.Order = "desc"
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			req.TenantParam.Tenant = c.Request().Header.Get("X-Tenant-ID")

			res, count, err := h.service.ListAPIKeys(c.Ctx(), &req)
			if err != nil {
				return err
			}

			c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

			return c.JSON(http.StatusOK, res)
		},
	}
}

func (h *Handler) updateAPIKey() *Route {
	return &Route{
		endpoint:              "/namespaces/:tenant/api-key/:key",
		method:                MethodPatch,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			return c.NoContent(200)

			var req requests.APIKeyChanges

			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			tenant := c.Request().Header.Get("X-Tenant-ID")

			var key *models.APIKey

			if err := guard.EvaluatePermission(c.Role(), guard.Actions.APIKey.Edit, func() error {
				var err error

				key, err = h.service.EditAPIKey(c.Ctx(), tenant, &req)

				return err
			}); err != nil {
				return err
			}

			return c.JSON(http.StatusOK, key)
		},
	}
}

func (h *Handler) deleteAPIKey() *Route {
	return &Route{
		endpoint:              "/namespaces/:tenant/api-key/:key",
		method:                MethodDelete,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			return c.NoContent(200)

			var req requests.APIKeyID

			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			tenant := c.Request().Header.Get("X-Tenant-ID")

			if err := guard.EvaluatePermission(c.Role(), guard.Actions.APIKey.Delete, func() error {
				return h.service.DeleteAPIKey(c.Ctx(), req.ID, tenant)
			}); err != nil {
				return err
			}

			return c.JSON(http.StatusOK, nil)
		},
	}
}
