package client

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const deviceToken = "Bearer device-token"

type tokenRecorder struct {
	mu       sync.Mutex
	received []string
}

func (r *tokenRecorder) record(req *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.received = append(r.received, req.Header.Get("Authorization"))
}

func (r *tokenRecorder) tokens() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return append([]string(nil), r.received...)
}

func newServer(t *testing.T, secure bool, handler http.Handler) *httptest.Server {
	t.Helper()

	server := httptest.NewUnstartedServer(handler)
	if secure {
		server.StartTLS()
	} else {
		server.Start()
	}
	t.Cleanup(server.Close)

	return server
}

func newTarget(t *testing.T, secure bool) (*httptest.Server, *tokenRecorder) {
	t.Helper()

	recorder := new(tokenRecorder)
	upgrader := websocket.Upgrader{}

	server := newServer(t, secure, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		recorder.record(req)

		if !websocket.IsWebSocketUpgrade(req) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"accepted"}`))

			return
		}

		conn, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			return
		}

		_ = conn.Close()
	}))

	return server, recorder
}

type reply struct {
	status   int
	location string
}

type redirector struct {
	server *httptest.Server
	answer atomic.Pointer[reply]
	hits   atomic.Int32
}

func newRedirector(t *testing.T, secure bool) *redirector {
	t.Helper()

	r := new(redirector)
	r.server = newServer(t, secure, http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		r.hits.Add(1)
		current := r.answer.Load()
		if current.location != "" {
			w.Header().Set("Location", current.location)
		}
		w.WriteHeader(current.status)
	}))

	return r
}

func (r *redirector) redirectTo(location string) {
	r.answerWith(http.StatusTemporaryRedirect, location)
}

func (r *redirector) answerWith(status int, location string) {
	r.answer.Store(&reply{status: status, location: location})
}

func onOtherHost(serverURL string) string {
	return strings.Replace(serverURL, "127.0.0.1", "localhost", 1)
}

func testTLSConfig(t *testing.T) *tls.Config {
	t.Helper()

	server := newServer(t, true, http.NotFoundHandler())
	httpTransport, ok := server.Client().Transport.(*http.Transport)
	require.True(t, ok)

	return httpTransport.TLSClientConfig
}

func trustTestCertificates(t *testing.T) {
	t.Helper()

	original := websocket.DefaultDialer
	t.Cleanup(func() { websocket.DefaultDialer = original })

	dialer := *original
	dialer.TLSClientConfig = testTLSConfig(t)
	websocket.DefaultDialer = &dialer
}

func withTestTransport(t *testing.T, retries int) Opt {
	t.Helper()

	config := testTLSConfig(t)

	return func(c *client) error {
		c.http.SetRetryCount(retries)
		c.http.SetRetryWaitTime(time.Millisecond)
		c.http.SetRetryMaxWaitTime(time.Millisecond)
		c.http.SetTimeout(5 * time.Second)
		c.http.SetTLSClientConfig(config)

		return nil
	}
}

type transport struct {
	name string
	send func(t *testing.T, address string) error
}

func transports() []transport {
	return []transport{
		{
			name: "tunnel dial",
			send: func(t *testing.T, address string) error {
				t.Helper()

				trustTestCertificates(t)

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				conn, _, err := DialContext(ctx, address, http.Header{"Authorization": {deviceToken}}) //nolint:bodyclose // gorilla/websocket documents that the handshake response body need not be closed.
				if err != nil {
					return err
				}

				return conn.Close()
			},
		},
		{
			name: "api client",
			send: func(t *testing.T, address string) error {
				t.Helper()

				cli, err := NewClient(address, withTestTransport(t, 0))
				require.NoError(t, err)

				_, err = cli.GetDeviceAuthStatus("device-token")

				return err
			},
		},
	}
}

func TestTheDeviceTokenGoesOnlyToTheHostTheAgentWasGiven(t *testing.T) {
	tests := []struct {
		description    string
		redirectSecure bool
		targetSecure   bool
		location       func(target *httptest.Server) string
		status         int
		reachesTarget  bool
		refusal        error
		failure        string
	}{
		{
			description:   "a redirect to another path on the same host is followed",
			location:      func(target *httptest.Server) string { return target.URL + "/elsewhere" },
			reachesTarget: true,
		},
		{
			description:   "a permanent redirect to another path on the same host is followed",
			status:        http.StatusPermanentRedirect,
			location:      func(target *httptest.Server) string { return target.URL + "/elsewhere" },
			reachesTarget: true,
		},
		{
			description:   "a redirect from http to https on the same host is followed",
			targetSecure:  true,
			location:      func(target *httptest.Server) string { return target.URL },
			reachesTarget: true,
		},
		{
			description: "a redirect to another host is refused",
			location:    func(target *httptest.Server) string { return onOtherHost(target.URL) },
			refusal:     errRedirectOtherHost,
		},
		{
			description:    "a redirect from https to http on the same host is refused",
			redirectSecure: true,
			location:       func(target *httptest.Server) string { return target.URL },
			refusal:        errRedirectDowngrade,
		},
		{
			description: "a redirect to a location that does not parse is refused",
			location:    func(*httptest.Server) string { return "http://[::1" },
			failure:     "missing ']' in host",
		},
	}

	for _, via := range transports() {
		for _, test := range tests {
			t.Run(via.name+": "+test.description, func(t *testing.T) {
				target, recorder := newTarget(t, test.targetSecure)
				redirect := newRedirector(t, test.redirectSecure)
				status := http.StatusTemporaryRedirect
				if test.status != 0 {
					status = test.status
				}
				redirect.answerWith(status, test.location(target))

				err := via.send(t, redirect.server.URL)
				if test.reachesTarget {
					require.NoError(t, err)
					assert.Equal(t, []string{deviceToken}, recorder.tokens())

					return
				}

				if test.refusal != nil {
					require.ErrorIs(t, err, test.refusal)
				} else {
					require.ErrorContains(t, err, test.failure)
				}
				assert.Empty(t, recorder.tokens())
				assert.Equal(t, int32(1), redirect.hits.Load())
			})
		}
	}
}

func TestTunnelDialReportsARedirectWithoutALocation(t *testing.T) {
	redirect := newRedirector(t, false)
	redirect.answerWith(http.StatusTemporaryRedirect, "")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _, err := DialContext(ctx, redirect.server.URL, http.Header{"Authorization": {deviceToken}}) //nolint:bodyclose // gorilla/websocket documents that the handshake response body need not be closed.
	require.ErrorIs(t, err, http.ErrNoLocation)
	assert.Equal(t, int32(1), redirect.hits.Load())
}

func TestRedirectsStopAfterThree(t *testing.T) {
	for _, via := range transports() {
		t.Run(via.name, func(t *testing.T) {
			redirect := newRedirector(t, false)
			redirect.redirectTo(redirect.server.URL)

			err := via.send(t, redirect.server.URL)
			require.ErrorIs(t, err, errTooManyRedirects)
			assert.Equal(t, int32(4), redirect.hits.Load())
		})
	}
}

func TestARefusedRedirectIsLoggedWithTheRuleThatRefusedIt(t *testing.T) {
	target, _ := newTarget(t, false)
	redirect := newRedirector(t, false)
	redirect.redirectTo(onOtherHost(target.URL))

	backend, hook := logtest.NewNullLogger()
	backend.SetLevel(logrus.DebugLevel)

	cli, err := NewClient(redirect.server.URL, withTestTransport(t, 1), WithLogger(backend))
	require.NoError(t, err)

	_, err = cli.GetDeviceAuthStatus("device-token")
	require.Error(t, err)

	var refusals []error
	for _, entry := range hook.AllEntries() {
		if logged, ok := entry.Data[logrus.ErrorKey].(error); ok {
			refusals = append(refusals, logged)
		}
	}

	require.NotEmpty(t, refusals)
	for _, logged := range refusals {
		assert.ErrorIs(t, logged, errRedirectOtherHost)
	}
}
