package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/api/apicontext"
)

const (
	GetStatsURL = "/stats"
)

func GetStats(c apicontext.Context) error {
	stats, err := c.Store().GetStats(c.Ctx())
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, stats)
}
