package main

import (
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnterpriseAddsMembersDirectly adds members through the enterprise member route, which makes
// a user who already has an account a member at once and invites only an email with none.
func TestEnterpriseAddsMembersDirectly(t *testing.T) {
	t.Parallel()

	compose := newEnterpriseEnvironment(t, t.Context(), environment.New(t, run))

	add := func(t *testing.T, email string) *resty.Response {
		t.Helper()

		resp, err := compose.R(t.Context()).
			SetBody(map[string]string{"email": email, "role": string(authorizer.RoleOperator)}).
			Post("/api/namespaces/" + ShellHubNamespace + "/members")
		require.NoError(t, err)

		return resp
	}

	invitations := func(t *testing.T) []models.MembershipInvitation {
		t.Helper()

		invitations := []models.MembershipInvitation{}
		resp, err := compose.R(t.Context()).SetResult(&invitations).Get("/api/namespaces/" + ShellHubNamespace + "/invitations")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		return invitations
	}

	const (
		existingUsername = "existing"
		existingEmail    = "existing@ossystems.com.br"
	)

	compose.NewUser(t, existingUsername, existingEmail, ShellHubPassword)

	t.Run("an existing user becomes a member with no invitation", func(t *testing.T) {
		resp := add(t, existingEmail)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		existing := signInto(t, compose, existingUsername, ShellHubPassword)
		assert.Equal(t, authorizer.RoleOperator, memberRoles(t, compose)[existing.ID])
		require.NotNil(t, existing.Role)
		assert.Equal(t, authorizer.RoleOperator, *existing.Role)
		assert.Empty(t, invitations(t))
	})

	t.Run("adding the member again answers conflict", func(t *testing.T) {
		resp := add(t, existingEmail)
		assert.Equal(t, http.StatusConflict, resp.StatusCode(), resp.String())
	})

	t.Run("an email with no account is invited", func(t *testing.T) {
		resp := add(t, "newcomer@ossystems.com.br")
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		pending := invitations(t)
		require.Len(t, pending, 1)
		assert.Equal(t, models.MembershipInvitationStatusPending, pending[0].Status)
	})
}
