package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/require"
)

const (
	leaverUsername = "leaver"
	leaverEmail    = "leaver@ossystems.com.br"
	leaverPassword = "password"

	leftBehindNamespaceName = "otherspace"
	leftBehindNamespace     = "00000000-0000-4000-0000-000000000001"

	neverLeftNamespaceName = "neverleftspace"
	neverLeftNamespace     = "00000000-0000-4000-0000-000000000002"
)

// TestLeaveNamespace covers both shapes the route answers. Leaving the namespace the caller's
// token is scoped to mints a replacement token, because the one they hold names a namespace they
// no longer belong to. Leaving any other namespace leaves that token valid, so there is nothing
// to send back and the route says so with its own status rather than an empty 200.
func TestLeaveNamespace(t *testing.T) {
	ctx := context.Background()

	compose := environment.New(t).Up(ctx)
	t.Cleanup(compose.Down)

	compose.NewUser(t, ShellHubUsername, ShellHubEmail, ShellHubPassword)
	compose.NewNamespace(t, ShellHubUsername, ShellHubNamespaceName, ShellHubNamespace, "")
	compose.NewNamespace(t, ShellHubUsername, leftBehindNamespaceName, leftBehindNamespace, "")
	compose.NewNamespace(t, ShellHubUsername, neverLeftNamespaceName, neverLeftNamespace, "")
	compose.NewUser(t, leaverUsername, leaverEmail, leaverPassword)
	compose.NewMember(t, leaverUsername, ShellHubNamespaceName, string(authorizer.RoleObserver))
	compose.NewMember(t, leaverUsername, leftBehindNamespaceName, string(authorizer.RoleObserver))
	compose.NewMember(t, leaverUsername, neverLeftNamespaceName, string(authorizer.RoleObserver))

	auth := compose.AuthUser(t, leaverUsername, leaverPassword)

	signedInto := models.UserAuthResponse{}

	resp, err := compose.Anonymous(ctx).
		SetAuthToken(auth.Token).
		SetResult(&signedInto).
		Get("/api/auth/token/" + ShellHubNamespace)

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.NotNil(t, signedInto.Tenant)
	require.Equal(t, ShellHubNamespace, *signedInto.Tenant)

	t.Run("leaving another namespace answers no content", func(t *testing.T) {
		resp, err := compose.Anonymous(ctx).
			SetAuthToken(signedInto.Token).
			Delete("/api/namespaces/" + leftBehindNamespace + "/members")

		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, resp.StatusCode())
		require.Empty(t, resp.Body())
	})

	t.Run("leaving the namespace the token names answers a replacement token", func(t *testing.T) {
		replacement := models.UserAuthResponse{}

		resp, err := compose.Anonymous(ctx).
			SetAuthToken(signedInto.Token).
			SetResult(&replacement).
			Delete("/api/namespaces/" + ShellHubNamespace + "/members")

		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode())
		require.NotEmpty(t, replacement.Token)
		require.NotEqual(t, signedInto.Token, replacement.Token)
		require.NotNil(t, replacement.Tenant)
		require.Equal(t, neverLeftNamespace, *replacement.Tenant)
	})
}
