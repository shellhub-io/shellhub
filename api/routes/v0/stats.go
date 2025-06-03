package v0

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
	c.Response().Writer.Header().Add("Content-Type", "text/x-shellscript")

	data, err := h.service.SystemDownloadInstallScript(c.Ctx())
	if err != nil {
		return err
	}

	return c.String(http.StatusOK, data)
}
