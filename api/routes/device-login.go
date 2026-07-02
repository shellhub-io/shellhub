package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
)

const (
	CreateDeviceLoginCodeURL = "/devices/auth/code"
	GetDeviceAuthStatusURL   = "/devices/auth/status"
	// NOTE: This path must not start with /api/login, /api/devices/auth or any
	// other nginx prefix location that disables the auth subrequest; it relies
	// on the generic /api location forwarding the user claims.
	ResolveDeviceLoginCodeURL = "/devices/login-code/:code"
)

// CreateDeviceLoginCode issues a short-lived code that deep-links this device
// into the console's accept page. It is authenticated with the device's own
// token; the gateway sets X-Device-UID and X-Tenant-ID from the device claims.
func (h *Handler) CreateDeviceLoginCode(c gateway.Context) error {
	uid := c.DeviceUID()
	tenant := c.Tenant()
	if uid == "" || tenant == nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	code, err := h.service.CreateDeviceLoginCode(c.Ctx(), uid, tenant.ID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, code)
}

// GetDeviceAuthStatus reports the device's current status to the device
// itself, so the agent can poll while it waits for acceptance.
func (h *Handler) GetDeviceAuthStatus(c gateway.Context) error {
	uid := c.DeviceUID()
	if uid == "" {
		return c.NoContent(http.StatusUnauthorized)
	}

	device, err := h.service.GetDevice(c.Ctx(), models.UID(uid))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, &models.DeviceAuthStatus{Status: device.Status})
}

// ResolveDeviceLoginCode resolves a login code into a device preview for the
// accept page. It requires a user token; membership in the device's namespace
// is enforced by the service.
func (h *Handler) ResolveDeviceLoginCode(c gateway.Context) error {
	var req requests.DeviceLoginCodeResolve
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	userID, ok := c.GetID()
	if !ok {
		return c.NoContent(http.StatusUnauthorized)
	}

	preview, err := h.service.ResolveDeviceLoginCode(c.Ctx(), userID, req.Code)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, preview)
}
