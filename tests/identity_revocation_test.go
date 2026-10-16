package main

import (
	"context"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

// TestSSHIdentityRevocation covers revoking an enrolled identity through the API: who may revoke
// whose, and what the key that held it meets at its next login. Revoking an API key's identities
// by deleting the key is covered by [TestAPIKeySSHAccess].
func TestSSHIdentityRevocation(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	compose := newSSHEnvironment(t, ctx, newNamespace)
	_, device := startAcceptedAgent(t, ctx, compose)

	sshid := deviceSSHID(device)

	t.Run("revoking one's own identity needs the enroll permission", func(t *testing.T) {
		member, signer := memberWithKey(t, compose, "self-revoker", authorizer.RoleOperator)
		fingerprint := ssh.FingerprintSHA256(signer.PublicKey())
		identity := identityByFingerprint(t, compose, fingerprint)

		observer := reassignRoleAndSignIn(t, compose, member, authorizer.RoleObserver)

		resp, err := revokeIdentityAs(t.Context(), compose, observer.Token, identity.ID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())
		require.Len(t, identitiesHolding(t, compose, fingerprint), 1)

		operator := reassignRoleAndSignIn(t, compose, member, authorizer.RoleOperator)

		resp, err = revokeIdentityAs(t.Context(), compose, operator.Token, identity.ID)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		assert.Empty(t, identitiesHolding(t, compose, fingerprint))
	})

	t.Run("revoking another member's identity needs the manage permission", func(t *testing.T) {
		_, signer := memberWithKey(t, compose, "revokee", authorizer.RoleOperator)
		fingerprint := ssh.FingerprintSHA256(signer.PublicKey())
		identity := identityByFingerprint(t, compose, fingerprint)

		operator := newMember(t, compose, "peer-revoker", authorizer.RoleOperator)

		resp, err := revokeIdentityAs(t.Context(), compose, operator.Token, identity.ID)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode(), resp.String())
		require.Len(t, identitiesHolding(t, compose, fingerprint), 1)

		administrator := newMember(t, compose, "admin-revoker", authorizer.RoleAdministrator)

		resp, err = revokeIdentityAs(t.Context(), compose, administrator.Token, identity.ID)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())
		assert.Empty(t, identitiesHolding(t, compose, fingerprint))
	})

	t.Run("a revoked key is held for approval at its next login instead of being let back in", func(t *testing.T) {
		signer, fingerprint := enrollOwnerKey(t, compose)
		identity := identityByFingerprint(t, compose, fingerprint)

		requireStraightThrough(t, compose, sshid, signer)

		resp, err := revokeIdentityAs(t.Context(), compose, "", identity.ID)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode(), resp.String())

		held := startLogin(t, compose, sshid, signer)
		prompt := held.awaitApproval(t)
		assert.Equal(t, models.SSHApprovalIdentity, prompt.kind,
			"the gateway should no longer know the key, and ask to enroll it as a new one")
		assert.Equal(t, models.SSHApprovalPending, approvalState(t, compose, prompt.code))

		assert.Empty(t, identitiesHolding(t, compose, fingerprint),
			"a login held for approval must not enroll the key again on its own")
	})
}

func revokeIdentityAs(ctx context.Context, compose *environment.DockerCompose, token, id string) (*resty.Response, error) {
	return approvalRequest(ctx, compose, token).Delete("/api/ssh-identities/" + id)
}
