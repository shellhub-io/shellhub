package host

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"sync/atomic"
	"syscall"
	"testing"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/stretchr/testify/assert"
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

func (f *fakeSession) EmulatedPty() bool { return false }

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

	sessionID, ok := id.(string)
	if !ok || sessionID != "test-session-id" {
		t.Fatalf("expected session ID %q, got %q", "test-session-id", id)
	}
}

// TestPtyFailureHint verifies the ptyFailureHint helper:
//   - wrapping syscall.ENOTTY yields a non-empty hint
//   - an error whose message contains "inappropriate ioctl for device" yields a non-empty hint
//   - an unrelated error yields an empty string
func TestPtyFailureHint(t *testing.T) {
	const wantHint = "the system may not support PTY allocation — ensure /dev/ptmx is accessible and the agent is not in a restricted environment"

	tests := []struct {
		name      string
		err       error
		wantEmpty bool
	}{
		{
			name:      "wrapped ENOTTY yields hint",
			err:       fmt.Errorf("wrapped: %w", syscall.ENOTTY),
			wantEmpty: false,
		},
		{
			name:      "message contains 'inappropriate ioctl for device' yields hint",
			err:       errors.New("inappropriate ioctl for device"),
			wantEmpty: false,
		},
		{
			name:      "unrelated error yields empty hint",
			err:       errors.New("some unrelated error"),
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hint := PtyFailureHint(tt.err)

			if tt.wantEmpty {
				assert.Empty(t, hint, "expected empty hint for unrelated error")
			} else {
				assert.Equal(t, wantHint, hint, "expected diagnostic hint for pty failure error")
			}
		})
	}
}

func TestPtyStartOptionsBoundsTheOwnerID(t *testing.T) {
	original := geteuidFn
	t.Cleanup(func() { geteuidFn = original })

	geteuidFn = func() int { return 0 }

	cases := []struct {
		description string
		uid         uint32
		wantOwner   bool
	}{
		{
			description: "hands the pty over for an ordinary account",
			uid:         1000,
			wantOwner:   true,
		},
		{
			description: "hands it over at the largest id an int can hold everywhere",
			uid:         math.MaxInt32,
			wantOwner:   true,
		},
		{
			description: "skips the hand-over rather than wrap the id negative",
			uid:         math.MaxInt32 + 1,
			wantOwner:   false,
		},
		{
			description: "skips (uid_t)-1, which is not an account",
			uid:         math.MaxUint32,
			wantOwner:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			// WithJobControl is always present, so an owner option makes it two.
			opts := ptyStartOptions(tc.uid)
			if tc.wantOwner {
				assert.Len(t, opts, 2)
			} else {
				assert.Len(t, opts, 1)
			}
		})
	}
}

func TestPtyStartOptionsSkipsTheHandOverWhenNotRoot(t *testing.T) {
	original := geteuidFn
	t.Cleanup(func() { geteuidFn = original })

	geteuidFn = func() int { return 1000 }

	assert.Len(t, ptyStartOptions(1000), 1)
}
