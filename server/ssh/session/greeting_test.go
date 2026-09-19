package session

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAwaitAgentGreetingHoldsTheHandshakeUntilTheAgentSpeaks is what keeps an agent from
// losing the first bytes of the handshake. Both ends of an SSH exchange write their version
// line before reading the other's, so the server's lands on a device that has not finished
// reading the header that opened the stream, and a device whose header decoder reads in
// blocks takes the version line with it. Hearing the device first makes that impossible:
// anything it sends proves it is past the header.
func TestAwaitAgentGreetingHoldsTheHandshakeUntilTheAgentSpeaks(t *testing.T) {
	ours, device := net.Pipe()
	defer ours.Close()   //nolint:errcheck // the assertions below report the failure, not the teardown
	defer device.Close() //nolint:errcheck // the assertions below report the failure, not the teardown

	returned := make(chan net.Conn, 1)

	go func() {
		greeted, err := awaitAgentGreeting(ours)
		if err != nil {
			close(returned)

			return
		}

		returned <- greeted
	}()

	select {
	case <-returned:
		t.Fatal("the handshake was let through before the agent had said anything")
	case <-time.After(200 * time.Millisecond):
	}

	const greeting = "SSH-2.0-agent\r\n"

	go device.Write([]byte(greeting)) //nolint:errcheck // a blocked write fails the read below

	greeted := <-returned
	require.NotNil(t, greeted, "hearing the agent out must not fail when the agent speaks")

	got := make([]byte, len(greeting))
	require.NoError(t, greeted.SetReadDeadline(time.Now().Add(5*time.Second))) //nolint:forbidigo // a deadline, so a regression fails the test instead of hanging it
	_, err := io.ReadFull(greeted, got)
	require.NoError(t, err)

	assert.Equal(t, greeting, string(got),
		"the byte read to hear the agent out must still reach the handshake")
}
