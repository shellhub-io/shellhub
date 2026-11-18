package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCloudEmailChange runs a cloud instance, the only edition that mails a new address before it
// takes effect. Changing an email needs the current password; the old address is told as soon as
// the change is asked for; and the account keeps its old email, and its old recovery email, until
// the link mailed to the new one is used, once.
func TestCloudEmailChange(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	compose := environment.New(t, run).WithEdition(environment.EditionCloud).Up(ctx)

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

	t.Run("keeps the recovery email until the link mailed to the new one is used", func(t *testing.T) {
		recoveryEmail := username + "-recovery" + stripeTestEmailDomain

		resp, err := asBearer(t, compose, token).
			SetBody(map[string]string{"recovery_email": recoveryEmail, "current_password": ShellHubPassword}).
			Patch("/api/users")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		assert.Empty(t, compose.AuthUser(t, username, ShellHubPassword).RecoveryEmail)

		confirmMailedAddress(t, compose, recoveryEmail)
		assert.Equal(t, recoveryEmail, compose.AuthUser(t, username, ShellHubPassword).RecoveryEmail)
	})
}

// TestCloudAccountChangeAsksForMFA runs a cloud instance and checks that an account with MFA
// enabled changes its password or an address only with a code from its authenticator or a recovery
// code, on top of its password, and that a recovery code is used up by the change it allows.
func TestCloudAccountChangeAsksForMFA(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	compose := environment.New(t, run).WithEdition(environment.EditionCloud).Up(ctx)

	user := newMFAUser(t, compose, environment.EditionCloud, "account-change-mfa")
	newEmail := user.username + "-new" + stripeTestEmailDomain

	patchUser := func(body map[string]string, code string) (int, string) {
		body["current_password"] = ShellHubPassword
		if code != "" {
			body["mfa_code"] = code
		}

		resp, err := asBearer(t, compose, user.token).SetBody(body).Patch("/api/users")
		require.NoError(t, err)

		return resp.StatusCode(), resp.String()
	}

	changeEmail := func(email, code string) (int, string) {
		return patchUser(map[string]string{"email": email}, code)
	}

	assertRefused := func(t *testing.T, status int, body, reason string) {
		t.Helper()

		assert.Equal(t, http.StatusBadRequest, status, body)
		assert.Contains(t, body, `"mfa_code":"`+reason+`"`, body)
	}

	t.Run("refuses a change without a code", func(t *testing.T) {
		status, body := changeEmail(newEmail, "")
		assertRefused(t, status, body, "is required while MFA is enabled")
	})

	t.Run("refuses a change with a wrong code", func(t *testing.T) {
		wrong := "000000"
		if totp(user.totpKey) == wrong {
			wrong = "111111"
		}

		status, body := changeEmail(newEmail, wrong)
		assertRefused(t, status, body, "is invalid")
	})

	t.Run("accepts a change with the authenticator's code", func(t *testing.T) {
		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			status, body := changeEmail(newEmail, totp(user.totpKey))
			assert.Equal(tt, http.StatusOK, status, body)
		}, time.Minute, 5*time.Second)

		link, ok := compose.AwaitEmail(t, newEmail, "/confirm-email").Link("/confirm-email")
		require.True(t, ok, "no confirmation link in the email to %s", newEmail)
		assert.NotEmpty(t, link.Query().Get("token"))
		assert.Len(t, compose.Emails(t, newEmail, "/confirm-email"), 1, "a change refused for its MFA code mails no link")
	})

	t.Run("accepts a change with a recovery code once", func(t *testing.T) {
		otherEmail := user.username + "-other" + stripeTestEmailDomain

		status, body := changeEmail(otherEmail, user.recoveryCodes[0])
		require.Equal(t, http.StatusOK, status, body)

		status, body = changeEmail(otherEmail, user.recoveryCodes[0])
		assertRefused(t, status, body, "is invalid")
	})

	t.Run("refuses a password change without a code", func(t *testing.T) {
		status, body := patchUser(map[string]string{"password": "new-" + ShellHubPassword}, "")
		assertRefused(t, status, body, "is required while MFA is enabled")
	})

	t.Run("accepts a password change with the authenticator's code", func(t *testing.T) {
		require.EventuallyWithT(t, func(tt *assert.CollectT) {
			status, body := patchUser(map[string]string{"password": "new-" + ShellHubPassword}, totp(user.totpKey))
			assert.Equal(tt, http.StatusOK, status, body)
		}, time.Minute, 5*time.Second)
	})
}

func confirmMailedAddress(t *testing.T, compose *environment.DockerCompose, address string) {
	t.Helper()

	link, ok := compose.AwaitEmail(t, address, "/confirm-email").Link("/confirm-email")
	require.True(t, ok, "no confirmation link in the email to %s", address)

	resp, err := compose.Anonymous(t.Context()).
		SetBody(map[string]string{"token": link.Query().Get("token")}).
		Post("/api/user/" + link.Query().Get("id") + "/confirm_email")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
}
