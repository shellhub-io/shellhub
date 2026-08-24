package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthCheckAnswersWithoutACredential(t *testing.T) {
	router, _, _ := authenticatedRouter(t)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api"+HealthCheckURL, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Body.String())
}
