package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const licenseNotAuthorizedConnection = "Connection blocked: this ShellHub instance's license does not authorize this service"

// TestCloudLicensing runs a cloud instance, which needs a license that grants cloud authorization,
// and reads what an installed license refuses once it no longer grants it: connecting to a device
// and accepting one. A case stores its license straight into the store, because the admin API on
// Cloud refuses one without the grant, and leaves a cloud license behind for the next.
func TestCloudLicensing(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	compose := environment.New(t, run).WithEdition(environment.EditionCloud).WithLicense(cloudLicense()).Up(ctx)
	t.Cleanup(compose.Down)

	compose.NewUser(t, ShellHubUsername, ShellHubUsername+stripeTestEmailDomain, ShellHubPassword)
	ownNamespace(t, ctx, compose, models.SSHAccessModeLegacy)

	signer := registerDeviceKey(t, ctx, compose)
	_, device := startAcceptedAgent(t, ctx, compose)

	t.Run("connecting under a cloud license succeeds", func(t *testing.T) {
		conn := dialDevice(t, ctx, compose, device, signer)
		defer conn.Close() //nolint:errcheck // the test is over once the command answered

		assert.Equal(t, "licensed\n", runOnDevice(t, conn, "echo licensed"))
	})

	t.Run("connecting under a license that does not authorize cloud is refused", func(t *testing.T) {
		underLicense(t, compose, environment.FullLicense(), func() {
			requireAccessDenied(t, compose, deviceSSHID(device), signer, licenseNotAuthorizedConnection)
		})
	})

	t.Run("accepting a device under a license that does not authorize cloud is refused", func(t *testing.T) {
		pending := enrollPendingDevice(t, compose, "cloud-unauthorized", "02:00:00:00:22:01")

		underLicense(t, compose, environment.FullLicense(), func() {
			resp, err := compose.PatchDeviceStatus(t.Context(), pending.UID, environment.DeviceActionAccept)
			require.NoError(t, err)
			require.Equal(t, http.StatusPaymentRequired, resp.StatusCode(), resp.String())
		})

		assert.Equal(t, models.DeviceStatusPending, requireDevice(t, compose, pending.UID).Status)
	})
}

func underLicense(t *testing.T, compose *environment.DockerCompose, license environment.License, during func()) {
	t.Helper()

	compose.StoreLicense(t, compose.SignLicense(t, license))
	defer compose.StoreLicense(t, compose.SignLicense(t, cloudLicense()))

	during()
}

func cloudLicense() environment.License {
	return fullLicenseWhere(func(license *environment.License) {
		license.Features.Cloud = true
	})
}
