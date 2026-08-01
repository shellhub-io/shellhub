package channels

import (
	"bytes"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/envs/envstest"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/server/ssh/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

// stream is one direction of a channel: writes are buffered, reads block until
// there is something to read or the write side is closed.
type stream struct {
	mu     sync.Mutex
	cond   *sync.Cond
	buf    bytes.Buffer
	closed bool
}

func newStream() *stream {
	s := &stream{} //nolint:exhaustruct
	s.cond = sync.NewCond(&s.mu)

	return s
}

func (s *stream) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return 0, io.EOF
	}

	n, err := s.buf.Write(p)
	s.cond.Broadcast()

	return n, err
}

func (s *stream) Read(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for s.buf.Len() == 0 && !s.closed {
		s.cond.Wait()
	}

	if s.buf.Len() == 0 {
		return 0, io.EOF
	}

	return s.buf.Read(p)
}

func (s *stream) close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closed = true
	s.cond.Broadcast()
}

func (s *stream) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.buf.String()
}

// duplex pairs the two halves of an io.ReadWriter that are independent in the
// SSH protocol, as gossh.Channel.Stderr is.
type duplex struct {
	in  *stream
	out *stream
}

func (d duplex) Read(p []byte) (int, error)  { return d.in.Read(p) }
func (d duplex) Write(p []byte) (int, error) { return d.out.Write(p) }

// fakeChannel is a gossh.Channel with each direction of each stream separate, so
// a test can feed the peer's output and inspect what was relayed, and can
// observe when the write side is closed.
type fakeChannel struct {
	dataIn    *stream
	dataOut   *stream
	stderrIn  *stream
	stderrOut *stream

	mu           sync.Mutex
	closeWritten bool
}

func newFakeChannel() *fakeChannel {
	return &fakeChannel{ //nolint:exhaustruct
		dataIn:    newStream(),
		dataOut:   newStream(),
		stderrIn:  newStream(),
		stderrOut: newStream(),
	}
}

func (f *fakeChannel) Read(p []byte) (int, error)  { return f.dataIn.Read(p) }
func (f *fakeChannel) Write(p []byte) (int, error) { return f.dataOut.Write(p) }

func (f *fakeChannel) Stderr() io.ReadWriter { return duplex{in: f.stderrIn, out: f.stderrOut} }

func (f *fakeChannel) Close() error {
	f.dataIn.close()
	f.stderrIn.close()

	return nil
}

func (f *fakeChannel) CloseWrite() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.closeWritten = true

	return nil
}

func (f *fakeChannel) closeWriteCalled() bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return f.closeWritten
}

func (*fakeChannel) SendRequest(_ string, _ bool, _ []byte) (bool, error) { return true, nil }

// quiet marks the peer as having nothing left to send.
func (f *fakeChannel) quiet() {
	f.dataIn.close()
	f.stderrIn.close()
}

func newPipeSession(t *testing.T) *session.Session {
	t.Helper()

	// Recording is a separate concern from the data path; the community edition
	// keeps the recorder out of the writer set.
	envstest.SetEdition(t, envs.Community)

	sess := &session.Session{ //nolint:exhaustruct
		UID:   "session-uid",
		Seats: session.NewSeats(),
		Data: session.Data{ //nolint:exhaustruct
			SSHID: "user@namespace.device",
			Device: &models.Device{ //nolint:exhaustruct
				UID:  "device-uid",
				Info: &models.DeviceInfo{Version: "latest"}, //nolint:exhaustruct
			},
		},
	}

	_, err := sess.Seats.NewSeat()
	require.NoError(t, err)

	return sess
}

func runPipe(t *testing.T, sess *session.Session, client, agent gossh.Channel) chan struct{} {
	t.Helper()

	done := make(chan bool, 1)
	finished := make(chan struct{})

	go func() {
		defer close(finished)

		pipe(sess, client, agent, 0, done)
	}()

	return finished
}

func waitFor(t *testing.T, finished chan struct{}) {
	t.Helper()

	select {
	case <-finished:
	case <-time.After(10 * time.Second):
		t.Fatal("pipe did not return")
	}
}

// TestPipeKeepsStreamsApart is the regression: joining the two streams with
// io.MultiReader meant stderr was only read after stdout reached EOF, so it
// arrived out of order and tagged as ordinary data.
func TestPipeKeepsStreamsApart(t *testing.T) {
	sess := newPipeSession(t)
	client, agent := newFakeChannel(), newFakeChannel()

	_, err := agent.dataIn.Write([]byte("to stdout"))
	require.NoError(t, err)

	_, err = agent.stderrIn.Write([]byte("to stderr"))
	require.NoError(t, err)

	agent.quiet()
	client.quiet()

	waitFor(t, runPipe(t, sess, client, agent))

	assert.Equal(t, "to stdout", client.dataOut.String())
	assert.Equal(t, "to stderr", client.stderrOut.String())
}

// TestPipeDrainsAgentStderrBeforeStdoutEOF covers the hang: extended data sits
// in x/crypto's buffer and consumes the channel window, so a chatty stderr
// stalls the whole session if nothing reads it until stdout is done.
func TestPipeDrainsAgentStderrBeforeStdoutEOF(t *testing.T) {
	sess := newPipeSession(t)
	client, agent := newFakeChannel(), newFakeChannel()

	_, err := agent.stderrIn.Write([]byte("early stderr"))
	require.NoError(t, err)

	client.quiet()

	finished := runPipe(t, sess, client, agent)

	assert.Eventually(t, func() bool {
		return client.stderrOut.String() == "early stderr"
	}, 10*time.Second, 10*time.Millisecond, "stderr was not relayed while stdout was still open")

	agent.quiet()

	waitFor(t, finished)
}

// TestPipeClosesWriteAfterBothStreams pins the ordering: CloseWrite makes every
// later write return EOF, so firing it when only stdout is drained truncates
// stderr.
func TestPipeClosesWriteAfterBothStreams(t *testing.T) {
	sess := newPipeSession(t)
	client, agent := newFakeChannel(), newFakeChannel()

	client.quiet()
	agent.dataIn.close()

	finished := runPipe(t, sess, client, agent)

	// stdout is already at EOF; the agent's stderr is not.
	assert.Never(t, client.closeWriteCalled, 200*time.Millisecond, 20*time.Millisecond,
		"client write side was closed while agent stderr was still open")

	_, err := agent.stderrIn.Write([]byte("late stderr"))
	require.NoError(t, err)

	agent.stderrIn.close()

	waitFor(t, finished)

	assert.True(t, client.closeWriteCalled())
	assert.Equal(t, "late stderr", client.stderrOut.String())
}
