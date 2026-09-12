package routes

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
)

// The health check route, relative to the API's base path.
const (
	HealthCheckURL = "/healthcheck"
)

// EvaluateHealth answers that the API is serving. It checks nothing behind the API, so it
// reports reachability rather than readiness.
func (h *Handler) EvaluateHealth(_ context.Context, _ scope.Scope, _ gateway.Actor, _ *requests.Empty) error {
	return nil
}
