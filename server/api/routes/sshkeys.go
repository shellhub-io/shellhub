package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/server/api/services"
	log "github.com/sirupsen/logrus"
)

// The public key routes, relative to the API's base path.
const (
	GetPublicKeysURL   = "/sshkeys/public-keys"
	CreatePublicKeyURL = "/sshkeys/public-keys"
	UpdatePublicKeyURL = "/sshkeys/public-keys/:fingerprint"
	DeletePublicKeyURL = "/sshkeys/public-keys/:fingerprint"
)

// The path parameter name these routes bind by.
const (
	ParamPublicKeyFingerprint = "fingerprint"
)

// GetPublicKeys serves the namespace's public keys.
func (h *Handler) GetPublicKeys(c *gateway.Context) error {
	req := new(requests.ListPublicKeys)

	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	if err := req.Filters.Unmarshal(); err != nil {
		log.WithError(err).WithField("filter", req.Filters.Raw).Warn("failed to decode public keys list filter")

		return c.NoContent(http.StatusBadRequest)
	}

	if err := query.ValidateFilters(&req.Filters, services.PublicKeyFilterFields); err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	req.Paginator.Normalize()

	list, count, err := h.service.ListPublicKeys(c.Ctx(), req)
	if err != nil {
		return err
	}

	c.Response().Header().Set("X-Total-Count", strconv.Itoa(count))

	return c.JSON(http.StatusOK, list)
}

// CreatePublicKey adds a public key, with the device and username rules restricting it.
func (h *Handler) CreatePublicKey(c *gateway.Context) error {
	var req requests.PublicKeyCreate
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	var tenant string
	if c.Tenant() != nil {
		tenant = c.Tenant().ID
		req.TenantID = tenant
	}

	res, err := h.service.CreatePublicKey(c.Ctx(), req, tenant)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// UpdatePublicKey changes a key's name or its restriction rules, not the key material.
func (h *Handler) UpdatePublicKey(c *gateway.Context) error {
	var req requests.PublicKeyUpdate
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

	res, err := h.service.UpdatePublicKey(c.Ctx(), req.Fingerprint, tenant, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, res)
}

// DeletePublicKey revokes a key by fingerprint.
func (h *Handler) DeletePublicKey(c *gateway.Context) error {
	var req requests.PublicKeyDelete
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

	if err := h.service.DeletePublicKey(c.Ctx(), req.Fingerprint, tenant); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
