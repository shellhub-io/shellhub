package main

import (
	"encoding/json"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const invitationLifetime = 7 * 24 * time.Hour

// TestInvitationLink covers the link an invitation to a namespace travels by: it resolves for seven
// days, is refused once it expires, and generating it again then issues a fresh link while the old
// one stays refused. Each case invites its own namespace, so the namespace holds one invitation.
func TestInvitationLink(t *testing.T) {
	t.Parallel()

	compose := newSSHEnvironment(t, t.Context(), models.SSHAccessModeLegacy)

	t.Run("a link resolves for seven days", func(t *testing.T) {
		tenant := newInvitingNamespace(t, compose, "invitation-lifetime")
		link := generateInvitationLink(t, compose, tenant, "lifetime@ossystems.com.br")

		require.Equal(t, http.StatusOK, resolveInvitation(t, compose, link).StatusCode())

		invitations := []models.MembershipInvitation{}
		resp, err := compose.R(t.Context()).SetResult(&invitations).Get("/api/namespaces/" + tenant + "/invitations")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		require.Len(t, invitations, 1)
		require.NotNil(t, invitations[0].ExpiresAt)
		assert.WithinDuration(t, wallClock.Now().Add(invitationLifetime), *invitations[0].ExpiresAt, 5*time.Minute)
	})

	t.Run("an expired link is refused, and generating it again issues a fresh one", func(t *testing.T) {
		tenant := newInvitingNamespace(t, compose, "invitation-expired")
		invitee := "expired@ossystems.com.br"
		expired := generateInvitationLink(t, compose, tenant, invitee)

		require.Equal(t, http.StatusOK, resolveInvitation(t, compose, expired).StatusCode())
		compose.ExpireMembershipInvitation(t, tenant)
		assert.Equal(t, http.StatusForbidden, resolveInvitation(t, compose, expired).StatusCode())

		fresh := generateInvitationLink(t, compose, tenant, invitee)

		assert.NotEqual(t, expired, fresh)
		assert.Equal(t, http.StatusOK, resolveInvitation(t, compose, fresh).StatusCode())
		assert.Equal(t, http.StatusForbidden, resolveInvitation(t, compose, expired).StatusCode())
	})

	t.Run("a link still valid is not generated again", func(t *testing.T) {
		tenant := newInvitingNamespace(t, compose, "invitation-valid")
		invitee := "valid@ossystems.com.br"
		link := generateInvitationLink(t, compose, tenant, invitee)

		resp, err := postInvitationLink(t, compose, tenant, invitee)
		require.NoError(t, err)
		assert.Equal(t, http.StatusConflict, resp.StatusCode(), resp.String())
		assert.Equal(t, http.StatusOK, resolveInvitation(t, compose, link).StatusCode())
	})
}

func newInvitingNamespace(t *testing.T, compose *environment.DockerCompose, name string) string {
	t.Helper()

	tenant := uuid.Generate()
	compose.NewNamespace(t, ShellHubUsername, name, tenant, "")

	return tenant
}

func postInvitationLink(t *testing.T, compose *environment.DockerCompose, tenant, email string) (*resty.Response, error) {
	t.Helper()

	return compose.R(t.Context()).
		SetBody(map[string]string{"email": email, "role": "observer"}).
		Post("/api/namespaces/" + tenant + "/invitations/links")
}

func generateInvitationLink(t *testing.T, compose *environment.DockerCompose, tenant, email string) string {
	t.Helper()

	resp, err := postInvitationLink(t, compose, tenant, email)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

	generated := struct {
		Link string `json:"link"`
	}{}
	require.NoError(t, json.Unmarshal(resp.Body(), &generated))
	require.NotEmpty(t, generated.Link)

	return generated.Link
}

func resolveInvitation(t *testing.T, compose *environment.DockerCompose, link string) *resty.Response {
	t.Helper()

	parsed, err := url.Parse(link)
	require.NoError(t, err)

	invite := parsed.Query().Get("invite")
	require.NotEmpty(t, invite, "the link %s carries no invite code", link)

	resp, err := compose.Anonymous(t.Context()).SetQueryParam("invite", invite).Get("/api/invitations/resolve")
	require.NoError(t, err)

	return resp
}
