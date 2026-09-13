package dialertest

import (
	"context"
	"errors"
	"net"
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

// TestAgentCompletesTheHandshake is the invariant the harness exists for: a caller that dials
// it gets a connection an SSH client can be built on, without a live agent anywhere.
func TestAgentCompletesTheHandshake(t *testing.T) {
	agent := NewAgent(t)

	conn, err := agent.DialTo(t.Context(), "tenant", "device", dialer.SSHOpenTarget{SessionID: "session"})
	require.NoError(t, err)

	client, chans, reqs, err := gossh.NewClientConn(conn, "tcp", clientConfig())
	require.NoError(t, err, "the fake agent must complete an SSH handshake")

	defer client.Close() //nolint:errcheck

	session, err := gossh.NewClient(client, chans, reqs).NewSession()
	require.NoError(t, err, "the fake agent must accept a session channel")

	assert.NoError(t, session.Close())
}

// TestAgentRecordsWhatItWasDialledFor pins the recording the SSH path's tests assert on.
func TestAgentRecordsWhatItWasDialledFor(t *testing.T) {
	agent := NewAgent(t)

	target := dialer.SSHOpenTarget{SessionID: "session"}

	_, err := agent.DialTo(t.Context(), "tenant", "device", target)
	require.NoError(t, err)

	assert.Equal(t, []Dial{{Tenant: "tenant", UID: "device", Target: target}}, agent.Dials())
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

// TestStubFailsEveryDial covers the failure paths the fake agent cannot reach.
func TestStubFailsEveryDial(t *testing.T) {
	for _, want := range []error{dialer.ErrInvalidArgument, dialer.ErrNoConnection, errors.New("unreachable")} {
		t.Run(want.Error(), func(t *testing.T) {
			stub := &Stub{Err: want} //nolint:exhaustruct

			conn, err := stub.DialTo(context.Background(), "tenant", "device", nil)

			assert.Equal(t, net.Conn(nil), conn)
			require.ErrorIs(t, err, want)
			assert.Len(t, stub.Dials(), 1)
		})
	}
}
