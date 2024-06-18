package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	GetDeviceListURL            = "/devices"
	GetDeviceURL                = "/devices/:uid"
	GetDeviceByPublicURLAddress = "/devices/public/:address"
	DeleteDeviceURL             = "/devices/:uid"
	RenameDeviceURL             = "/devices/:uid"
	OfflineDeviceURL            = "/devices/:uid/offline"
	LookupDeviceURL             = "/lookup"
	UpdateDeviceStatusURL       = "/devices/:uid/:status"
	CreateTagURL                = "/devices/:uid/tags"      // Add a tag to a device.
	UpdateTagURL                = "/devices/:uid/tags"      // Update device's tags with a new set.
	RemoveTagURL                = "/devices/:uid/tags/:tag" // Delete a tag from a device.
	UpdateDevice                = "/devices/:uid"
)

const (
	ParamDeviceID     = "uid"
	ParamDeviceStatus = "status"
	ParamTagName      = "name"
)

func (h *Handler) GetDeviceList(c gateway.Context) error {
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
}

func (h *Handler) GetDevice(c gateway.Context) error {
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
}

func (h *Handler) GetDeviceByPublicURLAddress(c gateway.Context) error {
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
}

func (h *Handler) DeleteDevice(c gateway.Context) error {
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

	if err := h.service.DeleteDevice(c.Ctx(), models.UID(req.UID), tenant); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) RenameDevice(c gateway.Context) error {
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

	if err := h.service.RenameDevice(c.Ctx(), models.UID(req.UID), req.Name, tenant); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) OfflineDevice(c gateway.Context) error {
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
}

func (h *Handler) LookupDevice(c gateway.Context) error {
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
}

func (h *Handler) UpdateDeviceStatus(c gateway.Context) error {
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

	if err := h.service.UpdateDeviceStatus(c.Ctx(), tenant, models.UID(req.UID), status[req.Status]); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) CreateDeviceTag(c gateway.Context) error {
	var req requests.DeviceCreateTag
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.service.CreateDeviceTag(c.Ctx(), models.UID(req.UID), req.Tag); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) RemoveDeviceTag(c gateway.Context) error {
	var req requests.DeviceRemoveTag
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.service.RemoveDeviceTag(c.Ctx(), models.UID(req.UID), req.Tag); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) UpdateDeviceTag(c gateway.Context) error {
	var req requests.DeviceUpdateTag
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.service.UpdateDeviceTag(c.Ctx(), models.UID(req.UID), req.Tags); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) UpdateDevice(c gateway.Context) error {
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

	if err := h.service.UpdateDevice(c.Ctx(), tenant, models.UID(req.UID), req.Name, req.PublicURL); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
