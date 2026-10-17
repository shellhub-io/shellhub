package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type vault struct {
	Data    string `json:"data"`
	Version int64  `json:"version"`
}

// TestEnterpriseVault covers the vault the server keeps for a user in a namespace. Each write of
// its data carries the version the writer last read, and the server refuses one based on a version
// another write has already replaced, so a session holding a stale copy cannot overwrite the keys
// another session saved.
func TestEnterpriseVault(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	compose := newEnterpriseEnvironment(t, ctx, environment.New(t, run))

	t.Run("a write based on a stale version is refused with 409 and the first write is kept", func(t *testing.T) {
		first := compose.AuthUser(t, ShellHubUsername, ShellHubPassword).Token
		second := compose.AuthUser(t, ShellHubUsername, ShellHubPassword).Token

		resp, err := asBearer(t, compose, first).
			SetBody(map[string]string{"meta": `{"salt":"c2FsdA=="}`}).
			Put("/api/vault/meta")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		read := readVault(t, compose, first).Version
		require.Equal(t, read, readVault(t, compose, second).Version, "the two sessions read different versions")

		firstData := `{"iv":"Zmlyc3Q=","ciphertext":"Zmlyc3Q="}`
		resp, err = asBearer(t, compose, first).
			SetBody(map[string]any{"data": firstData, "version": read}).
			Put("/api/vault/data")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		resp, err = asBearer(t, compose, second).
			SetBody(map[string]any{"data": `{"iv":"c2Vjb25k","ciphertext":"c2Vjb25k"}`, "version": read}).
			Put("/api/vault/data")
		require.NoError(t, err)
		assert.Equal(t, http.StatusConflict, resp.StatusCode(), resp.String())

		stored := readVault(t, compose, second)
		assert.Equal(t, firstData, stored.Data, "the stale write replaced the first writer's data")
		assert.Equal(t, read+1, stored.Version, "the stale write bumped the version")
	})
}

func readVault(t *testing.T, compose *environment.DockerCompose, token string) vault {
	t.Helper()

	stored := vault{}
	resp, err := asBearer(t, compose, token).SetResult(&stored).Get("/api/vault")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return stored
}
