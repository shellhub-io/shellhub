//go:build !docker

package host

import (
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
