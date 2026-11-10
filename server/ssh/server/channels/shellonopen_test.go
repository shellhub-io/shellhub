package channels

import (
	"slices"
	"testing"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/ssh/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

func withFollowingRequestsGrace(t *testing.T, grace time.Duration) {
	t.Helper()

	previous := followingRequestsGrace
	followingRequestsGrace = grace

	t.Cleanup(func() { followingRequestsGrace = previous })
}

func sessionOnDevice(t *testing.T, sess *fakeSession) (*gossh.Session, <-chan string) {
	t.Helper()

	sshClient, gatewayChannels := sshClientOverLoopback(t)
	agentClient, deviceChannels := sshClientOverLoopback(t)

	received := make(chan string, 8)

	go serveDeviceRefusingRequests(deviceChannels, received)

	type opened struct {
		session *gossh.Session
		err     error
	}

	sessions := make(chan opened, 1)

	go func() {
		sess, err := sshClient.NewSession()
		sessions <- opened{session: sess, err: err}
	}()

	accepted, clientRequests, err := (<-gatewayChannels).Accept()
	require.NoError(t, err)

	opening := <-sessions
	require.NoError(t, opening.err)

	agentChannel, agentRequests, err := agentClient.OpenChannel(SessionChannel, nil)
	require.NoError(t, err)

	sess.agent = &session.AgentChannel{Channel: agentChannel, Requests: agentRequests, Shell: sess.startedShell}
	sess.client = &session.ClientChannel{Channel: accepted, Requests: clientRequests}

	ctx := newStubContext()
	ctx.SetValue(gliderssh.ContextKeyConn, agentClient.Conn)

	go sessionChannel(ctx, sess, new(rejectingNewChannel))

	return opening.session, received
}

func sessionOnShellStartingDevice(t *testing.T) (*gossh.Session, *fakeSession, <-chan string) {
	t.Helper()

	sess := &fakeSession{startsShell: true} //nolint:exhaustruct // only the shell on open matters here
	client, received := sessionOnDevice(t, sess)

	return client, sess, received
}

func nextOnDevice(t *testing.T, received <-chan string) string {
	t.Helper()

	select {
	case name := <-received:
		return name
	case <-time.After(5 * time.Second):
		t.Fatal("the device received no request")

		return ""
	}
}

func shellSetup(sess *fakeSession) []string {
	sess.mu.Lock()
	defer sess.mu.Unlock()

	return append([]string(nil), sess.shellSetup...)
}

func reopened(sess *fakeSession) int {
	sess.mu.Lock()
	defer sess.mu.Unlock()

	return sess.reopened
}

var (
	clientPty    = gossh.Marshal(models.SSHPty{Term: "xterm-256color", Columns: 120, Rows: 40, Modelist: []byte{0}})
	clientResize = gossh.Marshal(models.SSHWindowChange{Columns: 100, Rows: 30})
)

func TestSessionChannelStartsTheShellOnOpen(t *testing.T) {
	t.Run("opens the shell with the setup a client sent ahead of its shell request", func(t *testing.T) {
		withFollowingRequestsGrace(t, 5*time.Second)

		client, sess, received := sessionOnShellStartingDevice(t)

		_, err := client.SendRequest("env", false, gossh.Marshal(struct{ Name, Value string }{"TZ", "UTC"}))
		require.NoError(t, err)
		_, err = client.SendRequest(PtyRequestType, false, clientPty)
		require.NoError(t, err)
		_, err = client.SendRequest("env", false, gossh.Marshal(struct{ Name, Value string }{"LANG", "C.UTF-8"}))
		require.NoError(t, err)

		ok, err := client.SendRequest(ShellRequestType, true, nil)
		require.NoError(t, err)
		assert.True(t, ok, "the shell the device started must be granted")

		_, err = client.SendRequest(WindowChangeRequestType, false, clientResize)
		require.NoError(t, err)

		assert.Equal(t, "env", nextOnDevice(t, received), "a request ahead of the pty reaches the device on its plain channel")
		assert.Equal(t, WindowChangeRequestType, nextOnDevice(t, received), "the pty, the env after it and the shell must not reach the device again")

		assert.Equal(t, []string{"env", PtyRequestType, "env"}, shellSetup(sess))
		assert.Zero(t, reopened(sess))
		assert.True(t, sess.called("SetSeatPty"))
	})

	t.Run("accepts the device's agent connections before the shell starts", func(t *testing.T) {
		withFollowingRequestsGrace(t, 5*time.Second)

		client, sess, _ := sessionOnShellStartingDevice(t)

		_, err := client.SendRequest(PtyRequestType, false, clientPty)
		require.NoError(t, err)
		_, err = client.SendRequest(AuthRequestOpenSSHRequest, false, nil)
		require.NoError(t, err)

		ok, err := client.SendRequest(ShellRequestType, true, nil)
		require.NoError(t, err)
		assert.True(t, ok)

		sess.mu.Lock()
		defer sess.mu.Unlock()

		forwarded := slices.Index(sess.calls, "ForwardAgent")
		require.GreaterOrEqual(t, forwarded, 0, "the agent request must reach ForwardAgent")
		assert.Less(t, forwarded, slices.Index(sess.calls, "StartShell"))
	})

	t.Run("leaves the shell to a client that waits for its pty to be granted", func(t *testing.T) {
		withFollowingRequestsGrace(t, time.Millisecond)

		client, sess, received := sessionOnShellStartingDevice(t)

		ok, err := client.SendRequest(PtyRequestType, true, clientPty)
		require.NoError(t, err)
		assert.True(t, ok)

		assert.Equal(t, PtyRequestType, nextOnDevice(t, received))

		_, err = client.SendRequest(ShellRequestType, true, nil)
		require.NoError(t, err)

		assert.Equal(t, ShellRequestType, nextOnDevice(t, received))
		assert.Empty(t, shellSetup(sess))
		assert.Zero(t, reopened(sess))
	})

	t.Run("starts no shell for a command from a client that waits for its pty to be granted", func(t *testing.T) {
		withFollowingRequestsGrace(t, time.Millisecond)

		client, sess, received := sessionOnShellStartingDevice(t)

		ok, err := client.SendRequest(PtyRequestType, true, clientPty)
		require.NoError(t, err)
		assert.True(t, ok)

		assert.Equal(t, PtyRequestType, nextOnDevice(t, received))

		ok, err = client.SendRequest(ExecRequestType, true, gossh.Marshal(struct{ Command string }{"id"}))
		require.NoError(t, err)
		assert.False(t, ok, "the command must get the device's answer")

		assert.Equal(t, ExecRequestType, nextOnDevice(t, received))
		assert.Empty(t, shellSetup(sess), "the device must not run a shell the client never asked for")
		assert.Zero(t, reopened(sess))
		assert.True(t, sess.called("SetSeatType"), "the command must mark the seat as an exec")
	})

	t.Run("keeps the plain channel when a command follows the pty", func(t *testing.T) {
		withFollowingRequestsGrace(t, 5*time.Second)

		client, sess, received := sessionOnShellStartingDevice(t)

		_, err := client.SendRequest(PtyRequestType, false, clientPty)
		require.NoError(t, err)

		ok, err := client.SendRequest(ExecRequestType, true, gossh.Marshal(struct{ Command string }{"id"}))
		require.NoError(t, err)
		assert.False(t, ok)

		assert.Equal(t, PtyRequestType, nextOnDevice(t, received))
		assert.Equal(t, ExecRequestType, nextOnDevice(t, received))
		assert.Empty(t, shellSetup(sess))
		assert.Zero(t, reopened(sess))
	})

	t.Run("keeps the channel the first pty runs on when a second pty and a shell follow", func(t *testing.T) {
		withFollowingRequestsGrace(t, 5*time.Second)

		client, sess, received := sessionOnShellStartingDevice(t)

		_, err := client.SendRequest(PtyRequestType, false, clientPty)
		require.NoError(t, err)
		_, err = client.SendRequest(PtyRequestType, false, clientPty)
		require.NoError(t, err)
		_, err = client.SendRequest(ShellRequestType, true, nil)
		require.NoError(t, err)

		assert.Equal(t, PtyRequestType, nextOnDevice(t, received))
		assert.Equal(t, PtyRequestType, nextOnDevice(t, received), "a second pty must reach the channel the pipe already runs on")
		assert.Equal(t, ShellRequestType, nextOnDevice(t, received))
		assert.Empty(t, shellSetup(sess))
	})

	t.Run("forwards the setup and shell to the device when it refused the shell on open", func(t *testing.T) {
		withFollowingRequestsGrace(t, 5*time.Second)

		sess := &fakeSession{startsShell: true, shellRefusals: 1} //nolint:exhaustruct // only the shell on open matters here
		client, received := sessionOnDevice(t, sess)

		_, err := client.SendRequest(PtyRequestType, false, clientPty)
		require.NoError(t, err)
		_, err = client.SendRequest("env", false, gossh.Marshal(struct{ Name, Value string }{"LANG", "C.UTF-8"}))
		require.NoError(t, err)
		_, err = client.SendRequest(ShellRequestType, true, nil)
		require.NoError(t, err)

		assert.Equal(t, PtyRequestType, nextOnDevice(t, received))
		assert.Equal(t, "env", nextOnDevice(t, received))
		assert.Equal(t, ShellRequestType, nextOnDevice(t, received))
		assert.Zero(t, reopened(sess))

		sess.mu.Lock()
		defer sess.mu.Unlock()

		assert.Zero(t, sess.shellRefusals, "the shell on open must have been tried")
	})
}
