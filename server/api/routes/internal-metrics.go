package routes

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	routesmiddleware "github.com/shellhub-io/shellhub/server/api/routes/middleware"
)

// InternalMetricsURL is where the process publishes its Prometheus registry.
//
// It sits under /internal rather than beside [WithMetrics]'s /metrics because
// the gateway answers 404 to /internal* deliberately. /metrics is unreachable
// from outside only because no proxy rule happens to match it today.
const InternalMetricsURL = "/internal/metrics"

func registerInternalMetrics(router *echo.Echo, authn *routesmiddleware.Authenticator) {
	router.GET(InternalMetricsURL, echo.WrapHandler(promhttp.Handler()))

	if authn != nil {
		authn.AllowAnonymous(http.MethodGet, InternalMetricsURL)
	}
}
