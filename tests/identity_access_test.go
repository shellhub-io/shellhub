package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net"
	"regexp"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func newSigner(t *testing.T) (ssh.Signer, string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	pub, err := ssh.NewPublicKey(&key.PublicKey)
	require.NoError(t, err)

	signer, err := ssh.NewSignerFromKey(key)
	require.NoError(t, err)

	return signer, string(ssh.MarshalAuthorizedKey(pub))
}

func enrollIdentity(t *testing.T, ctx context.Context, compose *environment.DockerCompose, token, name, data string) {
	t.Helper()

	resp, err := compose.R(ctx).
		SetAuthToken(token).
		SetBody(&requests.SSHIdentityCreate{Name: name, Data: data}).
		Post("/api/ssh-identities")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode())
}

func createAccessPolicy(t *testing.T, ctx context.Context, compose *environment.DockerCompose, req *requests.AccessPolicyCreate) {
	t.Helper()

	resp, err := compose.R(ctx).SetBody(req).Post("/api/access-policies")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode())
}

func awaitServerLogContains(t *testing.T, ctx context.Context, compose *environment.DockerCompose, substr string) {
	t.Helper()

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		reader, err := compose.Service(environment.ServiceServer).Logs(ctx)
		if !assert.NoError(tt, err) {
			return
		}

		defer func() { _ = reader.Close() }()

		logs, err := io.ReadAll(reader)
		assert.NoError(tt, err)
		assert.Contains(tt, string(logs), substr)
	}, 30*time.Second, 2*time.Second)
}

func dialSSH(ctx context.Context, addr, sshid string, signer ssh.Signer, banners chan<- string) error {
	config := &ssh.ClientConfig{
		User:            sshid,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // a test dialing its own throwaway server
		Timeout:         15 * time.Second,
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

var approvalCodePattern = regexp.MustCompile(`/ssh-identities/new/([2-9A-Z]{8})`)

// TestIdentityAccessPolicy covers what the Access Policies decide once a namespace is in the
// identity model: default-deny leaves a member without a grant outside, a deny beats the allow that
// would otherwise let the owner in, and a namespace stripped of every policy admits nobody. The
// authentication side of the model is covered by [TestSSHIdentityMode].
func TestIdentityAccessPolicy(t *testing.T) {
	t.Run("an unknown key is held for approval and let in once confirmed", func(t *testing.T) {
		ctx := context.Background()
		compose := newSSHEnvironment(t, ctx, models.SSHAccessModeIdentity)
		_, device := startAcceptedAgent(t, ctx, compose)

		signer, _ := newSigner(t)
		sshid := ShellHubAgentUsername + "@" + ShellHubNamespaceName + "." + device.Name
		addr := "localhost:" + compose.Env("SHELLHUB_SSH_PORT")

		banners := make(chan string, 8)
		dialed := make(chan error, 1)

		go func() { dialed <- dialSSH(ctx, addr, sshid, signer, banners) }()

		var code string

		require.Eventually(t, func() bool {
			select {
			case message := <-banners:
				if match := approvalCodePattern.FindStringSubmatch(message); match != nil {
					code = match[1]
				}
			case <-time.After(time.Second):
			}

			return code != ""
		}, 60*time.Second, 10*time.Millisecond, "the gateway never offered an approval code")

		resp, err := compose.R(ctx).Post("/api/ssh-approvals/" + code + "/confirm")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode())

		select {
		case err := <-dialed:
			require.NoError(t, err, "the login should resume once the approval is confirmed")
		case <-time.After(60 * time.Second):
			require.Fail(t, "the login never resumed after the approval was confirmed")
		}

		identities := []models.SSHIdentity{}

		resp, err = compose.R(ctx).SetResult(&identities).Get("/api/ssh-identities")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode())
		require.Len(t, identities, 1)
		assert.Equal(t, models.SSHIdentitySourceApproval, identities[0].Source)
	})

	t.Run("a member no policy grants is refused", func(t *testing.T) {
		ctx := context.Background()
		compose := newSSHEnvironment(t, ctx, models.SSHAccessModeIdentity)
		_, device := startAcceptedAgent(t, ctx, compose)

		compose.NewUser(t, "member", "member@ossystems.com.br", ShellHubPassword)
		compose.NewMember(t, "member", ShellHubNamespaceName, "operator")

		auth := compose.AuthUser(ctx, "member", ShellHubPassword)
		require.Equal(t, ShellHubNamespace, auth.Tenant)

		signer, data := newSigner(t)
		enrollIdentity(t, ctx, compose, auth.Token, "member", data)

		sshid := ShellHubAgentUsername + "@" + ShellHubNamespaceName + "." + device.Name
		err := dialSSH(ctx, "localhost:"+compose.Env("SHELLHUB_SSH_PORT"), sshid, signer, nil)
		require.Error(t, err)

		awaitServerLogContains(t, ctx, compose, "reason="+string(models.ReasonNoGrant))
	})

	t.Run("a deny policy beats the allow that would grant the owner", func(t *testing.T) {
		ctx := context.Background()
		compose := newSSHEnvironment(t, ctx, models.SSHAccessModeIdentity)
		_, device := startAcceptedAgent(t, ctx, compose)

		signer, data := newSigner(t)

		auth := compose.AuthUser(ctx, ShellHubUsername, ShellHubPassword)
		enrollIdentity(t, ctx, compose, auth.Token, "owner", data)

		createAccessPolicy(t, ctx, compose, &requests.AccessPolicyCreate{
			Name:    "no root",
			Subject: requests.AccessPolicySubject{Type: "all-members"},
			Logins:  []string{ShellHubAgentUsername},
			Action:  "deny",
		})

		sshid := ShellHubAgentUsername + "@" + ShellHubNamespaceName + "." + device.Name
		err := dialSSH(ctx, "localhost:"+compose.Env("SHELLHUB_SSH_PORT"), sshid, signer, nil)
		require.Error(t, err)

		awaitServerLogContains(t, ctx, compose, "reason="+string(models.ReasonDeniedByPolicy))
	})

	t.Run("a namespace with no policy at all refuses everyone", func(t *testing.T) {
		ctx := context.Background()
		compose := newSSHEnvironment(t, ctx, models.SSHAccessModeIdentity)
		_, device := startAcceptedAgent(t, ctx, compose)

		signer, data := newSigner(t)

		auth := compose.AuthUser(ctx, ShellHubUsername, ShellHubPassword)
		enrollIdentity(t, ctx, compose, auth.Token, "owner", data)

		policies := []models.AccessPolicy{}

		resp, err := compose.R(ctx).SetResult(&policies).Get("/api/access-policies")
		require.NoError(t, err)
		require.Equal(t, 200, resp.StatusCode())
		require.NotEmpty(t, policies, "a namespace born in identity is seeded with the owner policy")

		for _, policy := range policies {
			resp, err := compose.R(ctx).Delete("/api/access-policies/" + policy.ID)
			require.NoError(t, err)
			require.Equal(t, 200, resp.StatusCode())
		}

		sshid := ShellHubAgentUsername + "@" + ShellHubNamespaceName + "." + device.Name
		err = dialSSH(ctx, "localhost:"+compose.Env("SHELLHUB_SSH_PORT"), sshid, signer, nil)
		require.Error(t, err, "default-deny leaves nobody in when no policy grants anything")
	})
}
