package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/scope"
	"github.com/shellhub-io/shellhub/server/api/pkg/gateway"
	"github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEvaluateHealth(t *testing.T) {
	mock := mocks.NewMockService(t)
	h := NewHandler(mock, nil)

	require.NoError(t, h.EvaluateHealth(t.Context(), scope.NewUnbounded("test"), gateway.Actor{}, &requests.Empty{}))

	mock.AssertExpectations(t)
}

// TestHealthCheckAnswersWithoutACredential drives the registration rather than the handler: the
// health check is the route that declares both an unbounded scope and an anonymous actor, so it is
// where those two claims are proven to reach the wire.
func TestHealthCheckAnswersWithoutACredential(t *testing.T) {
	router, _, _ := authenticatedRouter(t)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api"+HealthCheckURL, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Body.String())
}
