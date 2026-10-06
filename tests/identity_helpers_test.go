package main

import (
	"context"
	"regexp"
	"slices"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

const (
	approvalChallengeName = "shellhub-approval"
	deniedChallengeName   = "shellhub-approval-denied"
)

const (
	accessDeniedReason = "An access policy does not allow this login."
)

var reauthPromptPattern = regexp.MustCompile(`/ssh-identities/confirm/([2-9A-Z]{8})`)

type challengePrompt struct {
	name        string
	instruction string
}

type approvalPrompt struct {
	code        string
	kind        models.SSHApprovalKind
	instruction string
}

type login struct {
	prompts chan challengePrompt
	answers chan string
	done    chan error
	seen    []challengePrompt
}

func startLogin(t *testing.T, compose *environment.DockerCompose, sshid string, signer ssh.Signer) *login {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), 3*approvalWait)

	l := &login{
		prompts: make(chan challengePrompt, 16),
		answers: make(chan string, 1),
		done:    make(chan error, 1),
	}

	go func() {
		l.done <- dialInteractive(ctx, compose.SSHAddress(), sshid, signer, l.prompts, l.answers)
	}()

	t.Cleanup(func() {
		cancel()
		<-l.done
	})

	return l
}

func dialInteractive(ctx context.Context, addr, sshid string, signer ssh.Signer, prompts chan<- challengePrompt, answers <-chan string) error {
	challenge := func(name, instruction string, questions []string, _ []bool) ([]string, error) {
		select {
		case prompts <- challengePrompt{name: name, instruction: instruction}:
		case <-ctx.Done():
			return nil, ctx.Err()
		}

		if name != approvalChallengeName || len(questions) == 0 {
			return make([]string, len(questions)), nil
		}

		select {
		case answer := <-answers:
			return []string{answer}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	return handshake(ctx, addr, sshid, []ssh.AuthMethod{ssh.PublicKeys(signer), ssh.KeyboardInteractive(challenge)})
}

func (l *login) awaitApproval(t *testing.T) approvalPrompt {
	t.Helper()

	deadline := time.After(approvalWait)

	for {
		select {
		case prompt := <-l.prompts:
			l.seen = append(l.seen, prompt)

			if prompt.name != approvalChallengeName {
				continue
			}

			if match := approvalCodePattern.FindStringSubmatch(prompt.instruction); match != nil {
				return approvalPrompt{code: match[1], kind: models.SSHApprovalIdentity, instruction: prompt.instruction}
			}

			if match := reauthPromptPattern.FindStringSubmatch(prompt.instruction); match != nil {
				return approvalPrompt{code: match[1], kind: models.SSHApprovalReauth, instruction: prompt.instruction}
			}

			require.Failf(t, "the approval prompt carries no code", "%q", prompt.instruction)
		case err := <-l.done:
			l.done <- err

			require.Failf(t, "the login ended before an approval was asked for", "err: %v", err)
		case <-deadline:
			require.Fail(t, "the gateway never asked for an approval")
		}
	}
}

func (l *login) answer(code string) {
	l.answers <- code
}

func (l *login) result(t *testing.T) error {
	t.Helper()

	var err error

	select {
	case err = <-l.done:
		l.done <- err
	case <-time.After(approvalWait):
		require.Fail(t, "the login never ended")
	}

	for {
		select {
		case prompt := <-l.prompts:
			l.seen = append(l.seen, prompt)
		default:
			return err
		}
	}
}

func (l *login) denial() string {
	for _, prompt := range slices.Backward(l.seen) {
		if prompt.name == deniedChallengeName {
			return prompt.instruction
		}
	}

	return ""
}

func (l *login) approvals() int {
	count := 0

	for _, prompt := range l.seen {
		if prompt.name == approvalChallengeName {
			count++
		}
	}

	return count
}

func requireStraightThrough(t *testing.T, compose *environment.DockerCompose, sshid string, signer ssh.Signer) {
	t.Helper()

	l := startLogin(t, compose, sshid, signer)
	require.NoError(t, l.result(t))
	require.Zero(t, l.approvals(), "the login was held for an approval it should not need")
}

func requireRefusedAtAuth(t *testing.T, compose *environment.DockerCompose, sshid string, signer ssh.Signer) {
	t.Helper()

	err := dialSSH(t.Context(), compose.SSHAddress(), sshid, []ssh.Signer{signer}, nil, nil)
	require.ErrorContains(t, err, "unable to authenticate")
}

func approvalRequest(ctx context.Context, compose *environment.DockerCompose, token string) *resty.Request {
	req := compose.R(ctx)
	if token != "" {
		req = req.SetAuthToken(token)
	}

	return req
}

func confirmApprovalAs(ctx context.Context, compose *environment.DockerCompose, token, code string, expiresIn *int) (*models.SSHApprovalConfirmation, *resty.Response, error) {
	confirmation := new(models.SSHApprovalConfirmation)

	req := approvalRequest(ctx, compose, token).SetResult(confirmation)
	if expiresIn != nil {
		req = req.SetBody(map[string]int{"expires_in": *expiresIn})
	}

	resp, err := req.Post("/api/ssh-approvals/" + code + "/confirm")

	return confirmation, resp, err
}

func identitiesHolding(t *testing.T, compose *environment.DockerCompose, fingerprint string) []models.SSHIdentity {
	t.Helper()

	identities := []models.SSHIdentity{}

	resp, err := compose.R(t.Context()).SetResult(&identities).Get("/api/ssh-identities")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode(), resp.String())

	return slices.DeleteFunc(identities, func(identity models.SSHIdentity) bool {
		return identity.Fingerprint != fingerprint
	})
}

func identityByFingerprint(t *testing.T, compose *environment.DockerCompose, fingerprint string) models.SSHIdentity {
	t.Helper()

	holding := identitiesHolding(t, compose, fingerprint)
	require.Len(t, holding, 1, "exactly one identity should hold the key %s", fingerprint)

	return holding[0]
}

func enrollOwnerKey(t *testing.T, compose *environment.DockerCompose) (ssh.Signer, string) {
	t.Helper()

	signer, data := newSigner(t)
	compose.EnrollIdentity(t, "owner", data)

	return signer, ssh.FingerprintSHA256(signer.PublicKey())
}

func newMember(t *testing.T, compose *environment.DockerCompose, username string, role authorizer.Role) *models.UserAuthResponse {
	t.Helper()

	compose.NewUser(t, username, username+"@ossystems.com.br", ShellHubPassword)
	compose.NewMember(t, username, ShellHubNamespaceName, string(role))

	return compose.AuthUser(t, username, ShellHubPassword)
}

func newAPIKeyIdentity(t *testing.T, compose *environment.DockerCompose, name string, singleUse bool) (*responses.CreateAPIKey, ssh.Signer) {
	t.Helper()

	key := compose.CreateAPIKey(t, &requests.CreateAPIKey{Name: name, ExpiresAt: -1})

	signer, data := newSigner(t)

	resp, err := compose.R(t.Context()).
		SetBody(map[string]any{"name": name, "data": data, "single_use": singleUse}).
		Post("/api/namespaces/api-key/" + name + "/ssh-identities")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode(), resp.String())

	return key, signer
}

func apiKeyIdentity(t *testing.T, compose *environment.DockerCompose, name string) models.SSHIdentity {
	t.Helper()

	identities := []models.SSHIdentity{}

	resp, err := compose.R(t.Context()).SetResult(&identities).Get("/api/namespaces/api-key/" + name + "/ssh-identities")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode(), resp.String())
	require.Len(t, identities, 1)

	return identities[0]
}

func userSubject(id string) requests.AccessPolicySubject {
	return requests.AccessPolicySubject{Type: string(models.PolicySubjectUser), Value: id}
}

func apiKeySubject(id string) requests.AccessPolicySubject {
	return requests.AccessPolicySubject{Type: string(models.PolicySubjectAPIKey), Value: id}
}

func grant(t *testing.T, compose *environment.DockerCompose, req *requests.AccessPolicyCreate) models.AccessPolicy {
	t.Helper()

	policy := compose.CreateAccessPolicy(t, req)
	t.Cleanup(func() { compose.DeleteAccessPolicy(t, policy.ID) })

	return policy
}

func requireRecent(t *testing.T, at *time.Time) {
	t.Helper()

	require.NotNil(t, at)
	assert.WithinDuration(t, time.Now(), *at, time.Minute) //nolint:forbidigo // compares a stamp the server wrote with its own wall clock
}
