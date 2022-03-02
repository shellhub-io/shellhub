package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/authorizer"
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

func (h *Handler) GetDevice(c gateway.Context) error {
	device, err := h.service.GetDevice(c.Ctx(), models.UID(c.Param(ParamDeviceID)))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, device)
}

func (h *Handler) DeleteDevice(c gateway.Context) error {
	tenantID := ""
	if c.Tenant() != nil {
		tenantID = c.Tenant().ID
	}

	err := guard.EvaluatePermission(c.Role(), authorizer.Actions.Device.Remove, func() error {
		err := h.service.DeleteDevice(c.Ctx(), models.UID(c.Param(ParamDeviceID)), tenantID)

		return err
	})
	if err != nil {
		switch err {
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
		default:
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}

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

	err := guard.EvaluatePermission(c.Role(), authorizer.Actions.Device.Rename, func() error {
		err := h.service.RenameDevice(c.Ctx(), models.UID(c.Param(ParamDeviceID)), req.Name, tenantID)

		return err
	})
	if err != nil {
		switch err {
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
		case services.ErrDuplicatedDeviceName:
			return c.NoContent(http.StatusConflict)
		case services.ErrInvalidFormat:
			return c.NoContent(http.StatusBadRequest)
		default:
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) OfflineDevice(c gateway.Context) error {
	if err := h.service.UpdateDeviceStatus(c.Ctx(), models.UID(c.Param(ParamDeviceID)), false); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

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
	err := guard.EvaluatePermission(c.Role(), authorizer.Actions.Device.Accept, func() error {
		err := h.service.UpdatePendingStatus(c.Ctx(), models.UID(c.Param(ParamDeviceID)), status[c.Param(ParamDeviceStatus)], tenantID)

		return err
	})
	if err != nil {
		switch err {
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
		case services.ErrBadRequest:
			return c.NoContent(http.StatusBadRequest)
		case services.ErrMaxDeviceCountReached:
			return c.NoContent(http.StatusPaymentRequired)
		}

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

	err := guard.EvaluatePermission(c.Role(), authorizer.Actions.Device.CreateTag, func() error {
		return h.service.CreateDeviceTag(c.Ctx(), models.UID(c.Param(ParamDeviceID)), req.Name)
	})
	if err != nil {
		switch err {
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
		case services.ErrInvalidFormat:
			return c.NoContent(http.StatusBadRequest)
		case services.ErrDeviceNotFound:
			return c.NoContent(http.StatusNotFound)
		case services.ErrMaxTagReached:
			return c.NoContent(http.StatusNotAcceptable)
		case services.ErrDuplicateTagName:
			return c.NoContent(http.StatusConflict)
		default:
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) RemoveDeviceTag(c gateway.Context) error {
	err := guard.EvaluatePermission(c.Role(), authorizer.Actions.Device.RemoveTag, func() error {
		return h.service.RemoveDeviceTag(c.Ctx(), models.UID(c.Param(ParamDeviceID)), c.Param(ParamTagName))
	})
	if err != nil {
		switch err {
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
		case services.ErrDeviceNotFound:
			return c.NoContent(http.StatusNotFound)
		case services.ErrTagNameNotFound:
			return c.NoContent(http.StatusNotFound)
		default:
			return err
		}
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

	err := guard.EvaluatePermission(c.Role(), authorizer.Actions.Device.UpdateTag, func() error {
		return h.service.UpdateDeviceTag(c.Ctx(), models.UID(c.Param(ParamDeviceID)), req.Tags)
	})
	if err != nil {
		switch err {
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
		case services.ErrInvalidFormat:
			return c.NoContent(http.StatusBadRequest)
		case services.ErrMaxTagReached:
			return c.NoContent(http.StatusNotAcceptable)
		case services.ErrDeviceNotFound:
			return c.NoContent(http.StatusNotFound)
		default:
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}
