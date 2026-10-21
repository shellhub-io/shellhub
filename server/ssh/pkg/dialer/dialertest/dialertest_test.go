package dialertest

import (
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/server/ssh/pkg/dialer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

var (
	_ dialer.TunnelDialer = (*Agent)(nil)
	_ dialer.TunnelDialer = (*Stub)(nil)
)

func clientConfig() *gossh.ClientConfig {
	return &gossh.ClientConfig{ //nolint:exhaustruct
		User:            "user",
		HostKeyCallback: gossh.InsecureIgnoreHostKey(), //nolint:gosec
	}
}

// TestSilentAgentNeverAnswers covers the device that holds a tunnel it can no longer serve:
// the handshake against it has nothing to wait on and only ends on a deadline.
func TestSilentAgentNeverAnswers(t *testing.T) {
	agent := NewSilentAgent(t)

	conn, err := agent.DialTo(t.Context(), "tenant", "device", nil)
	require.NoError(t, err)

	require.NoError(t, conn.SetDeadline(time.Now().Add(time.Second))) //nolint:forbidigo // a deadline, an elapsed-time measurement, or the clock mock itself

	_, _, _, err = gossh.NewClientConn(conn, "tcp", clientConfig())
	assert.Error(t, err, "a silent agent must not complete a handshake")
}

// TestAgentReportsABrokenTransport is what lets a caller assert that nothing wrote bytes the
// SSH transport cannot account for.
func TestAgentReportsABrokenTransport(t *testing.T) {
	agent := NewAgent(t)

	conn, err := agent.DialTo(t.Context(), "tenant", "device", nil)
	require.NoError(t, err)

	client, _, _, err := gossh.NewClientConn(conn, "tcp", clientConfig())
	require.NoError(t, err)

	defer client.Close() //nolint:errcheck

	require.True(t, agent.Serving(100*time.Millisecond))

	_, err = conn.Write([]byte("DELETE /ssh/close/session HTTP/1.1\r\nHost: agent\r\n\r\n"))
	require.NoError(t, err)

	assert.False(t, agent.Serving(time.Second), "a raw write into the SSH transport must break it")
}
