package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
)

const (
	GetStatsURL = "/stats"
)

func (h *Handler) GetStats(c gateway.Context) error {
	stats, err := h.service.GetStats(c.Ctx())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, stats)
}
