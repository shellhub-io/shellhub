package host

import (
	"bytes"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// capturingSession records what each stream received, so a test can tell stdout
// from stderr. The embedded fakeSession discards both.
type capturingSession struct {
	*fakeSession

	stdout *syncBuffer
	stderr *syncBuffer
}

func newCapturingSession() *capturingSession {
	return &capturingSession{
		fakeSession: newFakeSession("session-id", "user"),
		stdout:      new(syncBuffer),
		stderr:      new(syncBuffer),
	}
}

func (c *capturingSession) Write(p []byte) (int, error) { return c.stdout.Write(p) }

func (c *capturingSession) Stderr() io.ReadWriter { return c.stderr }

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *syncBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.buf.Write(p)
}

func (s *syncBuffer) Read(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.buf.Read(p)
}

func (s *syncBuffer) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.buf.Len()
}

func TestRelayOutputKeepsStreamsApart(t *testing.T) {
	session := newCapturingSession()

	stdout := bytes.NewBufferString("to stdout")
	stderr := bytes.NewBufferString("to stderr")

	wg := new(sync.WaitGroup)
	relayOutput(wg, session, stdout, stderr)
	wg.Wait()

	assert.Equal(t, "to stdout", session.stdout.String())
	assert.Equal(t, "to stderr", session.stderr.String())
}

// TestRelayOutputDrainsStderrWhileStdoutIsOpen is the regression this exists
// for. With io.MultiReader the stderr writer blocks once it fills the pipe
// buffer, because nothing reads it until stdout reaches EOF — and stdout only
// reaches EOF once the command exits, which it cannot do while blocked.
func TestRelayOutputDrainsStderrWhileStdoutIsOpen(t *testing.T) {
	session := newCapturingSession()

	stdoutReader, stdoutWriter, err := os.Pipe()
	require.NoError(t, err)

	stderrReader, stderrWriter, err := os.Pipe()
	require.NoError(t, err)

	wg := new(sync.WaitGroup)
	relayOutput(wg, session, stdoutReader, stderrReader)

	payload := bytes.Repeat([]byte("e"), 256*1024)

	written := make(chan error, 1)

	go func() {
		_, err := stderrWriter.Write(payload)
		written <- err
	}()

	select {
	case err := <-written:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("writing to stderr blocked while stdout was still open")
	}

	require.NoError(t, stderrWriter.Close())
	require.NoError(t, stdoutWriter.Close())

	wg.Wait()

	assert.Equal(t, len(payload), session.stderr.Len())
	assert.Zero(t, session.stdout.Len())
}

func (s *syncBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.buf.String()
}
