package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
)

const (
	GetStatsURL                       = "/stats"
	GetSystemInfoURL                  = "/info"
	GetSystemDownloadInstallScriptURL = "/install"
)

func (h *Handler) GetStats(c gateway.Context) error {
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

func (h *Handler) GetSystemInfo(c gateway.Context) error {
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

func (h *Handler) GetSystemDownloadInstallScript(c gateway.Context) error {
	req := new(requests.SystemInstallScript)
	if err := c.Bind(req); err != nil {
		return err
	}

	// Fall back to the request Host when the gateway didn't forward it, so the
	// served script can default SERVER_ADDRESS to the instance's own address.
	if req.Host == "" {
		req.Host = c.Request().Host
	}

	c.Response().Writer.Header().Add("Content-Type", "text/x-shellscript")

	data, err := h.service.SystemDownloadInstallScript(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.String(http.StatusOK, data)
}
