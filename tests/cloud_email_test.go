package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCloudEmailPartials reads an email the cloud sends through Mailpit, for the header, footer and
// copyright every template pulls in from cloud/templates/partials with mj-include.
func TestCloudEmailPartials(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	compose := environment.New(t, run).WithEdition(environment.EditionCloud).Up(ctx)
	t.Cleanup(compose.Down)

	t.Run("the password reset email carries the shared header, footer and copyright", func(t *testing.T) {
		email := "email-partials" + stripeTestEmailDomain
		compose.NewUser(t, "email-partials", email, ShellHubPassword)

		resp, err := compose.Anonymous(t.Context()).
			SetBody(map[string]string{"username": "email-partials"}).
			Post("/api/user/recover_password")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		reset := compose.AwaitEmail(t, email, "/update-password")

		logo, ok := reset.Image("ShellHub")
		require.True(t, ok, "the header's logo is missing: %s", reset.Text())
		assert.Equal(t, "https://docs.shellhub.io/img/logo.png", logo)

		text := reset.Text()
		assert.Contains(t, text, "Reset your password")
		assert.Contains(t, text, "This link is valid until", "the footer is missing")
		assert.Contains(t, text, "If you were not expecting this email, please ignore it.", "the footer is missing")
		assert.Contains(t, text, "© ShellHub", "the copyright is missing")
	})
}
