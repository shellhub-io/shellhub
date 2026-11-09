package dialer

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net"
	"testing"
	"time"

	"github.com/multiformats/go-multistream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func agentSpeaking(t *testing.T, protocols ...string) (net.Conn, <-chan map[string]string) {
	t.Helper()

	server, agent := net.Pipe()
	t.Cleanup(func() {
		_ = server.Close()
		_ = agent.Close()
	})

	headers := make(chan map[string]string, 1)
	mux := multistream.NewMultistreamMuxer[string]()

	for _, protocol := range protocols {
		mux.AddHandler(protocol, func(_ string, rwc io.ReadWriteCloser) error {
			var header map[string]string
			if err := json.NewDecoder(rwc).Decode(&header); err != nil {
				return err
			}

			headers <- header

			_, err := rwc.Write([]byte("SSH-2.0-agent\r\n"))

			return err
		})
	}

	go func() {
		_ = mux.Handle(agent)
		_ = agent.Close()
	}()

	return server, headers
}

func TestSSHOpenPipelined(t *testing.T) {
	t.Run("sends the selection and the envelope before the agent answers", func(t *testing.T) {
		server, agent := net.Pipe()
		t.Cleanup(func() {
			_ = server.Close()
			_ = agent.Close()
		})

		received := make(chan map[string]string, 1)

		go func() {
			reader := bufio.NewReader(agent)

			for range 2 {
				if _, err := multistream.ReadNextToken[string](reader); err != nil {
					return
				}
			}

			var header map[string]string
			if err := json.NewDecoder(reader).Decode(&header); err != nil {
				return
			}

			received <- header
		}()

		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		defer cancel()

		_, err := handshake(ctx, server, TransportVersion2, SSHOpenTarget{SessionID: "session"})
		require.NoError(t, err)

		select {
		case header := <-received:
			assert.Equal(t, map[string]string{"id": "session"}, header)
		case <-time.After(5 * time.Second):
			t.Fatal("the envelope never reached an agent that had not answered the selection")
		}
	})

	t.Run("hands back the agent's greeting after the selection's confirmation", func(t *testing.T) {
		server, headers := agentSpeaking(t, ProtoSSHOpen)

		conn, err := handshake(t.Context(), server, TransportVersion2, SSHOpenTarget{SessionID: "session"})
		require.NoError(t, err)

		greeting, err := bufio.NewReader(conn).ReadString('\n')
		require.NoError(t, err)
		assert.Equal(t, "SSH-2.0-agent\r\n", greeting)
		assert.Equal(t, map[string]string{"id": "session"}, <-headers)
	})

	t.Run("surfaces an agent refusing the protocol on the first read", func(t *testing.T) {
		server, _ := agentSpeaking(t, ProtoSSHClose)

		conn, err := handshake(t.Context(), server, TransportVersion2, SSHOpenTarget{SessionID: "session"})
		require.NoError(t, err)

		require.NoError(t, conn.SetDeadline(time.Now().Add(5*time.Second))) //nolint:forbidigo // a network deadline needs the wall clock

		_, err = conn.Read(make([]byte, 1))
		require.ErrorAs(t, err, &multistream.ErrNotSupported[string]{})
	})

	t.Run("closes without waiting for an agent that never answers the selection", func(t *testing.T) {
		server, agent := net.Pipe()
		t.Cleanup(func() {
			_ = server.Close()
			_ = agent.Close()
		})

		go func() { _, _ = io.Copy(io.Discard, agent) }()

		conn, err := handshake(t.Context(), server, TransportVersion2, SSHOpenTarget{SessionID: "session"})
		require.NoError(t, err)

		closed := make(chan error, 1)
		go func() { closed <- conn.Close() }()

		select {
		case err := <-closed:
			require.NoError(t, err)
		case <-time.After(5 * time.Second):
			t.Fatal("closing waited for a confirmation the agent never sends")
		}
	})
}

func TestSSHOpenEnvelope(t *testing.T) {
	cases := []struct {
		description string
		target      SSHOpenTarget
		expected    map[string]string
	}{
		{
			description: "names only the session for an agent left to check the login",
			target:      SSHOpenTarget{SessionID: "session"},
			expected:    map[string]string{"id": "session"},
		},
		{
			description: "carries the stream's secret",
			target:      SSHOpenTarget{SessionID: "session", Login: "secret"},
			expected:    map[string]string{"id": "session", "login": "secret"},
		},
		{
			description: "vouches for the login",
			target:      SSHOpenTarget{SessionID: "session", Login: "secret", Vouched: true},
			expected:    map[string]string{"id": "session", "login": "secret", "vouched": "true"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.target.envelope())
		})
	}
}
