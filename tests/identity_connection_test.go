package main

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

const deadIdentityLog = "failed to resolve the identity for the public key"

// TestSSHIdentityConnection covers what an enrolled identity does at connect time once it exists:
// it stops working when it expires or when its one session burns it, every connection stamps it,
// and it authenticates as the principal it is bound to and nobody else. Enrolling a key and the
// first connection with it are covered by [TestSSHIdentityMode] and [TestIdentityAccessPolicy].
func TestSSHIdentityConnection(t *testing.T) {
	ctx := context.Background()
	compose := newSSHEnvironment(t, ctx, models.SSHAccessModeIdentity)
	_, device := startAcceptedAgent(t, ctx, compose)

	sshid := deviceSSHID(device)

	t.Run("an expired identity is refused instead of being offered for enrollment again", func(t *testing.T) {
		signer, data := newSigner(t)
		fingerprint := ssh.FingerprintSHA256(signer.PublicKey())

		days := 1
		resp, err := compose.R(t.Context()).
			SetBody(&requests.SSHIdentityCreate{Name: "expiring", Data: data, ExpiresIn: &days}).
			Post("/api/ssh-identities")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode(), resp.String())

		requireStraightThrough(t, compose, sshid, signer)

		expireIdentityAndRequireRefused(t, compose, sshid, signer)

		identity := identityByFingerprint(t, compose, fingerprint)
		assert.False(t, identity.Active(time.Now()), //nolint:forbidigo // the expiry the server compares against its own wall clock
			"the identity should still be on record, and dead")
	})

	t.Run("a single-use identity is consumed by its first session", func(t *testing.T) {
		key, signer := newAPIKeyIdentity(t, compose, "oneshot", true)
		grant(t, compose, &requests.AccessPolicyCreate{
			Name:    "oneshot",
			Subject: apiKeySubject(key.ID),
			Logins:  []string{"*"},
		})

		fresh := apiKeyIdentity(t, compose, "oneshot")
		require.True(t, fresh.SingleUse)
		require.Nil(t, fresh.ConsumedAt)

		requireStraightThrough(t, compose, sshid, signer)

		requireRecent(t, apiKeyIdentity(t, compose, "oneshot").ConsumedAt)
	})

	t.Run("a consumed single-use identity is refused", func(t *testing.T) {
		key, signer := newAPIKeyIdentity(t, compose, "burnt", true)
		grant(t, compose, &requests.AccessPolicyCreate{
			Name:    "burnt",
			Subject: apiKeySubject(key.ID),
			Logins:  []string{"*"},
		})

		requireStraightThrough(t, compose, sshid, signer)
		require.NotNil(t, apiKeyIdentity(t, compose, "burnt").ConsumedAt)

		mark := compose.ServerLogMark(t)

		requireRefusedAtAuth(t, compose, sshid, signer)

		compose.AwaitServerLogLine(t, mark, deadIdentityLog, `error="ssh access denied by policy"`)
	})

	t.Run("every connection stamps the identity's last use", func(t *testing.T) {
		signer, fingerprint := enrollOwnerKey(t, compose)
		require.Nil(t, identityByFingerprint(t, compose, fingerprint).LastUsedAt)

		requireStraightThrough(t, compose, sshid, signer)
		requireRecent(t, identityByFingerprint(t, compose, fingerprint).LastUsedAt)

		compose.AgeSSHIdentityLastUse(t, fingerprint, 24*time.Hour)

		aged := identityByFingerprint(t, compose, fingerprint).LastUsedAt
		require.NotNil(t, aged)
		require.WithinDuration(t, time.Now().Add(-24*time.Hour), *aged, time.Minute) //nolint:forbidigo // the stamp was aged with the database's wall clock

		requireStraightThrough(t, compose, sshid, signer)
		requireRecent(t, identityByFingerprint(t, compose, fingerprint).LastUsedAt)
	})

	t.Run("a member's key authenticates as that member, not as the owner a policy grants", func(t *testing.T) {
		member := newMember(t, compose, "crossing", authorizer.RoleOperator)

		memberSigner, data := newSigner(t)
		compose.EnrollIdentityAs(t, member.Token, "member", data)

		ownerSigner, _ := enrollOwnerKey(t, compose)
		requireStraightThrough(t, compose, sshid, ownerSigner)

		mark := compose.ServerLogMark(t)

		requireRefusedAtAuth(t, compose, sshid, memberSigner)

		compose.AwaitServerLogLine(t, mark, "reason="+string(models.ReasonNoGrant), "user="+member.ID)

		identity := identityByFingerprint(t, compose, ssh.FingerprintSHA256(memberSigner.PublicKey()))
		assert.Equal(t, member.ID, identity.PrincipalID)
	})
}
