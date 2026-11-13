package host

import (
	"bytes"
	"os/exec"
	"sync/atomic"
	"testing"

	"github.com/shellhub-io/shellhub/agent/pkg/osauth"
	osauthMocks "github.com/shellhub-io/shellhub/agent/pkg/osauth/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func sftpSessionForShell(t *testing.T, shell string) (*Sessioner, *fakeSession, *atomic.Bool) {
	t.Helper()

	osauthMock := &osauthMocks.MockBackend{}
	osauth.Set(t, osauthMock)

	osauthMock.On("LookupUser", mock.AnythingOfType("string")).Return(&osauth.User{
		UID:      1000,
		GID:      1000,
		Username: "git",
		Shell:    shell,
		HomeDir:  "/home/git",
	}, nil)

	var serverStarted atomic.Bool

	deviceName := "test-device"
	s := NewSessioner(&deviceName, func() *exec.Cmd {
		serverStarted.Store(true)

		return exec.CommandContext(t.Context(), "/nonexistent/sftp-server")
	})

	return s, newFakeSession("session-sftp", "git"), &serverStarted
}

func TestSFTP_RefusesAnAccountConfinedByItsLoginShell(t *testing.T) {
	shells := []string{
		"/usr/sbin/nologin",
		"/sbin/nologin",
		"/bin/false",
		"/usr/bin/false",
		"/usr/bin/git-shell",
		"/bin/rbash",
		"/bin/rksh",
		"/bin/rzsh",
	}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			s, sess, serverStarted := sftpSessionForShell(t, shell)

			var stderr bytes.Buffer
			sess.stderr = &stderr

			require.Error(t, s.SFTP(sess))

			assert.Equal(t, "This account is not available for SFTP.\n", stderr.String())
			assert.False(t, serverStarted.Load(), "the sftp server must not start")
			assert.Equal(t, int32(1), atomic.LoadInt32(&sess.exitCalled), "session.Exit must be called")
			assert.Equal(t, int32(1), atomic.LoadInt32(&sess.exitCode))
		})
	}
}

func TestSFTP_ServesAnAccountWhoseShellDoesNotConfineIt(t *testing.T) {
	shells := []string{"/bin/sh", "/bin/bash", "/usr/bin/zsh", "/usr/bin/scponly", ""}

	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			s, sess, serverStarted := sftpSessionForShell(t, shell)

			_ = s.SFTP(sess)

			assert.True(t, serverStarted.Load(), "the sftp server must start")
		})
	}
}

func TestSFTP_OptOutServesARestrictedShell(t *testing.T) {
	cases := []struct {
		value  string
		serves bool
	}{
		{value: "true", serves: true},
		{value: "TRUE", serves: true},
		{value: "1", serves: false},
		{value: "false", serves: false},
		{value: "", serves: false},
	}

	for _, tc := range cases {
		t.Run(tc.value, func(t *testing.T) {
			t.Setenv("SHELLHUB_SFTP_ALLOW_RESTRICTED_SHELLS", tc.value)

			s, sess, serverStarted := sftpSessionForShell(t, "/usr/sbin/nologin")

			_ = s.SFTP(sess)

			assert.Equal(t, tc.serves, serverStarted.Load())
		})
	}
}
