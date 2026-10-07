package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var unknownSessionRecording = "/api/sessions/" + strings.Repeat("0", 64) + "/records/0"

type gatedRoute struct {
	feature  string
	path     string
	withhold func(features *environment.LicenseFeatures)
	unlocked int
}

var gatedRoutes = []gatedRoute{
	{
		feature:  "firewall",
		path:     "/api/firewall/rules",
		withhold: func(features *environment.LicenseFeatures) { features.FirewallRules = false },
		unlocked: http.StatusOK,
	},
	{
		feature:  "session recording",
		path:     unknownSessionRecording,
		withhold: func(features *environment.LicenseFeatures) { features.SessionRecording = false },
		unlocked: http.StatusNotFound,
	},
}

func testLicenseFeatureGating(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	for _, route := range gatedRoutes {
		t.Run(route.feature+" routes answer 402 without the feature", func(t *testing.T) {
			useLicense(t, compose, fullLicenseWhere(func(license *environment.License) {
				route.withhold(license.Features)
			}))

			requireStatus(t, compose, route.path, http.StatusPaymentRequired)
		})

		t.Run(route.feature+" routes answer with the feature", func(t *testing.T) {
			useLicense(t, compose, environment.FullLicense())

			requireStatus(t, compose, route.path, route.unlocked)
		})
	}

	t.Run("an expired license answers 402 on every gated route", func(t *testing.T) {
		useLicense(t, compose, expiredLicense())

		for _, route := range gatedRoutes {
			requireStatus(t, compose, route.path, http.StatusPaymentRequired)
		}
	})

	t.Run("the license endpoint answers under an expired license", func(t *testing.T) {
		license := expiredLicense()
		useLicense(t, compose, license)

		installed := compose.InstalledLicense(t)
		assert.Equal(t, license.ID, installed.ID)
		assert.True(t, installed.Expired)

		requireStatus(t, compose, "/admin/api/users", http.StatusPaymentRequired)
	})

	t.Run("the license endpoint installs a license when none is installed", func(t *testing.T) {
		useNoLicense(t, compose)
		requireStatus(t, compose, "/admin/api/users", http.StatusPaymentRequired)

		license := environment.FullLicense()
		compose.UseLicense(t, license)

		assert.Equal(t, license.ID, compose.InstalledLicense(t).ID)
		requireStatus(t, compose, "/admin/api/users", http.StatusOK)
	})
}

func requireStatus(t *testing.T, compose *environment.DockerCompose, path string, status int) {
	t.Helper()

	resp, err := compose.R(t.Context()).Get(path)
	require.NoError(t, err)
	require.Equal(t, status, resp.StatusCode(), "GET %s: %s", path, resp.String())
}
