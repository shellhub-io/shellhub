package main

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

const strayConfirmationCode = "23456789"

// TestSSHApprovalIdentity covers the approval that turns an unknown key into an identity, past the
// confirmation that lets the login in: a rejection or a lapsed window refuses it, a confirmation
// can bound the key's life, a key approved twice binds to one person only, only a member allowed
// to approve can decide and only a member can see the request, and the code the terminal shows is
// the one the console shows. Holding an unknown key and confirming it are covered by
// [TestIdentityAccessPolicy].
func TestSSHApprovalIdentity(t *testing.T) {
	ctx := context.Background()
	compose := newSSHEnvironment(t, ctx, models.SSHAccessModeIdentity)
	_, device := startAcceptedAgent(t, ctx, compose)

	sshid := deviceSSHID(device)
	owner := compose.AuthUser(t, ShellHubUsername, ShellHubPassword)

	t.Run("a rejected approval refuses the login and enrolls nothing", func(t *testing.T) {
		signer, _ := newSigner(t)
		parked := startLogin(t, compose, sshid, signer)
		prompt := parked.awaitApproval(t)

		resp, err := rejectApprovalAs(t.Context(), compose, "", prompt.code)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode(), resp.String())

		parked.answer(strayConfirmationCode)

		require.Error(t, parked.result(t))
		assert.Equal(t, rejectedReason, parked.denial())
		assert.Equal(t, models.SSHApprovalRejected, approvalState(t, compose, prompt.code))
		assert.Empty(t, identitiesHolding(t, compose, ssh.FingerprintSHA256(signer.PublicKey())))
	})

	t.Run("an approval past its window can no longer be confirmed and refuses the login", func(t *testing.T) {
		signer, _ := newSigner(t)
		parked := startLogin(t, compose, sshid, signer)
		prompt := parked.awaitApproval(t)

		pending, resp, err := getApprovalAs(t.Context(), compose, "", prompt.code)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode(), resp.String())
		require.Equal(t, models.SSHApprovalPending, pending.State)
		require.Positive(t, pending.ExpiresIn)
		require.LessOrEqual(t, pending.ExpiresIn, 90)

		compose.ExpireSSHApproval(t, prompt.code)

		_, resp, err = confirmApprovalAs(t.Context(), compose, "", prompt.code, nil)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode(), resp.String())

		parked.answer(strayConfirmationCode)

		require.Error(t, parked.result(t))
		assert.Equal(t, expiredReason, parked.denial())
		assert.Empty(t, identitiesHolding(t, compose, ssh.FingerprintSHA256(signer.PublicKey())))
	})

	t.Run("a confirmation with an expiry enrolls a key that stops working when it expires", func(t *testing.T) {
		signer, _ := newSigner(t)
		fingerprint := ssh.FingerprintSHA256(signer.PublicKey())

		parked := startLogin(t, compose, sshid, signer)
		prompt := parked.awaitApproval(t)

		days := 1
		confirmation, resp, err := confirmApprovalAs(t.Context(), compose, "", prompt.code, &days)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode(), resp.String())

		parked.answer(confirmation.ConfirmationCode)
		require.NoError(t, parked.result(t))

		identity := identityByFingerprint(t, compose, fingerprint)
		assert.Equal(t, models.SSHIdentitySourceApproval, identity.Source)
		require.NotNil(t, identity.ExpiresAt)
		assert.WithinDuration(t, time.Now().Add(24*time.Hour), *identity.ExpiresAt, time.Minute) //nolint:forbidigo // the expiry the server set from its own wall clock

		compose.ExpireSSHIdentity(t, fingerprint)

		mark := compose.ServerLogMark(t)

		refused := startLogin(t, compose, sshid, signer)
		require.Error(t, refused.result(t))
		assert.Zero(t, refused.approvals(), "an expired key must not be sent to enrollment again")

		compose.AwaitServerLogLine(t, mark, deadIdentityLog, `error="ssh access denied by policy"`)
	})

	t.Run("approving the same key twice for the same person enrolls it once", func(t *testing.T) {
		signer, _ := newSigner(t)

		first := startLogin(t, compose, sshid, signer)
		second := startLogin(t, compose, sshid, signer)

		firstPrompt := first.awaitApproval(t)
		secondPrompt := second.awaitApproval(t)
		require.NotEqual(t, firstPrompt.code, secondPrompt.code)

		first.answer(confirmApproval(t, compose, firstPrompt.code))
		second.answer(confirmApproval(t, compose, secondPrompt.code))

		require.NoError(t, first.result(t))
		require.NoError(t, second.result(t))

		fingerprint := ssh.FingerprintSHA256(signer.PublicKey())
		assert.Len(t, identitiesHolding(t, compose, fingerprint), 1)
		assert.Equal(t, owner.ID, identityByFingerprint(t, compose, fingerprint).PrincipalID)
	})

	t.Run("approving a key already bound to someone else fails and leaves it bound", func(t *testing.T) {
		member := newMember(t, compose, "approver", authorizer.RoleOperator)
		signer, _ := newSigner(t)

		first := startLogin(t, compose, sshid, signer)
		second := startLogin(t, compose, sshid, signer)

		firstPrompt := first.awaitApproval(t)
		secondPrompt := second.awaitApproval(t)

		first.answer(confirmApproval(t, compose, firstPrompt.code))
		require.NoError(t, first.result(t))

		_, resp, err := confirmApprovalAs(t.Context(), compose, member.Token, secondPrompt.code, nil)
		require.NoError(t, err)
		assert.Equal(t, 409, resp.StatusCode(), resp.String())

		assert.Equal(t, models.SSHApprovalPending, approvalState(t, compose, secondPrompt.code),
			"a failed confirmation must not consume the approval")

		fingerprint := ssh.FingerprintSHA256(signer.PublicKey())
		assert.Len(t, identitiesHolding(t, compose, fingerprint), 1)
		assert.Equal(t, owner.ID, identityByFingerprint(t, compose, fingerprint).PrincipalID)
	})

	t.Run("a member whose role cannot approve can neither confirm nor reject", func(t *testing.T) {
		observer := newMember(t, compose, "watcher", authorizer.RoleObserver)
		signer, _ := newSigner(t)

		parked := startLogin(t, compose, sshid, signer)
		prompt := parked.awaitApproval(t)

		_, resp, err := confirmApprovalAs(t.Context(), compose, observer.Token, prompt.code, nil)
		require.NoError(t, err)
		assert.Equal(t, 403, resp.StatusCode(), resp.String())

		resp, err = rejectApprovalAs(t.Context(), compose, observer.Token, prompt.code)
		require.NoError(t, err)
		assert.Equal(t, 403, resp.StatusCode(), resp.String())

		assert.Equal(t, models.SSHApprovalPending, approvalState(t, compose, prompt.code))
		assert.Empty(t, identitiesHolding(t, compose, ssh.FingerprintSHA256(signer.PublicKey())))

		parked.answer(confirmApproval(t, compose, prompt.code))
		require.NoError(t, parked.result(t))
	})

	t.Run("a person outside the namespace can neither see nor decide its approval", func(t *testing.T) {
		compose.NewUser(t, "outsider", "outsider@ossystems.com.br", ShellHubPassword)
		compose.NewNamespace(t, "outsider", "elsewhere", uuid.Generate(), models.SSHAccessModeIdentity)
		outsider := compose.AuthUser(t, "outsider", ShellHubPassword)

		signer, _ := newSigner(t)
		parked := startLogin(t, compose, sshid, signer)
		prompt := parked.awaitApproval(t)

		_, resp, err := getApprovalAs(t.Context(), compose, outsider.Token, prompt.code)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode(), resp.String())
		assert.NotContains(t, resp.String(), ssh.FingerprintSHA256(signer.PublicKey()))

		_, resp, err = confirmApprovalAs(t.Context(), compose, outsider.Token, prompt.code, nil)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode(), resp.String())

		resp, err = rejectApprovalAs(t.Context(), compose, outsider.Token, prompt.code)
		require.NoError(t, err)
		assert.Equal(t, 404, resp.StatusCode(), resp.String())

		assert.Equal(t, models.SSHApprovalPending, approvalState(t, compose, prompt.code))
	})

	t.Run("the console shows the security code and key the terminal shows", func(t *testing.T) {
		signer, _ := newSigner(t)
		fingerprint := ssh.FingerprintSHA256(signer.PublicKey())

		parked := startLogin(t, compose, sshid, signer)
		prompt := parked.awaitApproval(t)

		assert.Contains(t, prompt.instruction, "Security code:  "+prompt.code[:4]+" "+prompt.code[4:])
		assert.Contains(t, prompt.instruction, "Key:            "+fingerprint)

		approval, resp, err := getApprovalAs(t.Context(), compose, "", prompt.code)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode(), resp.String())

		assert.Equal(t, prompt.code, approval.Code)
		assert.Equal(t, fingerprint, approval.Fingerprint)
		assert.Equal(t, models.SSHApprovalIdentity, approval.Kind)
		assert.Equal(t, sshid, approval.SSHID)
		assert.Equal(t, ShellHubNamespaceName, approval.Namespace)
		assert.Equal(t, device.Name, approval.DeviceName)
		assert.Equal(t, ShellHubAgentUsername, approval.Username)
	})
}
