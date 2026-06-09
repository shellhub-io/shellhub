//go:build !docker

package host

import (
	"errors"
	"net"
	"os/exec"
	"sync/atomic"
	"testing"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/agent/pkg/osauth"
	osauthMocks "github.com/shellhub-io/shellhub/agent/pkg/osauth/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	gossh "golang.org/x/crypto/ssh"
)

// fakeGosshConn is a minimal implementation of gossh.Conn that can be embedded
// in a gossh.ServerConn so tests can inject a ServerConn into the session context
// without needing a real SSH network connection.
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
	// Block until the test is done (goroutine is abandoned when the test
	// function returns; the Go runtime cleans it up). This is acceptable
	// because the caller only needs Wait to not panic.
	select {}
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
	// Mock osauth so LookupUser succeeds without touching /etc/passwd.
	osauthMock := &osauthMocks.Backend{}
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
	cmds := make(map[string]*exec.Cmd)
	s := NewSessioner(&deviceName, cmds, nil)

	sess := newFakeSession("session-exec-npty", "root")
	sess.isPty = false
	sess.command = []string{"/bin/true"}
	sess.rawCommand = "/bin/true"

	// Inject a fakeGosshConn so the serverConn context lookup succeeds and
	// Exec() can proceed past that check to reach session.Exit().
	fakeConn := &gossh.ServerConn{Conn: &fakeGosshConn{}}
	sess.ctx.(*testSSHContext).SetValue(gliderssh.ContextKeyConn, fakeConn)

	var retErr error

	assert.NotPanics(t, func() {
		retErr = s.Exec(sess)
	}, "Exec() must not panic for a succeeding non-PTY command")

	assert.NoError(t, retErr, "Exec() must return nil for a succeeding command")
	assert.Equal(t, int32(1), atomic.LoadInt32(&sess.exitCalled), "session.Exit must be called")
	assert.Equal(t, int32(0), atomic.LoadInt32(&sess.exitCode), "session.Exit must be called with code 0 for /bin/true")
}
