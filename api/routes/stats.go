package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
)

const (
	GetStatsURL = "/stats"
)

// @summary Get ShellHub stats
// @description returns data about the ShellHub instance like how many devices are registed, online, pending, rejected and how many active sessions are connected.
// @tags community
// @Security jwt
// @produce json
// @success 200 {object} models.Stats
// @failure 401 {object} nil
// @failure 500 {object} nil
// @router /stats [get]
func (h *Handler) GetStats(c gateway.Context) error {
	stats, err := h.service.GetStats(c.Ctx())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, stats)
}
