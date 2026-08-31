package auth

import (
	"context"
	"net"
	"sync"
	"sync/atomic"
	"testing"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/server/ssh/session"
	"github.com/stretchr/testify/assert"
	gossh "golang.org/x/crypto/ssh"
)

type stubContext struct {
	context.Context
	sync.Mutex
	values map[any]any
}

func newStubContext() *stubContext {
	return &stubContext{
		Context: context.Background(),
		values:  make(map[any]any),
	}
}

func (s *stubContext) User() string          { return "user@namespace.device" }
func (s *stubContext) SessionID() string     { return "test-session-id" }
func (s *stubContext) ClientVersion() string { return "" }
func (s *stubContext) ServerVersion() string { return "" }
func (s *stubContext) RemoteAddr() net.Addr  { return nil }
func (s *stubContext) LocalAddr() net.Addr   { return nil }

func (s *stubContext) Permissions() *gliderssh.Permissions {
	return &gliderssh.Permissions{}
}

func (s *stubContext) SetValue(key, val any) {
	s.Lock()
	defer s.Unlock()

	s.values[key] = val
}

func (s *stubContext) Value(key any) any {
	s.Lock()
	defer s.Unlock()

	return s.values[key]
}

type recordingConn struct {
	net.Conn
	closed atomic.Bool
}

func (c *recordingConn) Close() error {
	c.closed.Store(true)

	return nil
}

func testKey(t *testing.T) gliderssh.PublicKey {
	t.Helper()

	const authorized = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIJ0eKgLLYAqQBiKtHqBQF3JzKlXjmVjJqB0eqzKlEnCf test"

	key, _, _, _, err := gossh.ParseAuthorizedKey([]byte(authorized))
	if err != nil {
		t.Fatalf("failed to parse the test key: %v", err)
	}

	return key
}

// A connection that has not been evaluated has nothing to authenticate against,
// so it is dropped rather than refused.
func TestHandlersDropAConnectionThatHasNotBeenEvaluated(t *testing.T) {
	key := testKey(t)

	cases := []struct {
		name string
		call func(gliderssh.Context) bool
	}{
		{
			name: "password",
			call: func(ctx gliderssh.Context) bool { return PasswordHandler(ctx, "secret") },
		},
		{
			name: "public key offer",
			call: func(ctx gliderssh.Context) bool { return PublicKeyOffer(ctx, key) },
		},
		{
			name: "public key verified",
			call: func(ctx gliderssh.Context) bool { return PublicKeyVerified(ctx, key) },
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newStubContext()
			conn := new(recordingConn)
			session.StoreConn(ctx, conn)

			assert.False(t, tt.call(ctx), "the handler must refuse")
			assert.True(t, conn.closed.Load(), "the handler must drop the connection")
		})
	}
}

// The guard must survive a context that never had a connection stored: a panic
// here takes down the whole process, since the HTTP API shares it.
func TestHandlersWithoutAStoredConnDoNotPanic(t *testing.T) {
	key := testKey(t)

	assert.NotPanics(t, func() {
		assert.False(t, PasswordHandler(newStubContext(), "secret"))
		assert.False(t, PublicKeyOffer(newStubContext(), key))
		assert.False(t, PublicKeyVerified(newStubContext(), key))
	})
}
