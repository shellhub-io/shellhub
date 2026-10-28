package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCloudAccountConfirmation runs a cloud instance, the only edition with open sign-up, for the
// code a new account confirms its email with.
func TestCloudAccountConfirmation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	compose := environment.New(t, run).WithEdition(environment.EditionCloud).Up(ctx)
	t.Cleanup(compose.Down)

	t.Run("the emailed code confirms the account and signs it in", func(t *testing.T) {
		email := signUp(t, compose, "confirm-code")
		code := confirmationCode(t, compose.AwaitEmail(t, email, "/confirm-account"))

		auth := requireConfirmed(t, compose, "confirm-code", code)

		resp, err := asBearer(t, compose, auth.Token).Get("/api/auth/user")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	})

	t.Run("the email confirms an account signed in with its email", func(t *testing.T) {
		email := signUp(t, compose, "confirm-by-email")
		code := confirmationCode(t, compose.AwaitEmail(t, email, "/confirm-account"))

		requireConfirmed(t, compose, email, code)
	})

	t.Run("wrong codes lock a pending and an unknown identifier out alike, on the third", func(t *testing.T) {
		email := signUp(t, compose, "confirm-wrong")
		code := confirmationCode(t, compose.AwaitEmail(t, email, "/confirm-account"))

		wrong := "000000"
		if code == wrong {
			wrong = "111111"
		}

		for _, identifier := range []string{"confirm-wrong", "confirm-unknown"} {
			for _, status := range []int{http.StatusUnauthorized, http.StatusUnauthorized, http.StatusTooManyRequests} {
				resp := confirmAccount(t, compose, identifier, wrong, nil)
				require.Equal(t, status, resp.StatusCode(), "%s: %s", identifier, resp.String())
			}
		}
	})

	t.Run("a resend delivers the same code, once a minute", func(t *testing.T) {
		email := signUp(t, compose, "confirm-resend")
		first := confirmationCode(t, compose.AwaitEmail(t, email, "/confirm-account"))

		resp := resend(t, compose, "confirm-resend")
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		emails := compose.AwaitEmails(t, email, "/confirm-account", 2)
		assert.Equal(t, first, confirmationCode(t, emails[0]))

		resp = resend(t, compose, "confirm-resend")
		assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode(), resp.String())
	})

	t.Run("a resend answers alike for an unknown and a confirmed account, and sends nothing", func(t *testing.T) {
		done := "confirm-done" + stripeTestEmailDomain
		compose.NewUser(t, "confirm-done", done, ShellHubPassword)

		for _, identifier := range []string{"confirm-nobody", "confirm-done"} {
			resp := resend(t, compose, identifier)
			assert.Equal(t, http.StatusOK, resp.StatusCode(), "%s: %s", identifier, resp.String())
		}

		pending := signUp(t, compose, "confirm-flushed")
		compose.AwaitEmail(t, pending, "/confirm-account")
		assert.Empty(t, compose.Emails(t, done, "/confirm-account"))
	})

	t.Run("a code never confirms a confirmed account", func(t *testing.T) {
		compose.NewUser(t, "confirm-already", "confirm-already"+stripeTestEmailDomain, ShellHubPassword)

		resp := confirmAccount(t, compose, "confirm-already", "123456", nil)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode(), resp.String())
	})

	t.Run("signing in to a pending account checks the password first", func(t *testing.T) {
		signUp(t, compose, "confirm-login")

		for password, status := range map[string]int{"wrong-password": http.StatusUnauthorized, ShellHubPassword: http.StatusForbidden} {
			resp, err := compose.Anonymous(t.Context()).
				SetBody(map[string]string{"username": "confirm-login", "password": password}).
				Post("/api/login")
			require.NoError(t, err)
			assert.Equal(t, status, resp.StatusCode(), "%s: %s", password, resp.String())
		}
	})
}

func signUp(t *testing.T, compose *environment.DockerCompose, username string) string {
	t.Helper()

	email := username + stripeTestEmailDomain

	resp := signUpWithEmail(t, compose, username, email)
	require.Equal(t, http.StatusNoContent, resp.StatusCode(), resp.String())

	return email
}

func confirmationCode(t *testing.T, email environment.Email) string {
	t.Helper()

	code, ok := email.Labelled("Confirmation code")
	require.True(t, ok, "the email carries no confirmation code: %s", email.Text())
	require.Len(t, code, 6)

	return code
}

func confirmAccount(t *testing.T, compose *environment.DockerCompose, identifier, code string, result any) *resty.Response {
	t.Helper()

	req := compose.Anonymous(t.Context()).
		SetBody(map[string]string{"identifier": identifier, "code": code})
	if result != nil {
		req.SetResult(result)
	}

	resp, err := req.Post("/api/user/confirm_account")
	require.NoError(t, err)

	return resp
}

func requireConfirmed(t *testing.T, compose *environment.DockerCompose, identifier, code string) *models.UserAuthResponse {
	t.Helper()

	auth := new(models.UserAuthResponse)
	resp := confirmAccount(t, compose, identifier, code, auth)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
	require.NotEmpty(t, auth.Token)

	return auth
}

func resend(t *testing.T, compose *environment.DockerCompose, identifier string) *resty.Response {
	t.Helper()

	resp, err := compose.Anonymous(t.Context()).
		SetBody(map[string]string{"identifier": identifier}).
		Post("/api/user/resend_email")
	require.NoError(t, err)

	return resp
}
