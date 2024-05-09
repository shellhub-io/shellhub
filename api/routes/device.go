package routes

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	ParamDeviceID     = "uid"
	ParamDeviceStatus = "status"
	ParamTagName      = "name"
)

func (h *Handler) getDevice() *Route {
	return &Route{
		endpoint:              "/devices/:uid",
		method:                MethodGet,
		group:                 GroupPublic,
		requiresAuthorization: true,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var req requests.DeviceGet
			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			device, err := h.service.GetDevice(c.Ctx(), models.UID(req.UID))
			if err != nil {
				return err
			}

			return c.JSON(http.StatusOK, device)
		},
	}
}

func (h *Handler) getDeviceByPublicAddress() *Route {
	return &Route{
		endpoint:              "/devices/public/:address",
		method:                MethodGet,
		group:                 GroupInternal,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var req requests.DevicePublicURLAddress
			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			url, err := h.service.GetDeviceByPublicURLAddress(c.Ctx(), req.PublicURLAddress)
			if err != nil {
				return err
			}

			return c.JSON(http.StatusOK, url)
		},
	}
}

func (h *Handler) lookupDevice() *Route {
	return &Route{
		endpoint:              "/lookup",
		method:                MethodGet,
		group:                 GroupInternal,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var req requests.DeviceLookup
			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			device, err := h.service.LookupDevice(c.Ctx(), req.Domain, req.Name)
			if err != nil {
				return err
			}

			return c.JSON(http.StatusOK, device)
		},
	}
}

func (h *Handler) listDevice() *Route {
	return &Route{
		endpoint:              "/devices",
		method:                MethodGet,
		group:                 GroupPublic,
		requiresAuthorization: true,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			type Query struct {
				Status models.DeviceStatus `query:"status"`
				query.Paginator
				query.Sorter
				query.Filters
			}

			query := Query{}

			if err := c.Bind(&query); err != nil {
				return err
			}

			query.Paginator.Normalize()
			query.Sorter.Normalize()

			if err := query.Filters.Unmarshal(); err != nil {
				return err
			}

			var tenant string
			if c.Tenant() != nil {
				tenant = c.Tenant().ID
			}

			devices, count, err := h.service.ListDevices(
				c.Ctx(),
				tenant,
				query.Status,
				query.Paginator,
				query.Filters,
				query.Sorter,
			)
			if err != nil {
				return err
			}

			c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

			return c.JSON(http.StatusOK, devices)
		},
	}
}

func (h *Handler) updateDevice() *Route {
	return &Route{
		endpoint:              "/devices/:uid",
		method:                MethodPut,
		group:                 GroupPublic,
		blockAPIKey:           false,
		requiresAuthorization: false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var req requests.DeviceUpdate
			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			var tenant string
			if c.Tenant() != nil {
				tenant = c.Tenant().ID
			}

			if err := guard.EvaluatePermission(c.Role(), guard.Actions.Device.Update, func() error {
				return h.service.UpdateDevice(c.Ctx(), tenant, models.UID(req.UID), req.Name, req.PublicURL)
			}); err != nil {
				return err
			}

			return c.NoContent(http.StatusOK)
		},
	}
}

func (h *Handler) renameDevice() *Route {
	return &Route{
		endpoint:              "/devices/:uid",
		method:                MethodPatch,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var req requests.DeviceRename
			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			var tenant string
			if c.Tenant() != nil {
				tenant = c.Tenant().ID
			}

			err := guard.EvaluatePermission(c.Role(), guard.Actions.Device.Rename, func() error {
				err := h.service.RenameDevice(c.Ctx(), models.UID(req.UID), req.Name, tenant)

				return err
			})
			if err != nil {
				return err
			}

			return c.NoContent(http.StatusOK)
		},
	}
}

func (h *Handler) offlineDevice() *Route {
	return &Route{
		endpoint:              "/devices/:uid/offline",
		method:                MethodPost,
		group:                 GroupInternal,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var req requests.DeviceOffline
			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			if err := h.service.OfflineDevice(c.Ctx(), models.UID(req.UID)); err != nil {
				return err
			}

			return c.NoContent(http.StatusOK)
		},
	}
}

func (h *Handler) updateDeviceStatus() *Route {
	return &Route{
		endpoint:              "/devices/:uid/:status",
		method:                MethodPatch,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var req requests.DeviceUpdateStatus
			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			var tenant string
			if c.Tenant() != nil {
				tenant = c.Tenant().ID
			}

			status := map[string]models.DeviceStatus{
				"accept":  models.DeviceStatusAccepted,
				"reject":  models.DeviceStatusRejected,
				"pending": models.DeviceStatusPending,
				"unused":  models.DeviceStatusUnused,
			}
			err := guard.EvaluatePermission(c.Role(), guard.Actions.Device.Accept, func() error {
				err := h.service.UpdateDeviceStatus(c.Ctx(), tenant, models.UID(req.UID), status[req.Status])

				return err
			})
			if err != nil {
				return err
			}

			return c.NoContent(http.StatusOK)
		},
	}
}

func (h *Handler) deleteDevice() *Route {
	return &Route{
		endpoint:              "/devices/:uid",
		method:                MethodDelete,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var req requests.DeviceDelete
			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			var tenant string
			if c.Tenant() != nil {
				tenant = c.Tenant().ID
			}

			err := guard.EvaluatePermission(c.Role(), guard.Actions.Device.Remove, func() error {
				err := h.service.DeleteDevice(c.Ctx(), models.UID(req.UID), tenant)

				return err
			})
			if err != nil {
				return err
			}

			return c.NoContent(http.StatusOK)
		},
	}
}

func (h *Handler) createDeviceTag() *Route {
	return &Route{
		endpoint:              "/devices/:uid/tags",
		method:                MethodPost,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var req requests.DeviceCreateTag
			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			err := guard.EvaluatePermission(c.Role(), guard.Actions.Device.CreateTag, func() error {
				return h.service.CreateDeviceTag(c.Ctx(), models.UID(req.UID), req.Tag)
			})
			if err != nil {
				return err
			}

			return c.NoContent(http.StatusOK)
		},
	}
}

func (h *Handler) updateDeviceTag() *Route {
	return &Route{
		endpoint:              "/devices/:uid/tags",
		method:                MethodPut,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var req requests.DeviceUpdateTag
			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			err := guard.EvaluatePermission(c.Role(), guard.Actions.Device.UpdateTag, func() error {
				return h.service.UpdateDeviceTag(c.Ctx(), models.UID(req.UID), req.Tags)
			})
			if err != nil {
				return err
			}

			return c.NoContent(http.StatusOK)
		},
	}
}

func (h *Handler) removeDeviceTag() *Route {
	return &Route{
		endpoint:              "/devices/:uid/tags/:tag",
		method:                MethodDelete,
		group:                 GroupPublic,
		requiresAuthorization: false,
		blockAPIKey:           false,
		middlewares:           []echo.MiddlewareFunc{},
		handler: func(c gateway.Context) error {
			var req requests.DeviceRemoveTag
			if err := c.Bind(&req); err != nil {
				return err
			}

			if err := c.Validate(&req); err != nil {
				return err
			}

			err := guard.EvaluatePermission(c.Role(), guard.Actions.Device.RemoveTag, func() error {
				return h.service.RemoveDeviceTag(c.Ctx(), models.UID(req.UID), req.Tag)
			})
			if err != nil {
				return err
			}

			return c.NoContent(http.StatusOK)
		},
	}
}
