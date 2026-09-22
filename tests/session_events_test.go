package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/bramvdbogaerde/go-scp"
	"github.com/pkg/sftp"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/tests/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ssh"
)

func listSessions(ctx context.Context, compose *environment.DockerCompose) ([]models.Session, error) {
	sessions := []models.Session{}

	resp, err := compose.R(ctx).SetResult(&sessions).Get("/api/sessions")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("the list answered %d, and the strict validator names the mismatch: %s", resp.StatusCode(), resp.String())
	}

	return sessions, nil
}

func currentSessions(t *testing.T, ctx context.Context, compose *environment.DockerCompose) []models.Session {
	t.Helper()

	var sessions []models.Session

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		var err error

		sessions, err = listSessions(ctx, compose)
		assert.NoError(tt, err)
	}, 30*time.Second, 1*time.Second)

	return sessions
}

func sessionAfter(t *testing.T, ctx context.Context, compose *environment.DockerCompose, before []models.Session, action func()) *models.Session {
	t.Helper()

	known := make(map[string]struct{}, len(before))
	for _, session := range before {
		known[session.UID] = struct{}{}
	}

	action()

	opened := &models.Session{} //nolint:exhaustruct // filled from the API response

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		sessions, err := listSessions(ctx, compose)
		if !assert.NoError(tt, err) {
			return
		}

		for _, session := range sessions {
			if _, seen := known[session.UID]; !seen {
				*opened = session

				return
			}
		}

		assert.Fail(tt, "the action opened no session the list did not already carry")
	}, 30*time.Second, 1*time.Second)

	return opened
}

func sessionDetail(t *testing.T, ctx context.Context, compose *environment.DockerCompose, uid string) models.Session {
	t.Helper()

	session := models.Session{} //nolint:exhaustruct // filled from the API response

	require.EventuallyWithT(t, func(tt *assert.CollectT) {
		resp, err := compose.R(ctx).SetResult(&session).Get("/api/sessions/" + uid)
		assert.NoError(tt, err)
		assert.Equal(tt, 200, resp.StatusCode(), "the strict validator names the mismatch: %s", resp.String())
		assert.NotEmpty(tt, session.Events.Items, "the timeline is written after the channel opens")
	}, 30*time.Second, 1*time.Second)

	return session
}

func itemOfType(events []models.SessionEvent, eventType models.SessionEventType) *models.SessionEvent {
	for i, event := range events {
		if event.Type == eventType {
			return &events[i]
		}
	}

	return nil
}

func payloadString(t *testing.T, event *models.SessionEvent, key string) string {
	t.Helper()

	require.NotNil(t, event, "no event of the wanted type is in the timeline")

	data, ok := event.Data.(map[string]any)
	require.True(t, ok, "the event payload is not an object: %T", event.Data)

	value, ok := data[key].(string)
	require.True(t, ok, "the payload carries no string under %q: %v", key, data)

	return value
}

func TestSessionDetailSaysWhatTheSessionDid(t *testing.T) {
	ctx := context.Background()

	compose := newSSHEnvironment(t, ctx, "legacy")
	_, device := startAcceptedAgent(t, ctx, compose)

	cases := []struct {
		name   string
		open   func(t *testing.T, conn *ssh.Client)
		expect func(t *testing.T, session models.Session)
	}{
		{
			name: "an interactive shell names the terminal the client asked for",
			open: func(t *testing.T, conn *ssh.Client) {
				t.Helper()

				sess, err := conn.NewSession()
				require.NoError(t, err)

				require.NoError(t, sess.RequestPty("xterm-256color", 30, 141, ssh.TerminalModes{ssh.ECHO: 1}))
				require.NoError(t, sess.Shell())
			},
			expect: func(t *testing.T, session models.Session) {
				t.Helper()

				assert.Equal(t, models.SessionEventTypePtyRequest, session.Events.First)
				assert.Equal(t, "xterm-256color",
					payloadString(t, itemOfType(session.Events.Items, models.SessionEventTypePtyRequest), "term"))
			},
		},
		{
			name: "a single command carries the command it ran",
			open: func(t *testing.T, conn *ssh.Client) {
				t.Helper()

				sess, err := conn.NewSession()
				require.NoError(t, err)

				_, err = sess.CombinedOutput("uptime")
				require.NoError(t, err)
			},
			expect: func(t *testing.T, session models.Session) {
				t.Helper()

				assert.Equal(t, models.SessionEventTypeExec, session.Events.First)
				assert.Equal(t, "uptime",
					payloadString(t, itemOfType(session.Events.Items, models.SessionEventTypeExec), "command"))
			},
		},
		{
			name: "an SFTP transfer opens with the subsystem",
			open: func(t *testing.T, conn *ssh.Client) {
				t.Helper()

				client, err := sftp.NewClient(conn)
				require.NoError(t, err)

				file, err := client.OpenFile("/tmp/sent-by-sftp", os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
				require.NoError(t, err)

				_, err = file.Write([]byte("sent"))
				require.NoError(t, err)

				require.NoError(t, file.Close())
				require.NoError(t, client.Close())
			},
			expect: func(t *testing.T, session models.Session) {
				t.Helper()

				assert.Equal(t, models.SessionEventTypeSubsystem, session.Events.First)
			},
		},
		{
			name: "a legacy scp is told from a plain command by its own payload",
			open: func(t *testing.T, conn *ssh.Client) {
				t.Helper()

				client, err := scp.NewClientBySSH(conn)
				require.NoError(t, err)

				file := bytes.NewBuffer(make([]byte, 64))
				require.NoError(t, client.CopyFilePassThru(ctx, file, "/tmp/sent-by-scp", "0644", io.LimitReader))

				client.Close()
			},
			expect: func(t *testing.T, session models.Session) {
				t.Helper()

				assert.Equal(t, models.SessionEventTypeExec, session.Events.First)
				assert.Contains(t,
					payloadString(t, itemOfType(session.Events.Items, models.SessionEventTypeExec), "command"),
					"scp",
					"only the payload separates a legacy transfer from an ordinary command")
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := currentSessions(t, ctx, compose)

			conn := dialDevice(t, ctx, compose, device)
			t.Cleanup(func() { _ = conn.Close() })

			opened := sessionAfter(t, ctx, compose, before, func() { tc.open(t, conn) })

			assert.Empty(t, opened.Events.Items, "the list carries no timeline")
			assert.Nil(t, opened.Principal, "the legacy access model authorizes no principal")

			session := sessionDetail(t, ctx, compose, opened.UID)
			tc.expect(t, session)

			for i := 1; i < len(session.Events.Items); i++ {
				assert.False(t, session.Events.Items[i].Timestamp.Before(session.Events.Items[i-1].Timestamp),
					"the timeline is oldest first")
			}

			for _, item := range session.Events.Items {
				assert.NotEqual(t, models.SessionEventTypePtyOutput, item.Type,
					"terminal output is the recording, not a timeline entry")
			}
		})
	}
}

func TestSessionPrincipalNamesWhoOpenedIt(t *testing.T) {
	ctx := context.Background()

	compose := newSSHEnvironment(t, ctx, models.SSHAccessModeIdentity)
	_, device := startAcceptedAgent(t, ctx, compose)

	owner := compose.AuthUser(t, ShellHubUsername, ShellHubPassword)

	signer, data := newSigner(t)
	compose.EnrollIdentity(t, "integration", data)

	before := currentSessions(t, ctx, compose)

	conn := dialClient(t, ctx, compose.SSHAddress(), &ssh.ClientConfig{ //nolint:exhaustruct // the remaining fields keep their defaults
		User:            deviceSSHID(device),
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec // the test stack's host key is ephemeral
	})
	t.Cleanup(func() { _ = conn.Close() })

	opened := sessionAfter(t, ctx, compose, before, func() {
		sess, err := conn.NewSession()
		require.NoError(t, err)

		_, err = sess.CombinedOutput("uptime")
		require.NoError(t, err)
	})

	require.NotNil(t, opened.Principal, "an identity-mode session is opened by someone")
	assert.Equal(t, models.Principal{Kind: models.PrincipalUser, ID: owner.ID}, *opened.Principal)
}
