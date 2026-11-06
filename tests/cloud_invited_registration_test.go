package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCloudInvitedRegistration runs a cloud instance, which mails both the invitation and the
// account confirmation, for an account registered from an invitation link.
func TestCloudInvitedRegistration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	compose := environment.New(t, run).WithEdition(environment.EditionCloud).Up(ctx)
	t.Cleanup(compose.Down)

	inviter := "inviter"
	compose.NewUser(t, inviter, inviter+stripeTestEmailDomain, ShellHubPassword)
	tenant := uuid.Generate()
	compose.NewNamespace(t, inviter, inviter, tenant)
	token := compose.AuthUser(t, inviter, ShellHubPassword).Token

	t.Run("an account registered from an invitation link joins, but signs in only after the emailed confirmation", func(t *testing.T) {
		username := "invitee"
		email := username + stripeTestEmailDomain

		generated := struct {
			Link string `json:"link"`
		}{}
		resp, err := asBearer(t, compose, token).
			SetBody(map[string]string{"email": email, "role": "observer"}).
			SetResult(&generated).
			Post("/api/namespaces/" + tenant + "/invitations/links")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		resp, err = compose.Anonymous(t.Context()).
			SetBody(map[string]string{
				"name":     username,
				"username": username,
				"password": ShellHubPassword,
				"sig":      inviteCode(t, generated.Link),
			}).
			Post("/api/register")
		require.NoError(t, err)
		require.Equal(t, http.StatusNoContent, resp.StatusCode(), resp.String())

		resp, err = compose.Anonymous(t.Context()).
			SetBody(map[string]string{"username": username, "password": ShellHubPassword}).
			Post("/api/login")
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())

		code := confirmationCode(t, compose.AwaitEmail(t, email, "/confirm-account"))
		auth := requireConfirmed(t, compose, username, code)
		require.NotNil(t, auth.Tenant)
		assert.Equal(t, tenant, *auth.Tenant)
	})
}
