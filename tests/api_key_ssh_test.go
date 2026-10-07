package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAPIKeySSHAccess covers what decides an API key's SSH login in the identity mode: no policy
// naming the key lets it in, a deny naming it beats the allow that would, its identity works until
// its expiry, and deleting the key takes its identities with it. That an every-member policy does
// not grant a key is covered by [TestAccessPolicyEvaluation], single-use identities by
// [TestSSHIdentityConnection], and that a key is never asked to re-authenticate by
// [TestAccessPolicyReauth].
func TestAPIKeySSHAccess(t *testing.T) {
	ctx := context.Background()
	compose := newSSHEnvironment(t, ctx, models.SSHAccessModeIdentity)
	_, device := startAcceptedAgent(t, ctx, compose)

	sshid := deviceSSHID(device)

	t.Run("a key no policy names is refused", func(t *testing.T) {
		key, signer := newAPIKeyIdentity(t, compose, "ungranted", false)

		mark := compose.ServerLogMark(t)
		requireRefusedAtAuth(t, compose, sshid, signer)
		compose.AwaitServerLogLine(t, mark, "reason="+string(models.ReasonNoGrant), "user="+key.ID)
	})

	t.Run("a deny policy naming the key beats the allow that names it", func(t *testing.T) {
		key, signer := newAPIKeyIdentity(t, compose, "denied", false)

		grant(t, compose, &requests.AccessPolicyCreate{
			Name:    "allowed",
			Subject: apiKeySubject(key.ID),
			Logins:  []string{"*"},
		})

		requireStraightThrough(t, compose, sshid, signer)

		grant(t, compose, &requests.AccessPolicyCreate{
			Name:    "denied",
			Subject: apiKeySubject(key.ID),
			Logins:  []string{"*"},
			Action:  string(models.PolicyActionDeny),
		})

		mark := compose.ServerLogMark(t)
		requireRefusedAtAuth(t, compose, sshid, signer)
		compose.AwaitServerLogLine(t, mark, "reason="+string(models.ReasonDeniedByPolicy), "user="+key.ID)
	})

	t.Run("a key's identity is accepted until its expiry and refused after", func(t *testing.T) {
		key, signer := enrollAPIKeyIdentity(t, compose, "expiring", new(1), false)

		grant(t, compose, &requests.AccessPolicyCreate{
			Name:    "expiring",
			Subject: apiKeySubject(key.ID),
			Logins:  []string{"*"},
		})

		live := apiKeyIdentity(t, compose, key.Name)
		require.NotNil(t, live.ExpiresAt)
		require.True(t, live.Active(time.Now()), //nolint:forbidigo // the expiry the server compares against its own wall clock
			"an identity a day from its expiry should be live")

		requireStraightThrough(t, compose, sshid, signer)

		expireIdentityAndRequireRefused(t, compose, sshid, signer)
	})

	t.Run("deleting a key revokes its identities", func(t *testing.T) {
		key, signer := newAPIKeyIdentity(t, compose, "revoked", false)

		policy := compose.CreateAccessPolicy(t, &requests.AccessPolicyCreate{
			Name:    "revoked",
			Subject: apiKeySubject(key.ID),
			Logins:  []string{"*"},
		})

		requireStraightThrough(t, compose, sshid, signer)

		resp, err := compose.R(t.Context()).Delete("/api/namespaces/api-key/" + key.Name)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		resp, err = compose.R(t.Context()).Get("/api/access-policies/" + policy.ID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode(), "the policy naming the key should go with it: %s", resp.String())

		unknown := startLogin(t, compose, sshid, signer)
		prompt := unknown.awaitApproval(t)
		assert.Equal(t, models.SSHApprovalIdentity, prompt.kind,
			"the gateway should no longer know the key, and ask to enroll it as a new one")
	})
}
