package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/server/api/services"
	log "github.com/sirupsen/logrus"
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
func (h *Handler) GetDeviceList(c *gateway.Context) error {
	req := new(requests.DeviceList)

	if err := c.Bind(req); err != nil {
		return err
	}

	req.Paginator.Normalize()
	req.Sorter.Normalize()

	if err := query.ValidateSorter(&req.Sorter, services.DeviceSortFields); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	if err := req.Filters.Unmarshal(); err != nil {
		log.WithError(err).WithField("filter", req.Filters.Raw).Warn("failed to decode device list filter")

		return c.NoContent(http.StatusBadRequest)
	}

	if err := query.ValidateFilters(&req.Filters, services.DeviceFilterFields); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	if c.QueryParam("connector") != "" {
		filter := []query.Filter{
			{
				Type: query.FilterTypeOperator,
				Params: &query.FilterOperator{
					Name: "and",
				},
			},
			{
				Type: query.FilterTypeProperty,
				Params: &query.FilterProperty{
					Name:     "platform",
					Operator: "eq",
					Value:    "connector",
				},
			},
		}

		req.Filters.Data = append(req.Filters.Data, filter...)
	} else {
		filter := []query.Filter{
			{
				Type: query.FilterTypeOperator,
				Params: &query.FilterOperator{
					Name: "and",
				},
			},
			{
				Type: query.FilterTypeProperty,
				Params: &query.FilterProperty{
					Name:     "platform",
					Operator: "ne",
					Value:    "connector",
				},
			},
		}

		req.Filters.Data = append(req.Filters.Data, filter...)
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	sc, err := c.AdminOrScope()
	if err != nil {
		return err
	}

	res, count, err := h.service.ListDevices(c.Ctx(), sc, req)
	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// GetDevice serves a single device by UID.
func (h *Handler) GetDevice(c *gateway.Context) error {
	var req requests.DeviceGet
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	sc, err := c.AdminOrScope()
	if err != nil {
		return err
	}

	device, err := h.service.GetDevice(c.Ctx(), sc, models.UID(req.UID))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, device)
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
