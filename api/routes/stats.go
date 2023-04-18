package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
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
	host := c.Request().Header.Get("X-Forwarded-Host")
	if host == "" {
		host = c.Request().Host
	}

	var port int
	if v := c.Request().Header.Get("X-Forwarded-Port"); v != "" {
		var err error
		port, err = strconv.Atoi(v)
		if err != nil {
			return err
		}
	}

	info, err := h.service.SystemGetInfo(c.Ctx(), host, port)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, info)
}

func (h *Handler) GetSystemDownloadInstallScript(c gateway.Context) error {
	c.Response().Writer.Header().Add("Content-Type", "text/x-shellscript")

	var req struct {
		Host                string `header:"X-Forwarded-Host"`
		Scheme              string `header:"X-Forwarded-Proto"`
		ForwardedPort       string `header:"X-Forwarded-Port"`
		TenantID            string `query:"tenant_id"`
		KeepAliveInternavel string `query:"keepalive_interval"`
		PreferredHostname   string `query:"preferred_hostname"`
		PreferredIdentity   string `query:"preferred_identity"`
	}

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
