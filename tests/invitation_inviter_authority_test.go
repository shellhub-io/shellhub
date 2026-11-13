package main

import (
	"net/http"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInvitationsOutlivingTheirInviter replays an administrator who issues administrator
// invitations, one for an existing account and one for an email with none, and is then removed.
// Removal cancels both, so neither the accept nor the registration admits anyone, while the
// owner's own invitation still admits its invitee.
func TestInvitationsOutlivingTheirInviter(t *testing.T) {
	t.Parallel()

	compose := newMembershipEnvironment(t)

	const departingUsername = "departing"
	compose.NewUser(t, departingUsername, "departing@ossystems.com.br", ShellHubPassword)
	compose.NewMember(t, departingUsername, ShellHubNamespaceName, string(authorizer.RoleAdministrator))

	const altUsername = "alt"
	altEmail := altUsername + "@ossystems.com.br"
	compose.NewUser(t, altUsername, altEmail, ShellHubPassword)

	owner := signInto(t, compose, ShellHubUsername, ShellHubPassword)
	departing := signInto(t, compose, departingUsername, ShellHubPassword)

	invite := func(t *testing.T, token, email string) string {
		t.Helper()

		generated := struct {
			Link string `json:"link"`
		}{}
		resp, err := asBearer(t, compose, token).
			SetBody(map[string]string{"email": email, "role": string(authorizer.RoleAdministrator)}).
			SetResult(&generated).
			Post("/api/namespaces/" + ShellHubNamespace + "/invitations/links")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		return inviteCode(t, generated.Link)
	}

	register := func(t *testing.T, username, code string) int {
		t.Helper()

		resp, err := compose.Anonymous(t.Context()).
			SetBody(map[string]string{"name": username, "username": username, "password": ShellHubPassword, "sig": code}).
			Post("/api/register")
		require.NoError(t, err)

		return resp.StatusCode()
	}

	altCode := invite(t, departing.Token, altEmail)
	sockCode := invite(t, departing.Token, "sock@ossystems.com.br")
	keptCode := invite(t, owner.Token, "kept@ossystems.com.br")

	resp, err := compose.R(t.Context()).Delete("/api/namespaces/" + ShellHubNamespace + "/members/" + departing.ID)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	invitations := namespaceInvitations(t, compose)
	require.Len(t, invitations, 3)
	for _, invitation := range invitations {
		want := models.MembershipInvitationStatusCancelled
		if invitation.InvitedBy == owner.ID {
			want = models.MembershipInvitationStatusPending
		}

		assert.Equal(t, want, invitation.Status, "the invitation %s sent", invitation.InvitedBy)
	}

	assert.Equal(t, http.StatusForbidden, register(t, "sock", sockCode))

	alt := compose.AuthUser(t, altUsername, ShellHubPassword)
	resp, err = asBearer(t, compose, alt.Token).
		SetBody(map[string]string{"code": altCode}).
		Patch("/api/namespaces/" + ShellHubNamespace + "/invitations/accept")
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode(), resp.String())

	assert.Equal(t, http.StatusOK, register(t, "kept", keptCode))

	roles := memberRoles(t, compose)
	assert.Len(t, roles, 2)
	assert.Equal(t, authorizer.RoleOwner, roles[owner.ID])
	assert.NotContains(t, roles, alt.ID)
	assert.NotContains(t, roles, departing.ID)
}
