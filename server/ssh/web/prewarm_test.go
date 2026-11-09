package web

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func finishedPrewarm(err error) *prewarm {
	p := &prewarm{browser: newLateConn(), done: make(chan struct{}), err: err, cancel: func() {}}
	close(p.done)

	return p
}

func TestLateConnWaitsForTheBrowser(t *testing.T) {
	late := newLateConn()

	written := make(chan error, 1)

	go func() {
		_, err := late.WriteMessage(&Message{Kind: messageKindSignature, Data: "challenge"})
		written <- err
	}()

	select {
	case err := <-written:
		t.Fatalf("a message went out before any browser was there to read it: %v", err)
	case <-time.After(100 * time.Millisecond):
	}

	late.abandon()

	select {
	case err := <-written:
		require.ErrorIs(t, err, ErrBridgeCredentialsNotFound)
	case <-time.After(5 * time.Second):
		t.Fatal("a browser that never came left the write blocked")
	}
}

func TestPrewarmClaim(t *testing.T) {
	t.Run("hands over what the dial produced", func(t *testing.T) {
		p := finishedPrewarm(ErrAuthentication)

		_, err := p.claim(NewConn(nil))
		assert.ErrorIs(t, err, ErrAuthentication)
	})

	t.Run("refuses a claim once the token has expired", func(t *testing.T) {
		p := finishedPrewarm(nil)
		p.expire()

		_, err := p.claim(NewConn(nil))
		assert.ErrorIs(t, err, ErrBridgeCredentialsNotFound)
	})

	t.Run("belongs only to the address its token was created from", func(t *testing.T) {
		p := finishedPrewarm(nil)
		p.ip = "198.51.100.7"

		assert.True(t, p.from("198.51.100.7"))
		assert.False(t, p.from("203.0.113.9"), "a browser elsewhere must log in again, under its own address")
	})

	t.Run("stops a login still under way when the token expires", func(t *testing.T) {
		p := finishedPrewarm(nil)

		var cancelled bool
		p.cancel = func() { cancelled = true }

		p.expire()

		assert.True(t, cancelled, "a handshake blocked on the device must not outlive the token")
	})

	t.Run("does not expire once claimed", func(t *testing.T) {
		p := finishedPrewarm(nil)

		_, err := p.claim(NewConn(nil))
		require.NoError(t, err)

		var cancelled bool
		p.cancel = func() { cancelled = true }

		p.expire()

		assert.False(t, cancelled, "a claimed login must not be stopped by its token's expiry")
	})
}
