package routes

import (
	"context"
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
)

// The instance information routes, relative to the API's base path.
const (
	GetStatsURL                       = "/stats"
	GetSystemInfoURL                  = "/info"
	GetSystemDownloadInstallScriptURL = "/install"
)

// GetStats serves the device and session counts of the namespace the request is scoped to. It
// returns the store's error when the counts cannot be read.
func (h *Handler) GetStats(ctx context.Context, sc scope.Scope, _ gateway.Actor, _ *struct{}) (*models.Stats, error) {
	return h.service.GetStats(ctx, sc)
}

// GetSystemInfo serves what the instance is and how to reach it. It is reachable
// unauthenticated, so the UI can configure itself before anyone logs in.
func (h *Handler) GetSystemInfo(c *gateway.Context) error {
	info, err := h.service.GetSystemInfo(c.Ctx(), c.Origin())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, info)
}

// GetSystemDownloadInstallScript serves the agent install script, with this instance's
// address and the caller's provisioning key already filled in.
func (h *Handler) GetSystemDownloadInstallScript(c *gateway.Context) error {
	req := new(requests.SystemInstallScript)
	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	c.Response().Header().Add("Content-Type", "text/x-shellscript")

	data, err := h.service.SystemDownloadInstallScript(c.Ctx(), req, c.Origin())
	if err != nil {
		return err
	}

	return c.String(http.StatusOK, data)
}
