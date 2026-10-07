package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnterpriseLicensing runs an enterprise instance under licenses the run's issuer signs and
// reads what each license lets through: the license file the server loads on startup, the routes
// a license gates, accepting and connecting to devices, and the regions a license allows. The
// cases share one stack whose GeoIP database locates every address in Brazil, so a case changes
// the installed license and leaves the full one behind for the next.
func TestEnterpriseLicensing(t *testing.T) {
	ctx := context.Background()

	startup := environment.FullLicense()
	compose := newEnterpriseEnvironment(t, ctx, environment.New(t, run).WithLicense(startup).WithEveryAddressIn("BR"))

	t.Run("license file", func(t *testing.T) { testLicenseFile(t, compose, startup) })
	t.Run("signature", func(t *testing.T) { testLicenseSignature(t, compose) })
	t.Run("feature gating", func(t *testing.T) { testLicenseFeatureGating(t, compose) })
	t.Run("device acceptance", func(t *testing.T) { testLicensedDeviceAcceptance(t, compose) })
	t.Run("device connection", func(t *testing.T) { testLicensedDeviceConnection(t, ctx, compose) })
}

// TestEnterpriseWithoutLicenseFile starts an enterprise instance with no license file configured:
// the server starts, stores no license, and so accepts no device until one is installed.
func TestEnterpriseWithoutLicenseFile(t *testing.T) {
	compose := newEnterpriseEnvironment(t, context.Background(), environment.New(t, run).WithoutLicense())

	assert.Zero(t, compose.LicenseRows(t))

	device := enrollPendingDevice(t, compose, "no-license-file", "02:00:00:00:21:e0")
	requireAcceptAnswers(t, compose, device, http.StatusPaymentRequired)
}

func newEnterpriseEnvironment(t *testing.T, ctx context.Context, configurator *environment.DockerComposeConfigurator) *environment.DockerCompose {
	t.Helper()

	compose := configurator.WithEdition(environment.EditionEnterprise).Up(ctx)
	t.Cleanup(compose.Down)

	compose.NewAdmin(t, ShellHubUsername, ShellHubEmail, ShellHubPassword)
	ownNamespace(t, ctx, compose, models.SSHAccessModeLegacy)

	return compose
}

func useLicense(t *testing.T, compose *environment.DockerCompose, license environment.License) {
	t.Helper()

	compose.UseLicense(t, license)
	t.Cleanup(func() { compose.UseLicense(t, environment.FullLicense()) })
}

func useNoLicense(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	compose.RemoveLicenses(t)
	t.Cleanup(func() { compose.UseLicense(t, environment.FullLicense()) })
}

func fullLicenseWhere(change func(license *environment.License)) environment.License {
	license := environment.FullLicense()
	change(&license)

	return license
}

func expiredLicense() environment.License {
	return fullLicenseWhere(func(license *environment.License) {
		license.ExpiresAt = clock.Now().Add(-time.Hour).Unix()
	})
}

func notYetStartedLicense() environment.License {
	return fullLicenseWhere(func(license *environment.License) {
		license.StartsAt = clock.Now().Add(24 * time.Hour).Unix()
	})
}

func licenseForDevices(devices int) environment.License {
	return fullLicenseWhere(func(license *environment.License) {
		license.Features.Devices = devices
	})
}

func acceptedDevices(t *testing.T, compose *environment.DockerCompose) int {
	t.Helper()

	namespace := new(responses.Namespace)

	resp, err := compose.R(t.Context()).SetResult(namespace).Get("/api/namespaces/" + ShellHubNamespace)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return int(namespace.DevicesAcceptedCount)
}

func testLicenseFile(t *testing.T, compose *environment.DockerCompose, startup environment.License) {
	t.Helper()

	t.Run("the server installs the license file on startup", func(t *testing.T) {
		installed := compose.InstalledLicense(t)

		assert.Equal(t, startup.ID, installed.ID)
		assert.False(t, installed.Expired)
		assert.Equal(t, 1, compose.LicenseRows(t))
	})

	t.Run("a restart with the same license file installs nothing", func(t *testing.T) {
		compose.RestartServer(t)

		assert.Equal(t, startup.ID, compose.InstalledLicense(t).ID)
		assert.Equal(t, 1, compose.LicenseRows(t))
	})

	t.Run("a restart with another license file replaces the installed license", func(t *testing.T) {
		replacement := environment.FullLicense()
		compose.WriteLicenseFile(t, compose.SignLicense(t, replacement))

		compose.RestartServer(t)

		assert.Equal(t, replacement.ID, compose.InstalledLicense(t).ID)
		assert.Equal(t, 2, compose.LicenseRows(t))
	})

	t.Run("an invalid license file stops the server from starting", func(t *testing.T) {
		installed := compose.InstalledLicense(t)
		rows := compose.LicenseRows(t)
		mark := compose.ServerLogMark(t)

		compose.WriteLicenseFile(t, []byte("not a license"))
		t.Cleanup(func() {
			compose.WriteLicenseFile(t, compose.SignLicense(t, installed.License))
			compose.RestartServer(t)
		})

		compose.RestartFailingServer(t)

		compose.AwaitServerLogLine(t, mark, "LICENSE_FILE is set but failed to load")
		require.Equal(t, rows, compose.LicenseRows(t))
	})
}
