package session

import (
	"errors"
	"io"
	"net"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	servicemocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

func drained(t *testing.T, sess *Session, reqs chan *gossh.Request, send int) {
	t.Helper()

	done := make(chan struct{})

	go func() {
		defer close(done)

		sess.drainAgentRequests(newStubContext(), reqs, nil)
	}()

	for range send {
		select {
		case reqs <- &gossh.Request{Type: "keepalive@shellhub.io"}: //nolint:exhaustruct
		case <-time.After(5 * time.Second):
			t.Error("the drain stopped consuming requests")

			close(reqs)

			return
		}
	}

	close(reqs)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("the drain did not return after the requests channel closed")
	}
}

// TestDrainAgentRequestsSurvivesKeepAliveFailure is the regression. The agent's
// requests are buffered by a handful before its mux blocks on the send, and the
// agent keeps sending for as long as its connection lives, so a loop that gave
// up on a transient failure would freeze every channel on the connection.
func TestDrainAgentRequestsSurvivesKeepAliveFailure(t *testing.T) {
	serviceMock := servicemocks.NewMockService(t)
	serviceMock.EXPECT().
		KeepAliveSession(mock.Anything, tenantScope, models.UID("test-uid")).
		Return(errors.New("store is down"))

	sess := newTestSession(serviceMock, nil)

	drained(t, sess, make(chan *gossh.Request), 32)
}

func TestDrainAgentRequestsRefusesOtherRequests(t *testing.T) {
	sess := newTestSession(servicemocks.NewMockService(t), nil)

	reqs := make(chan *gossh.Request)
	done := make(chan struct{})

	go func() {
		defer close(done)

		sess.drainAgentRequests(newStubContext(), reqs, nil)
	}()

	for range 32 {
		reqs <- &gossh.Request{Type: "tcpip-forward"} //nolint:exhaustruct
	}

	close(reqs)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("the drain did not return after the requests channel closed")
	}
}

func TestDrainAgentRequestsClosesTheClientConnectionWhenTheAgentConnectionEnds(t *testing.T) {
	sess := newTestSession(servicemocks.NewMockService(t), nil)

	server, client := net.Pipe()
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.SetReadDeadline(time.Now().Add(5*time.Second))) //nolint:forbidigo // a deadline on a real pipe, which reads the wall clock

	reqs := make(chan *gossh.Request)
	close(reqs)

	sess.drainAgentRequests(newStubContext(), reqs, server)

	_, err := client.Read(make([]byte, 1))
	require.ErrorIs(t, err, io.EOF, "the client connection outlived the agent connection")
}

func TestDrainAgentRequestsKeepsTheClientConnectionUntilItsChannelsClose(t *testing.T) {
	sess := newTestSession(servicemocks.NewMockService(t), nil)

	channel, err := sess.NewClientChannel(&fakeNewChannel{}, 0)
	require.NoError(t, err)

	server, client := net.Pipe()
	t.Cleanup(func() { _ = client.Close() })
	require.NoError(t, client.SetReadDeadline(time.Now().Add(5*time.Second))) //nolint:forbidigo // a deadline on a real pipe, which reads the wall clock

	reqs := make(chan *gossh.Request)
	close(reqs)

	drained := make(chan struct{})

	go func() {
		defer close(drained)

		sess.drainAgentRequests(newStubContext(), reqs, server)
	}()

	select {
	case <-drained:
		t.Fatal("the client connection closed before its channel could tell the client how the session ended")
	case <-time.After(100 * time.Millisecond):
	}

	require.NoError(t, channel.Close())

	select {
	case <-drained:
	case <-time.After(5 * time.Second):
		t.Fatal("the client connection stayed open after its last channel closed")
	}

	_, err = client.Read(make([]byte, 1))
	require.ErrorIs(t, err, io.EOF)
}

func TestDrainAgentRequestsEndsWithTheAgentConnection(t *testing.T) {
	serviceMock := servicemocks.NewMockService(t)
	serviceMock.EXPECT().
		KeepAliveSession(mock.Anything, tenantScope, models.UID("test-uid")).
		Return(nil)

	sess := newTestSession(serviceMock, nil)

	reqs := make(chan *gossh.Request)
	done := make(chan struct{})

	go func() {
		defer close(done)

		sess.drainAgentRequests(newStubContext(), reqs, nil)
	}()

	reqs <- &gossh.Request{Type: "keepalive"} //nolint:exhaustruct
	close(reqs)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("the drain outlived the agent connection")
	}
}
