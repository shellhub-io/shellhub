package channels

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"io"
	"net"
	"slices"
	"sync"
	"testing"
	"time"

	gliderssh "github.com/gliderlabs/ssh"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/ssh/session"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

type fakeSession struct {
	mu sync.Mutex

	newSeat       int
	newSeatErr    error
	agentChanErr  error
	clientChanErr error
	recordedErr   error

	calls    []string
	dropped  []int
	forwards chan gossh.NewChannel
	events   []recordedEvent

	agentConn net.Conn

	agentChannel gossh.Channel

	agent             *session.AgentChannel
	client            *session.ClientChannel
	onCloseAgentWrite func()

	edition envs.Edition
}

func (f *fakeSession) Edition() envs.Edition {
	return f.edition
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

	return f.newSeat, f.newSeatErr
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

	if f.agentConn != nil {
		return f.agentConn, nil
	}

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

func (f *fakeSession) Recorded(int) error           { return f.recordedErr }
func (f *fakeSession) Announce(gossh.Channel) error { return nil }

type recordedEvent struct {
	Type models.SessionEventType
	Seat int
}

func (f *fakeSession) Event(t string, _ any, seat int) {
	f.record("Event")

	f.mu.Lock()
	defer f.mu.Unlock()

	f.events = append(f.events, recordedEvent{Type: models.SessionEventType(t), Seat: seat})
}

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
	assert.Empty(t, sess.events, "a rejected forward records nothing")
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
			assert.Empty(t, sess.events, "a rejected forward records nothing")
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
	assert.Empty(t, sess.events, "a forward the agent could not open is not a port forward")
}

func TestDirectTCPIPChannelRecordsAnOpenedForwardOnASeatOfItsOwn(t *testing.T) {
	agent, device := net.Pipe()
	require.NoError(t, device.Close())

	newChan := &rejectingNewChannel{extraData: forwardPayload(t)}
	sess := &fakeSession{agentConn: agent, newSeat: 7}

	allow := func(gliderssh.Context, string, uint32) bool { return true }
	directTCPIPChannel(newStubContext(), sess, newChan, allow)

	assert.True(t, newChan.accepted, "the client channel must be accepted")
	assert.Equal(t, []recordedEvent{{Type: models.SessionEventTypeDirectTCPIP, Seat: 7}}, sess.events)
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

func sshClientOverLoopback(t *testing.T) (*gossh.Client, <-chan gossh.NewChannel) {
	t.Helper()

	_, key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	signer, err := gossh.NewSignerFromKey(key)
	require.NoError(t, err)

	serverConfig := &gossh.ServerConfig{NoClientAuth: true} //nolint:exhaustruct // a test server needs only a host key and no client auth
	serverConfig.AddHostKey(signer)

	listener, err := (&net.ListenConfig{}).Listen(t.Context(), "tcp", "127.0.0.1:0") //nolint:exhaustruct // the zero ListenConfig is the default one
	require.NoError(t, err)

	t.Cleanup(func() { _ = listener.Close() })

	type handshake struct {
		conn  *gossh.ServerConn
		chans <-chan gossh.NewChannel
		err   error
	}

	accepted := make(chan handshake, 1)

	go func() {
		serverSide, err := listener.Accept()
		if err != nil {
			accepted <- handshake{err: err} //nolint:exhaustruct // a failed accept has no connection

			return
		}

		conn, chans, requests, err := gossh.NewServerConn(serverSide, serverConfig)
		if err == nil {
			go gossh.DiscardRequests(requests)
		}

		accepted <- handshake{conn: conn, chans: chans, err: err}
	}()

	clientSide, err := (&net.Dialer{}).DialContext(t.Context(), "tcp", listener.Addr().String()) //nolint:exhaustruct // the zero Dialer is the default one
	require.NoError(t, err)

	clientConfig := &gossh.ClientConfig{User: "user", HostKeyCallback: gossh.InsecureIgnoreHostKey()} //nolint:exhaustruct,gosec // a loopback client to a NoClientAuth server needs only a user, and the server's host key is throwaway
	conn, chans, requests, err := gossh.NewClientConn(clientSide, listener.Addr().String(), clientConfig)
	require.NoError(t, err)

	server := <-accepted
	require.NoError(t, server.err)

	client := gossh.NewClient(conn, chans, requests)

	t.Cleanup(func() {
		_ = client.Close()
		_ = server.conn.Close()
	})

	return client, server.chans
}

type closeSignallingChannel struct {
	gossh.Channel

	closed    chan struct{}
	closeOnce sync.Once
}

func (c *closeSignallingChannel) Close() error {
	c.closeOnce.Do(func() { close(c.closed) })

	return c.Channel.Close()
}

type replyHoldingChannel struct {
	gossh.Channel

	release <-chan struct{}
}

func (c *replyHoldingChannel) SendRequest(name string, wantReply bool, payload []byte) (bool, error) {
	ok, err := c.Channel.SendRequest(name, wantReply, payload)

	select {
	case <-c.release:
	case <-time.After(500 * time.Millisecond):
	}

	return ok, err
}

func serveDeviceEndingOnExec(chans <-chan gossh.NewChannel, output string) {
	newChannel, ok := <-chans
	if !ok {
		return
	}

	channel, requests, err := newChannel.Accept()
	if err != nil {
		return
	}

	for req := range requests {
		if req.Type != ExecRequestType {
			_ = req.Reply(false, nil)

			continue
		}

		_ = req.Reply(true, nil)
		_, _ = channel.Write([]byte(output))
		_, _ = channel.SendRequest(ExitStatusRequest, false, gossh.Marshal(models.SSHExitStatus{Status: 0}))
		_ = channel.Close()
	}
}

func TestSessionChannelDeliversTheReplyAndOutputOfACommandTheDeviceEndsAtOnce(t *testing.T) {
	const deviceOutput = "command output"

	sshClient, gatewayChannels := sshClientOverLoopback(t)
	agentClient, deviceChannels := sshClientOverLoopback(t)

	go serveDeviceEndingOnExec(deviceChannels, deviceOutput)

	type opened struct {
		session *gossh.Session
		err     error
	}

	sessions := make(chan opened, 1)

	go func() {
		sess, err := sshClient.NewSession()
		sessions <- opened{session: sess, err: err}
	}()

	accepted, clientRequests, err := (<-gatewayChannels).Accept()
	require.NoError(t, err)

	opening := <-sessions
	require.NoError(t, opening.err)

	agentChannel, agentRequests, err := agentClient.OpenChannel(SessionChannel, nil)
	require.NoError(t, err)

	client := &closeSignallingChannel{Channel: accepted, closed: make(chan struct{})} //nolint:exhaustruct // closeOnce starts at its zero value

	sess := &fakeSession{ //nolint:exhaustruct // only the agent and client channels matter here
		agent:  &session.AgentChannel{Channel: &replyHoldingChannel{Channel: agentChannel, release: client.closed}, Requests: agentRequests},
		client: &session.ClientChannel{Channel: client, Requests: clientRequests},
	}

	ended := make(chan struct{})

	go func() {
		defer close(ended)

		sessionChannel(newStubContext(), sess, new(rejectingNewChannel))
	}()

	output, err := opening.session.Output("a command serveDeviceEndingOnExec ignores")
	require.NoError(t, err, "the client did not get the reply and exit status of its exec")
	assert.Equal(t, deviceOutput, string(output))

	select {
	case <-ended:
	case <-time.After(5 * time.Second):
		t.Fatal("the session channel did not end after the device ended the command")
	}
}
