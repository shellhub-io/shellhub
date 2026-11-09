package session

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

type passwordDevice struct {
	t         *testing.T
	signer    gossh.Signer
	publicKey string
	failFirst bool
	silent    bool
	stalls    bool

	mu       sync.Mutex
	dials    int
	closes   int
	attempts []string
	ended    chan struct{}
}

func newPasswordDevice(t *testing.T) *passwordDevice {
	t.Helper()

	pub, key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	signer, err := gossh.NewSignerFromKey(key)
	require.NoError(t, err)

	der, err := x509.MarshalPKIXPublicKey(pub)
	require.NoError(t, err)

	return &passwordDevice{
		t:         t,
		signer:    signer,
		publicKey: string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})),
		ended:     make(chan struct{}, 4),
	}
}

func (d *passwordDevice) DialTo(ctx context.Context, _, _ string, target dialer.Target) (net.Conn, error) { //nolint:ireturn // net.Conn is the interface TunnelDialer returns
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if _, closing := target.(dialer.SSHCloseTarget); closing {
		d.mu.Lock()
		d.closes++
		d.mu.Unlock()

		server, device := net.Pipe()
		_ = device.Close()

		return server, nil
	}

	d.mu.Lock()
	d.dials++
	first := d.dials == 1
	d.mu.Unlock()

	if first && d.failFirst {
		return nil, errors.New("the tunnel dropped")
	}

	if d.silent {
		return nil, errors.Join(dialer.ErrUnreachable, context.DeadlineExceeded)
	}

	server, device := tcpPair(d.t)

	go d.serve(device)

	return server, nil
}

func (d *passwordDevice) serve(device net.Conn) {
	defer func() { d.ended <- struct{}{} }()

	config := &gossh.ServerConfig{ //nolint:exhaustruct // a device that only takes passwords
		PasswordCallback: func(_ gossh.ConnMetadata, password []byte) (*gossh.Permissions, error) {
			d.mu.Lock()
			d.attempts = append(d.attempts, string(password))
			d.mu.Unlock()

			if string(password) != "secret" {
				return nil, errors.New("wrong password")
			}

			return &gossh.Permissions{}, nil //nolint:exhaustruct // nothing to grant
		},
	}
	config.AddHostKey(d.signer)
	conn, chans, reqs, err := gossh.NewServerConn(device, config)
	if err != nil {
		return
	}

	go gossh.DiscardRequests(reqs)

	for channel := range chans {
		if d.stalls {
			continue
		}

		_ = channel.Reject(gossh.Prohibited, "the test device runs nothing")
	}

	_ = conn.Close()
}

func (d *passwordDevice) closed() int {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.closes
}

func (d *passwordDevice) dialled() int {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.dials
}

func (d *passwordDevice) seen() (int, []string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	return d.dials, append([]string(nil), d.attempts...)
}

func warmSession(t *testing.T, device *passwordDevice) (*Session, *stubContext, context.CancelFunc) {
	t.Helper()

	sess := newTestSession(nil, device)
	sess.Device.PublicKey = device.publicKey

	ctx, cancel := context.WithCancel(t.Context())
	t.Cleanup(cancel)

	stub := &stubContext{Context: ctx, values: make(map[any]any)} //nolint:exhaustruct // the embedded mutex starts unlocked

	return sess, stub, cancel
}

func TestWarm(t *testing.T) {
	t.Run("authenticates only with the credential Auth delivers", func(t *testing.T) {
		device := newPasswordDevice(t)
		sess, ctx, _ := warmSession(t, device)

		sess.Warm(ctx)

		require.Eventually(t, func() bool { return device.dialled() == 1 }, 5*time.Second, 10*time.Millisecond)

		dials, attempts := device.seen()
		require.Equal(t, 1, dials, "the warm connection dials as soon as the session is evaluated")
		require.Empty(t, attempts, "nothing may authenticate before the client has")

		require.NoError(t, sess.connect(ctx, credential{password: "secret"}))

		dials, attempts = device.seen()
		assert.Equal(t, 1, dials, "the warm connection is the one that authenticates")
		assert.Equal(t, []string{"secret"}, attempts)
	})

	t.Run("reports a credential the device refuses without dialling again", func(t *testing.T) {
		device := newPasswordDevice(t)
		sess, ctx, _ := warmSession(t, device)

		sess.Warm(ctx)

		require.Error(t, sess.connect(ctx, credential{password: "wrong"}))

		dials, attempts := device.seen()
		assert.Equal(t, 1, dials)
		assert.Equal(t, []string{"wrong"}, attempts)
	})

	t.Run("dials again when the warm connection failed before authenticating", func(t *testing.T) {
		device := newPasswordDevice(t)
		device.failFirst = true
		sess, ctx, _ := warmSession(t, device)

		sess.Warm(ctx)

		require.NoError(t, sess.connect(ctx, credential{password: "secret"}))

		dials, attempts := device.seen()
		assert.Equal(t, 2, dials)
		assert.Equal(t, []string{"secret"}, attempts)
	})

	t.Run("does not dial again a device that did not answer the warm connection", func(t *testing.T) {
		device := newPasswordDevice(t)
		device.silent = true
		sess, ctx, _ := warmSession(t, device)

		sess.Warm(ctx)

		require.ErrorIs(t, sess.connect(ctx, credential{password: "secret"}), dialer.ErrUnreachable)

		dials, attempts := device.seen()
		assert.Equal(t, 1, dials)
		assert.Empty(t, attempts)
	})

	t.Run("closes the warm connection when the client leaves before authenticating", func(t *testing.T) {
		device := newPasswordDevice(t)
		sess, ctx, cancel := warmSession(t, device)

		sess.Warm(ctx)

		require.Eventually(t, func() bool { return device.dialled() == 1 }, 5*time.Second, 10*time.Millisecond)
		cancel()

		select {
		case <-device.ended:
		case <-time.After(5 * time.Second):
			t.Fatal("the warm connection outlived the client")
		}

		_, attempts := device.seen()
		assert.Empty(t, attempts)
		assert.Eventually(t, func() bool { return device.closed() == 1 }, 5*time.Second, 10*time.Millisecond,
			"the device must be asked to close the session the warm stream opened")
	})

	t.Run("closes a warm connection no client claimed in time and dials again for the late one", func(t *testing.T) {
		window := warmClaimWindow
		warmClaimWindow = 100 * time.Millisecond
		t.Cleanup(func() { warmClaimWindow = window })

		device := newPasswordDevice(t)
		sess, ctx, _ := warmSession(t, device)

		sess.Warm(ctx)

		select {
		case <-device.ended:
		case <-time.After(5 * time.Second):
			t.Fatal("the unclaimed warm connection outlived its window")
		}

		require.NoError(t, sess.connect(ctx, credential{password: "secret"}))

		dials, attempts := device.seen()
		assert.Equal(t, 2, dials)
		assert.Equal(t, []string{"secret"}, attempts)
		assert.Zero(t, device.closed(), "a client still authenticating may dial the same session again")
	})

	t.Run("finishes while the device has not answered a channel open", func(t *testing.T) {
		device := newPasswordDevice(t)
		device.stalls = true
		sess, ctx, _ := warmSession(t, device)
		sess.Events = NewEvents(sess.UID, sess.namespaceScope(), nil)

		sess.Warm(ctx)

		require.NoError(t, sess.connect(ctx, credential{password: "secret"}))

		opening := make(chan struct{})

		go func() {
			defer close(opening)

			_, _ = sess.NewAgentChannel(0)
		}()

		time.Sleep(100 * time.Millisecond)

		finished := make(chan struct{})

		go func() {
			defer close(finished)

			_ = sess.Finish()
		}()

		select {
		case <-finished:
		case <-time.After(5 * time.Second):
			t.Fatal("finishing the session waited for a channel the device never answered")
		}

		assert.Eventually(t, func() bool { return device.closed() == 1 }, 5*time.Second, 10*time.Millisecond,
			"the device must be asked to close the session")
	})

	t.Run("refuses a device connection dialled after the session finished", func(t *testing.T) {
		device := newPasswordDevice(t)
		sess, ctx, _ := warmSession(t, device)

		sess.agent.connMu.Lock()
		sess.agent.finished = true
		sess.agent.connMu.Unlock()

		require.ErrorIs(t, sess.connect(ctx, credential{password: "secret"}), ErrClientClosing)

		_, attempts := device.seen()
		assert.Empty(t, attempts)
		assert.Eventually(t, func() bool { return device.closed() == 1 }, 5*time.Second, 10*time.Millisecond,
			"the device must be asked to close the session nobody will finish")
	})
}
