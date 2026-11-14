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

// TestCloudEmailChange runs a cloud instance, the only edition that mails a new address before it
// takes effect. Changing an email needs the current password; the old address is told as soon as
// the change is asked for; and the account keeps its old email until the link mailed to the new
// one is used, once.
func TestCloudEmailChange(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	compose := environment.New(t, run).WithEdition(environment.EditionCloud).Up(ctx)
	t.Cleanup(compose.Down)

	username := "email-change-confirm"
	oldEmail := username + stripeTestEmailDomain
	newEmail := username + "-new" + stripeTestEmailDomain

	compose.NewUser(t, username, oldEmail, ShellHubPassword)
	compose.NewNamespace(t, username, username, uuid.Generate())
	token := compose.AuthUser(t, username, ShellHubPassword).Token

	t.Run("refuses an email change with a wrong current password", func(t *testing.T) {
		resp, err := asBearer(t, compose, token).
			SetBody(map[string]string{"email": newEmail, "current_password": "wrong-password"}).
			Patch("/api/users")
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())
	})

	t.Run("tells the old address at once and keeps it until the link mailed to the new one is used, once", func(t *testing.T) {
		resp, err := asBearer(t, compose, token).
			SetBody(map[string]string{"email": newEmail, "current_password": ShellHubPassword}).
			Patch("/api/users")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		assert.Equal(t, oldEmail, compose.AuthUser(t, username, ShellHubPassword).Email)

		notice := compose.AwaitEmail(t, oldEmail, "/forgot-password").Text()
		assert.Contains(t, notice, "asked to change the email address of your ShellHub account from "+oldEmail+" to "+newEmail, notice)

		link, ok := compose.AwaitEmail(t, newEmail, "/confirm-email").Link("/confirm-email")
		require.True(t, ok, "no confirmation link in the email to %s", newEmail)

		confirm := func() int {
			resp, err := compose.Anonymous(t.Context()).
				SetBody(map[string]string{"token": link.Query().Get("token")}).
				Post("/api/user/" + link.Query().Get("id") + "/confirm_email")
			require.NoError(t, err)

			return resp.StatusCode()
		}

		require.Equal(t, http.StatusOK, confirm())
		assert.Equal(t, newEmail, compose.AuthUser(t, username, ShellHubPassword).Email)
		assert.Equal(t, http.StatusUnauthorized, confirm(), "a confirmation link works once")
		assert.Len(t, compose.Emails(t, oldEmail, "/forgot-password"), 1, "confirming sends no second notice")
	})
}
