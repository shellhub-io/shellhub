package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateNamespace creates namespaces through the enterprise API as a user the admin CLI made,
// who may own any number of them until an instance administrator caps that number with the admin
// API. Every refusal leaves the caller's namespaces as they were.
func TestCreateNamespace(t *testing.T) {
	t.Parallel()

	compose := newEnterpriseEnvironment(t, t.Context(), environment.New(t, run))

	const creatorUsername = "creator"
	compose.NewUser(t, creatorUsername, "creator@ossystems.com.br", ShellHubPassword)
	creator := compose.AuthUser(t, creatorUsername, ShellHubPassword)

	create := func(t *testing.T, name string) (*models.Namespace, *resty.Response) {
		t.Helper()

		created := new(models.Namespace)
		resp, err := asBearer(t, compose, creator.Token).
			SetBody(map[string]string{"name": name}).
			SetResult(created).
			Post("/api/namespaces")
		require.NoError(t, err)

		return created, resp
	}

	limitNamespaces := func(t *testing.T, limit int) {
		t.Helper()

		resp, err := compose.R(t.Context()).
			SetBody(map[string]any{"max_namespaces": limit}).
			Put("/admin/api/users/" + creator.ID)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	}

	owned := func(t *testing.T) []string {
		t.Helper()

		namespaces := []models.Namespace{}
		resp, err := asBearer(t, compose, creator.Token).SetResult(&namespaces).Get("/api/namespaces")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		names := make([]string, 0, len(namespaces))
		for _, namespace := range namespaces {
			names = append(names, namespace.Name)
		}

		return names
	}

	refusal := func(t *testing.T, resp *resty.Response) string {
		t.Helper()

		refused := new(responses.Error)
		require.NoError(t, json.Unmarshal(resp.Body(), refused), resp.String())

		return refused.Message
	}

	t.Run("creates a namespace the caller owns, named in lowercase", func(t *testing.T) {
		created, resp := create(t, "CreatedSpace")
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		assert.Equal(t, "createdspace", created.Name)
		assert.Equal(t, creator.ID, created.Owner)
		assert.Equal(t, []string{"createdspace"}, owned(t))
	})

	t.Run("a name another namespace holds answers conflict, whatever its case", func(t *testing.T) {
		_, resp := create(t, "TestSpace")
		assert.Equal(t, http.StatusConflict, resp.StatusCode(), resp.String())
		assert.Equal(t, []string{"createdspace"}, owned(t))
	})

	t.Run("a user who owns as many namespaces as their limit is refused another", func(t *testing.T) {
		limitNamespaces(t, 1)

		_, resp := create(t, "overlimitspace")
		assert.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())
		assert.Equal(t, "namespace limit reached", refusal(t, resp))
		assert.Equal(t, []string{"createdspace"}, owned(t))
	})

	t.Run("a user whose limit is zero may not create a namespace", func(t *testing.T) {
		limitNamespaces(t, 0)

		_, resp := create(t, "forbiddenspace")
		assert.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())
		assert.Equal(t, "namespace creation not permitted for user", refusal(t, resp))
		assert.Equal(t, []string{"createdspace"}, owned(t))
	})
}
