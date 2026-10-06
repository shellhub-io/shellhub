package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testDeviceLoginCode(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("a pending device creates a code that resolves to it", func(t *testing.T) {
		device := authDevice(t, compose, newDeviceAuthRequest(t, "login-code", "02:00:00:00:91:01"))
		require.Equal(t, models.DeviceStatusPending, device.Status)

		code := createLoginCode(t, compose, device.Token)
		assert.Equal(t, int((10 * time.Minute).Seconds()), code.ExpiresIn)

		preview := requireLoginCodePreview(t, compose, code.Code)
		assert.Equal(t, models.DeviceLoginCodeKindDevice, preview.Kind)
		assert.Equal(t, device.UID, preview.UID)
		assert.Equal(t, "login-code", preview.Name)
		assert.Equal(t, ShellHubNamespace, preview.TenantID)
		assert.Equal(t, ShellHubNamespaceName, preview.Namespace)
		assert.Equal(t, models.DeviceStatusPending, preview.Status)
	})

	t.Run("a new code invalidates the one before it", func(t *testing.T) {
		device := authDevice(t, compose, newDeviceAuthRequest(t, "login-code-again", "02:00:00:00:91:02"))

		previous := createLoginCode(t, compose, device.Token)
		current := createLoginCode(t, compose, device.Token)
		require.NotEqual(t, previous.Code, current.Code)

		_, resp, err := resolveLoginCode(t.Context(), compose, previous.Code)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode(), resp.String())

		assert.Equal(t, device.UID, requireLoginCodePreview(t, compose, current.Code).UID)
	})

	t.Run("a code past its lifetime is not found", func(t *testing.T) {
		device := authDevice(t, compose, newDeviceAuthRequest(t, "login-code-expired", "02:00:00:00:91:03"))
		code := createLoginCode(t, compose, device.Token)

		compose.ExpireCacheEntryIn(t, loginCodeCacheKey(code.Code), codeLifetimeLeft)

		require.Equal(t, device.UID, requireLoginCodePreview(t, compose, code.Code).UID, "the code was gone before its lifetime ran out")

		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			_, resp, err := resolveLoginCode(t.Context(), compose, code.Code)
			if assert.NoError(tt, err) {
				assert.Equal(tt, http.StatusNotFound, resp.StatusCode(), resp.String())
			}
		}, codeLifetimeLeft+10*time.Second, time.Second)
	})

	t.Run("the device polls its own status until it is accepted", func(t *testing.T) {
		device := authDevice(t, compose, newDeviceAuthRequest(t, "login-code-polled", "02:00:00:00:91:04"))

		assert.Equal(t, models.DeviceStatusPending, requireDeviceAuthStatus(t, compose, device.Token))

		compose.UpdateDeviceStatus(t, device.UID, environment.DeviceActionAccept)

		assert.Equal(t, models.DeviceStatusAccepted, requireDeviceAuthStatus(t, compose, device.Token))
	})

	t.Run("a rejected device reads its rejection", func(t *testing.T) {
		device := authDevice(t, compose, newDeviceAuthRequest(t, "login-code-rejected", "02:00:00:00:91:05"))

		compose.UpdateDeviceStatus(t, device.UID, environment.DeviceActionReject)

		assert.Equal(t, models.DeviceStatusRejected, requireDeviceAuthStatus(t, compose, device.Token))
	})
}

func loginCodeCacheKey(code string) string {
	return "login_code/" + code
}

func createLoginCode(t *testing.T, compose *environment.DockerCompose, deviceToken string) *models.DeviceLoginCode {
	t.Helper()

	code := new(models.DeviceLoginCode)

	resp, err := asBearer(t, compose, deviceToken).SetResult(code).Post("/api/devices/auth/code")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	require.NotEmpty(t, code.Code)

	return code
}

func resolveLoginCode(ctx context.Context, compose *environment.DockerCompose, code string) (*models.DeviceLoginCodePreview, *resty.Response, error) {
	preview := new(models.DeviceLoginCodePreview)

	resp, err := compose.R(ctx).SetResult(preview).Get("/api/devices/login-code/" + code)

	return preview, resp, err
}

func requireLoginCodePreview(t *testing.T, compose *environment.DockerCompose, code string) *models.DeviceLoginCodePreview {
	t.Helper()

	preview, resp, err := resolveLoginCode(t.Context(), compose, code)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return preview
}

func requireDeviceAuthStatus(t *testing.T, compose *environment.DockerCompose, deviceToken string) models.DeviceStatus {
	t.Helper()

	status := new(models.DeviceAuthStatus)

	resp, err := asBearer(t, compose, deviceToken).SetResult(status).Get("/api/devices/auth/status")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return status.Status
}
