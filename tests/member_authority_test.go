package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMemberAuthority drives the member routes as members the route lets through, whose role the
// service still refuses for the member they name: the owner, whom nobody outranks, and the caller
// themselves. Every refusal leaves the namespace's members as they were.
func TestMemberAuthority(t *testing.T) {
	t.Parallel()

	compose := newMembershipEnvironment(t)

	const seniorUsername = "senior"
	compose.NewUser(t, seniorUsername, "senior@ossystems.com.br", ShellHubPassword)
	compose.NewMember(t, seniorUsername, ShellHubNamespaceName, string(authorizer.RoleAdministrator))

	owner := signInto(t, compose, ShellHubUsername, ShellHubPassword)
	senior := signInto(t, compose, seniorUsername, ShellHubPassword)
	membersPath := "/api/namespaces/" + ShellHubNamespace + "/members"

	refusals := []struct {
		description string
		request     func(t *testing.T) (*resty.Response, error)
	}{
		{
			description: "the owner cannot leave their namespace",
			request: func(t *testing.T) (*resty.Response, error) {
				t.Helper()

				return asBearer(t, compose, owner.Token).Delete(membersPath)
			},
		},
		{
			description: "an administrator cannot change the owner's role",
			request: func(t *testing.T) (*resty.Response, error) {
				t.Helper()

				return asBearer(t, compose, senior.Token).
					SetBody(map[string]string{"role": string(authorizer.RoleObserver)}).
					Patch(membersPath + "/" + owner.ID)
			},
		},
		{
			description: "an administrator cannot remove the owner",
			request: func(t *testing.T) (*resty.Response, error) {
				t.Helper()

				return asBearer(t, compose, senior.Token).Delete(membersPath + "/" + owner.ID)
			},
		},
		{
			description: "an administrator cannot change their own role",
			request: func(t *testing.T) (*resty.Response, error) {
				t.Helper()

				return asBearer(t, compose, senior.Token).
					SetBody(map[string]string{"role": string(authorizer.RoleOperator)}).
					Patch(membersPath + "/" + senior.ID)
			},
		},
		{
			description: "an administrator cannot remove themselves",
			request: func(t *testing.T) (*resty.Response, error) {
				t.Helper()

				return asBearer(t, compose, senior.Token).Delete(membersPath + "/" + senior.ID)
			},
		},
	}

	for _, tc := range refusals {
		t.Run(tc.description, func(t *testing.T) {
			resp, err := tc.request(t)
			require.NoError(t, err)
			assert.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())

			assert.Equal(t, map[string]authorizer.Role{
				owner.ID:  authorizer.RoleOwner,
				senior.ID: authorizer.RoleAdministrator,
			}, memberRoles(t, compose))
		})
	}
}

// TestReinviteMember invites an email that has no account through the member route, which leaves
// the invitation pending until the invitee accepts it or it expires.
func TestReinviteMember(t *testing.T) {
	t.Parallel()

	compose := newMembershipEnvironment(t)

	const invitee = "reinvited@ossystems.com.br"

	resp := addMember(t, compose, invitee, authorizer.RoleObserver)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	first := onlyInvitation(t, compose)
	require.Equal(t, models.MembershipInvitationStatusPending, first.Status)

	t.Run("inviting again while the invitation is pending answers conflict", func(t *testing.T) {
		resp := addMember(t, compose, invitee, authorizer.RoleOperator)
		assert.Equal(t, http.StatusConflict, resp.StatusCode(), resp.String())
		assert.Equal(t, first, onlyInvitation(t, compose))
	})

	t.Run("inviting again once the invitation expired resends the same invitation", func(t *testing.T) {
		compose.ExpireMembershipInvitation(t, ShellHubNamespace)

		resp := addMember(t, compose, invitee, authorizer.RoleOperator)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		resent := onlyInvitation(t, compose)
		assert.True(t, first.CreatedAt.Equal(resent.CreatedAt), "the resent invitation was created at %s, not %s", resent.CreatedAt, first.CreatedAt)
		assert.Equal(t, models.MembershipInvitationStatusPending, resent.Status)
		assert.Equal(t, authorizer.RoleOperator, resent.Role)
		require.NotNil(t, resent.ExpiresAt)
		assert.WithinDuration(t, wallClock.Now().Add(invitationLifetime), *resent.ExpiresAt, 5*time.Minute)
	})
}

func newMembershipEnvironment(t *testing.T) *environment.DockerCompose {
	t.Helper()

	compose := environment.New(t, run).Up(t.Context())
	t.Cleanup(compose.Down)

	compose.NewUser(t, ShellHubUsername, ShellHubEmail, ShellHubPassword)
	compose.NewNamespace(t, ShellHubUsername, ShellHubNamespaceName, ShellHubNamespace, "")
	compose.JWT(compose.AuthUser(t, ShellHubUsername, ShellHubPassword).Token)

	return compose
}

func addMember(t *testing.T, compose *environment.DockerCompose, email string, role authorizer.Role) *resty.Response {
	t.Helper()

	resp, err := compose.R(t.Context()).
		SetBody(map[string]string{"email": email, "role": string(role)}).
		Post("/api/namespaces/" + ShellHubNamespace + "/members")
	require.NoError(t, err)

	return resp
}

func memberRoles(t *testing.T, compose *environment.DockerCompose) map[string]authorizer.Role {
	t.Helper()

	members := []models.MemberView{}
	resp, err := compose.R(t.Context()).SetResult(&members).Get("/api/namespaces/" + ShellHubNamespace + "/members")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	roles := make(map[string]authorizer.Role, len(members))
	for _, member := range members {
		roles[member.ID] = member.Role
	}

	return roles
}

func namespaceInvitations(t *testing.T, compose *environment.DockerCompose) []models.MembershipInvitation {
	t.Helper()

	invitations := []models.MembershipInvitation{}
	resp, err := compose.R(t.Context()).SetResult(&invitations).Get("/api/namespaces/" + ShellHubNamespace + "/invitations")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	return invitations
}

func onlyInvitation(t *testing.T, compose *environment.DockerCompose) models.MembershipInvitation {
	t.Helper()

	invitations := namespaceInvitations(t, compose)
	require.Len(t, invitations, 1)

	return invitations[0]
}
