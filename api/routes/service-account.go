package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
)

const (
	ListServiceAccountsURL  = "/service-accounts"
	CreateServiceAccountURL = "/service-accounts"
	DeleteServiceAccountURL = "/service-accounts/:id"
)

// ListServiceAccounts returns the namespace's service accounts with their identities.
func (h *Handler) ListServiceAccounts(c gateway.Context) error {
	req := new(requests.ServiceAccountList)
	if err := c.Bind(req); err != nil {
		return err
	}

	if c.Tenant() != nil {
		req.TenantID = c.Tenant().ID
	}

	list, err := h.service.ListServiceAccounts(c.Ctx(), req)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(len(list)))

	return c.JSON(http.StatusOK, list)
}

// CreateServiceAccount creates a service account from a display name and an OpenSSH
// public key.
func (h *Handler) CreateServiceAccount(c gateway.Context) error {
	req := new(requests.ServiceAccountCreate)
	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if c.Tenant() != nil {
		req.TenantID = c.Tenant().ID
	}

	account, err := h.service.CreateServiceAccount(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, account)
}

// DeleteServiceAccount removes a service account and cascades to its membership and
// SSH identities.
func (h *Handler) DeleteServiceAccount(c gateway.Context) error {
	req := new(requests.ServiceAccountDelete)
	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if c.Tenant() != nil {
		req.TenantID = c.Tenant().ID
	}

	if err := h.service.DeleteServiceAccount(c.Ctx(), req); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
