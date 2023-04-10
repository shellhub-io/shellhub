package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/api/pkg/gateway"
)

const (
	HealthCheckURL = "/healthcheck"
)

func (h *Handler) EvaluateHealth(c gateway.Context) error {
	return c.NoContent(http.StatusOK)
}
