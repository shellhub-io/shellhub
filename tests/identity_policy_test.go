package main

import (
	"context"
	"net/netip"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

const unroutableCIDR = "203.0.113.0/24"

// TestAccessPolicyEvaluation covers what each part of an access policy matches: the subject, by
// person, role, every member or API key; the device, by tag; the login; and the source address.
// It also covers the two refusals no policy can lift, a person outside the namespace and a deny
// that cannot be read. Each case reaches the device as someone only its own policies grant, so the
// owner's starter policy never decides it. Default-deny, deny-over-allow and a member with no grant
// are covered by [TestIdentityAccessPolicy].
func TestAccessPolicyEvaluation(t *testing.T) {
	ctx := context.Background()
	compose := newSSHEnvironment(t, ctx, models.SSHAccessModeIdentity)
	_, device := startAcceptedAgent(t, ctx, compose)

	sshid := deviceSSHID(device)

	everyone := requests.AccessPolicySubject{Type: string(models.PolicySubjectAllMembers)}

	t.Run("a person approved as a member and removed before the login resumes is refused as a non-member", func(t *testing.T) {
		member := newMember(t, compose, "leaver", authorizer.RoleOperator)
		grant(t, compose, &requests.AccessPolicyCreate{Name: "everyone", Subject: everyone, Logins: []string{"*"}})

		signer, _ := newSigner(t)
		parked := startLogin(t, compose, sshid, signer)
		prompt := parked.awaitApproval(t)

		confirmation, resp, err := confirmApprovalAs(t.Context(), compose, member.Token, prompt.code, nil)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode(), resp.String())

		compose.RemoveMember(t, "leaver", ShellHubNamespaceName)

		mark := compose.ServerLogMark(t)

		parked.answer(confirmation.ConfirmationCode)

		require.Error(t, parked.result(t))
		assert.Equal(t, accessDeniedReason, parked.denial())

		compose.AwaitServerLogLine(t, mark, "reason="+string(models.ReasonNotAMember), "user="+member.ID)

		assert.Empty(t, identitiesHolding(t, compose, ssh.FingerprintSHA256(signer.PublicKey())),
			"leaving the namespace takes the identity the approval enrolled with it")
	})

	t.Run("a source address policy grants only the addresses it lists", func(t *testing.T) {
		member, signer := memberWithKey(t, compose, "sourced", authorizer.RoleOperator)

		ownerSigner, _ := enrollOwnerKey(t, compose)
		session := sessionAfter(t, t.Context(), compose, currentSessions(t, t.Context(), compose), func() {
			requireStraightThrough(t, compose, sshid, ownerSigner)
		})

		client, err := netip.ParseAddr(session.IPAddress)
		require.NoError(t, err, "the session should record the address the gateway saw")

		grant(t, compose, &requests.AccessPolicyCreate{
			Name:     "elsewhere",
			Subject:  userSubject(member.ID),
			Logins:   []string{"*"},
			SourceIP: []string{unroutableCIDR},
		})

		mark := compose.ServerLogMark(t)
		requireRefusedAtAuth(t, compose, sshid, signer)
		compose.AwaitServerLogLine(t, mark, "reason="+string(models.ReasonNoGrant), "user="+member.ID)

		grant(t, compose, &requests.AccessPolicyCreate{
			Name:     "here",
			Subject:  userSubject(member.ID),
			Logins:   []string{"*"},
			SourceIP: []string{netip.PrefixFrom(client.Unmap(), client.Unmap().BitLen()).String()},
		})

		requireStraightThrough(t, compose, sshid, signer)
	})

	t.Run("a policy listing logins grants those logins and no other", func(t *testing.T) {
		member, signer := memberWithKey(t, compose, "logins", authorizer.RoleOperator)

		grant(t, compose, &requests.AccessPolicyCreate{
			Name:    "root only",
			Subject: userSubject(member.ID),
			Logins:  []string{ShellHubAgentUsername},
		})

		requireStraightThrough(t, compose, sshid, signer)

		mark := compose.ServerLogMark(t)
		requireRefusedAtAuth(t, compose, "operator@"+ShellHubNamespaceName+"."+device.Name, signer)
		compose.AwaitServerLogLine(t, mark, "reason="+string(models.ReasonNoGrant), "user="+member.ID, "username=operator")
	})

	t.Run("a tag-filtered policy grants only the devices carrying the tag", func(t *testing.T) {
		member, signer := memberWithKey(t, compose, "tagged", authorizer.RoleOperator)

		createTag(t, compose, "production")

		grant(t, compose, &requests.AccessPolicyCreate{
			Name:    "production",
			Subject: userSubject(member.ID),
			Filter:  requests.AccessPolicyFilter{Tags: []string{"production"}},
			Logins:  []string{"*"},
		})

		mark := compose.ServerLogMark(t)
		requireRefusedAtAuth(t, compose, sshid, signer)
		compose.AwaitServerLogLine(t, mark, "reason="+string(models.ReasonNoGrant), "user="+member.ID)

		tagDevice(t, compose, device.UID, "production")

		requireStraightThrough(t, compose, sshid, signer)
	})

	t.Run("a role policy grants the members holding that role and no other", func(t *testing.T) {
		member, signer := memberWithKey(t, compose, "roled", authorizer.RoleOperator)

		grant(t, compose, &requests.AccessPolicyCreate{
			Name:    "administrators",
			Subject: requests.AccessPolicySubject{Type: string(models.PolicySubjectRole), Value: string(authorizer.RoleAdministrator)},
			Logins:  []string{"*"},
		})

		mark := compose.ServerLogMark(t)
		requireRefusedAtAuth(t, compose, sshid, signer)
		compose.AwaitServerLogLine(t, mark, "reason="+string(models.ReasonNoGrant), "user="+member.ID)

		grant(t, compose, &requests.AccessPolicyCreate{
			Name:    "operators",
			Subject: requests.AccessPolicySubject{Type: string(models.PolicySubjectRole), Value: string(authorizer.RoleOperator)},
			Logins:  []string{"*"},
		})

		requireStraightThrough(t, compose, sshid, signer)
	})

	t.Run("an every-member policy grants members and not API keys", func(t *testing.T) {
		_, memberSigner := memberWithKey(t, compose, "everyone", authorizer.RoleOperator)
		key, keySigner := newAPIKeyIdentity(t, compose, "automation", false)

		grant(t, compose, &requests.AccessPolicyCreate{Name: "everyone", Subject: everyone, Logins: []string{"*"}})

		requireStraightThrough(t, compose, sshid, memberSigner)

		mark := compose.ServerLogMark(t)
		requireRefusedAtAuth(t, compose, sshid, keySigner)
		compose.AwaitServerLogLine(t, mark, "reason="+string(models.ReasonNoGrant), "user="+key.ID)

		grant(t, compose, &requests.AccessPolicyCreate{
			Name:    "automation",
			Subject: apiKeySubject(key.ID),
			Logins:  []string{"*"},
		})

		requireStraightThrough(t, compose, sshid, keySigner)
	})

	t.Run("a deny policy that cannot be evaluated refuses the login", func(t *testing.T) {
		signer, _ := enrollOwnerKey(t, compose)

		deny := grant(t, compose, &requests.AccessPolicyCreate{
			Name:     "unreadable",
			Subject:  everyone,
			Logins:   []string{"*"},
			SourceIP: []string{unroutableCIDR},
			Action:   string(models.PolicyActionDeny),
		})

		requireStraightThrough(t, compose, sshid, signer)

		compose.BreakAccessPolicySourceIP(t, deny.ID)

		mark := compose.ServerLogMark(t)
		requireRefusedAtAuth(t, compose, sshid, signer)
		compose.AwaitServerLogLine(t, mark,
			"reason="+string(models.ReasonPolicyUnevaluable), "policy_name=unreadable")
	})
}

func memberWithKey(t *testing.T, compose *environment.DockerCompose, username string, role authorizer.Role) (*models.UserAuthResponse, ssh.Signer) {
	t.Helper()

	member := newMember(t, compose, username, role)

	signer, data := newSigner(t)
	compose.EnrollIdentityAs(t, member.Token, username, data)

	return member, signer
}

func createTag(t *testing.T, compose *environment.DockerCompose, name string) {
	t.Helper()

	resp, err := compose.R(t.Context()).SetBody(map[string]string{"name": name}).Post("/api/tags")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode(), resp.String())
}

func tagDevice(t *testing.T, compose *environment.DockerCompose, uid, tag string) {
	t.Helper()

	resp, err := compose.R(t.Context()).Post("/api/devices/" + uid + "/tags/" + tag)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode(), resp.String())

	t.Cleanup(func() {
		resp, err := compose.R(context.WithoutCancel(t.Context())).Delete("/api/devices/" + uid + "/tags/" + tag)
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode(), resp.String())
	})
}
