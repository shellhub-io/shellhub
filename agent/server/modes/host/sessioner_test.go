package host

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"testing"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/agent/pkg/osauth"
	osauthMocks "github.com/shellhub-io/shellhub/agent/pkg/osauth/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	gossh "golang.org/x/crypto/ssh"
)

// devNull is a no-op io.ReadWriter used as a stub for Stderr() in fakeSession.
type devNull struct{}

func (devNull) Read([]byte) (int, error) { return 0, io.EOF }

func (devNull) Write(p []byte) (int, error) { return len(p), nil }

// fakeSession is a hand-rolled test double that satisfies the gliderssh.Session
// interface. It allows unit tests to inject controlled values without spinning
// up a real SSH connection.
type fakeSession struct {
	user       string
	environ    []string
	pty        gliderssh.Pty
	winCh      <-chan gliderssh.Window
	isPty      bool
	remoteAddr net.Addr
	localAddr  net.Addr
	ctx        gliderssh.Context
	command    []string
	rawCommand string

	// exitCalled tracks whether Exit() was called and with what code.
	exitCalled int32 // atomic: 0 = not called, 1 = called
	exitCode   int32 // atomic: last code passed to Exit()
}

// --- gossh.Channel methods ---

func (f *fakeSession) Read(data []byte) (int, error) { return 0, io.EOF }

func (f *fakeSession) Write(data []byte) (int, error) { return len(data), nil }

func (f *fakeSession) Close() error { return nil }

func (f *fakeSession) CloseWrite() error { return nil }

func (f *fakeSession) SendRequest(name string, wantReply bool, payload []byte) (bool, error) {
	return false, nil
}

func (f *fakeSession) Stderr() io.ReadWriter { return devNull{} }

// --- gliderssh.Session methods ---

func (f *fakeSession) User() string { return f.user }

func (f *fakeSession) RemoteAddr() net.Addr { return f.remoteAddr }

func (f *fakeSession) LocalAddr() net.Addr { return f.localAddr }

func (f *fakeSession) Environ() []string { return f.environ }

func (f *fakeSession) Exit(code int) error {
	atomic.StoreInt32(&f.exitCalled, 1)
	atomic.StoreInt32(&f.exitCode, int32(code)) //nolint:gosec

	return nil
}

func (f *fakeSession) Command() []string { return f.command }

func (f *fakeSession) RawCommand() string { return f.rawCommand }

func (f *fakeSession) Subsystem() string { return "" }

func (f *fakeSession) PublicKey() gliderssh.PublicKey { return nil }

func (f *fakeSession) Context() gliderssh.Context { return f.ctx }

func (f *fakeSession) Permissions() gliderssh.Permissions {
	return gliderssh.Permissions{}
}

func (f *fakeSession) Pty() (gliderssh.Pty, <-chan gliderssh.Window, bool) {
	return f.pty, f.winCh, f.isPty
}

func (f *fakeSession) Signals(c chan<- gliderssh.Signal) {}

func (f *fakeSession) Break(c chan<- bool) {}

// newFakeSession constructs a fakeSession whose Context() returns a
// testSSHContext with gliderssh.ContextKeySessionID already set.
func newFakeSession(sessionID, user string) *fakeSession {
	ctx := &testSSHContext{
		Context:   context.Background(),
		user:      user,
		sessionID: sessionID,
	}
	ctx.SetValue(gliderssh.ContextKeySessionID, sessionID)

	return &fakeSession{
		user:       user,
		remoteAddr: &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345},
		localAddr:  &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 22},
		ctx:        ctx,
		winCh:      make(chan gliderssh.Window),
	}
}

// Ensure gossh.Channel is satisfied (compile-time check via assignment).
var _ gossh.Channel = (*fakeSession)(nil)

// TestFakeSessionCompiles is a compile-time check that fakeSession fully
// implements gliderssh.Session. The test body is intentionally empty — a build
// failure means the interface is incomplete.
func TestFakeSessionCompiles(t *testing.T) {
	t.Helper()

	var _ gliderssh.Session = (*fakeSession)(nil)

	sess := newFakeSession("test-session-id", "testuser")

	id := sess.Context().Value(gliderssh.ContextKeySessionID)
	if id == nil {
		t.Fatal("expected ContextKeySessionID to be set in fakeSession.Context()")
	}

	if id.(string) != "test-session-id" {
		t.Fatalf("expected session ID %q, got %q", "test-session-id", id)
	}
}

// TestShell_StartPtyError verifies that Shell() returns an error and does NOT
// panic when startPtyFn fails (e.g. "ptmx: inappropriate ioctl for device").
// Before the fix the nil *os.File returned by the stub was dereferenced, causing
// a panic.  After the fix there is an early-return guard that logs the error,
// calls session.Exit(1), and returns a wrapped error — all before any call that
// would dereference pts.
func TestShell_StartPtyError(t *testing.T) {
	// Restore the real startPtyFn after the test so other tests are not affected.
	origStartPty := startPtyFn

	t.Cleanup(func() {
		startPtyFn = origStartPty
	})

	// Stub startPtyFn to simulate a pty open failure.
	startPtyFn = func(_ *exec.Cmd, _ io.ReadWriter, _ <-chan gliderssh.Window) (*os.File, error) {
		return nil, errors.New("ptmx: inappropriate ioctl for device")
	}

	// Mock osauth so generateShellCmd (and any subsequent LookupUser call) can
	// return a real user without touching /etc/passwd.
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

	sess := newFakeSession("session-1", "root")

	var retErr error

	assert.NotPanics(t, func() {
		retErr = s.Shell(sess)
	})

	assert.NotNil(t, retErr, "Shell() must return a non-nil error when startPty fails")
	assert.True(t, strings.Contains(retErr.Error(), "pty"), "error should mention 'pty', got: %s", retErr.Error())
	assert.Empty(t, s.cmds, "s.cmds must be empty — the session must not have been registered")
}

// TestExec_InitPtyError verifies that Exec() with sIsPty=true does NOT panic and
// returns a non-nil error when initPtyFn fails. It also asserts that session.Exit(1)
// is called — the early-return guard must fire before any defer on nil *os.File.
func TestExec_InitPtyError(t *testing.T) {
	origInitPty := initPtyFn

	t.Cleanup(func() {
		initPtyFn = origInitPty
	})

	// Stub initPtyFn to simulate a pty open failure.
	initPtyFn = func(_ *exec.Cmd, _ io.ReadWriter, _ <-chan gliderssh.Window) (*os.File, *os.File, error) {
		return nil, nil, errors.New("ptmx: inappropriate ioctl for device")
	}

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

	sess := newFakeSession("session-exec-pty", "root")
	sess.isPty = true
	sess.command = []string{"/bin/echo", "hello"}
	sess.rawCommand = "/bin/echo hello"

	var retErr error

	assert.NotPanics(t, func() {
		retErr = s.Exec(sess)
	}, "Exec() must not panic when initPtyFn fails")

	assert.NotNil(t, retErr, "Exec() must return a non-nil error when initPtyFn fails")
	assert.True(
		t,
		strings.Contains(retErr.Error(), "pty") || strings.Contains(retErr.Error(), "ptmx"),
		"error should mention the pty failure, got: %s", retErr.Error(),
	)
	assert.Equal(t, int32(1), atomic.LoadInt32(&sess.exitCalled), "session.Exit must be called once")
	assert.Equal(t, int32(1), atomic.LoadInt32(&sess.exitCode), "session.Exit must be called with code 1")
}
