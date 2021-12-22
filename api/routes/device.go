package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/apicontext"
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
	GetTagsURL         = "/devices/tags"            // Get tags from all device.
	RenameTagURL       = "/devices/tags/:name"      // Rename a tag inside all devices.
	DeleteTagsURL      = "/devices/tags/:name"      // Delete all tag's occurrence inside all devices with a specif name.
)

const (
	ParamDeviceID     = "uid"
	ParamDeviceStatus = "status"
	ParamTagName      = "name"
)

type filterQuery struct {
	Filter string `query:"filter"`
	paginator.Query
	Status  string `query:"status"`
	SortBy  string `query:"sort_by"`
	OrderBy string `query:"order_by"`
}

func (h *Handler) GetDeviceList(c apicontext.Context) error {
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

func (h *Handler) GetDevice(c apicontext.Context) error {
	device, err := h.service.GetDevice(c.Ctx(), models.UID(c.Param(ParamDeviceID)))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, device)
}

func (h *Handler) DeleteDevice(c apicontext.Context) error {
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

func (h *Handler) RenameDevice(c apicontext.Context) error {
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

func (h *Handler) OfflineDevice(c apicontext.Context) error {
	if err := h.service.UpdateDeviceStatus(c.Ctx(), models.UID(c.Param(ParamDeviceID)), false); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) LookupDevice(c apicontext.Context) error {
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

func (h *Handler) UpdatePendingStatus(c apicontext.Context) error {
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

func (h *Handler) HeartbeatDevice(c apicontext.Context) error {
	return h.service.DeviceHeartbeat(c.Ctx(), models.UID(c.Param(ParamDeviceID)))
}

func (h *Handler) CreateTag(c apicontext.Context) error {
	var req struct {
		Name string `json:"name"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	err := guard.EvaluatePermission(c.Role(), authorizer.Actions.Device.CreateTag, func() error {
		return h.service.CreateTag(c.Ctx(), models.UID(c.Param(ParamDeviceID)), req.Name)
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

func (h *Handler) RemoveTag(c apicontext.Context) error {
	err := guard.EvaluatePermission(c.Role(), authorizer.Actions.Device.RemoveTag, func() error {
		return h.service.RemoveTag(c.Ctx(), models.UID(c.Param(ParamDeviceID)), c.Param(ParamTagName))
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

func (h *Handler) RenameTag(c apicontext.Context) error {
	var req struct {
		Name string `json:"name"`
	}

	tenant := ""
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	err := guard.EvaluatePermission(c.Role(), authorizer.Actions.Device.RenameTag, func() error {
		return h.service.RenameTag(c.Ctx(), tenant, c.Param(ParamTagName), req.Name)
	})
	if err != nil {
		switch err {
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
		case services.ErrInvalidFormat:
			return c.NoContent(http.StatusBadRequest)
		case services.ErrNoTags:
			return c.NoContent(http.StatusNotFound)
		case services.ErrTagNameNotFound:
			return c.NoContent(http.StatusNotFound)
		case services.ErrDuplicateTagName:
			return c.NoContent(http.StatusConflict)
		default:
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) UpdateTag(c apicontext.Context) error {
	var req struct {
		Tags []string `json:"tags"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	err := guard.EvaluatePermission(c.Role(), authorizer.Actions.Device.UpdateTag, func() error {
		return h.service.UpdateTag(c.Ctx(), models.UID(c.Param(ParamDeviceID)), req.Tags)
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

func (h *Handler) GetTags(c apicontext.Context) error {
	tenant := ""
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}

	tags, count, err := h.service.GetTags(c.Ctx(), tenant)
	if err != nil {
		switch err {
		case services.ErrNamespaceNotFound:
			return c.NoContent(http.StatusNotFound)
		default:
			return err
		}
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, tags)
}

func (h *Handler) DeleteTags(c apicontext.Context) error {
	tenant := ""
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}

	err := guard.EvaluatePermission(c.Role(), authorizer.Actions.Device.DeleteTag, func() error {
		return h.service.DeleteTags(c.Ctx(), tenant, c.Param(ParamTagName))
	})
	if err != nil {
		switch err {
		case guard.ErrForbidden:
			return c.NoContent(http.StatusForbidden)
		case services.ErrNamespaceNotFound:
			return c.NoContent(http.StatusNotFound)
		default:
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}
