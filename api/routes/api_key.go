package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
)

const (
	CreateAPIKeyURL = "/namespaces/:tenant/api-key"
	EditAPIKeyURL   = "/namespaces/:tenant/api-key/:key"
	ListAPIKeysURL  = "/namespaces/:tenant/api-key"
	DeleteAPIKeyURL = "/namespaces/:tenant/api-key/:key"
)

func (h *Handler) CreateAPIKey(c gateway.Context) error {
	var req requests.CreateAPIKey

	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	var userID string
	if v := c.ID(); v != nil {
		userID = v.ID
	}

	key := c.Request().Header.Get("X-API-KEY")
	tenant := c.Request().Header.Get("X-Tenant-ID")

	uid, err := h.service.CreateAPIKey(c.Ctx(), userID, tenant, key, &req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, uid)
}

func (h *Handler) ListAPIKeys(c gateway.Context) error {
	var req requests.APIKeyList

	if err := c.Bind(&req); err != nil {
		return err
	}

	req.Paginator.Normalize()
	req.Sorter.Normalize()

	if req.Sorter.By == "" {
		req.Sorter.By = "expires_in"
		req.Sorter.Order = "desc"
	}

	if v := c.ID(); v != nil {
		req.UserID = v.ID
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	req.TenantParam.Tenant = c.Request().Header.Get("X-Tenant-ID")

	res, count, err := h.service.ListAPIKeys(c.Ctx(), &req)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, res)
}

func (h *Handler) EditAPIKey(c gateway.Context) error {
	var req requests.APIKeyChanges

	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	key, err := h.service.EditAPIKey(c.Ctx(), &req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, key)
}

func (h *Handler) DeleteAPIKey(c gateway.Context) error {
	var req requests.APIKeyID

	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	tenant := c.Request().Header.Get("X-Tenant-ID")

	err := h.service.DeleteAPIKey(c.Ctx(), req.ID, tenant)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, nil)
}
