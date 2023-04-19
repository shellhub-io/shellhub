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
	stats, err := h.service.GetStats(c.Ctx())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, stats)
}

func (h *Handler) GetSystemInfo(c gateway.Context) error {
	var req requests.SystemGetInfo

	if err := c.Bind(&req); err != nil {
		return err
	}

	if req.Host == "" {
		req.Host = c.Request().Host
	}

	info, err := h.service.SystemGetInfo(c.Ctx(), req)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, info)
}

func (h *Handler) GetSystemDownloadInstallScript(c gateway.Context) error {
	c.Response().Writer.Header().Add("Content-Type", "text/x-shellscript")

	var req requests.SystemInstallScript

	if err := c.Bind(&req); err != nil {
		return err
	}

	if req.Host == "" {
		req.Host = c.Request().Host
	}

	if req.Scheme == "" {
		req.Scheme = "http"
	}

	if req.ForwardedPort != "" {
		req.Host = req.Host + ":" + req.ForwardedPort
	}

	tmpl, data, err := h.service.SystemDownloadInstallScript(c.Ctx(), req)
	if err != nil {
		return err
	}

	return tmpl.Execute(c.Response().Writer, data)
}
