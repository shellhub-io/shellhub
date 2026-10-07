package main

import (
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetup(t *testing.T) {
	compose := environment.New(t, run).Up(t.Context())
	t.Cleanup(compose.Down)

	setupDone := func(t *testing.T) bool {
		t.Helper()

		info := new(struct {
			Setup bool `json:"setup"`
		})
		resp, err := compose.Anonymous(t.Context()).SetResult(info).Get("/api/info")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		return info.Setup
	}

	setupRequest := func(username string) *requests.Setup {
		return &requests.Setup{
			Name:      username,
			Username:  username,
			Email:     username + "@ossystems.com.br",
			Password:  ShellHubPassword,
			Namespace: username,
		}
	}

	require.False(t, setupDone(t), "a fresh instance is not set up")

	auth := new(models.UserAuthResponse)
	resp, err := compose.Anonymous(t.Context()).SetBody(setupRequest("first")).SetResult(auth).Post("/api/setup")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	t.Run("signs the first user in as the administrator of their namespace", func(t *testing.T) {
		assert.Equal(t, "first", auth.User)
		assert.True(t, auth.Admin, "the first user is the instance's administrator")
		require.NotNil(t, auth.Tenant)

		namespace := new(models.Namespace)
		resp, err := asBearer(t, compose, auth.Token).SetResult(namespace).Get("/api/namespaces/" + *auth.Tenant)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		assert.Equal(t, "first", namespace.Name)

		assert.True(t, setupDone(t))
	})

	t.Run("a second setup is refused as already completed", func(t *testing.T) {
		resp, err := compose.Anonymous(t.Context()).SetBody(setupRequest("second")).Post("/api/setup")
		require.NoError(t, err)
		assert.Equal(t, http.StatusConflict, resp.StatusCode(), resp.String())

		resp, err = compose.Anonymous(t.Context()).
			SetBody(map[string]string{"username": "second", "password": ShellHubPassword}).
			Post("/api/login")
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode(), "the refused setup created no user: %s", resp.String())
	})
}
