package routes

import (
	"net/http"
	"strconv"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
)

const (
	GetStatsURL      = "/stats"
	GetSystemInfoURL = "/info"
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
