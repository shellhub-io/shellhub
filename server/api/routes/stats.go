package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
)

// The instance information routes, relative to the API's base path.
const (
	GetStatsURL                       = "/stats"
	GetSystemInfoURL                  = "/info"
	GetSystemDownloadInstallScriptURL = "/install"
)

// GetStats serves the namespace's device and session counts.
func (h *Handler) GetStats(c *gateway.Context) error {
	req := new(requests.GetStats)
	if err := c.Bind(req); err != nil {
		return err
	}

	if err := c.Validate(req); err != nil {
		return err
	}

	stats, err := h.service.GetStats(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, stats)
}

// GetSystemInfo serves what the instance is and how to reach it. It is reachable
// unauthenticated, so the UI can configure itself before anyone logs in.
func (h *Handler) GetSystemInfo(c *gateway.Context) error {
	req := new(requests.GetSystemInfo)
	if err := c.Bind(req); err != nil {
		return err
	}

	if req.Host == "" {
		req.Host = c.Request().Host
	}

	info, err := h.service.GetSystemInfo(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, info)
}

// GetSystemDownloadInstallScript serves the agent install script, with this instance's
// address and the caller's install key already filled in.
func (h *Handler) GetSystemDownloadInstallScript(c *gateway.Context) error {
	req := new(requests.SystemInstallScript)
	if err := c.Bind(req); err != nil {
		return err
	}

	if req.Host == "" {
		req.Host = c.Request().Host
	}

	c.Response().Header().Add("Content-Type", "text/x-shellscript")

	data, err := h.service.SystemDownloadInstallScript(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.String(http.StatusOK, data)
}
