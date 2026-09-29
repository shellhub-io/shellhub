package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	errs "github.com/shellhub-io/shellhub/server/api/routes/errors"
)

// The device pairing routes. Create and status are on the anonymous allowlist (the code is
// the secret); accept requires a user token.
const (
	CreateDevicePairingURL    = "/devices/pairing"
	GetDevicePairingStatusURL = "/devices/pairing/:code/status"
	AcceptDevicePairingURL    = "/devices/pairing/:code/accept"
)

// CreateDevicePairing stores the identity payload of a tenant-less agent and
// returns a short-lived pairing code. No device is created until a user
// accepts the pairing into a namespace.
func (h *Handler) CreateDevicePairing(c *gateway.Context) error {
	req := new(requests.DevicePairingCreate)
	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	pairing, err := h.service.CreateDevicePairing(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, pairing)
}

// GetDevicePairingStatus reports the pairing outcome to the agent polling it.
func (h *Handler) GetDevicePairingStatus(c *gateway.Context) error {
	var req requests.DevicePairingStatus
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	status, err := h.service.GetDevicePairingStatus(c.Ctx(), req.Code)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, status)
}

// AcceptDevicePairing materializes the pairing as a device in the namespace
// chosen by the user. Authorization against the chosen namespace happens in
// the service, since the session may be scoped to another tenant.
func (h *Handler) AcceptDevicePairing(c *gateway.Context) error {
	req := new(requests.DevicePairingAccept)
	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	userID, ok := c.GetID()
	if !ok {
		return errs.NewErrUnauthorized(nil)
	}

	accepted, err := h.service.AcceptDevicePairing(c.Ctx(), userID, req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, accepted)
}
