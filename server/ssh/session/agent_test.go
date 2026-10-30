package session

import (
	"io"
	"net"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

type recordingChannel struct {
	closed      bool
	closedWrite bool
}

func (c *recordingChannel) Read([]byte) (int, error)    { return 0, io.EOF }
func (c *recordingChannel) Write(p []byte) (int, error) { return len(p), nil }
func (c *recordingChannel) Stderr() io.ReadWriter       { return nil }

func (c *recordingChannel) SendRequest(_ string, _ bool, _ []byte) (bool, error) {
	return true, nil
}

func (c *recordingChannel) Close() error {
	c.closed = true

	return nil
}

func (c *recordingChannel) CloseWrite() error {
	c.closedWrite = true

	return nil
}

// Agents before v0.9.3 hang after an exec when only the write side is closed,
// so those get a full close instead.
func TestCloseAgentWrite(t *testing.T) {
	cases := []struct {
		name        string
		version     string
		seatType    string
		expectClose bool
	}{
		{
			name:        "closes fully for an exec on an agent older than v0.9.3",
			version:     "v0.9.2",
			seatType:    SeatTypeExec,
			expectClose: true,
		},
		{
			name:        "closes only the write side for an exec on v0.9.3",
			version:     "v0.9.3",
			seatType:    SeatTypeExec,
			expectClose: false,
		},
		{
			name:        "closes only the write side for an exec on a newer agent",
			version:     "v0.9.4",
			seatType:    SeatTypeExec,
			expectClose: false,
		},
		{
			name:        "closes only the write side for a shell on an old agent",
			version:     "v0.9.2",
			seatType:    "shell",
			expectClose: false,
		},
		{
			name:        "closes only the write side when the version does not parse",
			version:     "not-a-version",
			seatType:    SeatTypeExec,
			expectClose: false,
		},
		{
			name:        "closes only the write side when the version is empty",
			version:     "",
			seatType:    SeatTypeExec,
			expectClose: false,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			channel := new(recordingChannel)
			sess := newAgentSession(t, tt.version)

			seat, err := sess.NewSeat()
			require.NoError(t, err)
			sess.SetSeatType(seat, tt.seatType)
			agentChannel := &AgentChannel{Channel: channel} //nolint:exhaustruct // CloseAgentWrite never reads Requests, and the closed flag starts false
			sess.agent.channels[seat] = agentChannel

			require.NoError(t, sess.CloseAgentWrite(seat))

			assert.Equal(t, tt.expectClose, channel.closed, "full close")
			assert.Equal(t, !tt.expectClose, channel.closedWrite, "write-side close")
			assert.Equal(t, tt.expectClose, agentChannel.ClosedByGateway(), "closed by the gateway")
		})
	}
}

// A seat with no agent channel is reported rather than dereferenced.
func TestCloseAgentWriteOnAnUnknownSeat(t *testing.T) {
	channel := new(recordingChannel)
	sess := newAgentSession(t, "v0.9.2")

	require.ErrorIs(t, sess.CloseAgentWrite(404), ErrSeatNotFound)
	assert.False(t, channel.closed)
	assert.False(t, channel.closedWrite)
}

type halfClosableConn struct {
	net.Conn

	closed      bool
	closedWrite bool
}

func (c *halfClosableConn) Close() error {
	c.closed = true

	return nil
}

func (c *halfClosableConn) CloseWrite() error {
	c.closedWrite = true

	return nil
}

type fullCloseOnlyConn struct {
	net.Conn

	closed bool
}

func (c *fullCloseOnlyConn) Close() error {
	c.closed = true

	return nil
}

func TestCloseAgentForwardWrite(t *testing.T) {
	cases := []struct {
		name             string
		version          string
		expectCloseWrite bool
	}{
		{
			name:             "leaves the forward open on an agent that hangs up on EOF",
			version:          "v0.29.0",
			expectCloseWrite: false,
		},
		{
			name:             "closes the write side on the first agent that half-closes forwards",
			version:          "v0.29.1",
			expectCloseWrite: true,
		},
		{
			name:             "closes the write side on a newer agent",
			version:          "v0.29.2",
			expectCloseWrite: true,
		},
		{
			name:             "closes the write side when the version does not parse",
			version:          "not-a-version",
			expectCloseWrite: true,
		},
		{
			name:             "closes the write side when the version is empty",
			version:          "",
			expectCloseWrite: true,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			conn := new(halfClosableConn)
			sess := newAgentSession(t, tt.version)

			require.NoError(t, sess.CloseAgentForwardWrite(conn))

			assert.Equal(t, tt.expectCloseWrite, conn.closedWrite, "write-side close")
			assert.False(t, conn.closed, "a forward is never closed fully on EOF")
		})
	}
}

func TestCloseAgentForwardWriteClosesAConnThatCannotHalfClose(t *testing.T) {
	conn := new(fullCloseOnlyConn)
	sess := newAgentSession(t, "v0.29.1")

	require.NoError(t, sess.CloseAgentForwardWrite(conn))

	assert.True(t, conn.closed)
}

func newAgentSession(t *testing.T, version string) *Session {
	t.Helper()

	return &Session{
		clock: clock.New(),
		seats: newSeats(),
		agent: &Agent{channels: make(map[int]*AgentChannel)},
		Data: Data{
			Device: &models.Device{
				UID:  "device-uid",
				Info: &models.DeviceInfo{Version: version},
			},
		},
	}
}

var _ gossh.Channel = (*recordingChannel)(nil)
