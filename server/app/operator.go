package app

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const operatorMetricsPath = "/metrics"

func metricsAddress(configured string) string {
	if configured == "" {
		return defaultMetricsAddress
	}

	return configured
}

func operatorHandler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET "+operatorMetricsPath, promhttp.Handler())

	return mux
}
