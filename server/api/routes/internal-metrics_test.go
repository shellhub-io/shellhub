package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInternalMetricsNeedsNoCredential pins both decisions behind this
// endpoint: it is registered unconditionally, and it answers without one. An
// operator reaches it during an incident, from inside the deployment, with no
// token to hand; the gateway is what keeps it off the internet.
func TestInternalMetricsNeedsNoCredential(t *testing.T) {
	router, _, _ := authenticatedRouter(t)

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, InternalMetricsURL, nil))

	require.Equal(t, http.StatusOK, rec.Code)

	assert.Contains(t, rec.Body.String(), "go_goroutines")
	assert.Contains(t, rec.Body.String(), "process_resident_memory_bytes")
}
