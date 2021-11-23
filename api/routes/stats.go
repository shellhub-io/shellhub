package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/api/contexts"
)

const (
	GetStatsURL = "/stats"
)

func (h *Handler) GetStats(c contexts.EchoContext) error {
	stats, err := h.service.GetStats(c.Ctx())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, stats)
}
