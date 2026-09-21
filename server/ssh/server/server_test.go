package server

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/pires/go-proxyproto"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/api/services"
	servicemocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/banner"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer/dialertest"
	"github.com/shellhub-io/shellhub/server/ssh/pkg/webhandoff"
	"github.com/shellhub-io/shellhub/server/ssh/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

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
func (s *stubContext) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(203, 0, 113, 7), Port: 4444, Zone: ""}
}
func (s *stubContext) LocalAddr() net.Addr { return nil }
func (s *stubContext) Permissions() *gliderssh.Permissions {
	return &gliderssh.Permissions{}
}
func (s *stubContext) SetValue(_, _ any) {}

func newStubCtx(user string) gliderssh.Context {
	return &stubContext{
		Context:   context.Background(),
		user:      user,
		sessionID: "test-session-id",
	}
}

const validSSHID = "user@namespace.device"

func stubDeps() bannerDeps {
	return bannerDeps{
		newSession: session.NewSession,
		evaluate: func(_ *session.Session, _ gliderssh.Context) error {
			return nil
		},
	}
}

func reachableDevice(t *testing.T) *servicemocks.MockService {
	t.Helper()

	service := servicemocks.NewMockService(t)
	service.On("LookupDevice", mock.Anything, "namespace", "device").
		Return(&models.Device{UID: "device-uid", Name: "device", TenantID: "tenant-id", Online: true}, nil) //nolint:exhaustruct // NewSession reads only the fields it resolves the namespace from
	service.On("GetNamespace", mock.Anything, "tenant-id").
		Return(&models.Namespace{Name: "namespace", TenantID: "tenant-id"}, nil) //nolint:exhaustruct // NewSession reads only the fields it names the session with

	return service
}

func bannerKind(message string) banner.Kind {
	kind := banner.Classify(message)

	return kind
}

func TestBannerHandlerInvalidSSHID(t *testing.T) {
	h := newBannerHandler(nil, nil, nil, session.NewRegistry())
	result := h(newStubCtx("not-a-valid-sshid"))

	assert.Equal(t, banner.KindInvalidSSHID, bannerKind(result),
		"BannerHandler must return the KindInvalidSSHID banner for a malformed SSHID")
}

func TestBannerHandlerNewSessionFailure(t *testing.T) {
	deps := stubDeps()
	deps.newSession = func(_ gliderssh.Context, _ dialer.TunnelDialer, _ services.Service, _ *webhandoff.Store, _ *session.Registry) (*session.Session, error) {
		return nil, errors.New("api unreachable")
	}

	h := newBannerHandlerWithDeps(nil, nil, nil, session.NewRegistry(), deps)
	result := h(newStubCtx(validSSHID))

	assert.Equal(t, banner.KindConnectionFailed, bannerKind(result),
		"BannerHandler must return KindConnectionFailed when NewSession fails")
}

func offlineDevice(t *testing.T) *servicemocks.MockService {
	t.Helper()

	service := servicemocks.NewMockService(t)
	service.On("LookupDevice", mock.Anything, "namespace", "device").
		Return(&models.Device{UID: "device-uid", Name: "device", TenantID: "tenant-id", Online: false}, nil) //nolint:exhaustruct // NewSession reads only the fields it resolves the namespace from
	service.On("GetNamespace", mock.Anything, "tenant-id").
		Return(&models.Namespace{Name: "namespace", TenantID: "tenant-id"}, nil) //nolint:exhaustruct // NewSession reads only the fields it names the session with

	return service
}

func TestBannerHandlerOfflineDevice(t *testing.T) {
	agent := dialertest.NewAgent(t)

	h := newBannerHandlerWithDeps(agent, offlineDevice(t), nil, session.NewRegistry(), stubDeps())
	result := h(newStubCtx(validSSHID))

	assert.Equal(t, banner.KindConnectionFailed, bannerKind(result),
		"BannerHandler must return KindConnectionFailed when the device is not reporting in")
	assert.Empty(t, agent.Dials(),
		"an offline device must be refused from what the lookup already knows, without a tunnel")
}

func TestBannerHandlerEvaluateFailure(t *testing.T) {
	deps := stubDeps()
	deps.evaluate = func(_ *session.Session, _ gliderssh.Context) error {
		return errors.New("firewall block")
	}

	h := newBannerHandlerWithDeps(dialertest.NewAgent(t), reachableDevice(t), nil, session.NewRegistry(), deps)
	result := h(newStubCtx(validSSHID))

	assert.Equal(t, banner.KindAccessDenied, bannerKind(result),
		"BannerHandler must return KindAccessDenied when Evaluate fails")
}

func TestBannerHandlerSuccess(t *testing.T) {
	agent := dialertest.NewAgent(t)

	h := newBannerHandlerWithDeps(agent, reachableDevice(t), nil, session.NewRegistry(), stubDeps())
	result := h(newStubCtx(validSSHID))

	assert.Empty(t, result,
		"BannerHandler must return an empty string on the success path")
	assert.Empty(t, agent.Dials(),
		"the tunnel to the device must not be opened before the client has authenticated")
}

func TestBannerHandlerRecoversFromPanic(t *testing.T) {
	deps := stubDeps()
	deps.newSession = func(_ gliderssh.Context, _ dialer.TunnelDialer, _ services.Service, _ *webhandoff.Store, _ *session.Registry) (*session.Session, error) {
		panic("boom")
	}

	h := newBannerHandlerWithDeps(nil, nil, nil, session.NewRegistry(), deps)

	var result string
	require.NotPanics(t, func() {
		result = h(newStubCtx(validSSHID))
	})

	assert.Equal(t, banner.KindConnectionFailed, bannerKind(result),
		"BannerHandler must recover from a panic and fail only the connection")
}

func TestLoopbackProxyPolicy(t *testing.T) {
	cases := []struct {
		name     string
		upstream net.Addr
		want     proxyproto.Policy
	}{
		{"ipv4 loopback", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 4444, Zone: ""}, proxyproto.USE},
		{"ipv6 loopback", &net.TCPAddr{IP: net.IPv6loopback, Port: 4444, Zone: ""}, proxyproto.USE},
		{"public peer", &net.TCPAddr{IP: net.IPv4(203, 0, 113, 7), Port: 4444, Zone: ""}, proxyproto.REJECT},
		{"docker bridge peer", &net.TCPAddr{IP: net.IPv4(172, 18, 0, 1), Port: 4444, Zone: ""}, proxyproto.REJECT},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := loopbackProxyPolicy(proxyproto.ConnPolicyOptions{Upstream: tc.upstream}) //nolint:exhaustruct
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestProxyListenerAcceptsWithoutProxyHeader(t *testing.T) {
	raw, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	proxy := newProxyListener(raw)
	defer proxy.Close() //nolint:errcheck

	deadline := clock.Now().Add(5 * time.Second)

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

	conn, err := new(net.Dialer).DialContext(t.Context(), "tcp", proxy.Addr().String())
	require.NoError(t, err)

	defer conn.Close() //nolint:errcheck

	conn.SetDeadline(deadline) //nolint:errcheck

	_, err = conn.Write([]byte("hello"))
	require.NoError(t, err)

	buf := make([]byte, 5)
	n, err := conn.Read(buf)
	require.NoError(t, err)
	assert.Equal(t, "hello", string(buf[:n]))

	assert.NoError(t, <-done)
}
