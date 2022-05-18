package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
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
	CreateTagURL       = "/devices/:uid/tags"       // Add a tag to a device.
	UpdateTagURL       = "/devices/:uid/tags"       // Update device's tags with a new set.
	RemoveTagURL       = "/devices/:uid/tags/:name" // Delete a tag from a device.
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

// @summary Get a list of devices.
// @description Return a list of devices between a defined range. You can also set a filter, sort by, order by and get only devices with a specific status.
// @tags community,devices
// @security jwt
// @param filter query string false "Device's Filter field receives a base64 encoded JSON object" example(W3sidHlwZSI6InByb3BlcnR5IiwicGFyYW1zIjp7Im5hbWUiOiJuYW1lIiwib3BlcmF0b3IiOiJjb250YWlucyIsInZhbHVlIjoiZDAifX1d)
// @param status query string false "Device's status" enums(accepted, rejected, pending, unused) example(accepted)
// @param sort_by query string false "Device's property to sort of"
// @param order_by query string false "Device's list order" enums(asc, desc)
// @param page query int true "Page number" example(1)
// @param perPage query int true "Number of items per page" example(10)
// @produce json
// @success 200 {array} models.Device
// @header 200 {string} X-Total-Count "Device's total number"
// @failure 401 {object} nil
// @failure 500 {object} nil
// @router /devices [get]
func (h *Handler) GetDeviceList(c gateway.Context) error {
	query := filterQuery{}
	if err := c.Bind(&query); err != nil {
		return err
	}

	query.Normalize()

	devices, count, err := h.service.ListDevices(c.Ctx(), query.Query, query.Filter, query.Status, query.SortBy, query.OrderBy)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, devices)
}

// @summary Get a device.
// @description Return a device by its ID.
// @tags community,devices
// @security jwt
// @param uid path string true "Device's UID." example(13b0c8ea878e61ff849db69461795006a9594c8f6a6390ce0000100b0c9d7d0a)
// @produce json
// @success 200 {object} models.Device
// @failure 401 {object} nil
// @failure 403 {object} nil
// @failure 404 {object} nil
// @failure 500 {object} nil
// @router /devices/{uid} [get]
func (h *Handler) GetDevice(c gateway.Context) error {
	device, err := h.service.GetDevice(c.Ctx(), models.UID(c.Param(ParamDeviceID)))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, device)
}

// @summary Delete a device.
// @description Delete a device by its ID.
// @tags community,devices
// @param uid path string true "Device's UID." example(13b0c8ea878e61ff849db69461795006a9594c8f6a6390ce0000100b0c9d7d0a)
// @security jwt
// @produce json
// @success 200 {object} nil
// @failure 401 {object} nil
// @failure 403 {object} nil
// @failure 500 {object} nil
// @router /devices/{uid} [delete]
func (h *Handler) DeleteDevice(c gateway.Context) error {
	tenantID := ""
	if c.Tenant() != nil {
		tenantID = c.Tenant().ID
	}

	err := guard.EvaluatePermission(c.Role(), guard.Actions.Device.Remove, func() error {
		err := h.service.DeleteDevice(c.Ctx(), models.UID(c.Param(ParamDeviceID)), tenantID)

		return err
	})
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// @summary Rename a device.
// @description Rename a device by its ID.
// @tags community,devices
// @param uid path string true "Device's UID." example(13b0c8ea878e61ff849db69461795006a9594c8f6a6390ce0000100b0c9d7d0a)
// @accept json
// @param _ body object{name=string} true "Device's new name."
// @security jwt
// @produce json
// @success 200 {object} nil
// @failure 400 {object} nil
// @failure 401 {object} nil
// @failure 403 {object} nil
// @failure 404 {object} nil
// @failure 409 {object} nil
// @failure 500 {object} nil
// @router /devices/{uid} [patch]
func (h *Handler) RenameDevice(c gateway.Context) error {
	var req struct {
		Name string `json:"name"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	tenantID := ""
	if c.Tenant() != nil {
		tenantID = c.Tenant().ID
	}

	err := guard.EvaluatePermission(c.Role(), guard.Actions.Device.Rename, func() error {
		err := h.service.RenameDevice(c.Ctx(), models.UID(c.Param(ParamDeviceID)), req.Name, tenantID)

		return err
	})
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// @summary Update a device's status to online.
// @description Update a device's status to online.
// @tags community,devices
// @param uid path string true "Device's UID." example(13b0c8ea878e61ff849db69461795006a9594c8f6a6390ce0000100b0c9d7d0a)
// @security jwt
// @produce json
// @success 200 {object} nil
// @failure 401 {object} nil
// @failure 500 {object} nil
// @router /devices/{uid}/offline [post]
func (h *Handler) OfflineDevice(c gateway.Context) error {
	if err := h.service.UpdateDeviceStatus(c.Ctx(), models.UID(c.Param(ParamDeviceID)), false); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// @summary Lookup for device.
// @description Lookup for device.
// @tags community,devices
// @param domain query string false "Device's domain."
// @param name query string false "Device's name."
// @param username query string false "Device's username."
// @param ip_address query string false "Device's IP address."
// @produce json
// @success 200 {object} models.Device
// @failure 401 {object} nil
// @failure 403 {object} nil
// @failure 404 {object} nil
// @failure 500 {object} nil
// @router /devices/lookup [get]
func (h *Handler) LookupDevice(c gateway.Context) error {
	var query struct {
		Domain    string `query:"domain"`
		Name      string `query:"name"`
		Username  string `query:"username"`
		IPAddress string `query:"ip_address"`
	}

	if err := c.Bind(&query); err != nil {
		return err
	}

	device, err := h.service.LookupDevice(c.Ctx(), query.Domain, query.Name)
	if err == services.ErrNotFound {
		return c.NoContent(http.StatusNotFound)
	} else if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, device)
}

// @summary Update a device's status.
// @description Update a device's status.
// @tags community,devices
// @param uid path string true "Device's UID." example(13b0c8ea878e61ff849db69461795006a9594c8f6a6390ce0000100b0c9d7d0a)
// @param status query string true "Device's status." enums(accepted, rejected, pending, unused)
// @produce json
// @success 200 {object} nil
// @failure 400 {object} nil
// @failure 401 {object} nil
// @failure 402 {object} nil
// @failure 403 {object} nil
// @failure 404 {object} nil
// @failure 500 {object} nil
// @router /devices/{uid}/{status} [patch]
func (h *Handler) UpdatePendingStatus(c gateway.Context) error {
	tenantID := ""
	if c.Tenant() != nil {
		tenantID = c.Tenant().ID
	}

	status := map[string]string{
		"accept":  "accepted",
		"reject":  "rejected",
		"pending": "pending",
		"unused":  "unused",
	}
	err := guard.EvaluatePermission(c.Role(), guard.Actions.Device.Accept, func() error {
		err := h.service.UpdatePendingStatus(c.Ctx(), models.UID(c.Param(ParamDeviceID)), status[c.Param(ParamDeviceStatus)], tenantID)

		return err
	})
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) HeartbeatDevice(c gateway.Context) error {
	return h.service.DeviceHeartbeat(c.Ctx(), models.UID(c.Param(ParamDeviceID)))
}

func (h *Handler) CreateDeviceTag(c gateway.Context) error {
	var req struct {
		Name string `json:"name"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	err := guard.EvaluatePermission(c.Role(), guard.Actions.Device.CreateTag, func() error {
		return h.service.CreateDeviceTag(c.Ctx(), models.UID(c.Param(ParamDeviceID)), req.Name)
	})
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) RemoveDeviceTag(c gateway.Context) error {
	err := guard.EvaluatePermission(c.Role(), guard.Actions.Device.RemoveTag, func() error {
		return h.service.RemoveDeviceTag(c.Ctx(), models.UID(c.Param(ParamDeviceID)), c.Param(ParamTagName))
	})
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) UpdateDeviceTag(c gateway.Context) error {
	var req struct {
		Tags []string `json:"tags"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	err := guard.EvaluatePermission(c.Role(), guard.Actions.Device.UpdateTag, func() error {
		return h.service.UpdateDeviceTag(c.Ctx(), models.UID(c.Param(ParamDeviceID)), req.Tags)
	})
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
