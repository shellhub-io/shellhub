package main

import (
	"net/http"
	"testing"

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

	const (
		existingUsername = "existing"
		existingEmail    = "existing@ossystems.com.br"
	)

	compose.NewUser(t, existingUsername, existingEmail, ShellHubPassword)

	t.Run("an existing user becomes a member with no invitation", func(t *testing.T) {
		resp := addMember(t, compose, existingEmail, authorizer.RoleOperator)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		existing := signInto(t, compose, existingUsername, ShellHubPassword)
		assert.Equal(t, authorizer.RoleOperator, memberRoles(t, compose)[existing.ID])
		require.NotNil(t, existing.Role)
		assert.Equal(t, authorizer.RoleOperator, *existing.Role)
		assert.Empty(t, namespaceInvitations(t, compose))
	})

	t.Run("adding the member again answers conflict", func(t *testing.T) {
		resp := addMember(t, compose, existingEmail, authorizer.RoleOperator)
		assert.Equal(t, http.StatusConflict, resp.StatusCode(), resp.String())
	})

	t.Run("an email with no account is invited", func(t *testing.T) {
		resp := addMember(t, compose, "newcomer@ossystems.com.br", authorizer.RoleOperator)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		pending := namespaceInvitations(t, compose)
		require.Len(t, pending, 1)
		assert.Equal(t, models.MembershipInvitationStatusPending, pending[0].Status)
	})
}
