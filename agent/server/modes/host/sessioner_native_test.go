//go:build !docker

package host

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/agent/pkg/osauth"
	osauthMocks "github.com/shellhub-io/shellhub/agent/pkg/osauth/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

type fakeGosshConn struct{}

func (f *fakeGosshConn) User() string          { return "root" }
func (f *fakeGosshConn) SessionID() []byte     { return nil }
func (f *fakeGosshConn) ClientVersion() []byte { return nil }
func (f *fakeGosshConn) ServerVersion() []byte { return nil }
func (f *fakeGosshConn) RemoteAddr() net.Addr  { return &net.TCPAddr{} }
func (f *fakeGosshConn) LocalAddr() net.Addr   { return &net.TCPAddr{} }
func (f *fakeGosshConn) SendRequest(_ string, _ bool, _ []byte) (bool, []byte, error) {
	return false, nil, nil
}

func (f *fakeGosshConn) OpenChannel(_ string, _ []byte) (gossh.Channel, <-chan *gossh.Request, error) {
	return nil, nil, errors.New("not implemented")
}

func (f *fakeGosshConn) Close() error { return nil }

// Wait blocks until the returned channel is closed — used by tests to
// control when the "kill on disconnect" goroutine unblocks.
func (f *fakeGosshConn) Wait() error {
	select {}
}

// TestShell_DeniedCredentialSwitch verifies that Shell() returns a non-nil error
// and calls session.Exit(1) exactly once when checkCredentialSwitchFn returns an
// error. The gate must fire before any call to session.Pty() or
// generateShellCmd, so the test does NOT need a real ServerConn in the session
// context.
func TestShell_DeniedCredentialSwitch(t *testing.T) {
	origCheckCredentialSwitch := checkCredentialSwitchFn

	t.Cleanup(func() {
		checkCredentialSwitchFn = origCheckCredentialSwitch
	})

	checkCredentialSwitchFn = func() error {
		return errors.New("setgroups denied in unprivileged user namespace")
	}

	deviceName := "test-device"
	s := NewSessioner(&deviceName, nil)

	sess := newFakeSession("session-cred-switch", "root")

	var retErr error

	assert.NotPanics(t, func() {
		retErr = s.Shell(sess)
	}, "Shell() must not panic when credential switch is denied")

	require.Error(t, retErr, "Shell() must return a non-nil error when credential switch is denied")
	assert.Equal(t, int32(1), atomic.LoadInt32(&sess.exitCalled), "session.Exit must be called once")
	assert.Equal(t, int32(1), atomic.LoadInt32(&sess.exitCode), "session.Exit must be called with code 1")
}

// TestExec_DeniedCredentialSwitch verifies that Exec() returns a non-nil error,
// calls session.Exit(1) exactly once with code 1, when checkCredentialSwitchFn
// returns an error. The gate must fire as the FIRST statement inside Exec(),
// before LookupUser, session.Pty(), command.NewCmd, initPtyFn, or cmd.Start.
func TestExec_DeniedCredentialSwitch(t *testing.T) {
	origCheckCredentialSwitch := checkCredentialSwitchFn

	t.Cleanup(func() {
		checkCredentialSwitchFn = origCheckCredentialSwitch
	})

	checkCredentialSwitchFn = func() error {
		return errors.New("setgroups denied in unprivileged user namespace")
	}

	deviceName := "test-device"
	s := NewSessioner(&deviceName, nil)

	sess := newFakeSession("session-exec-cred-switch", "root")
	sess.command = []string{"/bin/true"}
	sess.rawCommand = "/bin/true"

	var retErr error

	assert.NotPanics(t, func() {
		retErr = s.Exec(sess)
	}, "Exec() must not panic when credential switch is denied")

	require.Error(t, retErr, "Exec() must return a non-nil error when credential switch is denied")
	assert.Equal(t, int32(1), atomic.LoadInt32(&sess.exitCalled), "session.Exit must be called once")
	assert.Equal(t, int32(1), atomic.LoadInt32(&sess.exitCode), "session.Exit must be called with code 1")
}

// TestHeredoc_StartFailure verifies that Heredoc() handles cmd.Start() failure
// without panicking. Before the fix, two nil-derefs were possible:
//  1. The kill-goroutine was launched before cmd.Start(), so cmd.Process was nil
//     when serverConn.Wait() returned, causing cmd.Process.Kill() to panic.
//  2. After cmd.Start() failed cmd.Wait() also failed and cmd.ProcessState was nil,
//     causing cmd.ProcessState.ExitCode() to panic.
//
// After the fix: cmd.Start() failure triggers an early-return — log.Warn + session.Exit(1)
// + return err — BEFORE launching any goroutine or reaching cmd.ProcessState.ExitCode().
func TestHeredoc_StartFailure(t *testing.T) {
	osauthMock := &osauthMocks.MockBackend{}
	osauth.Set(t, osauthMock)

	fakeUser := &osauth.User{
		UID:      0,
		GID:      0,
		Username: "root",
		Shell:    "/nonexistent/shell-that-does-not-exist",
		HomeDir:  "/root",
	}

	osauthMock.On("LookupUser", mock.AnythingOfType("string")).Return(fakeUser, nil).Maybe()
	osauthMock.On("ListGroups", mock.AnythingOfType("string")).Return([]uint32{}, nil).Maybe()

	deviceName := "test-device"
	s := NewSessioner(&deviceName, nil)

	sess := newFakeSession("session-heredoc-start-fail", "root")

	fakeConn := &gossh.ServerConn{Conn: &fakeGosshConn{}}
	testCtx, ok := sess.ctx.(*testSSHContext)
	require.True(t, ok)
	testCtx.SetValue(gliderssh.ContextKeyConn, fakeConn)

	var retErr error

	assert.NotPanics(t, func() {
		retErr = s.Heredoc(sess)
	}, "Heredoc() must not panic when cmd.Start() fails")

	require.Error(t, retErr, "Heredoc() must return a non-nil error when cmd.Start() fails")
	assert.Equal(t, int32(1), atomic.LoadInt32(&sess.exitCalled), "session.Exit must be called once")
	assert.Equal(t, int32(1), atomic.LoadInt32(&sess.exitCode), "session.Exit must be called with exit code 1")
}

// TestHeredoc_DeniedCredentialSwitch verifies that Heredoc() returns a non-nil
// error, calls session.Exit(1) exactly once with code 1, and returns before any
// further work (generateShellCmd, pipe creation, serverConn lookup, cmd.Start)
// when checkCredentialSwitchFn returns an error. Because the gate fires as the
// FIRST statement, no ServerConn injection into the session context is needed.
func TestHeredoc_DeniedCredentialSwitch(t *testing.T) {
	origCheckCredentialSwitch := checkCredentialSwitchFn

	t.Cleanup(func() {
		checkCredentialSwitchFn = origCheckCredentialSwitch
	})

	checkCredentialSwitchFn = func() error {
		return errors.New("setgroups denied in unprivileged user namespace")
	}

	deviceName := "test-device"
	s := NewSessioner(&deviceName, nil)

	sess := newFakeSession("session-heredoc-cred-switch", "root")

	var retErr error

	assert.NotPanics(t, func() {
		retErr = s.Heredoc(sess)
	}, "Heredoc() must not panic when credential switch is denied")

	require.Error(t, retErr, "Heredoc() must return a non-nil error when credential switch is denied")
	assert.Equal(t, int32(1), atomic.LoadInt32(&sess.exitCalled), "session.Exit must be called once")
	assert.Equal(t, int32(1), atomic.LoadInt32(&sess.exitCode), "session.Exit must be called with code 1")
}

// TestExec_NonPty_SucceedingCommand is a regression guard for the non-PTY path of
// Exec(). It verifies that:
//   - a real command starts and completes
//   - session.Exit is called with the command's actual exit code (0 for success)
//   - cmd.ProcessState nil-guard does not break the happy path
//
// This test is constrained to the native (non-docker) build because the docker
// variant of command.NewCmd wraps the binary inside /usr/bin/nsenter, which exits
// non-zero in environments that are not a real Docker-on-host setup (CI, dev
// containers, etc.).
func TestExec_NonPty_SucceedingCommand(t *testing.T) {
	osauthMock := &osauthMocks.MockBackend{}
	osauth.Set(t, osauthMock)

	fakeUser := &osauth.User{
		UID:      0,
		GID:      0,
		Username: "root",
		Shell:    "/bin/sh",
		HomeDir:  "/root",
	}

	osauthMock.On("LookupUser", mock.AnythingOfType("string")).Return(fakeUser, nil).Maybe()
	osauthMock.On("ListGroups", mock.AnythingOfType("string")).Return([]uint32{}, nil).Maybe()

	deviceName := "test-device"
	s := NewSessioner(&deviceName, nil)

	sess := newFakeSession("session-exec-npty", "root")
	sess.isPty = false
	sess.command = []string{"/bin/true"}
	sess.rawCommand = "/bin/true"

	fakeConn := &gossh.ServerConn{Conn: &fakeGosshConn{}}
	testCtx, ok := sess.ctx.(*testSSHContext)
	require.True(t, ok)
	testCtx.SetValue(gliderssh.ContextKeyConn, fakeConn)

	var retErr error

	assert.NotPanics(t, func() {
		retErr = s.Exec(sess)
	}, "Exec() must not panic for a succeeding non-PTY command")

	require.NoError(t, retErr, "Exec() must return nil for a succeeding command")
	assert.Equal(t, int32(1), atomic.LoadInt32(&sess.exitCalled), "session.Exit must be called")
	assert.Equal(t, int32(0), atomic.LoadInt32(&sess.exitCode), "session.Exit must be called with code 0 for /bin/true")
}

func TestSFTP_SendsTheSFTPServerExitCode(t *testing.T) {
	cases := []struct {
		name        string
		code        int32
		requireErrs require.ErrorAssertionFunc
	}{
		{name: "sftp server exits cleanly", code: 0, requireErrs: require.NoError},
		{name: "sftp server fails", code: 3, requireErrs: require.Error},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			osauthMock := &osauthMocks.MockBackend{}
			osauth.Set(t, osauthMock)

			osauthMock.On("LookupUser", mock.AnythingOfType("string")).Return(&osauth.User{
				UID:      0,
				GID:      0,
				Username: "root",
				Shell:    "/bin/sh",
				HomeDir:  "/root",
			}, nil)

			deviceName := "test-device"
			s := NewSessioner(&deviceName, func() *exec.Cmd {
				return exec.CommandContext(t.Context(), "/bin/sh", "-c", fmt.Sprintf("exit %d", tc.code)) //nolint:gosec // the exit code comes from the test table
			})

			sess := newFakeSession("session-sftp", "root")

			testCtx, ok := sess.ctx.(*testSSHContext)
			require.True(t, ok)
			testCtx.SetValue(gliderssh.ContextKeyConn, &gossh.ServerConn{Conn: &fakeGosshConn{}})

			tc.requireErrs(t, s.SFTP(sess))

			assert.Equal(t, int32(1), atomic.LoadInt32(&sess.exitCalled), "session.Exit must be called")
			assert.Equal(t, tc.code, atomic.LoadInt32(&sess.exitCode))
		})
	}
}

func TestExec_DeliversASignalToTheCommand(t *testing.T) {
	signals, done, sess := startInterruptTrapExec(t)

	signals <- gliderssh.SIGINT

	requireExecExitedWith(t, done, sess, 3)
}

func TestExec_IgnoresAnUnknownSignal(t *testing.T) {
	signals, done, sess := startInterruptTrapExec(t)

	signals <- gliderssh.Signal("BOGUS")

	select {
	case <-done:
		t.Fatal("an unknown signal ended the command")
	case <-time.After(500 * time.Millisecond):
	}

	signals <- gliderssh.SIGINT

	requireExecExitedWith(t, done, sess, 3)
}

func startInterruptTrapExec(t *testing.T) (chan<- gliderssh.Signal, <-chan struct{}, *fakeSession) {
	t.Helper()

	osauthMock := &osauthMocks.MockBackend{}
	osauth.Set(t, osauthMock)

	osauthMock.On("LookupUser", mock.AnythingOfType("string")).Return(&osauth.User{
		UID:      0,
		GID:      0,
		Username: "root",
		Shell:    "/bin/sh",
		HomeDir:  "/root",
	}, nil).Maybe()
	osauthMock.On("ListGroups", mock.AnythingOfType("string")).Return([]uint32{}, nil).Maybe()

	ready := filepath.Join(t.TempDir(), "ready")

	deviceName := "test-device"
	s := NewSessioner(&deviceName, nil)

	sess := newFakeSession("session-exec-signal", "root")
	sess.command = []string{"sh"}
	sess.rawCommand = fmt.Sprintf(`trap 'exit 3' INT; touch %s; for i in $(seq 1 100); do sleep 0.1; done`, ready)
	sess.signals = make(chan chan<- gliderssh.Signal, 1)

	testCtx, ok := sess.ctx.(*testSSHContext)
	require.True(t, ok)
	testCtx.SetValue(gliderssh.ContextKeyConn, &gossh.ServerConn{Conn: &fakeGosshConn{}})

	done := make(chan struct{})

	go func() {
		defer close(done)

		_ = s.Exec(sess)
	}()

	var signals chan<- gliderssh.Signal

	select {
	case signals = <-sess.signals:
	case <-time.After(10 * time.Second):
		t.Fatal("Exec never asked the session for its signals")
	}

	require.Eventually(t, func() bool {
		_, err := os.Stat(ready)

		return err == nil
	}, 10*time.Second, 10*time.Millisecond, "the command never installed its trap")

	return signals, done, sess
}

func requireExecExitedWith(t *testing.T, done <-chan struct{}, sess *fakeSession, code int32) {
	t.Helper()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("the command did not react to the signal")
	}

	assert.Equal(t, code, atomic.LoadInt32(&sess.exitCode))
}
