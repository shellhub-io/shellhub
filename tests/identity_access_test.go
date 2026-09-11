package main

import (
	"context"
	"net"
	"regexp"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

const approvalWait = 60 * time.Second

func dialSSH(ctx context.Context, addr, sshid string, signer ssh.Signer) error {
	return dialSSHForApproval(ctx, addr, sshid, signer, nil)
}

func dialSSHForApproval(ctx context.Context, addr, sshid string, signer ssh.Signer, banners chan<- string) error {
	config := &ssh.ClientConfig{
		User:            sshid,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // a test dialing its own throwaway server
	}

	if banners != nil {
		config.BannerCallback = func(message string) error {
			select {
			case banners <- message:
			default:
			}

			return nil
		}
	}

	dialer := new(net.Dialer)

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}

	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		_ = conn.Close()

		return err
	}

	_ = ssh.NewClient(c, chans, reqs).Close()

	return nil
}

func requireLoginRefused(t *testing.T, compose *environment.DockerCompose, device *models.Device, signer ssh.Signer, wantServerLogs ...string) {
	t.Helper()

	err := dialSSH(t.Context(), compose.SSHAddress(), deviceSSHID(device), signer)
	require.Error(t, err)

	for _, substr := range wantServerLogs {
		compose.AwaitServerLog(t, substr)
	}
}

var approvalCodePattern = regexp.MustCompile(`/ssh-identities/new/([2-9A-Z]{8})`)

// TestIdentityAccessPolicy covers what decides a login once a namespace is in the identity model:
// an unknown key is held for approval and let in once confirmed, default-deny leaves a member
// without a grant outside, a deny beats the allow that would otherwise let the owner in, and a
// namespace stripped of every policy admits nobody. Password refusal and a pre-enrolled identity
// connecting are covered by [TestSSHIdentityMode].
func TestIdentityAccessPolicy(t *testing.T) {
	t.Run("an unknown key is held for approval and let in once confirmed", func(t *testing.T) {
		ctx := context.Background()
		compose := newSSHEnvironment(t, ctx, models.SSHAccessModeIdentity)
		_, device := startAcceptedAgent(t, ctx, compose)

		signer, _ := newSigner(t)

		banners := make(chan string, 8)
		dialed := make(chan error, 1)

		go func() {
			dialed <- dialSSHForApproval(ctx, compose.SSHAddress(), deviceSSHID(device), signer, banners)
		}()

		var code string

		deadline := time.After(approvalWait)

		for code == "" {
			select {
			case message := <-banners:
				if match := approvalCodePattern.FindStringSubmatch(message); match != nil {
					code = match[1]
				}
			case err := <-dialed:
				require.NoError(t, err, "the login failed before an approval code was offered")
				require.Fail(t, "the login completed without an approval code being offered")
			case <-deadline:
				require.Fail(t, "the gateway never offered an approval code")
			}
		}

		resp, err := compose.R(ctx).Post("/api/ssh-approvals/" + code + "/confirm")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode())

		select {
		case err := <-dialed:
			require.NoError(t, err, "the login should resume once the approval is confirmed")
		case <-time.After(approvalWait):
			require.Fail(t, "the login never resumed after the approval was confirmed")
		}

		identities := []models.SSHIdentity{}

		resp, err = compose.R(ctx).SetResult(&identities).Get("/api/ssh-identities")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode())
		require.Len(t, identities, 1)
		assert.Equal(t, models.SSHIdentitySourceApproval, identities[0].Source)
		assert.Equal(t, ssh.FingerprintSHA256(signer.PublicKey()), identities[0].Fingerprint,
			"the identity should hold the key that dialled")
	})

	tests := []struct {
		name           string
		refuse         func(t *testing.T, compose *environment.DockerCompose) ssh.Signer
		wantServerLogs []string
	}{
		{
			name: "a member no policy grants is refused",
			refuse: func(t *testing.T, compose *environment.DockerCompose) ssh.Signer {
				t.Helper()

				compose.NewUser(t, "member", "member@ossystems.com.br", ShellHubPassword)
				compose.NewMember(t, "member", ShellHubNamespaceName, string(authorizer.RoleOperator))

				auth := compose.AuthUser(t, "member", ShellHubPassword)
				require.Equal(t, ShellHubNamespace, auth.Tenant)

				signer, data := newSigner(t)
				compose.EnrollIdentityAs(t, auth.Token, "member", data)

				return signer
			},
			wantServerLogs: []string{"reason=" + string(models.ReasonNoGrant)},
		},
		{
			name: "a deny policy beats the allow that would grant the owner",
			refuse: func(t *testing.T, compose *environment.DockerCompose) ssh.Signer {
				t.Helper()

				signer, data := newSigner(t)
				compose.EnrollIdentity(t, "owner", data)

				compose.CreateAccessPolicy(t, &requests.AccessPolicyCreate{
					Name:    "no root",
					Subject: requests.AccessPolicySubject{Type: string(models.PolicySubjectAllMembers)},
					Logins:  []string{ShellHubAgentUsername},
					Action:  string(models.PolicyActionDeny),
				})

				return signer
			},
			wantServerLogs: []string{"reason=" + string(models.ReasonDeniedByPolicy)},
		},
		{
			name: "a namespace with no policy at all refuses everyone",
			refuse: func(t *testing.T, compose *environment.DockerCompose) ssh.Signer {
				t.Helper()

				signer, data := newSigner(t)
				compose.EnrollIdentity(t, "owner", data)

				policies := []models.AccessPolicy{}

				resp, err := compose.R(t.Context()).SetResult(&policies).Get("/api/access-policies")
				require.NoError(t, err)
				require.Equal(t, 200, resp.StatusCode())
				require.NotEmpty(t, policies, "a namespace born in identity is seeded with the owner policy")

				for _, policy := range policies {
					resp, err := compose.R(t.Context()).Delete("/api/access-policies/" + policy.ID)
					require.NoError(t, err)
					require.Equal(t, 200, resp.StatusCode())
				}

				return signer
			},
			wantServerLogs: []string{
				"destination device did not pass the connection evaluation",
				`error="ssh access denied by policy"`,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			compose := newSSHEnvironment(t, ctx, models.SSHAccessModeIdentity)
			_, device := startAcceptedAgent(t, ctx, compose)

			signer := tc.refuse(t, compose)

			requireLoginRefused(t, compose, device, signer, tc.wantServerLogs...)
		})
	}
}
