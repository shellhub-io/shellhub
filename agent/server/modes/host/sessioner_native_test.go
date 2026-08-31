//go:build !docker

package host

import (
	"errors"
	"net"
	"sync/atomic"
	"testing"

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
	osauth.DefaultBackend = osauthMock

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
	osauth.DefaultBackend = osauthMock

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
