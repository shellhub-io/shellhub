package main

import (
	"net/http"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/pairingcode"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testDeviceCodeFormat(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("a code is read without its hyphen and in any case", func(t *testing.T) {
		pairing := startPairing(t, compose, newPairingRequest(t, "format-pairing", "02:00:00:00:92:01"))
		typed := typedLoosely(pairing.Code)

		status, resp, err := getPairingStatus(t.Context(), compose, typed)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		assert.Equal(t, models.DeviceStatusPending, status.Status)

		accepted := acceptPairing(t, compose.R(t.Context()), typed)
		assert.Equal(t, models.DeviceStatusAccepted, requireDevice(t, compose, accepted.UID).Status)

		device := authDevice(t, compose, newDeviceAuthRequest(t, "format-login", "02:00:00:00:92:02"))
		code := createLoginCode(t, compose, device.Token)

		assert.Equal(t, device.UID, requireLoginCodePreview(t, compose, typedLoosely(code.Code)).UID)
	})

	t.Run("a code with a character outside the alphabet is not found, even when the cache holds it", func(t *testing.T) {
		req := newPairingRequest(t, "format-outside", "02:00:00:00:92:03")
		pairing := startPairing(t, compose, req)
		require.Regexp(t, "^["+pairingcode.Alphabet+"]{8}$", pairing.Code)

		device := authDevice(t, compose, newDeviceAuthRequest(t, "format-outside-login", "02:00:00:00:92:04"))
		login := createLoginCode(t, compose, device.Token)

		for _, outside := range []string{"0", "O", "1", "I", "L", "U"} {
			mistyped := outside + pairing.Code[1:]
			compose.CopyCacheEntry(t, pairingCodeCacheKey(pairing.Code), pairingCodeCacheKey(mistyped))

			_, resp, err := postPairingAccept(compose.R(t.Context()), mistyped)
			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, resp.StatusCode(), "accepting %s: %s", mistyped, resp.String())

			mistyped = outside + login.Code[1:]
			compose.CopyCacheEntry(t, loginCodeCacheKey(login.Code), loginCodeCacheKey(mistyped))

			_, resp, err = resolveLoginCode(t.Context(), compose, mistyped)
			require.NoError(t, err)
			assert.Equal(t, http.StatusNotFound, resp.StatusCode(), "resolving %s: %s", mistyped, resp.String())
		}

		assert.NotContains(t, deviceNames(compose.ListDevices(t, models.DeviceStatusEmpty)), req.Hostname, "a code outside the alphabet enrolled a device")
	})
}

func typedLoosely(code string) string {
	return strings.ToLower(code[:4] + "-" + code[4:])
}
