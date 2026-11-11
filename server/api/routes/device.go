package routes

import (
	"context"
	"maps"
	"net/http"
	"net/url"
	"strconv"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	errs "github.com/shellhub-io/shellhub/server/api/routes/errors"
	"github.com/shellhub-io/shellhub/server/api/services"
	log "github.com/sirupsen/logrus"
)

// The device routes, relative to the API's base path.
const (
	GetDeviceListURL           = "/devices"
	GetDeviceURL               = "/devices/:uid"
	ResolveDeviceURL           = "/devices/resolve"
	DeviceLocationsURL         = "/devices/locations"
	DeleteDeviceURL            = "/devices/:uid"
	RenameDeviceURL            = "/devices/:uid"
	UpdateDeviceStatusURL      = "/devices/:uid/:status"
	UpdateDevice               = "/devices/:uid"
	SetDeviceCustomFieldURL    = "/devices/:uid/custom_fields/:key"
	DeleteDeviceCustomFieldURL = "/devices/:uid/custom_fields/:key"
	MakeTeamDeviceURL          = "/devices/:uid/owner"
)

const geoipSourceHeader = "X-GeoIP-Source"

const (
	pageSizeParam  = "page[size]"
	pageAfterParam = "page[after]"
)

var deviceListPageModeParams = []string{"page", "per_page", "sort_by", "order_by"}

// GetDeviceList serves the namespace's devices, filtered by status, connector and filter. In page
// mode it serves the numbered page sorted as requested, with the total in X-Total-Count. A
// page[size] or page[after] parameter selects cursor mode, which serves the devices in creation
// order after the cursor and hands back the next page in Link and X-Next-Cursor, leaving both out
// on the last page. Both modes answer 422 to a page, per_page, page[size] or connector that does
// not parse, 403 when the request carries no namespace outside the admin console, 400 to a status
// or filter they cannot read, and 404 when the status is not removed and reading the namespace's
// device limit fails, whether the namespace is gone or the store failed. Page mode also answers
// 400 to a sort_by it cannot sort by. Cursor mode answers 400 to page, per_page, sort_by or
// order_by in the same request, to a page[after] the server did not issue, and on the admin
// cross-namespace list. Any other store failure is a 500.
func (h *Handler) GetDeviceList(c *gateway.Context) error {
	req := new(requests.DeviceList)

	if err := c.Bind(req); err != nil {
		return err
	}

	params := c.QueryParams()
	if params.Has(pageSizeParam) || params.Has(pageAfterParam) {
		return h.listDevicesAfter(c, req, params)
	}

	return h.listDevicesByPage(c, req)
}

func (h *Handler) listDevicesByPage(c *gateway.Context, req *requests.DeviceList) error {
	req.Paginator.Normalize()
	req.Sorter.Normalize()

	if err := query.ValidateSorter(&req.Sorter, services.DeviceSortFields); err != nil {
		return errs.NewErrInvalidEntity(map[string]string{"sort_by": err.Error()})
	}

	if err := validateDeviceListRequest(c, req); err != nil {
		return err
	}

	sc, err := c.AdminOrScope()
	if err != nil {
		return err
	}

	res, count, err := h.service.ListDevices(c.Ctx(), sc, req)
	c.Response().Header().Set("X-Total-Count", strconv.FormatInt(count, 10))

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) listDevicesAfter(c *gateway.Context, req *requests.DeviceList, params url.Values) error {
	if fields := cursorModeConflicts(params); len(fields) > 0 {
		return errs.NewErrInvalidEntity(fields)
	}

	var after *query.Cursor
	if params.Has(pageAfterParam) {
		cursor, err := query.ParseCursor(req.After)
		if err != nil {
			return errs.NewErrInvalidEntity(map[string]string{pageAfterParam: err.Error()})
		}

		after = &cursor
	}

	if err := validateDeviceListRequest(c, req); err != nil {
		return err
	}

	sc, err := c.AdminOrScope()
	if err != nil {
		return err
	}

	if !sc.IsBounded() {
		sent := pageSizeParam
		if !params.Has(pageSizeParam) {
			sent = pageAfterParam
		}

		return errs.NewErrInvalidEntity(map[string]string{sent: "cursor pagination is only available on a namespace's device list"})
	}

	req.CursorPaginator.Normalize(h.deviceListMaxPageSize)

	devices, next, err := h.service.ListDevicesAfter(c.Ctx(), sc, req, after)
	if err != nil {
		return err
	}

	if next != nil {
		cursor := next.String()
		c.Response().Header().Set("Link", "<"+nextDevicePageURL(c, params, req.Size, cursor)+`>; rel="next"`)
		c.Response().Header().Set("X-Next-Cursor", cursor)
	}

	return c.JSON(http.StatusOK, devices)
}

func validateDeviceListRequest(c *gateway.Context, req *requests.DeviceList) error {
	if err := req.Filters.Unmarshal(); err != nil {
		log.WithError(err).WithField("filter", req.Filters.Raw).Warn("failed to decode device list filter")

		return errs.NewErrInvalidEntity(map[string]string{"filter": err.Error()})
	}

	if err := query.ValidateFilters(&req.Filters, services.DeviceFilterFields); err != nil {
		return errs.NewErrInvalidEntity(map[string]string{"filter": err.Error()})
	}

	return c.Validate(req)
}

func cursorModeConflicts(params url.Values) map[string]string {
	fields := map[string]string{}
	for _, name := range deviceListPageModeParams {
		if params.Has(name) {
			fields[name] = "cannot be combined with " + pageSizeParam + " or " + pageAfterParam
		}
	}

	return fields
}

func nextDevicePageURL(c *gateway.Context, params url.Values, size int64, cursor string) string {
	values := maps.Clone(params)
	values.Set(pageSizeParam, strconv.FormatInt(size, 10))
	values.Set(pageAfterParam, cursor)

	reached := c.Origin()
	next := url.URL{
		Scheme:   reached.Scheme,
		Host:     services.ClientFacingAddress(reached.Scheme, reached.Host, reached.Port),
		Path:     c.Request().URL.Path,
		RawQuery: values.Encode(),
	}

	return next.String()
}

// GetDevice returns the device with req's UID within sc. It returns [services.ErrDeviceNotFound]
// when the read fails for any reason: a device that does not exist or lies outside sc, a store
// failure and a cancelled ctx alike.
func (h *Handler) GetDevice(ctx context.Context, sc scope.Scope, _ gateway.Principal, req *requests.DeviceGet) (*models.Device, error) {
	return h.service.GetDevice(ctx, sc, models.UID(req.UID))
}

// GetDeviceAsAdmin returns the device with req's UID in any namespace to an instance administrator,
// sc spanning every namespace. It returns [services.ErrDeviceNotFound] as [Handler.GetDevice] does.
func (h *Handler) GetDeviceAsAdmin(ctx context.Context, sc scope.Scope, _ gateway.Admin, req *requests.DeviceGet) (*models.Device, error) {
	return h.service.GetDevice(ctx, sc, models.UID(req.UID))
}

// GetDeviceLocations serves the namespace's accepted devices counted by position, for drawing them
// on a map without reading every device, and names the GeoIP database that placed them in the
// X-GeoIP-Source header, left out when the server opened none. It answers 400 to a filter it cannot read, 403 when the
// request carries no namespace outside the admin console, and 500 to any other failure, the
// request's context ending while the store reads included.
func (h *Handler) GetDeviceLocations(c *gateway.Context) error {
	req := new(requests.DeviceLocationList)
	if err := c.Bind(req); err != nil {
		return err
	}

	if err := req.Filters.Unmarshal(); err != nil {
		return errs.NewErrInvalidEntity(map[string]string{"filter": err.Error()})
	}

	if err := query.ValidateFilters(&req.Filters, services.DeviceFilterFields); err != nil {
		return errs.NewErrInvalidEntity(map[string]string{"filter": err.Error()})
	}

	sc, err := c.AdminOrScope()
	if err != nil {
		return err
	}

	locations, err := h.service.ListDeviceLocations(c.Ctx(), sc, req)
	if err != nil {
		return err
	}

	if h.geoipSource != "" {
		c.Response().Header().Set(geoipSourceHeader, h.geoipSource)
	}

	return c.JSON(http.StatusOK, locations)
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

// MakeTeamDevice clears a device's owner, so it stays in the namespace when the member who paired
// it leaves.
func (h *Handler) MakeTeamDevice(c *gateway.Context) error {
	var req requests.DeviceParam
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

	if err := h.service.MakeTeamDevice(c.Ctx(), tenant, req.UID); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}
