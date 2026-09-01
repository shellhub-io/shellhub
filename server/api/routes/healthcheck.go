package routes

import (
	"net/http"

	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
)

// The health check route, relative to the API's base path.
const (
	HealthCheckURL = "/healthcheck"
)

// EvaluateHealth answers that the API is serving. It checks nothing behind the API, so it
// reports reachability rather than readiness.
func (h *Handler) EvaluateHealth(c *gateway.Context) error {
	return c.NoContent(http.StatusOK)
}
