package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/stretchr/testify/require"
)

func TestFOO(t *testing.T) {
	routes := []struct {
		endpoint string
		methods  []string
	}{
		// api-key routes
		{
			endpoint: "/api/namespaces/:tenant/api-key",
			methods:  []string{http.MethodGet, http.MethodPost},
		},
		{
			endpoint: "/api/namespaces/:tenant/api-key/:key",
			methods:  []string{http.MethodPatch, http.MethodDelete},
		},
		// // auth routes
		{
			endpoint: "/internal/auth",
			methods:  []string{http.MethodGet},
		},
		{
			endpoint: "/internal/auth/token/:id",
			methods:  []string{http.MethodGet},
		},
		{
			endpoint: "/api/auth/token/:tenant",
			methods:  []string{http.MethodGet},
		},
		{
			endpoint: "/api/auth/user",
			methods:  []string{http.MethodGet},
		},
		{
			endpoint: "/api/auth/user",
			methods:  []string{http.MethodPost},
		},
		{
			endpoint: "/api/login",
			methods:  []string{http.MethodPost},
		},
		{
			endpoint: "/api/auth/ssh",
			methods:  []string{http.MethodPost},
		},
		{
			endpoint: "/api/auth/device",
			methods:  []string{http.MethodPost},
		},
		{
			endpoint: "/api/devices/auth",
			methods:  []string{http.MethodPost},
		},
		// // device routes
		{
			endpoint: "/api/devices",
			methods:  []string{http.MethodGet},
		},
		{
			endpoint: "/api/devices/:uid",
			methods:  []string{http.MethodGet},
		},
		{
			endpoint: "/internal/devices/public/:address",
			methods:  []string{http.MethodGet},
		},
		{
			endpoint: "/internal/lookup",
			methods:  []string{http.MethodGet},
		},
		{
			endpoint: "/api/devices/:uid",
			methods:  []string{http.MethodPatch},
		},
		{
			endpoint: "/api/devices/:uid/:status",
			methods:  []string{http.MethodPatch},
		},
		{
			endpoint: "/api/devices/:uid",
			methods:  []string{http.MethodPut},
		},
		{
			endpoint: "/api/devices/:uid",
			methods:  []string{http.MethodDelete},
		},
		{
			endpoint: "/api/devices/:uid/tags",
			methods:  []string{http.MethodPost},
		},
		{
			endpoint: "/api/devices/:uid/tags",
			methods:  []string{http.MethodPut},
		},
		{
			endpoint: "/api/devices/:uid/tags/:tag",
			methods:  []string{http.MethodDelete},
		},
		{
			endpoint: "/internal/devices/:uid/offline",
			methods:  []string{http.MethodPost},
		},
		{
			endpoint: "/internal/lookup",
			methods:  []string{http.MethodGet},
		},
		// // namespace routes
		{
			endpoint: "/api/namespaces",
			methods:  []string{http.MethodGet},
		},
		{
			endpoint: "/api/namespaces/:tenant",
			methods:  []string{http.MethodGet},
		},
		{
			endpoint: "/api/namespaces",
			methods:  []string{http.MethodPost},
		},
		{
			endpoint: "/api/namespaces/:tenant",
			methods:  []string{http.MethodPut},
		},
		{
			endpoint: "/api/namespaces/:tenant",
			methods:  []string{http.MethodDelete},
		},
		{
			endpoint: "/api/namespaces/:tenant/members",
			methods:  []string{http.MethodPost},
		},
		{
			endpoint: "/api/namespaces/:tenant/members/:uid",
			methods:  []string{http.MethodPatch},
		},
		{
			endpoint: "/api/namespaces/:tenant/members/:uid",
			methods:  []string{http.MethodDelete},
		},
		{
			endpoint: "/api/users/security",
			methods:  []string{http.MethodGet},
		},
		{
			endpoint: "/api/users/security/:tenant",
			methods:  []string{http.MethodPut},
		},
		// // session routes
		{
			endpoint: "/api/sessions",
			methods:  []string{http.MethodGet},
		},
		{
			endpoint: "/api/sessions/:uid",
			methods:  []string{http.MethodGet},
		},
		{
			endpoint: "/api/sessions/:uid/play",
			methods:  []string{http.MethodGet},
		},
		{
			endpoint: "/internal/sessions",
			methods:  []string{http.MethodPost},
		},
		{
			endpoint: "/internal/sessions/:uid/finish",
			methods:  []string{http.MethodPost},
		},
		{
			endpoint: "/internal/sessions/:uid/keepalive",
			methods:  []string{http.MethodPost},
		},
		{
			endpoint: "/internal/sessions/:uid/record",
			methods:  []string{http.MethodPost},
		},
		{
			endpoint: "/internal/sessions/:uid",
			methods:  []string{http.MethodPatch},
		},
		{
			endpoint: "/api/sessions/:uid/record",
			methods:  []string{http.MethodDelete},
		},
		// // sshkeys routes
		{
			endpoint: "/api/sshkeys/public-keys",
			methods:  []string{http.MethodGet},
		},
		{
			endpoint: "/api/sshkeys/public-keys",
			methods:  []string{http.MethodPost},
		},
		{
			endpoint: "/api/sshkeys/public-keys/:fingerprint",
			methods:  []string{http.MethodPut},
		},
		{
			endpoint: "/api/sshkeys/public-keys/:fingerprint",
			methods:  []string{http.MethodDelete},
		},
		{
			endpoint: "/internal/sshkeys/public-keys/:fingerprint/:tenant",
			methods:  []string{http.MethodGet},
		},
		{
			endpoint: "/internal/sshkeys/public-keys/evaluate/:fingerprint/:username",
			methods:  []string{http.MethodPost},
		},
		{
			endpoint: "/internal/sshkeys/private-keys",
			methods:  []string{http.MethodPost},
		},
		{
			endpoint: "/api/sshkeys/public-keys/:fingerprint/tags",
			methods:  []string{http.MethodPost},
		},
		{
			endpoint: "/api/sshkeys/public-keys/:fingerprint/tags/:tag",
			methods:  []string{http.MethodDelete},
		},
		{
			endpoint: "/api/sshkeys/public-keys/:fingerprint/tags",
			methods:  []string{http.MethodPut},
		},
		// // stats routes
		{
			endpoint: "/api/stats",
			methods:  []string{http.MethodGet},
		},
		{
			endpoint: "/api/info",
			methods:  []string{http.MethodGet},
		},
		{
			endpoint: "/api/install",
			methods:  []string{http.MethodGet},
		},
		// // tags routes
		{
			endpoint: "/api/tags",
			methods:  []string{http.MethodGet},
		},
		{
			endpoint: "/api/tags/:tag",
			methods:  []string{http.MethodPut},
		},
		{
			endpoint: "/api/tags/:tag",
			methods:  []string{http.MethodDelete},
		},
		// // user routes
		{
			endpoint: "/api/users/:id/data",
			methods:  []string{http.MethodPatch},
		},
		{
			endpoint: "/api/users/:id/password",
			methods:  []string{http.MethodPatch},
		},
	}

	svcMock := new(mocks.Service)

	for _, tc := range routes {
		t.Run(tc.endpoint, func(t *testing.T) {
			for _, m := range tc.methods {
				req := httptest.NewRequest(m, tc.endpoint, nil)
				rec := httptest.NewRecorder()

				e := NewRouter(svcMock)
				e.ServeHTTP(rec, req)

				require.NotEqual(t, 404, rec.Result().StatusCode)
				require.NotEqual(t, 405, rec.Result().StatusCode)
			}
		})
	}
}
