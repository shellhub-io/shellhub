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

// registerInternalMetrics publishes the default registry, unconditionally.
//
// Unconditional and not behind METRICS, which ships false and was false on the
// host whose connection count could not be read. A flag an operator can trim
// out of an env file is one that turns out to be off during the incident that
// needs it.
//
// It costs nothing when nobody scrapes: the request histograms METRICS installs
// are middleware on every request, while this is a handler and a collector that
// reads counters the store already maintains. The default registry brings
// go_goroutines, go_memstats_* and process_resident_memory_bytes with it, none
// of which the server publishes today.
func registerInternalMetrics(router *echo.Echo, authn *routesmiddleware.Authenticator) {
	router.GET(InternalMetricsURL, echo.WrapHandler(promhttp.Handler()))

	if authn != nil {
		authn.AllowAnonymous(http.MethodGet, InternalMetricsURL)
	}
}
