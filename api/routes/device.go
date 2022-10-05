package routes

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/api/request"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	GetDeviceListURL   = "/devices"
	GetDeviceURL       = "/devices/:uid"
	DeleteDeviceURL    = "/devices/:uid"
	RenameDeviceURL    = "/devices/:uid"
	OfflineDeviceURL   = "/devices/:uid/offline"
	HeartbeatDeviceURL = "/devices/:uid/heartbeat"
	LookupDeviceURL    = "/lookup"
	UpdateStatusURL    = "/devices/:uid/:status"
	CreateTagURL       = "/devices/:uid/tags"      // Add a tag to a device.
	UpdateTagURL       = "/devices/:uid/tags"      // Update device's tags with a new set.
	RemoveTagURL       = "/devices/:uid/tags/:tag" // Delete a tag from a device.
)

const (
	ParamDeviceID     = "uid"
	ParamDeviceStatus = "status"
	ParamTagName      = "name"
)

type filterQuery struct {
	Filter  string `query:"filter"`
	Status  string `query:"status"`
	SortBy  string `query:"sort_by"`
	OrderBy string `query:"order_by"`
	paginator.Query
}

func (h *Handler) GetDeviceList(c gateway.Context) error {
	query := filterQuery{}
	if err := c.Bind(&query); err != nil {
		return err
	}

	query.Normalize()

	raw, err := base64.StdEncoding.DecodeString(query.Filter)
	if err != nil {
		return err
	}

	var filter []models.Filter
	if err := json.Unmarshal(raw, &filter); len(raw) > 0 && err != nil {
		return err
	}

	devices, count, err := h.service.ListDevices(c.Ctx(), query.Query, filter, query.Status, query.SortBy, query.OrderBy)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, devices)
}

func (h *Handler) GetDevice(c gateway.Context) error {
	var req request.DeviceGet
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
}

func (h *Handler) DeleteDevice(c gateway.Context) error {
	var req request.DeviceDelete
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
}

func (h *Handler) RenameDevice(c gateway.Context) error {
	var req request.DeviceRename
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
}

func (h *Handler) OfflineDevice(c gateway.Context) error {
	var req request.DeviceOffline
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.service.UpdateDeviceStatus(c.Ctx(), models.UID(req.UID), false); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) LookupDevice(c gateway.Context) error {
	var req request.DeviceLookup
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	device, err := h.service.LookupDevice(c.Ctx(), req.Domain, req.Name)
	if err == services.ErrNotFound {
		return c.NoContent(http.StatusNotFound)
	} else if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, device)
}

func (h *Handler) UpdatePendingStatus(c gateway.Context) error {
	var req request.DevicePendingStatus
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

	status := map[string]string{
		"accept":  "accepted",
		"reject":  "rejected",
		"pending": "pending",
		"unused":  "unused",
	}
	err := guard.EvaluatePermission(c.Role(), guard.Actions.Device.Accept, func() error {
		err := h.service.UpdatePendingStatus(c.Ctx(), models.UID(req.UID), status[req.Status], tenant)

		return err
	})
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) HeartbeatDevice(c gateway.Context) error {
	var req request.DeviceHeartbeat
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	return h.service.DeviceHeartbeat(c.Ctx(), models.UID(req.UID))
}

func (h *Handler) CreateDeviceTag(c gateway.Context) error {
	var req request.DeviceCreateTag
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
}

func (h *Handler) RemoveDeviceTag(c gateway.Context) error {
	var req request.DeviceRemoveTag
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
}

func (h *Handler) UpdateDeviceTag(c gateway.Context) error {
	var req request.DeviceUpdateTag
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
}
