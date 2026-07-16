//go:build internal_api

package server

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/ssh/pkg/banner"
	"github.com/shellhub-io/shellhub/ssh/pkg/dialer"
	"github.com/shellhub-io/shellhub/ssh/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubContext is a minimal gliderssh.Context implementation for tests.
// It only populates the fields the BannerHandler reads.
type stubContext struct {
	context.Context
	sync.Mutex
	user      string
	sessionID string
}

func (s *stubContext) User() string          { return s.user }
func (s *stubContext) SessionID() string     { return s.sessionID }
func (s *stubContext) ClientVersion() string { return "" }
func (s *stubContext) ServerVersion() string { return "" }
func (s *stubContext) RemoteAddr() net.Addr  { return nil }
func (s *stubContext) LocalAddr() net.Addr   { return nil }
func (s *stubContext) Permissions() *gliderssh.Permissions {
	return &gliderssh.Permissions{}
}
func (s *stubContext) SetValue(_, _ interface{}) {}

func newStubCtx(user string) gliderssh.Context {
	return &stubContext{
		Context:   context.Background(),
		user:      user,
		sessionID: "test-session-id",
	}
}

// validSSHID is a syntactically correct SSHID that passes target.NewTarget.
// It is not connected to any real device, so only the stub deps are called.
const validSSHID = "user@namespace.device"

// stubDeps returns a bannerDeps whose three operations are all stubs that
// succeed by default. Individual tests override the stub they want to fail.
func stubDeps() bannerDeps {
	return bannerDeps{
		newSession: func(_ gliderssh.Context, _ *dialer.Dialer, _ cache.Cache) (*session.Session, error) {
			return &session.Session{}, nil //nolint:exhaustruct
		},
		dial: func(_ *session.Session, _ gliderssh.Context) error {
			return nil
		},
		evaluate: func(_ *session.Session, _ gliderssh.Context) error {
			return nil
		},
	}
}

func TestBannerHandlerInvalidSSHID(t *testing.T) {
	h := newBannerHandler(nil, nil)
	result := h(newStubCtx("not-a-valid-sshid"))

	assert.Equal(t, banner.KindInvalidSSHID, banner.Classify(result),
		"BannerHandler must return the KindInvalidSSHID banner for a malformed SSHID")
}

func TestBannerHandlerNewSessionFailure(t *testing.T) {
	deps := stubDeps()
	deps.newSession = func(_ gliderssh.Context, _ *dialer.Dialer, _ cache.Cache) (*session.Session, error) {
		return nil, errors.New("api unreachable")
	}

	h := newBannerHandlerWithDeps(nil, nil, deps)
	result := h(newStubCtx(validSSHID))

	assert.Equal(t, banner.KindConnectionFailed, banner.Classify(result),
		"BannerHandler must return KindConnectionFailed when NewSession fails")
}

func TestBannerHandlerDialFailure(t *testing.T) {
	deps := stubDeps()
	deps.dial = func(_ *session.Session, _ gliderssh.Context) error {
		return errors.New("device offline")
	}

	h := newBannerHandlerWithDeps(nil, nil, deps)
	result := h(newStubCtx(validSSHID))

	assert.Equal(t, banner.KindConnectionFailed, banner.Classify(result),
		"BannerHandler must return KindConnectionFailed when Dial fails")
}

func TestBannerHandlerEvaluateFailure(t *testing.T) {
	deps := stubDeps()
	deps.evaluate = func(_ *session.Session, _ gliderssh.Context) error {
		return errors.New("firewall block")
	}

	h := newBannerHandlerWithDeps(nil, nil, deps)
	result := h(newStubCtx(validSSHID))

	assert.Equal(t, banner.KindAccessDenied, banner.Classify(result),
		"BannerHandler must return KindAccessDenied when Evaluate fails")
}

func TestBannerHandlerSuccess(t *testing.T) {
	h := newBannerHandlerWithDeps(nil, nil, stubDeps())
	result := h(newStubCtx(validSSHID))

	assert.Empty(t, result,
		"BannerHandler must return an empty string on the success path")
}

func TestProxyListenerAcceptsWithoutProxyHeader(t *testing.T) {
	raw, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	proxy := newProxyListener(raw)
	defer proxy.Close() //nolint:errcheck

	deadline := time.Now().Add(5 * time.Second)

	done := make(chan error, 1)
	go func() {
		conn, err := proxy.Accept()
		if err != nil {
			done <- err

			return
		}
		defer conn.Close() //nolint:errcheck

		conn.SetDeadline(deadline) //nolint:errcheck

		buf := make([]byte, 5)
		n, err := conn.Read(buf)
		if err != nil {
			done <- err

			return
		}

		_, err = conn.Write(buf[:n])
		done <- err
	}()

	conn, err := net.Dial("tcp", proxy.Addr().String())
	require.NoError(t, err)

	defer conn.Close() //nolint:errcheck

	conn.SetDeadline(deadline) //nolint:errcheck

	_, err = conn.Write([]byte("hello"))
	assert.NoError(t, err)

	buf := make([]byte, 5)
	n, err := conn.Read(buf)
	assert.NoError(t, err)
	assert.Equal(t, "hello", string(buf[:n]))

	assert.NoError(t, <-done)
}
