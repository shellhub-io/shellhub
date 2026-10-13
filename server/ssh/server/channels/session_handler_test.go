package channels

import (
	"context"
	"errors"
	"io"
	"net"
	"slices"
	"sync"
	"testing"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/envs/envstest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/ssh/session"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

type fakeSession struct {
	mu sync.Mutex

	newSeatErr    error
	agentChanErr  error
	clientChanErr error
	recordedErr   error

	calls    []string
	dropped  []int
	forwards chan gossh.NewChannel

	agentChannel gossh.Channel

	agent             *session.AgentChannel
	client            *session.ClientChannel
	onCloseAgentWrite func()
}

func (f *fakeSession) record(name string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.calls = append(f.calls, name)
}

func (f *fakeSession) called(name string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return slices.Contains(f.calls, name)
}

func (f *fakeSession) LogFields() log.Fields { return log.Fields{"uid": "fake"} }

func (f *fakeSession) NewSeat() (int, error) {
	f.record("NewSeat")

	return 0, f.newSeatErr
}

func (f *fakeSession) NewAgentChannel(_ string, _ int) (*session.AgentChannel, error) {
	f.record("NewAgentChannel")

	if f.agentChanErr != nil {
		return nil, f.agentChanErr
	}

	if f.agent != nil {
		return f.agent, nil
	}

	return &session.AgentChannel{Channel: &nopChannel{}, Requests: closedRequests()}, nil
}

func (f *fakeSession) NewClientChannel(_ gossh.NewChannel, _ int) (*session.ClientChannel, error) {
	f.record("NewClientChannel")

	if f.clientChanErr != nil {
		return nil, f.clientChanErr
	}

	if f.client != nil {
		return f.client, nil
	}

	return &session.ClientChannel{Channel: &nopChannel{}, Requests: closedRequests()}, nil
}

func (f *fakeSession) DropAgentChannel(seat int) {
	f.record("DropAgentChannel")

	f.mu.Lock()
	defer f.mu.Unlock()

	f.dropped = append(f.dropped, seat)
}

func (f *fakeSession) OpenAgentForwards(string) <-chan gossh.NewChannel { return f.forwards }

func (f *fakeSession) DialAgent(context.Context, string, string) (net.Conn, error) {
	f.record("DialAgent")

	return nil, errors.New("no agent")
}

func (f *fakeSession) CloseAgentWrite(int) error {
	f.record("CloseAgentWrite")

	if f.onCloseAgentWrite != nil {
		f.onCloseAgentWrite()

		return nil
	}

	if f.agentChannel == nil {
		return nil
	}

	return f.agentChannel.CloseWrite()
}

func (f *fakeSession) Recorded(int) error            { return f.recordedErr }
func (f *fakeSession) Announce(gossh.Channel) error  { return nil }
func (f *fakeSession) Event(string, any, int)        { f.record("Event") }
func (f *fakeSession) Seat(int) (session.Seat, bool) { return session.Seat{}, true }
func (f *fakeSession) SetSeatPty(int, bool)          { f.record("SetSeatPty") }
func (f *fakeSession) SetSeatType(int, string)       { f.record("SetSeatType") }

type nopChannel struct{}

func (*nopChannel) Read([]byte) (int, error)    { return 0, io.EOF }
func (*nopChannel) Write(p []byte) (int, error) { return len(p), nil }
func (*nopChannel) Close() error                { return nil }
func (*nopChannel) CloseWrite() error           { return nil }
func (*nopChannel) Stderr() io.ReadWriter       { return nil }

func (*nopChannel) SendRequest(_ string, _ bool, _ []byte) (bool, error) { return true, nil }

func closedRequests() <-chan *gossh.Request {
	requests := make(chan *gossh.Request)
	close(requests)

	return requests
}

type rejectingNewChannel struct {
	mu       sync.Mutex
	rejected bool
	reason   gossh.RejectionReason
	message  string
	accepted bool

	extraData []byte
}

func (n *rejectingNewChannel) Accept() (gossh.Channel, <-chan *gossh.Request, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.accepted = true

	return &nopChannel{}, closedRequests(), nil
}

func (n *rejectingNewChannel) Reject(reason gossh.RejectionReason, message string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.rejected = true
	n.reason = reason
	n.message = message

	return nil
}

func (n *rejectingNewChannel) ChannelType() string { return SessionChannel }
func (n *rejectingNewChannel) ExtraData() []byte   { return n.extraData }

type stubContext struct {
	context.Context
	sync.Mutex
	values map[any]any
}

func newStubContext() *stubContext {
	return &stubContext{Context: context.Background(), values: map[any]any{}}
}

func (s *stubContext) User() string          { return "user@namespace.device" }
func (s *stubContext) SessionID() string     { return "test-session-id" }
func (s *stubContext) ClientVersion() string { return "" }
func (s *stubContext) ServerVersion() string { return "" }
func (s *stubContext) RemoteAddr() net.Addr  { return nil }
func (s *stubContext) LocalAddr() net.Addr   { return nil }

func (s *stubContext) Permissions() *gliderssh.Permissions { return &gliderssh.Permissions{} }

func (s *stubContext) SetValue(key, val any) {
	s.Lock()
	defer s.Unlock()

	s.values[key] = val
}

func (s *stubContext) Value(key any) any {
	s.Lock()
	defer s.Unlock()

	return s.values[key]
}

func TestSessionChannelRejectsWhenTheSeatCannotBeCreated(t *testing.T) {
	sess := &fakeSession{newSeatErr: errors.New("no seats left")}
	newChan := new(rejectingNewChannel)

	sessionChannel(newStubContext(), sess, newChan)

	assert.True(t, newChan.rejected, "the channel must be rejected")
	assert.False(t, newChan.accepted, "the client channel must not be accepted")
	assert.False(t, sess.called("NewAgentChannel"), "no agent channel for a session with no seat")
}

// The agent channel is opened before the client's is accepted, so a device that
// dropped can still be reported: rejecting after accepting reaches nobody.
func TestSessionChannelRejectsWhenTheAgentChannelFails(t *testing.T) {
	sess := &fakeSession{agentChanErr: errors.New("agent is gone")}
	newChan := new(rejectingNewChannel)

	sessionChannel(newStubContext(), sess, newChan)

	assert.True(t, newChan.rejected, "the channel must be rejected")
	assert.False(t, newChan.accepted, "the client channel must not be accepted")
	require.True(t, sess.called("NewAgentChannel"))
	assert.False(t, sess.called("NewClientChannel"), "the client channel must not be opened")
}

// A client channel that fails after the agent channel opened must not leave the
// agent side live behind an unusable seat.
func TestSessionChannelDropsTheAgentChannelWhenTheClientChannelFails(t *testing.T) {
	sess := &fakeSession{clientChanErr: errors.New("client hung up")}
	newChan := new(rejectingNewChannel)

	sessionChannel(newStubContext(), sess, newChan)

	assert.True(t, newChan.rejected, "the channel must be rejected")
	require.True(t, sess.called("DropAgentChannel"), "the agent channel must be dropped")
	assert.Equal(t, []int{0}, sess.dropped)
}

var _ Session = (*fakeSession)(nil)

func forwardPayload(t *testing.T) []byte {
	t.Helper()

	return gossh.Marshal(&forwardData{
		DestAddr:   "10.0.0.1",
		DestPort:   8080,
		OriginAddr: "127.0.0.1",
		OriginPort: 1234,
	})
}

func TestDirectTCPIPChannelRejectsAnUnparsablePayload(t *testing.T) {
	newChan := &rejectingNewChannel{extraData: []byte{0x00}}
	sess := new(fakeSession)

	allow := func(gliderssh.Context, string, uint32) bool { return true }
	directTCPIPChannel(newStubContext(), sess, newChan, allow)

	assert.True(t, newChan.rejected, "the channel must be rejected")
	assert.Equal(t, gossh.ConnectionFailed, newChan.reason)
	assert.False(t, sess.called("DialAgent"), "a payload that does not parse must not reach the agent")
}

func TestDirectTCPIPChannelRejectsWhenForwardingIsDisabled(t *testing.T) {
	cases := map[string]gliderssh.LocalPortForwardingCallback{
		"no callback configured": nil,
		"callback refuses":       func(gliderssh.Context, string, uint32) bool { return false },
	}

	for name, allow := range cases {
		t.Run(name, func(t *testing.T) {
			newChan := &rejectingNewChannel{extraData: forwardPayload(t)}
			sess := new(fakeSession)

			directTCPIPChannel(newStubContext(), sess, newChan, allow)

			assert.True(t, newChan.rejected, "the channel must be rejected")
			assert.Equal(t, gossh.Prohibited, newChan.reason)
			assert.False(t, newChan.accepted, "the client channel must not be accepted")
			assert.False(t, sess.called("DialAgent"), "a refused forward must not reach the agent")
		})
	}
}

// The agent is dialled before the client's channel is accepted, so a device
// that cannot reach the destination is reported rather than silently accepted.
func TestDirectTCPIPChannelRejectsWhenTheAgentCannotDial(t *testing.T) {
	newChan := &rejectingNewChannel{extraData: forwardPayload(t)}
	sess := new(fakeSession)

	allow := func(gliderssh.Context, string, uint32) bool { return true }
	directTCPIPChannel(newStubContext(), sess, newChan, allow)

	require.True(t, sess.called("DialAgent"))
	assert.True(t, newChan.rejected, "the channel must be rejected")
	assert.Equal(t, gossh.ConnectionFailed, newChan.reason)
	assert.False(t, newChan.accepted, "the client channel must not be accepted")
}

func TestSessionChannelEndsWhenItsChannelsEndBeforeAShellOrExecStarts(t *testing.T) {
	ended := make(chan struct{})

	go func() {
		defer close(ended)

		sessionChannel(newStubContext(), &fakeSession{}, new(rejectingNewChannel))
	}()

	select {
	case <-ended:
	case <-time.After(5 * time.Second):
		t.Fatal("the session channel outlived both of its ends")
	}
}

type clientEnd struct {
	*fakeChannel

	mu        sync.Mutex
	sent      []string
	requests  chan *gossh.Request
	closeOnce sync.Once
}

func (c *clientEnd) SendRequest(name string, _ bool, _ []byte) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.sent = append(c.sent, name)

	return true, nil
}

func (c *clientEnd) Close() error {
	c.closeOnce.Do(func() { close(c.requests) })

	return c.fakeChannel.Close()
}

func (c *clientEnd) received(name string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return slices.Contains(c.sent, name)
}

func TestSessionChannelDoesNotReportTheDeviceEndingAChannelTheGatewayClosed(t *testing.T) {
	envstest.SetEdition(t, envs.Community)

	agentRequests := make(chan *gossh.Request)
	agent := &session.AgentChannel{Channel: newFakeChannel(), Requests: agentRequests}

	client := &clientEnd{fakeChannel: newFakeChannel(), requests: make(chan *gossh.Request, 1)} //nolint:exhaustruct // mu, sent and closeOnce start at their zero values
	client.quiet()
	client.requests <- &gossh.Request{Type: ExecRequestType, Payload: gossh.Marshal(models.SSHCommand{Command: "sleep 5"})} //nolint:exhaustruct // the exec wants no reply

	sess := &fakeSession{ //nolint:exhaustruct // only the agent and client fakes matter here
		agent:  agent,
		client: &session.ClientChannel{Channel: client, Requests: client.requests},
		onCloseAgentWrite: func() {
			_ = agent.Close()
			close(agentRequests)
		},
	}

	ended := make(chan struct{})

	go func() {
		defer close(ended)

		sessionChannel(newStubContext(), sess, new(rejectingNewChannel))
	}()

	select {
	case <-ended:
	case <-time.After(5 * time.Second):
		t.Fatal("the session channel did not end after the gateway closed its agent channel")
	}

	assert.False(t, client.received(ExitSignalRequest), "the client was told the device ended a channel the gateway closed")
}
