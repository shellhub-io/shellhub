package session

import (
	"errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	servicemocks "github.com/shellhub-io/shellhub/server/api/services/mocks"
	"github.com/stretchr/testify/mock"
	gossh "golang.org/x/crypto/ssh"
)

// drained feeds the loop and asserts it keeps consuming until the channel is
// closed, which is the only thing that may end it.
func drained(t *testing.T, sess *Session, reqs chan *gossh.Request, send int) {
	t.Helper()

	done := make(chan struct{})

	go func() {
		defer close(done)

		sess.drainAgentRequests(newStubContext(), reqs)
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
		KeepAliveSession(mock.Anything, models.UID("test-uid")).
		Return(errors.New("store is down"))

	sess := newTestSession(serviceMock)

	drained(t, sess, make(chan *gossh.Request), 32)
}

func TestDrainAgentRequestsRefusesOtherRequests(t *testing.T) {
	sess := newTestSession(servicemocks.NewMockService(t))

	reqs := make(chan *gossh.Request)
	done := make(chan struct{})

	go func() {
		defer close(done)

		sess.drainAgentRequests(newStubContext(), reqs)
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

func TestDrainAgentRequestsEndsWithTheAgentConnection(t *testing.T) {
	serviceMock := servicemocks.NewMockService(t)
	serviceMock.EXPECT().
		KeepAliveSession(mock.Anything, models.UID("test-uid")).
		Return(nil)

	sess := newTestSession(serviceMock)

	reqs := make(chan *gossh.Request)
	done := make(chan struct{})

	go func() {
		defer close(done)

		sess.drainAgentRequests(newStubContext(), reqs)
	}()

	reqs <- &gossh.Request{Type: "keepalive"} //nolint:exhaustruct
	close(reqs)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("the drain outlived the agent connection")
	}
}
