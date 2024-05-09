package routes

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
)

func (h *Handler) listTags() *Route {
	return &Route{
		endpoint:              "/tags",
		method:                MethodGet,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var tenant string
			if t := c.Tenant(); t != nil {
				tenant = t.ID
			}

			tags, count, err := h.service.GetTags(c.Ctx(), tenant)
			if err != nil {
				return err
			}

			c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

			return c.JSON(http.StatusOK, tags)
		},
	}
}

func (h *Handler) updateTag() *Route {
	return &Route{
		endpoint:              "/tags/:tag",
		method:                MethodPut,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var req requests.TagRename

			var tenant string
			if t := c.Tenant(); t != nil {
				tenant = t.ID
			}

			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			err := guard.EvaluatePermission(c.Role(), guard.Actions.Device.RenameTag, func() error {
				return h.service.RenameTag(c.Ctx(), tenant, req.Tag, req.NewTag)
			})
			if err != nil {
				return err
			}

			return c.NoContent(http.StatusOK)
		},
	}
}

func (h *Handler) deleteTag() *Route {
	return &Route{
		endpoint:              "/tags/:tag",
		method:                MethodDelete,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var req requests.TagDelete
			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			var tenant string
			if t := c.Tenant(); t != nil {
				tenant = t.ID
			}

			err := guard.EvaluatePermission(c.Role(), guard.Actions.Device.DeleteTag, func() error {
				return h.service.DeleteTag(c.Ctx(), tenant, req.Tag)
			})
			if err != nil {
				return err
			}

			return c.NoContent(http.StatusOK)
		},
	}
}
