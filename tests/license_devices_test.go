package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const licenseBlockedConnection = "Connection blocked: your ShellHub instance has exceeded the maximum number of devices allowed by your license"

func testLicensedDeviceAcceptance(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	cases := []struct {
		description string
		license     func(t *testing.T) environment.License
		status      int
	}{
		{
			description: "accepting a device within the license limit succeeds",
			license: func(t *testing.T) environment.License {
				t.Helper()

				return licenseForDevices(acceptedDevices(t, compose) + 1)
			},
			status: http.StatusOK,
		},
		{
			description: "accepting a device with the fleet at the license limit is refused",
			license: func(t *testing.T) environment.License {
				t.Helper()

				return licenseForDevices(acceptedDevices(t, compose))
			},
			status: http.StatusPaymentRequired,
		},
		{
			description: "accepting a device under an unlimited license succeeds",
			license: func(t *testing.T) environment.License {
				t.Helper()

				return licenseForDevices(environment.UnlimitedDevices)
			},
			status: http.StatusOK,
		},
		{
			description: "accepting a device under an expired license is refused",
			license: func(t *testing.T) environment.License {
				t.Helper()

				return expiredLicense()
			},
			status: http.StatusPaymentRequired,
		},
		{
			description: "accepting a device under a license not yet started is refused",
			license: func(t *testing.T) environment.License {
				t.Helper()

				return notYetStartedLicense()
			},
			status: http.StatusPaymentRequired,
		},
	}

	for i, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			device := enrollPendingDevice(t, compose, fmt.Sprintf("licensed-%d", i), fmt.Sprintf("02:00:00:00:21:%02x", i))
			useLicense(t, compose, tc.license(t))

			requireAcceptAnswers(t, compose, device, tc.status)
		})
	}

	t.Run("accepting a device with no license installed is refused", func(t *testing.T) {
		device := enrollPendingDevice(t, compose, "unlicensed", "02:00:00:00:21:f0")
		useNoLicense(t, compose)

		requireAcceptAnswers(t, compose, device, http.StatusPaymentRequired)
	})

	t.Run("accepting a device succeeds when the license cannot be read", func(t *testing.T) {
		device := enrollPendingDevice(t, compose, "license-unreadable", "02:00:00:00:21:f1")
		useLicense(t, compose, licenseForDevices(acceptedDevices(t, compose)))
		compose.BreakLicenseStore(t)

		requireAcceptAnswers(t, compose, device, http.StatusOK)
	})
}

func enrollPendingDevice(t *testing.T, compose *environment.DockerCompose, hostname, mac string) models.Device {
	t.Helper()

	device := requireDevice(t, compose, authDevice(t, compose, newDeviceAuthRequest(t, hostname, mac)).UID)
	require.Equal(t, models.DeviceStatusPending, device.Status)
	t.Cleanup(func() { compose.DeleteDevice(t, device.UID) })

	return device
}

func requireAcceptAnswers(t *testing.T, compose *environment.DockerCompose, device models.Device, status int) {
	t.Helper()

	resp, err := compose.PatchDeviceStatus(t.Context(), device.UID, environment.DeviceActionAccept)
	require.NoError(t, err)
	require.Equal(t, status, resp.StatusCode(), resp.String())

	want := models.DeviceStatusAccepted
	if status != http.StatusOK {
		want = models.DeviceStatusPending
	}

	assert.Equal(t, want, requireDevice(t, compose, device.UID).Status)
}

func testLicensedDeviceConnection(t *testing.T, ctx context.Context, compose *environment.DockerCompose) {
	t.Helper()

	signer := registerDeviceKey(t, ctx, compose)
	_, device := startAcceptedAgent(t, ctx, compose)

	allowed := []struct {
		description string
		license     func(t *testing.T) environment.License
	}{
		{
			description: "connecting with the fleet within the license limit succeeds",
			license: func(t *testing.T) environment.License {
				t.Helper()

				return licenseForDevices(acceptedDevices(t, compose))
			},
		},
		{
			description: "connecting under an unlimited license succeeds",
			license: func(t *testing.T) environment.License {
				t.Helper()

				return licenseForDevices(environment.UnlimitedDevices)
			},
		},
		{
			description: "connecting under a license not yet started succeeds",
			license: func(t *testing.T) environment.License {
				t.Helper()

				return notYetStartedLicense()
			},
		},
		{
			description: "connecting from a region the license allows succeeds",
			license: func(t *testing.T) environment.License {
				t.Helper()

				return licenseInRegions("BR")
			},
		},
		{
			description: "connecting under a license that names no region succeeds from anywhere",
			license: func(t *testing.T) environment.License {
				t.Helper()

				return licenseInRegions()
			},
		},
	}

	for _, tc := range allowed {
		t.Run(tc.description, func(t *testing.T) {
			useLicense(t, compose, tc.license(t))

			conn := dialDevice(t, ctx, compose, device, signer)
			defer conn.Close() //nolint:errcheck // the test is over once the command answered

			assert.Equal(t, "licensed\n", runOnDevice(t, conn, "echo licensed"))
		})
	}

	refused := []struct {
		description string
		install     func(t *testing.T)
		reason      string
	}{
		{
			description: "connecting with the fleet over the license limit is refused",
			install: func(t *testing.T) {
				t.Helper()

				useLicense(t, compose, licenseForDevices(acceptedDevices(t, compose)-1))
			},
			reason: licenseBlockedConnection,
		},
		{
			description: "connecting with no license installed is refused",
			install: func(t *testing.T) {
				t.Helper()

				useNoLicense(t, compose)
			},
			reason: licenseBlockedConnection,
		},
		{
			description: "connecting under an expired license is refused",
			install: func(t *testing.T) {
				t.Helper()

				useLicense(t, compose, expiredLicense())
			},
			reason: licenseBlockedConnection,
		},
		{
			description: "connecting from a region the license does not allow is refused",
			install: func(t *testing.T) {
				t.Helper()

				useLicense(t, compose, licenseInRegions("US"))
			},
			reason: "destination device is blocked by a firewall rule",
		},
	}

	for _, tc := range refused {
		t.Run(tc.description, func(t *testing.T) {
			tc.install(t)

			requireAccessDenied(t, compose, deviceSSHID(device), signer, tc.reason)
		})
	}
}

func licenseInRegions(regions ...string) environment.License {
	return fullLicenseWhere(func(license *environment.License) {
		license.AllowedRegions = regions
	})
}
