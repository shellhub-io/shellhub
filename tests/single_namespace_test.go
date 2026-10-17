package main

import (
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSingleNamespaceInstance sets a community instance up through the API, which binds the
// instance to the namespace the setup creates. A bound instance holds that namespace alone, so an
// account belongs to it or to nothing.
func TestSingleNamespaceInstance(t *testing.T) {
	t.Parallel()

	compose := environment.New(t, run).Up(t.Context())
	t.Cleanup(compose.Down)

	const (
		boundUsername  = "bound"
		boundNamespace = "boundspace"
	)

	setup := new(models.UserAuthResponse)
	resp, err := compose.Anonymous(t.Context()).
		SetBody(&requests.Setup{
			Name:      boundUsername,
			Username:  boundUsername,
			Email:     boundUsername + "@ossystems.com.br",
			Password:  ShellHubPassword,
			Namespace: boundNamespace,
		}).
		SetResult(setup).
		Post("/api/setup")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	require.NotNil(t, setup.Tenant)

	tenant := *setup.Tenant

	t.Run("the API offers no route to create a namespace", func(t *testing.T) {
		resp, err := asBearer(t, compose, setup.Token).
			SetBody(map[string]string{"name": "secondspace"}).
			Post("/api/namespaces")
		require.NoError(t, err)
		assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode(), resp.String())
	})

	t.Run("the admin CLI refuses a second namespace", func(t *testing.T) {
		err := compose.CreateNamespace(t.Context(), boundUsername, "secondspace", uuid.Generate())
		assert.ErrorContains(t, err, "this instance does not support multi-tenancy")

		namespaces := []models.Namespace{}
		resp, err := asBearer(t, compose, setup.Token).SetResult(&namespaces).Get("/api/namespaces")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		require.Len(t, namespaces, 1)
		assert.Equal(t, boundNamespace, namespaces[0].Name)
	})

	t.Run("removing a member deletes their account, which belongs to no other namespace", func(t *testing.T) {
		const orphanUsername = "orphan"
		compose.NewUser(t, orphanUsername, "orphan@ossystems.com.br", ShellHubPassword)
		compose.NewMember(t, orphanUsername, boundNamespace, string(authorizer.RoleObserver))
		orphan := compose.AuthUser(t, orphanUsername, ShellHubPassword)

		resp, err := asBearer(t, compose, setup.Token).Delete("/api/namespaces/" + tenant + "/members/" + orphan.ID)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		resp, err = compose.Anonymous(t.Context()).
			SetBody(map[string]string{"username": orphanUsername, "password": ShellHubPassword}).
			Post("/api/login")
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode(), "the removed member's account still signs in: %s", resp.String())
	})
}
