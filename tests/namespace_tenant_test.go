package main

import (
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/require"
)

func TestNamespaceTenantIsServerMinted(t *testing.T) {
	t.Parallel()

	compose := newEnterpriseEnvironment(t, t.Context(), environment.New(t, run))

	compose.DeleteNamespace(t, ShellHubNamespaceName)

	const successorUsername = "successor"
	compose.NewUser(t, successorUsername, "successor@ossystems.com.br", ShellHubPassword)
	successor := compose.AuthUser(t, successorUsername, ShellHubPassword)

	created := new(models.Namespace)
	resp, err := asBearer(t, compose, successor.Token).
		SetBody(map[string]string{"name": "successor", "tenant": ShellHubNamespace}).
		SetResult(created).
		Post("/api/namespaces")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	require.NotEqual(t, ShellHubNamespace, created.TenantID, "the namespace took the tenant its creator asked for")

	t.Run("a device configured with the deleted tenant does not enroll in the new namespace", func(t *testing.T) {
		compose.JWT(signIntoTenant(t, compose, successorUsername, ShellHubPassword, created.TenantID).Token)
		requireAuthAnswers(t, compose, newDeviceAuthRequest(t, "former-tenant", "02:00:00:00:04:01"), http.StatusNotFound)
	})
}
