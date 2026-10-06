package main

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSSHApprovalReauth covers the approval a login waits on when a policy asks an enrolled key to
// re-authenticate: proving the factor refreshes the identity's re-authentication, the endpoint that
// enrolls keys cannot stand in for it, and a lapsed window refuses the login without refreshing
// anything. When a policy asks for one is covered by [TestAccessPolicyReauth].
func TestSSHApprovalReauth(t *testing.T) {
	ctx := context.Background()
	compose := newSSHEnvironment(t, ctx, models.SSHAccessModeIdentity)
	_, device := startAcceptedAgent(t, ctx, compose)

	sshid := deviceSSHID(device)

	t.Run("re-authenticating refreshes the identity's last re-authentication", func(t *testing.T) {
		requireReauth(t, compose, "hourly", new(3600))
		signer, fingerprint := enrollOwnerKey(t, compose)

		compose.AgeSSHIdentityReauth(t, fingerprint, 24*time.Hour)
		stale := identityByFingerprint(t, compose, fingerprint).LastReauthAt
		require.NotNil(t, stale)

		parked := startLogin(t, compose, sshid, signer)
		prompt := parked.awaitApproval(t)
		require.Equal(t, models.SSHApprovalReauth, prompt.kind)

		parked.answer(reauthenticate(t, compose, fingerprint, prompt.code))
		require.NoError(t, parked.result(t))

		requireRecent(t, identityByFingerprint(t, compose, fingerprint).LastReauthAt)
		assert.Equal(t, models.SSHApprovalConfirmed, approvalState(t, compose, prompt.code))
	})

	t.Run("a re-authentication cannot be confirmed through the enrollment endpoint", func(t *testing.T) {
		requireReauth(t, compose, "always", nil)
		signer, fingerprint := enrollOwnerKey(t, compose)

		parked := startLogin(t, compose, sshid, signer)
		prompt := parked.awaitApproval(t)
		require.Equal(t, models.SSHApprovalReauth, prompt.kind)

		_, resp, err := confirmApprovalAs(t.Context(), compose, "", prompt.code, nil)
		require.NoError(t, err)
		assert.Equal(t, 403, resp.StatusCode(), resp.String())

		approval, resp, err := getApprovalAs(t.Context(), compose, "", prompt.code)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode(), resp.String())
		assert.Equal(t, models.SSHApprovalReauth, approval.Kind)
		assert.Equal(t, models.SSHApprovalPending, approval.State)
		assert.Nil(t, identityByFingerprint(t, compose, fingerprint).LastReauthAt)

		parked.answer(reauthenticate(t, compose, fingerprint, prompt.code))
		require.NoError(t, parked.result(t))
	})

	t.Run("a re-authentication past its window is refused and refreshes nothing", func(t *testing.T) {
		requireReauth(t, compose, "always", nil)
		signer, fingerprint := enrollOwnerKey(t, compose)

		parked := startLogin(t, compose, sshid, signer)
		prompt := parked.awaitApproval(t)
		require.Equal(t, models.SSHApprovalReauth, prompt.kind)

		compose.ExpireSSHApproval(t, prompt.code)

		_, resp, err := postReauth(t.Context(), compose, fingerprint, prompt.code)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode(), resp.String())
		assert.Nil(t, identityByFingerprint(t, compose, fingerprint).LastReauthAt,
			"a step-up that released nothing must not refresh the identity")

		parked.answer(strayConfirmationCode)

		require.Error(t, parked.result(t))
		assert.Equal(t, expiredReason, parked.denial())
	})
}
