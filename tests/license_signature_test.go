package main

import (
	"encoding/base64"
	"net/http"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testLicenseSignature(t *testing.T, compose *environment.DockerCompose) {
	t.Helper()

	t.Run("a license the issuer signed is installed", func(t *testing.T) {
		license := environment.FullLicense()

		resp, err := compose.PostLicense(t.Context(), compose.SignLicense(t, license))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		assert.Equal(t, license.ID, compose.InstalledLicense(t).ID)
	})

	cases := []struct {
		description string
		contents    func(t *testing.T) []byte
	}{
		{
			description: "a license whose body changed after signing is refused",
			contents: func(t *testing.T) []byte {
				t.Helper()

				signed := decodeLicense(t, compose.SignLicense(t, environment.FullLicense()))
				tampered := strings.Replace(string(signed), `"devices":-1`, `"devices":99`, 1)
				require.NotEqual(t, string(signed), tampered, "the license body holds no device limit to tamper with")

				return []byte(base64.StdEncoding.EncodeToString([]byte(tampered)))
			},
		},
		{
			description: "a license shorter than its signature is refused",
			contents: func(t *testing.T) []byte {
				t.Helper()

				signed := decodeLicense(t, compose.SignLicense(t, environment.FullLicense()))

				return []byte(base64.StdEncoding.EncodeToString(signed[:100]))
			},
		},
		{
			description: "a license that is not base64 is refused",
			contents: func(t *testing.T) []byte {
				t.Helper()

				return []byte("%%% not base64 %%%")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			installed := compose.InstalledLicense(t)
			rows := compose.LicenseRows(t)

			resp, err := compose.PostLicense(t.Context(), tc.contents(t))
			require.NoError(t, err)
			require.Equal(t, http.StatusBadRequest, resp.StatusCode(), resp.String())

			assert.Equal(t, installed.ID, compose.InstalledLicense(t).ID)
			assert.Equal(t, rows, compose.LicenseRows(t))
		})
	}
}

func decodeLicense(t *testing.T, signed []byte) []byte {
	t.Helper()

	decoded, err := base64.StdEncoding.DecodeString(string(signed))
	require.NoError(t, err)

	return decoded
}
