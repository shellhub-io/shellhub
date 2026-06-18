//go:build !freebsd

package host

import (
	"context"
	"io"
	"net"
	"strings"
	"sync"
	"testing"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/agent/pkg/osauth"
	osauthMocks "github.com/shellhub-io/shellhub/agent/pkg/osauth/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubSession is a minimal implementation of gliderssh.Session for testing generateShellCmd.
type stubSession struct {
	user string
	envs []string
	ctx  gliderssh.Context
}

// gossh.Channel methods — not exercised by generateShellCmd.
func (s *stubSession) Read(_ []byte) (int, error) { return 0, io.EOF }

func (s *stubSession) Write(_ []byte) (int, error) { return 0, nil }

func (s *stubSession) Close() error { return nil }

func (s *stubSession) CloseWrite() error { return nil }

func (s *stubSession) SendRequest(_ string, _ bool, _ []byte) (bool, error) { return false, nil }

func (s *stubSession) Stderr() io.ReadWriter { return nil }

// gliderssh.Session methods.
func (s *stubSession) User() string { return s.user }

func (s *stubSession) RemoteAddr() net.Addr { return &net.TCPAddr{} }

func (s *stubSession) LocalAddr() net.Addr { return &net.TCPAddr{} }

func (s *stubSession) Environ() []string { return s.envs }

func (s *stubSession) Exit(_ int) error { return nil }

func (s *stubSession) Command() []string { return nil }

func (s *stubSession) RawCommand() string { return "" }

func (s *stubSession) Subsystem() string { return "" }

func (s *stubSession) PublicKey() gliderssh.PublicKey { return nil }

func (s *stubSession) Context() gliderssh.Context { return s.ctx }

func (s *stubSession) Permissions() gliderssh.Permissions { return gliderssh.Permissions{} }

func (s *stubSession) Pty() (gliderssh.Pty, <-chan gliderssh.Window, bool) {
	return gliderssh.Pty{}, nil, false
}

func (s *stubSession) Signals(_ chan<- gliderssh.Signal) {}

func (s *stubSession) Break(_ chan<- bool) {}

// Ensure stubSession satisfies the gliderssh.Session interface at compile time.
var _ gliderssh.Session = (*stubSession)(nil)

// stubSSHContext is a minimal gliderssh.Context for testing.
type stubSSHContext struct {
	context.Context
	*sync.Mutex
}

func (c *stubSSHContext) User() string { return "" }

func (c *stubSSHContext) SessionID() string { return "" }

func (c *stubSSHContext) ClientVersion() string { return "" }

func (c *stubSSHContext) ServerVersion() string { return "" }

func (c *stubSSHContext) RemoteAddr() net.Addr { return &net.TCPAddr{} }

func (c *stubSSHContext) LocalAddr() net.Addr { return &net.TCPAddr{} }

func (c *stubSSHContext) Permissions() *gliderssh.Permissions { return nil }

func (c *stubSSHContext) SetValue(_, _ interface{}) {}

// Ensure stubSSHContext satisfies the gliderssh.Context interface at compile time.
var _ gliderssh.Context = (*stubSSHContext)(nil)

func newStubContext() gliderssh.Context {
	return &stubSSHContext{
		Context: context.Background(),
		Mutex:   &sync.Mutex{},
	}
}

// TestGenerateShellCmdFiltersClientEnv verifies that generateShellCmd passes only the
// acceptClientEnv-filtered variables — not raw session.Environ() — to the command.
func TestGenerateShellCmdFiltersClientEnv(t *testing.T) {
	mock := osauthMocks.NewBackend(t)
	osauth.DefaultBackend = mock

	mock.On("LookupUser", "testuser").Return(&osauth.User{
		Username: "testuser",
		Shell:    "/bin/sh",
		HomeDir:  "/tmp",
	}, nil)
	mock.On("ListGroups", "testuser").Return([]uint32{}, nil)

	tests := []struct {
		name       string
		sessionEnv []string
		wantInEnv  []string
		wantAbsent []string
	}{
		{
			name:       "forbidden vars are filtered out, safe vars survive",
			sessionEnv: []string{"LD_PRELOAD=/evil.so", "LANG=en_US.UTF-8", "PATH=/tmp/evil", "LC_ALL=C"},
			wantInEnv:  []string{"LANG=en_US.UTF-8", "LC_ALL=C"},
			wantAbsent: []string{"LD_PRELOAD=/evil.so", "PATH=/tmp/evil"},
		},
		{
			name:       "SSH_AUTH_SOCK from client env is dropped but may be re-added from context",
			sessionEnv: []string{"SSH_AUTH_SOCK=/tmp/client.sock"},
			wantAbsent: []string{"SSH_AUTH_SOCK=/tmp/client.sock"},
		},
		{
			name:       "empty client env produces a cmd with no extra vars",
			sessionEnv: []string{},
			wantInEnv:  []string{},
			wantAbsent: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &stubSession{
				user: "testuser",
				envs: tt.sessionEnv,
				ctx:  newStubContext(),
			}

			cmd := generateShellCmd("mydevice", session, "xterm")
			require.NotNil(t, cmd, "generateShellCmd must not return nil for a valid user")

			for _, want := range tt.wantInEnv {
				assert.True(t, containsEnvEntry(cmd.Env, want),
					"expected %q in cmd.Env but not found; cmd.Env=%v", want, cmd.Env)
			}

			for _, absent := range tt.wantAbsent {
				assert.False(t, containsEnvEntry(cmd.Env, absent),
					"expected %q NOT in cmd.Env but it was found; cmd.Env=%v", absent, cmd.Env)
			}
		})
	}
}

// containsEnvEntry reports whether the exact entry e appears in envs.
func containsEnvEntry(envs []string, e string) bool {
	for _, v := range envs {
		if v == e {
			return true
		}
	}

	return false
}

// TestGenerateShellCmdExcludesForbiddenVarsPresentInCmdEnv verifies that dangerous
// variables injected through session.Environ() do not appear in cmd.Env even when
// they look like legitimate KEY=value pairs.
func TestGenerateShellCmdExcludesForbiddenVarsPresentInCmdEnv(t *testing.T) {
	mock := osauthMocks.NewBackend(t)
	osauth.DefaultBackend = mock

	mock.On("LookupUser", "testuser").Return(&osauth.User{
		Username: "testuser",
		Shell:    "/bin/sh",
		HomeDir:  "/tmp",
	}, nil)
	mock.On("ListGroups", "testuser").Return([]uint32{}, nil)

	dangerousEnv := []string{
		"LD_PRELOAD=/inject.so",
		"LD_AUDIT=/audit.so",
		"BASH_ENV=/etc/evil",
		"GODEBUG=inittrace=1",
		"HOME=/tmp/evil",
		"USER=root",
		"TERM=evil",
	}

	session := &stubSession{
		user: "testuser",
		envs: dangerousEnv,
		ctx:  newStubContext(),
	}

	cmd := generateShellCmd("mydevice", session, "xterm")
	require.NotNil(t, cmd)

	for _, danger := range dangerousEnv {
		name := strings.SplitN(danger, "=", 2)[0]

		for _, e := range cmd.Env {
			if strings.HasPrefix(e, name+"=") {
				// HOME, USER, TERM are set by NewCmd itself from the user struct, so they
				// can appear in cmd.Env — but must not carry the dangerous client-supplied value.
				// All other dangerous vars must be absent entirely.
				if name != "HOME" && name != "USER" && name != "TERM" {
					t.Errorf("dangerous var %q from session.Environ() reached cmd.Env as %q", danger, e)
				}
			}
		}
	}
}
