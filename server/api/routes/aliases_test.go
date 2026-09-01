package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/server/api/pkg/responses"
	serviceMocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestRootAliasesReachCanonicalRoutes pins the URLs ShellHub publishes at the
// root of its domain. The console boots by calling /info and /healthcheck, and
// /install.sh is an address already in the wild, so an alias that stops
// resolving is a broken console or a broken installer, not a degraded one.
func TestRootAliasesReachCanonicalRoutes(t *testing.T) {
	tests := []struct {
		description string
		target      string
		mock        func(*serviceMocks.MockService)
	}{
		{
			description: "healthcheck",
			target:      "/healthcheck",
		},
		{
			description: "system info",
			target:      "/info",
			mock: func(service *serviceMocks.MockService) {
				service.
					On("GetSystemInfo", mock.Anything, mock.Anything).
					Return(&responses.SystemInfo{}, nil).
					Once()
			},
		},
		{
			description: "install script",
			target:      "/install.sh",
			mock: func(service *serviceMocks.MockService) {
				service.
					On("SystemDownloadInstallScript", mock.Anything, mock.Anything).
					Return("#!/bin/sh", nil).
					Once()
			},
		},
		{
			description: "the install script's former name",
			target:      "/kickstart.sh",
			mock: func(service *serviceMocks.MockService) {
				service.
					On("SystemDownloadInstallScript", mock.Anything, mock.Anything).
					Return("#!/bin/sh", nil).
					Once()
			},
		},
		{
			description: "install script keeps its query string",
			target:      "/install.sh?tenant_id=00000000-0000-4000-0000-000000000000&preferred_hostname=host&preferred_identity=00:11:22:33:44:55",
			mock: func(service *serviceMocks.MockService) {
				service.
					On("SystemDownloadInstallScript", mock.Anything, &requests.SystemInstallScript{
						Host:              "example.com",
						TenantID:          "00000000-0000-4000-0000-000000000000",
						PreferredHostname: "host",
						PreferredIdentity: "00:11:22:33:44:55",
					}).
					Return("#!/bin/sh", nil).
					Once()
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			router, _, service := authenticatedRouter(t)
			if tc.mock != nil {
				tc.mock(service)
			}

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, tc.target, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusOK, rec.Code)
		})
	}
}

// TestRootAliasesDoNotMatchAsSuffix guards the ^ anchor on every rule: Echo
// compiles a rewrite key into an unanchored regex, so "/info" alone would also
// rewrite any request whose path merely ends in /info -- including one a device
// serves through a web endpoint tunnel.
func TestRootAliasesDoNotMatchAsSuffix(t *testing.T) {
	for _, target := range []string{"/api/devices/info", "/tunnel/healthcheck", "/downloads/install.sh"} {
		t.Run(target, func(t *testing.T) {
			router, _, _ := authenticatedRouter(t)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.NotEqual(t, http.StatusOK, rec.Code)
		})
	}
}

// TestInstallScriptRejectsMalformedQuery guards the anonymous installer route: its query
// parameters are rendered into a script the operator pipes into a root shell, so a value outside
// the shape the parameter is known to have never reaches the renderer (GHSA-w6jh-83pc-x59g). The
// service is left without expectations on purpose -- reaching it fails the test.
func TestInstallScriptRejectsMalformedQuery(t *testing.T) {
	for _, target := range []string{
		"/install.sh?tenant_id=x$(id)",
		"/install.sh?tenant_id=not-a-uuid",
		"/install.sh?preferred_hostname=host`id`",
		"/install.sh?preferred_identity=a\"%0Aid%0A%23",
	} {
		t.Run(target, func(t *testing.T) {
			router, _, _ := authenticatedRouter(t)

			req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
	}
}
