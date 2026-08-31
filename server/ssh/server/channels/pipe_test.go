package channels

import (
	"bytes"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/envs/envstest"
	"github.com/shellhub-io/shellhub/server/ssh/session"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gossh "golang.org/x/crypto/ssh"
)

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

type duplex struct {
	in  *stream
	out *stream
}

func (d duplex) Read(p []byte) (int, error)  { return d.in.Read(p) }
func (d duplex) Write(p []byte) (int, error) { return d.out.Write(p) }

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

func (f *fakeChannel) quiet() {
	f.dataIn.close()
	f.stderrIn.close()
}

func newPipeSession(t *testing.T) *fakeSession {
	t.Helper()

	envstest.SetEdition(t, envs.Community)

	return new(fakeSession)
}

func captureLogs(t *testing.T) *test.Hook {
	t.Helper()

	hook := test.NewGlobal()
	level := log.GetLevel()
	log.SetLevel(log.DebugLevel)

	t.Cleanup(func() {
		log.SetLevel(level)
		hook.Reset()
	})

	return hook
}

func runPipe(t *testing.T, sess *fakeSession, client, agent gossh.Channel) chan struct{} {
	t.Helper()

	sess.agentChannel = agent

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

// TestPipeReportsExpectedRecordingSkips is the regression: a session that was
// never going to be recorded — recording off for the namespace, or no pty to
// record — was logged as a recording failure, so on a fleet of non-interactive
// sessions every single one warned that a compliance feature had failed.
func TestPipeReportsExpectedRecordingSkips(t *testing.T) {
	tests := []struct {
		description string
		recordedErr error
	}{
		{
			description: "recording disabled for the namespace",
			recordedErr: session.ErrRecordingDisabled,
		},
		{
			description: "seat has no pty to record",
			recordedErr: session.ErrRecordingNoPty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			hook := captureLogs(t)

			sess := newPipeSession(t)
			envstest.SetEdition(t, envs.Enterprise)

			sess.recordedErr = tt.recordedErr

			client, agent := newFakeChannel(), newFakeChannel()

			client.quiet()
			agent.quiet()

			waitFor(t, runPipe(t, sess, client, agent))

			var skipped *log.Entry

			for _, entry := range hook.AllEntries() {
				assert.NotEqual(t, log.WarnLevel, entry.Level, "expected recording skip reported as a failure: %s", entry.Message)

				if entry.Message == "session will not be recorded" {
					skipped = entry
				}
			}

			require.NotNil(t, skipped, "expected the recording skip to be reported")
			assert.Equal(t, log.DebugLevel, skipped.Level)
		})
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

	assert.Never(t, client.closeWriteCalled, 200*time.Millisecond, 20*time.Millisecond,
		"client write side was closed while agent stderr was still open")

	_, err := agent.stderrIn.Write([]byte("late stderr"))
	require.NoError(t, err)

	agent.stderrIn.close()

	waitFor(t, finished)

	assert.True(t, client.closeWriteCalled())
	assert.Equal(t, "late stderr", client.stderrOut.String())
}
