package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/envs"
	serviceMocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
)

var editions = []envs.Edition{envs.Community, envs.Enterprise, envs.Cloud}

func TestEveryRouteStatesItsAccess(t *testing.T) {
	handler := NewHandler(serviceMocks.NewMockService(t))

	for _, edition := range editions {
		for _, r := range handler.publicRoutes(edition) {
			assert.NotEqual(t, accessUnstated, r.access.kind, "%s: %s %s states no access", edition, r.method, r.path)
		}
	}
}

func TestNoRouteIsMountedTwice(t *testing.T) {
	handler := NewHandler(serviceMocks.NewMockService(t))

	for _, edition := range editions {
		seen := make(map[string]bool)

		for _, r := range handler.publicRoutes(edition) {
			address := r.method + " " + r.path
			assert.False(t, seen[address], "%s: %s is mounted twice", edition, address)
			seen[address] = true
		}
	}
}

func TestEveryPublicRouteComesFromTheTable(t *testing.T) {
	for _, edition := range editions {
		service := serviceMocks.NewMockService(t)
		router := newTestRouter(t, service, WithEdition(edition))

		table := make(map[string]bool)
		for _, r := range NewHandler(service).publicRoutes(edition) {
			table[r.method+" "+publicAPIPrefix+r.path] = true
		}

		for _, mounted := range router.Router().Routes() {
			if !strings.HasPrefix(mounted.Path, publicAPIPrefix+"/") {
				continue
			}

			address := mounted.Method + " " + mounted.Path
			assert.True(t, table[address], "%s: %s is mounted outside the route table", edition, address)
		}
	}
}

func TestAuthenticationRoutesServeAnonymousCallers(t *testing.T) {
	router, _, _ := authenticatedRouter(t)

	for _, path := range []string{AuthDeviceURL, AuthDeviceURLV2, AuthLocalUserURL, AuthLocalUserURLV2} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, publicAPIPrefix+path, strings.NewReader("{}"))
			req.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.NotEqual(t, http.StatusUnauthorized, rec.Code)
		})
	}
}
