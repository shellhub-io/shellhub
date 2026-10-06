package main

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// TestAccessPolicyReauth covers when a policy's require_reauth holds an enrolled key for a fresh
// re-authentication: every time without a period, only once the period lapses with one, never for
// an API key, and at the shortest period when several policies ask. What the re-authentication
// approval itself does is covered by [TestSSHApprovalReauth].
func TestAccessPolicyReauth(t *testing.T) {
	ctx := context.Background()
	compose := newSSHEnvironment(t, ctx, models.SSHAccessModeIdentity)
	_, device := startAcceptedAgent(t, ctx, compose)

	sshid := deviceSSHID(device)

	t.Run("a policy with no period asks on every login", func(t *testing.T) {
		requireReauth(t, compose, "always", nil)
		signer, fingerprint := enrollOwnerKey(t, compose)

		loginThroughReauth(t, compose, sshid, signer)
		requireRecent(t, identityByFingerprint(t, compose, fingerprint).LastReauthAt)

		loginThroughReauth(t, compose, sshid, signer)
	})

	t.Run("a policy with a period skips a fresh re-authentication and asks for a stale one", func(t *testing.T) {
		requireReauth(t, compose, "hourly", new(3600))
		signer, fingerprint := enrollOwnerKey(t, compose)

		loginThroughReauth(t, compose, sshid, signer)
		requireStraightThrough(t, compose, sshid, signer)

		compose.AgeSSHIdentityReauth(t, fingerprint, time.Hour+time.Second)

		loginThroughReauth(t, compose, sshid, signer)
	})

	t.Run("the period's window closes when the period elapses", func(t *testing.T) {
		const period = 60

		requireReauth(t, compose, "minutely", new(period))
		signer, fingerprint := enrollOwnerKey(t, compose)

		compose.AgeSSHIdentityReauth(t, fingerprint, (period-5)*time.Second)
		reauthedAt := identityByFingerprint(t, compose, fingerprint).LastReauthAt
		require.NotNil(t, reauthedAt)

		requireStraightThrough(t, compose, sshid, signer)

		time.Sleep(time.Until(reauthedAt.Add(period*time.Second + time.Second)))

		loginThroughReauth(t, compose, sshid, signer)
	})

	t.Run("an API key is never asked to re-authenticate", func(t *testing.T) {
		key, signer := newAPIKeyIdentity(t, compose, "unattended", false)
		grant(t, compose, &requests.AccessPolicyCreate{
			Name:          "unattended",
			Subject:       apiKeySubject(key.ID),
			Logins:        []string{"*"},
			RequireReauth: true,
		})

		requireStraightThrough(t, compose, sshid, signer)
		requireStraightThrough(t, compose, sshid, signer)

		require.Nil(t, apiKeyIdentity(t, compose, "unattended").LastReauthAt)
	})

	t.Run("the shorter of two periods decides", func(t *testing.T) {
		requireReauth(t, compose, "hourly", new(3600))
		requireReauth(t, compose, "minutely", new(60))
		signer, fingerprint := enrollOwnerKey(t, compose)

		compose.AgeSSHIdentityReauth(t, fingerprint, 30*time.Second)
		requireStraightThrough(t, compose, sshid, signer)

		compose.AgeSSHIdentityReauth(t, fingerprint, 2*time.Minute)
		loginThroughReauth(t, compose, sshid, signer)
	})

	t.Run("a policy with no period outweighs one with a period", func(t *testing.T) {
		requireReauth(t, compose, "hourly", new(3600))
		requireReauth(t, compose, "always", nil)
		signer, fingerprint := enrollOwnerKey(t, compose)

		compose.AgeSSHIdentityReauth(t, fingerprint, 30*time.Second)
		loginThroughReauth(t, compose, sshid, signer)
	})
}

func requireReauth(t *testing.T, compose *environment.DockerCompose, name string, period *int) {
	t.Helper()

	grant(t, compose, &requests.AccessPolicyCreate{
		Name:          name,
		Subject:       requests.AccessPolicySubject{Type: string(models.PolicySubjectAllMembers)},
		Logins:        []string{"*"},
		RequireReauth: true,
		ReauthPeriod:  period,
	})
}

func loginThroughReauth(t *testing.T, compose *environment.DockerCompose, sshid string, signer ssh.Signer) {
	t.Helper()

	parked := startLogin(t, compose, sshid, signer)

	prompt := parked.awaitApproval(t)
	require.Equal(t, models.SSHApprovalReauth, prompt.kind, "the login was held for something other than a re-authentication")

	parked.answer(reauthenticate(t, compose, ssh.FingerprintSHA256(signer.PublicKey()), prompt.code))

	require.NoError(t, parked.result(t), "the login should resume once the re-authentication is confirmed")
}
