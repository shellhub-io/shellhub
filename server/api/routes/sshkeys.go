package routes

import (
	"context"
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
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
func (h *Handler) GetPublicKeys(ctx context.Context, _ scope.Scope, _ gateway.Actor, req *requests.ListPublicKeys) ([]models.PublicKey, int, error) {
	return h.service.ListPublicKeys(ctx, req)
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
