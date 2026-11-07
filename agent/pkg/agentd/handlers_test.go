package agentd

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/shellhub-io/shellhub/agent/pkg/keygen"
	"github.com/shellhub-io/shellhub/agent/pkg/tunnel"
	"github.com/shellhub-io/shellhub/agent/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func dialError(errno syscall.Errno) error {
	return &net.OpError{Op: "dial", Net: "tcp", Err: os.NewSyscallError("connect", errno)}
}

func TestDialFailureNamesTheCauseWithoutTheAddress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		description string
		err         error
		expected    error
	}{
		{
			description: "nothing listens on the port",
			err:         dialError(syscall.ECONNREFUSED),
			expected:    ErrProxyConnectionRefused,
		},
		{
			description: "the connect times out in the kernel",
			err:         dialError(syscall.ETIMEDOUT),
			expected:    ErrProxyConnectionTimedOut,
		},
		{
			description: "the host has no route",
			err:         dialError(syscall.EHOSTUNREACH),
			expected:    ErrProxyHostUnreachable,
		},
		{
			description: "the network has no route",
			err:         dialError(syscall.ENETUNREACH),
			expected:    ErrProxyHostUnreachable,
		},
		{
			description: "the name does not resolve",
			err:         &net.OpError{Op: "dial", Net: "tcp", Err: &net.DNSError{Name: "app.internal", IsNotFound: true}},
			expected:    ErrProxyHostNotFound,
		},
		{
			description: "any other failure",
			err:         errors.New("dial tcp 172.17.0.3:8080: something else"),
			expected:    ErrProxyDial,
		},
	}

	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			t.Parallel()

			assert.ErrorIs(t, dialFailure(tc.err), tc.expected)
		})
	}
}

type recordedStream struct {
	in  io.Reader
	out bytes.Buffer
}

func (s *recordedStream) Read(p []byte) (int, error) { return s.in.Read(p) }

func (s *recordedStream) Write(p []byte) (int, error) { return s.out.Write(p) }

func (*recordedStream) Close() error { return nil }

func TestHTTPProxyHandlerV2AnswersARefusedTargetWithItsCause(t *testing.T) {
	t.Parallel()

	listener, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	host, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)
	require.NoError(t, listener.Close())

	headers, err := json.Marshal(map[string]string{"id": "request", "host": host, "port": port})
	require.NoError(t, err)

	stream := &recordedStream{in: bytes.NewReader(append(headers, '\n'))}

	require.NoError(t, httpProxyHandlerV2(&Agent{})(tunnel.NewContext(t.Context(), stream), stream))

	assert.JSONEq(t, `{"error":"connection refused"}`, stream.out.String())
}

func TestSSHHandlerV2ForgetsASessionWhoseStreamEnded(t *testing.T) {
	t.Parallel()

	agent := &Agent{server: server.NewServer(nil, &server.Config{})}

	gateway, device := net.Pipe()

	headers, err := json.Marshal(map[string]string{"id": "session"})
	require.NoError(t, err)

	go func() {
		_, _ = gateway.Write(append(headers, '\n'))
		_ = gateway.Close()
	}()

	require.NoError(t, sshHandlerV2(agent)(tunnel.NewContext(t.Context(), device), device))

	_, kept := agent.server.Sessions.Load("session")
	assert.False(t, kept, "a stream that ended must not stay in the agent's sessions")
}

func TestSSHHandlerV2LeavesASessionThatTookOverTheID(t *testing.T) {
	t.Parallel()

	hostKey := filepath.Join(t.TempDir(), "host.key")
	require.NoError(t, keygen.GeneratePrivateKey(hostKey))

	agent := &Agent{server: server.NewServer(nil, &server.Config{PrivateKey: hostKey})}

	gateway, device := net.Pipe()
	replacement, _ := net.Pipe()

	headers, err := json.Marshal(map[string]string{"id": "session"})
	require.NoError(t, err)

	go func() {
		_, _ = gateway.Write(append(headers, '\n'))
		_, _ = gateway.Read(make([]byte, 1))
		agent.server.Sessions.Store("session", replacement)
		_ = gateway.Close()
	}()

	require.NoError(t, sshHandlerV2(agent)(tunnel.NewContext(t.Context(), device), device))

	kept, ok := agent.server.Sessions.Load("session")
	require.True(t, ok, "a stream that took over the id must survive the old stream ending")
	assert.Same(t, replacement, kept)
}
