package routes

import (
	"context"
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
)

// The device routes, relative to the API's base path.
const (
	GetDeviceListURL           = "/devices"
	GetDeviceURL               = "/devices/:uid"
	ResolveDeviceURL           = "/devices/resolve"
	DeleteDeviceURL            = "/devices/:uid"
	RenameDeviceURL            = "/devices/:uid"
	UpdateDeviceStatusURL      = "/devices/:uid/:status"
	UpdateDevice               = "/devices/:uid"
	SetDeviceCustomFieldURL    = "/devices/:uid/custom_fields/:key"
	DeleteDeviceCustomFieldURL = "/devices/:uid/custom_fields/:key"
)

// The path parameter names these routes bind by.
const (
	ParamDeviceID             = "uid"
	ParamDeviceStatus         = "status"
	ParamDeviceCustomFieldKey = "key"
)

// GetDeviceList serves the namespace's devices, filtered, sorted and paginated as requested.
func (h *Handler) GetDeviceList(ctx context.Context, sc scope.Scope, _ gateway.Actor, req *requests.DeviceList) ([]models.Device, int, error) {
	return h.service.ListDevices(ctx, sc, req)
}

// GetDevice serves a single device by UID.
func (h *Handler) GetDevice(ctx context.Context, sc scope.Scope, _ gateway.Actor, req *requests.DeviceGet) (*models.Device, error) {
	return h.service.GetDevice(ctx, sc, models.UID(req.UID))
}

// ResolveDevice serves the device matching a name or SSHID, for callers that have a name
// rather than a UID.
func (h *Handler) ResolveDevice(c *gateway.Context) error {
	var req requests.ResolveDevice
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	device, err := h.service.ResolveDevice(c.Ctx(), &req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, device)
}

// DeleteDevice removes a device, ending its access to the namespace.
func (h *Handler) DeleteDevice(c *gateway.Context) error {
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

// RenameDevice changes a device's name, which is part of the SSHID clients connect by.
func (h *Handler) RenameDevice(c *gateway.Context) error {
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

// UpdateDeviceStatus accepts or rejects a pending device, or returns an accepted one to the
// pending list.
func (h *Handler) UpdateDeviceStatus(c *gateway.Context) error {
	req := new(requests.DeviceUpdateStatus)

	if err := c.Bind(req); err != nil {
		return err
	}

	status := map[string]string{
		"accept":  string(models.DeviceStatusAccepted),
		"reject":  string(models.DeviceStatusRejected),
		"pending": string(models.DeviceStatusPending),
		"unused":  string(models.DeviceStatusUnused),
	}

	req.Status = status[req.Status]

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.service.UpdateDeviceStatus(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// UpdateDevice changes a device's editable attributes.
func (h *Handler) UpdateDevice(c *gateway.Context) error {
	req := new(requests.DeviceUpdate)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.service.UpdateDevice(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// SetDeviceCustomField sets one operator-defined field on a device.
func (h *Handler) SetDeviceCustomField(c *gateway.Context) error {
	req := new(requests.DeviceSetCustomField)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.service.SetDeviceCustomField(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// DeleteDeviceCustomField removes one operator-defined field from a device.
func (h *Handler) DeleteDeviceCustomField(c *gateway.Context) error {
	req := new(requests.DeviceDeleteCustomField)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := h.service.DeleteDeviceCustomField(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
