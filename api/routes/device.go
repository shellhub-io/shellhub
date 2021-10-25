package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/pkg/authorizer"

	"github.com/shellhub-io/shellhub/api/apicontext"
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
	CreateTagURL       = "/devices/:uid/tags"
	DeleteTagURL       = "/devices/:uid/tags/:name"
	RenameTagURL       = "/devices/tags/:name"
	ListTagURL         = "/devices/tags"
	UpdateTagURL       = "/devices/:uid/tags"
	GetTagsURL         = "/devices/tags"
	DeleteAllTagsURL   = "/devices/tags/:name"
)

const TenantIDHeader = "X-Tenant-ID"

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
	device, err := h.service.GetDevice(c.Ctx(), models.UID(c.Param("uid")))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, device)
}

func (h *Handler) DeleteDevice(c apicontext.Context) error {
	userID := ""
	tenantID := ""
	if c.ID() != nil && c.Tenant() != nil {
		userID = c.ID().ID
		tenantID = c.Tenant().ID
	}

	err := h.service.CheckPermission(c.Ctx(), tenantID, userID, authorizer.Actions.Device.Remove, func() error {
		err := h.service.DeleteDevice(c.Ctx(), models.UID(c.Param("uid")), tenantID)

		return err
	})
	if err != nil {
		switch err {
		case services.ErrForbidden:
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

	userID := ""
	tenantID := ""
	if c.ID() != nil && c.Tenant() != nil {
		userID = c.ID().ID
		tenantID = c.Tenant().ID
	}

	err := h.service.CheckPermission(c.Ctx(), tenantID, userID, authorizer.Actions.Device.Rename, func() error {
		err := h.service.RenameDevice(c.Ctx(), models.UID(c.Param("uid")), req.Name, tenantID)

		return err
	})
	if err != nil {
		switch err {
		case services.ErrForbidden:
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
	if err := h.service.UpdateDeviceStatus(c.Ctx(), models.UID(c.Param("uid")), false); err != nil {
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
	userID := ""
	tenantID := ""
	if c.ID() != nil && c.Tenant() != nil {
		userID = c.ID().ID
		tenantID = c.Tenant().ID
	}

	status := map[string]string{
		"accept":  "accepted",
		"reject":  "rejected",
		"pending": "pending",
		"unused":  "unused",
	}
	err := h.service.CheckPermission(c.Ctx(), tenantID, userID, authorizer.Actions.Device.Accept, func() error {
		err := h.service.UpdatePendingStatus(c.Ctx(), models.UID(c.Param("uid")), status[c.Param("status")], tenantID)

		return err
	})
	if err != nil {
		switch err {
		case services.ErrForbidden:
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
	return h.service.DeviceHeartbeat(c.Ctx(), models.UID(c.Param("uid")))
}

func (h *Handler) CreateTag(c apicontext.Context) error {
	var req struct {
		Name string `json:"name"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := h.service.CreateTag(c.Ctx(), models.UID(c.Param("uid")), req.Name); err != nil {
		switch err {
		case services.ErrUnauthorized:
			return c.NoContent(http.StatusForbidden)
		case services.ErrInvalidFormat:
			return c.NoContent(http.StatusBadRequest)
		case services.ErrMaxTagReached:
			return c.NoContent(http.StatusNotAcceptable)
		case services.ErrDuplicateTagName:
			return c.NoContent(http.StatusConflict)
		}

		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) DeleteTag(c apicontext.Context) error {
	if err := h.service.DeleteTag(c.Ctx(), models.UID(c.Param("uid")), c.Param("name")); err != nil {
		switch err {
		case services.ErrUnauthorized:
			return c.NoContent(http.StatusForbidden)
		case services.ErrNotFound:
			return c.NoContent(http.StatusNotFound)
		}

		return err
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

	if err := h.service.RenameTag(c.Ctx(), tenant, c.Param("name"), req.Name); err != nil {
		switch err {
		case services.ErrUnauthorized:
			return c.NoContent(http.StatusForbidden)
		case services.ErrInvalidFormat:
			return c.NoContent(http.StatusBadRequest)
		case services.ErrNotFound:
			return c.NoContent(http.StatusNotFound)
		}

		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) ListTag(c apicontext.Context) error {
	tags, count, err := h.service.ListTag(c.Ctx())
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, tags)
}

func (h *Handler) UpdateTag(c apicontext.Context) error {
	var req struct {
		Tags []string `json:"tags"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := h.service.UpdateTag(c.Ctx(), models.UID(c.Param("uid")), req.Tags); err != nil {
		switch err {
		case services.ErrUnauthorized:
			return c.NoContent(http.StatusForbidden)
		case services.ErrInvalidFormat:
			return c.NoContent(http.StatusBadRequest)
		case services.ErrMaxTagReached:
			return c.NoContent(http.StatusForbidden)
		case services.ErrNotFound:
			return c.NoContent(http.StatusNotFound)
		}

		return err
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) GetTags(c apicontext.Context) error {
	tenant := ""
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}

	tags, count, err := h.service.GetTags(c.Ctx(), tenant)
	if err == services.ErrUnauthorized {
		return c.NoContent(http.StatusForbidden)
	} else if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, tags)
}

func (h *Handler) DeleteAllTags(c apicontext.Context) error {
	tenant := ""
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}

	if err := h.service.DeleteAllTags(c.Ctx(), tenant, c.Param("name")); err == services.ErrUnauthorized {
		return c.NoContent(http.StatusForbidden)
	} else if err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
