package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOperatorHandlerServesTheRegistry(t *testing.T) {
	rec := httptest.NewRecorder()
	operatorHandler().ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/metrics", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "go_goroutines")
	assert.Contains(t, rec.Body.String(), "process_resident_memory_bytes")
}

func TestOperatorHandlerServesNothingElse(t *testing.T) {
	for _, path := range []string{"/", "/internal/metrics", "/api/info"} {
		rec := httptest.NewRecorder()
		operatorHandler().ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil))

		assert.Equal(t, http.StatusNotFound, rec.Code, path)
	}
}

func TestMetricsAddressDefaultsToLoopback(t *testing.T) {
	assert.Equal(t, "127.0.0.1:9091", metricsAddress(""))
	assert.Equal(t, ":9091", metricsAddress(":9091"))
}
