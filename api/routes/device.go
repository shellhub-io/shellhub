package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/apicontext"
	"github.com/shellhub-io/shellhub/api/deviceadm"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	GetDeviceListURL = "/devices"
	GetDeviceURL     = "/devices/:uid"
	DeleteDeviceURL  = "/devices/:uid"
	RenameDeviceURL  = "/devices/:uid"
	OfflineDeviceURL = "/devices/:uid/offline"
	LookupDeviceURL  = "/lookup"
	UpdateStatusURL  = "/devices/:uid/:status"
)

const TenantIDHeader = "X-Tenant-ID"

type filterQuery struct {
	Filter string `query:"filter"`
	paginator.Query
	Status  string `query:"status"`
	SortBy  string `query:"sort_by"`
	OrderBy string `query:"order_by"`
}

func GetDeviceList(c apicontext.Context) error {
	svc := deviceadm.NewService(c.Store())

	query := filterQuery{}
	if err := c.Bind(&query); err != nil {
		return err
	}

	query.Normalize()

	devices, count, err := svc.ListDevices(c.Ctx(), query.Query, query.Filter, query.Status, query.SortBy, query.OrderBy)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, devices)
}

func GetDevice(c apicontext.Context) error {
	svc := deviceadm.NewService(c.Store())

	device, err := svc.GetDevice(c.Ctx(), models.UID(c.Param("uid")))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, device)
}

func DeleteDevice(c apicontext.Context) error {
	svc := deviceadm.NewService(c.Store())

	tenant := ""
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}

	username := ""
	if v := c.Username(); v != nil {
		username = v.ID
	}

	if err := svc.DeleteDevice(c.Ctx(), models.UID(c.Param("uid")), tenant, username); err != nil {
		if err == deviceadm.ErrUnauthorized {
			return c.NoContent(http.StatusForbidden)
		}

		return err
	}

	return nil
}

func RenameDevice(c apicontext.Context) error {
	var req struct {
		Name string `json:"name"`
	}

	if err := c.Bind(&req); err != nil {
		return err
	}

	tenant := ""
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}

	username := ""
	if v := c.Username(); v != nil {
		username = v.ID
	}

	svc := deviceadm.NewService(c.Store())

	err := svc.RenameDevice(c.Ctx(), models.UID(c.Param("uid")), req.Name, tenant, username)
	switch err {
	case deviceadm.ErrUnauthorized:
		return c.NoContent(http.StatusForbidden)
	case deviceadm.ErrDuplicatedDeviceName:
		return c.NoContent(http.StatusConflict)
	default:
		return err
	}
}

func OfflineDevice(c apicontext.Context) error {
	svc := deviceadm.NewService(c.Store())

	if err := svc.UpdateDeviceStatus(c.Ctx(), models.UID(c.Param("uid")), false); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, nil)
}

func LookupDevice(c apicontext.Context) error {
	var query struct {
		Domain    string `query:"domain"`
		Name      string `query:"name"`
		Username  string `query:"username"`
		IPAddress string `query:"ip_address"`
	}

	if err := c.Bind(&query); err != nil {
		return err
	}

	svc := deviceadm.NewService(c.Store())

	device, err := svc.LookupDevice(c.Ctx(), query.Domain, query.Name)
	if err != nil {
		return nil
	}

	return c.JSON(http.StatusOK, device)
}

func UpdatePendingStatus(c apicontext.Context) error {
	svc := deviceadm.NewService(c.Store())

	tenant := ""
	if v := c.Tenant(); v != nil {
		tenant = v.ID
	}

	username := ""
	if v := c.Username(); v != nil {
		username = v.ID
	}

	status := map[string]string{
		"accept":  "accepted",
		"reject":  "rejected",
		"pending": "pending",
		"unused":  "unused",
	}

	if err := svc.UpdatePendingStatus(c.Ctx(), models.UID(c.Param("uid")), status[c.Param("status")], tenant, username); err != nil {
		if err == deviceadm.ErrUnauthorized {
			return c.NoContent(http.StatusForbidden)
		}

		return err
	}
	return c.JSON(http.StatusOK, nil)
}
